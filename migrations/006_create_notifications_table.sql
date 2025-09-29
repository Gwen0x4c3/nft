-- Migration: 006_create_notifications_table
-- Description: Create the notifications table for real-time user notifications
-- Created: 2025-09-29
-- Dependencies: 001_create_users_table

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    message VARCHAR(500) NOT NULL,
    data JSONB DEFAULT '{}',
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Foreign key constraints
ALTER TABLE notifications ADD CONSTRAINT fk_notifications_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX IF NOT EXISTS idx_notifications_user_created ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_type_created ON notifications(type, created_at DESC);

-- JSONB indexes for efficient data queries
CREATE INDEX IF NOT EXISTS idx_notifications_data_nft_id ON notifications USING GIN ((data->'nft_id')) WHERE data ? 'nft_id';
CREATE INDEX IF NOT EXISTS idx_notifications_data_auction_id ON notifications USING GIN ((data->'auction_id')) WHERE data ? 'auction_id';
CREATE INDEX IF NOT EXISTS idx_notifications_data_bid_id ON notifications USING GIN ((data->'bid_id')) WHERE data ? 'bid_id';

-- General JSONB index for flexible queries
CREATE INDEX IF NOT EXISTS idx_notifications_data_gin ON notifications USING GIN (data);

-- Check constraints
ALTER TABLE notifications ADD CONSTRAINT chk_notifications_type CHECK (
    type IN ('bid_placed', 'bid_outbid', 'auction_won', 'auction_ended', 'nft_sold')
);
ALTER TABLE notifications ADD CONSTRAINT chk_notifications_title_length CHECK (
    LENGTH(title) >= 1 AND LENGTH(title) <= 100
);
ALTER TABLE notifications ADD CONSTRAINT chk_notifications_message_length CHECK (
    LENGTH(message) >= 1 AND LENGTH(message) <= 500
);

-- Comments
COMMENT ON TABLE notifications IS 'Real-time user notifications for platform events';
COMMENT ON COLUMN notifications.id IS 'Primary key for the notification';
COMMENT ON COLUMN notifications.user_id IS 'Foreign key to users table';
COMMENT ON COLUMN notifications.type IS 'Notification type: bid_placed, bid_outbid, auction_won, auction_ended, nft_sold';
COMMENT ON COLUMN notifications.title IS 'Notification title (1-100 characters)';
COMMENT ON COLUMN notifications.message IS 'Notification message (1-500 characters)';
COMMENT ON COLUMN notifications.data IS 'Additional context data as JSON';
COMMENT ON COLUMN notifications.is_read IS 'Whether notification has been read by user';
COMMENT ON COLUMN notifications.created_at IS 'Timestamp when notification was created';

-- Function to clean up old read notifications
CREATE OR REPLACE FUNCTION cleanup_old_notifications()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete read notifications older than 90 days
    DELETE FROM notifications 
    WHERE is_read = true 
    AND created_at < CURRENT_TIMESTAMP - INTERVAL '90 days';
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_notifications() IS 'Removes read notifications older than 90 days';

-- Function to mark all notifications as read for a user
CREATE OR REPLACE FUNCTION mark_all_notifications_read(target_user_id BIGINT)
RETURNS INTEGER AS $$
DECLARE
    updated_count INTEGER;
BEGIN
    UPDATE notifications 
    SET is_read = true 
    WHERE user_id = target_user_id AND is_read = false;
    
    GET DIAGNOSTICS updated_count = ROW_COUNT;
    
    RETURN updated_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION mark_all_notifications_read(BIGINT) IS 'Marks all unread notifications as read for a specific user';

-- Function to get unread notification count for a user
CREATE OR REPLACE FUNCTION get_unread_notification_count(target_user_id BIGINT)
RETURNS INTEGER AS $$
DECLARE
    unread_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO unread_count
    FROM notifications 
    WHERE user_id = target_user_id AND is_read = false;
    
    RETURN unread_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_unread_notification_count(BIGINT) IS 'Returns count of unread notifications for a specific user';

-- Function to create notifications based on common templates
CREATE OR REPLACE FUNCTION create_notification(
    p_user_id BIGINT,
    p_type VARCHAR(50),
    p_title VARCHAR(100),
    p_message VARCHAR(500),
    p_data JSONB DEFAULT '{}'
)
RETURNS BIGINT AS $$
DECLARE
    notification_id BIGINT;
BEGIN
    -- Validate notification type
    IF p_type NOT IN ('bid_placed', 'bid_outbid', 'auction_won', 'auction_ended', 'nft_sold') THEN
        RAISE EXCEPTION 'Invalid notification type: %', p_type;
    END IF;
    
    -- Insert notification
    INSERT INTO notifications (user_id, type, title, message, data)
    VALUES (p_user_id, p_type, p_title, p_message, p_data)
    RETURNING id INTO notification_id;
    
    RETURN notification_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_notification(BIGINT, VARCHAR, VARCHAR, VARCHAR, JSONB) IS 'Creates a new notification with validation';

-- View for notification summary by type
CREATE OR REPLACE VIEW notification_summary AS
SELECT 
    type,
    COUNT(*) as total_count,
    COUNT(*) FILTER (WHERE is_read = false) as unread_count,
    COUNT(*) FILTER (WHERE is_read = true) as read_count,
    MIN(created_at) as oldest_notification,
    MAX(created_at) as newest_notification
FROM notifications
GROUP BY type
ORDER BY total_count DESC;

COMMENT ON VIEW notification_summary IS 'Summary statistics for notifications by type';

-- View for recent notifications with parsed data
CREATE OR REPLACE VIEW recent_notifications AS
SELECT 
    n.id,
    n.user_id,
    u.username,
    n.type,
    n.title,
    n.message,
    n.data,
    CASE 
        WHEN n.data ? 'nft_id' THEN (n.data->>'nft_id')::BIGINT
        ELSE NULL
    END as nft_id,
    CASE 
        WHEN n.data ? 'auction_id' THEN (n.data->>'auction_id')::BIGINT
        ELSE NULL
    END as auction_id,
    CASE 
        WHEN n.data ? 'amount' THEN n.data->>'amount'
        ELSE NULL
    END as amount_wei,
    n.is_read,
    n.created_at
FROM notifications n
JOIN users u ON n.user_id = u.id
WHERE n.created_at > CURRENT_TIMESTAMP - INTERVAL '30 days'
ORDER BY n.created_at DESC;

COMMENT ON VIEW recent_notifications IS 'Recent notifications with parsed data fields and user details';