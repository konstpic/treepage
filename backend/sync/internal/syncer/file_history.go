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

type GitFileRevision struct {
	CommitSHA   string `json:"commit_sha"`
	AuthorName  string `json:"author_name"`
	Message     string `json:"message"`
	CreatedAt   string `json:"created_at"`
}

func (s *Syncer) FileHistory(ctx context.Context, repoID, relPath string, limit int) ([]GitFileRevision, error) {
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	var repo models.Repository
	if err := s.db.WithContext(ctx).First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, err
	}
	cloneDir, err := s.ensureRepoClone(ctx, &repo, repo.ID+"-history", 80)
	if err != nil {
		return nil, err
	}
	gitPath := filepath.ToSlash(filepath.Join(repo.DocsPath, strings.TrimPrefix(strings.ReplaceAll(relPath, "\\", "/"), "/")))
	out, err := s.gitOutput(ctx, cloneDir, "log", "--follow",
		fmt.Sprintf("-n%d", limit), "--format=%H%x09%an%x09%at%x09%s", "--", gitPath)
	if err != nil {
		return nil, err
	}
	return parseGitLog(out), nil
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
	cloneDir, err := s.ensureRepoClone(ctx, &repo, repo.ID+"-history", 80)
	if err != nil {
		return "", err
	}
	gitPath := filepath.ToSlash(filepath.Join(repo.DocsPath, strings.TrimPrefix(strings.ReplaceAll(relPath, "\\", "/"), "/")))
	spec := fmt.Sprintf("%s:%s", commitSHA, gitPath)
	rev, err := s.gitOutput(ctx, cloneDir, "rev-parse", spec)
	if err != nil {
		return "", err
	}
	out, err := s.gitOutput(ctx, cloneDir, "cat-file", "-p", strings.TrimSpace(rev))
	if err != nil {
		return "", err
	}
	return out, nil
}

func (s *Syncer) ensureRepoClone(ctx context.Context, repo *models.Repository, dirSuffix string, depth int) (string, error) {
	cloneDir := filepath.Join(s.workDir, dirSuffix)
	token := resolveCredential(repo.AccessTokenRef, s.token)
	cloneURL := authCloneURL(repo.URL, token)
	branch := repo.Branch
	if branch == "" {
		branch = "main"
	}
	if _, err := os.Stat(cloneDir); os.IsNotExist(err) {
		args := []string{"clone", "--depth", strconv.Itoa(depth), "--branch", branch, cloneURL, cloneDir}
		if err := s.runGit(ctx, s.workDir, args...); err != nil {
			return "", err
		}
		return cloneDir, nil
	}
	_ = s.runGit(ctx, cloneDir, "fetch", "--depth", strconv.Itoa(depth), "origin", branch)
	_ = s.runGit(ctx, cloneDir, "checkout", branch)
	_ = s.runGit(ctx, cloneDir, "reset", "--hard", "origin/"+branch)
	return cloneDir, nil
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
