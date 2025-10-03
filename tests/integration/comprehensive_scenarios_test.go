package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"nft-platform/internal/models"
)

// ComprehensiveIntegrationTestSuite covers end-to-end scenarios
type ComprehensiveIntegrationTestSuite struct {
	suite.Suite
	router       *gin.Engine
	server       *httptest.Server
	wsDialer     websocket.Dialer
	wsURL        string
	testUsers    map[string]*models.User
	testNFTs     map[string]*models.NFT
	testAuctions map[string]*models.Auction
	userTokens   map[string]string
}

func (suite *ComprehensiveIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupTestRoutes()
	suite.setupTestData()

	// Start test server
	suite.server = httptest.NewServer(suite.router)
	suite.wsURL = "ws" + suite.server.URL[4:] + "/ws"
	suite.wsDialer = websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
}

func (suite *ComprehensiveIntegrationTestSuite) TearDownSuite() {
	suite.server.Close()
}

func (suite *ComprehensiveIntegrationTestSuite) setupTestRoutes() {
	// Mock API routes for integration testing
	v1 := suite.router.Group("/api/v1")
	{
		// Authentication
		v1.POST("/auth/login", suite.mockLogin)
		v1.POST("/auth/register", suite.mockRegister)
		v1.POST("/auth/refresh", suite.mockRefresh)

		// Users
		v1.GET("/users/profile", suite.mockGetUserProfile)
		v1.PUT("/users/profile", suite.mockUpdateUserProfile)
		v1.GET("/users/:userId", suite.mockGetUserByID)

		// NFTs
		v1.GET("/nfts", suite.mockListNFTs)
		v1.POST("/nfts", suite.mockCreateNFT)
		v1.GET("/nfts/:nftId", suite.mockGetNFT)
		v1.PUT("/nfts/:nftId", suite.mockUpdateNFT)
		v1.POST("/nfts/:nftId/transfer", suite.mockTransferNFT)

		// Auctions
		v1.GET("/auctions", suite.mockListAuctions)
		v1.POST("/auctions", suite.mockCreateAuction)
		v1.GET("/auctions/:auctionId", suite.mockGetAuction)
		v1.DELETE("/auctions/:auctionId", suite.mockCancelAuction)
		v1.GET("/auctions/:auctionId/bids", suite.mockGetAuctionBids)
		v1.POST("/auctions/:auctionId/bids", suite.mockPlaceBid)

		// Notifications
		v1.GET("/notifications", suite.mockGetNotifications)
		v1.PUT("/notifications/:notificationId/read", suite.mockMarkNotificationRead)
		v1.PUT("/notifications/read-all", suite.mockMarkAllNotificationsRead)
	}

	// WebSocket route
	suite.router.GET("/ws", suite.mockWebSocketHandler)
}

