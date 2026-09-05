#!/bin/bash
set -e

mkdir -p services/news-service/internal/domain
mkdir -p services/news-service/internal/ingestion/connectors
mkdir -p services/news-service/internal/repository/postgres

# Update go.mod
cat << 'MOD' > services/news-service/go.mod
module github.com/global-news/news-service

go 1.21

require (
	github.com/google/uuid v1.3.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.9
)
MOD

# 1. Domain Models
cat << 'DOMAIN' > services/news-service/internal/domain/ingestion.go
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Source struct {
	ID         uuid.UUID
	Name       string
	SourceType string
	BaseURL    string
}

type SourceItem struct {
	ID          uuid.UUID
	SourceID    uuid.UUID
	ExternalID  string
	URL         string
	Title       string
	RawMetadata []byte
	PublishedAt time.Time
	CollectedAt time.Time
}

type ThreatEvent struct {
	ID              uuid.UUID
	Title           string
	Description     string
	EventType       string
	Category        string
	Severity        string
	Confidence      float64
	OccurredAt      time.Time
	DetectedAt      time.Time
	Latitude        float64
	Longitude       float64
	Country         string
	LocationDetails string
	Status          string
	Metadata        []byte
}

type ExternalRecord struct {
	ExternalID  string
	URL         string
	Title       string
	PublishedAt time.Time
	RawMetadata []byte
}

type Connector interface {
	Name() string
	SourceType() string
	BaseURL() string
	Fetch(ctx context.Context) ([]ExternalRecord, error)
	Normalize(record ExternalRecord) (*ThreatEvent, error)
}

type IngestionRepository interface {
	GetOrCreateSource(ctx context.Context, name, sourceType, baseURL string) (uuid.UUID, error)
	SaveItemAndEvent(ctx context.Context, sourceID uuid.UUID, item *SourceItem, event *ThreatEvent) error
}
DOMAIN

# 2. HTTP Client
cat << 'HTTP' > services/news-service/internal/ingestion/http_client.go
package ingestion

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type HTTPClient struct {
	client  *http.Client
	retries int
}

func NewHTTPClient(timeout time.Duration, retries int) *HTTPClient {
	return &HTTPClient{
		client:  &http.Client{Timeout: timeout},
		retries: retries,
	}
}

func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for i := 0; i <= c.retries; i++ {
		if i > 0 {
			slog.Info("Retrying request", "url", url, "attempt", i)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(i) * time.Second):
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "GlobePulse-Ingestion/1.0")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
			continue
		}

		return body, nil
	}
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
HTTP

# 3. USGS Connector
cat << 'USGS' > services/news-service/internal/ingestion/connectors/usgs.go
package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/global-news/news-service/internal/ingestion"
	"github.com/google/uuid"
)

type USGSConnector struct {
	client *ingestion.HTTPClient
	apiURL string
}

func NewUSGSConnector() *USGSConnector {
	return &USGSConnector{
		client: ingestion.NewHTTPClient(10*time.Second, 3),
		apiURL: "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson",
	}
}

func (c *USGSConnector) Name() string       { return "USGS Earthquakes" }
func (c *USGSConnector) SourceType() string { return "api" }
func (c *USGSConnector) BaseURL() string    { return "https://earthquake.usgs.gov" }

type usgsFeature struct {
	ID         string `json:"id"`
	Properties struct {
		Title   string  `json:"title"`
		Url     string  `json:"url"`
		Mag     float64 `json:"mag"`
		Time    int64   `json:"time"`
		Place   string  `json:"place"`
	} `json:"properties"`
	Geometry struct {
		Coordinates []float64 `json:"coordinates"` // [lon, lat, depth]
	} `json:"geometry"`
}

