package domain

import (
	"context"
)

type Enricher interface {
	Name() string
	Enrich(ctx context.Context, record ExternalRecord, event *ThreatEvent) error
}

type EnrichmentPipeline interface {
	Enrich(ctx context.Context, record ExternalRecord, event *ThreatEvent) error
}
