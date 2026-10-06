package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultUserAgent = "vps-prices-bot (+https://github.com/lesichkovm/vps-prices)"

// HTTPClient wraps http.Client with custom User-Agent and context support.
type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(timeout time.Duration) *HTTPClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &HTTPClient{
		client: &http.Client{Timeout: timeout},
	}
}

func (c *HTTPClient) Client() *http.Client {
	return c.client
}

func (c *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %s: %w", url, err)
	}

	req.Header.Set("User-Agent", DefaultUserAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body for %s: %w", url, err)
	}

	return body, nil
}
