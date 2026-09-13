package extraction

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/google/uuid"
)

var (
	// Strict URL pattern for text scanning
	urlRegex = regexp.MustCompile(`https?://[^\s<>"'` + "`" + `]+`)

	// Strict IPv4 boundary pattern
	ipv4Regex = regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)

	// Standard IPv6 boundary pattern
	ipv6Regex = regexp.MustCompile(`\b(?:[a-fA-F0-9]{1,4}:){7}[a-fA-F0-9]{1,4}\b|\b(?:[a-fA-F0-9]{1,4}:)*:[a-fA-F0-9]{1,4}\b`)
)

type DeterministicExtractor struct{}

func NewDeterministicExtractor() *DeterministicExtractor {
	return &DeterministicExtractor{}
}

func (e *DeterministicExtractor) Name() string {
	return "deterministic_extractor"
}

// Extract extracts intelligence entities from canonical ThreatEvent data without mutating the event.
func (e *DeterministicExtractor) Extract(ctx context.Context, event *domain.ThreatEvent) ([]domain.IntelligenceEntity, error) {
	if event == nil {
		return nil, nil
	}

	var entities []domain.IntelligenceEntity
	seen := make(map[string]bool) // Key: entity_type + ":" + normalized_value

	addEntity := func(entity domain.IntelligenceEntity) {
		key := string(entity.Type) + ":" + entity.NormalizedValue
		if !seen[key] {
			seen[key] = true
			if entity.ID == uuid.Nil {
				entity.ID = uuid.New()
			}
			entities = append(entities, entity)
		}
	}

	// Parse event metadata if present
	var meta map[string]interface{}
	if len(event.Metadata) > 0 {
		_ = json.Unmarshal(event.Metadata, &meta)
	}

	// 1. Extract Country Entity from structured event.Country
	if event.Country != "" {
		var normCountry string
		var err error

		// Step 7 authoritative ISO country code check first: avoid conflicting country authority
		if meta != nil {
			if geo, ok := meta["geo"].(map[string]interface{}); ok {
				if iso, ok := geo["iso_country_code"].(string); ok && len(strings.TrimSpace(iso)) == 2 {
					normCountry = strings.ToUpper(strings.TrimSpace(iso))
				}
			}
		}
		if normCountry == "" {
			normCountry, err = NormalizeCountry(event.Country)
		}

		if err == nil && normCountry != "" {
			// Extraction confidence reflects certainty of entity extraction from the event model
			conf := domain.ConfidenceHigh
			if meta != nil {
				if geo, ok := meta["geo"].(map[string]interface{}); ok {
					if geo["country_source"] == "heuristic" {
						conf = domain.ConfidenceMedium
					}
				}
			}

			countryMeta := map[string]interface{}{
				"extracted_at": time.Now().UTC().Format(time.RFC3339),
				"source_field": "country",
				"country_name": event.Country,
				"iso_code":     normCountry,
			}
			// Maintain geographic confidence as a separate dimension in metadata without conflating with extraction confidence
			if meta != nil {
				if geo, ok := meta["geo"].(map[string]interface{}); ok {
					if geoConf, ok := geo["location_confidence"].(string); ok && geoConf != "" {
						countryMeta["geo_confidence"] = geoConf
					}
				}
			}

			metaJSON, _ := json.Marshal(countryMeta)

			addEntity(domain.IntelligenceEntity{
				Type:             domain.EntityTypeCountry,
				Value:            event.Country,
				NormalizedValue:  normCountry,
				Confidence:       conf,
				ExtractionMethod: domain.MethodStructured,
				SourceField:      "country",
				Metadata:         metaJSON,
			})
		}
	}

	// 2. Extract Location Entity from structured event.LocationDetails
	if event.LocationDetails != "" {
		normLoc, err := NormalizeLocation(event.LocationDetails)
		if err == nil {
			// Textual location entity extraction confidence reflects certainty of extraction from the structured field.
			// Textual location is NOT automatically classified as geographically exact simply because the event has coordinates.
			// Event coordinate confidence remains strictly governed by Step 7.
			conf := domain.ConfidenceHigh

			locMeta := map[string]interface{}{
				"extracted_at": time.Now().UTC().Format(time.RFC3339),
				"source_field": "location_details",
			}
			if !event.HasNoLocation && event.Latitude != nil && event.Longitude != nil {
				locMeta["event_coordinates"] = []float64{*event.Latitude, *event.Longitude}
			}
			if meta != nil {
				if geo, ok := meta["geo"].(map[string]interface{}); ok {
					if geoConf, ok := geo["location_confidence"].(string); ok && geoConf != "" {
						locMeta["geo_confidence"] = geoConf
					}
				}
			}

			metaJSON, _ := json.Marshal(locMeta)

			addEntity(domain.IntelligenceEntity{
				Type:             domain.EntityTypeLocation,
				Value:            event.LocationDetails,
				NormalizedValue:  normLoc,
				Confidence:       conf,
				ExtractionMethod: domain.MethodStructured,
				SourceField:      "location_details",
				Metadata:         metaJSON,
			})
		}
	}

	// 3. Extract Provenance URL and Domain from event.Metadata
	if meta != nil {
		if prov, ok := meta["provenance"].(map[string]interface{}); ok {
			if sourceURL, ok := prov["source_url"].(string); ok && sourceURL != "" {
				e.extractURLAndDomain(sourceURL, "provenance.source_url", domain.MethodStructured, domain.ConfidenceExact, addEntity)
			}
		}
	}

	// 4. Extract URLs, Domains, and IPs from Title and Description
	scanText := func(text, sourceField string) {
		if text == "" {
			return
		}

		// A. Scan URLs
		urls := urlRegex.FindAllString(text, -1)
		for _, rawURL := range urls {
			e.extractURLAndDomain(rawURL, sourceField, domain.MethodRegex, domain.ConfidenceHigh, addEntity)
		}

		// B. Scan IPv4
		ipv4Matches := ipv4Regex.FindAllString(text, -1)
		for _, rawIP := range ipv4Matches {
			ip := net.ParseIP(rawIP)
			if ip != nil && ip.To4() != nil {
				normIP, err := NormalizeIP(rawIP)
				if err == nil {
					metaJSON, _ := json.Marshal(map[string]interface{}{
						"extracted_at": time.Now().UTC().Format(time.RFC3339),
						"source_field": sourceField,
						"ip_version":   "IPv4",
					})

					addEntity(domain.IntelligenceEntity{
						Type:             domain.EntityTypeIP,
						Value:            rawIP,
						NormalizedValue:  normIP,
						Confidence:       domain.ConfidenceHigh,
						ExtractionMethod: domain.MethodRegex,
						SourceField:      sourceField,
						Metadata:         metaJSON,
					})
				}
			}
		}

		// C. Scan IPv6
		ipv6Matches := ipv6Regex.FindAllString(text, -1)
		for _, rawIP := range ipv6Matches {
			ip := net.ParseIP(rawIP)
			if ip != nil && ip.To4() == nil {
				normIP, err := NormalizeIP(rawIP)
				if err == nil {
					metaJSON, _ := json.Marshal(map[string]interface{}{
						"extracted_at": time.Now().UTC().Format(time.RFC3339),
						"source_field": sourceField,
						"ip_version":   "IPv6",
					})

					addEntity(domain.IntelligenceEntity{
						Type:             domain.EntityTypeIP,
						Value:            rawIP,
						NormalizedValue:  normIP,
						Confidence:       domain.ConfidenceHigh,
						ExtractionMethod: domain.MethodRegex,
						SourceField:      sourceField,
						Metadata:         metaJSON,
					})
				}
			}
		}
	}

	scanText(event.Title, "title")
	scanText(event.Description, "description")

	return entities, nil
}

