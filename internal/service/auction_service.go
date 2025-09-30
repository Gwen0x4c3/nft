package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"nft-platform/internal/models"
	"nft-platform/internal/repository"
	"nft-platform/internal/types"
	"nft-platform/pkg/validation"
)

// AuctionService handles auction-related business operations
type AuctionService struct {
	auctionRepo      *repository.AuctionRepository
	nftRepo          *repository.NFTRepository
	userRepo         *repository.UserRepository
	transferRepo     *repository.TransferRepository
	validator        *validation.CustomValidator
	blockchainSvc    BlockchainService
	notificationSvc  *NotificationService
	minimumIncrement *big.Int
}

// AuctionServiceConfig represents configuration for the auction service
type AuctionServiceConfig struct {
	BlockchainService   BlockchainService
	NotificationService *NotificationService
	MinimumIncrement    *big.Int
}

// NewAuctionService creates a new AuctionService instance
func NewAuctionService(
	auctionRepo *repository.AuctionRepository,
	nftRepo *repository.NFTRepository,
	userRepo *repository.UserRepository,
	transferRepo *repository.TransferRepository,
	config *AuctionServiceConfig,
) *AuctionService {
	// Default minimum increment is 0.01 ETH (in Wei)
	minimumIncrement := new(big.Int)
	minimumIncrement.SetString("10000000000000000", 10) // 0.01 ETH in Wei

	if config != nil && config.MinimumIncrement != nil {
		minimumIncrement = config.MinimumIncrement
	}

	var notificationService *NotificationService
	if config != nil {
		notificationService = config.NotificationService
	}

	return &AuctionService{
		auctionRepo:      auctionRepo,
		nftRepo:          nftRepo,
		userRepo:         userRepo,
		transferRepo:     transferRepo,
		validator:        validation.NewCustomValidator(),
		blockchainSvc:    config.BlockchainService,
		notificationSvc:  notificationService,
		minimumIncrement: minimumIncrement,
	}
}

// CreateAuctionRequest represents the auction creation request
type CreateAuctionRequest struct {
	NFTID        uint      `json:"nft_id" validate:"required"`
	StartPrice   string    `json:"start_price" validate:"required"`
	ReservePrice string    `json:"reserve_price,omitempty"`
	StartTime    time.Time `json:"start_time" validate:"required"`
	EndTime      time.Time `json:"end_time" validate:"required"`
}

// PlaceBidRequest represents the bid placement request
type PlaceBidRequest struct {
	Amount string `json:"amount" validate:"required"`
}

// BidResponse represents the bid placement response
type BidResponse struct {
	Bid             *models.Bid `json:"bid"`
	IsHighest       bool        `json:"is_highest"`
	PreviousHighBid string      `json:"previous_high_bid"`
	AuctionUpdated  bool        `json:"auction_updated"`
}

// AuctionListResponse represents auction list with pagination
type AuctionListResponse struct {
	Data       []models.Auction     `json:"data"`
	Pagination types.PaginationInfo `json:"pagination"`
}

