# LLM Service

The LLM Service provides a unified interface for multiple Large Language Model providers in the QuantumLayer Platform. It supports text completion, embeddings, chat completion, and batch processing with automatic fallback between providers.

## Features

- **Multi-Provider Support**: Azure OpenAI, Ollama, and Mock providers
- **Automatic Fallback**: Seamless failover between providers
- **Health Monitoring**: Continuous health checks and provider status tracking
- **Batch Processing**: Efficient batch processing for multiple requests
- **Comprehensive Metrics**: Detailed usage and performance metrics
- **Tenant Isolation**: Multi-tenant request handling and isolation

## Supported Providers

### Azure OpenAI
- Full OpenAI API compatibility
- Production-ready with enterprise features
- Requires `AZURE_OPENAI_API_KEY` and `AZURE_OPENAI_ENDPOINT`

### Ollama
- Local and self-hosted LLM support
- Configurable via `OLLAMA_BASE_URL` and `OLLAMA_MODEL`
- Default: `http://localhost:11434` with `llama3` model

### Mock Provider
- Development and testing support
- Always available as fallback
- Generates realistic mock responses

## API Endpoints

### Core LLM Operations

- `POST /api/v1/tenants/{tenantId}/completion` - Text completion
- `POST /api/v1/tenants/{tenantId}/embedding` - Generate embeddings
- `POST /api/v1/tenants/{tenantId}/chat/completion` - Chat completion
- `POST /api/v1/tenants/{tenantId}/batch` - Batch processing

### Service Management

- `GET /api/v1/providers` - List available providers
- `GET /api/v1/status` - Service status and provider health
- `GET /api/v1/metrics` - Usage metrics and statistics

### Health & Monitoring

- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus-compatible metrics

## Configuration

Configure the service using environment variables:

### Required
- `PORT`: Server port (default: 8085)

### Azure OpenAI
- `AZURE_OPENAI_API_KEY`: Azure OpenAI API key
- `AZURE_OPENAI_ENDPOINT`: Azure OpenAI endpoint URL
- `AZURE_OPENAI_MODEL`: Model name (default: gpt-4)

### Ollama
- `OLLAMA_BASE_URL`: Ollama server URL (default: http://localhost:11434)
- `OLLAMA_MODEL`: Model name (default: llama3)

### Logging
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `LOG_FORMAT`: Log format (json, console)

### Database (Optional)
- `DATABASE_URL`: Database connection for tenant resolution

## Request Examples

### Text Completion
```json
POST /api/v1/tenants/tenant1/completion
{
  "prompt": "Explain quantum computing",
  "max_tokens": 1000,
  "temperature": 0.7,
  "system_prompt": "You are a helpful technical assistant."
}
```

### Chat Completion
```json
POST /api/v1/tenants/tenant1/chat/completion
{
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "What is machine learning?"}
  ],
  "max_tokens": 500,
  "temperature": 0.3
}
```

### Generate Embedding
```json
POST /api/v1/tenants/tenant1/embedding
{
  "text": "This is text to embed"
}
```

### Batch Processing
```json
POST /api/v1/tenants/tenant1/batch
{
  "type": "completion",
  "requests": [
    {"prompt": "Question 1", "max_tokens": 100},
    {"prompt": "Question 2", "max_tokens": 100}
  ]
}
```

## Response Format

All responses include provider information and usage metrics:

```json
{
  "content": "Response content",
  "model": "gpt-4",
  "provider": "azure-openai",
  "response_time": "1.5s",
  "request_id": "comp_1234567890",
  "usage": {
    "prompt_tokens": 15,
    "completion_tokens": 50,
    "total_tokens": 65
  }
}
```

## Docker

Build the service:
```bash
docker build -t llm-service .
```

Run with Azure OpenAI:
```bash
docker run -p 8085:8085 \
  -e AZURE_OPENAI_API_KEY=your-key \
  -e AZURE_OPENAI_ENDPOINT=your-endpoint \
  llm-service
```

Run with Ollama:
```bash
docker run -p 8085:8085 \
  -e OLLAMA_BASE_URL=http://host.docker.internal:11434 \
  llm-service
```

## Development

Run locally:
```bash
go run ./services/llm-service/cmd/main.go
```

Test with mock provider:
```bash
curl -X POST http://localhost:8085/api/v1/tenants/test/completion \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello world"}'
```

## Health Monitoring

The service provides comprehensive health monitoring:

- **Provider Health**: Individual provider status and response times
- **Service Metrics**: Request counts, error rates, and latency statistics
- **Automatic Recovery**: Unhealthy providers are automatically disabled and re-enabled when recovered

## Security

- **Tenant Isolation**: All operations are scoped to specific tenants
- **Input Validation**: All requests are validated for security and correctness
- **Provider Isolation**: Provider failures don't affect other providers
- **API Key Security**: Secure handling of provider API keys

## Performance

- **Connection Pooling**: Efficient HTTP connection management
- **Concurrent Processing**: Parallel processing for batch operations
- **Timeout Management**: Configurable timeouts for different operations
- **Automatic Failover**: Sub-second failover between providers

## Monitoring and Observability

- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Metrics Collection**: Detailed metrics for monitoring and alerting
- **Health Checks**: Kubernetes-compatible health check endpoints
- **Tracing**: Request tracing across provider calls