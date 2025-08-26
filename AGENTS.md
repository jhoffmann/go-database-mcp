# Agent Development Guide

## Build Commands
- **Build**: `go build -o bin/server ./cmd/server`
- **Run**: `./bin/server` or `go run ./cmd/server`
- **Test**: `go test ./...` (single test: `go test -run TestName ./pkg/path`)
- **Format**: `go fmt ./...`
- **Vet**: `go vet ./...`

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
- No tests directory found - add tests alongside source files with `_test.go` suffix
- Configuration uses environment variables with envconfig tags and .env file support