# Database MCP Server Implementation Plan

## Project Overview
A Go 1.25.0 database MCP server supporting MySQL and PostgreSQL with environment/file configuration, TLS, and authentication.

## Architecture Design

### 1. Project Structure
```
go-database-mcp/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go        # Configuration structures
│   │   └── loader.go        # Environment and .env file loading
│   ├── database/
│   │   ├── connection.go    # Database connection management
│   │   ├── mysql.go         # MySQL-specific implementation
│   │   ├── postgres.go      # PostgreSQL-specific implementation
│   │   └── interface.go     # Database interface
│   ├── auth/
│   │   ├── auth.go          # Authentication handlers
│   │   └── tls.go           # TLS configuration
│   └── handlers/
│       ├── query.go         # SQL query tools
│       ├── schema.go        # Schema inspection tools
│       └── admin.go         # Database admin tools
├── pkg/
│   └── types/               # Shared types and schemas
├── deployments/
│   └── docker/
│       ├── Dockerfile       # Multi-stage Docker build
│       └── docker-compose.yml # Development environment
├── go.mod
├── go.sum
├── .dockerignore
└── README.md
```

### 2. Configuration System
- **Environment Variables**: Support standard DB_* prefixed variables
- **.env File Support**: Load from `.env` file in working directory
- **Hierarchical Config**: Environment variables override .env file values
- **Validation**: Comprehensive config validation with helpful error messages

**Key Configuration Fields:**
```go
type Config struct {
    Database DatabaseConfig `json:"database"`
    Server   ServerConfig   `json:"server"`
    Auth     AuthConfig     `json:"auth"`
    TLS      TLSConfig      `json:"tls"`
}

type DatabaseConfig struct {
    Type         string // "mysql" or "postgres"
    Host         string
    Port         int
    Database     string
    Username     string
    Password     string
    MaxConns     int
    MaxIdleConns int
    SSLMode      string // postgres: disable/require/verify-ca/verify-full
}
```

### 3. Database Support Architecture
- **Unified Interface**: Common database interface for both MySQL and PostgreSQL
- **Driver Abstraction**: Separate implementations for each database type
- **Connection Pooling**: Built-in connection pool management
- **Query Builder**: Safe SQL query construction with parameter binding

### 4. TLS and Authentication Framework
- **TLS Configuration**: Support for client certificates and custom CA
- **Database Authentication**: Username/password, certificate-based auth
- **MCP Authentication**: Optional MCP-level authentication
- **Security Best Practices**: No credential logging, secure connection handling

### 5. MCP Tools Implementation

**Core Tools:**
1. **query** - Execute SQL queries with result formatting
2. **describe_table** - Get table schema and metadata
3. **list_tables** - List all tables in database
4. **list_databases** - List available databases
5. **explain_query** - Get query execution plan
6. **get_table_data** - Fetch table data with pagination

### 6. Docker Containerization

**Container Features:**
- **Multi-stage Build**: Optimized production image with minimal attack surface
- **Security**: Non-root user execution, minimal base image (distroless/alpine)
- **Configuration**: Environment-based configuration for containers
- **Health Checks**: Built-in container health monitoring
- **Volume Support**: Optional volume mounts for TLS certificates and config files

**Docker Architecture:**
```dockerfile
# Multi-stage build with Go 1.25.0
FROM golang:1.25.0-alpine AS builder
# Build optimized binary with CGO disabled

FROM gcr.io/distroless/static-debian12:latest
# Minimal production image with non-root user
```

**Deployment Options:**
- **Standalone Container**: Single container deployment
- **Docker Compose**: Multi-service development environment
- **Docker Swarm**: Container orchestration support

## Implementation Roadmap

### Phase 1: Foundation (Week 1)
1. **Project Setup**
   - Initialize Go module with Go 1.25.0
   - Set up project structure
   - Add MCP Go SDK dependency

2. **Configuration System**
   - Implement config structures
   - Add environment variable loading
   - Add .env file support with `godotenv`
   - Add configuration validation

3. **Basic MCP Server**
   - Create main server entry point
   - Implement basic MCP server with health check
   - Add logging infrastructure

### Phase 2: Database Integration (Week 2)
1. **Database Abstraction Layer**
   - Define database interface
   - Implement connection management
   - Add connection pooling

