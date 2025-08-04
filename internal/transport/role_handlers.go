package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateRole(ctx context.Context, req *CreateRoleRequest) (*RoleWithPermissions, error) {
	role := &common.Role{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		OwnerID:     int(req.OwnerId),
	}

	roleWithPerms, err := s.repositories.CreateRole(ctx, role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create role: %v", err)
	}

	return convertRoleWithPermissions(roleWithPerms), nil
}

func (s *Server) AddPermission(ctx context.Context, req *AddPermissionRequest) (*RoleWithPermissions, error) {
	// Получаем permission для определения его scope
	perm, err := s.repositories.GetPermissionByID(ctx, int(req.PermissionId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get permission: %v", err)
	}
	if perm == nil {
		return nil, status.Errorf(codes.NotFound, "permission not found")
	}

	// Создаем временный permission для передачи в AddPermission
	tmpPerm := &common.Permission{ID: int(req.PermissionId)}
	if perm.OrganizationID != nil {
		tmpPerm.OrganizationID = perm.OrganizationID
	} else if perm.TeamID != nil {
		tmpPerm.TeamID = perm.TeamID
	}

	if err := s.repositories.AddPermission(ctx, int(req.RoleId), tmpPerm); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add permission: %v", err)
	}

	// Возвращаем обновленную роль
	roleWithPerms, err := s.repositories.GetRole(ctx, int(req.RoleId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
	}

	return convertRoleWithPermissions(roleWithPerms), nil
}

func (s *Server) GetRole(ctx context.Context, req *GetRoleRequest) (*RoleWithPermissions, error) {
	role, err := s.repositories.GetRole(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
	}
	if role == nil {
		return nil, status.Errorf(codes.NotFound, "role not found")
	}

	return convertRoleWithPermissions(role), nil
}

func (s *Server) GetRoleByName(ctx context.Context, req *GetRoleByNameRequest) (*Role, error) {
	role, err := s.repositories.GetRoleByName(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
	}
	if role == nil {
		return nil, status.Errorf(codes.NotFound, "role not found")
	}

	return &Role{
		Id:          int32(role.ID),
		Name:        role.Name,
		Description: role.Description,
		IsActive:    role.IsActive,
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
		OwnerId:     int32(role.OwnerID),
	}, nil
}

func (s *Server) UpdateRole(ctx context.Context, req *UpdateRoleRequest) (*Role, error) {
	var (
		name        *string
		description *string
		isActive    *bool
	)

	if req.Name != nil {
		name = req.Name
	}
	if req.Description != nil {
		description = req.Description
	}
	if req.IsActive != nil {
		isActive = req.IsActive
	}

	role, err := s.repositories.UpdateRole(ctx, int(req.Id), name, description, isActive)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}

	return &Role{
		Id:          int32(role.ID),
		Name:        role.Name,
		Description: role.Description,
		IsActive:    role.IsActive,
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
		OwnerId:     int32(role.OwnerID),
	}, nil
}

func (s *Server) DeleteRole(ctx context.Context, req *DeleteRoleRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteRole(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}
	return &emptypb.Empty{}, nil
}

//func (s *Server) RemovePermission(ctx context.Context, req *RemovePermissionRequest) (*RoleWithPermissions, error) {
//	if err := s.repositories.RemovePermission(ctx, int(req.RoleId), req.PermissionIds}); err != nil {
//		return nil, status.Errorf(codes.Internal, "failed to remove permission: %v", err)
//	}
//
//	roleWithPerms, err := s.repositories.GetRole(ctx, int(req.RoleId))
//	if err != nil {
//		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
//	}
//
//	return convertRoleWithPermissions(roleWithPerms), nil
//}

func (s *Server) ListRoles(ctx context.Context, req *ListRolesRequest) (*ListRolesResponse, error) {
	roles, err := s.repositories.ListRoles(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	resp := &ListRolesResponse{}
	for _, role := range roles {
		resp.Roles = append(resp.Roles, &Role{
			Id:          int32(role.ID),
			Name:        role.Name,
			Description: role.Description,
			IsActive:    role.IsActive,
			CreatedAt:   timestamppb.New(role.CreatedAt),
			UpdatedAt:   timestamppb.New(role.UpdatedAt),
			OwnerId:     int32(role.OwnerID),
		})
	}

	return resp, nil
}

func (s *Server) ListRolesByScope(ctx context.Context, req *ListRolesByScopeRequest) (*ListRolesWithPermissionsResponse, error) {
	var scope common.RoleScope

	switch s := req.Scope.(type) {
	case *ListRolesByScopeRequest_OrganizationId:
		orgID := int(s.OrganizationId)
		scope.OrganizationID = &orgID
	case *ListRolesByScopeRequest_TeamId:
		teamID := int(s.TeamId)
		scope.TeamID = &teamID
	}

	roles, err := s.repositories.ListRolesByScope(ctx, scope)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles by scope: %v", err)
	}

	resp := &ListRolesWithPermissionsResponse{}
	for _, role := range roles {
		resp.Roles = append(resp.Roles, convertRoleWithPermissions(role))
	}

	return resp, nil
}

func (s *Server) AssignRoleToUser(ctx context.Context, req *AssignRoleRequest) (*emptypb.Empty, error) {
	if err := s.repositories.AssignRoleToUser(ctx, int(req.UserId), int(req.RoleId)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign role to user: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) RemoveRoleFromUser(ctx context.Context, req *RemoveRoleRequest) (*emptypb.Empty, error) {
	if err := s.repositories.RemoveRoleFromUser(ctx, int(req.UserId), int(req.RoleId)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove role from user: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserRoles(ctx context.Context, req *GetUserRolesRequest) (*ListRolesResponse, error) {
	roles, err := s.repositories.GetUserRoles(ctx, int(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user roles: %v", err)
	}

	resp := &ListRolesResponse{}
	for _, role := range roles {
		resp.Roles = append(resp.Roles, &Role{
			Id:          int32(role.ID),
			Name:        role.Name,
			Description: role.Description,
			IsActive:    role.IsActive,
			CreatedAt:   timestamppb.New(role.CreatedAt),
			UpdatedAt:   timestamppb.New(role.UpdatedAt),
			OwnerId:     int32(role.OwnerID),
		})
	}

	return resp, nil
}

// Вспомогательные функции преобразования
func convertRoleWithPermissions(r *common.RoleWithPermissions) *RoleWithPermissions {
	role := &Role{
		Id:          int32(r.Role.ID),
		Name:        r.Role.Name,
		Description: r.Role.Description,
		IsActive:    r.Role.IsActive,
		CreatedAt:   timestamppb.New(r.Role.CreatedAt),
		UpdatedAt:   timestamppb.New(r.Role.UpdatedAt),
		OwnerId:     int32(r.Role.OwnerID),
	}

	permissions := make([]*Permission, len(r.Permissions))
	for i, p := range r.Permissions {
		perm := &Permission{
			Id:          int32(p.ID),
			Name:        p.Name,
			Description: p.Description,
			Read:        p.Read,
			Write:       p.Write,
		}

		if p.OrganizationID != nil {
			perm.Scope = &Permission_OrganizationId{OrganizationId: int32(*p.OrganizationID)}
		} else if p.TeamID != nil {
			perm.Scope = &Permission_TeamId{TeamId: int32(*p.TeamID)}
		}

		permissions[i] = perm
	}

	return &RoleWithPermissions{
		Role:        role,
		Permissions: permissions,
	}
}
