package models

import (
	"math/big"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Transfer represents NFT ownership transfer history
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
	NFT  NFT   `json:"nft" gorm:"foreignKey:NFTID"`
	From *User `json:"from" gorm:"foreignKey:FromID"`
	To   User  `json:"to" gorm:"foreignKey:ToID"`
}

// BeforeCreate is a GORM hook that runs before creating a new transfer
func (t *Transfer) BeforeCreate(tx *gorm.DB) error {
	// Validate the transfer using the validator package
	validate := validator.New()
	if err := validate.Struct(t); err != nil {
		return err
	}

	// Business validation rules
	switch t.Type {
	case "mint":
		// Mint transfers must have null FromID
		if t.FromID != nil {
			return gorm.ErrInvalidData
		}
	case "sale":
		// Sale transfers must have a price and FromID
		if t.Price == nil || t.Price.Sign() <= 0 {
			return gorm.ErrInvalidData
		}
		if t.FromID == nil {
			return gorm.ErrInvalidData
		}
	case "transfer":
		// Regular transfers must have FromID but price can be null
		if t.FromID == nil {
			return gorm.ErrInvalidData
		}
	default:
		return gorm.ErrInvalidData
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a transfer
func (t *Transfer) BeforeUpdate(tx *gorm.DB) error {
	// Validate the transfer using the validator package
	validate := validator.New()
	return validate.Struct(t)
}

// TableName specifies the table name for the Transfer model
func (Transfer) TableName() string {
	return "transfers"
}

// GetPriceWei returns the price in Wei as a string
func (t *Transfer) GetPriceWei() string {
	if t.Price == nil {
		return "0"
	}
	return t.Price.String()
}

// SetPriceWei sets the price from a Wei string value
func (t *Transfer) SetPriceWei(priceWei string) error {
	if priceWei == "" || priceWei == "0" {
		t.Price = nil
		return nil
	}

	price := new(big.Int)
	if _, ok := price.SetString(priceWei, 10); !ok {
		return gorm.ErrInvalidData
	}

	// Validate that the price is positive for sales
	if t.Type == "sale" && price.Sign() <= 0 {
		return gorm.ErrInvalidData
	}

	t.Price = price
	return nil
}

// IsMint returns true if this is a mint transfer
func (t *Transfer) IsMint() bool {
	return t.Type == "mint"
}

// IsSale returns true if this is a sale transfer
func (t *Transfer) IsSale() bool {
	return t.Type == "sale"
}

// IsTransfer returns true if this is a regular transfer (no sale)
func (t *Transfer) IsTransfer() bool {
	return t.Type == "transfer"
}

// HasPrice returns true if this transfer has a price associated
func (t *Transfer) HasPrice() bool {
	return t.Price != nil && t.Price.Sign() > 0
}

// GetFromUserID returns the FromID as a value, or 0 if nil (for mint transfers)
func (t *Transfer) GetFromUserID() uint {
	if t.FromID == nil {
		return 0
	}
	return *t.FromID
}

// SetFromUserID sets the FromID, handling nil for mint transfers
func (t *Transfer) SetFromUserID(fromID uint) {
	if fromID == 0 {
		t.FromID = nil
	} else {
		t.FromID = &fromID
	}
}