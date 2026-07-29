package service

import (
	"context"
	"fmt"
	"time"

	"github.com/global-news/news-service/internal/domain"
	"github.com/global-news/news-service/pkg/logger"
	types "github.com/global-news/shared-types"
	"go.uber.org/zap"
)

type newsService struct {
	newsRepo domain.NewsRepository
	cache    domain.CacheRepository
	cacheTTL time.Duration
}

func NewNewsService(nr domain.NewsRepository, cache domain.CacheRepository, ttlMinutes int) domain.NewsUseCase {
	return &newsService{
		newsRepo: nr,
		cache:    cache,
		cacheTTL: time.Duration(ttlMinutes) * time.Minute,
	}
}

func (s *newsService) GetNews(ctx context.Context, filter domain.NewsFilter) ([]types.NewsArticle, error) {
	cacheKey := fmt.Sprintf("news:country:%s:limit:%d:offset:%d", filter.CountryCode, filter.Limit, filter.Offset)

	var cachedNews []types.NewsArticle
	if err := s.cache.Get(ctx, cacheKey, &cachedNews); err == nil {
		logger.Log.Debug("Cache hit for news fetch", zap.String("key", cacheKey))
		return cachedNews, nil
	}

	news, err := s.newsRepo.Fetch(ctx, filter)
	if err != nil {
		logger.Log.Error("Failed to fetch news from DB", zap.Error(err))
		return nil, err
	}

	if len(news) > 0 {
		_ = s.cache.Set(ctx, cacheKey, news, s.cacheTTL)
	}

	return news, nil
}

func (s *newsService) GetArticleByID(ctx context.Context, id string) (*types.NewsArticle, error) {
	cacheKey := fmt.Sprintf("news:article:%s", id)

	var cachedArticle types.NewsArticle
	if err := s.cache.Get(ctx, cacheKey, &cachedArticle); err == nil {
		return &cachedArticle, nil
	}

	article, err := s.newsRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, article, s.cacheTTL)
	return article, nil
}

func (s *newsService) CreateArticle(ctx context.Context, article *types.NewsArticle) error {
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()

	err := s.newsRepo.Create(ctx, article)
	if err != nil {
		logger.Log.Error("Failed to create article", zap.Error(err))
		return err
	}
	return nil
}
