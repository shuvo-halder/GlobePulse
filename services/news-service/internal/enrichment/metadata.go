package enrichment

import (
	"context"
	"encoding/json"

	"github.com/global-news/news-service/internal/domain"
)

type MetadataEnricher struct{}

func NewMetadataEnricher() *MetadataEnricher {
	return &MetadataEnricher{}
}

func (e *MetadataEnricher) Name() string {
	return "MetadataEnricher"
}

func (e *MetadataEnricher) Enrich(ctx context.Context, record domain.ExternalRecord, event *domain.ThreatEvent) error {
	var meta map[string]interface{}
	if len(event.Metadata) > 0 {
		if err := json.Unmarshal(event.Metadata, &meta); err != nil {
			meta = make(map[string]interface{})
		}
	} else {
		meta = make(map[string]interface{})
	}

	flags := []string{}

	if event.HasNoLocation {
		flags = append(flags, "missing_location")
	}
	if event.EventTimeUnknown {
		flags = append(flags, "missing_event_time")
	}
	if record.PublishedAt.IsZero() {
		flags = append(flags, "source_timestamp_missing")
	}
	if event.Title == "" {
		flags = append(flags, "missing_title")
	}

	// Idempotent merge of flags
	existingFlagsSet := make(map[string]bool)
	if existing, ok := meta["data_quality_flags"]; ok {
		if arr, isArr := existing.([]interface{}); isArr {
			for _, f := range arr {
				if strFlag, ok := f.(string); ok {
					existingFlagsSet[strFlag] = true
				}
			}
		}
	}

	for _, f := range flags {
		existingFlagsSet[f] = true
	}

	var finalFlags []string
	for f := range existingFlagsSet {
		finalFlags = append(finalFlags, f)
	}

	meta["data_quality_flags"] = finalFlags

	b, err := json.Marshal(meta)
	if err == nil {
		event.Metadata = b
	}

	return nil
}
