# Tasks: NFT Platform

**Input**: Design documents from `/specs/001-nft/`
**Prerequisites**: research.md ✅, data-model.md ✅, contracts/ ✅

## Execution Flow (main)

```
1. Load available design documents:
   → research.md: Tech stack (Gin, GORM, Redis, Kafka)
   → data-model.md: 6 entities (User, NFT, Auction, Bid, Transfer, Notification)
   → contracts/: OpenAPI spec + WebSocket events + gRPC contracts + README.md
2. Generate tasks by category:
   → Setup: Go project, Gin framework, dependencies
   → Tests: 23 contract tests, 8 integration tests
   → Core: 6 models, 4 services, 15 endpoints
   → Integration: DB, middleware, WebSocket, blockchain
   → Polish: unit tests, performance, docs
3. Apply task rules:
   → Contract tests [P] - different files
   → Model creation [P] - different files
   → Sequential for shared service files
4. Number tasks T001-T089
5. TDD approach: tests before implementation
```

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Phase 3.1: Setup ✅

- [x] T001 Create Go project structure with cmd/, internal/, pkg/, configs/, migrations/ directories
- [x] T002 Initialize Go module with Gin, GORM, Redis, Kafka, and blockchain dependencies
- [x] T003 [P] Configure golangci-lint and gofmt for code quality
- [x] T004 [P] Setup PostgreSQL connection and GORM configuration in internal/config/database.go
- [x] T005 [P] Setup Redis connection configuration in internal/config/redis.go
- [x] T006 [P] Setup Kafka producer/consumer configuration in internal/config/kafka.go
- [x] T007 [P] Setup blockchain client configuration in internal/config/blockchain.go

## Phase 3.2: Tests First (TDD) ✅ COMPLETED

**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation** ✅ DONE

### Authentication Contract Tests

- [x] T008 [P] Contract test POST /auth/login in tests/contract/auth_login_test.go
- [x] T009 [P] Contract test POST /auth/register in tests/contract/auth_register_test.go
- [x] T010 [P] Contract test POST /auth/refresh in tests/contract/auth_refresh_test.go

### User Contract Tests

- [x] T011 [P] Contract test GET /users/profile in tests/contract/users_test.go
- [x] T012 [P] Contract test PUT /users/profile in tests/contract/users_test.go
- [x] T013 [P] Contract test GET /users/{userId} in tests/contract/users_test.go

### NFT Contract Tests

- [x] T014 [P] Contract test GET /nfts in tests/contract/nfts_test.go
- [x] T015 [P] Contract test POST /nfts in tests/contract/nfts_test.go
- [x] T016 [P] Contract test GET /nfts/{nftId} in tests/contract/nfts_test.go
- [x] T017 [P] Contract test PUT /nfts/{nftId} in tests/contract/nfts_test.go
- [x] T018 [P] Contract test POST /nfts/{nftId}/transfer in tests/contract/nfts_test.go

### Auction Contract Tests

- [x] T019 [P] Contract test GET /auctions in tests/contract/auctions_bids_test.go
- [x] T020 [P] Contract test POST /auctions in tests/contract/auctions_bids_test.go
- [x] T021 [P] Contract test GET /auctions/{auctionId} in tests/contract/auctions_bids_test.go
- [x] T022 [P] Contract test DELETE /auctions/{auctionId} in tests/contract/auctions_bids_test.go

### Bid Contract Tests

- [x] T023 [P] Contract test GET /auctions/{auctionId}/bids in tests/contract/auctions_bids_test.go
- [x] T024 [P] Contract test POST /auctions/{auctionId}/bids in tests/contract/auctions_bids_test.go

### Notification Contract Tests

- [x] T025 [P] Contract test GET /notifications in tests/contract/notifications_websocket_test.go
- [x] T026 [P] Contract test PUT /notifications/{notificationId}/read in tests/contract/notifications_websocket_test.go
- [x] T027 [P] Contract test PUT /notifications/read-all in tests/contract/notifications_websocket_test.go

### WebSocket Contract Tests

- [x] T028 [P] Contract test WebSocket /ws connection in tests/contract/notifications_websocket_test.go
- [x] T029 [P] Contract test WebSocket bid_placed event in tests/contract/notifications_websocket_test.go
- [x] T030 [P] Contract test WebSocket auction_status_changed event in tests/contract/notifications_websocket_test.go

### Integration Tests

- [x] T031 [P] Integration test user registration and login flow in tests/integration/flows_test.go
- [x] T032 [P] Integration test NFT minting and ownership in tests/integration/flows_test.go
- [x] T033 [P] Integration test auction creation and bidding in tests/integration/flows_test.go
- [x] T034 [P] Integration test NFT transfer and history in tests/integration/flows_test.go
- [x] T035 [P] Integration test notification delivery in tests/integration/flows_test.go
- [x] T036 [P] Integration test WebSocket real-time updates in tests/integration/flows_test.go
- [x] T037 [P] Integration test blockchain interaction in tests/integration/flows_test.go
- [x] T038 [P] Integration test auction ending and settlement in tests/integration/flows_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Database Models

