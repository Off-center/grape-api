package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"grape-api/internal/apperror"
	domainauth "grape-api/internal/domain/auth"
	"time"
	"uuid"
)

const (
	SESSION_EXPIRES_IN       = 24 * 7 * time.Hour
	SETUP_EXPIRES_IN         = 15 * time.Minute
	INITIAL_SETUP_EXPIRES_IN = 15 * time.Minute
	UPDATE_AGE               = 24 * time.Hour
)

type SessionService struct {
	authRepository domainauth.IAuthRepository
}

func NewSessionService(authRepository domainauth.IAuthRepository) *SessionService {
	return &SessionService{authRepository: authRepository}
}

func (s *SessionService) GenerateSessionToken() string {
	return rand.Text()
}

func (s *SessionService) HashSessionToken(token string) [32]byte {
	return sha256.Sum256([]byte(token))
}

func (s *SessionService) GenerateExpirationTime(sessionType domainauth.SessionType) (time.Time, error) {
	var expiresIn time.Duration

	switch sessionType {
	case domainauth.SESSION:
		expiresIn = SESSION_EXPIRES_IN
	case domainauth.SETUP:
		expiresIn = SETUP_EXPIRES_IN
	case domainauth.INITIAL_SETUP:
		expiresIn = INITIAL_SETUP_EXPIRES_IN
	default:
		return time.Time{}, apperror.New(apperror.CodeBadRequest, "invalid session type")
	}

	return time.Now().Add(expiresIn), nil
}

func (s *SessionService) VerifyNeedRefresh(session *domainauth.Session) (bool, error) {
	if session == nil {
		return false, apperror.New(apperror.CodeBadRequest, "invalid session")
	}
	if session.Type != domainauth.SESSION {
		return false, apperror.New(apperror.CodeBadRequest, "session cannot be refreshed")
	}

	return time.Since(session.UpdatedAt) >= UPDATE_AGE, nil
}

func (s *SessionService) VerifySession(ctx context.Context, token string, expectedType domainauth.SessionType) (*domainauth.SessionWithUser, error) {
	if token == "" {
		return nil, apperror.New(apperror.CodeUnauthorized, "invalid session")
	}
	if _, err := s.GenerateExpirationTime(expectedType); err != nil {
		return nil, err
	}

	tokenHash := s.HashSessionToken(token)
	result, err := s.authRepository.FindSessionByTokenHashAndType(ctx, tokenHash[:], expectedType)

	if err != nil {
		return nil, apperror.New(apperror.CodeUnauthorized, "invalid session")
	}
	if result == nil || result.Session == nil || result.Session.Type != expectedType {
		return nil, apperror.New(apperror.CodeUnauthorized, "invalid session")
	}

	if !result.Session.ExpiresAt.After(time.Now()) {
		return nil, apperror.New(apperror.CodeUnauthorized, "session expired")
	}
	if expectedType == domainauth.SESSION && (result.Session.UserID == nil || result.UserInfo == nil) {
		return nil, apperror.New(apperror.CodeUnauthorized, "invalid session")
	}
	if expectedType != domainauth.SESSION && result.Session.UserID != nil {
		return nil, apperror.New(apperror.CodeUnauthorized, "invalid session")
	}

	return result, nil
}

func (s *SessionService) RefreshSession(ctx context.Context, sessionID uuid.UUID, sessionType domainauth.SessionType) error {
	if sessionID == uuid.Nil() {
		return apperror.New(apperror.CodeBadRequest, "invalid session")
	}
	if sessionType != domainauth.SESSION {
		return apperror.New(apperror.CodeBadRequest, "session cannot be refreshed")
	}

	expiresAt, err := s.GenerateExpirationTime(domainauth.SESSION)
	if err != nil {
		return err
	}

	return s.authRepository.RefreshSession(ctx, domainauth.RefreshSessionData{
		SessionID:    sessionID,
		NewExpiresAt: expiresAt,
	})
}
