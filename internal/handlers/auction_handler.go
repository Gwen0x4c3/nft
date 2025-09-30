package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nft-platform/internal/repository"
	"nft-platform/internal/service"
)

// AuctionHandler handles auction-related HTTP requests
type AuctionHandler struct {
	auctionService *service.AuctionService
}

// NewAuctionHandler creates a new auction handler
func NewAuctionHandler(auctionService *service.AuctionService) *AuctionHandler {
	return &AuctionHandler{
		auctionService: auctionService,
	}
}

// ListAuctions godoc
// @Summary List auctions with filtering and pagination
// @Description Retrieves a paginated list of auctions with optional filtering by status and seller
// @Tags Auctions
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status" Enums(pending, active, ended, cancelled)
// @Param seller_id query int false "Filter by seller ID"
// @Success 200 {object} service.AuctionListResponse "Auctions retrieved"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auctions [get]
func (h *AuctionHandler) ListAuctions(c *gin.Context) {
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
	filter := repository.AuctionFilter{}

	// Parse optional filters
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	if sellerID := c.Query("seller_id"); sellerID != "" {
		if id, err := strconv.ParseUint(sellerID, 10, 32); err == nil {
			uid := uint(id)
			filter.SellerID = &uid
		}
	}

	// Get auctions from service
	auctionsResp, err := h.auctionService.ListAuctions(context.Background(), &filter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to retrieve auctions",
			Code:    "FETCH_FAILED",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Return response with pagination
	c.JSON(http.StatusOK, auctionsResp)
}

// CreateAuction godoc
// @Summary Create new auction
// @Description Creates a new auction for an NFT owned by the authenticated user
// @Tags Auctions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param auction body service.CreateAuctionRequest true "Auction creation data"
// @Success 201 {object} models.Auction "Auction created"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - not NFT owner"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auctions [post]
func (h *AuctionHandler) CreateAuction(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	sellerID := userID.(uint)

	// Parse request body
	var req service.CreateAuctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Create auction
	auction, err := h.auctionService.CreateAuction(context.Background(), sellerID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "CREATE_FAILED"

		// Check for specific errors
		if err.Error() == "NFT not found" || err.Error() == "nft not found" {
			statusCode = http.StatusNotFound
			code = "NFT_NOT_FOUND"
		} else if err.Error() == "not NFT owner" || err.Error() == "user is not the owner of this NFT" {
			statusCode = http.StatusForbidden
			code = "FORBIDDEN"
		} else if err.Error() == "NFT is already in auction" {
			statusCode = http.StatusConflict
			code = "ALREADY_IN_AUCTION"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "create_failed",
			Message: "Failed to create auction",
			Code:    code,
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, auction)
}

// GetAuction godoc
// @Summary Get auction by ID
// @Description Retrieves detailed information about a specific auction
// @Tags Auctions
// @Accept json
// @Produce json
// @Param auctionId path int true "Auction ID"
// @Success 200 {object} models.Auction "Auction found"
// @Failure 400 {object} ErrorResponse "Invalid auction ID"
// @Failure 404 {object} ErrorResponse "Auction not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auctions/{auctionId} [get]
func (h *AuctionHandler) GetAuction(c *gin.Context) {
	// Parse auction ID from path parameter
	auctionID, err := strconv.ParseUint(c.Param("auctionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid auction ID",
			Code:    "INVALID_ID",
		})
		return
	}

	// Get auction from service
	auction, err := h.auctionService.GetAuction(context.Background(), uint(auctionID))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "Auction not found",
			Code:    "NOT_FOUND",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, auction)
}

// CancelAuction godoc
// @Summary Cancel auction
// @Description Cancels an auction (seller only, before first bid)
// @Tags Auctions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param auctionId path int true "Auction ID"
// @Success 200 {object} map[string]interface{} "Auction cancelled"
// @Failure 400 {object} ErrorResponse "Invalid auction ID"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - not auction seller"
// @Failure 404 {object} ErrorResponse "Auction not found"
// @Failure 409 {object} ErrorResponse "Cannot cancel auction with bids"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auctions/{auctionId} [delete]
func (h *AuctionHandler) CancelAuction(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User authentication required",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	sellerID := userID.(uint)

	// Parse auction ID from path parameter
	auctionID, err := strconv.ParseUint(c.Param("auctionId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid auction ID",
			Code:    "INVALID_ID",
		})
		return
	}

	// Cancel auction through service
	err = h.auctionService.CancelAuction(context.Background(), uint(auctionID), sellerID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "CANCEL_FAILED"

		// Check for specific errors
		if err.Error() == "auction not found" {
			statusCode = http.StatusNotFound
			code = "NOT_FOUND"
		} else if err.Error() == "not auction seller" || err.Error() == "only seller can cancel auction" {
			statusCode = http.StatusForbidden
			code = "FORBIDDEN"
		} else if err.Error() == "cannot cancel auction with bids" || err.Error() == "auction already has bids" {
			statusCode = http.StatusConflict
			code = "HAS_BIDS"
		} else if err.Error() == "auction already ended or cancelled" {
			statusCode = http.StatusConflict
			code = "ALREADY_ENDED"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "cancel_failed",
			Message: "Failed to cancel auction",
			Code:    code,
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Auction cancelled successfully",
		"auction_id": auctionID,
	})
}
