# NFT Platform API Documentation

**Version**: 1.0.0
**Base URL**: `http://localhost:8080/api/v1`
**Authentication**: JWT Bearer Token

## Overview

The NFT Platform API provides endpoints for creating, managing, and trading NFTs through auction-based marketplace functionality. It includes user authentication, NFT minting, auction management, bidding, and real-time notifications via WebSocket.

## Authentication

All protected endpoints require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

### Authentication Endpoints

#### Login
```http
POST /auth/login
```

**Request Body:**
```json
{
  "wallet_address": "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
  "signature": "0x1234567890abcdef...",
  "message": "Sign this message to authenticate"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "wallet_address": "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
    "avatar": "https://example.com/avatar.jpg",
    "bio": "Test user bio",
    "is_verified": false,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### Register
```http
POST /auth/register
```

**Request Body:**
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "wallet_address": "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
  "signature": "0x1234567890abcdef...",
  "message": "Sign this message to register",
  "avatar": "https://example.com/avatar.jpg",
  "bio": "Test user bio"
}
```

**Response:** Same as login endpoint

#### Refresh Token
```http
POST /auth/refresh
```

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response:** Same as login endpoint

## User Endpoints

### Get Current User Profile
```http
GET /users/profile
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 1,
  "username": "testuser",
  "email": "test@example.com",
  "wallet_address": "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
  "avatar": "https://example.com/avatar.jpg",
  "bio": "Test user bio",
  "is_verified": false,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Update User Profile
```http
PUT /users/profile
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "username": "newusername",
  "avatar": "https://example.com/new-avatar.jpg",
  "bio": "Updated bio"
}
```

**Response:** Updated user profile

### Get User by ID
```http
GET /users/{userId}
```

**Response:** User profile (public information only)

## NFT Endpoints

### List NFTs
```http
GET /nfts?page=1&limit=20&creator_id=1&owner_id=2&is_for_sale=true&sort=created_at&order=desc
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `creator_id` (int): Filter by creator ID
- `owner_id` (int): Filter by owner ID
- `is_for_sale` (boolean): Filter NFTs for sale
- `sort` (string): Sort field (created_at, price, title)
- `order` (string): Sort order (asc, desc)

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "token_id": "123",
      "contract_address": "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
      "creator_id": 1,
      "owner_id": 2,
      "title": "My NFT",
      "description": "A beautiful NFT artwork",
      "image_url": "https://example.com/nft-image.jpg",
      "metadata_uri": "https://example.com/metadata.json",
      "price": "1000000000000000000",
      "is_for_sale": true,
      "royalty": 5,
      "mint_tx_hash": "0x1234567890abcdef...",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "creator": {
        "id": 1,
        "username": "creator",
        "wallet_address": "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D"
      },
      "owner": {
        "id": 2,
        "username": "owner",
        "wallet_address": "0x842d35Cc6634C0532925a3b8D4E7E0E0e9e0dF2D"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

### Create NFT (Mint)
```http
POST /nfts
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

**Form Data:**
- `title` (string, required): NFT title (1-100 characters)
- `description` (string, optional): NFT description (max 1000 characters)
- `image` (file, required): NFT image file
- `metadata` (object, required): NFT metadata JSON
- `royalty` (int, optional): Creator royalty percentage (0-10, default: 0)

**Response:**
```json
{
  "id": 1,
  "token_id": "123",
  "contract_address": "0x742d35Cc6634C0532925a3b8D4E7E0E0e9e0dF1D",
  "creator_id": 1,
  "owner_id": 1,
  "title": "My NFT",
  "description": "A beautiful NFT artwork",
  "image_url": "https://example.com/nft-image.jpg",
  "metadata_uri": "https://example.com/metadata.json",
  "price": null,
  "is_for_sale": false,
  "royalty": 5,
  "mint_tx_hash": "0x1234567890abcdef...",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### Get NFT by ID
```http
GET /nfts/{nftId}
```

**Response:** Complete NFT details with relationships

### Update NFT
```http
PUT /nfts/{nftId}
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "price": "2000000000000000000",
  "is_for_sale": true
}
```

**Response:** Updated NFT details

### Transfer NFT
```http
POST /nfts/{nftId}/transfer
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "to_address": "0x842d35Cc6634C0532925a3b8D4E7E0E0e9e0dF2D"
}
```

**Response:**
```json
{
  "id": 1,
  "nft_id": 1,
  "from_id": 1,
  "to_id": 2,
  "tx_hash": "0x1234567890abcdef...",
  "price": "2000000000000000000",
  "type": "transfer",
  "created_at": "2024-01-01T00:00:00Z"
}
```

## Auction Endpoints

### List Auctions
```http
GET /auctions?page=1&limit=20&status=active&seller_id=1
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `status` (string): Filter by status (pending, active, ended, cancelled)
- `seller_id` (int): Filter by seller ID

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "nft_id": 1,
      "seller_id": 1,
      "start_price": "1000000000000000000",
      "reserve_price": "1500000000000000000",
      "current_bid": "2000000000000000000",
      "start_time": "2024-01-01T12:00:00Z",
      "end_time": "2024-01-02T12:00:00Z",
      "status": "active",
      "winner_id": null,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "nft": {
        "id": 1,
        "title": "Auction NFT",
        "image_url": "https://example.com/nft-image.jpg"
      },
      "seller": {
        "id": 1,
        "username": "seller"
      },
      "bids": [
        {
          "id": 1,
          "auction_id": 1,
          "bidder_id": 2,
          "amount": "2000000000000000000",
          "status": "confirmed",
          "created_at": "2024-01-01T13:00:00Z"
        }
      ]
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3,
    "has_next": true,
    "has_prev": false
  }
}
```

### Create Auction
```http
POST /auctions
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "nft_id": 1,
  "start_price": "1000000000000000000",
  "reserve_price": "1500000000000000000",
  "start_time": "2024-01-01T12:00:00Z",
  "end_time": "2024-01-02T12:00:00Z"
}
```

**Response:** Complete auction details

### Get Auction by ID
```http
GET /auctions/{auctionId}
```

**Response:** Complete auction details with NFT, seller, and bids

### Cancel Auction
```http
DELETE /auctions/{auctionId}
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "Auction cancelled successfully",
  "auction_id": 1
}
```

## Bid Endpoints

### Get Auction Bids
```http
GET /auctions/{auctionId}/bids?page=1&limit=50
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "auction_id": 1,
      "bidder_id": 2,
      "amount": "2000000000000000000",
      "status": "confirmed",
      "tx_hash": "0x1234567890abcdef...",
      "created_at": "2024-01-01T13:00:00Z",
      "updated_at": "2024-01-01T13:00:00Z",
      "bidder": {
        "id": 2,
        "username": "bidder"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 25,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

### Place Bid
```http
POST /auctions/{auctionId}/bids
Authorization: Bearer <token>
```

**Request Body:**
```json
{
  "amount": "2500000000000000000"
}
```

**Response:**
```json
{
  "bid": {
    "id": 2,
    "auction_id": 1,
    "bidder_id": 3,
    "amount": "2500000000000000000",
    "status": "confirmed",
    "tx_hash": "0x2345678901bcdef...",
    "created_at": "2024-01-01T14:00:00Z"
  },
  "is_highest": true,
  "previous_high_bid": "2000000000000000000",
  "auction_updated": true
}
```

## Notification Endpoints

### Get User Notifications
```http
GET /notifications?page=1&limit=50&unread_only=false
Authorization: Bearer <token>
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 50, max: 100)
- `unread_only` (boolean): Filter only unread notifications

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "type": "bid_placed",
      "title": "New Bid Received",
      "message": "A new bid of 2.5 ETH has been placed on your auction",
      "data": {
        "auction_id": 1,
        "nft_id": 1,
        "nft_title": "Auction NFT",
        "bid_id": 2,
        "amount": "2500000000000000000",
        "bidder_id": 3
      },
      "is_read": false,
      "created_at": "2024-01-01T14:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 10,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  }
}
```

### Mark Notification as Read
```http
PUT /notifications/{notificationId}/read
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "Notification marked as read"
}
```

### Mark All Notifications as Read
```http
PUT /notifications/read-all
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "All notifications marked as read"
}
```

## WebSocket Events

### Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=<access_token>');
```

