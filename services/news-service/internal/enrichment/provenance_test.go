package enrichment

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/global-news/news-service/internal/domain"
)

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
