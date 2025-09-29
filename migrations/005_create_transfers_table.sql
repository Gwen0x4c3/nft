-- Migration: 005_create_transfers_table
-- Description: Create the transfers table for NFT ownership transfer history
-- Created: 2025-09-29
-- Dependencies: 001_create_users_table, 002_create_nfts_table

CREATE TABLE IF NOT EXISTS transfers (
    id BIGSERIAL PRIMARY KEY,
    nft_id BIGINT NOT NULL,
    from_id BIGINT, -- Nullable for mint transactions
    to_id BIGINT NOT NULL,
    tx_hash VARCHAR(66) NOT NULL, -- Ethereum transaction hash (0x + 64 hex chars)
    price NUMERIC(78,0), -- Big integer for Wei amounts (nullable for non-sale transfers)
    type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Foreign key constraints
ALTER TABLE transfers ADD CONSTRAINT fk_transfers_nft_id FOREIGN KEY (nft_id) REFERENCES nfts(id) ON DELETE CASCADE;
ALTER TABLE transfers ADD CONSTRAINT fk_transfers_from_id FOREIGN KEY (from_id) REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE transfers ADD CONSTRAINT fk_transfers_to_id FOREIGN KEY (to_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_transfers_nft_id ON transfers(nft_id);
CREATE INDEX IF NOT EXISTS idx_transfers_from_id ON transfers(from_id) WHERE from_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transfers_to_id ON transfers(to_id);
CREATE INDEX IF NOT EXISTS idx_transfers_type ON transfers(type);
CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON transfers(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transfers_tx_hash ON transfers(tx_hash);
CREATE INDEX IF NOT EXISTS idx_transfers_price ON transfers(price DESC) WHERE price IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_transfers_nft_created ON transfers(nft_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transfers_user_from ON transfers(from_id, created_at DESC) WHERE from_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_transfers_user_to ON transfers(to_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transfers_type_created ON transfers(type, created_at DESC);

-- Check constraints
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_type CHECK (
    type IN ('mint', 'sale', 'transfer')
);
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_tx_hash_format CHECK (
    tx_hash ~ '^0x[a-fA-F0-9]{64}$'
);
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_price_positive CHECK (
    price IS NULL OR price > 0
);

-- Business logic constraints based on transfer type
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_mint_from_null CHECK (
    type != 'mint' OR from_id IS NULL
);
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_sale_has_price CHECK (
    type != 'sale' OR (price IS NOT NULL AND price > 0)
);
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_sale_has_from CHECK (
    type != 'sale' OR from_id IS NOT NULL
);
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_transfer_has_from CHECK (
    type != 'transfer' OR from_id IS NOT NULL
);

-- Prevent self-transfers (except for mints)
ALTER TABLE transfers ADD CONSTRAINT chk_transfers_not_self CHECK (
    type = 'mint' OR from_id != to_id
);

-- Unique constraint on transaction hash to prevent duplicate processing
CREATE UNIQUE INDEX IF NOT EXISTS idx_transfers_tx_hash_unique ON transfers(tx_hash);

-- Comments
COMMENT ON TABLE transfers IS 'NFT ownership transfer history tracking';
COMMENT ON COLUMN transfers.id IS 'Primary key for the transfer';
COMMENT ON COLUMN transfers.nft_id IS 'Foreign key to nfts table';
COMMENT ON COLUMN transfers.from_id IS 'Foreign key to users table (sender, null for mints)';
COMMENT ON COLUMN transfers.to_id IS 'Foreign key to users table (recipient)';
COMMENT ON COLUMN transfers.tx_hash IS 'Ethereum transaction hash (unique)';
COMMENT ON COLUMN transfers.price IS 'Transfer price in Wei (null for non-sale transfers)';
COMMENT ON COLUMN transfers.type IS 'Transfer type: mint, sale, transfer';
COMMENT ON COLUMN transfers.created_at IS 'Timestamp when transfer was recorded';

-- Function to update NFT owner when a transfer is recorded
CREATE OR REPLACE FUNCTION update_nft_owner_on_transfer()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the NFT owner to the recipient of this transfer
    UPDATE nfts 
    SET owner_id = NEW.to_id,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.nft_id;
    
    -- For sale transfers, clear the for_sale status and price
    IF NEW.type = 'sale' THEN
        UPDATE nfts 
        SET is_for_sale = FALSE,
            price = NULL,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = NEW.nft_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER transfer_update_nft_owner
    AFTER INSERT ON transfers
    FOR EACH ROW
    EXECUTE FUNCTION update_nft_owner_on_transfer();

-- Function to validate transfer consistency
CREATE OR REPLACE FUNCTION validate_transfer_consistency()
RETURNS TRIGGER AS $$
DECLARE
    current_owner_id BIGINT;
    nft_exists BOOLEAN;
BEGIN
    -- Check if NFT exists
    SELECT EXISTS(SELECT 1 FROM nfts WHERE id = NEW.nft_id) INTO nft_exists;
    IF NOT nft_exists THEN
        RAISE EXCEPTION 'NFT with id % does not exist', NEW.nft_id;
    END IF;
    
    -- For non-mint transfers, validate that from_id is the current owner
    IF NEW.type != 'mint' THEN
        SELECT owner_id INTO current_owner_id FROM nfts WHERE id = NEW.nft_id;
        
        IF NEW.from_id != current_owner_id THEN
            RAISE EXCEPTION 'Transfer from_id (%) does not match current NFT owner (%)', NEW.from_id, current_owner_id;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER transfer_consistency_validation
    BEFORE INSERT ON transfers
    FOR EACH ROW
    EXECUTE FUNCTION validate_transfer_consistency();

-- View for transfer history with user details
CREATE OR REPLACE VIEW transfer_history AS
SELECT 
    t.id,
    t.nft_id,
    n.title as nft_title,
    n.token_id,
    t.from_id,
    u_from.username as from_username,
    u_from.wallet_address as from_wallet,
    t.to_id,
    u_to.username as to_username,
    u_to.wallet_address as to_wallet,
    t.tx_hash,
    t.price,
    t.type,
    t.created_at
FROM transfers t
JOIN nfts n ON t.nft_id = n.id
LEFT JOIN users u_from ON t.from_id = u_from.id
JOIN users u_to ON t.to_id = u_to.id
ORDER BY t.created_at DESC;

COMMENT ON VIEW transfer_history IS 'Complete transfer history with user and NFT details';