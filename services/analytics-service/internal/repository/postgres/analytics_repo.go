package postgres

import (
	"context"
	"time"

	types "github.com/global-news/shared-types"

	"github.com/global-news/analytics-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type analyticsRepository struct {
	db *sqlx.DB
}

func NewAnalyticsRepository(db *sqlx.DB) domain.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) SaveEvent(ctx context.Context, event *types.AnalyticsEvent) error {
	query := `
		INSERT INTO analytics_events (id, event_type, payload, country_code, sentiment, timestamp)
		VALUES (:id, :event_type, :payload, :country_code, :sentiment, :timestamp)
	`
	_, err := r.db.NamedExecContext(ctx, query, event)
	return err
}

func (r *analyticsRepository) GetCountryMetrics(ctx context.Context, countryCode string, startDate, endDate time.Time) ([]types.CountryMetrics, error) {
	var metrics []types.CountryMetrics
	query := `
		SELECT country_code, date, total_events, avg_sentiment, trending_score 
		FROM country_metrics 
		WHERE country_code = $1 AND date >= $2 AND date <= $3
		ORDER BY date ASC
	`
	err := r.db.SelectContext(ctx, &metrics, query, countryCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	return metrics, err
}

func (r *analyticsRepository) GetGlobalMetrics(ctx context.Context, date time.Time) ([]types.CountryMetrics, error) {
	var metrics []types.CountryMetrics
	query := `
		SELECT country_code, date, total_events, avg_sentiment, trending_score 
		FROM country_metrics 
		WHERE date = $1
		ORDER BY total_events DESC
	`
	err := r.db.SelectContext(ctx, &metrics, query, date.Format("2006-01-02"))
	return metrics, err
}

func (r *analyticsRepository) GenerateHeatmap(ctx context.Context, date time.Time) ([]types.HeatmapData, error) {
	// A placeholder for heatmap generation, ideally joining with a countries table for lat/lng
	// Returning empty here, in reality, it would execute a complex spatial query or join.
	var heatmap []types.HeatmapData
	query := `
		SELECT c.lat, c.lng, m.total_events as intensity
		FROM country_metrics m
		JOIN countries c ON m.country_code = c.code
		WHERE m.date = $1
	`
	err := r.db.SelectContext(ctx, &heatmap, query, date.Format("2006-01-02"))
	// Ignore err if countries table is in a different service database, we might need to handle it differently
	// Assuming it's in the same DB or we have lat/lng cached.
	return heatmap, err
}

func (r *analyticsRepository) AggregateDailyMetrics(ctx context.Context, date time.Time) error {
	query := `
		INSERT INTO country_metrics (country_code, date, total_events, avg_sentiment, trending_score)
		SELECT 
			country_code,
			DATE(timestamp) as date,
			COUNT(*) as total_events,
			AVG(sentiment) as avg_sentiment,
			COUNT(*) * ABS(AVG(sentiment)) as trending_score
		FROM analytics_events
		WHERE DATE(timestamp) = $1
		GROUP BY country_code, DATE(timestamp)
		ON CONFLICT (country_code, date) DO UPDATE SET
			total_events = EXCLUDED.total_events,
			avg_sentiment = EXCLUDED.avg_sentiment,
			trending_score = EXCLUDED.trending_score
	`
	_, err := r.db.ExecContext(ctx, query, date.Format("2006-01-02"))
	return err
}
