# NFT Platform Deployment Guide

**Version**: 1.0.0
**Last Updated**: January 1, 2024

This guide provides comprehensive instructions for deploying the NFT Platform in various environments, from development to production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Development Deployment](#development-deployment)
4. [Production Deployment](#production-deployment)
5. [Docker Deployment](#docker-deployment)
6. [Kubernetes Deployment](#kubernetes-deployment)
7. [Monitoring and Logging](#monitoring-and-logging)
8. [Security Configuration](#security-configuration)
9. [Backup and Recovery](#backup-and-recovery)
10. [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

**Minimum Requirements:**
- CPU: 2 cores
- RAM: 4GB
- Storage: 20GB SSD
- Network: 100 Mbps

**Recommended Requirements:**
- CPU: 4+ cores
- RAM: 8GB+
- Storage: 50GB+ SSD
- Network: 1 Gbps

### Software Requirements

- **Go**: 1.21+ (for building from source)
- **Docker**: 20.10+ (for containerized deployment)
- **Docker Compose**: 2.0+ (for local development)
- **Kubernetes**: 1.24+ (for production clusters)
- **Node.js**: 18+ (for frontend if applicable)
- **PostgreSQL**: 14+ (database)
- **Redis**: 6.0+ (caching)
- **Apache Kafka**: 3.0+ (messaging)

### External Services

- **Ethereum Node**: Sepolia testnet or mainnet endpoint
- **IPFS**: For decentralized storage (optional)
- **CDN**: For static asset delivery (recommended)
- **Load Balancer**: For high availability (production)

## Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/nft-platform/nft-platform.git
cd nft-platform
```

### 2. Environment Configuration

Create environment configuration file:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Application
APP_ENV=production
APP_PORT=8080
APP_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=nft_platform
DB_PASSWORD=your_secure_password
DB_NAME=nft_platform_db
DB_SSL_MODE=require

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password
REDIS_DB=0

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC_EVENTS=nft-events
KAFKA_TOPIC_NOTIFICATIONS=nft-notifications

# JWT
JWT_SECRET=your_very_secure_jwt_secret_key_here
JWT_ACCESS_TOKEN_DURATION=3600
JWT_REFRESH_TOKEN_DURATION=86400

# Ethereum
ETHEREUM_NETWORK=sepolia
ETHEREUM_RPC_URL=https://sepolia.infura.io/v3/YOUR_PROJECT_ID
ETHEREUM_PRIVATE_KEY=your_private_key_here
CONTRACT_ADDRESS=0x...

# File Storage
UPLOAD_DIR=./uploads
MAX_FILE_SIZE=10485760
ALLOWED_FILE_TYPES=jpg,jpeg,png,gif

# Security
CORS_ALLOWED_ORIGINS=https://yourdomain.com
RATE_LIMIT_REQUESTS_PER_MINUTE=100

# Monitoring
LOG_LEVEL=info
LOG_FORMAT=json
METRICS_ENABLED=true
HEALTH_CHECK_ENABLED=true
```

### 3. Database Setup

#### PostgreSQL Setup

```bash
# Install PostgreSQL (Ubuntu/Debian)
sudo apt update
sudo apt install postgresql postgresql-contrib

# Create database and user
sudo -u postgres psql
CREATE DATABASE nft_platform_db;
CREATE USER nft_platform WITH PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE nft_platform_db TO nft_platform;
\q
```

#### Redis Setup

```bash
# Install Redis (Ubuntu/Debian)
sudo apt install redis-server

# Configure Redis
sudo nano /etc/redis/redis.conf
# Set password and other security settings
sudo systemctl restart redis-server
```

#### Kafka Setup

```bash
# Download and extract Kafka
wget https://downloads.apache.org/kafka/3.5.1/kafka_2.13-3.5.1.tgz
tar -xzf kafka_2.13-3.5.1.tgz
sudo mv kafka_2.13-3.5.1 /opt/kafka

# Start Kafka services
/opt/kafka/bin/zookeeper-server-start.sh -daemon /opt/kafka/config/zookeeper.properties
/opt/kafka/bin/kafka-server-start.sh -daemon /opt/kafka/config/server.properties

# Create topics
/opt/kafka/bin/kafka-topics.sh --create --topic nft-events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
/opt/kafka/bin/kafka-topics.sh --create --topic nft-notifications --bootstrap-server localhost:9092 --partitions 2 --replication-factor 1
```

## Development Deployment

### Using Docker Compose (Recommended)

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### From Source

```bash
# Install dependencies
go mod download

# Build application
make build

# Run database migrations
make migrate-up

# Start server
make run-server
```

### Environment File for Development

```env
APP_ENV=development
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=nft_platform
DB_PASSWORD=dev_password
DB_NAME=nft_platform_dev
REDIS_HOST=localhost
REDIS_PORT=6379
KAFKA_BROKERS=localhost:9092
JWT_SECRET=dev_secret_key
ETHEREUM_NETWORK=sepolia
LOG_LEVEL=debug
```

## Production Deployment

### 1. Build Application

```bash
# Build for production
make build

# Or cross-compile
GOOS=linux GOARCH=amd64 go build -o nft-platform cmd/server/main.go
```

### 2. Database Migration

```bash
# Run migrations in production
./nft-platform migrate up

# Or using migrate tool
migrate -path migrations -database "postgres://user:password@host:port/dbname?sslmode=require" up
```

### 3. Systemd Service

Create systemd service file:

```bash
sudo nano /etc/systemd/system/nft-platform.service
```

```ini
[Unit]
Description=NFT Platform Server
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=nft-platform
Group=nft-platform
WorkingDirectory=/opt/nft-platform
ExecStart=/opt/nft-platform/nft-platform server
Restart=always
RestartSec=5
Environment=APP_ENV=production

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/nft-platform/uploads

[Install]
WantedBy=multi-user.target
```

Enable and start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable nft-platform
sudo systemctl start nft-platform
sudo systemctl status nft-platform
```

### 4. Nginx Reverse Proxy

Create Nginx configuration:

```bash
sudo nano /etc/nginx/sites-available/nft-platform
```

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /path/to/ssl/cert.pem;
    ssl_certificate_key /path/to/ssl/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;

    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static files
    location /uploads/ {
        alias /opt/nft-platform/uploads/;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

Enable site:

```bash
sudo ln -s /etc/nginx/sites-available/nft-platform /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Docker Deployment

### 1. Dockerfile

```dockerfile
# Multi-stage build
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o nft-platform cmd/server/main.go

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/nft-platform .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env.production .env

EXPOSE 8080

CMD ["./nft-platform", "server"]
```

### 2. Docker Compose for Production

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DB_HOST=postgres
      - REDIS_HOST=redis
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - postgres
      - redis
      - kafka
    restart: unless-stopped
    volumes:
      - ./uploads:/app/uploads
    networks:
      - nft-network

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: nft_platform_db
      POSTGRES_USER: nft_platform
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    restart: unless-stopped
    networks:
      - nft-network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped
    networks:
      - nft-network

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    restart: unless-stopped
    networks:
      - nft-network

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    restart: unless-stopped
    networks:
      - nft-network

volumes:
  postgres_data:
  redis_data:

networks:
  nft-network:
    driver: bridge
```

### 3. Deploy with Docker Compose

```bash
# Create production compose file
cp docker-compose.yml docker-compose.prod.yml

# Edit production settings
nano docker-compose.prod.yml

# Deploy
docker-compose -f docker-compose.prod.yml up -d

# Scale if needed
docker-compose -f docker-compose.prod.yml up -d --scale app=3
```

## Kubernetes Deployment

### 1. Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: nft-platform
```

### 2. ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: nft-platform-config
  namespace: nft-platform
data:
  APP_ENV: "production"
  APP_PORT: "8080"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "nft_platform_db"
  REDIS_HOST: "redis-service"
  REDIS_PORT: "6379"
  KAFKA_BROKERS: "kafka-service:9092"
```

### 3. Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: nft-platform-secrets
  namespace: nft-platform
type: Opaque
data:
  DB_PASSWORD: <base64-encoded-password>
  REDIS_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-jwt-secret>
  ETHEREUM_PRIVATE_KEY: <base64-encoded-private-key>
```

### 4. Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nft-platform
  namespace: nft-platform
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nft-platform
  template:
    metadata:
      labels:
        app: nft-platform
    spec:
      containers:
      - name: nft-platform
        image: nft-platform:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: nft-platform-config
        - secretRef:
            name: nft-platform-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 5. Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nft-platform-service
  namespace: nft-platform
spec:
  selector:
    app: nft-platform
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

### 6. Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nft-platform-ingress
  namespace: nft-platform
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/proxy-body-size: "10m"
spec:
  tls:
  - hosts:
    - yourdomain.com
    secretName: nft-platform-tls
  rules:
  - host: yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nft-platform-service
            port:
              number: 80
```

### 7. Deploy to Kubernetes

```bash
# Apply all configurations
kubectl apply -f k8s/

# Check deployment
kubectl get pods -n nft-platform
kubectl get services -n nft-platform
kubectl get ingress -n nft-platform

# View logs
kubectl logs -f deployment/nft-platform -n nft-platform
```

## Monitoring and Logging

### 1. Prometheus Metrics

Add to main application:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

// Add metrics endpoint
router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### 2. Grafana Dashboard

Create monitoring dashboard:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboard
  namespace: monitoring
data:
  nft-platform.json: |
    {
      "dashboard": {
        "title": "NFT Platform Metrics",
        "panels": [
          {
            "title": "Request Rate",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(http_requests_total[5m])"
              }
            ]
          },
          {
            "title": "Response Time",
            "type": "graph",
            "targets": [
              {
                "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
              }
            ]
          }
        ]
      }
    }
```

### 3. Structured Logging

Configure structured logging:

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

// Use structured logging
logger.Info("User authenticated",
    zap.String("user_id", "123"),
    zap.String("ip", "192.168.1.1"),
    zap.Duration("response_time", 150*time.Millisecond),
)
```

### 4. Log Aggregation with ELK Stack

```yaml
# docker-compose.logging.yml
version: '3.8'

services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.5.0
    environment:
      - discovery.type=single-node
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"

  logstash:
    image: docker.elastic.co/logstash/logstash:8.5.0
    volumes:
      - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf
    ports:
      - "5044:5044"

  kibana:
    image: docker.elastic.co/kibana/kibana:8.5.0
    ports:
      - "5601:5601"
    environment:
      ELASTICSEARCH_HOSTS: http://elasticsearch:9200
```

## Security Configuration

### 1. SSL/TLS Setup

```bash
# Generate self-signed certificate (for testing)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/nft-platform.key \
  -out /etc/ssl/certs/nft-platform.crt

# Or use Let's Encrypt
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com
```

### 2. Firewall Configuration

```bash
# UFW firewall setup
sudo ufw enable
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny 5432/tcp  # PostgreSQL (internal only)
sudo ufw deny 6379/tcp  # Redis (internal only)
sudo ufw deny 9092/tcp  # Kafka (internal only)
```

### 3. Security Headers

Add security headers to Nginx:

```nginx
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "no-referrer-when-downgrade" always;
add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

### 4. Database Security

```sql
-- Create read-only user for reporting
CREATE USER nft_platform_readonly WITH PASSWORD 'secure_readonly_password';
GRANT CONNECT ON DATABASE nft_platform_db TO nft_platform_readonly;
GRANT USAGE ON SCHEMA public TO nft_platform_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO nft_platform_readonly;

-- Enable row-level security
ALTER TABLE auctions ENABLE ROW LEVEL SECURITY;
CREATE POLICY auction_policy ON auctions FOR ALL USING (seller_id = current_user_id());
```

## Backup and Recovery

### 1. Database Backup

```bash
#!/bin/bash
# backup.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
DB_NAME="nft_platform_db"

# Create backup directory
mkdir -p $BACKUP_DIR

# Database backup
pg_dump -h localhost -U nft_platform -d $DB_NAME > $BACKUP_DIR/db_backup_$DATE.sql

# Compress backup
gzip $BACKUP_DIR/db_backup_$DATE.sql

# Remove old backups (keep last 7 days)
find $BACKUP_DIR -name "db_backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: $BACKUP_DIR/db_backup_$DATE.sql.gz"
```

### 2. Automated Backup with Cron

```bash
# Edit crontab
crontab -e

# Add backup schedule (daily at 2 AM)
0 2 * * * /opt/nft-platform/scripts/backup.sh
```

### 3. File Backup

```bash
#!/bin/bash
# backup-files.sh

DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups"
UPLOAD_DIR="/opt/nft-platform/uploads"

# Create backup
tar -czf $BACKUP_DIR/uploads_backup_$DATE.tar.gz $UPLOAD_DIR

# Sync to remote storage (AWS S3 example)
aws s3 sync $BACKUP_DIR/ s3://your-backup-bucket/nft-platform/

echo "Files backup completed: $BACKUP_DIR/uploads_backup_$DATE.tar.gz"
```

### 4. Recovery Procedures

```bash
# Restore database
gunzip -c /backups/db_backup_20240101_020000.sql.gz | psql -h localhost -U nft_platform -d nft_platform_db

# Restore files
tar -xzf /backups/uploads_backup_20240101_020000.tar.gz -C /
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

```bash
# Check PostgreSQL status
sudo systemctl status postgresql

# Check connection
psql -h localhost -U nft_platform -d nft_platform_db

# View PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

#### 2. Redis Connection Failed

```bash
# Check Redis status
sudo systemctl status redis-server

# Test connection
redis-cli -a your_password ping

# View Redis logs
sudo tail -f /var/log/redis/redis-server.log
```

#### 3. Kafka Connection Issues

```bash
# Check Kafka topics
/opt/kafka/bin/kafka-topics.sh --list --bootstrap-server localhost:9092

# Test Kafka connectivity
/opt/kafka/bin/kafka-console-producer.sh --broker-list localhost:9092 --topic test
```

#### 4. High Memory Usage

```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head

# Monitor application memory
go tool pprof http://localhost:8080/debug/pprof/heap
```

#### 5. Slow Database Queries

```sql
-- Enable slow query logging
ALTER SYSTEM SET log_min_duration_statement = 1000;
SELECT pg_reload_conf();

-- View slow queries
SELECT query, mean_time, calls
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

### Health Checks

```bash
# Application health
curl -f http://localhost:8080/health || exit 1

# Database health
pg_isready -h localhost -p 5432 -U nft_platform || exit 1

# Redis health
redis-cli -a your_password ping || exit 1
```

### Performance Monitoring

```bash
# CPU and memory usage
htop

# Network connections
netstat -tulpn | grep :8080

# Disk usage
df -h

# Application logs
tail -f /var/log/nft-platform/app.log
```

## Support

For deployment support:

- **Documentation**: [docs.nftplatform.com](https://docs.nftplatform.com)
- **GitHub Issues**: [github.com/nft-platform/issues](https://github.com/nft-platform/issues)
- **Community**: [discord.gg/nftplatform](https://discord.gg/nftplatform)
- **Email**: support@nftplatform.com

---

**Note**: This deployment guide covers the most common deployment scenarios. For specific requirements or custom deployments, please refer to the detailed documentation or contact the support team.