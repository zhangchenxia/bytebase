package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/pkg/errors"
)

// ApprovalFlow is the API message for approval flow.
type ApprovalFlow struct {
	ID                int32
	Title             string
	Description       string
	SensitivityLevel  storepb.SensitiveDataLevel
	Nodes             []*ApprovalNode
	Enabled           bool
	CreatorID         int32
	UpdaterID         int32
	CreateTime        int64
	UpdateTime        int64
}

// ApprovalNode is the API message for approval node.
type ApprovalNode struct {
	ID                  int32
	FlowID              int32
	Title               string
	Description         string
	ApproverType        storepb.ApproverType
	ApproverIDs         []string
	RequiredApprovals   int32
	CreatorID           int32
	UpdaterID           int32
	CreateTime          int64
	UpdateTime          int64
}

// ApprovalFlowExecution is the API message for approval flow execution.
type ApprovalFlowExecution struct {
	ID                int32
	FlowID            int32
	IssueID           int32
	SensitivityLevel  storepb.SensitiveDataLevel
	Status            storepb.ApprovalFlowExecutionStatus
	NodeExecutions    []*ApprovalNodeExecution
	CreatorID         int32
	UpdaterID         int32
	CreateTime        int64
	UpdateTime        int64
}

// ApprovalNodeExecution is the API message for approval node execution.
type ApprovalNodeExecution struct {
	ID                int32
	ExecutionID       int32
	NodeID            int32
	Status            storepb.ApprovalNodeExecutionStatus
	Approvals         []*Approval
	CreatorID         int32
	UpdaterID         int32
	CreateTime        int64
	UpdateTime        int64
}

// Approval is the API message for approval.
type Approval struct {
	ID          int32
	NodeExecutionID  int32
	UserID      int32
	Status      storepb.ApprovalStatus
	Comment     string
	CreateTime  int64
}

// CreateApprovalFlow creates a new approval flow.
func (s *Store) CreateApprovalFlow(ctx context.Context, create *ApprovalFlow) (*ApprovalFlow, error) {
	if create.Title == "" {
		return nil, errors.New("title is required")
	}
	if create.SensitivityLevel == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return nil, errors.New("sensitivity level is required")
	}
	if len(create.Nodes) == 0 {
		return nil, errors.New("at least one node is required")
	}

	// Check if there's already an approval flow for the same sensitivity level
	var existingID int32
	query := `SELECT id FROM approval_flow WHERE sensitivity_level = ?`
	if err := s.db.QueryRowContext(ctx, query, create.SensitivityLevel).Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to check existing approval flow")
	}
	if existingID > 0 {
		return nil, errors.New("an approval flow for the same sensitivity level already exists")
	}

	// Start a transaction
	 tx, err := s.db.BeginTx(ctx, nil)
	 if err != nil {
		 return nil, errors.Wrap(err, "failed to begin transaction")
	 }
	 defer func() {
		 if r := recover(); r != nil {
			 tx.Rollback()
			 panic(r)
		 } else if err != nil {
			 tx.Rollback()
		 } else {
			 err = tx.Commit()
		 }
	 }()

	// Insert the new approval flow
	query = `INSERT INTO approval_flow (title, description, sensitivity_level, enabled, creator_id, updater_id) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := tx.ExecContext(ctx, query, create.Title, create.Description, create.SensitivityLevel, create.Enabled, create.CreatorID, create.UpdaterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create approval flow")
	}

	flowID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert ID for approval flow")
	}

	// Insert the approval nodes
	for _, node := range create.Nodes {
		// Convert approver IDs to a comma-separated string
		approverIDsStr := strings.Join(node.ApproverIDs, ",")

		// Insert the node
		query = `INSERT INTO approval_node (flow_id, title, description, approver_type, approver_ids, required_approvals, creator_id, updater_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.ExecContext(ctx, query, flowID, node.Title, node.Description, node.ApproverType, approverIDsStr, node.RequiredApprovals, create.CreatorID, create.UpdaterID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create approval node")
		}
	}

	// Retrieve the newly created approval flow with nodes
	flow, err := s.getApprovalFlowWithNodes(ctx, int32(flowID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve newly created approval flow")
	}

	return flow, nil
}

