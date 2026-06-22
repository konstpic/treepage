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

type Slack struct {
	url    string
	client *http.Client
}

func NewSlackFromEnv() *Slack {
	url := strings.TrimSpace(os.Getenv("NOTIFY_SLACK_WEBHOOK_URL"))
	if url == "" {
		return nil
	}
	return &Slack{url: url, client: &http.Client{Timeout: 5 * time.Second}}
}

func (s *Slack) Notify(ctx context.Context, p Payload) {
	if s == nil || s.url == "" {
		return
	}
	text := "*" + p.Title + "*\n" + p.Body
	body, _ := json.Marshal(map[string]string{"text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
