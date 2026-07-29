package postgres

import (
	"context"

	"github.com/global-news/auth-service/internal/domain"
	types "github.com/global-news/shared-types"
	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *types.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, is_verified, created_at, updated_at) 
		VALUES (:id, :email, :password_hash, :first_name, :last_name, :role, :is_verified, :created_at, :updated_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE email = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*types.User, error) {
	var user types.User
	query := `SELECT * FROM users WHERE id = $1 LIMIT 1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *types.User) error {
	query := `
		UPDATE users 
		SET email = :email, password_hash = :password_hash, first_name = :first_name, 
		    last_name = :last_name, role = :role, is_verified = :is_verified, updated_at = :updated_at
		WHERE id = :id
	`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}