// ListApprovalFlows lists all approval flows.
func (s *Store) ListApprovalFlows(ctx context.Context, filter *ListApprovalFlowsFilter) ([]*ApprovalFlow, error) {
	query := `SELECT id, title, description, sensitivity_level, enabled, creator_id, updater_id, create_time, update_time FROM approval_flow`
	args := []interface{}{}

	whereClauses := []string{}
	if filter != nil {
		if filter.SensitivityLevel != storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
			whereClauses = append(whereClauses, "sensitivity_level = ?")
			args = append(args, filter.SensitivityLevel)
		}
		if filter.Enabled != nil {
			whereClauses = append(whereClauses, "enabled = ?")
			args = append(args, *filter.Enabled)
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY create_time DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list approval flows")
	}
	defer rows.Close()

	var flows []*ApprovalFlow
	for rows.Next() {
		var flow ApprovalFlow
		if err := rows.Scan(&flow.ID, &flow.Title, &flow.Description, &flow.SensitivityLevel, &flow.Enabled, &flow.CreatorID, &flow.UpdaterID, &flow.CreateTime, &flow.UpdateTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan approval flow")
		}

		// Retrieve the nodes for this flow
		nodes, err := s.getApprovalNodes(ctx, flow.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get approval nodes")
		}
		flow.Nodes = nodes

		flows = append(flows, &flow)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over approval flows")
	}

	return flows, nil
}

// GetApprovalFlow gets an approval flow by ID.
func (s *Store) GetApprovalFlow(ctx context.Context, id int32) (*ApprovalFlow, error) {
	if id <= 0 {
		return nil, errors.New("invalid approval flow ID")
	}

	return s.getApprovalFlowWithNodes(ctx, id)
}

// UpdateApprovalFlow updates an approval flow.
func (s *Store) UpdateApprovalFlow(ctx context.Context, update *ApprovalFlow) (*ApprovalFlow, error) {
	if update.ID <= 0 {
		return nil, errors.New("invalid approval flow ID")
	}

	// Check if the flow exists
	_, err := s.getApprovalFlowWithNodes(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	// Check if there's another flow with the same sensitivity level
	var existingID int32
	query := `SELECT id FROM approval_flow WHERE sensitivity_level = ? AND id != ?`
	if err := s.db.QueryRowContext(ctx, query, update.SensitivityLevel, update.ID).Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to check existing approval flow")
	}
	if existingID > 0 {
		return nil, errors.New("an approval flow for the same sensitivity level already exists")
	}

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Update the approval flow
	query = `UPDATE approval_flow SET title = ?, description = ?, sensitivity_level = ?, enabled = ?, updater_id = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, query, update.Title, update.Description, update.SensitivityLevel, update.Enabled, update.UpdaterID, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update approval flow")
	}

	// Delete existing nodes
	query = `DELETE FROM approval_node WHERE flow_id = ?`
	_, err = tx.ExecContext(ctx, query, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to delete existing approval nodes")
	}

	// Insert the updated nodes
	for _, node := range update.Nodes {
		// Convert approver IDs to a comma-separated string
		approverIDsStr := strings.Join(node.ApproverIDs, ",")

		// Insert the node
		query = `INSERT INTO approval_node (flow_id, title, description, approver_type, approver_ids, required_approvals, creator_id, updater_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.ExecContext(ctx, query, update.ID, node.Title, node.Description, node.ApproverType, approverIDsStr, node.RequiredApprovals, update.CreatorID, update.UpdaterID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create approval node")
		}
	}

	// Retrieve the updated approval flow with nodes
	flow, err := s.getApprovalFlowWithNodes(ctx, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve updated approval flow")
	}

	return flow, nil
}

// DeleteApprovalFlow deletes an approval flow by ID.
func (s *Store) DeleteApprovalFlow(ctx context.Context, id int32) error {
	if id <= 0 {
		return errors.New("invalid approval flow ID")
	}

	// Check if the flow exists
	_, err := s.getApprovalFlowWithNodes(ctx, id)
	if err != nil {
		return err
	}

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Delete all nodes for this flow
	query := `DELETE FROM approval_node WHERE flow_id = ?`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete approval nodes")
	}

	// Delete the flow
	query = `DELETE FROM approval_flow WHERE id = ?`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete approval flow")
	}

	return nil
}

// ListApprovalFlowsFilter is the filter for listing approval flows.
type ListApprovalFlowsFilter struct {
	SensitivityLevel  storepb.SensitiveDataLevel
	Enabled            *bool
}

