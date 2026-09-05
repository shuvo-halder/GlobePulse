package connectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUSGSConnector_FetchAndNormalize(t *testing.T) {
	mockResponse := `{
		"type": "FeatureCollection",
		"features": [
			{
				"type": "Feature",
				"properties": {
					"mag": 4.5,
					"place": "10km NE of Testville, CA",
					"time": 1697624100000,
					"updated": 1697624150000,
					"url": "https://earthquake.usgs.gov/earthquakes/eventpage/us12345",
					"detail": "https://earthquake.usgs.gov/earthquakes/feed/v1.0/detail/us12345.geojson",
					"felt": null,
					"cdi": null,
					"mmi": null,
					"alert": null,
					"status": "reviewed",
					"tstmic": 0,
					"net": "us",
					"code": "12345",
					"ids": ",us12345,",
					"sources": ",us,",
					"types": ",origin,phase-data,",
					"nst": 10,
					"dmin": 0.5,
					"rms": 0.5,
					"gap": 50,
					"magType": "mwr",
					"type": "earthquake",
					"title": "M 4.5 - 10km NE of Testville, CA"
				},
				"geometry": {
					"type": "Point",
					"coordinates": [
						-120.123,
						35.456,
						10.5
					]
				},
				"id": "us12345"
			},
			{
				"type": "Feature",
				"properties": {
					"mag": 7.1,
					"place": "Offshore Test Island",
					"time": 1697624200000,
					"updated": 1697624250000,
					"url": "https://earthquake.usgs.gov/earthquakes/eventpage/us67890",
					"alert": "red",
					"status": "automatic",
					"type": "earthquake",
					"title": "M 7.1 - Offshore Test Island"
				},
				"geometry": {
					"type": "Point",
					"coordinates": [
						140.0,
						-10.0,
						30.0
					]
				},
				"id": "us67890"
			}
		]
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	connector := NewUSGSConnector()
	connector.apiURL = ts.URL // override URL for testing

	ctx := context.Background()
	records, err := connector.Fetch(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Test Event 1
	rec1 := records[0]
	if rec1.ExternalID != "us12345" {
		t.Errorf("expected external ID 'us12345', got %v", rec1.ExternalID)
	}
	if rec1.URL != "https://earthquake.usgs.gov/earthquakes/eventpage/us12345" {
		t.Errorf("expected URL mismatch, got %v", rec1.URL)
	}

	event1, err := connector.Normalize(rec1)
	if err != nil {
		t.Fatalf("expected no error on normalize, got %v", err)
	}
	if event1.Longitude != -120.123 {
		t.Errorf("expected lon -120.123, got %v", event1.Longitude)
	}
	if event1.Latitude != 35.456 {
		t.Errorf("expected lat 35.456, got %v", event1.Latitude)
	}
	
	expectedOccurredAt := time.UnixMilli(1697624100000)
	if !event1.OccurredAt.Equal(expectedOccurredAt) {
		t.Errorf("expected OccurredAt %v, got %v", expectedOccurredAt, event1.OccurredAt)
	}
	if event1.Severity != "medium" {
		t.Errorf("expected severity 'medium', got %v", event1.Severity)
	}
	if event1.EventType != "earthquake" {
		t.Errorf("expected event type 'earthquake', got %v", event1.EventType)
	}

	// Test Event 2
	rec2 := records[1]
	event2, err := connector.Normalize(rec2)
	if err != nil {
		t.Fatalf("expected no error on normalize, got %v", err)
	}
	if event2.Severity != "critical" {
		t.Errorf("expected severity 'critical', got %v", event2.Severity)
	}
}

func TestUSGSConnector_InvalidCoordinates(t *testing.T) {
	connector := NewUSGSConnector()
	
	// Test missing coordinates
	missingCoordsJSON := `{"id": "test1", "properties": {"mag": 4.5}, "geometry": {"coordinates": []}}`
	_, err := connector.Normalize(domain.ExternalRecord{RawMetadata: []byte(missingCoordsJSON)})
	if err == nil {
		t.Errorf("expected error for missing coordinates")
	}

	// Test out of bounds lat
	invalidLatJSON := `{"id": "test2", "properties": {"mag": 4.5}, "geometry": {"coordinates": [0, 95]}}`
	_, err = connector.Normalize(domain.ExternalRecord{RawMetadata: []byte(invalidLatJSON)})
	if err == nil {
		t.Errorf("expected error for invalid latitude")
	}

	// Test out of bounds lon
	invalidLonJSON := `{"id": "test3", "properties": {"mag": 4.5}, "geometry": {"coordinates": [190, 0]}}`
	_, err = connector.Normalize(domain.ExternalRecord{RawMetadata: []byte(invalidLonJSON)})
	if err == nil {
		t.Errorf("expected error for invalid longitude")
	}
}

func TestUSGSConnector_HTTPFailure(t *testing.T) {
	connector := NewUSGSConnector()
	connector.apiURL = "http://localhost:12345/nonexistent" // Invalid URL

	ctx := context.Background()
	_, err := connector.Fetch(ctx)
	if err == nil {
		t.Fatalf("expected HTTP error")
	}
}

func TestUSGSConnector_EmptyFeatureCollection(t *testing.T) {
	mockResponse := `{"type": "FeatureCollection", "features": []}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	connector := NewUSGSConnector()
	connector.apiURL = ts.URL

	ctx := context.Background()
	records, err := connector.Fetch(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}
