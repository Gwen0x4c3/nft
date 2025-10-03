package unit

import (
	"context"
	"math/big"
	"testing"
	"time"

	"nft-platform/internal/models"
	"nft-platform/internal/repository"
	"nft-platform/internal/service"
	"nft-platform/internal/types"
	"nft-platform/pkg/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuctionRepository is a mock implementation of AuctionRepository
type MockAuctionRepository struct {
	mock.Mock
}

func (m *MockAuctionRepository) Create(auction *models.Auction) error {
	args := m.Called(auction)
	return args.Error(0)
}

func (m *MockAuctionRepository) GetByID(id uint) (*models.Auction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) GetWithBids(id uint) (*models.Auction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) Update(auction *models.Auction) error {
	args := m.Called(auction)
	return args.Error(0)
}

func (m *MockAuctionRepository) UpdateStatus(id uint, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockAuctionRepository) UpdateCurrentBid(id uint, amount *big.Int, bidderID uint) error {
	args := m.Called(id, amount, bidderID)
	return args.Error(0)
}

func (m *MockAuctionRepository) List(filter *repository.AuctionFilter, offset, limit int) ([]models.Auction, error) {
	args := m.Called(filter, offset, limit)
	return args.Get(0).([]models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) Count(filter *repository.AuctionFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAuctionRepository) GetActiveAuctions(offset, limit int) ([]models.Auction, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) HasActiveAuctionForNFT(nftID uint) (bool, error) {
	args := m.Called(nftID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuctionRepository) GetExpiredAuctions(limit int) ([]models.Auction, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Auction), args.Error(1)
}

func (m *MockAuctionRepository) GetDB() interface{} {
	args := m.Called()
	return args.Get(0)
}

// MockNFTRepository is a mock implementation of NFTRepository
type MockNFTRepository struct {
	mock.Mock
}

func (m *MockNFTRepository) Create(nft *models.NFT) error {
	args := m.Called(nft)
	return args.Error(0)
}

func (m *MockNFTRepository) GetByID(id uint) (*models.NFT, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NFT), args.Error(1)
}

func (m *MockNFTRepository) Update(nft *models.NFT) error {
	args := m.Called(nft)
	return args.Error(0)
}

func (m *MockNFTRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockNFTRepository) List(filter *repository.NFTFilter, offset, limit int) ([]models.NFT, error) {
	args := m.Called(filter, offset, limit)
	return args.Get(0).([]models.NFT), args.Error(1)
}

func (m *MockNFTRepository) Count(filter *repository.NFTFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNFTRepository) GetByOwnerID(ownerID uint, offset, limit int) ([]models.NFT, error) {
	args := m.Called(ownerID, offset, limit)
	return args.Get(0).([]models.NFT), args.Error(1)
}

func (m *MockNFTRepository) GetByCreatorID(creatorID uint, offset, limit int) ([]models.NFT, error) {
	args := m.Called(creatorID, offset, limit)
	return args.Get(0).([]models.NFT), args.Error(1)
}

// MockBidRepository is a mock implementation of BidRepository
type MockBidRepository struct {
	mock.Mock
}

func (m *MockBidRepository) Create(bid *models.Bid) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockBidRepository) GetByID(id uint) (*models.Bid, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bid), args.Error(1)
}

func (m *MockBidRepository) GetByAuctionID(auctionID uint, offset, limit int) ([]models.Bid, error) {
	args := m.Called(auctionID, offset, limit)
	return args.Get(0).([]models.Bid), args.Error(1)
}

func (m *MockBidRepository) GetByBidderID(bidderID uint, offset, limit int) ([]models.Bid, error) {
	args := m.Called(bidderID, offset, limit)
	return args.Get(0).([]models.Bid), args.Error(1)
}

func (m *MockBidRepository) GetHighestBid(auctionID uint) (*models.Bid, error) {
	args := m.Called(auctionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bid), args.Error(1)
}

