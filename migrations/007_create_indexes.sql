-- Migration: 007_create_indexes
-- Description: Create additional performance indexes as defined in data-model.md
-- Created: 2025-09-29
-- Dependencies: All table creation migrations

-- Note: Many of these indexes are already created in the table creation migrations
-- This file ensures all indexes from data-model.md are present and adds any missing ones

-- =============================================================================
-- USER INDEXES (from data-model.md)
-- =============================================================================

-- Most user indexes already created in 001_create_users_table.sql
-- Adding any missing performance indexes

CREATE INDEX IF NOT EXISTS idx_users_username_lower ON users(LOWER(username));
CREATE INDEX IF NOT EXISTS idx_users_verified_users ON users(is_verified, created_at DESC) WHERE is_verified = true;

-- =============================================================================
-- NFT INDEXES (from data-model.md) 
-- =============================================================================

-- Core NFT indexes already created in 002_create_nfts_table.sql
-- Adding comprehensive performance indexes

-- Indexes for NFT marketplace queries
CREATE INDEX IF NOT EXISTS idx_nfts_sale_price ON nfts(price DESC, created_at DESC) WHERE is_for_sale = true;
CREATE INDEX IF NOT EXISTS idx_nfts_creator_recent ON nfts(creator_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_nfts_owner_recent ON nfts(owner_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_nfts_royalty ON nfts(royalty) WHERE royalty > 0;

-- Indexes for search and filtering
CREATE INDEX IF NOT EXISTS idx_nfts_title_search ON nfts USING GIN (to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_nfts_description_search ON nfts USING GIN (to_tsvector('english', description)) WHERE description IS NOT NULL;

-- Contract-specific indexes
CREATE INDEX IF NOT EXISTS idx_nfts_contract_token ON nfts(contract_address, token_id);

-- =============================================================================
-- AUCTION INDEXES (from data-model.md)
-- =============================================================================

-- Most auction indexes already created in 003_create_auctions_table.sql
-- Adding specialized performance indexes

-- Indexes for active auction queries
CREATE INDEX IF NOT EXISTS idx_auctions_ending_soon ON auctions(end_time ASC) 
    WHERE status = 'active' AND end_time > CURRENT_TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_auctions_recently_ended ON auctions(end_time DESC) 
    WHERE status = 'ended' AND end_time > CURRENT_TIMESTAMP - INTERVAL '7 days';

-- Indexes for auction analytics
CREATE INDEX IF NOT EXISTS idx_auctions_reserve_met ON auctions(status, current_bid, reserve_price) 
    WHERE reserve_price IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_auctions_no_bids ON auctions(status, current_bid) WHERE current_bid = 0;

-- Seller performance indexes
CREATE INDEX IF NOT EXISTS idx_auctions_seller_status ON auctions(seller_id, status, created_at DESC);

-- =============================================================================
-- BID INDEXES (from data-model.md)
-- =============================================================================

-- Most bid indexes already created in 004_create_bids_table.sql
-- Adding comprehensive bidding analysis indexes

-- Indexes for bidding history and analytics
CREATE INDEX IF NOT EXISTS idx_bids_confirmed_amount ON bids(amount DESC, created_at DESC) WHERE status = 'confirmed';
CREATE INDEX IF NOT EXISTS idx_bids_pending_old ON bids(created_at ASC) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_bids_failed ON bids(auction_id, created_at DESC) WHERE status = 'failed';

-- User bidding behavior indexes
CREATE INDEX IF NOT EXISTS idx_bids_bidder_won ON bids(bidder_id, created_at DESC) 
    WHERE status = 'confirmed' AND bidder_id IN (SELECT winner_id FROM auctions WHERE winner_id IS NOT NULL);
CREATE INDEX IF NOT EXISTS idx_bids_bidder_activity ON bids(bidder_id, status, created_at DESC);

-- =============================================================================
-- TRANSFER INDEXES (from data-model.md)
-- =============================================================================

-- Most transfer indexes already created in 005_create_transfers_table.sql
-- Adding comprehensive transfer tracking indexes

-- Sales analytics indexes
CREATE INDEX IF NOT EXISTS idx_transfers_sales_recent ON transfers(created_at DESC, price DESC) WHERE type = 'sale';
CREATE INDEX IF NOT EXISTS idx_transfers_sales_volume ON transfers(price DESC, created_at DESC) WHERE type = 'sale' AND price IS NOT NULL;

-- Mint tracking indexes
CREATE INDEX IF NOT EXISTS idx_transfers_mints_recent ON transfers(to_id, created_at DESC) WHERE type = 'mint';
CREATE INDEX IF NOT EXISTS idx_transfers_mints_daily ON transfers(DATE(created_at), to_id) WHERE type = 'mint';

-- User trading activity indexes
CREATE INDEX IF NOT EXISTS idx_transfers_user_sent ON transfers(from_id, type, created_at DESC) WHERE from_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transfers_user_received ON transfers(to_id, type, created_at DESC);

-- NFT ownership history indexes
CREATE INDEX IF NOT EXISTS idx_transfers_nft_chronological ON transfers(nft_id, created_at ASC);

-- =============================================================================
-- NOTIFICATION INDEXES (from data-model.md)
-- =============================================================================

-- Most notification indexes already created in 006_create_notifications_table.sql
-- Adding specialized notification query indexes

-- User notification management indexes
CREATE INDEX IF NOT EXISTS idx_notifications_priority ON notifications(user_id, type, created_at DESC) 
    WHERE type IN ('auction_won', 'bid_outbid');
CREATE INDEX IF NOT EXISTS idx_notifications_recent_unread ON notifications(created_at DESC) 
    WHERE is_read = false AND created_at > CURRENT_TIMESTAMP - INTERVAL '24 hours';

-- Notification cleanup indexes
CREATE INDEX IF NOT EXISTS idx_notifications_old_read ON notifications(created_at ASC) 
    WHERE is_read = true AND created_at < CURRENT_TIMESTAMP - INTERVAL '30 days';

-- =============================================================================
-- CROSS-TABLE PERFORMANCE INDEXES
-- =============================================================================

-- Indexes for complex queries involving multiple tables

-- User portfolio queries
CREATE INDEX IF NOT EXISTS idx_user_portfolio_nfts ON nfts(owner_id, is_for_sale, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_created_nfts ON nfts(creator_id, created_at DESC);

-- Auction participation queries  
CREATE INDEX IF NOT EXISTS idx_user_auction_participation ON bids(bidder_id, auction_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_auction_wins ON auctions(winner_id, status, end_time DESC) WHERE winner_id IS NOT NULL;

-- Market activity queries
CREATE INDEX IF NOT EXISTS idx_market_active_auctions ON auctions(status, end_time, current_bid DESC) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_market_recent_sales ON transfers(type, created_at DESC, price DESC) WHERE type = 'sale';

-- =============================================================================
-- FULL-TEXT SEARCH INDEXES
-- =============================================================================

-- Full-text search for NFTs
CREATE INDEX IF NOT EXISTS idx_nfts_fulltext ON nfts USING GIN (
    to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(description, ''))
);

-- Full-text search for users
CREATE INDEX IF NOT EXISTS idx_users_fulltext ON users USING GIN (
    to_tsvector('english', COALESCE(username, '') || ' ' || COALESCE(bio, ''))
);

-- =============================================================================
-- SPECIALIZED BUSINESS LOGIC INDEXES
-- =============================================================================

-- Indexes for auction expiration processing
CREATE INDEX IF NOT EXISTS idx_auctions_expiring ON auctions(end_time ASC, status) 
    WHERE status IN ('pending', 'active');

-- Indexes for royalty calculations
CREATE INDEX IF NOT EXISTS idx_nfts_with_royalties ON nfts(creator_id, royalty) WHERE royalty > 0;

-- Indexes for bid validation
CREATE INDEX IF NOT EXISTS idx_active_auction_bids ON bids(auction_id, amount DESC, created_at DESC) 
    WHERE auction_id IN (SELECT id FROM auctions WHERE status = 'active');

-- Indexes for notification delivery optimization
CREATE INDEX IF NOT EXISTS idx_notifications_delivery ON notifications(user_id, type, is_read, created_at DESC)
    WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '7 days';

-- =============================================================================
-- ANALYTICS AND REPORTING INDEXES
-- =============================================================================

-- Daily activity tracking
CREATE INDEX IF NOT EXISTS idx_daily_nft_mints ON transfers(DATE(created_at)) WHERE type = 'mint';
CREATE INDEX IF NOT EXISTS idx_daily_sales_volume ON transfers(DATE(created_at), price) WHERE type = 'sale' AND price IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_daily_auction_activity ON auctions(DATE(created_at), status);

-- User engagement metrics
CREATE INDEX IF NOT EXISTS idx_user_last_activity ON (
    SELECT user_id, MAX(created_at) as last_active FROM (
        SELECT user_id, created_at FROM notifications
        UNION ALL
        SELECT bidder_id as user_id, created_at FROM bids
        UNION ALL
        SELECT creator_id as user_id, created_at FROM nfts
    ) activities GROUP BY user_id
);

-- Market trend analysis
CREATE INDEX IF NOT EXISTS idx_price_trends ON transfers(created_at DESC, price DESC) 
    WHERE type = 'sale' AND price IS NOT NULL AND created_at > CURRENT_TIMESTAMP - INTERVAL '30 days';

-- =============================================================================
-- COMMENTS
-- =============================================================================

COMMENT ON INDEX idx_users_username_lower IS 'Case-insensitive username searches';
COMMENT ON INDEX idx_nfts_sale_price IS 'NFT marketplace sorting by price';
COMMENT ON INDEX idx_nfts_fulltext IS 'Full-text search across NFT title and description';
COMMENT ON INDEX idx_auctions_ending_soon IS 'Find auctions ending soon for notifications';
COMMENT ON INDEX idx_transfers_sales_volume IS 'Sales volume analysis and reporting';
COMMENT ON INDEX idx_notifications_priority IS 'High-priority notification delivery';

-- =============================================================================
-- INDEX MAINTENANCE
-- =============================================================================

-- Function to monitor index usage
CREATE OR REPLACE FUNCTION get_unused_indexes()
RETURNS TABLE(
    schemaname TEXT,
    tablename TEXT,
    indexname TEXT,
    index_size TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.schemaname::TEXT,
        s.tablename::TEXT,
        s.indexname::TEXT,
        pg_size_pretty(pg_relation_size(i.indexrelid))::TEXT as index_size
    FROM pg_stat_user_indexes s
    JOIN pg_index i ON s.indexrelid = i.indexrelid
    WHERE s.idx_scan = 0
    AND i.indisunique = false
    AND i.indisprimary = false
    ORDER BY pg_relation_size(i.indexrelid) DESC;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_unused_indexes() IS 'Returns indexes that have never been used (candidates for removal)';

-- =============================================================================
-- VALIDATION
-- =============================================================================

-- Verify all required indexes exist
DO $$
DECLARE
    missing_indexes TEXT[] := ARRAY[]::TEXT[];
BEGIN
    -- Check for critical indexes from data-model.md
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_users_wallet_addr') THEN
        missing_indexes := array_append(missing_indexes, 'idx_users_wallet_addr');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_nfts_creator_id') THEN
        missing_indexes := array_append(missing_indexes, 'idx_nfts_creator_id');
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_auctions_status') THEN
        missing_indexes := array_append(missing_indexes, 'idx_auctions_status');
    END IF;
    
    IF array_length(missing_indexes, 1) > 0 THEN
        RAISE WARNING 'Missing critical indexes: %', array_to_string(missing_indexes, ', ');
    END IF;
    
    RAISE NOTICE 'Index validation completed. Total indexes created in this migration: %', 
        (SELECT COUNT(*) FROM pg_indexes WHERE tablename IN ('users', 'nfts', 'auctions', 'bids', 'transfers', 'notifications'));
END;
$$;