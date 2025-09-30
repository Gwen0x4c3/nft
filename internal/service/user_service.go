package service

import (
	"errors"
	"fmt"
	"time"

	"nft-platform/internal/models"
	"nft-platform/internal/repository"
	"nft-platform/pkg/auth"
	"nft-platform/pkg/crypto"
	"nft-platform/pkg/validation"
)

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *models.User `json:"user"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	WalletAddress string `json:"wallet_address" validate:"required,eth_addr"`
	Signature     string `json:"signature" validate:"required"`
	Message       string `json:"message" validate:"required"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username      string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email         string `json:"email" validate:"required,email"`
	WalletAddress string `json:"wallet_address" validate:"required,eth_addr"`
	Signature     string `json:"signature" validate:"required"`
	Message       string `json:"message" validate:"required"`
	Avatar        string `json:"avatar,omitempty" validate:"omitempty,url"`
	Bio           string `json:"bio,omitempty" validate:"omitempty,max=500"`
}

// UpdateUserRequest represents the user profile update request
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=30,alphanum"`
	Avatar   string `json:"avatar,omitempty" validate:"omitempty,url"`
	Bio      string `json:"bio,omitempty" validate:"omitempty,max=500"`
}

// RefreshTokenRequest represents the token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UserService handles user authentication and management operations
type UserService struct {
	userRepo     *repository.UserRepository
	validator    *validation.CustomValidator
	jwtManager   *auth.JWTManager
	signVerifier *crypto.SignatureVerifier
	messageGen   *crypto.MessageGenerator
}

// ServiceConfig represents configuration for the user service
type ServiceConfig struct {
	JWTManager   *auth.JWTManager
	SignVerifier *crypto.SignatureVerifier
	MessageGen   *crypto.MessageGenerator
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo *repository.UserRepository, config *ServiceConfig) *UserService {
	// Use provided dependencies or create default ones
	jwtManager := config.JWTManager
	if jwtManager == nil {
		var err error
		jwtManager, err = auth.NewJWTManagerWithKeys("nft-platform")
		if err != nil {
			panic(fmt.Sprintf("Failed to create JWT manager: %v", err))
		}
	}

	signVerifier := config.SignVerifier
	if signVerifier == nil {
		signVerifier = crypto.NewSignatureVerifier()
	}

	messageGen := config.MessageGen
	if messageGen == nil {
		messageGen = crypto.NewMessageGenerator()
	}

	return &UserService{
		userRepo:     userRepo,
		validator:    validation.NewCustomValidator(),
		jwtManager:   jwtManager,
		signVerifier: signVerifier,
		messageGen:   messageGen,
	}
}

// Register creates a new user account after validating wallet signature
func (s *UserService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify signature
	if err := s.signVerifier.VerifyEthereumSignature(req.WalletAddress, req.Signature, req.Message); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Check for existing users
	if exists, err := s.userRepo.ExistsByUsername(req.Username); err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	} else if exists {
		return nil, errors.New("username already exists")
	}

	if exists, err := s.userRepo.ExistsByEmail(req.Email); err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	} else if exists {
		return nil, errors.New("email already exists")
	}

	if exists, err := s.userRepo.ExistsByWalletAddress(req.WalletAddress); err != nil {
		return nil, fmt.Errorf("failed to check wallet address: %w", err)
	} else if exists {
		return nil, errors.New("wallet address already exists")
	}

	// Create new user
	user := &models.User{
		Username:   req.Username,
		Email:      req.Email,
		WalletAddr: req.WalletAddress,
		Avatar:     req.Avatar,
		Bio:        req.Bio,
		IsVerified: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// Login authenticates user with wallet signature
func (s *UserService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify signature
	if err := s.signVerifier.VerifyEthereumSignature(req.WalletAddress, req.Signature, req.Message); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Find user by wallet address
	user, err := s.userRepo.GetByWalletAddress(req.WalletAddress)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// RefreshToken generates new access token from refresh token
func (s *UserService) RefreshToken(req *RefreshTokenRequest) (*AuthResponse, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse and validate refresh token
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken, auth.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get current user data
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate new tokens
	return s.generateAuthResponse(user)
}

// GetProfile returns the user profile by ID
func (s *UserService) GetProfile(userID uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	return user, nil
}

// GetUserByID returns a user by ID (public endpoint)
func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(userID uint, req *UpdateUserRequest) (*models.User, error) {
	// Validate request
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get current user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check for username conflicts if updating username
	if req.Username != "" && req.Username != user.Username {
		if exists, err := s.userRepo.ExistsByUsername(req.Username); err != nil {
			return nil, fmt.Errorf("failed to check username: %w", err)
		} else if exists {
			return nil, errors.New("username already exists")
		}
		user.Username = req.Username
	}

	// Update optional fields
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return user, nil
}

// VerifyUser marks a user as verified (admin operation)
func (s *UserService) VerifyUser(userID uint) error {
	return s.userRepo.UpdateVerificationStatus(userID, true)
}

// UnverifyUser removes verification from a user (admin operation)
func (s *UserService) UnverifyUser(userID uint) error {
	return s.userRepo.UpdateVerificationStatus(userID, false)
}

// ValidateToken validates and parses a JWT token
func (s *UserService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString, auth.AccessToken)
}

// generateAuthResponse creates access and refresh tokens for a user
func (s *UserService) generateAuthResponse(user *models.User) (*AuthResponse, error) {
	tokenPair, err := s.jwtManager.GenerateTokenPair(
		user.ID,
		user.WalletAddr,
		user.Username,
		user.IsVerified,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         user,
	}, nil
}

// Helper methods for backwards compatibility and convenience

// ExportJWTPrivateKeyPEM exports JWT private key as PEM string
func (s *UserService) ExportJWTPrivateKeyPEM() string {
	return s.jwtManager.ExportPrivateKeyPEM()
}

// ExportJWTPublicKeyPEM exports JWT public key as PEM string
func (s *UserService) ExportJWTPublicKeyPEM() string {
	return s.jwtManager.ExportPublicKeyPEM()
}

// GenerateAuthMessage generates a standard authentication message
func (s *UserService) GenerateAuthMessage(action string) string {
	return s.messageGen.GenerateAuthMessage(action)
}

// Package-level utility functions for backwards compatibility

// GenerateNonce generates a cryptographically secure random nonce
func GenerateNonce() string {
	return crypto.GenerateNonce()
}

// GenerateAuthMessage generates a standard authentication message
func GenerateAuthMessage(action string) string {
	return crypto.GenerateAuthMessage(action)
}
