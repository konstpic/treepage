package syncer

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/contenthash"
	"github.com/konstpic/treepage/backend/pkg/models"
	"github.com/konstpic/treepage/backend/pkg/ragindex"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var ErrSyncConflict = errors.New("sync conflict: document has pending local changes")

type SyncResult struct {
	FilesProcessed   int
	ConflictsSkipped int
}

func (s *Syncer) SyncRepository(ctx context.Context, repoID, trigger string) (*SyncResult, error) {
	result := &SyncResult{}
	var repo models.Repository
	if err := s.db.WithContext(ctx).First(&repo, "id = ?", repoID).Error; err != nil {
		return result, err
	}

	job := models.SyncJob{RepositoryID: repoID, Status: "running", TriggerType: trigger}
	now := time.Now()
	job.StartedAt = &now
	if err := s.db.WithContext(ctx).Create(&job).Error; err != nil {
		return result, err
	}

	cloneDir := filepath.Join(s.workDir, repo.ID)
	_ = os.RemoveAll(cloneDir)

	cloneURL := repo.URL
	token := resolveCredential(repo.AccessTokenRef, s.token)
	if token != "" && strings.HasPrefix(cloneURL, "https://") {
		cloneURL = strings.Replace(cloneURL, "https://", fmt.Sprintf("https://oauth2:%s@", token), 1)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--branch", repo.Branch, cloneURL, cloneDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		syncErr := fmt.Errorf("git clone: %w: %s", err, string(out))
		s.finishJob(ctx, &job, &repo, result, syncErr)
		return result, syncErr
	}

	docsRoot := filepath.Join(cloneDir, repo.DocsPath)
	if _, statErr := os.Stat(docsRoot); statErr != nil {
		syncErr := fmt.Errorf("docs path not found: %s", repo.DocsPath)
		s.finishJob(ctx, &job, &repo, result, syncErr)
		return result, syncErr
	}

	repoHEAD, _ := s.gitOutput(ctx, cloneDir, "rev-parse", "HEAD")

	seenSlugs := make(map[string]struct{})

	err := filepath.WalkDir(docsRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(docsRoot, path)
		slug := slugify(strings.TrimSuffix(rel, ".md"))
		seenSlugs[slug] = struct{}{}
		applied, upsertErr := s.upsertDocument(ctx, repo, rel, string(content), strings.TrimSpace(repoHEAD))
		if errors.Is(upsertErr, ErrSyncConflict) {
			result.ConflictsSkipped++
			s.logger.Info("sync skipped document with pending local changes", zap.String("path", rel))
			return nil
		}
		if upsertErr != nil {
			s.logger.Warn("document upsert failed", zap.String("path", rel), zap.Error(upsertErr))
			return nil
		}
		if applied {
			result.FilesProcessed++
		}
		return nil
	})

	if err == nil {
		s.removeOrphanDocuments(ctx, repo, seenSlugs)
	}

	s.finishJob(ctx, &job, &repo, result, err)
	return result, err
}

func (s *Syncer) upsertDocument(ctx context.Context, repo models.Repository, relPath, content, commitSHA string) (bool, error) {
	title := strings.TrimSuffix(filepath.Base(relPath), ".md")
	if h1 := extractH1(content); h1 != "" {
		title = h1
	}
	slug := slugify(strings.TrimSuffix(relPath, ".md"))
	tags := extractTags(content)
	gitHash := contenthash.SHA256(content)
	now := time.Now()

	var doc models.Document
	err := s.db.WithContext(ctx).
		Where("space_id = ? AND slug = ?", repo.SpaceID, slug).
		First(&doc).Error

	if err == gorm.ErrRecordNotFound {
		doc = models.Document{
			SpaceID: repo.SpaceID, RepositoryID: &repo.ID,
			Slug: slug, Title: title, Path: relPath,
			Content: content, Tags: tags, IsPublished: true,
			SyncedContentHash: gitHash, HasPendingChanges: false, LastSyncedAt: &now,
			CommitSHA: commitSHA,
		}
		if err := s.db.WithContext(ctx).Create(&doc).Error; err != nil {
			return true, err
		}
		if err := ragindex.IndexDocument(ctx, s.db, &doc); err != nil {
			s.logger.Warn("rag index failed after create", zap.String("slug", slug), zap.Error(err))
		}
		return true, nil
	}
	if err != nil {
		return false, err
	}

	if doc.HasPendingChanges {
		return false, ErrSyncConflict
	}

	doc.Title = title
	doc.Content = content
	doc.Tags = tags
	doc.RepositoryID = &repo.ID
	doc.Path = relPath
	doc.SyncedContentHash = gitHash
	doc.HasPendingChanges = false
	doc.LastSyncedAt = &now
	if commitSHA != "" {
		doc.CommitSHA = commitSHA
	}
	if err := s.db.WithContext(ctx).Save(&doc).Error; err != nil {
		return true, err
	}
	if err := ragindex.IndexDocument(ctx, s.db, &doc); err != nil {
		s.logger.Warn("rag index failed after update", zap.String("slug", slug), zap.Error(err))
	}
	return true, nil
}

