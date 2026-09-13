package extraction

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/google/uuid"
)

func TestDeterministicExtractor_URLAndDomain(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	metaJSON, _ := json.Marshal(map[string]interface{}{
		"provenance": map[string]interface{}{
			"source_url": "https://www.reuters.com/world/asia-pacific/earthquake-hits-japan-2026-09-13/",
		},
	})

	event := &domain.ThreatEvent{
		ID:          uuid.New(),
		Title:       "M 6.5 Earthquake Reported Near Honshu",
		Description: "For details visit https://earthquake.usgs.gov/earthquakes/eventpage/us7000test and check alert mirror http://backup.usgs.gov/feed",
		Metadata:    metaJSON,
	}

	entities, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract returned unexpected error: %v", err)
	}

	// Verify we extracted URLs and their corresponding domains
	foundProvenanceURL := false
	foundProvenanceDomain := false
	foundTextURL := false
	foundTextDomain := false

	for _, e := range entities {
		if e.Type == domain.EntityTypeURL {
			if e.NormalizedValue == "https://www.reuters.com/world/asia-pacific/earthquake-hits-japan-2026-09-13/" {
				foundProvenanceURL = true
				if e.Confidence != domain.ConfidenceExact {
					t.Errorf("Expected ConfidenceExact for structured provenance URL, got %s", e.Confidence)
				}
				if e.ExtractionMethod != domain.MethodStructured {
					t.Errorf("Expected MethodStructured for provenance URL, got %s", e.ExtractionMethod)
				}
			}
			if e.NormalizedValue == "https://earthquake.usgs.gov/earthquakes/eventpage/us7000test" {
				foundTextURL = true
			}
		}
		if e.Type == domain.EntityTypeDomain {
			if e.NormalizedValue == "www.reuters.com" {
				foundProvenanceDomain = true
				if e.ExtractionMethod != domain.MethodDeterministic {
					t.Errorf("Expected MethodDeterministic for derived domain, got %s", e.ExtractionMethod)
				}
			}
			if e.NormalizedValue == "earthquake.usgs.gov" {
				foundTextDomain = true
			}
		}
	}

	if !foundProvenanceURL {
		t.Errorf("Expected provenance URL entity to be extracted")
	}
	if !foundProvenanceDomain {
		t.Errorf("Expected provenance domain entity to be extracted")
	}
	if !foundTextURL {
		t.Errorf("Expected text URL entity to be extracted")
	}
	if !foundTextDomain {
		t.Errorf("Expected text domain entity to be extracted")
	}
}

func TestDeterministicExtractor_IPAddresses(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	event := &domain.ThreatEvent{
		ID:          uuid.New(),
		Title:       "C2 node 198.51.100.25 detected communicating with 2001:db8::1",
		Description: "Telemetry software version 3.14.1 running on host with IP 203.0.113.5",
	}

	entities, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract unexpected error: %v", err)
	}

	foundIPv4 := make(map[string]bool)
	foundIPv6 := make(map[string]bool)
	foundVersion := false

	for _, e := range entities {
		if e.Type == domain.EntityTypeIP {
			if e.NormalizedValue == "198.51.100.25" {
				foundIPv4[e.NormalizedValue] = true
			}
			if e.NormalizedValue == "203.0.113.5" {
				foundIPv4[e.NormalizedValue] = true
			}
			if e.NormalizedValue == "2001:db8::1" {
				foundIPv6[e.NormalizedValue] = true
			}
			if e.NormalizedValue == "3.14.1" || e.Value == "3.14.1" {
				foundVersion = true
			}
		}
	}

	if !foundIPv4["198.51.100.25"] || !foundIPv4["203.0.113.5"] {
		t.Errorf("Expected IPv4 addresses to be extracted, got %v", foundIPv4)
	}
	if !foundIPv6["2001:db8::1"] {
		t.Errorf("Expected IPv6 2001:db8::1 to be extracted, got %v", foundIPv6)
	}
	if foundVersion {
		t.Errorf("Software version number 3.14.1 must NOT be extracted as an IP entity")
	}
}

