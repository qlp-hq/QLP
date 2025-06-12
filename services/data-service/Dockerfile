# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the data service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o data-service ./services/data-service/cmd

# Final stage
FROM alpine:latest

# Install ca-certificates and curl for HTTPS and health checks
RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

# Copy the binary
COPY --from=builder /app/data-service .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run the service
CMD ["./data-service"]