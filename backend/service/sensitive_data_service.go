package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bytebase/bytebase/backend/common"
	"github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SensitiveDataService implements the sensitive data service.
type SensitiveDataService struct {
	store store.Store
}

// NewSensitiveDataService creates a new SensitiveDataService.
func NewSensitiveDataService(store store.Store) *SensitiveDataService {
	return &SensitiveDataService{
		store: store,
	}
}

// CreateSensitiveDataRule implements the CreateSensitiveDataRule method.
func (s *SensitiveDataService) CreateSensitiveDataRule(ctx context.Context, request *v1.CreateSensitiveDataRuleRequest) (*v1.SensitiveDataRule, error) {
	// Validate request
	if request.Rule == nil {
		return nil, status.Errorf(codes.InvalidArgument, "rule is required")
	}

	// Check if rule with same name already exists
	_, err := s.store.GetSensitiveDataRuleByName(ctx, request.Rule.Name)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "rule with name %q already exists", request.Rule.Name)
	}
	if err != store.ErrNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check existing rule: %v", err)
	}

	// Create the rule
	createdRule, err := s.store.CreateSensitiveDataRule(ctx, request.Rule)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create rule: %v", err)
	}

	return createdRule, nil
}

// GetSensitiveDataRule implements the GetSensitiveDataRule method.
func (s *SensitiveDataService) GetSensitiveDataRule(ctx context.Context, request *v1.GetSensitiveDataRuleRequest) (*v1.SensitiveDataRule, error) {
	// Validate request
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	// Get the rule
	rule, err := s.store.GetSensitiveDataRuleByName(ctx, request.Name)
	if err == store.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "rule not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rule: %v", err)
	}

	return rule, nil
}

// ListSensitiveDataRules implements the ListSensitiveDataRules method.
func (s *SensitiveDataService) ListSensitiveDataRules(ctx context.Context, request *v1.ListSensitiveDataRulesRequest) (*v1.ListSensitiveDataRulesResponse, error) {
	// Validate request
	if request.Parent == "" {
		return nil, status.Errorf(codes.InvalidArgument, "parent is required")
	}

	// List rules
	rules, err := s.store.ListSensitiveDataRules(ctx, request.Parent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list rules: %v", err)
	}

	return &v1.ListSensitiveDataRulesResponse{
		Rules: rules,
	}, nil
}

// UpdateSensitiveDataRule implements the UpdateSensitiveDataRule method.
func (s *SensitiveDataService) UpdateSensitiveDataRule(ctx context.Context, request *v1.UpdateSensitiveDataRuleRequest) (*v1.SensitiveDataRule, error) {
	// Validate request
	if request.Rule == nil {
		return nil, status.Errorf(codes.InvalidArgument, "rule is required")
	}
	if request.Rule.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "rule name is required")
	}

	// Update the rule
	updatedRule, err := s.store.UpdateSensitiveDataRule(ctx, request.Rule, request.UpdateMask)
	if err == store.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "rule not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update rule: %v", err)
	}

	return updatedRule, nil
}

// DeleteSensitiveDataRule implements the DeleteSensitiveDataRule method.
func (s *SensitiveDataService) DeleteSensitiveDataRule(ctx context.Context, request *v1.DeleteSensitiveDataRuleRequest) (*v1.DeleteSensitiveDataRuleResponse, error) {
	// Validate request
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	// Delete the rule
	err := s.store.DeleteSensitiveDataRule(ctx, request.Name)
	if err == store.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "rule not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete rule: %v", err)
	}

	return &v1.DeleteSensitiveDataRuleResponse{}, nil
}

