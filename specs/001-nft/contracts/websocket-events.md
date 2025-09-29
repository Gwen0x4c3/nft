# WebSocket Events Specification

## Overview

This document defines the real-time WebSocket events used in the NFT platform for bidirectional communication between the server and clients.

## Connection

### Endpoint

```
ws://localhost:8080/ws?token={jwt_token}
wss://api.nftplatform.com/ws?token={jwt_token}
```

### Authentication

- JWT token must be provided as a query parameter
- Token is validated on connection establishment
- Invalid tokens result in connection rejection (4001 close code)

### Connection Lifecycle

1. Client initiates WebSocket connection with valid JWT token
2. Server validates token and establishes connection
3. Server sends initial `connection_established` event
4. Client can subscribe to specific event types
5. Server sends relevant events based on subscriptions and user context

## Message Format

All WebSocket messages follow this JSON structure:

```json
{
  "event": "event_type",
  "data": {
    // Event-specific data
  },
  "timestamp": "2023-12-01T10:00:00Z",
  "id": "unique_message_id"
}
```

## Client-to-Server Events

### 1. Subscribe to Events

```json
{
  "event": "subscribe",
  "data": {
    "types": ["bid_placed", "auction_ended", "nft_transferred"],
    "filters": {
      "auction_ids": [1, 2, 3],
      "nft_ids": [10, 20, 30],
      "user_ids": [100, 200]
    }
  }
}
```

### 2. Unsubscribe from Events

```json
{
  "event": "unsubscribe",
  "data": {
    "types": ["bid_placed"]
  }
}
```

### 3. Ping

```json
{
  "event": "ping",
  "data": {}
}
```

## Server-to-Client Events

### 1. Connection Established

```json
{
  "event": "connection_established",
  "data": {
    "user_id": 123,
    "connection_id": "conn_456789"
  },
  "timestamp": "2023-12-01T10:00:00Z",
  "id": "msg_001"
}
```

### 2. Subscription Confirmed

```json
{
  "event": "subscription_confirmed",
  "data": {
    "types": ["bid_placed", "auction_ended"],
    "filters": {
      "auction_ids": [1, 2, 3]
    }
  },
  "timestamp": "2023-12-01T10:00:00Z",
  "id": "msg_002"
}
```

### 3. Pong

```json
{
  "event": "pong",
  "data": {},
  "timestamp": "2023-12-01T10:00:00Z",
  "id": "msg_003"
}
```

## Auction Events

### 1. New Bid Placed

```json
{
  "event": "bid_placed",
  "data": {
    "auction_id": 123,
    "bid": {
      "id": 456,
      "bidder_id": 789,
      "bidder_username": "crypto_collector",
      "bidder_avatar": "https://example.com/avatar.jpg",
      "amount": "1500000000000000000",
      "amount_formatted": "1.5 ETH",
      "tx_hash": "0x1234567890abcdef...",
      "created_at": "2023-12-01T10:00:00Z"
    },
    "auction": {
      "id": 123,
      "current_bid": "1500000000000000000",
      "current_bid_formatted": "1.5 ETH",
      "bid_count": 5,
      "end_time": "2023-12-01T18:00:00Z"
    },
    "nft": {
      "id": 101,
      "title": "Digital Masterpiece #1",
      "image_url": "https://example.com/nft.jpg"
    }
  },
  "timestamp": "2023-12-01T10:00:00Z",
  "id": "msg_004"
}
```

### 2. Auction Status Changed

```json
{
  "event": "auction_status_changed",
  "data": {
    "auction_id": 123,
    "old_status": "active",
    "new_status": "ended",
    "winner": {
      "id": 789,
      "username": "crypto_collector",
      "avatar": "https://example.com/avatar.jpg"
    },
    "winning_bid": "2000000000000000000",
    "winning_bid_formatted": "2.0 ETH",
    "end_time": "2023-12-01T18:00:00Z",
    "nft": {
      "id": 101,
      "title": "Digital Masterpiece #1",
      "image_url": "https://example.com/nft.jpg"
    }
  },
  "timestamp": "2023-12-01T18:00:00Z",
  "id": "msg_005"
}
```

### 3. Auction Extended

```json
{
  "event": "auction_extended",
  "data": {
    "auction_id": 123,
    "old_end_time": "2023-12-01T18:00:00Z",
    "new_end_time": "2023-12-01T18:10:00Z",
    "extension_minutes": 10,
    "reason": "Last minute bid received",
    "nft": {
      "id": 101,
      "title": "Digital Masterpiece #1"
    }
  },
  "timestamp": "2023-12-01T17:59:30Z",
  "id": "msg_006"
}
```

