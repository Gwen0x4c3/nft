package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	Requests       int                       // Number of requests allowed
	Duration       time.Duration             // Time window duration
	KeyGenerator   func(*gin.Context) string // Function to generate rate limit key
	OnLimitReached func(*gin.Context)        // Custom handler when limit is reached
	SkipOnError    bool                      // Whether to skip rate limiting on Redis errors
}

// RateLimitMiddleware provides rate limiting functionality using Redis
type RateLimitMiddleware struct {
	redisClient *redis.Client
}

// NewRateLimitMiddleware creates a new rate limit middleware instance
func NewRateLimitMiddleware(redisClient *redis.Client) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		redisClient: redisClient,
	}
}

// DefaultRateLimit provides a default rate limiting configuration
// 100 requests per minute per IP
func (r *RateLimitMiddleware) DefaultRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: 100,
		Duration: time.Minute,
		KeyGenerator: func(c *gin.Context) string {
			return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		},
	}
	return r.RateLimit(config)
}

// UserRateLimit provides rate limiting per authenticated user
// 1000 requests per minute per user
func (r *RateLimitMiddleware) UserRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: 1000,
		Duration: time.Minute,
		KeyGenerator: func(c *gin.Context) string {
			// Try to get user from context first
			if user, exists := GetUserFromContext(c); exists {
				return fmt.Sprintf("rate_limit:user:%d", user.UserID)
			}
			// Fall back to IP-based rate limiting for unauthenticated users
			return fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		},
	}
	return r.RateLimit(config)
}

// StrictRateLimit provides strict rate limiting for sensitive operations
// 10 requests per minute per user
func (r *RateLimitMiddleware) StrictRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: 10,
		Duration: time.Minute,
		KeyGenerator: func(c *gin.Context) string {
			if user, exists := GetUserFromContext(c); exists {
				return fmt.Sprintf("rate_limit:strict:user:%d", user.UserID)
			}
			return fmt.Sprintf("rate_limit:strict:ip:%s", c.ClientIP())
		},
	}
	return r.RateLimit(config)
}

// BlockchainRateLimit provides rate limiting for blockchain operations
// 10 transactions per minute per user to prevent spam and gas costs
func (r *RateLimitMiddleware) BlockchainRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: 10,
		Duration: time.Minute,
		KeyGenerator: func(c *gin.Context) string {
			if user, exists := GetUserFromContext(c); exists {
				return fmt.Sprintf("rate_limit:blockchain:user:%d", user.UserID)
			}
			return fmt.Sprintf("rate_limit:blockchain:ip:%s", c.ClientIP())
		},
		OnLimitReached: func(c *gin.Context) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Blockchain operation rate limit exceeded. Please wait before trying again.",
				"code":        "BLOCKCHAIN_RATE_LIMIT_EXCEEDED",
				"retry_after": 60,
			})
		},
	}
	return r.RateLimit(config)
}

// WebSocketRateLimit provides rate limiting for WebSocket connections
// 100 messages per minute per connection
func (r *RateLimitMiddleware) WebSocketRateLimit() gin.HandlerFunc {
	config := RateLimitConfig{
		Requests: 100,
		Duration: time.Minute,
		KeyGenerator: func(c *gin.Context) string {
			if user, exists := GetUserFromContext(c); exists {
				return fmt.Sprintf("rate_limit:websocket:user:%d", user.UserID)
			}
			return fmt.Sprintf("rate_limit:websocket:ip:%s", c.ClientIP())
		},
	}
	return r.RateLimit(config)
}

