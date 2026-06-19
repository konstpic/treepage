package syncer

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

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

func (s *Syncer) SyncRepository(ctx context.Context, repoID, trigger string) error {
	var repo models.Repository
	if err := s.db.WithContext(ctx).First(&repo, "id = ?", repoID).Error; err != nil {
		return err
	}

	job := models.SyncJob{RepositoryID: repoID, Status: "running", TriggerType: trigger}
	now := time.Now()
	job.StartedAt = &now
	if err := s.db.WithContext(ctx).Create(&job).Error; err != nil {
		return err
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
		s.finishJob(ctx, &job, &repo, 0, fmt.Errorf("git clone: %w: %s", err, string(out)))
		return err
	}

	docsRoot := filepath.Join(cloneDir, repo.DocsPath)
	if _, statErr := os.Stat(docsRoot); statErr != nil {
		syncErr := fmt.Errorf("docs path not found: %s", repo.DocsPath)
		s.finishJob(ctx, &job, &repo, 0, syncErr)
		return syncErr
	}
	count := 0
	err := filepath.WalkDir(docsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(docsRoot, path)
		if err := s.upsertDocument(ctx, repo, rel, string(content)); err != nil {
			s.logger.Warn("document upsert failed", zap.String("path", rel), zap.Error(err))
			return nil
		}
		count++
		return nil
	})

	s.finishJob(ctx, &job, &repo, count, err)
	return err
}

func (s *Syncer) upsertDocument(ctx context.Context, repo models.Repository, relPath, content string) error {
	title := strings.TrimSuffix(filepath.Base(relPath), ".md")
	if h1 := extractH1(content); h1 != "" {
		title = h1
	}
	slug := slugify(strings.TrimSuffix(relPath, ".md"))
	tags := extractTags(content)
	mermaidBlocks := mermaidRe.FindAllString(content, -1)
	_ = mermaidBlocks

	var doc models.Document
	err := s.db.WithContext(ctx).
		Where("space_id = ? AND slug = ?", repo.SpaceID, slug).
		First(&doc).Error

	if err == gorm.ErrRecordNotFound {
		doc = models.Document{
			SpaceID: repo.SpaceID, RepositoryID: &repo.ID,
			Slug: slug, Title: title, Path: relPath,
			Content: content, Tags: tags, IsPublished: true,
		}
		return s.db.WithContext(ctx).Create(&doc).Error
	}
	if err != nil {
		return err
	}
	doc.Title = title
	doc.Content = content
	doc.Tags = tags
	doc.RepositoryID = &repo.ID
	doc.Path = relPath
	return s.db.WithContext(ctx).Save(&doc).Error
}

func (s *Syncer) finishJob(ctx context.Context, job *models.SyncJob, repo *models.Repository, count int, syncErr error) {
	finished := time.Now()
	job.FinishedAt = &finished
	job.FilesProcessed = count
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

// resolveCredential uses per-repo ref from admin UI: env var name or literal token value.
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
		if err := s.SyncRepository(ctx, repo.ID, "scheduled"); err != nil {
			s.logger.Warn("scheduled sync failed", zap.String("repo", repo.Name), zap.Error(err))
		}
	}
}
