package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"grape-api/internal/domain/instance"
	"grape-api/pkg"
	"uuid"
)

type PgxInstanceRepository struct {
	db *sql.DB
}

var _ instance.IInstanceRepository = (*PgxInstanceRepository)(nil)

func NewPgxInstanceRepository(db *sql.DB) *PgxInstanceRepository {
	return &PgxInstanceRepository{db: db}
}

func (r *PgxInstanceRepository) Exists(ctx context.Context) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM instances
		)
	`).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("check instance existence: %w", err)
	}

	return exists, nil
}

func (r *PgxInstanceRepository) CreateIfAbsent(ctx context.Context, data *instance.Instance) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO instances (id, setup_status, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`, data.ID(), data.SetupStatus, data.CreatedAt, data.UpdatedAt)

	if err != nil {
		return fmt.Errorf("create instance if absent: %w", err)
	}

	return nil
}

func (r *PgxInstanceRepository) List(ctx context.Context) ([]*instance.Instance, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, description, setup_status, registration_method,
		       master_user_id, created_at, updated_at
		FROM instances
		ORDER BY created_at, id
	`)

	if err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}

	defer rows.Close()

	instances := make([]*instance.Instance, 0)

	for rows.Next() {
		foundInstance, err := scanInstance(rows)

		if err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}

		instances = append(instances, foundInstance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate instances: %w", err)
	}

	return instances, nil
}

func (r *PgxInstanceRepository) Update(ctx context.Context, data *instance.Instance) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE instances
		SET name = $1,
		    description = $2,
		    setup_status = $3,
		    registration_method = $4,
		    master_user_id = $5,
		    updated_at = $6
		WHERE id = $7
	`, data.Name, data.Description, data.SetupStatus, data.RegistrationMethod,
		data.MasterUserID, data.UpdatedAt, data.ID())

	if err != nil {
		return fmt.Errorf("update instance: %w", err)
	}

	return nil
}

type instanceRowScanner interface {
	Scan(dest ...any) error
}

func scanInstance(row instanceRowScanner) (*instance.Instance, error) {
	var foundInstance instance.Instance
	var id uuid.UUID
	var name sql.NullString
	var description sql.NullString
	var masterUserID *uuid.UUID

	err := row.Scan(
		&id,
		&name,
		&description,
		&foundInstance.SetupStatus,
		&foundInstance.RegistrationMethod,
		&masterUserID,
		&foundInstance.CreatedAt,
		&foundInstance.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	foundInstance.Entity = pkg.Restore(id)

	if name.Valid {
		foundInstance.Name = name.String
	}

	if description.Valid {
		foundInstance.Description = description.String
	}

	if masterUserID != nil {
		foundInstance.MasterUserID = *masterUserID
	}

	return &foundInstance, nil
}
