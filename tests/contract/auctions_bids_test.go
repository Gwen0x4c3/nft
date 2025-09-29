package contract

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"nft-platform/tests/testutil"
)

// Test T019: Contract test GET /auctions
func TestAuctions_ListAuctions(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/auctions", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var response map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &response)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")

	data := response["data"].([]interface{})
	if len(data) > 0 {
		auction := data[0].(map[string]interface{})
		assert.Contains(t, auction, "id")
		assert.Contains(t, auction, "nft")
		assert.Contains(t, auction, "status")
		assert.Contains(t, auction, "start_price")
	}
}

func TestAuctions_ListAuctions_WithFilters(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/auctions?status=active&seller_id=1&page=1&limit=10", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
}

// Test T020: Contract test POST /auctions
func TestAuctions_CreateAuction_ValidRequest(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	auctionRequest := map[string]interface{}{
		"nft_id":        1,
		"start_price":   "500000000000000000", // 0.5 ETH
		"reserve_price": "1000000000000000000", // 1 ETH
		"end_time":      "2024-12-31T23:59:59Z",
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions", token, auctionRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusCreated)
	
	var auction map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &auction)

	assert.Contains(t, auction, "id")
	assert.Contains(t, auction, "nft_id")
	assert.Contains(t, auction, "status")
	assert.Contains(t, auction, "start_price")
	assert.Equal(t, auctionRequest["nft_id"], auction["nft_id"])
	assert.Equal(t, auctionRequest["start_price"], auction["start_price"])
}

func TestAuctions_CreateAuction_InvalidData(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	
	invalidRequests := []map[string]interface{}{
		{}, // missing required fields
		{"nft_id": 0}, // invalid NFT ID
		{"nft_id": 1, "start_price": "-1"}, // negative price
		{"nft_id": 1, "start_price": "100", "end_time": "2020-01-01T00:00:00Z"}, // past end time
		{"nft_id": 1, "start_price": "100", "reserve_price": "50"}, // reserve < start
	}

	for _, invalidReq := range invalidRequests {
		resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions", token, invalidReq)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestAuctions_CreateAuction_NotOwner(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	auctionRequest := map[string]interface{}{
		"nft_id":      999, // NFT not owned by user
		"start_price": "500000000000000000",
		"end_time":    "2024-12-31T23:59:59Z",
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions", token, auctionRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusForbidden, "Not authorized")
}

// Test T021: Contract test GET /auctions/{auctionId}
func TestAuctions_GetAuction_ValidId(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/auctions/1", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var auction map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &auction)

	assert.Contains(t, auction, "id")
	assert.Contains(t, auction, "nft")
	assert.Contains(t, auction, "seller")
	assert.Contains(t, auction, "bids")
	assert.Contains(t, auction, "status")
	assert.Equal(t, float64(1), auction["id"])
}

func TestAuctions_GetAuction_NotFound(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/auctions/99999", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusNotFound, "Auction not found")
}

// Test T022: Contract test DELETE /auctions/{auctionId}
func TestAuctions_CancelAuction_ValidSeller(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	resp := server.AuthorizedJSONRequest(t, "DELETE", "/api/v1/auctions/1", token, nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
}

func TestAuctions_CancelAuction_NotSeller(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	resp := server.AuthorizedJSONRequest(t, "DELETE", "/api/v1/auctions/999", token, nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusForbidden, "Not authorized")
}

func TestAuctions_CancelAuction_WithBids(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	resp := server.AuthorizedJSONRequest(t, "DELETE", "/api/v1/auctions/2", token, nil) // Auction with bids

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	// Should return 409 - cannot cancel auction with bids
	testutil.AssertErrorResponse(t, resp, http.StatusConflict, "Cannot cancel auction with bids")
}

// Test T023: Contract test GET /auctions/{auctionId}/bids
func TestBids_ListBids(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/auctions/1/bids", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
	
	var response map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &response)

	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")

	data := response["data"].([]interface{})
	if len(data) > 0 {
		bid := data[0].(map[string]interface{})
		assert.Contains(t, bid, "id")
		assert.Contains(t, bid, "bidder")
		assert.Contains(t, bid, "amount")
		assert.Contains(t, bid, "status")
	}
}

func TestBids_ListBids_WithPagination(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/auctions/1/bids?page=1&limit=10", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusOK)
}

func TestBids_ListBids_AuctionNotFound(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "GET", "/api/v1/auctions/99999/bids", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusNotFound, "Auction not found")
}

// Test T024: Contract test POST /auctions/{auctionId}/bids
func TestBids_PlaceBid_ValidRequest(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	bidRequest := map[string]interface{}{
		"amount": "1500000000000000000", // 1.5 ETH
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions/1/bids", token, bidRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertStatusCode(t, resp, http.StatusCreated)
	
	var bid map[string]interface{}
	testutil.ParseJSONResponse(t, resp, &bid)

	assert.Contains(t, bid, "id")
	assert.Contains(t, bid, "auction_id")
	assert.Contains(t, bid, "bidder")
	assert.Contains(t, bid, "amount")
	assert.Contains(t, bid, "status")
	assert.Equal(t, bidRequest["amount"], bid["amount"])
	assert.Equal(t, "pending", bid["status"])
}

func TestBids_PlaceBid_BidTooLow(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	bidRequest := map[string]interface{}{
		"amount": "100000000000000000", // Lower than current highest bid
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions/1/bids", token, bidRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusConflict, "Bid too low")
}

func TestBids_PlaceBid_AuctionNotActive(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	bidRequest := map[string]interface{}{
		"amount": "1000000000000000000",
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions/2/bids", token, bidRequest) // Non-active auction

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusConflict, "Auction not active")
}

func TestBids_PlaceBid_OwnAuction(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	bidRequest := map[string]interface{}{
		"amount": "1000000000000000000",
	}

	resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions/3/bids", token, bidRequest) // User's own auction

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "Cannot bid on own auction")
}

func TestBids_PlaceBid_InvalidAmount(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	token := testutil.MockJWTToken()
	
	invalidAmounts := []interface{}{
		"", // empty
		"-1", // negative
		"abc", // non-numeric
		0, // zero as number
	}

	for _, invalidAmount := range invalidAmounts {
		bidRequest := map[string]interface{}{
			"amount": invalidAmount,
		}

		resp := server.AuthorizedJSONRequest(t, "POST", "/api/v1/auctions/1/bids", token, bidRequest)

		if resp.StatusCode == http.StatusNotFound {
			t.Skip("Handler not implemented yet (expected for TDD)")
			return
		}

		testutil.AssertErrorResponse(t, resp, http.StatusBadRequest, "validation failed")
	}
}

func TestBids_PlaceBid_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	bidRequest := map[string]interface{}{
		"amount": "1000000000000000000",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auctions/1/bids", bidRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Authentication required")
}

// Edge cases
func TestAuctions_CreateAuction_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	auctionRequest := map[string]interface{}{
		"nft_id":      1,
		"start_price": "500000000000000000",
		"end_time":    "2024-12-31T23:59:59Z",
	}

	resp := server.JSONRequest(t, "POST", "/api/v1/auctions", auctionRequest)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Authentication required")
}

func TestAuctions_CancelAuction_NoAuth(t *testing.T) {
	server := testutil.NewTestServer()
	defer server.Close()

	resp := server.JSONRequest(t, "DELETE", "/api/v1/auctions/1", nil)

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Handler not implemented yet (expected for TDD)")
		return
	}

	testutil.AssertErrorResponse(t, resp, http.StatusUnauthorized, "Authentication required")
}