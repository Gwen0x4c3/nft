package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nft-platform/internal/repository"
	"nft-platform/internal/service"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetNotifications godoc
// @Summary Get user notifications
// @Description Retrieves notifications for the authenticated user with optional filtering
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Param unread_only query bool false "Filter unread notifications only" default(false)
// @Success 200 {object} service.NotificationListResponse "Notifications retrieved"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /notifications [get]
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userIDUint := userID.(uint)

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Validate and limit page size
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Build filter for this user
	filter := repository.NotificationFilter{
		UserID: &userIDUint,
	}

	// Parse unread_only filter
	if unreadOnly := c.Query("unread_only"); unreadOnly != "" {
		if val, err := strconv.ParseBool(unreadOnly); err == nil && val {
			isRead := false
			filter.IsRead = &isRead
		}
	}

	// Get notifications from service
	notificationsResp, err := h.notificationService.GetUserNotifications(context.Background(), userIDUint, &filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to retrieve notifications",
			Code:    "FETCH_FAILED",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, notificationsResp)
}

// MarkNotificationAsRead godoc
// @Summary Mark notification as read
// @Description Marks a specific notification as read for the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationId path int true "Notification ID"
// @Success 200 {object} map[string]interface{} "Notification marked as read"
// @Failure 400 {object} ErrorResponse "Invalid notification ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - not notification owner"
// @Failure 404 {object} ErrorResponse "Notification not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /notifications/{notificationId}/read [put]
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userIDUint := userID.(uint)

	// Parse notification ID from path parameter
	notificationID, err := strconv.ParseUint(c.Param("notificationId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid notification ID",
			Code:    "INVALID_ID",
		})
		return
	}

	// Mark notification as read through service
	err = h.notificationService.MarkAsRead(context.Background(), uint(notificationID), userIDUint)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "UPDATE_FAILED"

		// Check for specific errors
		if err.Error() == "notification not found" {
			statusCode = http.StatusNotFound
			code = "NOT_FOUND"
		} else if err.Error() == "not notification owner" || err.Error() == "user is not the owner of this notification" {
			statusCode = http.StatusForbidden
			code = "FORBIDDEN"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to mark notification as read",
			Code:    code,
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Notification marked as read",
		"notification_id": notificationID,
	})
}

// MarkAllNotificationsAsRead godoc
// @Summary Mark all notifications as read
// @Description Marks all notifications as read for the authenticated user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "All notifications marked as read"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /notifications/read-all [put]
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userIDUint := userID.(uint)

	// Mark all notifications as read through service
	count, err := h.notificationService.MarkAllAsRead(context.Background(), userIDUint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "update_failed",
			Message: "Failed to mark all notifications as read",
			Code:    "UPDATE_FAILED",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
		"count":   count,
	})
}