## NFT Events

### 1. NFT Transferred

```json
{
  "event": "nft_transferred",
  "data": {
    "nft_id": 101,
    "from_user": {
      "id": 456,
      "username": "original_owner",
      "wallet_address": "0xabcd..."
    },
    "to_user": {
      "id": 789,
      "username": "new_owner",
      "wallet_address": "0x1234..."
    },
    "transfer_type": "sale",
    "price": "2000000000000000000",
    "price_formatted": "2.0 ETH",
    "tx_hash": "0x9876543210fedcba...",
    "nft": {
      "id": 101,
      "title": "Digital Masterpiece #1",
      "image_url": "https://example.com/nft.jpg"
    }
  },
  "timestamp": "2023-12-01T18:05:00Z",
  "id": "msg_007"
}
```

### 2. NFT Listed for Sale

```json
{
  "event": "nft_listed",
  "data": {
    "nft_id": 101,
    "owner": {
      "id": 456,
      "username": "nft_seller",
      "avatar": "https://example.com/avatar.jpg"
    },
    "price": "1000000000000000000",
    "price_formatted": "1.0 ETH",
    "nft": {
      "id": 101,
      "title": "Digital Masterpiece #1",
      "image_url": "https://example.com/nft.jpg",
      "description": "A beautiful digital artwork"
    }
  },
  "timestamp": "2023-12-01T09:00:00Z",
  "id": "msg_008"
}
```

### 3. NFT Unlisted

```json
{
  "event": "nft_unlisted",
  "data": {
    "nft_id": 101,
    "owner": {
      "id": 456,
      "username": "nft_owner"
    },
    "nft": {
      "id": 101,
      "title": "Digital Masterpiece #1"
    }
  },
  "timestamp": "2023-12-01T11:00:00Z",
  "id": "msg_009"
}
```

## Notification Events

### 1. Personal Notification

```json
{
  "event": "notification",
  "data": {
    "id": 123,
    "type": "bid_outbid",
    "title": "You've been outbid!",
    "message": "Someone placed a higher bid on 'Digital Masterpiece #1'",
    "action_url": "/auction/123",
    "priority": "high",
    "read": false,
    "related_entity": {
      "type": "auction",
      "id": 123,
      "title": "Digital Masterpiece #1"
    }
  },
  "timestamp": "2023-12-01T10:30:00Z",
  "id": "msg_010"
}
```

### 2. System Announcement

```json
{
  "event": "system_announcement",
  "data": {
    "id": 456,
    "type": "maintenance",
    "title": "Scheduled Maintenance",
    "message": "The platform will undergo maintenance from 2 AM to 4 AM UTC",
    "start_time": "2023-12-02T02:00:00Z",
    "end_time": "2023-12-02T04:00:00Z",
    "severity": "info"
  },
  "timestamp": "2023-12-01T15:00:00Z",
  "id": "msg_011"
}
```

## User Activity Events

### 1. User Online Status

```json
{
  "event": "user_status_changed",
  "data": {
    "user_id": 789,
    "username": "crypto_collector",
    "status": "online",
    "last_seen": "2023-12-01T10:00:00Z"
  },
  "timestamp": "2023-12-01T10:00:00Z",
  "id": "msg_012"
}
```

### 2. User Profile Updated

```json
{
  "event": "user_profile_updated",
  "data": {
    "user_id": 789,
    "username": "crypto_collector",
    "changes": {
      "avatar": "https://example.com/new_avatar.jpg",
      "display_name": "Crypto Collector Pro"
    }
  },
  "timestamp": "2023-12-01T11:00:00Z",
  "id": "msg_013"
}
```

## Collection Events

### 1. New NFT Minted

```json
{
  "event": "nft_minted",
  "data": {
    "nft_id": 202,
    "collection_id": 10,
    "creator": {
      "id": 456,
      "username": "digital_artist",
      "avatar": "https://example.com/avatar.jpg"
    },
    "nft": {
      "id": 202,
      "title": "New Creation #202",
      "image_url": "https://example.com/nft202.jpg",
      "description": "Latest artwork from the collection",
      "token_id": "202",
      "contract_address": "0xabcd1234..."
    },
    "collection": {
      "id": 10,
      "name": "Digital Art Collection",
      "total_items": 202
    }
  },
  "timestamp": "2023-12-01T12:00:00Z",
  "id": "msg_014"
}
```

### 2. Collection Floor Price Changed

