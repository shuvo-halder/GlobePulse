package postgres

import (
	"context"

	"github.com/global-news/auth-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type auditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) domain.AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) LogAction(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, user_id, action, ip_address, user_agent, created_at)
		VALUES (:id, :user_id, :action, :ip_address, :user_agent, :created_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, log)
	return err
}

func (r *auditRepository) GetUserLogs(ctx context.Context, userID string, limit, offset int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	query := `SELECT * FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &logs, query, userID, limit, offset)
	return logs, err
}
