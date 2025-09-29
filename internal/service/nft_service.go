package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"nft-platform/internal/models"
	"nft-platform/internal/repository"
	"nft-platform/pkg/validation"
)

// NFTService handles NFT-related business operations
type NFTService struct {
	nftRepo       repository.NFTRepository
	userRepo      repository.UserRepository
	transferRepo  *repository.TransferRepository
	validator     *validation.CustomValidator
	blockchainSvc BlockchainService
	ipfsService   IPFSService
}


// NewNFTService creates a new NFT service instance
func NewNFTService(
	nftRepo repository.NFTRepository,
	userRepo repository.UserRepository,
	transferRepo *repository.TransferRepository,
	config *NFTServiceConfig,
) *NFTService {
	return &NFTService{
		nftRepo:       nftRepo,
		userRepo:      userRepo,
		transferRepo:  transferRepo,
		validator:     validation.NewCustomValidator(),
		blockchainSvc: config.BlockchainService,
		ipfsService:   config.IPFSService,
	}
}

// MintNFT creates a new NFT and mints it on the blockchain
func (s *NFTService) MintNFT(ctx context.Context, creatorID uint, req *MintNFTRequest) (*models.NFT, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify creator exists
	creator, err := s.userRepo.GetByID(creatorID)
	if err != nil {
		return nil, fmt.Errorf("creator not found: %w", err)
	}

	// Upload image to IPFS
	var imageURL string
	if req.Image != nil && s.ipfsService != nil {
		file, err := req.Image.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open image file: %w", err)
		}
		defer file.Close()

		imageURL, err = s.ipfsService.UploadFile(ctx, file, req.Image.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image to IPFS: %w", err)
		}
	} else {
		// For testing/development, use a placeholder
		imageURL = "https://placeholder.example.com/nft-image.jpg"
	}

	// Create metadata
	metadata := NFTMetadata{
		Name:        req.Title,
		Description: req.Description,
		Image:       imageURL,
	}

	// Add custom attributes from request
	if req.Metadata != nil {
		if attrs, ok := req.Metadata["attributes"].([]interface{}); ok {
			for _, attr := range attrs {
				if attrMap, ok := attr.(map[string]interface{}); ok {
					metadata.Attributes = append(metadata.Attributes, MetadataAttribute{
						TraitType: fmt.Sprintf("%v", attrMap["trait_type"]),
						Value:     attrMap["value"],
					})
				}
			}
		}
		if props, ok := req.Metadata["properties"].(map[string]interface{}); ok {
			metadata.Properties = props
		}
	}

	// Upload metadata to IPFS
	var metadataURI string
	if s.ipfsService != nil {
		metadataURI, err = s.ipfsService.UploadJSON(ctx, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to upload metadata to IPFS: %w", err)
		}
	} else {
		// For testing/development, use a placeholder
		metadataURI = "https://placeholder.example.com/nft-metadata.json"
	}

	// Get next token ID from blockchain
	var tokenID string
	if s.blockchainSvc != nil {
		tokenID, err = s.blockchainSvc.GetNextTokenID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next token ID: %w", err)
		}
	} else {
		// For testing/development, generate a random token ID
		tokenID = s.generateTokenID()
	}

	// Create NFT record
	nft := &models.NFT{
		TokenID:      tokenID,
		ContractAddr: s.getContractAddress(),
		CreatorID:    creatorID,
		OwnerID:      creatorID, // Creator is initial owner
		Title:        req.Title,
		Description:  req.Description,
		ImageURL:     imageURL,
		MetadataURI:  metadataURI,
		Price:        nil, // Not for sale initially
		IsForSale:    false,
		Royalty:      req.Royalty,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save NFT to database first (will be updated with mint transaction hash later)
	if err := s.nftRepo.Create(nft); err != nil {
		return nil, fmt.Errorf("failed to create NFT record: %w", err)
	}

	// Mint NFT on blockchain
	if s.blockchainSvc != nil {
		tx, err := s.blockchainSvc.MintNFT(ctx, creator.WalletAddr, metadataURI)
		if err != nil {
			// Rollback NFT creation if blockchain mint fails
			s.nftRepo.Delete(nft.ID)
			return nil, fmt.Errorf("failed to mint NFT on blockchain: %w", err)
		}

		// Update NFT with transaction hash
		nft.MintTxHash = tx.Hash
		if err := s.nftRepo.Update(nft); err != nil {
			// Log error but don't fail the mint since blockchain transaction succeeded
			fmt.Printf("Warning: failed to update NFT with transaction hash: %v", err)
		}

		// Create mint transfer record
		transfer := &models.Transfer{
			NFTID:     nft.ID,
			FromID:    nil, // Null for mint
			ToID:      creatorID,
			TxHash:    tx.Hash,
			Price:     nil,
			Type:      "mint",
			CreatedAt: time.Now(),
		}

		if err := s.transferRepo.Create(transfer); err != nil {
			// Log error but don't fail the mint
			fmt.Printf("Warning: failed to create transfer record: %v", err)
		}
	}

	// Load complete NFT with relationships
	return s.nftRepo.GetByID(nft.ID)
}

