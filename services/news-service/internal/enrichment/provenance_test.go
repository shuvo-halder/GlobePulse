package enrichment

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/global-news/news-service/internal/domain"
)

// TEST 8 & Provenance Idempotency
func TestProvenanceEnricher_Idempotent(t *testing.T) {
	enricher := NewProvenanceEnricher()

	record := domain.ExternalRecord{
		ExternalID:  "ext-123",
		URL:         "http://example.com/ext-123",
		PublishedAt: time.Now(),
	}

	event := &domain.ThreatEvent{}

	err := enricher.Enrich(context.Background(), record, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	var meta1 map[string]interface{}
	json.Unmarshal(event.Metadata, &meta1)
	prov1 := meta1["provenance"].(map[string]interface{})

	if prov1["external_id"] != "ext-123" {
		t.Errorf("Expected external_id ext-123, got %v", prov1["external_id"])
	}

	originalNormalizedAt := prov1["normalized_at"]

	// Run again (simulate idempotent rerun)
	time.Sleep(10 * time.Millisecond)
	err = enricher.Enrich(context.Background(), record, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	var meta2 map[string]interface{}
	json.Unmarshal(event.Metadata, &meta2)
	prov2 := meta2["provenance"].(map[string]interface{})

	if prov2["normalized_at"] != originalNormalizedAt {
		t.Errorf("Expected normalized_at to remain constant across idempotent runs, got %v vs %v", prov2["normalized_at"], originalNormalizedAt)
	}
}

// TEST 5 — GDELT source country
// Input: sourcecountry = US in raw metadata
// Expected: source country preserved in provenance metadata.
// But: event country MUST NOT automatically become US.
func TestProvenanceEnricher_GDELTSourceCountryPreserved(t *testing.T) {
	enricher := NewProvenanceEnricher()

	rawGDELT := []byte(`{"url":"https://example.com/story","sourcecountry":"US"}`)
	pubTime := time.Date(2023, 10, 18, 12, 0, 0, 0, time.UTC)
	record := domain.ExternalRecord{
		ExternalID:  "https://example.com/story",
		URL:         "https://example.com/story",
		PublishedAt: pubTime,
		RawMetadata: rawGDELT,
	}

	event := &domain.ThreatEvent{
		Country:       "", // Not fabricated
		HasNoLocation: true,
	}

	err := enricher.Enrich(context.Background(), record, event)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	// Event country must NOT become US
	if event.Country != "" {
		t.Errorf("GDELT sourcecountry must NOT become event country, got %v", event.Country)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	prov := meta["provenance"].(map[string]interface{})

	if prov["source_country"] != "US" {
		t.Errorf("Expected provenance to preserve source_country 'US', got %v", prov["source_country"])
	}
	if prov["external_id"] != "https://example.com/story" {
		t.Errorf("Expected external_id preserved, got %v", prov["external_id"])
	}
}

// TEST 10 — Provenance preservation
// Verify that enrichment does not modify:
// - external_id
// - source URL
// - raw source metadata
// - source publication timestamp
func TestProvenanceEnricher_PreservesOriginalEvidence(t *testing.T) {
	enricher := NewProvenanceEnricher()

	originalRaw := []byte(`{"original":"payload","immutable":true}`)
	originalPubTime := time.Date(2024, 1, 15, 8, 30, 0, 0, time.UTC)
	record := domain.ExternalRecord{
		ExternalID:  "evidence-999",
		URL:         "https://intel.source.org/report/999",
		PublishedAt: originalPubTime,
		RawMetadata: originalRaw,
	}

	event := &domain.ThreatEvent{
		Title: "Threat Observed",
	}

	err := enricher.Enrich(context.Background(), record, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	// Verify record was untouched
	if record.ExternalID != "evidence-999" {
		t.Errorf("Record external_id was mutated")
	}
	if record.URL != "https://intel.source.org/report/999" {
		t.Errorf("Record URL was mutated")
	}
	if !record.PublishedAt.Equal(originalPubTime) {
		t.Errorf("Record PublishedAt was mutated")
	}
	if !bytes.Equal(record.RawMetadata, originalRaw) {
		t.Errorf("Record RawMetadata was mutated")
	}
}
