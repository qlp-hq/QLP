# Stage 1: Build the application
FROM golang:1.24-alpine AS builder
ARG SERVICE_PATH
ENV SERVICE_PATH=${SERVICE_PATH}

WORKDIR /app

# Copy all Go module and workspace files first
COPY go.mod go.sum go.work go.work.sum ./

# Download dependencies. This layer is cached if the mod/work files don't change.
RUN go mod download

# Copy the entire source code
COPY . .

# Ensure the SERVICE_PATH is provided
RUN if [ -z "${SERVICE_PATH}" ]; then echo "Error: SERVICE_PATH build-arg is not set." && exit 1; fi

# --- DIAGNOSTICS ---
RUN echo "--- Go Version ---" && go version
RUN echo "--- Go Env ---" && go env
RUN echo "--- File Listing ---" && ls -lR

# Build the specific service binary using 'go install', which is more reliable for workspaces.
# The binary will be placed in /go/bin/$(basename $SERVICE_PATH)
RUN CGO_ENABLED=0 go build -o /app/main -ldflags="-w -s" "./${SERVICE_PATH}"

# Stage 2: Create the final, minimal image
FROM alpine:latest
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /home/appuser
# Copy the built binary from the builder's default Go bin directory
COPY --from=builder /app/main ./main
USER appuser
CMD ["./main"]