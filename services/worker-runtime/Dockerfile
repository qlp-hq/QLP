# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install Docker CLI for container management
RUN apk add --no-cache docker-cli

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the worker runtime service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker-runtime ./services/worker-runtime/cmd

# Final stage
FROM alpine:latest

# Install Docker CLI and other dependencies
RUN apk --no-cache add ca-certificates docker-cli curl bash

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/worker-runtime .

# Create work directory for sandbox operations
RUN mkdir -p /tmp/qlp-sandbox

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8081/health || exit 1

# Run the service
CMD ["./worker-runtime"]