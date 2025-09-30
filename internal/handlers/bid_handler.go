package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nft-platform/internal/service"
)

// BidHandler handles bid-related HTTP requests
type BidHandler struct {
	auctionService *service.AuctionService
}

// NewBidHandler creates a new bid handler
func NewBidHandler(auctionService *service.AuctionService) *BidHandler {
	return &BidHandler{
		auctionService: auctionService,
	}
}

// GetAuctionBids godoc
// @Summary Get auction bids
// @Description Retrieves all bids for a specific auction with pagination
// @Tags Bids
// @Accept json
// @Produce json
// @Param auctionId path int true "Auction ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} service.BidListResponse "Bids retrieved"
// @Failure 400 {object} ErrorResponse "Invalid parameters"
// @Failure 404 {object} ErrorResponse "Auction not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auctions/{auctionId}/bids [get]
func (h *BidHandler) GetAuctionBids(c *gin.Context) {
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

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	// Validate and limit page size
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Get bids from service
	bids, err := h.auctionService.GetAuctionBids(context.Background(), uint(auctionID), page, limit)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "FETCH_FAILED"

		// Check for specific errors
		if err.Error() == "auction not found" {
			statusCode = http.StatusNotFound
			code = "AUCTION_NOT_FOUND"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to retrieve bids",
			Code:    code,
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Return bids as array response
	c.JSON(http.StatusOK, gin.H{
		"data": bids,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(bids),
		},
	})
}

// PlaceBid godoc
// @Summary Place bid on auction
// @Description Places a new bid on an active auction
// @Tags Bids
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param auctionId path int true "Auction ID"
// @Param bid body service.PlaceBidRequest true "Bid data"
// @Success 201 {object} service.BidResponse "Bid placed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request or bid amount"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - seller cannot bid"
// @Failure 404 {object} ErrorResponse "Auction not found"
// @Failure 409 {object} ErrorResponse "Bid too low or auction not active"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auctions/{auctionId}/bids [post]
func (h *BidHandler) PlaceBid(c *gin.Context) {
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

	bidderID := userID.(uint)

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

	// Parse request body
	var req service.PlaceBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request format",
			Code:    "INVALID_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Place bid through service
	bidResponse, err := h.auctionService.PlaceBid(context.Background(), uint(auctionID), bidderID, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		code := "BID_FAILED"

		// Check for specific errors
		if err.Error() == "auction not found" {
			statusCode = http.StatusNotFound
			code = "AUCTION_NOT_FOUND"
		} else if err.Error() == "seller cannot bid on own auction" || err.Error() == "cannot bid on own auction" {
			statusCode = http.StatusForbidden
			code = "FORBIDDEN"
		} else if err.Error() == "auction is not active" || err.Error() == "auction not active" {
			statusCode = http.StatusConflict
			code = "AUCTION_NOT_ACTIVE"
		} else if err.Error() == "bid amount too low" || err.Error() == "bid must be higher than current bid" {
			statusCode = http.StatusConflict
			code = "BID_TOO_LOW"
		} else if err.Error() == "invalid bid amount" {
			statusCode = http.StatusBadRequest
			code = "INVALID_AMOUNT"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   "bid_failed",
			Message: "Failed to place bid",
			Code:    code,
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, bidResponse)
}
