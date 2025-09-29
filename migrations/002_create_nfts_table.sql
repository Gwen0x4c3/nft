-- Migration: 002_create_nfts_table
-- Description: Create the nfts table for digital assets stored on blockchain
-- Created: 2025-09-29
-- Dependencies: 001_create_users_table

CREATE TABLE IF NOT EXISTS nfts (
    id BIGSERIAL PRIMARY KEY,
    token_id VARCHAR(255) NOT NULL,
    contract_address VARCHAR(42) NOT NULL,
    creator_id BIGINT NOT NULL,
    owner_id BIGINT NOT NULL,
    title VARCHAR(100) NOT NULL,
    description VARCHAR(1000),
    image_url VARCHAR(2048) NOT NULL,
    metadata_uri VARCHAR(2048) NOT NULL,
    price NUMERIC(78,0), -- Big integer for Wei amounts (up to 78 digits)
    is_for_sale BOOLEAN DEFAULT FALSE,
    royalty SMALLINT DEFAULT 0,
    mint_tx_hash VARCHAR(66), -- Ethereum transaction hash format (0x + 64 hex chars)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Foreign key constraints
ALTER TABLE nfts ADD CONSTRAINT fk_nfts_creator_id FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE nfts ADD CONSTRAINT fk_nfts_owner_id FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE RESTRICT;

-- Unique constraints
CREATE UNIQUE INDEX IF NOT EXISTS idx_nfts_token_id_unique ON nfts(token_id);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_nfts_creator_id ON nfts(creator_id);
CREATE INDEX IF NOT EXISTS idx_nfts_owner_id ON nfts(owner_id);
CREATE INDEX IF NOT EXISTS idx_nfts_contract_address ON nfts(contract_address);
CREATE INDEX IF NOT EXISTS idx_nfts_for_sale ON nfts(is_for_sale) WHERE is_for_sale = true;
CREATE INDEX IF NOT EXISTS idx_nfts_created_at ON nfts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_nfts_price ON nfts(price DESC) WHERE price IS NOT NULL;

-- Check constraints
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_title_length CHECK (LENGTH(title) >= 1 AND LENGTH(title) <= 100);
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_description_length CHECK (LENGTH(description) <= 1000);
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_contract_address_format CHECK (contract_address ~ '^0x[a-fA-F0-9]{40}$');
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_image_url_format CHECK (image_url ~ '^https?://');
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_metadata_uri_format CHECK (metadata_uri ~ '^https?://');
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_price_positive CHECK (price IS NULL OR price > 0);
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_royalty_range CHECK (royalty >= 0 AND royalty <= 10);
ALTER TABLE nfts ADD CONSTRAINT chk_nfts_mint_tx_hash_format CHECK (
    mint_tx_hash IS NULL OR 
    mint_tx_hash ~ '^0x[a-fA-F0-9]{64}$'
);

-- Comments
COMMENT ON TABLE nfts IS 'Digital assets created by users and stored on blockchain';
COMMENT ON COLUMN nfts.id IS 'Primary key for the NFT';
COMMENT ON COLUMN nfts.token_id IS 'Unique blockchain token identifier';
COMMENT ON COLUMN nfts.contract_address IS 'Ethereum contract address where NFT is deployed';
COMMENT ON COLUMN nfts.creator_id IS 'Foreign key to users table (NFT creator)';
COMMENT ON COLUMN nfts.owner_id IS 'Foreign key to users table (current owner)';
COMMENT ON COLUMN nfts.title IS 'NFT title (1-100 characters)';
COMMENT ON COLUMN nfts.description IS 'NFT description (max 1000 characters)';
COMMENT ON COLUMN nfts.image_url IS 'URL to NFT image';
COMMENT ON COLUMN nfts.metadata_uri IS 'URL to NFT metadata JSON';
COMMENT ON COLUMN nfts.price IS 'Sale price in Wei (null if not for sale)';
COMMENT ON COLUMN nfts.is_for_sale IS 'Whether NFT is listed for direct sale';
COMMENT ON COLUMN nfts.royalty IS 'Creator royalty percentage (0-10%)';
COMMENT ON COLUMN nfts.mint_tx_hash IS 'Ethereum transaction hash of minting transaction';
COMMENT ON COLUMN nfts.created_at IS 'Timestamp when NFT was created';
COMMENT ON COLUMN nfts.updated_at IS 'Timestamp when NFT was last updated';

-- Update trigger for updated_at
CREATE TRIGGER update_nfts_updated_at 
    BEFORE UPDATE ON nfts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();