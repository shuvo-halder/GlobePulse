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
		if r.ID == "" {
			continue // skip invalid records
		}

		raw, _ := json.Marshal(r)
		
		var pubTime time.Time
		if parsed, err := time.Parse(time.RFC3339, r.Fields.Date.Created); err == nil {
			pubTime = parsed
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
		EventType:        "humanitarian_report",
		Category:         "humanitarian",
		Severity:         "unknown",
		Confidence:       50.0,
		OccurredAt:       time.Time{}, // Unknown actual event time
		DetectedAt:       record.PublishedAt,
		HasNoLocation:    true,
		EventTimeUnknown: true,
		LocationDetails:  "See report for specific geography",
		Status:           "active",
	}, nil
}
