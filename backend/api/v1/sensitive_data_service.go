package v1

import (
	"context"

	"connectrpc.com/connect"
	"github.com/pkg/errors"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/generated-go/v1/v1connect"
	"github.com/bytebase/bytebase/backend/store"
)

// SensitiveDataService implements the sensitive data service.
type SensitiveDataService struct {
	v1connect.UnimplementedSensitiveDataServiceHandler
	store *store.Store
}

// NewSensitiveDataService creates a new SensitiveDataService.
func NewSensitiveDataService(store *store.Store) *SensitiveDataService {
	return &SensitiveDataService{
		store: store,
	}
}

// CreateSensitiveDataRule creates a sensitive data rule.
func (s *SensitiveDataService) CreateSensitiveDataRule(ctx context.Context, req *connect.Request[v1pb.CreateSensitiveDataRuleRequest]) (*connect.Response[v1pb.SensitiveDataRule], error) {
	// TODO: Implement CreateSensitiveDataRule
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("CreateSensitiveDataRule not implemented"))
}

// ListSensitiveDataRules lists all sensitive data rules.
func (s *SensitiveDataService) ListSensitiveDataRules(ctx context.Context, req *connect.Request[v1pb.ListSensitiveDataRulesRequest]) (*connect.Response[v1pb.ListSensitiveDataRulesResponse], error) {
	// TODO: Implement ListSensitiveDataRules
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("ListSensitiveDataRules not implemented"))
}

// GetSensitiveDataRule gets a sensitive data rule.
func (s *SensitiveDataService) GetSensitiveDataRule(ctx context.Context, req *connect.Request[v1pb.GetSensitiveDataRuleRequest]) (*connect.Response[v1pb.SensitiveDataRule], error) {
	// TODO: Implement GetSensitiveDataRule
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("GetSensitiveDataRule not implemented"))
}

// UpdateSensitiveDataRule updates a sensitive data rule.
func (s *SensitiveDataService) UpdateSensitiveDataRule(ctx context.Context, req *connect.Request[v1pb.UpdateSensitiveDataRuleRequest]) (*connect.Response[v1pb.SensitiveDataRule], error) {
	// TODO: Implement UpdateSensitiveDataRule
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("UpdateSensitiveDataRule not implemented"))
}

// DeleteSensitiveDataRule deletes a sensitive data rule.
func (s *SensitiveDataService) DeleteSensitiveDataRule(ctx context.Context, req *connect.Request[v1pb.DeleteSensitiveDataRuleRequest]) (*connect.Response[v1pb.DeleteSensitiveDataRuleResponse], error) {
	// TODO: Implement DeleteSensitiveDataRule
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("DeleteSensitiveDataRule not implemented"))
}

// CreateSensitiveDataField creates a sensitive data field.
func (s *SensitiveDataService) CreateSensitiveDataField(ctx context.Context, req *connect.Request[v1pb.CreateSensitiveDataFieldRequest]) (*connect.Response[v1pb.SensitiveDataField], error) {
	// TODO: Implement CreateSensitiveDataField
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("CreateSensitiveDataField not implemented"))
}

// ListSensitiveDataFields lists all sensitive data fields.
func (s *SensitiveDataService) ListSensitiveDataFields(ctx context.Context, req *connect.Request[v1pb.ListSensitiveDataFieldsRequest]) (*connect.Response[v1pb.ListSensitiveDataFieldsResponse], error) {
	// TODO: Implement ListSensitiveDataFields
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("ListSensitiveDataFields not implemented"))
}

// GetSensitiveDataField gets a sensitive data field.
func (s *SensitiveDataService) GetSensitiveDataField(ctx context.Context, req *connect.Request[v1pb.GetSensitiveDataFieldRequest]) (*connect.Response[v1pb.SensitiveDataField], error) {
	// TODO: Implement GetSensitiveDataField
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("GetSensitiveDataField not implemented"))
}

// UpdateSensitiveDataField updates a sensitive data field.
func (s *SensitiveDataService) UpdateSensitiveDataField(ctx context.Context, req *connect.Request[v1pb.UpdateSensitiveDataFieldRequest]) (*connect.Response[v1pb.SensitiveDataField], error) {
	// TODO: Implement UpdateSensitiveDataField
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("UpdateSensitiveDataField not implemented"))
}

// DeleteSensitiveDataField deletes a sensitive data field.
func (s *SensitiveDataService) DeleteSensitiveDataField(ctx context.Context, req *connect.Request[v1pb.DeleteSensitiveDataFieldRequest]) (*connect.Response[v1pb.DeleteSensitiveDataFieldResponse], error) {
	// TODO: Implement DeleteSensitiveDataField
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("DeleteSensitiveDataField not implemented"))
}
