package domain

import (
	"context"
	"time"

	types "github.com/global-news/shared-types"
)

type NewsFilter struct {
	CountryCode string
	Topic       string
	Query       string
	Limit       int
	Offset      int
}

type NewsRepository interface {
	Fetch(ctx context.Context, filter NewsFilter) ([]types.NewsArticle, error)
	GetByID(ctx context.Context, id string) (*types.NewsArticle, error)
	Create(ctx context.Context, article *types.NewsArticle) error
}

type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
}

type NewsUseCase interface {
	GetNews(ctx context.Context, filter NewsFilter) ([]types.NewsArticle, error)
	GetArticleByID(ctx context.Context, id string) (*types.NewsArticle, error)
	CreateArticle(ctx context.Context, article *types.NewsArticle) error
}
