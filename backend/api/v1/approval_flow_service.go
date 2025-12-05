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

// ApprovalFlowService implements the approval flow service.
type ApprovalFlowService struct {
	v1pb.UnimplementedApprovalFlowServiceServer

	store *store.Store
}

// NewApprovalFlowService creates a new approval flow service.
func NewApprovalFlowService(store *store.Store) *ApprovalFlowService {
	return &ApprovalFlowService{
		store: store,
	}
}

// ListApprovalFlows implements the ListApprovalFlows method.
func (s *ApprovalFlowService) ListApprovalFlows(ctx context.Context, req *v1pb.ListApprovalFlowsRequest) (*v1pb.ListApprovalFlowsResponse, error) {
	// Parse the filter
	var filter *store.ListApprovalFlowsFilter
	if req.Filter != nil {
		filter = &store.ListApprovalFlowsFilter{
			SensitivityLevel: req.Filter.SensitivityLevel,
			Enabled:           req.Filter.Enabled,
		}
	}

	// Retrieve the approval flows from the store
	flows, err := s.store.ListApprovalFlows(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve approval flows: %v", err)
	}

	// Convert the store flows to proto flows
	var protoFlows []*v1pb.ApprovalFlow
	for _, flow := range flows {
		protoFlow := &v1pb.ApprovalFlow{
			Name:              fmt.Sprintf("approvalFlows/%d", flow.ID),
			Uid:               fmt.Sprintf("%d", flow.ID),
			Title:             flow.Title,
			Description:       flow.Description,
			SensitivityLevel:  flow.SensitivityLevel,
			Enabled:           flow.Enabled,
			CreateTime:        timestampFromInt64(flow.CreateTime),
			UpdateTime:        timestampFromInt64(flow.UpdateTime),
		}

		// Convert the store nodes to proto nodes
		var protoNodes []*v1pb.ApprovalNode
		for _, node := range flow.Nodes {
			protoNode := &v1pb.ApprovalNode{
				Name:                fmt.Sprintf("approvalNodes/%d", node.ID),
				Uid:                 fmt.Sprintf("%d", node.ID),
				Title:               node.Title,
				Description:         node.Description,
				ApproverType:        node.ApproverType,
				ApproverIds:         node.ApproverIDs,
				RequiredApprovals:   node.RequiredApprovals,
				CreateTime:          timestampFromInt64(node.CreateTime),
				UpdateTime:          timestampFromInt64(node.UpdateTime),
			}

			protoNodes = append(protoNodes, protoNode)
		}
		protoFlow.Nodes = protoNodes

		protoFlows = append(protoFlows, protoFlow)
	}

	// Create the response
	resp := &v1pb.ListApprovalFlowsResponse{
		ApprovalFlows: protoFlows,
	}

	return resp, nil
}

// GetApprovalFlow implements the GetApprovalFlow method.
func (s *ApprovalFlowService) GetApprovalFlow(ctx context.Context, req *v1pb.GetApprovalFlowRequest) (*v1pb.GetApprovalFlowResponse, error) {
	// Parse the flow ID from the name
	id, err := parseResourceID(req.Name, "approvalFlows")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse approval flow name: %v", err)
	}

	// Retrieve the approval flow from the store
	flow, err := s.store.GetApprovalFlow(ctx, id)
	if err != nil {
		if err == common.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "approval flow not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve approval flow: %v", err)
	}

	// Convert the store flow to a proto flow
	protoFlow := &v1pb.ApprovalFlow{
		Name:              fmt.Sprintf("approvalFlows/%d", flow.ID),
		Uid:               fmt.Sprintf("%d", flow.ID),
		Title:             flow.Title,
		Description:       flow.Description,
		SensitivityLevel:  flow.SensitivityLevel,
		Enabled:           flow.Enabled,
		CreateTime:        timestampFromInt64(flow.CreateTime),
		UpdateTime:        timestampFromInt64(flow.UpdateTime),
	}

	// Convert the store nodes to proto nodes
	var protoNodes []*v1pb.ApprovalNode
	for _, node := range flow.Nodes {
		protoNode := &v1pb.ApprovalNode{
			Name:                fmt.Sprintf("approvalNodes/%d", node.ID),
			Uid:                 fmt.Sprintf("%d", node.ID),
			Title:               node.Title,
			Description:         node.Description,
			ApproverType:        node.ApproverType,
			ApproverIds:         node.ApproverIDs,
			RequiredApprovals:   node.RequiredApprovals,
			CreateTime:          timestampFromInt64(node.CreateTime),
			UpdateTime:          timestampFromInt64(node.UpdateTime),
		}

		protoNodes = append(protoNodes, protoNode)
	}
	protoFlow.Nodes = protoNodes

	// Create the response
	resp := &v1pb.GetApprovalFlowResponse{
		ApprovalFlow: protoFlow,
	}

	return resp, nil
}

