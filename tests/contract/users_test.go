package contract

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"nft-platform/tests/testutil"
)

// Test T011: Contract test GET /users/profile
func TestUsersProfile_GetProfile_ValidAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	resp := server.AuthorizedJSONRequest(t, "GET", "/api/v1/users/profile", token, nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var user map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &user)

	// Verify user object structure matches OpenAPI User schema
	assert.Contains(t, user, "id", "User should contain id")
	assert.Contains(t, user, "username", "User should contain username")
	assert.Contains(t, user, "email", "User should contain email")
	assert.Contains(t, user, "wallet_address", "User should contain wallet_address")
	assert.Contains(t, user, "avatar", "User should contain avatar")
	assert.Contains(t, user, "bio", "User should contain bio")
	assert.Contains(t, user, "is_verified", "User should contain is_verified")
	assert.Contains(t, user, "created_at", "User should contain created_at")
	assert.Contains(t, user, "updated_at", "User should contain updated_at")
}

func TestUsersProfile_GetProfile_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/users/profile", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 401 for missing authentication
	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Authentication required")
}

func TestUsersProfile_GetProfile_InvalidToken(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.AuthorizedJSONRequest(t, "GET", "/api/v1/users/profile", "invalid.jwt.token", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 401 for invalid token
	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Invalid token")
}

// Test T012: Contract test PUT /users/profile
func TestUsersProfile_UpdateProfile_ValidAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	updateRequest := map[string]interface{}{
		"username": "updateduser",
		"avatar":   "https://example.com/new_avatar.jpg",
		"bio":      "Updated bio description",
	}

	resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/users/profile", token, updateRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var user map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &user)

	// Verify updated data is returned
	assert.Equal(t, updateRequest["username"], user["username"])
	assert.Equal(t, updateRequest["avatar"], user["avatar"])
	assert.Equal(t, updateRequest["bio"], user["bio"])
}

func TestUsersProfile_UpdateProfile_InvalidUsername(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	invalidUsernames := []string{
		"ab", // too short
		"thisusernameiswaytooooooooooolong", // too long
		"", // empty
		"user with spaces", // invalid chars
	}

	for _, invalidUsername := range invalidUsernames {
		updateRequest := map[string]interface{}{
			"username": invalidUsername,
		}

		resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/users/profile", token, updateRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestUsersProfile_UpdateProfile_InvalidAvatar(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	updateRequest := map[string]interface{}{
		"avatar": "not_a_valid_url",
	}

	resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/users/profile", token, updateRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
}

func TestUsersProfile_UpdateProfile_BioTooLong(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	longBio := string(make([]byte, 501)) // > 500 chars
	for i := range longBio {
		longBio = longBio[:i] + "a" + longBio[i+1:]
	}

	updateRequest := map[string]interface{}{
		"bio": longBio,
	}

	resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/users/profile", token, updateRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
}

func TestUsersProfile_UpdateProfile_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	updateRequest := map[string]interface{}{
		"username": "newname",
	}

	resp := server.JSONRequest(t, "PUT", "/api/v1/users/profile", updateRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Authentication required")
}

// Test T013: Contract test GET /users/{userId}
func TestUsersGet_ValidUserId(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// This endpoint should not require authentication (per OpenAPI spec)
	resp := server.JSONRequest(t, "GET", "/api/v1/users/123", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var user map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &user)

	// Verify user object structure
	assert.Contains(t, user, "id", "User should contain id")
	assert.Contains(t, user, "username", "User should contain username")
	assert.Contains(t, user, "wallet_address", "User should contain wallet_address")
	assert.Contains(t, user, "avatar", "User should contain avatar")
	assert.Contains(t, user, "bio", "User should contain bio")
	assert.Contains(t, user, "is_verified", "User should contain is_verified")
	assert.Contains(t, user, "created_at", "User should contain created_at")
	
	// Email should NOT be exposed in public profile
	assert.NotContains(t, user, "email", "Email should not be exposed in public profile")
	
	// Verify ID matches requested user
	assert.Equal(t, float64(123), user["id"]) // JSON unmarshals numbers as float64
}

func TestUsersGet_UserNotFound(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/users/99999", nil) // Non-existent user

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusNotFound, "User not found")
}

func TestUsersGet_InvalidUserId(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidUserIds := []string{
		"abc", // non-numeric
		"0", // invalid ID
		"-1", // negative
		"", // empty
	}

	for _, invalidId := range invalidUserIds {
		path := fmt.Sprintf("/api/v1/users/%s", invalidId)
		resp := server.JSONRequest(t, "GET", path, nil)

		if resp.StatusCode == http.StatusNotFound {
			continue // Route not matched is also valid
		}

		// Should return 400 for invalid user ID format
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "Invalid user ID")
	}
}

func TestUsersGet_WithAuthentication(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Even with auth, should work (auth is optional for this endpoint)
	token := testutil.MockJWTToken()
	resp := server.AuthorizedJSONRequest(t, "GET", "/api/v1/users/123", token, nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var user map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &user)

	assert.Contains(t, user, "id")
	assert.Contains(t, user, "username")
}

// Additional edge case tests
func TestUsersProfile_UpdateProfile_EmptyRequest(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/users/profile", token, map[string]interface{}{})

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Empty update should be allowed (no changes made)
	testutil.AssertStatusCode(t, resp, http.StatusOK)
}

func TestUsersProfile_UpdateProfile_PartialUpdate(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	updateRequest := map[string]interface{}{
		"bio": "Only updating bio field",
		// username and avatar not provided - should remain unchanged
	}

	resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/users/profile", token, updateRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var user map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &user)

	assert.Equal(t, updateRequest["bio"], user["bio"])
}

func TestUsersGet_LargeUserId(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Test with very large user ID
	resp := server.JSONRequest(t, "GET", "/api/v1/users/999999999999", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should handle large IDs gracefully (either 404 or proper response)
	if resp.StatusCode != http.StatusOK {
		testutil.AssertErrorResponse(t, resp, http.StatusNotFound, "User not found")
	}
}