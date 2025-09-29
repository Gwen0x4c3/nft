package repository

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"nft-platform/internal/models"
	"gorm.io/gorm"
)

// AuctionFilter represents filtering options for auction queries
type AuctionFilter struct {
	Status         *string
	SellerID       *uint
	NFTID          *uint
	MinStartPrice  *big.Int
	MaxStartPrice  *big.Int
	IsActive       *bool
	HasBids        *bool
	EndingAfter    *time.Time
	EndingBefore   *time.Time
	HasReservePrice *bool
}

// AuctionRepository handles database operations for Auction entities
type AuctionRepository struct {
	db *gorm.DB
}

// NewAuctionRepository creates a new AuctionRepository instance
func NewAuctionRepository(db *gorm.DB) *AuctionRepository {
	return &AuctionRepository{db: db}
}

// Create creates a new auction in the database
func (r *AuctionRepository) Create(auction *models.Auction) error {
	if err := r.db.Create(auction).Error; err != nil {
		return fmt.Errorf("failed to create auction: %w", err)
	}
	return nil
}

// GetByID retrieves an auction by its ID with related data
func (r *AuctionRepository) GetByID(id uint) (*models.Auction, error) {
	var auction models.Auction
	if err := r.db.Preload("NFT").Preload("Seller").Preload("Winner").
		Preload("Bids", func(db *gorm.DB) *gorm.DB {
			return db.Order("amount DESC, created_at ASC")
		}).
		Preload("Bids.Bidder").
		First(&auction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("auction with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get auction by id: %w", err)
	}
	return &auction, nil
}

// Update updates an existing auction
func (r *AuctionRepository) Update(auction *models.Auction) error {
	if err := r.db.Save(auction).Error; err != nil {
		return fmt.Errorf("failed to update auction: %w", err)
	}
	return nil
}

// Delete deletes an auction from the database
func (r *AuctionRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Auction{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete auction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("auction with id %d not found", id)
	}
	return nil
}

// List retrieves auctions with filtering and pagination
func (r *AuctionRepository) List(filter *AuctionFilter, offset, limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	query := r.db.Model(&models.Auction{}).
		Preload("NFT").
		Preload("Seller").
		Preload("Winner")

	// Apply filters
	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.SellerID != nil {
			query = query.Where("seller_id = ?", *filter.SellerID)
		}
		if filter.NFTID != nil {
			query = query.Where("nft_id = ?", *filter.NFTID)
		}
		if filter.MinStartPrice != nil {
			query = query.Where("start_price >= ?", filter.MinStartPrice.String())
		}
		if filter.MaxStartPrice != nil {
			query = query.Where("start_price <= ?", filter.MaxStartPrice.String())
		}
		if filter.IsActive != nil && *filter.IsActive {
			query = query.Where("status = ? AND start_time <= ? AND end_time > ?",
				"active", time.Now(), time.Now())
		}
		if filter.HasBids != nil {
			if *filter.HasBids {
				query = query.Where("current_bid > 0")
			} else {
				query = query.Where("current_bid = 0")
			}
		}
		if filter.EndingAfter != nil {
			query = query.Where("end_time > ?", *filter.EndingAfter)
		}
		if filter.EndingBefore != nil {
			query = query.Where("end_time < ?", *filter.EndingBefore)
		}
		if filter.HasReservePrice != nil {
			if *filter.HasReservePrice {
				query = query.Where("reserve_price IS NOT NULL AND reserve_price > 0")
			} else {
				query = query.Where("reserve_price IS NULL OR reserve_price = 0")
			}
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

	if err := query.Order("created_at DESC").Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to list auctions: %w", err)
	}

	return auctions, nil
}

