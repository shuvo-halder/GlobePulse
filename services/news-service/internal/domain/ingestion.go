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
	LocationDetails  string
	Status           string
	Metadata         []byte
	HasNoLocation    bool
	EventTimeUnknown bool
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