func TestDeterministicExtractor_CountryAndLocation(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	lat := 35.6762
	lon := 139.6503
	metaJSON, _ := json.Marshal(map[string]interface{}{
		"geo": map[string]interface{}{
			"location_confidence": "exact",
			"iso_country_code":    "JP",
		},
	})

	event := &domain.ThreatEvent{
		ID:              uuid.New(),
		Country:         "Japan",
		LocationDetails: "Tokyo Bay Region",
		Latitude:        &lat,
		Longitude:       &lon,
		HasNoLocation:   false,
		Metadata:        metaJSON,
	}

	entities, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract unexpected error: %v", err)
	}

	var countryEntity, locEntity *domain.IntelligenceEntity
	for i := range entities {
		if entities[i].Type == domain.EntityTypeCountry {
			countryEntity = &entities[i]
		}
		if entities[i].Type == domain.EntityTypeLocation {
			locEntity = &entities[i]
		}
	}

	if countryEntity == nil {
		t.Fatalf("Expected country entity to be extracted")
	}
	if countryEntity.NormalizedValue != "JP" {
		t.Errorf("Expected normalized country JP, got %s", countryEntity.NormalizedValue)
	}
	if countryEntity.Value != "Japan" {
		t.Errorf("Expected raw country value 'Japan', got %s", countryEntity.Value)
	}
	// Extraction confidence reflects extraction certainty from the structured field, NOT geographic confidence
	if countryEntity.Confidence != domain.ConfidenceHigh {
		t.Errorf("Expected ConfidenceHigh for structured country entity extraction, got %s", countryEntity.Confidence)
	}

	// Verify geographic confidence remains separated in metadata
	var cMeta map[string]interface{}
	_ = json.Unmarshal(countryEntity.Metadata, &cMeta)
	if cMeta["geo_confidence"] != "exact" {
		t.Errorf("Expected geo_confidence 'exact' preserved in metadata, got %v", cMeta["geo_confidence"])
	}

	if locEntity == nil {
		t.Fatalf("Expected location entity to be extracted")
	}
	if locEntity.NormalizedValue != "tokyo bay region" {
		t.Errorf("Expected normalized location 'tokyo bay region', got %s", locEntity.NormalizedValue)
	}
	// Textual location entity confidence must NOT be elevated to ConfidenceExact simply because coordinates exist
	if locEntity.Confidence != domain.ConfidenceHigh {
		t.Errorf("Expected ConfidenceHigh for textual location entity, got %s", locEntity.Confidence)
	}

	var lMeta map[string]interface{}
	_ = json.Unmarshal(locEntity.Metadata, &lMeta)
	if lMeta["geo_confidence"] != "exact" {
		t.Errorf("Expected geo_confidence 'exact' preserved in location metadata, got %v", lMeta["geo_confidence"])
	}
}

func TestExtractionConfidence_Vs_GeographicConfidence_Separation(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	lat := 35.6762
	lon := 139.6503
	metaJSON, _ := json.Marshal(map[string]interface{}{
		"geo": map[string]interface{}{
			"location_confidence": "exact", // Geographic coordinate confidence from Step 7
			"iso_country_code":    "JP",
			"country_source":      "heuristic",
		},
	})

	event := &domain.ThreatEvent{
		ID:              uuid.New(),
		Title:           "Seismic tremor observed",
		Severity:        "low", // Threat severity dimension is independent
		Country:         "Japan",
		LocationDetails: "45km NW of Tokyo", // Textual representation is not exact point
		Latitude:        &lat,               // Coordinates are exact point
		Longitude:       &lon,
		HasNoLocation:   false,
		Metadata:        metaJSON,
	}

	entities, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract unexpected error: %v", err)
	}

	var locEntity, countryEntity *domain.IntelligenceEntity
	for i := range entities {
		if entities[i].Type == domain.EntityTypeLocation {
			locEntity = &entities[i]
		}
		if entities[i].Type == domain.EntityTypeCountry {
			countryEntity = &entities[i]
		}
	}

	if locEntity == nil {
		t.Fatalf("Expected location entity")
	}
	// 1. Textual location must NOT be elevated to exact just because event has coordinates
	if locEntity.Confidence == domain.ConfidenceExact {
		t.Errorf("Textual location entity '45km NW of Tokyo' must NOT be ConfidenceExact merely because coordinates exist")
	}
	if locEntity.Confidence != domain.ConfidenceHigh {
		t.Errorf("Expected ConfidenceHigh for structured LocationDetails extraction, got %s", locEntity.Confidence)
	}

	// 2. Geographic confidence remains separate in metadata
	var lMeta map[string]interface{}
	_ = json.Unmarshal(locEntity.Metadata, &lMeta)
	if lMeta["geo_confidence"] != "exact" {
		t.Errorf("Expected Step 7 geo_confidence to be recorded separately in metadata")
	}

	// 3. Country with heuristic source reflects extraction confidence Medium, not elevated by geo exact
	if countryEntity == nil {
		t.Fatalf("Expected country entity")
	}
	if countryEntity.Confidence != domain.ConfidenceMedium {
		t.Errorf("Expected ConfidenceMedium for heuristic country extraction, got %s", countryEntity.Confidence)
	}

	// 4. Source authority does not alter threat severity
	if event.Severity != "low" {
		t.Errorf("Threat severity must remain untouched by entity extraction: %s", event.Severity)
	}
}