// CreateApprovalFlow implements the CreateApprovalFlow method.
func (s *SensitiveDataService) CreateApprovalFlow(ctx context.Context, request *v1.CreateApprovalFlowRequest) (*v1.ApprovalFlow, error) {
	// Validate request
	if request.Flow == nil {
		return nil, status.Errorf(codes.InvalidArgument, "flow is required")
	}

	// Check if flow with same name already exists
	_, err := s.store.GetApprovalFlowByName(ctx, request.Flow.Name)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "flow with name %q already exists", request.Flow.Name)
	}
	if err != store.ErrNotFound {
		return nil, status.Errorf(codes.Internal, "failed to check existing flow: %v", err)
	}

	// Create the flow
	createdFlow, err := s.store.CreateApprovalFlow(ctx, request.Flow)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create flow: %v", err)
	}

	return createdFlow, nil
}

// GetApprovalFlow implements the GetApprovalFlow method.
func (s *SensitiveDataService) GetApprovalFlow(ctx context.Context, request *v1.GetApprovalFlowRequest) (*v1.ApprovalFlow, error) {
	// Validate request
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	// Get the flow
	flow, err := s.store.GetApprovalFlowByName(ctx, request.Name)
	if err == store.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "flow not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get flow: %v", err)
	}

	return flow, nil
}

// ListApprovalFlows implements the ListApprovalFlows method.
func (s *SensitiveDataService) ListApprovalFlows(ctx context.Context, request *v1.ListApprovalFlowsRequest) (*v1.ListApprovalFlowsResponse, error) {
	// Validate request
	if request.Parent == "" {
		return nil, status.Errorf(codes.InvalidArgument, "parent is required")
	}

	// List flows
	flows, err := s.store.ListApprovalFlows(ctx, request.Parent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list flows: %v", err)
	}

	return &v1.ListApprovalFlowsResponse{
		Flows: flows,
	}, nil
}

// UpdateApprovalFlow implements the UpdateApprovalFlow method.
func (s *SensitiveDataService) UpdateApprovalFlow(ctx context.Context, request *v1.UpdateApprovalFlowRequest) (*v1.ApprovalFlow, error) {
	// Validate request
	if request.Flow == nil {
		return nil, status.Errorf(codes.InvalidArgument, "flow is required")
	}
	if request.Flow.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "flow name is required")
	}

	// Update the flow
	updatedFlow, err := s.store.UpdateApprovalFlow(ctx, request.Flow, request.UpdateMask)
	if err == store.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "flow not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update flow: %v", err)
	}

	return updatedFlow, nil
}

// DeleteApprovalFlow implements the DeleteApprovalFlow method.
func (s *SensitiveDataService) DeleteApprovalFlow(ctx context.Context, request *v1.DeleteApprovalFlowRequest) (*v1.DeleteApprovalFlowResponse, error) {
	// Validate request
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	// Delete the flow
	err := s.store.DeleteApprovalFlow(ctx, request.Name)
	if err == store.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "flow not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete flow: %v", err)
	}

	return &v1.DeleteApprovalFlowResponse{}, nil
}

