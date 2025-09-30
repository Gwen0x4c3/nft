package types

import (
	"nft-platform/internal/models"
)

// CreateNotificationRequest represents the notification creation request
type CreateNotificationRequest struct {
	UserID  uint                   `json:"user_id" validate:"required"`
	Type    string                 `json:"type" validate:"required,oneof=bid_placed bid_outbid auction_won auction_ended nft_sold"`
	Title   string                 `json:"title" validate:"required,max=100"`
	Message string                 `json:"message" validate:"required,max=500"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NotificationListResponse represents notification list with pagination
type NotificationListResponse struct {
	Data       []models.Notification `json:"data"`
	Pagination PaginationInfo        `json:"pagination"`
}

// BulkNotificationRequest represents bulk notification creation request
type BulkNotificationRequest struct {
	Notifications []CreateNotificationRequest `json:"notifications" validate:"required,min=1,max=100"`
}
