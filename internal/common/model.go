package common

import "time"

type UserID int
type TeamID int
type OrgID int
type User struct {
	ID       UserID
	Name     string
	Password string
}

type Role struct {
	ID          int
	Name        string
	Description *string
	IsActive    *bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	OwnerID     UserID
}
type PermissionReadWrite struct {
	Read  bool
	Write bool
}
type Permission struct {
	ID             int
	Name           string
	Description    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Read           bool
	Write          bool
	OrganizationID *int
	TeamID         *int
}

type Organization struct {
	ID          int
	ProjectName string
	OwnerID     UserID
}

type Team struct {
	ID             int
	TeamName       string
	OwnerID        UserID
	Folder         *string
	OrganizationID int
}

type Application struct {
	ID          int
	Name        string
	Description *string
	TeamID      int
}

type Version struct {
	ID            int
	ApplicationID int
	Version       string
}

type Scan struct {
	ID        int
	ScanDate  time.Time
	VersionID int
}

type ScanInfo struct {
	ID     int
	ScanID int
}

type ScanRule struct {
	ID                    int
	ApplicationID         int
	TeamID                int
	OrganizationID        int
	SCAScanEnabled        *bool
	SASTScanEnabled       *bool
	AllowIncrementalScans *bool
	AllowSASTEmptyCode    *bool
	ExcludeDirRegexpQueue []string
	ForcedDoOwnSBOM       *bool
	ActiveBlockingSCA     *bool
}

type RoleScope struct {
	OrganizationID *int
	TeamID         *int
}

type RoleWithPermissions struct {
	Role        *Role
	Permissions []*Permission
	Scope       RoleScope
}
