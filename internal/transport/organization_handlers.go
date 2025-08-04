package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*Organization, error) {
	org := &common.Organization{
		ProjectName: req.ProjectName,
		OwnerID:     common.UserID(req.OwnerId),
	}

	if err := s.repositories.CreateOrganization(ctx, org); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create organization: %v", err)
	}

	return &Organization{
		Id:          int32(org.ID),
		ProjectName: org.ProjectName,
		OwnerId:     int32(org.OwnerID),
	}, nil
}

func (s *Server) GetOrganization(ctx context.Context, req *GetOrganizationRequest) (*Organization, error) {
	org, err := s.repositories.GetOrganizationByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get organization: %v", err)
	}
	if org == nil {
		return nil, status.Errorf(codes.NotFound, "organization not found")
	}

	return &Organization{
		Id:          int32(org.ID),
		ProjectName: org.ProjectName,
		OwnerId:     int32(org.OwnerID),
	}, nil
}

func (s *Server) GetOrganizationByName(ctx context.Context, req *GetOrganizationByNameRequest) (*Organization, error) {
	org, err := s.repositories.GetOrganizationByName(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get organization: %v", err)
	}
	if org == nil {
		return nil, status.Errorf(codes.NotFound, "organization not found")
	}

	return &Organization{
		Id:          int32(org.ID),
		ProjectName: org.ProjectName,
		OwnerId:     int32(org.OwnerID),
	}, nil
}

func (s *Server) UpdateOrganization(ctx context.Context, req *UpdateOrganizationRequest) (*Organization, error) {
	currentOrg, err := s.repositories.GetOrganizationByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current organization: %v", err)
	}
	if currentOrg == nil {
		return nil, status.Errorf(codes.NotFound, "organization not found")
	}

	updatedOrg := &common.Organization{
		ID: int(req.Id),
	}

	if req.ProjectName != nil {
		updatedOrg.ProjectName = *req.ProjectName
	} else {
		updatedOrg.ProjectName = currentOrg.ProjectName
	}

	if req.OwnerId != nil {
		updatedOrg.OwnerID = common.UserID(*req.OwnerId)
	} else {
		updatedOrg.OwnerID = currentOrg.OwnerID
	}

	if err := s.repositories.UpdateOrganization(ctx, updatedOrg); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update organization: %v", err)
	}

	return &Organization{
		Id:          int32(updatedOrg.ID),
		ProjectName: updatedOrg.ProjectName,
		OwnerId:     int32(updatedOrg.OwnerID),
	}, nil
}

func (s *Server) DeleteOrganization(ctx context.Context, req *DeleteOrganizationRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteOrganization(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete organization: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListOrganizations(ctx context.Context, req *ListOrganizationsRequest) (*ListOrganizationsResponse, error) {
	orgs, err := s.repositories.ListOrganizations(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list organizations: %v", err)
	}

	resp := &ListOrganizationsResponse{}
	for _, org := range orgs {
		resp.Organizations = append(resp.Organizations, &Organization{
			Id:          int32(org.ID),
			ProjectName: org.ProjectName,
			OwnerId:     int32(org.OwnerID),
		})
	}

	return resp, nil
}

func (s *Server) ListOrganizationsByOwner(ctx context.Context, req *ListByOwnerRequest) (*ListOrganizationsResponse, error) {
	orgs, err := s.repositories.ListOrganizationsByOwner(ctx, common.UserID(req.OwnerId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list organizations by owner: %v", err)
	}

	resp := &ListOrganizationsResponse{}
	for _, org := range orgs {
		resp.Organizations = append(resp.Organizations, &Organization{
			Id:          int32(org.ID),
			ProjectName: org.ProjectName,
			OwnerId:     int32(org.OwnerID),
		})
	}

	return resp, nil
}