func (c *USGSConnector) Fetch(ctx context.Context) ([]domain.ExternalRecord, error) {
	body, err := c.client.Get(ctx, c.apiURL)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Features []usgsFeature `json:"features"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	var records []domain.ExternalRecord
	for _, f := range payload.Features {
		raw, _ := json.Marshal(f)
		records = append(records, domain.ExternalRecord{
			ExternalID:  f.ID,
			URL:         f.Properties.Url,
			Title:       f.Properties.Title,
			PublishedAt: time.UnixMilli(f.Properties.Time),
			RawMetadata: raw,
		})
	}
	return records, nil
}

func (c *USGSConnector) Normalize(record domain.ExternalRecord) (*domain.ThreatEvent, error) {
	var f usgsFeature
	if err := json.Unmarshal(record.RawMetadata, &f); err != nil {
		return nil, err
	}

	var lat, lon float64
	if len(f.Geometry.Coordinates) >= 2 {
		lon = f.Geometry.Coordinates[0]
		lat = f.Geometry.Coordinates[1]
	}

	severity := "low"
	if f.Properties.Mag > 6.0 {
		severity = "critical"
	} else if f.Properties.Mag > 4.5 {
		severity = "high"
	} else if f.Properties.Mag > 3.0 {
		severity = "medium"
	}

	return &domain.ThreatEvent{
		ID:              uuid.New(),
		Title:           f.Properties.Title,
		Description:     fmt.Sprintf("Earthquake of magnitude %v at %s", f.Properties.Mag, f.Properties.Place),
		EventType:       "earthquake",
		Category:        "natural_disaster",
		Severity:        severity,
		Confidence:      100.0,
		OccurredAt:      record.PublishedAt,
		DetectedAt:      time.Now(),
		Latitude:        lat,
		Longitude:       lon,
		LocationDetails: f.Properties.Place,
		Status:          "active",
	}, nil
}
USGS

# 4. GDELT Connector
cat << 'GDELT' > services/news-service/internal/ingestion/connectors/gdelt.go
package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/global-news/news-service/internal/ingestion"
	"github.com/google/uuid"
)

type GDELTConnector struct {
	client *ingestion.HTTPClient
	apiURL string
}

func NewGDELTConnector() *GDELTConnector {
	return &GDELTConnector{
		client: ingestion.NewHTTPClient(15*time.Second, 3),
		apiURL: "https://api.gdeltproject.org/api/v2/geo/geo?query=tone<-5&format=geojson",
	}
}

func (c *GDELTConnector) Name() string       { return "GDELT Project" }
func (c *GDELTConnector) SourceType() string { return "api" }
func (c *GDELTConnector) BaseURL() string    { return "https://gdeltproject.org" }

type gdeltFeature struct {
	Type       string `json:"type"`
	Properties struct {
		Name     string `json:"name"`
		Html     string `json:"html"`
		Url      string `json:"url"`
		Date     string `json:"date"` // Usually omitted in basic geojson or passed in html
	} `json:"properties"`
	Geometry struct {
		Coordinates []float64 `json:"coordinates"` // [lon, lat]
	} `json:"geometry"`
}

func (c *GDELTConnector) Fetch(ctx context.Context) ([]domain.ExternalRecord, error) {
	body, err := c.client.Get(ctx, c.apiURL)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Features []gdeltFeature `json:"features"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	var records []domain.ExternalRecord
	for i, f := range payload.Features {
		raw, _ := json.Marshal(f)
		
		// Fallback for ID since GDELT geojson doesn't always provide a stable unique ID per node
		extID := f.Properties.Url
		if extID == "" {
			extID = fmt.Sprintf("gdelt-geo-%d-%d", time.Now().Unix(), i)
		}

		records = append(records, domain.ExternalRecord{
			ExternalID:  extID,
			URL:         f.Properties.Url,
			Title:       f.Properties.Name,
			PublishedAt: time.Now(), // GDELT v2 live geojson is current 15min window
			RawMetadata: raw,
		})
	}
	return records, nil
}

func (c *GDELTConnector) Normalize(record domain.ExternalRecord) (*domain.ThreatEvent, error) {
	var f gdeltFeature
	if err := json.Unmarshal(record.RawMetadata, &f); err != nil {
		return nil, err
	}

	var lat, lon float64
	if len(f.Geometry.Coordinates) >= 2 {
		lon = f.Geometry.Coordinates[0]
		lat = f.Geometry.Coordinates[1]
	}

	return &domain.ThreatEvent{
		ID:              uuid.New(),
		Title:           f.Properties.Name,
		Description:     "Geopolitical event detected via GDELT media monitoring.",
		EventType:       "geopolitical_event",
		Category:        "news",
		Severity:        "unknown",
		Confidence:      70.0,
		OccurredAt:      record.PublishedAt,
		DetectedAt:      time.Now(),
		Latitude:        lat,
		Longitude:       lon,
		LocationDetails: f.Properties.Name, // Name in GDELT GeoJSON is often the location
		Status:          "active",
	}, nil
}
GDELT

# 5. ReliefWeb Connector
cat << 'RELIEFWEB' > services/news-service/internal/ingestion/connectors/reliefweb.go
package connectors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/global-news/news-service/internal/ingestion"
	"github.com/google/uuid"
)

