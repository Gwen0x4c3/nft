package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"nft-platform/internal/config"
)

// Producer handles publishing events to Kafka topics
type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
	config *config.KafkaConfig
}

// NewProducer creates a new Kafka event producer
func NewProducer(cfg *config.KafkaConfig, logger *zap.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{},
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		RequiredAcks: cfg.RequiredAcks,
		MaxAttempts:  cfg.RetryMax,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		Async:        cfg.Async,
		ErrorLogger: kafka.LoggerFunc(func(format string, args ...interface{}) {
			logger.Error(fmt.Sprintf(format, args...))
		}),
	}

	return &Producer{
		writer: writer,
		logger: logger,
		config: cfg,
	}
}

// PublishEvent publishes an event to the appropriate Kafka topic
func (p *Producer) PublishEvent(ctx context.Context, event *Event) error {
	// Validate event
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}

	// Determine topic based on event type
	topic, err := p.getTopicForEventType(event.Type)
	if err != nil {
		return fmt.Errorf("failed to determine topic for event type %s: %w", event.Type, err)
	}

	// Serialize event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Create Kafka message
	kafkaMessage := kafka.Message{
		Topic: topic,
		Key:   []byte(event.AggregateID),
		Value: eventBytes,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(string(event.Type))},
			{Key: "event_version", Value: []byte(event.Version)},
			{Key: "timestamp", Value: []byte(event.Timestamp.Format(time.RFC3339))},
			{Key: "correlation_id", Value: []byte(event.CorrelationID)},
		},
	}

	// Add user ID header if present
	if event.UserID != nil {
		kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
			Key:   "user_id",
			Value: []byte(fmt.Sprintf("%d", *event.UserID)),
		})
	}

	// Add aggregate type header
	if event.AggregateType != "" {
		kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
			Key:   "aggregate_type",
			Value: []byte(event.AggregateType),
		})
	}

	// Publish message
	if err := p.writer.WriteMessages(ctx, kafkaMessage); err != nil {
		p.logger.Error("failed to publish event",
			zap.Error(err),
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)),
			zap.String("topic", topic),
		)
		return fmt.Errorf("failed to publish event to topic %s: %w", topic, err)
	}

	p.logger.Info("event published successfully",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)),
		zap.String("topic", topic),
		zap.String("aggregate_id", event.AggregateID),
	)

	return nil
}

// PublishBidPlaced publishes a bid placed event
func (p *Producer) PublishBidPlaced(ctx context.Context, bidData *BidPlacedData) error {
	event := &Event{
		ID:            generateEventID(),
		Type:          EventTypeBidPlaced,
		AggregateType: "auction",
		AggregateID:   fmt.Sprintf("%d", bidData.AuctionID),
		Version:       "1.0",
		Timestamp:     time.Now(),
		UserID:        &bidData.BidderID,
		Data:          bidData,
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

// PublishAuctionEnded publishes an auction ended event
func (p *Producer) PublishAuctionEnded(ctx context.Context, auctionData *AuctionEndedData) error {
	event := &Event{
		ID:            generateEventID(),
		Type:          EventTypeAuctionEnded,
		AggregateType: "auction",
		AggregateID:   fmt.Sprintf("%d", auctionData.AuctionID),
		Version:       "1.0",
		Timestamp:     time.Now(),
		UserID:        auctionData.WinnerID,
		Data:          auctionData,
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

// PublishNFTMinted publishes an NFT minted event
func (p *Producer) PublishNFTMinted(ctx context.Context, nftData *NFTMintedData) error {
	event := &Event{
		ID:            generateEventID(),
		Type:          EventTypeNFTMinted,
		AggregateType: "nft",
		AggregateID:   fmt.Sprintf("%d", nftData.NFTID),
		Version:       "1.0",
		Timestamp:     time.Now(),
		UserID:        &nftData.CreatorID,
		Data:          nftData,
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

// PublishNFTTransferred publishes an NFT transferred event
func (p *Producer) PublishNFTTransferred(ctx context.Context, transferData *NFTTransferredData) error {
	event := &Event{
		ID:            generateEventID(),
		Type:          EventTypeNFTTransferred,
		AggregateType: "nft",
		AggregateID:   fmt.Sprintf("%d", transferData.NFTID),
		Version:       "1.0",
		Timestamp:     time.Now(),
		UserID:        &transferData.ToID,
		Data:          transferData,
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

// PublishUserRegistered publishes a user registered event
func (p *Producer) PublishUserRegistered(ctx context.Context, userData *UserRegisteredData) error {
	event := &Event{
		ID:            generateEventID(),
		Type:          EventTypeUserRegistered,
		AggregateType: "user",
		AggregateID:   fmt.Sprintf("%d", userData.UserID),
		Version:       "1.0",
		Timestamp:     time.Now(),
		UserID:        &userData.UserID,
		Data:          userData,
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

// PublishNotification publishes a notification event
func (p *Producer) PublishNotification(ctx context.Context, notificationData *NotificationData) error {
	event := &Event{
		ID:            generateEventID(),
		Type:          EventTypeNotification,
		AggregateType: "notification",
		AggregateID:   fmt.Sprintf("%d", notificationData.NotificationID),
		Version:       "1.0",
		Timestamp:     time.Now(),
		UserID:        &notificationData.UserID,
		Data:          notificationData,
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

// PublishBlockchainEvent publishes a blockchain-related event
func (p *Producer) PublishBlockchainEvent(ctx context.Context, blockchainData *BlockchainEventData) error {
	event := &Event{
		ID:            generateEventID(),
		Type:          EventTypeBlockchainEvent,
		AggregateType: "blockchain",
		AggregateID:   blockchainData.TransactionHash,
		Version:       "1.0",
		Timestamp:     time.Now(),
		Data:          blockchainData,
		CorrelationID: generateCorrelationID(),
	}

	return p.PublishEvent(ctx, event)
}

// getTopicForEventType maps event types to Kafka topics
func (p *Producer) getTopicForEventType(eventType EventType) (string, error) {
	switch eventType {
	case EventTypeBidPlaced:
		return config.TopicBidPlaced, nil
	case EventTypeAuctionEnded:
		return config.TopicAuctionEnded, nil
	case EventTypeNFTMinted:
		return config.TopicNFTMinted, nil
	case EventTypeNFTTransferred:
		return config.TopicNFTTransferred, nil
	case EventTypeUserRegistered:
		return config.TopicUserRegistered, nil
	case EventTypeNotification:
		return config.TopicNotification, nil
	case EventTypeBlockchainEvent:
		return config.TopicBlockchainEvents, nil
	default:
		return "", fmt.Errorf("unknown event type: %s", eventType)
	}
}

// Close closes the producer and releases resources
func (p *Producer) Close() error {
	if err := p.writer.Close(); err != nil {
		p.logger.Error("failed to close Kafka producer", zap.Error(err))
		return fmt.Errorf("failed to close Kafka producer: %w", err)
	}

	p.logger.Info("Kafka producer closed successfully")
	return nil
}

// HealthCheck checks the health of the Kafka producer
func (p *Producer) HealthCheck(ctx context.Context) error {
	// Try to get broker information to verify connectivity
	conn, err := kafka.Dial("tcp", p.config.Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka broker: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to get broker information: %w", err)
	}

	p.logger.Debug("Kafka producer health check passed")
	return nil
}

