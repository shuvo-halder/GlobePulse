package ingestion

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type HTTPClient struct {
	client     *http.Client
	retries    int
	backoffBase time.Duration
	backoffMax  time.Duration
}

func NewHTTPClient(timeout time.Duration, retries int) *HTTPClient {
	return &HTTPClient{
		client:      &http.Client{Timeout: timeout},
		retries:     retries,
		backoffBase: 1 * time.Second,
		backoffMax:  30 * time.Second,
	}
}

func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	var delay time.Duration

	for i := 0; i <= c.retries; i++ {
		if i > 0 {
			// Apply delay before retry, respecting context
			slog.Warn("Retrying HTTP request", 
				"url", url, 
				"attempt", i, 
				"delay", delay.String(), 
				"error", lastErr)
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", "GlobePulse-Ingestion/1.0")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network error: %w", err)
			delay = c.calculateBackoff(i)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			delay = c.calculateBackoff(i)
			continue
		}

		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			
			if resp.StatusCode == http.StatusTooManyRequests {
				delay = c.parseRetryAfter(resp.Header.Get("Retry-After"))
				if delay == 0 {
					delay = c.calculateBackoff(i)
				}
				continue
			}

			// Don't blindly retry 400s unless it's a 429, 408, or 5xx
			if resp.StatusCode < 500 && resp.StatusCode != http.StatusRequestTimeout {
				// Fatal 4xx error (e.g. 401, 403, 404)
				return nil, lastErr
			}
			
			delay = c.calculateBackoff(i)
			continue
		}

		return body, nil
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *HTTPClient) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: base * 2^(attempt-1)
	backoff := float64(c.backoffBase) * float64(uint(1)<<(attempt-1))
	if backoff > float64(c.backoffMax) {
		backoff = float64(c.backoffMax)
	}

	// Add jitter (up to 20% of the backoff)
	jitter := rand.Float64() * 0.2 * backoff
	
	// Final delay = backoff - jitter/2 (to vary slightly up or down)
	finalDelay := time.Duration(backoff - jitter/2)
	
	if finalDelay < c.backoffBase {
		finalDelay = c.backoffBase
	}
	
	return finalDelay
}

func (c *HTTPClient) parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	// Try parsing as integer seconds
	if secs, err := strconv.Atoi(header); err == nil {
		return time.Duration(secs) * time.Second
	}
	// Try parsing as HTTP date
	if t, err := time.Parse(http.TimeFormat, header); err == nil {
		delay := time.Until(t)
		if delay > 0 {
			return delay
		}
	}
	return 0
}
