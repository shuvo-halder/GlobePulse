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
		Title   string   `json:"title"`
		Url     string   `json:"url"`
		Mag     *float64 `json:"mag"`
		Time    int64    `json:"time"`
		Updated int64    `json:"updated"`
		Place   string   `json:"place"`
		Alert   string   `json:"alert"`
		Type    string   `json:"type"`
		Status  string   `json:"status"`
	} `json:"properties"`
	Geometry struct {
		Type        string    `json:"type"`
		Coordinates []float64 `json:"coordinates"` // [lon, lat, depth]
	} `json:"geometry"`
}

type usgsMetadata struct {
	Magnitude *float64 `json:"magnitude"`
	Depth     *float64 `json:"depth,omitempty"`
	Alert     string   `json:"alert,omitempty"`
	Updated   int64    `json:"updated,omitempty"`
	Type      string   `json:"type,omitempty"`
	Status    string   `json:"status,omitempty"`
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
		return nil, fmt.Errorf("failed to parse USGS geojson: %w", err)
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

	if len(f.Geometry.Coordinates) < 2 {
		return nil, fmt.Errorf("missing coordinates for USGS event %s", f.ID)
	}

	lon := f.Geometry.Coordinates[0]
	lat := f.Geometry.Coordinates[1]

	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, fmt.Errorf("invalid coordinates for USGS event %s: lat=%v, lon=%v", f.ID, lat, lon)
	}

	var depth *float64
	if len(f.Geometry.Coordinates) >= 3 {
		d := f.Geometry.Coordinates[2]
		depth = &d
	}

	// Prepare metadata for source-specific attributes like magnitude and depth
	meta := usgsMetadata{
		Magnitude: f.Properties.Mag,
		Depth:     depth,
		Alert:     f.Properties.Alert,
		Updated:   f.Properties.Updated,
		Type:      f.Properties.Type,
		Status:    f.Properties.Status,
	}
	metaBytes, _ := json.Marshal(meta)

	// Severity determination - keeping it deterministic as a preliminary rule
	severity := "unknown"
	if f.Properties.Mag != nil {
		mag := *f.Properties.Mag
		if mag > 6.0 {
			severity = "critical"
		} else if mag > 4.5 {
			severity = "high"
		} else if mag > 3.0 {
			severity = "medium"
		} else {
			severity = "low"
		}
	}

	// Determine event type
	eventType := "earthquake"
	if f.Properties.Type != "" {
		eventType = f.Properties.Type
	}

	// DetectedAt (when USGS updated/detected it) vs OccurredAt (actual event time)
	detectedAt := time.Now()
	if f.Properties.Updated > 0 {
		detectedAt = time.UnixMilli(f.Properties.Updated)
	}

	return &domain.ThreatEvent{
		ID:              uuid.New(),
		Title:           f.Properties.Title,
		Description:     f.Properties.Title, // Using title which usually contains "M X.X - Place"
		EventType:       eventType,
		Category:        "natural_disaster",
		Severity:        severity,
		Confidence:      100.0,
		OccurredAt:      record.PublishedAt,
		DetectedAt:      detectedAt,
		Latitude:        lat,
		Longitude:       lon,
		LocationDetails: f.Properties.Place,
		Status:          f.Properties.Status,
		Metadata:        metaBytes,
	}, nil
}