// CreateApprovalFlow implements the CreateApprovalFlow method.
func (s *ApprovalFlowService) CreateApprovalFlow(ctx context.Context, req *v1pb.CreateApprovalFlowRequest) (*v1pb.CreateApprovalFlowResponse, error) {
	// Parse the parent resource (should be "instances/{instance}" or similar)
	// For now, we'll ignore the parent and just create the flow

	// Validate the request body
	if req.ApprovalFlow == nil {
		return nil, status.Errorf(codes.InvalidArgument, "approval flow is required")
	}
	if req.ApprovalFlow.Title == "" {
		return nil, status.Errorf(codes.InvalidArgument, "title is required")
	}
	if req.ApprovalFlow.SensitivityLevel == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "sensitivity level is required")
	}
	if len(req.ApprovalFlow.Nodes) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "at least one node is required")
	}

	// Convert the proto flow to a store flow
	storeFlow := &store.ApprovalFlow{
		Title:             req.ApprovalFlow.Title,
		Description:       req.ApprovalFlow.Description,
		SensitivityLevel:  req.ApprovalFlow.SensitivityLevel,
		Enabled:           req.ApprovalFlow.Enabled,
	}

	// Convert the proto nodes to store nodes
	var storeNodes []*store.ApprovalNode
	for _, node := range req.ApprovalFlow.Nodes {
		storeNode := &store.ApprovalNode{
			Title:               node.Title,
			Description:         node.Description,
			ApproverType:        node.ApproverType,
			ApproverIDs:         node.ApproverIds,
			RequiredApprovals:   node.RequiredApprovals,
		}
		storeNodes = append(storeNodes, storeNode)
	}
	storeFlow.Nodes = storeNodes

	// Set the creator ID from the context
	creatorID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user ID: %v", err)
	}
	storeFlow.CreatorID = creatorID
	storeFlow.UpdaterID = creatorID

	// Create the approval flow in the store
	createdFlow, err := s.store.CreateApprovalFlow(ctx, storeFlow)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create approval flow: %v", err)
	}

	// Convert the created store flow to a proto flow
	protoFlow := &v1pb.ApprovalFlow{
		Name:              fmt.Sprintf("approvalFlows/%d", createdFlow.ID),
		Uid:               fmt.Sprintf("%d", createdFlow.ID),
		Title:             createdFlow.Title,
		Description:       createdFlow.Description,
		SensitivityLevel:  createdFlow.SensitivityLevel,
		Enabled:           createdFlow.Enabled,
		CreateTime:        timestampFromInt64(createdFlow.CreateTime),
		UpdateTime:        timestampFromInt64(createdFlow.UpdateTime),
	}

	// Convert the created store nodes to proto nodes
	var createdProtoNodes []*v1pb.ApprovalNode
	for _, node := range createdFlow.Nodes {
		createdProtoNode := &v1pb.ApprovalNode{
			Name:                fmt.Sprintf("approvalNodes/%d", node.ID),
			Uid:                 fmt.Sprintf("%d", node.ID),
			Title:               node.Title,
			Description:         node.Description,
			ApproverType:        node.ApproverType,
			ApproverIds:         node.ApproverIDs,
			RequiredApprovals:   node.RequiredApprovals,
			CreateTime:          timestampFromInt64(node.CreateTime),
			UpdateTime:          timestampFromInt64(node.UpdateTime),
		}

		createdProtoNodes = append(createdProtoNodes, createdProtoNode)
	}
	protoFlow.Nodes = createdProtoNodes

	// Create the response
	resp := &v1pb.CreateApprovalFlowResponse{
		ApprovalFlow: protoFlow,
	}

	return resp, nil
}

