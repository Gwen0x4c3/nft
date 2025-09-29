# NFT Platform API Contracts

This directory contains the API contract specifications for the NFT platform, defining the interfaces between different services and components.

## Overview

The NFT platform uses a microservices architecture with gRPC for inter-service communication and WebSocket for real-time client updates. All service interfaces are defined using Protocol Buffers (protobuf) for type safety and cross-language compatibility.

## Service Contracts

### 1. User Service (`user-service.proto`)

Handles all user-related operations including authentication, profile management, and user preferences.

**Key Features:**
- User registration and authentication
- Profile management with avatar and bio
- Wallet address management
- User preferences and settings
- Follow/unfollow functionality
- Activity tracking

**Main RPCs:**
- `RegisterUser` - Create new user account
- `AuthenticateUser` - User login and JWT token generation
- `GetUserProfile` - Retrieve user profile information
- `UpdateUserProfile` - Update user profile data
- `LinkWallet` / `UnlinkWallet` - Manage wallet connections
- `FollowUser` / `UnfollowUser` - Social features

### 2. NFT Service (`nft-service.proto`)

Manages NFT collections, metadata, ownership, and marketplace operations.

**Key Features:**
- NFT creation and minting
- Collection management
- Metadata handling with IPFS integration
- Ownership tracking
- Marketplace listings and sales
- Auction functionality
- Search and filtering

**Main RPCs:**
- `CreateNFT` - Create and mint new NFTs
- `GetNFT` - Retrieve NFT details and metadata
- `ListNFTs` - Search and browse NFTs with filters
- `TransferNFT` - Handle NFT ownership transfers
- `CreateCollection` - Create NFT collections
- `ListForSale` / `RemoveFromSale` - Marketplace operations
- `CreateAuction` / `PlaceBid` - Auction functionality

### 3. Blockchain Service (`blockchain-service.proto`)

Provides blockchain interaction capabilities including smart contract operations, transaction management, and event monitoring.

**Key Features:**
- Smart contract deployment and interaction
- NFT minting, transfer, and burning operations
- Transaction submission and monitoring
- Gas estimation and price tracking
- Event monitoring and log retrieval
- Batch operations for efficiency
- Marketplace and auction contract interactions
- Royalty management
- Metadata operations

**Main RPCs:**
- `DeployContract` - Deploy new smart contracts
- `MintNFT` / `TransferNFT` / `BurnNFT` - Core NFT operations
- `SubmitTransaction` - Submit blockchain transactions
- `GetTransactionStatus` - Monitor transaction status
- `WatchEvents` - Real-time blockchain event monitoring
- `BatchTransfer` / `BatchMint` - Efficient batch operations
- `CreateMarketplaceListing` - Marketplace integration
- `PlaceBid` / `FinalizeAuction` - Auction operations

### 4. WebSocket Events (`websocket-events.md`)

Defines real-time event specifications for client-server communication via WebSocket connections.

**Key Features:**
- Real-time auction updates (bids, status changes)
- NFT transfer notifications
- User activity updates
- System announcements
- Personal notifications
- Collection events
- Transaction status updates
- Error handling and connection management

**Event Types:**
- `bid_placed` - New auction bids
- `auction_status_changed` - Auction state updates
- `nft_transferred` - Ownership changes
- `nft_listed` / `nft_unlisted` - Marketplace updates
- `notification` - Personal user notifications
- `system_announcement` - Platform-wide announcements
- `user_status_changed` - User online/offline status
- `collection_floor_price_changed` - Market data updates

## Data Models

### Common Types

All services share common data types and enums:

- **User**: Complete user profile with wallet addresses and preferences
- **NFT**: NFT metadata, ownership, and marketplace information
- **Collection**: NFT collection details and statistics
- **Transaction**: Blockchain transaction information
- **Auction**: Auction details with bidding history
- **Bid**: Individual bid information
- **Contract**: Smart contract metadata

### Enums

- **UserStatus**: `ACTIVE`, `SUSPENDED`, `DELETED`
- **NFTStatus**: `DRAFT`, `MINTED`, `LISTED`, `SOLD`, `BURNED`
- **AuctionStatus**: `DRAFT`, `ACTIVE`, `ENDED`, `CANCELLED`
- **TransactionStatus**: `PENDING`, `CONFIRMED`, `FAILED`, `REVERTED`
- **ContractType**: `ERC721`, `ERC1155`, `ERC20`, `CUSTOM`

