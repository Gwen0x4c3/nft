package service

import (
	"mime/multipart"

	"nft-platform/internal/models"
)

// MintNFTRequest represents the NFT minting request
type MintNFTRequest struct {
	Title       string                 `json:"title" validate:"required,min=1,max=100"`
	Description string                 `json:"description" validate:"max=1000"`
	Image       *multipart.FileHeader  `json:"-"` // File upload
	Metadata    map[string]interface{} `json:"metadata"`
	Royalty     uint8                  `json:"royalty" validate:"max=10"`
}

// UpdateNFTRequest represents the NFT update request
type UpdateNFTRequest struct {
	Price     *string `json:"price,omitempty"` // Wei amount as string
	IsForSale *bool   `json:"is_for_sale,omitempty"`
}

// TransferNFTRequest represents the NFT transfer request
type TransferNFTRequest struct {
	ToAddress string `json:"to_address" validate:"required,eth_addr"`
}

// NFTListResponse represents the NFT list response
type NFTListResponse struct {
	Data       []models.NFT   `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NFTMetadata represents the standard NFT metadata format (ERC-721/OpenSea compatible)
type NFTMetadata struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Image           string                 `json:"image"`
	ExternalURL     string                 `json:"external_url,omitempty"`
	AnimationURL    string                 `json:"animation_url,omitempty"`
	Attributes      []MetadataAttribute    `json:"attributes,omitempty"`
	Properties      map[string]interface{} `json:"properties,omitempty"`
	BackgroundColor string                 `json:"background_color,omitempty"`
	YouTubeURL      string                 `json:"youtube_url,omitempty"`
}

// MetadataAttribute represents an NFT attribute (trait)
type MetadataAttribute struct {
	TraitType   string      `json:"trait_type"`
	Value       interface{} `json:"value"`
	DisplayType string      `json:"display_type,omitempty"` // "boost_number", "boost_percentage", "number", "date"
	MaxValue    interface{} `json:"max_value,omitempty"`
}

// NFTServiceConfig represents configuration for the NFT service
type NFTServiceConfig struct {
	BlockchainService BlockchainService
	IPFSService       IPFSService
	DefaultGatewayURL string
	MaxFileSize       int64  // Max file size in bytes
	AllowedFileTypes  []string
	AutoPinToIPFS     bool
}

// NFTTransferHistory represents transfer history response
type NFTTransferHistory struct {
	Transfers  []models.Transfer `json:"transfers"`
	Pagination PaginationInfo    `json:"pagination"`
}

// NFTSearchRequest represents search parameters for NFTs
type NFTSearchRequest struct {
	Query        string   `json:"query,omitempty"`
	CreatorID    *uint    `json:"creator_id,omitempty"`
	OwnerID      *uint    `json:"owner_id,omitempty"`
	IsForSale    *bool    `json:"is_for_sale,omitempty"`
	MinPrice     *string  `json:"min_price,omitempty"` // Wei as string
	MaxPrice     *string  `json:"max_price,omitempty"` // Wei as string
	Categories   []string `json:"categories,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	SortBy       string   `json:"sort_by,omitempty"` // "created_at", "price", "name"
	SortOrder    string   `json:"sort_order,omitempty"` // "asc", "desc"
	Page         int      `json:"page,omitempty"`
	Limit        int      `json:"limit,omitempty"`
}

// NFTStatsResponse represents NFT statistics
type NFTStatsResponse struct {
	TotalNFTs      int64  `json:"total_nfts"`
	TotalCreators  int64  `json:"total_creators"`
	TotalOwners    int64  `json:"total_owners"`
	TotalForSale   int64  `json:"total_for_sale"`
	TotalVolume    string `json:"total_volume"` // Total trading volume in Wei
	AveragePrice   string `json:"average_price"` // Average price in Wei
	FloorPrice     string `json:"floor_price"` // Lowest price in Wei
	TotalTransfers int64  `json:"total_transfers"`
}

// NFTCollectionStats represents statistics for an NFT collection
type NFTCollectionStats struct {
	CreatorID      uint   `json:"creator_id"`
	CreatorName    string `json:"creator_name"`
	TotalNFTs      int64  `json:"total_nfts"`
	TotalSold      int64  `json:"total_sold"`
	TotalVolume    string `json:"total_volume"`
	FloorPrice     string `json:"floor_price"`
	AveragePrice   string `json:"average_price"`
	LastSalePrice  string `json:"last_sale_price"`
	LastSaleDate   int64  `json:"last_sale_date"`
}

// NFTValidationError represents validation errors for NFT operations
type NFTValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

func (e NFTValidationError) Error() string {
	return e.Message
}

// NFTOperationResult represents the result of an NFT operation
type NFTOperationResult struct {
	Success       bool     `json:"success"`
	NFT           *models.NFT `json:"nft,omitempty"`
	Transfer      *models.Transfer `json:"transfer,omitempty"`
	TransactionHash string `json:"transaction_hash,omitempty"`
	IPFSHash      string   `json:"ipfs_hash,omitempty"`
	Errors        []NFTValidationError `json:"errors,omitempty"`
	Message       string   `json:"message,omitempty"`
}

// Predefined metadata attribute display types
const (
	DisplayTypeNumber     = "number"
	DisplayTypeBoostNumber = "boost_number"
	DisplayTypeBoostPercentage = "boost_percentage" 
	DisplayTypeDate       = "date"
	DisplayTypeString     = ""  // Default, empty string
)

// Common NFT categories
var (
	NFTCategoryArt         = "Art"
	NFTCategoryMusic       = "Music"
	NFTCategoryDomainNames = "Domain Names"
	NFTCategoryVirtualWorlds = "Virtual Worlds"
	NFTCategoryTradingCards = "Trading Cards"
	NFTCategoryCollectibles = "Collectibles"
	NFTCategorySports      = "Sports"
	NFTCategoryUtility     = "Utility"
	
	// All available categories
	AvailableCategories = []string{
		NFTCategoryArt,
		NFTCategoryMusic,
		NFTCategoryDomainNames,
		NFTCategoryVirtualWorlds,
		NFTCategoryTradingCards,
		NFTCategoryCollectibles,
		NFTCategorySports,
		NFTCategoryUtility,
	}
)

// Common sort options for NFT listings
const (
	SortByCreatedAt = "created_at"
	SortByPrice     = "price"
	SortByName      = "name"
	SortByUpdatedAt = "updated_at"
	SortByPopularity = "popularity"
	
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// File validation constants
const (
	MaxImageSize = 10 * 1024 * 1024 // 10MB
	MaxVideoSize = 100 * 1024 * 1024 // 100MB
	MaxAudioSize = 50 * 1024 * 1024  // 50MB
)

// Allowed file types for NFT uploads
var (
	AllowedImageTypes = []string{
		"image/jpeg",
		"image/png", 
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}
	
	AllowedVideoTypes = []string{
		"video/mp4",
		"video/webm",
		"video/ogg",
	}
	
	AllowedAudioTypes = []string{
		"audio/mpeg",
		"audio/wav",
		"audio/ogg",
	}
	
	AllowedFileTypes = append(append(AllowedImageTypes, AllowedVideoTypes...), AllowedAudioTypes...)
)