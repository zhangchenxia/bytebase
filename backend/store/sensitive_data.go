package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/bytebase/bytebase/backend/common"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/store/model"
	"github.com/bytebase/bytebase/backend/utils"
	"github.com/pkg/errors"
)

// SensitiveDataRule is the API message for sensitive data rule.
type SensitiveDataRule struct {
	ID              int32
	Title           string
	Description     string
	Level           storepb.SensitiveDataLevel
	DatabaseType    string
	TablePattern    string
	ColumnPattern   string
	DataType        string
	RegexPattern    string
	Enabled         bool
	CreatorID       int32
	UpdaterID       int32
	CreateTime      int64
	UpdateTime      int64
}

// SensitiveDataField is the API message for sensitive data field.
type SensitiveDataField struct {
	ID          int32
	TableID     int32
	TableName   string
	FieldName   string
	Level       storepb.SensitiveDataLevel
	RuleID      int32
}

// SensitiveDataChange is the API message for sensitive data change.
type SensitiveDataChange struct {
	ID          int32
	IssueID     int32
	TableID     int32
	TableName   string
	FieldName   string
	Level       storepb.SensitiveDataLevel
	ChangeType  string
	OldValue    string
	NewValue    string
}

// CreateSensitiveDataRule creates a new sensitive data rule.
func (s *Store) CreateSensitiveDataRule(ctx context.Context, create *SensitiveDataRule) (*SensitiveDataRule, error) {
	if create.Title == "" {
		return nil, errors.New("title is required")
	}
	if create.Level == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return nil, errors.New("level is required")
	}

	// Check if there's already a rule with the same title
	var existingID int32
	query := `SELECT id FROM sensitive_data_rule WHERE title = ?`
	if err := s.db.QueryRowContext(ctx, query, create.Title).Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to check existing sensitive data rule")
	}
	if existingID > 0 {
		return nil, errors.New("a rule with the same title already exists")
	}

	// Insert the new rule
	query = `INSERT INTO sensitive_data_rule (title, description, level, database_type, table_pattern, column_pattern, data_type, regex_pattern, enabled, creator_id, updater_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.ExecContext(ctx, query, create.Title, create.Description, create.Level, create.DatabaseType, create.TablePattern, create.ColumnPattern, create.DataType, create.RegexPattern, create.Enabled, create.CreatorID, create.UpdaterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sensitive data rule")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert ID for sensitive data rule")
	}

	// Retrieve the newly created rule
	return s.GetSensitiveDataRule(ctx, int32(id))
}

// ListSensitiveDataRules lists all sensitive data rules.
func (s *Store) ListSensitiveDataRules(ctx context.Context, filter *ListSensitiveDataRulesFilter) ([]*SensitiveDataRule, error) {
	query := `SELECT id, title, description, level, database_type, table_pattern, column_pattern, data_type, regex_pattern, enabled, creator_id, updater_id, create_time, update_time FROM sensitive_data_rule`
	args := []interface{}{}

	whereClauses := []string{}
	if filter != nil {
		if filter.Level != storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
			whereClauses = append(whereClauses, "level = ?")
			args = append(args, filter.Level)
		}
		if filter.DatabaseType != "" {
			whereClauses = append(whereClauses, "database_type = ?")
			args = append(args, filter.DatabaseType)
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
		return nil, errors.Wrap(err, "failed to list sensitive data rules")
	}
	defer rows.Close()

	var rules []*SensitiveDataRule
	for rows.Next() {
		var rule SensitiveDataRule
		if err := rows.Scan(&rule.ID, &rule.Title, &rule.Description, &rule.Level, &rule.DatabaseType, &rule.TablePattern, &rule.ColumnPattern, &rule.DataType, &rule.RegexPattern, &rule.Enabled, &rule.CreatorID, &rule.UpdaterID, &rule.CreateTime, &rule.UpdateTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan sensitive data rule")
		}
		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over sensitive data rules")
	}

	return rules, nil
}

// GetSensitiveDataRule gets a sensitive data rule by ID.
func (s *Store) GetSensitiveDataRule(ctx context.Context, id int32) (*SensitiveDataRule, error) {
	if id <= 0 {
		return nil, errors.New("invalid sensitive data rule ID")
	}

	query := `SELECT id, title, description, level, database_type, table_pattern, column_pattern, data_type, regex_pattern, enabled, creator_id, updater_id, create_time, update_time FROM sensitive_data_rule WHERE id = ?`
	var rule SensitiveDataRule
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&rule.ID, &rule.Title, &rule.Description, &rule.Level, &rule.DatabaseType, &rule.TablePattern, &rule.ColumnPattern, &rule.DataType, &rule.RegexPattern, &rule.Enabled, &rule.CreatorID, &rule.UpdaterID, &rule.CreateTime, &rule.UpdateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("sensitive data rule not found")
		}
		return nil, errors.Wrap(err, "failed to get sensitive data rule")
	}

	return &rule, nil
}

// UpdateSensitiveDataRule updates a sensitive data rule.
func (s *Store) UpdateSensitiveDataRule(ctx context.Context, update *SensitiveDataRule) (*SensitiveDataRule, error) {
	if update.ID <= 0 {
		return nil, errors.New("invalid sensitive data rule ID")
	}

	// Check if the rule exists
	_, err := s.GetSensitiveDataRule(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	// Check if there's another rule with the same title
	var existingID int32
	query := `SELECT id FROM sensitive_data_rule WHERE title = ? AND id != ?`
	if err := s.db.QueryRowContext(ctx, query, update.Title, update.ID).Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to check existing sensitive data rule")
	}
	if existingID > 0 {
		return nil, errors.New("a rule with the same title already exists")
	}

	// Update the rule
	query = `UPDATE sensitive_data_rule SET title = ?, description = ?, level = ?, database_type = ?, table_pattern = ?, column_pattern = ?, data_type = ?, regex_pattern = ?, enabled = ?, updater_id = ? WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, update.Title, update.Description, update.Level, update.DatabaseType, update.TablePattern, update.ColumnPattern, update.DataType, update.RegexPattern, update.Enabled, update.UpdaterID, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update sensitive data rule")
	}

	// Retrieve the updated rule
	return s.GetSensitiveDataRule(ctx, update.ID)
}

