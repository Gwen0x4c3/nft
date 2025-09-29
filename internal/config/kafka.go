package config

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConfig holds the Kafka configuration
type KafkaConfig struct {
	Brokers         []string
	ClientID        string
	GroupID         string
	MinBytes        int
	MaxBytes        int
	MaxWait         time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	BatchSize       int
	BatchTimeout    time.Duration
	RequiredAcks    kafka.RequiredAcks
	RetryMax        int
	RetryBackoffMax time.Duration
	Async           bool
}

// LoadKafkaConfig loads Kafka configuration from environment variables
func LoadKafkaConfig() *KafkaConfig {
	brokers := strings.Split(getEnvString("KAFKA_BROKERS", "localhost:9092"), ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	return &KafkaConfig{
		Brokers:         brokers,
		ClientID:        getEnvString("KAFKA_CLIENT_ID", "nft-platform"),
		GroupID:         getEnvString("KAFKA_GROUP_ID", "nft-platform-group"),
		MinBytes:        getEnvInt("KAFKA_MIN_BYTES", 1),
		MaxBytes:        getEnvInt("KAFKA_MAX_BYTES", 10*1024*1024), // 10MB
		MaxWait:         time.Duration(getEnvInt("KAFKA_MAX_WAIT_MS", 500)) * time.Millisecond,
		ReadTimeout:     time.Duration(getEnvInt("KAFKA_READ_TIMEOUT_S", 10)) * time.Second,
		WriteTimeout:    time.Duration(getEnvInt("KAFKA_WRITE_TIMEOUT_S", 10)) * time.Second,
		BatchSize:       getEnvInt("KAFKA_BATCH_SIZE", 100),
		BatchTimeout:    time.Duration(getEnvInt("KAFKA_BATCH_TIMEOUT_MS", 1000)) * time.Millisecond,
		RequiredAcks:    kafka.RequiredAcks(getEnvInt("KAFKA_REQUIRED_ACKS", 1)),
		RetryMax:        getEnvInt("KAFKA_RETRY_MAX", 3),
		RetryBackoffMax: time.Duration(getEnvInt("KAFKA_RETRY_BACKOFF_MS", 1000)) * time.Millisecond,
		Async:           getEnvBool("KAFKA_ASYNC", false),
	}
}

// Topic names for different event types
const (
	TopicBidPlaced        = "bid-placed"
	TopicAuctionEnded     = "auction-ended"
	TopicNFTMinted        = "nft-minted"
	TopicNFTTransferred   = "nft-transferred"
	TopicUserRegistered   = "user-registered"
	TopicNotification     = "notification"
	TopicBlockchainEvents = "blockchain-events"
)

// Event types for message routing
type EventType string

const (
	EventTypeBidPlaced      EventType = "bid_placed"
	EventTypeAuctionEnded   EventType = "auction_ended"
	EventTypeNFTMinted      EventType = "nft_minted"
	EventTypeNFTTransferred EventType = "nft_transferred"
	EventTypeUserRegistered EventType = "user_registered"
	EventTypeNotification   EventType = "notification"
)

// Message represents a Kafka message structure
type Message struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    *uint       `json:"user_id,omitempty"`
	Data      interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// KafkaProducer provides methods for producing messages
type KafkaProducer struct {
	writer *kafka.Writer
	config *KafkaConfig
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(config *KafkaConfig) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(config.Brokers...),
		Balancer:               &kafka.Hash{},
		BatchSize:              config.BatchSize,
		BatchTimeout:           config.BatchTimeout,
		RequiredAcks:           config.RequiredAcks,
		MaxAttempts:            config.RetryMax,
		WriteTimeout:           config.WriteTimeout,
		ReadTimeout:            config.ReadTimeout,
		Async:                  config.Async,
		Logger:                 kafka.LoggerFunc(log.Printf),
		ErrorLogger:            kafka.LoggerFunc(log.Printf),
	}

	return &KafkaProducer{
		writer: writer,
		config: config,
	}
}

// PublishMessage publishes a message to the specified topic
func (p *KafkaProducer) PublishMessage(ctx context.Context, topic string, message *Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMessage := kafka.Message{
		Topic: topic,
		Key:   []byte(message.ID),
		Value: messageBytes,
		Headers: []kafka.Header{
			{Key: "type", Value: []byte(message.Type)},
			{Key: "timestamp", Value: []byte(message.Timestamp.Format(time.RFC3339))},
		},
	}

	if message.UserID != nil {
		kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
			Key:   "user_id",
			Value: []byte(fmt.Sprintf("%d", *message.UserID)),
		})
	}

	return p.writer.WriteMessages(ctx, kafkaMessage)
}

// PublishBidPlaced publishes a bid placed event
func (p *KafkaProducer) PublishBidPlaced(ctx context.Context, bidData interface{}) error {
	message := &Message{
		ID:        generateMessageID(),
		Type:      EventTypeBidPlaced,
		Timestamp: time.Now(),
		Data:      bidData,
	}
	return p.PublishMessage(ctx, TopicBidPlaced, message)
}