// Count returns the total number of auctions matching the filter
func (r *AuctionRepository) Count(filter *AuctionFilter) (int64, error) {
	var count int64
	query := r.db.Model(&models.Auction{})

	// Apply same filters as List
	if filter != nil {
		if filter.Status != nil {
			query = query.Where("status = ?", *filter.Status)
		}
		if filter.SellerID != nil {
			query = query.Where("seller_id = ?", *filter.SellerID)
		}
		if filter.NFTID != nil {
			query = query.Where("nft_id = ?", *filter.NFTID)
		}
		if filter.MinStartPrice != nil {
			query = query.Where("start_price >= ?", filter.MinStartPrice.String())
		}
		if filter.MaxStartPrice != nil {
			query = query.Where("start_price <= ?", filter.MaxStartPrice.String())
		}
		if filter.IsActive != nil && *filter.IsActive {
			query = query.Where("status = ? AND start_time <= ? AND end_time > ?",
				"active", time.Now(), time.Now())
		}
		if filter.HasBids != nil {
			if *filter.HasBids {
				query = query.Where("current_bid > 0")
			} else {
				query = query.Where("current_bid = 0")
			}
		}
		if filter.EndingAfter != nil {
			query = query.Where("end_time > ?", *filter.EndingAfter)
		}
		if filter.EndingBefore != nil {
			query = query.Where("end_time < ?", *filter.EndingBefore)
		}
		if filter.HasReservePrice != nil {
			if *filter.HasReservePrice {
				query = query.Where("reserve_price IS NOT NULL AND reserve_price > 0")
			} else {
				query = query.Where("reserve_price IS NULL OR reserve_price = 0")
			}
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count auctions: %w", err)
	}

	return count, nil
}

// GetActiveAuctions retrieves currently active auctions
func (r *AuctionRepository) GetActiveAuctions(offset, limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	now := time.Now()
	query := r.db.Where("status = ? AND start_time <= ? AND end_time > ?", "active", now, now).
		Preload("NFT").
		Preload("Seller").
		Preload("Winner")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("end_time ASC").Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to get active auctions: %w", err)
	}

	return auctions, nil
}

// GetEndingSoon retrieves auctions ending within the specified duration
func (r *AuctionRepository) GetEndingSoon(within time.Duration, limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	now := time.Now()
	endTime := now.Add(within)

	query := r.db.Where("status = ? AND end_time > ? AND end_time <= ?", "active", now, endTime).
		Preload("NFT").
		Preload("Seller").
		Preload("Winner")

	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(50) // Default limit
	}

	if err := query.Order("end_time ASC").Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to get auctions ending soon: %w", err)
	}

	return auctions, nil
}

// GetExpiredAuctions retrieves auctions that have ended but status hasn't been updated
func (r *AuctionRepository) GetExpiredAuctions(limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	now := time.Now()

	query := r.db.Where("status IN (?, ?) AND end_time <= ?", "pending", "active", now).
		Preload("NFT").
		Preload("Seller").
		Preload("Winner")

	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("end_time ASC").Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to get expired auctions: %w", err)
	}

	return auctions, nil
}

// GetBySeller retrieves auctions created by a specific seller
func (r *AuctionRepository) GetBySeller(sellerID uint, offset, limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	query := r.db.Where("seller_id = ?", sellerID).
		Preload("NFT").
		Preload("Seller").
		Preload("Winner")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to get auctions by seller: %w", err)
	}

	return auctions, nil
}

// GetByNFT retrieves auctions for a specific NFT
func (r *AuctionRepository) GetByNFT(nftID uint, offset, limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	query := r.db.Where("nft_id = ?", nftID).
		Preload("NFT").
		Preload("Seller").
		Preload("Winner").
		Preload("Bids", func(db *gorm.DB) *gorm.DB {
			return db.Order("amount DESC, created_at ASC").Limit(10) // Limit bids for performance
		}).
		Preload("Bids.Bidder")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(50) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to get auctions by NFT: %w", err)
	}

	return auctions, nil
}

// GetWithBids retrieves an auction with all its bids
func (r *AuctionRepository) GetWithBids(id uint) (*models.Auction, error) {
	var auction models.Auction
	if err := r.db.Preload("NFT").
		Preload("Seller").
		Preload("Winner").
		Preload("Bids", func(db *gorm.DB) *gorm.DB {
			return db.Order("amount DESC, created_at ASC")
		}).
		Preload("Bids.Bidder").
		First(&auction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("auction with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get auction with bids: %w", err)
	}
	return &auction, nil
}

