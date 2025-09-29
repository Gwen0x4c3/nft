# Data Model: NFT Platform

**Date**: 2025-09-29  
**Context**: Entity design for NFT minting and auction platform

## Core Entities

### User

**Purpose**: Platform users who can mint, buy, and sell NFTs

```go
type User struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Username    string    `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=30"`
    Email       string    `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
    WalletAddr  string    `json:"wallet_address" gorm:"uniqueIndex;not null" validate:"required,eth_addr"`
    Avatar      string    `json:"avatar" validate:"url"`
    Bio         string    `json:"bio" validate:"max=500"`
    IsVerified  bool      `json:"is_verified" gorm:"default:false"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relationships
    NFTs        []NFT     `json:"nfts" gorm:"foreignKey:CreatorID"`
    Bids        []Bid     `json:"bids" gorm:"foreignKey:BidderID"`
}
```

**Validation Rules**:

- Username: 3-30 characters, alphanumeric + underscore
- Email: Valid email format, unique
- WalletAddr: Valid Ethereum address format, unique
- Avatar: Valid URL or empty
- Bio: Max 500 characters

**State Transitions**: None (user state is static)

### NFT

**Purpose**: Digital assets created by users, stored on blockchain

```go
type NFT struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    TokenID         string    `json:"token_id" gorm:"uniqueIndex;not null"`
    ContractAddr    string    `json:"contract_address" gorm:"not null" validate:"required,eth_addr"`
    CreatorID       uint      `json:"creator_id" gorm:"not null"`
    OwnerID         uint      `json:"owner_id" gorm:"not null"`
    Title           string    `json:"title" gorm:"not null" validate:"required,min=1,max=100"`
    Description     string    `json:"description" validate:"max=1000"`
    ImageURL        string    `json:"image_url" gorm:"not null" validate:"required,url"`
    MetadataURI     string    `json:"metadata_uri" gorm:"not null" validate:"required,url"`
    Price           *big.Int  `json:"price" gorm:"type:numeric(78,0)"` // Wei amount
    IsForSale       bool      `json:"is_for_sale" gorm:"default:false"`
    Royalty         uint8     `json:"royalty" gorm:"default:0" validate:"max=10"` // 0-10%
    MintTxHash      string    `json:"mint_tx_hash" validate:"eth_tx_hash"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`

    // Relationships
    Creator         User      `json:"creator" gorm:"foreignKey:CreatorID"`
    Owner           User      `json:"owner" gorm:"foreignKey:OwnerID"`
    Auctions        []Auction `json:"auctions" gorm:"foreignKey:NFTID"`
    TransferHistory []Transfer `json:"transfers" gorm:"foreignKey:NFTID"`
}
```

**Validation Rules**:

- TokenID: Unique blockchain token identifier
- ContractAddr: Valid Ethereum contract address
- Title: 1-100 characters, required
- Description: Max 1000 characters
- ImageURL: Valid URL, required
- MetadataURI: Valid URL pointing to NFT metadata JSON
- Price: Positive value in Wei, nullable
- Royalty: 0-10% percentage
- MintTxHash: Valid Ethereum transaction hash

**State Transitions**:

- `Minting` → `Owned` (after successful blockchain mint)
- `Owned` → `ForSale` (when listed for direct sale)
- `Owned` → `InAuction` (when auction created)
- `ForSale` → `Owned` (when unlisted)
- `InAuction` → `Owned` (when auction ends)

### Auction

**Purpose**: Time-bound competitive bidding for NFTs

```go
type Auction struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    NFTID       uint      `json:"nft_id" gorm:"not null"`
    SellerID    uint      `json:"seller_id" gorm:"not null"`
    StartPrice  *big.Int  `json:"start_price" gorm:"type:numeric(78,0);not null" validate:"required"`
    ReservePrice *big.Int `json:"reserve_price" gorm:"type:numeric(78,0)"` // Minimum acceptable price
    CurrentBid  *big.Int  `json:"current_bid" gorm:"type:numeric(78,0);default:0"`
    StartTime   time.Time `json:"start_time" gorm:"not null"`
    EndTime     time.Time `json:"end_time" gorm:"not null"`
    Status      string    `json:"status" gorm:"default:pending" validate:"oneof=pending active ended cancelled"`
    WinnerID    *uint     `json:"winner_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`

    // Relationships
    NFT         NFT       `json:"nft" gorm:"foreignKey:NFTID"`
    Seller      User      `json:"seller" gorm:"foreignKey:SellerID"`
    Winner      *User     `json:"winner" gorm:"foreignKey:WinnerID"`
    Bids        []Bid     `json:"bids" gorm:"foreignKey:AuctionID"`
}
```

**Validation Rules**:

- StartPrice: Must be positive, required
- ReservePrice: Must be >= StartPrice if set
- StartTime: Must be in the future when created
- EndTime: Must be after StartTime, minimum 1 hour duration
- Status: One of: pending, active, ended, cancelled

**State Transitions**:

- `pending` → `active` (when StartTime reached)
- `active` → `ended` (when EndTime reached)
- `pending/active` → `cancelled` (by seller before first bid)

### Bid

**Purpose**: Individual bid records within auctions

```go
type Bid struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    AuctionID uint      `json:"auction_id" gorm:"not null"`
    BidderID  uint      `json:"bidder_id" gorm:"not null"`
    Amount    *big.Int  `json:"amount" gorm:"type:numeric(78,0);not null" validate:"required"`
    TxHash    string    `json:"tx_hash" validate:"eth_tx_hash"`
    Status    string    `json:"status" gorm:"default:pending" validate:"oneof=pending confirmed failed"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`

    // Relationships
    Auction   Auction   `json:"auction" gorm:"foreignKey:AuctionID"`
    Bidder    User      `json:"bidder" gorm:"foreignKey:BidderID"`
}
```

**Validation Rules**:

- Amount: Must be greater than current highest bid
- TxHash: Valid Ethereum transaction hash when confirmed
- Status: One of: pending, confirmed, failed

**State Transitions**:

- `pending` → `confirmed` (after blockchain confirmation)
- `pending` → `failed` (if blockchain transaction fails)

### Transfer

**Purpose**: NFT ownership transfer history

```go
type Transfer struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    NFTID     uint      `json:"nft_id" gorm:"not null"`
    FromID    *uint     `json:"from_id"` // Null for mint transactions
    ToID      uint      `json:"to_id" gorm:"not null"`
    TxHash    string    `json:"tx_hash" gorm:"not null" validate:"required,eth_tx_hash"`
    Price     *big.Int  `json:"price" gorm:"type:numeric(78,0)"` // Null for non-sale transfers
    Type      string    `json:"type" gorm:"not null" validate:"oneof=mint sale transfer"`
    CreatedAt time.Time `json:"created_at"`

    // Relationships
    NFT       NFT       `json:"nft" gorm:"foreignKey:NFTID"`
    From      *User     `json:"from" gorm:"foreignKey:FromID"`
    To        User      `json:"to" gorm:"foreignKey:ToID"`
}
```

**Validation Rules**:

- TxHash: Valid Ethereum transaction hash, required
- Price: Must be positive if Type is "sale"
- Type: One of: mint, sale, transfer
- FromID: Must be null only for mint type

### Notification

**Purpose**: Real-time user notifications for platform events

```go
type Notification struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    UserID    uint      `json:"user_id" gorm:"not null"`
    Type      string    `json:"type" gorm:"not null" validate:"oneof=bid_placed bid_outbid auction_won auction_ended nft_sold"`
    Title     string    `json:"title" gorm:"not null" validate:"required,max=100"`
    Message   string    `json:"message" gorm:"not null" validate:"required,max=500"`
    Data      JSON      `json:"data" gorm:"type:jsonb"` // Additional context data
    IsRead    bool      `json:"is_read" gorm:"default:false"`
    CreatedAt time.Time `json:"created_at"`

    // Relationships
    User      User      `json:"user" gorm:"foreignKey:UserID"`
}
```

**Validation Rules**:

- Type: One of predefined notification types
- Title: Max 100 characters, required
- Message: Max 500 characters, required
- Data: Valid JSON object with context information

## Relationships Summary

### One-to-Many

- User → NFTs (as creator)
- User → NFTs (as owner)
- User → Bids
- User → Notifications
- NFT → Auctions
- NFT → Transfers
- Auction → Bids

### Many-to-One

- NFT → User (creator)
- NFT → User (owner)
- Auction → NFT
- Auction → User (seller)
- Bid → Auction
- Bid → User (bidder)
- Transfer → NFT
- Notification → User

### Optional Relationships

- Auction → User (winner, nullable)
- Transfer → User (from, nullable for mints)

## Database Indexes

### Performance Indexes

```sql
-- User lookups
CREATE INDEX idx_users_wallet_addr ON users(wallet_address);
CREATE INDEX idx_users_email ON users(email);

