# Database MCP Server Implementation Plan

## Project Overview
A Go 1.25.0 database MCP server supporting MySQL and PostgreSQL with environment/file configuration. Uses stdio transport for Model Context Protocol communication.

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
│   └── handlers/
│       ├── query.go         # SQL query tools
│       ├── schema.go        # Schema inspection tools
│       └── admin.go         # Database admin tools
├── pkg/
│   └── types/               # Shared types and schemas
├── go.mod
├── go.sum
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
    SSLMode      string // postgres: disable/require/verify-ca/verify-full; mysql: none/required/preferred
}
```

### 3. Database Support Architecture
- **Unified Interface**: Common database interface for both MySQL and PostgreSQL
- **Driver Abstraction**: Separate implementations for each database type
- **Connection Pooling**: Built-in connection pool management
- **Query Builder**: Safe SQL query construction with parameter binding

### 4. Security Framework
- **Database SSL/TLS**: Support for encrypted database connections
- **Credential Security**: Environment-based credential loading, no credential logging
- **SQL Safety**: Parameter binding to prevent SQL injection
- **Stdio Transport**: Uses JSON-RPC over stdin/stdout for MCP communication

### 5. MCP Tools Implementation

**Core Tools:**
1. **query** - Execute SQL queries with result formatting
2. **describe_table** - Get table schema and metadata
3. **list_tables** - List all tables in database
4. **list_databases** - List available databases
5. **explain_query** - Get query execution plan
6. **get_table_data** - Fetch table data with pagination

### 6. Static Binary Distribution
- **Static Compilation**: CGO disabled for portable binaries
- **Cross-platform Support**: Linux, macOS, Windows builds
- **Minimal Dependencies**: Self-contained executable with no external runtime requirements
- **Simple Deployment**: Single binary distribution

## Implementation Roadmap (Shift-Left Testing Approach)

### Phase 1: Foundation with Testing Infrastructure (Week 1)
1. **Project Setup**
   - Initialize Go module with Go 1.25.0
   - Set up project structure
   - Add MCP Go SDK dependency
   - **Set up testing framework and coverage tools**

2. **Configuration System**
   - Implement config structures with comprehensive tests
   - Add environment variable loading with validation tests
   - Add .env file support with `godotenv` and test different scenarios
   - Add configuration validation with table-driven tests for all error cases
   - **Achieve 80%+ coverage on config package before proceeding**

3. **Basic MCP Server**
   - Create main server entry point with stdio transport
   - Implement basic MCP server initialization with testable design
   - Add logging infrastructure with structured testing
   - **Write unit tests for server initialization and configuration handling**

**Phase 1 Testing Deliverables:**
   - Complete unit test suite for configuration system (90%+ coverage)
   - Mock framework setup for external dependencies
   - CI/CD pipeline with automated testing
   - Test coverage reporting and quality gates

### Phase 2: Database Integration with Comprehensive Testing (Week 2)
1. **Database Abstraction Layer**
   - Define database interface with testability in mind
   - Implement connection management with dependency injection
   - Add connection pooling with configurable parameters
   - **Write interface tests and connection manager unit tests**

2. **MySQL Support**
   - Add `go-sql-driver/mysql` dependency
   - Implement MySQL-specific connection with SSL support
   - **Create comprehensive unit tests for MySQL implementation**
   - **Test DSN building, connection pooling, and error handling**

3. **PostgreSQL Support**
   - Add `lib/pq` dependency  
   - Implement PostgreSQL-specific connection with SSL support
   - **Create comprehensive unit tests for PostgreSQL implementation**
   - **Test connection strings, SSL modes, and error scenarios**

**Phase 2 Testing Deliverables:**
   - Unit tests for database interface implementations (80%+ coverage)
   - Mock database implementations for testing business logic
   - Connection manager tests with various configurations
   - Integration test framework setup (using test containers)

### Phase 3: Core MCP Tools with Test-First Development (Week 3)
1. **Basic Query Tools**
   - **Write tests first for expected query tool behavior**
   - Implement `query` tool with parameter binding to pass tests
   - Add result formatting (JSON, table format) with format validation tests
   - Implement `list_tables` tool with mock database testing

2. **Schema Tools**
   - **Write schema inspection tests before implementation**
   - Implement `describe_table` tool to satisfy test requirements
   - Add `list_databases` tool with comprehensive test coverage
   - Implement column metadata extraction with data validation tests

3. **Advanced Query Tools**
   - **Create test cases for complex query scenarios first**
   - Add `explain_query` tool with database-specific test mocking
   - Implement `get_table_data` with pagination testing
   - Add query validation and sanitization with security-focused tests

**Phase 3 Testing Deliverables:**
   - Test-driven development for all MCP tools
   - Integration tests with real database instances
   - Performance tests for pagination and large result sets
   - Security tests for SQL injection prevention

### Phase 4: Security Hardening with Security Testing (Week 4)
1. **SQL Safety**
   - **Write security tests before implementing safety measures**
   - Parameter binding implementation with injection attack tests
   - Query sanitization with malicious input testing
   - Input validation with boundary and edge case testing

2. **Credential Security**
   - **Create tests for credential exposure scenarios**
   - Secure environment variable handling with leak detection tests
   - No credential logging/exposure with audit trail testing
   - Connection string building with sanitization validation

3. **Query Safety**
   - **Write tests for query complexity and performance limits**
   - Query complexity limits (optional) with stress testing
   - Safe error messaging with information disclosure tests

**Phase 4 Testing Deliverables:**
   - Security test suite with penetration testing scenarios
   - Performance tests for connection pooling and query limits
   - Credential security audit with automated leak detection
   - Error handling tests ensuring no sensitive data exposure

### Phase 5: Static Binary Builds with Build Testing (Week 5)
1. **Build System**
   - Configure static compilation with CGO_ENABLED=0
   - Set up cross-platform build targets with validation tests
   - Optimize binary size with build flags and size regression tests
   - **Test builds on multiple platforms and architectures**

2. **Distribution**
   - Create release automation with checksum validation
   - Generate checksums for binaries with integrity tests
   - Package binaries for different platforms with deployment tests
   - **Automated testing of deployment packages**

**Phase 5 Testing Deliverables:**
   - Automated build verification on multiple platforms
   - Binary integrity and security scanning
   - Deployment and installation testing
   - Performance benchmarking of static builds

### Phase 6: Final Testing and Documentation (Week 6)
1. **Final Testing Suite**
   - End-to-end integration tests with real database scenarios
   - Load testing and performance benchmarking
   - Container-based test environment with CI/CD integration
   - **Final coverage analysis ensuring 80%+ across all components**

2. **Documentation with Examples**
   - MCP tool documentation with tested examples
   - Configuration guide with validated sample configurations
   - MCP client integration guide with working examples

3. **Quality Assurance**
   - Final security audit and penetration testing
   - Performance profiling and optimization validation
   - User acceptance testing with real-world scenarios

**Phase 6 Testing Deliverables:**
   - Complete test suite with 80%+ coverage across all packages
   - Performance benchmarks and load testing results
   - Security audit report with all issues resolved
   - Documentation with tested and verified examples

## Key Dependencies
- `github.com/modelcontextprotocol/go-sdk/mcp` - MCP SDK
- `github.com/go-sql-driver/mysql` - MySQL driver
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/joho/godotenv` - .env file support
- `github.com/kelseyhightower/envconfig` - Environment config parsing

