package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/common"
	"github.com/bytebase/bytebase/backend/common/qb"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// SensitiveDataRuleMessage is the message for sensitive data rule.
type SensitiveDataRuleMessage struct {
	ID          int
	Name        string
	ProjectID   string
	Rule        *v1pb.SensitiveDataRule
	CreatorID   int
	UpdaterID   int
	CreateTime  int64
	UpdateTime  int64
}

// CreateSensitiveDataRule creates a new sensitive data rule.
func (s *Store) CreateSensitiveDataRule(ctx context.Context, rule *v1pb.SensitiveDataRule) (*v1pb.SensitiveDataRule, error) {
	// Validate rule
	if rule.Name == "" {
		return nil, errors.New("rule name is required")
	}
	if rule.Project == "" {
		return nil, errors.New("project is required")
	}

	// Parse project ID from project name
	projectID, err := common.GetProjectIDFromName(rule.Project)
	if err != nil {
		return nil, errors.Wrap(err, "invalid project name")
	}

	// Marshal rule to JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal rule")
	}

	// Insert into database
	query := `
		INSERT INTO sensitive_data_rule (name, project_id, rule, creator_id, updater_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, create_time, update_time
	`

	// Get current user ID from context
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user ID")
	}

	var id int
	var createTime, updateTime int64
	err = s.dbConnManager.GetDB().QueryRowContext(ctx, query, rule.Name, projectID, ruleJSON, currentUserID, currentUserID).Scan(&id, &createTime, &updateTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sensitive data rule")
	}

	// Return the created rule
	return rule, nil
}

// GetSensitiveDataRuleByName gets a sensitive data rule by name.
func (s *Store) GetSensitiveDataRuleByName(ctx context.Context, name string) (*v1pb.SensitiveDataRule, error) {
	// Validate name
	if name == "" {
		return nil, errors.New("rule name is required")
	}

	// Parse project ID from name
	projectID, err := common.GetProjectIDFromName(name)
	if err != nil {
		return nil, errors.Wrap(err, "invalid rule name")
	}

	// Query from database
	query := `
		SELECT id, rule, create_time, update_time
		FROM sensitive_data_rule
		WHERE name = $1 AND project_id = $2
	`

	var id int
	var ruleJSON []byte
	var createTime, updateTime int64
	err = s.dbConnManager.GetDB().QueryRowContext(ctx, query, name, projectID).Scan(&id, &ruleJSON, &createTime, &updateTime)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get sensitive data rule")
	}

	// Unmarshal rule
	var rule v1pb.SensitiveDataRule
	err = json.Unmarshal(ruleJSON, &rule)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal rule")
	}

	return &rule, nil
}

// ListSensitiveDataRules lists all sensitive data rules for a project.
func (s *Store) ListSensitiveDataRules(ctx context.Context, parent string) ([]*v1pb.SensitiveDataRule, error) {
	// Validate parent
	if parent == "" {
		return nil, errors.New("parent is required")
	}

	// Parse project ID from parent
	projectID, err := common.GetProjectIDFromName(parent)
	if err != nil {
		return nil, errors.Wrap(err, "invalid parent name")
	}

	// Query from database
	query := `
		SELECT id, rule
		FROM sensitive_data_rule
		WHERE project_id = $1
		ORDER BY create_time DESC
	`

	rows, err := s.dbConnManager.GetDB().QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sensitive data rules")
	}
	defer rows.Close()

	// Unmarshal rules
	var rules []*v1pb.SensitiveDataRule
	for rows.Next() {
		var id int
		var ruleJSON []byte
		err := rows.Scan(&id, &ruleJSON)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan sensitive data rule")
		}

		var rule v1pb.SensitiveDataRule
		err = json.Unmarshal(ruleJSON, &rule)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal rule")
		}

		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate sensitive data rules")
	}

	return rules, nil
}

// UpdateSensitiveDataRule updates a sensitive data rule.
func (s *Store) UpdateSensitiveDataRule(ctx context.Context, rule *v1pb.SensitiveDataRule, updateMask *fieldmaskpb.FieldMask) (*v1pb.SensitiveDataRule, error) {
	// Validate rule
	if rule.Name == "" {
		return nil, errors.New("rule name is required")
	}

	// Parse project ID from name
	projectID, err := common.GetProjectIDFromName(rule.Name)
	if err != nil {
		return nil, errors.Wrap(err, "invalid rule name")
	}

	// Marshal rule to JSON
	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal rule")
	}

	// Update database
	query := `
		UPDATE sensitive_data_rule
		SET rule = $1, updater_id = $2, update_time = CURRENT_TIMESTAMP
		WHERE name = $3 AND project_id = $4
		RETURNING update_time
	`

	// Get current user ID from context
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user ID")
	}

	var updateTime int64
	err = s.dbConnManager.GetDB().QueryRowContext(ctx, query, ruleJSON, currentUserID, rule.Name, projectID).Scan(&updateTime)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to update sensitive data rule")
	}

	// Return the updated rule
	return rule, nil
}