func TestEntityProvenance_MultiEventIsolation(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	// Event A: domain mentioned in Title
	eventA := &domain.ThreatEvent{
		ID:          uuid.New(),
		Title:       "Suspected phishing host at https://example.com/login",
		Description: "Monitoring incident report",
	}

	// Event B: domain mentioned in structured Provenance URL
	provMeta, _ := json.Marshal(map[string]interface{}{
		"provenance": map[string]interface{}{
			"source_url": "https://example.com/bulletin/2026",
		},
	})
	eventB := &domain.ThreatEvent{
		ID:          uuid.New(),
		Title:       "Official Advisory",
		Description: "Official advisory bulletin",
		Metadata:    provMeta,
	}

	entitiesA, err := extractor.Extract(ctx, eventA)
	if err != nil {
		t.Fatalf("Event A extraction error: %v", err)
	}

	entitiesB, err := extractor.Extract(ctx, eventB)
	if err != nil {
		t.Fatalf("Event B extraction error: %v", err)
	}

	var domainA, domainB *domain.IntelligenceEntity
	for i := range entitiesA {
		if entitiesA[i].Type == domain.EntityTypeDomain && entitiesA[i].NormalizedValue == "example.com" {
			domainA = &entitiesA[i]
		}
	}
	for i := range entitiesB {
		if entitiesB[i].Type == domain.EntityTypeDomain && entitiesB[i].NormalizedValue == "example.com" {
			domainB = &entitiesB[i]
		}
	}

	if domainA == nil || domainB == nil {
		t.Fatalf("Expected both events to extract domain 'example.com'")
	}

	// Both share the identical canonical global identity
	if domainA.NormalizedValue != domainB.NormalizedValue {
		t.Errorf("Global normalized value must match: %s vs %s", domainA.NormalizedValue, domainB.NormalizedValue)
	}
	if domainA.Type != domainB.Type {
		t.Errorf("Global entity type must match: %s vs %s", domainA.Type, domainB.Type)
	}

	// Event-specific context remains distinct and isolated
	if domainA.SourceField != "title" {
		t.Errorf("Event A source field must be 'title', got %s", domainA.SourceField)
	}
	if domainB.SourceField != "provenance.source_url" {
		t.Errorf("Event B source field must be 'provenance.source_url', got %s", domainB.SourceField)
	}

	// Event B's extraction does not overwrite or mutate Event A's evidence
	if domainA.SourceField == domainB.SourceField {
		t.Errorf("Event-specific source fields must remain isolated between events")
	}
}

func TestURLQuery_PreservationAndRepeatedParams(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	rawURL := "https://example.com/api?z=9&a=1&item=first&item=second"
	event := &domain.ThreatEvent{
		ID:          uuid.New(),
		Title:       "Advisory containing " + rawURL,
		Description: "Detailed report",
	}

	entities, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}

	var urlEntity *domain.IntelligenceEntity
	for i := range entities {
		if entities[i].Type == domain.EntityTypeURL {
			urlEntity = &entities[i]
		}
	}

	if urlEntity == nil {
		t.Fatalf("Expected URL entity to be extracted")
	}

	// Verify raw URL is preserved in Value
	if urlEntity.Value != rawURL {
		t.Errorf("Expected raw URL preserved in Value: want %q, got %q", rawURL, urlEntity.Value)
	}

	// Verify query order and repeated parameters are preserved in NormalizedValue
	if urlEntity.NormalizedValue != rawURL {
		t.Errorf("Expected query order and repeated params preserved in NormalizedValue: want %q, got %q", rawURL, urlEntity.NormalizedValue)
	}
}

