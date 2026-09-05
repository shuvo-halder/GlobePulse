package ingestion

import (
	"context"
	"log/slog"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/google/uuid"
)

type Scheduler struct {
	connectors []domain.Connector
	repo       domain.IngestionRepository
	interval   time.Duration
}

func NewScheduler(repo domain.IngestionRepository, interval time.Duration) *Scheduler {
	return &Scheduler{
		repo:       repo,
		interval:   interval,
	}
}

func (s *Scheduler) Register(c domain.Connector) {
	s.connectors = append(s.connectors, c)
}

func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("Starting ingestion scheduler")
	s.runOnce(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping ingestion scheduler")
			return
		case <-ticker.C:
			s.runOnce(ctx)
		}
	}
}

func (s *Scheduler) runOnce(ctx context.Context) {
	for _, c := range s.connectors {
		slog.Info("Running connector", "name", c.Name())

		sourceID, err := s.repo.GetOrCreateSource(ctx, c.Name(), c.SourceType(), c.BaseURL())
		if err != nil {
			slog.Error("Failed to get/create source", "name", c.Name(), "error", err)
			continue
		}

		records, err := c.Fetch(ctx)
		if err != nil {
			slog.Error("Connector fetch failed", "name", c.Name(), "error", err)
			continue
		}

		slog.Info("Fetched records", "name", c.Name(), "count", len(records))

		var savedCount int
		for _, rec := range records {
			event, err := c.Normalize(rec)
			if err != nil {
				slog.Error("Failed to normalize record", "external_id", rec.ExternalID, "error", err)
				continue
			}

			item := &domain.SourceItem{
				ID:          uuid.New(),
				ExternalID:  rec.ExternalID,
				URL:         rec.URL,
				Title:       rec.Title,
				RawMetadata: rec.RawMetadata,
				PublishedAt: rec.PublishedAt,
			}

			err = s.repo.SaveItemAndEvent(ctx, sourceID, item, event)
			if err != nil {
				slog.Error("Failed to save item/event", "error", err)
			} else {
				savedCount++
			}
		}
		slog.Info("Ingestion cycle complete", "name", c.Name(), "processed", len(records))
	}
}
