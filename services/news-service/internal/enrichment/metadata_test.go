package enrichment

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/global-news/news-service/internal/domain"
)

func TestMetadataEnricher_QualityFlags(t *testing.T) {
	enricher := NewMetadataEnricher()
	
	record := domain.ExternalRecord{
		// Missing PublishedAt
	}

	event := &domain.ThreatEvent{
		HasNoLocation:    true,
		EventTimeUnknown: true,
		Title:            "", // Missing title
	}

	err := enricher.Enrich(context.Background(), record, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	flagsInterface := meta["data_quality_flags"].([]interface{})
	
	flagsMap := make(map[string]bool)
	for _, f := range flagsInterface {
		flagsMap[f.(string)] = true
	}

	expectedFlags := []string{
		"missing_location",
		"missing_event_time",
		"source_timestamp_missing",
		"missing_title",
	}

	for _, ef := range expectedFlags {
		if !flagsMap[ef] {
			t.Errorf("Expected flag %v to be set", ef)
		}
	}
}

func TestMetadataEnricher_Idempotent(t *testing.T) {
	enricher := NewMetadataEnricher()
	
	record := domain.ExternalRecord{
		PublishedAt: time.Now(),
	}

	event := &domain.ThreatEvent{
		HasNoLocation: true,
		Title:         "Valid Title",
	}

	err := enricher.Enrich(context.Background(), record, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	// Run again to ensure no duplicate flags in array
	err = enricher.Enrich(context.Background(), record, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	flagsInterface := meta["data_quality_flags"].([]interface{})
	
	if len(flagsInterface) != 1 {
		t.Errorf("Expected exactly 1 flag (missing_location) despite multiple runs, got %d", len(flagsInterface))
	}
	
	if flagsInterface[0].(string) != "missing_location" {
		t.Errorf("Expected missing_location flag, got %v", flagsInterface[0])
	}
}
