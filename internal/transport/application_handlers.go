package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateApplication(ctx context.Context, req *CreateApplicationRequest) (*Application, error) {
	app := &common.Application{
		Name:        req.Name,
		Description: req.Description,
		TeamID:      int(req.TeamId),
	}

	if err := s.repositories.CreateApplication(ctx, app); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create application: %v", err)
	}

	return &Application{
		Id:          int32(app.ID),
		Name:        app.Name,
		Description: app.Description,
		TeamId:      int32(app.TeamID),
	}, nil
}

func (s *Server) GetApplication(ctx context.Context, req *GetApplicationRequest) (*Application, error) {
	app, err := s.repositories.GetApplicationByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get application: %v", err)
	}
	if app == nil {
		return nil, status.Errorf(codes.NotFound, "application not found")
	}

	return &Application{
		Id:          int32(app.ID),
		Name:        app.Name,
		Description: app.Description,
		TeamId:      int32(app.TeamID),
	}, nil
}

func (s *Server) GetApplicationByName(ctx context.Context, req *GetApplicationByNameRequest) (*Application, error) {
	app, err := s.repositories.GetApplicationByName(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get application: %v", err)
	}
	if app == nil {
		return nil, status.Errorf(codes.NotFound, "application not found")
	}

	return &Application{
		Id:          int32(app.ID),
		Name:        app.Name,
		Description: app.Description,
		TeamId:      int32(app.TeamID),
	}, nil
}

func (s *Server) UpdateApplication(ctx context.Context, req *UpdateApplicationRequest) (*Application, error) {
	// Получаем текущее состояние приложения
	currentApp, err := s.repositories.GetApplicationByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current application: %v", err)
	}
	if currentApp == nil {
		return nil, status.Errorf(codes.NotFound, "application not found")
	}

	// Подготавливаем обновленные данные
	updatedApp := &common.Application{
		ID: int(req.Id),
	}

	// Обрабатываем Name
	if req.Name != nil {
		updatedApp.Name = *req.Name
	} else {
		updatedApp.Name = currentApp.Name
	}

	// Обрабатываем Description
	if req.Description != nil {
		updatedApp.Description = req.Description
	} else {
		updatedApp.Description = currentApp.Description
	}

	// Обрабатываем TeamID
	if req.TeamId != nil {
		updatedApp.TeamID = int(*req.TeamId)
	} else {
		updatedApp.TeamID = currentApp.TeamID
	}

	// Обновляем приложение
	if err := s.repositories.UpdateApplication(ctx, updatedApp); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update application: %v", err)
	}

	// Возвращаем обновленное приложение
	return &Application{
		Id:          int32(updatedApp.ID),
		Name:        updatedApp.Name,
		Description: updatedApp.Description,
		TeamId:      int32(updatedApp.TeamID),
	}, nil
}

func (s *Server) DeleteApplication(ctx context.Context, req *DeleteApplicationRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteApplication(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete application: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListApplications(ctx context.Context, req *ListApplicationsRequest) (*ListApplicationsResponse, error) {
	apps, err := s.repositories.ListApplications(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list applications: %v", err)
	}

	resp := &ListApplicationsResponse{}
	for _, app := range apps {
		resp.Applications = append(resp.Applications, &Application{
			Id:          int32(app.ID),
			Name:        app.Name,
			Description: app.Description,
			TeamId:      int32(app.TeamID),
		})
	}

	return resp, nil
}

func (s *Server) ListApplicationsByTeam(ctx context.Context, req *ListByParentRequest) (*ListApplicationsResponse, error) {
	apps, err := s.repositories.ListApplicationsByTeam(ctx, int(req.ParentId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list applications by team: %v", err)
	}

	resp := &ListApplicationsResponse{}
	for _, app := range apps {
		resp.Applications = append(resp.Applications, &Application{
			Id:          int32(app.ID),
			Name:        app.Name,
			Description: app.Description,
			TeamId:      int32(app.TeamID),
		})
	}

	return resp, nil
}
