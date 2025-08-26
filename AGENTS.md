# Agent Development Guide

## Build Commands

- **Build**: `go build -o build/server ./cmd/server`
- **Run**: `./build/server` or `go run ./cmd/server`
- **Test**: `go test ./...` (single test: `go test -run TestName ./pkg/path`)
- **Coverage**: `go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out`
- **Format**: `go fmt ./...`
- **Vet**: `go vet ./...`

## Code Quality Requirements

- **Test Coverage**: We must achieve **over 80% test coverage** to satisfy quality gate requirements
- **Shift-Left Testing**: Testing must be integrated from the earliest phases of development:
  - Write tests immediately after implementing each function or method
  - Test-driven development (TDD) approach when possible
  - Never defer testing to later phases - test as you build
  - Each pull request must include corresponding tests
  - Continuous testing throughout development, not just at the end
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

- **Shift-Left Approach**: Testing is integrated into every development phase, not deferred to the end
- **Unit Tests**: Test individual functions and methods in isolation immediately after implementation
- **Integration Tests**: Test database operations with real connections (using test containers)
- **Interface Mocking**: Create mocks for database interface to test business logic without real connections
- **Table-driven Tests**: Use table-driven tests for comprehensive input/output validation
- **Error Path Testing**: Ensure all error conditions are tested and covered
- **Configuration Testing**: Test all configuration validation and loading scenarios
- **Continuous Validation**: Run tests frequently during development to catch issues early

## Development Workflow

1. **Design Phase**: Consider testability when designing interfaces and functions
2. **Implementation**: Write tests immediately after implementing each component
3. **Validation**: Run coverage analysis to ensure 80%+ coverage before moving to next feature
4. **Integration**: Test integration points as they are developed, not at the end
5. **Quality Gates**: No code is considered "done" without corresponding tests
