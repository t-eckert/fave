# Fave

Fave is a tiny bookmark manager written in Go. There are many like it, but this one is mine.

## Features

### Server
- RESTful HTTP API for bookmark management
- Persistent storage with automatic snapshots
- HTTP Basic Authentication support
- Graceful shutdown with signal handling
- Structured logging with `log/slog`
- CORS support for web clients
- Health check endpoint

### CLI Client
- Full CRUD operations (add, list, get, update, delete)
- Rich flag support for descriptions and tags
- Automatic tag deduplication
- Multi-source configuration (flags, env vars, config file)
- Retry logic with exponential backoff
- Connection pooling and timeouts

### Development
- Standard library only (no external dependencies except testing)
- Comprehensive test suite (>40 tests)
- Performance benchmarks (>20 benchmarks)
- High test coverage (>60% overall)

## Installation

```bash
go install github.com/t-eckert/fave@latest
```

Or build from source:

```bash
git clone https://github.com/t-eckert/fave.git
cd fave
go build
```

## Usage

### Start the Server

```bash
# With defaults
fave serve

# With custom configuration
fave serve --port 9090 --password secret123 --log-level debug

# With config file
fave serve --config config.json

# With environment variables
export FAVE_PORT=9090
export FAVE_AUTH_PASSWORD=secret123
fave serve
```

### CLI Client Commands

The Fave CLI provides commands to interact with a running server.

#### Add Bookmarks

```bash
# Basic add
fave add "My Website" "https://example.com"

# Add with description
fave add -d "Great article on Go" "Go Best Practices" "https://golang.org"
fave add --description "Useful tool" "Tool Name" "https://tool.com"

# Add with tags (can specify multiple times)
fave add -t golang -t programming "Learn Go" "https://golang.org"
fave add --tag web --tag tutorial "Web Tutorial" "https://example.com"

# Add with both description and tags
fave add -d "Comprehensive guide" -t golang -t guide "Go Guide" "https://go.dev"

# Tags are automatically deduplicated
fave add -t golang -t tutorial -t golang "Go Tutorial" "https://example.com"
# Results in tags: [golang, tutorial]

# Connect to remote server
fave add --host http://remote:8080 --password secret123 "Remote Bookmark" "https://example.com"
```

#### List Bookmarks

```bash
# List all bookmarks from default server (localhost:8080)
fave list

# List from remote server
fave list --host http://remote:8080 --password secret123
```

#### Get Bookmark by ID

```bash
# Get bookmark with ID 1
fave get 1

# Get from remote server
fave get 42 --host http://remote:8080 --password secret123
```

#### Update Bookmarks

```bash
# Update name and URL
fave update 1 "Updated Name" "https://newurl.com"

# Update with description
fave update -d "New description" 1 "Updated" "https://url.com"

# Update with tags
fave update -t updated -t v2 1 "Version 2" "https://v2.example.com"

# Update with description and tags
fave update -d "Latest version" -t v2 -t stable 1 "Stable Release" "https://example.com"

# Update on remote server
fave update --host http://remote:8080 42 "Updated" "https://example.com"
```

#### Delete Bookmarks

```bash
# Delete bookmark with ID 1
fave delete 1

# Delete from remote server
fave delete 42 --host http://remote:8080 --password secret123
```

#### Health Check

```bash
# Check if server is healthy
fave health

# Check remote server
fave health --host http://remote:8080
```

### Client Configuration

The CLI client can be configured using:

1. **CLI flags** (highest priority) - `--host`, `--password`, etc.
2. **Environment variables** - `FAVE_HOST`, `FAVE_PASSWORD`, etc.
3. **Config file** - `~/.fave/client-config.json`
4. **Defaults** - `http://localhost:8080` with no auth

#### Environment Variables

```bash
export FAVE_HOST=http://localhost:8080
export FAVE_PASSWORD=secret123
export FAVE_TIMEOUT=30s
export FAVE_RETRY_ATTEMPTS=3

# Now all commands use these settings
fave list
fave add "Example" "https://example.com"
```

#### Config File

Create `~/.fave/client-config.json`:

```json
{
  "host": "http://localhost:8080",
  "password": "secret123",
  "timeout": "30s",
  "retry_attempts": 3,
  "retry_delay": "1s"
}
```

Then run commands without flags:

