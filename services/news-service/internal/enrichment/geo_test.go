package enrichment

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/global-news/news-service/internal/domain"
)

func TestGeoEnricher_AuthoritativeCoordinates(t *testing.T) {
	enricher := NewGeoEnricher()
	
	event := &domain.ThreatEvent{
		Latitude:      40.123,
		Longitude:     141.456,
		HasNoLocation: false, // Authoritative USGS behavior
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	if event.Latitude != 40.123 || event.Longitude != 141.456 {
		t.Errorf("GeoEnricher overwrote authoritative coordinates: %v, %v", event.Latitude, event.Longitude)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["location_confidence"] != "exact" {
		t.Errorf("Expected 'exact' location confidence, got %v", geo["location_confidence"])
	}
}

func TestGeoEnricher_UnknownLocationRemainsUnknown(t *testing.T) {
	enricher := NewGeoEnricher()
	
	event := &domain.ThreatEvent{
		Latitude:      0,
		Longitude:     0,
		HasNoLocation: true, // Typical news article missing coordinates
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	if event.Latitude != 0 || event.Longitude != 0 {
		t.Errorf("GeoEnricher incorrectly populated missing coordinates")
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["location_confidence"] != "unknown" {
		t.Errorf("Expected 'unknown' location confidence, got %v", geo["location_confidence"])
	}
}

func TestGeoEnricher_CountryResolution(t *testing.T) {
	enricher := NewGeoEnricher()
	
	event := &domain.ThreatEvent{
		LocationDetails: "Some city in Japan",
		HasNoLocation:   true,
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil, got %v", err)
	}

	if event.Country != "Japan" {
		t.Errorf("Expected country 'Japan', got %v", event.Country)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["iso_country_code"] != "JP" {
		t.Errorf("Expected ISO code 'JP', got %v", geo["iso_country_code"])
	}

	if geo["location_confidence"] != "country" {
		t.Errorf("Expected 'country' location confidence, got %v", geo["location_confidence"])
	}
}
