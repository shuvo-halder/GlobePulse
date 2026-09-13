package extraction

import (
	"context"
	"sort"

	"github.com/global-news/news-service/internal/domain"
)

type Pipeline struct {
	extractors []domain.EntityExtractor
}

func NewPipeline(extractors ...domain.EntityExtractor) *Pipeline {
	return &Pipeline{extractors: extractors}
}

// Extract executes all registered extractors sequentially, respects context cancellation,
// deduplicates entities deterministically by (Type, NormalizedValue), and returns a deterministically sorted slice.
func (p *Pipeline) Extract(ctx context.Context, event *domain.ThreatEvent) ([]domain.IntelligenceEntity, error) {
	if event == nil {
		return nil, nil
	}

	entityMap := make(map[string]domain.IntelligenceEntity) // key: type + ":" + normalized_value

	for _, extractor := range p.extractors {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		extracted, err := extractor.Extract(ctx, event)
		if err != nil {
			// Non-fatal extractor failure: log/continue with remaining extractors
			continue
		}

		for _, entity := range extracted {
			key := string(entity.Type) + ":" + entity.NormalizedValue
			if existing, exists := entityMap[key]; exists {
				// If existing entity has lower confidence, upgrade it
				if confidenceRank(entity.Confidence) > confidenceRank(existing.Confidence) {
					entityMap[key] = entity
				}
			} else {
				entityMap[key] = entity
			}
		}
	}

	// Deterministic sorting by Type, then NormalizedValue
	results := make([]domain.IntelligenceEntity, 0, len(entityMap))
	for _, entity := range entityMap {
		results = append(results, entity)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Type != results[j].Type {
			return results[i].Type < results[j].Type
		}
		return results[i].NormalizedValue < results[j].NormalizedValue
	})

	return results, nil
}

func confidenceRank(c domain.ExtractionConfidence) int {
	switch c {
	case domain.ConfidenceExact:
		return 4
	case domain.ConfidenceHigh:
		return 3
	case domain.ConfidenceMedium:
		return 2
	case domain.ConfidenceLow:
		return 1
	default:
		return 0
	}
}
