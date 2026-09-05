package connectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReliefWebConnector_FetchAndNormalize(t *testing.T) {
	mockResponse := `{
		"data": [
			{
				"id": "12345",
				"fields": {
					"title": "Humanitarian crisis update",
					"date": {
						"created": "2023-10-18T10:15:00+00:00"
					},
					"url": "https://reliefweb.int/report/12345"
				}
			},
			{
				"id": "",
				"fields": {
					"title": "Invalid article without ID"
				}
			}
		]
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	connector := NewReliefWebConnector()
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
	if rec.ExternalID != "12345" {
		t.Errorf("expected external ID to be '12345', got %v", rec.ExternalID)
	}
	if rec.Title != "Humanitarian crisis update" {
		t.Errorf("expected title mismatch, got %v", rec.Title)
	}
	expectedTime, _ := time.Parse(time.RFC3339, "2023-10-18T10:15:00+00:00")
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
	if event.EventType != "humanitarian_report" {
		t.Errorf("expected event type humanitarian_report, got %v", event.EventType)
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
	if !event.OccurredAt.IsZero() {
		t.Errorf("expected OccurredAt to be zero for humanitarian_report")
	}
}
