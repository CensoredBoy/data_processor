package data_processor

import (
	"context"
	"data_processor/internal/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateScanRule(ctx context.Context, req *CreateScanRuleRequest) (*ScanRule, error) {
	rule := &common.ScanRule{
		ApplicationID:              common.Int32PtrToIntPtr(req.ApplicationId),
		TeamID:                     common.Int32PtrToIntPtr(req.TeamId),
		OrganizationID:             common.Int32PtrToIntPtr(&req.OrganizationId),
		SCAScanEnabled:             req.ScaScanEnabled,
		SASTScanEnabled:            req.SastScanEnabled,
		ApplicationPostfix:         req.ApplicationPostfix,
		AllowUnsafeExtDistribs:     req.AllowUnsafeExtDistribs,
		IgnoreRepositoryMembership: req.IgnoreRepositoryMembership,
		AllowIncrementalScans:      req.AllowIncrementalScans,
		AllowSASTEmptyCode:         req.AllowSastEmptyCode,
		ExcludeDirRegexpQueue:      req.ExcludeDirRegexpQueue,
		ForcedDoOwnSBOM:            req.ForcedDoOwnSbom,
		ActiveBlockingSCA:          req.ActiveBlockingSca,
	}

	if err := s.scanRuleRepo.CreateScanRule(ctx, rule); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create scan rule: %v", err)
	}

	return convertScanRuleToProto(rule), nil
}

func (s *Server) GetScanRule(ctx context.Context, req *GetScanRuleRequest) (*ScanRule, error) {
	rule, err := s.scanRuleRepo.GetScanRuleByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan rule: %v", err)
	}
	if rule == nil {
		return nil, status.Errorf(codes.NotFound, "scan rule not found")
	}

	return convertScanRuleToProto(rule), nil
}

func (s *Server) GetScanRuleByComposite(ctx context.Context, req *GetScanRuleByCompositeRequest) (*ScanRule, error) {
	rule, err := s.scanRuleRepo.GetScanRuleByComposite(
		ctx,
		int(req.ApplicationId),
		int(req.TeamId),
		int(req.OrganizationId),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan rule: %v", err)
	}
	if rule == nil {
		return nil, status.Errorf(codes.NotFound, "scan rule not found")
	}

	return convertScanRuleToProto(rule), nil
}

func (s *Server) UpdateScanRule(ctx context.Context, req *UpdateScanRuleRequest) (*ScanRule, error) {
	// Простое обновление без дополнительной логики
	updatedRule := &common.ScanRule{
		ID:                         int(req.Id),
		ApplicationID:              common.Int32PtrToIntPtr(req.ApplicationId),
		TeamID:                     common.Int32PtrToIntPtr(req.TeamId),
		OrganizationID:             common.Int32PtrToIntPtr(req.OrganizationId),
		SCAScanEnabled:             req.ScaScanEnabled,
		SASTScanEnabled:            req.SastScanEnabled,
		ApplicationPostfix:         req.ApplicationPostfix,
		AllowUnsafeExtDistribs:     req.AllowUnsafeExtDistribs,
		IgnoreRepositoryMembership: req.IgnoreRepositoryMembership,
		AllowIncrementalScans:      req.AllowIncrementalScans,
		AllowSASTEmptyCode:         req.AllowSastEmptyCode,
		ExcludeDirRegexpQueue:      req.ExcludeDirRegexpQueue,
		ForcedDoOwnSBOM:            req.ForcedDoOwnSbom,
		ActiveBlockingSCA:          req.ActiveBlockingSca,
	}

	if err := s.scanRuleRepo.UpdateScanRule(ctx, updatedRule); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update scan rule: %v", err)
	}

	// Получаем обновленное правило для возврата
	rule, err := s.scanRuleRepo.GetScanRuleByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated scan rule: %v", err)
	}

	return convertScanRuleToProto(rule), nil
}

func (s *Server) DeleteScanRule(ctx context.Context, req *DeleteScanRuleRequest) (*emptypb.Empty, error) {
	if err := s.scanRuleRepo.DeleteScanRule(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete scan rule: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListScanRules(ctx context.Context, req *ListScanRulesRequest) (*ListScanRulesResponse, error) {
	rules, err := s.scanRuleRepo.ListScanRules(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list scan rules: %v", err)
	}

	// Применяем пагинацию
	total := len(rules)
	offset := int(req.Offset)
	if offset > total {
		offset = total
	}
	end := offset + int(req.Limit)
	if end > total {
		end = total
	}
	paginatedRules := rules[offset:end]

	resp := &ListScanRulesResponse{
		ScanRules:  make([]*ScanRule, 0, len(paginatedRules)),
		TotalCount: int32(total),
	}

	for _, rule := range paginatedRules {
		resp.ScanRules = append(resp.ScanRules, convertScanRuleToProto(rule))
	}

	return resp, nil
}

func convertScanRuleToProto(rule *common.ScanRule) *ScanRule {
	protoRule := &ScanRule{
		Id:                         int32(rule.ID),
		ApplicationId:              common.Int32PtrFromIntPtr(rule.ApplicationID),
		TeamId:                     common.Int32PtrFromIntPtr(rule.TeamID),
		OrganizationId:             common.Int32FromIntPtr(rule.OrganizationID),
		ScaScanEnabled:             rule.SCAScanEnabled,
		SastScanEnabled:            rule.SASTScanEnabled,
		ApplicationPostfix:         rule.ApplicationPostfix,
		AllowUnsafeExtDistribs:     rule.AllowUnsafeExtDistribs,
		IgnoreRepositoryMembership: rule.IgnoreRepositoryMembership,
		AllowIncrementalScans:      rule.AllowIncrementalScans,
		AllowSastEmptyCode:         rule.AllowSASTEmptyCode,
		ExcludeDirRegexpQueue:      rule.ExcludeDirRegexpQueue,
		ForcedDoOwnSbom:            rule.ForcedDoOwnSBOM,
		ActiveBlockingSca:          rule.ActiveBlockingSCA,
	}
	return protoRule
}

func getUpdatedBoolValue(newVal *bool, currentVal *bool) *bool {
	if newVal != nil {
		val := newVal
		return val
	}
	return currentVal
}
