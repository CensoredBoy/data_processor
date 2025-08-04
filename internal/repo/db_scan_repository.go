package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ IScanRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateScan(ctx context.Context, scan *common.Scan) error {
	query := `INSERT INTO scans (scan_date, version_id) VALUES ($1, $2) RETURNING id`
	return r.pool.QueryRow(ctx, query, scan.ScanDate, scan.VersionID).Scan(&scan.ID)
}

func (r *PgxRepository) GetScanByID(ctx context.Context, id int) (*common.Scan, error) {
	query := `SELECT id, scan_date, version_id FROM scans WHERE id = $1`
	scan := &common.Scan{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&scan.ID, &scan.ScanDate, &scan.VersionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return scan, nil
}

func (r *PgxRepository) UpdateScan(ctx context.Context, scan *common.Scan) error {
	query := `UPDATE scans SET scan_date = $1, version_id = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, scan.ScanDate, scan.VersionID, scan.ID)
	return err
}

func (r *PgxRepository) DeleteScan(ctx context.Context, id int) error {
	query := `DELETE FROM scans WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) ListScans(ctx context.Context, versionID int) ([]*common.Scan, error) {
	query := `SELECT id, scan_date, version_id FROM scans WHERE version_id = $1`
	rows, err := r.pool.Query(ctx, query, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scans, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Scan, error) {
		var scan common.Scan
		err := row.Scan(&scan.ID, &scan.ScanDate, &scan.VersionID)
		return &scan, err
	})
	if err != nil {
		return nil, err
	}
	return scans, nil
}
