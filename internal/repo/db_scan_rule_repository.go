package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
)

var _ IScanRuleRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateScanRule(ctx context.Context, rule *common.ScanRule) error {
	query := `INSERT INTO scan_rules (
        application_id, team_id, organization_id,
        sca_scan_enabled, sast_scan_enabled, 
        application_postfix, allow_unsafe_ext_distribs,
        ignore_repository_membership, allow_incremental_scans,
        allow_sast_empty_code, exclude_dir_regexp_queue, 
        forced_do_own_sbom, active_blocking_sca
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`

	// Обрабатываем nil-массив как пустой массив
	excludeDirs := rule.ExcludeDirRegexpQueue
	if excludeDirs == nil {
		excludeDirs = []string{}
	}

	return r.pool.QueryRow(ctx, query,
		rule.ApplicationID,
		rule.TeamID,
		rule.OrganizationID,
		rule.SCAScanEnabled,
		rule.SASTScanEnabled,
		rule.ApplicationPostfix,
		rule.AllowUnsafeExtDistribs,
		rule.IgnoreRepositoryMembership,
		rule.AllowIncrementalScans,
		rule.AllowSASTEmptyCode,
		pq.Array(excludeDirs),
		rule.ForcedDoOwnSBOM,
		rule.ActiveBlockingSCA,
	).Scan(&rule.ID)
}

func (r *PgxRepository) GetScanRuleByID(ctx context.Context, id int) (*common.ScanRule, error) {
	query := `SELECT 
        id, application_id, team_id, organization_id,
        sca_scan_enabled, sast_scan_enabled, 
        application_postfix, allow_unsafe_ext_distribs,
        ignore_repository_membership, allow_incremental_scans,
        allow_sast_empty_code, exclude_dir_regexp_queue, 
        forced_do_own_sbom, active_blocking_sca
    FROM scan_rules WHERE id = $1`

	rule := &common.ScanRule{}
	var excludeDirs []string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rule.ID,
		&rule.ApplicationID,
		&rule.TeamID,
		&rule.OrganizationID,
		&rule.SCAScanEnabled,
		&rule.SASTScanEnabled,
		&rule.ApplicationPostfix,
		&rule.AllowUnsafeExtDistribs,
		&rule.IgnoreRepositoryMembership,
		&rule.AllowIncrementalScans,
		&rule.AllowSASTEmptyCode,
		pq.Array(&excludeDirs),
		&rule.ForcedDoOwnSBOM,
		&rule.ActiveBlockingSCA,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	rule.ExcludeDirRegexpQueue = excludeDirs
	return rule, nil
}

