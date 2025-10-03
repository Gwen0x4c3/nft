package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"nft-platform/internal/handlers"
	"nft-platform/internal/models"
	"nft-platform/internal/repository"
	"nft-platform/internal/service"
)

// Mock performance test dependencies
type MockPerfAuctionRepository struct {
	mock.Mock
}

func (m *MockPerfAuctionRepository) Create(auction *models.Auction) error {
	args := m.Called(auction)
	return args.Error(0)
}

func (m *MockPerfAuctionRepository) GetByID(id uint) (*models.Auction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Auction), args.Error(1)
}

func (m *MockPerfAuctionRepository) GetWithBids(id uint) (*models.Auction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Auction), args.Error(1)
}

func (m *MockPerfAuctionRepository) Update(auction *models.Auction) error {
	args := m.Called(auction)
	return args.Error(0)
}

func (m *MockPerfAuctionRepository) UpdateStatus(id uint, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockPerfAuctionRepository) UpdateCurrentBid(id uint, amount *big.Int, bidderID uint) error {
	args := m.Called(id, amount, bidderID)
	return args.Error(0)
}

func (m *MockPerfAuctionRepository) List(filter *repository.AuctionFilter, offset, limit int) ([]models.Auction, error) {
	args := m.Called(filter, offset, limit)
	return args.Get(0).([]models.Auction), args.Error(1)
}

func (m *MockPerfAuctionRepository) Count(filter *repository.AuctionFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPerfAuctionRepository) GetActiveAuctions(offset, limit int) ([]models.Auction, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]models.Auction), args.Error(1)
}

func (m *MockPerfAuctionRepository) HasActiveAuctionForNFT(nftID uint) (bool, error) {
	args := m.Called(nftID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPerfAuctionRepository) GetExpiredAuctions(limit int) ([]models.Auction, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Auction), args.Error(1)
}

func (m *MockPerfAuctionRepository) GetDB() interface{} {
	args := m.Called()
	return args.Get(0)
}

type MockPerfNFTRepository struct {
	mock.Mock
}

func (m *MockPerfNFTRepository) Create(nft *models.NFT) error {
	args := m.Called(nft)
	return args.Error(0)
}

func (m *MockPerfNFTRepository) GetByID(id uint) (*models.NFT, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NFT), args.Error(1)
}

func (m *MockPerfNFTRepository) Update(nft *models.NFT) error {
	args := m.Called(nft)
	return args.Error(0)
}

func (m *MockPerfNFTRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPerfNFTRepository) List(filter *repository.NFTFilter, offset, limit int) ([]models.NFT, error) {
	args := m.Called(filter, offset, limit)
	return args.Get(0).([]models.NFT), args.Error(1)
}

func (m *MockPerfNFTRepository) Count(filter *repository.NFTFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPerfNFTRepository) GetByOwnerID(ownerID uint, offset, limit int) ([]models.NFT, error) {
	args := m.Called(ownerID, offset, limit)
	return args.Get(0).([]models.NFT), args.Error(1)
}

func (m *MockPerfNFTRepository) GetByCreatorID(creatorID uint, offset, limit int) ([]models.NFT, error) {
	args := m.Called(creatorID, offset, limit)
	return args.Get(0).([]models.NFT), args.Error(1)
}

type MockPerfUserRepository struct {
	mock.Mock
}

func (m *MockPerfUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockPerfUserRepository) GetByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockPerfUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockPerfUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockPerfUserRepository) GetByWalletAddress(address string) (*models.User, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockPerfUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockPerfUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPerfUserRepository) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockPerfUserRepository) ExistsByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockPerfUserRepository) ExistsByWalletAddress(address string) (bool, error) {
	args := m.Called(address)
	return args.Bool(0), args.Error(1)
}

func (m *MockPerfUserRepository) UpdateVerificationStatus(id uint, verified bool) error {
	args := m.Called(id, verified)
	return args.Error(0)
}

type MockPerfTransferRepository struct {
	mock.Mock
}

func (m *MockPerfTransferRepository) Create(transfer *models.Transfer) error {
	args := m.Called(transfer)
	return args.Error(0)
}

