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

// Consumer handles consuming events from Kafka topics
type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
	config *config.KafkaConfig
}

// EventHandler defines the interface for handling events
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
}

// BlockchainEventHandler handles blockchain-specific events
type BlockchainEventHandler interface {
	EventHandler
	HandleTransactionConfirmed(ctx context.Context, data *BlockchainEventData) error
	HandleTransactionFailed(ctx context.Context, data *BlockchainEventData) error
	HandleNFTMinted(ctx context.Context, data *NFTMintedData) error
	HandleNFTTransferred(ctx context.Context, data *NFTTransferredData) error
}

// NotificationEventHandler handles notification events
type NotificationEventHandler interface {
	EventHandler
	HandleBidPlaced(ctx context.Context, data *BidPlacedData) error
	HandleAuctionEnded(ctx context.Context, data *AuctionEndedData) error
	HandleUserNotification(ctx context.Context, data *NotificationData) error
}

// NewConsumer creates a new Kafka event consumer
func NewConsumer(cfg *config.KafkaConfig, topics []string, logger *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		GroupTopics:    topics,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		ReadBackoffMin: 100 * time.Millisecond,
		ReadBackoffMax: 1 * time.Second,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
		Logger: kafka.LoggerFunc(func(format string, args ...interface{}) {
			logger.Debug(fmt.Sprintf(format, args...))
		}),
		ErrorLogger: kafka.LoggerFunc(func(format string, args ...interface{}) {
			logger.Error(fmt.Sprintf(format, args...))
		}),
	})

	return &Consumer{
		reader: reader,
		logger: logger,
		config: cfg,
	}
}

// StartConsuming starts consuming messages from Kafka topics
func (c *Consumer) StartConsuming(ctx context.Context, handler EventHandler) error {
	c.logger.Info("starting Kafka consumer",
		zap.Strings("topics", c.reader.Config().GroupTopics),
		zap.String("group_id", c.reader.Config().GroupID),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("shutting down Kafka consumer")
			return ctx.Err()
		default:
			message, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c.logger.Error("failed to read message", zap.Error(err))
				continue
			}

			if err := c.processMessage(ctx, message, handler); err != nil {
				c.logger.Error("failed to process message",
					zap.Error(err),
					zap.String("topic", message.Topic),
					zap.Int64("offset", message.Offset),
					zap.Int("partition", message.Partition),
				)
				// Continue processing other messages even if one fails
				continue
			}
		}
	}
}

// processMessage processes a single Kafka message
func (c *Consumer) processMessage(ctx context.Context, message kafka.Message, handler EventHandler) error {
	// Extract event type from headers
	eventType := c.extractHeaderValue(message.Headers, "event_type")
	if eventType == "" {
		return fmt.Errorf("missing event_type header")
	}

	// Parse event
	var event Event
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Log message receipt
	c.logger.Debug("received event",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)),
		zap.String("topic", message.Topic),
		zap.Int64("offset", message.Offset),
		zap.String("aggregate_id", event.AggregateID),
	)

	// Handle event based on type
	if err := handler.Handle(ctx, &event); err != nil {
		c.logger.Error("event handler failed",
			zap.Error(err),
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.Type)),
		)
		return fmt.Errorf("event handler failed: %w", err)
	}

	// Commit message if synchronous processing
	if err := c.reader.CommitMessages(ctx, message); err != nil {
		c.logger.Error("failed to commit message",
			zap.Error(err),
			zap.String("event_id", event.ID),
			zap.String("topic", message.Topic),
		)
		return fmt.Errorf("failed to commit message: %w", err)
	}

	c.logger.Debug("event processed and committed successfully",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)),
	)

	return nil
}

// extractHeaderValue extracts a value from message headers
func (c *Consumer) extractHeaderValue(headers []kafka.Header, key string) string {
	for _, header := range headers {
		if header.Key == key {
			return string(header.Value)
		}
	}
	return ""
}

// Close closes the consumer and releases resources
func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		c.logger.Error("failed to close Kafka consumer", zap.Error(err))
		return fmt.Errorf("failed to close Kafka consumer: %w", err)
	}

	c.logger.Info("Kafka consumer closed successfully")
	return nil
}

// HealthCheck checks the health of the Kafka consumer
func (c *Consumer) HealthCheck(ctx context.Context) error {
	// Try to get broker information to verify connectivity
	conn, err := kafka.Dial("tcp", c.config.Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka broker: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to get broker information: %w", err)
	}

	c.logger.Debug("Kafka consumer health check passed")
	return nil
}