// getApprovalFlowWithNodes gets an approval flow by ID with its nodes.
func (s *Store) getApprovalFlowWithNodes(ctx context.Context, id int32) (*ApprovalFlow, error) {
	query := `SELECT id, title, description, sensitivity_level, enabled, creator_id, updater_id, create_time, update_time FROM approval_flow WHERE id = ?`
	var flow ApprovalFlow
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&flow.ID, &flow.Title, &flow.Description, &flow.SensitivityLevel, &flow.Enabled, &flow.CreatorID, &flow.UpdaterID, &flow.CreateTime, &flow.UpdateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("approval flow not found")
		}
		return nil, errors.Wrap(err, "failed to get approval flow")
	}

	// Retrieve the nodes for this flow
	nodes, err := s.getApprovalNodes(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get approval nodes")
	}
	flow.Nodes = nodes

	return &flow, nil
}

// getApprovalNodes gets all approval nodes for a given flow ID.
func (s *Store) getApprovalNodes(ctx context.Context, flowID int32) ([]*ApprovalNode, error) {
	query := `SELECT id, flow_id, title, description, approver_type, approver_ids, required_approvals, creator_id, updater_id, create_time, update_time FROM approval_node WHERE flow_id = ? ORDER BY id`

	rows, err := s.db.QueryContext(ctx, query, flowID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list approval nodes")
	}
	defer rows.Close()

	var nodes []*ApprovalNode
	for rows.Next() {
		var node ApprovalNode
		var approverIDsStr string
		if err := rows.Scan(&node.ID, &node.FlowID, &node.Title, &node.Description, &node.ApproverType, &approverIDsStr, &node.RequiredApprovals, &node.CreatorID, &node.UpdaterID, &node.CreateTime, &node.UpdateTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan approval node")
		}

		// Convert comma-separated string to slice of strings
		if approverIDsStr != "" {
			node.ApproverIDs = strings.Split(approverIDsStr, ",")
		} else {
			node.ApproverIDs = []string{}
		}

		nodes = append(nodes, &node)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over approval nodes")
	}

	return nodes, nil
}