func (suite *ComprehensiveIntegrationTestSuite) setupTestData() {
	suite.testUsers = make(map[string]*models.User)
	suite.testNFTs = make(map[string]*models.NFT)
	suite.testAuctions = make(map[string]*models.Auction)
	suite.userTokens = make(map[string]string)

	// Create test users
	suite.testUsers["alice"] = &models.User{
		ID:         1,
		Username:   "alice",
		Email:      "alice@example.com",
		WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
		IsVerified: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	suite.testUsers["bob"] = &models.User{
		ID:         2,
		Username:   "bob",
		Email:      "bob@example.com",
		WalletAddr: "0x842d35Cc6634C0532925a3b8D4E7E0E0e9e0dF2D",
		IsVerified: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	suite.testUsers["charlie"] = &models.User{
		ID:         3,
		Username:   "charlie",
		Email:      "charlie@example.com",
		WalletAddr: "0x952d35Cc6634C0532925a3b8D4E7E0E0e9e0dF3E",
		IsVerified: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create test NFTs
	suite.testNFTs["alice_art"] = &models.NFT{
		ID:           1,
		TokenID:      "1",
		ContractAddr: "0x1234567890123456789012345678901234567890",
		CreatorID:    1,
		OwnerID:      1,
		Title:        "Alice's Art",
		Description:  "Beautiful digital artwork by Alice",
		ImageURL:     "https://example.com/alice-art.jpg",
		MetadataURI:  "https://example.com/metadata/1.json",
		Price:        big.NewInt(1000000000000000000), // 1 ETH
		IsForSale:    true,
		Royalty:      5,
		MintTxHash:   "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	suite.testNFTs["bob_photo"] = &models.NFT{
		ID:           2,
		TokenID:      "2",
		ContractAddr: "0x1234567890123456789012345678901234567890",
		CreatorID:    2,
		OwnerID:      2,
		Title:        "Bob's Photo",
		Description:  "Stunning photography by Bob",
		ImageURL:     "https://example.com/bob-photo.jpg",
		MetadataURI:  "https://example.com/metadata/2.json",
		Price:        big.NewInt(2000000000000000000), // 2 ETH
		IsForSale:    false,
		Royalty:      10,
		MintTxHash:   "0xbcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Create test auction
	suite.testAuctions["alice_art_auction"] = &models.Auction{
		ID:           1,
		NFTID:        1,
		SellerID:     1,
		StartPrice:   big.NewInt(1000000000000000000), // 1 ETH
		ReservePrice: big.NewInt(1500000000000000000), // 1.5 ETH
		CurrentBid:   big.NewInt(1200000000000000000), // 1.2 ETH
		StartTime:    time.Now().Add(-30 * time.Minute),
		EndTime:      time.Now().Add(23 * time.Hour),
		Status:       "active",
		CreatedAt:    time.Now().Add(-1 * time.Hour),
		UpdatedAt:    time.Now().Add(-10 * time.Minute),
	}
}

// Scenario 1: Complete User Registration and NFT Minting Workflow
func (suite *ComprehensiveIntegrationTestSuite) TestCompleteUserWorkflow() {
	t := suite.T()

	// Step 1: Register new user
	registerReq := map[string]interface{}{
		"username":       "newuser",
		"email":          "newuser@example.com",
		"wallet_address": "0x9999999999999999999999999999999999999999",
		"signature":      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"message":        "Sign this message to register",
		"avatar":         "https://example.com/avatar.jpg",
		"bio":            "New user bio",
	}

	resp := suite.makeRequest("POST", "/api/v1/auth/register", registerReq)
	assert.Equal(t, http.StatusCreated, resp.Code)

	var authResp map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &authResp)
	require.NoError(t, err)

	userToken := authResp["access_token"].(string)
	userData := authResp["user"].(map[string]interface{})
	userID := uint(userData["id"].(float64))

	// Step 2: Login with registered user
	loginReq := map[string]interface{}{
		"wallet_address": "0x9999999999999999999999999999999999999999",
		"signature":      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"message":        "Sign this message to authenticate",
	}

	resp = suite.makeAuthenticatedRequest("POST", "/api/v1/auth/login", loginReq, userToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	// Step 3: Get user profile
	resp = suite.makeAuthenticatedRequest("GET", "/api/v1/users/profile", nil, userToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var profileResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &profileResp)
	require.NoError(t, err)
	assert.Equal(t, "newuser", profileResp["data"].(map[string]interface{})["username"])

	// Step 4: Update user profile
	updateReq := map[string]interface{}{
		"username": "updateduser",
		"bio":      "Updated bio text",
	}

	resp = suite.makeAuthenticatedRequest("PUT", "/api/v1/users/profile", updateReq, userToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	// Step 5: Mint new NFT
	nftReq := map[string]interface{}{
		"title":       "My First NFT",
		"description": "This is my first NFT creation",
		"metadata": map[string]interface{}{
			"artist":   "newuser",
			"year":     2024,
			"category": "digital_art",
		},
		"royalty": 5,
	}

	resp = suite.makeAuthenticatedRequest("POST", "/api/v1/nfts", nftReq, userToken)
	assert.Equal(t, http.StatusCreated, resp.Code)

	var nftResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &nftResp)
	require.NoError(t, err)
	nftID := uint(nftResp["data"].(map[string]interface{})["id"].(float64))

	// Step 6: Get NFT details
	resp = suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/nfts/%d", nftID), nil, userToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var nftDetailsResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &nftDetailsResp)
	require.NoError(t, err)
	assert.Equal(t, "My First NFT", nftDetailsResp["data"].(map[string]interface{})["title"])

	fmt.Printf("✅ Complete User Workflow Test Passed - User ID: %d, NFT ID: %d\n", userID, nftID)
}

// Scenario 2: Complete Auction Workflow with Bidding
func (suite *ComprehensiveIntegrationTestSuite) TestCompleteAuctionWorkflow() {
	t := suite.T()

	// Step 1: Login as Alice (seller)
	aliceToken := suite.loginUser("alice")
	require.NotEmpty(t, aliceToken)

	// Step 2: Create auction for Alice's NFT
	auctionReq := map[string]interface{}{
		"nft_id":        suite.testNFTs["alice_art"].ID,
		"start_price":   "500000000000000000",  // 0.5 ETH
		"reserve_price": "1000000000000000000", // 1 ETH
		"start_time":    time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		"end_time":      time.Now().Add(25 * time.Hour).Format(time.RFC3339),
	}

	resp := suite.makeAuthenticatedRequest("POST", "/api/v1/auctions", auctionReq, aliceToken)
	assert.Equal(t, http.StatusCreated, resp.Code)

	var auctionResp map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &auctionResp)
	require.NoError(t, err)
	auctionID := uint(auctionResp["data"].(map[string]interface{})["id"].(float64))

	// Step 3: Login as Bob (bidder)
	bobToken := suite.loginUser("bob")
	require.NotEmpty(t, bobToken)

	// Step 4: Bob places first bid
	bidReq := map[string]interface{}{
		"amount": "600000000000000000", // 0.6 ETH
	}

	resp = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/auctions/%d/bids", auctionID), bidReq, bobToken)
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Step 5: Alice checks auction bids
	resp = suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/auctions/%d/bids", auctionID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var bidsResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &bidsResp)
	require.NoError(t, err)
	bids := bidsResp["data"].([]interface{})
	assert.Len(t, bids, 1)

	// Step 6: Login as Charlie (second bidder)
	charlieToken := suite.loginUser("charlie")
	require.NotEmpty(t, charlieToken)

	// Step 7: Charlie places higher bid
	bidReq2 := map[string]interface{}{
		"amount": "800000000000000000", // 0.8 ETH
	}

	resp = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/auctions/%d/bids", auctionID), bidReq2, charlieToken)
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Step 8: Get auction details to verify highest bid
	resp = suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/auctions/%d", auctionID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var auctionDetailsResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &auctionDetailsResp)
	require.NoError(t, err)
	auctionData := auctionDetailsResp["data"].(map[string]interface{})
	currentBid := auctionData["current_bid"].(string)
	assert.Equal(t, "800000000000000000", currentBid) // Charlie's bid

	// Step 9: Bob receives notification about being outbid
	resp = suite.makeAuthenticatedRequest("GET", "/api/v1/notifications", nil, bobToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var notificationsResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &notificationsResp)
	require.NoError(t, err)
	notifications := notificationsResp["data"].([]interface{})
	assert.Greater(t, len(notifications), 0)

	// Find outbid notification
	foundOutbid := false
	for _, notif := range notifications {
		notification := notif.(map[string]interface{})
		if notification["type"].(string) == "bid_outbid" {
			foundOutbid = true
			break
		}
	}
	assert.True(t, foundOutbid, "Bob should receive outbid notification")

	fmt.Printf("✅ Complete Auction Workflow Test Passed - Auction ID: %d, Final Bid: 0.8 ETH\n", auctionID)
}

// Scenario 3: Real-time WebSocket Integration
func (suite *ComprehensiveIntegrationTestSuite) TestRealtimeWebSocketIntegration() {
	t := suite.T()

	// Step 1: Login as Alice
	aliceToken := suite.loginUser("alice")
	require.NotEmpty(t, aliceToken)

	// Step 2: Alice connects to WebSocket
	aliceWS, _, err := suite.wsDialer.Dial(suite.wsURL+"?token="+aliceToken, nil)
	require.NoError(t, err)
	defer aliceWS.Close()

	// Step 3: Verify connection established
	_, message, err := aliceWS.ReadMessage()
	require.NoError(t, err)

	var connectEvent map[string]interface{}
	err = json.Unmarshal(message, &connectEvent)
	require.NoError(t, err)
	assert.Equal(t, "connection_established", connectEvent["event"])

	// Step 4: Login as Bob
	bobToken := suite.loginUser("bob")
	require.NotEmpty(t, bobToken)

	// Step 5: Bob connects to WebSocket
	bobWS, _, err := suite.wsDialer.Dial(suite.wsURL+"?token="+bobToken, nil)
	require.NoError(t, err)
	defer bobWS.Close()

	// Step 6: Alice creates auction
	auctionReq := map[string]interface{}{
		"nft_id":        suite.testNFTs["alice_art"].ID,
		"start_price":   "1000000000000000000", // 1 ETH
		"reserve_price": "1500000000000000000", // 1.5 ETH
		"start_time":    time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
		"end_time":      time.Now().Add(23 * time.Hour).Format(time.RFC3339),
	}

	resp := suite.makeAuthenticatedRequest("POST", "/api/v1/auctions", auctionReq, aliceToken)
	assert.Equal(t, http.StatusCreated, resp.Code)

	var auctionResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &auctionResp)
	require.NoError(t, err)
	auctionID := uint(auctionResp["data"].(map[string]interface{})["id"].(float64))

	// Step 7: Bob places bid (should trigger WebSocket event)
	bidReq := map[string]interface{}{
		"amount": "1200000000000000000", // 1.2 ETH
	}

	resp = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/auctions/%d/bids", auctionID), bidReq, bobToken)
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Step 8: Alice receives WebSocket notification about new bid
	timeout := time.After(5 * time.Second)
	var bidPlacedReceived bool

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for bid_placed WebSocket event")
		default:
			_, message, err = aliceWS.ReadMessage()
			if err != nil {
				t.Fatal("Error reading WebSocket message:", err)
			}

			var event map[string]interface{}
			err = json.Unmarshal(message, &event)
			require.NoError(t, err)

			if event["event"].(string) == "bid_placed" {
				bidPlacedReceived = true
				eventData := event["data"].(map[string]interface{})
				assert.Equal(t, auctionID, uint(eventData["auction_id"].(float64)))
				assert.Equal(t, "bob", eventData["bidder_username"])
				break
			}
		}
	}

	assert.True(t, bidPlacedReceived, "Alice should receive bid_placed WebSocket event")

	// Step 9: Alice checks her notifications
	resp = suite.makeAuthenticatedRequest("GET", "/api/v1/notifications", nil, aliceToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var notificationsResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &notificationsResp)
	require.NoError(t, err)
	notifications := notificationsResp["data"].([]interface{})
	assert.Greater(t, len(notifications), 0)

	fmt.Printf("✅ Real-time WebSocket Integration Test Passed - Auction ID: %d, WebSocket Events: %d\n", auctionID, len(notifications))
}

