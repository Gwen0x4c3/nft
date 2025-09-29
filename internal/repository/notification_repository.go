package repository

import (
	"errors"
	"fmt"
	"time"

	"nft-platform/internal/models"
	"gorm.io/gorm"
)

// NotificationFilter represents filtering options for notification queries
type NotificationFilter struct {
	UserID    *uint
	Type      *string
	IsRead    *bool
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// NotificationRepository handles database operations for Notification entities
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository instance
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification in the database
func (r *NotificationRepository) Create(notification *models.Notification) error {
	if err := r.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

// GetByID retrieves a notification by its ID
func (r *NotificationRepository) GetByID(id uint) (*models.Notification, error) {
	var notification models.Notification
	if err := r.db.Preload("User").First(&notification, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("notification with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get notification by id: %w", err)
	}
	return &notification, nil
}

// Update updates an existing notification
func (r *NotificationRepository) Update(notification *models.Notification) error {
	if err := r.db.Save(notification).Error; err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}
	return nil
}

// Delete deletes a notification from the database
func (r *NotificationRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Notification{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}
	return nil
}

// List retrieves notifications with filtering and pagination
func (r *NotificationRepository) List(filter *NotificationFilter, offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	query := r.db.Model(&models.Notification{}).Preload("User")

	// Apply filters
	if filter != nil {
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.IsRead != nil {
			query = query.Where("is_read = ?", *filter.IsRead)
		}
		if filter.CreatedAfter != nil {
			query = query.Where("created_at > ?", *filter.CreatedAfter)
		}
		if filter.CreatedBefore != nil {
			query = query.Where("created_at < ?", *filter.CreatedBefore)
		}
	}

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	return notifications, nil
}

// Count returns the total number of notifications matching the filter
func (r *NotificationRepository) Count(filter *NotificationFilter) (int64, error) {
	var count int64
	query := r.db.Model(&models.Notification{})

	// Apply same filters as List
	if filter != nil {
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.IsRead != nil {
			query = query.Where("is_read = ?", *filter.IsRead)
		}
		if filter.CreatedAfter != nil {
			query = query.Where("created_at > ?", *filter.CreatedAfter)
		}
		if filter.CreatedBefore != nil {
			query = query.Where("created_at < ?", *filter.CreatedBefore)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	return count, nil
}

// GetByUser retrieves notifications for a specific user
func (r *NotificationRepository) GetByUser(userID uint, offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	query := r.db.Where("user_id = ?", userID).Preload("User")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications by user: %w", err)
	}

	return notifications, nil
}

// GetUnreadByUser retrieves unread notifications for a specific user
func (r *NotificationRepository) GetUnreadByUser(userID uint, offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	query := r.db.Where("user_id = ? AND is_read = ?", userID, false).Preload("User")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}

	return notifications, nil
}

// GetByType retrieves notifications of a specific type
func (r *NotificationRepository) GetByType(notificationType string, offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	query := r.db.Where("type = ?", notificationType).Preload("User")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications by type: %w", err)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(id uint) error {
	result := r.db.Model(&models.Notification{}).Where("id = ?", id).Update("is_read", true)
	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as read: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}
	return nil
}

// MarkAsUnread marks a notification as unread
func (r *NotificationRepository) MarkAsUnread(id uint) error {
	result := r.db.Model(&models.Notification{}).Where("id = ?", id).Update("is_read", false)
	if result.Error != nil {
		return fmt.Errorf("failed to mark notification as unread: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}
	return nil
}

// MarkAllAsReadForUser marks all notifications as read for a specific user
func (r *NotificationRepository) MarkAllAsReadForUser(userID uint) (int64, error) {
	result := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark all notifications as read: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get unread notification count: %w", err)
	}
	return count, nil
}

// GetRecentNotifications retrieves recent notifications across all users
func (r *NotificationRepository) GetRecentNotifications(limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	if limit <= 0 {
		limit = 50 // Default limit
	}

	if err := r.db.Preload("User").Order("created_at DESC").Limit(limit).Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent notifications: %w", err)
	}

	return notifications, nil
}

// GetNotificationsByTimeRange retrieves notifications within a time range
func (r *NotificationRepository) GetNotificationsByTimeRange(startTime, endTime time.Time, offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	query := r.db.Where("created_at >= ? AND created_at <= ?", startTime, endTime).Preload("User")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications by time range: %w", err)
	}

	return notifications, nil
}

// DeleteOldReadNotifications deletes read notifications older than the specified duration
func (r *NotificationRepository) DeleteOldReadNotifications(olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	
	result := r.db.Where("is_read = ? AND created_at < ?", true, cutoffTime).Delete(&models.Notification{})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete old notifications: %w", result.Error)
	}
	
	return result.RowsAffected, nil
}

// BulkCreate creates multiple notifications in a single transaction
func (r *NotificationRepository) BulkCreate(notifications []models.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	if err := r.db.CreateInBatches(notifications, 100).Error; err != nil {
		return fmt.Errorf("failed to bulk create notifications: %w", err)
	}
	
	return nil
}

// Exists checks if a notification exists by ID
func (r *NotificationRepository) Exists(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Notification{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check notification existence: %w", err)
	}
	return count > 0, nil
}

// GetNotificationsByDataField retrieves notifications containing specific data field
func (r *NotificationRepository) GetNotificationsByDataField(fieldName string, fieldValue interface{}, offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	query := r.db.Where("data->>? = ?", fieldName, fmt.Sprintf("%v", fieldValue)).Preload("User")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications by data field: %w", err)
	}

	return notifications, nil
}

// GetNotificationStats returns notification statistics
func (r *NotificationRepository) GetNotificationStats() (*NotificationStats, error) {
	stats := &NotificationStats{}

	// Total notifications
	if err := r.db.Model(&models.Notification{}).Count(&stats.TotalNotifications).Error; err != nil {
		return nil, fmt.Errorf("failed to count total notifications: %w", err)
	}

	// Unread notifications
	if err := r.db.Model(&models.Notification{}).Where("is_read = ?", false).Count(&stats.UnreadNotifications).Error; err != nil {
		return nil, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	// Read notifications
	if err := r.db.Model(&models.Notification{}).Where("is_read = ?", true).Count(&stats.ReadNotifications).Error; err != nil {
		return nil, fmt.Errorf("failed to count read notifications: %w", err)
	}

	// Notifications by type
	var typeStats []TypeStat
	if err := r.db.Model(&models.Notification{}).
		Select("type, count(*) as count").
		Group("type").
		Scan(&typeStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications by type: %w", err)
	}
	stats.NotificationsByType = typeStats

	return stats, nil
}

// GetNotificationTypeCounts returns counts grouped by notification type
func (r *NotificationRepository) GetNotificationTypeCounts(userID *uint) (map[string]int64, error) {
	var results []struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	query := r.db.Model(&models.Notification{}).
		Select("type, count(*) as count").
		Group("type")

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get notification type counts: %w", err)
	}

	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.Type] = result.Count
	}

	return counts, nil
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalNotifications    int64      `json:"total_notifications"`
	UnreadNotifications   int64      `json:"unread_notifications"`
	ReadNotifications     int64      `json:"read_notifications"`
	NotificationsByType   []TypeStat `json:"notifications_by_type"`
}

// TypeStat represents notification count by type
type TypeStat struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}