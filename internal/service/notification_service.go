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
