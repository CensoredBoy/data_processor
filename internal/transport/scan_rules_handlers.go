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
		ApplicationID:         int(req.ApplicationId),
		TeamID:                int(req.TeamId),
		OrganizationID:        int(req.OrganizationId),
		SCAScanEnabled:        req.ScaScanEnabled,
		SASTScanEnabled:       req.SastScanEnabled,
		AllowIncrementalScans: req.AllowIncrementalScans,
		AllowSASTEmptyCode:    req.AllowSastEmptyCode,
		ExcludeDirRegexpQueue: req.ExcludeDirRegexpQueue,
		ForcedDoOwnSBOM:       req.ForcedDoOwnSbom,
		ActiveBlockingSCA:     req.ActiveBlockingSca,
	}

	if err := s.repositories.CreateScanRule(ctx, rule); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create scan rule: %v", err)
	}

	return convertScanRuleToProto(rule), nil
}

func (s *Server) GetScanRule(ctx context.Context, req *GetScanRuleRequest) (*ScanRule, error) {
	rule, err := s.repositories.GetScanRuleByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan rule: %v", err)
	}
	if rule == nil {
		return nil, status.Errorf(codes.NotFound, "scan rule not found")
	}

	return convertScanRuleToProto(rule), nil
}

func (s *Server) GetScanRuleByComposite(ctx context.Context, req *GetScanRuleByCompositeRequest) (*ScanRule, error) {
	rule, err := s.repositories.GetScanRuleByComposite(
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
	currentRule, err := s.repositories.GetScanRuleByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current scan rule: %v", err)
	}
	if currentRule == nil {
		return nil, status.Errorf(codes.NotFound, "scan rule not found")
	}

	updatedRule := &common.ScanRule{
		ID: int(req.Id),
	}

	// Обновляем основные поля
	if req.ApplicationId != nil {
		if _, err := s.repositories.GetApplicationByID(ctx, int(*req.ApplicationId)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "application with id %d not found", *req.ApplicationId)
		}
		updatedRule.ApplicationID = int(*req.ApplicationId)
	} else {
		updatedRule.ApplicationID = currentRule.ApplicationID
	}

	if req.TeamId != nil {
		if _, err := s.repositories.GetTeamByID(ctx, int(*req.TeamId)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "team with id %d not found", *req.TeamId)
		}
		updatedRule.TeamID = int(*req.TeamId)
	} else {
		updatedRule.TeamID = currentRule.TeamID
	}

	if req.OrganizationId != nil {
		if _, err := s.repositories.GetOrganizationByID(ctx, int(*req.OrganizationId)); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "organization with id %d not found", *req.OrganizationId)
		}
		updatedRule.OrganizationID = int(*req.OrganizationId)
	} else {
		updatedRule.OrganizationID = currentRule.OrganizationID
	}

	// Обновляем дополнительные поля
	updatedRule.SCAScanEnabled = getUpdatedBoolValue(req.ScaScanEnabled, currentRule.SCAScanEnabled)
	updatedRule.SASTScanEnabled = getUpdatedBoolValue(req.SastScanEnabled, currentRule.SASTScanEnabled)
	updatedRule.AllowIncrementalScans = getUpdatedBoolValue(req.AllowIncrementalScans, currentRule.AllowIncrementalScans)
	updatedRule.AllowSASTEmptyCode = getUpdatedBoolValue(req.AllowSastEmptyCode, currentRule.AllowSASTEmptyCode)
	updatedRule.ForcedDoOwnSBOM = getUpdatedBoolValue(req.ForcedDoOwnSbom, currentRule.ForcedDoOwnSBOM)
	updatedRule.ActiveBlockingSCA = getUpdatedBoolValue(req.ActiveBlockingSca, currentRule.ActiveBlockingSCA)

	if req.ExcludeDirRegexpQueue != nil {
		updatedRule.ExcludeDirRegexpQueue = req.ExcludeDirRegexpQueue
	} else {
		updatedRule.ExcludeDirRegexpQueue = currentRule.ExcludeDirRegexpQueue
	}

	// Проверяем уникальность
	if existingRule, err := s.repositories.GetScanRuleByComposite(
		ctx,
		updatedRule.ApplicationID,
		updatedRule.TeamID,
		updatedRule.OrganizationID,
	); err == nil && existingRule != nil && existingRule.ID != updatedRule.ID {
		return nil, status.Errorf(
			codes.AlreadyExists,
			"scan rule for application %d, team %d and organization %d already exists",
			updatedRule.ApplicationID,
			updatedRule.TeamID,
			updatedRule.OrganizationID,
		)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check scan rule uniqueness: %v", err)
	}

	if err := s.repositories.UpdateScanRule(ctx, updatedRule); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update scan rule: %v", err)
	}

	return convertScanRuleToProto(updatedRule), nil
}

func (s *Server) DeleteScanRule(ctx context.Context, req *DeleteScanRuleRequest) (*emptypb.Empty, error) {
	if err := s.repositories.DeleteScanRule(ctx, int(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete scan rule: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListScanRules(ctx context.Context, req *ListScanRulesRequest) (*ListScanRulesResponse, error) {
	rules, err := s.repositories.ListScanRules(ctx)
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
		Id:                    int32(rule.ID),
		ApplicationId:         int32(rule.ApplicationID),
		TeamId:                int32(rule.TeamID),
		OrganizationId:        int32(rule.OrganizationID),
		ExcludeDirRegexpQueue: rule.ExcludeDirRegexpQueue,
		ScaScanEnabled:        rule.SCAScanEnabled,
		SastScanEnabled:       rule.SASTScanEnabled,
		AllowIncrementalScans: rule.AllowIncrementalScans,
		AllowSastEmptyCode:    rule.AllowSASTEmptyCode,
		ForcedDoOwnSbom:       rule.ForcedDoOwnSBOM,
		ActiveBlockingSca:     rule.ActiveBlockingSCA,
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
