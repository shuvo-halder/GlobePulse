package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/global-news/auth-service/internal/domain"
	"github.com/go-redis/redis/v8"
)

type sessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(addr, password string) domain.SessionRepository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       1,
	})
	return &sessionRepository{client: rdb}
}

func (r *sessionRepository) sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func (r *sessionRepository) userSessionsKey(userID string) string {
	return fmt.Sprintf("user_sessions:%s", userID)
}

func (r *sessionRepository) CreateSession(ctx context.Context, session *domain.Session, expiration time.Duration) error {
	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, r.sessionKey(session.SessionID), bytes, expiration)
	pipe.SAdd(ctx, r.userSessionsKey(session.UserID), session.SessionID)
	_, err = pipe.Exec(ctx)

	return err
}

func (r *sessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	val, err := r.client.Get(ctx, r.sessionKey(sessionID)).Result()
	if err != nil {
		return nil, err
	}
	var session domain.Session
	err = json.Unmarshal([]byte(val), &session)
	return &session, err
}

func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	session, err := r.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, r.sessionKey(sessionID))
	pipe.SRem(ctx, r.userSessionsKey(session.UserID), sessionID)
	_, err = pipe.Exec(ctx)

	return err
}

func (r *sessionRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	sessions, err := r.client.SMembers(ctx, r.userSessionsKey(userID)).Result()
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		return nil
	}

	pipe := r.client.TxPipeline()
	for _, sessionID := range sessions {
		pipe.Del(ctx, r.sessionKey(sessionID))
	}
	pipe.Del(ctx, r.userSessionsKey(userID))
	_, err = pipe.Exec(ctx)

	return err
}
