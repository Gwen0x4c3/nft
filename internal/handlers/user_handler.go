package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nft-platform/internal/service"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Retrieves the authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User "User profile"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	// Convert to uint
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get user profile from service
	user, err := h.userService.GetUserByID(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "fetch_failed",
			Message: "Failed to retrieve user profile",
			Code:    "FETCH_FAILED",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Return user profile (password will be excluded by JSON tags)
	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Updates the authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.UpdateUserRequest true "Profile update request"
// @Success 200 {object} models.User "Updated user profile"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Bind request body
	var req service.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid request payload",
			Code:    "BAD_REQUEST",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Update user profile
	user, err := h.userService.UpdateProfile(uid, &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorCode := "UPDATE_FAILED"
		message := err.Error()

		// Handle specific error cases
		if contains(message, "validation") {
			statusCode = http.StatusBadRequest
			errorCode = "VALIDATION_FAILED"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Retrieves a user's public profile information by user ID
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} models.User "User profile"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users/{userId} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	// Parse user ID from URL parameter
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_parameter",
			Message: "Invalid user ID",
			Code:    "INVALID_PARAMETER",
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Get user from service
	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		statusCode := http.StatusNotFound
		errorCode := "USER_NOT_FOUND"
		message := "User not found"

		if !contains(err.Error(), "not found") {
			statusCode = http.StatusInternalServerError
			errorCode = "FETCH_FAILED"
			message = "Failed to retrieve user"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorCode,
			Message: message,
			Code:    errorCode,
			Details: map[string]interface{}{"error": err.Error()},
		})
		return
	}

	// Return user profile
	c.JSON(http.StatusOK, user)
}
