package service

import (
	"context"
	"fmt"
	"time"

	"github.com/global-news/analytics-service/internal/domain"
	"github.com/global-news/analytics-service/pkg/logger"
	types "github.com/global-news/shared-types"
	"go.uber.org/zap"
)

type analyticsService struct {
	repo     domain.AnalyticsRepository
	cache    domain.CacheRepository
	cacheTTL time.Duration
}

func NewAnalyticsService(repo domain.AnalyticsRepository, cache domain.CacheRepository, ttlMinutes int) domain.AnalyticsUseCase {
	return &analyticsService{
		repo:     repo,
		cache:    cache,
		cacheTTL: time.Duration(ttlMinutes) * time.Minute,
	}
}

func (s *analyticsService) ProcessEvent(ctx context.Context, event *types.AnalyticsEvent) error {
	logger.Log.Debug("Processing event", zap.String("event_id", event.ID))
	return s.repo.SaveEvent(ctx, event)
}

func (s *analyticsService) GetCountryMetrics(ctx context.Context, countryCode string, startDate, endDate time.Time) ([]types.CountryMetrics, error) {
	cacheKey := fmt.Sprintf("metrics:country:%s:%s:%s", countryCode, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	var cached []types.CountryMetrics
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	metrics, err := s.repo.GetCountryMetrics(ctx, countryCode, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(metrics) > 0 {
		_ = s.cache.Set(ctx, cacheKey, metrics, s.cacheTTL)
	}
	return metrics, nil
}

func (s *analyticsService) GetGlobalMetrics(ctx context.Context, date time.Time) ([]types.CountryMetrics, error) {
	cacheKey := fmt.Sprintf("metrics:global:%s", date.Format("2006-01-02"))

	var cached []types.CountryMetrics
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	metrics, err := s.repo.GetGlobalMetrics(ctx, date)
	if err != nil {
		return nil, err
	}

	if len(metrics) > 0 {
		_ = s.cache.Set(ctx, cacheKey, metrics, s.cacheTTL)
	}
	return metrics, nil
}

func (s *analyticsService) GetHeatmapData(ctx context.Context, date time.Time) ([]types.HeatmapData, error) {
	cacheKey := fmt.Sprintf("heatmap:%s", date.Format("2006-01-02"))

	var cached []types.HeatmapData
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	heatmap, err := s.repo.GenerateHeatmap(ctx, date)
	if err != nil {
		return nil, err
	}

	// Handle case where heatmap is empty but no error
	if heatmap == nil {
		heatmap = []types.HeatmapData{}
	}

	_ = s.cache.Set(ctx, cacheKey, heatmap, s.cacheTTL)
	return heatmap, nil
}
