# QLP API Gateway

The API Gateway serves as the central entry point for all client requests to the QuantumLayer Platform microservices. It provides routing, authentication, rate limiting, circuit breaking, and load balancing capabilities.

## Features

- **Service Discovery & Routing**: Automatic routing to backend microservices
- **Authentication**: JWT, API Key, and Basic Auth support
- **Rate Limiting**: Configurable rate limiting by IP, tenant, user, or API key
- **Circuit Breaking**: Fault tolerance with automatic service failure detection
- **Load Balancing**: Round-robin and weighted load balancing
- **Health Monitoring**: Real-time health checks for all backend services
- **CORS Support**: Configurable Cross-Origin Resource Sharing
- **Request/Response Transformation**: Header manipulation and path rewriting
- **Comprehensive Logging**: Structured logging with request tracing
- **Metrics & Monitoring**: Built-in metrics and status endpoints

## Architecture

The API Gateway coordinates requests across these microservices:

- **Data Service** (port 8081) - Intent and data management
- **Worker Runtime Service** (port 8082) - Runtime operations and execution
- **Packaging Service** (port 8083) - Capsule and quantum drops packaging
- **Orchestrator Service** (port 8084) - Workflow orchestration and DAG execution
- **LLM Service** (port 8085) - AI completion, embeddings, and chat
- **Agent Service** (port 8086) - Dynamic agent lifecycle management
- **Validation Service** (port 8087) - Validation and quality assurance

## Configuration

Configure the gateway using environment variables:

### Core Settings
- `GATEWAY_PORT`: Gateway port (default: 8080)
- `ENVIRONMENT`: Environment name (development, staging, production)

### Feature Toggles
- `ENABLE_CORS`: Enable CORS support (default: true)
- `ENABLE_RATE_LIMIT`: Enable rate limiting (default: true)
- `ENABLE_AUTH`: Enable authentication (default: false)

