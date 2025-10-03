package messaging

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Auction events
	EventTypeBidPlaced    EventType = "bid_placed"
	EventTypeAuctionEnded EventType = "auction_ended"
	EventTypeAuctionExtended EventType = "auction_extended"

	// NFT events
	EventTypeNFTMinted      EventType = "nft_minted"
	EventTypeNFTTransferred EventType = "nft_transferred"
	EventTypeNFTListed      EventType = "nft_listed"
	EventTypeNFTUnlisted    EventType = "nft_unlisted"

	// User events
	EventTypeUserRegistered   EventType = "user_registered"
	EventTypeUserUpdated      EventType = "user_updated"
	EventTypeUserStatusChanged EventType = "user_status_changed"

	// Notification events
	EventTypeNotification EventType = "notification"

	// Blockchain events
	EventTypeBlockchainEvent EventType = "blockchain_event"
	EventTypeTransactionConfirmed EventType = "transaction_confirmed"
	EventTypeTransactionFailed EventType = "transaction_failed"

	// Collection events
	EventTypeCollectionFloorPriceChanged EventType = "collection_floor_price_changed"
)

// Event represents a domain event
type Event struct {
	ID            string      `json:"id"`
	Type          EventType   `json:"type"`
	AggregateType string      `json:"aggregate_type"`
	AggregateID   string      `json:"aggregate_id"`
	Version       string      `json:"version"`
	Timestamp     time.Time   `json:"timestamp"`
	UserID        *uint       `json:"user_id,omitempty"`
	Data          interface{} `json:"data"`
	CorrelationID string      `json:"correlation_id"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Validate validates the event structure
func (e *Event) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("event ID is required")
	}
	if e.Type == "" {
		return fmt.Errorf("event type is required")
	}
	if e.AggregateID == "" {
		return fmt.Errorf("aggregate ID is required")
	}
	if e.Version == "" {
		return fmt.Errorf("event version is required")
	}
	if e.Timestamp.IsZero() {
		return fmt.Errorf("event timestamp is required")
	}
	if e.CorrelationID == "" {
		return fmt.Errorf("correlation ID is required")
	}
	if e.Data == nil {
		return fmt.Errorf("event data is required")
	}
	return nil
}

// BidPlacedData represents the data for a bid placed event
type BidPlacedData struct {
	AuctionID        uint    `json:"auction_id"`
	BidID            uint    `json:"bid_id"`
	BidderID         uint    `json:"bidder_id"`
	BidderUsername   string  `json:"bidder_username"`
	BidderAvatar     *string `json:"bidder_avatar,omitempty"`
	Amount           string  `json:"amount"`
	AmountFormatted  string  `json:"amount_formatted"`
	PreviousBidID    *uint   `json:"previous_bid_id,omitempty"`
	PreviousAmount   *string `json:"previous_amount,omitempty"`
	TransactionHash  *string `json:"transaction_hash,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// AuctionEndedData represents the data for an auction ended event
type AuctionEndedData struct {
	AuctionID         uint      `json:"auction_id"`
	WinnerID          *uint     `json:"winner_id,omitempty"`
	WinnerUsername    *string   `json:"winner_username,omitempty"`
	WinnerAvatar      *string   `json:"winner_avatar,omitempty"`
	WinningBidID      *uint     `json:"winning_bid_id,omitempty"`
	WinningBidAmount  *string   `json:"winning_bid_amount,omitempty"`
	WinningBidFormatted *string `json:"winning_bid_formatted,omitempty"`
	ReservePriceMet   bool      `json:"reserve_price_met"`
	TotalBids         int       `json:"total_bids"`
	EndTime           time.Time `json:"end_time"`
	SettlementTxHash  *string   `json:"settlement_tx_hash,omitempty"`
}

// NFTMintedData represents the data for an NFT minted event
type NFTMintedData struct {
	NFTID           uint      `json:"nft_id"`
	TokenID         string    `json:"token_id"`
	ContractAddr    string    `json:"contract_address"`
	CreatorID       uint      `json:"creator_id"`
	CreatorUsername string    `json:"creator_username"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	ImageURL        string    `json:"image_url"`
	MetadataURI     string    `json:"metadata_uri"`
	CollectionID    *uint     `json:"collection_id,omitempty"`
	CollectionName  *string   `json:"collection_name,omitempty"`
	MintTxHash      string    `json:"mint_tx_hash"`
	BlockNumber     uint64    `json:"block_number"`
	CreatedAt       time.Time `json:"created_at"`
}

// NFTTransferredData represents the data for an NFT transferred event
type NFTTransferredData struct {
	NFTID           uint      `json:"nft_id"`
	TokenID         string    `json:"token_id"`
	ContractAddr    string    `json:"contract_address"`
	FromID          *uint     `json:"from_id,omitempty"`
	FromUsername    *string   `json:"from_username,omitempty"`
	ToID            uint      `json:"to_id"`
	ToUsername      string    `json:"to_username"`
	TransferType    string    `json:"transfer_type"` // "mint", "sale", "transfer", "burn"
	Price           *string   `json:"price,omitempty"`
	PriceFormatted  *string   `json:"price_formatted,omitempty"`
	TransactionHash string    `json:"transaction_hash"`
	BlockNumber     uint64    `json:"block_number"`
	CreatedAt       time.Time `json:"created_at"`
}

// UserRegisteredData represents the data for a user registered event
type UserRegisteredData struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	WalletAddr   string    `json:"wallet_address"`
	Avatar       *string   `json:"avatar,omitempty"`
	Bio          *string   `json:"bio,omitempty"`
	IsVerified   bool      `json:"is_verified"`
	ReferralCode *string   `json:"referral_code,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// NotificationData represents the data for a notification event
type NotificationData struct {
	NotificationID uint   `json:"notification_id"`
	UserID         uint   `json:"user_id"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	ActionURL      *string `json:"action_url,omitempty"`
	Priority       string `json:"priority"`
	RelatedType    *string `json:"related_type,omitempty"`
	RelatedID      *uint   `json:"related_id,omitempty"`
	Channels       []string `json:"channels"` // ["websocket", "email", "push"]
}

// BlockchainEventData represents the data for a blockchain event
type BlockchainEventData struct {
	EventName       string                 `json:"event_name"`
	ContractAddr    string                 `json:"contract_address"`
	TransactionHash string                 `json:"transaction_hash"`
	BlockNumber     uint64                 `json:"block_number"`
	LogIndex        uint                   `json:"log_index"`
	Parameters      map[string]interface{} `json:"parameters"`
	Timestamp       time.Time              `json:"timestamp"`
	Processed       bool                   `json:"processed"`
}

// CollectionFloorPriceChangedData represents the data for a collection floor price change event
type CollectionFloorPriceChangedData struct {
	CollectionID            uint    `json:"collection_id"`
	CollectionName          string  `json:"collection_name"`
	OldFloorPrice           string  `json:"old_floor_price"`
	OldFloorPriceFormatted  string  `json:"old_floor_price_formatted"`
	NewFloorPrice           string  `json:"new_floor_price"`
	NewFloorPriceFormatted  string  `json:"new_floor_price_formatted"`
	ChangePercentage        float64 `json:"change_percentage"`
	Volume24h               string  `json:"volume_24h"`
	Volume24hFormatted      string  `json:"volume_24h_formatted"`
	NFTsAffectingFloor      []uint  `json:"nfts_affecting_floor,omitempty"`
}

// generateEventID generates a unique event ID
func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("evt_%s", hex.EncodeToString(bytes))
}

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("corr_%s", hex.EncodeToString(bytes))
}