// EvaluateSensitiveDataChange implements the EvaluateSensitiveDataChange method.
func (s *SensitiveDataService) EvaluateSensitiveDataChange(ctx context.Context, request *v1.EvaluateSensitiveDataChangeRequest) (*v1.EvaluateSensitiveDataChangeResponse, error) {
	// Validate request
	if request.Parent == "" {
		return nil, status.Errorf(codes.InvalidArgument, "parent is required")
	}
	if request.Change == nil || request.Change.Sql == "" {
		return nil, status.Errorf(codes.InvalidArgument, "change.sql is required")
	}
	if request.Database == nil || request.Database.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "database.name is required")
	}

	// 1. Parse SQL to identify affected tables and fields
	// TODO: Implement proper SQL parsing using the parser plugin
	// For now, we'll extract table and field information from common SQL patterns
	tables, fields, err := s.extractTablesAndFields(request.Change.Sql)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to parse SQL: %v", err)
	}

	// 2. Query all sensitive data rules for the project
	rules, err := s.store.ListSensitiveDataRules(ctx, request.Parent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list sensitive data rules: %v", err)
	}

	// 3. Match fields against sensitive data rules
	var matchingRules []*v1.SensitiveDataRule
	var highestLevel v1.SensitiveDataLevel = v1.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED

	for _, table := range tables {
		for _, field := range fields {
			for _, rule := range rules {
				// Check if rule matches the table and field
				if s.isRuleMatch(rule, table, field) {
					matchingRules = append(matchingRules, rule)
					// Update highest sensitive level
					if rule.Level > highestLevel {
						highestLevel = rule.Level
					}
				}
			}
		}
	}

	// 4. Determine required approval flow based on highest sensitive level
	var requiredApprovalFlow *v1.ApprovalFlow
	if highestLevel != v1.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		// Get approval flow for the highest sensitive level
		approvalFlows, err := s.store.ListApprovalFlows(ctx, request.Parent)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to list approval flows: %v", err)
		}

		// Find the approval flow that matches the highest sensitive level
		for _, flow := range approvalFlows {
			if flow.Level == highestLevel {
				requiredApprovalFlow = flow
				break
			}
		}

		// If no matching approval flow found, use default based on level
		if requiredApprovalFlow == nil {
			requiredApprovalFlow = s.getDefaultApprovalFlow(request.Parent, highestLevel)
		}
	}

	// 5. Build the response
	response := &v1.EvaluateSensitiveDataChangeResponse{
		AffectsSensitiveData: len(matchingRules) > 0,
		HighestSensitiveLevel: highestLevel,
		MatchingRules:         matchingRules,
		RequiredApprovalFlow:  requiredApprovalFlow,
	}

	return response, nil
}

// extractTablesAndFields extracts tables and fields from SQL statement
func (s *SensitiveDataService) extractTablesAndFields(sql string) ([]string, []string, error) {
	// TODO: Implement proper SQL parsing using the parser plugin
	// For now, we'll use a simple regex-based approach to extract common patterns
	var tables []string
	var fields []string

	// Extract tables from FROM clause
	fromRegex := regexp.MustCompile(`FROM\s+([a-zA-Z0-9_\.]+)`)
	matches := fromRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	// Extract fields from SELECT clause
	selectRegex := regexp.MustCompile(`SELECT\s+([^FROM]+)`)
	selectMatches := selectRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range selectMatches {
		if len(match) > 1 {
			// Split fields by commas and trim whitespace
			fieldList := strings.Split(match[1], ",")
			for _, field := range fieldList {
				field = strings.TrimSpace(field)
				// Skip wildcards
				if field != "*" {
					// Extract field name (handle aliases)
					fieldName := strings.Split(field, " ")[0]
					fields = append(fields, fieldName)
				}
			}
		}
	}

	// Extract fields from UPDATE clause
	updateRegex := regexp.MustCompile(`UPDATE\s+([a-zA-Z0-9_\.]+)\s+SET\s+([^WHERE]+)`)
	updateMatches := updateRegex.FindAllStringSubmatch(sql, -1)
	for _, match := range updateMatches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
		if len(match) > 2 {
			// Split fields by commas and trim whitespace
			fieldList := strings.Split(match[2], ",")
			for _, field := range fieldList {
				field = strings.TrimSpace(field)
				// Extract field name (before = sign)
				fieldName := strings.Split(field, "=")[0]
				fieldName = strings.TrimSpace(fieldName)
				fields = append(fields, fieldName)
			}
		}
	}

	// Remove duplicates
	tableMap := make(map[string]bool)
	uniqueTables := []string
	for _, table := range tables {
		if !tableMap[table] {
			tableMap[table] = true
			uniqueTables = append(uniqueTables, table)
		}
	}

	fieldMap := make(map[string]bool)
	uniqueFields := []string
	for _, field := range fields {
		if !fieldMap[field] {
			fieldMap[field] = true
			uniqueFields = append(uniqueFields, field)
		}
	}

	return uniqueTables, uniqueFields, nil
}

