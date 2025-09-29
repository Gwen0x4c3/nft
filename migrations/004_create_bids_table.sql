-- Migration: 004_create_bids_table
-- Description: Create the bids table for individual bid records within auctions
-- Created: 2025-09-29
-- Dependencies: 001_create_users_table, 003_create_auctions_table

CREATE TABLE IF NOT EXISTS bids (
    id BIGSERIAL PRIMARY KEY,
    auction_id BIGINT NOT NULL,
    bidder_id BIGINT NOT NULL,
    amount NUMERIC(78,0) NOT NULL, -- Big integer for Wei amounts
    tx_hash VARCHAR(66), -- Ethereum transaction hash (0x + 64 hex chars)
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Foreign key constraints
ALTER TABLE bids ADD CONSTRAINT fk_bids_auction_id FOREIGN KEY (auction_id) REFERENCES auctions(id) ON DELETE CASCADE;
ALTER TABLE bids ADD CONSTRAINT fk_bids_bidder_id FOREIGN KEY (bidder_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_bids_auction_id ON bids(auction_id);
CREATE INDEX IF NOT EXISTS idx_bids_bidder_id ON bids(bidder_id);
CREATE INDEX IF NOT EXISTS idx_bids_amount ON bids(amount DESC);
CREATE INDEX IF NOT EXISTS idx_bids_status ON bids(status);
CREATE INDEX IF NOT EXISTS idx_bids_created_at ON bids(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bids_tx_hash ON bids(tx_hash) WHERE tx_hash IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_bids_auction_amount ON bids(auction_id, amount DESC);
CREATE INDEX IF NOT EXISTS idx_bids_auction_status ON bids(auction_id, status);
CREATE INDEX IF NOT EXISTS idx_bids_bidder_created ON bids(bidder_id, created_at DESC);

-- Check constraints
ALTER TABLE bids ADD CONSTRAINT chk_bids_status CHECK (
    status IN ('pending', 'confirmed', 'failed')
);
ALTER TABLE bids ADD CONSTRAINT chk_bids_amount_positive CHECK (amount > 0);
ALTER TABLE bids ADD CONSTRAINT chk_bids_tx_hash_format CHECK (
    tx_hash IS NULL OR tx_hash ~ '^0x[a-fA-F0-9]{64}$'
);

-- Business logic constraints
-- Prevent bidder from bidding on their own auction
ALTER TABLE bids ADD CONSTRAINT chk_bids_not_own_auction CHECK (
    bidder_id != (SELECT seller_id FROM auctions WHERE auctions.id = bids.auction_id)
);

-- Unique constraint to prevent duplicate pending bids from same user on same auction
CREATE UNIQUE INDEX IF NOT EXISTS idx_bids_unique_pending ON bids(auction_id, bidder_id) 
WHERE status = 'pending';

-- Comments
COMMENT ON TABLE bids IS 'Individual bid records within auctions';
COMMENT ON COLUMN bids.id IS 'Primary key for the bid';
COMMENT ON COLUMN bids.auction_id IS 'Foreign key to auctions table';
COMMENT ON COLUMN bids.bidder_id IS 'Foreign key to users table (bidder)';
COMMENT ON COLUMN bids.amount IS 'Bid amount in Wei';
COMMENT ON COLUMN bids.tx_hash IS 'Ethereum transaction hash when confirmed';
COMMENT ON COLUMN bids.status IS 'Bid status: pending, confirmed, failed';
COMMENT ON COLUMN bids.created_at IS 'Timestamp when bid was created';
COMMENT ON COLUMN bids.updated_at IS 'Timestamp when bid was last updated';

-- Update trigger for updated_at
CREATE TRIGGER update_bids_updated_at 
    BEFORE UPDATE ON bids 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to update auction current_bid when a bid is confirmed
CREATE OR REPLACE FUNCTION update_auction_current_bid()
RETURNS TRIGGER AS $$
BEGIN
    -- When a bid is confirmed, update the auction's current_bid if this is the highest
    IF NEW.status = 'confirmed' AND (OLD.status != 'confirmed' OR OLD.status IS NULL) THEN
        UPDATE auctions 
        SET current_bid = (
            SELECT COALESCE(MAX(amount), 0) 
            FROM bids 
            WHERE auction_id = NEW.auction_id AND status = 'confirmed'
        ),
        winner_id = (
            SELECT bidder_id 
            FROM bids 
            WHERE auction_id = NEW.auction_id AND status = 'confirmed' 
            ORDER BY amount DESC, created_at ASC 
            LIMIT 1
        )
        WHERE id = NEW.auction_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER bid_update_auction
    AFTER INSERT OR UPDATE ON bids
    FOR EACH ROW
    EXECUTE FUNCTION update_auction_current_bid();

-- Function to validate bid amount against current auction state
CREATE OR REPLACE FUNCTION validate_bid_amount()
RETURNS TRIGGER AS $$
DECLARE
    auction_record auctions%ROWTYPE;
    current_highest_bid NUMERIC(78,0);
    min_increment NUMERIC(78,0) := 1000000000000000; -- 0.001 ETH in Wei as minimum increment
BEGIN
    -- Get auction details
    SELECT * INTO auction_record FROM auctions WHERE id = NEW.auction_id;
    
    -- Check if auction exists and is active
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Auction not found';
    END IF;
    
    IF auction_record.status != 'active' THEN
        RAISE EXCEPTION 'Auction is not active';
    END IF;
    
    -- Check if auction is still within time bounds
    IF CURRENT_TIMESTAMP >= auction_record.end_time THEN
        RAISE EXCEPTION 'Auction has ended';
    END IF;
    
    -- Get current highest bid
    SELECT COALESCE(MAX(amount), 0) INTO current_highest_bid 
    FROM bids 
    WHERE auction_id = NEW.auction_id AND status = 'confirmed';
    
    -- If no bids yet, compare against start price
    IF current_highest_bid = 0 THEN
        current_highest_bid = auction_record.start_price;
    END IF;
    
    -- Validate bid amount
    IF NEW.amount <= current_highest_bid THEN
        RAISE EXCEPTION 'Bid amount must be higher than current highest bid';
    END IF;
    
    -- Ensure minimum increment
    IF NEW.amount < current_highest_bid + min_increment THEN
        RAISE EXCEPTION 'Bid increment too small';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER bid_amount_validation
    BEFORE INSERT ON bids
    FOR EACH ROW
    EXECUTE FUNCTION validate_bid_amount();