// UpdateApprovalFlow implements the UpdateApprovalFlow method.
func (s *ApprovalFlowService) UpdateApprovalFlow(ctx context.Context, req *v1pb.UpdateApprovalFlowRequest) (*v1pb.UpdateApprovalFlowResponse, error) {
	// Parse the flow ID from the name
	id, err := parseResourceID(req.ApprovalFlow.Name, "approvalFlows")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse approval flow name: %v", err)
	}

	// Retrieve the existing approval flow from the store
	existingFlow, err := s.store.GetApprovalFlow(ctx, id)
	if err != nil {
		if err == common.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "approval flow not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve approval flow: %v", err)
	}

	// Apply the update mask to determine which fields to update
	// For now, we'll just update all fields from the request
	// In a real implementation, we would use the update mask to only update the specified fields
	updateFlow := &store.ApprovalFlow{
		ID:                 existingFlow.ID,
		Title:              req.ApprovalFlow.Title,
		Description:        req.ApprovalFlow.Description,
		SensitivityLevel:   req.ApprovalFlow.SensitivityLevel,
		Enabled:            req.ApprovalFlow.Enabled,
	}

	// Convert the proto nodes to store nodes
	var storeNodes []*store.ApprovalNode
	for _, node := range req.ApprovalFlow.Nodes {
		storeNode := &store.ApprovalNode{
			Title:               node.Title,
			Description:         node.Description,
			ApproverType:        node.ApproverType,
			ApproverIDs:         node.ApproverIds,
			RequiredApprovals:   node.RequiredApprovals,
		}
		storeNodes = append(storeNodes, storeNode)
	}
	updateFlow.Nodes = storeNodes

	// Set the updater ID from the context
	updaterID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current user ID: %v", err)
	}
	updateFlow.UpdaterID = updaterID

	// Update the approval flow in the store
	updatedFlow, err := s.store.UpdateApprovalFlow(ctx, updateFlow)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update approval flow: %v", err)
	}

	// Convert the updated store flow to a proto flow
	protoFlow := &v1pb.ApprovalFlow{
		Name:              fmt.Sprintf("approvalFlows/%d", updatedFlow.ID),
		Uid:               fmt.Sprintf("%d", updatedFlow.ID),
		Title:             updatedFlow.Title,
		Description:       updatedFlow.Description,
		SensitivityLevel:  updatedFlow.SensitivityLevel,
		Enabled:           updatedFlow.Enabled,
		CreateTime:        timestampFromInt64(updatedFlow.CreateTime),
		UpdateTime:        timestampFromInt64(updatedFlow.UpdateTime),
	}

	// Convert the updated store nodes to proto nodes
	var updatedProtoNodes []*v1pb.ApprovalNode
	for _, node := range updatedFlow.Nodes {
		updatedProtoNode := &v1pb.ApprovalNode{
			Name:                fmt.Sprintf("approvalNodes/%d", node.ID),
			Uid:                 fmt.Sprintf("%d", node.ID),
			Title:               node.Title,
			Description:         node.Description,
			ApproverType:        node.ApproverType,
			ApproverIds:         node.ApproverIDs,
			RequiredApprovals:   node.RequiredApprovals,
			CreateTime:          timestampFromInt64(node.CreateTime),
			UpdateTime:          timestampFromInt64(node.UpdateTime),
		}

		updatedProtoNodes = append(updatedProtoNodes, updatedProtoNode)
	}
	protoFlow.Nodes = updatedProtoNodes

	// Create the response
	resp := &v1pb.UpdateApprovalFlowResponse{
		ApprovalFlow: protoFlow,
	}

	return resp, nil
}

// DeleteApprovalFlow implements the DeleteApprovalFlow method.
func (s *ApprovalFlowService) DeleteApprovalFlow(ctx context.Context, req *v1pb.DeleteApprovalFlowRequest) (*v1pb.DeleteApprovalFlowResponse, error) {
	// Parse the flow ID from the name
	id, err := parseResourceID(req.Name, "approvalFlows")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse approval flow name: %v", err)
	}

	// Delete the approval flow from the store
	if err := s.store.DeleteApprovalFlow(ctx, id); err != nil {
		if err == common.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "approval flow not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to delete approval flow: %v", err)
	}

	// Create the response
	resp := &v1pb.DeleteApprovalFlowResponse{}

	return resp, nil
}