-- NFT queries
CREATE INDEX idx_nfts_creator_id ON nfts(creator_id);
CREATE INDEX idx_nfts_owner_id ON nfts(owner_id);
CREATE INDEX idx_nfts_for_sale ON nfts(is_for_sale) WHERE is_for_sale = true;
CREATE INDEX idx_nfts_created_at ON nfts(created_at DESC);

-- Auction queries
CREATE INDEX idx_auctions_status ON auctions(status);
CREATE INDEX idx_auctions_end_time ON auctions(end_time);
CREATE INDEX idx_auctions_nft_id ON auctions(nft_id);

-- Bid queries
CREATE INDEX idx_bids_auction_id ON bids(auction_id);
CREATE INDEX idx_bids_bidder_id ON bids(bidder_id);
CREATE INDEX idx_bids_amount ON bids(amount DESC);

-- Transfer history
CREATE INDEX idx_transfers_nft_id ON transfers(nft_id);
CREATE INDEX idx_transfers_from_id ON transfers(from_id);
CREATE INDEX idx_transfers_to_id ON transfers(to_id);

-- Notifications
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;
```

## Data Consistency Rules

### Business Rules

1. **NFT Ownership**: An NFT can only have one owner at a time
2. **Auction Integrity**: Only NFT owner can create auction for their NFT
3. **Bid Validation**: Bids must exceed current highest bid by minimum increment
4. **Transfer Tracking**: Every NFT ownership change must create Transfer record
5. **Royalty Enforcement**: Creator royalties automatically calculated on sales

### Referential Integrity

1. All foreign keys must reference existing records
2. NFT creator and owner must be valid users
3. Auction seller must be NFT owner
4. Bid bidder cannot be auction seller
5. Transfer records must maintain ownership chain

### Concurrency Handling

1. **Optimistic Locking**: Use version fields for auction updates
2. **Atomic Bids**: Bid placement and auction update in single transaction
3. **Queue Processing**: Blockchain events processed sequentially
4. **Cache Invalidation**: Redis cache updated after database changes

---

**Model Status**: ✅ Complete - All entities defined with validation and relationships
