# Go-Next-ERP

Enterprise Resource Planning (ERP) system built with Go, featuring a comprehensive 4A permission system (Authentication, Authorization, Accounting, Audit).

## Features

### Core Modules
- **4A Permission System**
  - Authentication (JWT-based auth)
  - Authorization (RBAC, ABAC, ReBAC)
  - Accounting (Session management)
  - Audit (Comprehensive audit logging)

- **Organization Management**
  - Multi-level organization structure
  - Position management
  - Employee management
  - Department hierarchy

- **Workflow Engine**
  - Dynamic workflow definitions
  - State machine-based execution
  - Flexible node types (approval, notification, etc.)

### Technical Features
- Multi-tenant architecture
- PostgreSQL database with connection pooling
- Redis caching (optional)
- RESTful API with Swagger documentation
- Comprehensive logging with structured output
- Database migration system
- Graceful shutdown
- Health checks and metrics

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 16 (or use Docker)
- Redis 7 (optional, or use Docker)

### 1. Start Infrastructure Services

```bash
# Start PostgreSQL, Redis, MinIO, and Milvus
docker-compose up -d

# Check service status
make docker-ps
```

### 2. Configure Environment

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your configuration
# Default values work with docker-compose setup
```

### 3. Run Database Migrations

```bash
# Build migration tool
make migrate-build

# Run migrations
make migrate-up

# Check migration status
make migrate-status
```

### 4. Build and Run

```bash
# Build the application
make build

# Run the server
make run

# Or run in development mode with hot reload
make dev
```

The server will start on http://localhost:8080

### 5. Access Documentation

- **API Documentation**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics
- **Version Info**: http://localhost:8080/version

## Development

### Project Structure

```
.
├── cmd/
│   ├── api/           # HTTP API server
│   └── migrate/       # Database migration tool
├── internal/
│   ├── auth/          # 4A permission system
│   ├── middleware/    # HTTP middlewares
│   └── organization/  # Organization module
├── pkg/
│   ├── cache/         # Redis cache abstraction
│   ├── database/      # PostgreSQL abstraction
│   ├── logger/        # Structured logging
│   ├── migrate/       # Migration management
│   └── workflow/      # Workflow engine
├── docs/              # Swagger documentation
└── test/              # Integration tests
```

### Available Commands

```bash
# Development
make install        # Install dependencies
make build          # Build application
make run            # Run application
make dev            # Run with hot reload

# Testing
make test           # Run all tests
make test-coverage  # Generate coverage report
make test-unit      # Run unit tests only
make bench          # Run benchmarks

# Database
make migrate-up     # Run migrations
make migrate-status # Check migration status
make db-create      # Create database
make db-drop        # Drop database

# Code Quality
make fmt            # Format code
make vet            # Run go vet
make lint           # Run golangci-lint

# Documentation
make swag-init      # Generate Swagger docs
make swag-fmt       # Format Swagger comments

# Docker
make docker-up      # Start all services
make docker-down    # Stop all services
make docker-logs    # View all logs
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage
open coverage.html

# Run specific package tests
go test -v ./internal/auth/...
go test -v ./pkg/workflow/...

# Run benchmarks
make bench
```

### Generating API Documentation

```bash
# Generate Swagger documentation
make swag-init

# The docs will be available at:
# http://localhost:8080/swagger/index.html
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/change-password` - Change password

### User Management
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user details
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Role Management
- `GET /api/v1/roles` - List roles
- `POST /api/v1/roles` - Create role
- `PUT /api/v1/roles/:id` - Update role
- `DELETE /api/v1/roles/:id` - Delete role
- `POST /api/v1/roles/assign` - Assign roles to user
- `POST /api/v1/roles/remove` - Remove roles from user

### Permission Management
- `GET /api/v1/permissions` - List permissions
- `POST /api/v1/permissions` - Create permission
- `POST /api/v1/permissions/assign` - Assign permissions to role
- `POST /api/v1/permissions/remove` - Remove permissions from role

### Organization Management
- `GET /api/v1/organizations` - List organizations
- `POST /api/v1/organizations` - Create organization
- `GET /api/v1/organizations/tree` - Get organization tree
- `PUT /api/v1/organizations/:id` - Update organization
- `DELETE /api/v1/organizations/:id` - Delete organization

See the full API documentation at `/swagger/index.html` when the server is running.

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `15000` |
| `DB_NAME` | Database name | `erp` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres123` |
| `REDIS_ENABLED` | Enable Redis cache | `false` |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `15001` |
| `REDIS_PASSWORD` | Redis password | `redis123` |
| `JWT_SECRET_KEY` | JWT secret key | `your-secret-key-change-me-in-production` |