// PublishAuctionEnded publishes an auction ended event
func (p *KafkaProducer) PublishAuctionEnded(ctx context.Context, auctionData interface{}) error {
	message := &Message{
		ID:        generateMessageID(),
		Type:      EventTypeAuctionEnded,
		Timestamp: time.Now(),
		Data:      auctionData,
	}
	return p.PublishMessage(ctx, TopicAuctionEnded, message)
}

// PublishNFTMinted publishes an NFT minted event
func (p *KafkaProducer) PublishNFTMinted(ctx context.Context, nftData interface{}) error {
	message := &Message{
		ID:        generateMessageID(),
		Type:      EventTypeNFTMinted,
		Timestamp: time.Now(),
		Data:      nftData,
	}
	return p.PublishMessage(ctx, TopicNFTMinted, message)
}

// PublishNotification publishes a notification event
func (p *KafkaProducer) PublishNotification(ctx context.Context, userID uint, notificationData interface{}) error {
	message := &Message{
		ID:        generateMessageID(),
		Type:      EventTypeNotification,
		Timestamp: time.Now(),
		UserID:    &userID,
		Data:      notificationData,
	}
	return p.PublishMessage(ctx, TopicNotification, message)
}

// Close closes the producer
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// KafkaConsumer provides methods for consuming messages
type KafkaConsumer struct {
	reader *kafka.Reader
	config *KafkaConfig
}

// NewKafkaConsumer creates a new Kafka consumer for the specified topics
func NewKafkaConsumer(config *KafkaConfig, topics ...string) *KafkaConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.Brokers,
		GroupID:        config.GroupID,
		GroupTopics:    topics,
		MinBytes:       config.MinBytes,
		MaxBytes:       config.MaxBytes,
		MaxWait:        config.MaxWait,
		Logger:         kafka.LoggerFunc(log.Printf),
		ErrorLogger:    kafka.LoggerFunc(log.Printf),
	})

	return &KafkaConsumer{
		reader: reader,
		config: config,
	}
}

// MessageHandler defines the interface for handling messages
type MessageHandler func(ctx context.Context, message *Message) error

// ConsumeMessages starts consuming messages and processes them with the provided handler
func (c *KafkaConsumer) ConsumeMessages(ctx context.Context, handler MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			kafkaMessage, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			var message Message
			if err := json.Unmarshal(kafkaMessage.Value, &message); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			if err := handler(ctx, &message); err != nil {
				log.Printf("Error handling message: %v", err)
				// In production, you might want to send to a dead letter queue
				continue
			}

			// Commit the message
			if err := c.reader.CommitMessages(ctx, kafkaMessage); err != nil {
				log.Printf("Error committing message: %v", err)
			}
		}
	}
}

// Close closes the consumer
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// TopicManager provides methods for managing Kafka topics
type TopicManager struct {
	conn *kafka.Conn
}

// NewTopicManager creates a new topic manager
func NewTopicManager(brokers []string) (*TopicManager, error) {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}

	return &TopicManager{conn: conn}, nil
}

// CreateTopics creates the required topics if they don't exist
func (tm *TopicManager) CreateTopics() error {
	topics := []kafka.TopicConfig{
		{Topic: TopicBidPlaced, NumPartitions: 3, ReplicationFactor: 1},
		{Topic: TopicAuctionEnded, NumPartitions: 3, ReplicationFactor: 1},
		{Topic: TopicNFTMinted, NumPartitions: 3, ReplicationFactor: 1},
		{Topic: TopicNFTTransferred, NumPartitions: 3, ReplicationFactor: 1},
		{Topic: TopicUserRegistered, NumPartitions: 3, ReplicationFactor: 1},
		{Topic: TopicNotification, NumPartitions: 3, ReplicationFactor: 1},
		{Topic: TopicBlockchainEvents, NumPartitions: 3, ReplicationFactor: 1},
	}

	return tm.conn.CreateTopics(topics...)
}

// HealthCheck checks Kafka connectivity
func (tm *TopicManager) HealthCheck() error {
	_, err := tm.conn.Brokers()
	return err
}

// Close closes the topic manager connection
func (tm *TopicManager) Close() error {
	return tm.conn.Close()
}

// Helper functions

func generateMessageID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// CreateTestKafkaProducer creates a Kafka producer for testing
func CreateTestKafkaProducer() *KafkaProducer {
	testConfig := &KafkaConfig{
		Brokers:         []string{getEnvString("TEST_KAFKA_BROKERS", "localhost:9092")},
		ClientID:        "nft-platform-test",
		BatchSize:       1,
		BatchTimeout:    100 * time.Millisecond,
		RequiredAcks:    kafka.RequireOne,
		RetryMax:        1,
		WriteTimeout:    5 * time.Second,
		ReadTimeout:     5 * time.Second,
		Async:           false,
	}

	return NewKafkaProducer(testConfig)
}

// CreateTestKafkaConsumer creates a Kafka consumer for testing
func CreateTestKafkaConsumer(topics ...string) *KafkaConsumer {
	testConfig := &KafkaConfig{
		Brokers:     []string{getEnvString("TEST_KAFKA_BROKERS", "localhost:9092")},
		GroupID:     "nft-platform-test-group",
		MinBytes:    1,
		MaxBytes:    1024 * 1024,
		MaxWait:     500 * time.Millisecond,
		ReadTimeout: 5 * time.Second,
	}

	return NewKafkaConsumer(testConfig, topics...)
}