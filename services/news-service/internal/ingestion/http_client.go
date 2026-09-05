package ingestion

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type HTTPClient struct {
	client  *http.Client
	retries int
}

func NewHTTPClient(timeout time.Duration, retries int) *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{Timeout: timeout},
		retries: retries,
	}
}

func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for i := 0; i <= c.retries; i++ {
		if i > 0 {
			slog.Info("Retrying request", "url", url, "attempt", i)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i) * time.Second):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "GlobePulse-Ingestion/1.0")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		return body, nil
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