### Event Types

#### Connection Established
```json
{
  "event": "connection_established",
  "data": {
    "id": "connection-uuid"
  },
  "timestamp": "2024-01-01T12:00:00Z",
  "id": "event-uuid"
}
```

#### Bid Placed
```json
{
  "event": "bid_placed",
  "data": {
    "auction_id": 1,
    "nft_id": 1,
    "nft_title": "Auction NFT",
    "bid_id": 2,
    "amount": "2500000000000000000",
    "bidder_id": 3,
    "bidder_username": "bidder"
  },
  "timestamp": "2024-01-01T14:00:00Z",
  "id": "event-uuid"
}
```

#### Auction Status Changed
```json
{
  "event": "auction_status_changed",
  "data": {
    "auction_id": 1,
    "nft_id": 1,
    "nft_title": "Auction NFT",
    "old_status": "active",
    "new_status": "ended",
    "winner_id": 3,
    "final_bid": "3000000000000000000"
  },
  "timestamp": "2024-01-02T12:00:00Z",
  "id": "event-uuid"
}
```

#### NFT Transferred
```json
{
  "event": "nft_transferred",
  "data": {
    "nft_id": 1,
    "nft_title": "Transferred NFT",
    "from_user_id": 1,
    "to_user_id": 2,
    "tx_hash": "0x1234567890abcdef...",
    "price": "3000000000000000000"
  },
  "timestamp": "2024-01-01T15:00:00Z",
  "id": "event-uuid"
}
```