func (m *MockPerfTransferRepository) GetByID(id uint) (*models.Transfer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transfer), args.Error(1)
}

func (m *MockPerfTransferRepository) GetByNFTID(nftID uint, offset, limit int) ([]models.Transfer, error) {
	args := m.Called(nftID, offset, limit)
	return args.Get(0).([]models.Transfer), args.Error(1)
}

func (m *MockPerfTransferRepository) GetByFromID(fromID uint, offset, limit int) ([]models.Transfer, error) {
	args := m.Called(fromID, offset, limit)
	return args.Get(0).([]models.Transfer), args.Error(1)
}

func (m *MockPerfTransferRepository) GetByToID(toID uint, offset, limit int) ([]models.Transfer, error) {
	args := m.Called(toID, offset, limit)
	return args.Get(0).([]models.Transfer), args.Error(1)
}

// AuctionPerformanceTestSuite contains performance tests for auction endpoints
type AuctionPerformanceTestSuite struct {
	suite.Suite
	router           *gin.Engine
	auctionHandler   *handlers.AuctionHandler
	bidHandler       *handlers.BidHandler
	mockAuctionRepo  *MockPerfAuctionRepository
	mockNFTRepo      *MockPerfNFTRepository
	mockUserRepo     *MockPerfUserRepository
	mockTransferRepo *MockPerfTransferRepository
}

func (suite *AuctionPerformanceTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Create mocks
	suite.mockAuctionRepo = new(MockPerfAuctionRepository)
	suite.mockNFTRepo = new(MockPerfNFTRepository)
	suite.mockUserRepo = new(MockPerfUserRepository)
	suite.mockTransferRepo = new(MockPerfTransferRepository)

	// Create auction service
	minimumIncrement := big.NewInt(10000000000000000) // 0.01 ETH
	config := &service.AuctionServiceConfig{
		MinimumIncrement: minimumIncrement,
	}

	auctionService := service.NewAuctionService(
		suite.mockAuctionRepo,
		suite.mockNFTRepo,
		suite.mockUserRepo,
		suite.mockTransferRepo,
		config,
	)

	// Create handlers
	suite.auctionHandler = handlers.NewAuctionHandler(auctionService)
	suite.bidHandler = handlers.NewBidHandler(auctionService)

	// Setup routes
	v1 := suite.router.Group("/api/v1")
	{
		auctions := v1.Group("/auctions")
		{
			auctions.GET("", suite.auctionHandler.ListAuctions)
			auctions.POST("", suite.auctionHandler.CreateAuction)
			auctions.GET("/:auctionId", suite.auctionHandler.GetAuction)
			auctions.DELETE("/:auctionId", suite.auctionHandler.CancelAuction)
			auctions.GET("/:auctionId/bids", suite.bidHandler.GetAuctionBids)
			auctions.POST("/:auctionId/bids", suite.bidHandler.PlaceBid)
		}
	}
}

// Helper method to measure response time
func (suite *AuctionPerformanceTestSuite) measureResponseTime(req *http.Request) (*httptest.ResponseRecorder, time.Duration) {
	start := time.Now()
	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	duration := time.Since(start)
	return resp, duration
}

// Helper method to create authenticated request
func (suite *AuctionPerformanceTestSuite) createAuthenticatedRequest(method, path string, body io.Reader, userID uint) *http.Request {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")

	// Mock authentication by setting user_id in context
	// In real implementation, this would be done by middleware
	if userID > 0 {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("user_id", userID)
		req = req.WithContext(c.Request.Context())
	}

	return req
}

// TestListAuctionsPerformance tests GET /auctions endpoint performance
func (suite *AuctionPerformanceTestSuite) TestListAuctionsPerformance() {
	// Create mock data
	mockAuctions := make([]models.Auction, 100)
	for i := 0; i < 100; i++ {
		mockAuctions[i] = models.Auction{
			ID:         uint(i + 1),
			NFTID:      uint(i + 1),
			SellerID:   1,
			StartPrice: big.NewInt(int64(i+1) * 1000000000000000000), // i+1 ETH
			Status:     "active",
			StartTime:  time.Now().Add(-time.Hour),
			EndTime:    time.Now().Add(time.Hour),
		}
	}

	// Setup mock expectations
	suite.mockAuctionRepo.On("List", mock.Anything, 0, 20).Return(mockAuctions[:20], nil)
	suite.mockAuctionRepo.On("Count", mock.Anything).Return(int64(100), nil)

	// Test performance
	req := httptest.NewRequest("GET", "/api/v1/auctions?page=1&limit=20", nil)
	resp, duration := suite.measureResponseTime(req)

	// Assertions
	suite.Equal(http.StatusOK, resp.Code)
	suite.Less(duration, 200*time.Millisecond, "ListAuctions should respond in <200ms")

	fmt.Printf("ListAuctions response time: %v\n", duration)
}

