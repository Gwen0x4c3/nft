package models

import (
	"math/big"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Auction represents time-bound competitive bidding for NFTs
type Auction struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	NFTID        uint      `json:"nft_id" gorm:"not null"`
	SellerID     uint      `json:"seller_id" gorm:"not null"`
	StartPrice   *big.Int  `json:"start_price" gorm:"type:numeric(78,0);not null" validate:"required"`
	ReservePrice *big.Int  `json:"reserve_price" gorm:"type:numeric(78,0)"` // Minimum acceptable price
	CurrentBid   *big.Int  `json:"current_bid" gorm:"type:numeric(78,0);default:0"`
	StartTime    time.Time `json:"start_time" gorm:"not null"`
	EndTime      time.Time `json:"end_time" gorm:"not null"`
	Status       string    `json:"status" gorm:"default:pending" validate:"oneof=pending active ended cancelled"`
	WinnerID     *uint     `json:"winner_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	NFT    NFT   `json:"nft" gorm:"foreignKey:NFTID"`
	Seller User  `json:"seller" gorm:"foreignKey:SellerID"`
	Winner *User `json:"winner" gorm:"foreignKey:WinnerID"`
	Bids   []Bid `json:"bids" gorm:"foreignKey:AuctionID"`
}

// BeforeCreate is a GORM hook that runs before creating a new auction
func (a *Auction) BeforeCreate(tx *gorm.DB) error {
	// Validate the auction using the validator package
	validate := validator.New()
	if err := validate.Struct(a); err != nil {
		return err
	}

	// Business validation rules
	if a.EndTime.Before(a.StartTime) {
		return gorm.ErrInvalidData
	}

	// Minimum auction duration of 1 hour
	if a.EndTime.Sub(a.StartTime) < time.Hour {
		return gorm.ErrInvalidData
	}

	// Reserve price must be >= start price if set
	if a.ReservePrice != nil && a.StartPrice != nil {
		if a.ReservePrice.Cmp(a.StartPrice) < 0 {
			return gorm.ErrInvalidData
		}
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating an auction
func (a *Auction) BeforeUpdate(tx *gorm.DB) error {
	// Validate the auction using the validator package
	validate := validator.New()
	return validate.Struct(a)
}

// TableName specifies the table name for the Auction model
func (Auction) TableName() string {
	return "auctions"
}

// GetStartPriceWei returns the start price in Wei as a string
func (a *Auction) GetStartPriceWei() string {
	if a.StartPrice == nil {
		return "0"
	}
	return a.StartPrice.String()
}

// SetStartPriceWei sets the start price from a Wei string value
func (a *Auction) SetStartPriceWei(priceWei string) error {
	price := new(big.Int)
	if _, ok := price.SetString(priceWei, 10); !ok {
		return gorm.ErrInvalidData
	}
	a.StartPrice = price
	return nil
}

// GetReservePriceWei returns the reserve price in Wei as a string
func (a *Auction) GetReservePriceWei() string {
	if a.ReservePrice == nil {
		return "0"
	}
	return a.ReservePrice.String()
}

// SetReservePriceWei sets the reserve price from a Wei string value
func (a *Auction) SetReservePriceWei(priceWei string) error {
	price := new(big.Int)
	if _, ok := price.SetString(priceWei, 10); !ok {
		return gorm.ErrInvalidData
	}
	a.ReservePrice = price
	return nil
}

// GetCurrentBidWei returns the current bid in Wei as a string
func (a *Auction) GetCurrentBidWei() string {
	if a.CurrentBid == nil {
		return "0"
	}
	return a.CurrentBid.String()
}

// SetCurrentBidWei sets the current bid from a Wei string value
func (a *Auction) SetCurrentBidWei(bidWei string) error {
	bid := new(big.Int)
	if _, ok := bid.SetString(bidWei, 10); !ok {
		return gorm.ErrInvalidData
	}
	a.CurrentBid = bid
	return nil
}

// IsActive returns true if the auction is currently active
func (a *Auction) IsActive() bool {
	now := time.Now()
	return a.Status == "active" && now.After(a.StartTime) && now.Before(a.EndTime)
}

// HasReservePrice returns true if the auction has a reserve price set
func (a *Auction) HasReservePrice() bool {
	return a.ReservePrice != nil && a.ReservePrice.Sign() > 0
}

// IsReserveMet returns true if the current bid meets or exceeds the reserve price
func (a *Auction) IsReserveMet() bool {
	if !a.HasReservePrice() {
		return true // No reserve price means it's always met
	}
	if a.CurrentBid == nil {
		return false
	}
	return a.CurrentBid.Cmp(a.ReservePrice) >= 0
}