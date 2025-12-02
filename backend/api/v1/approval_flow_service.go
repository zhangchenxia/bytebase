package v1

import (
	"context"

	"connectrpc.com/connect"
	"github.com/pkg/errors"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/generated-go/v1/v1connect"
	"github.com/bytebase/bytebase/backend/store"
)

// ApprovalFlowService implements the approval flow service.
type ApprovalFlowService struct {
	v1connect.UnimplementedApprovalFlowServiceHandler
	store *store.Store
}

// NewApprovalFlowService creates a new ApprovalFlowService.
func NewApprovalFlowService(store *store.Store) *ApprovalFlowService {
	return &ApprovalFlowService{
		store: store,
	}
}

// GetApprovalConfig gets the approval configuration.
func (s *ApprovalFlowService) GetApprovalConfig(ctx context.Context, req *connect.Request[v1pb.GetApprovalConfigRequest]) (*connect.Response[v1pb.ApprovalConfig], error) {
	// TODO: Implement GetApprovalConfig
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("GetApprovalConfig not implemented"))
}

// UpdateApprovalConfig updates the approval configuration.
func (s *ApprovalFlowService) UpdateApprovalConfig(ctx context.Context, req *connect.Request[v1pb.UpdateApprovalConfigRequest]) (*connect.Response[v1pb.ApprovalConfig], error) {
	// TODO: Implement UpdateApprovalConfig
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("UpdateApprovalConfig not implemented"))
}

// ListApprovalRecords lists all approval records.
func (s *ApprovalFlowService) ListApprovalRecords(ctx context.Context, req *connect.Request[v1pb.ListApprovalRecordsRequest]) (*connect.Response[v1pb.ListApprovalRecordsResponse], error) {
	// TODO: Implement ListApprovalRecords
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("ListApprovalRecords not implemented"))
}

// GetApprovalRecord gets a specific approval record.
func (s *ApprovalFlowService) GetApprovalRecord(ctx context.Context, req *connect.Request[v1pb.GetApprovalRecordRequest]) (*connect.Response[v1pb.ApprovalRecord], error) {
	// TODO: Implement GetApprovalRecord
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("GetApprovalRecord not implemented"))
}

// ApproveRecord approves an approval record.
func (s *ApprovalFlowService) ApproveRecord(ctx context.Context, req *connect.Request[v1pb.ApproveRecordRequest]) (*connect.Response[v1pb.ApprovalRecord], error) {
	// TODO: Implement ApproveRecord
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("ApproveRecord not implemented"))
}

// RejectRecord rejects an approval record.
func (s *ApprovalFlowService) RejectRecord(ctx context.Context, req *connect.Request[v1pb.RejectRecordRequest]) (*connect.Response[v1pb.ApprovalRecord], error) {
	// TODO: Implement RejectRecord
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("RejectRecord not implemented"))
}
