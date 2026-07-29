package domain

import (
	"context"
	"time"

	types "github.com/global-news/shared-types"
)

type AnalyticsRepository interface {
	SaveEvent(ctx context.Context, event *types.AnalyticsEvent) error
	GetCountryMetrics(ctx context.Context, countryCode string, startDate, endDate time.Time) ([]types.CountryMetrics, error)
	GetGlobalMetrics(ctx context.Context, date time.Time) ([]types.CountryMetrics, error)
	GenerateHeatmap(ctx context.Context, date time.Time) ([]types.HeatmapData, error)
	AggregateDailyMetrics(ctx context.Context, date time.Time) error
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
}

type AnalyticsUseCase interface {
	ProcessEvent(ctx context.Context, event *types.AnalyticsEvent) error
	GetCountryMetrics(ctx context.Context, countryCode string, startDate, endDate time.Time) ([]types.CountryMetrics, error)
	GetGlobalMetrics(ctx context.Context, date time.Time) ([]types.CountryMetrics, error)
	GetHeatmapData(ctx context.Context, date time.Time) ([]types.HeatmapData, error)
}
