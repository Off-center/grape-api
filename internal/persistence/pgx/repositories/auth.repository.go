package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"grape-api/internal/apperror"
	"grape-api/internal/domain/auth"
	"grape-api/pkg"
	"uuid"
)

type PgxAuthRepository struct {
	db *sql.DB
}

var _ auth.IAuthRepository = (*PgxAuthRepository)(nil)

func NewPgxAuthRepository(db *sql.DB) *PgxAuthRepository {
	return &PgxAuthRepository{
		db: db,
	}
}

func (r *PgxAuthRepository) FindAccountByUserIDAndIssuer(ctx context.Context, userID uuid.UUID, issuer auth.AccountIssuer) (*auth.Account, error) {
	var account auth.Account
	var accountID uuid.UUID

	err := r.db.QueryRowContext(ctx, `
        SELECT id, user_id, issuer, password_hash, created_at, updated_at
        FROM accounts
        WHERE user_id = $1 AND issuer = $2
    `, userID, issuer).Scan(&accountID, &account.UserID, &account.Issuer, &account.PasswordHash, &account.CreatedAt, &account.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.New(apperror.CodeNotFound, "account not found")
	}

	if err != nil {
		return nil, fmt.Errorf("find account: %w", err)
	}

	account.Entity = pkg.Restore(accountID)

	return &account, nil
}

func (r *PgxAuthRepository) CreateAccount(ctx context.Context, data *auth.Account) error {
	result, err := r.db.ExecContext(ctx, `
        INSERT INTO accounts (id, user_id, issuer, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, data.ID(), data.UserID, data.Issuer, data.PasswordHash, data.CreatedAt, data.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperror.New(apperror.CodeBadRequest, "account not created")
	}

	return nil
}

func (r *PgxAuthRepository) CreateSession(ctx context.Context, data *auth.Session) error {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (id, session_type, user_id, ip_address, user_agent, token_hash, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, data.ID(), data.Type, data.UserID, data.IPAddress, data.UserAgent, data.TokenHash, data.ExpiresAt, data.CreatedAt, data.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperror.New(apperror.CodeBadRequest, "session not created")
	}

	return nil
}

func (r *PgxAuthRepository) FindSessionByTokenHashAndType(ctx context.Context, tokenHash []byte, sessionType auth.SessionType) (*auth.SessionWithUser, error) {
	var session auth.Session
	var sessionID uuid.UUID
	var sessionUserID *uuid.UUID
	var ipAddress sql.NullString
	var userAgent sql.NullString
	var userID *uuid.UUID
	var username sql.NullString
	var userRole sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT s.id, s.session_type, s.user_id, s.ip_address, s.user_agent, s.token_hash,
		       s.expires_at, s.created_at, s.updated_at,
		       u.id, u.username, u.role::text
		FROM sessions s
		LEFT JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.session_type = $2
	`, tokenHash, sessionType).Scan(
		&sessionID,
		&session.Type,
		&sessionUserID,
		&ipAddress,
		&userAgent,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.UpdatedAt,
		&userID,
		&username,
		&userRole,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.New(apperror.CodeNotFound, "session not found")
	}

	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	session.Entity = pkg.Restore(sessionID)
	session.UserID = sessionUserID

	if ipAddress.Valid {
		session.IPAddress = &ipAddress.String
	}

	if userAgent.Valid {
		session.UserAgent = &userAgent.String
	}

	var userInfo *auth.UserInfo

	if userID != nil {
		userInfo = &auth.UserInfo{
			ID:       *userID,
			Username: username.String,
			Role:     userRole.String,
		}
	}

	return &auth.SessionWithUser{
		Session:  &session,
		UserInfo: userInfo,
	}, nil
}

func (r *PgxAuthRepository) DeleteSessionByTokenHashAndType(ctx context.Context, tokenHash []byte, sessionType auth.SessionType) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE token_hash = $1 AND session_type = $2
	`, tokenHash, sessionType)

	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperror.New(apperror.CodeNotFound, "session not found")
	}

	return nil
}

func (r *PgxAuthRepository) ReplaceInitialSetupSession(ctx context.Context, data *auth.Session) error {
	if data == nil || data.Type != auth.INITIAL_SETUP || data.UserID != nil {
		return apperror.New(apperror.CodeBadRequest, "invalid initial setup session")
	}

	result, err := r.db.ExecContext(ctx, `
		WITH deleted AS (
			DELETE FROM sessions
			WHERE session_type = $1
		)
		INSERT INTO sessions (id, session_type, user_id, ip_address, user_agent, token_hash, expires_at, created_at, updated_at)
		VALUES ($2, $1, $3, $4, $5, $6, $7, $8, $9)
	`, auth.INITIAL_SETUP, data.ID(), data.UserID, data.IPAddress, data.UserAgent, data.TokenHash, data.ExpiresAt, data.CreatedAt, data.UpdatedAt)
	if err != nil {
		return fmt.Errorf("replace initial setup session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return apperror.New(apperror.CodeBadRequest, "initial setup session not created")
	}

	return nil
}

func (r *PgxAuthRepository) ExchangeInitialSetupSession(ctx context.Context, initialTokenHash []byte, setupSession *auth.Session) error {
	if setupSession == nil || setupSession.Type != auth.SETUP || setupSession.UserID != nil {
		return apperror.New(apperror.CodeBadRequest, "invalid setup session")
	}

	result, err := r.db.ExecContext(ctx, `
		WITH consumed AS (
			DELETE FROM sessions
			WHERE token_hash = $1 AND session_type = $2
			RETURNING expires_at
		)
		INSERT INTO sessions (id, session_type, user_id, ip_address, user_agent, token_hash, expires_at, created_at, updated_at)
		SELECT $3, $4, $5, $6, $7, $8, $9, $10, $11
		FROM consumed
		WHERE consumed.expires_at > NOW()
	`,
		initialTokenHash,
		auth.INITIAL_SETUP,
		setupSession.ID(),
		setupSession.Type,
		setupSession.UserID,
		setupSession.IPAddress,
		setupSession.UserAgent,
		setupSession.TokenHash,
		setupSession.ExpiresAt,
		setupSession.CreatedAt,
		setupSession.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("exchange initial setup session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected != 1 {
		return apperror.New(apperror.CodeUnauthorized, "invalid initial setup token")
	}

	return nil
}

func (r *PgxAuthRepository) RefreshSession(ctx context.Context, data auth.RefreshSessionData) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE sessions
		SET expires_at = $1, updated_at = NOW()
		WHERE id = $2 AND session_type = $3
	`, data.NewExpiresAt, data.SessionID, auth.SESSION)

	if err != nil {
		return fmt.Errorf("refresh session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return apperror.New(apperror.CodeNotFound, "session not found")
	}

	return nil
}
