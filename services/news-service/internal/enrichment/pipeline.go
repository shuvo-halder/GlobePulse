package enrichment

import (
	"context"
	"log/slog"

	"github.com/global-news/news-service/internal/domain"
)

type Pipeline struct {
	enrichers []domain.Enricher
}

func NewPipeline(enrichers ...domain.Enricher) *Pipeline {
	return &Pipeline{
		enrichers: enrichers,
	}
}

func (p *Pipeline) Enrich(ctx context.Context, record domain.ExternalRecord, event *domain.ThreatEvent) error {
	for _, enricher := range p.enrichers {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := enricher.Enrich(ctx, record, event)
		if err != nil {
			// Enrichment failures are generally non-fatal to the ingestion process,
			// but we log them. If a specific enricher considers its failure fatal,
			// it should be handled differently, but standard enrichment is best-effort.
			slog.Warn("Enricher failed (non-fatal)",
				"enricher", enricher.Name(),
				"external_id", record.ExternalID,
				"error", err)
		}
	}
	return nil
}