## Authentication & Authorization

All services use JWT-based authentication with the following claims:
- `user_id`: Unique user identifier
- `wallet_address`: Primary wallet address
- `roles`: User roles and permissions
- `exp`: Token expiration time

## Error Handling

Standardized error codes across all services:
- `UNAUTHENTICATED`: Missing or invalid authentication
- `PERMISSION_DENIED`: Insufficient permissions
- `NOT_FOUND`: Requested resource not found
- `ALREADY_EXISTS`: Resource already exists
- `INVALID_ARGUMENT`: Invalid request parameters
- `RESOURCE_EXHAUSTED`: Rate limits or quotas exceeded
- `INTERNAL`: Server internal errors

## Rate Limiting

- **gRPC Services**: 1000 requests per minute per user
- **WebSocket Connections**: 100 messages per minute per connection
- **Blockchain Operations**: 10 transactions per minute per user

## Development Guidelines

### Protocol Buffer Best Practices

1. **Versioning**: Use semantic versioning for breaking changes
2. **Field Numbers**: Never reuse field numbers
3. **Naming**: Use snake_case for field names
4. **Documentation**: Include comments for all messages and fields
5. **Validation**: Use field options for validation rules

### Service Design Principles

1. **Single Responsibility**: Each service has a focused domain
2. **Stateless**: Services should be stateless for scalability
3. **Idempotent**: Operations should be idempotent where possible
4. **Graceful Degradation**: Handle partial failures gracefully
5. **Circuit Breakers**: Implement circuit breakers for external dependencies

### Testing Requirements

1. **Unit Tests**: Test all message serialization/deserialization
2. **Integration Tests**: Test service interactions end-to-end
3. **Contract Tests**: Verify API contracts between services
4. **Load Tests**: Test under expected production loads
5. **Chaos Tests**: Test resilience to failures

## Deployment Considerations

### Service Discovery

Services use consul for service discovery with health checks:
- Health check endpoints for all services
- Automatic service registration/deregistration
- Load balancing with health-aware routing

### Monitoring & Observability

- **Metrics**: Prometheus metrics for all services
- **Tracing**: Distributed tracing with Jaeger
- **Logging**: Structured logging with correlation IDs
- **Alerts**: Alert rules for critical service metrics

### Security

- **TLS**: All gRPC communications use TLS 1.3
- **Authentication**: JWT tokens with short expiration times
- **Authorization**: Role-based access control (RBAC)
- **Rate Limiting**: Per-user and per-service rate limits
- **Input Validation**: Strict validation of all inputs

## Future Enhancements

### Planned Features

1. **Multi-chain Support**: Extend blockchain service for multiple chains
2. **Advanced Analytics**: Add analytics service for platform metrics
3. **Social Features**: Expand social functionality (comments, likes)
4. **Mobile API**: Optimize APIs for mobile applications
5. **GraphQL Gateway**: Add GraphQL layer for flexible queries

### Performance Optimizations

1. **Caching**: Implement Redis caching for frequently accessed data
2. **Database Optimization**: Add read replicas and connection pooling
3. **CDN Integration**: Use CDN for static assets and metadata
4. **Batch Operations**: Expand batch operations for better efficiency
5. **Streaming**: Use gRPC streaming for large data transfers

## Getting Started

### Prerequisites

- Protocol Buffers compiler (protoc) v3.19+
- Go 1.19+ for Go service generation
- Node.js 16+ for JavaScript/TypeScript generation

### Code Generation

```bash
# Generate Go code
protoc --go_out=. --go-grpc_out=. *.proto

# Generate TypeScript code
protoc --plugin=protoc-gen-ts=./node_modules/.bin/protoc-gen-ts \
       --ts_out=service=grpc-web:. *.proto
```

### Service Implementation

Each service should implement:
1. gRPC server with all defined RPCs
2. Health check endpoint
3. Metrics collection
4. Proper error handling
5. Request/response logging

For detailed implementation examples, see the service implementation directories in the main codebase.
