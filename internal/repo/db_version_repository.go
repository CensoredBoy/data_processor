package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ IVersionRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateVersion(ctx context.Context, version *common.Version) error {
	query := `INSERT INTO versions (application_id, version) VALUES ($1, $2) RETURNING id`
	return r.pool.QueryRow(ctx, query, version.ApplicationID, version.Version).Scan(&version.ID)
}

func (r *PgxRepository) GetVersionByID(ctx context.Context, id int) (*common.Version, error) {
	query := `SELECT id, application_id, version FROM versions WHERE id = $1`
	version := &common.Version{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&version.ID, &version.ApplicationID, &version.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return version, nil
}

func (r *PgxRepository) GetVersionByNumber(ctx context.Context, appID int, version string) (*common.Version, error) {
	query := `SELECT id, application_id, version FROM versions WHERE application_id = $1 AND version = $2`
	ver := &common.Version{}
	err := r.pool.QueryRow(ctx, query, appID, version).Scan(&ver.ID, &ver.ApplicationID, &ver.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return ver, nil
}

func (r *PgxRepository) UpdateVersion(ctx context.Context, version *common.Version) error {
	query := `UPDATE versions SET application_id = $1, version = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, version.ApplicationID, version.Version, version.ID)
	return err
}

func (r *PgxRepository) DeleteVersion(ctx context.Context, id int) error {
	query := `DELETE FROM versions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) ListVersions(ctx context.Context, appID int) ([]*common.Version, error) {
	query := `SELECT id, application_id, version FROM versions WHERE application_id = $1`
	rows, err := r.pool.Query(ctx, query, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Version, error) {
		var version common.Version
		err := row.Scan(&version.ID, &version.ApplicationID, &version.Version)
		return &version, err
	})
	if err != nil {
		return nil, err
	}
	return versions, nil
}
