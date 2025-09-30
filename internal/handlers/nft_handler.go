package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nft-platform/internal/repository"
	"nft-platform/internal/service"
	"nft-platform/internal/types"
)

// NFTHandler handles NFT-related HTTP requests
type NFTHandler struct {
	nftService *service.NFTService
}

// NewNFTHandler creates a new NFT handler
func NewNFTHandler(nftService *service.NFTService) *NFTHandler {
	return &NFTHandler{
		nftService: nftService,
	}
}

// ListNFTs godoc
// @Summary List NFTs with filtering and pagination
// @Description Retrieves a paginated list of NFTs with optional filtering
// @Tags NFTs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param creator_id query int false "Filter by creator ID"
// @Param owner_id query int false "Filter by owner ID"
// @Param is_for_sale query bool false "Filter by sale status"
// @Param sort query string false "Sort field" Enums(created_at, price, title) default(created_at)
// @Param order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} NFTListResponse "NFTs retrieved"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /nfts [get]
func (h *NFTHandler) ListNFTs(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and limit page size
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Build filter
	filter := repository.NFTFilter{}

	// Parse optional filters
	if creatorID := c.Query("creator_id"); creatorID != "" {
		if id, err := strconv.ParseUint(creatorID, 10, 32); err == nil {
			uid := uint(id)
			filter.CreatorID = &uid
		}
	}

	if ownerID := c.Query("owner_id"); ownerID != "" {
		if id, err := strconv.ParseUint(ownerID, 10, 32); err == nil {
			uid := uint(id)
			filter.OwnerID = &uid
		}
	}

	if isForSale := c.Query("is_for_sale"); isForSale != "" {
		if val, err := strconv.ParseBool(isForSale); err == nil {
			filter.IsForSale = &val
		}
	}

	// Get NFTs from service
	nftsResp, err := h.nftService.ListNFTs(context.Background(), &filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to retrieve NFTs",
			Code:    "FETCH_FAILED",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Return response with pagination already included
	c.JSON(http.StatusOK, nftsResp)
}

// MintNFT godoc
// @Summary Mint new NFT
// @Description Creates and mints a new NFT on the blockchain
// @Tags NFTs
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string true "NFT title"
// @Param description formData string false "NFT description"
// @Param image formData file true "NFT image file"
// @Param metadata formData string false "NFT metadata JSON"
// @Param royalty formData int false "Royalty percentage (0-10)"
// @Success 201 {object} models.NFT "NFT minted successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /nfts [post]
func (h *NFTHandler) MintNFT(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse multipart form",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Get form fields
	title := c.PostForm("title")
	description := c.PostForm("description")
	royaltyStr := c.DefaultPostForm("royalty", "0")

	// Parse royalty
	royalty, err := strconv.ParseUint(royaltyStr, 10, 8)
	if err != nil || royalty > 10 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_royalty",
			Message: "Royalty must be between 0 and 10",
			Code:    "INVALID_PARAMETER",
		})
		return
	}

	// Get image file
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_image",
			Message: "Image file is required",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Build mint request
	mintReq := &types.MintNFTRequest{
		Title:       title,
		Description: description,
		Image:       file,
		Royalty:     uint8(royalty),
	}

	// Parse metadata if provided
	if metadataStr := c.PostForm("metadata"); metadataStr != "" {
		// In a real implementation, parse JSON metadata
		// For now, we'll skip this or use a simple map
		mintReq.Metadata = map[string]interface{}{
			"raw": metadataStr,
		}
	}

	// Call service to mint NFT
	nft, err := h.nftService.MintNFT(context.Background(), uid, mintReq)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "MINT_FAILED"
		message := err.Error()

		if contains(message, "validation") {
			statusCode = http.StatusBadRequest
			errorCode = "VALIDATION_FAILED"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
		})
		return
	}

	c.JSON(http.StatusCreated, nft)
}

