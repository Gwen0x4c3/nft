# NFT Platform

A comprehensive NFT minting and auction platform built with Go, featuring real-time bidding, blockchain integration, and WebSocket communication.

## Features

- **NFT Management**: Mint, transfer, and manage NFTs on Ethereum blockchain
- **Auction System**: Create and participate in real-time auctions with automatic bidding
- **Real-time Updates**: WebSocket-based real-time notifications and auction updates
- **User Authentication**: JWT-based authentication with session management
- **Event Streaming**: Kafka-based event streaming for scalable architecture
- **Caching**: Redis-based caching for improved performance
- **API Documentation**: Complete OpenAPI/Swagger documentation
- **gRPC Support**: High-performance gRPC endpoints alongside REST APIs

## Tech Stack

### Backend
- **Go 1.21+** - Main programming language
- **Gin** - HTTP web framework
- **GORM** - ORM for database operations
- **PostgreSQL** - Primary database
- **Redis** - Caching and session management
- **Apache Kafka** - Event streaming and messaging
- **go-ethereum** - Ethereum blockchain interaction
- **WebSockets** - Real-time communication
- **gRPC** - High-performance API endpoints

### Blockchain
- **Ethereum Sepolia Testnet** - Development blockchain
- **Solidity** - Smart contract development
- **OpenZeppelin** - Security-audited contract libraries

## Project Structure

```
├── cmd/                    # Application entry points
│   ├── server/            # HTTP server
│   └── grpc-server/       # gRPC server
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── models/           # Database models
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic layer
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middleware
│   ├── websocket/        # WebSocket management
│   ├── blockchain/       # Blockchain integration
│   ├── messaging/        # Kafka messaging
│   ├── grpc/            # gRPC services
│   └── proto/           # Generated protobuf code
├── pkg/                  # Public packages
├── configs/             # Configuration files
├── migrations/          # Database migrations
├── tests/               # Test files
│   ├── contract/       # Contract tests
│   ├── integration/    # Integration tests
│   ├── unit/          # Unit tests
│   └── performance/   # Performance tests
├── docs/               # Documentation
└── specs/              # Design specifications
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+
- Redis 6+
- Apache Kafka (or compatible)
- Node.js (for frontend development)
- Docker & Docker Compose (recommended)

### Environment Setup

1. Copy the environment template:
```bash
cp .env.example .env
```

2. Update the `.env` file with your configuration:
   - Database credentials
   - Redis connection details
   - Kafka brokers
   - Ethereum RPC endpoints
   - Private keys (for development only)

### Development Setup

1. Install dependencies:
```bash
go mod download
```

2. Generate protobuf code:
```bash
make proto-gen
```

3. Set up the database:
```bash
# Create database
createdb nft_platform

# Run migrations
make migrate-up
```

4. Start the services:
```bash
# Start HTTP server
make run-server

# Or start gRPC server
make run-grpc-server
```

### Docker Setup (Recommended)

The easiest way to get started is using Docker Compose:

1. Start development environment:
```bash
# Start infrastructure services (DB, Redis, Kafka, Blockchain)
make docker-dev

# Or use the helper script directly
./scripts/docker-dev.sh up
```

2. Start with application servers:
```bash
# Start infrastructure + application servers
make docker-dev-app

# Or start with hot reload for development
make docker-dev-hot
```

3. Stop services:
```bash
make docker-down

# Or clean stop (removes volumes)
make docker-down-clean
```

#### Available Services

Once started, the following services will be available:

- **PostgreSQL**: localhost:5432 (main) / localhost:5433 (test)
- **Redis**: localhost:6379
- **Kafka**: localhost:9092
- **Kafka UI**: http://localhost:8081
- **Hardhat Node**: http://localhost:8545
- **HTTP Server**: http://localhost:8080 (when app profile is active)
- **gRPC Server**: localhost:9000 (when app profile is active)
- **Prometheus**: http://localhost:9090 (when monitoring profile is active)
- **Grafana**: http://localhost:3000 (when monitoring profile is active)

#### Docker Profiles

The compose file uses profiles to organize services:

- **Default**: Infrastructure services (DB, Redis, Kafka, Blockchain)
- **app**: Application servers (HTTP + gRPC)
- **hot-reload**: Development with hot reload
- **monitoring**: Prometheus + Grafana
- **migration**: Database migration service
- **proxy**: Nginx reverse proxy

#### Helper Commands

Use the development helper script for common tasks:

```bash
# Show all available commands
./scripts/docker-dev.sh help

# View logs
./scripts/docker-dev.sh logs [service]

# Open shell in container
./scripts/docker-dev.sh shell postgres

# Database operations
./scripts/docker-dev.sh db-shell      # PostgreSQL CLI
./scripts/docker-dev.sh redis-shell   # Redis CLI
./scripts/docker-dev.sh kafka-topics  # List Kafka topics

# Clean everything
./scripts/docker-dev.sh clean
./scripts/docker-dev.sh reset         # Clean + rebuild
```

## Development

### Code Quality

We use several tools to maintain code quality:

- **golangci-lint** - Comprehensive Go linting
- **gofmt** - Code formatting
- **goimports** - Import management

Run code quality checks:
```bash
make lint
make fmt
```

### Testing

The project follows Test-Driven Development (TDD) principles:

```bash
# Run all tests
make test

# Run specific test suites
make test-contract      # Contract tests
make test-integration   # Integration tests

# Run with coverage
make test-coverage
```

### Available Commands

```bash
make help              # Show all available commands
make build             # Build the application
make test              # Run all tests
make lint              # Run linter
make fmt               # Format code
make proto-gen         # Generate protobuf code
make migrate-up        # Run database migrations
make migrate-down      # Rollback migrations
make clean             # Clean build artifacts
make docker-build      # Build Docker images
```

## API Documentation

### REST API
The REST API documentation is available via OpenAPI/Swagger. After starting the server, visit:
- http://localhost:8080/swagger/index.html

### gRPC API
gRPC services are documented in the proto files located in `specs/001-nft/contracts/`.

### WebSocket Events
WebSocket event documentation is available in `specs/001-nft/contracts/websocket-events.md`.

## Architecture

### Phase 3.1: Setup ✅
- [x] Project structure and dependencies
- [x] Database configuration (PostgreSQL + GORM)
- [x] Redis configuration and session management
- [x] Kafka event streaming setup
- [x] Blockchain client configuration
- [x] Code quality tools (golangci-lint, Makefile)

### Phase 3.2: Tests First (TDD)
- Contract tests for all API endpoints
- Integration tests for user flows
- WebSocket event testing
- gRPC contract tests

### Phase 3.3: Core Implementation
- Database models and migrations
- Repository pattern implementation
- Business logic services
- REST API handlers
- Authentication middleware

### Phase 3.4: Integration
- WebSocket real-time communication
- Blockchain smart contract integration
- Kafka event processing
- Server setup and routing

### Phase 3.5: Polish
- Comprehensive unit tests
- Performance optimization
- Security audit
- Documentation updates

## Contributing

1. Follow the existing code style and patterns
2. Write tests for all new functionality
3. Update documentation as needed
4. Run `make pre-commit` before submitting changes

## Security Considerations

- Never commit private keys or sensitive data
- Use environment variables for all secrets
- Follow blockchain security best practices
- Implement proper input validation
- Use HTTPS in production
- Regular security audits

## Monitoring and Logging

- Structured logging with correlation IDs
- Health check endpoints
- Prometheus metrics ready
- Error tracking and alerting

## License

[Add your license here]

## Support

For questions or support, please [create an issue](https://github.com/your-org/nft-platform/issues) on GitHub.