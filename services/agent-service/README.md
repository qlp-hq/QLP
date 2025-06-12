# Agent Service

The Agent Service manages the lifecycle of dynamic agents that execute various tasks in the QuantumLayer Platform. It provides creation, execution, monitoring, and control capabilities for different types of agents including code generation, infrastructure, testing, documentation, analysis, and deployment validation agents.

## Features

- **Dynamic Agent Creation**: Create specialized agents for different task types
- **Agent Execution**: Execute agents with LLM integration and real-time monitoring
- **Lifecycle Management**: Full agent lifecycle from creation to completion
- **Deployment Validation**: Specialized agents for Azure deployment validation
- **Batch Operations**: Create and manage multiple agents simultaneously
- **Real-time Monitoring**: Comprehensive metrics and status tracking
- **Tenant Isolation**: Multi-tenant agent management and security

## Agent Types

### Dynamic Agents
- **Code Generation**: Creates complete project structures with code
- **Infrastructure**: Generates Kubernetes, Terraform, Docker configurations
- **Testing**: Creates comprehensive test suites and test data
- **Documentation**: Generates technical documentation and API docs
- **Analysis**: Performs code analysis and generates reports

### Specialized Agents
- **Deployment Validator**: Validates deployments on Azure with real infrastructure
- **Security Scanner**: Performs security analysis and vulnerability detection
- **Quality Analyzer**: Analyzes code quality and provides recommendations

## API Endpoints

### Core Agent Operations

- `POST /api/v1/tenants/{tenantId}/agents` - Create a new agent
- `GET /api/v1/tenants/{tenantId}/agents` - List agents with pagination
- `GET /api/v1/tenants/{tenantId}/agents/{agentId}` - Get agent details

### Agent Execution and Control

- `POST /api/v1/tenants/{tenantId}/agents/{agentId}/execute` - Execute an agent
- `POST /api/v1/tenants/{tenantId}/agents/{agentId}/cancel` - Cancel agent execution
- `POST /api/v1/tenants/{tenantId}/agents/{agentId}/retry` - Retry failed agent

### Batch Operations

- `POST /api/v1/tenants/{tenantId}/agents/batch` - Create multiple agents

### Specialized Agents

- `POST /api/v1/tenants/{tenantId}/agents/deployment-validator` - Create deployment validator

### Service Management

- `GET /api/v1/status` - Service status and statistics
- `GET /api/v1/metrics` - Service metrics and performance data

### Health & Monitoring

- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus-compatible metrics

## Configuration

Configure the service using environment variables:

### Required
- `PORT`: Server port (default: 8086)
- `LLM_SERVICE_URL`: URL of the LLM service (default: http://localhost:8085)

### Optional
- `DATABASE_URL`: Database connection for tenant resolution
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `LOG_FORMAT`: Log format (json, console)

## Request Examples

### Create Code Generation Agent
```json
POST /api/v1/tenants/tenant1/agents
{
  "task_id": "task-001",
  "task_type": "codegen",
  "task_description": "Create a REST API server in Go",
  "priority": "high",
  "project_context": {
    "project_type": "web_api",
    "tech_stack": ["Go", "HTTP", "JSON"],
    "requirements": ["RESTful API", "Authentication"],
    "architecture": "microservices"
  },
  "configuration": {
    "enable_sandbox": true,
    "enable_validation": true,
    "timeout": "5m"
  }
}
```

### Execute Agent
```json
POST /api/v1/tenants/tenant1/agents/QLI-AGT-123456-001/execute
{
  "parameters": {
    "execution_mode": "standard"
  }
}
```

### Create Deployment Validator
```json
POST /api/v1/tenants/tenant1/agents/deployment-validator
{
  "agent_id": "validator-001",
  "capsule_data": {
    "id": "capsule-123",
    "name": "test-deployment",
    "files": {
      "main.go": "package main...",
      "Dockerfile": "FROM golang:1.21..."
    },
    "technologies": ["go", "docker"]
  },
  "config": {
    "azure_config": {
      "subscription_id": "sub-123",
      "location": "westeurope"
    },
    "cost_limit_usd": 5.0,
    "ttl": "15m"
  }
}
```

### Batch Create Agents
```json
POST /api/v1/tenants/tenant1/agents/batch
{
  "agents": [
    {
      "task_id": "task-001",
      "task_type": "codegen",
      "task_description": "Create API server"
    },
    {
      "task_id": "task-002", 
      "task_type": "test",
      "task_description": "Create test suite"
    }
  ]
}
```

## Agent Lifecycle

1. **Creation**: Agent is created with specified configuration and context
2. **Ready**: Agent is initialized and ready for execution
3. **Executing**: Agent is running and processing the task
4. **Completed**: Agent has finished successfully with output
5. **Failed**: Agent execution failed with error details
6. **Cancelled**: Agent was cancelled before completion

## Response Format

Agent responses include comprehensive information:

```json
{
  "agent_id": "QLI-AGT-123456-001",
  "status": "completed",
  "message": "Agent execution completed",
  "agent": {
    "id": "QLI-AGT-123456-001",
    "task_id": "task-001",
    "task_type": "codegen",
    "status": "completed",
    "created_at": "2024-01-01T12:00:00Z",
    "completed_at": "2024-01-01T12:05:00Z",
    "duration": "5m",
    "output": "Generated code output...",
    "metrics": {
      "llm_tokens_used": 1500,
      "total_execution_time": "5m",
      "validation_score": 85
    }
  }
}
```

## Docker

Build the service:
```bash
docker build -t agent-service .
```

Run the service:
```bash
docker run -p 8086:8086 \
  -e LLM_SERVICE_URL=http://llm-service:8085 \
  agent-service
```

## Development

Run locally:
```bash
go run ./services/agent-service/cmd/main.go
```

Test agent creation:
```bash
curl -X POST http://localhost:8086/api/v1/tenants/test/agents \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "test-001",
    "task_type": "codegen", 
    "task_description": "Create Hello World app"
  }'
```

## Metrics and Monitoring

The service provides comprehensive metrics:

- **Agent Statistics**: Creation, execution, completion rates
- **Performance Metrics**: Execution times, success rates, error rates
- **Resource Usage**: Memory, CPU, and token consumption
- **Agent Distribution**: By type, status, and tenant

## Architecture

### Factory Pattern
- **AgentFactory**: Manages agent creation and lifecycle
- **Dynamic Agents**: General-purpose agents for various tasks
- **Specialized Agents**: Purpose-built agents for specific operations

### Execution Flow
1. **Request Validation**: Validate agent creation parameters
2. **Agent Initialization**: Create agent with proper configuration
3. **LLM Integration**: Execute tasks using LLM service
4. **Output Processing**: Process and validate agent output
5. **Metrics Collection**: Track performance and usage statistics

## Security

- **Tenant Isolation**: All operations scoped to specific tenants
- **Input Validation**: Comprehensive validation of all requests
- **Resource Limits**: Configurable timeouts and resource constraints
- **Access Control**: Tenant-based access control for agent operations

## Integration

The Agent Service integrates with:

- **LLM Service**: For intelligent task execution
- **Validation Service**: For output quality assessment
- **Orchestrator Service**: For workflow coordination
- **Packaging Service**: For output packaging and delivery

## Error Handling

Comprehensive error handling includes:

- **Validation Errors**: Invalid parameters or configurations
- **Execution Errors**: LLM failures, timeouts, resource constraints
- **System Errors**: Service unavailability, network issues
- **Business Logic Errors**: Task-specific validation failures