package unit

import (
	"math/big"
	"testing"
	"time"

	"nft-platform/internal/models"
	"nft-platform/pkg/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBidValidation tests the Bid model validation
func TestBidValidation(t *testing.T) {
	validator := validation.NewCustomValidator()

	tests := []struct {
		name        string
		bid         models.Bid
		expectError bool
		errorField  string
	}{
		{
			name: "Valid bid",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(1000000000000000000), // 1 ETH in Wei
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "Valid bid with transaction hash",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(1000000000000000000),
				TxHash:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				Status:    "confirmed",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "Invalid status",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(1000000000000000000),
				Status:    "invalid_status",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorField:  "Status",
		},
		{
			name: "Nil amount",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    nil,
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
		},
		{
			name: "Zero amount",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(0),
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
		},
		{
			name: "Negative amount",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(-1000000000000000000),
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
		},
		{
			name: "Invalid transaction hash - too short",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(1000000000000000000),
				TxHash:    "0x123",
				Status:    "confirmed",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorField:  "TxHash",
		},
		{
			name: "Invalid transaction hash - non-hex",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(1000000000000000000),
				TxHash:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdeG",
				Status:    "confirmed",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: true,
			errorField:  "TxHash",
		},
		{
			name: "Valid transaction hash without 0x prefix",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(1000000000000000000),
				TxHash:    "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				Status:    "confirmed",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
		{
			name: "Empty transaction hash (should be valid for pending bids)",
			bid: models.Bid{
				AuctionID: 1,
				BidderID:  2,
				Amount:    big.NewInt(1000000000000000000),
				TxHash:    "",
				Status:    "pending",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(&tt.bid)

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

// TestBidHelperMethods tests bid helper methods
func TestBidHelperMethods(t *testing.T) {
	t.Run("GetAmountWei", func(t *testing.T) {
		bid := &models.Bid{
			Amount: big.NewInt(1000000000000000000),
		}

		result := bid.GetAmountWei()
		assert.Equal(t, "1000000000000000000", result)

		bid.Amount = nil
		result = bid.GetAmountWei()
		assert.Equal(t, "0", result)
	})

	t.Run("SetAmountWei", func(t *testing.T) {
		bid := &models.Bid{}

		err := bid.SetAmountWei("1000000000000000000")
		assert.NoError(t, err)
		assert.Equal(t, int64(1000000000000000000), bid.Amount.Int64())

		err = bid.SetAmountWei("invalid")
		assert.Error(t, err)

		err = bid.SetAmountWei("-100")
		assert.Error(t, err)

		err = bid.SetAmountWei("0")
		assert.Error(t, err)
	})

	t.Run("IsConfirmed", func(t *testing.T) {
		tests := []struct {
			name     string
			bid      models.Bid
			expected bool
		}{
			{
				name: "Confirmed bid",
				bid: models.Bid{
					Status: "confirmed",
				},
				expected: true,
			},
			{
				name: "Pending bid",
				bid: models.Bid{
					Status: "pending",
				},
				expected: false,
			},
			{
				name: "Failed bid",
				bid: models.Bid{
					Status: "failed",
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.bid.IsConfirmed()
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsPending", func(t *testing.T) {
		tests := []struct {
			name     string
			bid      models.Bid
			expected bool
		}{
			{
				name: "Pending bid",
				bid: models.Bid{
					Status: "pending",
				},
				expected: true,
			},
			{
				name: "Confirmed bid",
				bid: models.Bid{
					Status: "confirmed",
				},
				expected: false,
			},
			{
				name: "Failed bid",
				bid: models.Bid{
					Status: "failed",
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.bid.IsPending()
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("HasFailed", func(t *testing.T) {
		tests := []struct {
			name     string
			bid      models.Bid
			expected bool
		}{
			{
				name: "Failed bid",
				bid: models.Bid{
					Status: "failed",
				},
				expected: true,
			},
			{
				name: "Pending bid",
				bid: models.Bid{
					Status: "pending",
				},
				expected: false,
			},
			{
				name: "Confirmed bid",
				bid: models.Bid{
					Status: "confirmed",
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.bid.HasFailed()
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// TestBidAmountValidation tests bid amount validation logic
func TestBidAmountValidation(t *testing.T) {
	minimumIncrement := big.NewInt(10000000000000000) // 0.01 ETH in Wei

	tests := []struct {
		name             string
		bidAmount        string
		currentHighestBid string
		expectError      bool
	}{
		{
			name:             "Valid first bid",
			bidAmount:        "1000000000000000000", // 1 ETH
			currentHighestBid: "0",
			expectError:      false,
		},
		{
			name:             "Valid bid above minimum increment",
			bidAmount:        "1100000000000000000", // 1.1 ETH
			currentHighestBid: "1000000000000000000", // 1 ETH
			expectError:      false,
		},
		{
			name:             "Invalid bid below minimum increment",
			bidAmount:        "1005000000000000000", // 1.005 ETH
			currentHighestBid: "1000000000000000000", // 1 ETH
			expectError:      true,
		},
		{
			name:             "Invalid bid equal to current highest bid",
			bidAmount:        "1000000000000000000", // 1 ETH
			currentHighestBid: "1000000000000000000", // 1 ETH
			expectError:      true,
		},
		{
			name:             "Invalid bid below current highest bid",
			bidAmount:        "900000000000000000", // 0.9 ETH
			currentHighestBid: "1000000000000000000", // 1 ETH
			expectError:      true,
		},
		{
			name:             "Valid bid with no current highest bid",
			bidAmount:        "500000000000000000", // 0.5 ETH
			currentHighestBid: "0",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bid := &models.Bid{}
			err := bid.SetAmountWei(tt.bidAmount)
			require.NoError(t, err)

			var currentBid *big.Int
			if tt.currentHighestBid != "0" {
				currentBid = new(big.Int)
				_, ok := currentBid.SetString(tt.currentHighestBid, 10)
				require.True(t, ok)
			}

			err = bid.ValidateAmount(currentBid, minimumIncrement)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBidEdgeCases tests edge cases and boundary conditions
func TestBidEdgeCases(t *testing.T) {
	t.Run("Very large bid amounts", func(t *testing.T) {
		bid := &models.Bid{}
		veryLargeAmount := "1000000000000000000000000000000000000000000000000" // 1000 ETH in Wei

		err := bid.SetAmountWei(veryLargeAmount)
		assert.NoError(t, err)
		assert.Equal(t, veryLargeAmount, bid.GetAmountWei())
	})

	t.Run("Minimum bid amount (1 Wei)", func(t *testing.T) {
		bid := &models.Bid{}

		err := bid.SetAmountWei("1")
		assert.NoError(t, err)
		assert.Equal(t, "1", bid.GetAmountWei())
	})

	t.Run("Bid amount boundary with minimum increment", func(t *testing.T) {
		bid := &models.Bid{}
		minimumIncrement := big.NewInt(10000000000000000) // 0.01 ETH

		// Set current highest bid to 1 ETH
		currentBid := big.NewInt(1000000000000000000)
		minimumValidBid := new(big.Int).Add(currentBid, minimumIncrement)

		// Test exactly minimum valid bid
		err := bid.SetAmountWei(minimumValidBid.String())
		require.NoError(t, err)

		err = bid.ValidateAmount(currentBid, minimumIncrement)
		assert.NoError(t, err)

		// Test one wei below minimum
		invalidBid := new(big.Int).Sub(minimumValidBid, big.NewInt(1))
		err = bid.SetAmountWei(invalidBid.String())
		require.NoError(t, err)

		err = bid.ValidateAmount(currentBid, minimumIncrement)
		assert.Error(t, err)
	})

	t.Run("Bid validation with nil current bid", func(t *testing.T) {
		bid := &models.Bid{}
		minimumIncrement := big.NewInt(10000000000000000)

		err := bid.SetAmountWei("1000000000000000000")
		require.NoError(t, err)

		// Should pass validation even with nil current bid
		err = bid.ValidateAmount(nil, minimumIncrement)
		assert.NoError(t, err)
	})

	t.Run("Bid with zero current bid", func(t *testing.T) {
		bid := &models.Bid{}
		minimumIncrement := big.NewInt(10000000000000000)
		currentBid := big.NewInt(0)

		err := bid.SetAmountWei("500000000000000000")
		require.NoError(t, err)

		// Should pass validation when current bid is zero
		err = bid.ValidateAmount(currentBid, minimumIncrement)
		assert.NoError(t, err)
	})
}

// TestBidTransactionHashValidation tests transaction hash validation
func TestBidTransactionHashValidation(t *testing.T) {
	validator := validation.NewCustomValidator()

	tests := []struct {
		name        string
		txHash      string
		expectValid bool
	}{
		{
			name:        "Valid 32-byte hash with 0x prefix",
			txHash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectValid: true,
		},
		{
			name:        "Valid 32-byte hash without 0x prefix",
			txHash:      "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectValid: true,
		},
		{
			name:        "Valid hash with mixed case",
			txHash:      "0x1234567890AbCdEf1234567890AbCdEf1234567890AbCdEf1234567890AbCdEf",
			expectValid: true,
		},
		{
			name:        "Invalid - too short",
			txHash:      "0x1234567890abcdef",
			expectValid: false,
		},
		{
			name:        "Invalid - too long",
			txHash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			expectValid: false,
		},
		{
			name:        "Invalid - contains non-hex characters",
			txHash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdeG",
			expectValid: false,
		},
		{
			name:        "Empty string",
			txHash:      "",
			expectValid: true, // Should be valid for optional fields
		},
		{
			name:        "Only 0x prefix",
			txHash:      "0x",
			expectValid: false,
		},
		{
			name:        "Invalid - contains spaces",
			txHash:      "0x1234567890abcdef 1234567890abcdef1234567890abcdef1234567890abcdef",
			expectValid: false,
		},
		{
			name:        "Invalid - 31 bytes (62 hex chars)",
			txHash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Var(tt.txHash, "eth_tx_hash")

			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestBidStatusTransitions tests bid status transitions
func TestBidStatusTransitions(t *testing.T) {
	t.Run("Pending to confirmed transition", func(t *testing.T) {
		bid := &models.Bid{
			Status: "pending",
		}

		assert.True(t, bid.IsPending())
		assert.False(t, bid.IsConfirmed())
		assert.False(t, bid.HasFailed())

		// Simulate confirmation
		bid.Status = "confirmed"

		assert.False(t, bid.IsPending())
		assert.True(t, bid.IsConfirmed())
		assert.False(t, bid.HasFailed())
	})

	t.Run("Pending to failed transition", func(t *testing.T) {
		bid := &models.Bid{
			Status: "pending",
		}

		assert.True(t, bid.IsPending())
		assert.False(t, bid.IsConfirmed())
		assert.False(t, bid.HasFailed())

		// Simulate failure
		bid.Status = "failed"

		assert.False(t, bid.IsPending())
		assert.False(t, bid.IsConfirmed())
		assert.True(t, bid.HasFailed())
	})

	t.Run("Confirmed bid status", func(t *testing.T) {
		bid := &models.Bid{
			Status:    "confirmed",
			Amount:    big.NewInt(1000000000000000000),
			TxHash:    "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.False(t, bid.IsPending())
		assert.True(t, bid.IsConfirmed())
		assert.False(t, bid.HasFailed())
		assert.Equal(t, "1000000000000000000", bid.GetAmountWei())
	})

	t.Run("Failed bid status", func(t *testing.T) {
		bid := &models.Bid{
			Status:    "failed",
			Amount:    big.NewInt(1000000000000000000),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.False(t, bid.IsPending())
		assert.False(t, bid.IsConfirmed())
		assert.True(t, bid.HasFailed())
	})
}

// TestBidStringOperations tests string parsing and formatting
func TestBidStringOperations(t *testing.T) {
	t.Run("SetAmountWei with various formats", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
			valid    bool
		}{
			{"0", 0, false},        // Zero should be invalid
			{"1", 1, true},         // Minimum valid amount
			{"100", 100, true},
			{"1000000000000000000", 1000000000000000000, true}, // 1 ETH
			{"001", 1, true},        // Leading zeros should be handled
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				bid := &models.Bid{}
				err := bid.SetAmountWei(tt.input)

				if tt.valid {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected, bid.Amount.Int64())
				} else {
					assert.Error(t, err)
				}
			})
		}
	})

	t.Run("GetAmountWei consistency", func(t *testing.T) {
		testAmounts := []string{
			"1",
			"1000000000000000000", // 1 ETH
			"500000000000000000000", // 500 ETH
		}

		for _, amount := range testAmounts {
			t.Run(amount, func(t *testing.T) {
				bid := &models.Bid{}
				err := bid.SetAmountWei(amount)
				require.NoError(t, err)

				retrievedAmount := bid.GetAmountWei()
				assert.Equal(t, amount, retrievedAmount)
			})
		}
	})
}