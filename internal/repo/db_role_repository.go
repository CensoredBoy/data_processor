package repo

import (
	"context"
	"data_processor/internal/common"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
)

var _ IRoleRepository = (*PgxRepository)(nil)

func (r *PgxRepository) CreateRole(ctx context.Context, role *common.Role) (*common.RoleWithPermissions, error) {
	query := `INSERT INTO roles (name, description, is_active, owner_id) 
              VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	roleWithPerms := &common.RoleWithPermissions{
		Role:        role,
		Permissions: []*common.Permission{},
	}

	err := r.pool.QueryRow(ctx, query,
		role.Name, role.Description, role.IsActive, role.OwnerID,
	).Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return roleWithPerms, nil
}
func (r *PgxRepository) UpdateRole(ctx context.Context, roleID int, name, description *string, isActive *bool) (*common.Role, error) {
	query := `
        UPDATE roles SET 
            name = COALESCE($1, name),
            description = COALESCE($2, description),
            is_active = COALESCE($3, is_active),
            updated_at = NOW()
        WHERE id = $4
        RETURNING id, name, description, is_active, created_at, updated_at, owner_id
    `

	role := &common.Role{}
	err := r.pool.QueryRow(ctx, query, name, description, isActive, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &role.OwnerID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return role, nil
}

func (r *PgxRepository) GetRole(ctx context.Context, id int) (*common.RoleWithPermissions, error) {
	// Получаем базовую информацию о роли
	roleQuery := `SELECT id, name, description, is_active, created_at, updated_at, owner_id 
                  FROM roles WHERE id = $1`

	role := &common.Role{}
	err := r.pool.QueryRow(ctx, roleQuery, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &role.OwnerID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	// Получаем permissions для роли через JOIN с таблицами связей
	permQuery := `
        SELECT 
            p.id, p.name, p.description, p.created_at, p.updated_at, 
            p.read, p.write,
            rpo.organisation_id as org_id,
            rpt.team_id as team_id
        FROM permissions p
        LEFT JOIN roles_permission_organisation rpo ON p.id = rpo.permission_id AND rpo.role_id = $1
        LEFT JOIN roles_permission_team rpt ON p.id = rpt.permission_id AND rpt.role_id = $1
        WHERE rpo.role_id IS NOT NULL OR rpt.role_id IS NOT NULL
    `

	rows, err := r.pool.Query(ctx, permQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*common.Permission
	for rows.Next() {
		var perm common.Permission
		var orgID, teamID *int

		err := rows.Scan(
			&perm.ID, &perm.Name, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt,
			&perm.Read, &perm.Write, &orgID, &teamID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}

		// Устанавливаем scope на основе полученных данных
		if orgID != nil {
			perm.OrganizationID = orgID
		} else if teamID != nil {
			perm.TeamID = teamID
		}

		permissions = append(permissions, &perm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return &common.RoleWithPermissions{
		Role:        role,
		Permissions: permissions,
	}, nil
}
func (r *PgxRepository) DeleteRole(ctx context.Context, roleID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Удаляем связи с permissions
	_, err = tx.Exec(ctx, `DELETE FROM roles_permission_organisation WHERE role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete organization permissions: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM roles_permission_team WHERE role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete team permissions: %w", err)
	}

	// Удаляем связи с пользователями
	_, err = tx.Exec(ctx, `DELETE FROM users_roles WHERE role_id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete user role assignments: %w", err)
	}

	// Удаляем саму роль
	_, err = tx.Exec(ctx, `DELETE FROM roles WHERE id = $1`, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) GetRoleByName(ctx context.Context, name string) (*common.Role, error) {
	query := `SELECT id, name, description, is_active, created_at, updated_at, owner_id 
              FROM roles WHERE name = $1`

	role := &common.Role{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.IsActive,
		&role.CreatedAt, &role.UpdatedAt, &role.OwnerID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	return role, nil
}

func (r *PgxRepository) AddPermission(ctx context.Context, roleID int, permission *common.Permission) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Проверяем существование permission
	var exists bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM permissions WHERE id = $1)`, permission.ID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("permission with ID %d does not exist", permission.ID)
	}

	// Добавляем связь в зависимости от типа permission
	if permission.OrganizationID != nil {
		// Проверяем существование организации
		var orgExists bool
		err = tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM organizations WHERE id = $1)`,
			*permission.OrganizationID,
		).Scan(&orgExists)

		if err != nil {
			return fmt.Errorf("failed to check organization existence: %w", err)
		}
		if !orgExists {
			return fmt.Errorf("organization with ID %d does not exist", *permission.OrganizationID)
		}

		_, err = tx.Exec(ctx, `
            INSERT INTO roles_permission_organisation (role_id, organisation_id, permission_id)
            VALUES ($1, $2, $3)
            ON CONFLICT DO NOTHING`,
			roleID, *permission.OrganizationID, permission.ID,
		)
	} else if permission.TeamID != nil {
		// Проверяем существование команды
		var teamExists bool
		err = tx.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`,
			*permission.TeamID,
		).Scan(&teamExists)

		if err != nil {
			return fmt.Errorf("failed to check team existence: %w", err)
		}
		if !teamExists {
			return fmt.Errorf("team with ID %d does not exist", *permission.TeamID)
		}

		_, err = tx.Exec(ctx, `
            INSERT INTO roles_permission_team (role_id, team_id, permission_id)
            VALUES ($1, $2, $3)
            ON CONFLICT DO NOTHING`,
			roleID, *permission.TeamID, permission.ID,
		)
	} else {
		// Глобальный permission
		_, err = tx.Exec(ctx, `
            INSERT INTO roles_permission_organisation (role_id, permission_id)
            VALUES ($1, $2)
            ON CONFLICT DO NOTHING`,
			roleID, permission.ID,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to add permission: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) RemovePermission(ctx context.Context, roleID int, permissionID int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Удаляем из обеих таблиц (на случай если permission был в обеих)
	_, err = tx.Exec(ctx, `DELETE FROM roles_permission_organisation WHERE role_id = $1 AND permission_id = $2`,
		roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove organization permission: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM roles_permission_team WHERE role_id = $1 AND permission_id = $2`,
		roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove team permission: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) ListRolesByScope(ctx context.Context, scope common.RoleScope) ([]*common.RoleWithPermissions, error) {
	var query string
	var args []interface{}

	// Формируем запрос в зависимости от scope
	if scope.OrganizationID != nil {
		query = `
            SELECT DISTINCT r.id, r.name, r.description, r.is_active, 
                   r.created_at, r.updated_at, r.owner_id
            FROM roles r
            JOIN roles_permission_organisation rpo ON r.id = rpo.role_id
            WHERE rpo.organisation_id = $1
        `
		args = []interface{}{*scope.OrganizationID}
	} else if scope.TeamID != nil {
		query = `
            SELECT DISTINCT r.id, r.name, r.description, r.is_active, 
                   r.created_at, r.updated_at, r.owner_id
            FROM roles r
            JOIN roles_permission_team rpt ON r.id = rpt.role_id
            WHERE rpt.team_id = $1
        `
		args = []interface{}{*scope.TeamID}
	} else {
		query = `SELECT id, name, description, is_active, created_at, updated_at, owner_id FROM roles`
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Role, error) {
		var role common.Role
		err := row.Scan(
			&role.ID, &role.Name, &role.Description, &role.IsActive,
			&role.CreatedAt, &role.UpdatedAt, &role.OwnerID,
		)
		return &role, err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan roles: %w", err)
	}

	// Получаем полную информацию для каждой роли
	var result []*common.RoleWithPermissions
	for _, role := range roles {
		fullRole, err := r.GetRole(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get role details: %w", err)
		}
		result = append(result, fullRole)
	}

	return result, nil
}

func (r *PgxRepository) AssignRoleToUser(ctx context.Context, userID common.UserID, roleID int) error {
	// Проверяем существование пользователя и роли
	var userExists, roleExists bool

	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&userExists)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}
	if !userExists {
		return fmt.Errorf("user with ID %d does not exist", userID)
	}

	err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`, roleID).Scan(&roleExists)
	if err != nil {
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	if !roleExists {
		return fmt.Errorf("role with ID %d does not exist", roleID)
	}

	_, err = r.pool.Exec(ctx, `
        INSERT INTO users_roles (user_id, role_id) 
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING`,
		userID, roleID,
	)

	if err != nil {
		return fmt.Errorf("failed to assign role to user: %w", err)
	}

	return nil
}

func (r *PgxRepository) RemoveRoleFromUser(ctx context.Context, userID common.UserID, roleID int) error {
	_, err := r.pool.Exec(ctx, `
        DELETE FROM users_roles 
        WHERE user_id = $1 AND role_id = $2`,
		userID, roleID,
	)

	if err != nil {
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	return nil
}

func (r *PgxRepository) GetUserRoles(ctx context.Context, userID common.UserID) ([]*common.Role, error) {
	query := `
        SELECT r.id, r.name, r.description, r.is_active, r.created_at, r.updated_at, r.owner_id
        FROM roles r
        JOIN users_roles ur ON r.id = ur.role_id
        WHERE ur.user_id = $1
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	roles, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Role, error) {
		var role common.Role
		err := row.Scan(
			&role.ID, &role.Name, &role.Description, &role.IsActive,
			&role.CreatedAt, &role.UpdatedAt, &role.OwnerID,
		)
		return &role, err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan user roles: %w", err)
	}

	return roles, nil
}

func (r *PgxRepository) ListRoles(ctx context.Context) ([]*common.Role, error) {
	query := `SELECT id, name, description, is_active, created_at, updated_at, owner_id FROM roles`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*common.Role, error) {
		var role common.Role
		err := row.Scan(
			&role.ID, &role.Name, &role.Description, &role.IsActive,
			&role.CreatedAt, &role.UpdatedAt, &role.OwnerID,
		)
		return &role, err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan roles: %w", err)
	}

	return roles, nil
}