```bash
fave list
fave add "Example" "https://example.com"
```

## Server Configuration

The Fave server can be configured in multiple ways, with the following precedence:

1. Command-line flags (highest)
2. Environment variables
3. Configuration file
4. Default values (lowest)

### Configuration Options

| Option | Flag | Environment Variable | Default | Description |
|--------|------|---------------------|---------|-------------|
| Port | `--port` | `FAVE_PORT` | `8080` | Server port |
| Host | `--host` | `FAVE_HOST` | `localhost` | Server host |
| Store File | `--store-file` | `FAVE_STORE_FILE` | `./data/bookmarks.json` | Path to bookmarks storage file |
| Password | `--password` | `FAVE_AUTH_PASSWORD` | `` (no auth) | Authentication password |
| Log Level | `--log-level` | `FAVE_LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| Log JSON | `--log-json` | `FAVE_LOG_JSON` | `false` | Output logs as JSON |
| Snapshot Interval | `--snapshot-interval` | `FAVE_SNAPSHOT_INTERVAL` | `1s` | Snapshot save interval (e.g., 1s, 5s, 1m) |

### Command-Line Flags

```bash
fave serve --port 8080 \
           --host localhost \
           --store-file ./data/bookmarks.json \
           --password secret123 \
           --log-level info \
           --log-json \
           --snapshot-interval 5s
```

### Environment Variables

```bash
export FAVE_PORT=8080
export FAVE_HOST=localhost
export FAVE_STORE_FILE=./data/bookmarks.json
export FAVE_AUTH_PASSWORD=secret123
export FAVE_LOG_LEVEL=info
export FAVE_LOG_JSON=true
export FAVE_SNAPSHOT_INTERVAL=5s

fave serve
```

### Configuration File

Create a `config.json`:

```json
{
  "port": "8080",
  "host": "localhost",
  "store_file": "./data/bookmarks.json",
  "auth_password": "secret123",
  "log_level": "info",
  "log_json": false,
  "snapshot_interval": "5s"
}
```

Then run:

```bash
fave serve --config config.json
```

See `config.example.json` for a complete example.

### Authentication

When `auth_password` is set, all API endpoints (except `/health`) require HTTP Basic Authentication:

```bash
# Using curl
curl -u user:secret123 http://localhost:8080/bookmarks

# Using JavaScript
fetch('http://localhost:8080/bookmarks', {
  headers: {
    'Authorization': 'Basic ' + btoa('user:secret123')
  }
})
```

Note: The username can be any value; only the password is validated.

### Graceful Shutdown

The server handles SIGINT (Ctrl+C) and SIGTERM gracefully:

1. Stops accepting new requests
2. Waits for active requests to complete (up to 30s)
3. Saves final snapshot to disk
4. Exits cleanly

```bash
# Send SIGINT
Ctrl+C

# Or send SIGTERM
kill -TERM <pid>
```

## API Reference

All endpoints return JSON. Errors follow this format:

```json
{
  "error": "Error message here"
}
```

### Endpoints

#### Health Check

```http
GET /health
```

Returns server health status. Does not require authentication.

**Response (200 OK):**
```json
{
  "status": "healthy"
}
```

#### List All Bookmarks

```http
GET /bookmarks
```

Returns all bookmarks.

**Response (200 OK):**
```json
{
  "1": {
    "url": "https://example.com",
    "name": "Example",
    "description": "An example bookmark",
    "tags": ["example", "test"]
  },
  "2": {
    "url": "https://golang.org",
    "name": "Go",
    "description": "The Go Programming Language",
    "tags": ["golang", "programming"]
  }
}
```

#### Get Bookmark by ID

```http
GET /bookmarks/{id}
```

Returns a specific bookmark.

**Response (200 OK):**
```json
{
  "url": "https://example.com",
  "name": "Example",
  "description": "An example bookmark",
  "tags": ["example", "test"]
}
```

**Response (404 Not Found):**
```json
{
  "error": "Bookmark not found"
}
```

#### Create Bookmark

```http
POST /bookmarks
Content-Type: application/json

{
  "url": "https://example.com",
  "name": "Example",
  "description": "An example bookmark",
  "tags": ["example", "test"]
}
```

Creates a new bookmark.

**Response (201 Created):**
```json
{
  "id": 1
}
```

**Response (400 Bad Request):**
```json
{
  "error": "Bookmark name is required"
}
```

#### Update Bookmark

```http
PUT /bookmarks/{id}
Content-Type: application/json

