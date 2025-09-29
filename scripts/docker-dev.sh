#!/bin/bash
set -e

# NFT Platform Docker Development Helper Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker Desktop."
        exit 1
    fi
}

# Display help
show_help() {
    cat << EOF
NFT Platform Docker Development Helper

Usage: $0 [COMMAND]

Commands:
    up              Start all development services
    down            Stop all services
    restart         Restart all services
    logs            Show logs for all services
    logs [service]  Show logs for specific service
    shell [service] Open shell in service container
    clean           Remove all containers, networks, and volumes
    reset           Clean and rebuild everything
    status          Show status of all services
    db-shell        Open PostgreSQL shell
    redis-shell     Open Redis CLI
    kafka-topics    List Kafka topics
    help            Show this help message

Examples:
    $0 up                    # Start all services
    $0 logs nft-server      # Show logs for HTTP server
    $0 shell postgres       # Open bash in postgres container
    $0 clean                # Remove everything
    $0 reset                # Clean rebuild

Services:
    - postgres: PostgreSQL database
    - postgres_test: Test database
    - redis: Redis cache
    - kafka: Kafka message broker
    - kafka-ui: Kafka UI (http://localhost:8081)
    - hardhat-node: Local Ethereum node
    - nft-server: HTTP API server (when using --profile app)
    - nft-grpc-server: gRPC server (when using --profile app)

EOF
}

# Start services
start_services() {
    print_status "Starting NFT Platform development environment..."
    
    check_docker
    
    # Start core services (databases, cache, messaging)
    print_status "Starting core services (DB, Redis, Kafka, Blockchain)..."
    docker-compose up -d postgres postgres_test redis zookeeper kafka kafka-ui hardhat-node
    
    print_status "Waiting for services to be healthy..."
    sleep 10
    
    # Check health status
    print_status "Checking service health..."
    docker-compose ps
    
    print_success "Core services started successfully!"
    
    echo ""
    print_status "Available services:"
    echo "  - PostgreSQL: localhost:5432"
    echo "  - PostgreSQL Test: localhost:5433"
    echo "  - Redis: localhost:6379"
    echo "  - Kafka: localhost:9092"
    echo "  - Kafka UI: http://localhost:8081"
    echo "  - Hardhat Node: http://localhost:8545"
    echo ""
    print_warning "To start the application servers, run:"
    echo "  docker-compose --profile app up -d"
    echo "  or: docker-compose --profile hot-reload up -d (for hot reload)"
}

# Stop services
stop_services() {
    print_status "Stopping all services..."
    docker-compose --profile "*" down
    print_success "All services stopped."
}

# Restart services
restart_services() {
    print_status "Restarting all services..."
    stop_services
    start_services
}

# Show logs
show_logs() {
    if [ -z "$1" ]; then
        print_status "Showing logs for all services..."
        docker-compose logs -f
    else
        print_status "Showing logs for $1..."
        docker-compose logs -f "$1"
    fi
}

# Open shell in container
open_shell() {
    if [ -z "$1" ]; then
        print_error "Please specify a service name"
        exit 1
    fi
    
    print_status "Opening shell in $1..."
    docker-compose exec "$1" /bin/sh
}

# Clean everything
clean_everything() {
    print_warning "This will remove all containers, networks, and volumes!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_status "Cleaning everything..."
        docker-compose --profile "*" down -v --remove-orphans
        docker system prune -f
        print_success "Everything cleaned."
    else
        print_status "Cancelled."
    fi
}

# Reset (clean and rebuild)
reset_everything() {
    print_status "Resetting everything..."
    clean_everything
    docker-compose build --no-cache
    start_services
    print_success "Reset complete."
}

# Show status
show_status() {
    print_status "Service status:"
    docker-compose ps
}

# Database shell
db_shell() {
    print_status "Opening PostgreSQL shell..."
    docker-compose exec postgres psql -U postgres -d nft_platform
}

# Redis shell
redis_shell() {
    print_status "Opening Redis CLI..."
    docker-compose exec redis redis-cli
}

# List Kafka topics
kafka_topics() {
    print_status "Kafka topics:"
    docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list
}

# Main script logic
case "$1" in
    up|start)
        start_services
        ;;
    down|stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    logs)
        show_logs "$2"
        ;;
    shell)
        open_shell "$2"
        ;;
    clean)
        clean_everything
        ;;
    reset)
        reset_everything
        ;;
    status)
        show_status
        ;;
    db-shell)
        db_shell
        ;;
    redis-shell)
        redis_shell
        ;;
    kafka-topics)
        kafka_topics
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac