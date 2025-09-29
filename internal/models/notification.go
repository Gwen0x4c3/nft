package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// JSON is a custom type for handling JSON data in GORM
type JSON map[string]interface{}

// Value implements the driver.Valuer interface for JSON
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSON
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSON)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSON", value)
	}

	return json.Unmarshal(bytes, j)
}

// Notification represents real-time user notifications for platform events
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
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// BeforeCreate is a GORM hook that runs before creating a new notification
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	// Validate the notification using the validator package
	validate := validator.New()
	if err := validate.Struct(n); err != nil {
		return err
	}

	// Initialize data field if nil
	if n.Data == nil {
		n.Data = make(JSON)
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a notification
func (n *Notification) BeforeUpdate(tx *gorm.DB) error {
	// Validate the notification using the validator package
	validate := validator.New()
	return validate.Struct(n)
}

// TableName specifies the table name for the Notification model
func (Notification) TableName() string {
	return "notifications"
}

// SetData sets the data field from a map
func (n *Notification) SetData(data map[string]interface{}) {
	if data == nil {
		n.Data = make(JSON)
	} else {
		n.Data = JSON(data)
	}
}

// GetData returns the data field as a map
func (n *Notification) GetData() map[string]interface{} {
	return map[string]interface{}(n.Data)
}

// AddDataField adds a single field to the data
func (n *Notification) AddDataField(key string, value interface{}) {
	if n.Data == nil {
		n.Data = make(JSON)
	}
	n.Data[key] = value
}

// GetDataField retrieves a single field from the data
func (n *Notification) GetDataField(key string) (interface{}, bool) {
	if n.Data == nil {
		return nil, false
	}
	value, exists := n.Data[key]
	return value, exists
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	n.IsRead = true
}

// MarkAsUnread marks the notification as unread
func (n *Notification) MarkAsUnread() {
	n.IsRead = false
}

// IsUnread returns true if the notification is unread
func (n *Notification) IsUnread() bool {
	return !n.IsRead
}

// IsBidNotification returns true if this is a bid-related notification
func (n *Notification) IsBidNotification() bool {
	return n.Type == "bid_placed" || n.Type == "bid_outbid"
}

// IsAuctionNotification returns true if this is an auction-related notification
func (n *Notification) IsAuctionNotification() bool {
	return n.Type == "auction_won" || n.Type == "auction_ended"
}

// IsSaleNotification returns true if this is a sale-related notification
func (n *Notification) IsSaleNotification() bool {
	return n.Type == "nft_sold"
}

// NotificationData represents structured data for different notification types
type NotificationData struct {
	NFTID     *uint   `json:"nft_id,omitempty"`
	AuctionID *uint   `json:"auction_id,omitempty"`
	BidID     *uint   `json:"bid_id,omitempty"`
	Amount    *string `json:"amount,omitempty"`    // Wei amount as string
	BidderID  *uint   `json:"bidder_id,omitempty"`
	SellerID  *uint   `json:"seller_id,omitempty"`
	BuyerID   *uint   `json:"buyer_id,omitempty"`
	TxHash    *string `json:"tx_hash,omitempty"`
}

// SetStructuredData sets the notification data from a structured object
func (n *Notification) SetStructuredData(data NotificationData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var mapData map[string]interface{}
	if err := json.Unmarshal(jsonData, &mapData); err != nil {
		return err
	}

	n.Data = JSON(mapData)
	return nil
}

// GetStructuredData retrieves the notification data as a structured object
func (n *Notification) GetStructuredData() (*NotificationData, error) {
	if n.Data == nil {
		return &NotificationData{}, nil
	}

	jsonData, err := json.Marshal(n.Data)
	if err != nil {
		return nil, err
	}

	var data NotificationData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}