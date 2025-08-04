package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"github.com/jackc/pgx/v5"
)

var _ IScanRuleRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateScanRule(ctx context.Context, rule *common.ScanRule) error {
	query := `INSERT INTO scan_rules (
		application_id, team_id, organization_id,
		sca_scan_enabled, sast_scan_enabled, allow_incremental_scans,
		allow_sast_empty_code, exclude_dir_regexp_queue, forced_do_own_sbom,
		active_blocking_sca
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`

	return r.pool.QueryRow(ctx, query,
		rule.ApplicationID,
		rule.TeamID,
		rule.OrganizationID,
		rule.SCAScanEnabled,
		rule.SASTScanEnabled,
		rule.AllowIncrementalScans,
		rule.AllowSASTEmptyCode,
		rule.ExcludeDirRegexpQueue,
		rule.ForcedDoOwnSBOM,
		rule.ActiveBlockingSCA,
	).Scan(&rule.ID)
}

func (r *PgxRepository) GetScanRuleByID(ctx context.Context, id int) (*common.ScanRule, error) {
	query := `SELECT 
		id, application_id, team_id, organization_id,
		sca_scan_enabled, sast_scan_enabled, allow_incremental_scans,
		allow_sast_empty_code, exclude_dir_regexp_queue, forced_do_own_sbom,
		active_blocking_sca
	FROM scan_rules WHERE id = $1`

	rule := &common.ScanRule{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rule.ID,
		&rule.ApplicationID,
		&rule.TeamID,
		&rule.OrganizationID,
		&rule.SCAScanEnabled,
		&rule.SASTScanEnabled,
		&rule.AllowIncrementalScans,
		&rule.AllowSASTEmptyCode,
		&rule.ExcludeDirRegexpQueue,
		&rule.ForcedDoOwnSBOM,
		&rule.ActiveBlockingSCA,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rule, nil
}

func (r *PgxRepository) UpdateScanRule(ctx context.Context, rule *common.ScanRule) error {
	query := `UPDATE scan_rules SET 
		application_id = $1, 
		team_id = $2, 
		organization_id = $3,
		sca_scan_enabled = $4,
		sast_scan_enabled = $5,
		allow_incremental_scans = $6,
		allow_sast_empty_code = $7,
		exclude_dir_regexp_queue = $8,
		forced_do_own_sbom = $9,
		active_blocking_sca = $10
	WHERE id = $11`

	_, err := r.pool.Exec(ctx, query,
		rule.ApplicationID,
		rule.TeamID,
		rule.OrganizationID,
		rule.SCAScanEnabled,
		rule.SASTScanEnabled,
		rule.AllowIncrementalScans,
		rule.AllowSASTEmptyCode,
		rule.ExcludeDirRegexpQueue,
		rule.ForcedDoOwnSBOM,
		rule.ActiveBlockingSCA,
		rule.ID,
	)
	return err
}

func (r *PgxRepository) DeleteScanRule(ctx context.Context, id int) error {
	query := `DELETE FROM scan_rules WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) ListScanRules(ctx context.Context) ([]*common.ScanRule, error) {
	query := `SELECT 
		id, application_id, team_id, organization_id,
		sca_scan_enabled, sast_scan_enabled, allow_incremental_scans,
		allow_sast_empty_code, exclude_dir_regexp_queue, forced_do_own_sbom,
		active_blocking_sca
	FROM scan_rules`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.ScanRule, error) {
		var rule common.ScanRule
		err := row.Scan(
			&rule.ID,
			&rule.ApplicationID,
			&rule.TeamID,
			&rule.OrganizationID,
			&rule.SCAScanEnabled,
			&rule.SASTScanEnabled,
			&rule.AllowIncrementalScans,
			&rule.AllowSASTEmptyCode,
			&rule.ExcludeDirRegexpQueue,
			&rule.ForcedDoOwnSBOM,
			&rule.ActiveBlockingSCA,
		)
		return &rule, err
	})
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *PgxRepository) GetScanRuleByComposite(ctx context.Context, appID, teamID, orgID int) (*common.ScanRule, error) {
	query := `SELECT 
		id, application_id, team_id, organization_id,
		sca_scan_enabled, sast_scan_enabled, allow_incremental_scans,
		allow_sast_empty_code, exclude_dir_regexp_queue, forced_do_own_sbom,
		active_blocking_sca
	FROM scan_rules 
	WHERE application_id = $1 AND team_id = $2 AND organization_id = $3`

	rule := &common.ScanRule{}
	err := r.pool.QueryRow(ctx, query, appID, teamID, orgID).Scan(
		&rule.ID,
		&rule.ApplicationID,
		&rule.TeamID,
		&rule.OrganizationID,
		&rule.SCAScanEnabled,
		&rule.SASTScanEnabled,
		&rule.AllowIncrementalScans,
		&rule.AllowSASTEmptyCode,
		&rule.ExcludeDirRegexpQueue,
		&rule.ForcedDoOwnSBOM,
		&rule.ActiveBlockingSCA,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rule, nil
}
