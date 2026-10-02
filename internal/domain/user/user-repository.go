package user

import (
	"context"
	"uuid"
)

type IUserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, data User) (*User, error)
	ListByRole(ctx context.Context, role USER_ROLE) ([]*User, error)
}
