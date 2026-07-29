package domain

import (
	"context"
	"time"
)

type AuditLog struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type AuditRepository interface {
	LogAction(ctx context.Context, log *AuditLog) error
	GetUserLogs(ctx context.Context, userID string, limit, offset int) ([]AuditLog, error)
}
