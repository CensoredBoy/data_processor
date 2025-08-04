package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

var _ IPermissionRepository = (*PgxRepository)(nil)

func (r *PgxRepository) GetPermissionByID(ctx context.Context, id int) (*common.Permission, error) {
	query := `
        SELECT 
            p.id, p.name, p.description, p.created_at, p.updated_at, p.read, p.write,
            rpo.organisation_id, rpt.team_id
        FROM permissions p
        LEFT JOIN roles_permission_organisation rpo ON p.id = rpo.permission_id
        LEFT JOIN roles_permission_team rpt ON p.id = rpt.permission_id
        WHERE p.id = $1
    `

	perm := &common.Permission{}
	var orgID, teamID *int

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&perm.ID, &perm.Name, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt,
		&perm.Read, &perm.Write, &orgID, &teamID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Устанавливаем только одно из значений
	if orgID != nil {
		perm.OrganizationID = orgID
	} else if teamID != nil {
		perm.TeamID = teamID
	}

	return perm, nil
}

func (r *PgxRepository) CreatePermission(ctx context.Context, perm *common.Permission) error {
	// Валидируем permission
	if err := perm.Validate(); err != nil {
		return err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Создаем сам permission
	err = tx.QueryRow(ctx, `
        INSERT INTO permissions (name, description, read, write) 
        VALUES ($1, $2, $3, $4) 
        RETURNING id, created_at, updated_at`,
		perm.Name, perm.Description, perm.Read, perm.Write,
	).Scan(&perm.ID, &perm.CreatedAt, &perm.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}

	// 2. Если permission привязан к организации
	if perm.OrganizationID != nil {
		// Проверяем существует ли организация
		var orgExists bool
		err = tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`,
			*perm.OrganizationID,
		).Scan(&orgExists)

		if err != nil {
			return fmt.Errorf("failed to check organization existence: %w", err)
		}
		if !orgExists {
			return fmt.Errorf("organization with ID %d does not exist", *perm.OrganizationID)
		}

		// НЕ создаем связь с организацией здесь, так как permission еще не привязан к роли
		// Связь будет создана при добавлении permission к роли
	} else if perm.TeamID != nil {
		// Проверяем существует ли команда
		var teamExists bool
		err = tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`,
			*perm.TeamID,
		).Scan(&teamExists)

		if err != nil {
			return fmt.Errorf("failed to check team existence: %w", err)
		}
		if !teamExists {
			return fmt.Errorf("team with ID %d does not exist", *perm.TeamID)
		}

		// НЕ создаем связь с командой здесь, так как permission еще не привязан к роли
		// Связь будет создана при добавлении permission к роли
	}

	return tx.Commit(ctx)
}
func (r *PgxRepository) GetPermissionByName(ctx context.Context, name string) (*common.Permission, error) {
	query := `SELECT id, name, description, created_at, updated_at, read, write 
	          FROM permissions WHERE name = $1`

	perm := &common.Permission{}
	err := r.pool.QueryRow(ctx, query, name).
		Scan(&perm.ID, &perm.Name, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt, &perm.Read, &perm.Write)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return perm, nil
}

func (r *PgxRepository) UpdatePermission(ctx context.Context, permission *common.Permission) error {
	query := `UPDATE permissions SET 
		name = $1, 
		description = $2, 
		read = $3, 
		write = $4,
		updated_at = NOW()
		WHERE id = $5`
	_, err := r.pool.Exec(ctx, query, permission.Name, permission.Description,
		permission.Read, permission.Write, permission.ID)
	return err
}

func (r *PgxRepository) DeletePermission(ctx context.Context, id int) error {
	query := `DELETE FROM permissions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *PgxRepository) GetTeamPermissions(ctx context.Context, userID common.UserID, teamID common.TeamID) ([]common.PermissionReadWrite, error) {
	const query = `SELECT p.read, p.write FROM users_roles ur
		JOIN roles_permission_team rpt ON ur.role_id = rpt.role_id
		JOIN permissions p ON rpt.permission_id = p.id
		WHERE ur.user_id = $1 AND rpt.team_id = $2`

	rows, err := r.pool.Query(ctx, query, userID, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []common.PermissionReadWrite
	for rows.Next() {
		var p common.PermissionReadWrite
		if err := rows.Scan(&p.Read, &p.Write); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}

	return perms, nil
}

func (r *PgxRepository) GetOrganizationPermissions(ctx context.Context, userID common.UserID, orgID common.OrgID) ([]common.PermissionReadWrite, error) {
	const query = `SELECT p.read, p.write FROM users_roles ur
		JOIN roles_permission_organisation rpo ON ur.role_id = rpo.role_id
		JOIN permissions p ON rpo.permission_id = p.id
		WHERE ur.user_id = $1 AND rpo.organisation_id = $2`

	rows, err := r.pool.Query(ctx, query, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to query organization permissions: %w", err)
	}
	defer rows.Close()

	var perms []common.PermissionReadWrite
	for rows.Next() {
		var p common.PermissionReadWrite
		if err := rows.Scan(&p.Read, &p.Write); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		perms = append(perms, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return perms, nil

}

func (r *PgxRepository) ListPermissions(ctx context.Context) ([]*common.Permission, error) {
	query := `
        SELECT 
            p.id, p.name, p.description, p.created_at, p.updated_at, p.read, p.write,
            rpo.organisation_id, rpt.team_id
        FROM permissions p
        LEFT JOIN roles_permission_organisation rpo ON p.id = rpo.permission_id
        LEFT JOIN roles_permission_team rpt ON p.id = rpt.permission_id
    `

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	perms, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Permission, error) {
		var perm common.Permission
		var orgID, teamID *int

		err := row.Scan(
			&perm.ID, &perm.Name, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt,
			&perm.Read, &perm.Write, &orgID, &teamID)

		if err != nil {
			return nil, err
		}

		// Устанавливаем только одно из значений
		if orgID != nil {
			perm.OrganizationID = orgID
		} else if teamID != nil {
			perm.TeamID = teamID
		}

		return &perm, nil
	})

	return perms, err
}
