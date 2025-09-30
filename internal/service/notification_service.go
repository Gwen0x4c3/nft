package service

import (
	"context"
	"fmt"
	"time"

	"nft-platform/internal/models"
	"nft-platform/internal/repository"
	"nft-platform/internal/types"
	"nft-platform/pkg/validation"
)

// NotificationService handles notification-related business operations
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	userRepo         *repository.UserRepository
	validator        *validation.CustomValidator
	wsManager        WebSocketManager
	messagingService MessagingService
}

// NotificationServiceConfig represents configuration for the notification service
type NotificationServiceConfig struct {
	WebSocketManager WebSocketManager
	MessagingService MessagingService
}

// NewNotificationService creates a new NotificationService instance
func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	config *NotificationServiceConfig,
) *NotificationService {
	service := &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		validator:        validation.NewCustomValidator(),
	}

	if config != nil {
		service.wsManager = config.WebSocketManager
		service.messagingService = config.MessagingService
	}

	return service
}

// SendNotification creates and sends a notification to a user
func (s *NotificationService) SendNotification(ctx context.Context, userID uint, notificationType string, title string, message string, data map[string]interface{}) (*models.Notification, error) {
	// Verify user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Create notification request
	req := &types.CreateNotificationRequest{
		UserID:  userID,
		Type:    notificationType,
		Title:   title,
		Message: message,
		Data:    data,
	}

	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create notification
	notification := &models.Notification{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	// Set data if provided
	if data != nil {
		notification.SetData(data)
	}

	// Save to database
	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send real-time notification via WebSocket
	if s.wsManager != nil {
		wsData := map[string]interface{}{
			"notification": notification,
			"type":         "new_notification",
		}
		if err := s.wsManager.SendToUser(userID, "notification", wsData); err != nil {
			// Log error but don't fail the notification creation
			fmt.Printf("Warning: failed to send WebSocket notification to user %d: %v\n", userID, err)
		}
	}

	// Publish to messaging service for external processing
	if s.messagingService != nil {
		if err := s.messagingService.PublishNotification(ctx, notification); err != nil {
			// Log error but don't fail the notification creation
			fmt.Printf("Warning: failed to publish notification to messaging service: %v\n", err)
		}
	}

	return notification, nil
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(ctx context.Context, req *types.CreateNotificationRequest) (*models.Notification, error) {
	return s.SendNotification(ctx, req.UserID, req.Type, req.Title, req.Message, req.Data)
}

// NotificationListResponse represents notification list with pagination
type NotificationListResponse struct {
	Data       []models.Notification `json:"data"`
	Pagination types.PaginationInfo  `json:"pagination"`
}

// GetUserNotifications retrieves notifications for a user with pagination
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uint, filter *repository.NotificationFilter, page, limit int) (*NotificationListResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get notifications from repository
	notifications, err := s.notificationRepo.List(filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	// Get total count
	total, err := s.notificationRepo.Count(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &NotificationListResponse{
		Data: notifications,
		Pagination: types.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uint, userID uint) error {
	// Get notification
	notification, err := s.notificationRepo.GetByID(notificationID)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	// Verify ownership
	if notification.UserID != userID {
		return fmt.Errorf("user is not the owner of this notification")
	}

	// Mark as read
	notification.IsRead = true

	// Update in database
	if err := s.notificationRepo.Update(notification); err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uint) (int, error) {
	// Update all unread notifications
	count, err := s.notificationRepo.MarkAllAsReadForUser(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return int(count), nil
}
