# Payment Gateway Service

A production-ready payment gateway service built with Go that demonstrates asynchronous processing, idempotent operations, and reliable message handling.

## Features

- **Payment Creation**: Create payments with amount, currency (ETB/USD), and unique reference
- **Asynchronous Processing**: Payments are processed via RabbitMQ message queue
- **Idempotent Operations**: Payments are never processed more than once, even with message redelivery or concurrent workers
- **Database Transactions**: Row-level locking ensures consistency during concurrent processing
- **RESTful API**: Clean HTTP API for payment operations
- **Docker Support**: Complete containerized deployment with Docker Compose

## Architecture

```
Client -> API Server -> PostgreSQL (Store Payment)
                   -> RabbitMQ (Queue Processing)

Worker <- RabbitMQ <- API Server

Worker -> PostgreSQL (Process Payment with Idempotency)
```

## Tech Stack

- **Backend**: Go with Echo framework
- **Database**: PostgreSQL with row-level locking
- **Message Queue**: RabbitMQ for async processing
- **Containerization**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)

### Environment Configuration

The application uses environment variables for configuration. Create a `.env` file in the project root:

```bash
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=payment_gateway
DB_SSLMODE=disable

# RabbitMQ Configuration
RABBITMQ_HOST=rabbitmq
RABBITMQ_PORT=5672
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
RABBITMQ_VHOST=/

# API Server Configuration
API_PORT=8080

# Development/Production Environment
ENV=development
LOG_LEVEL=info
```

**Note**: Docker Compose automatically loads environment variables from the `.env` file in the project root.

### Run with Docker Compose

1. **Start all services**:
   ```bash
   docker compose up --build
   ```

   This will start:
   - PostgreSQL database (port 5432)
   - RabbitMQ (ports 5672, 15672 for management UI)
   - API server (port 8080)
   - Worker service
   - Adminer database manager (port 8081)

2. **Check service health**:
   ```bash
   curl http://localhost:8080/health
   ```

3. **Access management interfaces**:
   - **API Documentation**: http://localhost:8080/swagger/index.html
   - **Database Manager**: http://localhost:8081
   - **Message Queue Monitor**: http://localhost:15672

## API Usage

### Create Payment

```bash
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 100.50,
    "currency": "USD",
    "reference": "order-12345"
  }'
```

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "PENDING"
}
```

### Get Payment Status

```bash
curl http://localhost:8080/api/v1/payments/{payment-id}
```

**Response**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 100.50,
  "currency": "USD",
  "reference": "order-12345",
  "status": "SUCCESS",
  "created_at": "2024-01-03T10:00:00Z"
}
```

## Testing Idempotency

The system ensures payments are processed exactly once:

1. **Create a payment** - status will be `PENDING`
2. **Wait for async processing** - status changes to `SUCCESS` or `FAILED`
3. **Multiple workers or message redelivery** - payment status remains unchanged
4. **Concurrent processing attempts** - database locking prevents race conditions

### Test Concurrent Processing

```bash
# Terminal 1: Start monitoring
watch -n 1 "curl -s http://localhost:8080/api/v1/payments/{payment-id} | jq .status"

# Terminal 2: Create payment
curl -X POST http://localhost:8080/api/v1/payments \
  -H "Content-Type: application/json" \
  -d '{"amount": 50.00, "currency": "ETB", "reference": "test-concurrent"}'

# Terminal 3: Simulate multiple workers by running extra worker instances
docker compose up --scale worker=3
```

### Visual Monitoring

For easier testing and debugging, use the web interfaces:

1. **API Testing**: http://localhost:8080/swagger/index.html
2. **Database Viewer**: http://localhost:8081 (Adminer)
3. **Queue Monitor**: http://localhost:15672 (RabbitMQ Management)

Watch messages flow from API → RabbitMQ → Worker → Database in real-time!

## Development

### Local Development Setup

1. **Start dependencies**:
   ```bash
   docker compose up postgres rabbitmq -d
   ```

2. **Run API server**:
   ```bash
   go run cmd/api/main.go
   ```

3. **Run worker** (in another terminal):
   ```bash
   go run cmd/worker/main.go
   ```

### Project Structure

```
├── cmd/
│   ├── api/           # API server entry point
│   └── worker/        # Worker service entry point
├── internal/
│   ├── config/        # Configuration management
│   ├── database/      # DB connection and repository
│   ├── models/        # Data models and DTOs
│   └── rabbitmq/      # Message queue client and messaging
├── pkg/
│   ├── handlers/      # HTTP request handlers
│   └── worker/        # Background processing logic
├── docker/            # Docker files and database schema
└── docker-compose.yml # Multi-service orchestration
```

## Database Schema

