package repo

import (
	"context"
	"data_processor/internal/common"
)

// UserRepository handles user operations
type IUserRepository interface {
	GetUserID(ctx context.Context, user *common.User) (*common.UserID, error)
	CreateUser(ctx context.Context, user *common.User) error
	GetUserByID(ctx context.Context, id common.UserID) (*common.User, error)
	GetUserByName(ctx context.Context, name string) (*common.User, error)
	UpdateUser(ctx context.Context, user *common.User) error
	DeleteUser(ctx context.Context, id common.UserID) error
	ListUsers(ctx context.Context) ([]*common.User, error)
}

// RoleRepository handles role operations
type IRoleRepository interface {
	CreateRole(ctx context.Context, role *common.Role) (*common.RoleWithPermissions, error)
	UpdateRole(ctx context.Context, roleID int, name, description *string, isActive *bool) (*common.Role, error)
	GetRole(ctx context.Context, roleID int) (*common.RoleWithPermissions, error)
	DeleteRole(ctx context.Context, roleID int) error
	GetRoleByName(ctx context.Context, name string) (*common.Role, error)
	AddPermission(ctx context.Context, roleID int, permission *common.Permission) error
	RemovePermission(ctx context.Context, roleID int, permissionID int) error
	ListRolesByScope(ctx context.Context, scope common.RoleScope) ([]*common.RoleWithPermissions, error)
	AssignRoleToUser(ctx context.Context, userID common.UserID, roleID int) error
	RemoveRoleFromUser(ctx context.Context, userID common.UserID, roleID int) error
	GetUserRoles(ctx context.Context, userID common.UserID) ([]*common.Role, error)
	ListRoles(ctx context.Context) ([]*common.Role, error)
}

// PermissionRepository handles permission operations
type IPermissionRepository interface {
	CreatePermission(ctx context.Context, permission *common.Permission) error
	GetPermissionByID(ctx context.Context, id int) (*common.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*common.Permission, error)
	UpdatePermission(ctx context.Context, permission *common.Permission) error
	DeletePermission(ctx context.Context, id int) error
	ListPermissions(ctx context.Context) ([]*common.Permission, error)
	GetTeamPermissions(ctx context.Context, userID common.UserID, teamID common.TeamID) ([]common.PermissionReadWrite, error)
	GetOrganizationPermissions(ctx context.Context, userID common.UserID, orgID common.OrgID) ([]common.PermissionReadWrite, error)
}

// OrganizationRepository handles organization operations
type IOrganizationRepository interface {
	CreateOrganization(ctx context.Context, org *common.Organization) error
	GetOrganizationByID(ctx context.Context, id int) (*common.Organization, error)
	GetOrganizationByName(ctx context.Context, name string) (*common.Organization, error)
	UpdateOrganization(ctx context.Context, org *common.Organization) error
	DeleteOrganization(ctx context.Context, id int) error
	ListOrganizations(ctx context.Context) ([]*common.Organization, error)
	ListOrganizationsByOwner(ctx context.Context, ownerID common.UserID) ([]*common.Organization, error)
}

// TeamRepository handles team operations
type ITeamRepository interface {
	CreateTeam(ctx context.Context, team *common.Team) error
	GetTeamByID(ctx context.Context, id int) (*common.Team, error)
	GetTeamByName(ctx context.Context, name string) (*common.Team, error)
	UpdateTeam(ctx context.Context, team *common.Team) error
	DeleteTeam(ctx context.Context, id int) error
	ListTeams(ctx context.Context) ([]*common.Team, error)
	ListTeamsByOrganization(ctx context.Context, orgID int) ([]*common.Team, error)
	ListTeamsByOwner(ctx context.Context, ownerID int) ([]*common.Team, error)
}

// ApplicationRepository handles application operations
type IApplicationRepository interface {
	CreateApplication(ctx context.Context, app *common.Application) error
	GetApplicationByID(ctx context.Context, id int) (*common.Application, error)
	GetApplicationByName(ctx context.Context, name string) (*common.Application, error)
	UpdateApplication(ctx context.Context, app *common.Application) error
	DeleteApplication(ctx context.Context, id int) error
	ListApplications(ctx context.Context) ([]*common.Application, error)
	ListApplicationsByTeam(ctx context.Context, teamID int) ([]*common.Application, error)
}

// VersionRepository handles version operations
type IVersionRepository interface {
	CreateVersion(ctx context.Context, version *common.Version) error
	GetVersionByID(ctx context.Context, id int) (*common.Version, error)
	GetVersionByNumber(ctx context.Context, appID int, version string) (*common.Version, error)
	UpdateVersion(ctx context.Context, version *common.Version) error
	DeleteVersion(ctx context.Context, id int) error
	ListVersions(ctx context.Context, appID int) ([]*common.Version, error)
}

// ScanRepository handles scan operations
type IScanRepository interface {
	CreateScan(ctx context.Context, scan *common.Scan) error
	GetScanByID(ctx context.Context, id int) (*common.Scan, error)
	UpdateScan(ctx context.Context, scan *common.Scan) error
	DeleteScan(ctx context.Context, id int) error
	ListScans(ctx context.Context, versionID int) ([]*common.Scan, error)
}

// ScanInfoRepository handles scan info operations
type IScanInfoRepository interface {
	CreateScanInfo(ctx context.Context, scanInfo *common.ScanInfo) error
	GetScanInfoByID(ctx context.Context, id int) (*common.ScanInfo, error)
	GetScanInfoByScanID(ctx context.Context, scanID int) (*common.ScanInfo, error)
	UpdateScanInfo(ctx context.Context, scanInfo *common.ScanInfo) error
	DeleteScanInfo(ctx context.Context, id int) error
}

// ScanRuleRepository handles scan rule operations
type IScanRuleRepository interface {
	CreateScanRule(ctx context.Context, rule *common.ScanRule) error
	GetScanRuleByID(ctx context.Context, id int) (*common.ScanRule, error)
	UpdateScanRule(ctx context.Context, rule *common.ScanRule) error
	DeleteScanRule(ctx context.Context, id int) error
	ListScanRules(ctx context.Context) ([]*common.ScanRule, error)
	GetScanRuleByComposite(ctx context.Context, appID, teamID, orgID int) (*common.ScanRule, error)
}
