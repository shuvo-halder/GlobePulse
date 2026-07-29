package domain

import (
	"context"

	types "github.com/global-news/shared-types"
)

type UserRepository interface {
	Create(ctx context.Context, user *types.User) error
	GetByEmail(ctx context.Context, email string) (*types.User, error)
	GetByID(ctx context.Context, id string) (*types.User, error)
	Update(ctx context.Context, user *types.User) error
}
