package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSConfig defines CORS configuration options
type CORSConfig struct {
	// AllowOrigins is a list of origins a cross-domain request can be executed from.
	// If the special "*" value is present in the list, all origins will be allowed.
	// Default value is []
	AllowOrigins []string

	// AllowOriginFunc is a custom function to validate the origin. It takes the origin
	// as an argument and returns true if allowed or false otherwise. If this option is
	// set, the value of AllowOrigins is ignored.
	AllowOriginFunc func(origin string) bool

	// AllowMethods is a list of methods the client is allowed to use with
	// cross-domain requests. Default value is simple methods (GET and POST)
	AllowMethods []string

	// AllowHeaders is list of non simple headers the client is allowed to use with
	// cross-domain requests.
	AllowHeaders []string

	// ExposeHeaders indicates which headers are safe to expose to the API of a CORS
	// API specification
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached
	MaxAge time.Duration
}

// DefaultCORSConfig returns a default CORS configuration for development
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001", "http://127.0.0.1:3000"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
			"Connection",
			"Host",
			"Referer",
			"User-Agent",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Retry-After",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// ProductionCORSConfig returns a CORS configuration suitable for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Accept-Encoding",
			"Accept-Language",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           24 * time.Hour,
	}
}

// StrictCORSConfig returns a strict CORS configuration with minimal permissions
func StrictCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Content-Type",
			"Authorization",
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           1 * time.Hour,
	}
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			handlePreflight(c, config, origin)
			return
		}

		// Handle actual requests
		handleCORS(c, config, origin)
		c.Next()
	}
}

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// handlePreflight handles OPTIONS preflight requests
func handlePreflight(c *gin.Context, config CORSConfig, origin string) {
	if !isOriginAllowed(origin, config) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Set Access-Control-Allow-Origin
	if config.AllowCredentials && origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
	} else if len(config.AllowOrigins) > 0 && config.AllowOrigins[0] == "*" {
		c.Header("Access-Control-Allow-Origin", "*")
	} else if origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
	}

	// Set Access-Control-Allow-Methods
	if len(config.AllowMethods) > 0 {
		c.Header("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ","))
	}

	// Set Access-Control-Allow-Headers
	requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
	if requestHeaders != "" && len(config.AllowHeaders) > 0 {
		allowedHeaders := filterRequestedHeaders(requestHeaders, config.AllowHeaders)
		if len(allowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ","))
		}
	} else if len(config.AllowHeaders) > 0 {
		c.Header("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ","))
	}

	// Set Access-Control-Allow-Credentials
	if config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// Set Access-Control-Max-Age
	if config.MaxAge > 0 {
		c.Header("Access-Control-Max-Age", strconv.Itoa(int(config.MaxAge.Seconds())))
	}

	c.Status(http.StatusOK)
}

// handleCORS handles actual CORS requests
func handleCORS(c *gin.Context, config CORSConfig, origin string) {
	if !isOriginAllowed(origin, config) {
		// Don't abort, just don't set CORS headers
		return
	}

	// Set Access-Control-Allow-Origin
	if config.AllowCredentials && origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
	} else if len(config.AllowOrigins) > 0 && config.AllowOrigins[0] == "*" {
		c.Header("Access-Control-Allow-Origin", "*")
	} else if origin != "" {
		c.Header("Access-Control-Allow-Origin", origin)
	}

	// Set Access-Control-Expose-Headers
	if len(config.ExposeHeaders) > 0 {
		c.Header("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ","))
	}

	// Set Access-Control-Allow-Credentials
	if config.AllowCredentials {
		c.Header("Access-Control-Allow-Credentials", "true")
	}

	// Set Vary header to indicate that the response varies based on the Origin header
	varyHeader := c.GetHeader("Vary")
	if varyHeader == "" {
		c.Header("Vary", "Origin")
	} else if !strings.Contains(varyHeader, "Origin") {
		c.Header("Vary", varyHeader+", Origin")
	}
}

// isOriginAllowed checks if the origin is allowed
func isOriginAllowed(origin string, config CORSConfig) bool {
	if origin == "" {
		return true // Same-origin requests
	}

	// Use custom function if provided
	if config.AllowOriginFunc != nil {
		return config.AllowOriginFunc(origin)
	}

	// Check wildcard
	if len(config.AllowOrigins) > 0 && config.AllowOrigins[0] == "*" {
		return true
	}

	// Check exact matches
	for _, allowedOrigin := range config.AllowOrigins {
		if allowedOrigin == origin {
			return true
		}
		
		// Support for wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// filterRequestedHeaders filters requested headers against allowed headers
func filterRequestedHeaders(requestHeaders string, allowedHeaders []string) []string {
	if requestHeaders == "" {
		return nil
	}

	requested := strings.Split(requestHeaders, ",")
	var allowed []string

	for _, header := range requested {
		header = strings.TrimSpace(header)
		header = strings.ToLower(header)
		
		for _, allowedHeader := range allowedHeaders {
			if strings.ToLower(allowedHeader) == header {
				allowed = append(allowed, allowedHeader)
				break
			}
		}
	}

	return allowed
}

// SecurityHeaders adds security-related headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Strict transport security (HTTPS only)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Content security policy (basic policy)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' wss: ws:")
		
		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions policy
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// AllowAllCORS allows all origins (for development only)
func AllowAllCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			return
		}

		c.Next()
	}
}

// DynamicCORS creates a CORS middleware that can dynamically determine allowed origins
func DynamicCORS(allowOriginFunc func(origin string) bool) gin.HandlerFunc {
	config := CORSConfig{
		AllowOriginFunc: allowOriginFunc,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	return CORSWithConfig(config)
}