// GetNFT retrieves an NFT by ID
func (s *NFTService) GetNFT(ctx context.Context, nftID uint) (*models.NFT, error) {
	nft, err := s.nftRepo.GetByID(nftID)
	if err != nil {
		return nil, fmt.Errorf("NFT not found: %w", err)
	}
	return nft, nil
}

// GetNFTByTokenID retrieves an NFT by token ID
func (s *NFTService) GetNFTByTokenID(ctx context.Context, tokenID string) (*models.NFT, error) {
	nft, err := s.nftRepo.GetByTokenID(tokenID)
	if err != nil {
		return nil, fmt.Errorf("NFT not found: %w", err)
	}
	return nft, nil
}

// ListNFTs retrieves NFTs with filtering and pagination
func (s *NFTService) ListNFTs(ctx context.Context, filter *repository.NFTFilter, page, limit int) (*NFTListResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Default limit
	}

	offset := (page - 1) * limit

	// Get NFTs
	nfts, err := s.nftRepo.List(filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list NFTs: %w", err)
	}

	// Get total count
	total, err := s.nftRepo.Count(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count NFTs: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &NFTListResponse{
		Data: nfts,
		Pagination: PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetNFTsByCreator retrieves NFTs created by a specific user
func (s *NFTService) GetNFTsByCreator(ctx context.Context, creatorID uint, page, limit int) (*NFTListResponse, error) {
	filter := &repository.NFTFilter{
		CreatorID: &creatorID,
	}
	return s.ListNFTs(ctx, filter, page, limit)
}

// GetNFTsByOwner retrieves NFTs owned by a specific user
func (s *NFTService) GetNFTsByOwner(ctx context.Context, ownerID uint, page, limit int) (*NFTListResponse, error) {
	filter := &repository.NFTFilter{
		OwnerID: &ownerID,
	}
	return s.ListNFTs(ctx, filter, page, limit)
}

// UpdateNFT updates NFT details (only by owner)
func (s *NFTService) UpdateNFT(ctx context.Context, nftID uint, userID uint, req *UpdateNFTRequest) (*models.NFT, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get NFT
	nft, err := s.nftRepo.GetByID(nftID)
	if err != nil {
		return nil, fmt.Errorf("NFT not found: %w", err)
	}

	// Check ownership
	if nft.OwnerID != userID {
		return nil, fmt.Errorf("not authorized to update this NFT")
	}

	// Update price if provided
	if req.Price != nil {
		if *req.Price == "" || *req.Price == "0" {
			nft.Price = nil
		} else {
			price := new(big.Int)
			if _, ok := price.SetString(*req.Price, 10); !ok {
				return nil, fmt.Errorf("invalid price format")
			}
			if price.Sign() <= 0 {
				return nil, fmt.Errorf("price must be positive")
			}
			nft.Price = price
		}
	}

	// Update for sale status if provided
	if req.IsForSale != nil {
		nft.IsForSale = *req.IsForSale
		// If setting for sale, ensure price is set
		if *req.IsForSale && (nft.Price == nil || nft.Price.Sign() <= 0) {
			return nil, fmt.Errorf("price must be set when marking NFT for sale")
		}
	}

	nft.UpdatedAt = time.Now()

	// Save updates
	if err := s.nftRepo.Update(nft); err != nil {
		return nil, fmt.Errorf("failed to update NFT: %w", err)
	}

	return nft, nil
}

// TransferNFT transfers NFT ownership
func (s *NFTService) TransferNFT(ctx context.Context, nftID uint, fromUserID uint, req *TransferNFTRequest) (*models.Transfer, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get NFT
	nft, err := s.nftRepo.GetByID(nftID)
	if err != nil {
		return nil, fmt.Errorf("NFT not found: %w", err)
	}

	// Check ownership
	if nft.OwnerID != fromUserID {
		return nil, fmt.Errorf("not authorized to transfer this NFT")
	}

	// Get sender user details
	fromUser, err := s.userRepo.GetByID(fromUserID)
	if err != nil {
		return nil, fmt.Errorf("sender not found: %w", err)
	}

	// Check if transferring to self
	if fromUser.WalletAddr == req.ToAddress {
		return nil, fmt.Errorf("cannot transfer NFT to yourself")
	}

	// Find recipient by wallet address
	toUser, err := s.userRepo.GetByWalletAddress(req.ToAddress)
	if err != nil {
		return nil, fmt.Errorf("recipient not found: %w", err)
	}

	// Execute blockchain transfer
	var txHash string
	if s.blockchainSvc != nil {
		tx, err := s.blockchainSvc.TransferNFT(ctx, nft.TokenID, fromUser.WalletAddr, req.ToAddress)
		if err != nil {
			return nil, fmt.Errorf("blockchain transfer failed: %w", err)
		}
		txHash = tx.Hash
	} else {
		// For testing/development, generate a mock transaction hash
		txHash = s.generateMockTxHash()
	}

	// Update NFT ownership
	nft.OwnerID = toUser.ID
	nft.IsForSale = false // Remove from sale when transferred
	nft.Price = nil
	nft.UpdatedAt = time.Now()

	if err := s.nftRepo.Update(nft); err != nil {
		return nil, fmt.Errorf("failed to update NFT ownership: %w", err)
	}

	// Create transfer record
	transfer := &models.Transfer{
		NFTID:     nft.ID,
		FromID:    &fromUserID,
		ToID:      toUser.ID,
		TxHash:    txHash,
		Price:     nil, // No price for regular transfers
		Type:      "transfer",
		CreatedAt: time.Now(),
	}

	if err := s.transferRepo.Create(transfer); err != nil {
		return nil, fmt.Errorf("failed to create transfer record: %w", err)
	}

	return transfer, nil
}

// GetNFTTransferHistory retrieves transfer history for an NFT
func (s *NFTService) GetNFTTransferHistory(ctx context.Context, nftID uint, page, limit int) ([]models.Transfer, error) {
	// Validate NFT exists
	_, err := s.nftRepo.GetByID(nftID)
	if err != nil {
		return nil, fmt.Errorf("NFT not found: %w", err)
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	transfers, err := s.transferRepo.GetByNFTID(nftID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get transfer history: %w", err)
	}

	return transfers, nil
}

// GetUserTransfers retrieves transfer history for a user
func (s *NFTService) GetUserTransfers(ctx context.Context, userID uint, page, limit int) ([]models.Transfer, error) {
	// Validate user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	transfers, err := s.transferRepo.GetByUserID(userID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user transfers: %w", err)
	}

	return transfers, nil
}

// ValidateNFTOwnership checks if a user owns a specific NFT
func (s *NFTService) ValidateNFTOwnership(ctx context.Context, nftID uint, userID uint) error {
	nft, err := s.nftRepo.GetByID(nftID)
	if err != nil {
		return fmt.Errorf("NFT not found: %w", err)
	}

	if nft.OwnerID != userID {
		return fmt.Errorf("user does not own this NFT")
	}

	return nil
}

// Helper methods

// generateTokenID generates a unique token ID for testing
func (s *NFTService) generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateMockTxHash generates a mock transaction hash for testing
func (s *NFTService) generateMockTxHash() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return "0x" + hex.EncodeToString(bytes)
}

// getContractAddress returns the NFT contract address
func (s *NFTService) getContractAddress() string {
	// This would normally come from blockchain config
	return "0x1234567890123456789012345678901234567890"
}

// ParseImageFile validates and processes uploaded image files
func (s *NFTService) ParseImageFile(file *multipart.FileHeader) error {
	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return fmt.Errorf("image file too large (max 10MB)")
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return nil
		}
	}

	return fmt.Errorf("invalid image format (supported: jpg, png, gif, webp)")
}

// FormatPriceWei formats price from big.Int to string for JSON responses
func (s *NFTService) FormatPriceWei(price *big.Int) string {
	if price == nil {
		return "0"
	}
	return price.String()
}

// ParsePriceWei parses price from string to big.Int
func (s *NFTService) ParsePriceWei(priceStr string) (*big.Int, error) {
	if priceStr == "" || priceStr == "0" {
		return nil, nil
	}

	price := new(big.Int)
	if _, ok := price.SetString(priceStr, 10); !ok {
		return nil, fmt.Errorf("invalid price format")
	}

	if price.Sign() <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}

	return price, nil
}