// CreateApprovalFlowExecution creates a new approval flow execution.
func (s *Store) CreateApprovalFlowExecution(ctx context.Context, create *ApprovalFlowExecution) (*ApprovalFlowExecution, error) {
	if create.FlowID <= 0 {
		return nil, errors.New("flow ID is required")
	}
	if create.IssueID <= 0 {
		return nil, errors.New("issue ID is required")
	}
	if create.SensitivityLevel == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return nil, errors.New("sensitivity level is required")
	}
	if create.Status == storepb.ApprovalFlowExecutionStatus_APPROVAL_FLOW_EXECUTION_STATUS_UNSPECIFIED {
		return nil, errors.New("status is required")
	}

	// Check if the flow exists
	_, err := s.getApprovalFlowWithNodes(ctx, create.FlowID)
	if err != nil {
		return nil, err
	}

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Insert the new approval flow execution
	query := `INSERT INTO approval_flow_execution (flow_id, issue_id, sensitivity_level, status, creator_id, updater_id) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := tx.ExecContext(ctx, query, create.FlowID, create.IssueID, create.SensitivityLevel, create.Status, create.CreatorID, create.UpdaterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create approval flow execution")
	}

	executionID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert ID for approval flow execution")
	}

	// Retrieve the nodes for the associated flow
	nodes, err := s.getApprovalNodes(ctx, create.FlowID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get approval nodes")
	}

	// Insert the approval node executions
	for _, node := range nodes {
		// Insert the node execution
		query = `INSERT INTO approval_node_execution (execution_id, node_id, status, creator_id, updater_id) VALUES (?, ?, ?, ?, ?)`
		status := storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_PENDING
		if len(nodes) == 1 || node.ID == nodes[0].ID {
			// Set the first node to in progress
			status = storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_IN_PROGRESS
		}
		_, err = tx.ExecContext(ctx, query, executionID, node.ID, status, create.CreatorID, create.UpdaterID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create approval node execution")
		}
	}

	// Retrieve the newly created approval flow execution with node executions
	execution, err := s.getApprovalFlowExecutionWithNodeExecutions(ctx, int32(executionID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve newly created approval flow execution")
	}

	return execution, nil
}

// ListApprovalFlowExecutions lists all approval flow executions.
func (s *Store) ListApprovalFlowExecutions(ctx context.Context, filter *ListApprovalFlowExecutionsFilter) ([]*ApprovalFlowExecution, error) {
	query := `SELECT id, flow_id, issue_id, sensitivity_level, status, creator_id, updater_id, create_time, update_time FROM approval_flow_execution`
	args := []interface{}{}

	whereClauses := []string{}
	if filter != nil {
		if filter.FlowID > 0 {
			whereClauses = append(whereClauses, "flow_id = ?")
			args = append(args, filter.FlowID)
		}
		if filter.IssueID > 0 {
			whereClauses = append(whereClauses, "issue_id = ?")
			args = append(args, filter.IssueID)
		}
		if filter.SensitivityLevel != storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
			whereClauses = append(whereClauses, "sensitivity_level = ?")
			args = append(args, filter.SensitivityLevel)
		}
		if filter.Status != storepb.ApprovalFlowExecutionStatus_APPROVAL_FLOW_EXECUTION_STATUS_UNSPECIFIED {
			whereClauses = append(whereClauses, "status = ?")
			args = append(args, filter.Status)
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY create_time DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list approval flow executions")
	}
	defer rows.Close()

	var executions []*ApprovalFlowExecution
	for rows.Next() {
		var execution ApprovalFlowExecution
		if err := rows.Scan(&execution.ID, &execution.FlowID, &execution.IssueID, &execution.SensitivityLevel, &execution.Status, &execution.CreatorID, &execution.UpdaterID, &execution.CreateTime, &execution.UpdateTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan approval flow execution")
		}

		// Retrieve the node executions for this execution
		nodeExecutions, err := s.getApprovalNodeExecutions(ctx, execution.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get approval node executions")
		}
		execution.NodeExecutions = nodeExecutions

		executions = append(executions, &execution)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over approval flow executions")
	}

	return executions, nil
}

// GetApprovalFlowExecution gets an approval flow execution by ID.
func (s *Store) GetApprovalFlowExecution(ctx context.Context, id int32) (*ApprovalFlowExecution, error) {
	if id <= 0 {
		return nil, errors.New("invalid approval flow execution ID")
	}

	return s.getApprovalFlowExecutionWithNodeExecutions(ctx, id)
}

// UpdateApprovalFlowExecution updates an approval flow execution.
func (s *Store) UpdateApprovalFlowExecution(ctx context.Context, update *ApprovalFlowExecution) (*ApprovalFlowExecution, error) {
	if update.ID <= 0 {
		return nil, errors.New("invalid approval flow execution ID")
	}

	// Check if the execution exists
	_, err := s.getApprovalFlowExecutionWithNodeExecutions(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	// Update the approval flow execution
	query := `UPDATE approval_flow_execution SET status = ?, updater_id = ? WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, update.Status, update.UpdaterID, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update approval flow execution")
	}

	// Retrieve the updated approval flow execution with node executions
	execution, err := s.getApprovalFlowExecutionWithNodeExecutions(ctx, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve updated approval flow execution")
	}

	return execution, nil
}

// ListApprovalFlowExecutionsFilter is the filter for listing approval flow executions.
type ListApprovalFlowExecutionsFilter struct {
	FlowID             int32
	IssueID            int32
	SensitivityLevel   storepb.SensitiveDataLevel
	Status              storepb.ApprovalFlowExecutionStatus
}

// getApprovalFlowExecutionWithNodeExecutions gets an approval flow execution by ID with its node executions.
func (s *Store) getApprovalFlowExecutionWithNodeExecutions(ctx context.Context, id int32) (*ApprovalFlowExecution, error) {
	query := `SELECT id, flow_id, issue_id, sensitivity_level, status, creator_id, updater_id, create_time, update_time FROM approval_flow_execution WHERE id = ?`
	var execution ApprovalFlowExecution
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&execution.ID, &execution.FlowID, &execution.IssueID, &execution.SensitivityLevel, &execution.Status, &execution.CreatorID, &execution.UpdaterID, &execution.CreateTime, &execution.UpdateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("approval flow execution not found")
		}
		return nil, errors.Wrap(err, "failed to get approval flow execution")
	}

	// Retrieve the node executions for this execution
	nodeExecutions, err := s.getApprovalNodeExecutions(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get approval node executions")
	}
	execution.NodeExecutions = nodeExecutions

	return &execution, nil
}

