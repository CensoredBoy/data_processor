package data_processor

import (
	"context"
	"data_processor/internal/common"
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	existingRule, err := s.scanRuleRepo.GetScanRuleByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan rule: %v", err)
	}
	if existingRule == nil {
		return nil, status.Errorf(codes.NotFound, "scan rule not found")
	}

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

	var extraComment string
	if req.Comment != nil {
		extraComment = *req.Comment
	}
	commentBody := buildScanRuleComment(existingRule, updatedRule, extraComment)

	var userIDPtr *int
	if req.UserId != nil {
		idVal := int(*req.UserId)
		userIDPtr = &idVal
	}

	if commentBody != nil && (userIDPtr == nil || *userIDPtr == 0) {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required when comment is present")
	}

	_, err = s.scanRuleRepo.UpdateScanRule(ctx, updatedRule, userIDPtr, commentBody)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update scan rule: %v", err)
	}

	// Получаем обновленное правило для возврата
	rule, err := s.scanRuleRepo.GetScanRuleByID(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated scan rule: %v", err)
	}

	return convertScanRuleToProto(rule), nil
}

func (s *Server) GetScanRuleWithComments(ctx context.Context, req *GetScanRuleRequest) (*ScanRuleWithComments, error) {
	rule, err := s.scanRuleRepo.GetScanRuleWithComments(ctx, int(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get scan rule with comments: %v", err)
	}
	if rule == nil {
		return nil, status.Errorf(codes.NotFound, "scan rule not found")
	}

	resp := &ScanRuleWithComments{
		Rule: convertScanRuleToProto(rule),
	}

	for _, comment := range rule.Comments {
		resp.Comments = append(resp.Comments, convertCommentToProto(comment))
	}

	return resp, nil
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
	protoRule.LatestCommentId = common.Int32PtrFromIntPtr(rule.LatestCommentID)
	return protoRule
}

func convertCommentToProto(comment *common.Comment) *Comment {
	protoComment := &Comment{
		Id:         int32(comment.ID),
		ScanRuleId: int32(comment.ScanRuleID),
		UserId:     int32(comment.UserID),
		Text:       comment.Text,
		CreatedAt:  timestamppb.New(comment.CreatedAt),
	}
	protoComment.PreviousCommentId = common.Int32PtrFromIntPtr(comment.PreviousCommentID)
	return protoComment
}

func buildScanRuleComment(oldRule, newRule *common.ScanRule, rawComment string) *string {
	diffs := buildScanRuleDiff(oldRule, newRule)
	trimmed := strings.TrimSpace(rawComment)

	lines := make([]string, 0, len(diffs)+1)
	lines = append(lines, diffs...)

	if trimmed != "" {
		lines = append(lines, fmt.Sprintf("Comment: %s", trimmed))
	}

	if len(lines) == 0 {
		return nil
	}

	comment := strings.Join(lines, "\n")
	return &comment
}

func buildScanRuleDiff(oldRule, newRule *common.ScanRule) []string {
	var diffs []string

	compareIntPtr := func(field string, oldVal, newVal *int) {
		if !equalIntPtr(oldVal, newVal) {
			diffs = append(diffs, fmt.Sprintf("%s: %s -> %s", field, formatIntPtr(oldVal), formatIntPtr(newVal)))
		}
	}

	compareBoolPtr := func(field string, oldVal, newVal *bool) {
		if !equalBoolPtr(oldVal, newVal) {
			diffs = append(diffs, fmt.Sprintf("%s: %s -> %s", field, formatBoolPtr(oldVal), formatBoolPtr(newVal)))
		}
	}

	compareStringPtr := func(field string, oldVal, newVal *string) {
		if !equalStringPtr(oldVal, newVal) {
			diffs = append(diffs, fmt.Sprintf("%s: %s -> %s", field, formatStringPtr(oldVal), formatStringPtr(newVal)))
		}
	}

	formatSlice := func(v []string) string {
		if v == nil {
			return "null"
		}
		if len(v) == 0 {
			return ""
		}
		return strings.Join(v, ",")
	}

	if !reflect.DeepEqual(oldRule.ExcludeDirRegexpQueue, newRule.ExcludeDirRegexpQueue) {
		diffs = append(diffs, fmt.Sprintf("%s: [%s] -> [%s]", "exclude_dir_regexp_queue", formatSlice(oldRule.ExcludeDirRegexpQueue), formatSlice(newRule.ExcludeDirRegexpQueue)))
	}

	compareIntPtr("application_id", oldRule.ApplicationID, newRule.ApplicationID)
	compareIntPtr("team_id", oldRule.TeamID, newRule.TeamID)
	compareIntPtr("organization_id", oldRule.OrganizationID, newRule.OrganizationID)
	compareBoolPtr("sca_scan_enabled", oldRule.SCAScanEnabled, newRule.SCAScanEnabled)
	compareBoolPtr("sast_scan_enabled", oldRule.SASTScanEnabled, newRule.SASTScanEnabled)
	compareStringPtr("application_postfix", oldRule.ApplicationPostfix, newRule.ApplicationPostfix)
	compareBoolPtr("allow_unsafe_ext_distribs", oldRule.AllowUnsafeExtDistribs, newRule.AllowUnsafeExtDistribs)
	compareBoolPtr("ignore_repository_membership", oldRule.IgnoreRepositoryMembership, newRule.IgnoreRepositoryMembership)
	compareBoolPtr("allow_incremental_scans", oldRule.AllowIncrementalScans, newRule.AllowIncrementalScans)
	compareBoolPtr("allow_sast_empty_code", oldRule.AllowSASTEmptyCode, newRule.AllowSASTEmptyCode)
	compareBoolPtr("forced_do_own_sbom", oldRule.ForcedDoOwnSBOM, newRule.ForcedDoOwnSBOM)
	compareBoolPtr("active_blocking_sca", oldRule.ActiveBlockingSCA, newRule.ActiveBlockingSCA)

	return diffs
}

func equalIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalBoolPtr(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalStringPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func formatIntPtr(v *int) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprintf("%d", *v)
}

func formatBoolPtr(v *bool) string {
	if v == nil {
		return "null"
	}
	return fmt.Sprintf("%t", *v)
}

func formatStringPtr(v *string) string {
	if v == nil {
		return "null"
	}
	return *v
}
