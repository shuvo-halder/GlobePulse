package service

import (
	"context"
	"fmt"
	"time"

	"github.com/global-news/country-service/internal/domain"
	"github.com/global-news/country-service/pkg/logger"
	types "github.com/global-news/shared-types"
	"go.uber.org/zap"
)

type countryService struct {
	repo     domain.CountryRepository
	cache    domain.CacheRepository
	cacheTTL time.Duration
}

func NewCountryService(cr domain.CountryRepository, cache domain.CacheRepository, ttlMinutes int) domain.CountryUseCase {
	return &countryService{
		repo:     cr,
		cache:    cache,
		cacheTTL: time.Duration(ttlMinutes) * time.Minute,
	}
}

func (s *countryService) GetCountries(ctx context.Context, filter domain.CountryFilter) ([]types.Country, error) {
	cacheKey := fmt.Sprintf("countries:q:%s:r:%s:s:%s:l:%d:o:%d", filter.Query, filter.Region, filter.SortBy, filter.Limit, filter.Offset)

	var cached []types.Country
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		logger.Log.Debug("Cache hit for countries", zap.String("key", cacheKey))
		return cached, nil
	}

	countries, err := s.repo.Fetch(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(countries) > 0 {
		_ = s.cache.Set(ctx, cacheKey, countries, s.cacheTTL)
	}

	return countries, nil
}

func (s *countryService) GetCountryDetails(ctx context.Context, code string) (*types.Country, error) {
	cacheKey := fmt.Sprintf("country:%s", code)

	var cached types.Country
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return &cached, nil
	}

	country, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, country, s.cacheTTL)
	return country, nil
}

func (s *countryService) GetRankings(ctx context.Context, criteria string, limit int) ([]types.Country, error) {
	cacheKey := fmt.Sprintf("rankings:c:%s:l:%d", criteria, limit)

	var cached []types.Country
	if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
		return cached, nil
	}

	countries, err := s.repo.GetRankings(ctx, criteria, limit)
	if err != nil {
		return nil, err
	}

	if len(countries) > 0 {
		_ = s.cache.Set(ctx, cacheKey, countries, s.cacheTTL)
	}

	return countries, nil
}