```sql
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL CHECK (currency IN ('ETB', 'USD')),
    reference VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'SUCCESS', 'FAILED')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Monitoring & Database Access

### Web Interfaces

#### Adminer (Database Management)
Access the Adminer database management interface at: http://localhost:8081

**Features:**
- Web-based database browser and editor
- Execute SQL queries directly
- View table structures and relationships
- Export data in various formats

**Login Credentials:**
- **System**: PostgreSQL
- **Server**: postgres (Docker service name)
- **Username**: postgres
- **Password**: postgres
- **Database**: payment_gateway

**Quick Start:**
1. Open http://localhost:8081
2. Select "PostgreSQL" from system dropdown
3. Enter server: `postgres`, username: `postgres`, password: `postgres`
4. Click "Login"

#### RabbitMQ Management UI
Access the RabbitMQ management interface at: http://localhost:15672
- **Username**: guest
- **Password**: guest

**Key Monitoring Areas:**
- **Queues**: Monitor `payment_processing` queue
- **Connections**: See active API and worker connections
- **Overview**: Message rates and system health

### Logs

View logs for all services:
```bash
docker compose logs -f
```

View specific service logs:
```bash
docker compose logs -f api
docker compose logs -f worker
docker compose logs -f postgres
docker compose logs -f rabbitmq
```

## Production Considerations

- **Payment Processing**: In production, integrate with real payment processors (Stripe, PayPal, etc.)
- **Error Handling**: Implement exponential backoff for failed payment processing
- **Monitoring**: Add metrics collection and health checks
- **Security**: Implement authentication, rate limiting, and input sanitization
- **Scalability**: Add load balancing and database connection pooling
- **Dead Letter Queues**: Handle permanently failed payment processing

## API Documentation

Interactive API documentation is available via Swagger UI:

**URL**: http://localhost:8080/swagger/index.html

The documentation includes:
- Complete endpoint specifications
- Request/response examples
- Error response details
- Interactive testing interface

## Troubleshooting

### Common Issues and Solutions

#### RabbitMQ Deprecation Warnings
**Issue**: Seeing warnings about `management_metrics_collection` being deprecated

**Solution**: The project uses RabbitMQ 3.13 which eliminates these warnings. If you encounter them:
```bash
# Update to the latest version
docker compose pull rabbitmq
docker compose up --build rabbitmq
```

#### Port Conflicts
**Issue**: Ports 8080, 5432, 5672, or 15672 already in use

**Solution**: Check what's using the ports and stop those services:
```bash
# Find what's using port 8080
lsof -i :8080

# Or change ports in docker-compose.yml
ports:
  - "8081:8080"  # Change host port
```

#### Database Connection Issues
**Issue**: API can't connect to PostgreSQL

**Solution**: Ensure PostgreSQL is healthy:
```bash
docker compose ps
docker compose logs postgres

# Restart if needed
docker compose restart postgres
```

#### Build Failures
**Issue**: Go compilation errors

**Solution**: Clear Docker cache and rebuild:
```bash
docker system prune -f
docker compose build --no-cache
```

#### Payment Processing Delays
**Issue**: Payments stay in PENDING status

**Solution**: Check worker service:
```bash
docker compose logs worker
docker compose ps

# Ensure worker is running
docker compose restart worker
```

### Getting Help

1. **Check logs**: `docker compose logs -f`
2. **Verify services**: `docker compose ps`
3. **Test API**: `curl http://localhost:8080/health`
4. **Check RabbitMQ**: http://localhost:15672

## Technology Choices

### Why These Technologies?

| Component | Choice | Reasoning |
|-----------|--------|-----------|
| **Go** | Backend Language | High performance, strong concurrency, excellent for microservices |
| **Echo** | Web Framework | Lightweight, fast, good middleware support |
| **PostgreSQL** | Database | ACID transactions, row-level locking, JSON support |
| **RabbitMQ** | Message Queue | Reliable delivery, management UI, proven in production |
| **Docker** | Containerization | Consistent environments, easy scaling, isolation |
| **sqlc** | Database Code Gen | Type-safe SQL, compile-time verification, zero runtime errors |

### Performance Expectations

- **API Response Time**: <50ms for simple requests
- **Payment Processing**: <5 seconds for async processing
- **Concurrent Requests**: Handles 1000+ concurrent connections
- **Database Queries**: <10ms average response time

### Security Considerations

- **Input Validation**: Comprehensive validation using struct tags
- **SQL Injection Protection**: Parameterized queries via sqlc
- **Container Security**: Non-root users, minimal base images
- **Network Security**: Internal service communication

## Contributing

### Development Workflow

1. **Fork and clone** the repository
2. **Create feature branch**: `git checkout -b feature/your-feature`
3. **Make changes** following the existing patterns
4. **Test thoroughly**: `./test_api.sh`
5. **Update documentation** if needed
6. **Submit pull request**

### Code Standards

- **Go formatting**: `go fmt`
- **Import organization**: `goimports`
- **Linting**: Follow Go best practices
- **Documentation**: Swagger annotations for API changes
- **Testing**: Unit tests for new functionality

### Commit Messages

```
feat: add payment refund functionality
fix: resolve RabbitMQ connection timeout
docs: update API documentation
refactor: improve error handling in handlers
```

## License

This project is for demonstration purposes. See individual component licenses for production use.

## Acknowledgments

- **Go Community**: For excellent documentation and libraries
- **Docker**: For containerization technology
- **PostgreSQL**: For robust database functionality
- **RabbitMQ**: For reliable message queuing

---

**Built for Senior Golang Developer Assessment** 🚀
