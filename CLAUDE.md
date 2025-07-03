# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build System

This repository uses `just` as the build system. Key commands:

- `just build <target>` - Build a specific target (e.g., `just build kuzu-memory`)
- `just build-static <target>` - Build with embedded C libraries (semi-static)
- `just build-release <target>` - Build optimized release binary
- `just build-linux-static <target>` - Cross-compile fully static binary for Linux (requires Docker)
- `just run <target> [args]` - Run a target, building first if needed
- `just dev <target> [args]` - Run target in development mode with `go run`

Available targets:
- `date-time` - Simple MCP server that returns current date/time
- `kuzu-memory` - MCP server with Kuzu graph database integration
- `duckdb` - DuckDB-based MCP server with dynamic configuration

## Architecture

This is a Go project implementing multiple MCP (Model Context Protocol) servers:

### MCP Server Structure
Each server is implemented as a separate command in `cmd/`:
- Built using `github.com/mark3labs/mcp-go` framework
- Servers communicate via stdio protocol
- Each server registers tools and/or prompts with the MCP framework

### Key Components

**kuzu-memory server** (`cmd/kuzu-memory/`):
- Most complex server with graph database integration
- Uses Kuzu graph database for memory storage
- Implements both tools (`getSchema`, `query`) and prompts (`generateKuzuCypher`)
- Configuration via command line args or `KUZU_DB_PATH` environment variable
- Read-only mode controlled by `KUZU_READ_ONLY=true` environment variable
- Internal structure in `internal/kuzu-memory/`:
  - `config/` - Configuration parsing
  - `db/` - Kuzu database wrapper
  - `tools/` - MCP tool handlers
  - `prompts/` - MCP prompt handlers

**duckdb server** (`cmd/duckdb/`):
- Dynamic DuckDB MCP server with session-based configuration
- Tools: `configure` (setup database/directory), `query` (SQL execution), `list_files` (file discovery)
- Prompt: `duckdb-initial-prompt` (comprehensive DuckDB guidance)
- Supports local DuckDB files, CSV/Parquet/JSON direct querying, `:memory:` databases
- No static CLI arguments - all configuration via MCP protocol
- Internal structure in `internal/duckdb/`:
  - `state/` - Session state management
  - `database/` - DuckDB connection wrapper
  - `tools/` - MCP tool handlers (configure, query, list_files)
  - `prompt/` - MCP prompt templates

**date-time server** (`cmd/date-time/`):
- Simple server that provides current timestamp
- Single tool: `current_date_time`
- Returns ISO 8601 formatted datetime

## Development

- Go 1.24.3 required
- Dependencies managed via `go.mod`
- No test framework currently configured
- No linting configuration present
- Standard Go project structure with `internal/` for private packages
- uses git flow: git flow feature start <feature>; git flow feature finish <feature>

## Best practices

- use `go vet` to check code quality
- use `golangci-lint` to lint code
- use `go doc` if you are unsure how to properly use som package
- after changes or adding new features, update CLAUDE.md
- also, after changes or adding new features, update nortes in Obsididan

## Environment Variables

- `KUZU_DB_PATH` - Path to Kuzu database (fallback if not provided as CLI arg)
- `KUZU_READ_ONLY` - Set to "true" to enable read-only mode for kuzu-memory server