// TestGetAuctionPerformance tests GET /auctions/{id} endpoint performance
func (suite *AuctionPerformanceTestSuite) TestGetAuctionPerformance() {
	// Create mock auction
	mockAuction := &models.Auction{
		ID:         1,
		NFTID:      1,
		SellerID:   1,
		StartPrice: big.NewInt(1000000000000000000), // 1 ETH
		CurrentBid: big.NewInt(1500000000000000000), // 1.5 ETH
		Status:     "active",
		StartTime:  time.Now().Add(-time.Hour),
		EndTime:    time.Now().Add(time.Hour),
		NFT: models.NFT{
			ID:      1,
			Title:   "Test NFT",
			TokenID: "123",
		},
		Seller: models.User{
			ID:       1,
			Username: "seller",
		},
	}

	// Setup mock expectations
	suite.mockAuctionRepo.On("GetByID", uint(1)).Return(mockAuction, nil)

	// Test performance
	req := httptest.NewRequest("GET", "/api/v1/auctions/1", nil)
	resp, duration := suite.measureResponseTime(req)

	// Assertions
	suite.Equal(http.StatusOK, resp.Code)
	suite.Less(duration, 200*time.Millisecond, "GetAuction should respond in <200ms")

	fmt.Printf("GetAuction response time: %v\n", duration)
}

// TestCreateAuctionPerformance tests POST /auctions endpoint performance
func (suite *AuctionPerformanceTestSuite) TestCreateAuctionPerformance() {
	// Create request body
	reqBody := service.CreateAuctionRequest{
		NFTID:        1,
		StartPrice:   "1000000000000000000", // 1 ETH
		ReservePrice: "1500000000000000000", // 1.5 ETH
		StartTime:    time.Now().Add(time.Hour),
		EndTime:      time.Now().Add(25 * time.Hour),
	}

	body, _ := json.Marshal(reqBody)

	// Create mock NFT
	mockNFT := &models.NFT{
		ID:        1,
		OwnerID:   1,
		TokenID:   "123",
		Title:     "Test NFT",
		IsForSale: true,
	}

	// Create mock auction
	mockAuction := &models.Auction{
		ID:           1,
		NFTID:        1,
		SellerID:     1,
		StartPrice:   big.NewInt(1000000000000000000),
		ReservePrice: big.NewInt(1500000000000000000),
		CurrentBid:   big.NewInt(0),
		Status:       "pending",
		StartTime:    reqBody.StartTime,
		EndTime:      reqBody.EndTime,
	}

	// Setup mock expectations
	suite.mockNFTRepo.On("GetByID", uint(1)).Return(mockNFT, nil)
	suite.mockAuctionRepo.On("HasActiveAuctionForNFT", uint(1)).Return(false, nil)
	suite.mockAuctionRepo.On("Create", mock.AnythingOfType("*models.Auction")).Return(nil)
	suite.mockNFTRepo.On("Update", mock.AnythingOfType("*models.NFT")).Return(nil)
	suite.mockAuctionRepo.On("GetByID", mock.AnythingOfType("uint")).Return(mockAuction, nil)

	// Test performance
	req := suite.createAuthenticatedRequest("POST", "/api/v1/auctions", bytes.NewBuffer(body), 1)
	resp, duration := suite.measureResponseTime(req)

	// Assertions
	suite.Equal(http.StatusCreated, resp.Code)
	suite.Less(duration, 200*time.Millisecond, "CreateAuction should respond in <200ms")

	fmt.Printf("CreateAuction response time: %v\n", duration)
}