#### Personal Notification
```json
{
  "event": "notification",
  "data": {
    "id": 1,
    "type": "bid_outbid",
    "title": "You've Been Outbid",
    "message": "Your bid on Auction NFT has been exceeded",
    "data": {
      "auction_id": 1,
      "nft_id": 1,
      "new_bid_amount": "2500000000000000000"
    }
  },
  "timestamp": "2024-01-01T14:00:00Z",
  "id": "event-uuid"
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "additional": "context"
  }
}
```

### Common Error Codes

- `UNAUTHORIZED` (401): Authentication required or invalid token
- `FORBIDDEN` (403): Insufficient permissions
- `NOT_FOUND` (404): Resource not found
- `INVALID_REQUEST` (400): Invalid request parameters
- `CONFLICT` (409): Resource conflict (e.g., duplicate)
- `INTERNAL_ERROR` (500): Server error

## Rate Limiting

API endpoints are rate-limited to prevent abuse:
- **Authentication endpoints**: 10 requests per minute per IP
- **General endpoints**: 100 requests per minute per user
- **WebSocket connections**: 100 messages per minute per connection

## Data Types

### Monetary Values
All monetary values are represented as strings in Wei (the smallest unit of Ether):
- 1 ETH = 1,000,000,000,000,000,000 Wei
- Example: `"1000000000000000000"` = 1 ETH

### Time Format
All timestamps use ISO 8601 format in UTC:
- Example: `"2024-01-01T12:00:00Z"`

### Pagination
All list endpoints support pagination with these response fields:
- `page`: Current page number
- `limit`: Items per page
- `total`: Total number of items
- `total_pages`: Total number of pages
- `has_next`: Whether there are more pages
- `has_prev`: Whether there are previous pages

## SDK and Libraries

### JavaScript/TypeScript
```bash
npm install nft-platform-js-sdk
```

### Go
```bash
go get github.com/nft-platform/go-sdk
```

### Python
```bash
pip install nft-platform-py-sdk
```

## Support

For API support and documentation updates:
- GitHub Issues: [nft-platform/issues](https://github.com/nft-platform/issues)
- Documentation: [docs.nftplatform.com](https://docs.nftplatform.com)
- Developer Discord: [discord.gg/nftplatform](https://discord.gg/nftplatform)

## Changelog

### v1.0.0 (2024-01-01)
- Initial API release
- Authentication endpoints
- NFT management
- Auction functionality
- Bidding system
- Real-time notifications via WebSocket
- Comprehensive error handling

---

**Last Updated**: January 1, 2024
**API Version**: 1.0.0