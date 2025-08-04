package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
	user := &common.User{
		Name:     req.Name,
		Password: req.Password,
	}

	if err := s.repositories.CreateUser(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &User{
		Id:   int32(user.ID),
		Name: user.Name,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *GetUserRequest) (*User, error) {
	user, err := s.repositories.GetUserByID(ctx, common.UserID(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &User{
		Id:   int32(user.ID),
		Name: user.Name,
	}, nil
}

func (s *Server) GetUserByName(ctx context.Context, req *GetUserByNameRequest) (*User, error) {
	user, err := s.repositories.GetUserByName(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}
	if user == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	return &User{
		Id:   int32(user.ID),
		Name: user.Name,
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, req *UpdateUserRequest) (*User, error) {
	// Получаем текущие данные пользователя
	currentUser, err := s.repositories.GetUserByID(ctx, common.UserID(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user data: %v", err)
	}
	if currentUser == nil {
		return nil, status.Errorf(codes.NotFound, "user not found")
	}

	// Подготавливаем обновленные данные
	updatedUser := &common.User{
		ID: common.UserID(req.Id),
	}

	// Обрабатываем Name
	if req.Name != nil {
		updatedUser.Name = *req.Name // Используем новое значение
	} else {
		updatedUser.Name = currentUser.Name // Сохраняем текущее значение
	}

	// Обрабатываем Password
	if req.Password != nil {
		updatedUser.Password = *req.Password // Используем новый пароль
	} else {
		updatedUser.Password = currentUser.Password // Сохраняем текущий пароль
	}

	// Выполняем обновление
	if err := s.repositories.UpdateUser(ctx, updatedUser); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &User{
		Id:   int32(updatedUser.ID),
		Name: updatedUser.Name,
	}, nil
}
func (s *Server) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteUser(ctx, common.UserID(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	users, err := s.repositories.ListUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	resp := &ListUsersResponse{}
	for _, user := range users {
		resp.Users = append(resp.Users, &User{
			Id:   int32(user.ID),
			Name: user.Name,
		})
	}

	return resp, nil
}
