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

	"github.com/konstpic/treepage/backend/pkg/contenthash"
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
	if err := s.runGit(ctx, s.workDir, "clone", "--branch", baseBranch, cloneURL, cloneDir); err != nil {
		return nil, fmt.Errorf("git clone: %w", err)
	}

	authorName := "TreePage"
	authorEmail := "treepage@local"
	_ = s.runGit(ctx, cloneDir, "config", "user.name", authorName)
	_ = s.runGit(ctx, cloneDir, "config", "user.email", authorEmail)

	if err := s.runGit(ctx, cloneDir, "fetch", "origin"); err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}

	branchExisted := s.remoteBranchExists(ctx, cloneDir, input.Branch)
	if branchExisted {
		if err := s.runGit(ctx, cloneDir, "checkout", "-B", input.Branch, "origin/"+input.Branch); err != nil {
			return nil, fmt.Errorf("git checkout existing branch: %w", err)
		}
	} else if err := s.runGit(ctx, cloneDir, "checkout", "-b", input.Branch); err != nil {
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

	hasStagedChanges := s.gitHasStagedChanges(ctx, cloneDir)
	if hasStagedChanges {
		if err := s.runGit(ctx, cloneDir, "commit", "-m", input.CommitMessage); err != nil {
			return nil, fmt.Errorf("git commit: %w", err)
		}
	} else if !branchExisted {
		return nil, fmt.Errorf("no changes to publish: document content matches %s", baseBranch)
	} else {
		s.logger.Info("publish skipped commit; branch already up to date", zap.String("branch", input.Branch))
	}

	sha, err := s.gitOutput(ctx, cloneDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}

	if hasStagedChanges {
		if err := s.runGit(ctx, cloneDir, "push", "-u", "origin", input.Branch); err != nil {
			return nil, fmt.Errorf("git push: %w", err)
		}
	} else if branchExisted {
		// Branch unchanged — remote already has this content.
		_ = s.runGit(ctx, cloneDir, "push", "-u", "origin", input.Branch)
	}

	result := &PublishResult{
		Branch:    input.Branch,
		CommitSHA: strings.TrimSpace(sha),
	}
	if !hasStagedChanges && branchExisted {
		result.Message = "Branch already contains these changes; opening or reusing pull request."
	}

	if input.DocumentID != "" {
		hash := contenthash.SHA256(input.Content)
		now := time.Now()
		s.db.WithContext(ctx).Model(&models.Document{}).
			Where("id = ?", input.DocumentID).
			Updates(map[string]interface{}{
				"commit_sha":          result.CommitSHA,
				"synced_content_hash": hash,
				"has_pending_changes": false,
				"last_synced_at":      now,
			})
	}

	prTitle := strings.TrimSpace(input.PRTitle)
	if prTitle == "" {
		prTitle = input.CommitMessage
	}

	switch strings.ToLower(repo.Provider) {
	case "github", "":
		if prURL, err := createGitHubPR(ctx, repo.URL, token, baseBranch, input.Branch, prTitle, input.PRBody); err != nil {
			if existing, findErr := findGitHubPR(ctx, repo.URL, token, baseBranch, input.Branch); findErr == nil && existing != "" {
				result.PRURL = existing
			} else {
				s.logger.Warn("github pr creation failed", zap.Error(err))
				result.Message = "Changes pushed; create a PR manually in your Git provider."
			}
		} else {
			result.PRURL = prURL
		}
	case "gitlab":
		if prURL, err := createGitLabMR(ctx, repo.URL, token, baseBranch, input.Branch, prTitle, input.PRBody); err != nil {
			if existing, findErr := findGitLabMR(ctx, repo.URL, token, input.Branch); findErr == nil && existing != "" {
				result.PRURL = existing
			} else {
				s.logger.Warn("gitlab mr creation failed", zap.Error(err))
				result.Message = "Changes pushed; create a merge request manually in GitLab."
			}
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

func parseGitLabProject(rawURL string) (string, error) {
	u := strings.TrimSuffix(strings.TrimSpace(rawURL), ".git")
	var path string
	if strings.HasPrefix(u, "git@") {
		if idx := strings.LastIndex(u, ":"); idx >= 0 {
			path = u[idx+1:]
		}
	} else {
		parsed, err := url.Parse(u)
		if err != nil {
			return "", err
		}
		path = strings.Trim(parsed.Path, "/")
	}
	if path == "" {
		return "", fmt.Errorf("cannot parse gitlab project from url")
	}
	return url.PathEscape(path), nil
}

func gitLabAPIBase(rawURL string) string {
	u := strings.TrimSuffix(strings.TrimSpace(rawURL), ".git")
	if strings.HasPrefix(u, "git@") {
		parts := strings.SplitN(strings.TrimPrefix(u, "git@"), ":", 2)
		if len(parts) == 2 {
			return "https://" + parts[0] + "/api/v4"
		}
	}
	parsed, err := url.Parse(u)
	if err != nil || parsed.Host == "" {
		return "https://gitlab.com/api/v4"
	}
	return parsed.Scheme + "://" + parsed.Host + "/api/v4"
}

func createGitLabMR(ctx context.Context, repoURL, token, base, head, title, body string) (string, error) {
	project, err := parseGitLabProject(repoURL)
	if err != nil {
		return "", err
	}
	payload := map[string]string{
		"title":         title,
		"source_branch": head,
		"target_branch": base,
		"description":   body,
	}
	raw, _ := json.Marshal(payload)
	apiBase := gitLabAPIBase(repoURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/projects/%s/merge_requests", apiBase, project),
		bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	respBody, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("gitlab api %d: %s", res.StatusCode, strings.TrimSpace(string(respBody)))
	}
	var parsed struct {
		WebURL string `json:"web_url"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	return parsed.WebURL, nil
}

func (s *Syncer) remoteBranchExists(ctx context.Context, dir, branch string) bool {
	out, err := s.gitOutput(ctx, dir, "ls-remote", "--heads", "origin", branch)
	return err == nil && strings.TrimSpace(out) != ""
}

func (s *Syncer) gitHasStagedChanges(ctx context.Context, dir string) bool {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = dir
	return cmd.Run() != nil
}

func findGitHubPR(ctx context.Context, repoURL, token, base, head string) (string, error) {
	owner, repo, err := parseGitHubRepo(repoURL)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("state", "open")
	q.Set("head", fmt.Sprintf("%s:%s", owner, head))
	q.Set("base", base)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?%s", owner, repo, q.Encode()), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("github api %d", res.StatusCode)
	}
	var pulls []struct {
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(body, &pulls); err != nil {
		return "", err
	}
	if len(pulls) == 0 {
		return "", fmt.Errorf("no open pull request found")
	}
	return pulls[0].HTMLURL, nil
}

func findGitLabMR(ctx context.Context, repoURL, token, sourceBranch string) (string, error) {
	project, err := parseGitLabProject(repoURL)
	if err != nil {
		return "", err
	}
	apiBase := gitLabAPIBase(repoURL)
	q := url.Values{}
	q.Set("source_branch", sourceBranch)
	q.Set("state", "opened")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/projects/%s/merge_requests?%s", apiBase, project, q.Encode()), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("gitlab api %d", res.StatusCode)
	}
	var mrs []struct {
		WebURL string `json:"web_url"`
	}
	if err := json.Unmarshal(body, &mrs); err != nil {
		return "", err
	}
	if len(mrs) == 0 {
		return "", fmt.Errorf("no open merge request found")
	}
	return mrs[0].WebURL, nil
}
