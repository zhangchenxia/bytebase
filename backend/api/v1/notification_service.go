package v1

import (
	"context"

	"connectrpc.com/connect"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/generated-go/v1/v1connect"
	"github.com/bytebase/bytebase/backend/store"
)

// NotificationService implements the notification service.
type NotificationService struct {
	v1connect.UnimplementedNotificationServiceHandler
	store *store.Store
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(store *store.Store) *NotificationService {
	return &NotificationService{
		store: store,
	}
}

// ListNotifications lists notifications for the current user.
func (s *NotificationService) ListNotifications(ctx context.Context, req *connect.Request[v1pb.ListNotificationsRequest]) (*connect.Response[v1pb.ListNotificationsResponse], error) {
	// TODO: Get the current user ID from the context
	currentUserID := int32(1) // Temporary placeholder

	// Call the store to list notifications
	notifications, err := s.store.ListNotifications(ctx, currentUserID, req.Msg.Read)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert store notifications to proto notifications
	var protoNotifications []*v1pb.Notification
	for _, notification := range notifications {
		protoNotification := &v1pb.Notification{
			Name:                        notification.Name(),
			Uid:                         notification.Uid(),
			Type:                        v1pb.NotificationType(v1pb.NotificationType_value[string(notification.Type)]),
			Title:                       notification.Title,
			Content:                     notification.Content,
			RecipientId:                 notification.RecipientID,
			ApprovalFlowExecutionId:    notification.ApprovalFlowExecutionID,
			ApprovalNodeExecutionId:    notification.ApprovalNodeExecutionID,
			ApprovalId:                  notification.ApprovalID,
			Read:                        notification.Read,
			CreateTime:                  timestampFromInt64(notification.CreateTime),
			UpdateTime:                  timestampFromInt64(notification.UpdateTime),
		}

		protoNotifications = append(protoNotifications, protoNotification)
	}

	// Create the response
	resp := &v1pb.ListNotificationsResponse{
		Notifications: protoNotifications,
	}

	return connect.NewResponse(resp), nil
}

// UpdateNotification updates a notification.
func (s *NotificationService) UpdateNotification(ctx context.Context, req *connect.Request[v1pb.UpdateNotificationRequest]) (*connect.Response[v1pb.UpdateNotificationResponse], error) {
	// TODO: Get the current user ID from the context
	currentUserID := int32(1) // Temporary placeholder

	// Parse the notification ID from the name
	id, err := parseResourceID(req.Msg.Name, "notifications")
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse notification name: %v", err)
	}

	// Create the update request
	update := &store.Notification{
		ID:          id,
		Read:        req.Msg.Notification.Read,
		UpdaterID:   currentUserID,
	}

	// Call the store to update the notification
	notification, err := s.store.UpdateNotification(ctx, update)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert store notification to proto notification
	protoNotification := &v1pb.Notification{
		Name:                        notification.Name(),
		Uid:                         notification.Uid(),
		Type:                        v1pb.NotificationType(v1pb.NotificationType_value[string(notification.Type)]),
		Title:                       notification.Title,
		Content:                     notification.Content,
		RecipientId:                 notification.RecipientID,
		ApprovalFlowExecutionId:    notification.ApprovalFlowExecutionID,
		ApprovalNodeExecutionId:    notification.ApprovalNodeExecutionID,
		ApprovalId:                  notification.ApprovalID,
		Read:                        notification.Read,
		CreateTime:                  timestampFromInt64(notification.CreateTime),
		UpdateTime:                  timestampFromInt64(notification.UpdateTime),
	}

	// Create the response
	resp := &v1pb.UpdateNotificationResponse{
		Notification: protoNotification,
	}

	return connect.NewResponse(resp), nil
}

// Notification methods
func (n *store.Notification) Name() string {
	return fmt.Sprintf("notifications/%d", n.ID)
}

func (n *store.Notification) Uid() string {
	return fmt.Sprintf("%d", n.ID)
}
