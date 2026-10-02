package authcontext

import (
	"context"
	"uuid"
)

type contextKey struct{}

type User struct {
	ID       uuid.UUID
	Username string
	Role     string
}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, contextKey{}, user)
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(contextKey{}).(User)

	return user, ok
}
