package auth

import (
	"errors"
	"grape-api/pkg"
	"time"
	"uuid"
)

type AccountIssuer string

const (
	CREDENTIALS AccountIssuer = "CREDENTIALS"
)

type SessionType string

const (
	SESSION       SessionType = "SESSION"
	SETUP         SessionType = "SETUP"
	INITIAL_SETUP SessionType = "INITIAL_SETUP"
)

var ErrInvalidSession = errors.New("invalid session")

type Session struct {
	pkg.Entity

	Type      SessionType
	UserID    *uuid.UUID
	TokenHash []byte
	IPAddress *string
	UserAgent *string
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Account struct {
	pkg.Entity

	UserID       uuid.UUID
	Issuer       AccountIssuer
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUserSession(userID uuid.UUID, tokenHash []byte, ipAddress string, userAgent string, expiresAt time.Time) (Session, error) {
	if userID == uuid.Nil() || len(tokenHash) == 0 {
		return Session{}, ErrInvalidSession
	}

	return newSession(SESSION, &userID, tokenHash, optionalString(ipAddress), optionalString(userAgent), expiresAt), nil
}

func NewSetupSession(tokenHash []byte, ipAddress string, userAgent string, expiresAt time.Time) (Session, error) {
	if len(tokenHash) == 0 {
		return Session{}, ErrInvalidSession
	}

	return newSession(SETUP, nil, tokenHash, optionalString(ipAddress), optionalString(userAgent), expiresAt), nil
}

func NewInitialSetupSession(tokenHash []byte, expiresAt time.Time) (Session, error) {
	if len(tokenHash) == 0 {
		return Session{}, ErrInvalidSession
	}

	return newSession(INITIAL_SETUP, nil, tokenHash, nil, nil, expiresAt), nil
}

func newSession(sessionType SessionType, userID *uuid.UUID, tokenHash []byte, ipAddress *string, userAgent *string, expiresAt time.Time) Session {
	now := time.Now()

	return Session{
		Entity:    pkg.NewEntity(),
		Type:      sessionType,
		UserID:    userID,
		TokenHash: tokenHash,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func NewAccount(userID uuid.UUID, issuer AccountIssuer, passwordHash string) Account {
	return Account{
		Entity:       pkg.NewEntity(),
		UserID:       userID,
		Issuer:       issuer,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}
