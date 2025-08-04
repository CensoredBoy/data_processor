package repo

import (
	"context"
	"data_processor/internal/common"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, repo *PgxRepository) *common.User {
	user := &common.User{
		Name:     "test_user",
		Password: "password",
	}
	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err)
	return user
}

func createTestOrg(t *testing.T, repo *PgxRepository, ownerID common.UserID) *common.Organization {
	org := &common.Organization{
		ProjectName: "test_org",
		OwnerID:     ownerID,
	}
	err := repo.CreateOrganization(context.Background(), org)
	require.NoError(t, err)
	return org
}

func createTestTeam(t *testing.T, repo *PgxRepository, ownerID common.UserID, orgID int) *common.Team {
	f := "/test"
	team := &common.Team{
		TeamName:       "test_team",
		OwnerID:        ownerID,
		Folder:         &f,
		OrganizationID: orgID,
	}
	err := repo.CreateTeam(context.Background(), team)
	require.NoError(t, err)
	return team
}

func TestUserRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	t.Run("Create and Get User", func(t *testing.T) {
		user := &common.User{
			Name:     "test_user",
			Password: "secure_password",
		}

		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		id, err := repo.GetUserID(ctx, user)
		require.NoError(t, err)
		assert.Equal(t, common.UserID(1), *id)
		fetchedUser, err := repo.GetUserByID(ctx, *id)
		require.NoError(t, err)
		assert.Equal(t, *id, fetchedUser.ID)
		assert.Equal(t, user.Name, fetchedUser.Name)

	})

	t.Run("Get User By Name", func(t *testing.T) {
		user := &common.User{
			Name:     "unique_user",
			Password: "pass123",
		}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		fetchedUser, err := repo.GetUserByName(ctx, user.Name)
		require.NoError(t, err)
		assert.Equal(t, user.ID, fetchedUser.ID)
	})

	t.Run("Update and Delete User", func(t *testing.T) {
		user := &common.User{
			Name:     "to_update",
			Password: "old_pass",
		}
		err := repo.CreateUser(ctx, user)
		require.NoError(t, err)

		user.Password = "new_pass"
		err = repo.UpdateUser(ctx, user)
		require.NoError(t, err)

		updatedUser, err := repo.GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "new_pass", updatedUser.Password)

		err = repo.DeleteUser(ctx, user.ID)
		require.NoError(t, err)

		deletedUser, err := repo.GetUserByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedUser)
	})

	t.Run("List Users", func(t *testing.T) {
		// Создаем несколько пользователей
		users := []*common.User{
			{Name: "user1", Password: "pass1"},
			{Name: "user2", Password: "pass2"},
		}

		for _, u := range users {
			err := repo.CreateUser(ctx, u)
			require.NoError(t, err)
		}

		userList, err := repo.ListUsers(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(userList), 2)
	})
}

func TestPermissionRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	user := createTestUser(t, repo)
	org := createTestOrg(t, repo, user.ID)
	_ = createTestTeam(t, repo, user.ID, org.ID)

	t.Run("Create and Get Permission", func(t *testing.T) {
		desc := "Test permission"
		perm := &common.Permission{
			Name:        "read",
			Description: &desc,
			Read:        true,
			Write:       false,
		}

		err := repo.CreatePermission(ctx, perm)
		require.NoError(t, err)
		assert.NotZero(t, perm.ID)

		fetchedPerm, err := repo.GetPermissionByID(ctx, perm.ID)
		require.NoError(t, err)
		assert.Equal(t, perm.ID, fetchedPerm.ID)
		assert.Equal(t, perm.Name, fetchedPerm.Name)
	})

	t.Run("Update and Delete Permission", func(t *testing.T) {
		perm := &common.Permission{
			Name:  "to_update",
			Read:  true,
			Write: false,
		}
		err := repo.CreatePermission(ctx, perm)
		require.NoError(t, err)

		perm.Write = true
		err = repo.UpdatePermission(ctx, perm)
		require.NoError(t, err)

		updatedPerm, err := repo.GetPermissionByID(ctx, perm.ID)
		require.NoError(t, err)
		assert.True(t, updatedPerm.Write)

		err = repo.DeletePermission(ctx, perm.ID)
		require.NoError(t, err)

		deletedPerm, err := repo.GetPermissionByID(ctx, perm.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedPerm)
	})

	t.Run("List Permissions", func(t *testing.T) {
		// Создаем несколько permissions
		perms := []*common.Permission{
			{Name: "perm1", Read: true},
			{Name: "perm2", Write: true},
		}

		for _, p := range perms {
			err := repo.CreatePermission(ctx, p)
			require.NoError(t, err)
		}

		permList, err := repo.ListPermissions(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(permList), 2)
	})
}

func TestOrganizationRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	user := createTestUser(t, repo)

	t.Run("Create and Get Organization", func(t *testing.T) {
		org := &common.Organization{
			ProjectName: "test_org",
			OwnerID:     user.ID,
		}

		err := repo.CreateOrganization(ctx, org)
		require.NoError(t, err)
		assert.NotZero(t, org.ID)

		fetchedOrg, err := repo.GetOrganizationByID(ctx, org.ID)
		require.NoError(t, err)
		assert.Equal(t, org.ID, fetchedOrg.ID)
		assert.Equal(t, org.ProjectName, fetchedOrg.ProjectName)
	})

	t.Run("Update and Delete Organization", func(t *testing.T) {
		org := &common.Organization{
			ProjectName: "to_update",
			OwnerID:     user.ID,
		}
		err := repo.CreateOrganization(ctx, org)
		require.NoError(t, err)

		org.ProjectName = "updated_name"
		err = repo.UpdateOrganization(ctx, org)
		require.NoError(t, err)

		updatedOrg, err := repo.GetOrganizationByID(ctx, org.ID)
		require.NoError(t, err)
		assert.Equal(t, "updated_name", updatedOrg.ProjectName)

		err = repo.DeleteOrganization(ctx, org.ID)
		require.NoError(t, err)

		deletedOrg, err := repo.GetOrganizationByID(ctx, org.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedOrg)
	})

	t.Run("List Organizations by Owner", func(t *testing.T) {
		// Создаем несколько организаций
		orgs := []*common.Organization{
			{ProjectName: "org1", OwnerID: user.ID},
			{ProjectName: "org2", OwnerID: user.ID},
		}

		for _, o := range orgs {
			err := repo.CreateOrganization(ctx, o)
			require.NoError(t, err)
		}

		orgList, err := repo.ListOrganizationsByOwner(ctx, user.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(orgList), 2)
	})
}

func TestTeamRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	user := createTestUser(t, repo)
	org := createTestOrg(t, repo, user.ID)

	t.Run("Create and Get Team", func(t *testing.T) {
		f := "/dev"
		team := &common.Team{
			TeamName:       "dev",
			OwnerID:        user.ID,
			Folder:         &f,
			OrganizationID: org.ID,
		}

		err := repo.CreateTeam(ctx, team)
		require.NoError(t, err)
		assert.NotZero(t, team.ID)

		fetchedTeam, err := repo.GetTeamByID(ctx, team.ID)
		require.NoError(t, err)
		assert.Equal(t, team.ID, fetchedTeam.ID)
		assert.Equal(t, team.TeamName, fetchedTeam.TeamName)
	})

	t.Run("List Teams by Organization", func(t *testing.T) {
		// Создаем несколько команд
		teams := []*common.Team{
			{TeamName: "team1", OwnerID: user.ID, OrganizationID: org.ID},
			{TeamName: "team2", OwnerID: user.ID, OrganizationID: org.ID},
		}

		for _, te := range teams {
			err := repo.CreateTeam(ctx, te)
			require.NoError(t, err)
		}

		teamList, err := repo.ListTeamsByOrganization(ctx, org.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(teamList), 2)
	})
}

func TestApplicationRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	user := createTestUser(t, repo)
	org := createTestOrg(t, repo, user.ID)
	team := createTestTeam(t, repo, user.ID, org.ID)

	t.Run("Create and Get Application", func(t *testing.T) {
		d := "Test application"
		app := &common.Application{
			Name:        "app1",
			Description: &d,
			TeamID:      team.ID,
		}

		err := repo.CreateApplication(ctx, app)
		require.NoError(t, err)
		assert.NotZero(t, app.ID)

		fetchedApp, err := repo.GetApplicationByID(ctx, app.ID)
		require.NoError(t, err)
		assert.Equal(t, app.ID, fetchedApp.ID)
		assert.Equal(t, app.Name, fetchedApp.Name)
	})

	t.Run("List Applications by Team", func(t *testing.T) {
		// Создаем несколько приложений
		apps := []*common.Application{
			{Name: "app1", TeamID: team.ID},
			{Name: "app2", TeamID: team.ID},
		}

		for _, a := range apps {
			err := repo.CreateApplication(ctx, a)
			require.NoError(t, err)
		}

		appList, err := repo.ListApplicationsByTeam(ctx, team.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(appList), 2)
	})
}

func TestVersionRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	user := createTestUser(t, repo)
	org := createTestOrg(t, repo, user.ID)
	team := createTestTeam(t, repo, user.ID, org.ID)
	app := &common.Application{
		Name:   "app",
		TeamID: team.ID,
	}
	err := repo.CreateApplication(ctx, app)
	require.NoError(t, err)

	t.Run("Create and Get Version", func(t *testing.T) {
		version := &common.Version{
			ApplicationID: app.ID,
			Version:       "1.0.0",
		}

		err := repo.CreateVersion(ctx, version)
		require.NoError(t, err)
		assert.NotZero(t, version.ID)

		fetchedVersion, err := repo.GetVersionByID(ctx, version.ID)
		require.NoError(t, err)
		assert.Equal(t, version.ID, fetchedVersion.ID)
		assert.Equal(t, version.Version, fetchedVersion.Version)
	})

	t.Run("List Versions", func(t *testing.T) {
		// Создаем несколько версий
		versions := []*common.Version{
			{ApplicationID: app.ID, Version: "1.0.1"},
			{ApplicationID: app.ID, Version: "1.0.2"},
		}

		for _, v := range versions {
			err := repo.CreateVersion(ctx, v)
			require.NoError(t, err)
		}

		versionList, err := repo.ListVersions(ctx, app.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(versionList), 2)
	})
}

func TestScanRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	user := createTestUser(t, repo)
	org := createTestOrg(t, repo, user.ID)
	team := createTestTeam(t, repo, user.ID, org.ID)
	app := &common.Application{
		Name:   "app",
		TeamID: team.ID,
	}
	err := repo.CreateApplication(ctx, app)
	require.NoError(t, err)
	version := &common.Version{
		ApplicationID: app.ID,
		Version:       "1.0.0",
	}
	err = repo.CreateVersion(ctx, version)
	require.NoError(t, err)

	t.Run("Create and Get Scan", func(t *testing.T) {
		scan := &common.Scan{
			ScanDate:  time.Now(),
			VersionID: version.ID,
		}

		err := repo.CreateScan(ctx, scan)
		require.NoError(t, err)
		assert.NotZero(t, scan.ID)

		fetchedScan, err := repo.GetScanByID(ctx, scan.ID)
		require.NoError(t, err)
		assert.Equal(t, scan.ID, fetchedScan.ID)
		assert.Equal(t, scan.VersionID, fetchedScan.VersionID)
	})

	t.Run("List Scans", func(t *testing.T) {
		// Создаем несколько сканов
		scans := []*common.Scan{
			{ScanDate: time.Now(), VersionID: version.ID},
			{ScanDate: time.Now().Add(-time.Hour), VersionID: version.ID},
		}

		for _, s := range scans {
			err := repo.CreateScan(ctx, s)
			require.NoError(t, err)
		}

		scanList, err := repo.ListScans(ctx, version.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(scanList), 2)
	})
}

func TestScanInfoRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	user := createTestUser(t, repo)
	org := createTestOrg(t, repo, user.ID)
	team := createTestTeam(t, repo, user.ID, org.ID)
	app := &common.Application{
		Name:   "app",
		TeamID: team.ID,
	}
	err := repo.CreateApplication(ctx, app)
	require.NoError(t, err)
	version := &common.Version{
		ApplicationID: app.ID,
		Version:       "1.0.0",
	}
	err = repo.CreateVersion(ctx, version)
	require.NoError(t, err)
	scan := &common.Scan{
		ScanDate:  time.Now(),
		VersionID: version.ID,
	}
	err = repo.CreateScan(ctx, scan)
	require.NoError(t, err)

	t.Run("Create and Get ScanInfo", func(t *testing.T) {
		scanInfo := &common.ScanInfo{
			ScanID: scan.ID,
		}

		err := repo.CreateScanInfo(ctx, scanInfo)
		require.NoError(t, err)
		assert.NotZero(t, scanInfo.ID)

		fetchedScanInfo, err := repo.GetScanInfoByScanID(ctx, scan.ID)
		require.NoError(t, err)
		assert.Equal(t, scanInfo.ID, fetchedScanInfo.ID)
		assert.Equal(t, scanInfo.ScanID, fetchedScanInfo.ScanID)
	})
}

