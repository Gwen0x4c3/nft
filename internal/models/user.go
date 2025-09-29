package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// User represents a platform user who can mint, buy, and sell NFTs
type User struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Username   string    `json:"username" gorm:"uniqueIndex;not null" validate:"required,min=3,max=30"`
	Email      string    `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	WalletAddr string    `json:"wallet_address" gorm:"uniqueIndex;not null" validate:"required,eth_addr"`
	Avatar     string    `json:"avatar" validate:"url"`
	Bio        string    `json:"bio" validate:"max=500"`
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	NFTs []NFT `json:"nfts" gorm:"foreignKey:CreatorID"`
	Bids []Bid `json:"bids" gorm:"foreignKey:BidderID"`
}

// BeforeCreate is a GORM hook that runs before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Validate the user using the validator package
	validate := validator.New()
	return validate.Struct(u)
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Validate the user using the validator package
	validate := validator.New()
	return validate.Struct(u)
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}