// getApprovalNodeExecutions gets all approval node executions for a given execution ID.
func (s *Store) getApprovalNodeExecutions(ctx context.Context, executionID int32) ([]*ApprovalNodeExecution, error) {
	query := `SELECT id, execution_id, node_id, status, creator_id, updater_id, create_time, update_time FROM approval_node_execution WHERE execution_id = ? ORDER BY id`

	rows, err := s.db.QueryContext(ctx, query, executionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list approval node executions")
	}
	defer rows.Close()

	var nodeExecutions []*ApprovalNodeExecution
	for rows.Next() {
		var nodeExecution ApprovalNodeExecution
		if err := rows.Scan(&nodeExecution.ID, &nodeExecution.ExecutionID, &nodeExecution.NodeID, &nodeExecution.Status, &nodeExecution.CreatorID, &nodeExecution.UpdaterID, &nodeExecution.CreateTime, &nodeExecution.UpdateTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan approval node execution")
		}

		// Retrieve the approvals for this node execution
		approvals, err := s.getApprovals(ctx, nodeExecution.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get approvals")
		}
		nodeExecution.Approvals = approvals

		nodeExecutions = append(nodeExecutions, &nodeExecution)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over approval node executions")
	}

	return nodeExecutions, nil
}

// CreateApproval creates a new approval.
func (s *Store) CreateApproval(ctx context.Context, create *Approval) (*Approval, error) {
	if create.NodeExecutionID <= 0 {
		return nil, errors.New("node execution ID is required")
	}
	if create.UserID <= 0 {
		return nil, errors.New("user ID is required")
	}
	if create.Status == storepb.ApprovalStatus_APPROVAL_STATUS_UNSPECIFIED {
		return nil, errors.New("status is required")
	}

	// Check if the node execution exists
	_, err := s.getApprovalNodeExecution(ctx, create.NodeExecutionID)
	if err != nil {
		return nil, err
	}

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Insert the new approval
	query := `INSERT INTO approval (node_execution_id, user_id, status, comment) VALUES (?, ?, ?, ?)`
	result, err := tx.ExecContext(ctx, query, create.NodeExecutionID, create.UserID, create.Status, create.Comment)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create approval")
	}

	approvalID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert ID for approval")
	}

	// Retrieve the node execution to update its status
	nodeExecution, err := s.getApprovalNodeExecution(ctx, create.NodeExecutionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get approval node execution")
	}

	// Retrieve all approvals for this node execution
	approvals, err := s.getApprovals(ctx, create.NodeExecutionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get approvals")
	}

	// Add the new approval to the list
	var newApproval Approval
	if err := tx.QueryRowContext(ctx, `SELECT id, node_execution_id, user_id, status, comment, create_time FROM approval WHERE id = ?`, approvalID).Scan(&newApproval.ID, &newApproval.NodeExecutionID, &newApproval.UserID, &newApproval.Status, &newApproval.Comment, &newApproval.CreateTime); err != nil {
		return nil, errors.Wrap(err, "failed to get new approval")
	}
	approvals = append(approvals, &newApproval)

	// Update the node execution status based on the approvals
	var nodeExecutionStatus storepb.ApprovalNodeExecutionStatus
	var approvedCount int32
	var rejectedCount int32
	for _, approval := range approvals {
		if approval.Status == storepb.ApprovalStatus_APPROVED {
			approvedCount++
		} else if approval.Status == storepb.ApprovalStatus_REJECTED {
			rejectedCount++
		}
	}

	// Get the required approvals for this node
	var requiredApprovals int32
	query = `SELECT required_approvals FROM approval_node WHERE id = ?`
	if err := tx.QueryRowContext(ctx, query, nodeExecution.NodeID).Scan(&requiredApprovals); err != nil {
		return nil, errors.Wrap(err, "failed to get required approvals for node")
	}

	if rejectedCount > 0 {
		// If any approval is rejected, the node is rejected
		nodeExecutionStatus = storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_REJECTED
	} else if approvedCount >= requiredApprovals {
		// If enough approvals are received, the node is approved
		nodeExecutionStatus = storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_APPROVED
	} else {
		// Otherwise, the node remains in progress
		nodeExecutionStatus = storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_IN_PROGRESS
	}

	// Update the node execution status
	query = `UPDATE approval_node_execution SET status = ? WHERE id = ?`
	_, err = tx.ExecContext(ctx, query, nodeExecutionStatus, create.NodeExecutionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update approval node execution status")
	}

	// If the node is approved, move to the next node
	if nodeExecutionStatus == storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_APPROVED {
		// Get the next node in the flow
		var nextNodeID int32
		query = `SELECT id FROM approval_node WHERE flow_id = (SELECT flow_id FROM approval_flow_execution WHERE id = ?) AND id > ? ORDER BY id LIMIT 1`
		if err := tx.QueryRowContext(ctx, query, nodeExecution.ExecutionID, nodeExecution.NodeID).Scan(&nextNodeID); err != nil && err != sql.ErrNoRows {
			return nil, errors.Wrap(err, "failed to get next approval node")
		}

		if nextNodeID > 0 {
			// Find the node execution for the next node
			var nextNodeExecutionID int32
			query = `SELECT id FROM approval_node_execution WHERE execution_id = ? AND node_id = ?`
			if err := tx.QueryRowContext(ctx, query, nodeExecution.ExecutionID, nextNodeID).Scan(&nextNodeExecutionID); err != nil {
				return nil, errors.Wrap(err, "failed to get next approval node execution")
			}

			// Update the next node execution status to in progress
			query = `UPDATE approval_node_execution SET status = ? WHERE id = ?`
			_, err = tx.ExecContext(ctx, query, storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_IN_PROGRESS, nextNodeExecutionID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to update next approval node execution status")
			}
		} else {
			// If there are no more nodes, update the flow execution status to approved
			query = `UPDATE approval_flow_execution SET status = ? WHERE id = ?`
			_, err = tx.ExecContext(ctx, query, storepb.ApprovalFlowExecutionStatus_APPROVED, nodeExecution.ExecutionID)
			if err != nil {
				return nil, errors.Wrap(err, "failed to update approval flow execution status")
			}
		}
	} else if nodeExecutionStatus == storepb.ApprovalNodeExecutionStatus_APPROVAL_NODE_EXECUTION_STATUS_REJECTED {
		// If the node is rejected, update the flow execution status to rejected
		query = `UPDATE approval_flow_execution SET status = ? WHERE id = ?`
		_, err = tx.ExecContext(ctx, query, storepb.ApprovalFlowExecutionStatus_REJECTED, nodeExecution.ExecutionID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update approval flow execution status")
		}
	}

	return &newApproval, nil
}