// TestPlaceBidPerformance tests POST /auctions/{id}/bids endpoint performance
func (suite *AuctionPerformanceTestSuite) TestPlaceBidPerformance() {
	// Create request body
	reqBody := service.PlaceBidRequest{
		Amount: "2000000000000000000", // 2 ETH
	}

	body, _ := json.Marshal(reqBody)

	// Create mock auction
	mockAuction := &models.Auction{
		ID:         1,
		NFTID:      1,
		SellerID:   1,
		StartPrice: big.NewInt(1000000000000000000), // 1 ETH
		CurrentBid: big.NewInt(1500000000000000000), // 1.5 ETH
		Status:     "active",
		StartTime:  time.Now().Add(-time.Hour),
		EndTime:    time.Now().Add(time.Hour),
		NFT: models.NFT{
			ID:      1,
			Title:   "Test NFT",
			TokenID: "123",
		},
	}

	// Create mock bidder
	mockBidder := &models.User{
		ID:         2,
		Username:   "bidder",
		WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
	}

	// Create mock bid response
	mockBidResponse := &service.BidResponse{
		Bid: &models.Bid{
			ID:        1,
			AuctionID: 1,
			BidderID:  2,
			Amount:    big.NewInt(2000000000000000000),
			Status:    "confirmed",
			TxHash:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		IsHighest:       true,
		PreviousHighBid: "1500000000000000000",
		AuctionUpdated:  true,
	}

	// Setup mock expectations
	suite.mockAuctionRepo.On("GetWithBids", uint(1)).Return(mockAuction, nil)
	suite.mockUserRepo.On("GetByID", uint(2)).Return(mockBidder, nil)
	suite.mockAuctionRepo.On("GetDB").Return(&MockPerfBidRepository{})
	suite.mockAuctionRepo.On("UpdateCurrentBid", uint(1), big.NewInt(2000000000000000000), uint(2)).Return(nil)

	// Test performance
	req := suite.createAuthenticatedRequest("POST", "/api/v1/auctions/1/bids", bytes.NewBuffer(body), 2)
	resp, duration := suite.measureResponseTime(req)

	// Assertions
	suite.Equal(http.StatusCreated, resp.Code)
	suite.Less(duration, 200*time.Millisecond, "PlaceBid should respond in <200ms")

	fmt.Printf("PlaceBid response time: %v\n", duration)
}

// TestGetAuctionBidsPerformance tests GET /auctions/{id}/bids endpoint performance
func (suite *AuctionPerformanceTestSuite) TestGetAuctionBidsPerformance() {
	// Create mock bids
	mockBids := make([]models.Bid, 50)
	for i := 0; i < 50; i++ {
		mockBids[i] = models.Bid{
			ID:        uint(i + 1),
			AuctionID: 1,
			BidderID:  uint(i%5 + 2),                               // 5 different bidders
			Amount:    big.NewInt(int64(i+2) * 100000000000000000), // Bids from 0.2 to 5.2 ETH
			Status:    "confirmed",
			TxHash:    fmt.Sprintf("0x%d%064d", i, i),
		}
	}

	// Create mock auction
	mockAuction := &models.Auction{
		ID:         1,
		NFTID:      1,
		SellerID:   1,
		StartPrice: big.NewInt(1000000000000000000),
		Status:     "active",
	}

	// Setup mock expectations
	suite.mockAuctionRepo.On("GetByID", uint(1)).Return(mockAuction, nil)

	// Mock bid repository
	mockBidRepo := &MockPerfBidRepository{}
	mockBidRepo.On("GetByAuctionID", uint(1), 0, 50).Return(mockBids, nil)
	suite.mockAuctionRepo.On("GetDB").Return(mockBidRepo)

	// Test performance
	req := httptest.NewRequest("GET", "/api/v1/auctions/1/bids?page=1&limit=50", nil)
	resp, duration := suite.measureResponseTime(req)

	// Assertions
	suite.Equal(http.StatusOK, resp.Code)
	suite.Less(duration, 200*time.Millisecond, "GetAuctionBids should respond in <200ms")

	fmt.Printf("GetAuctionBids response time: %v\n", duration)
}

// TestCancelAuctionPerformance tests DELETE /auctions/{id} endpoint performance
func (suite *AuctionPerformanceTestSuite) TestCancelAuctionPerformance() {
	// Create mock auction
	mockAuction := &models.Auction{
		ID:       1,
		NFTID:    1,
		SellerID: 1,
		Status:   "pending",
		Bids:     []models.Bid{}, // No bids
		NFT: models.NFT{
			ID:      1,
			Title:   "Test NFT",
			TokenID: "123",
		},
	}

	// Setup mock expectations
	suite.mockAuctionRepo.On("GetWithBids", uint(1)).Return(mockAuction, nil)
	suite.mockAuctionRepo.On("UpdateStatus", uint(1), "cancelled").Return(nil)

	// Test performance
	req := suite.createAuthenticatedRequest("DELETE", "/api/v1/auctions/1", nil, 1)
	resp, duration := suite.measureResponseTime(req)

	// Assertions
	suite.Equal(http.StatusOK, resp.Code)
	suite.Less(duration, 200*time.Millisecond, "CancelAuction should respond in <200ms")

	fmt.Printf("CancelAuction response time: %v\n", duration)
}

// TestConcurrentRequests tests performance under concurrent load
func (suite *AuctionPerformanceTestSuite) TestConcurrentRequests() {
	// Create mock auction
	mockAuction := &models.Auction{
		ID:         1,
		NFTID:      1,
		SellerID:   1,
		StartPrice: big.NewInt(1000000000000000000),
		CurrentBid: big.NewInt(1500000000000000000),
		Status:     "active",
		StartTime:  time.Now().Add(-time.Hour),
		EndTime:    time.Now().Add(time.Hour),
		NFT: models.NFT{
			ID:      1,
			Title:   "Test NFT",
			TokenID: "123",
		},
		Seller: models.User{
			ID:       1,
			Username: "seller",
		},
	}

	// Setup mock expectations for concurrent access
	suite.mockAuctionRepo.On("GetByID", uint(1)).Return(mockAuction, nil)

	// Test concurrent GET requests
	concurrency := 50
	results := make(chan time.Duration, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/v1/auctions/1", nil)
			_, duration := suite.measureResponseTime(req)
			results <- duration
		}()
	}

	// Collect results
	var maxDuration time.Duration
	var totalDuration time.Duration
	for i := 0; i < concurrency; i++ {
		duration := <-results
		totalDuration += duration
		if duration > maxDuration {
			maxDuration = duration
		}
	}

	avgDuration := totalDuration / time.Duration(concurrency)

	fmt.Printf("Concurrent GET requests (%d):\n", concurrency)
	fmt.Printf("  Average response time: %v\n", avgDuration)
	fmt.Printf("  Max response time: %v\n", maxDuration)

	// Assertions
	suite.Less(maxDuration, 200*time.Millisecond, "Max response time should be <200ms under load")
	suite.Less(avgDuration, 150*time.Millisecond, "Average response time should be <150ms under load")
}

