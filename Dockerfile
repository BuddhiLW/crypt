# Build stage
FROM golang:1.23.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata libjpeg-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o crypt ./cmd/crypt

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata libjpeg-dev

# Create non-root user
RUN addgroup -g 1001 -S crypt && \
    adduser -u 1001 -S crypt -G crypt

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/crypt .

# Change ownership to non-root user
RUN chown -R crypt:crypt /app

# Switch to non-root user
USER crypt

# Expose port (if needed)
# EXPOSE 8080

# Set entrypoint
ENTRYPOINT ["./crypt"]

# Default command
CMD ["help"]