// GetApproval gets an approval by ID.
func (s *Store) GetApproval(ctx context.Context, id int32) (*Approval, error) {
	if id <= 0 {
		return nil, errors.New("invalid approval ID")
	}

	query := `SELECT id, node_execution_id, user_id, status, comment, create_time FROM approval WHERE id = ?`
	var approval Approval
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&approval.ID, &approval.NodeExecutionID, &approval.UserID, &approval.Status, &approval.Comment, &approval.CreateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("approval not found")
		}
		return nil, errors.Wrap(err, "failed to get approval")
	}

	return &approval, nil
}

// getApprovalNodeExecution gets an approval node execution by ID.
func (s *Store) getApprovalNodeExecution(ctx context.Context, id int32) (*ApprovalNodeExecution, error) {
	query := `SELECT id, execution_id, node_id, status, creator_id, updater_id, create_time, update_time FROM approval_node_execution WHERE id = ?`
	var nodeExecution ApprovalNodeExecution
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&nodeExecution.ID, &nodeExecution.ExecutionID, &nodeExecution.NodeID, &nodeExecution.Status, &nodeExecution.CreatorID, &nodeExecution.UpdaterID, &nodeExecution.CreateTime, &nodeExecution.UpdateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("approval node execution not found")
		}
		return nil, errors.Wrap(err, "failed to get approval node execution")
	}

	return &nodeExecution, nil
}

// getApprovals gets all approvals for a given node execution ID.
func (s *Store) getApprovals(ctx context.Context, nodeExecutionID int32) ([]*Approval, error) {
	query := `SELECT id, node_execution_id, user_id, status, comment, create_time FROM approval WHERE node_execution_id = ? ORDER BY create_time DESC`

	rows, err := s.db.QueryContext(ctx, query, nodeExecutionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list approvals")
	}
	defer rows.Close()

	var approvals []*Approval
	for rows.Next() {
		var approval Approval
		if err := rows.Scan(&approval.ID, &approval.NodeExecutionID, &approval.UserID, &approval.Status, &approval.Comment, &approval.CreateTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan approval")
		}

		approvals = append(approvals, &approval)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over approvals")
	}

	return approvals, nil
}
