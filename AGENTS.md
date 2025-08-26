# Agent Development Guide

## Build Commands

- **Build**: `go build -o build/server ./cmd/server`
- **Run**: `./build/server` or `go run ./cmd/server`
- **Test**: `go test ./...` (single test: `go test -run TestName ./pkg/path`)
- **Coverage**: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- **Format**: `go fmt ./...`
- **Vet**: `go vet ./...`

## Code Quality Requirements

- **Test Coverage**: We must achieve **over 80% test coverage** to satisfy corporate quality gate requirements
- **Testable Design**: Code must be designed with testability in mind:
  - Use dependency injection for database connections and external dependencies
  - Create interfaces for external dependencies (file system, network, etc.)
  - Design functions to be pure when possible (deterministic input/output)
  - Separate business logic from I/O operations
  - Use constructor functions that accept dependencies as parameters
  - Avoid global state and singleton patterns
  - Make private functions testable through public interfaces

## Code Style

- **Imports**: Standard library first, then external, then internal (`github.com/jhoffmann/go-database-mcp/internal/...`)
- **Naming**: PascalCase for exports, camelCase for private, snake_case for JSON tags
- **Types**: Define structs with JSON tags using envconfig pattern for config
- **Error handling**: Always check errors with `if err != nil` and wrap with `fmt.Errorf("context: %w", err)`
- **Logging**: Use `log.Printf()` with structured messages
- **Context**: Pass `context.Context` as first parameter in functions that might block

## Project Structure

- `cmd/server/`: Main application entry point
- `internal/`: Private application code (config, database, handlers, auth)
- Tests: Add tests alongside source files with `_test.go` suffix
- Configuration uses environment variables with envconfig tags and .env file support

## Testing Strategy

- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test database operations with real connections (using test containers)
- **Interface Mocking**: Create mocks for database interface to test business logic without real connections
- **Table-driven Tests**: Use table-driven tests for comprehensive input/output validation
- **Error Path Testing**: Ensure all error conditions are tested and covered
- **Configuration Testing**: Test all configuration validation and loading scenarios

