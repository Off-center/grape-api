package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"grape-api/internal/apperror"
	"grape-api/internal/domain/user"
	"grape-api/pkg"
	"uuid"
)

type PgxUserRepository struct {
	db *sql.DB
}

var _ user.IUserRepository = (*PgxUserRepository)(nil)

func NewPgxUserRepository(db *sql.DB) *PgxUserRepository {
	return &PgxUserRepository{db: db}
}

func (r *PgxUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	foundUser, err := scanUser(r.db.QueryRowContext(ctx, `
		SELECT id, name, username, image_url, role, access_status,
		       created_at, updated_at
		FROM users
		WHERE id = $1
	`, id))

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.New(apperror.CodeNotFound, "user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return foundUser, nil
}

func (r *PgxUserRepository) FindByUsername(ctx context.Context, username string) (*user.User, error) {
	foundUser, err := scanUser(r.db.QueryRowContext(ctx, `
		SELECT id, name, username, image_url, role, access_status,
		        created_at, updated_at
		FROM users
		WHERE username = $1
	`, username))

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.New(apperror.CodeNotFound, "user not found")
	}

	if err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return foundUser, nil
}

func (r *PgxUserRepository) Create(ctx context.Context, data user.User) (*user.User, error) {
	createdUser, err := scanUser(r.db.QueryRowContext(ctx, `
		INSERT INTO users (
			id, name, username, image_url, role, access_status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, name, username, image_url, role, access_status,
		        created_at, updated_at
	`,
		data.ID(),
		data.Name,
		data.Username,
		data.ImageURL,
		data.Role,
		data.AccessStatus,
		data.CreatedAt,
		data.UpdatedAt,
	))

	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.New(apperror.CodeBadRequest, "user not created")
	}

	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return createdUser, nil
}

func (r *PgxUserRepository) ListByRole(ctx context.Context, role user.USER_ROLE) ([]*user.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, username, image_url, role, access_status,
		       created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at, id
	`, role)

	if err != nil {
		return nil, fmt.Errorf("list users by role: %w", err)
	}

	defer rows.Close()

	users := make([]*user.User, 0)

	for rows.Next() {
		foundUser, err := scanUser(rows)

		if err != nil {
			return nil, fmt.Errorf("scan user by role: %w", err)
		}

		users = append(users, foundUser)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users by role: %w", err)
	}

	return users, nil
}

type userRowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row userRowScanner) (*user.User, error) {
	var foundUser user.User
	var id uuid.UUID
	var imageURL sql.NullString

	err := row.Scan(
		&id,
		&foundUser.Name,
		&foundUser.Username,
		&imageURL,
		&foundUser.Role,
		&foundUser.AccessStatus,
		&foundUser.CreatedAt,
		&foundUser.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	foundUser.Entity = pkg.Restore(id)

	if imageURL.Valid {
		foundUser.ImageURL = imageURL.String
	}

	return &foundUser, nil
}
