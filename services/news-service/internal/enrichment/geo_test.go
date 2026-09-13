package enrichment

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/global-news/news-service/internal/domain"
)

// TEST 1 — Unknown coordinates
// Input: no coordinates
// Expected: latitude = nil, longitude = nil.
// Explicitly assert latitude != 0, longitude != 0 (never use 0,0 as sentinel).
func TestGeoEnricher_UnknownCoordinates_NilNotZero(t *testing.T) {
	enricher := NewGeoEnricher()

	event := &domain.ThreatEvent{
		Latitude:      nil,
		Longitude:     nil,
		HasNoLocation: true,
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if event.Latitude != nil {
		t.Fatalf("Expected latitude to be nil/NULL, got %v (must not use 0 or non-nil as sentinel)", *event.Latitude)
	}
	if event.Longitude != nil {
		t.Fatalf("Expected longitude to be nil/NULL, got %v (must not use 0 or non-nil as sentinel)", *event.Longitude)
	}
	if !event.HasNoLocation {
		t.Errorf("Expected HasNoLocation to remain true")
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["location_confidence"] != "unknown" {
		t.Errorf("Expected 'unknown' location confidence, got %v", geo["location_confidence"])
	}
}

// TEST 2 — Valid coordinate at 0,0
// Input: legitimate coordinate latitude = 0, longitude = 0 (Null Island)
// Expected: location is VALID, confidence = exact. Proves 0,0 is not treated as missing.
func TestGeoEnricher_ValidCoordinateAt00(t *testing.T) {
	enricher := NewGeoEnricher()

	lat := 0.0
	lon := 0.0
	event := &domain.ThreatEvent{
		Latitude:      &lat,
		Longitude:     &lon,
		HasNoLocation: false,
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if event.Latitude == nil || *event.Latitude != 0.0 {
		t.Fatalf("Expected latitude 0.0 to be preserved, got %v", event.Latitude)
	}
	if event.Longitude == nil || *event.Longitude != 0.0 {
		t.Fatalf("Expected longitude 0.0 to be preserved, got %v", event.Longitude)
	}
	if event.HasNoLocation {
		t.Errorf("Expected HasNoLocation to remain false for valid 0,0 coordinates")
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["location_confidence"] != "exact" {
		t.Errorf("Expected 'exact' location confidence for valid 0,0 coordinates, got %v", geo["location_confidence"])
	}
}

// TEST 3 — USGS authoritative coordinates
// Input: valid USGS coordinates
// Expected: coordinates preserved exactly, confidence = exact, GeoEnricher leaves coordinates unchanged.
func TestGeoEnricher_USGSAuthoritativeCoordinatesPreserved(t *testing.T) {
	enricher := NewGeoEnricher()

	lat := 35.456
	lon := -120.123
	event := &domain.ThreatEvent{
		Latitude:        &lat,
		Longitude:       &lon,
		HasNoLocation:   false,
		LocationDetails: "10km NE of Testville, CA",
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if event.Latitude == nil || *event.Latitude != 35.456 {
		t.Errorf("USGS latitude was modified: got %v, expected 35.456", event.Latitude)
	}
	if event.Longitude == nil || *event.Longitude != -120.123 {
		t.Errorf("USGS longitude was modified: got %v, expected -120.123", event.Longitude)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["location_confidence"] != "exact" {
		t.Errorf("Expected 'exact' location confidence for authoritative USGS coordinates, got %v", geo["location_confidence"])
	}
}

// TEST 4 — Country-only location
// Input: country = Japan, latitude = nil, longitude = nil
// Expected: location_confidence = country, NOT exact
func TestGeoEnricher_CountryOnlyLocation(t *testing.T) {
	enricher := NewGeoEnricher()

	event := &domain.ThreatEvent{
		Country:       "Japan",
		Latitude:      nil,
		Longitude:     nil,
		HasNoLocation: true,
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if event.Latitude != nil || event.Longitude != nil {
		t.Errorf("Country-only event should not have fabricated coordinates, got lat=%v lon=%v", event.Latitude, event.Longitude)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["location_confidence"] != "country" {
		t.Errorf("Expected 'country' location confidence, got %v", geo["location_confidence"])
	}
	if geo["location_confidence"] == "exact" {
		t.Errorf("Country-only event MUST NOT have 'exact' location confidence")
	}
}

// TEST 6 — Unknown country
// Input: no reliable country information
// Expected: country remains unknown/empty, no fabricated value
func TestGeoEnricher_UnknownCountry(t *testing.T) {
	enricher := NewGeoEnricher()

	event := &domain.ThreatEvent{
		Country:         "",
		LocationDetails: "Unspecified international waters",
		Latitude:        nil,
		Longitude:       nil,
		HasNoLocation:   true,
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, event)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if event.Country != "" {
		t.Errorf("Expected country to remain empty/unknown, got %v", event.Country)
	}

	var meta map[string]interface{}
	json.Unmarshal(event.Metadata, &meta)
	geo := meta["geo"].(map[string]interface{})

	if geo["location_confidence"] != "unknown" {
		t.Errorf("Expected 'unknown' location confidence, got %v", geo["location_confidence"])
	}
}

// TEST 7 — Text heuristic safety
// Verify:
// - does not overwrite authoritative coordinates
// - does not mark location exact without coordinates
// - produces at most a lower-confidence inference (country)
func TestGeoEnricher_TextHeuristicSafety(t *testing.T) {
	enricher := NewGeoEnricher()

	// 1. With authoritative coordinates and USGS place ending in "Japan"
	lat := 40.5
	lon := 142.1
	eventAuthoritative := &domain.ThreatEvent{
		Latitude:        &lat,
		Longitude:       &lon,
		HasNoLocation:   false,
		LocationDetails: "45 km E of Hachinohe, Japan",
	}

	err := enricher.Enrich(context.Background(), domain.ExternalRecord{}, eventAuthoritative)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if eventAuthoritative.Latitude == nil || *eventAuthoritative.Latitude != 40.5 {
		t.Errorf("Heuristic overwrote authoritative latitude")
	}
	if eventAuthoritative.Longitude == nil || *eventAuthoritative.Longitude != 142.1 {
		t.Errorf("Heuristic overwrote authoritative longitude")
	}
	if eventAuthoritative.Country != "Japan" {
		t.Errorf("Expected country to be resolved to 'Japan', got %v", eventAuthoritative.Country)
	}

	var metaAuth map[string]interface{}
	json.Unmarshal(eventAuthoritative.Metadata, &metaAuth)
	geoAuth := metaAuth["geo"].(map[string]interface{})
	if geoAuth["country_source"] != "heuristic" {
		t.Errorf("Expected country_source 'heuristic', got %v", geoAuth["country_source"])
	}
	if geoAuth["location_confidence"] != "exact" {
		t.Errorf("Expected 'exact' location confidence when authoritative coordinates exist, got %v", geoAuth["location_confidence"])
	}

	// 2. Text heuristic without coordinates MUST NOT mark confidence exact
	eventNoCoords := &domain.ThreatEvent{
		Latitude:        nil,
		Longitude:       nil,
		HasNoLocation:   true,
		LocationDetails: "Protests reported near Manila, Philippines",
	}

	err = enricher.Enrich(context.Background(), domain.ExternalRecord{}, eventNoCoords)
	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if eventNoCoords.Latitude != nil || eventNoCoords.Longitude != nil {
		t.Errorf("Heuristic must not fabricate coordinates")
	}
	if eventNoCoords.Country != "Philippines" {
		t.Errorf("Expected country 'Philippines', got %v", eventNoCoords.Country)
	}

	var metaNoCoords map[string]interface{}
	json.Unmarshal(eventNoCoords.Metadata, &metaNoCoords)
	geoNoCoords := metaNoCoords["geo"].(map[string]interface{})

	if geoNoCoords["location_confidence"] != "country" {
		t.Errorf("Expected 'country' confidence for text-only heuristic, got %v", geoNoCoords["location_confidence"])
	}
	if geoNoCoords["location_confidence"] == "exact" {
		t.Errorf("Text heuristic without coordinates must NEVER produce 'exact' confidence")
	}
}
