package repo

import (
	"context"
	"data_processor/internal/common"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Запускаем контейнер PostgreSQL
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	// Получаем строку подключения
	connStr, err := pgContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Создаем пул подключений
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Применяем миграции
	err = applyMigrations(pool)
	require.NoError(t, err)

	// Возвращаем пул и функцию очистки
	return pool, func() {
		pool.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}
}

func applyMigrations(pool *pgxpool.Pool) error {
	// Здесь должна быть реализация применения миграций
	// В тестах можно использовать простой SQL для создания таблиц
	_, err := pool.Exec(context.Background(), `CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       password VARCHAR(255) NOT NULL
);

CREATE TABLE roles (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(255),
                       description VARCHAR(512),
                       is_active BOOLEAN DEFAULT true,
                       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                       owner_id INTEGER,
                       FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE permissions (
                             id SERIAL PRIMARY KEY,
                             name VARCHAR(255) NOT NULL,
                             description VARCHAR(512),
                             created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                             updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                             read BOOLEAN DEFAULT false,
                             write BOOLEAN DEFAULT false
);

CREATE TABLE organizations (
                               id SERIAL PRIMARY KEY,
                               project_name VARCHAR(255) NOT NULL,
                               owner_id INTEGER NOT NULL,
                               FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE teams (
                       id SERIAL PRIMARY KEY,
                       team_name VARCHAR(255) NOT NULL,
                       owner_id INTEGER NOT NULL,
                       folder VARCHAR(255),
                       organization_id INTEGER NOT NULL,
                       FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
                       FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);


CREATE TABLE users_roles (
                             user_id INTEGER NOT NULL,
                             role_id INTEGER NOT NULL,
                             FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                             FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE roles_permission_organisation (
                                               role_id INTEGER,
                                               organisation_id INTEGER NOT NULL,
                                               permission_id INTEGER UNIQUE NOT NULL,
                                               FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
                                               FOREIGN KEY (organisation_id) REFERENCES organizations(id) ON DELETE CASCADE,
                                               FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);
CREATE TABLE roles_permission_team (
                                       role_id INTEGER,
                                       team_id INTEGER NOT NULL,
                                       permission_id INTEGER UNIQUE NOT NULL,
                                       FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
                                       FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
                                       FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE TABLE applications (
                              id SERIAL PRIMARY KEY,
                              name VARCHAR(255) NOT NULL,
                              description VARCHAR(512),
                              team_id INTEGER NOT NULL,
                              FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL

);

CREATE TABLE versions (
                          id SERIAL PRIMARY KEY,
                          application_id INTEGER NOT NULL,
                          version VARCHAR(50) NOT NULL,
                          FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
);

CREATE TABLE scans (
                       id SERIAL PRIMARY KEY,
                       scan_date DATE NOT NULL DEFAULT CURRENT_DATE,
                       version_id INTEGER NOT NULL,
                       FOREIGN KEY (version_id) REFERENCES versions(id) ON DELETE CASCADE
);

CREATE TABLE scan_info (
                           id SERIAL PRIMARY KEY,
                           scan_id INTEGER NOT NULL,
                           FOREIGN KEY (scan_id) REFERENCES scans(id) ON DELETE CASCADE
);

CREATE TABLE scan_rules (
                            id SERIAL PRIMARY KEY,
                            application_id INTEGER NOT NULL,
                            team_id INTEGER NOT NULL,
                            organization_id INTEGER NOT NULL,
                            sca_scan_enabled BOOLEAN,
                            sast_scan_enabled BOOLEAN,
                            allow_incremental_scans BOOLEAN,
                            allow_sast_empty_code BOOLEAN,
                            exclude_dir_regexp_queue VARCHAR(255) ARRAY,
                            forced_do_own_sbom BOOLEAN,
                            active_blocking_sca BOOLEAN,
                            FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
                            FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
                            FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);`)
	return err
}

func TestRoleRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	// Создаем тестовые данные
	user := &common.User{Name: "test_user", Password: "pass"}
	err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	org := &common.Organization{ProjectName: "test_org", OwnerID: user.ID}
	err = repo.CreateOrganization(ctx, org)
	require.NoError(t, err)

	team := &common.Team{TeamName: "test_team", OwnerID: user.ID, OrganizationID: org.ID}
	err = repo.CreateTeam(ctx, team)
	require.NoError(t, err)

	t.Run("Create and Get Role", func(t *testing.T) {
		desc := "Administrator role"
		active := true
		role := &common.Role{
			Name:        "admin",
			Description: &desc,
			IsActive:    &active,
			OwnerID:     user.ID,
		}

		// Создаем роль
		createdRole, err := repo.CreateRole(ctx, role)
		require.NoError(t, err)
		assert.NotZero(t, createdRole.Role.ID)
		assert.Equal(t, role.Name, createdRole.Role.Name)

		// Получаем роль
		fetchedRole, err := repo.GetRole(ctx, createdRole.Role.ID)
		require.NoError(t, err)
		assert.Equal(t, createdRole.Role.ID, fetchedRole.Role.ID)
		assert.Equal(t, role.Name, fetchedRole.Role.Name)
		assert.Empty(t, fetchedRole.Permissions)
	})

	t.Run("Add and Remove Permissions", func(t *testing.T) {
		// Создаем роль
		role := &common.Role{Name: "manager", OwnerID: user.ID}
		createdRole, err := repo.CreateRole(ctx, role)
		require.NoError(t, err)
		desc := "Organization permission"
		// Создаем permission для организации
		orgPerm := &common.Permission{
			Name:           "org_perm",
			Description:    &desc,
			Read:           true,
			Write:          false,
			OrganizationID: &org.ID,
		}
		err = repo.CreatePermission(ctx, orgPerm)
		require.NoError(t, err)

		// Добавляем permission к роли
		err = repo.AddPermission(ctx, createdRole.Role.ID, orgPerm)
		require.NoError(t, err)

		// Проверяем что permission добавился
		roleWithPerms, err := repo.GetRole(ctx, createdRole.Role.ID)
		require.NoError(t, err)
		require.Len(t, roleWithPerms.Permissions, 1)
		assert.Equal(t, orgPerm.ID, roleWithPerms.Permissions[0].ID)

		// Удаляем permission
		err = repo.RemovePermission(ctx, createdRole.Role.ID, orgPerm.ID)
		require.NoError(t, err)

		// Проверяем что permission удалился
		roleWithPerms, err = repo.GetRole(ctx, createdRole.Role.ID)
		require.NoError(t, err)
		assert.Empty(t, roleWithPerms.Permissions)
	})

	t.Run("Assign and Remove Role from User", func(t *testing.T) {
		// Создаем роль
		role := &common.Role{Name: "developer", OwnerID: user.ID}
		createdRole, err := repo.CreateRole(ctx, role)
		require.NoError(t, err)

		// Создаем второго пользователя
		user2 := &common.User{Name: "test_user2", Password: "pass"}
		err = repo.CreateUser(ctx, user2)
		require.NoError(t, err)

		// Назначаем роль
		err = repo.AssignRoleToUser(ctx, user2.ID, createdRole.Role.ID)
		require.NoError(t, err)

		// Проверяем назначение
		roles, err := repo.GetUserRoles(ctx, user2.ID)
		require.NoError(t, err)
		require.Len(t, roles, 1)
		assert.Equal(t, createdRole.Role.ID, roles[0].ID)

		// Удаляем роль
		err = repo.RemoveRoleFromUser(ctx, user2.ID, createdRole.Role.ID)
		require.NoError(t, err)

		// Проверяем удаление
		roles, err = repo.GetUserRoles(ctx, user2.ID)
		require.NoError(t, err)
		assert.Empty(t, roles)
	})

	t.Run("List Roles by Scope", func(t *testing.T) {
		// Создаем несколько ролей с разными scope
		globalRole := &common.Role{Name: "global_role", OwnerID: user.ID}
		_, err := repo.CreateRole(ctx, globalRole)
		require.NoError(t, err)

		orgRole := &common.Role{Name: "org_role", OwnerID: user.ID}
		orgRoleWithPerms, err := repo.CreateRole(ctx, orgRole)
		require.NoError(t, err)

		// Добавляем org permission
		orgPerm := &common.Permission{
			Name:           "org_perm2",
			OrganizationID: &org.ID,
		}
		err = repo.CreatePermission(ctx, orgPerm)
		require.NoError(t, err)
		err = repo.AddPermission(ctx, orgRoleWithPerms.Role.ID, orgPerm)
		require.NoError(t, err)

		// Тестируем фильтрацию
		tests := []struct {
			name     string
			scope    common.RoleScope
			expected int
		}{
			{"Global scope", common.RoleScope{}, 5},
			{"Org scope", common.RoleScope{OrganizationID: &org.ID}, 1},
			{"Team scope", common.RoleScope{TeamID: &team.ID}, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				roles, err := repo.ListRolesByScope(ctx, tt.scope)
				require.NoError(t, err)
				assert.Len(t, roles, tt.expected)
			})
		}
	})

	t.Run("Update and Delete Role", func(t *testing.T) {
		desc := "Original description"
		active := true
		role := &common.Role{
			Name:        "updatable_role",
			Description: &desc,
			IsActive:    &active,
			OwnerID:     user.ID,
		}

		// Создаем роль
		createdRole, err := repo.CreateRole(ctx, role)
		require.NoError(t, err)

		// Обновляем роль
		newDesc := "Updated description"
		updatedRole, err := repo.UpdateRole(ctx, createdRole.Role.ID, nil, &newDesc, nil)
		require.NoError(t, err)
		assert.Equal(t, newDesc, *updatedRole.Description)

		// Удаляем роль
		err = repo.DeleteRole(ctx, createdRole.Role.ID)
		require.NoError(t, err)

		// Проверяем что роль удалена
		deletedRole, err := repo.GetRole(ctx, createdRole.Role.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedRole)
	})
}

// Дополнительные тесты для edge cases
func TestRoleRepositoryEdgeCases(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	t.Run("Get Non-Existent Role", func(t *testing.T) {
		role, err := repo.GetRole(ctx, 9999)
		require.NoError(t, err)
		assert.Nil(t, role)
	})

	t.Run("Add Permission to Non-Existent Role", func(t *testing.T) {
		perm := &common.Permission{ID: 1} // Несуществующий permission
		err := repo.AddPermission(ctx, 9999, perm)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not exist")
	})

	t.Run("Delete Non-Existent Role", func(t *testing.T) {
		err := repo.DeleteRole(ctx, 9999)
		require.NoError(t, err)
	})
}
