-- Migration: 003_create_auctions_table
-- Description: Create the auctions table for time-bound competitive bidding on NFTs
-- Created: 2025-09-29
-- Dependencies: 001_create_users_table, 002_create_nfts_table

CREATE TABLE IF NOT EXISTS auctions (
    id BIGSERIAL PRIMARY KEY,
    nft_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    start_price NUMERIC(78,0) NOT NULL, -- Big integer for Wei amounts
    reserve_price NUMERIC(78,0), -- Minimum acceptable price (nullable)
    current_bid NUMERIC(78,0) DEFAULT 0, -- Current highest bid
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    winner_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Foreign key constraints
ALTER TABLE auctions ADD CONSTRAINT fk_auctions_nft_id FOREIGN KEY (nft_id) REFERENCES nfts(id) ON DELETE RESTRICT;
ALTER TABLE auctions ADD CONSTRAINT fk_auctions_seller_id FOREIGN KEY (seller_id) REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE auctions ADD CONSTRAINT fk_auctions_winner_id FOREIGN KEY (winner_id) REFERENCES users(id) ON DELETE SET NULL;

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_auctions_nft_id ON auctions(nft_id);
CREATE INDEX IF NOT EXISTS idx_auctions_seller_id ON auctions(seller_id);
CREATE INDEX IF NOT EXISTS idx_auctions_status ON auctions(status);
CREATE INDEX IF NOT EXISTS idx_auctions_end_time ON auctions(end_time);
CREATE INDEX IF NOT EXISTS idx_auctions_start_time ON auctions(start_time);
CREATE INDEX IF NOT EXISTS idx_auctions_current_bid ON auctions(current_bid DESC);
CREATE INDEX IF NOT EXISTS idx_auctions_created_at ON auctions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_auctions_winner_id ON auctions(winner_id) WHERE winner_id IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_auctions_status_end_time ON auctions(status, end_time) WHERE status IN ('pending', 'active');
CREATE INDEX IF NOT EXISTS idx_auctions_active ON auctions(start_time, end_time) WHERE status = 'active';

-- Check constraints
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_status CHECK (
    status IN ('pending', 'active', 'ended', 'cancelled')
);
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_start_price_positive CHECK (start_price > 0);
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_reserve_price_positive CHECK (
    reserve_price IS NULL OR reserve_price > 0
);
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_current_bid_non_negative CHECK (current_bid >= 0);
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_end_after_start CHECK (end_time > start_time);
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_minimum_duration CHECK (
    end_time > start_time + INTERVAL '1 hour'
);
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_reserve_gte_start CHECK (
    reserve_price IS NULL OR reserve_price >= start_price
);

-- Business logic constraints
ALTER TABLE auctions ADD CONSTRAINT chk_auctions_winner_only_when_ended CHECK (
    (status != 'ended') OR (winner_id IS NOT NULL) OR (current_bid = 0)
);

-- Comments
COMMENT ON TABLE auctions IS 'Time-bound competitive bidding for NFTs';
COMMENT ON COLUMN auctions.id IS 'Primary key for the auction';
COMMENT ON COLUMN auctions.nft_id IS 'Foreign key to nfts table';
COMMENT ON COLUMN auctions.seller_id IS 'Foreign key to users table (auction creator)';
COMMENT ON COLUMN auctions.start_price IS 'Starting bid price in Wei';
COMMENT ON COLUMN auctions.reserve_price IS 'Minimum acceptable price in Wei (optional)';
COMMENT ON COLUMN auctions.current_bid IS 'Current highest bid in Wei';
COMMENT ON COLUMN auctions.start_time IS 'When auction becomes active';
COMMENT ON COLUMN auctions.end_time IS 'When auction ends';
COMMENT ON COLUMN auctions.status IS 'Auction status: pending, active, ended, cancelled';
COMMENT ON COLUMN auctions.winner_id IS 'Foreign key to users table (auction winner)';
COMMENT ON COLUMN auctions.created_at IS 'Timestamp when auction was created';
COMMENT ON COLUMN auctions.updated_at IS 'Timestamp when auction was last updated';

-- Update trigger for updated_at
CREATE TRIGGER update_auctions_updated_at 
    BEFORE UPDATE ON auctions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically update auction status based on time
CREATE OR REPLACE FUNCTION update_auction_status()
RETURNS TRIGGER AS $$
BEGIN
    -- Auto-activate auctions when start_time is reached
    IF OLD.status = 'pending' AND NEW.start_time <= CURRENT_TIMESTAMP THEN
        NEW.status = 'active';
    END IF;
    
    -- Auto-end auctions when end_time is reached
    IF OLD.status = 'active' AND NEW.end_time <= CURRENT_TIMESTAMP THEN
        NEW.status = 'ended';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auction_status_update
    BEFORE UPDATE ON auctions
    FOR EACH ROW
    EXECUTE FUNCTION update_auction_status();