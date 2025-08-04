package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateVersion(ctx context.Context, req *CreateVersionRequest) (*Version, error) {
	version := &common.Version{
		ApplicationID: int(req.ApplicationId),
		Version:       req.Version,
	}

	if err := s.repositories.CreateVersion(ctx, version); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create version: %v", err)
	}

	return &Version{
		Id:            int32(version.ID),
		ApplicationId: int32(version.ApplicationID),
		Version:       version.Version,
	}, nil
}

func (s *Server) GetVersion(ctx context.Context, req *GetVersionRequest) (*Version, error) {
	version, err := s.repositories.GetVersionByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get version: %v", err)
	}
	if version == nil {
		return nil, status.Errorf(codes.NotFound, "version not found")
	}

	return &Version{
		Id:            int32(version.ID),
		ApplicationId: int32(version.ApplicationID),
		Version:       version.Version,
	}, nil
}

func (s *Server) GetVersionByNumber(ctx context.Context, req *GetVersionByNumberRequest) (*Version, error) {
	version, err := s.repositories.GetVersionByNumber(ctx, int(req.ApplicationId), req.Version)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get version: %v", err)
	}
	if version == nil {
		return nil, status.Errorf(codes.NotFound, "version not found")
	}

	return &Version{
		Id:            int32(version.ID),
		ApplicationId: int32(version.ApplicationID),
		Version:       version.Version,
	}, nil
}

func (s *Server) UpdateVersion(ctx context.Context, req *UpdateVersionRequest) (*Version, error) {
	// Получаем текущую версию
	currentVersion, err := s.repositories.GetVersionByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current version: %v", err)
	}
	if currentVersion == nil {
		return nil, status.Errorf(codes.NotFound, "version not found")
	}

	// Подготавливаем обновленные данные
	updatedVersion := &common.Version{
		ID: int(req.Id),
	}

	// Обрабатываем ApplicationID (если не передано - оставляем текущее)
	if req.ApplicationId != nil {
		updatedVersion.ApplicationID = int(*req.ApplicationId)
	} else {
		updatedVersion.ApplicationID = currentVersion.ApplicationID
	}

	// Обрабатываем Version (если не передано - оставляем текущее)
	if req.Version != nil {
		updatedVersion.Version = *req.Version
	} else {
		updatedVersion.Version = currentVersion.Version
	}

	// Проверяем, что версия с такими параметрами не существует
	if existingVersion, err := s.repositories.GetVersionByNumber(
		ctx,
		updatedVersion.ApplicationID,
		updatedVersion.Version,
	); err == nil && existingVersion != nil && existingVersion.ID != updatedVersion.ID {
		return nil, status.Errorf(
			codes.AlreadyExists,
			"version %s already exists for this application",
			updatedVersion.Version,
		)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check version uniqueness: %v", err)
	}

	// Обновляем версию
	if err := s.repositories.UpdateVersion(ctx, updatedVersion); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update version: %v", err)
	}

	// Возвращаем обновленную версию
	return &Version{
		Id:            int32(updatedVersion.ID),
		ApplicationId: int32(updatedVersion.ApplicationID),
		Version:       updatedVersion.Version,
	}, nil
}

func (s *Server) DeleteVersion(ctx context.Context, req *DeleteVersionRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteVersion(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete version: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListVersions(ctx context.Context, req *ListVersionsRequest) (*ListVersionsResponse, error) {
	versions, err := s.repositories.ListVersions(ctx, int(req.ApplicationId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list versions: %v", err)
	}

	resp := &ListVersionsResponse{}
	for _, version := range versions {
		resp.Versions = append(resp.Versions, &Version{
			Id:            int32(version.ID),
			ApplicationId: int32(version.ApplicationID),
			Version:       version.Version,
		})
	}

	return resp, nil
}
