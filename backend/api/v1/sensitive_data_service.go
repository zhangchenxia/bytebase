package v1

import (
	"context"
	"fmt"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/store"
	"github.com/bytebase/bytebase/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SensitiveDataService implements the sensitive data service.
type SensitiveDataService struct {
	v1pb.UnimplementedSensitiveDataServiceServer

	store *store.Store
}

// NewSensitiveDataService creates a new sensitive data service.
func NewSensitiveDataService(store *store.Store) *SensitiveDataService {
	return &SensitiveDataService{
		store: store,
	}
}

// ListSensitiveDataRules implements the ListSensitiveDataRules method.
func (s *SensitiveDataService) ListSensitiveDataRules(ctx context.Context, req *v1pb.ListSensitiveDataRulesRequest) (*v1pb.ListSensitiveDataRulesResponse, error) {
	// Parse the filter
	var filter *store.ListSensitiveDataRulesFilter
	if req.Filter != nil {
		filter = &store.ListSensitiveDataRulesFilter{
			Level:       req.Filter.Level,
			TableName:   req.Filter.TableName,
		}
	}

	// Retrieve the sensitive data rules from the store
	rules, err := s.store.ListSensitiveDataRules(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve sensitive data rules: %v", err)
	}

	// Convert the store rules to proto rules
	var protoRules []*v1pb.SensitiveDataRule
	for _, rule := range rules {
		protoRule := &v1pb.SensitiveDataRule{
			Name:        fmt.Sprintf("sensitiveDataRules/%d", rule.ID),
			Uid:         fmt.Sprintf("%d", rule.ID),
			Title:       rule.Title,
			Description: rule.Description,
			Level:       rule.Level,
			Enabled:     rule.Enabled,
			TableName:   rule.TableName,
			CreateTime:  timestampFromInt64(rule.CreateTime),
			UpdateTime:  timestampFromInt64(rule.UpdateTime),
		}

		// Convert the store fields to proto fields
		var protoFields []*v1pb.SensitiveDataField
		for _, field := range rule.Fields {
			protoField := &v1pb.SensitiveDataField{
				FieldName: field.FieldName,
				DataType:  field.DataType,
				Regex:     field.Regex,
			}
			protoFields = append(protoFields, protoField)
		}
		protoRule.Fields = protoFields

		protoRules = append(protoRules, protoRule)
	}

	// Create the response
	resp := &v1pb.ListSensitiveDataRulesResponse{
		SensitiveDataRules: protoRules,
	}

	return resp, nil
}

// GetSensitiveDataRule implements the GetSensitiveDataRule method.
func (s *SensitiveDataService) GetSensitiveDataRule(ctx context.Context, req *v1pb.GetSensitiveDataRuleRequest) (*v1pb.GetSensitiveDataRuleResponse, error) {
	// Parse the rule ID from the name
	id, err := parseResourceID(req.Name, "sensitiveDataRules")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse sensitive data rule name: %v", err)
	}

	// Retrieve the sensitive data rule from the store
	rule, err := s.store.GetSensitiveDataRule(ctx, id)
	if err != nil {
		if err == common.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "sensitive data rule not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve sensitive data rule: %v", err)
	}

	// Convert the store rule to a proto rule
	protoRule := &v1pb.SensitiveDataRule{
		Name:        fmt.Sprintf("sensitiveDataRules/%d", rule.ID),
		Uid:         fmt.Sprintf("%d", rule.ID),
		Title:       rule.Title,
		Description: rule.Description,
		Level:       rule.Level,
		Enabled:     rule.Enabled,
		TableName:   rule.TableName,
		CreateTime:  timestampFromInt64(rule.CreateTime),
		UpdateTime:  timestampFromInt64(rule.UpdateTime),
	}

	// Convert the store fields to proto fields
	var protoFields []*v1pb.SensitiveDataField
	for _, field := range rule.Fields {
		protoField := &v1pb.SensitiveDataField{
			FieldName: field.FieldName,
			DataType:  field.DataType,
			Regex:     field.Regex,
		}
		protoFields = append(protoFields, protoField)
	}
	protoRule.Fields = protoFields

	// Create the response
	resp := &v1pb.GetSensitiveDataRuleResponse{
		SensitiveDataRule: protoRule,
	}

	return resp, nil
}

