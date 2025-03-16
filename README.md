# Go User API

A robust, production-ready User Management API service built with Golang following Clean Architecture principles. This service provides comprehensive user management capabilities including authentication, user CRUD operations, and token management.

## Features

- **User Management**
  - User registration and profile management
  - Role-based access control
  - User status management (active, inactive, blocked)
  
- **Authentication & Authorization**
  - Secure authentication using PASETO tokens (more secure alternative to JWT)
  - Access and refresh token functionality
  - Token revocation and logout capabilities
  
- **Robust Infrastructure**
  - MongoDB persistence layer
  - Redis cache for improved performance
  - Comprehensive logging with zerolog
  - API versioning ready
  
- **Security**
  - Secure password hashing with bcrypt
  - Protection against common web vulnerabilities
  - HTTPS support
  
- **Production-Ready**
  - Docker and Docker Compose support
  - Configurable via environment variables
  - Healthcheck endpoint
  - Graceful shutdown

## Technology Stack

- **Backend**: Go 1.24+
- **Web Framework**: Fiber v2
- **Database**: MongoDB 8
- **Cache**: Redis 7
- **Authentication**: PASETO (Platform-Agnostic Security Tokens)
- **Containerization**: Docker
- **Logging**: zerolog

## Architecture

The project follows Clean Architecture principles, organized into layers:

- **Entities**: Core business objects (User, Token)
- **Use Cases**: Application business rules
- **Interfaces**: Adapters for external systems (repositories, API handlers)
- **Infrastructure**: Implementation details (database connections, caching)

This architecture ensures:
- Separation of concerns
- Testability
- Independence from external frameworks
- Flexibility to change infrastructure details without affecting business logic

## Project Structure

```
.
├── api/                  # API layer (HTTP handlers, middleware, routing)
├── cmd/                  # Application entry points
├── config/               # Configuration handling
├── internal/             # Internal packages (not exported)
│   ├── domain/           # Domain layer (entities, use cases, repositories interfaces)
│   │   ├── entity/       # Domain entities
│   │   ├── repository/   # Repository interfaces and implementations
│   │   ├── service/      # Domain services
│   │   └── usecase/      # Business logic
│   ├── infrastructure/   # Infrastructure layer
│   │   ├── cache/        # Cache implementations (Redis)
│   │   └── db/           # Database implementations (MongoDB)
│   ├── logger/           # Logging functionality
│   └── utils/            # Utility functions
├── scripts/              # Scripts for setup, deployment, etc.
├── server/               # Server setup and lifecycle management
├── Dockerfile            # Docker configuration
├── docker-compose.yaml   # Docker Compose configuration
├── go.mod                # Go module definition
├── go.sum                # Go module checksum
└── Makefile              # Build and development commands
```

## Requirements

- Go 1.24 or higher
- MongoDB 8+
- Redis 7+
- Docker and Docker Compose (for containerized development)

## Getting Started

### Local Development Setup

1. **Clone the repository**

```bash
git clone https://github.com/your-username/go-user-api.git
cd go-user-api
```

2. **Install dependencies**

```bash
make deps
```

3. **Set up environment variables**

```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Start supporting services (MongoDB, Redis) with Docker**

```bash
make docker-up
```

5. **Run the application**

```bash
make run
# Or for hot reload during development
make dev
```

### Using Docker Compose

To run the entire application stack including MongoDB and Redis:

```bash
docker-compose up -d
```

## Configuration

The application is configured via environment variables. See `.env.example` for all available options.

Key configuration options:

```
# Application
APP_ENV=development  # development, staging, production
APP_DEBUG=true

# HTTP Server
HTTP_PORT=8080

# Database
DB_TYPE=mongodb
DB_HOST=localhost
DB_PORT=27017
DB_USERNAME=mongo
DB_PASSWORD=mongo
DB_DATABASE=user_service

# Cache
CACHE_TYPE=redis
CACHE_HOST=localhost
CACHE_PORT=6379

# Security
ACCESS_TOKEN_EXPIRATION_MINUTES=15
REFRESH_TOKEN_EXPIRATION_DAYS=7
```

## API Endpoints

### Authentication

- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout (requires authentication)
- `POST /api/v1/auth/logout-all` - Logout from all devices (requires authentication)

### User Management

- `POST /api/v1/users/register` - Register a new user
- `GET /api/v1/users/:id` - Get user by ID (requires authentication)
- `PUT /api/v1/users/:id` - Update user (requires authentication)
- `DELETE /api/v1/users/:id` - Delete user (requires authentication)
- `GET /api/v1/users` - List users with pagination (requires authentication)
- `PUT /api/v1/users/:id/password` - Change user password (requires authentication)
- `PUT /api/v1/users/:id/status` - Update user status (requires authentication)

### Healthcheck

- `GET /api/health` - Server health check

## Development

### Available Make Commands

```
make help              # Display available commands
make build             # Build the application
make run               # Run the application
make dev               # Run with hot reload
make test              # Run tests
make test-coverage     # Run tests with coverage
make lint              # Run linter
make docker-build      # Build Docker image
make docker-up         # Start Docker containers
make docker-down       # Stop Docker containers
make docker-logs       # Show Docker logs
```

### Generating Keys

The application uses PASETO tokens which require Ed25519 keys. To generate new keys:

```bash
go run scripts/generate-keys.go
```

Add the generated keys to your `.env` file:

```
PASETO_PRIVATE_KEY=your_generated_private_key
PASETO_PUBLIC_KEY=your_generated_public_key
```

## Deployment

### Docker Deployment

The included Dockerfile and docker-compose.yaml provide everything needed to deploy the application.

Build and push the Docker image:

```bash
make docker-build
make docker-push
```

### Server Deployment

1. Set up MongoDB and Redis
2. Configure environment variables
3. Build and deploy the application

```bash
make build
./go-user-api
```

## License

[MIT License](LICENSE)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Security

For security issues, please contact the maintainers directly instead of opening a public issue.