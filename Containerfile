# Build stage
FROM docker.io/library/golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o fave .

# Create data directory
RUN mkdir -p /data && chmod 777 /data

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/fave .

# Copy data directory with correct permissions
COPY --from=builder --chown=65532:65532 /data /data

# Expose default port
EXPOSE 8080

# Set container-specific defaults
ENV FAVE_HOST=0.0.0.0
ENV FAVE_STORE_FILE=/data/bookmarks.json

# Run the server by default
ENTRYPOINT ["/app/fave"]
CMD ["serve"]
