package domain

import (
	"context"
	"time"

	types "github.com/global-news/shared-types"
)

type CountryFilter struct {
	Query  string
	Region string
	Limit  int
	Offset int
	SortBy string
}

type CountryRepository interface {
	Fetch(ctx context.Context, filter CountryFilter) ([]types.Country, error)
	GetByCode(ctx context.Context, code string) (*types.Country, error)
	GetRankings(ctx context.Context, criteria string, limit int) ([]types.Country, error)
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
}

type CountryUseCase interface {
	GetCountries(ctx context.Context, filter CountryFilter) ([]types.Country, error)
	GetCountryDetails(ctx context.Context, code string) (*types.Country, error)
	GetRankings(ctx context.Context, criteria string, limit int) ([]types.Country, error)
}
