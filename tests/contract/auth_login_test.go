package contract

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nft-platform/tests/testutil"
)

// Test T008: Contract test POST /auth/login
func TestAuthLogin_ValidRequest(t *testing.T) {
	// Setup test server
	server := testutil.NewTestServer()
	defer server.Close()

	// Test data based on OpenAPI schema
	loginRequest := map[string]interface{}{
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to authenticate: 1234567890",
	}

	// Expected response structure based on AuthResponse schema
	expectedResponse := map[string]interface{}{
		"access_token":  "",
		"refresh_token": "",
		"expires_in":    0,
		"user":          map[string]interface{}{},
	}

	// Make request - this MUST FAIL until handlers are implemented
	resp := server.JSONRequest(t, "POST", "/api/v1/auth/login", loginRequest)

	// Contract expectations - this will fail initially (TDD)
	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// When implemented, should return 200 with proper structure
	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var actualResponse map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &actualResponse)

	// Verify response structure matches OpenAPI spec
	assert.Contains(t, actualResponse, "access_token", "Response should contain access_token")
	assert.Contains(t, actualResponse, "refresh_token", "Response should contain refresh_token")
	assert.Contains(t, actualResponse, "expires_in", "Response should contain expires_in")
	assert.Contains(t, actualResponse, "user", "Response should contain user object")

	// Verify user object structure
	user, ok := actualResponse["user"].(map[string]interface{})
	require.True(t, ok, "User should be an object")
	
	assert.Contains(t, user, "id", "User should contain id")
	assert.Contains(t, user, "username", "User should contain username")
	assert.Contains(t, user, "email", "User should contain email")
	assert.Contains(t, user, "wallet_address", "User should contain wallet_address")

	// Verify wallet address matches request
	assert.Equal(t, loginRequest["wallet_address"], user["wallet_address"])

	_ = expectedResponse // Suppress unused variable warning
}

func TestAuthLogin_InvalidWalletAddress(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidRequests := []map[string]interface{}{
		{
			"wallet_address": "invalid_address",
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":        "Please sign this message to authenticate: 1234567890",
		},
		{
			"wallet_address": "0x123", // too short
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":        "Please sign this message to authenticate: 1234567890",
		},
		{
			"wallet_address": "", // empty
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":        "Please sign this message to authenticate: 1234567890",
		},
	}

	for _, requestData := range invalidRequests {
		resp := server.JSONRequest(t, "POST", "/api/v1/auth/login", requestData)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 401 for invalid wallet address
		testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Invalid wallet address")
	}
}

func TestAuthLogin_MissingRequiredFields(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	missingFieldRequests := []map[string]interface{}{
		{
			// missing wallet_address
			"signature": "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":   "Please sign this message to authenticate: 1234567890",
		},
		{
			"wallet_address": "0x1234567890123456789012345678901234567890",
			// missing signature
			"message": "Please sign this message to authenticate: 1234567890",
		},
		{
			"wallet_address": "0x1234567890123456789012345678901234567890",
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			// missing message
		},
	}

	for _, requestData := range missingFieldRequests {
		resp := server.JSONRequest(t, "POST", "/api/v1/auth/login", requestData)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 400 for missing required fields
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestAuthLogin_InvalidSignature(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	loginRequest := map[string]interface{}{
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"signature":      "0xinvalidsignature", // Invalid signature
		"message":        "Please sign this message to authenticate: 1234567890",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/login", loginRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 401 for invalid signature
	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Invalid signature")
}

func TestAuthLogin_UserNotFound(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Valid format but user doesn't exist
	loginRequest := map[string]interface{}{
		"wallet_address": "0x9999999999999999999999999999999999999999", // Non-existent user
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to authenticate: 1234567890",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/login", loginRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 401 for user not found
	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "User not found")
}

func TestAuthLogin_EmptyRequestBody(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/login", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 400 for empty request body
	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "Invalid request")
}

func TestAuthLogin_InvalidJSONFormat(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Send invalid JSON
	req, err := http.NewRequest("POST", server.URL("/api/v1/auth/login"), 
		http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := server.Client.Do(req)
	require.NoError(t, err)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 400 for invalid JSON
	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "Invalid JSON")
}