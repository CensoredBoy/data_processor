package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ IScanInfoRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateScanInfo(ctx context.Context, scanInfo *common.ScanInfo) error {
	query := `INSERT INTO scan_info (scan_id) VALUES ($1) RETURNING id`
	return r.pool.QueryRow(ctx, query, scanInfo.ScanID).Scan(&scanInfo.ID)
}

func (r *PgxRepository) GetScanInfoByID(ctx context.Context, id int) (*common.ScanInfo, error) {
	query := `SELECT id, scan_id FROM scan_info WHERE id = $1`
	scanInfo := &common.ScanInfo{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&scanInfo.ID, &scanInfo.ScanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return scanInfo, nil
}

func (r *PgxRepository) GetScanInfoByScanID(ctx context.Context, scanID int) (*common.ScanInfo, error) {
	query := `SELECT id, scan_id FROM scan_info WHERE scan_id = $1`
	scanInfo := &common.ScanInfo{}
	err := r.pool.QueryRow(ctx, query, scanID).Scan(&scanInfo.ID, &scanInfo.ScanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return scanInfo, nil
}

func (r *PgxRepository) UpdateScanInfo(ctx context.Context, scanInfo *common.ScanInfo) error {
	query := `UPDATE scan_info SET scan_id = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, scanInfo.ScanID, scanInfo.ID)
	return err
}

func (r *PgxRepository) DeleteScanInfo(ctx context.Context, id int) error {
	query := `DELETE FROM scan_info WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