type ReliefWebConnector struct {
	client *ingestion.HTTPClient
	apiURL string
}

func NewReliefWebConnector() *ReliefWebConnector {
	return &ReliefWebConnector{
		client: ingestion.NewHTTPClient(10*time.Second, 3),
		apiURL: "https://api.reliefweb.int/v1/reports?appname=globepulse&limit=10&preset=latest",
	}
}

func (c *ReliefWebConnector) Name() string       { return "ReliefWeb" }
func (c *ReliefWebConnector) SourceType() string { return "api" }
func (c *ReliefWebConnector) BaseURL() string    { return "https://reliefweb.int" }

type rwReport struct {
	ID     string `json:"id"`
	Fields struct {
		Title string `json:"title"`
		Date  struct {
			Created string `json:"created"`
		} `json:"date"`
		Url string `json:"url"`
	} `json:"fields"`
}

func (c *ReliefWebConnector) Fetch(ctx context.Context) ([]domain.ExternalRecord, error) {
	body, err := c.client.Get(ctx, c.apiURL)
	if err != nil {
		return nil, err
	}

	var payload struct {
		Data []rwReport `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	var records []domain.ExternalRecord
	for _, r := range payload.Data {
		raw, _ := json.Marshal(r)
		
		pubTime, _ := time.Parse(time.RFC3339, r.Fields.Date.Created)
		if pubTime.IsZero() {
			pubTime = time.Now()
		}

		records = append(records, domain.ExternalRecord{
			ExternalID:  r.ID,
			URL:         r.Fields.Url,
			Title:       r.Fields.Title,
			PublishedAt: pubTime,
			RawMetadata: raw,
		})
	}
	return records, nil
}

func (c *ReliefWebConnector) Normalize(record domain.ExternalRecord) (*domain.ThreatEvent, error) {
	var r rwReport
	if err := json.Unmarshal(record.RawMetadata, &r); err != nil {
		return nil, err
	}

	return &domain.ThreatEvent{
		ID:              uuid.New(),
		Title:           r.Fields.Title,
		Description:     "Humanitarian report/update via ReliefWeb.",
		EventType:       "humanitarian_event",
		Category:        "humanitarian",
		Severity:        "medium",
		Confidence:      90.0,
		OccurredAt:      record.PublishedAt,
		DetectedAt:      time.Now(),
		LocationDetails: "See report for specific geography",
		Status:          "active",
	}, nil
}
RELIEFWEB

# 6. Repository
cat << 'REPO' > services/news-service/internal/repository/postgres/ingestion_repo.go
package postgres

import (
	"context"

	"github.com/global-news/news-service/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IngestionRepository struct {
	db *sqlx.DB
}

func NewIngestionRepository(db *sqlx.DB) *IngestionRepository {
	return &IngestionRepository{db: db}
}

func (r *IngestionRepository) GetOrCreateSource(ctx context.Context, name, sourceType, baseURL string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO sources (id, name, source_type, base_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, uuid.New(), name, sourceType, baseURL).Scan(&id)
	return id, err
}

func (r *IngestionRepository) SaveItemAndEvent(ctx context.Context, sourceID uuid.UUID, item *domain.SourceItem, event *domain.ThreatEvent) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Source Item (Ignore if conflict)
	var storedItemID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO source_items (id, source_id, external_id, url, title, raw_metadata, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (source_id, external_id) DO NOTHING
		RETURNING id
	`, item.ID, sourceID, item.ExternalID, item.URL, item.Title, item.RawMetadata, item.PublishedAt).Scan(&storedItemID)

	if err != nil {
		// No row returned means it was a duplicate due to DO NOTHING. We safely ignore.
		return nil
	}

	// 2. Insert Threat Event
	_, err = tx.ExecContext(ctx, `
		INSERT INTO threat_events (id, title, description, event_type, category, severity, confidence, occurred_at, detected_at, latitude, longitude, location_details, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, event.ID, event.Title, event.Description, event.EventType, event.Category, event.Severity, event.Confidence, event.OccurredAt, event.DetectedAt, event.Latitude, event.Longitude, event.LocationDetails, event.Status)
	if err != nil {
		return err
	}

	// 3. Link them
	_, err = tx.ExecContext(ctx, `
		INSERT INTO threat_event_source_items (threat_event_id, source_item_id)
		VALUES ($1, $2)
	`, event.ID, storedItemID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
REPO

# 7. Scheduler
cat << 'SCHED' > services/news-service/internal/ingestion/scheduler.go
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
SCHED