// UpdateStatus updates the status of an auction
func (r *AuctionRepository) UpdateStatus(id uint, status string) error {
	result := r.db.Model(&models.Auction{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("failed to update auction status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("auction with id %d not found", id)
	}
	return nil
}

// UpdateCurrentBid updates the current bid and winner of an auction
func (r *AuctionRepository) UpdateCurrentBid(id uint, amount *big.Int, winnerID uint) error {
	updates := map[string]interface{}{
		"current_bid": amount.String(),
		"winner_id":   winnerID,
	}

	result := r.db.Model(&models.Auction{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update auction current bid: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("auction with id %d not found", id)
	}
	return nil
}

// Exists checks if an auction exists by ID
func (r *AuctionRepository) Exists(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Auction{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check auction existence: %w", err)
	}
	return count > 0, nil
}

// HasActiveAuctionForNFT checks if an NFT has an active auction
func (r *AuctionRepository) HasActiveAuctionForNFT(nftID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Auction{}).
		Where("nft_id = ? AND status IN (?, ?)", nftID, "pending", "active").
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check active auction for NFT: %w", err)
	}
	return count > 0, nil
}

// GetHighestBidderForAuction retrieves the highest bidder for an auction
func (r *AuctionRepository) GetHighestBidderForAuction(auctionID uint) (*models.User, *big.Int, error) {
	var bid models.Bid
	if err := r.db.Where("auction_id = ? AND status = ?", auctionID, "confirmed").
		Preload("Bidder").
		Order("amount DESC, created_at ASC").
		First(&bid).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil // No bids found
		}
		return nil, nil, fmt.Errorf("failed to get highest bidder: %w", err)
	}
	return &bid.Bidder, bid.Amount, nil
}

// GetAuctionStats returns auction statistics
func (r *AuctionRepository) GetAuctionStats() (*AuctionStats, error) {
	stats := &AuctionStats{}

	// Total auctions
	if err := r.db.Model(&models.Auction{}).Count(&stats.TotalAuctions).Error; err != nil {
		return nil, fmt.Errorf("failed to count total auctions: %w", err)
	}

	// Active auctions
	now := time.Now()
	if err := r.db.Model(&models.Auction{}).
		Where("status = ? AND start_time <= ? AND end_time > ?", "active", now, now).
		Count(&stats.ActiveAuctions).Error; err != nil {
		return nil, fmt.Errorf("failed to count active auctions: %w", err)
	}

	// Ended auctions
	if err := r.db.Model(&models.Auction{}).
		Where("status = ?", "ended").
		Count(&stats.EndedAuctions).Error; err != nil {
		return nil, fmt.Errorf("failed to count ended auctions: %w", err)
	}

	// Auctions with bids
	if err := r.db.Model(&models.Auction{}).
		Where("current_bid > 0").
		Count(&stats.AuctionsWithBids).Error; err != nil {
		return nil, fmt.Errorf("failed to count auctions with bids: %w", err)
	}

	return stats, nil
}

// GetTopAuctionsByCurrent bid retrieves auctions with highest current bids
func (r *AuctionRepository) GetTopAuctionsByCurrentBid(limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if err := r.db.Where("current_bid > 0").
		Preload("NFT").
		Preload("Seller").
		Preload("Winner").
		Order("current_bid DESC").
		Limit(limit).
		Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to get top auctions by current bid: %w", err)
	}

	return auctions, nil
}

// GetRecentAuctions retrieves recently created auctions
func (r *AuctionRepository) GetRecentAuctions(limit int) ([]models.Auction, error) {
	var auctions []models.Auction
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if err := r.db.Preload("NFT").
		Preload("Seller").
		Preload("Winner").
		Order("created_at DESC").
		Limit(limit).
		Find(&auctions).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent auctions: %w", err)
	}

	return auctions, nil
}

// AuctionStats represents auction statistics
type AuctionStats struct {
	TotalAuctions     int64 `json:"total_auctions"`
	ActiveAuctions    int64 `json:"active_auctions"`
	EndedAuctions     int64 `json:"ended_auctions"`
	AuctionsWithBids  int64 `json:"auctions_with_bids"`
}