// DeleteSensitiveDataRule deletes a sensitive data rule by ID.
func (s *Store) DeleteSensitiveDataRule(ctx context.Context, id int32) error {
	if id <= 0 {
		return errors.New("invalid sensitive data rule ID")
	}

	// Check if the rule exists
	_, err := s.GetSensitiveDataRule(ctx, id)
	if err != nil {
		return err
	}

	// Delete the rule
	query := `DELETE FROM sensitive_data_rule WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete sensitive data rule")
	}

	return nil
}

// ListSensitiveDataRulesFilter is the filter for listing sensitive data rules.
type ListSensitiveDataRulesFilter struct {
	Level         storepb.SensitiveDataLevel
	DatabaseType  string
	Enabled       *bool
}

// CreateSensitiveDataField creates a new sensitive data field.
func (s *Store) CreateSensitiveDataField(ctx context.Context, create *SensitiveDataField) (*SensitiveDataField, error) {
	if create.TableID <= 0 {
		return nil, errors.New("table ID is required")
	}
	if create.FieldName == "" {
		return nil, errors.New("field name is required")
	}
	if create.Level == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return nil, errors.New("level is required")
	}
	if create.RuleID <= 0 {
		return nil, errors.New("rule ID is required")
	}

	// Check if there's already a field with the same table ID and field name
	var existingID int32
	query := `SELECT id FROM sensitive_data_field WHERE table_id = ? AND field_name = ?`
	if err := s.db.QueryRowContext(ctx, query, create.TableID, create.FieldName).Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to check existing sensitive data field")
	}
	if existingID > 0 {
		return nil, errors.New("a field with the same table ID and field name already exists")
	}

	// Insert the new field
	query = `INSERT INTO sensitive_data_field (table_id, table_name, field_name, level, rule_id) VALUES (?, ?, ?, ?, ?)`
	result, err := s.db.ExecContext(ctx, query, create.TableID, create.TableName, create.FieldName, create.Level, create.RuleID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sensitive data field")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert ID for sensitive data field")
	}

	// Retrieve the newly created field
	return s.GetSensitiveDataField(ctx, int32(id))
}

// ListSensitiveDataFields lists all sensitive data fields.
func (s *Store) ListSensitiveDataFields(ctx context.Context, filter *ListSensitiveDataFieldsFilter) ([]*SensitiveDataField, error) {
	query := `SELECT id, table_id, table_name, field_name, level, rule_id FROM sensitive_data_field`
	args := []interface{}{}

	whereClauses := []string{}
	if filter != nil {
		if filter.TableID > 0 {
			whereClauses = append(whereClauses, "table_id = ?")
			args = append(args, filter.TableID)
		}
		if filter.Level != storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
			whereClauses = append(whereClauses, "level = ?")
			args = append(args, filter.Level)
		}
		if filter.RuleID > 0 {
			whereClauses = append(whereClauses, "rule_id = ?")
			args = append(args, filter.RuleID)
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY table_name, field_name"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sensitive data fields")
	}
	defer rows.Close()

	var fields []*SensitiveDataField
	for rows.Next() {
		var field SensitiveDataField
		if err := rows.Scan(&field.ID, &field.TableID, &field.TableName, &field.FieldName, &field.Level, &field.RuleID); err != nil {
			return nil, errors.Wrap(err, "failed to scan sensitive data field")
		}
		fields = append(fields, &field)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over sensitive data fields")
	}

	return fields, nil
}

// GetSensitiveDataField gets a sensitive data field by ID.
func (s *Store) GetSensitiveDataField(ctx context.Context, id int32) (*SensitiveDataField, error) {
	if id <= 0 {
		return nil, errors.New("invalid sensitive data field ID")
	}

	query := `SELECT id, table_id, table_name, field_name, level, rule_id FROM sensitive_data_field WHERE id = ?`
	var field SensitiveDataField
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&field.ID, &field.TableID, &field.TableName, &field.FieldName, &field.Level, &field.RuleID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("sensitive data field not found")
		}
		return nil, errors.Wrap(err, "failed to get sensitive data field")
	}

	return &field, nil
}

// UpdateSensitiveDataField updates a sensitive data field.
func (s *Store) UpdateSensitiveDataField(ctx context.Context, update *SensitiveDataField) (*SensitiveDataField, error) {
	if update.ID <= 0 {
		return nil, errors.New("invalid sensitive data field ID")
	}

	// Check if the field exists
	_, err := s.GetSensitiveDataField(ctx, update.ID)
	if err != nil {
		return nil, err
	}

	// Check if there's another field with the same table ID and field name
	var existingID int32
	query := `SELECT id FROM sensitive_data_field WHERE table_id = ? AND field_name = ? AND id != ?`
	if err := s.db.QueryRowContext(ctx, query, update.TableID, update.FieldName, update.ID).Scan(&existingID); err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "failed to check existing sensitive data field")
	}
	if existingID > 0 {
		return nil, errors.New("a field with the same table ID and field name already exists")
	}

	// Update the field
	query = `UPDATE sensitive_data_field SET table_id = ?, table_name = ?, field_name = ?, level = ?, rule_id = ? WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, update.TableID, update.TableName, update.FieldName, update.Level, update.RuleID, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update sensitive data field")
	}

	// Retrieve the updated field
	return s.GetSensitiveDataField(ctx, update.ID)
}

