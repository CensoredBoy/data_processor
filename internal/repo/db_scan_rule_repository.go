package repo

import (
	"context"
	"data_processor/internal/common"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
		excludeDirs,
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
        forced_do_own_sbom, active_blocking_sca, latest_comment_id
    FROM scan_rules WHERE id = $1`

	rule := &common.ScanRule{}
	var excludeDirs []string
	var latestCommentID sql.NullInt32

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
		&excludeDirs,
		&rule.ForcedDoOwnSBOM,
		&rule.ActiveBlockingSCA,
		&latestCommentID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	rule.ExcludeDirRegexpQueue = excludeDirs
	if latestCommentID.Valid {
		val := int(latestCommentID.Int32)
		rule.LatestCommentID = &val
	}
	return rule, nil
}

func (r *PgxRepository) UpdateScanRule(ctx context.Context, rule *common.ScanRule, userID *int, commentText *string) (*common.Comment, error) {
	updateQuery := `UPDATE scan_rules SET 
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

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var latestCommentID sql.NullInt32
	err = tx.QueryRow(ctx, `SELECT latest_comment_id FROM scan_rules WHERE id = $1 FOR UPDATE`, rule.ID).Scan(&latestCommentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("scan rule not found")
		}
		return nil, err
	}

	_, err = tx.Exec(ctx, updateQuery,
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
		excludeDirs,
		rule.ForcedDoOwnSBOM,
		rule.ActiveBlockingSCA,
		rule.ID,
	)
	if err != nil {
		return nil, err
	}

	if latestCommentID.Valid {
		val := int(latestCommentID.Int32)
		rule.LatestCommentID = &val
	} else {
		rule.LatestCommentID = nil
	}

	var createdComment *common.Comment

	var trimmedComment string
	if commentText != nil {
		trimmedComment = strings.TrimSpace(*commentText)
	}

	if latestCommentID.Valid {
		val := int(latestCommentID.Int32)
		rule.LatestCommentID = &val
	} else {
		rule.LatestCommentID = nil
	}

	if trimmedComment != "" {
		if userID == nil || *userID == 0 {
			return nil, fmt.Errorf("user id is required for comment")
		}

		var prevValue interface{}
		if latestCommentID.Valid {
			prevValue = int(latestCommentID.Int32)
		}

		newComment := &common.Comment{
			ScanRuleID: rule.ID,
			Text:       trimmedComment,
			UserID:     *userID,
		}

		var prev sql.NullInt32
		err = tx.QueryRow(ctx, `INSERT INTO comments (scan_rule_id, previous_comment_id, user_id, comment) 
				VALUES ($1, $2, $3, $4) RETURNING id, previous_comment_id, created_at`,
			newComment.ScanRuleID, prevValue, newComment.UserID, newComment.Text,
		).Scan(&newComment.ID, &prev, &newComment.CreatedAt)
		if err != nil {
			return nil, err
		}

		if prev.Valid {
			val := int(prev.Int32)
			newComment.PreviousCommentID = &val
		}

		_, err = tx.Exec(ctx, `UPDATE scan_rules SET latest_comment_id = $1 WHERE id = $2`, newComment.ID, rule.ID)
		if err != nil {
			return nil, err
		}

		createdComment = newComment
		rule.LatestCommentID = &newComment.ID
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return createdComment, nil
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
        forced_do_own_sbom, active_blocking_sca, latest_comment_id
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
		var latestCommentID sql.NullInt32

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
			&excludeDirs,
			&rule.ForcedDoOwnSBOM,
			&rule.ActiveBlockingSCA,
			&latestCommentID,
		)
		if err != nil {
			return nil, err
		}

		rule.ExcludeDirRegexpQueue = excludeDirs
		if latestCommentID.Valid {
			val := int(latestCommentID.Int32)
			rule.LatestCommentID = &val
		}
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
		var excludeDirs []string

		err := r.pool.QueryRow(ctx, query, args...).Scan(
			&rule.SCAScanEnabled,
			&rule.SASTScanEnabled,
			&rule.ApplicationPostfix,
			&rule.AllowUnsafeExtDistribs,
			&rule.IgnoreRepositoryMembership,
			&rule.AllowIncrementalScans,
			&rule.AllowSASTEmptyCode,
			&excludeDirs,
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

func (r *PgxRepository) GetScanRuleWithComments(ctx context.Context, id int) (*common.ScanRule, error) {
	rule, err := r.GetScanRuleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, `
        SELECT id, scan_rule_id, previous_comment_id, user_id, comment, created_at
        FROM comments
        WHERE scan_rule_id = $1
        ORDER BY created_at ASC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*common.Comment
	for rows.Next() {
		var c common.Comment
		var prev sql.NullInt32
		var created time.Time

		if err := rows.Scan(&c.ID, &c.ScanRuleID, &prev, &c.UserID, &c.Text, &created); err != nil {
			return nil, err
		}
		if prev.Valid {
			val := int(prev.Int32)
			c.PreviousCommentID = &val
		}
		c.CreatedAt = created
		comments = append(comments, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rule.Comments = comments
	return rule, nil
}