### Database Configuration

The application uses PostgreSQL with the following configuration:
- Max connections: 50
- Min connections: 10
- Connection timeout: 30s
- Statement timeout: 60s

### Cache Configuration

Redis cache is optional but recommended for production:
- Supports standalone and cluster modes
- Configurable TTL per cache key
- Automatic connection retry

## Deployment

### Building for Production

```bash
# Build for current platform
make build

# Build for Linux
make build-linux

# Build for macOS
make build-darwin

# Build for Windows
make build-windows

# Build for all platforms
make build-all
```

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Start all services
docker-compose up -d

# View logs
make docker-logs

# Stop services
docker-compose down
```

### Environment Setup

1. Set `GIN_MODE=release` in production
2. Use strong `JWT_SECRET_KEY`
3. Enable Redis caching for better performance
4. Configure proper database connection limits
5. Set up log rotation
6. Use HTTPS/TLS for API endpoints

## Monitoring

### Health Checks

```bash
# Check application health
curl http://localhost:8080/health

# Response includes:
# - Overall status
# - Database connectivity
# - Redis connectivity (if enabled)
```

### Metrics

```bash
# Get system metrics
curl http://localhost:8080/metrics

# Includes:
# - Memory usage
# - Goroutine count
# - GC statistics
# - CPU info
```

### Logging

The application uses structured logging with the following levels:
- `DEBUG` - Detailed debugging information
- `INFO` - General informational messages
- `WARN` - Warning messages
- `ERROR` - Error messages

Logs include:
- Request/response logging with duration
- Database query logging
- Authentication events
- Audit trail for all operations

## Testing

### Test Coverage

The project maintains comprehensive test coverage across all layers:

| Layer | Coverage | Status | Test Cases |
|-------|----------|--------|-----------|
| **Handler Layer** | ~70-78% | ✅ Excellent | 226+ tests |
| **Repository Layer** | High | ✅ Complete | 135+ tests |
| **Service Layer** | Partial | ⚠️ In Progress | - |
| **Overall** | ~70% | ✅ Good | 360+ tests |

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific module tests
go test -v ./internal/organization/handler/...
go test -v ./internal/approval/handler/...
go test -v ./internal/form/handler/...
go test -v ./internal/notification/handler/...
go test -v ./internal/auth/handler/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Architecture

All tests follow consistent patterns:

**AAA Pattern** (Arrange-Act-Assert):
```go
t.Run("Create successfully", func(t *testing.T) {
    // Arrange - Setup mock and data
    mockService.On("Create", mock.Anything, req).Return(expected, nil)

    // Act - Execute the operation
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("POST", "/api/v1/resource", body)
    router.ServeHTTP(w, req)

    // Assert - Verify results
    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
})
```

**Mock Services**:
- All Handler tests use interface-based mocking
- Powered by `testify/mock` for precise control
- Consistent mock patterns across all modules

**Test Reports**:
- `FINAL_TEST_REPORT.md` - Comprehensive test summary
- `AUTH_REFACTORING_COMPLETE.md` - Auth module architecture details
- `PROJECT_STATUS_SUMMARY.md` - Overall project status

### Key Testing Achievements

✅ **Handler Layer**: 8 handlers fully tested with 226+ test cases
✅ **Repository Layer**: 7 repositories with 135+ test cases
✅ **Architecture Consistency**: All modules follow the same testable design
✅ **Mock Infrastructure**: Complete mock service implementations
✅ **Test Coverage**: Maintained above 70% overall

### Auth Module Refactoring

The Auth module underwent a major architectural refactoring to improve testability:

**Before**:
```go
type AuthHandler struct {
    authService *authentication.Service  // Concrete type
}
```

**After**:
```go
type AuthHandler struct {
    authService authentication.AuthenticationService  // Interface
}
```

This change enabled:
- ✅ Unit testing with mock services
- ✅ Consistent architecture across all modules
- ✅ Better adherence to SOLID principles
- ✅ Improved maintainability

See `AUTH_REFACTORING_COMPLETE.md` for detailed information.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Follow Go best practices and idioms
- Write unit tests for new features
- Update documentation for API changes
- Run `make fmt` and `make lint` before committing
- Keep test coverage above 80%

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Email: support@go-next-erp.com

## Acknowledgments

Built with:
- [Gin](https://github.com/gin-gonic/gin) - Web framework
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver
- [go-redis](https://github.com/redis/go-redis) - Redis client
- [Swag](https://github.com/swaggo/swag) - Swagger documentation
- [Zap](https://github.com/uber-go/zap) - Structured logging
