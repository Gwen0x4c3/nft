package models

import (
	"math/big"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// NFT represents a digital asset created by users, stored on blockchain
type NFT struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	TokenID      string    `json:"token_id" gorm:"uniqueIndex;not null"`
	ContractAddr string    `json:"contract_address" gorm:"not null" validate:"required,eth_addr"`
	CreatorID    uint      `json:"creator_id" gorm:"not null"`
	OwnerID      uint      `json:"owner_id" gorm:"not null"`
	Title        string    `json:"title" gorm:"not null" validate:"required,min=1,max=100"`
	Description  string    `json:"description" validate:"max=1000"`
	ImageURL     string    `json:"image_url" gorm:"not null" validate:"required,url"`
	MetadataURI  string    `json:"metadata_uri" gorm:"not null" validate:"required,url"`
	Price        *big.Int  `json:"price" gorm:"type:numeric(78,0)"` // Wei amount
	IsForSale    bool      `json:"is_for_sale" gorm:"default:false"`
	Royalty      uint8     `json:"royalty" gorm:"default:0" validate:"max=10"` // 0-10%
	MintTxHash   string    `json:"mint_tx_hash" validate:"eth_tx_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relationships
	Creator         User       `json:"creator" gorm:"foreignKey:CreatorID"`
	Owner           User       `json:"owner" gorm:"foreignKey:OwnerID"`
	Auctions        []Auction  `json:"auctions" gorm:"foreignKey:NFTID"`
	TransferHistory []Transfer `json:"transfers" gorm:"foreignKey:NFTID"`
}

// BeforeCreate is a GORM hook that runs before creating a new NFT
func (n *NFT) BeforeCreate(tx *gorm.DB) error {
	// Validate the NFT using the validator package
	validate := validator.New()
	return validate.Struct(n)
}

// BeforeUpdate is a GORM hook that runs before updating an NFT
func (n *NFT) BeforeUpdate(tx *gorm.DB) error {
	// Validate the NFT using the validator package
	validate := validator.New()
	return validate.Struct(n)
}

// TableName specifies the table name for the NFT model
func (NFT) TableName() string {
	return "nfts"
}

// GetPriceWei returns the price in Wei as a string for safe JSON serialization
func (n *NFT) GetPriceWei() string {
	if n.Price == nil {
		return "0"
	}
	return n.Price.String()
}

// SetPriceWei sets the price from a Wei string value
func (n *NFT) SetPriceWei(priceWei string) error {
	price := new(big.Int)
	if _, ok := price.SetString(priceWei, 10); !ok {
		return gorm.ErrInvalidData
	}
	n.Price = price
	return nil
}