// MultiEventHandler dispatches events to appropriate handlers
type MultiEventHandler struct {
	blockchainHandler   BlockchainEventHandler
	notificationHandler NotificationEventHandler
	fallbackHandler     EventHandler
	logger              *zap.Logger
}

// NewMultiEventHandler creates a new multi-event handler
func NewMultiEventHandler(
	blockchainHandler BlockchainEventHandler,
	notificationHandler NotificationEventHandler,
	fallbackHandler EventHandler,
	logger *zap.Logger,
) *MultiEventHandler {
	return &MultiEventHandler{
		blockchainHandler:   blockchainHandler,
		notificationHandler: notificationHandler,
		fallbackHandler:     fallbackHandler,
		logger:              logger,
	}
}

// Handle dispatches events to the appropriate handler
func (h *MultiEventHandler) Handle(ctx context.Context, event *Event) error {
	switch event.Type {
	case EventTypeTransactionConfirmed, EventTypeTransactionFailed:
		if h.blockchainHandler != nil {
			var data BlockchainEventData
			if err := h.unmarshalEventData(event.Data, &data); err != nil {
				return fmt.Errorf("failed to unmarshal blockchain event data: %w", err)
			}

			if event.Type == EventTypeTransactionConfirmed {
				return h.blockchainHandler.HandleTransactionConfirmed(ctx, &data)
			}
			return h.blockchainHandler.HandleTransactionFailed(ctx, &data)
		}

	case EventTypeNFTMinted:
		if h.blockchainHandler != nil {
			var data NFTMintedData
			if err := h.unmarshalEventData(event.Data, &data); err != nil {
				return fmt.Errorf("failed to unmarshal NFT minted data: %w", err)
			}
			return h.blockchainHandler.HandleNFTMinted(ctx, &data)
		}

	case EventTypeNFTTransferred:
		if h.blockchainHandler != nil {
			var data NFTTransferredData
			if err := h.unmarshalEventData(event.Data, &data); err != nil {
				return fmt.Errorf("failed to unmarshal NFT transferred data: %w", err)
			}
			return h.blockchainHandler.HandleNFTTransferred(ctx, &data)
		}

	case EventTypeBidPlaced:
		if h.notificationHandler != nil {
			var data BidPlacedData
			if err := h.unmarshalEventData(event.Data, &data); err != nil {
				return fmt.Errorf("failed to unmarshal bid placed data: %w", err)
			}
			return h.notificationHandler.HandleBidPlaced(ctx, &data)
		}

	case EventTypeAuctionEnded:
		if h.notificationHandler != nil {
			var data AuctionEndedData
			if err := h.unmarshalEventData(event.Data, &data); err != nil {
				return fmt.Errorf("failed to unmarshal auction ended data: %w", err)
			}
			return h.notificationHandler.HandleAuctionEnded(ctx, &data)
		}

	case EventTypeNotification:
		if h.notificationHandler != nil {
			var data NotificationData
			if err := h.unmarshalEventData(event.Data, &data); err != nil {
				return fmt.Errorf("failed to unmarshal notification data: %w", err)
			}
			return h.notificationHandler.HandleUserNotification(ctx, &data)
		}

	default:
		if h.fallbackHandler != nil {
			return h.fallbackHandler.Handle(ctx, event)
		}
		h.logger.Warn("no handler found for event type",
			zap.String("event_type", string(event.Type)),
			zap.String("event_id", event.ID),
		)
	}

	return nil
}

// unmarshalEventData unmarshals event data into the provided struct
func (h *MultiEventHandler) unmarshalEventData(data interface{}, target interface{}) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	return nil
}

// DefaultFallbackHandler provides a default fallback handler
type DefaultFallbackHandler struct {
	logger *zap.Logger
}

// NewDefaultFallbackHandler creates a new default fallback handler
func NewDefaultFallbackHandler(logger *zap.Logger) *DefaultFallbackHandler {
	return &DefaultFallbackHandler{
		logger: logger,
	}
}

// Handle handles events that don't have specific handlers
func (h *DefaultFallbackHandler) Handle(ctx context.Context, event *Event) error {
	h.logger.Info("handling event with default fallback handler",
		zap.String("event_id", event.ID),
		zap.String("event_type", string(event.Type)),
		zap.String("aggregate_type", event.AggregateType),
		zap.String("aggregate_id", event.AggregateID),
	)

	// Default implementation just logs the event
	// In a real implementation, you might want to:
	// - Store the event for later processing
	// - Send to a dead letter queue
	// - Trigger alerts for unknown event types

	return nil
}

