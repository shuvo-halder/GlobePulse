package connectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGDELTConnector_FetchAndNormalize(t *testing.T) {
	mockResponse := `{
		"articles": [
			{
				"url": "https://example.com/news/123",
				"url_mobile": "https://m.example.com/news/123",
				"title": "Major crisis reported in test region",
				"seendate": "20231018T101500Z",
				"domain": "example.com",
				"language": "English",
				"sourcecountry": "Test Country"
			},
			{
				"url": "",
				"title": "Invalid article without URL"
			}
		]
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	connector := NewGDELTConnector()
	connector.apiURL = ts.URL // override URL for testing

	ctx := context.Background()
	records, err := connector.Fetch(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should only parse valid records
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec.ExternalID != "https://example.com/news/123" {
		t.Errorf("expected external ID to be url, got %v", rec.ExternalID)
	}
	if rec.Title != "Major crisis reported in test region" {
		t.Errorf("expected title mismatch, got %v", rec.Title)
	}
	expectedTime, _ := time.Parse("20060102T150405Z", "20231018T101500Z")
	if !rec.PublishedAt.Equal(expectedTime) {
		t.Errorf("expected time %v, got %v", expectedTime, rec.PublishedAt)
	}

	event, err := connector.Normalize(rec)
	if err != nil {
		t.Fatalf("expected no error on normalize, got %v", err)
	}

	if event.Title != rec.Title {
		t.Errorf("expected event title %v, got %v", rec.Title, event.Title)
	}
	if event.EventType != "news_signal" {
		t.Errorf("expected event type news_signal, got %v", event.EventType)
	}
	if event.Country != "Test Country" {
		t.Errorf("expected country 'Test Country', got %v", event.Country)
	}
	if !event.HasNoLocation {
		t.Errorf("expected HasNoLocation to be true")
	}
	if !event.EventTimeUnknown {
		t.Errorf("expected EventTimeUnknown to be true")
	}
	if !event.DetectedAt.Equal(expectedTime) {
		t.Errorf("expected DetectedAt %v, got %v", expectedTime, event.DetectedAt)
	}
}
