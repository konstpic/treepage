package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

// Webhook sends JSON POST notifications to NOTIFY_WEBHOOK_URL when configured.
type Webhook struct {
	url    string
	client *http.Client
}

func NewWebhookFromEnv() *Webhook {
	url := strings.TrimSpace(os.Getenv("NOTIFY_WEBHOOK_URL"))
	if url == "" {
		return nil
	}
	return &Webhook{
		url: url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type Payload struct {
	Type         string  `json:"type"`
	Title        string  `json:"title"`
	Body         string  `json:"body"`
	UserID       string  `json:"user_id"`
	ResourceType *string `json:"resource_type,omitempty"`
	ResourceID   *string `json:"resource_id,omitempty"`
}

func (w *Webhook) Notify(ctx context.Context, p Payload) {
	if w == nil || w.url == "" {
		return
	}
	body, err := json.Marshal(p)
	if err != nil {
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if secret := strings.TrimSpace(os.Getenv("NOTIFY_WEBHOOK_SECRET")); secret != "" {
		req.Header.Set("X-Webhook-Secret", secret)
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
