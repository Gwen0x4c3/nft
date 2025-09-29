package repository

import (
	"fmt"
	"nft-platform/internal/models"
	"strings"

	"gorm.io/gorm"
)

// TransferRepository handles database operations for Transfer entities
type TransferRepository struct {
	db *gorm.DB
}

// NewTransferRepository creates a new TransferRepository instance
func NewTransferRepository(db *gorm.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

// Create creates a new transfer in the database
func (r *TransferRepository) Create(transfer *models.Transfer) error {
	if err := r.db.Create(transfer).Error; err != nil {
		return fmt.Errorf("failed to create transfer: %w", err)
	}
	return nil
}

// GetByID retrieves a transfer by its ID
func (r *TransferRepository) GetByID(id uint) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.Preload("NFT").Preload("From").Preload("To").First(&transfer, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transfer with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get transfer by id: %w", err)
	}

	return &transfer, nil
}

// GetByNFTID retrieves transfers for a specific NFT with pagination
func (r *TransferRepository) GetByNFTID(nftID uint, offset, limit int) ([]models.Transfer, error) {
	var transfers []models.Transfer
	query := r.db.Where("nft_id = ?", nftID).
		Preload("NFT").
		Preload("From").
		Preload("To")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get transfers by NFT ID: %w", err)
	}

	return transfers, nil
}

// GetByUserID retrieves transfers where user is sender or recipient
func (r *TransferRepository) GetByUserID(userID uint, offset, limit int) ([]models.Transfer, error) {
	var transfers []models.Transfer
	query := r.db.Where("from_id = ? OR to_id = ?", userID, userID).
		Preload("NFT").
		Preload("From").
		Preload("To")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get transfers by user ID: %w", err)
	}

	return transfers, nil
}

// GetByTxHash retrieves a transfer by transaction hash
func (r *TransferRepository) GetByTxHash(txHash string) (*models.Transfer, error) {
	var transfer models.Transfer
	if err := r.db.Where("tx_hash = ?", txHash).
		Preload("NFT").
		Preload("From").
		Preload("To").
		First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transfer with tx_hash %s not found", txHash)
		}
		return nil, fmt.Errorf("failed to get transfer by tx_hash: %w", err)
	}
	return &transfer, nil
}

// Update updates an existing transfer
func (r *TransferRepository) Update(transfer *models.Transfer) error {
	if err := r.db.Save(transfer).Error; err != nil {
		return fmt.Errorf("failed to update transfer: %w", err)
	}
	return nil
}

// Delete deletes a transfer from the database
func (r *TransferRepository) Delete(id uint) error {
	result := r.db.Delete(&models.Transfer{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete transfer: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("transfer with id %d not found", id)
	}
	return nil
}

// GetMintTransfers retrieves all mint transfers (where FromID is null)
func (r *TransferRepository) GetMintTransfers(offset, limit int) ([]models.Transfer, error) {
	var transfers []models.Transfer
	query := r.db.Where("from_id IS NULL AND type = ?", "mint").
		Preload("NFT").
		Preload("To")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get mint transfers: %w", err)
	}

	return transfers, nil
}

// GetSaleTransfers retrieves all sale transfers (where Price is not null)
func (r *TransferRepository) GetSaleTransfers(offset, limit int) ([]models.Transfer, error) {
	var transfers []models.Transfer
	query := r.db.Where("price IS NOT NULL AND type = ?", "sale").
		Preload("NFT").
		Preload("From").
		Preload("To")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get sale transfers: %w", err)
	}

	return transfers, nil
}

// CountByNFTID returns the total number of transfers for an NFT
func (r *TransferRepository) CountByNFTID(nftID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Transfer{}).Where("nft_id = ?", nftID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count transfers by NFT ID: %w", err)
	}
	return count, nil
}

// CountByUserID returns the total number of transfers for a user
func (r *TransferRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Transfer{}).
		Where("from_id = ? OR to_id = ?", userID, userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count transfers by user ID: %w", err)
	}
	return count, nil
}

// GetRecentTransfers retrieves recent transfers across all NFTs
func (r *TransferRepository) GetRecentTransfers(limit int) ([]models.Transfer, error) {
	var transfers []models.Transfer
	query := r.db.Preload("NFT").Preload("From").Preload("To")

	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(50) // Default limit for recent transfers
	}

	if err := query.Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent transfers: %w", err)
	}

	return transfers, nil
}

// GetTransfersByType retrieves transfers of a specific type
func (r *TransferRepository) GetTransfersByType(transferType string, offset, limit int) ([]models.Transfer, error) {
	var transfers []models.Transfer

	// Validate transfer type
	validTypes := []string{"mint", "transfer", "sale"}
	isValidType := false
	for _, validType := range validTypes {
		if strings.EqualFold(transferType, validType) {
			transferType = validType
			isValidType = true
			break
		}
	}

	if !isValidType {
		return nil, fmt.Errorf("invalid transfer type: %s", transferType)
	}

	query := r.db.Where("type = ?", transferType).
		Preload("NFT").
		Preload("From").
		Preload("To")

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get transfers by type: %w", err)
	}

	return transfers, nil
}

