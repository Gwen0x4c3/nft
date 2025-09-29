-- Migration: 001_create_users_table
-- Description: Create the users table for platform users who can mint, buy, and sell NFTs
-- Created: 2025-09-29

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(30) NOT NULL,
    email VARCHAR(255) NOT NULL,
    wallet_address VARCHAR(42) NOT NULL,
    avatar VARCHAR(2048),
    bio VARCHAR(500),
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Unique constraints
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_unique ON users(username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_wallet_address_unique ON users(wallet_address);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_users_wallet_addr ON users(wallet_address);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_is_verified ON users(is_verified) WHERE is_verified = true;

-- Check constraints
ALTER TABLE users ADD CONSTRAINT chk_users_username_length CHECK (LENGTH(username) >= 3 AND LENGTH(username) <= 30);
ALTER TABLE users ADD CONSTRAINT chk_users_email_format CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');
ALTER TABLE users ADD CONSTRAINT chk_users_wallet_address_format CHECK (wallet_address ~ '^0x[a-fA-F0-9]{40}$');
ALTER TABLE users ADD CONSTRAINT chk_users_bio_length CHECK (LENGTH(bio) <= 500);

-- Comments
COMMENT ON TABLE users IS 'Platform users who can mint, buy, and sell NFTs';
COMMENT ON COLUMN users.id IS 'Primary key for the user';
COMMENT ON COLUMN users.username IS 'Unique username (3-30 characters)';
COMMENT ON COLUMN users.email IS 'Unique email address';
COMMENT ON COLUMN users.wallet_address IS 'Unique Ethereum wallet address (0x format)';
COMMENT ON COLUMN users.avatar IS 'URL to user avatar image';
COMMENT ON COLUMN users.bio IS 'User biography (max 500 characters)';
COMMENT ON COLUMN users.is_verified IS 'Whether the user is verified';
COMMENT ON COLUMN users.created_at IS 'Timestamp when user was created';
COMMENT ON COLUMN users.updated_at IS 'Timestamp when user was last updated';

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();