// DeleteSensitiveDataRule deletes a sensitive data rule.
func (s *Store) DeleteSensitiveDataRule(ctx context.Context, name string) error {
	// Validate name
	if name == "" {
		return errors.New("rule name is required")
	}

	// Parse project ID from name
	projectID, err := common.GetProjectIDFromName(name)
	if err != nil {
		return errors.Wrap(err, "invalid rule name")
	}

	// Delete from database
	query := `
		DELETE FROM sensitive_data_rule
		WHERE name = $1 AND project_id = $2
	`

	result, err := s.dbConnManager.GetDB().ExecContext(ctx, query, name, projectID)
	if err != nil {
		return errors.Wrap(err, "failed to delete sensitive data rule")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// ApprovalFlowMessage is the message for approval flow.
type ApprovalFlowMessage struct {
	ID          int
	Name        string
	ProjectID   string
	Flow        *v1pb.ApprovalFlow
	CreatorID   int
	UpdaterID   int
	CreateTime  int64
	UpdateTime  int64
}

// CreateApprovalFlow creates a new approval flow.
func (s *Store) CreateApprovalFlow(ctx context.Context, flow *v1pb.ApprovalFlow) (*v1pb.ApprovalFlow, error) {
	// Validate flow
	if flow.Name == "" {
		return nil, errors.New("flow name is required")
	}
	if flow.Project == "" {
		return nil, errors.New("project is required")
	}

	// Parse project ID from project name
	projectID, err := common.GetProjectIDFromName(flow.Project)
	if err != nil {
		return nil, errors.Wrap(err, "invalid project name")
	}

	// Marshal flow to JSON
	flowJSON, err := json.Marshal(flow)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal flow")
	}

	// Insert into database
	query := `
		INSERT INTO approval_flow (name, project_id, flow, creator_id, updater_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, create_time, update_time
	`

	// Get current user ID from context
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user ID")
	}

	var id int
	var createTime, updateTime int64
	err = s.dbConnManager.GetDB().QueryRowContext(ctx, query, flow.Name, projectID, flowJSON, currentUserID, currentUserID).Scan(&id, &createTime, &updateTime)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create approval flow")
	}

	// Return the created flow
	return flow, nil
}

// GetApprovalFlowByName gets an approval flow by name.
func (s *Store) GetApprovalFlowByName(ctx context.Context, name string) (*v1pb.ApprovalFlow, error) {
	// Validate name
	if name == "" {
		return nil, errors.New("flow name is required")
	}

	// Parse project ID from name
	projectID, err := common.GetProjectIDFromName(name)
	if err != nil {
		return nil, errors.Wrap(err, "invalid flow name")
	}

	// Query from database
	query := `
		SELECT id, flow, create_time, update_time
		FROM approval_flow
		WHERE name = $1 AND project_id = $2
	`

	var id int
	var flowJSON []byte
	var createTime, updateTime int64
	err = s.dbConnManager.GetDB().QueryRowContext(ctx, query, name, projectID).Scan(&id, &flowJSON, &createTime, &updateTime)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get approval flow")
	}

	// Unmarshal flow
	var flow v1pb.ApprovalFlow
	err = json.Unmarshal(flowJSON, &flow)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal flow")
	}

	return &flow, nil
}

// ListApprovalFlows lists all approval flows for a project.
func (s *Store) ListApprovalFlows(ctx context.Context, parent string) ([]*v1pb.ApprovalFlow, error) {
	// Validate parent
	if parent == "" {
		return nil, errors.New("parent is required")
	}

	// Parse project ID from parent
	projectID, err := common.GetProjectIDFromName(parent)
	if err != nil {
		return nil, errors.Wrap(err, "invalid parent name")
	}

	// Query from database
	query := `
		SELECT id, flow
		FROM approval_flow
		WHERE project_id = $1
		ORDER BY create_time DESC
	`

	rows, err := s.dbConnManager.GetDB().QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list approval flows")
	}
	defer rows.Close()

	// Unmarshal flows
	var flows []*v1pb.ApprovalFlow
	for rows.Next() {
		var id int
		var flowJSON []byte
		err := rows.Scan(&id, &flowJSON)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan approval flow")
		}

		var flow v1pb.ApprovalFlow
		err = json.Unmarshal(flowJSON, &flow)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal flow")
		}

		flows = append(flows, &flow)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate approval flows")
	}

	return flows, nil
}

// UpdateApprovalFlow updates an approval flow.
func (s *Store) UpdateApprovalFlow(ctx context.Context, flow *v1pb.ApprovalFlow, updateMask *fieldmaskpb.FieldMask) (*v1pb.ApprovalFlow, error) {
	// Validate flow
	if flow.Name == "" {
		return nil, errors.New("flow name is required")
	}

	// Parse project ID from name
	projectID, err := common.GetProjectIDFromName(flow.Name)
	if err != nil {
		return nil, errors.Wrap(err, "invalid flow name")
	}

	// Marshal flow to JSON
	flowJSON, err := json.Marshal(flow)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal flow")
	}

	// Update database
	query := `
		UPDATE approval_flow
		SET flow = $1, updater_id = $2, update_time = CURRENT_TIMESTAMP
		WHERE name = $3 AND project_id = $4
		RETURNING update_time
	`

	// Get current user ID from context
	currentUserID, err := getCurrentUserID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current user ID")
	}

	var updateTime int64
	err = s.dbConnManager.GetDB().QueryRowContext(ctx, query, flowJSON, currentUserID, flow.Name, projectID).Scan(&updateTime)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to update approval flow")
	}

	// Return the updated flow
	return flow, nil
}

// DeleteApprovalFlow deletes an approval flow.
func (s *Store) DeleteApprovalFlow(ctx context.Context, name string) error {
	// Validate name
	if name == "" {
		return errors.New("flow name is required")
	}

	// Parse project ID from name
	projectID, err := common.GetProjectIDFromName(name)
	if err != nil {
		return errors.Wrap(err, "invalid flow name")
	}

	// Delete from database
	query := `
		DELETE FROM approval_flow
		WHERE name = $1 AND project_id = $2
	`

	result, err := s.dbConnManager.GetDB().ExecContext(ctx, query, name, projectID)
	if err != nil {
		return errors.Wrap(err, "failed to delete approval flow")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// getCurrentUserID gets the current user ID from context.
func getCurrentUserID(ctx context.Context) (int, error) {
	// TODO: Implement getCurrentUserID
	// For now, return a mock user ID
	return 1, nil
}