// RateLimit implements sliding window rate limiting using Redis
func (r *RateLimitMiddleware) RateLimit(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate rate limit key
		key := config.KeyGenerator(c)

		// Check current request count using sliding window
		allowed, remaining, resetTime, err := r.checkRateLimit(key, config.Requests, config.Duration)

		if err != nil {
			// Log error and decide whether to skip or block
			if config.SkipOnError {
				c.Header("X-RateLimit-Error", "true")
				c.Next()
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Rate limit check failed",
				"code":  "RATE_LIMIT_ERROR",
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			retryAfter := resetTime - time.Now().Unix()
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))

			if config.OnLimitReached != nil {
				config.OnLimitReached(c)
			} else {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       "Rate limit exceeded",
					"code":        "RATE_LIMIT_EXCEEDED",
					"retry_after": retryAfter,
				})
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkRateLimit implements sliding window rate limiting algorithm
func (r *RateLimitMiddleware) checkRateLimit(key string, limit int, window time.Duration) (allowed bool, remaining int, resetTime int64, err error) {
	ctx := context.Background()
	now := time.Now()
	windowStart := now.Add(-window)

	// Redis pipeline for atomic operations
	pipe := r.redisClient.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current entries in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d:%d", now.UnixNano(), now.UnixNano()%1000000), // Ensure uniqueness
	})

	// Set expiration
	pipe.Expire(ctx, key, window+time.Minute) // Extra buffer to handle clock skew

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, fmt.Errorf("redis pipeline execution failed: %w", err)
	}

	// Get count result
	currentCount, err := countCmd.Result()
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to get count: %w", err)
	}

	// Calculate remaining and reset time
	remaining = limit - int(currentCount) - 1 // -1 for current request
	if remaining < 0 {
		remaining = 0
	}

	resetTime = now.Add(window).Unix()
	allowed = int(currentCount) < limit

	return allowed, remaining, resetTime, nil
}

// CustomRateLimit allows creating custom rate limiting configurations
func (r *RateLimitMiddleware) CustomRateLimit(requests int, duration time.Duration, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	config := RateLimitConfig{
		Requests:     requests,
		Duration:     duration,
		KeyGenerator: keyFunc,
		SkipOnError:  true, // Default to skip on Redis errors for custom limits
	}
	return r.RateLimit(config)
}

// BurstRateLimit allows burst traffic with token bucket algorithm
func (r *RateLimitMiddleware) BurstRateLimit(burst int, refillRate float64, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		allowed, err := r.checkTokenBucket(key, burst, refillRate)

		if err != nil {
			c.Header("X-RateLimit-Error", "true")
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"code":  "BURST_RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkTokenBucket implements token bucket algorithm using Redis
func (r *RateLimitMiddleware) checkTokenBucket(key string, burst int, refillRate float64) (bool, error) {
	ctx := context.Background()
	now := time.Now()

	// Lua script for atomic token bucket check
	luaScript := `
		local key = KEYS[1]
		local burst = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1]) or burst
		local last_refill = tonumber(bucket[2]) or now
		
		-- Calculate tokens to add based on time passed
		local time_passed = (now - last_refill) / 1000000000 -- Convert to seconds
		local tokens_to_add = math.floor(time_passed * refill_rate)
		tokens = math.min(burst, tokens + tokens_to_add)
		
		-- Check if we can consume a token
		if tokens >= 1 then
			tokens = tokens - 1
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 3600) -- 1 hour expiration
			return 1
		else
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, 3600)
			return 0
		end
	`

	result, err := r.redisClient.Eval(ctx, luaScript, []string{key}, burst, refillRate, now.UnixNano()).Result()
	if err != nil {
		return false, fmt.Errorf("token bucket check failed: %w", err)
	}

	return result.(int64) == 1, nil
}

// GetRateLimitStatus returns current rate limit status for debugging
func (r *RateLimitMiddleware) GetRateLimitStatus(key string, window time.Duration) (current int, err error) {
	ctx := context.Background()
	windowStart := time.Now().Add(-window)

	count, err := r.redisClient.ZCount(ctx, key, strconv.FormatInt(windowStart.UnixNano(), 10), "+inf").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get rate limit status: %w", err)
	}

	return int(count), nil
}
