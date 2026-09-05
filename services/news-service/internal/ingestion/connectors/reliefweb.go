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