func TestStep7CountryCompatibility(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	metaJSON, _ := json.Marshal(map[string]interface{}{
		"geo": map[string]interface{}{
			"iso_country_code":    "PH",
			"location_confidence": "country",
		},
	})

	event := &domain.ThreatEvent{
		ID:       uuid.New(),
		Title:    "Typhoon signal in Luzon",
		Country:  "Philippines",
		Metadata: metaJSON,
	}

	entities, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract error: %v", err)
	}

	var countryEntity *domain.IntelligenceEntity
	for i := range entities {
		if entities[i].Type == domain.EntityTypeCountry {
			countryEntity = &entities[i]
		}
	}

	if countryEntity == nil {
		t.Fatalf("Expected country entity")
	}
	if countryEntity.NormalizedValue != "PH" {
		t.Errorf("Expected Step 7 ISO code 'PH', got %s", countryEntity.NormalizedValue)
	}
	if countryEntity.Value != "Philippines" {
		t.Errorf("Expected raw country 'Philippines', got %s", countryEntity.Value)
	}
	if countryEntity.Confidence != domain.ConfidenceHigh {
		t.Errorf("Expected ConfidenceHigh for structured country, got %s", countryEntity.Confidence)
	}
}

func TestDeterministicExtractor_ProseIgnored(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	event := &domain.ThreatEvent{
		ID:          uuid.New(),
		Title:       "President John Doe spoke with Red Cross officials yesterday",
		Description: "A diplomatic delegation met in an undisclosed location with Cyber Command operatives.",
	}

	entities, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract unexpected error: %v", err)
	}

	for _, e := range entities {
		if e.Type == domain.EntityTypePerson || e.Type == domain.EntityTypeOrganization || e.Type == domain.EntityTypeGroup {
			t.Fatalf("Prose extraction for %s must NOT occur without reliable structured ground: found %s", e.Type, e.Value)
		}
	}
}

func TestDeterministicExtractor_AuthoritativeEventImmutability(t *testing.T) {
	extractor := NewDeterministicExtractor()
	ctx := context.Background()

	lat := 14.5995
	lon := 120.9842
	occTime := time.Date(2026, 9, 13, 10, 0, 0, 0, time.UTC)
	detTime := time.Date(2026, 9, 13, 10, 5, 0, 0, time.UTC)

	event := &domain.ThreatEvent{
		ID:              uuid.New(),
		Title:           "Flooding in Manila",
		Description:     "Report published on https://reliefweb.int/report/123",
		Country:         "Philippines",
		LocationDetails: "Manila City",
		Latitude:        &lat,
		Longitude:       &lon,
		OccurredAt:      occTime,
		DetectedAt:      detTime,
		HasNoLocation:   false,
	}

	// Capture values prior to extraction
	origCountry := event.Country
	origLoc := event.LocationDetails
	origLat := *event.Latitude
	origLon := *event.Longitude
	origOcc := event.OccurredAt
	origDet := event.DetectedAt

	_, err := extractor.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Extract unexpected error: %v", err)
	}

	// Assert complete immutability of ThreatEvent
	if event.Country != origCountry {
		t.Errorf("ThreatEvent.Country mutated! Was %s, now %s", origCountry, event.Country)
	}
	if event.LocationDetails != origLoc {
		t.Errorf("ThreatEvent.LocationDetails mutated! Was %s, now %s", origLoc, event.LocationDetails)
	}
	if event.Latitude == nil || *event.Latitude != origLat {
		t.Errorf("ThreatEvent.Latitude mutated!")
	}
	if event.Longitude == nil || *event.Longitude != origLon {
		t.Errorf("ThreatEvent.Longitude mutated!")
	}
	if !event.OccurredAt.Equal(origOcc) {
		t.Errorf("ThreatEvent.OccurredAt mutated!")
	}
	if !event.DetectedAt.Equal(origDet) {
		t.Errorf("ThreatEvent.DetectedAt mutated!")
	}
}