// TestLargeDatasetPerformance tests performance with large datasets
func (suite *AuctionPerformanceTestSuite) TestLargeDatasetPerformance() {
	// Create large mock dataset (1000 auctions)
	mockAuctions := make([]models.Auction, 1000)
	for i := 0; i < 1000; i++ {
		mockAuctions[i] = models.Auction{
			ID:         uint(i + 1),
			NFTID:      uint(i + 1),
			SellerID:   uint(i%10 + 1), // 10 different sellers
			StartPrice: big.NewInt(int64(i+1) * 1000000000000000000),
			Status:     []string{"active", "ended", "pending"}[i%3],
			StartTime:  time.Now().Add(-time.Hour * time.Duration(i%24)),
			EndTime:    time.Now().Add(time.Hour * time.Duration(i%24)),
		}
	}

	// Setup mock expectations
	suite.mockAuctionRepo.On("List", mock.Anything, 0, 100).Return(mockAuctions[:100], nil)
	suite.mockAuctionRepo.On("Count", mock.Anything).Return(int64(1000), nil)

	// Test performance with large dataset
	req := httptest.NewRequest("GET", "/api/v1/auctions?page=1&limit=100", nil)
	resp, duration := suite.measureResponseTime(req)

	// Assertions
	suite.Equal(http.StatusOK, resp.Code)
	suite.Less(duration, 200*time.Millisecond, "ListAuctions with large dataset should respond in <200ms")

	fmt.Printf("ListAuctions (1000 records) response time: %v\n", duration)
}

