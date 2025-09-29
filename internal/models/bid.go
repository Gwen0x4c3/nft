package models

import (
	"math/big"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Bid represents individual bid records within auctions
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
	Auction Auction `json:"auction" gorm:"foreignKey:AuctionID"`
	Bidder  User    `json:"bidder" gorm:"foreignKey:BidderID"`
}

// BeforeCreate is a GORM hook that runs before creating a new bid
func (b *Bid) BeforeCreate(tx *gorm.DB) error {
	// Validate the bid using the validator package
	validate := validator.New()
	if err := validate.Struct(b); err != nil {
		return err
	}

	// Business validation rules
	if b.Amount == nil || b.Amount.Sign() <= 0 {
		return gorm.ErrInvalidData
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a bid
func (b *Bid) BeforeUpdate(tx *gorm.DB) error {
	// Validate the bid using the validator package
	validate := validator.New()
	return validate.Struct(b)
}

// TableName specifies the table name for the Bid model
func (Bid) TableName() string {
	return "bids"
}

// GetAmountWei returns the bid amount in Wei as a string
func (b *Bid) GetAmountWei() string {
	if b.Amount == nil {
		return "0"
	}
	return b.Amount.String()
}

// SetAmountWei sets the bid amount from a Wei string value
func (b *Bid) SetAmountWei(amountWei string) error {
	amount := new(big.Int)
	if _, ok := amount.SetString(amountWei, 10); !ok {
		return gorm.ErrInvalidData
	}
	
	// Validate that the amount is positive
	if amount.Sign() <= 0 {
		return gorm.ErrInvalidData
	}
	
	b.Amount = amount
	return nil
}

// IsConfirmed returns true if the bid has been confirmed on the blockchain
func (b *Bid) IsConfirmed() bool {
	return b.Status == "confirmed"
}

// IsPending returns true if the bid is still pending blockchain confirmation
func (b *Bid) IsPending() bool {
	return b.Status == "pending"
}

// HasFailed returns true if the bid failed to be processed on the blockchain
func (b *Bid) HasFailed() bool {
	return b.Status == "failed"
}

// ValidateAmount validates that the bid amount meets the minimum requirements
func (b *Bid) ValidateAmount(currentHighestBid *big.Int, minimumIncrement *big.Int) error {
	if b.Amount == nil {
		return gorm.ErrInvalidData
	}

	// Must be positive
	if b.Amount.Sign() <= 0 {
		return gorm.ErrInvalidData
	}

	// Must exceed current highest bid
	if currentHighestBid != nil {
		requiredAmount := new(big.Int).Add(currentHighestBid, minimumIncrement)
		if b.Amount.Cmp(requiredAmount) < 0 {
			return gorm.ErrInvalidData
		}
	}

	return nil
}