# NFT Platform Testing Guide

**Version**: 1.0.0
**Last Updated**: January 1, 2024

This guide provides comprehensive information about testing strategies, test types, and how to run various test scenarios for the NFT Platform.

## Table of Contents

1. [Test Strategy Overview](#test-strategy-overview)
2. [Test Types](#test-types)
3. [Running Tests](#running-tests)
4. [Test Scenarios](#test-scenarios)
5. [Test Coverage](#test-coverage)
6. [Performance Testing](#performance-testing)
7. [Debugging Tests](#debugging-tests)
8. [CI/CD Integration](#cicd-integration)

## Test Strategy Overview

The NFT Platform follows a comprehensive testing strategy that ensures high-quality, reliable software:

### Testing Pyramid

1. **Unit Tests** (70%): Fast, isolated tests for individual components
2. **Integration Tests** (20%): Tests for component interactions
3. **End-to-End Tests** (10%): Complete user journey tests

### Testing Principles

- **Test-Driven Development (TDD)**: Write tests before implementation
- **Red-Green-Refactor**: Continuous test improvement cycle
- **Independent Tests**: Tests should not depend on each other
- **Deterministic Tests**: Tests should produce consistent results
- **Fast Feedback**: Tests should run quickly for development efficiency

## Test Types

### 1. Unit Tests

**Location**: `tests/unit/`

**Purpose**: Test individual functions, methods, and components in isolation.

**Characteristics**:
- Fast execution (milliseconds)
- Isolated from external dependencies
- Mock external services
- High coverage target (80%+)

**Example**:
```bash
# Run all unit tests
make test

# Run specific unit test file
go test ./tests/unit/user_validation_test.go -v

# Run with coverage
go test -cover ./tests/unit/... -coverprofile=unit-coverage.out
```

### 2. Contract Tests

**Location**: `tests/contract/`

**Purpose**: Test API contracts and HTTP interfaces.

**Characteristics**:
- Test API request/response formats
- Validate HTTP status codes
- Test error handling
- Fast execution

**Example**:
```bash
# Run all contract tests
go test ./tests/contract/... -v

# Run specific contract test
go test ./tests/contract/auth_login_test.go -v
```

### 3. Integration Tests

**Location**: `tests/integration/`

**Purpose**: Test interactions between different components and services.

**Characteristics**:
- Test real database connections
- Test service interactions
- Test API endpoint chains
- Medium execution time (seconds)

**Example**:
```bash
# Run all integration tests
make test-integration

# Run with timeout
go test ./tests/integration/... -timeout=30m
```

### 4. Performance Tests

**Location**: `tests/performance/`

**Purpose**: Test system performance under load.

**Characteristics**:
- Benchmark response times
- Test concurrent load
- Memory and CPU usage monitoring
- Extended execution time

**Example**:
```bash
# Run performance tests
go test -bench=. ./tests/performance/... -benchtime=10s

# Run WebSocket load tests
go test -v ./tests/performance/websocket_load_test.go
```

### 5. Comprehensive Integration Scenarios

**Location**: `tests/integration/comprehensive_scenarios_test.go`

**Purpose**: Test complete user journeys and business workflows.

**Characteristics**:
- End-to-end user flows
- Real-time interactions
- WebSocket integration
- Complex multi-step scenarios

## Running Tests

### Individual Test Types

```bash
# Unit tests only
go test ./tests/unit/...

# Contract tests only
go test ./tests/contract/...

# Integration tests only
go test ./tests/integration/...

# Performance tests only
go test ./tests/performance/...

# Comprehensive scenarios only
go test ./tests/integration/comprehensive_scenarios_test.go
```

### Combined Test Execution

```bash
# All tests
make test

# Integration tests only
make test-integration

# Comprehensive scenarios only
make test-comprehensive

# All tests with detailed reporting
make test-all
```

### Test Execution Options

```bash
# Verbose output
go test -v ./...

# Run specific test file
go test -v ./tests/unit/user_validation_test.go

# Run with coverage
go test -cover ./tests/... -coverprofile=coverage.out

# Run with race detection
go test -race ./tests/...

# Run with timeout
go test -timeout=30m ./tests/integration/...
```

## Test Scenarios

### 1. Complete User Workflow

**File**: `comprehensive_scenarios_test.go - TestCompleteUserWorkflow`

**Purpose**: Test the complete user registration and NFT minting process.

**Steps**:
1. User registration with wallet authentication
2. User login and token generation
3. User profile management
4. NFT creation (minting)
5. NFT management and updates

**Expected Results**:
- User successfully registers and authenticates
- NFT is minted and associated with user
- All operations complete within reasonable time limits
- Error handling works correctly

### 2. Complete Auction Workflow

**File**: `comprehensive_scenarios_test.go - TestCompleteAuctionWorkflow`

**Purpose**: Test the complete auction lifecycle including bidding.

**Steps**:
1. Seller creates auction for NFT
2. Multiple bidders place bids
3. Bid validation and minimum increments
4. Auction status updates
5. Notifications for bid events

**Expected Results**:
- Auction is created successfully
- Bids are validated and processed
- Highest bid updates correctly
- All participants receive appropriate notifications

### 3. Real-time WebSocket Integration

**File**: `comprehensive_scenarios_test.go - TestRealtimeWebSocketIntegration`

**Purpose**: Test WebSocket functionality and real-time event delivery.

**Steps**:
1. Users establish WebSocket connections
2. Server broadcasts auction events
3. Clients receive real-time updates
4. Connection stability under load
5. Message delivery verification

**Expected Results**:
- WebSocket connections establish successfully
- Real-time events are delivered promptly
- Multiple connections are handled simultaneously
- Connection cleanup works properly

### 4. NFT Transfer Workflow

**File**: `comprehensive_scenarios_test.go - TestNFTTransferWorkflow`

**Purpose**: Test NFT ownership transfer between users.

**Steps**:
1. Current owner initiates transfer
2. Ownership is validated and transferred
3. Transfer history is recorded
4. Both parties receive notifications
5. NFT metadata is updated

**Expected Results**:
- Transfer is processed successfully
- Ownership changes are reflected immediately
- Transfer history is maintained
- All stakeholders are notified

### 5. Error Handling and Edge Cases

**File**: `comprehensive_scenarios_test.go - TestErrorHandlingAndEdgeCases`

**Purpose**: Test system behavior under error conditions.

**Test Cases**:
- Invalid authentication attempts
- Unauthorized resource access
- Invalid input data handling
- Resource not found scenarios
- Constraint violations
- Network failure simulation

**Expected Results**:
- Appropriate HTTP status codes returned
- Error messages are descriptive and helpful
- System remains stable under error conditions
- No sensitive information leaked

### 6. Performance and Load Testing

**File**: `comprehensive_scenarios_test.go - TestPerformanceAndLoad`

**Purpose**: Test system performance under various load conditions.

**Test Scenarios**:
- Concurrent user registrations
- Rapid API call sequences
- WebSocket connection load
- Database query performance
- Memory usage under load

**Expected Results**:
- System handles concurrent operations efficiently
- Response times remain within acceptable limits
- Memory usage stays within expected bounds
- No resource exhaustion occurs

## Test Coverage

### Coverage Goals

- **Unit Tests**: 80%+ code coverage
- **Integration Tests**: 70%+ code coverage
- **End-to-End Tests**: All critical user journeys

### Generating Coverage Reports

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Generate coverage for specific package
go test -coverprofile=coverage.out ./internal/service/...
go tool cover -html=coverage.out -o service-coverage.html
```

### Coverage Analysis

```bash
# View coverage summary
go tool cover -func=coverage.out

# View uncovered functions
go tool cover -mode=count coverage.out

# Identify low-coverage areas
go tool cover -mode=set coverage.out | grep -E "missed|uncovered"
```

## Performance Testing

### Response Time Targets

- **API Endpoints**: <200ms (95th percentile)
- **WebSocket Events**: <100ms
- **Database Queries**: <50ms
- **Authentication**: <300ms

### Load Testing Scenarios

```bash
# Benchmark specific endpoint
go test -bench=BenchmarkGetNFTs ./tests/performance/...

# Benchmark with custom parameters
go test -bench=BenchmarkListAuctions -benchtime=30s ./tests/performance/...

# Run performance tests
make test-performance
```

### Monitoring During Tests

```bash
# Monitor CPU and memory usage
top -p $(pgrep -f "go test") -o %CPU,%MEM

# Monitor network connections
netstat -tulpn | grep :8080

# Monitor database connections
psql -c "SELECT count(*) FROM pg_stat_activity WHERE state = 'active';"
```

## Debugging Tests

### Running Tests in Debug Mode

```bash
# Run with detailed output
go test -v -run=TestSpecificFunction ./tests/unit/user_validation_test.go

# Run with race detection
go test -race ./tests/unit/user_validation_test.go

# Run with specific test data
USER_ID=1 go test -v ./tests/unit/user_validation_test.go
```

### Test Isolation

```bash
# Run tests in parallel
go test -parallel 4 ./tests/unit/...

# Clean test cache
go clean -testcache
```

### Common Debugging Issues

#### Import Errors
```bash
# Check for missing dependencies
go mod tidy
go mod download

# Verify module structure
go list ./...
```

#### Database Connection Issues
```bash
# Check database status
pg_isready -h localhost -p 5432 -U nft_platform

# Check connection logs
tail -f /var/log/postgresql/postgresql-15-main.log
```

#### WebSocket Issues
```bash
# Test WebSocket connection
wscat -c "localhost:8080/ws" -H "Authorization: Bearer <token>"

# Check WebSocket server logs
tail -f /var/log/nft-platform/websocket.log
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install dependencies
        run: go mod download
      - name: Run tests
        run: make test-all
      - name: Upload coverage reports
        uses: actions/upload-artifact@v3
        with:
          name: coverage-reports
          path: test-results/
```

### Docker Test Environment

```dockerfile
FROM golang:1.21-alpine AS test

WORKDIR /app
COPY . .
COPY go.mod go.sum ./
RUN go mod download

# Install test dependencies
RUN go install github.com/golang/mock/mockgen@latest

# Run tests
CMD ["make", "test-all"]
```

### Kubernetes Test Jobs

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: nft-platform-tests
spec:
  template:
    spec:
      containers:
      - name: test-runner
        image: nft-platform:test
        command: ["make", "test-all"]
        env:
        - DB_HOST=postgres-service
        - REDIS_HOST=redis-service
```

## Test Data Management

### Test Data Factories

**Location**: `tests/testutil/common.go`

**Purpose**: Create consistent test data across all test types.

**Examples**:
```go
// Create test user
user := testutil.TestUser(1, "alice")

// Create test NFT
nft := testutil.TestNFT(1, 1, 1, "Test NFT")

// Create test auction
auction := testutil.TestAuction(1, 1, 1, "active")
```

### Database Test Setup

```go
// Setup test database
func setupTestDB(t *testing.T) *sql.DB {
    db := sql.Open("postgres", testDBConnStr)

    // Run migrations
    migrate.Exec(t, db, "up")

    // Truncate tables
    db.Exec("TRUNCATE TABLE users CASCADE;")
    db.Exec("TRUNCATE TABLE nfts CASCADE;")

    return db
}
```

### Mock Services

```go
// Mock repository for testing
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
    args := m.Called(user)
    return args.Error(0)
}
```

## Test Environment Configuration

### Environment Variables

```bash
# Test database
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_NAME=nft_platform_test
TEST_DB_USER=test_user
TEST_DB_PASSWORD=test_password

# Redis
TEST_REDIS_HOST=localhost
TEST_REDIS_PORT=6379
TEST_REDIS_DB=0

# Test configuration
APP_ENV=test
LOG_LEVEL=debug
JWT_SECRET=test_secret_key
```

### Test Configuration Files

```yaml
# test-config.yaml
database:
  host: localhost
  port: 5432
  name: nft_platform_test
  user: test_user
  password: test_password
  sslmode: disable

redis:
  host: localhost
  port: 6379
  db: 0

jwt:
  secret: test_secret_key
  access_duration: 1h
  refresh_duration: 24h
```

## Best Practices

### 1. Test Organization

- **Structure tests by feature/functionality**
- **Use descriptive test names**
- **Group related tests together**
- **Keep test files focused and manageable**

### 2. Test Data Management

- **Use factories for test data creation**
- **Ensure test data consistency**
- **Clean up test data after tests**
- **Use deterministic test data where possible**

### 3. Assertion Strategy

- **Use descriptive assertion messages**
- **Test both positive and negative cases**
- **Verify both state and behavior**
- **Use appropriate assertions for each test**

### 4. Test Independence

- **Tests should not depend on each other**
- **Tests should be runnable in any order**
- **Tests should clean up after themselves**
- **Use proper setup and teardown methods**

### 5. Performance Considerations

- **Keep tests fast and focused**
- **Use appropriate timeouts for integration tests**
- **Monitor test execution time**
- **Avoid unnecessary I/O operations**

### 6. Documentation

- **Document complex test scenarios**
- **Explain test purposes and expectations**
- **Maintain test case documentation**
- **Document any special test setup requirements**

## Troubleshooting

### Common Test Issues

#### Import Path Problems
```bash
# Check module structure
go list ./...

# Verify package imports
go mod verify
```

#### Database Connection Issues
```bash
# Check database connectivity
psql -h localhost -U test_user -d nft_platform_test

# Check table existence
psql -U test_user -d nft_platform_test -c "\dt"
```

#### Port Conflicts
```bash
# Check port usage
netstat -tulpn | grep :8080

# Kill processes using port
fuser -k :8080
```

#### Memory Issues
```bash
# Check memory usage
free -h

# Check Go process memory
ps aux | grep "[nft-platform]" | awk '{print $6}'
```

### Test Failure Analysis

1. **Check test logs**
   ```bash
   tail -f test-results/*.log
   ```

2. **Run single failing test**
   ```bash
   go test -v ./tests/unit/failing_test.go -run TestFailingFunction
   ```

3. **Check for race conditions**
   ```bash
   go test -race ./tests/unit/failing_test.go
   ```

4. **Enable verbose output**
   ```bash
   go test -v -run TestFailingFunction ./tests/unit/failing_test.go
   ```

## Continuous Improvement

### Code Coverage Goals

- **Maintain 80%+ coverage for critical paths**
- **Address low coverage areas**
- **Focus on business logic coverage**
- **Include edge cases in test coverage**

### Test Performance Goals

- **Keep test execution under 5 minutes for full suite**
- **Maintain <100ms average for unit tests**
- **Monitor and optimize slow tests**
- **Regular performance regression testing**

### Quality Assurance

- **Regular code reviews of test code**
- **Update tests when code changes**
- **Refactor tests for maintainability**
- **Add tests for new features immediately**

---

**Last Updated**: January 1, 2024
**Version**: 1.0.0

For questions or issues with testing, please refer to the development team or create an issue in the project repository.