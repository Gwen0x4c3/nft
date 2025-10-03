package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"nft-platform/internal/config"
	"nft-platform/internal/handlers"
	"nft-platform/internal/middleware"
	"nft-platform/internal/messaging"
	"nft-platform/internal/models"
	"nft-platform/internal/repository"
	"nft-platform/internal/service"
	"nft-platform/internal/websocket"
	"nft-platform/pkg/auth"
)

// simpleMessagingService adapts the messaging.Producer to service.MessagingService interface
type simpleMessagingService struct {
	producer *messaging.Producer
}

func (s *simpleMessagingService) PublishNotification(ctx context.Context, notification *models.Notification) error {
	// Convert notification to Event format expected by Producer
	event := &messaging.Event{
		Type:          messaging.EventTypeNotification,
		AggregateType: "notification",
		AggregateID:   fmt.Sprintf("%d", notification.ID),
		Version:       "1.0",
		Timestamp:     time.Now(),
		UserID:        &notification.UserID,
		Data:          notification,
		CorrelationID: generateID(),
	}
	return s.producer.PublishEvent(ctx, event)
}

func (s *simpleMessagingService) PublishEvent(ctx context.Context, event string, data interface{}) error {
	// Convert data to Event format expected by Producer
	msgEvent := &messaging.Event{
		Type:          messaging.EventType(event),
		AggregateType: "event",
		AggregateID:   generateID(),
		Version:       "1.0",
		Timestamp:     time.Now(),
		Data:          data,
		CorrelationID: generateID(),
	}
	return s.producer.PublishEvent(ctx, msgEvent)
}

// generateID generates a random ID for events
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func main() {
	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	zap.ReplaceGlobals(logger)
	zap.L().Info("Starting NFT Platform Server")

	// Load configurations
	dbConfig := config.LoadDatabaseConfig()
	redisConfig := config.LoadRedisConfig()
	kafkaConfig := config.LoadKafkaConfig()

	// Initialize database connection
	db, err := config.NewDatabase(dbConfig)
	if err != nil {
		zap.L().Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run database migrations
	if err := config.AutoMigrate(db,
		&models.User{},
		&models.NFT{},
		&models.Auction{},
		&models.Bid{},
		&models.Transfer{},
		&models.Notification{},
	); err != nil {
		zap.L().Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize Redis client
	redisClient, err := config.NewRedisClient(redisConfig)
	if err != nil {
		zap.L().Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize Kafka producer and consumer
	kafkaProducer := messaging.NewProducer(kafkaConfig, logger)
	defer kafkaProducer.Close()

	kafkaConsumer := messaging.NewConsumer(kafkaConfig, []string{"blockchain-events", "auction-events", "notification-events"}, logger)
	defer kafkaConsumer.Close()

	// Initialize JWT manager
	jwtManager, err := auth.NewJWTManagerWithKeys("nft-platform")
	if err != nil {
		zap.L().Fatal("Failed to initialize JWT manager", zap.Error(err))
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	nftRepo := repository.NewNFTRepository(db)
	auctionRepo := repository.NewAuctionRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	transferRepo := repository.NewTransferRepository(db)

	// Initialize blockchain service first
	blockchainService := service.NewMockBlockchainService()

	// Initialize services with proper configurations
	userServiceConfig := &service.ServiceConfig{
		JWTManager: jwtManager,
	}
	userService := service.NewUserService(userRepo, userServiceConfig)

	nftServiceConfig := &service.NFTServiceConfig{
		BlockchainService: blockchainService,
		IPFSService:       service.NewMockIPFSService(),
		DefaultGatewayURL: "https://ipfs.io/ipfs/",
		MaxFileSize:       10 * 1024 * 1024, // 10MB
		AllowedFileTypes:  []string{"jpg", "jpeg", "png", "gif"},
		AutoPinToIPFS:     true,
	}
	nftService := service.NewNFTService(nftRepo, userRepo, transferRepo, nftServiceConfig)

	auctionServiceConfig := &service.AuctionServiceConfig{}
	auctionService := service.NewAuctionService(auctionRepo, nftRepo, userRepo, transferRepo, auctionServiceConfig)

	wsManager := websocket.NewManager()

	// Create a simple messaging service adapter for notification service
	messagingService := &simpleMessagingService{producer: kafkaProducer}

	notificationServiceConfig := &service.NotificationServiceConfig{
		MessagingService: messagingService,
	}
	notificationService := service.NewNotificationService(notificationRepo, userRepo, notificationServiceConfig)

	// Initialize message consumer for blockchain events
	messageConsumer := messaging.NewConsumer(kafkaConfig, []string{"blockchain-events", "auction-events", "notification-events"}, logger)

	// Start background services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start message consumer
	go func() {
		if err := messageConsumer.StartConsuming(ctx, nil); err != nil {
			zap.L().Error("Message consumer stopped", zap.Error(err))
		}
	}()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService)
	userHandler := handlers.NewUserHandler(userService)
	nftHandler := handlers.NewNFTHandler(nftService)
	auctionHandler := handlers.NewAuctionHandler(auctionService)
	bidHandler := handlers.NewBidHandler(auctionService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	// Simple CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// User routes
		users := v1.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.GET("/:userId", userHandler.GetUserByID)
		}

		// NFT routes
		nfts := v1.Group("/nfts")
		{
			nfts.GET("", nftHandler.ListNFTs) // Public access
			nfts.Use(authMiddleware.RequireAuth())
			{
				nfts.POST("", nftHandler.MintNFT)
				nfts.GET("/:nftId", nftHandler.GetNFTByID)
				nfts.PUT("/:nftId", nftHandler.UpdateNFT)
				nfts.POST("/:nftId/transfer", nftHandler.TransferNFT)
			}
		}

		// Auction routes
		auctions := v1.Group("/auctions")
		{
			auctions.GET("", auctionHandler.ListAuctions) // Public access
			auctions.Use(authMiddleware.RequireAuth())
			{
				auctions.POST("", auctionHandler.CreateAuction)
				auctions.GET("/:auctionId", auctionHandler.GetAuction)
				auctions.DELETE("/:auctionId", auctionHandler.CancelAuction)
			}
		}

		// Bid routes
		bids := v1.Group("/auctions/:auctionId/bids")
		bids.Use(authMiddleware.RequireAuth())
		{
			bids.GET("", bidHandler.GetAuctionBids)
			bids.POST("", bidHandler.PlaceBid)
		}

		// Notification routes
		notifications := v1.Group("/notifications")
		notifications.Use(authMiddleware.RequireAuth())
		{
			notifications.GET("", notificationHandler.GetNotifications)
			notifications.PUT("/:notificationId/read", notificationHandler.MarkNotificationAsRead)
			notifications.PUT("/read-all", notificationHandler.MarkAllNotificationsAsRead)
		}
	}

	// WebSocket endpoint
	router.GET("/ws", wsManager.HandleConnection)

	// Get server port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		zap.L().Info("Starting HTTP server", zap.String("port", port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("Shutting down server...")

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		zap.L().Error("Server forced to shutdown", zap.Error(err))
	}

	// Cancel background services
	cancel()

	// Give some time for cleanup
	time.Sleep(2 * time.Second)

	zap.L().Info("Server exited")
}

// Helper functions for environment variables
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}