func (r *PgxRepository) UpdateScanRule(ctx context.Context, rule *common.ScanRule) error {
	query := `UPDATE scan_rules SET 
        application_id = $1, 
        team_id = $2, 
        organization_id = $3,
        sca_scan_enabled = $4,
        sast_scan_enabled = $5,
        application_postfix = $6,
        allow_unsafe_ext_distribs = $7,
        ignore_repository_membership = $8,
        allow_incremental_scans = $9,
        allow_sast_empty_code = $10,
        exclude_dir_regexp_queue = $11,
        forced_do_own_sbom = $12,
        active_blocking_sca = $13
    WHERE id = $14`

	// Обрабатываем nil-массив как пустой массив
	excludeDirs := rule.ExcludeDirRegexpQueue
	if excludeDirs == nil {
		excludeDirs = []string{}
	}

	_, err := r.pool.Exec(ctx, query,
		rule.ApplicationID,
		rule.TeamID,
		rule.OrganizationID,
		rule.SCAScanEnabled,
		rule.SASTScanEnabled,
		rule.ApplicationPostfix,
		rule.AllowUnsafeExtDistribs,
		rule.IgnoreRepositoryMembership,
		rule.AllowIncrementalScans,
		rule.AllowSASTEmptyCode,
		pq.Array(excludeDirs),
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
        sca_scan_enabled, sast_scan_enabled, 
        application_postfix, allow_unsafe_ext_distribs,
        ignore_repository_membership, allow_incremental_scans,
        allow_sast_empty_code, exclude_dir_regexp_queue, 
        forced_do_own_sbom, active_blocking_sca
    FROM scan_rules`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*common.ScanRule
	for rows.Next() {
		var rule common.ScanRule
		var excludeDirs []string

		err := rows.Scan(
			&rule.ID,
			&rule.ApplicationID,
			&rule.TeamID,
			&rule.OrganizationID,
			&rule.SCAScanEnabled,
			&rule.SASTScanEnabled,
			&rule.ApplicationPostfix,
			&rule.AllowUnsafeExtDistribs,
			&rule.IgnoreRepositoryMembership,
			&rule.AllowIncrementalScans,
			&rule.AllowSASTEmptyCode,
			pq.Array(&excludeDirs),
			&rule.ForcedDoOwnSBOM,
			&rule.ActiveBlockingSCA,
		)
		if err != nil {
			return nil, err
		}

		rule.ExcludeDirRegexpQueue = excludeDirs
		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *PgxRepository) GetScanRuleByComposite(ctx context.Context, appID, teamID, orgID int) (*common.ScanRule, error) {
	result := &common.ScanRule{
		ApplicationID:  &appID,
		TeamID:         &teamID,
		OrganizationID: &orgID,
	}

	// Вспомогательная функция для получения правил
	getRules := func(query string, args ...interface{}) (*common.ScanRule, error) {
		var rule common.ScanRule
		var excludeDirs pgtype.FlatArray[string] // Используем pgtype для массивов

		err := r.pool.QueryRow(ctx, query, args...).Scan(
			&rule.SCAScanEnabled,
			&rule.SASTScanEnabled,
			&rule.ApplicationPostfix,
			&rule.AllowUnsafeExtDistribs,
			&rule.IgnoreRepositoryMembership,
			&rule.AllowIncrementalScans,
			&rule.AllowSASTEmptyCode,
			&excludeDirs, // Используем pgtype напрямую
			&rule.ForcedDoOwnSBOM,
			&rule.ActiveBlockingSCA,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}

		rule.ExcludeDirRegexpQueue = excludeDirs
		return &rule, nil
	}

	// 1. Получаем правило для организации
	orgRule, err := getRules(`
        SELECT sca_scan_enabled, sast_scan_enabled, application_postfix,
               allow_unsafe_ext_distribs, ignore_repository_membership,
               allow_incremental_scans, allow_sast_empty_code,
               exclude_dir_regexp_queue, forced_do_own_sbom, active_blocking_sca
        FROM scan_rules
        WHERE organization_id = $1 AND team_id IS NULL AND application_id IS NULL`,
		orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get org rule: %w", err)
	}

	// 2. Получаем правило для команды
	var teamRule *common.ScanRule
	if teamID != 0 {
		teamRule, err = getRules(`
            SELECT sca_scan_enabled, sast_scan_enabled, application_postfix,
                   allow_unsafe_ext_distribs, ignore_repository_membership,
                   allow_incremental_scans, allow_sast_empty_code,
                   exclude_dir_regexp_queue, forced_do_own_sbom, active_blocking_sca
            FROM scan_rules
            WHERE organization_id = $1 AND team_id = $2 AND application_id IS NULL`,
			orgID, teamID)
		if err != nil {
			return nil, fmt.Errorf("failed to get team rule: %w", err)
		}
	}

	// 3. Получаем правило для приложения
	var appRule *common.ScanRule
	if appID != 0 {
		appRule, err = getRules(`
            SELECT sca_scan_enabled, sast_scan_enabled, application_postfix,
                   allow_unsafe_ext_distribs, ignore_repository_membership,
                   allow_incremental_scans, allow_sast_empty_code,
                   exclude_dir_regexp_queue, forced_do_own_sbom, active_blocking_sca
            FROM scan_rules
            WHERE organization_id = $1 AND team_id = $2 AND application_id = $3`,
			orgID, teamID, appID)
		if err != nil {
			return nil, fmt.Errorf("failed to get app rule: %w", err)
		}
	}

	// Объединяем правила
	mergeRule := func(dest, src *common.ScanRule) {
		if src == nil {
			return
		}
		if src.SCAScanEnabled != nil {
			dest.SCAScanEnabled = src.SCAScanEnabled
		}
		if src.SASTScanEnabled != nil {
			dest.SASTScanEnabled = src.SASTScanEnabled
		}
		if src.ApplicationPostfix != nil {
			dest.ApplicationPostfix = src.ApplicationPostfix
		}
		if src.AllowUnsafeExtDistribs != nil {
			dest.AllowUnsafeExtDistribs = src.AllowUnsafeExtDistribs
		}
		if src.IgnoreRepositoryMembership != nil {
			dest.IgnoreRepositoryMembership = src.IgnoreRepositoryMembership
		}
		if src.AllowIncrementalScans != nil {
			dest.AllowIncrementalScans = src.AllowIncrementalScans
		}
		if src.AllowSASTEmptyCode != nil {
			dest.AllowSASTEmptyCode = src.AllowSASTEmptyCode
		}
		if len(src.ExcludeDirRegexpQueue) > 0 {
			dest.ExcludeDirRegexpQueue = src.ExcludeDirRegexpQueue
		}
		if src.ForcedDoOwnSBOM != nil {
			dest.ForcedDoOwnSBOM = src.ForcedDoOwnSBOM
		}
		if src.ActiveBlockingSCA != nil {
			dest.ActiveBlockingSCA = src.ActiveBlockingSCA
		}
	}

	mergeRule(result, orgRule)
	mergeRule(result, teamRule)
	mergeRule(result, appRule)

	return result, nil
}
