# Orchestrator Service

The Orchestrator Service is responsible for workflow execution, task coordination, and DAG (Directed Acyclic Graph) validation in the QuantumLayer Platform.

## Features

- **Workflow Execution**: Manages complete workflow lifecycles from start to completion
- **Task Coordination**: Coordinates task execution based on dependencies and priorities
- **DAG Validation**: Validates task graphs for cycles and dependency correctness
- **Workflow Control**: Provides pause, resume, cancel, and retry operations
- **Progress Tracking**: Real-time workflow and task progress monitoring
- **Metrics Collection**: Comprehensive metrics for workflow performance analysis

## API Endpoints

### Workflow Management

- `POST /api/v1/tenants/{tenantId}/workflows` - Execute a new workflow
- `GET /api/v1/tenants/{tenantId}/workflows` - List workflows with pagination
- `GET /api/v1/tenants/{tenantId}/workflows/{workflowId}` - Get workflow details

### Workflow Control

- `POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/pause` - Pause workflow execution
- `POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/resume` - Resume paused workflow
- `POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/cancel` - Cancel workflow execution
- `POST /api/v1/tenants/{tenantId}/workflows/{workflowId}/retry` - Retry failed task

### DAG Operations

- `POST /api/v1/dag/validate` - Validate DAG structure

### Metrics

- `GET /api/v1/tenants/{tenantId}/workflows/{workflowId}/metrics` - Get workflow metrics

### Health & Monitoring

- `GET /health` - Health check endpoint
- `GET /metrics` - Service metrics endpoint

## Configuration

The service can be configured using environment variables:

- `PORT`: Server port (default: 8084)
- `DATABASE_URL`: Database connection string for tenant resolution
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `LOG_FORMAT`: Log format (json, console)

## Docker

Build the service:
```bash
docker build -t orchestrator-service .
```

Run the service:
```bash
docker run -p 8084:8084 orchestrator-service
```

## Development

Run the service locally:
```bash
go run ./services/orchestrator-service/cmd/main.go
```

## Architecture

The service consists of several key components:

### Engines

- **DAGEngine**: Handles DAG validation, cycle detection, and execution order generation
- **WorkflowEngine**: Manages workflow execution state, task coordination, and control operations

### Handlers

- **OrchestratorHandler**: HTTP request handlers for all workflow and DAG operations

### Client

- **OrchestratorClient**: HTTP client for communicating with the orchestrator service

## Workflow Execution Flow

1. **Validation**: Incoming workflow requests are validated for DAG correctness
2. **Initialization**: Workflow execution context is created and stored
3. **Task Scheduling**: Tasks are scheduled based on dependencies and priority
4. **Execution**: Tasks are executed asynchronously with concurrency limits
5. **Monitoring**: Progress is tracked and metrics are collected
6. **Completion**: Final results are aggregated and workflow is marked complete

## Error Handling

The service implements comprehensive error handling:

- **Validation Errors**: Invalid DAG structures or missing required fields
- **Execution Errors**: Task failures, timeouts, and resource constraints
- **Control Errors**: Invalid state transitions (e.g., pausing completed workflows)
- **System Errors**: Infrastructure failures and service unavailability

## Security

- **Tenant Isolation**: All operations are scoped to specific tenants
- **Input Validation**: All requests are validated for security and correctness
- **Access Control**: Tenant-based access control for workflow operations