func TestPipeline_IdentityDeduplicationAndIdempotency(t *testing.T) {
	pipeline := NewPipeline(NewDeterministicExtractor())
	ctx := context.Background()

	event := &domain.ThreatEvent{
		ID:          uuid.New(),
		Title:       "Alert from https://EXAMPLE.COM/feed and https://example.com/feed",
		Description: "Additional link at https://example.com/other",
	}

	// First pass
	entities1, err := pipeline.Extract(ctx, event)
	if err != nil {
		t.Fatalf("First pass unexpected error: %v", err)
	}

	// Second pass
	entities2, err := pipeline.Extract(ctx, event)
	if err != nil {
		t.Fatalf("Second pass unexpected error: %v", err)
	}

	// Idempotency: both passes must yield exactly the same number of entities with identical values
	if len(entities1) != len(entities2) {
		t.Fatalf("Idempotency failure: pass 1 gave %d entities, pass 2 gave %d", len(entities1), len(entities2))
	}

	domainCount := 0
	urlCount := 0
	for _, e := range entities1 {
		if e.Type == domain.EntityTypeDomain {
			domainCount++
			if e.NormalizedValue != "example.com" {
				t.Errorf("Expected normalized domain 'example.com', got %s", e.NormalizedValue)
			}
		}
		if e.Type == domain.EntityTypeURL {
			urlCount++
		}
	}

	// Case variation deduplication: EXAMPLE.COM and example.com must yield exactly 1 domain entity
	if domainCount != 1 {
		t.Errorf("Expected exactly 1 deduplicated domain entity, got %d", domainCount)
	}

	// Distinct URLs remain distinct: /feed vs /other
	if urlCount != 2 {
		t.Errorf("Expected 2 distinct URL entities (/feed and /other), got %d", urlCount)
	}
}

func TestRegression_Step7AndStep6Guarantees(t *testing.T) {
	pipeline := NewPipeline(NewDeterministicExtractor())
	ctx := context.Background()

	t.Run("Valid 0,0 Null Island Coordinates Preserved", func(t *testing.T) {
		lat := 0.0
		lon := 0.0
		event := &domain.ThreatEvent{
			ID:            uuid.New(),
			Latitude:      &lat,
			Longitude:     &lon,
			HasNoLocation: false,
		}

		_, err := pipeline.Extract(ctx, event)
		if err != nil {
			t.Fatalf("Extract unexpected error: %v", err)
		}

		if event.Latitude == nil || *event.Latitude != 0.0 {
			t.Errorf("Null Island latitude was cleared or modified: %v", event.Latitude)
		}
		if event.Longitude == nil || *event.Longitude != 0.0 {
			t.Errorf("Null Island longitude was cleared or modified: %v", event.Longitude)
		}
	})

	t.Run("Nil Coordinates Remain Nil", func(t *testing.T) {
		event := &domain.ThreatEvent{
			ID:            uuid.New(),
			Latitude:      nil,
			Longitude:     nil,
			HasNoLocation: true,
		}

		_, err := pipeline.Extract(ctx, event)
		if err != nil {
			t.Fatalf("Extract unexpected error: %v", err)
		}

		if event.Latitude != nil || event.Longitude != nil {
			t.Errorf("Nil coordinates fabricated during entity extraction")
		}
	})

	t.Run("GDELT SourceCountry Does Not Create Event Country Entity", func(t *testing.T) {
		provMeta, _ := json.Marshal(map[string]interface{}{
			"provenance": map[string]interface{}{
				"source_country": "US",
				"source_url":     "https://gdeltproject.org/data/article1",
			},
		})

		event := &domain.ThreatEvent{
			ID:       uuid.New(),
			Country:  "", // GDELT event country is unknown
			Metadata: provMeta,
		}

		entities, err := pipeline.Extract(ctx, event)
		if err != nil {
			t.Fatalf("Extract unexpected error: %v", err)
		}

		// Verify NO country entity was created for the event
		for _, e := range entities {
			if e.Type == domain.EntityTypeCountry {
				t.Fatalf("GDELT sourcecountry must NOT be extracted as an event country entity! Found: %s", e.Value)
			}
		}

		// ThreatEvent.Country must remain empty
		if event.Country != "" {
			t.Fatalf("ThreatEvent.Country must remain empty, got %s", event.Country)
		}
	})

	t.Run("Missing OccurredAt Not Fabricated", func(t *testing.T) {
		event := &domain.ThreatEvent{
			ID:               uuid.New(),
			OccurredAt:       time.Time{},
			EventTimeUnknown: true,
		}

		_, err := pipeline.Extract(ctx, event)
		if err != nil {
			t.Fatalf("Extract unexpected error: %v", err)
		}

		if !event.OccurredAt.IsZero() {
			t.Errorf("OccurredAt was fabricated during extraction!")
		}
	})
}
