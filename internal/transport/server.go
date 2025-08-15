package data_processor

import "data_processor/internal/repo"

type Server struct {
	UnimplementedUserServiceServer
	UnimplementedOrganizationServiceServer
	UnimplementedTeamServiceServer
	UnimplementedApplicationServiceServer
	UnimplementedVersionServiceServer
	UnimplementedScanServiceServer
	UnimplementedScanInfoServiceServer
	UnimplementedScanRuleServiceServer
	UnimplementedPermissionServiceServer
	UnimplementedRoleServiceServer

	userRepo        repo.IUserRepository
	permRepo        repo.IPermissionRepository
	roleRepo        repo.IRoleRepository
	orgRepo         repo.IOrganizationRepository
	applicationRepo repo.IApplicationRepository
	scanRepo        repo.IScanRepository
	scanRuleRepo    repo.IScanRuleRepository
	scanInfoRepo    repo.IScanInfoRepository
	teamRepo        repo.ITeamRepository
	versionRepo     repo.IVersionRepository
}

func NewServer(userRepo repo.IUserRepository,
	permRepo repo.IPermissionRepository,
	roleRepo repo.IRoleRepository,
	orgRepo repo.IOrganizationRepository,
	applicationRepo repo.IApplicationRepository,
	scanRepo repo.IScanRepository,
	scanRuleRepo repo.IScanRuleRepository,
	scanInfoRepo repo.IScanInfoRepository,
	teamRepo repo.ITeamRepository,
	versionRepo repo.IVersionRepository) *Server {
	return &Server{
		userRepo:        userRepo,
		permRepo:        permRepo,
		roleRepo:        roleRepo,
		orgRepo:         orgRepo,
		applicationRepo: applicationRepo,
		scanRepo:        scanRepo,
		scanRuleRepo:    scanRuleRepo,
		scanInfoRepo:    scanInfoRepo,
		teamRepo:        teamRepo,
		versionRepo:     versionRepo,
	}
}
