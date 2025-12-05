package store

import (
	"context"
	"database/sql"
	"strings"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/pkg/errors"
)

// NotificationType is the type of notification.
type NotificationType string

const (
	// NotificationTypeApprovalRequest is the notification type for new approval requests.
	NotificationTypeApprovalRequest NotificationType = "approval_request"
	// NotificationTypeApprovalResult is the notification type for approval results.
	NotificationTypeApprovalResult NotificationType = "approval_result"
	// NotificationTypeApprovalFlowComplete is the notification type for approval flow completion.
	NotificationTypeApprovalFlowComplete NotificationType = "approval_flow_complete"
	// NotificationTypeApprovalNodeReminder is the notification type for approval node reminders.
	NotificationTypeApprovalNodeReminder NotificationType = "approval_node_reminder"
)

// Notification is the API message for notification.
type Notification struct {
	ID          int32
	Type        NotificationType
	Title       string
	Content     string
	RecipientID int32
	ApprovalFlowExecutionID int32
	ApprovalNodeExecutionID int32
	ApprovalID  int32
	Read        bool
	CreatorID   int32
	UpdaterID   int32
	CreateTime  int64
	UpdateTime  int64
}

// CreateNotification creates a new notification.
func (s *Store) CreateNotification(ctx context.Context, create *Notification) (*Notification, error) {
	if create.Type == "" {
		return nil, errors.New("notification type is required")
	}
	if create.Title == "" {
		return nil, errors.New("notification title is required")
	}
	if create.RecipientID <= 0 {
		return nil, errors.New("notification recipient ID is required")
	}

	// Insert the new notification
	query := `INSERT INTO notification (type, title, content, recipient_id, approval_flow_execution_id, approval_node_execution_id, approval_id, read, creator_id, updater_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	result, err := s.db.ExecContext(ctx, query, create.Type, create.Title, create.Content, create.RecipientID, create.ApprovalFlowExecutionID, create.ApprovalNodeExecutionID, create.ApprovalID, create.Read, create.CreatorID, create.UpdaterID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create notification")
	}

	notificationID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get last insert ID for notification")
	}

	// Retrieve the newly created notification
	notification, err := s.getNotification(ctx, int32(notificationID))
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve newly created notification")
	}

	return notification, nil
}

// ListNotifications lists notifications for a given recipient.
func (s *Store) ListNotifications(ctx context.Context, recipientID int32, read *bool) ([]*Notification, error) {
	if recipientID <= 0 {
		return nil, errors.New("recipient ID is required")
	}

	query := `SELECT id, type, title, content, recipient_id, approval_flow_execution_id, approval_node_execution_id, approval_id, read, creator_id, updater_id, create_time, update_time FROM notification WHERE recipient_id = ?`
	args := []interface{}{recipientID}

	if read != nil {
		query += " AND read = ?"
		args = append(args, *read)
	}

	query += " ORDER BY create_time DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list notifications")
	}
	defer rows.Close()

	var notifications []*Notification
	for rows.Next() {
		var notification Notification
		if err := rows.Scan(&notification.ID, &notification.Type, &notification.Title, &notification.Content, &notification.RecipientID, &notification.ApprovalFlowExecutionID, &notification.ApprovalNodeExecutionID, &notification.ApprovalID, &notification.Read, &notification.CreatorID, &notification.UpdaterID, &notification.CreateTime, &notification.UpdateTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan notification")
		}

		notifications = append(notifications, &notification)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over notifications")
	}

	return notifications, nil
}

// UpdateNotification updates a notification.
func (s *Store) UpdateNotification(ctx context.Context, update *Notification) (*Notification, error) {
	if update.ID <= 0 {
		return nil, errors.New("notification ID is required")
	}

	// Build the update query
	query := `UPDATE notification SET `
	args := []interface{}{}

	fields := []string{}
	if update.Read != nil {
		fields = append(fields, "read = ?")
		args = append(args, *update.Read)
	}
	if update.UpdaterID > 0 {
		fields = append(fields, "updater_id = ?")
		args = append(args, update.UpdaterID)
	}

	if len(fields) == 0 {
		return nil, errors.New("no fields to update")
	}

	query += strings.Join(fields, ", ")
	query += " WHERE id = ?"
	args = append(args, update.ID)

	// Execute the update query
	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update notification")
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rows affected for notification update")
	}
	if rowsAffected == 0 {
		return nil, errors.New("notification not found")
	}

	// Retrieve the updated notification
	notification, err := s.getNotification(ctx, update.ID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve updated notification")
	}

	return notification, nil
}

// DeleteNotification deletes a notification.
func (s *Store) DeleteNotification(ctx context.Context, id int32) error {
	if id <= 0 {
		return errors.New("notification ID is required")
	}

	// Delete the notification
	query := `DELETE FROM notification WHERE id = ?`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, "failed to delete notification")
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected for notification delete")
	}
	if rowsAffected == 0 {
		return errors.New("notification not found")
	}

	return nil
}

// getNotification retrieves a notification by ID.
func (s *Store) getNotification(ctx context.Context, id int32) (*Notification, error) {
	if id <= 0 {
		return nil, errors.New("notification ID is required")
	}

	query := `SELECT id, type, title, content, recipient_id, approval_flow_execution_id, approval_node_execution_id, approval_id, read, creator_id, updater_id, create_time, update_time FROM notification WHERE id = ?`

	var notification Notification
	if err := s.db.QueryRowContext(ctx, query, id).Scan(&notification.ID, &notification.Type, &notification.Title, &notification.Content, &notification.RecipientID, &notification.ApprovalFlowExecutionID, &notification.ApprovalNodeExecutionID, &notification.ApprovalID, &notification.Read, &notification.CreatorID, &notification.UpdaterID, &notification.CreateTime, &notification.UpdateTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("notification not found")
		}
		return nil, errors.Wrap(err, "failed to scan notification")
	}

	return &notification, nil
}