{
  "url": "https://example.com",
  "name": "Updated Example",
  "description": "An updated bookmark",
  "tags": ["example", "test", "updated"]
}
```

Updates an existing bookmark.

**Response (200 OK):**
```json
{
  "id": 1
}
```

**Response (404 Not Found):**
```json
{
  "error": "Bookmark not found"
}
```

#### Delete Bookmark

```http
DELETE /bookmarks/{id}
```

Deletes a bookmark.

**Response (200 OK):**
```json
{
  "id": 1
}
```

**Response (404 Not Found):**
```json
{
  "error": "Bookmark not found"
}
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific tests
go test -run TestGetBookmarks ./internal/server

# Run integration tests only
go test -run Integration ./internal/server
```

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./internal/server
go test -bench=. ./internal/store

# Run specific benchmark
go test -bench=BenchmarkGetBookmarks ./internal/server

# Run with memory profiling
go test -bench=. -benchmem ./internal/server
```

### Test Coverage

The project maintains high test coverage:

- Client package: ~42% (18 tests + 8 benchmarks)
- Server package: ~64% (20 tests + 7 benchmarks)
- Store package: ~89% (comprehensive tests + 9 benchmarks)

### Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── serve.go           # Server command
│   ├── add.go             # Add bookmark command (with -d/-t flags)
│   ├── list.go            # List bookmarks command
│   ├── get.go             # Get bookmark command
│   ├── update.go          # Update bookmark command (with -d/-t flags)
│   ├── delete.go          # Delete bookmark command
│   ├── health.go          # Health check command
│   └── utils/             # Shared utilities
│       ├── config.go      # Client config loader
│       ├── flags.go       # Custom flag types
│       └── format.go      # Output formatting
├── internal/
│   ├── bookmark.go        # Bookmark data structure
│   ├── client/            # HTTP client
│   │   ├── client.go      # Client implementation
│   │   ├── config.go      # Client configuration
│   │   ├── errors.go      # Error types
│   │   ├── client_test.go # Client tests (~18 tests)
│   │   └── client_bench_test.go # Client benchmarks (~8 benchmarks)
│   ├── server/            # HTTP server
│   │   ├── server.go      # Server implementation
│   │   ├── config.go      # Configuration system
│   │   ├── middleware.go  # HTTP middleware
│   │   ├── store_interface.go  # Store abstraction
│   │   ├── server_test.go      # Handler tests (~20 tests)
│   │   ├── integration_test.go # Integration tests (~5 tests)
│   │   ├── server_bench_test.go # Benchmarks (~7 benchmarks)
│   │   └── mock_store_test.go  # Mock for testing
│   └── store/             # Bookmark storage
│       ├── store.go       # Store implementation
│       ├── store_test.go  # Store tests
│       └── store_bench_test.go # Store benchmarks (~9 benchmarks)
├── main.go                # Entry point
├── config.example.json    # Example server configuration
└── README.md              # This file
```

## Architecture

### Server

The server uses Go's standard library `net/http` with custom middleware for:

- Request/response logging
- Panic recovery
- CORS support
- HTTP Basic Authentication

### Storage

Bookmarks are stored in memory and persisted to disk as JSON:

- In-memory storage with `sync.RWMutex` for thread safety
- Automatic snapshots at configurable intervals
- Atomic file writes (temp file + rename) to prevent corruption
- Loaded from disk on startup if file exists

### Testing

Comprehensive test suite with:

- **Client tests**: Unit tests for HTTP client (~18 tests), performance benchmarks (~8 benchmarks)
- **Server tests**: Unit tests for HTTP handlers (~20 tests), integration tests (~5 tests), benchmarks (~7 benchmarks)
- **Store tests**: Unit tests with comprehensive coverage, benchmarks for all operations (~9 benchmarks)
- Mock implementations for dependency injection
- Table-driven tests for multiple scenarios
- Modern `b.Loop()` syntax for all benchmarks

## License

MIT

## Contributing

This is a personal project, but contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Write tests for your changes
4. Ensure all tests pass
5. Submit a pull request

## Acknowledgments

Built with ❤️ using only Go's standard library (except for testing dependencies).
