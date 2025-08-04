package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateScanInfo(ctx context.Context, req *CreateScanInfoRequest) (*ScanInfo, error) {
	scanInfo := &common.ScanInfo{
		ScanID: int(req.ScanId),
	}

	if err := s.repositories.CreateScanInfo(ctx, scanInfo); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create scan info: %v", err)
	}

	return &ScanInfo{
		Id:     int32(scanInfo.ID),
		ScanId: int32(scanInfo.ScanID),
	}, nil
}

func (s *Server) GetScanInfo(ctx context.Context, req *GetScanInfoRequest) (*ScanInfo, error) {
	scanInfo, err := s.repositories.GetScanInfoByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan info: %v", err)
	}
	if scanInfo == nil {
		return nil, status.Errorf(codes.NotFound, "scan info not found")
	}

	return &ScanInfo{
		Id:     int32(scanInfo.ID),
		ScanId: int32(scanInfo.ScanID),
	}, nil
}

func (s *Server) GetScanInfoByScan(ctx context.Context, req *GetScanInfoByScanRequest) (*ScanInfo, error) {
	scanInfo, err := s.repositories.GetScanInfoByScanID(ctx, int(req.ScanId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan info: %v", err)
	}
	if scanInfo == nil {
		return nil, status.Errorf(codes.NotFound, "scan info not found")
	}

	return &ScanInfo{
		Id:     int32(scanInfo.ID),
		ScanId: int32(scanInfo.ScanID),
	}, nil
}

func (s *Server) UpdateScanInfo(ctx context.Context, req *UpdateScanInfoRequest) (*ScanInfo, error) {
	// Получаем текущую информацию о сканировании
	currentScanInfo, err := s.repositories.GetScanInfoByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current scan info: %v", err)
	}
	if currentScanInfo == nil {
		return nil, status.Errorf(codes.NotFound, "scan info not found")
	}

	// Подготавливаем обновленные данные
	updatedScanInfo := &common.ScanInfo{
		ID: int(req.Id),
	}

	// Обрабатываем ScanID (если не передано - оставляем текущее)
	if req.ScanId != nil {
		// Проверяем существование скана
		if _, err := s.repositories.GetScanByID(ctx, int(*req.ScanId)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "scan with id %d not found", req.ScanId)
		}
		updatedScanInfo.ScanID = int(*req.ScanId)
	} else {
		updatedScanInfo.ScanID = currentScanInfo.ScanID
	}

	// Обновляем информацию о сканировании
	if err := s.repositories.UpdateScanInfo(ctx, updatedScanInfo); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update scan info: %v", err)
	}

	// Возвращаем обновленную информацию
	return &ScanInfo{
		Id:     int32(updatedScanInfo.ID),
		ScanId: int32(updatedScanInfo.ScanID),
	}, nil
}

func (s *Server) DeleteScanInfo(ctx context.Context, req *DeleteScanInfoRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteScanInfo(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete scan info: %v", err)
	}
	return &emptypb.Empty{}, nil
}
