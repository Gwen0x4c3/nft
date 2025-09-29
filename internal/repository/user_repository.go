package repository

import (
	"errors"
	"fmt"
	"strings"

	"nft-platform/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles database operations for User entities
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			if strings.Contains(err.Error(), "username") {
				return fmt.Errorf("username already exists: %w", err)
			}
			if strings.Contains(err.Error(), "email") {
				return fmt.Errorf("email already exists: %w", err)
			}
			if strings.Contains(err.Error(), "wallet_address") {
				return fmt.Errorf("wallet address already exists: %w", err)
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// GetByUsername retrieves a user by their username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with username %s not found", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// GetByWalletAddress retrieves a user by their wallet address
func (r *UserRepository) GetByWalletAddress(walletAddr string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("wallet_address = ?", walletAddr).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with wallet address %s not found", walletAddr)
		}
		return nil, fmt.Errorf("failed to get user by wallet address: %w", err)
	}
	return &user, nil
}

// Update updates an existing user
func (r *UserRepository) Update(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			if strings.Contains(err.Error(), "username") {
				return fmt.Errorf("username already exists: %w", err)
			}
			if strings.Contains(err.Error(), "email") {
				return fmt.Errorf("email already exists: %w", err)
			}
			if strings.Contains(err.Error(), "wallet_address") {
				return fmt.Errorf("wallet address already exists: %w", err)
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete deletes a user from the database
func (r *UserRepository) Delete(id uint) error {
	result := r.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}

// List retrieves users with pagination and filtering
func (r *UserRepository) List(offset, limit int, verified *bool) ([]models.User, error) {
	var users []models.User
	query := r.db.Model(&models.User{})

	// Apply verification filter if provided
	if verified != nil {
		query = query.Where("is_verified = ?", *verified)
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

	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users, optionally filtered by verification status
func (r *UserRepository) Count(verified *bool) (int64, error) {
	var count int64
	query := r.db.Model(&models.User{})

	if verified != nil {
		query = query.Where("is_verified = ?", *verified)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// Search searches for users by username or email with partial matching
func (r *UserRepository) Search(query string, offset, limit int) ([]models.User, error) {
	var users []models.User
	searchPattern := "%" + strings.ToLower(query) + "%"

	dbQuery := r.db.Model(&models.User{}).
		Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern)

	// Apply pagination
	if offset > 0 {
		dbQuery = dbQuery.Offset(offset)
	}
	if limit > 0 {
		dbQuery = dbQuery.Limit(limit)
	} else {
		dbQuery = dbQuery.Limit(50) // Default limit for search
	}

	if err := dbQuery.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// GetWithNFTs retrieves a user with their created NFTs
func (r *UserRepository) GetWithNFTs(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("NFTs").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user with NFTs: %w", err)
	}
	return &user, nil
}

// GetWithBids retrieves a user with their bid history
func (r *UserRepository) GetWithBids(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Preload("Bids").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user with bids: %w", err)
	}
	return &user, nil
}

// Exists checks if a user exists by ID
func (r *UserRepository) Exists(id uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence by email: %w", err)
	}
	return count > 0, nil
}

// ExistsByUsername checks if a user exists by username
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence by username: %w", err)
	}
	return count > 0, nil
}

// ExistsByWalletAddress checks if a user exists by wallet address
func (r *UserRepository) ExistsByWalletAddress(walletAddr string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("wallet_address = ?", walletAddr).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check user existence by wallet address: %w", err)
	}
	return count > 0, nil
}

// UpdateVerificationStatus updates the verification status of a user
func (r *UserRepository) UpdateVerificationStatus(id uint, verified bool) error {
	result := r.db.Model(&models.User{}).Where("id = ?", id).Update("is_verified", verified)
	if result.Error != nil {
		return fmt.Errorf("failed to update verification status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}

// GetRecentUsers retrieves recently registered users
func (r *UserRepository) GetRecentUsers(limit int) ([]models.User, error) {
	var users []models.User
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if err := r.db.Order("created_at DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent users: %w", err)
	}

	return users, nil
}

// GetVerifiedUsers retrieves all verified users
func (r *UserRepository) GetVerifiedUsers(offset, limit int) ([]models.User, error) {
	var users []models.User
	query := r.db.Where("is_verified = ?", true)

	// Apply pagination
	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	} else {
		query = query.Limit(100) // Default limit
	}

	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get verified users: %w", err)
	}

	return users, nil
}