package auth

import (
	"context"
	"time"
	"uuid"
)

type RefreshSessionData struct {
	SessionID    uuid.UUID
	NewExpiresAt time.Time
}

type UserInfo struct {
	ID       uuid.UUID
	Username string
	Role     string
}

type SessionWithUser struct {
	Session  *Session
	UserInfo *UserInfo
}

type IAuthRepository interface {
	FindAccountByUserIDAndIssuer(ctx context.Context, userID uuid.UUID, issuer AccountIssuer) (*Account, error)
	CreateAccount(ctx context.Context, data *Account) error
	CreateSession(ctx context.Context, data *Session) error
	FindSessionByTokenHashAndType(ctx context.Context, tokenHash []byte, sessionType SessionType) (*SessionWithUser, error)
	DeleteSessionByTokenHashAndType(ctx context.Context, tokenHash []byte, sessionType SessionType) error
	ReplaceInitialSetupSession(ctx context.Context, data *Session) error
	ExchangeInitialSetupSession(ctx context.Context, initialTokenHash []byte, setupSession *Session) error
	RefreshSession(ctx context.Context, data RefreshSessionData) error
}
