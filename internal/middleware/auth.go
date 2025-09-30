package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"nft-platform/pkg/auth"

	"github.com/gin-gonic/gin"
)

// UserContextKey is the key used to store user info in context
type UserContextKey string

const (
	UserKey UserContextKey = "user"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth middleware that requires valid JWT access token
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return a.authenticate(true)
}

// OptionalAuth middleware that validates JWT if present but doesn't require it
func (a *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return a.authenticate(false)
}

// authenticate is the core authentication logic
func (a *AuthMiddleware) authenticate(required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			if required {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Authorization header is required",
					"code":  "MISSING_AUTH_HEADER",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Check for Bearer token format
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			if required {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid authorization header format. Expected: Bearer <token>",
					"code":  "INVALID_AUTH_FORMAT",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		token := tokenParts[1]

		// Validate the JWT token
		claims, err := a.jwtManager.ValidateToken(token, auth.AccessToken)
		if err != nil {
			if required {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid or expired token",
					"code":  "INVALID_TOKEN",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Store user information in context
		c.Set(string(UserKey), claims)

		// Add user info to request context for gRPC calls
		ctx := context.WithValue(c.Request.Context(), UserKey, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetUserFromContext extracts user claims from Gin context
func GetUserFromContext(c *gin.Context) (*auth.Claims, bool) {
	user, exists := c.Get(string(UserKey))
	if !exists {
		return nil, false
	}

	claims, ok := user.(*auth.Claims)
	return claims, ok
}

// GetUserFromRequestContext extracts user claims from request context
func GetUserFromRequestContext(ctx context.Context) (*auth.Claims, bool) {
	user := ctx.Value(UserKey)
	if user == nil {
		return nil, false
	}

	claims, ok := user.(*auth.Claims)
	return claims, ok
}

// RequireVerifiedUser middleware that requires user to be verified
func (a *AuthMiddleware) RequireVerifiedUser() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User authentication required",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		if !user.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Verified user status required",
				"code":  "USER_NOT_VERIFIED",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// RequireUserOwnership middleware that checks if user owns the resource
func (a *AuthMiddleware) RequireUserOwnership(userIDParam string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User authentication required",
				"code":  "NOT_AUTHENTICATED",
			})
			c.Abort()
			return
		}

		// Get user ID from URL parameter
		userID := c.Param(userIDParam)
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User ID parameter is required",
				"code":  "MISSING_USER_ID",
			})
			c.Abort()
			return
		}

		// Check if the authenticated user matches the resource owner
		if userID != fmt.Sprintf("%d", user.UserID) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied: resource ownership required",
				"code":  "INSUFFICIENT_PERMISSIONS",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// RefreshTokenMiddleware validates refresh tokens
func (a *AuthMiddleware) RefreshTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		// Check for Bearer token format
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate the refresh token
		claims, err := a.jwtManager.ValidateToken(token, auth.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired refresh token",
				"code":  "INVALID_REFRESH_TOKEN",
			})
			c.Abort()
			return
		}

		// Store user information in context
		c.Set(string(UserKey), claims)

		// Add user info to request context
		ctx := context.WithValue(c.Request.Context(), UserKey, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
