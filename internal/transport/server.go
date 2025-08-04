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

	repositories *repo.PgxRepository
}

func NewServer(repo *repo.PgxRepository) *Server {
	return &Server{
		repositories: repo,
	}
}
