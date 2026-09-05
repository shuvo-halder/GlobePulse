package postgres

import (
	"context"
	"database/sql"

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

func (r *IngestionRepository) SaveItemAndEvent(ctx context.Context, sourceID uuid.UUID, item *domain.SourceItem, event *domain.ThreatEvent) (domain.SaveResult, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
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
		if err == sql.ErrNoRows {
			// No row returned means it was a duplicate due to DO NOTHING. We safely ignore.
			return domain.SaveDuplicate, nil
		}
		return 0, err
	}

	// 2. Insert Threat Event
	var dbLat, dbLon interface{} = event.Latitude, event.Longitude
	if event.HasNoLocation {
		dbLat, dbLon = nil, nil
	}
	
	var dbOccurredAt interface{} = event.OccurredAt
	if event.EventTimeUnknown {
		dbOccurredAt = nil
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO threat_events (id, title, description, event_type, category, severity, confidence, occurred_at, detected_at, latitude, longitude, location_details, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, event.ID, event.Title, event.Description, event.EventType, event.Category, event.Severity, event.Confidence, dbOccurredAt, event.DetectedAt, dbLat, dbLon, event.LocationDetails, event.Status)
	if err != nil {
		return 0, err
	}

	// 3. Link them
	_, err = tx.ExecContext(ctx, `
		INSERT INTO threat_event_source_items (threat_event_id, source_item_id)
		VALUES ($1, $2)
	`, event.ID, storedItemID)
	if err != nil {
		return 0, err
	}

	return domain.SaveInserted, tx.Commit()
}