// DeleteSensitiveDataField deletes a sensitive data field by ID.
func (s *Store) DeleteSensitiveDataField(ctx context.Context, id int32) error {
	if id <= 0 {
		return errors.New("invalid sensitive data field ID")
	}

	// Check if the field exists
	_, err := s.GetSensitiveDataField(ctx, id)
	if err != nil {
		return err
	}

	// Delete the field
	query := `DELETE FROM sensitive_data_field WHERE id = ?`
	_, err = s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete sensitive data field")
	}

	return nil
}

// ListSensitiveDataFieldsFilter is the filter for listing sensitive data fields.
type ListSensitiveDataFieldsFilter struct {
	TableID  int32
	Level    storepb.SensitiveDataLevel
	RuleID   int32
}

// CreateSensitiveDataChange creates a new sensitive data change.
func (s *Store) CreateSensitiveDataChange(ctx context.Context, create *SensitiveDataChange) (*SensitiveDataChange, error) {
	if create.IssueID <= 0 {
		return nil, errors.New("issue ID is required")
	}
	if create.TableID <= 0 {
		return nil, errors.New("table ID is required")
	}
	if create.FieldName == "" {
		return nil, errors.New("field name is required")
	}
	if create.Level == storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
		return nil, errors.New("level is required")
	}
	if create.ChangeType == "" {
		return nil, errors.New("change type is required")
	}

	// Insert the new change
	query := `INSERT INTO sensitive_data_change (issue_id, table_id, table_name, field_name, level, change_type, old_value, new_value) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.ExecContext(ctx, query, create.IssueID, create.TableID, create.TableName, create.FieldName, create.Level, create.ChangeType, create.OldValue, create.NewValue)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create sensitive data change")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert ID for sensitive data change")
	}

	// Retrieve the newly created change
	return s.GetSensitiveDataChange(ctx, int32(id))
}

// ListSensitiveDataChanges lists all sensitive data changes.
func (s *Store) ListSensitiveDataChanges(ctx context.Context, filter *ListSensitiveDataChangesFilter) ([]*SensitiveDataChange, error) {
	query := `SELECT id, issue_id, table_id, table_name, field_name, level, change_type, old_value, new_value FROM sensitive_data_change`
	args := []interface{}{}

	whereClauses := []string{}
	if filter != nil {
		if filter.IssueID > 0 {
			whereClauses = append(whereClauses, "issue_id = ?")
			args = append(args, filter.IssueID)
		}
		if filter.TableID > 0 {
			whereClauses = append(whereClauses, "table_id = ?")
			args = append(args, filter.TableID)
		}
		if filter.Level != storepb.SensitiveDataLevel_SENSITIVE_DATA_LEVEL_UNSPECIFIED {
			whereClauses = append(whereClauses, "level = ?")
			args = append(args, filter.Level)
		}
		if filter.ChangeType != "" {
			whereClauses = append(whereClauses, "change_type = ?")
			args = append(args, filter.ChangeType)
		}
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY id DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sensitive data changes")
	}
	defer rows.Close()

	var changes []*SensitiveDataChange
	for rows.Next() {
		var change SensitiveDataChange
		if err := rows.Scan(&change.ID, &change.IssueID, &change.TableID, &change.TableName, &change.FieldName, &change.Level, &change.ChangeType, &change.OldValue, &change.NewValue); err != nil {
			return nil, errors.Wrap(err, "failed to scan sensitive data change")
		}
		changes = append(changes, &change)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over sensitive data changes")
	}

	return changes, nil
}

// GetSensitiveDataChange gets a sensitive data change by ID.
func (s *Store) GetSensitiveDataChange(ctx context.Context, id int32) (*SensitiveDataChange, error) {
	if id <= 0 {
		return nil, errors.New("invalid sensitive data change ID")
	}

	query := `SELECT id, issue_id, table_id, table_name, field_name, level, change_type, old_value, new_value FROM sensitive_data_change WHERE id = ?`
	var change SensitiveDataChange
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&change.ID, &change.IssueID, &change.TableID, &change.TableName, &change.FieldName, &change.Level, &change.ChangeType, &change.OldValue, &change.NewValue); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("sensitive data change not found")
		}
		return nil, errors.Wrap(err, "failed to get sensitive data change")
	}

	return &change, nil
}

// ListSensitiveDataChangesFilter is the filter for listing sensitive data changes.
type ListSensitiveDataChangesFilter struct {
	IssueID     int32
	TableID     int32
	Level       storepb.SensitiveDataLevel
	ChangeType  string
}
