package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/global-news/country-service/internal/domain"
	"github.com/go-redis/redis/v8"
)

type cacheRepository struct {
	client *redis.Client
}

func NewCacheRepository(addr, password string) domain.CacheRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       2,
	})
	return &cacheRepository{client: rdb}
}

func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, bytes, expiration).Err()
}

func (r *cacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}
