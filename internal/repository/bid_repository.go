package repository

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"nft-platform/internal/models"

	"gorm.io/gorm"
)

// BidFilter represents filtering options for bid queries
type BidFilter struct {
	AuctionID     *uint
	BidderID      *uint
	Status        *string
	MinAmount     *big.Int
	MaxAmount     *big.Int
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// BidRepository handles database operations for Bid entities
type BidRepository struct {
	db *gorm.DB
}

// NewBidRepository creates a new BidRepository instance
func NewBidRepository(db *gorm.DB) *BidRepository {
	return &BidRepository{db: db}
}

// Create creates a new bid in the database
func (r *BidRepository) Create(bid *models.Bid) error {
	if err := r.db.Create(bid).Error; err != nil {
		return fmt.Errorf("failed to create bid: %w", err)
	}
	return nil
}

// GetByID retrieves a bid by its ID with related data
func (r *BidRepository) GetByID(id uint) (*models.Bid, error) {
	var bid models.Bid
	if err := r.db.Preload("Auction").Preload("Bidder").First(&bid, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bid with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get bid by id: %w", err)
	}
	return &bid, nil
}

// Update updates an existing bid
func (r *BidRepository) Update(bid *models.Bid) error {
	if err := r.db.Save(bid).Error; err != nil {
		return fmt.Errorf("failed to update bid: %w", err)
	}
	return nil
}

// Delete deletes a bid from the database
func (r *BidRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Bid{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete bid: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("bid with id %d not found", id)
	}
	return nil
}

// List retrieves bids with filtering and pagination
func (r *BidRepository) List(filter *BidFilter, offset, limit int) ([]models.Bid, error) {
	var bids []models.Bid
	query := r.db.Model(&models.Bid{}).
		Preload("Auction").
		Preload("Bidder")

	// Apply filters
	if filter != nil {
		if filter.AuctionID != nil {
			query = query.Where("auction_id = ?", *filter.AuctionID)
		}
		if filter.BidderID != nil {
			query = query.Where("bidder_id = ?", *filter.BidderID)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.MinAmount != nil {
			query = query.Where("amount >= ?", filter.MinAmount.String())
		}
		if filter.MaxAmount != nil {
			query = query.Where("amount <= ?", filter.MaxAmount.String())
		}
		if filter.CreatedAfter != nil {
			query = query.Where("created_at > ?", *filter.CreatedAfter)
		}
		if filter.CreatedBefore != nil {
			query = query.Where("created_at < ?", *filter.CreatedBefore)
		}
	}

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("amount DESC, created_at ASC").Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to list bids: %w", err)
	}

	return bids, nil
}

// Count returns the total number of bids matching the filter
func (r *BidRepository) Count(filter *BidFilter) (int64, error) {
	var count int64
	query := r.db.Model(&models.Bid{})

	// Apply same filters as List
	if filter != nil {
		if filter.AuctionID != nil {
			query = query.Where("auction_id = ?", *filter.AuctionID)
		}
		if filter.BidderID != nil {
			query = query.Where("bidder_id = ?", *filter.BidderID)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.MinAmount != nil {
			query = query.Where("amount >= ?", filter.MinAmount.String())
		}
		if filter.MaxAmount != nil {
			query = query.Where("amount <= ?", filter.MaxAmount.String())
		}
		if filter.CreatedAfter != nil {
			query = query.Where("created_at > ?", *filter.CreatedAfter)
		}
		if filter.CreatedBefore != nil {
			query = query.Where("created_at < ?", *filter.CreatedBefore)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count bids: %w", err)
	}

	return count, nil
}

// GetByAuctionID retrieves bids for a specific auction
func (r *BidRepository) GetByAuctionID(auctionID uint, offset, limit int) ([]models.Bid, error) {
	var bids []models.Bid
	query := r.db.Where("auction_id = ?", auctionID).
		Preload("Auction").
		Preload("Bidder")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("amount DESC, created_at ASC").Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get bids by auction: %w", err)
	}

	return bids, nil
}

// GetByBidderID retrieves bids placed by a specific bidder
func (r *BidRepository) GetByBidderID(bidderID uint, offset, limit int) ([]models.Bid, error) {
	var bids []models.Bid
	query := r.db.Where("bidder_id = ?", bidderID).
		Preload("Auction").
		Preload("Bidder")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get bids by bidder: %w", err)
	}

	return bids, nil
}

// GetHighestBidForAuction retrieves the highest confirmed bid for an auction
func (r *BidRepository) GetHighestBidForAuction(auctionID uint) (*models.Bid, error) {
	var bid models.Bid
	if err := r.db.Where("auction_id = ? AND status = ?", auctionID, "confirmed").
		Preload("Auction").
		Preload("Bidder").
		Order("amount DESC, created_at ASC").
		First(&bid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No bids found
		}
		return nil, fmt.Errorf("failed to get highest bid: %w", err)
	}
	return &bid, nil
}

