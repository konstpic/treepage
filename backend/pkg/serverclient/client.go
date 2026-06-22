package serverclient

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client calls backend-server internal APIs (search reindex after sync).
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewFromEnv() *Client {
	url := strings.TrimRight(os.Getenv("SERVER_SERVICE_URL"), "/")
	if url == "" {
		url = "http://backend-server:8082"
	}
	token := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if token == "" {
		return nil
	}
	return &Client{
		baseURL: url,
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) ReindexDocument(ctx context.Context, docID string) error {
	if c == nil || docID == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/internal/documents/"+docID+"/reindex", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("reindex document %s: HTTP %d", docID, resp.StatusCode)
	}
	return nil
}

func (c *Client) DeleteDocumentIndex(ctx context.Context, docID string) error {
	if c == nil || docID == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/internal/documents/"+docID+"/search-index", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete search index %s: HTTP %d", docID, resp.StatusCode)
	}
	return nil
}