// MockPerfBidRepository is a mock for bid repository
type MockPerfBidRepository struct {
	mock.Mock
}

func (m *MockPerfBidRepository) Create(bid *models.Bid) error {
	args := m.Called(bid)
	return args.Error(0)
}

func (m *MockPerfBidRepository) GetByID(id uint) (*models.Bid, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bid), args.Error(1)
}

func (m *MockPerfBidRepository) GetByAuctionID(auctionID uint, offset, limit int) ([]models.Bid, error) {
	args := m.Called(auctionID, offset, limit)
	return args.Get(0).([]models.Bid), args.Error(1)
}

func (m *MockPerfBidRepository) GetByBidderID(bidderID uint, offset, limit int) ([]models.Bid, error) {
	args := m.Called(bidderID, offset, limit)
	return args.Get(0).([]models.Bid), args.Error(1)
}

func (m *MockPerfBidRepository) GetHighestBid(auctionID uint) (*models.Bid, error) {
	args := m.Called(auctionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bid), args.Error(1)
}

// TestRunner runs the performance test suite
func TestAuctionPerformanceSuite(t *testing.T) {
	suite.Run(t, new(AuctionPerformanceTestSuite))
}

// BenchmarkListAuctions benchmarks the ListAuctions endpoint
func BenchmarkListAuctions(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockAuctionRepo := new(MockPerfAuctionRepository)
	mockNFTRepo := new(MockPerfNFTRepository)
	mockUserRepo := new(MockPerfUserRepository)
	mockTransferRepo := new(MockPerfTransferRepository)

	minimumIncrement := big.NewInt(10000000000000000)
	config := &service.AuctionServiceConfig{
		MinimumIncrement: minimumIncrement,
	}

	auctionService := service.NewAuctionService(
		mockAuctionRepo,
		mockNFTRepo,
		mockUserRepo,
		mockTransferRepo,
		config,
	)

	auctionHandler := handlers.NewAuctionHandler(auctionService)

	router := gin.New()
	router.GET("/api/v1/auctions", auctionHandler.ListAuctions)

	// Setup mock data
	mockAuctions := make([]models.Auction, 20)
	for i := 0; i < 20; i++ {
		mockAuctions[i] = models.Auction{
			ID:         uint(i + 1),
			NFTID:      uint(i + 1),
			SellerID:   1,
			StartPrice: big.NewInt(int64(i+1) * 1000000000000000000),
			Status:     "active",
		}
	}

	mockAuctionRepo.On("List", mock.Anything, 0, 20).Return(mockAuctions, nil)
	mockAuctionRepo.On("Count", mock.Anything).Return(int64(100), nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/auctions?page=1&limit=20", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", resp.Code)
		}
	}
}

// BenchmarkGetAuction benchmarks the GetAuction endpoint
func BenchmarkGetAuction(b *testing.B) {
	gin.SetMode(gin.TestMode)

	mockAuctionRepo := new(MockPerfAuctionRepository)
	mockNFTRepo := new(MockPerfNFTRepository)
	mockUserRepo := new(MockPerfUserRepository)
	mockTransferRepo := new(MockPerfTransferRepository)

	minimumIncrement := big.NewInt(10000000000000000)
	config := &service.AuctionServiceConfig{
		MinimumIncrement: minimumIncrement,
	}

	auctionService := service.NewAuctionService(
		mockAuctionRepo,
		mockNFTRepo,
		mockUserRepo,
		mockTransferRepo,
		config,
	)

	auctionHandler := handlers.NewAuctionHandler(auctionService)

	router := gin.New()
	router.GET("/api/v1/auctions/:auctionId", auctionHandler.GetAuction)

	// Setup mock data
	mockAuction := &models.Auction{
		ID:         1,
		NFTID:      1,
		SellerID:   1,
		StartPrice: big.NewInt(1000000000000000000),
		Status:     "active",
	}

	mockAuctionRepo.On("GetByID", uint(1)).Return(mockAuction, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/auctions/1", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", resp.Code)
		}
	}
}