### Service URLs
- `DATA_SERVICE_URL`: Data service URL (default: http://localhost:8081)
- `WORKER_SERVICE_URL`: Worker service URL (default: http://localhost:8082)
- `PACKAGING_SERVICE_URL`: Packaging service URL (default: http://localhost:8083)
- `ORCHESTRATOR_SERVICE_URL`: Orchestrator service URL (default: http://localhost:8084)
- `LLM_SERVICE_URL`: LLM service URL (default: http://localhost:8085)
- `AGENT_SERVICE_URL`: Agent service URL (default: http://localhost:8086)
- `VALIDATION_SERVICE_URL`: Validation service URL (default: http://localhost:8087)

### Authentication
- `AUTH_ENABLED`: Enable authentication (default: false)
- `AUTH_TYPE`: Authentication type (jwt, api_key, basic)
- `JWT_SECRET`: JWT signing secret
- `JWT_ISSUER`: JWT issuer name (default: qlp-gateway)

### Rate Limiting
- `RATE_LIMIT_ENABLED`: Enable rate limiting (default: true)
- `RATE_LIMIT_RPS`: Requests per second (default: 100)
- `RATE_LIMIT_BURST`: Burst size (default: 200)
- `RATE_LIMIT_WINDOW`: Time window (default: 1m)
- `RATE_LIMIT_KEY_FUNC`: Rate limit key function (ip, tenant, user, api_key)

### Circuit Breaker
- `CIRCUIT_BREAKER_ENABLED`: Enable circuit breaker (default: true)
- `CIRCUIT_BREAKER_FAILURE_THRESHOLD`: Failure threshold percentage (default: 5)
- `CIRCUIT_BREAKER_RECOVERY_TIMEOUT`: Recovery timeout (default: 30s)
- `CIRCUIT_BREAKER_MONITORING_PERIOD`: Monitoring period (default: 10s)
- `CIRCUIT_BREAKER_MIN_REQUESTS`: Minimum requests for evaluation (default: 3)

### Timeouts
- `TIMEOUT_READ`: Read timeout (default: 30s)
- `TIMEOUT_WRITE`: Write timeout (default: 30s)
- `TIMEOUT_IDLE`: Idle timeout (default: 60s)
- `TIMEOUT_REQUEST`: Request timeout (default: 300s)

### Health Checks
- `HEALTH_CHECK_ENABLED`: Enable health checks (default: true)
- `HEALTH_CHECK_INTERVAL`: Check interval (default: 30s)
- `HEALTH_CHECK_TIMEOUT`: Check timeout (default: 5s)
- `HEALTH_CHECK_PATH`: Health check path (default: /health)

## API Endpoints

### Gateway Management

- `GET /health` - Gateway health check
- `GET /metrics` - Prometheus-compatible metrics
- `GET /api/v1/status` - Detailed gateway status
- `GET /api/v1/services` - Backend service information
- `GET /api/v1/config` - Gateway configuration (non-sensitive)

### Proxied Service Routes

All service routes are proxied through the gateway with the `/api/v1/tenants/{tenantId}` prefix:

#### Data Service
- `GET|POST|PUT|DELETE /api/v1/tenants/{tenantId}/intents/**`

#### Worker Runtime Service  
- `GET|POST /api/v1/tenants/{tenantId}/runtime/**`

#### Packaging Service
- `GET|POST /api/v1/tenants/{tenantId}/capsules/**`
- `GET|POST /api/v1/tenants/{tenantId}/quantum-drops/**`

#### Orchestrator Service
- `GET|POST|PUT|DELETE /api/v1/tenants/{tenantId}/workflows/**`
- `POST /api/v1/dag/validate`

#### LLM Service
- `POST /api/v1/tenants/{tenantId}/completion`
- `POST /api/v1/tenants/{tenantId}/embedding`
- `POST /api/v1/tenants/{tenantId}/chat/**`
- `GET /api/v1/providers`

#### Agent Service
- `GET|POST /api/v1/tenants/{tenantId}/agents/**`

#### Validation Service
- `POST /api/v1/tenants/{tenantId}/validate`

## Docker

Build the gateway:
```bash
docker build -t qlp-api-gateway .
```

Run the gateway:
```bash
docker run -p 8080:8080 \
  -e DATA_SERVICE_URL=http://data-service:8081 \
  -e WORKER_SERVICE_URL=http://worker-service:8082 \
  -e PACKAGING_SERVICE_URL=http://packaging-service:8083 \
  -e ORCHESTRATOR_SERVICE_URL=http://orchestrator-service:8084 \
  -e LLM_SERVICE_URL=http://llm-service:8085 \
  -e AGENT_SERVICE_URL=http://agent-service:8086 \
  -e VALIDATION_SERVICE_URL=http://validation-service:8087 \
  qlp-api-gateway
```

## Development

Run locally:
```bash
go run ./api-gateway/cmd/main.go
```

Test gateway health:
```bash
curl http://localhost:8080/health
```

Test service routing:
```bash
curl http://localhost:8080/api/v1/tenants/test/intents
```

## Authentication

### JWT Authentication

Include JWT token in requests:
```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/tenants/test/intents
```

### API Key Authentication

Include API key in requests:
```bash
curl -H "Authorization: Bearer <api-key>" \
  http://localhost:8080/api/v1/tenants/test/intents

# Or using X-API-Key header
curl -H "X-API-Key: <api-key>" \
  http://localhost:8080/api/v1/tenants/test/intents
```

### Basic Authentication

Include basic auth credentials:
```bash
curl -u "username:password" \
  http://localhost:8080/api/v1/tenants/test/intents
```

## Rate Limiting

The gateway returns rate limit headers:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1609459200
```

When rate limit is exceeded:
```json
{
  "error": "rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Circuit Breaker

Services are automatically monitored for failures. When a service becomes unhealthy:

```json
{
  "error": "service temporarily unavailable", 
  "code": "CIRCUIT_BREAKER_OPEN",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

Circuit breaker states:
- **Closed**: Normal operation, requests flow through
- **Open**: Service is failing, requests are rejected
- **Half-Open**: Testing if service has recovered

## Monitoring

### Health Check Response

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "services": {
    "data": {
      "healthy": true,
      "last_check": "2024-01-01T12:00:00Z",
      "last_error": null
    }
  }
}
```

### Status Response

```json
{
  "gateway": {
    "status": "running",
    "version": "1.0.0", 
    "uptime": "2h30m15s",
    "environment": "production",
    "features": {
      "authentication": false,
      "rate_limit": true,
      "circuit_breaker": true,
      "cors": true,
      "health_checks": true
    }
  },
  "services": {
    "data": {...},
    "worker": {...}
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Metrics Response

```json
{
  "gateway": {
    "total_requests": 1000,
    "successful_requests": 950,
    "failed_requests": 50,
    "average_response_time_ms": 150.5,
    "requests_per_second": 10.2,
    "error_rate_percent": 5.0
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Error Handling

Standard error response format:

```json
{
  "error": "description of the error",
  "code": "ERROR_CODE",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

Common error codes:
- `ROUTE_NOT_FOUND`: Requested route does not exist
- `METHOD_NOT_ALLOWED`: HTTP method not allowed for route
- `SERVICE_NOT_FOUND`: Backend service not configured
- `SERVICE_UNAVAILABLE`: Backend service is down
- `RATE_LIMIT_EXCEEDED`: Rate limit exceeded
- `CIRCUIT_BREAKER_OPEN`: Circuit breaker is open
- `AUTH_FAILED`: Authentication failed
- `BACKEND_ERROR`: Error from backend service

## Load Balancing

The gateway supports multiple load balancing strategies:

- **Round Robin**: Distributes requests evenly across instances
- **Weighted**: Distributes based on instance weights
- **Least Connections**: Routes to instance with fewest active connections

Configure service instances:
```json
{
  "instances": [
    {
      "id": "data-1",
      "url": "http://data-service-1:8081",
      "weight": 100,
      "healthy": true
    },
    {
      "id": "data-2", 
      "url": "http://data-service-2:8081",
      "weight": 100,
      "healthy": true
    }
  ]
}
```

## Security

- **Input Validation**: All requests are validated before proxying
- **Header Sanitization**: Malicious headers are removed
- **Request/Response Logging**: Comprehensive audit trail
- **Rate Limiting**: Prevents abuse and DDoS attacks
- **Circuit Breaking**: Protects against cascading failures
- **Authentication**: Configurable auth mechanisms
- **CORS**: Configurable cross-origin policies

## Performance

- **Connection Pooling**: Efficient backend connections
- **Request Multiplexing**: Concurrent request handling
- **Caching**: Response caching for improved performance
- **Compression**: Automatic response compression
- **Keep-Alive**: HTTP keep-alive for reduced latency
- **Timeout Management**: Configurable timeouts prevent hanging requests

## Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qlp-api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: qlp-api-gateway
  template:
    metadata:
      labels:
        app: qlp-api-gateway
    spec:
      containers:
      - name: api-gateway
        image: qlp-api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: ENABLE_AUTH
          value: "true"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Docker Compose

```yaml
version: '3.8'
services:
  api-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - ENABLE_AUTH=true
      - DATA_SERVICE_URL=http://data-service:8081
      - WORKER_SERVICE_URL=http://worker-service:8082
    depends_on:
      - data-service
      - worker-service
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```