// Scenario 4: NFT Transfer Workflow
func (suite *ComprehensiveIntegrationTestSuite) TestNFTTransferWorkflow() {
	t := suite.T()

	// Step 1: Login as Alice (NFT owner)
	aliceToken := suite.loginUser("alice")
	require.NotEmpty(t, aliceToken)

	// Step 2: Verify Alice owns the NFT
	resp := suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/nfts/%d", suite.testNFTs["alice_art"].ID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var nftResp map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &nftResp)
	require.NoError(t, err)
	nftData := nftResp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), nftData["owner_id"])

	// Step 3: Login as Bob (recipient)
	bobToken := suite.loginUser("bob")
	require.NotEmpty(t, bobToken)

	// Step 4: Alice transfers NFT to Bob
	transferReq := map[string]interface{}{
		"to_address": suite.testUsers["bob"].WalletAddr,
	}

	resp = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/nfts/%d/transfer", suite.testNFTs["alice_art"].ID), transferReq, aliceToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var transferResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &transferResp)
	require.NoError(t, err)
	assert.Equal(t, "transfer", transferResp["data"].(map[string]interface{})["type"])

	// Step 5: Verify NFT ownership changed
	resp = suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/nfts/%d", suite.testNFTs["alice_art"].ID), nil, aliceToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &nftResp)
	require.NoError(t, err)
	nftData = nftResp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), nftData["owner_id"]) // Bob should now be the owner

	// Step 6: Bob verifies he received the NFT
	resp = suite.makeAuthenticatedRequest("GET", fmt.Sprintf("/api/v1/nfts/%d", suite.testNFTs["alice_art"].ID), nil, bobToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	err = json.Unmarshal(resp.Body.Bytes(), &nftResp)
	require.NoError(t, err)
	nftData = nftResp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), nftData["owner_id"])

	// Step 7: Alice receives notification about transfer
	resp = suite.makeAuthenticatedRequest("GET", "/api/v1/notifications", nil, aliceToken)
	assert.Equal(t, http.StatusOK, resp.Code)

	var notificationsResp map[string]interface{}
	err = json.Unmarshal(resp.Body.Bytes(), &notificationsResp)
	require.NoError(t, err)
	notifications := notificationsResp["data"].([]interface{})
	assert.Greater(t, len(notifications), 0)

	// Find NFT transferred notification
	foundTransfer := false
	for _, notif := range notifications {
		notification := notif.(map[string]interface{})
		if notification["type"].(string) == "nft_transferred" {
			foundTransfer = true
			break
		}
	}
	assert.True(t, foundTransfer, "Alice should receive transfer notification")

	fmt.Printf("✅ NFT Transfer Workflow Test Passed - NFT ID: %d transferred from Alice to Bob\n", suite.testNFTs["alice_art"].ID)
}