func (s *Syncer) removeOrphanDocuments(ctx context.Context, repo models.Repository, seen map[string]struct{}) {
	var docs []models.Document
	if err := s.db.WithContext(ctx).
		Where("repository_id = ? AND space_id = ?", repo.ID, repo.SpaceID).
		Find(&docs).Error; err != nil {
		s.logger.Warn("list repo documents for orphan cleanup failed", zap.Error(err))
		return
	}
	for _, doc := range docs {
		if _, ok := seen[doc.Slug]; ok {
			continue
		}
		if doc.HasPendingChanges {
			s.logger.Info("orphan document skipped (pending local changes)", zap.String("slug", doc.Slug))
			continue
		}
		if err := s.db.WithContext(ctx).Delete(&doc).Error; err != nil {
			s.logger.Warn("orphan document delete failed", zap.String("slug", doc.Slug), zap.Error(err))
		} else {
			s.logger.Info("orphan document removed after sync", zap.String("slug", doc.Slug))
		}
	}
}

func (s *Syncer) finishJob(ctx context.Context, job *models.SyncJob, repo *models.Repository, result *SyncResult, syncErr error) {
	finished := time.Now()
	job.FinishedAt = &finished
	job.FilesProcessed = result.FilesProcessed
	job.ConflictsSkipped = result.ConflictsSkipped
	if syncErr != nil {
		job.Status = "failed"
		job.ErrorMessage = syncErr.Error()
		repo.LastSyncStatus = "failed"
		repo.LastSyncError = syncErr.Error()
	} else {
		job.Status = "completed"
		repo.LastSyncStatus = "completed"
		repo.LastSyncError = ""
	}
	now := time.Now()
	repo.LastSyncAt = &now
	s.db.WithContext(ctx).Save(job)
	s.db.WithContext(ctx).Save(repo)
}

var mermaidRe = regexp.MustCompile("(?s)```mermaid\\n(.*?)```")

type Syncer struct {
	db      *gorm.DB
	workDir string
	token   string
	logger  *zap.Logger
}

func New(db *gorm.DB, workDir, token string, logger *zap.Logger) *Syncer {
	_ = os.MkdirAll(workDir, 0o755)
	return &Syncer{db: db, workDir: workDir, token: token, logger: logger}
}

func extractH1(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

func extractTags(content string) []string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "tags:") {
			raw := strings.TrimPrefix(line, "tags:")
			raw = strings.TrimPrefix(raw, "Tags:")
			parts := strings.Split(raw, ",")
			var tags []string
			for _, p := range parts {
				t := strings.TrimSpace(p)
				if t != "" {
					tags = append(tags, t)
				}
			}
			return tags
		}
	}
	return nil
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "/", "-")
	return strings.Trim(slugRe.ReplaceAllString(s, "-"), "-")
}

func resolveCredential(ref, fallback string) string {
	if ref != "" {
		if v := os.Getenv(ref); v != "" {
			return v
		}
		return ref
	}
	return fallback
}

func (s *Syncer) RunScheduled(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncAll(ctx)
		}
	}
}

func (s *Syncer) syncAll(ctx context.Context) {
	var repos []models.Repository
	if err := s.db.WithContext(ctx).Where("enabled = ?", true).Find(&repos).Error; err != nil {
		s.logger.Error("list repos failed", zap.Error(err))
		return
	}
	for _, repo := range repos {
		if _, err := s.SyncRepository(ctx, repo.ID, "scheduled"); err != nil {
			s.logger.Warn("scheduled sync failed", zap.String("repo", repo.Name), zap.Error(err))
		}
	}
}
