# Database MCP Server

> **⚠️ WARNING**: This software is designed for use with testing and development databases only. It makes no assurances that it won't break your database or cause data loss. **DO NOT use this software against production databases.** Always use dedicated test/development database instances when working with AI assistants.

A Model Context Protocol (MCP) server that provides secure database connectivity for MySQL and PostgreSQL databases. Enable AI assistants in Claude Code, Cursor, and other agentic editors to query, analyze, and interact with your databases safely.

## Features

- **Multi-Database Support**: Connect to MySQL and PostgreSQL databases
- **Secure Access Control**: Restrict access to specific databases and control connection limits
- **Rich Schema Inspection**: Get detailed table schemas, indexes, and metadata
- **Query Execution**: Run SELECT, INSERT, UPDATE, and DELETE operations
- **Query Planning**: View execution plans for performance optimization
- **Paginated Data Access**: Efficiently browse large datasets with pagination

## Installation

```bash
go install github.com/jhoffmann/go-database-mcp@latest
```

## Configuration

Configure the database connection using a connection string.

Create a `.env` file in your project directory:

#### PostgreSQL Examples

```bash
# Basic connection
DB_CONNECTION_STRING=postgresql://myuser:mypassword@localhost:5432/myapp

# With SSL requirements
DB_CONNECTION_STRING=postgresql://myuser:mypassword@localhost:5432/myapp?sslmode=require

# Optional: Allow access to additional databases
# DB_ALLOWED_NAMES=testdb,devdb,staging
```

#### MySQL Examples

```bash
# Basic connection
DB_CONNECTION_STRING=mysql://myuser:mypassword@localhost:3306/myapp

# Remote database with SSL
DB_CONNECTION_STRING=mysql://myuser:mypassword@db.example.com:3306/myapp?sslmode=require

# Optional: Allow access to additional databases
# DB_ALLOWED_NAMES=testdb,devdb,staging
```

### Environment Variables

| Variable               | Description                                              | Required | Default  | Notes                                         |
| ---------------------- | -------------------------------------------------------- | -------- | -------- | --------------------------------------------- |
| `DB_CONNECTION_STRING` | Full database connection URL (postgresql:// or mysql://) | Yes      | -        | Primary configuration method                  |
| `DB_SSL_MODE`          | SSL/TLS mode (`none`, `prefer`, `require`)               | No       | `prefer` | Can be set in connection string or separately |
| `DB_MAX_CONNS`         | Maximum open connections                                 | No       | 10       | Connection pool setting                       |
| `DB_MAX_IDLE_CONNS`    | Maximum idle connections                                 | No       | 5        | Connection pool setting                       |
| `DB_ALLOWED_NAMES`     | Comma-separated list of additional allowed databases     | No       | -        | Security setting                              |

## Integration with Agentic Editors

### OpenCode (https://opencode.ai)

Add the MCP server to your OpenCode configuration:

1. Create or edit `~/.config/opencode/opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "database-mcp": {
      "type": "local",
      "command": ["go-database-mcp"],
      "enabled": true,
      "environment": {
        "DB_CONNECTION_STRING": "postgresql://myuser:mypassword@localhost:5432/myapp"
      }
    }
  }
}
```

2. Restart OpenCode to load the MCP server

### Claude Code (https://www.anthropic.com/claude-code)

Add the MCP server to your Claude Code configuration:

1. Create or edit `~/.config/claude-code/mcp_servers.json`:

```json
{
  "mcpServers": {
    "database": {
      "command": "go-database-mcp",
      "args": [],
      "env": {
        "DB_CONNECTION_STRING": "postgresql://myuser:mypassword@localhost:5432/myapp"
      }
    }
  }
}
```

2. Restart Claude Code to load the MCP server

### Cursor IDE (https://cursor.com)

Configure the MCP server in your Cursor settings:

1. Open Cursor Settings → Extensions → MCP
2. Add a new MCP server configuration:

```json
{
  "name": "Database Server",
  "command": "go-database-mcp",
  "args": [],
  "env": {
    "DB_CONNECTION_STRING": "mysql://app_user:secure_password@localhost:3306/ecommerce"
  }
}
```

### Generic MCP Client Integration

For other MCP-compatible tools, configure the server with:

- **Command**: `go-database-mcp` (assuming it's in your PATH after `go install`)
- **Transport**: stdio
- **Environment**: Set `DB_CONNECTION_STRING` with your database connection URL

## Available MCP Tools

Once connected, the following tools become available to your AI assistant:

- `database_connection_info` - Get current database connection details
- `database_list_databases` - List all available databases
- `database_list_tables` - List tables in the current database
- `database_describe_table` - Get detailed schema for a specific table
- `database_get_table_data` - Retrieve paginated table data
- `database_query` - Execute SQL queries with optional parameters
- `database_explain_query` - Get query execution plans

## Usage Examples

### Ask your AI assistant:

**Schema Exploration:**

- "What tables are in my database?"
- "Describe the structure of the users table"
- "Show me the indexes on the orders table"

**Data Analysis:**

- "What are the top 10 customers by total order value?"
- "Show me user registration trends over the last 6 months"
- "Find all orders placed in the last week"

**Query Optimization:**

- "Explain the execution plan for this slow query"
- "How can I optimize this JOIN operation?"
- "What indexes should I add to improve performance?"

## Security Considerations

- **Database Access Control**: Use `DB_ALLOWED_NAMES` to restrict which databases can be accessed
- **User Permissions**: Create database users with minimal required permissions
- **Connection Limits**: Set appropriate `DB_MAX_CONNS` to prevent connection exhaustion
- **SSL/TLS**: Always use encrypted connections when available (`DB_SSL_MODE=require`). Available modes: `none` (no encryption, default), `prefer` (attempt SSL, fallback to unencrypted), `require` (mandatory SSL)
- **Environment Variables**: Store sensitive credentials in environment variables, not in code
