package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisConfig holds the Redis configuration
type RedisConfig struct {
	Host         string
	Port         int
	Password     string
	Database     int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
	IdleTimeout  time.Duration
}

// LoadRedisConfig loads Redis configuration from environment variables
func LoadRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:         getEnvString("REDIS_HOST", "localhost"),
		Port:         getEnvInt("REDIS_PORT", 6379),
		Password:     getEnvString("REDIS_PASSWORD", ""),
		Database:     getEnvInt("REDIS_DB", 0),
		PoolSize:     getEnvInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: getEnvInt("REDIS_MIN_IDLE_CONNS", 2),
		MaxRetries:   getEnvInt("REDIS_MAX_RETRIES", 3),
		DialTimeout:  time.Duration(getEnvInt("REDIS_DIAL_TIMEOUT", 5)) * time.Second,
		ReadTimeout:  time.Duration(getEnvInt("REDIS_READ_TIMEOUT", 3)) * time.Second,
		WriteTimeout: time.Duration(getEnvInt("REDIS_WRITE_TIMEOUT", 3)) * time.Second,
		PoolTimeout:  time.Duration(getEnvInt("REDIS_POOL_TIMEOUT", 4)) * time.Second,
		IdleTimeout:  time.Duration(getEnvInt("REDIS_IDLE_TIMEOUT", 300)) * time.Second,
	}
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.Database,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
		IdleTimeout:  config.IdleTimeout,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Redis connected: %s:%d", config.Host, config.Port)
	return client, nil
}

// RedisService provides Redis operations
type RedisService struct {
	client *redis.Client
}

// NewRedisService creates a new Redis service
func NewRedisService(client *redis.Client) *RedisService {
	return &RedisService{client: client}
}

// Set stores a key-value pair with expiration
func (r *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (r *RedisService) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete removes a key
func (r *RedisService) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (r *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// SetNX sets a key only if it doesn't exist (for locking)
func (r *RedisService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiration).Result()
}

// Increment increments a counter
func (r *RedisService) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// IncrementBy increments a counter by a specific value
func (r *RedisService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// Expire sets expiration for a key
func (r *RedisService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the time to live for a key
func (r *RedisService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// HSet sets field in hash
func (r *RedisService) HSet(ctx context.Context, key, field string, value interface{}) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HGet gets field from hash
func (r *RedisService) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from hash
func (r *RedisService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel deletes field from hash
func (r *RedisService) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// Publish publishes a message to a channel (for pub/sub)
func (r *RedisService) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to channels
func (r *RedisService) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// ZAdd adds member to sorted set
func (r *RedisService) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	return r.client.ZAdd(ctx, key, &redis.Z{Score: score, Member: member}).Err()
}

// ZRange gets members from sorted set by rank
func (r *RedisService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore gets members from sorted set by score
func (r *RedisService) ZRangeByScore(ctx context.Context, key string, min, max string) ([]string, error) {
	return r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()
}

// HealthCheck checks if Redis is accessible
func (r *RedisService) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Session management helpers
const (
	SessionKeyPrefix = "session:"
	UserSessionsKey  = "user_sessions:"
	RateLimitPrefix  = "rate_limit:"
	CachePrefix      = "cache:"
)

// SetSession stores user session
func (r *RedisService) SetSession(ctx context.Context, sessionID string, userID uint, expiration time.Duration) error {
	sessionKey := SessionKeyPrefix + sessionID
	userSessionsKey := UserSessionsKey + fmt.Sprintf("%d", userID)

	// Store session data
	if err := r.HSet(ctx, sessionKey, "user_id", userID); err != nil {
		return err
	}
	if err := r.HSet(ctx, sessionKey, "created_at", time.Now().Unix()); err != nil {
		return err
	}
	if err := r.Expire(ctx, sessionKey, expiration); err != nil {
		return err
	}

	// Add to user's session list
	return r.client.SAdd(ctx, userSessionsKey, sessionID).Err()
}

// GetSession retrieves session data
func (r *RedisService) GetSession(ctx context.Context, sessionID string) (map[string]string, error) {
	sessionKey := SessionKeyPrefix + sessionID
	return r.HGetAll(ctx, sessionKey)
}

// DeleteSession removes a session
func (r *RedisService) DeleteSession(ctx context.Context, sessionID string, userID uint) error {
	sessionKey := SessionKeyPrefix + sessionID
	userSessionsKey := UserSessionsKey + fmt.Sprintf("%d", userID)

	// Remove session data
	if err := r.Delete(ctx, sessionKey); err != nil {
		return err
	}

	// Remove from user's session list
	return r.client.SRem(ctx, userSessionsKey, sessionID).Err()
}

// CreateTestRedisClient creates a Redis client for testing
func CreateTestRedisClient() (*redis.Client, error) {
	testConfig := &RedisConfig{
		Host:         getEnvString("TEST_REDIS_HOST", "localhost"),
		Port:         getEnvInt("TEST_REDIS_PORT", 6379),
		Password:     getEnvString("TEST_REDIS_PASSWORD", ""),
		Database:     getEnvInt("TEST_REDIS_DB", 1), // Use different DB for testing
		PoolSize:     5,
		MinIdleConns: 1,
		MaxRetries:   2,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolTimeout:  2 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return NewRedisClient(testConfig)
}