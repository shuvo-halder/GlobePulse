package ingestion

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/google/uuid"
)

type Scheduler struct {
	connectors []domain.Connector
	repo       domain.IngestionRepository
	enricher   domain.EnrichmentPipeline
	interval   time.Duration
	stats      map[string]*ConnectorStats
	statsMu    sync.RWMutex
	wg         sync.WaitGroup
}

func NewScheduler(repo domain.IngestionRepository, enricher domain.EnrichmentPipeline, interval time.Duration) *Scheduler {
	return &Scheduler{
		repo:     repo,
		enricher: enricher,
		interval: interval,
		stats:    make(map[string]*ConnectorStats),
	}
}

func (s *Scheduler) Register(c domain.Connector) {
	s.connectors = append(s.connectors, c)
	s.statsMu.Lock()
	defer s.statsMu.Unlock()
	s.stats[c.Name()] = NewConnectorStats(c.Name())
}

func (s *Scheduler) GetStats() map[string]interface{} {
	s.statsMu.RLock()
	defer s.statsMu.RUnlock()

	result := make(map[string]interface{})
	for name, stats := range s.stats {
		result[name] = stats.Snapshot()
	}
	return result
}

func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("Starting ingestion scheduler")

	// Trigger an initial run for all connectors immediately
	for _, c := range s.connectors {
		s.launchConnector(ctx, c)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping ingestion scheduler, waiting for active jobs to finish")
			s.wg.Wait()
			slog.Info("Ingestion scheduler stopped cleanly")
			return
		case <-ticker.C:
			for _, c := range s.connectors {
				s.launchConnector(ctx, c)
			}
		}
	}
}

func (s *Scheduler) launchConnector(ctx context.Context, c domain.Connector) {
	s.statsMu.RLock()
	st, ok := s.stats[c.Name()]
	s.statsMu.RUnlock()
	
	if !ok {
		return
	}

	if !st.TryStart() {
		slog.Warn("Skipping connector run due to overlap", "name", c.Name())
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		
		// Panic recovery for isolation
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Connector panic", "name", c.Name(), "panic", r)
				st.MarkFailure(0, nil)
			}
		}()

		start := time.Now()
		slog.Info("Running connector", "name", c.Name())

		sourceID, err := s.repo.GetOrCreateSource(ctx, c.Name(), c.SourceType(), c.BaseURL())
		if err != nil {
			slog.Error("Failed to get/create source", "name", c.Name(), "error", err)
			st.MarkFailure(time.Since(start), err)
			return
		}

		records, err := c.Fetch(ctx)
		if err != nil {
			slog.Error("Connector fetch failed", "name", c.Name(), "error", err)
			st.MarkFailure(time.Since(start), err)
			return
		}

		slog.Info("Fetched records", "name", c.Name(), "count", len(records))

		var fetched, normalized, inserted, duplicate, rejected int
		fetched = len(records)

		for _, rec := range records {
			event, err := c.Normalize(rec)
			if err != nil {
				slog.Error("Failed to normalize record", "external_id", rec.ExternalID, "error", err)
				rejected++
				continue
			}
			normalized++

			if s.enricher != nil {
				if err := s.enricher.Enrich(ctx, rec, event); err != nil {
					slog.Warn("Enrichment failure", "external_id", rec.ExternalID, "error", err)
					// Non-fatal, we continue to save the normalized event
				}
			}

			item := &domain.SourceItem{
				ID:          uuid.New(),
				ExternalID:  rec.ExternalID,
				URL:         rec.URL,
				Title:       rec.Title,
				RawMetadata: rec.RawMetadata,
				PublishedAt: rec.PublishedAt,
			}

			saveResult, err := s.repo.SaveItemAndEvent(ctx, sourceID, item, event)
			if err != nil {
				slog.Error("Failed to save item/event", "error", err)
				rejected++
			} else {
				if saveResult == domain.SaveDuplicate {
					duplicate++
				} else {
					inserted++
				}
			}
		}

		duration := time.Since(start)
		slog.Info("Ingestion cycle complete", 
			"name", c.Name(), 
			"duration", duration.String(),
			"fetched", fetched, 
			"normalized", normalized, 
			"inserted", inserted,
			"duplicate", duplicate,
			"rejected", rejected)

		st.MarkSuccess(duration, fetched, normalized, inserted, duplicate, rejected)
	}()
}
