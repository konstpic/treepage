package syncclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
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

func (c *Client) TriggerSync(ctx context.Context, repoID string) (int, []byte, error) {
	url := fmt.Sprintf("%s/api/sync/repositories/%s", c.baseURL, repoID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return 0, nil, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, body, nil
}