// MockBlockchainService is a mock implementation of BlockchainService
type MockBlockchainService struct {
	mock.Mock
}

func (m *MockBlockchainService) PlaceBid(ctx context.Context, tokenID string, walletAddr string, amount *big.Int) (*types.Transaction, error) {
	args := m.Called(ctx, tokenID, walletAddr, amount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBlockchainService) TransferNFT(ctx context.Context, tokenID string, fromAddr, toAddr string) (*types.Transaction, error) {
	args := m.Called(ctx, tokenID, fromAddr, toAddr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

// MockNotificationService is a mock implementation of NotificationService
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(ctx context.Context, userID uint, notificationType, title, message string, data map[string]interface{}) error {
	args := m.Called(ctx, userID, notificationType, title, message, data)
	return args.Error(0)
}

// TestAuctionValidation tests auction model validation
func TestAuctionValidation(t *testing.T) {
	validator := validation.NewCustomValidator()

	tests := []struct {
		name        string
		auction     models.Auction
		expectError bool
		errorField  string
	}{
		{
			name: "Valid auction",
			auction: models.Auction{
				NFTID:      1,
				SellerID:   1,
				StartPrice: big.NewInt(1000000000000000000), // 1 ETH in Wei
				StartTime:  time.Now().Add(time.Hour),
				EndTime:    time.Now().Add(25 * time.Hour),
				Status:     "pending",
			},
			expectError: false,
		},
		{
			name: "Invalid status",
			auction: models.Auction{
				NFTID:      1,
				SellerID:   1,
				StartPrice: big.NewInt(1000000000000000000),
				StartTime:  time.Now().Add(time.Hour),
				EndTime:    time.Now().Add(25 * time.Hour),
				Status:     "invalid_status",
			},
			expectError: true,
			errorField:  "Status",
		},
		{
			name: "End time before start time",
			auction: models.Auction{
				NFTID:      1,
				SellerID:   1,
				StartPrice: big.NewInt(1000000000000000000),
				StartTime:  time.Now().Add(2 * time.Hour),
				EndTime:    time.Now().Add(time.Hour),
				Status:     "pending",
			},
			expectError: true,
		},
		{
			name: "Duration too short (less than 1 hour)",
			auction: models.Auction{
				NFTID:      1,
				SellerID:   1,
				StartPrice: big.NewInt(1000000000000000000),
				StartTime:  time.Now().Add(time.Hour),
				EndTime:    time.Now().Add(90 * time.Minute),
				Status:     "pending",
			},
			expectError: true,
		},
		{
			name: "Reserve price less than start price",
			auction: models.Auction{
				NFTID:        1,
				SellerID:     1,
				StartPrice:   big.NewInt(2000000000000000000), // 2 ETH
				ReservePrice: big.NewInt(1000000000000000000), // 1 ETH
				StartTime:    time.Now().Add(time.Hour),
				EndTime:      time.Now().Add(25 * time.Hour),
				Status:       "pending",
			},
			expectError: true,
		},
		{
			name: "Valid auction with reserve price equal to start price",
			auction: models.Auction{
				NFTID:        1,
				SellerID:     1,
				StartPrice:   big.NewInt(1000000000000000000),
				ReservePrice: big.NewInt(1000000000000000000),
				StartTime:    time.Now().Add(time.Hour),
				EndTime:      time.Now().Add(25 * time.Hour),
				Status:       "pending",
			},
			expectError: false,
		},
		{
			name: "Valid auction without reserve price",
			auction: models.Auction{
				NFTID:      1,
				SellerID:   1,
				StartPrice: big.NewInt(1000000000000000000),
				StartTime:  time.Now().Add(time.Hour),
				EndTime:    time.Now().Add(25 * time.Hour),
				Status:     "pending",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(&tt.auction)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
					assert.Contains(t, err.Error(), tt.errorField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAuctionCreation tests auction creation logic
func TestAuctionCreation(t *testing.T) {
	mockAuctionRepo := &repository.AuctionRepository{db: nil} // Use real repo with nil DB for testing
	mockNFTRepo := &repository.NFTRepository{db: nil}
	mockUserRepo := &repository.UserRepository{db: nil}
	mockTransferRepo := &repository.TransferRepository{db: nil}
	mockNotificationSvc := new(MockNotificationService)

	// Create mock for repository methods using test doubles
	mockAuctionRepoDB := new(MockAuctionRepository)
	mockNFTRepoDB := new(MockNFTRepository)
	mockUserRepoDB := new(MockUserRepository)

	minimumIncrement := big.NewInt(10000000000000000) // 0.01 ETH in Wei

	config := &service.AuctionServiceConfig{
		NotificationService: mockNotificationSvc,
		MinimumIncrement:    minimumIncrement,
	}

	auctionService := service.NewAuctionService(
		mockAuctionRepo,
		mockNFTRepo,
		mockUserRepo,
		mockTransferRepo,
		config,
	)

	ctx := context.Background()
	sellerID := uint(1)
	nftID := uint(1)

	t.Run("Successful auction creation", func(t *testing.T) {
		startTime := time.Now().Add(time.Hour)
		endTime := time.Now().Add(25 * time.Hour)

		req := &service.CreateAuctionRequest{
			NFTID:        nftID,
			StartPrice:   "1000000000000000000", // 1 ETH
			ReservePrice: "1500000000000000000", // 1.5 ETH
			StartTime:    startTime,
			EndTime:      endTime,
		}

		nft := &models.NFT{
			ID:        nftID,
			OwnerID:   sellerID,
			TokenID:   "123",
			Title:     "Test NFT",
			IsForSale: true,
		}

		createdAuction := &models.Auction{
			ID:           1,
			NFTID:        nftID,
			SellerID:     sellerID,
			StartPrice:   big.NewInt(1000000000000000000),
			ReservePrice: big.NewInt(1500000000000000000),
			CurrentBid:   big.NewInt(0),
			StartTime:    startTime,
			EndTime:      endTime,
			Status:       "pending",
		}

		// Setup mocks
		mockNFTRepo.On("GetByID", nftID).Return(nft, nil)
		mockAuctionRepo.On("HasActiveAuctionForNFT", nftID).Return(false, nil)
		mockAuctionRepo.On("Create", mock.AnythingOfType("*models.Auction")).Return(nil)
		mockNFTRepo.On("Update", mock.AnythingOfType("*models.NFT")).Return(nil)
		mockAuctionRepo.On("GetByID", mock.AnythingOfType("uint")).Return(createdAuction, nil)
		mockNotificationSvc.On("SendNotification", ctx, sellerID, "auction_started",
			"Auction Created", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil)

		auction, err := auctionService.CreateAuction(ctx, sellerID, req)

		assert.NoError(t, err)
		assert.NotNil(t, auction)
		assert.Equal(t, nftID, auction.NFTID)
		assert.Equal(t, sellerID, auction.SellerID)
		assert.Equal(t, "pending", auction.Status)

		mockNFTRepo.AssertExpectations(t)
		mockAuctionRepo.AssertExpectations(t)
		mockNotificationSvc.AssertExpectations(t)
	})

	t.Run("Failed auction creation - not NFT owner", func(t *testing.T) {
		req := &service.CreateAuctionRequest{
			NFTID:      nftID,
			StartPrice: "1000000000000000000",
			StartTime:  time.Now().Add(time.Hour),
			EndTime:    time.Now().Add(25 * time.Hour),
		}

		nft := &models.NFT{
			ID:      nftID,
			OwnerID: 2, // Different owner
			TokenID: "123",
			Title:   "Test NFT",
		}

		mockNFTRepo.On("GetByID", nftID).Return(nft, nil)

		auction, err := auctionService.CreateAuction(ctx, sellerID, req)

		assert.Error(t, err)
		assert.Nil(t, auction)
		assert.Contains(t, err.Error(), "not authorized to auction this NFT")

		mockNFTRepo.AssertExpectations(t)
	})

	t.Run("Failed auction creation - NFT has active auction", func(t *testing.T) {
		req := &service.CreateAuctionRequest{
			NFTID:      nftID,
			StartPrice: "1000000000000000000",
			StartTime:  time.Now().Add(time.Hour),
			EndTime:    time.Now().Add(25 * time.Hour),
		}

		nft := &models.NFT{
			ID:      nftID,
			OwnerID: sellerID,
			TokenID: "123",
			Title:   "Test NFT",
		}

		mockNFTRepo.On("GetByID", nftID).Return(nft, nil)
		mockAuctionRepo.On("HasActiveAuctionForNFT", nftID).Return(true, nil)

		auction, err := auctionService.CreateAuction(ctx, sellerID, req)

		assert.Error(t, err)
		assert.Nil(t, auction)
		assert.Contains(t, err.Error(), "NFT already has an active auction")

		mockNFTRepo.AssertExpectations(t)
		mockAuctionRepo.AssertExpectations(t)
	})

	t.Run("Failed auction creation - invalid start time", func(t *testing.T) {
		req := &service.CreateAuctionRequest{
			NFTID:      nftID,
			StartPrice: "1000000000000000000",
			StartTime:  time.Now().Add(-time.Hour), // Past time
			EndTime:    time.Now().Add(24 * time.Hour),
		}

		nft := &models.NFT{
			ID:      nftID,
			OwnerID: sellerID,
			TokenID: "123",
			Title:   "Test NFT",
		}

		mockNFTRepo.On("GetByID", nftID).Return(nft, nil)
		mockAuctionRepo.On("HasActiveAuctionForNFT", nftID).Return(false, nil)

		auction, err := auctionService.CreateAuction(ctx, sellerID, req)

		assert.Error(t, err)
		assert.Nil(t, auction)
		assert.Contains(t, err.Error(), "start time must be in the future")

		mockNFTRepo.AssertExpectations(t)
		mockAuctionRepo.AssertExpectations(t)
	})

	t.Run("Failed auction creation - duration too short", func(t *testing.T) {
		req := &service.CreateAuctionRequest{
			NFTID:      nftID,
			StartPrice: "1000000000000000000",
			StartTime:  time.Now().Add(time.Hour),
			EndTime:    time.Now().Add(90 * time.Minute), // Only 30 minutes
		}

		nft := &models.NFT{
			ID:      nftID,
			OwnerID: sellerID,
			TokenID: "123",
			Title:   "Test NFT",
		}

		mockNFTRepo.On("GetByID", nftID).Return(nft, nil)
		mockAuctionRepo.On("HasActiveAuctionForNFT", nftID).Return(false, nil)

		auction, err := auctionService.CreateAuction(ctx, sellerID, req)

		assert.Error(t, err)
		assert.Nil(t, auction)
		assert.Contains(t, err.Error(), "auction duration must be at least 1 hour")

		mockNFTRepo.AssertExpectations(t)
		mockAuctionRepo.AssertExpectations(t)
	})
}

// TestBidPlacement tests bid placement logic
func TestBidPlacement(t *testing.T) {
	mockAuctionRepo := new(MockAuctionRepository)
	mockNFTRepo := new(MockNFTRepository)
	mockUserRepo := new(MockUserRepository)
	mockTransferRepo := new(MockTransferRepository)
	mockBlockchainSvc := new(MockBlockchainService)
	mockNotificationSvc := new(MockNotificationService)

	minimumIncrement := big.NewInt(10000000000000000) // 0.01 ETH in Wei

	config := &service.AuctionServiceConfig{
		BlockchainService:   mockBlockchainSvc,
		NotificationService: mockNotificationSvc,
		MinimumIncrement:    minimumIncrement,
	}

	auctionService := service.NewAuctionService(
		mockAuctionRepo,
		mockNFTRepo,
		mockUserRepo,
		mockTransferRepo,
		config,
	)

	ctx := context.Background()
	auctionID := uint(1)
	bidderID := uint(2)
	sellerID := uint(1)

	t.Run("Successful bid placement", func(t *testing.T) {
		req := &service.PlaceBidRequest{
			Amount: "2000000000000000000", // 2 ETH
		}

		nft := &models.NFT{
			ID:      1,
			TokenID: "123",
			Title:   "Test NFT",
		}

		auction := &models.Auction{
			ID:         auctionID,
			NFTID:      1,
			SellerID:   sellerID,
			StartPrice: big.NewInt(1000000000000000000), // 1 ETH
			CurrentBid: big.NewInt(1500000000000000000), // 1.5 ETH
			StartTime:  time.Now().Add(-time.Hour),
			EndTime:    time.Now().Add(time.Hour),
			Status:     "active",
			NFT:        *nft,
		}

		bidder := &models.User{
			ID:         bidderID,
			Username:   "bidder",
			WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
		}

		tx := &types.Transaction{
			Hash: "0x1234567890abcdef",
		}

		createdBid := &models.Bid{
			ID:        1,
			AuctionID: auctionID,
			BidderID:  bidderID,
			Amount:    big.NewInt(2000000000000000000),
			Status:    "confirmed",
			TxHash:    tx.Hash,
		}

		// Setup mocks
		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)
		mockUserRepo.On("GetByID", bidderID).Return(bidder, nil)
		mockBlockchainSvc.On("PlaceBid", ctx, nft.TokenID, bidder.WalletAddr, big.NewInt(2000000000000000000)).Return(tx, nil)
		mockAuctionRepo.On("GetDB").Return(nil)
		mockAuctionRepo.On("UpdateCurrentBid", auctionID, big.NewInt(2000000000000000000), bidderID).Return(nil)
		mockNotificationSvc.On("SendNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Mock bid repository
		mockBidRepo := new(MockBidRepository)
		mockBidRepo.On("Create", mock.AnythingOfType("*models.Bid")).Return(nil)
		mockBidRepo.On("GetByID", mock.AnythingOfType("uint")).Return(createdBid, nil)
		mockAuctionRepo.On("GetDB").Return(func() interface{} {
			return &MockBidRepository{} // This is a simplified mock for GetDB
		})

		response, err := auctionService.PlaceBid(ctx, auctionID, bidderID, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Bid)
		assert.True(t, response.IsHighest)
		assert.True(t, response.AuctionUpdated)

		mockAuctionRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockBlockchainSvc.AssertExpectations(t)
		mockNotificationSvc.AssertExpectations(t)
	})

	t.Run("Failed bid placement - auction not active", func(t *testing.T) {
		req := &service.PlaceBidRequest{
			Amount: "2000000000000000000",
		}

		auction := &models.Auction{
			ID:         auctionID,
			NFTID:      1,
			SellerID:   sellerID,
			StartPrice: big.NewInt(1000000000000000000),
			Status:     "pending", // Not active
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)

		response, err := auctionService.PlaceBid(ctx, auctionID, bidderID, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "auction is not active")

		mockAuctionRepo.AssertExpectations(t)
	})

	t.Run("Failed bid placement - auction ended", func(t *testing.T) {
		req := &service.PlaceBidRequest{
			Amount: "2000000000000000000",
		}

		auction := &models.Auction{
			ID:         auctionID,
			NFTID:      1,
			SellerID:   sellerID,
			StartPrice: big.NewInt(1000000000000000000),
			EndTime:    time.Now().Add(-time.Hour), // Ended
			Status:     "active",
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)

		response, err := auctionService.PlaceBid(ctx, auctionID, bidderID, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "auction has ended")

		mockAuctionRepo.AssertExpectations(t)
	})

	t.Run("Failed bid placement - seller bidding", func(t *testing.T) {
		req := &service.PlaceBidRequest{
			Amount: "2000000000000000000",
		}

		auction := &models.Auction{
			ID:         auctionID,
			NFTID:      1,
			SellerID:   sellerID,
			StartPrice: big.NewInt(1000000000000000000),
			StartTime:  time.Now().Add(-time.Hour),
			EndTime:    time.Now().Add(time.Hour),
			Status:     "active",
		}

		seller := &models.User{
			ID: sellerID,
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)
		mockUserRepo.On("GetByID", sellerID).Return(seller, nil)

		response, err := auctionService.PlaceBid(ctx, auctionID, sellerID, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "seller cannot bid on their own auction")

		mockAuctionRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failed bid placement - bid too low", func(t *testing.T) {
		req := &service.PlaceBidRequest{
			Amount: "1000000000000000000", // 1 ETH (same as start price)
		}

		auction := &models.Auction{
			ID:         auctionID,
			NFTID:      1,
			SellerID:   sellerID,
			StartPrice: big.NewInt(1000000000000000000), // 1 ETH
			CurrentBid: big.NewInt(0),
			StartTime:  time.Now().Add(-time.Hour),
			EndTime:    time.Now().Add(time.Hour),
			Status:     "active",
		}

		bidder := &models.User{
			ID: bidderID,
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)
		mockUserRepo.On("GetByID", bidderID).Return(bidder, nil)

		response, err := auctionService.PlaceBid(ctx, auctionID, bidderID, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "bid must be at least")

		mockAuctionRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

// TestAuctionHelpers tests auction helper methods
func TestAuctionHelpers(t *testing.T) {
	t.Run("GetStartPriceWei", func(t *testing.T) {
		auction := &models.Auction{
			StartPrice: big.NewInt(1000000000000000000),
		}

		result := auction.GetStartPriceWei()
		assert.Equal(t, "1000000000000000000", result)

		auction.StartPrice = nil
		result = auction.GetStartPriceWei()
		assert.Equal(t, "0", result)
	})

	t.Run("SetStartPriceWei", func(t *testing.T) {
		auction := &models.Auction{}

		err := auction.SetStartPriceWei("1000000000000000000")
		assert.NoError(t, err)
		assert.Equal(t, int64(1000000000000000000), auction.StartPrice.Int64())

		err = auction.SetStartPriceWei("invalid")
		assert.Error(t, err)

		err = auction.SetStartPriceWei("-100")
		assert.Error(t, err)
	})

	t.Run("IsActive", func(t *testing.T) {
		now := time.Now()

		tests := []struct {
			name     string
			auction  models.Auction
			expected bool
		}{
			{
				name: "Active auction",
				auction: models.Auction{
					Status:    "active",
					StartTime: now.Add(-time.Hour),
					EndTime:   now.Add(time.Hour),
				},
				expected: true,
			},
			{
				name: "Not started yet",
				auction: models.Auction{
					Status:    "active",
					StartTime: now.Add(time.Hour),
					EndTime:   now.Add(2 * time.Hour),
				},
				expected: false,
			},
			{
				name: "Already ended",
				auction: models.Auction{
					Status:    "active",
					StartTime: now.Add(-2 * time.Hour),
					EndTime:   now.Add(-time.Hour),
				},
				expected: false,
			},
			{
				name: "Wrong status",
				auction: models.Auction{
					Status:    "pending",
					StartTime: now.Add(-time.Hour),
					EndTime:   now.Add(time.Hour),
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.auction.IsActive()
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("HasReservePrice", func(t *testing.T) {
		auction := &models.Auction{
			ReservePrice: big.NewInt(1000000000000000000),
		}
		assert.True(t, auction.HasReservePrice())

		auction.ReservePrice = big.NewInt(0)
		assert.False(t, auction.HasReservePrice())

		auction.ReservePrice = nil
		assert.False(t, auction.HasReservePrice())
	})

	t.Run("IsReserveMet", func(t *testing.T) {
		tests := []struct {
			name     string
			auction  models.Auction
			expected bool
		}{
			{
				name: "No reserve price",
				auction: models.Auction{
					ReservePrice: nil,
					CurrentBid:   big.NewInt(1000000000000000000),
				},
				expected: true,
			},
			{
				name: "Reserve price met",
				auction: models.Auction{
					ReservePrice: big.NewInt(1000000000000000000),
					CurrentBid:   big.NewInt(1500000000000000000),
				},
				expected: true,
			},
			{
				name: "Reserve price not met",
				auction: models.Auction{
					ReservePrice: big.NewInt(2000000000000000000),
					CurrentBid:   big.NewInt(1500000000000000000),
				},
				expected: false,
			},
			{
				name: "No current bid",
				auction: models.Auction{
					ReservePrice: big.NewInt(1000000000000000000),
					CurrentBid:   big.NewInt(0),
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.auction.IsReserveMet()
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// TestAuctionCancellation tests auction cancellation logic
func TestAuctionCancellation(t *testing.T) {
	mockAuctionRepo := new(MockAuctionRepository)
	mockNFTRepo := new(MockNFTRepository)
	mockUserRepo := new(MockUserRepository)
	mockTransferRepo := new(MockTransferRepository)
	mockNotificationSvc := new(MockNotificationService)

	config := &service.AuctionServiceConfig{
		NotificationService: mockNotificationSvc,
	}

	auctionService := service.NewAuctionService(
		mockAuctionRepo,
		mockNFTRepo,
		mockUserRepo,
		mockTransferRepo,
		config,
	)

	ctx := context.Background()
	auctionID := uint(1)
	sellerID := uint(1)

	t.Run("Successful auction cancellation", func(t *testing.T) {
		nft := &models.NFT{
			ID:      1,
			Title:   "Test NFT",
			TokenID: "123",
		}

		auction := &models.Auction{
			ID:       auctionID,
			NFTID:    1,
			SellerID: sellerID,
			Status:   "pending",
			Bids:     []models.Bid{}, // No bids
			NFT:      *nft,
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)
		mockAuctionRepo.On("UpdateStatus", auctionID, "cancelled").Return(nil)
		mockNotificationSvc.On("SendNotification", ctx, sellerID, "auction_cancelled",
			"Auction Cancelled", mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil)

		err := auctionService.CancelAuction(ctx, auctionID, sellerID)

		assert.NoError(t, err)

		mockAuctionRepo.AssertExpectations(t)
		mockNotificationSvc.AssertExpectations(t)
	})

	t.Run("Failed cancellation - not authorized", func(t *testing.T) {
		auction := &models.Auction{
			ID:       auctionID,
			SellerID: sellerID + 1, // Different seller
			Status:   "pending",
			Bids:     []models.Bid{},
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)

		err := auctionService.CancelAuction(ctx, auctionID, sellerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authorized to cancel this auction")

		mockAuctionRepo.AssertExpectations(t)
	})

	t.Run("Failed cancellation - auction ended", func(t *testing.T) {
		auction := &models.Auction{
			ID:       auctionID,
			SellerID: sellerID,
			Status:   "ended",
			Bids:     []models.Bid{},
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)

		err := auctionService.CancelAuction(ctx, auctionID, sellerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auction cannot be cancelled")

		mockAuctionRepo.AssertExpectations(t)
	})

	t.Run("Failed cancellation - has confirmed bids", func(t *testing.T) {
		auction := &models.Auction{
			ID:       auctionID,
			SellerID: sellerID,
			Status:   "active",
			Bids: []models.Bid{
				{
					Status: "confirmed",
				},
			},
		}

		mockAuctionRepo.On("GetWithBids", auctionID).Return(auction, nil)

		err := auctionService.CancelAuction(ctx, auctionID, sellerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot cancel auction with confirmed bids")

		mockAuctionRepo.AssertExpectations(t)
	})
}

