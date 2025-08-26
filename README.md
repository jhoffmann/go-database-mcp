# Database MCP Server

A Go implementation of a Model Context Protocol (MCP) server that provides database connectivity for MySQL and PostgreSQL. Uses stdio transport for communication with MCP clients.

## Configuration

Create a `.env` file (see `.env.example`) or set environment variables:

```bash
# Database Configuration
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=myuser
DB_PASSWORD=mypassword


```

## Build and Run

```bash
# Build the server (static binary)
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/mcp-database-server ./cmd/server

# Run with environment variables
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=mydb
export DB_USER=myuser
export DB_PASSWORD=mypass
./bin/mcp-database-server

# Or run with .env file
cp .env.example .env
# Edit .env with your settings
./bin/mcp-database-server
```

## MCP Client Integration

This server uses stdio transport and communicates via JSON-RPC over stdin/stdout. It's designed to be used with MCP clients like Claude Desktop, IDEs with MCP support, or custom applications that can spawn and communicate with MCP servers.

## Project Structure

```
go-database-mcp/
├── cmd/server/           # Main application entry point
├── internal/
│   ├── config/          # Configuration structures and loading
│   ├── database/        # Database abstraction layer
│   ├── auth/           # Authentication handlers
│   └── handlers/       # MCP tool handlers
├── pkg/types/          # Shared types and schemas
└── README.md
```
