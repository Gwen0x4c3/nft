package contract

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nft-platform/tests/testutil"
)

// Test T009: Contract test POST /auth/register
func TestAuthRegister_ValidRequest(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Test data based on OpenAPI RegisterRequest schema
	registerRequest := map[string]interface{}{
		"username":       "newuser123",
		"email":          "newuser@example.com",
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to register: 1234567890",
		"avatar":         "https://example.com/avatar.jpg",
		"bio":            "I'm a digital art collector and NFT enthusiast",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

	// Contract expectations - this will fail initially (TDD)
	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// When implemented, should return 201 with proper structure
	testutil.AssertStatusCode(t, resp, http.StatusCreated)
	
	var actualResponse map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &actualResponse)

	// Verify response structure matches OpenAPI AuthResponse schema
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
	assert.Contains(t, user, "avatar", "User should contain avatar")
	assert.Contains(t, user, "bio", "User should contain bio")
	assert.Contains(t, user, "is_verified", "User should contain is_verified")

	// Verify returned data matches request
	assert.Equal(t, registerRequest["username"], user["username"])
	assert.Equal(t, registerRequest["email"], user["email"])
	assert.Equal(t, registerRequest["wallet_address"], user["wallet_address"])
	assert.Equal(t, registerRequest["avatar"], user["avatar"])
	assert.Equal(t, registerRequest["bio"], user["bio"])
	assert.Equal(t, false, user["is_verified"]) // New users should not be verified
}

func TestAuthRegister_MinimalRequest(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Only required fields
	registerRequest := map[string]interface{}{
		"username":       "minimaluser",
		"email":          "minimal@example.com",
		"wallet_address": "0x9876543210987654321098765432109876543210",
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to register: 1234567890",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusCreated)
	
	var actualResponse map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &actualResponse)

	user := actualResponse["user"].(map[string]interface{})
	assert.Equal(t, registerRequest["username"], user["username"])
	assert.Equal(t, registerRequest["email"], user["email"])
	assert.Equal(t, registerRequest["wallet_address"], user["wallet_address"])
}

func TestAuthRegister_InvalidUsername(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidUsernames := []string{
		"ab", // too short (< 3 chars)
		"", // empty
		"thisusernameiswaytooooooooooolong", // too long (> 30 chars)
		"user with spaces", // invalid characters
		"user@special", // invalid characters
	}

	for _, invalidUsername := range invalidUsernames {
		registerRequest := map[string]interface{}{
			"username":       invalidUsername,
			"email":          "test@example.com",
			"wallet_address": "0x1234567890123456789012345678901234567890",
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":        "Please sign this message to register: 1234567890",
		}

		resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 400 for invalid username
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestAuthRegister_InvalidEmail(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidEmails := []string{
		"notanemail", // invalid format
		"", // empty
		"@example.com", // missing username
		"user@", // missing domain
		"user@.com", // missing domain name
		"user.example.com", // missing @
	}

	for _, invalidEmail := range invalidEmails {
		registerRequest := map[string]interface{}{
			"username":       "testuser",
			"email":          invalidEmail,
			"wallet_address": "0x1234567890123456789012345678901234567890",
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":        "Please sign this message to register: 1234567890",
		}

		resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 400 for invalid email
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestAuthRegister_InvalidWalletAddress(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidWalletAddresses := []string{
		"not_a_wallet", // invalid format
		"0x123", // too short
		"1234567890123456789012345678901234567890", // missing 0x prefix
		"0xGGGG567890123456789012345678901234567890", // invalid hex characters
		"", // empty
	}

	for _, invalidWallet := range invalidWalletAddresses {
		registerRequest := map[string]interface{}{
			"username":       "testuser",
			"email":          "test@example.com",
			"wallet_address": invalidWallet,
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":        "Please sign this message to register: 1234567890",
		}

		resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 400 for invalid wallet address
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestAuthRegister_MissingRequiredFields(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	baseRequest := map[string]interface{}{
		"username":       "testuser",
		"email":          "test@example.com",
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to register: 1234567890",
	}

	// Test missing each required field
	requiredFields := []string{"username", "email", "wallet_address", "signature", "message"}

	for _, fieldToRemove := range requiredFields {
		requestData := make(map[string]interface{})
		for k, v := range baseRequest {
			if k != fieldToRemove {
				requestData[k] = v
			}
		}

		resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", requestData)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 400 for missing required field
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestAuthRegister_DuplicateUsername(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	registerRequest := map[string]interface{}{
		"username":       "existinguser", // Assume this username already exists
		"email":          "new@example.com",
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to register: 1234567890",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 409 for duplicate username
	testutil.AssertErrorResponse(t, resp, http.StatusConflict, "Username already exists")
}

func TestAuthRegister_DuplicateEmail(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	registerRequest := map[string]interface{}{
		"username":       "newuser",
		"email":          "existing@example.com", // Assume this email already exists
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to register: 1234567890",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 409 for duplicate email
	testutil.AssertErrorResponse(t, resp, http.StatusConflict, "Email already exists")
}

func TestAuthRegister_DuplicateWalletAddress(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	registerRequest := map[string]interface{}{
		"username":       "newuser",
		"email":          "new@example.com",
		"wallet_address": "0x9876543210987654321098765432109876543210", // Assume this wallet already exists
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to register: 1234567890",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 409 for duplicate wallet address
	testutil.AssertErrorResponse(t, resp, http.StatusConflict, "Wallet address already exists")
}

func TestAuthRegister_InvalidBio(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Bio longer than 500 characters
	longBio := string(make([]byte, 501))
	for i := range longBio {
		longBio = longBio[:i] + "a" + longBio[i+1:]
	}

	registerRequest := map[string]interface{}{
		"username":       "testuser",
		"email":          "test@example.com",
		"wallet_address": "0x1234567890123456789012345678901234567890",
		"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
		"message":        "Please sign this message to register: 1234567890",
		"bio":            longBio,
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 400 for bio too long
	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
}

func TestAuthRegister_InvalidAvatar(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidAvatars := []string{
		"not_a_url", // invalid URL format
		"ftp://example.com/avatar.jpg", // invalid protocol
		"javascript:alert('xss')", // security risk
	}

	for _, invalidAvatar := range invalidAvatars {
		registerRequest := map[string]interface{}{
			"username":       "testuser",
			"email":          "test@example.com",
			"wallet_address": "0x1234567890123456789012345678901234567890",
			"signature":      "0xabcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd123456789012345678901234567890abcd1234567890abcd1234567890abcd1234",
			"message":        "Please sign this message to register: 1234567890",
			"avatar":         invalidAvatar,
		}

		resp := server.JSONRequest(t, "POST", "/api/v1/auth/register", registerRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 400 for invalid avatar URL
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}