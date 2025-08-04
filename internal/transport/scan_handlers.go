package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateScan(ctx context.Context, req *CreateScanRequest) (*Scan, error) {
	scan := &common.Scan{
		ScanDate:  req.ScanDate.AsTime(),
		VersionID: int(req.VersionId),
	}

	if err := s.repositories.CreateScan(ctx, scan); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create scan: %v", err)
	}

	return &Scan{
		Id:        int32(scan.ID),
		ScanDate:  timestamppb.New(scan.ScanDate),
		VersionId: int32(scan.VersionID),
	}, nil
}

func (s *Server) GetScan(ctx context.Context, req *GetScanRequest) (*Scan, error) {
	scan, err := s.repositories.GetScanByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan: %v", err)
	}
	if scan == nil {
		return nil, status.Errorf(codes.NotFound, "scan not found")
	}

	return &Scan{
		Id:        int32(scan.ID),
		ScanDate:  timestamppb.New(scan.ScanDate),
		VersionId: int32(scan.VersionID),
	}, nil
}

func (s *Server) UpdateScan(ctx context.Context, req *UpdateScanRequest) (*Scan, error) {
	// Получаем текущий скан
	currentScan, err := s.repositories.GetScanByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current scan: %v", err)
	}
	if currentScan == nil {
		return nil, status.Errorf(codes.NotFound, "scan not found")
	}

	// Подготавливаем обновленные данные
	updatedScan := &common.Scan{
		ID: int(req.Id),
	}

	// Обрабатываем ScanDate (если не передано - оставляем текущее)
	if req.ScanDate != nil {
		updatedScan.ScanDate = req.ScanDate.AsTime()
	} else {
		updatedScan.ScanDate = currentScan.ScanDate
	}

	// Обрабатываем VersionID (если не передано - оставляем текущее)
	if req.VersionId != nil {
		// Проверяем существование версии
		if _, err := s.repositories.GetVersionByID(ctx, int(*req.VersionId)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "version with id %d not found", req.VersionId)
		}
		updatedScan.VersionID = int(*req.VersionId)
	} else {
		updatedScan.VersionID = currentScan.VersionID
	}

	// Обновляем скан
	if err := s.repositories.UpdateScan(ctx, updatedScan); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update scan: %v", err)
	}

	// Возвращаем обновленный скан
	return &Scan{
		Id:        int32(updatedScan.ID),
		ScanDate:  timestamppb.New(updatedScan.ScanDate),
		VersionId: int32(updatedScan.VersionID),
	}, nil
}

func (s *Server) DeleteScan(ctx context.Context, req *DeleteScanRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteScan(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete scan: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListScans(ctx context.Context, req *ListScansRequest) (*ListScansResponse, error) {
	scans, err := s.repositories.ListScans(ctx, int(req.VersionId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list scans: %v", err)
	}

	resp := &ListScansResponse{}
	for _, scan := range scans {
		resp.Scans = append(resp.Scans, &Scan{
			Id:        int32(scan.ID),
			ScanDate:  timestamppb.New(scan.ScanDate),
			VersionId: int32(scan.VersionID),
		})
	}

	return resp, nil
}
