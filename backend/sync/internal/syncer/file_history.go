package syncer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
)

const (
	historyCacheTTL      = 3 * time.Minute
	cloneRefreshInterval = 5 * time.Minute
	historyCloneDepth    = 200
)

type GitFileRevision struct {
	CommitSHA  string `json:"commit_sha"`
	AuthorName string `json:"author_name"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
}

type historyCacheEntry struct {
	items     []GitFileRevision
	expiresAt time.Time
}

func (s *Syncer) FileHistory(ctx context.Context, repoID, relPath string, limit int) ([]GitFileRevision, error) {
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	cacheKey := fmt.Sprintf("%s:%s:%d", repoID, relPath, limit)
	if items, ok := s.cachedHistory(cacheKey); ok {
		return items, nil
	}

	var repo models.Repository
	if err := s.db.WithContext(ctx).First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, err
	}
	cloneDir, err := s.ensureRepoClone(ctx, &repo, repo.ID+"-history", historyCloneDepth)
	if err != nil {
		return nil, err
	}
	gitPath := gitDocPath(&repo, relPath)
	out, err := s.gitOutput(ctx, cloneDir, "log", "--follow",
		fmt.Sprintf("-n%d", limit), "--format=%H%x09%an%x09%at%x09%s", "--", gitPath)
	if err != nil {
		return nil, err
	}
	items := parseGitLog(out)
	s.storeHistory(cacheKey, items)
	return items, nil
}

func (s *Syncer) FileContentAt(ctx context.Context, repoID, relPath, commitSHA string) (string, error) {
	commitSHA = strings.TrimSpace(commitSHA)
	if commitSHA == "" {
		return "", fmt.Errorf("commit sha is required")
	}
	var repo models.Repository
	if err := s.db.WithContext(ctx).First(&repo, "id = ?", repoID).Error; err != nil {
		return "", err
	}
	cloneDir, err := s.ensureRepoClone(ctx, &repo, repo.ID+"-history", historyCloneDepth)
	if err != nil {
		return "", err
	}
	if err := s.ensureCommit(ctx, cloneDir, commitSHA); err != nil {
		return "", err
	}
	gitPath := gitDocPath(&repo, relPath)
	out, err := s.gitOutput(ctx, cloneDir, "show", fmt.Sprintf("%s:%s", commitSHA, gitPath))
	if err != nil {
		return "", err
	}
	return out, nil
}

func gitDocPath(repo *models.Repository, relPath string) string {
	return filepath.ToSlash(filepath.Join(repo.DocsPath, strings.TrimPrefix(strings.ReplaceAll(relPath, "\\", "/"), "/")))
}

func (s *Syncer) cachedHistory(key string) ([]GitFileRevision, bool) {
	s.historyMu.Lock()
	defer s.historyMu.Unlock()
	entry, ok := s.historyCache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	out := make([]GitFileRevision, len(entry.items))
	copy(out, entry.items)
	return out, true
}

func (s *Syncer) storeHistory(key string, items []GitFileRevision) {
	s.historyMu.Lock()
	defer s.historyMu.Unlock()
	if s.historyCache == nil {
		s.historyCache = map[string]historyCacheEntry{}
	}
	copied := make([]GitFileRevision, len(items))
	copy(copied, items)
	s.historyCache[key] = historyCacheEntry{items: copied, expiresAt: time.Now().Add(historyCacheTTL)}
}

func (s *Syncer) ensureRepoClone(ctx context.Context, repo *models.Repository, dirSuffix string, depth int) (string, error) {
	cloneDir := filepath.Join(s.workDir, dirSuffix)
	branch := repo.Branch
	if branch == "" {
		branch = "main"
	}

	s.cloneMu.Lock()
	defer s.cloneMu.Unlock()

	if s.cloneFetched == nil {
		s.cloneFetched = map[string]time.Time{}
	}

	if _, err := os.Stat(cloneDir); os.IsNotExist(err) {
		token := resolveCredential(repo.AccessTokenRef, s.token)
		cloneURL := authCloneURL(repo.URL, token)
		args := []string{"clone", "--depth", strconv.Itoa(depth), "--branch", branch, cloneURL, cloneDir}
		if err := s.runGit(ctx, s.workDir, args...); err != nil {
			return "", err
		}
		s.cloneFetched[dirSuffix] = time.Now()
		return cloneDir, nil
	}

	if time.Since(s.cloneFetched[dirSuffix]) < cloneRefreshInterval {
		return cloneDir, nil
	}

	_ = s.runGit(ctx, cloneDir, "fetch", "--depth", strconv.Itoa(depth), "origin", branch)
	s.cloneFetched[dirSuffix] = time.Now()
	return cloneDir, nil
}

func (s *Syncer) ensureCommit(ctx context.Context, cloneDir, commitSHA string) error {
	if _, err := s.gitOutput(ctx, cloneDir, "cat-file", "-e", commitSHA+"^{commit}"); err == nil {
		return nil
	}
	return s.runGit(ctx, cloneDir, "fetch", "--depth=1", "origin", commitSHA)
}

func parseGitLog(raw string) []GitFileRevision {
	var items []GitFileRevision
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 4 {
			continue
		}
		sec, _ := strconv.ParseInt(parts[2], 10, 64)
		items = append(items, GitFileRevision{
			CommitSHA:  parts[0],
			AuthorName: parts[1],
			Message:    parts[3],
			CreatedAt:  time.Unix(sec, 0).UTC().Format(time.RFC3339),
		})
	}
	return items
}
