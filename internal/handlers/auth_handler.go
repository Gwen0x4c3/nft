package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nft-platform/internal/service"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Login godoc
// @Summary User login with wallet signature
// @Description Authenticates a user using wallet signature verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "Login request payload"
// @Success 200 {object} service.AuthResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid signature or credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Call user service to authenticate
	authResp, err := h.userService.Login(&req)
	if err != nil {
		// Check error type and return appropriate status
		statusCode := http.StatusUnauthorized
		errorCode := "INVALID_CREDENTIALS"
		message := err.Error()

		// You can add more specific error handling here based on error types
		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
		})
		return
	}

	// Return successful authentication response
	c.JSON(http.StatusOK, authResp)
}

// Register godoc
// @Summary Register new user
// @Description Creates a new user account after validating wallet signature
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "Registration request payload"
// @Success 201 {object} service.AuthResponse "User registered successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Call user service to register
	authResp, err := h.userService.Register(&req)
	if err != nil {
		statusCode := http.StatusBadRequest
		errorCode := "REGISTRATION_FAILED"
		message := err.Error()

		// Check for conflict errors (user already exists)
		if contains(message, "already exists") || contains(message, "duplicate") {
			statusCode = http.StatusConflict
			errorCode = "USER_ALREADY_EXISTS"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
		})
		return
	}

	// Return successful registration response
	c.JSON(http.StatusCreated, authResp)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Generates a new access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body service.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} service.AuthResponse "Token refreshed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid or expired refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req service.RefreshTokenRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Call user service to refresh token
	authResp, err := h.userService.RefreshToken(&req)
	if err != nil {
		statusCode := http.StatusUnauthorized
		errorCode := "INVALID_REFRESH_TOKEN"
		message := err.Error()

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
		})
		return
	}

	// Return new token pair
	c.JSON(http.StatusOK, authResp)
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    string                 `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Helper function to check if string contains substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(str) > len(substr) && (str[:len(substr)] == substr || str[len(str)-len(substr):] == substr || containsHelper(str, substr)))
}

func containsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
