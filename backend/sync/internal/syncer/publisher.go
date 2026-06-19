package syncer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/models"
	"go.uber.org/zap"
)

type PublishInput struct {
	DocumentID    string `json:"document_id"`
	Path          string `json:"path"`
	Content       string `json:"content"`
	Branch        string `json:"branch"`
	CommitMessage string `json:"commit_message"`
	PRTitle       string `json:"pr_title"`
	PRBody        string `json:"pr_body"`
}

type PublishResult struct {
	Branch   string `json:"branch"`
	CommitSHA string `json:"commit_sha,omitempty"`
	PRURL    string `json:"pr_url,omitempty"`
	Message  string `json:"message,omitempty"`
}

var branchNameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)

func (s *Syncer) PublishDocument(ctx context.Context, repoID string, input PublishInput) (*PublishResult, error) {
	if strings.TrimSpace(input.Path) == "" {
		return nil, fmt.Errorf("document path is required")
	}
	if strings.TrimSpace(input.Branch) == "" {
		return nil, fmt.Errorf("branch name is required")
	}
	if !branchNameRe.MatchString(input.Branch) {
		return nil, fmt.Errorf("invalid branch name")
	}
	if strings.TrimSpace(input.CommitMessage) == "" {
		return nil, fmt.Errorf("commit message is required")
	}

	var repo models.Repository
	if err := s.db.WithContext(ctx).First(&repo, "id = ?", repoID).Error; err != nil {
		return nil, err
	}

	baseBranch := repo.Branch
	if baseBranch == "" {
		baseBranch = "main"
	}

	token := resolveCredential(repo.AccessTokenRef, s.token)
	if token == "" {
		return nil, fmt.Errorf("git access token is not configured")
	}

	cloneDir := filepath.Join(s.workDir, repo.ID+"-publish")
	_ = os.RemoveAll(cloneDir)

	cloneURL := authCloneURL(repo.URL, token)
	if err := s.runGit(ctx, s.workDir, "clone", "--branch", baseBranch, "--single-branch", cloneURL, cloneDir); err != nil {
		return nil, fmt.Errorf("git clone: %w", err)
	}

	authorName := "TreePage"
	authorEmail := "treepage@local"
	_ = s.runGit(ctx, cloneDir, "config", "user.name", authorName)
	_ = s.runGit(ctx, cloneDir, "config", "user.email", authorEmail)

	if err := s.runGit(ctx, cloneDir, "checkout", "-b", input.Branch); err != nil {
		return nil, fmt.Errorf("git checkout: %w", err)
	}

	relPath := filepath.FromSlash(strings.TrimPrefix(input.Path, "/"))
	targetFile := filepath.Join(cloneDir, repo.DocsPath, relPath)
	if err := os.MkdirAll(filepath.Dir(targetFile), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(targetFile, []byte(input.Content), 0o644); err != nil {
		return nil, err
	}

	relInRepo := filepath.ToSlash(filepath.Join(repo.DocsPath, relPath))
	if err := s.runGit(ctx, cloneDir, "add", relInRepo); err != nil {
		return nil, fmt.Errorf("git add: %w", err)
	}

	if err := s.runGit(ctx, cloneDir, "commit", "-m", input.CommitMessage); err != nil {
		return nil, fmt.Errorf("git commit: %w", err)
	}

	sha, err := s.gitOutput(ctx, cloneDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}

	if err := s.runGit(ctx, cloneDir, "push", "-u", "origin", input.Branch); err != nil {
		return nil, fmt.Errorf("git push: %w", err)
	}

	result := &PublishResult{
		Branch:    input.Branch,
		CommitSHA: strings.TrimSpace(sha),
	}

	if input.DocumentID != "" {
		s.db.WithContext(ctx).Model(&models.Document{}).
			Where("id = ?", input.DocumentID).
			Update("commit_sha", result.CommitSHA)
	}

	prTitle := strings.TrimSpace(input.PRTitle)
	if prTitle == "" {
		prTitle = input.CommitMessage
	}

	switch strings.ToLower(repo.Provider) {
	case "github", "":
		if prURL, err := createGitHubPR(ctx, repo.URL, token, baseBranch, input.Branch, prTitle, input.PRBody); err != nil {
			s.logger.Warn("github pr creation failed", zap.Error(err))
			result.Message = "Changes pushed; create a PR manually in your Git provider."
		} else {
			result.PRURL = prURL
		}
	default:
		result.Message = "Changes pushed; create a PR manually in your Git provider."
	}

	_ = os.RemoveAll(cloneDir)
	return result, nil
}

func authCloneURL(rawURL, token string) string {
	if token == "" || !strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}
	return strings.Replace(rawURL, "https://", fmt.Sprintf("https://oauth2:%s@", token), 1)
}

func (s *Syncer) runGit(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (s *Syncer) gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func parseGitHubRepo(rawURL string) (owner, name string, err error) {
	u := strings.TrimSuffix(strings.TrimSpace(rawURL), ".git")
	if strings.HasPrefix(u, "git@github.com:") {
		parts := strings.Split(strings.TrimPrefix(u, "git@github.com:"), "/")
		if len(parts) >= 2 {
			return parts[0], parts[1], nil
		}
	}
	parsed, parseErr := url.Parse(u)
	if parseErr != nil {
		return "", "", parseErr
	}
	path := strings.Trim(parsed.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("cannot parse github repo from url")
	}
	return parts[0], parts[1], nil
}

func createGitHubPR(ctx context.Context, repoURL, token, base, head, title, body string) (string, error) {
	owner, repo, err := parseGitHubRepo(repoURL)
	if err != nil {
		return "", err
	}

	payload := map[string]string{
		"title": title,
		"head":  head,
		"base":  base,
		"body":  body,
	}
	raw, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo),
		bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("github api %d: %s", res.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	return parsed.HTMLURL, nil
}