- [x] T039 [P] User model with GORM tags in internal/models/user.go
- [x] T040 [P] NFT model with relationships in internal/models/nft.go
- [x] T041 [P] Auction model with bid tracking in internal/models/auction.go
- [x] T042 [P] Bid model with amount validation in internal/models/bid.go
- [x] T043 [P] Transfer model with history tracking in internal/models/transfer.go
- [x] T044 [P] Notification model with JSON data in internal/models/notification.go

### Database Migrations

- [x] T045 Create database migration for users table in migrations/001_create_users_table.sql
- [x] T046 Create database migration for nfts table in migrations/002_create_nfts_table.sql
- [x] T047 Create database migration for auctions table in migrations/003_create_auctions_table.sql
- [x] T048 Create database migration for bids table in migrations/004_create_bids_table.sql
- [x] T049 Create database migration for transfers table in migrations/005_create_transfers_table.sql
- [x] T050 Create database migration for notifications table in migrations/006_create_notifications_table.sql
- [x] T051 Create database indexes migration in migrations/007_create_indexes.sql

### Repository Layer

- [x] T052 [P] User repository with CRUD operations in internal/repository/user_repository.go
- [x] T053 [P] NFT repository with filtering in internal/repository/nft_repository.go
- [x] T054 [P] Auction repository with status queries in internal/repository/auction_repository.go
- [x] T055 [P] Notification repository with pagination in internal/repository/notification_repository.go

### Service Layer

- [x] T056 User service with authentication logic in internal/service/user_service.go
- [x] T057 NFT service with minting and transfer logic in internal/service/nft_service.go
- [x] T058 Auction service with bidding logic in internal/service/auction_service.go
- [x] T059 Notification service with real-time delivery in internal/service/notification_service.go

### Middleware

- [x] T060 [P] JWT authentication middleware in internal/middleware/auth.go
- [x] T061 [P] Rate limiting middleware in internal/middleware/rate_limit.go
- [x] T062 [P] CORS middleware configuration in internal/middleware/cors.go
- [x] T063 [P] Request logging middleware in internal/middleware/logging.go

### API Handlers - Authentication

- [x] T064 POST /auth/login handler in internal/handlers/auth_handler.go
- [x] T065 POST /auth/register handler in internal/handlers/auth_handler.go
- [x] T066 POST /auth/refresh handler in internal/handlers/auth_handler.go

### API Handlers - Users

- [x] T067 GET /users/profile handler in internal/handlers/user_handler.go
- [x] T068 PUT /users/profile handler in internal/handlers/user_handler.go
- [x] T069 GET /users/{userId} handler in internal/handlers/user_handler.go

### API Handlers - NFTs

- [x] T070 GET /nfts handler with filtering in internal/handlers/nft_handler.go
- [x] T071 POST /nfts handler with file upload in internal/handlers/nft_handler.go
- [x] T072 GET /nfts/{nftId} handler in internal/handlers/nft_handler.go
- [x] T073 PUT /nfts/{nftId} handler in internal/handlers/nft_handler.go
- [x] T074 POST /nfts/{nftId}/transfer handler in internal/handlers/nft_handler.go

### API Handlers - Auctions & Bids

- [x] T075 GET /auctions handler with status filtering in internal/handlers/auction_handler.go
- [x] T076 POST /auctions handler in internal/handlers/auction_handler.go
- [x] T077 GET /auctions/{auctionId} handler in internal/handlers/auction_handler.go
- [x] T078 DELETE /auctions/{auctionId} handler in internal/handlers/auction_handler.go
- [x] T079 GET /auctions/{auctionId}/bids handler in internal/handlers/bid_handler.go
- [x] T080 POST /auctions/{auctionId}/bids handler in internal/handlers/bid_handler.go

### API Handlers - Notifications

- [x] T081 GET /notifications handler in internal/handlers/notification_handler.go
- [x] T082 PUT /notifications/{notificationId}/read handler in internal/handlers/notification_handler.go
- [x] T083 PUT /notifications/read-all handler in internal/handlers/notification_handler.go

## Phase 3.4: Integration

- [ ] T084 WebSocket connection manager in internal/websocket/manager.go
- [ ] T085 WebSocket event broadcasting in internal/websocket/events.go
- [ ] T086 Blockchain client with smart contract interaction in internal/blockchain/client.go
- [ ] T087 Kafka message producer for events in internal/messaging/producer.go
- [ ] T088 Kafka message consumer for blockchain events in internal/messaging/consumer.go
- [ ] T089 Main application server setup in cmd/server/main.go

## Phase 3.5: Polish

- [ ] T090 [P] Unit tests for user validation in tests/unit/user_validation_test.go
- [ ] T091 [P] Unit tests for auction logic in tests/unit/auction_logic_test.go
- [ ] T092 [P] Unit tests for bid validation in tests/unit/bid_validation_test.go
- [ ] T093 [P] Performance tests for auction endpoints (<200ms) in tests/performance/auction_perf_test.go
- [ ] T094 [P] Load tests for WebSocket connections in tests/performance/websocket_load_test.go
- [ ] T095 [P] Update API documentation in docs/api.md
- [ ] T096 [P] Update deployment guide in docs/deployment.md
- [ ] T097 Remove code duplication and optimize imports
- [ ] T098 Run comprehensive integration testing scenarios
- [ ] T099 Security audit and vulnerability testing