// ListApprovalFlowExecutions implements the ListApprovalFlowExecutions method.
func (s *ApprovalFlowService) ListApprovalFlowExecutions(ctx context.Context, req *v1pb.ListApprovalFlowExecutionsRequest) (*v1pb.ListApprovalFlowExecutionsResponse, error) {
	// Parse the filter
	var filter *store.ListApprovalFlowExecutionsFilter
	if req.Filter != nil {
		filter = &store.ListApprovalFlowExecutionsFilter{
			FlowID:             req.Filter.FlowId,
			IssueID:            req.Filter.IssueId,
			SensitivityLevel:   req.Filter.SensitivityLevel,
			Status:              req.Filter.Status,
		}
	}

	// Retrieve the approval flow executions from the store
	executions, err := s.store.ListApprovalFlowExecutions(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve approval flow executions: %v", err)
	}

	// Convert the store executions to proto executions
	var protoExecutions []*v1pb.ApprovalFlowExecution
	for _, execution := range executions {
		protoExecution := &v1pb.ApprovalFlowExecution{
			Name:              fmt.Sprintf("approvalFlowExecutions/%d", execution.ID),
			Uid:               fmt.Sprintf("%d", execution.ID),
			FlowId:            execution.FlowID,
			IssueId:           execution.IssueID,
			SensitivityLevel:  execution.SensitivityLevel,
			Status:            execution.Status,
			CreateTime:        timestampFromInt64(execution.CreateTime),
			UpdateTime:        timestampFromInt64(execution.UpdateTime),
		}

		// Convert the store node executions to proto node executions
		var protoNodeExecutions []*v1pb.ApprovalNodeExecution
		for _, nodeExecution := range execution.NodeExecutions {
			protoNodeExecution := &v1pb.ApprovalNodeExecution{
				Name:                fmt.Sprintf("approvalNodeExecutions/%d", nodeExecution.ID),
				Uid:                 fmt.Sprintf("%d", nodeExecution.ID),
				NodeId:              nodeExecution.NodeID,
				Status:              nodeExecution.Status,
				CreateTime:          timestampFromInt64(nodeExecution.CreateTime),
				UpdateTime:          timestampFromInt64(nodeExecution.UpdateTime),
			}

			// Convert the store approvals to proto approvals
			var protoApprovals []*v1pb.Approval
			for _, approval := range nodeExecution.Approvals {
				protoApproval := &v1pb.Approval{
					Name:        fmt.Sprintf("approvals/%d", approval.ID),
					Uid:         fmt.Sprintf("%d", approval.ID),
					UserId:      approval.UserID,
					Status:      approval.Status,
					Comment:     approval.Comment,
					CreateTime:  timestampFromInt64(approval.CreateTime),
				}

				protoApprovals = append(protoApprovals, protoApproval)
			}
			protoNodeExecution.Approvals = protoApprovals

			protoNodeExecutions = append(protoNodeExecutions, protoNodeExecution)
		}
		protoExecution.NodeExecutions = protoNodeExecutions

		protoExecutions = append(protoExecutions, protoExecution)
	}

	// Create the response
	resp := &v1pb.ListApprovalFlowExecutionsResponse{
		ApprovalFlowExecutions: protoExecutions,
	}

	return resp, nil
}

