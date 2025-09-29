package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	SSLMode      string
	Timezone     string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	LogLevel     logger.LogLevel
}

// LoadDatabaseConfig loads database configuration from environment variables
func LoadDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:         getEnvString("DB_HOST", "localhost"),
		Port:         getEnvInt("DB_PORT", 5432),
		User:         getEnvString("DB_USER", "postgres"),
		Password:     getEnvString("DB_PASSWORD", "postgres"),
		DBName:       getEnvString("DB_NAME", "nft_platform"),
		SSLMode:      getEnvString("DB_SSL_MODE", "disable"),
		Timezone:     getEnvString("DB_TIMEZONE", "UTC"),
		MaxOpenConns: getEnvInt("DB_MAX_OPEN_CONNS", 100),
		MaxIdleConns: getEnvInt("DB_MAX_IDLE_CONNS", 10),
		MaxLifetime:  time.Duration(getEnvInt("DB_MAX_LIFETIME_MINUTES", 60)) * time.Minute,
		LogLevel:     getDatabaseLogLevel(),
	}
}

// NewDatabase creates a new database connection
func NewDatabase(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
		config.Timezone,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.MaxLifetime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected: %s:%d/%s", config.Host, config.Port, config.DBName)
	return db, nil
}

// AutoMigrate runs database migrations for all models
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}
	return nil
}

// HealthCheck checks if the database is accessible
func (dc *DatabaseConfig) HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// CreateTestDatabase creates a test database for testing
func CreateTestDatabase() (*gorm.DB, error) {
	testConfig := &DatabaseConfig{
		Host:         getEnvString("TEST_DB_HOST", "localhost"),
		Port:         getEnvInt("TEST_DB_PORT", 5432),
		User:         getEnvString("TEST_DB_USER", "postgres"),
		Password:     getEnvString("TEST_DB_PASSWORD", "postgres"),
		DBName:       getEnvString("TEST_DB_NAME", "nft_platform_test"),
		SSLMode:      "disable",
		Timezone:     "UTC",
		MaxOpenConns: 10,
		MaxIdleConns: 2,
		MaxLifetime:  5 * time.Minute,
		LogLevel:     logger.Silent,
	}

	return NewDatabase(testConfig)
}

// Helper functions
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDatabaseLogLevel() logger.LogLevel {
	level := getEnvString("DB_LOG_LEVEL", "warn")
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}