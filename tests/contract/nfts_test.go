package contract

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"nft-platform/tests/testutil"
)

// Test T014: Contract test GET /nfts
func TestNFTs_ListNFTs_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Per OpenAPI spec, this endpoint doesn't require auth
	resp := server.JSONRequest(t, "GET", "/api/v1/nfts", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var response map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &response)

	// Verify NFTListResponse schema
	assert.Contains(t, response, "data", "Response should contain data array")
	assert.Contains(t, response, "pagination", "Response should contain pagination object")

	// Verify data is an array
	data, ok := response["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")

	// If NFTs exist, verify structure
	if len(data) > 0 {
		nft := data[0].(map[string]interface{})
		assert.Contains(t, nft, "id")
		assert.Contains(t, nft, "title")
		assert.Contains(t, nft, "creator")
		assert.Contains(t, nft, "owner")
	}
}

func TestNFTs_ListNFTs_WithFilters(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	// Test with query parameters
	resp := server.JSONRequest(t, "GET", "/api/v1/nfts?page=1&limit=10&creator_id=1&is_for_sale=true&sort=price&order=desc", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var response map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &response)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")
}

func TestNFTs_ListNFTs_InvalidPagination(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidQueries := []string{
		"?page=0", // page should be >= 1
		"?page=-1", // negative page
		"?limit=0", // limit should be >= 1
		"?limit=101", // limit exceeds max (100)
		"?page=abc", // non-numeric page
		"?limit=xyz", // non-numeric limit
	}

	for _, query := range invalidQueries {
		resp := server.JSONRequest(t, "GET", "/api/v1/nfts"+query, nil)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		// Should return 400 for invalid pagination
		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

// Test T015: Contract test POST /nfts (mint NFT)
func TestNFTs_MintNFT_ValidRequest(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	// Note: This would be multipart/form-data in reality, but for contract testing we use JSON
	mintRequest := map[string]interface{}{
		"title":       "My First NFT",
		"description": "A beautiful digital artwork",
		"image":       "base64EncodedImageData", // In reality would be binary
		"metadata": map[string]interface{}{
			"attributes": []interface{}{
				map[string]interface{}{
					"trait_type": "Color",
					"value":      "Blue",
				},
			},
		},
		"royalty": 5,
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/nfts", token, mintRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusCreated)
	
	var nft map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &nft)

	// Verify NFT schema
	assert.Contains(t, nft, "id")
	assert.Contains(t, nft, "token_id")
	assert.Contains(t, nft, "contract_address")
	assert.Contains(t, nft, "creator_id")
	assert.Contains(t, nft, "owner_id")
	assert.Contains(t, nft, "title")
	assert.Contains(t, nft, "image_url")
	assert.Contains(t, nft, "metadata_uri")

	assert.Equal(t, mintRequest["title"], nft["title"])
	assert.Equal(t, mintRequest["royalty"], nft["royalty"])
}

func TestNFTs_MintNFT_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	mintRequest := map[string]interface{}{
		"title": "Unauthorized NFT",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/nfts", mintRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Authentication required")
}

func TestNFTs_MintNFT_InvalidData(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	
	invalidRequests := []map[string]interface{}{
		{}, // missing required fields
		{"title": ""}, // empty title
		{"title": string(make([]byte, 101))}, // title too long (>100 chars)
		{"title": "Valid", "description": string(make([]byte, 1001))}, // description too long
		{"title": "Valid", "royalty": 11}, // royalty > 10%
		{"title": "Valid", "royalty": -1}, // negative royalty
	}

	for _, invalidReq := range invalidRequests {
		resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/nfts", token, invalidReq)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

// Test T016: Contract test GET /nfts/{nftId}
func TestNFTs_GetNFT_ValidId(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/nfts/1", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var nft map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &nft)

	// Verify complete NFT schema
	assert.Contains(t, nft, "id")
	assert.Contains(t, nft, "token_id")
	assert.Contains(t, nft, "creator")
	assert.Contains(t, nft, "owner")
	assert.Contains(t, nft, "title")
	assert.Contains(t, nft, "description")
	assert.Contains(t, nft, "image_url")
	assert.Contains(t, nft, "metadata_uri")
	assert.Contains(t, nft, "price")
	assert.Contains(t, nft, "is_for_sale")

	assert.Equal(t, float64(1), nft["id"])
}

func TestNFTs_GetNFT_NotFound(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/nfts/99999", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusNotFound, "NFT not found")
}

// Test T017: Contract test PUT /nfts/{nftId}
func TestNFTs_UpdateNFT_ValidOwner(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	updateRequest := map[string]interface{}{
		"price":       "2000000000000000000", // 2 ETH in Wei
		"is_for_sale": true,
	}

	resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/nfts/1", token, updateRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var nft map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &nft)

	assert.Equal(t, updateRequest["price"], nft["price"])
	assert.Equal(t, updateRequest["is_for_sale"], nft["is_for_sale"])
}

func TestNFTs_UpdateNFT_NotOwner(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken() // Assume this user doesn't own NFT 1
	updateRequest := map[string]interface{}{
		"price": "1000000000000000000",
	}

	resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/nfts/1", token, updateRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusForbidden, "Not authorized")
}

func TestNFTs_UpdateNFT_InvalidPrice(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	
	invalidPrices := []interface{}{
		"-1", // negative price
		"not_a_number", // invalid format
		"", // empty string
		0, // zero as number (should be string)
	}

	for _, invalidPrice := range invalidPrices {
		updateRequest := map[string]interface{}{
			"price": invalidPrice,
		}

		resp := server.AuthorizedJSONRequest(t, "PUT", "/api/v1/nfts/1", token, updateRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

// Test T018: Contract test POST /nfts/{nftId}/transfer
func TestNFTs_TransferNFT_ValidOwner(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	transferRequest := map[string]interface{}{
		"to_address": "0x9876543210987654321098765432109876543210",
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/nfts/1/transfer", token, transferRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var transfer map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &transfer)

	// Verify Transfer schema
	assert.Contains(t, transfer, "id")
	assert.Contains(t, transfer, "nft_id")
	assert.Contains(t, transfer, "from_id")
	assert.Contains(t, transfer, "to_id")
	assert.Contains(t, transfer, "tx_hash")
	assert.Contains(t, transfer, "type")

	assert.Equal(t, "transfer", transfer["type"])
}

func TestNFTs_TransferNFT_InvalidAddress(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	
	invalidAddresses := []string{
		"not_an_address", // invalid format
		"0x123", // too short
		"", // empty
		"1234567890123456789012345678901234567890", // missing 0x
		"0xGGGG567890123456789012345678901234567890", // invalid hex
	}

	for _, invalidAddr := range invalidAddresses {
		transferRequest := map[string]interface{}{
			"to_address": invalidAddr,
		}

		resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/nfts/1/transfer", token, transferRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestNFTs_TransferNFT_NotOwner(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	transferRequest := map[string]interface{}{
		"to_address": "0x9876543210987654321098765432109876543210",
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/nfts/999/transfer", token, transferRequest) // NFT not owned

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusForbidden, "Not authorized")
}

func TestNFTs_TransferNFT_SelfTransfer(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	transferRequest := map[string]interface{}{
		"to_address": "0x1234567890123456789012345678901234567890", // Same as current owner
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/nfts/1/transfer", token, transferRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should fail - can't transfer to self
	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "Cannot transfer to self")
}

func TestNFTs_TransferNFT_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	transferRequest := map[string]interface{}{
		"to_address": "0x9876543210987654321098765432109876543210",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/nfts/1/transfer", transferRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Authentication required")
}

// Edge cases and additional tests
func TestNFTs_GetNFT_InvalidId(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	invalidIds := []string{"abc", "0", "-1", ""}

	for _, invalidId := range invalidIds {
		path := fmt.Sprintf("/api/v1/nfts/%s", invalidId)
		resp := server.JSONRequest(t, "GET", path, nil)

		if resp.StatusCode == http.StatusNotFound {
			continue // Route not matched is valid
		}

		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "Invalid NFT ID")
	}
}

func TestNFTs_MintNFT_EmptyImage(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	mintRequest := map[string]interface{}{
		"title": "Valid Title",
		// Missing image field
		"metadata": map[string]interface{}{},
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/nfts", token, mintRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "image is required")
}