// CreateSensitiveDataRule implements the CreateSensitiveDataRule method.
func (s *SensitiveDataService) CreateSensitiveDataRule(ctx context.Context, req *v1pb.CreateSensitiveDataRuleRequest) (*v1pb.CreateSensitiveDataRuleResponse, error) {
	// Parse the parent resource (should be "instances/{instance}" or similar)
	// For now, we'll ignore the parent and just create the rule

	// Validate the request body
	if req.SensitiveDataRule == nil {
		return nil, status.Errorf(codes.InvalidArgument, "sensitive data rule is required")
	}
	if req.SensitiveDataRule.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}
	if req.SensitiveDataRule.Level == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "sensitivity level is required")
	}
	if len(req.SensitiveDataRule.Fields) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "at least one field is required")
	}

	// Convert the proto rule to a store rule
	storeRule := &store.SensitiveDataRule{
		Title:       req.SensitiveDataRule.Title,
		Description: req.SensitiveDataRule.Description,
		Level:       req.SensitiveDataRule.Level,
		Enabled:     req.SensitiveDataRule.Enabled,
		TableName:   req.SensitiveDataRule.TableName,
	}

	// Convert the proto fields to store fields
	var storeFields []*store.SensitiveDataField
	for _, field := range req.SensitiveDataRule.Fields {
		storeField := &store.SensitiveDataField{
			FieldName: field.FieldName,
			DataType:  field.DataType,
			Regex:     field.Regex,
		}
		storeFields = append(storeFields, storeField)
	}
	storeRule.Fields = storeFields

	// Set the creator ID from the context
	creatorID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user ID: %v", err)
	}
	storeRule.CreatorID = creatorID
	storeRule.UpdaterID = creatorID

	// Create the sensitive data rule in the store
	createdRule, err := s.store.CreateSensitiveDataRule(ctx, storeRule)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create sensitive data rule: %v", err)
	}

	// Convert the created store rule to a proto rule
	protoRule := &v1pb.SensitiveDataRule{
		Name:        fmt.Sprintf("sensitiveDataRules/%d", createdRule.ID),
		Uid:         fmt.Sprintf("%d", createdRule.ID),
		Title:       createdRule.Title,
		Description: createdRule.Description,
		Level:       createdRule.Level,
		Enabled:     createdRule.Enabled,
		TableName:   createdRule.TableName,
		CreateTime:  timestampFromInt64(createdRule.CreateTime),
		UpdateTime:  timestampFromInt64(createdRule.UpdateTime),
	}

	// Convert the created store fields to proto fields
	var createdProtoFields []*v1pb.SensitiveDataField
	for _, field := range createdRule.Fields {
		createdProtoField := &v1pb.SensitiveDataField{
			FieldName: field.FieldName,
			DataType:  field.DataType,
			Regex:     field.Regex,
		}
		createdProtoFields = append(createdProtoFields, createdProtoField)
	}
	protoRule.Fields = createdProtoFields

	// Create the response
	resp := &v1pb.CreateSensitiveDataRuleResponse{
		SensitiveDataRule: protoRule,
	}

	return resp, nil
}

// UpdateSensitiveDataRule implements the UpdateSensitiveDataRule method.
func (s *SensitiveDataService) UpdateSensitiveDataRule(ctx context.Context, req *v1pb.UpdateSensitiveDataRuleRequest) (*v1pb.UpdateSensitiveDataRuleResponse, error) {
	// Parse the rule ID from the name
	id, err := parseResourceID(req.SensitiveDataRule.Name, "sensitiveDataRules")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse sensitive data rule name: %v", err)
	}

	// Retrieve the existing sensitive data rule from the store
	existingRule, err := s.store.GetSensitiveDataRule(ctx, id)
	if err != nil {
		if err == common.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "sensitive data rule not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve sensitive data rule: %v", err)
	}

	// Apply the update mask to determine which fields to update
	// For now, we'll just update all fields from the request
	// In a real implementation, we would use the update mask to only update the specified fields
	updateRule := &store.SensitiveDataRule{
		ID:          existingRule.ID,
		Title:       req.SensitiveDataRule.Title,
		Description: req.SensitiveDataRule.Description,
		Level:       req.SensitiveDataRule.Level,
		Enabled:     req.SensitiveDataRule.Enabled,
		TableName:   req.SensitiveDataRule.TableName,
	}

	// Convert the proto fields to store fields
	var storeFields []*store.SensitiveDataField
	for _, field := range req.SensitiveDataRule.Fields {
		storeField := &store.SensitiveDataField{
			FieldName: field.FieldName,
			DataType:  field.DataType,
			Regex:     field.Regex,
		}
		storeFields = append(storeFields, storeField)
	}
	updateRule.Fields = storeFields

	// Set the updater ID from the context
	updaterID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user ID: %v", err)
	}
	updateRule.UpdaterID = updaterID

	// Update the sensitive data rule in the store
	updatedRule, err := s.store.UpdateSensitiveDataRule(ctx, updateRule)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update sensitive data rule: %v", err)
	}

	// Convert the updated store rule to a proto rule
	protoRule := &v1pb.SensitiveDataRule{
		Name:        fmt.Sprintf("sensitiveDataRules/%d", updatedRule.ID),
		Uid:         fmt.Sprintf("%d", updatedRule.ID),
		Title:       updatedRule.Title,
		Description: updatedRule.Description,
		Level:       updatedRule.Level,
		Enabled:     updatedRule.Enabled,
		TableName:   updatedRule.TableName,
		CreateTime:  timestampFromInt64(updatedRule.CreateTime),
		UpdateTime:  timestampFromInt64(updatedRule.UpdateTime),
	}

	// Convert the updated store fields to proto fields
	var updatedProtoFields []*v1pb.SensitiveDataField
	for _, field := range updatedRule.Fields {
		updatedProtoField := &v1pb.SensitiveDataField{
			FieldName: field.FieldName,
			DataType:  field.DataType,
			Regex:     field.Regex,
		}
		updatedProtoFields = append(updatedProtoFields, updatedProtoField)
	}
	protoRule.Fields = updatedProtoFields

	// Create the response
	resp := &v1pb.UpdateSensitiveDataRuleResponse{
		SensitiveDataRule: protoRule,
	}

	return resp, nil
}

// DeleteSensitiveDataRule implements the DeleteSensitiveDataRule method.
func (s *SensitiveDataService) DeleteSensitiveDataRule(ctx context.Context, req *v1pb.DeleteSensitiveDataRuleRequest) (*v1pb.DeleteSensitiveDataRuleResponse, error) {
	// Parse the rule ID from the name
	id, err := parseResourceID(req.Name, "sensitiveDataRules")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse sensitive data rule name: %v", err)
	}

	// Delete the sensitive data rule from the store
	if err := s.store.DeleteSensitiveDataRule(ctx, id); err != nil {
		if err == common.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "sensitive data rule not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete sensitive data rule: %v", err)
	}

	// Create the response
	resp := &v1pb.DeleteSensitiveDataRuleResponse{}

	return resp, nil
}
