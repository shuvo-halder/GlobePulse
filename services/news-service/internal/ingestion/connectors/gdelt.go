package connectors

import (
	"context"
	"encoding/json"
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
		apiURL: "https://api.gdeltproject.org/api/v2/doc/doc?query=(conflict OR crisis OR attack)&mode=artlist&maxrecords=50&format=json",
	}
}

func (c *GDELTConnector) Name() string       { return "GDELT Project" }
func (c *GDELTConnector) SourceType() string { return "api" }
func (c *GDELTConnector) BaseURL() string    { return "https://gdeltproject.org" }

type gdeltArticle struct {
	Url           string `json:"url"`
	UrlMobile     string `json:"url_mobile"`
	Title         string `json:"title"`
	Seendate      string `json:"seendate"` // Format: YYYYMMDDTHHMMSSZ
	Domain        string `json:"domain"`
	Language      string `json:"language"`
	Sourcecountry string `json:"sourcecountry"`
}

type gdeltResponse struct {
	Articles []gdeltArticle `json:"articles"`
}

func (c *GDELTConnector) Fetch(ctx context.Context) ([]domain.ExternalRecord, error) {
	body, err := c.client.Get(ctx, c.apiURL)
	if err != nil {
		return nil, err
	}

	var payload gdeltResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	var records []domain.ExternalRecord
	for _, a := range payload.Articles {
		if a.Url == "" {
			continue // skip invalid records
		}
		
		raw, _ := json.Marshal(a)
		
		extID := a.Url

		publishedAt, err := time.Parse("20060102T150405Z", a.Seendate)
		if err != nil {
			publishedAt = time.Now()
		}

		records = append(records, domain.ExternalRecord{
			ExternalID:  extID,
			URL:         a.Url,
			Title:       a.Title,
			PublishedAt: publishedAt,
			RawMetadata: raw,
		})
	}
	return records, nil
}

func (c *GDELTConnector) Normalize(record domain.ExternalRecord) (*domain.ThreatEvent, error) {
	var a gdeltArticle
	if err := json.Unmarshal(record.RawMetadata, &a); err != nil {
		return nil, err
	}

	return &domain.ThreatEvent{
		ID:               uuid.New(),
		Title:            a.Title,
		Description:      "Event detected via GDELT media monitoring.",
		EventType:        "news_signal",
		Category:         "news",
		Severity:         "unknown",
		Confidence:       50.0,
		OccurredAt:       time.Time{}, // ignored due to EventTimeUnknown = true
		DetectedAt:       record.PublishedAt,
		Latitude:         0, // ignored due to HasNoLocation = true
		Longitude:        0, // ignored due to HasNoLocation = true
		HasNoLocation:    true,
		EventTimeUnknown: true,
		Country:          a.Sourcecountry,
		LocationDetails:  a.Sourcecountry,
		Status:           "active",
	}, nil
}
