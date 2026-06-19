package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Config struct {
	Enabled bool
	BaseURL string
	APIKey  string
	Model   string
}

type Client struct {
	cfg  Config
	http *http.Client
}

func LoadConfigFromEnv() Config {
	enabled := os.Getenv("EMBEDDING_ENABLED") == "true"
	if os.Getenv("EMBEDDING_ENABLED") == "" && os.Getenv("LLM_ENABLED") == "true" {
		enabled = true
	}
	baseURL := os.Getenv("EMBEDDING_API_URL")
	if baseURL == "" {
		baseURL = os.Getenv("LLM_API_URL")
	}
	model := os.Getenv("EMBEDDING_MODEL")
	if model == "" {
		model = "nomic-embed-text"
	}
	return Config{
		Enabled: enabled,
		BaseURL: normalizeBaseURL(baseURL),
		APIKey:  firstNonEmpty(os.Getenv("EMBEDDING_API_KEY"), os.Getenv("LLM_API_KEY")),
		Model:   model,
	}
}

func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://127.0.0.1:11434"
	}
	cfg.BaseURL = strings.TrimRight(normalizeBaseURL(cfg.BaseURL), "/")
	return &Client{cfg: cfg, http: &http.Client{Timeout: 2 * time.Minute}}
}

func (c *Client) Available() bool {
	return c.cfg.Enabled && c.cfg.BaseURL != ""
}

func (c *Client) Embed(ctx context.Context, text string) (Vector, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("empty text")
	}
	if !c.Available() {
		return nil, errors.New("embeddings not configured")
	}
	if isOllamaNative(c.cfg.BaseURL) {
		return c.embedOllama(ctx, text)
	}
	return c.embedOpenAI(ctx, text)
}

func (c *Client) embedOllama(ctx context.Context, text string) (Vector, error) {
	base := ollamaRoot(c.cfg.BaseURL)
	body, _ := json.Marshal(map[string]string{"model": c.cfg.Model, "prompt": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("embeddings HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	return float64sToVector(parsed.Embedding), nil
}

func (c *Client) embedOpenAI(ctx context.Context, text string) (Vector, error) {
	body, _ := json.Marshal(map[string]any{
		"model": c.cfg.Model,
		"input": text,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("embeddings HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) == 0 {
		return nil, errors.New("empty embedding response")
	}
	return float64sToVector(parsed.Data[0].Embedding), nil
}

func normalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		baseURL = strings.TrimSuffix(baseURL, "/chat/completions")
	}
	if strings.HasSuffix(baseURL, "/embeddings") {
		baseURL = strings.TrimSuffix(baseURL, "/embeddings")
	}
	return strings.TrimRight(baseURL, "/")
}

func isOllamaNative(baseURL string) bool {
	u := strings.ToLower(baseURL)
	return strings.Contains(u, ":11434") || strings.Contains(u, ":11343") ||
		strings.Contains(u, "localhost") || strings.Contains(u, "127.0.0.1") || strings.Contains(u, "192.168.")
}

func ollamaRoot(baseURL string) string {
	baseURL = normalizeBaseURL(baseURL)
	if strings.HasSuffix(baseURL, "/v1") {
		return strings.TrimSuffix(baseURL, "/v1")
	}
	return baseURL
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func float64sToVector(v []float64) Vector {
	out := make(Vector, len(v))
	for i, x := range v {
		out[i] = float32(x)
	}
	return out
}