// GetApprovalFlowExecution implements the GetApprovalFlowExecution method.
func (s *ApprovalFlowService) GetApprovalFlowExecution(ctx context.Context, req *v1pb.GetApprovalFlowExecutionRequest) (*v1pb.GetApprovalFlowExecutionResponse, error) {
	// Parse the execution ID from the name
	id, err := parseResourceID(req.Name, "approvalFlowExecutions")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse approval flow execution name: %v", err)
	}

	// Retrieve the approval flow execution from the store
	execution, err := s.store.GetApprovalFlowExecution(ctx, id)
	if err != nil {
		if err == common.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "approval flow execution not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "failed to retrieve approval flow execution: %v", err)
	}

	// Convert the store execution to a proto execution
	protoExecution := &v1pb.ApprovalFlowExecution{
		Name:              fmt.Sprintf("approvalFlowExecutions/%d", execution.ID),
		Uid:               fmt.Sprintf("%d", execution.ID),
		FlowId:            execution.FlowID,
		IssueId:           execution.IssueID,
		SensitivityLevel:  execution.SensitivityLevel,
		Status:            execution.Status,
		CreateTime:        timestampFromInt64(execution.CreateTime),
		UpdateTime:        timestampFromInt64(execution.UpdateTime),
	}

	// Convert the store node executions to proto node executions
	var protoNodeExecutions []*v1pb.ApprovalNodeExecution
	for _, nodeExecution := range execution.NodeExecutions {
		protoNodeExecution := &v1pb.ApprovalNodeExecution{
			Name:                fmt.Sprintf("approvalNodeExecutions/%d", nodeExecution.ID),
			Uid:                 fmt.Sprintf("%d", nodeExecution.ID),
			NodeId:              nodeExecution.NodeID,
			Status:              nodeExecution.Status,
			CreateTime:          timestampFromInt64(nodeExecution.CreateTime),
			UpdateTime:          timestampFromInt64(nodeExecution.UpdateTime),
		}

		// Convert the store approvals to proto approvals
		var protoApprovals []*v1pb.Approval
		for _, approval := range nodeExecution.Approvals {
			protoApproval := &v1pb.Approval{
				Name:        fmt.Sprintf("approvals/%d", approval.ID),
				Uid:         fmt.Sprintf("%d", approval.ID),
				UserId:      approval.UserID,
				Status:      approval.Status,
				Comment:     approval.Comment,
				CreateTime:  timestampFromInt64(approval.CreateTime),
			}

			protoApprovals = append(protoApprovals, protoApproval)
		}
		protoNodeExecution.Approvals = protoApprovals

		protoNodeExecutions = append(protoNodeExecutions, protoNodeExecution)
	}
	protoExecution.NodeExecutions = protoNodeExecutions

	// Create the response
	resp := &v1pb.GetApprovalFlowExecutionResponse{
		ApprovalFlowExecution: protoExecution,
	}

	return resp, nil
}

// CreateApproval implements the CreateApproval method.
func (s *ApprovalFlowService) CreateApproval(ctx context.Context, req *v1pb.CreateApprovalRequest) (*v1pb.CreateApprovalResponse, error) {
	// Parse the parent resource (should be "approvalNodeExecutions/{nodeExecution}" or similar)
	// For now, we'll ignore the parent and just create the approval

	// Validate the request body
	if req.Approval == nil {
		return nil, status.Errorf(codes.InvalidArgument, "approval is required")
	}
	if req.Approval.NodeExecutionId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "node execution ID is required")
	}
	if req.Approval.UserId <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "user ID is required")
	}
	if req.Approval.Status == storepb.ApprovalStatus_APPROVAL_STATUS_UNSPECIFIED {
		return nil, status.Errorf(codes.InvalidArgument, "status is required")
	}

	// Convert the proto approval to a store approval
	storeApproval := &store.Approval{
		NodeExecutionID: req.Approval.NodeExecutionId,
		UserID:           req.Approval.UserId,
		Status:           req.Approval.Status,
		Comment:          req.Approval.Comment,
	}

	// Create the approval in the store
	createdApproval, err := s.store.CreateApproval(ctx, storeApproval)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create approval: %v", err)
	}

	// Convert the created store approval to a proto approval
	protoApproval := &v1pb.Approval{
		Name:        fmt.Sprintf("approvals/%d", createdApproval.ID),
		Uid:         fmt.Sprintf("%d", createdApproval.ID),
		NodeExecutionId: createdApproval.NodeExecutionID,
		UserId:      createdApproval.UserID,
		Status:      createdApproval.Status,
		Comment:     createdApproval.Comment,
		CreateTime:  timestampFromInt64(createdApproval.CreateTime),
	}

	// Create the response
	resp := &v1pb.CreateApprovalResponse{
		Approval: protoApproval,
	}

	return resp, nil
}