func (e *DeterministicExtractor) extractURLAndDomain(
	rawURL string,
	sourceField string,
	method domain.ExtractionMethod,
	urlConf domain.ExtractionConfidence,
	addEntity func(domain.IntelligenceEntity),
) {
	normURL, err := NormalizeURL(rawURL)
	if err != nil {
		return
	}

	metaJSON, _ := json.Marshal(map[string]interface{}{
		"extracted_at": time.Now().UTC().Format(time.RFC3339),
		"source_field": sourceField,
	})

	addEntity(domain.IntelligenceEntity{
		Type:             domain.EntityTypeURL,
		Value:            rawURL,
		NormalizedValue:  normURL,
		Confidence:       urlConf,
		ExtractionMethod: method,
		SourceField:      sourceField,
		Metadata:         metaJSON,
	})

	// Extract Domain from the valid URL
	parsed, err := url.Parse(normURL)
	if err == nil && parsed.Host != "" {
		normDomain, err := NormalizeDomain(parsed.Host)
		if err == nil {
			domainMetaJSON, _ := json.Marshal(map[string]interface{}{
				"extracted_at": time.Now().UTC().Format(time.RFC3339),
				"source_field": sourceField,
				"parent_url":   normURL,
			})

			addEntity(domain.IntelligenceEntity{
				Type:             domain.EntityTypeDomain,
				Value:            parsed.Host,
				NormalizedValue:  normDomain,
				Confidence:       domain.ConfidenceHigh,
				ExtractionMethod: domain.MethodDeterministic,
				SourceField:      sourceField,
				Metadata:         domainMetaJSON,
			})
		}
	}
}