// CreateAuction creates a new auction for an NFT
func (s *AuctionService) CreateAuction(ctx context.Context, sellerID uint, req *CreateAuctionRequest) (*models.Auction, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get NFT and verify ownership
	nft, err := s.nftRepo.GetByID(req.NFTID)
	if err != nil {
		return nil, fmt.Errorf("NFT not found: %w", err)
	}

	if nft.OwnerID != sellerID {
		return nil, fmt.Errorf("not authorized to auction this NFT")
	}

	// Check if NFT already has an active auction
	hasActiveAuction, err := s.auctionRepo.HasActiveAuctionForNFT(req.NFTID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing auctions: %w", err)
	}
	if hasActiveAuction {
		return nil, fmt.Errorf("NFT already has an active auction")
	}

	// Parse and validate prices
	startPrice, err := s.parsePrice(req.StartPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid start price: %w", err)
	}

	var reservePrice *big.Int
	if req.ReservePrice != "" {
		reservePrice, err = s.parsePrice(req.ReservePrice)
		if err != nil {
			return nil, fmt.Errorf("invalid reserve price: %w", err)
		}
		if reservePrice.Cmp(startPrice) < 0 {
			return nil, fmt.Errorf("reserve price must be greater than or equal to start price")
		}
	}

	// Validate auction timing
	now := time.Now()
	if req.StartTime.Before(now) {
		return nil, fmt.Errorf("start time must be in the future")
	}
	if req.EndTime.Before(req.StartTime) {
		return nil, fmt.Errorf("end time must be after start time")
	}
	if req.EndTime.Sub(req.StartTime) < time.Hour {
		return nil, fmt.Errorf("auction duration must be at least 1 hour")
	}

	// Create auction
	auction := &models.Auction{
		NFTID:        req.NFTID,
		SellerID:     sellerID,
		StartPrice:   startPrice,
		ReservePrice: reservePrice,
		CurrentBid:   big.NewInt(0),
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Status:       "pending",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.auctionRepo.Create(auction); err != nil {
		return nil, fmt.Errorf("failed to create auction: %w", err)
	}

	// Remove NFT from direct sale when auction is created
	nft.IsForSale = false
	nft.Price = nil
	if err := s.nftRepo.Update(nft); err != nil {
		// Log warning but don't fail auction creation
		fmt.Printf("Warning: failed to update NFT sale status: %v", err)
	}

	// Send notification to seller about auction creation
	if s.notificationSvc != nil {
		notificationData := map[string]interface{}{
			"auction_id": auction.ID,
			"nft_id":     nft.ID,
			"nft_title":  nft.Title,
		}
		s.notificationSvc.SendNotification(ctx, sellerID, "auction_started",
			"Auction Created",
			fmt.Sprintf("Your auction for %s has been created", nft.Title),
			notificationData)
	}

	// Load auction with relationships
	return s.auctionRepo.GetByID(auction.ID)
}

// GetAuction retrieves an auction by ID
func (s *AuctionService) GetAuction(ctx context.Context, auctionID uint) (*models.Auction, error) {
	return s.auctionRepo.GetByID(auctionID)
}

// ListAuctions retrieves auctions with filtering and pagination
func (s *AuctionService) ListAuctions(ctx context.Context, filter *repository.AuctionFilter, page, limit int) (*AuctionListResponse, error) {
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get auctions
	auctions, err := s.auctionRepo.List(filter, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list auctions: %w", err)
	}

	// Get total count
	total, err := s.auctionRepo.Count(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count auctions: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &AuctionListResponse{
		Data: auctions,
		Pagination: types.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetActiveAuctions retrieves currently active auctions
func (s *AuctionService) GetActiveAuctions(ctx context.Context, page, limit int) (*AuctionListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Get active auctions
	auctions, err := s.auctionRepo.GetActiveAuctions(offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get active auctions: %w", err)
	}

	// For total count, we need to count active auctions
	isActive := true
	filter := &repository.AuctionFilter{IsActive: &isActive}
	total, err := s.auctionRepo.Count(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count active auctions: %w", err)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return &AuctionListResponse{
		Data: auctions,
		Pagination: types.PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// PlaceBid places a bid on an auction
func (s *AuctionService) PlaceBid(ctx context.Context, auctionID uint, bidderID uint, req *PlaceBidRequest) (*BidResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get auction with relationships
	auction, err := s.auctionRepo.GetWithBids(auctionID)
	if err != nil {
		return nil, fmt.Errorf("auction not found: %w", err)
	}

	// Validate auction status and timing
	if auction.Status != "active" {
		return nil, fmt.Errorf("auction is not active")
	}

	now := time.Now()
	if now.Before(auction.StartTime) {
		return nil, fmt.Errorf("auction has not started yet")
	}
	if now.After(auction.EndTime) {
		return nil, fmt.Errorf("auction has ended")
	}

	// Verify bidder exists and is not the seller
	bidder, err := s.userRepo.GetByID(bidderID)
	if err != nil {
		return nil, fmt.Errorf("bidder not found: %w", err)
	}

	if bidderID == auction.SellerID {
		return nil, fmt.Errorf("seller cannot bid on their own auction")
	}

	// Parse and validate bid amount
	bidAmount, err := s.parsePrice(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid bid amount: %w", err)
	}

	// Validate bid amount against current highest bid
	requiredAmount := new(big.Int).Set(auction.StartPrice)
	if auction.CurrentBid != nil && auction.CurrentBid.Sign() > 0 {
		requiredAmount.Add(auction.CurrentBid, s.minimumIncrement)
	}

	if bidAmount.Cmp(requiredAmount) < 0 {
		return nil, fmt.Errorf("bid must be at least %s Wei", requiredAmount.String())
	}

	// Create bid record
	bid := &models.Bid{
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    bidAmount,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Execute blockchain transaction if blockchain service is available
	if s.blockchainSvc != nil {
		tx, err := s.blockchainSvc.PlaceBid(ctx, auction.NFT.TokenID, bidder.WalletAddr, bidAmount)
		if err != nil {
			return nil, fmt.Errorf("blockchain bid failed: %w", err)
		}
		bid.TxHash = tx.Hash
		bid.Status = "confirmed"
	} else {
		// For testing, generate mock transaction hash
		bid.TxHash = s.generateMockTxHash()
		bid.Status = "confirmed"
	}

	// Create bid in repository (with its own transaction)
	bidRepo := repository.NewBidRepository(s.auctionRepo.GetDB())
	if err := bidRepo.Create(bid); err != nil {
		return nil, fmt.Errorf("failed to create bid record: %w", err)
	}

	// Store previous high bid for response
	previousHighBid := auction.GetCurrentBidWei()

	// Update auction with new highest bid
	isHighest := true
	auctionUpdated := false
	if err := s.auctionRepo.UpdateCurrentBid(auctionID, bidAmount, bidderID); err != nil {
		fmt.Printf("Warning: failed to update auction current bid: %v", err)
	} else {
		auctionUpdated = true
	}

	// Send notifications
	if s.notificationSvc != nil {
		s.sendBidNotifications(ctx, auction, bid, bidder)
	}

	// Check if auction should be extended (anti-sniping)
	if now.Add(15 * time.Minute).After(auction.EndTime) {
		newEndTime := now.Add(15 * time.Minute)
		auction.EndTime = newEndTime
		if err := s.auctionRepo.Update(auction); err != nil {
			fmt.Printf("Warning: failed to extend auction end time: %v", err)
		}
	}

	// Load bid with relationships
	bidWithRelations, err := bidRepo.GetByID(bid.ID)
	if err != nil {
		// Return bid without relations if loading fails
		bidWithRelations = bid
	}

	return &BidResponse{
		Bid:             bidWithRelations,
		IsHighest:       isHighest,
		PreviousHighBid: previousHighBid,
		AuctionUpdated:  auctionUpdated,
	}, nil
}

// GetAuctionBids retrieves bids for an auction
func (s *AuctionService) GetAuctionBids(ctx context.Context, auctionID uint, page, limit int) ([]models.Bid, error) {
	// Verify auction exists
	_, err := s.auctionRepo.GetByID(auctionID)
	if err != nil {
		return nil, fmt.Errorf("auction not found: %w", err)
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	bidRepo := repository.NewBidRepository(s.auctionRepo.GetDB())
	return bidRepo.GetByAuctionID(auctionID, offset, limit)
}

// CancelAuction cancels an auction (only if no bids placed)
func (s *AuctionService) CancelAuction(ctx context.Context, auctionID uint, sellerID uint) error {
	// Get auction
	auction, err := s.auctionRepo.GetWithBids(auctionID)
	if err != nil {
		return fmt.Errorf("auction not found: %w", err)
	}

	// Verify seller authorization
	if auction.SellerID != sellerID {
		return fmt.Errorf("not authorized to cancel this auction")
	}

	// Check if auction can be cancelled
	if auction.Status == "ended" || auction.Status == "cancelled" {
		return fmt.Errorf("auction cannot be cancelled")
	}

	// Check if there are any confirmed bids
	for _, bid := range auction.Bids {
		if bid.Status == "confirmed" {
			return fmt.Errorf("cannot cancel auction with confirmed bids")
		}
	}

	// Update auction status
	if err := s.auctionRepo.UpdateStatus(auctionID, "cancelled"); err != nil {
		return fmt.Errorf("failed to cancel auction: %w", err)
	}

	// Send notification to seller
	if s.notificationSvc != nil {
		notificationData := map[string]interface{}{
			"auction_id": auction.ID,
			"nft_id":     auction.NFT.ID,
			"nft_title":  auction.NFT.Title,
		}
		s.notificationSvc.SendNotification(ctx, sellerID, "auction_cancelled",
			"Auction Cancelled",
			fmt.Sprintf("Your auction for %s has been cancelled", auction.NFT.Title),
			notificationData)
	}

	return nil
}

// EndAuction processes auction ending and determines winner
func (s *AuctionService) EndAuction(ctx context.Context, auctionID uint) (*models.Auction, error) {
	// Get auction with bids
	auction, err := s.auctionRepo.GetWithBids(auctionID)
	if err != nil {
		return nil, fmt.Errorf("auction not found: %w", err)
	}

	// Check if auction should be ended
	if auction.Status == "ended" || auction.Status == "cancelled" {
		return auction, nil // Already processed
	}

	now := time.Now()
	if now.Before(auction.EndTime) && auction.Status == "active" {
		return nil, fmt.Errorf("auction has not ended yet")
	}

	// Find highest confirmed bid
	var winningBid *models.Bid
	for _, bid := range auction.Bids {
		if bid.Status == "confirmed" {
			if winningBid == nil || bid.Amount.Cmp(winningBid.Amount) > 0 {
				winningBid = &bid
			}
		}
	}

	// Check if reserve price was met
	hasWinner := false
	if winningBid != nil {
		if auction.HasReservePrice() {
			hasWinner = winningBid.Amount.Cmp(auction.ReservePrice) >= 0
		} else {
			hasWinner = true
		}
	}

	// Update auction status and winner
	auction.Status = "ended"
	auction.UpdatedAt = now

	if hasWinner {
		auction.WinnerID = &winningBid.BidderID
		auction.CurrentBid = winningBid.Amount

		// Transfer NFT ownership to winner
		if err := s.transferNFTToWinner(ctx, auction, winningBid); err != nil {
			fmt.Printf("Warning: failed to transfer NFT to winner: %v", err)
		}
	}

	// Update auction in database
	if err := s.auctionRepo.Update(auction); err != nil {
		return nil, fmt.Errorf("failed to update auction: %w", err)
	}

	// Send notifications
	if s.notificationSvc != nil {
		s.sendAuctionEndNotifications(ctx, auction, winningBid, hasWinner)
	}

	return auction, nil
}

// ProcessExpiredAuctions processes auctions that have expired but haven't been ended
func (s *AuctionService) ProcessExpiredAuctions(ctx context.Context) error {
	expiredAuctions, err := s.auctionRepo.GetExpiredAuctions(50)
	if err != nil {
		return fmt.Errorf("failed to get expired auctions: %w", err)
	}

	for _, auction := range expiredAuctions {
		if _, err := s.EndAuction(ctx, auction.ID); err != nil {
			fmt.Printf("Error processing expired auction %d: %v\n", auction.ID, err)
		}
	}

	return nil
}

// Helper methods

func (s *AuctionService) parsePrice(priceStr string) (*big.Int, error) {
	if priceStr == "" {
		return nil, fmt.Errorf("price cannot be empty")
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

func (s *AuctionService) generateMockTxHash() string {
	// For testing purposes
	return fmt.Sprintf("0x%d%d", time.Now().Unix(), time.Now().Nanosecond())
}

func (s *AuctionService) sendBidNotifications(ctx context.Context, auction *models.Auction, bid *models.Bid, bidder *models.User) {
	nftTitle := "Unknown NFT"
	if auction.NFT.Title != "" {
		nftTitle = auction.NFT.Title
	}

	bidData := map[string]interface{}{
		"auction_id": auction.ID,
		"nft_id":     auction.NFTID,
		"nft_title":  nftTitle,
		"bid_id":     bid.ID,
		"amount":     bid.GetAmountWei(),
		"bidder_id":  bid.BidderID,
	}

	// Notify seller about new bid
	s.notificationSvc.SendNotification(ctx, auction.SellerID, "bid_placed",
		"New Bid Received",
		fmt.Sprintf("New bid of %s Wei placed on your auction for %s", bid.GetAmountWei(), nftTitle),
		bidData)

	// Notify previous highest bidder that they've been outbid
	if auction.WinnerID != nil && *auction.WinnerID != bid.BidderID {
		s.notificationSvc.SendNotification(ctx, *auction.WinnerID, "bid_outbid",
			"You've Been Outbid",
			fmt.Sprintf("Your bid on %s has been exceeded", nftTitle),
			bidData)
	}
}

func (s *AuctionService) sendAuctionEndNotifications(ctx context.Context, auction *models.Auction, winningBid *models.Bid, hasWinner bool) {
	nftTitle := "Unknown NFT"
	if auction.NFT.Title != "" {
		nftTitle = auction.NFT.Title
	}

	auctionData := map[string]interface{}{
		"auction_id": auction.ID,
		"nft_id":     auction.NFTID,
		"nft_title":  nftTitle,
	}

	if hasWinner && winningBid != nil {
		// Notify winner
		winnerData := auctionData
		winnerData["amount"] = winningBid.GetAmountWei()
		s.notificationSvc.SendNotification(ctx, winningBid.BidderID, "auction_won",
			"Auction Won!",
			fmt.Sprintf("Congratulations! You won the auction for %s", nftTitle),
			winnerData)

		// Notify seller
		sellerData := auctionData
		sellerData["amount"] = winningBid.GetAmountWei()
		sellerData["buyer_id"] = winningBid.BidderID
		s.notificationSvc.SendNotification(ctx, auction.SellerID, "nft_sold",
			"NFT Sold",
			fmt.Sprintf("Your NFT %s has been sold for %s Wei", nftTitle, winningBid.GetAmountWei()),
			sellerData)
	} else {
		// Notify seller that auction ended without winner
		s.notificationSvc.SendNotification(ctx, auction.SellerID, "auction_ended",
			"Auction Ended",
			fmt.Sprintf("Your auction for %s has ended without meeting the reserve price", nftTitle),
			auctionData)
	}
}

func (s *AuctionService) transferNFTToWinner(ctx context.Context, auction *models.Auction, winningBid *models.Bid) error {
	// Get NFT
	nft, err := s.nftRepo.GetByID(auction.NFTID)
	if err != nil {
		return fmt.Errorf("failed to get NFT: %w", err)
	}

	// Update NFT ownership
	nft.OwnerID = winningBid.BidderID
	nft.IsForSale = false
	nft.Price = nil
	nft.UpdatedAt = time.Now()

	if err := s.nftRepo.Update(nft); err != nil {
		return fmt.Errorf("failed to update NFT ownership: %w", err)
	}

	// Create transfer record
	transfer := &models.Transfer{
		NFTID:     nft.ID,
		FromID:    &auction.SellerID,
		ToID:      winningBid.BidderID,
		TxHash:    winningBid.TxHash,
		Price:     winningBid.Amount,
		Type:      "sale",
		CreatedAt: time.Now(),
	}

	if err := s.transferRepo.Create(transfer); err != nil {
		return fmt.Errorf("failed to create transfer record: %w", err)
	}

	return nil
}
