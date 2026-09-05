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

	// 1. Determine Location Confidence
	if !event.HasNoLocation {
		// We trust the authoritative source connector to have set HasNoLocation=false 
		// only if the coordinates are real.
		geoMeta["location_confidence"] = "exact"
	} else {
		// Ensure zeroed coordinates so no 0,0 sentinel makes it to a confusing JSON state, 
		// though DB layer maps to NULL because of HasNoLocation=true.
		event.Latitude = 0
		event.Longitude = 0
		
		if event.Country != "" {
			geoMeta["location_confidence"] = "country"
		} else {
			geoMeta["location_confidence"] = "unknown"
		}
	}

	// 2. Simple Country Resolution
	// Example USGS Place: "45 km E of Hachinohe, Japan"
	if event.Country == "" && event.LocationDetails != "" {
		// Simple static heuristic to serve as foundation for future country-service integration
		loc := strings.TrimSpace(event.LocationDetails)
		if strings.HasSuffix(loc, "Japan") {
			event.Country = "Japan"
			geoMeta["iso_country_code"] = "JP"
		} else if strings.HasSuffix(loc, "Philippines") {
			event.Country = "Philippines"
			geoMeta["iso_country_code"] = "PH"
		} else if strings.HasSuffix(loc, "Indonesia") {
			event.Country = "Indonesia"
			geoMeta["iso_country_code"] = "ID"
		}

		// Update confidence if we found a country and didn't have exact coordinates
		if event.Country != "" && event.HasNoLocation {
			geoMeta["location_confidence"] = "country"
		}
	}

	meta["geo"] = geoMeta

	b, err := json.Marshal(meta)
	if err == nil {
		event.Metadata = b
	}

	return nil
}