// GetNFTByID godoc
// @Summary Get NFT by ID
// @Description Retrieves detailed information about a specific NFT
// @Tags NFTs
// @Accept json
// @Produce json
// @Param nftId path int true "NFT ID"
// @Success 200 {object} models.NFT "NFT found"
// @Failure 400 {object} ErrorResponse "Invalid NFT ID"
// @Failure 404 {object} ErrorResponse "NFT not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /nfts/{nftId} [get]
func (h *NFTHandler) GetNFTByID(c *gin.Context) {
	// Parse NFT ID from URL
	nftIDStr := c.Param("nftId")
	nftID, err := strconv.ParseUint(nftIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid NFT ID",
			Code:    "INVALID_PARAMETER",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Get NFT from service
	nft, err := h.nftService.GetNFT(context.Background(), uint(nftID))
	if err != nil {
		statusCode := http.StatusNotFound
		errorCode := "NFT_NOT_FOUND"
		message := "NFT not found"

		if !contains(err.Error(), "not found") {
			statusCode = http.StatusInternalServerError
			errorCode = "FETCH_FAILED"
			message = "Failed to retrieve NFT"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, nft)
}

// UpdateNFT godoc
// @Summary Update NFT (owner only)
// @Description Updates NFT pricing and sale status (owner only)
// @Tags NFTs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param nftId path int true "NFT ID"
// @Param request body types.UpdateNFTRequest true "Update request"
// @Success 200 {object} models.NFT "NFT updated"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - not owner"
// @Failure 404 {object} ErrorResponse "NFT not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /nfts/{nftId} [put]
func (h *NFTHandler) UpdateNFT(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Parse NFT ID
	nftIDStr := c.Param("nftId")
	nftID, err := strconv.ParseUint(nftIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid NFT ID",
			Code:    "INVALID_PARAMETER",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Bind request body
	var req types.UpdateNFTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Update NFT
	nft, err := h.nftService.UpdateNFT(context.Background(), uint(nftID), uid, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "UPDATE_FAILED"
		message := err.Error()

		if contains(message, "not found") {
			statusCode = http.StatusNotFound
			errorCode = "NFT_NOT_FOUND"
		} else if contains(message, "not owner") || contains(message, "permission") {
			statusCode = http.StatusForbidden
			errorCode = "FORBIDDEN"
		} else if contains(message, "validation") {
			statusCode = http.StatusBadRequest
			errorCode = "VALIDATION_FAILED"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
		})
		return
	}

	c.JSON(http.StatusOK, nft)
}

// TransferNFT godoc
// @Summary Transfer NFT to another user
// @Description Initiates NFT transfer to another wallet address
// @Tags NFTs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param nftId path int true "NFT ID"
// @Param request body types.TransferNFTRequest true "Transfer request"
// @Success 200 {object} models.Transfer "Transfer initiated"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - not owner"
// @Failure 404 {object} ErrorResponse "NFT not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /nfts/{nftId}/transfer [post]
func (h *NFTHandler) TransferNFT(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Parse NFT ID
	nftIDStr := c.Param("nftId")
	nftID, err := strconv.ParseUint(nftIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid NFT ID",
			Code:    "INVALID_PARAMETER",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Bind request body
	var req types.TransferNFTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Transfer NFT
	transfer, err := h.nftService.TransferNFT(context.Background(), uint(nftID), uid, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "TRANSFER_FAILED"
		message := err.Error()

		if contains(message, "not found") {
			statusCode = http.StatusNotFound
			errorCode = "NFT_NOT_FOUND"
		} else if contains(message, "not owner") || contains(message, "permission") {
			statusCode = http.StatusForbidden
			errorCode = "FORBIDDEN"
		} else if contains(message, "validation") {
			statusCode = http.StatusBadRequest
			errorCode = "VALIDATION_FAILED"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
		})
		return
	}

	c.JSON(http.StatusOK, transfer)
}

// NFTListResponse represents the NFT list response with pagination
type NFTListResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}