## Dependencies

### Phase Dependencies

- Setup (T001-T007) before all other phases
- Tests (T008-T038) before implementation (T039-T089)
- Models (T039-T044) before repositories (T052-T055)
- Repositories before services (T056-T059)
- Services before handlers (T064-T083)
- Integration (T084-T089) after core implementation
- Polish (T090-T099) after integration

### Specific Blocking Dependencies

- T039-T044 (models) block T052-T055 (repositories)
- T052-T055 (repositories) block T056-T059 (services)
- T056-T059 (services) block T064-T083 (handlers)
- T060-T063 (middleware) block T089 (main server)
- T084-T088 (integration) block T089 (main server)

## Parallel Example

```bash
# Phase 3.2 - Contract Tests (can run all together):
Task: "Contract test POST /auth/login in tests/contract/auth_login_test.go"
Task: "Contract test POST /auth/register in tests/contract/auth_register_test.go"
Task: "Contract test GET /users/profile in tests/contract/users_profile_test.go"
Task: "Contract test GET /nfts in tests/contract/nfts_list_test.go"

# Phase 3.3 - Models (can run all together):
Task: "User model with GORM tags in internal/models/user.go"
Task: "NFT model with relationships in internal/models/nft.go"
Task: "Auction model with bid tracking in internal/models/auction.go"
```

## Notes

- [P] tasks = different files, no dependencies
- Verify all contract tests fail before implementing
- Blockchain integration requires Sepolia testnet connection
- WebSocket requires goroutine-based connection management
- Use big.Int for cryptocurrency amounts (Wei)
- Implement proper error handling for all blockchain operations
- Redis used for session management and caching
- Kafka used for event streaming and notifications

## Validation Checklist

_GATE: Checked before task execution_

- [x] All 23 API endpoints have corresponding contract tests
- [x] All 6 entities have model creation tasks
- [x] All contract tests come before implementation
- [x] Parallel tasks target different files
- [x] Each task specifies exact file path
- [x] WebSocket events covered in contract tests
- [x] Blockchain integration included
- [x] Database migrations planned
- [x] TDD approach maintained throughout

---

## Additional gRPC/Proto Tasks

### Setup and Code Generation

- [ ] T100 [P] Install protoc and plugins (protoc-gen-go, protoc-gen-go-grpc, optional: buf) and add Makefile targets in Makefile
- [ ] T101 [P] Generate Go code from protos to internal/proto/ (user, nft, auction, blockchain, notification)

### gRPC Contract Tests (one per proto contract file)

- [ ] T102 [P] gRPC contract tests for user-service.proto in tests/contract/grpc/user_service_contract_test.go
- [ ] T103 [P] gRPC contract tests for nft-service.proto in tests/contract/grpc/nft_service_contract_test.go
- [ ] T104 [P] gRPC contract tests for auction-service.proto in tests/contract/grpc/auction_service_contract_test.go
- [ ] T105 [P] gRPC contract tests for blockchain-service.proto in tests/contract/grpc/blockchain_service_contract_test.go
- [ ] T106 [P] gRPC contract tests for notification-service.proto in tests/contract/grpc/notification_service_contract_test.go

### gRPC Server Implementation

- [ ] T107 Implement UserService gRPC server in internal/grpc/user/server.go
- [ ] T108 Implement NFTService gRPC server in internal/grpc/nft/server.go
- [ ] T109 Implement AuctionService gRPC server in internal/grpc/auction/server.go
- [ ] T110 Implement BlockchainService gRPC server in internal/grpc/blockchain/server.go
- [ ] T111 Implement NotificationService gRPC server in internal/grpc/notification/server.go
- [ ] T112 Bootstrap gRPC server with reflection and interceptors in cmd/grpc-server/main.go
- [ ] T113 [P] gRPC health check service implementation in internal/grpc/health/health.go
- [ ] T114 Register all gRPC services and wire dependencies in internal/grpc/server/register.go

### CI, Docs, and Examples

- [ ] T115 [P] Add proto linting to CI (buf or protolint) with config at .proto-lint or buf.yaml
- [ ] T116 [P] Document gRPC endpoints and usage in docs/grpc.md

### Parallel Example (gRPC)

```bash
# Run these gRPC contract tests in parallel
Task: "gRPC contract tests for user-service.proto in tests/contract/grpc/user_service_contract_test.go"
Task: "gRPC contract tests for nft-service.proto in tests/contract/grpc/nft_service_contract_test.go"
Task: "gRPC contract tests for auction-service.proto in tests/contract/grpc/auction_service_contract_test.go"
```

## Updated Validation Checklist (including gRPC)

- [x] All 5 proto contracts have corresponding gRPC contract test tasks
- [x] Protos are compiled and codegen targets exist
- [x] gRPC servers implemented and registered
- [x] Health checks available for gRPC
- [x] gRPC documentation added

---

**Task Status**: ✅ Updated - 116 tasks generated with REST + gRPC coverage and proper dependencies/parallelization
