package enrichment

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/global-news/news-service/internal/domain"
)

type GeoEnricher struct{}

func NewGeoEnricher() *GeoEnricher {
	return &GeoEnricher{}
}

func (e *GeoEnricher) Name() string {
	return "GeoEnricher"
}

func (e *GeoEnricher) Enrich(ctx context.Context, record domain.ExternalRecord, event *domain.ThreatEvent) error {
	var meta map[string]interface{}
	if len(event.Metadata) > 0 {
		if err := json.Unmarshal(event.Metadata, &meta); err != nil {
			meta = make(map[string]interface{})
		}
	} else {
		meta = make(map[string]interface{})
	}

	geoMeta := map[string]interface{}{}
	if existing, ok := meta["geo"]; ok {
		if geoMap, isMap := existing.(map[string]interface{}); isMap {
			geoMeta = geoMap
		}
	}

	// 1. Authoritative Coordinate Validation
	hasValidCoordinates := false
	if !event.HasNoLocation && event.Latitude != nil && event.Longitude != nil {
		lat := *event.Latitude
		lon := *event.Longitude
		if lat >= -90 && lat <= 90 && lon >= -180 && lon <= 180 {
			hasValidCoordinates = true
		}
	}

	if !hasValidCoordinates {
		// Strict invariant: Unknown coordinates MUST be nil. Never use 0,0 as sentinel.
		event.Latitude = nil
		event.Longitude = nil
		event.HasNoLocation = true
	}

	// 2. Country Resolution via Non-Authoritative Heuristic (Fallback)
	// Higher-authority data must never be overwritten by lower-authority heuristics.
	if event.Country == "" && event.LocationDetails != "" {
		loc := strings.TrimSpace(event.LocationDetails)
		var inferredCountry, isoCode string
		if strings.HasSuffix(loc, "Japan") {
			inferredCountry = "Japan"
			isoCode = "JP"
		} else if strings.HasSuffix(loc, "Philippines") {
			inferredCountry = "Philippines"
			isoCode = "PH"
		} else if strings.HasSuffix(loc, "Indonesia") {
			inferredCountry = "Indonesia"
			isoCode = "ID"
		}

		if inferredCountry != "" {
			event.Country = inferredCountry
			geoMeta["iso_country_code"] = isoCode
			geoMeta["country_source"] = "heuristic"
		}
	}

	// 3. Location Confidence Determination
	if hasValidCoordinates {
		// If valid coordinates exist, confidence is exact (unless explicitly marked approximate upstream)
		if currentConf, ok := geoMeta["location_confidence"].(string); !ok || currentConf != "approximate" {
			geoMeta["location_confidence"] = "exact"
		}
	} else if event.Country != "" {
		// Resolved only to country level without exact coordinates
		geoMeta["location_confidence"] = "country"
	} else if currentConf, ok := geoMeta["location_confidence"].(string); ok && currentConf == "approximate" {
		geoMeta["location_confidence"] = "approximate"
	} else {
		// No reliable location
		geoMeta["location_confidence"] = "unknown"
	}

	meta["geo"] = geoMeta

	b, err := json.Marshal(meta)
	if err == nil {
		event.Metadata = b
	}

	return nil
}
