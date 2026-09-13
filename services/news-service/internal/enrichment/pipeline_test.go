package enrichment

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/global-news/news-service/internal/domain"
)

// TEST 8 — Full Pipeline Idempotency
// Execute full enrichment pipeline twice.
// Expected: same semantic output, no duplicate arrays/flags/metadata.
func TestPipeline_Idempotency(t *testing.T) {
	pipe := NewPipeline(
		NewProvenanceEnricher(),
		NewGeoEnricher(),
		NewMetadataEnricher(),
	)

	pubTime := time.Date(2023, 10, 18, 12, 0, 0, 0, time.UTC)
	record := domain.ExternalRecord{
		ExternalID:  "pipe-test-1",
		URL:         "https://example.com/item/1",
		PublishedAt: pubTime,
		RawMetadata: []byte(`{"sourcecountry":"US"}`),
	}

	lat := 35.0
	lon := 139.0
	event := &domain.ThreatEvent{
		Title:           "Seismic Event",
		EventType:       "earthquake",
		Latitude:        &lat,
		Longitude:       &lon,
		LocationDetails: "45 km E of Hachinohe, Japan",
		HasNoLocation:   false,
	}

	ctx := context.Background()

	// Run 1
	err := pipe.Enrich(ctx, record, event)
	if err != nil {
		t.Fatalf("First enrichment pass failed: %v", err)
	}

	metaPass1 := make([]byte, len(event.Metadata))
	copy(metaPass1, event.Metadata)
	countryPass1 := event.Country
	latPass1 := *event.Latitude
	lonPass1 := *event.Longitude

	// Run 2
	err = pipe.Enrich(ctx, record, event)
	if err != nil {
		t.Fatalf("Second enrichment pass failed: %v", err)
	}

	metaPass2 := event.Metadata

	if event.Country != countryPass1 {
		t.Errorf("Country changed across runs: %v vs %v", event.Country, countryPass1)
	}
	if *event.Latitude != latPass1 || *event.Longitude != lonPass1 {
		t.Errorf("Coordinates changed across runs")
	}
	if !bytes.Equal(metaPass1, metaPass2) {
		t.Errorf("Metadata changed across idempotent runs:\nPass 1: %s\nPass 2: %s", string(metaPass1), string(metaPass2))
	}
}
