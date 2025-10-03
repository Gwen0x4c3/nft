# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Core Development Workflow
- `make setup` - Initial setup (deps + proto-gen)
- `make build` - Build both HTTP and gRPC servers
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make pre-commit` - Run all pre-commit checks (fmt + lint + test)

### Code Quality
- `make fmt` - Format code with gofmt and goimports
- `make lint` - Run golangci-lint
- `make lint-fix` - Run golangci-lint with auto-fix

### Running Services
- `make run-server` - Start HTTP server (localhost:8080)
- `make run-grpc-server` - Start gRPC server (localhost:9000)
- `make dev` - Run with hot reload using air

### Docker Development (Recommended)
- `make docker-dev` - Start infrastructure services (DB, Redis, Kafka, Blockchain)
- `make docker-dev-app` - Start infrastructure + application servers
- `make docker-dev-hot` - Start with hot reload for development
- `make docker-down` - Stop all services
- `make docker-down-clean` - Stop and clean volumes

### Database Operations
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback migrations
- `make migrate-create NAME=migration_name` - Create new migration

### Protocol Buffers
- `make proto-gen` - Generate Go code from proto files
- `make proto-lint` - Lint proto files with buf

### Testing
- `make test-contract` - Run contract tests only
- `make test-integration` - Run integration tests only

## Architecture Overview

### Project Structure
The project follows a clean architecture pattern with clear separation of concerns:

- **cmd/** - Application entry points (server, grpc-server, migrate)
- **internal/** - Private application code organized by layer
  - **config/** - Configuration management (database, redis, kafka, blockchain)
  - **models/** - Database models (GORM entities)
  - **repository/** - Data access layer (database operations)
  - **service/** - Business logic layer
  - **handlers/** - HTTP handlers for REST API
  - **middleware/** - HTTP middleware (auth, cors, logging, rate limiting)
  - **websocket/** - WebSocket management for real-time features
  - **blockchain/** - Ethereum blockchain integration
  - **messaging/** - Kafka event streaming
  - **types/** - Type definitions and enums
- **specs/** - Design specifications and contracts
- **tests/** - Test files organized by type (contract, integration, unit, performance)

### Key Technologies
- **Go 1.25.1** - Main programming language
- **Gin** - HTTP web framework for REST API
- **GORM** - ORM for PostgreSQL database operations
- **Redis** - Caching and session management
- **Apache Kafka** - Event streaming and messaging
- **go-ethereum** - Ethereum blockchain interaction
- **WebSockets** - Real-time communication
- **gRPC** - High-performance API endpoints
- **JWT** - Authentication tokens

### Database Architecture
- PostgreSQL as primary database
- GORM for ORM operations
- Migration system for schema management
- Repository pattern for data access

### Event-Driven Architecture
- Kafka for event streaming between services
- Event producers and consumers for async processing
- WebSocket manager for real-time client updates

### Blockchain Integration
- Ethereum Sepolia testnet for development
- Smart contract interaction through go-ethereum
- IPFS service for decentralized storage

### Configuration Management
- Environment-based configuration
- Separate config files for different services
- Support for development/production profiles

### Development Workflow
1. Use Docker Compose for local development environment
2. Test-Driven Development with contract and integration tests
3. Code quality enforced through golangci-lint
4. Protocol buffer generation for gRPC services
5. Migration-based database schema management

### API Design
- REST API with OpenAPI/Swagger documentation
- gRPC services for high-performance communication
- WebSocket events for real-time updates
- JWT-based authentication with middleware

### Testing Strategy
- Contract tests for API endpoints
- Integration tests for user flows
- Unit tests for business logic
- Performance tests for scalability

## Key Development Notes

### Environment Setup
- Copy `.env.example` to `.env` and configure with local values
- Use Docker Compose for infrastructure services
- Generate protobuf code before running servers

### Code Organization
- Follow repository pattern for data access
- Implement business logic in service layer
- Use dependency injection for clean architecture
- Separate types and models for clarity

### Real-time Features
- WebSocket manager handles client connections
- Event-driven updates through Kafka
- Notification system for user alerts

### Security Considerations
- JWT tokens for authentication
- Rate limiting middleware
- Input validation through Gin validators
- Environment variables for sensitive data