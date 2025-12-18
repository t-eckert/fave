# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Fave is a bookmark manager with a client-server architecture written in Go. The project uses only the Go standard library.

**Key Architecture:**
- **Server**: RESTful HTTP API with in-memory storage and JSON persistence
- **Client**: HTTP client library with retry logic and authentication
- **CLI**: Command-line interface that uses the client library to interact with the server

## Building and Testing

```bash
# Build the binary
go build

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests verbosely
go test -v ./...

# Run specific package tests
go test ./internal/server
go test ./internal/client
go test ./internal/store

# Run specific test
go test -run TestGetBookmarks ./internal/server

# Run benchmarks
go test -bench=. ./internal/server
go test -bench=. ./internal/client
go test -bench=. ./internal/store

# Run benchmarks with memory stats
go test -bench=. -benchmem ./internal/server

# Build container image
podman build -t fave:latest .
docker build -t fave:latest .

# Run container
podman run -d -p 8080:8080 -v fave-data:/data fave:latest
```

**Container Notes:**
- Uses distroless base image (`gcr.io/distroless/static-debian12:nonroot`)
- Runs as UID 65532 (nonroot user)
- No shell or debugging tools available (use `--entrypoint` override for debugging)
- Project has no external dependencies (uses only Go standard library), so `go.sum` may not exist

**CI/CD:**
- GitHub Actions workflow runs on push to main and pull requests
- Tests on multiple Go versions (1.23, 1.24, 1.25) and OS (Linux, macOS, Windows)
- Runs linting (go vet, gofmt, staticcheck)
- Runs benchmarks
- Builds container image on main branch
- Coverage reports uploaded to Codecov (optional)

## Architecture

### Three-Layer Design

1. **Storage Layer** (`internal/store/`):
   - In-memory storage with `sync.RWMutex` for thread safety
   - Periodic snapshots to JSON file on disk
   - Atomic file writes (temp file + rename) to prevent corruption
   - Auto-incremented integer IDs

2. **Server Layer** (`internal/server/`):
   - HTTP handlers for RESTful API
   - Middleware chain: Recovery → Logging → CORS → Auth
   - Uses `StoreInterface` for dependency injection and testing
   - Background goroutine for periodic snapshots
   - Graceful shutdown handling

3. **Client Layer** (`internal/client/`):
   - HTTP client with connection pooling and timeouts
   - Retry logic with exponential backoff
   - HTTP Basic Authentication support
   - Structured error types (`ClientError`, sentinel errors)

### Configuration System

Both server and client use **multi-source configuration** with precedence:
1. CLI flags (highest priority)
2. Environment variables
3. Config file (JSON)
   - Client: `~/.config/fave/client.json`
   - Server: Path specified via `--config` flag
4. Default values (lowest priority)

**Implementation pattern:**
- Single `FlagSet` with all flags defined
- Use `fs.Visit()` to track explicitly set flags
- Only apply CLI flag values if explicitly set (avoids overriding config file with defaults)

### Middleware Pattern

Server uses function composition for middleware:

```go
type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler
```

Middlewares are applied in reverse order so the first middleware wraps outermost.

### Authentication

- **HTTP Basic Authentication**: Username ignored, only password validated
- **Public Read Mode**: When `public: true`, GET requests bypass auth, but POST/PUT/DELETE still require auth
- **Health endpoint**: Always bypasses auth

## Code Conventions

### Benchmark Tests

All benchmarks use modern `b.Loop()` syntax (not `for i := 0; i < b.N; i++`):

```go
func BenchmarkExample(b *testing.B) {
    for b.Loop() {
        // benchmark code
    }
}
```

Use `b.RunParallel()` for concurrent benchmarks.

### Error Handling

- Client uses structured errors (`internal/client/errors.go`)
- Sentinel errors: `ErrNotFound`, `ErrUnauthorized`, `ErrBadRequest`, etc.
- Server returns JSON errors: `{"error": "message"}`

### Testing Strategy

- **Unit tests**: Mock `StoreInterface` for server testing
- **Integration tests**: Full server + real store
- **Table-driven tests**: Multiple test cases in slices
- **httptest.Server**: Mock HTTP servers for client testing

### Flag Deduplication

Tags support multiple `-t` flags with automatic deduplication:

```go
// cmd/utils/flags.go
type StringSlice []string
func DeduplicateStrings(input []string) []string
```

## Project Structure

```
.
├── cmd/                    # CLI commands
│   ├── serve.go           # Server command
│   ├── add.go             # Add bookmark (-d, -t flags)
│   ├── update.go          # Update bookmark (-d, -t flags)
│   ├── list.go            # List bookmarks
│   ├── get.go             # Get bookmark by ID
│   ├── delete.go          # Delete bookmark
│   ├── health.go          # Health check
│   └── utils/             # Shared CLI utilities
│       ├── config.go      # Client config loader
│       ├── flags.go       # Custom flag types (StringSlice)
│       └── format.go      # Output formatting
├── internal/
│   ├── bookmark.go        # Core Bookmark type
│   ├── client/            # HTTP client library
│   │   ├── client.go
│   │   ├── config.go      # Multi-source config
│   │   ├── errors.go      # Structured errors
│   │   ├── client_test.go
│   │   └── client_bench_test.go
│   ├── server/            # HTTP server
│   │   ├── server.go      # Server and handlers
│   │   ├── config.go      # Multi-source config
│   │   ├── middleware.go  # HTTP middleware
│   │   ├── store_interface.go  # Abstraction for testing
│   │   ├── server_test.go      # Unit tests
│   │   ├── integration_test.go # Integration tests
│   │   ├── server_bench_test.go
│   │   └── mock_store_test.go  # Mock for testing
│   └── store/             # In-memory storage
│       ├── store.go
│       ├── store_test.go
│       └── store_bench_test.go
├── main.go                # CLI entry point
└── config.example.json    # Example server config
```

## API Reference

**JSON API:**
```
GET    /health              → {"status": "healthy"}
GET    /bookmarks           → map[int]Bookmark
GET    /bookmarks/{id}      → Bookmark
POST   /bookmarks           → {"id": int}
PUT    /bookmarks/{id}      → {"id": int}
DELETE /bookmarks/{id}      → {"id": int}
```

## Key Implementation Details

### Server Lifecycle

1. `New()` creates server, starts background snapshot goroutine
2. `Start()` begins HTTP listening (blocking)
3. `Close()` gracefully shuts down:
   - Stops snapshot loop
   - Saves final snapshot
   - Shuts down HTTP server with 30s timeout

### Store Snapshot Strategy

- Background goroutine with configurable interval (default: 1s)
- Atomic writes: write to temp file, then rename
- Final snapshot on graceful shutdown

### Client Retry Logic

- Exponential backoff with configurable attempts and delay
- Retries on network errors and 5xx responses
- Does not retry 4xx client errors

### Configuration Precedence Bug

Previously used two `FlagSet` instances which caused `--config` flag to not be recognized on second parse. Fixed by using single `FlagSet` with `fs.Visit()` to track explicit flags.