// isRuleMatch checks if a sensitive data rule matches a table and field
func (s *SensitiveDataService) isRuleMatch(rule *v1.SensitiveDataRule, table, field string) bool {
	// Check table match if specified
	if rule.Table != "" {
		// Rule table format: projects/{project}/databases/{database}/tables/{table}
		// Extract just the table name from the rule
		ruleTableParts := strings.Split(rule.Table, "/")
		ruleTableName := ruleTableParts[len(ruleTableParts)-1]
		if ruleTableName != table {
			return false
		}
	}

	// Check field pattern match if specified
	if rule.FieldPattern != "" {
		// Handle wildcard patterns (e.g., "email*", "*_password")
		pattern := strings.ReplaceAll(rule.FieldPattern, "*", ".*")
		pattern = "^" + pattern + "$"
		match, _ := regexp.MatchString(pattern, field)
		if !match {
			return false
		}
	}

	// If no table or field pattern specified, match all fields
	return true
}

// getDefaultApprovalFlow returns a default approval flow based on sensitive level
func (s *SensitiveDataService) getDefaultApprovalFlow(parent string, level v1.SensitiveDataLevel) *v1.ApprovalFlow {
	var steps []*v1.ApprovalStep
	var flowName string

	switch level {
	case v1.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_HIGH:
		// High sensitivity requires 2-level approval
		flowName = "high-sensitivity-flow"
		steps = []*v1.ApprovalStep{
			{
				Step:         1,
				Approvers:    []string{"roles/security-admin"},
				ApprovalType: v1.ApprovalStep_APPROVAL_TYPE_ALL,
			},
			{
				Step:         2,
				Approvers:    []string{"roles/dba"},
				ApprovalType: v1.ApprovalStep_APPROVAL_TYPE_ALL,
			},
		}
	case v1.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_MEDIUM:
		// Medium sensitivity requires 1-level approval
		flowName = "medium-sensitivity-flow"
		steps = []*v1.ApprovalStep{
			{
				Step:         1,
				Approvers:    []string{"roles/dba"},
				ApprovalType: v1.ApprovalStep_APPROVAL_TYPE_ALL,
			},
		}
	case v1.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_LOW:
		// Low sensitivity requires 1-level approval (self-approval)
		flowName = "low-sensitivity-flow"
		steps = []*v1.ApprovalStep{
			{
				Step:         1,
				Approvers:    []string{"self"},
				ApprovalType: v1.ApprovalStep_APPROVAL_TYPE_ALL,
			},
		}
	default:
		return nil
	}

	return &v1.ApprovalFlow{
		Name:        fmt.Sprintf("%s/approvalFlows/%s", parent, flowName),
		Title:       fmt.Sprintf("%s Approval Flow", strings.Title(strings.ToLower(level.String()))),
		Description: fmt.Sprintf("Default approval flow for %s sensitivity data", strings.ToLower(level.String())),
		Level:       level,
		Steps:       steps,
	}
}