func TestScanRuleRepository(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPgxRepository(pool)
	ctx := context.Background()

	// Создаем тестовые данные
	user := createTestUser(t, repo)
	org := createTestOrg(t, repo, user.ID)
	team := createTestTeam(t, repo, user.ID, org.ID)
	app := &common.Application{
		Name:   "test-app",
		TeamID: team.ID,
	}
	err := repo.CreateApplication(ctx, app)
	require.NoError(t, err)

	// Вспомогательные переменные для boolean полей
	scaEnabled := true
	sastEnabled := false
	allowIncremental := true
	allowEmptyCode := false
	forcedSBOM := true
	activeBlocking := false

	t.Run("Create and Get ScanRule", func(t *testing.T) {
		rule := &common.ScanRule{
			ApplicationID:         app.ID,
			TeamID:                team.ID,
			OrganizationID:        org.ID,
			SCAScanEnabled:        &scaEnabled,
			SASTScanEnabled:       &sastEnabled,
			AllowIncrementalScans: &allowIncremental,
			AllowSASTEmptyCode:    &allowEmptyCode,
			ExcludeDirRegexpQueue: []string{"node_modules", "vendor"},
			ForcedDoOwnSBOM:       &forcedSBOM,
			ActiveBlockingSCA:     &activeBlocking,
		}

		// Тестируем создание правила
		err := repo.CreateScanRule(ctx, rule)
		require.NoError(t, err)
		assert.NotZero(t, rule.ID)

		// Тестируем получение по составному ключу
		fetchedRule, err := repo.GetScanRuleByComposite(ctx, app.ID, team.ID, org.ID)
		require.NoError(t, err)
		require.NotNil(t, fetchedRule)

		// Проверяем все поля
		assert.Equal(t, rule.ID, fetchedRule.ID)
		assert.Equal(t, rule.ApplicationID, fetchedRule.ApplicationID)
		assert.Equal(t, rule.TeamID, fetchedRule.TeamID)
		assert.Equal(t, rule.OrganizationID, fetchedRule.OrganizationID)
		assert.Equal(t, *rule.SCAScanEnabled, *fetchedRule.SCAScanEnabled)
		assert.Equal(t, *rule.SASTScanEnabled, *fetchedRule.SASTScanEnabled)
		assert.Equal(t, *rule.AllowIncrementalScans, *fetchedRule.AllowIncrementalScans)
		assert.Equal(t, *rule.AllowSASTEmptyCode, *fetchedRule.AllowSASTEmptyCode)
		assert.Equal(t, rule.ExcludeDirRegexpQueue, fetchedRule.ExcludeDirRegexpQueue)
		assert.Equal(t, *rule.ForcedDoOwnSBOM, *fetchedRule.ForcedDoOwnSBOM)
		assert.Equal(t, *rule.ActiveBlockingSCA, *fetchedRule.ActiveBlockingSCA)

		// Тестируем получение по ID
		fetchedById, err := repo.GetScanRuleByID(ctx, rule.ID)
		require.NoError(t, err)
		require.NotNil(t, fetchedById)
		assert.Equal(t, rule.ID, fetchedById.ID)
	})

	t.Run("Update ScanRule", func(t *testing.T) {
		// Создаем тестовое правило
		rule := &common.ScanRule{
			ApplicationID:         app.ID,
			TeamID:                team.ID,
			OrganizationID:        org.ID,
			SCAScanEnabled:        &scaEnabled,
			SASTScanEnabled:       &sastEnabled,
			AllowIncrementalScans: &allowIncremental,
		}
		err := repo.CreateScanRule(ctx, rule)
		require.NoError(t, err)

		// Обновляем поля
		newSastEnabled := true
		newAllowIncremental := false
		newExcludeDirs := []string{"build", "dist"}
		updatedRule := &common.ScanRule{
			ID:                    rule.ID,
			ApplicationID:         app.ID,
			TeamID:                team.ID,
			OrganizationID:        org.ID,
			SCAScanEnabled:        rule.SCAScanEnabled,
			SASTScanEnabled:       &newSastEnabled,
			AllowIncrementalScans: &newAllowIncremental,
			ExcludeDirRegexpQueue: newExcludeDirs,
		}

		err = repo.UpdateScanRule(ctx, updatedRule)
		require.NoError(t, err)

		// Проверяем обновленные данные
		fetched, err := repo.GetScanRuleByID(ctx, rule.ID)
		require.NoError(t, err)
		assert.Equal(t, *updatedRule.SASTScanEnabled, *fetched.SASTScanEnabled)
		assert.Equal(t, *updatedRule.AllowIncrementalScans, *fetched.AllowIncrementalScans)
		assert.Equal(t, updatedRule.ExcludeDirRegexpQueue, fetched.ExcludeDirRegexpQueue)
	})

	t.Run("List and Delete ScanRule", func(t *testing.T) {
		// Создаем несколько правил
		rule1 := &common.ScanRule{
			ApplicationID:  app.ID,
			TeamID:         team.ID,
			OrganizationID: org.ID,
			SCAScanEnabled: &scaEnabled,
		}
		rule2 := &common.ScanRule{
			ApplicationID:   app.ID,
			TeamID:          team.ID,
			OrganizationID:  org.ID,
			SASTScanEnabled: &sastEnabled,
		}
		require.NoError(t, repo.CreateScanRule(ctx, rule1))
		require.NoError(t, repo.CreateScanRule(ctx, rule2))

		// Тестируем получение списка
		rules, err := repo.ListScanRules(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(rules), 2)

		// Тестируем удаление
		err = repo.DeleteScanRule(ctx, rule1.ID)
		require.NoError(t, err)

		// Проверяем, что правило удалено
		deletedRule, err := repo.GetScanRuleByID(ctx, rule1.ID)
		require.NoError(t, err)
		assert.Nil(t, deletedRule)
	})

	t.Run("Edge Cases", func(t *testing.T) {
		// Тест на NULL значения
		nullRule := &common.ScanRule{
			ApplicationID:  app.ID,
			TeamID:         team.ID,
			OrganizationID: org.ID,
			// Остальные поля nil
		}
		err := repo.CreateScanRule(ctx, nullRule)
		require.NoError(t, err)

		fetchedNull, err := repo.GetScanRuleByID(ctx, nullRule.ID)
		require.NoError(t, err)
		assert.Nil(t, fetchedNull.SCAScanEnabled)
		assert.Nil(t, fetchedNull.SASTScanEnabled)
		assert.Nil(t, fetchedNull.ExcludeDirRegexpQueue)
	})
}