// GetConfirmedBidsByAuction retrieves all confirmed bids for an auction
func (r *BidRepository) GetConfirmedBidsByAuction(auctionID uint, offset, limit int) ([]models.Bid, error) {
	var bids []models.Bid
	query := r.db.Where("auction_id = ? AND status = ?", auctionID, "confirmed").
		Preload("Auction").
		Preload("Bidder")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("amount DESC, created_at ASC").Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get confirmed bids: %w", err)
	}

	return bids, nil
}

// GetPendingBids retrieves all pending bids that need blockchain confirmation
func (r *BidRepository) GetPendingBids(limit int) ([]models.Bid, error) {
	var bids []models.Bid
	query := r.db.Where("status = ?", "pending").
		Preload("Auction").
		Preload("Bidder")

	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(50) // Default limit
	}

	if err := query.Order("created_at ASC").Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending bids: %w", err)
	}

	return bids, nil
}

// UpdateStatus updates the status of a bid
func (r *BidRepository) UpdateStatus(id uint, status string) error {
	result := r.db.Model(&models.Bid{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update bid status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("bid with id %d not found", id)
	}
	return nil
}

// UpdateTxHash updates the transaction hash of a bid
func (r *BidRepository) UpdateTxHash(id uint, txHash string) error {
	result := r.db.Model(&models.Bid{}).Where("id = ?", id).Update("tx_hash", txHash)
	if result.Error != nil {
		return fmt.Errorf("failed to update bid tx hash: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("bid with id %d not found", id)
	}
	return nil
}

// Exists checks if a bid exists by ID
func (r *BidRepository) Exists(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Bid{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check bid existence: %w", err)
	}
	return count > 0, nil
}

// GetRecentBids retrieves recently placed bids
func (r *BidRepository) GetRecentBids(limit int) ([]models.Bid, error) {
	var bids []models.Bid
	if limit <= 0 {
		limit = 20 // Default limit
	}

	if err := r.db.Preload("Auction").
		Preload("Bidder").
		Order("created_at DESC").
		Limit(limit).
		Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent bids: %w", err)
	}

	return bids, nil
}

// GetTopBidsByAmount retrieves bids with highest amounts
func (r *BidRepository) GetTopBidsByAmount(limit int) ([]models.Bid, error) {
	var bids []models.Bid
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if err := r.db.Where("status = ?", "confirmed").
		Preload("Auction").
		Preload("Bidder").
		Order("amount DESC").
		Limit(limit).
		Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get top bids: %w", err)
	}

	return bids, nil
}

// GetBidStats returns bid statistics
func (r *BidRepository) GetBidStats() (*BidStats, error) {
	stats := &BidStats{}

	// Total bids
	if err := r.db.Model(&models.Bid{}).Count(&stats.TotalBids).Error; err != nil {
		return nil, fmt.Errorf("failed to count total bids: %w", err)
	}

	// Confirmed bids
	if err := r.db.Model(&models.Bid{}).
		Where("status = ?", "confirmed").
		Count(&stats.ConfirmedBids).Error; err != nil {
		return nil, fmt.Errorf("failed to count confirmed bids: %w", err)
	}

	// Pending bids
	if err := r.db.Model(&models.Bid{}).
		Where("status = ?", "pending").
		Count(&stats.PendingBids).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending bids: %w", err)
	}

	// Failed bids
	if err := r.db.Model(&models.Bid{}).
		Where("status = ?", "failed").
		Count(&stats.FailedBids).Error; err != nil {
		return nil, fmt.Errorf("failed to count failed bids: %w", err)
	}

	return stats, nil
}

// GetActiveAuctionBids retrieves bids for currently active auctions
func (r *BidRepository) GetActiveAuctionBids(offset, limit int) ([]models.Bid, error) {
	var bids []models.Bid
	query := r.db.Joins("JOIN auctions ON bids.auction_id = auctions.id").
		Where("auctions.status = ? AND auctions.start_time <= ? AND auctions.end_time > ?",
			"active", time.Now(), time.Now()).
		Preload("Auction").
		Preload("Bidder")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("bids.created_at DESC").Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get active auction bids: %w", err)
	}

	return bids, nil
}

// CountBidsByAuction returns the number of bids for a specific auction
func (r *BidRepository) CountBidsByAuction(auctionID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Bid{}).Where("auction_id = ?", auctionID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count bids for auction: %w", err)
	}
	return count, nil
}

// GetBidsByTimeRange retrieves bids within a time range
func (r *BidRepository) GetBidsByTimeRange(startTime, endTime time.Time, offset, limit int) ([]models.Bid, error) {
	var bids []models.Bid
	query := r.db.Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Preload("Auction").
		Preload("Bidder")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&bids).Error; err != nil {
		return nil, fmt.Errorf("failed to get bids by time range: %w", err)
	}

	return bids, nil
}

// BidStats represents bid statistics
type BidStats struct {
	TotalBids     int64 `json:"total_bids"`
	ConfirmedBids int64 `json:"confirmed_bids"`
	PendingBids   int64 `json:"pending_bids"`
	FailedBids    int64 `json:"failed_bids"`
}
