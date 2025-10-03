package testutil

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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"nft-platform/internal/models"
)

// TestUser creates a test user with default values
func TestUser(id uint, username string) *models.User {
	return &models.User{
		ID:         id,
		Username:   username,
		Email:      fmt.Sprintf("%s@example.com", username),
		WalletAddr: fmt.Sprintf("0x%040x", id), // Generate test wallet address
		Avatar:     fmt.Sprintf("https://example.com/avatars/%s.jpg", username),
		Bio:        fmt.Sprintf("Test user %s", username),
		IsVerified: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// TestNFT creates a test NFT with default values
func TestNFT(id uint, creatorID, ownerID uint, title string) *models.NFT {
	return &models.NFT{
		ID:           id,
		TokenID:      fmt.Sprintf("%d", id),
		ContractAddr: "0x1234567890123456789012345678901234567890",
		CreatorID:    creatorID,
		OwnerID:      ownerID,
		Title:        title,
		Description:  fmt.Sprintf("Test NFT: %s", title),
		ImageURL:     fmt.Sprintf("https://example.com/nfts/%d.jpg", id),
		MetadataURI:  fmt.Sprintf("https://example.com/metadata/%d.json", id),
		Price:        big.NewInt(int64(id) * 1000000000000000000), // id ETH in Wei
		IsForSale:    id%2 == 0,
		Royalty:      uint8(id % 11), // 0-10%
		MintTxHash:   fmt.Sprintf("0x%064x", id),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// TestAuction creates a test auction with default values
func TestAuction(id uint, nftID, sellerID uint, status string) *models.Auction {
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now().Add(time.Hour)

	return &models.Auction{
		ID:           id,
		NFTID:        nftID,
		SellerID:     sellerID,
		StartPrice:   big.NewInt(1000000000000000000),            // 1 ETH
		ReservePrice: big.NewInt(1500000000000000000),            // 1.5 ETH
		CurrentBid:   big.NewInt(int64(id) * 500000000000000000), // Variable current bid
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       status,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// TestBid creates a test bid with default values
func TestBid(id uint, auctionID, bidderID uint, amount int64) *models.Bid {
	return &models.Bid{
		ID:        id,
		AuctionID: auctionID,
		BidderID:  bidderID,
		Amount:    big.NewInt(amount),
		Status:    "confirmed",
		TxHash:    fmt.Sprintf("0x%064x", id),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestNotification creates a test notification with default values
func TestNotification(id uint, userID uint, notificationType string) *models.Notification {
	return &models.Notification{
		ID:        id,
		UserID:    userID,
		Type:      notificationType,
		Title:     fmt.Sprintf("Test %s", notificationType),
		Message:   fmt.Sprintf("This is a test %s notification", notificationType),
		IsRead:    false,
		CreatedAt: time.Now(),
	}
}

// MockTransaction creates a mock blockchain transaction
// func MockTransaction(hash string) *types.Transaction {
// 	return &types.Transaction{
// 		Hash:   hash,
// 		Status: "confirmed",
// 	}
// }

// JSONMarshalHelper marshals data to JSON with error handling for tests
func JSONMarshalHelper(t *testing.T, data interface{}) []byte {
	result, err := json.Marshal(data)
	require.NoError(t, err)
	return result
}

// CreateJSONRequest creates a HTTP request with JSON body
func CreateJSONRequest(method, url string, body interface{}) *http.Request {
	jsonBody := JSONMarshalHelper(nil, body)
	req := httptest.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// CreateGinContext creates a gin context for testing
func CreateGinContext(method, path string, body interface{}, params gin.Params) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	if body != nil {
		jsonBody := JSONMarshalHelper(nil, body)
		c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, path, nil)
	}

	c.Params = params
	return c
}

// AssertRecorderJSONResponse asserts that response contains valid JSON with expected data
func AssertRecorderJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, expectedData interface{}) {
	assert.Equal(t, expectedStatus, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

	if expectedData != nil {
		var actualData interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &actualData)
		require.NoError(t, err)
		assert.Equal(t, expectedData, actualData)
	}
}

// AssertRecorderErrorResponse asserts that response contains expected error structure
func AssertRecorderErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus int, expectedCode string) {
	assert.Equal(t, expectedStatus, recorder.Code)
	assert.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["success"].(bool))

	errorData := response["error"].(map[string]interface{})
	assert.Equal(t, expectedCode, errorData["code"])
	assert.NotEmpty(t, errorData["message"])
}

// WebSocketTestHelper provides utilities for WebSocket testing
type WebSocketTestHelper struct {
	Dialer websocket.Dialer
	URL    string
}

// NewWebSocketTestHelper creates a new WebSocket test helper
func NewWebSocketTestHelper(url string) *WebSocketTestHelper {
	return &WebSocketTestHelper{
		Dialer: websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
		URL: url,
	}
}

// Connect establishes a WebSocket connection for testing
func (h *WebSocketTestHelper) Connect(t *testing.T, token string) *websocket.Conn {
	conn, _, err := h.Dialer.Dial(h.URL+"?token="+token, nil)
	require.NoError(t, err)
	return conn
}

// ReadMessage reads and parses a WebSocket message
func (h *WebSocketTestHelper) ReadMessage(t *testing.T, conn *websocket.Conn) (map[string]interface{}, error) {
	_, message, err := conn.ReadMessage()
	require.NoError(t, err)

	var data map[string]interface{}
	err = json.Unmarshal(message, &data)
	require.NoError(t, err)

	return data, nil
}

// AssertWebSocketEvent asserts that a WebSocket event matches expected values
func (h *WebSocketTestHelper) AssertWebSocketEvent(t *testing.T, conn *websocket.Conn, expectedEvent string, expectedData map[string]interface{}) {
	data, err := h.ReadMessage(t, conn)
	require.NoError(t, err)

	assert.Equal(t, expectedEvent, data["event"])
	assert.NotEmpty(t, data["timestamp"])
	assert.NotEmpty(t, data["id"])

	if expectedData != nil {
		assert.Equal(t, expectedData, data["data"])
	}
}

// MockFactory provides common mock creation utilities
type MockFactory struct{}

// NewMockFactory creates a new mock factory
func NewMockFactory() *MockFactory {
	return &MockFactory{}
}

// CreateUserMock creates a mock user with specified properties
func (f *MockFactory) CreateUserMock(id uint, username string) *models.User {
	user := TestUser(id, username)
	user.CreatedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	user.UpdatedAt = user.CreatedAt
	return user
}

// CreateNFTMock creates a mock NFT with specified properties
func (f *MockFactory) CreateNFTMock(id uint, creatorID, ownerID uint, title string) *models.NFT {
	nft := TestNFT(id, creatorID, ownerID, title)
	nft.CreatedAt = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	nft.UpdatedAt = nft.CreatedAt
	return nft
}

// TimeHelper provides time-related utilities for testing
type TimeHelper struct{}

// NewTimeHelper creates a new time helper
func NewTimeHelper() *TimeHelper {
	return &TimeHelper{}
}

// FixedTime returns a fixed time for consistent testing
func (h *TimeHelper) FixedTime() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

// ParseTime parses a time string for testing
func (h *TimeHelper) ParseTime(timeStr string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05Z", timeStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid time format: %v", err))
	}
	return t
}

// AssertTimeEqual asserts that two times are equal within a tolerance
func (h *TimeHelper) AssertTimeEqual(t *testing.T, expected, actual time.Time, tolerance time.Duration) {
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	assert.Less(t, diff, tolerance, "Times should be equal within tolerance")
}

// DatabaseHelper provides database-related utilities for testing
type DatabaseHelper struct{}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper() *DatabaseHelper {
	return &DatabaseHelper{}
}

// MockDB creates a mock database connection
func (h *DatabaseHelper) MockDB() *mock.Mock {
	return &mock.Mock{}
}

// TestDatabaseConfig returns test database configuration
func (h *DatabaseHelper) TestDatabaseConfig() map[string]string {
	return map[string]string{
		"host":     "localhost",
		"port":     "5432",
		"user":     "test_user",
		"password": "test_password",
		"dbname":   "test_db",
		"sslmode":  "disable",
	}
}

// AssertDatabaseError asserts database error handling
func (h *DatabaseHelper) AssertDatabaseError(t *testing.T, err error, expectedError string) {
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedError)
}

// StringHelper provides string manipulation utilities for testing
type StringHelper struct{}

// NewStringHelper creates a new string helper
func NewStringHelper() *StringHelper {
	return &StringHelper{}
}

// RandomString generates a random string of specified length
func (h *StringHelper) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// RandomEmail generates a random email address
func (h *StringHelper) RandomEmail() string {
	return fmt.Sprintf("test_%s@example.com", h.RandomString(10))
}

// RandomWalletAddress generates a random Ethereum wallet address
func (h *StringHelper) RandomWalletAddress() string {
	return fmt.Sprintf("0x%s", h.RandomString(40))
}

// AssertStringContains asserts that a string contains a substring
func (h *StringHelper) AssertStringContains(t *testing.T, str, substring string) {
	assert.Contains(t, str, substring)
}

// AssertStringLength asserts string length constraints
func (h *StringHelper) AssertStringLength(t *testing.T, str string, min, max int) {
	length := len(str)
	assert.GreaterOrEqual(t, length, min, "String should be at least %d characters", min)
	assert.LessOrEqual(t, length, max, "String should be at most %d characters", max)
}