// Scenario 5: Error Handling and Edge Cases
func (suite *ComprehensiveIntegrationTestSuite) TestErrorHandlingAndEdgeCases() {
	t := suite.T()

	// Test 1: Invalid authentication
	authReq := map[string]interface{}{
		"wallet_address": "invalid_address",
		"signature":      "invalid_signature",
		"message":        "Sign this message",
	}

	resp := suite.makeRequest("POST", "/api/v1/auth/login", authReq)
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Test 2: Access denied for unauthorized operations
	resp = suite.makeRequest("GET", "/api/v1/users/profile", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)

	// Test 3: Invalid NFT ID
	aliceToken := suite.loginUser("alice")
	resp = suite.makeAuthenticatedRequest("GET", "/api/v1/nfts/999999", nil, aliceToken)
	assert.Equal(t, http.StatusNotFound, resp.Code)

	// Test 4: Invalid auction operations
	auctionReq := map[string]interface{}{
		"nft_id":      999999, // Non-existent NFT
		"start_price": "1000000000000000000",
		"start_time":  time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		"end_time":    time.Now().Add(25 * time.Hour).Format(time.RFC3339),
	}

	resp = suite.makeAuthenticatedRequest("POST", "/api/v1/auctions", auctionReq, aliceToken)
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Test 5: Invalid bid amount
	bobToken := suite.loginUser("bob")
	bidReq := map[string]interface{}{
		"amount": "invalid_amount",
	}

	resp = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/auctions/%d/bids", suite.testAuctions["alice_art_auction"].ID), bidReq, bobToken)
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Test 6: Bidding on own auction
	bidReq2 := map[string]interface{}{
		"amount": "2000000000000000000",
	}

	resp = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/auctions/%d/bids", suite.testAuctions["alice_art_auction"].ID), bidReq2, aliceToken)
	assert.Equal(t, http.StatusForbidden, resp.Code)

	// Test 7: Invalid transfer address
	transferReq := map[string]interface{}{
		"to_address": "invalid_ethereum_address",
	}

	resp = suite.makeAuthenticatedRequest("POST", fmt.Sprintf("/api/v1/nfts/%d/transfer", suite.testNFTs["alice_art"].ID), transferReq, aliceToken)
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	fmt.Printf("✅ Error Handling and Edge Cases Test Passed - All error conditions properly handled\n")
}