// ApproveIssue implements the ApproveIssue method.
func (s *SensitiveDataService) ApproveIssue(ctx context.Context, request *v1.ApproveIssueRequest) (*v1.ApproveIssueResponse, error) {
	// Validate request
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if request.Approver == "" {
		return nil, status.Errorf(codes.InvalidArgument, "approver is required")
	}

	// Get the current approval flow for the issue
	approvalFlow, err := s.store.GetApprovalFlow(ctx, request.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "approval flow not found: %v", err)
	}

	// Check if the approver is authorized for the current step
	currentStep := approvalFlow.CurrentStep
	if currentStep < 1 || currentStep > len(approvalFlow.Steps) {
		return nil, status.Errorf(codes.FailedPrecondition, "invalid current step: %d", currentStep)
	}

	step := approvalFlow.Steps[currentStep-1]
	isAuthorized := false
	for _, approver := range step.Approvers {
		if approver == request.Approver || approver == "self" {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		return nil, status.Errorf(codes.PermissionDenied, "approver %s is not authorized to approve this step", request.Approver)
	}

	// Create approval record
	approvalRecord := &v1.ApprovalRecord{
		Name:         fmt.Sprintf("%s/approvalRecords/%s", request.Name, time.Now().Format("20060102150405")),
		Issue:        request.Name,
		Step:         currentStep,
		Approver:     request.Approver,
		Status:       v1.ApprovalRecord_STATUS_APPROVED,
		ApprovalTime: time.Now().Format(time.RFC3339),
		Comment:      request.Comment,
	}

	if err := s.store.CreateApprovalRecord(ctx, approvalRecord); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create approval record: %v", err)
	}

	// Move to next step or complete approval
	if currentStep == len(approvalFlow.Steps) {
		// All steps approved, complete the approval flow
		approvalFlow.Status = v1.ApprovalFlow_STATUS_APPROVED
		approvalFlow.CompletedTime = time.Now().Format(time.RFC3339)
	} else {
		// Move to next step
		approvalFlow.CurrentStep = currentStep + 1
	}

	// Update approval flow
	if err := s.store.UpdateApprovalFlow(ctx, approvalFlow); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update approval flow: %v", err)
	}

	return &v1.ApproveIssueResponse{
		ApprovalFlow: approvalFlow,
		ApprovalRecord: approvalRecord,
	}, nil
}

// RejectIssue implements the RejectIssue method.
func (s *SensitiveDataService) RejectIssue(ctx context.Context, request *v1.RejectIssueRequest) (*v1.RejectIssueResponse, error) {
	// Validate request
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if request.Approver == "" {
		return nil, status.Errorf(codes.InvalidArgument, "approver is required")
	}

	// Get the current approval flow for the issue
	approvalFlow, err := s.store.GetApprovalFlow(ctx, request.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "approval flow not found: %v", err)
	}

	// Check if the approval flow is already completed or rejected
	if approvalFlow.Status == v1.ApprovalFlow_STATUS_APPROVED || approvalFlow.Status == v1.ApprovalFlow_STATUS_REJECTED {
		return nil, status.Errorf(codes.FailedPrecondition, "approval flow is already %s", approvalFlow.Status)
	}

	// Check if the approver is authorized for the current step
	currentStep := approvalFlow.CurrentStep
	if currentStep < 1 || currentStep > len(approvalFlow.Steps) {
		return nil, status.Errorf(codes.FailedPrecondition, "invalid current step: %d", currentStep)
	}

	step := approvalFlow.Steps[currentStep-1]
	isAuthorized := false
	for _, approver := range step.Approvers {
		if approver == request.Approver || approver == "self" {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		return nil, status.Errorf(codes.PermissionDenied, "approver %s is not authorized to reject this step", request.Approver)
	}

	// Create approval record
	approvalRecord := &v1.ApprovalRecord{
		Name:         fmt.Sprintf("%s/approvalRecords/%s", request.Name, time.Now().Format("20060102150405")),
		Issue:        request.Name,
		Step:         currentStep,
		Approver:     request.Approver,
		Status:       v1.ApprovalRecord_STATUS_REJECTED,
		ApprovalTime: time.Now().Format(time.RFC3339),
		Comment:      request.Comment,
	}

	if err := s.store.CreateApprovalRecord(ctx, approvalRecord); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create approval record: %v", err)
	}

	// Reject the approval flow
	approvalFlow.Status = v1.ApprovalFlow_STATUS_REJECTED
	approvalFlow.CompletedTime = time.Now().Format(time.RFC3339)

	// Update approval flow
	if err := s.store.UpdateApprovalFlow(ctx, approvalFlow); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update approval flow: %v", err)
	}

	return &v1.RejectIssueResponse{
		ApprovalFlow: approvalFlow,
		ApprovalRecord: approvalRecord,
	}, nil
}