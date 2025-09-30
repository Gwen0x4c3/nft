package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer represents a test HTTP server
type TestServer struct {
	Server *httptest.Server
	Client *http.Client
	Router *gin.Engine
}

// NewTestServer creates a new test server with all routes registered
func NewTestServer() *TestServer {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup routes using the test router setup function
	SetupTestRoutes(router)
	
	server := httptest.NewServer(router)
	
	return &TestServer{
		Server: server,
		Client: server.Client(),
		Router: router,
	}
}

// Close closes the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// URL returns the server URL with the given path
func (ts *TestServer) URL(path string) string {
	return ts.Server.URL + path
}

// JSONRequest performs a JSON HTTP request and returns the response
func (ts *TestServer) JSONRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, ts.URL(path), reqBody)
	require.NoError(t, err, "Failed to create request")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := ts.Client.Do(req)
	require.NoError(t, err, "Failed to perform request")

	return resp
}

// AuthorizedJSONRequest performs a JSON HTTP request with JWT token
func (ts *TestServer) AuthorizedJSONRequest(t *testing.T, method, path, token string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, ts.URL(path), reqBody)
	require.NoError(t, err, "Failed to create request")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := ts.Client.Do(req)
	require.NoError(t, err, "Failed to perform request")

	return resp
}

// AssertStatusCode asserts the response status code
func AssertStatusCode(t *testing.T, resp *http.Response, expectedStatus int) {
	assert.Equal(t, expectedStatus, resp.StatusCode, "Unexpected status code")
}

// ParseJSONResponse parses JSON response body into the target struct
func ParseJSONResponse(t *testing.T, resp *http.Response, target interface{}) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	err = json.Unmarshal(body, target)
	require.NoError(t, err, fmt.Sprintf("Failed to unmarshal JSON: %s", string(body)))
}

// AssertJSONResponse asserts response status and parses JSON body
func AssertJSONResponse(t *testing.T, resp *http.Response, expectedStatus int, target interface{}) {
	AssertStatusCode(t, resp, expectedStatus)
	ParseJSONResponse(t, resp, target)
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// AssertErrorResponse asserts error response format and status
func AssertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedError string) *ErrorResponse {
	AssertStatusCode(t, resp, expectedStatus)
	
	var errorResp ErrorResponse
	ParseJSONResponse(t, resp, &errorResp)
	
	assert.Contains(t, errorResp.Error, expectedError, "Error message mismatch")
	return &errorResp
}

// MockJWTToken returns a mock JWT token for testing
func MockJWTToken() string {
	// This will fail until JWT implementation is added
	return "mock.jwt.token"
}

// MockUserData returns mock user data for testing
func MockUserData() map[string]interface{} {
	return map[string]interface{}{
		"id":             1,
		"username":       "testuser",
		"email":          "test@example.com",
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"avatar":         "https://example.com/avatar.jpg",
		"bio":            "Test user bio",
		"is_verified":    false,
		"created_at":     "2023-01-01T00:00:00Z",
		"updated_at":     "2023-01-01T00:00:00Z",
	}
}

// MockNFTData returns mock NFT data for testing
func MockNFTData() map[string]interface{} {
	return map[string]interface{}{
		"id":               1,
		"token_id":         "1",
		"contract_address": "0x1234567890123456789012345678901234567890",
		"creator_id":       1,
		"owner_id":         1,
		"title":            "Test NFT",
		"description":      "Test NFT description",
		"image_url":        "https://example.com/nft.jpg",
		"metadata_uri":     "https://example.com/metadata.json",
		"price":            "1000000000000000000", // 1 ETH in Wei
		"is_for_sale":      true,
		"royalty":          5,
		"mint_tx_hash":     "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456",
		"created_at":       "2023-01-01T00:00:00Z",
		"updated_at":       "2023-01-01T00:00:00Z",
	}
}

// MockAuctionData returns mock auction data for testing
func MockAuctionData() map[string]interface{} {
	return map[string]interface{}{
		"id":            1,
		"nft_id":        1,
		"seller_id":     1,
		"start_price":   "500000000000000000", // 0.5 ETH
		"reserve_price": "1000000000000000000", // 1 ETH
		"current_bid":   "0",
		"start_time":    "2023-01-01T12:00:00Z",
		"end_time":      "2023-01-02T12:00:00Z",
		"status":        "pending",
		"winner_id":     nil,
		"created_at":    "2023-01-01T00:00:00Z",
		"updated_at":    "2023-01-01T00:00:00Z",
	}
}

// SetupTestRoutes registers all API routes for testing
func SetupTestRoutes(router *gin.Engine) {
	// Note: This is a placeholder. In a real application, you would:
	// 1. Initialize all repositories with a test database
	// 2. Initialize all services with the repositories
	// 3. Initialize all handlers with the services
	// 4. Register all routes with the handlers
	
	// For now, we'll create a minimal setup to allow routes to be registered
	// The actual implementation will be done when the main server setup is complete
	
	api := router.Group("/api/v1")
	
	// These routes will be implemented when the main application wiring is done
	// For now, they will return 404, which is expected in the TDD approach
	_ = api
}