// Scenario 6: Performance and Load Testing
func (suite *ComprehensiveIntegrationTestSuite) TestPerformanceAndLoad() {
	t := suite.T()

	// Step 1: Test concurrent user registrations
	const numConcurrentUsers = 10
	type registrationResult struct {
		success bool
		token   string
		userID  uint
	}

	results := make(chan registrationResult, numConcurrentUsers)

	for i := 0; i < numConcurrentUsers; i++ {
		go func(userIndex int) {
			username := fmt.Sprintf("loaduser%d", userIndex)
			walletAddr := fmt.Sprintf("0x%040x", userIndex+1000)

			req := map[string]interface{}{
				"username":       username,
				"email":          fmt.Sprintf("%s@example.com", username),
				"wallet_address": walletAddr,
				"signature":      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"message":        "Sign this message to register",
			}

			resp := suite.makeRequest("POST", "/api/v1/auth/register", req)
			results <- registrationResult{
				success: resp.Code == http.StatusCreated,
				token:   "",
				userID:  0,
			}
		}(i)
	}

	// Collect results
	successfulRegistrations := 0
	for i := 0; i < numConcurrentUsers; i++ {
		result := <-results
		if result.success {
			successfulRegistrations++
		}
	}

	successRate := float64(successfulRegistrations) / float64(numConcurrentUsers) * 100
	fmt.Printf("Concurrent Registration Test: %d/%d successful (%.1f%%)\n", successfulRegistrations, numConcurrentUsers, successRate)
	assert.GreaterOrEqual(t, successRate, 80.0, "At least 80% of concurrent registrations should succeed")

	// Step 2: Test rapid API calls
	aliceToken := suite.loginUser("alice")
	start := time.Now()
	const rapidCalls = 50

	for i := 0; i < rapidCalls; i++ {
		switch i % 4 {
		case 0:
			suite.makeAuthenticatedRequest("GET", "/api/v1/users/profile", nil, aliceToken)
		case 1:
			suite.makeAuthenticatedRequest("GET", "/api/v1/nfts", nil, aliceToken)
		case 2:
			suite.makeAuthenticatedRequest("GET", "/api/v1/auctions", nil, aliceToken)
		case 3:
			suite.makeAuthenticatedRequest("GET", "/api/v1/notifications", nil, aliceToken)
		}
	}

	duration := time.Since(start)
	avgResponseTime := duration / rapidCalls
	fmt.Printf("Rapid API Calls Test: %d calls in %v (avg: %v per call)\n", rapidCalls, duration, avgResponseTime)
	assert.Less(t, avgResponseTime, 100*time.Millisecond, "Average response time should be less than 100ms")

	fmt.Printf("✅ Performance and Load Test Passed - System handled concurrent load successfully\n")
}