```json
{
  "event": "collection_floor_price_changed",
  "data": {
    "collection_id": 10,
    "collection_name": "Digital Art Collection",
    "old_floor_price": "500000000000000000",
    "old_floor_price_formatted": "0.5 ETH",
    "new_floor_price": "750000000000000000",
    "new_floor_price_formatted": "0.75 ETH",
    "change_percentage": 50.0,
    "volume_24h": "5000000000000000000",
    "volume_24h_formatted": "5.0 ETH"
  },
  "timestamp": "2023-12-01T13:00:00Z",
  "id": "msg_015"
}
```

## Error Events

### 1. Transaction Failed

```json
{
  "event": "transaction_failed",
  "data": {
    "transaction_id": "tx_123",
    "type": "bid_placement",
    "error_code": "INSUFFICIENT_FUNDS",
    "error_message": "Insufficient balance to place bid",
    "related_entity": {
      "type": "auction",
      "id": 123,
      "title": "Digital Masterpiece #1"
    },
    "retry_possible": true
  },
  "timestamp": "2023-12-01T14:00:00Z",
  "id": "msg_016"
}
```

### 2. Connection Error

```json
{
  "event": "connection_error",
  "data": {
    "error_code": "RATE_LIMIT_EXCEEDED",
    "error_message": "Too many requests, please slow down",
    "retry_after": 30
  },
  "timestamp": "2023-12-01T14:30:00Z",
  "id": "msg_017"
}
```

## Event Subscription Management

### Event Types

Clients can subscribe to the following event types:

- `bid_placed` - New bids on auctions
- `auction_status_changed` - Auction state changes
- `auction_extended` - Auction time extensions
- `nft_transferred` - NFT ownership transfers
- `nft_listed` - NFTs listed for sale
- `nft_unlisted` - NFTs removed from sale
- `nft_minted` - New NFTs created
- `notification` - Personal notifications
- `system_announcement` - Platform announcements
- `user_status_changed` - User online/offline status
- `user_profile_updated` - User profile changes
- `collection_floor_price_changed` - Collection floor price updates
- `transaction_failed` - Failed transactions
- `connection_error` - Connection-related errors

### Subscription Filters

Clients can filter events using the following criteria:

- `auction_ids`: Array of auction IDs to monitor
- `nft_ids`: Array of NFT IDs to monitor
- `collection_ids`: Array of collection IDs to monitor
- `user_ids`: Array of user IDs to monitor
- `creator_ids`: Array of creator IDs to monitor

### Example Subscription

```json
{
  "event": "subscribe",
  "data": {
    "types": ["bid_placed", "auction_status_changed", "nft_transferred"],
    "filters": {
      "auction_ids": [123, 456],
      "collection_ids": [10, 20],
      "user_ids": [789]
    }
  }
}
```

## Rate Limiting

- Maximum 100 messages per minute per connection
- Burst limit: 20 messages per 10 seconds
- Exceeded limits result in temporary connection throttling
- Persistent violations may result in connection termination

## Error Codes

### Connection Errors

- `4000`: Generic error
- `4001`: Authentication failed
- `4002`: Invalid token
- `4003`: Token expired
- `4004`: Rate limit exceeded
- `4005`: Invalid message format
- `4006`: Subscription limit exceeded

### Message Errors

- `4100`: Invalid event type
- `4101`: Missing required fields
- `4102`: Invalid filter criteria
- `4103`: Subscription not found

## Best Practices

### Client Implementation

1. **Reconnection Strategy**: Implement exponential backoff for reconnections
2. **Message Buffering**: Buffer messages during disconnections
3. **Heartbeat**: Send ping messages every 30 seconds to maintain connection
4. **Subscription Management**: Unsubscribe from events when no longer needed
5. **Error Handling**: Gracefully handle connection errors and invalid messages

### Server Implementation

1. **Connection Limits**: Limit concurrent connections per user
2. **Message Queuing**: Queue messages for offline users (with TTL)
3. **Event Filtering**: Efficiently filter events based on subscriptions
4. **Resource Management**: Clean up resources on connection close
5. **Monitoring**: Track connection metrics and event delivery rates

## Security Considerations

1. **Authentication**: Validate JWT tokens on every connection
2. **Authorization**: Ensure users only receive events they're authorized to see
3. **Rate Limiting**: Protect against abuse with proper rate limiting
4. **Input Validation**: Validate all incoming message formats
5. **Audit Logging**: Log all connection events and message exchanges

## Testing

### Unit Tests

- Message serialization/deserialization
- Event filtering logic
- Subscription management
- Rate limiting enforcement

### Integration Tests

- End-to-end event delivery
- Connection lifecycle management
- Error handling scenarios
- Performance under load

### Load Testing

- Concurrent connection limits
- Message throughput
- Memory usage under high load
- Connection recovery scenarios
