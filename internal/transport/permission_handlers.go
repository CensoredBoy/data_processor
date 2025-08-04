package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)

func (s *Server) CreatePermission(ctx context.Context, req *CreatePermissionRequest) (*Permission, error) {
	perm := &common.Permission{
		Name:        req.Name,
		Description: req.Description,
		Read:        req.Read,
		Write:       req.Write,
	}

	switch scope := req.Scope.(type) {
	case *CreatePermissionRequest_OrganizationId:
		orgID := int(scope.OrganizationId)
		perm.OrganizationID = &orgID
	case *CreatePermissionRequest_TeamId:
		teamID := int(scope.TeamId)
		perm.TeamID = &teamID
	}

	if err := s.repositories.CreatePermission(ctx, perm); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create permission: %v", err)
	}

	resp := &Permission{
		Id:          int32(perm.ID),
		Name:        perm.Name,
		Description: perm.Description,
		Read:        perm.Read,
		Write:       perm.Write,
	}

	if perm.OrganizationID != nil {
		resp.Scope = &Permission_OrganizationId{OrganizationId: int32(*perm.OrganizationID)}
	} else if perm.TeamID != nil {
		resp.Scope = &Permission_TeamId{TeamId: int32(*perm.TeamID)}
	}

	return resp, nil
}

func (s *Server) GetPermission(ctx context.Context, req *GetPermissionRequest) (*Permission, error) {
	perm, err := s.repositories.GetPermissionByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get permission: %v", err)
	}
	if perm == nil {
		return nil, status.Errorf(codes.NotFound, "permission not found")
	}

	resp := &Permission{
		Id:          int32(perm.ID),
		Name:        perm.Name,
		Description: perm.Description,
		Read:        perm.Read,
		Write:       perm.Write,
	}

	if perm.OrganizationID != nil {
		resp.Scope = &Permission_OrganizationId{OrganizationId: int32(*perm.OrganizationID)}
	} else if perm.TeamID != nil {
		resp.Scope = &Permission_TeamId{TeamId: int32(*perm.TeamID)}
	}

	return resp, nil
}

func (s *Server) GetTeamPermissions(
	ctx context.Context,
	req *GetTeamPermissionsRequest,
) (*GetPermissionsResponse, error) {
	userID := common.UserID(req.UserId)

	teamID := common.TeamID(req.TeamId)

	perms, err := s.repositories.GetTeamPermissions(ctx, userID, teamID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get team permissions: %v", err)
	}

	return &GetPermissionsResponse{
		Permissions: toPermissionProtos(perms),
	}, nil
}
func toPermissionProtos(perms []common.PermissionReadWrite) []*PermissionReadWrite {
	result := make([]*PermissionReadWrite, 0, len(perms))
	for _, p := range perms {
		result = append(result, &PermissionReadWrite{
			Read:  p.Read,
			Write: p.Write,
		})
	}
	return result
}
func (s *Server) GetOrganizationPermissions(
	ctx context.Context,
	req *GetOrganizationPermissionsRequest,
) (*GetPermissionsResponse, error) {
	userID := common.UserID(req.UserId)

	orgID := common.OrgID(req.OrgId)

	perms, err := s.repositories.GetOrganizationPermissions(ctx, userID, orgID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get organization permissions: %v", err)
	}

	return &GetPermissionsResponse{
		Permissions: toPermissionProtos(perms),
	}, nil
}

func (s *Server) GetPermissionByName(ctx context.Context, req *GetPermissionByNameRequest) (*Permission, error) {
	// Валидация входных данных
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "permission name cannot be empty")
	}

	// Получаем permission из репозитория
	perm, err := s.repositories.GetPermissionByName(ctx, req.Name)
	if err != nil {
		log.Printf("Error getting permission by name '%s': %v", req.Name, err)
		return nil, status.Errorf(codes.Internal, "failed to get permission")
	}
	if perm == nil {
		return nil, status.Errorf(codes.NotFound, "permission '%s' not found", req.Name)
	}

	// Создаем response объект
	resp := &Permission{
		Id:          int32(perm.ID),
		Name:        perm.Name,
		Description: perm.Description,
		Read:        perm.Read,
		Write:       perm.Write,
	}

	// Устанавливаем scope в зависимости от типа permission
	switch {
	case perm.OrganizationID != nil:
		resp.Scope = &Permission_OrganizationId{OrganizationId: int32(*perm.OrganizationID)}
	case perm.TeamID != nil:
		resp.Scope = &Permission_TeamId{TeamId: int32(*perm.TeamID)}
	default:
		log.Printf("Permission '%s' has no defined scope", req.Name)
		return nil, status.Errorf(codes.Internal, "permission scope not defined")
	}

	return resp, nil
}

