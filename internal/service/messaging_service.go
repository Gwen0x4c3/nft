package service

import (
	"context"

	"nft-platform/internal/models"
)

// MessagingService interface for external messaging (Kafka, etc.)
type MessagingService interface {
	PublishNotification(ctx context.Context, notification *models.Notification) error
	PublishEvent(ctx context.Context, event string, data interface{}) error
}
