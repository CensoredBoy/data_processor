package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateTeam(ctx context.Context, req *CreateTeamRequest) (*Team, error) {
	team := &common.Team{
		TeamName:       req.TeamName,
		OwnerID:        common.UserID(req.OwnerId),
		Folder:         req.Folder,
		OrganizationID: int(req.OrganizationId),
	}

	if err := s.repositories.CreateTeam(ctx, team); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create team: %v", err)
	}

	return &Team{
		Id:             int32(team.ID),
		TeamName:       team.TeamName,
		OwnerId:        int32(team.OwnerID),
		Folder:         team.Folder,
		OrganizationId: int32(team.OrganizationID),
	}, nil
}

func (s *Server) GetTeam(ctx context.Context, req *GetTeamRequest) (*Team, error) {
	team, err := s.repositories.GetTeamByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team: %v", err)
	}
	if team == nil {
		return nil, status.Errorf(codes.NotFound, "team not found")
	}

	return &Team{
		Id:             int32(team.ID),
		TeamName:       team.TeamName,
		OwnerId:        int32(team.OwnerID),
		Folder:         team.Folder,
		OrganizationId: int32(team.OrganizationID),
	}, nil
}

func (s *Server) GetTeamByName(ctx context.Context, req *GetTeamByNameRequest) (*Team, error) {
	team, err := s.repositories.GetTeamByName(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team: %v", err)
	}
	if team == nil {
		return nil, status.Errorf(codes.NotFound, "team not found")
	}

	return &Team{
		Id:             int32(team.ID),
		TeamName:       team.TeamName,
		OwnerId:        int32(team.OwnerID),
		Folder:         team.Folder,
		OrganizationId: int32(team.OrganizationID),
	}, nil
}

func (s *Server) UpdateTeam(ctx context.Context, req *UpdateTeamRequest) (*Team, error) {
	currentTeam, err := s.repositories.GetTeamByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current team: %v", err)
	}
	if currentTeam == nil {
		return nil, status.Errorf(codes.NotFound, "team not found")
	}

	updatedTeam := &common.Team{
		ID: int(req.Id),
	}

	if req.TeamName != nil {
		updatedTeam.TeamName = *req.TeamName
	} else {
		updatedTeam.TeamName = currentTeam.TeamName
	}

	if req.OwnerId != nil {
		updatedTeam.OwnerID = common.UserID(*req.OwnerId)
	} else {
		updatedTeam.OwnerID = currentTeam.OwnerID
	}

	if req.Folder != nil {
		updatedTeam.Folder = req.Folder
	} else {
		updatedTeam.Folder = currentTeam.Folder
	}

	if req.OrganizationId != nil {
		updatedTeam.OrganizationID = int(*req.OrganizationId)
	} else {
		updatedTeam.OrganizationID = currentTeam.OrganizationID
	}

	if err := s.repositories.UpdateTeam(ctx, updatedTeam); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update team: %v", err)
	}

	return &Team{
		Id:             int32(updatedTeam.ID),
		TeamName:       updatedTeam.TeamName,
		OwnerId:        int32(updatedTeam.OwnerID),
		Folder:         updatedTeam.Folder,
		OrganizationId: int32(updatedTeam.OrganizationID),
	}, nil
}

func (s *Server) DeleteTeam(ctx context.Context, req *DeleteTeamRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteTeam(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete team: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListTeams(ctx context.Context, req *ListTeamsRequest) (*ListTeamsResponse, error) {
	teams, err := s.repositories.ListTeams(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list teams: %v", err)
	}

	resp := &ListTeamsResponse{}
	for _, team := range teams {
		resp.Teams = append(resp.Teams, &Team{
			Id:             int32(team.ID),
			TeamName:       team.TeamName,
			OwnerId:        int32(team.OwnerID),
			Folder:         team.Folder,
			OrganizationId: int32(team.OrganizationID),
		})
	}

	return resp, nil
}

func (s *Server) ListTeamsByOrganization(ctx context.Context, req *ListByParentRequest) (*ListTeamsResponse, error) {
	teams, err := s.repositories.ListTeamsByOrganization(ctx, int(req.ParentId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list teams by organization: %v", err)
	}

	resp := &ListTeamsResponse{}
	for _, team := range teams {
		resp.Teams = append(resp.Teams, &Team{
			Id:             int32(team.ID),
			TeamName:       team.TeamName,
			OwnerId:        int32(team.OwnerID),
			Folder:         team.Folder,
			OrganizationId: int32(team.OrganizationID),
		})
	}

	return resp, nil
}

func (s *Server) ListTeamsByOwner(ctx context.Context, req *ListByOwnerRequest) (*ListTeamsResponse, error) {
	teams, err := s.repositories.ListTeamsByOwner(ctx, int(req.OwnerId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list teams by owner: %v", err)
	}

	resp := &ListTeamsResponse{}
	for _, team := range teams {
		resp.Teams = append(resp.Teams, &Team{
			Id:             int32(team.ID),
			TeamName:       team.TeamName,
			OwnerId:        int32(team.OwnerID),
			Folder:         team.Folder,
			OrganizationId: int32(team.OrganizationID),
		})
	}

	return resp, nil
}