2. **MySQL Support**
   - Add `go-sql-driver/mysql` dependency
   - Implement MySQL-specific connection
   - Add SSL/TLS support for MySQL

3. **PostgreSQL Support**
   - Add `lib/pq` dependency
   - Implement PostgreSQL-specific connection
   - Add SSL/TLS support for PostgreSQL

### Phase 3: Core MCP Tools (Week 3)
1. **Basic Query Tools**
   - Implement `query` tool with parameter binding
   - Add result formatting (JSON, table format)
   - Implement `list_tables` tool

2. **Schema Tools**
   - Implement `describe_table` tool
   - Add `list_databases` tool
   - Implement column metadata extraction

3. **Advanced Query Tools**
   - Add `explain_query` tool
   - Implement `get_table_data` with pagination
   - Add query validation and sanitization

### Phase 4: Security and Authentication (Week 4)
1. **TLS Implementation**
   - Client certificate support
   - Custom CA configuration
   - TLS version enforcement

2. **Authentication Framework**
   - Database credential management
   - Secure credential storage
   - Connection string building

3. **Security Hardening**
   - SQL injection prevention
   - Query complexity limits
   - Rate limiting for queries

### Phase 5: Docker Containerization (Week 5)
1. **Container Development**
   - Create multi-stage Dockerfile
   - Implement health check endpoints
   - Add .dockerignore for optimized builds

2. **Container Security**
   - Non-root user execution
   - Minimal base image selection
   - Security scanning integration

3. **Orchestration Support**
   - Docker Compose for development
   - Docker Swarm support for production scaling
   - Environment-based configuration

### Phase 6: Testing and Documentation (Week 6)
1. **Comprehensive Testing**
   - Unit tests for all components
   - Integration tests with real databases
   - Container-based test environment

2. **Documentation**
   - API documentation
   - Configuration guide
   - Docker deployment guide
   - Security best practices guide

3. **Examples and Demos**
   - Sample configurations
   - Docker Compose setup
   - Docker Swarm deployment examples
   - Usage examples

## Key Dependencies
- `github.com/modelcontextprotocol/go-sdk/mcp` - MCP SDK
- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/joho/godotenv` - .env file support
- `github.com/kelseyhightower/envconfig` - Environment config parsing

## Docker Configuration

### Environment Variables for Container
```bash
# Database Configuration
DB_TYPE=postgres          # or mysql
DB_HOST=database         # container hostname
DB_PORT=5432            # database port
DB_NAME=myapp           # database name
DB_USER=myuser          # database username
DB_PASSWORD=mypassword  # database password

# Server Configuration
MCP_HOST=0.0.0.0        # bind to all interfaces in container
MCP_PORT=8080           # internal container port

# TLS Configuration (optional)
TLS_CERT_PATH=/certs/server.crt
TLS_KEY_PATH=/certs/server.key
TLS_CA_PATH=/certs/ca.crt
```

### Docker Build Commands
```bash
# Build production image
docker build -f deployments/docker/Dockerfile -t go-database-mcp:latest .

# Run standalone container
docker run -d \
  --name mcp-server \
  -p 8080:8080 \
  -e DB_TYPE=postgres \
  -e DB_HOST=host.docker.internal \
  -e DB_PORT=5432 \
  -e DB_NAME=myapp \
  -e DB_USER=myuser \
  -e DB_PASSWORD=mypassword \
  go-database-mcp:latest

# Run with Docker Compose
docker-compose -f deployments/docker/docker-compose.yml up -d
```

### Container Health Check
- **Endpoint**: `GET /health`
- **Response**: JSON with database connectivity status
- **Timeout**: 30 seconds
- **Interval**: 10 seconds
- **Retries**: 3

## Success Criteria
- ✅ Supports both MySQL and PostgreSQL
- ✅ Loads configuration from environment variables and .env files
- ✅ Secure TLS and authentication support
- ✅ Complete set of database query and schema tools
- ✅ Comprehensive error handling and logging
- ✅ Production-ready security measures
- ✅ Docker containerization with security best practices
- ✅ Multi-stage optimized Docker builds
- ✅ Container orchestration support (Docker Compose, Docker Swarm)
- ✅ Full test coverage including container tests
- ✅ Clear documentation and deployment examples