## Build Configuration

### Environment Variables
```bash
# Database Configuration
DB_TYPE=postgres          # or mysql
DB_HOST=localhost         # database hostname
DB_PORT=5432             # database port
DB_NAME=myapp            # database name
DB_USER=myuser           # database username
DB_PASSWORD=mypassword   # database password
DB_SSL_MODE=prefer       # SSL mode for database connection
```

### Static Binary Build Commands
```bash
# Build for current platform
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/mcp-database-server ./cmd/server

# Build for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/mcp-database-server-linux ./cmd/server

# Build for macOS
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o bin/mcp-database-server-darwin ./cmd/server

# Build for Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/mcp-database-server.exe ./cmd/server
```

## Success Criteria with Quality Gates

### Functional Requirements
- ✅ Supports both MySQL and PostgreSQL with SSL connections
- ✅ Loads configuration from environment variables and .env files  
- ✅ Complete set of database query and schema tools
- ✅ SQL injection prevention through parameter binding
- ✅ Comprehensive error handling and logging
- ✅ Static binary compilation for easy distribution
- ✅ Cross-platform build support (Linux, macOS, Windows)
- ✅ Stdio-based MCP transport for client integration

### Quality Requirements (Shift-Left Testing)
- ✅ **80%+ test coverage across all packages** (corporate quality gate requirement)
- ✅ **Unit tests written immediately after each implementation** (not deferred)
- ✅ **Integration tests for all database operations**
- ✅ **Security tests for SQL injection and credential exposure**
- ✅ **Performance tests for connection pooling and query limits**
- ✅ **Cross-platform build validation and deployment testing**
- ✅ **Comprehensive error path testing and edge case coverage**
- ✅ **Testable architecture with dependency injection and mocking support**

### Documentation and Examples
- ✅ Clear documentation with tested and verified examples
- ✅ MCP client integration guide with working code samples
- ✅ Configuration guide with validated sample configurations
- ✅ Security best practices documentation