package enrichment

import (
	"context"
	"encoding/json"
	"time"

	"github.com/global-news/news-service/internal/domain"
)

type ProvenanceEnricher struct{}

func NewProvenanceEnricher() *ProvenanceEnricher {
	return &ProvenanceEnricher{}
}

func (e *ProvenanceEnricher) Name() string {
	return "ProvenanceEnricher"
}

func (e *ProvenanceEnricher) Enrich(ctx context.Context, record domain.ExternalRecord, event *domain.ThreatEvent) error {
	var meta map[string]interface{}
	if len(event.Metadata) > 0 {
		if err := json.Unmarshal(event.Metadata, &meta); err != nil {
			meta = make(map[string]interface{})
		}
	} else {
		meta = make(map[string]interface{})
	}

	// Only add provenance if it doesn't already exist to preserve idempotency
	if _, exists := meta["provenance"]; !exists {
		prov := map[string]interface{}{
			"external_id":         record.ExternalID,
			"source_url":          record.URL,
			"normalized_at":       time.Now().UTC().Format(time.RFC3339),
			"source_published_at": record.PublishedAt.UTC().Format(time.RFC3339),
		}

		// Extract source context if present in raw metadata (e.g. GDELT sourcecountry)
		if len(record.RawMetadata) > 0 {
			var rawMap map[string]interface{}
			if err := json.Unmarshal(record.RawMetadata, &rawMap); err == nil {
				if sc, ok := rawMap["sourcecountry"].(string); ok && sc != "" {
					prov["source_country"] = sc
				}
			}
		}

		meta["provenance"] = prov
		
		b, err := json.Marshal(meta)
		if err == nil {
			event.Metadata = b
		}
		return err
	}
	return nil
}