func (s *Server) UpdatePermission(ctx context.Context, req *UpdatePermissionRequest) (*Permission, error) {
	// 1. Получаем текущее состояние permission
	currentPerm, err := s.repositories.GetPermissionByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current permission: %v", err)
	}
	if currentPerm == nil {
		return nil, status.Errorf(codes.NotFound, "permission with id %d not found", req.Id)
	}

	// 2. Подготавливаем обновленные данные
	updatedPerm := &common.Permission{
		ID: int(req.Id),
	}

	// 3. Обрабатываем Name
	if req.Name != nil {
		updatedPerm.Name = *req.Name
	} else {
		updatedPerm.Name = currentPerm.Name
	}

	// 4. Обрабатываем Description
	if req.Description != nil {
		updatedPerm.Description = req.Description
	} else {
		updatedPerm.Description = currentPerm.Description
	}

	// 5. Обрабатываем Read/Write
	if req.Read != nil {
		updatedPerm.Read = *req.Read
	} else {
		updatedPerm.Read = currentPerm.Read
	}

	if req.Write != nil {
		updatedPerm.Write = *req.Write
	} else {
		updatedPerm.Write = currentPerm.Write
	}

	// 6. Обрабатываем Scope
	switch scope := req.Scope.(type) {
	case *UpdatePermissionRequest_OrganizationId:
		orgID := int(scope.OrganizationId)
		// Проверяем существование организации
		if _, err := s.repositories.GetOrganizationByID(ctx, orgID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "organization with id %d not found", orgID)
		}
		updatedPerm.OrganizationID = &orgID
		updatedPerm.TeamID = nil
	case *UpdatePermissionRequest_TeamId:
		teamID := int(scope.TeamId)
		// Проверяем существование команды
		if _, err := s.repositories.GetTeamByID(ctx, teamID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "team with id %d not found", teamID)
		}
		updatedPerm.TeamID = &teamID
		updatedPerm.OrganizationID = nil
	default:
		// Сохраняем текущий scope если не указан новый
		updatedPerm.OrganizationID = currentPerm.OrganizationID
		updatedPerm.TeamID = currentPerm.TeamID
	}

	// 7. Проверяем уникальность имени permission
	if existingPerm, err := s.repositories.GetPermissionByName(ctx, updatedPerm.Name); err == nil && existingPerm != nil && existingPerm.ID != updatedPerm.ID {
		return nil, status.Errorf(codes.AlreadyExists, "permission with name '%s' already exists", updatedPerm.Name)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check permission uniqueness: %v", err)
	}

	// 8. Обновляем permission в репозитории
	if err := s.repositories.UpdatePermission(ctx, updatedPerm); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update permission: %v", err)
	}

	// 9. Формируем ответ
	resp := &Permission{
		Id:          int32(updatedPerm.ID),
		Name:        updatedPerm.Name,
		Description: updatedPerm.Description,
		Read:        updatedPerm.Read,
		Write:       updatedPerm.Write,
	}

	if updatedPerm.OrganizationID != nil {
		resp.Scope = &Permission_OrganizationId{OrganizationId: int32(*updatedPerm.OrganizationID)}
	} else if updatedPerm.TeamID != nil {
		resp.Scope = &Permission_TeamId{TeamId: int32(*updatedPerm.TeamID)}
	}

	return resp, nil
}

func (s *Server) DeletePermission(ctx context.Context, req *DeletePermissionRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeletePermission(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete permission: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListPermissions(ctx context.Context, req *ListPermissionsRequest) (*ListPermissionsResponse, error) {
	perms, err := s.repositories.ListPermissions(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions: %v", err)
	}

	resp := &ListPermissionsResponse{}
	for _, perm := range perms {
		p := &Permission{
			Id:          int32(perm.ID),
			Name:        perm.Name,
			Description: perm.Description,
			Read:        perm.Read,
			Write:       perm.Write,
		}

		if perm.OrganizationID != nil {
			p.Scope = &Permission_OrganizationId{OrganizationId: int32(*perm.OrganizationID)}
		} else if perm.TeamID != nil {
			p.Scope = &Permission_TeamId{TeamId: int32(*perm.TeamID)}
		}

		resp.Permissions = append(resp.Permissions, p)
	}

	return resp, nil
}
