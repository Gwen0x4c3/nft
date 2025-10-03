package unit

import (
	"testing"

	"nft-platform/internal/models"
	"nft-platform/internal/service"
	"nft-platform/pkg/auth"
	"nft-platform/pkg/crypto"
	"nft-platform/pkg/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByWalletAddress(address string) (*models.User, error) {
	args := m.Called(address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByWalletAddress(address string) (bool, error) {
	args := m.Called(address)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UpdateVerificationStatus(id uint, verified bool) error {
	args := m.Called(id, verified)
	return args.Error(0)
}

// MockSignatureVerifier is a mock implementation of SignatureVerifier
type MockSignatureVerifier struct {
	mock.Mock
}

func (m *MockSignatureVerifier) VerifyEthereumSignature(address, signature, message string) error {
	args := m.Called(address, signature, message)
	return args.Error(0)
}

// TestUserValidation tests the User model validation
func TestUserValidation(t *testing.T) {
	validator := validation.NewCustomValidator()

	tests := []struct {
		name        string
		user        models.User
		expectError bool
		errorField  string
	}{
		{
			name: "Valid user",
			user: models.User{
				Username:   "testuser",
				Email:      "test@example.com",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
				Avatar:     "https://example.com/avatar.jpg",
				Bio:        "Test user bio",
			},
			expectError: false,
		},
		{
			name: "Invalid username - too short",
			user: models.User{
				Username:   "ab",
				Email:      "test@example.com",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			},
			expectError: true,
			errorField:  "Username",
		},
		{
			name: "Invalid username - too long",
			user: models.User{
				Username:   "thisusernameistoolongforvalidation",
				Email:      "test@example.com",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			},
			expectError: true,
			errorField:  "Username",
		},
		{
			name: "Invalid username - special characters",
			user: models.User{
				Username:   "test@user",
				Email:      "test@example.com",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			},
			expectError: true,
			errorField:  "Username",
		},
		{
			name: "Invalid email format",
			user: models.User{
				Username:   "testuser",
				Email:      "invalid-email",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			},
			expectError: true,
			errorField:  "Email",
		},
		{
			name: "Invalid wallet address - wrong format",
			user: models.User{
				Username:   "testuser",
				Email:      "test@example.com",
				WalletAddr: "invalid_address",
			},
			expectError: true,
			errorField:  "WalletAddr",
		},
		{
			name: "Invalid wallet address - too short",
			user: models.User{
				Username:   "testuser",
				Email:      "test@example.com",
				WalletAddr: "0x123",
			},
			expectError: true,
			errorField:  "WalletAddr",
		},
		{
			name: "Invalid avatar URL",
			user: models.User{
				Username:   "testuser",
				Email:      "test@example.com",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
				Avatar:     "not-a-url",
			},
			expectError: true,
			errorField:  "Avatar",
		},
		{
			name: "Invalid bio - too long",
			user: models.User{
				Username:   "testuser",
				Email:      "test@example.com",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
				Bio:        string(make([]byte, 501)), // 501 characters
			},
			expectError: true,
			errorField:  "Bio",
		},
		{
			name: "Valid user with optional fields empty",
			user: models.User{
				Username:   "testuser",
				Email:      "test@example.com",
				WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
				Avatar:     "",
				Bio:        "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Struct(&tt.user)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorField)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestUserServiceValidation tests user service request validation
func TestUserServiceValidation(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockSignVerifier := new(MockSignatureVerifier)

	jwtManager, err := auth.NewJWTManagerWithKeys("nft-platform")
	require.NoError(t, err)

	messageGen := crypto.NewMessageGenerator()

	config := &service.ServiceConfig{
		JWTManager:   jwtManager,
		SignVerifier: mockSignVerifier,
		MessageGen:   messageGen,
	}

	userService := service.NewUserService(mockRepo, config)

	t.Run("Valid register request", func(t *testing.T) {
		req := &service.RegisterRequest{
			Username:      "testuser",
			Email:         "test@example.com",
			WalletAddress: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			Signature:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Message:       "Sign this message to authenticate",
			Avatar:        "https://example.com/avatar.jpg",
			Bio:           "Test user bio",
		}

		// Setup mocks
		mockRepo.On("ExistsByUsername", req.Username).Return(false, nil)
		mockRepo.On("ExistsByEmail", req.Email).Return(false, nil)
		mockRepo.On("ExistsByWalletAddress", req.WalletAddress).Return(false, nil)
		mockSignVerifier.On("VerifyEthereumSignature", req.WalletAddress, req.Signature, req.Message).Return(nil)
		mockRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)

		resp, err := userService.Register(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.NotNil(t, resp.User)

		mockRepo.AssertExpectations(t)
		mockSignVerifier.AssertExpectations(t)
	})

	t.Run("Invalid register request - username too short", func(t *testing.T) {
		req := &service.RegisterRequest{
			Username:      "ab",
			Email:         "test@example.com",
			WalletAddress: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			Signature:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Message:       "Sign this message to authenticate",
		}

		resp, err := userService.Register(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Invalid register request - invalid email", func(t *testing.T) {
		req := &service.RegisterRequest{
			Username:      "testuser",
			Email:         "invalid-email",
			WalletAddress: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			Signature:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Message:       "Sign this message to authenticate",
		}

		resp, err := userService.Register(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Invalid register request - invalid wallet address", func(t *testing.T) {
		req := &service.RegisterRequest{
			Username:      "testuser",
			Email:         "test@example.com",
			WalletAddress: "invalid-address",
			Signature:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Message:       "Sign this message to authenticate",
		}

		resp, err := userService.Register(req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Valid login request", func(t *testing.T) {
		req := &service.LoginRequest{
			WalletAddress: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
			Signature:     "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			Message:       "Sign this message to authenticate",
		}

		user := &models.User{
			ID:         1,
			Username:   "testuser",
			Email:      "test@example.com",
			WalletAddr: req.WalletAddress,
		}

		// Setup mocks
		mockSignVerifier.On("VerifyEthereumSignature", req.WalletAddress, req.Signature, req.Message).Return(nil)
		mockRepo.On("GetByWalletAddress", req.WalletAddress).Return(user, nil)

		resp, err := userService.Login(req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)

		mockRepo.AssertExpectations(t)
		mockSignVerifier.AssertExpectations(t)
	})

	t.Run("Valid update user request", func(t *testing.T) {
		userID := uint(1)
		req := &service.UpdateUserRequest{
			Username: "newusername",
			Avatar:   "https://example.com/new-avatar.jpg",
			Bio:      "Updated bio",
		}

		user := &models.User{
			ID:         userID,
			Username:   "oldusername",
			Email:      "test@example.com",
			WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
		}

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(user, nil)
		mockRepo.On("ExistsByUsername", req.Username).Return(false, nil)
		mockRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

		updatedUser, err := userService.UpdateProfile(userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, req.Username, updatedUser.Username)
		assert.Equal(t, req.Avatar, updatedUser.Avatar)
		assert.Equal(t, req.Bio, updatedUser.Bio)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid update user request - username too short", func(t *testing.T) {
		userID := uint(1)
		req := &service.UpdateUserRequest{
			Username: "ab",
		}

		user := &models.User{
			ID:         userID,
			Username:   "oldusername",
			Email:      "test@example.com",
			WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
		}

		mockRepo.On("GetByID", userID).Return(user, nil)

		updatedUser, err := userService.UpdateProfile(userID, req)

		assert.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.Contains(t, err.Error(), "validation failed")

		mockRepo.AssertExpectations(t)
	})
}

// TestCustomValidatorFunctions tests the custom validation functions
func TestCustomValidatorFunctions(t *testing.T) {
	t.Run("IsValidEthereumAddress", func(t *testing.T) {
		tests := []struct {
			name     string
			address  string
			expected bool
		}{
			{"Valid address", "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D", true},
			{"Valid address with checksum", "0x742D35Cc6634C0532925A3b8D4E7E0e0e9e0dF1D", true},
			{"Invalid address - too short", "0x123", false},
			{"Invalid address - non-hex", "0x742d35Cc6634C0532925a3b8D4E7E0G0e9e0dF1D", false},
			{"Invalid address - missing 0x", "742d35Cc6634C0532925a3b8D4E7E0e9e0dF1D", false},
			{"Empty address", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validation.IsValidEthereumAddress(tt.address)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsValidUsername", func(t *testing.T) {
		tests := []struct {
			name     string
			username string
			expected bool
		}{
			{"Valid username", "testuser", true},
			{"Valid username with underscore", "test_user", true},
			{"Valid username with dash", "test-user", true},
			{"Valid username with numbers", "test123", true},
			{"Too short", "ab", false},
			{"Too long", "thisusernameistoolongforvalidation", false},
			{"Starts with special char", "_testuser", false},
			{"Ends with special char", "testuser_", false},
			{"Consecutive special chars", "test__user", false},
			{"Contains special char", "test@user", false},
			{"Empty", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validation.IsValidUsername(tt.username)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsValidEmail", func(t *testing.T) {
		tests := []struct {
			name     string
			email    string
			expected bool
		}{
			{"Valid email", "test@example.com", true},
			{"Valid email with subdomain", "test@mail.example.com", true},
			{"Valid email with numbers", "test123@example.com", true},
			{"Invalid email - no @", "testexample.com", false},
			{"Invalid email - no domain", "test@", false},
			{"Invalid email - no user", "@example.com", false},
			{"Invalid email - special chars", "test@exa$mple.com", false},
			{"Empty", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validation.IsValidEmail(tt.email)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsValidWalletSignature", func(t *testing.T) {
		tests := []struct {
			name      string
			signature string
			expected  bool
		}{
			{"Valid signature with 0x", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", true},
			{"Valid signature without 0x", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", true},
			{"Invalid signature - too short", "0x123", false},
			{"Invalid signature - too long", "0x" + string(make([]byte, 140)), false},
			{"Invalid signature - non-hex", "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdeG", false},
			{"Empty", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validation.IsValidWalletSignature(tt.signature)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsValidURL", func(t *testing.T) {
		tests := []struct {
			name     string
			url      string
			expected bool
		}{
			{"Valid HTTPS URL", "https://example.com", true},
			{"Valid HTTP URL", "http://example.com", true},
			{"Valid URL with path", "https://example.com/path/to/image.jpg", true},
			{"Invalid URL - no protocol", "example.com", false},
			{"Invalid URL - no domain", "https://", false},
			{"Empty", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := validation.IsValidURL(tt.url)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

// TestUserEdgeCases tests edge cases and boundary conditions
func TestUserEdgeCases(t *testing.T) {
	validator := validation.NewCustomValidator()

	t.Run("Username boundary tests", func(t *testing.T) {
		// Test minimum length (3 characters)
		user := models.User{
			Username:   "abc",
			Email:      "test@example.com",
			WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
		}
		err := validator.Struct(&user)
		assert.NoError(t, err)

		// Test maximum length (30 characters)
		user.Username = "abcdefghijklmnopqrstuvwxyzab"
		err = validator.Struct(&user)
		assert.NoError(t, err)

		// Test just over maximum length (31 characters)
		user.Username = "abcdefghijklmnopqrstuvwxyzabc"
		err = validator.Struct(&user)
		assert.Error(t, err)
	})

	t.Run("Bio boundary tests", func(t *testing.T) {
		user := models.User{
			Username:   "testuser",
			Email:      "test@example.com",
			WalletAddr: "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
		}

		// Test maximum length (500 characters)
		user.Bio = string(make([]byte, 500))
		for i := range user.Bio {
			user.Bio = user.Bio[:i] + "a" + user.Bio[i+1:]
		}
		err := validator.Struct(&user)
		assert.NoError(t, err)

		// Test just over maximum length (501 characters)
		user.Bio = string(make([]byte, 501))
		for i := range user.Bio {
			user.Bio = user.Bio[:i] + "a" + user.Bio[i+1:]
		}
		err = validator.Struct(&user)
		assert.Error(t, err)
	})

	t.Run("Wallet address case sensitivity", func(t *testing.T) {
		user := models.User{
			Username:   "testuser",
			Email:      "test@example.com",
			WalletAddr: "0x742D35Cc6634C0532925A3b8D4E7E0e0e9e0dF1D", // Mixed case
		}
		err := validator.Struct(&user)
		assert.NoError(t, err)
	})
}

