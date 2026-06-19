package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	cfg.BaseURL = normalizeBaseURL(cfg.BaseURL)
	if cfg.Model == "" {
		cfg.Model = "gpt-4o-mini"
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 10 * time.Minute},
	}
}

func normalizeBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		baseURL = strings.TrimSuffix(baseURL, "/chat/completions")
	}
	return strings.TrimRight(baseURL, "/")
}

func (c *Client) Available() bool {
	if !c.cfg.Enabled || c.cfg.BaseURL == "" {
		return false
	}
	if c.cfg.APIKey != "" {
		return true
	}
	return isLocalProvider(c.cfg.BaseURL)
}

func isLocalProvider(baseURL string) bool {
	u := strings.ToLower(baseURL)
	if strings.Contains(u, "localhost") ||
		strings.Contains(u, "127.0.0.1") ||
		strings.Contains(u, "host.docker.internal") {
		return true
	}
	// Private RFC1918 ranges and common local LLM ports (Ollama, LM Studio, etc.).
	if strings.Contains(u, ":11434") || strings.Contains(u, ":11343") {
		return true
	}
	for _, prefix := range []string{"192.168.", "10.", "172.16.", "172.17.", "172.18.", "172.19.", "172.2", "172.30.", "172.31."} {
		if strings.Contains(u, prefix) {
			return true
		}
	}
	return false
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) ChatJSON(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if !c.Available() {
		return "", errors.New("LLM is not configured (set LLM_ENABLED=true and LLM_API_URL; LLM_API_KEY for cloud providers)")
	}

	reqBody := chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.2,
	}
	// json_object works on OpenAI; many Ollama models ignore it — prompt enforces JSON.
	if !isLocalProvider(c.cfg.BaseURL) {
		reqBody.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	// Ollama: native JSON mode via extra body field when supported.
	if isLocalProvider(c.cfg.BaseURL) {
		return c.chatOllamaJSON(ctx, systemPrompt, userPrompt)
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		var errBody chatResponse
		_ = json.Unmarshal(raw, &errBody)
		if errBody.Error != nil && errBody.Error.Message != "" {
			return "", fmt.Errorf("LLM API error: %s", errBody.Error.Message)
		}
		return "", fmt.Errorf("LLM API HTTP %d", res.StatusCode)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM returned empty response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// Chat returns plain-text/markdown completion (no JSON mode).
func (c *Client) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if !c.Available() {
		return "", errors.New("LLM is not configured")
	}
	if isLocalProvider(c.cfg.BaseURL) {
		return c.chatOllamaPlain(ctx, systemPrompt, userPrompt)
	}
	reqBody := chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("LLM API HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM returned empty response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func (c *Client) chatOllamaPlain(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	payload := map[string]any{
		"model": c.cfg.Model,
		"messages": []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		"temperature": 0.3,
		"stream":      false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("LLM API HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM returned empty response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// chatOllamaJSON uses a plain map so we can pass Ollama's "format":"json".
func (c *Client) chatOllamaJSON(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	payload := map[string]any{
		"model": c.cfg.Model,
		"messages": []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		"temperature": 0.2,
		"format":      "json",
		"stream":      false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode >= 400 {
		return "", fmt.Errorf("LLM API HTTP %d: %s", res.StatusCode, string(raw))
	}
	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM returned empty response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func LoadConfigFromEnv(enabled bool, baseURL, apiKey, model string) Config {
	return Config{
		Enabled: enabled,
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	}
}