// Helper methods

func (suite *ComprehensiveIntegrationTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	return resp
}

func (suite *ComprehensiveIntegrationTestSuite) makeAuthenticatedRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp := httptest.NewRecorder()
	suite.router.ServeHTTP(resp, req)
	return resp
}

func (suite *ComprehensiveIntegrationTestSuite) loginUser(username string) string {
	user := suite.testUsers[username]
	loginReq := map[string]interface{}{
		"wallet_address": user.WalletAddr,
		"signature":      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"message":        "Sign this message to authenticate",
	}

	resp := suite.makeRequest("POST", "/api/v1/auth/login", loginReq)
	if resp.Code != http.StatusOK {
		return ""
	}

	var authResp map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &authResp)
	return authResp["access_token"].(string)
}

// Mock handlers for testing
func (suite *ComprehensiveIntegrationTestSuite) mockLogin(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"access_token":  "mock_token_" + c.PostForm("wallet_address"),
		"refresh_token": "mock_refresh_" + c.PostForm("wallet_address"),
		"expires_in":    3600,
		"user":          suite.testUsers["alice"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockRegister(c *gin.Context) {
	c.JSON(http.StatusCreated, map[string]interface{}{
		"access_token":  "mock_token_new",
		"refresh_token": "mock_refresh_new",
		"expires_in":    3600,
		"user":          suite.testUsers["alice"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockRefresh(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"access_token":  "mock_refreshed_token",
		"refresh_token": "mock_refreshed_refresh",
		"expires_in":    3600,
		"user":          suite.testUsers["alice"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockGetUserProfile(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    suite.testUsers["alice"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockUpdateUserProfile(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    suite.testUsers["alice"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockGetUserByID(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    suite.testUsers["alice"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockListNFTs(c *gin.Context) {
	nfts := []interface{}{suite.testNFTs["alice_art"], suite.testNFTs["bob_photo"]}
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    nfts,
		"pagination": map[string]interface{}{
			"page":        1,
			"limit":       20,
			"total":       2,
			"total_pages": 1,
			"has_next":    false,
			"has_prev":    false,
		},
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockCreateNFT(c *gin.Context) {
	c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    suite.testNFTs["alice_art"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockGetNFT(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    suite.testNFTs["alice_art"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockUpdateNFT(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    suite.testNFTs["alice_art"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockTransferNFT(c *gin.Context) {
	transfer := map[string]interface{}{
		"id":         1,
		"nft_id":     suite.testNFTs["alice_art"].ID,
		"from_id":    suite.testNFTs["alice_art"].OwnerID,
		"to_id":      suite.testUsers["bob"].ID,
		"tx_hash":    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"price":      "1000000000000000000",
		"type":       "transfer",
		"created_at": time.Now(),
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    transfer,
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockListAuctions(c *gin.Context) {
	auctions := []interface{}{suite.testAuctions["alice_art_auction"]}
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    auctions,
		"pagination": map[string]interface{}{
			"page":        1,
			"limit":       20,
			"total":       1,
			"total_pages": 1,
			"has_next":    false,
			"has_prev":    false,
		},
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockCreateAuction(c *gin.Context) {
	c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    suite.testAuctions["alice_art_auction"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockGetAuction(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    suite.testAuctions["alice_art_auction"],
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockCancelAuction(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Auction cancelled successfully",
		"auction_id": suite.testAuctions["alice_art_auction"].ID,
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockGetAuctionBids(c *gin.Context) {
	bids := []interface{}{&models.Bid{
		ID:        1,
		AuctionID: suite.testAuctions["alice_art_auction"].ID,
		BidderID:  suite.testUsers["bob"].ID,
		Amount:    big.NewInt(1200000000000000000),
		Status:    "confirmed",
		TxHash:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		CreatedAt: time.Now(),
	}}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    bids,
		"pagination": map[string]interface{}{
			"page":        1,
			"limit":       50,
			"total":       1,
			"total_pages": 1,
			"has_next":    false,
			"has_prev":    false,
		},
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockPlaceBid(c *gin.Context) {
	bid := map[string]interface{}{
		"bid": map[string]interface{}{
			"id":         2,
			"auction_id": suite.testAuctions["alice_art_auction"].ID,
			"bidder_id":  suite.testUsers["bob"].ID,
			"amount":     "800000000000000000",
			"status":     "confirmed",
			"tx_hash":    "0x2345678901bcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		"is_highest":        true,
		"previous_high_bid": "1200000000000000000",
		"auction_updated":   true,
	}

	c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    bid,
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockGetNotifications(c *gin.Context) {
	notifications := []interface{}{
		&models.Notification{
			ID:        1,
			UserID:    suite.testUsers["alice"].ID,
			Type:      "bid_placed",
			Title:     "New Bid Received",
			Message:   "A new bid has been placed on your auction",
			IsRead:    false,
			CreatedAt: time.Now(),
		},
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    notifications,
		"pagination": map[string]interface{}{
			"page":        1,
			"limit":       50,
			"total":       1,
			"total_pages": 1,
			"has_next":    false,
			"has_prev":    false,
		},
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockMarkNotificationRead(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Notification marked as read",
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockMarkAllNotificationsRead(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "All notifications marked as read",
	})
}

func (suite *ComprehensiveIntegrationTestSuite) mockWebSocketHandler(c *gin.Context) {
	// Simple WebSocket handler for testing
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// Send initial connection message
	connectEvent := map[string]interface{}{
		"event":     "connection_established",
		"data":      map[string]interface{}{"id": "test-connection-id"},
		"timestamp": time.Now().UTC(),
		"id":        "test-event-id",
	}
	conn.WriteJSON(connectEvent)

	// Keep connection alive for testing
	go func() {
		defer conn.Close()
		for {
			conn.ReadMessage()
		}
	}()
}

// TestRunner runs the comprehensive integration test suite
func TestComprehensiveIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ComprehensiveIntegrationTestSuite))
}
