package domain

import (
	"context"
	"time"

	types "github.com/global-news/shared-types"
)

type Session struct {
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *Session, expiration time.Duration) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteUserSessions(ctx context.Context, userID string) error
}

type AuthUseCase interface {
	Register(ctx context.Context, email, password, firstName, lastName string) (*types.User, error)
	Login(ctx context.Context, email, password string) (string, string, error)
	Logout(ctx context.Context, sessionID string) error
	VerifyEmail(ctx context.Context, token string) error
	ResetPassword(ctx context.Context, email, newPassword string) error
	GetProfile(ctx context.Context, userID string) (*types.User, error)
}
