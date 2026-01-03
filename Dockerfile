# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
# -ldflags="-s -w" strips debug info to reduce binary size
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w" \
    -o bot \
    ./cmd/bot

# Runtime stage - use minimal alpine image
FROM alpine:latest

# Add metadata labels
LABEL maintainer="Telegram Bot Go"
LABEL description="Telegram Bot with AI integration - Go implementation"

# Install runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    sqlite-libs \
    tzdata && \
    # Create non-root user for security
    addgroup -g 1000 botuser && \
    adduser -D -u 1000 -G botuser botuser

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bot .

# Create data directory for SQLite with proper permissions
RUN mkdir -p /app/data && \
    chown -R botuser:botuser /app

# Switch to non-root user
USER botuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Set environment variables
ENV DB_PATH=/app/data/bot.db \
    PORT=8080

# Run the bot
CMD ["./bot"]
