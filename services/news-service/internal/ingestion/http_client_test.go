package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestHTTPClient_TemporaryFailureThenSuccess(t *testing.T) {
	var attempt int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := atomic.AddInt32(&attempt, 1)
		if a < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewHTTPClient(2*time.Second, 3)
	// Lower backoff for fast tests
	client.backoffBase = 10 * time.Millisecond
	client.backoffMax = 50 * time.Millisecond

	ctx := context.Background()
	body, err := client.Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("Expected success after retries, got err: %v", err)
	}
	if string(body) != "success" {
		t.Errorf("Expected 'success', got %s", body)
	}
	if atomic.LoadInt32(&attempt) != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}
}

func TestHTTPClient_429Retry(t *testing.T) {
	var attempt int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a := atomic.AddInt32(&attempt, 1)
		if a == 1 {
			w.Header().Set("Retry-After", "1") // 1 second
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewHTTPClient(2*time.Second, 3)
	
	start := time.Now()
	body, err := client.Get(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Expected success after retry, got err: %v", err)
	}
	duration := time.Since(start)

	if string(body) != "success" {
		t.Errorf("Expected 'success', got %s", body)
	}
	if atomic.LoadInt32(&attempt) != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempt)
	}
	if duration < 1*time.Second {
		t.Errorf("Expected duration >= 1s due to Retry-After, got %v", duration)
	}
}

func TestHTTPClient_RetryExhaustion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPClient(2*time.Second, 2)
	client.backoffBase = 10 * time.Millisecond
	client.backoffMax = 50 * time.Millisecond

	_, err := client.Get(context.Background(), server.URL)
	if err == nil {
		t.Fatal("Expected error after exhaustion, got nil")
	}
}

func TestHTTPClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewHTTPClient(2*time.Second, 5)
	client.backoffBase = 500 * time.Millisecond // High enough to catch cancellation
	client.backoffMax = 1 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	
	errCh := make(chan error, 1)
	go func() {
		_, err := client.Get(ctx, server.URL)
		errCh <- err
	}()

	// Wait briefly to let the first attempt fail and enter the retry delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	err := <-errCh
	if err != context.Canceled {
		t.Fatalf("Expected context.Canceled, got: %v", err)
	}
}
