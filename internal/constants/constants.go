package constants

import "time"

// API Configuration
const (
	APIVersion      = "v1"
	APIBasePath     = "/api/" + APIVersion
	DefaultPort     = 8080
	ReadTimeout     = 15 * time.Second
	WriteTimeout    = 15 * time.Second
	IdleTimeout     = 60 * time.Second
	MaxHeaderBytes  = 1 << 20 // 1MB
)

// Rate Limiting
const (
	DefaultRateLimit   = 100 // requests per minute
	AuthRateLimit      = 10  // requests per minute
	UploadRateLimit    = 20  // requests per minute
	BidRateLimit       = 30  // requests per minute
)

// Pagination
const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
	DefaultPage      = 1
)

// File Upload
const (
	MaxFileSize        = 10 * 1024 * 1024 // 10MB
	AllowedImageTypes  = "jpg,jpeg,png,gif,webp"
	UploadPath         = "./uploads"
	ImageQuality       = 85
	ThumbnailSize      = 300
)

// Blockchain
const (
	EthereumDecimals  = 18
	WeiPerEth        = 1000000000000000000
	MinBidIncrement  = 10000000000000000 // 0.01 ETH
	DefaultGasLimit  = 21000
	MaxGasPrice      = 100000000000 // 100 Gwei
)

// Time Constants
const (
	OneMinute  = time.Minute
	FiveMinutes = 5 * time.Minute
	TenMinutes = 10 * time.Minute
	OneHour    = time.Hour
	OneDay     = 24 * time.Hour
	OneWeek    = 7 * OneDay
	OneMonth   = 30 * OneDay
)

// WebSocket
const (
	WebSocketReadLimit  = 1024
	WebSocketWriteLimit = 1024
	MaxConnections     = 1000
	MessageRateLimit   = 100 // messages per minute
)

// Security
const (
	JWTAccessTokenDuration  = 1 * OneHour
	JWTRefreshTokenDuration = 24 * OneHour
	MinUsernameLength       = 3
	MaxUsernameLength       = 30
	MinBioLength            = 0
	MaxBioLength            = 500
	MinRoyalty              = 0
	MaxRoyalty              = 10
)

// Auction
const (
	MinAuctionDuration = 1 * OneHour
	MaxAuctionDuration = 7 * OneDay
	MinStartPrice      = 100000000000000000 // 0.1 ETH
	AuctionExtension   = 15 * time.Minute
)

// Notification Types
const (
	NotificationBidPlaced      = "bid_placed"
	NotificationBidOutbid      = "bid_outbid"
	NotificationAuctionWon     = "auction_won"
	NotificationAuctionEnded   = "auction_ended"
	NotificationNFTSold        = "nft_sold"
	NotificationNFTTransferred = "nft_transferred"
	NotificationFollow        = "follow"
	NotificationSystem        = "system"
)

// User Status
const (
	UserActive    = "active"
	UserSuspended = "suspended"
	UserDeleted   = "deleted"
)

// NFT Status
const (
	NFTDraft   = "draft"
	NFTMinted  = "minted"
	NFTListed  = "listed"
	NFTSold    = "sold"
	NFTBurned  = "burned"
)

// Auction Status
const (
	AuctionPending   = "pending"
	AuctionActive    = "active"
	AuctionEnded     = "ended"
	AuctionCancelled = "cancelled"
)

// Bid Status
const (
	BidPending   = "pending"
	BidConfirmed = "confirmed"
	BidFailed    = "failed"
)

// Transfer Type
const (
	TransferMint   = "mint"
	TransferSale   = "sale"
	TransferTransfer = "transfer"
	TransferAuction = "auction"
)

// Error Messages
const (
	MsgInvalidRequest     = "Invalid request format"
	MsgMissingAuth        = "Authentication required"
	MsgUnauthorized       = "Unauthorized access"
	MsgForbidden          = "Access forbidden"
	MsgNotFound           = "Resource not found"
	MsgAlreadyExists      = "Resource already exists"
	MsgValidationFailed   = "Validation failed"
	MsgInternalServerError = "Internal server error"
	MsgDatabaseError      = "Database operation failed"
	MsgNetworkError       = "Network operation failed"
	MsgTimeoutError       = "Operation timed out"
)

// Validation Messages
const (
	MsgRequiredField     = "This field is required"
	MsgInvalidEmail       = "Invalid email format"
	MsgInvalidAddress     = "Invalid Ethereum address"
	MsgInvalidSignature   = "Invalid signature format"
	MsgInvalidImage       = "Invalid image format"
	MsgFileTooLarge       = "File size exceeds maximum limit"
	MsgUsernameTooShort   = "Username must be at least 3 characters"
	MsgUsernameTooLong    = "Username cannot exceed 30 characters"
	MsgBioTooLong         = "Bio cannot exceed 500 characters"
	MsgInvalidPrice       = "Invalid price format"
	MsgInvalidBidAmount   = "Invalid bid amount"
	MsgBidTooLow         = "Bid amount is too low"
	MsgAuctionNotActive   = "Auction is not active"
	MsgCannotBidOwnAuction = "Cannot bid on your own auction"
)

// Success Messages
const (
	MsgLoginSuccess      = "Login successful"
	MsgRegistrationSuccess = "Registration successful"
	MsgNFTCreated        = "NFT created successfully"
	MsgAuctionCreated    = "Auction created successfully"
	MsgBidPlaced         = "Bid placed successfully"
	MsgProfileUpdated    = "Profile updated successfully"
	MsgAuctionCancelled  = "Auction cancelled successfully"
	MsgNotificationsRead = "All notifications marked as read"
)

// Email Templates
const (
	EmailWelcome      = "welcome"
	EmailResetPassword = "reset_password"
	EmailBidPlaced    = "bid_placed"
	EmailAuctionWon   = "auction_won"
	EmailNFTSold      = "nft_sold"
)

// Log Levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// Cache Keys
const (
	UserSessionPrefix   = "user:session:"
	UserProfilePrefix   = "user:profile:"
	NFTCachePrefix      = "nft:"
	AuctionCachePrefix  = "auction:"
	NotificationPrefix  = "notification:"
	RateLimitPrefix     = "ratelimit:"
)

// Database Connection Pool Settings
const (
	MaxOpenConns     = 100
	MaxIdleConns     = 10
	ConnMaxLifetime  = 5 * OneHour
	ConnMaxIdleTime  = 10 * time.Minute
)

// Redis Connection Pool Settings
const (
	RedisPoolSize     = 20
	RedisMinIdleConns = 5
	RedisPoolTimeout  = 5 * time.Second
	RedisIdleTimeout  = 300 * time.Second
)

// Kafka Settings
const (
	KafkaProducerTimeout = 10 * time.Second
	KafkaConsumerTimeout = 5 * time.Second
	KafkaBatchSize       = 100
	KafkaBatchTimeout    = 10 * time.Millisecond
)

// CORS Settings
const (
	DefaultAllowedOrigins = "*"
	MaxAllowedOrigins     = 10
	DefaultAllowedMethods = "GET,POST,PUT,DELETE,OPTIONS"
	DefaultAllowedHeaders = "Origin,Content-Type,Accept,Authorization"
)

// Metrics
const (
	MetricsPath         = "/metrics"
	HealthCheckPath     = "/health"
	ReadinessCheckPath  = "/ready"
	MetricsNamespace    = "nft_platform"
)

// Environment
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Default Values
const (
	DefaultUserAvatar  = "/assets/default-avatar.png"
	DefaultNFTImage   = "/assets/default-nft.png"
	DefaultPagination = DefaultPageLimit
	DefaultTimeout    = 30 * time.Second
)