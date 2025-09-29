package repository

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"nft-platform/internal/models"
	"gorm.io/gorm"
)

// NFTFilter represents filtering options for NFT queries
type NFTFilter struct {
	CreatorID   *uint
	OwnerID     *uint
	IsForSale   *bool
	MinPrice    *big.Int
	MaxPrice    *big.Int
	SearchQuery string
	Royalty     *uint8
}

// NFTRepository handles database operations for NFT entities
type NFTRepository struct {
	db *gorm.DB
}

// NewNFTRepository creates a new NFTRepository instance
func NewNFTRepository(db *gorm.DB) *NFTRepository {
	return &NFTRepository{db: db}
}

// Create creates a new NFT in the database
func (r *NFTRepository) Create(nft *models.NFT) error {
	if err := r.db.Create(nft).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") && strings.Contains(err.Error(), "token_id") {
			return fmt.Errorf("NFT with token ID already exists: %w", err)
		}
		return fmt.Errorf("failed to create NFT: %w", err)
	}
	return nil
}

// GetByID retrieves an NFT by its ID
func (r *NFTRepository) GetByID(id uint) (*models.NFT, error) {
	var nft models.NFT
	if err := r.db.Preload("Creator").Preload("Owner").First(&nft, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("NFT with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get NFT by id: %w", err)
	}
	return &nft, nil
}

// GetByTokenID retrieves an NFT by its token ID
func (r *NFTRepository) GetByTokenID(tokenID string) (*models.NFT, error) {
	var nft models.NFT
	if err := r.db.Preload("Creator").Preload("Owner").Where("token_id = ?", tokenID).First(&nft).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("NFT with token ID %s not found", tokenID)
		}
		return nil, fmt.Errorf("failed to get NFT by token ID: %w", err)
	}
	return &nft, nil
}

// Update updates an existing NFT
func (r *NFTRepository) Update(nft *models.NFT) error {
	if err := r.db.Save(nft).Error; err != nil {
		return fmt.Errorf("failed to update NFT: %w", err)
	}
	return nil
}

// Delete deletes an NFT from the database
func (r *NFTRepository) Delete(id uint) error {
	result := r.db.Delete(&models.NFT{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete NFT: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("NFT with id %d not found", id)
	}
	return nil
}

// List retrieves NFTs with filtering and pagination
func (r *NFTRepository) List(filter *NFTFilter, offset, limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	query := r.db.Model(&models.NFT{}).Preload("Creator").Preload("Owner")

	// Apply filters
	if filter != nil {
		if filter.CreatorID != nil {
			query = query.Where("creator_id = ?", *filter.CreatorID)
		}
		if filter.OwnerID != nil {
			query = query.Where("owner_id = ?", *filter.OwnerID)
		}
		if filter.IsForSale != nil {
			query = query.Where("is_for_sale = ?", *filter.IsForSale)
		}
		if filter.MinPrice != nil {
			query = query.Where("price >= ?", filter.MinPrice.String())
		}
		if filter.MaxPrice != nil {
			query = query.Where("price <= ?", filter.MaxPrice.String())
		}
		if filter.SearchQuery != "" {
			searchPattern := "%" + strings.ToLower(filter.SearchQuery) + "%"
			query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
		}
		if filter.Royalty != nil {
			query = query.Where("royalty = ?", *filter.Royalty)
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

	if err := query.Order("created_at DESC").Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to list NFTs: %w", err)
	}

	return nfts, nil
}

// Count returns the total number of NFTs matching the filter
func (r *NFTRepository) Count(filter *NFTFilter) (int64, error) {
	var count int64
	query := r.db.Model(&models.NFT{})

	// Apply filters
	if filter != nil {
		if filter.CreatorID != nil {
			query = query.Where("creator_id = ?", *filter.CreatorID)
		}
		if filter.OwnerID != nil {
			query = query.Where("owner_id = ?", *filter.OwnerID)
		}
		if filter.IsForSale != nil {
			query = query.Where("is_for_sale = ?", *filter.IsForSale)
		}
		if filter.MinPrice != nil {
			query = query.Where("price >= ?", filter.MinPrice.String())
		}
		if filter.MaxPrice != nil {
			query = query.Where("price <= ?", filter.MaxPrice.String())
		}
		if filter.SearchQuery != "" {
			searchPattern := "%" + strings.ToLower(filter.SearchQuery) + "%"
			query = query.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
		}
		if filter.Royalty != nil {
			query = query.Where("royalty = ?", *filter.Royalty)
		}
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count NFTs: %w", err)
	}

	return count, nil
}

// GetByCreator retrieves NFTs created by a specific user
func (r *NFTRepository) GetByCreator(creatorID uint, offset, limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	query := r.db.Where("creator_id = ?", creatorID).Preload("Creator").Preload("Owner")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to get NFTs by creator: %w", err)
	}

	return nfts, nil
}

// GetByOwner retrieves NFTs owned by a specific user
func (r *NFTRepository) GetByOwner(ownerID uint, offset, limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	query := r.db.Where("owner_id = ?", ownerID).Preload("Creator").Preload("Owner")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to get NFTs by owner: %w", err)
	}

	return nfts, nil
}

// GetForSale retrieves NFTs that are currently for sale
func (r *NFTRepository) GetForSale(offset, limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	query := r.db.Where("is_for_sale = ?", true).Preload("Creator").Preload("Owner")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("price ASC").Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to get NFTs for sale: %w", err)
	}

	return nfts, nil
}

