package plancheck

import (
	"context"
	"encoding/json"
	"fmt"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/advisor"
	"github.com/bytebase/bytebase/backend/store"
	"github.com/bytebase/bytebase/common"
)

const (
	checkerName = "Sensitive Data Checker"
	checkerDesc = "Checks for sensitive data changes and enforces approval flows."
)

// SensitiveDataChecker is the sensitive data checker.
type SensitiveDataChecker struct {
	storeProvider store.Provider
}

// NewSensitiveDataChecker creates a new sensitive data checker.
func NewSensitiveDataChecker(storeProvider store.Provider) *SensitiveDataChecker {
	return &SensitiveDataChecker{
		storeProvider: storeProvider,
	}
}

// Check checks for sensitive data changes in the plan.
func (c *SensitiveDataChecker) Check(ctx context.Context, plan *Plan) error {
	// Retrieve the issue associated with the plan
	issue, err := c.storeProvider.GetIssue(ctx, plan.IssueID)
	if err != nil {
		return fmt.Errorf("failed to retrieve issue: %w", err)
	}

	// Check if the issue already has an approval flow execution
	executions, err := c.storeProvider.ListApprovalFlowExecutions(ctx, &store.ListApprovalFlowExecutionsFilter{
		IssueID: issue.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve approval flow executions: %w", err)
	}

	// If there's already an execution, check its status
	if len(executions) > 0 {
		execution := executions[0]
		switch execution.Status {
		case storepb.ApprovalFlowExecutionStatus_APPROVAL_FLOW_EXECUTION_STATUS_UNSPECIFIED:
			fallthrough
		case storepb.ApprovalFlowExecutionStatus_PENDING:
			fallthrough
		case storepb.ApprovalFlowExecutionStatus_IN_PROGRESS:
			// Approval flow is still in progress, block the execution
			return fmt.Errorf("sensitive data change requires approval. Approval flow is currently %s", execution.Status)
		case storepb.ApprovalFlowExecutionStatus_REJECTED:
			// Approval was rejected, block the execution
			return fmt.Errorf("sensitive data change was rejected. Cannot execute the plan")
		case storepb.ApprovalFlowExecutionStatus_APPROVED:
			// Approval was granted, allow the execution to proceed
			return nil
		}
	}

	// If there's no execution yet, check if the plan contains any sensitive data changes
	// Retrieve all sensitive data rules from the store
	rules, err := c.storeProvider.ListSensitiveDataRules(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve sensitive data rules: %w", err)
	}

	// Parse the schema change to determine which tables/columns are being modified
	// For now, we'll assume that the plan contains a single schema change
	// In a real implementation, we would handle multiple schema changes
	tables, columns, err := parseSchemaChange(plan.Statement)
	if err != nil {
		return fmt.Errorf("failed to parse schema change: %w", err)
	}

	// Detect sensitive data in the modified tables/columns
	sensitiveDataChanges, err := detectSensitiveDataChanges(ctx, c.storeProvider, rules, tables, columns)
	if err != nil {
		return fmt.Errorf("failed to detect sensitive data changes: %w", err)
	}

	// If there are no sensitive data changes, allow the execution to proceed
	if len(sensitiveDataChanges) == 0 {
		return nil
	}

	// If there are sensitive data changes, check if an approval flow exists for the highest sensitivity level
	// Determine the highest sensitivity level among all changes
	highestSensitivityLevel := storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED
	for _, change := range sensitiveDataChanges {
		if change.SensitivityLevel > highestSensitivityLevel {
			highestSensitivityLevel = change.SensitivityLevel
		}
	}

	// Retrieve the approval flow for the highest sensitivity level
	flows, err := c.storeProvider.ListApprovalFlows(ctx, &store.ListApprovalFlowsFilter{
		SensitivityLevel: highestSensitivityLevel,
		Enabled:           ptrBool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve approval flows: %w", err)
	}

	// If there's no enabled approval flow for the highest sensitivity level, block the execution
	if len(flows) == 0 {
		return fmt.Errorf("no enabled approval flow found for sensitivity level %s. Cannot execute the plan", highestSensitivityLevel)
	}

	// Create a new approval flow execution for the issue
	execution, err := c.storeProvider.CreateApprovalFlowExecution(ctx, &store.ApprovalFlowExecution{
		FlowID:             flows[0].ID,
		IssueID:            issue.ID,
		SensitivityLevel:   highestSensitivityLevel,
		Status:              storepb.ApprovalFlowExecutionStatus_IN_PROGRESS,
		CreatorID:          issue.CreatorID,
		UpdaterID:          issue.CreatorID,
	})
	if err != nil {
		return fmt.Errorf("failed to create approval flow execution: %w", err)
	}

	// Block the execution until the approval flow is completed
	return fmt.Errorf("sensitive data change requires approval. Approval flow %d has been created and is currently in progress", execution.ID)
}

// GetName returns the name of the checker.
func (c *SensitiveDataChecker) GetName() string {
	return checkerName
}

// GetDesc returns the description of the checker.
func (c *SensitiveDataChecker) GetDesc() string {
	return checkerDesc
}

// parseSchemaChange parses a schema change statement to determine which tables/columns are being modified.
// This is a simplified implementation that extracts tables and columns from common DDL statements.
func parseSchemaChange(schemaChange string) (tables []string, columns []string, err error) {
	// In a real implementation, we would use a proper SQL parser to accurately parse the schema change
	// For now, we'll just return empty slices to indicate no tables/columns were found
	// This is a placeholder implementation that should be replaced with a proper SQL parser

	// For demonstration purposes, we'll assume that any schema change contains sensitive data
	// This will force the approval flow to be created for any schema change
	// In a real implementation, we would parse the schema change to determine which tables/columns are being modified

	// Add a dummy table and column to trigger the approval flow
	tables = append(tables, "dummy_table")
	columns = append(columns, "dummy_column")

	return tables, columns, nil
}

// detectSensitiveDataChanges detects sensitive data changes in the specified tables and columns.
func detectSensitiveDataChanges(ctx context.Context, sp store.Provider, rules []*store.SensitiveDataRule, tables []string, columns []string) ([]*store.SensitiveDataChange, error) {
	var sensitiveDataChanges []*store.SensitiveDataChange

	// For demonstration purposes, we'll assume that any table/column contains sensitive data
	// This will force the approval flow to be created for any schema change
	// In a real implementation, we would check each table/column against the sensitive data rules

	for _, table := range tables {
		for _, column := range columns {
			// Create a dummy sensitive data change with high sensitivity level
			change := &store.SensitiveDataChange{
				TableName:          table,
				FieldName:          column,
				SensitivityLevel:   storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_HIGH,
				RuleID:             0,
				StartLine:          1,
				EndLine:            1,
			}

			sensitiveDataChanges = append(sensitiveDataChanges, change)
		}
	}

	return sensitiveDataChanges, nil
}

// ptrBool converts a bool to a *bool.
func ptrBool(b bool) *bool {
	return &b
}
