package syncclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/konstpic/treepage/backend/pkg/internalauth"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Minute},
	}
}

func (c *Client) applyInternalAuth(req *http.Request) {
	if name, token := internalauth.ClientHeader(); token != "" {
		req.Header.Set(name, token)
	}
}

func (c *Client) TriggerSync(ctx context.Context, repoID string) (int, []byte, error) {
	url := fmt.Sprintf("%s/api/sync/repositories/%s", c.baseURL, repoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return 0, nil, err
	}
	c.applyInternalAuth(req)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, body, nil
}

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
	Branch    string `json:"branch"`
	CommitSHA string `json:"commit_sha,omitempty"`
	PRURL     string `json:"pr_url,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (c *Client) PublishDocument(ctx context.Context, repoID string, input PublishInput) (int, *PublishResult, []byte, error) {
	url := fmt.Sprintf("%s/api/sync/repositories/%s/publish", c.baseURL, repoID)
	raw, err := json.Marshal(input)
	if err != nil {
		return 0, nil, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyInternalAuth(req)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return res.StatusCode, nil, body, nil
	}
	var result PublishResult
	if err := json.Unmarshal(body, &result); err != nil {
		return res.StatusCode, nil, body, err
	}
	return res.StatusCode, &result, body, nil
}

type GitFileRevision struct {
	CommitSHA  string `json:"commit_sha"`
	AuthorName string `json:"author_name"`
	Message    string `json:"message"`
	CreatedAt  string `json:"created_at"`
}

func (c *Client) FileHistory(ctx context.Context, repoID, path string, limit int) ([]GitFileRevision, error) {
	url := fmt.Sprintf("%s/api/sync/repositories/%s/file-history?path=%s&limit=%d",
		c.baseURL, repoID, urlQueryEscape(path), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.applyInternalAuth(req)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("sync file-history %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var parsed struct {
		Items []GitFileRevision `json:"items"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return parsed.Items, nil
}

func (c *Client) FileContentAt(ctx context.Context, repoID, path, sha string) (string, error) {
	url := fmt.Sprintf("%s/api/sync/repositories/%s/file-content?path=%s&sha=%s",
		c.baseURL, repoID, urlQueryEscape(path), urlQueryEscape(sha))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	c.applyInternalAuth(req)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return "", fmt.Errorf("sync file-content %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var parsed struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}
	return parsed.Content, nil
}

func urlQueryEscape(s string) string {
	return url.QueryEscape(s)
}