// GetWithTransfers retrieves an NFT with its transfer history
func (r *NFTRepository) GetWithTransfers(id uint) (*models.NFT, error) {
	var nft models.NFT
	if err := r.db.Preload("Creator").Preload("Owner").Preload("TransferHistory").First(&nft, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("NFT with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get NFT with transfers: %w", err)
	}
	return &nft, nil
}

// GetWithAuctions retrieves an NFT with its auction history
func (r *NFTRepository) GetWithAuctions(id uint) (*models.NFT, error) {
	var nft models.NFT
	if err := r.db.Preload("Creator").Preload("Owner").Preload("Auctions").First(&nft, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("NFT with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get NFT with auctions: %w", err)
	}
	return &nft, nil
}

// UpdateSaleStatus updates the sale status and price of an NFT
func (r *NFTRepository) UpdateSaleStatus(id uint, isForSale bool, price *big.Int) error {
	updates := map[string]interface{}{
		"is_for_sale": isForSale,
	}

	if price != nil {
		updates["price"] = price.String()
	} else {
		updates["price"] = nil
	}

	result := r.db.Model(&models.NFT{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update NFT sale status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("NFT with id %d not found", id)
	}
	return nil
}

// UpdateOwner updates the owner of an NFT
func (r *NFTRepository) UpdateOwner(id uint, newOwnerID uint) error {
	result := r.db.Model(&models.NFT{}).Where("id = ?", id).Update("owner_id", newOwnerID)
	if result.Error != nil {
		return fmt.Errorf("failed to update NFT owner: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("NFT with id %d not found", id)
	}
	return nil
}

// Search performs full-text search on NFT title and description
func (r *NFTRepository) Search(query string, offset, limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	searchPattern := "%" + strings.ToLower(query) + "%"

	dbQuery := r.db.Model(&models.NFT{}).
		Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern).
		Preload("Creator").Preload("Owner")

	// Apply pagination
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}
	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	} else {
		dbQuery = dbQuery.Limit(50) // Default limit for search
	}

	if err := dbQuery.Order("created_at DESC").Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to search NFTs: %w", err)
	}

	return nfts, nil
}

// Exists checks if an NFT exists by ID
func (r *NFTRepository) Exists(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.NFT{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check NFT existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByTokenID checks if an NFT exists by token ID
func (r *NFTRepository) ExistsByTokenID(tokenID string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.NFT{}).Where("token_id = ?", tokenID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check NFT existence by token ID: %w", err)
	}
	return count > 0, nil
}

// GetRecentNFTs retrieves recently created NFTs
func (r *NFTRepository) GetRecentNFTs(limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if err := r.db.Preload("Creator").Preload("Owner").Order("created_at DESC").Limit(limit).Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent NFTs: %w", err)
	}

	return nfts, nil
}

// GetTopSellingNFTs retrieves NFTs with highest prices that are for sale
func (r *NFTRepository) GetTopSellingNFTs(limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if err := r.db.Where("is_for_sale = ? AND price IS NOT NULL", true).
		Preload("Creator").Preload("Owner").
		Order("price DESC").Limit(limit).Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to get top selling NFTs: %w", err)
	}

	return nfts, nil
}

// GetNFTsByContract retrieves NFTs from a specific contract
func (r *NFTRepository) GetNFTsByContract(contractAddr string, offset, limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	query := r.db.Where("contract_address = ?", contractAddr).Preload("Creator").Preload("Owner")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to get NFTs by contract: %w", err)
	}

	return nfts, nil
}

// GetNFTsWithRoyalties retrieves NFTs that have royalties set
func (r *NFTRepository) GetNFTsWithRoyalties(offset, limit int) ([]models.NFT, error) {
	var nfts []models.NFT
	query := r.db.Where("royalty > 0").Preload("Creator").Preload("Owner")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("royalty DESC").Find(&nfts).Error; err != nil {
		return nil, fmt.Errorf("failed to get NFTs with royalties: %w", err)
	}

	return nfts, nil
}

// GetMarketplaceStats returns marketplace statistics
func (r *NFTRepository) GetMarketplaceStats() (*MarketplaceStats, error) {
	stats := &MarketplaceStats{}

	// Total NFTs count
	if err := r.db.Model(&models.NFT{}).Count(&stats.TotalNFTs).Error; err != nil {
		return nil, fmt.Errorf("failed to count total NFTs: %w", err)
	}

	// NFTs for sale count
	if err := r.db.Model(&models.NFT{}).Where("is_for_sale = ?", true).Count(&stats.NFTsForSale).Error; err != nil {
		return nil, fmt.Errorf("failed to count NFTs for sale: %w", err)
	}

	// Unique creators count
	if err := r.db.Model(&models.NFT{}).Distinct("creator_id").Count(&stats.UniqueCreators).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique creators: %w", err)
	}

	// Unique owners count
	if err := r.db.Model(&models.NFT{}).Distinct("owner_id").Count(&stats.UniqueOwners).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique owners: %w", err)
	}

	return stats, nil
}

// MarketplaceStats represents marketplace statistics
type MarketplaceStats struct {
	TotalNFTs      int64 `json:"total_nfts"`
	NFTsForSale    int64 `json:"nfts_for_sale"`
	UniqueCreators int64 `json:"unique_creators"`
	UniqueOwners   int64 `json:"unique_owners"`
}