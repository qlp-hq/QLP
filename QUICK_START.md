# QLP Quick Start Guide 🚀

Get the entire QuantumLayer Platform microservices architecture running in minutes!

## Prerequisites

- Docker & Docker Compose
- curl (for testing)
- jq (optional, for prettier JSON output)

```bash
# Install jq on macOS
brew install jq

# Install jq on Ubuntu/Debian
sudo apt-get install jq
```

## 🎯 One-Command Startup

```bash
./scripts/dev-setup.sh
```

This will:
- ✅ Check prerequisites
- 🧹 Clean up any existing containers
- 🏗️ Build all 7 microservices
- 🚀 Start all services in the correct order
- 🔍 Perform health checks
- 📊 Display service status and URLs

## 🧪 Test Everything

Once services are running, test the entire platform:

```bash
./scripts/test-api.sh
```

This will test:
- 🌐 API Gateway routing and middleware
- 🤖 LLM Service (completion, embedding, chat)
- 🎭 Agent Service (create, execute agents)
- ✅ Validation Service (code validation)
- 📦 Packaging Service (capsule creation)
- 🔄 Orchestrator Service (workflow management)
- 🗄️ Data Service (intent management)
- ⚙️ Worker Service (runtime operations)

## 🌍 Service URLs

Once running, access services at:

| Service | URL | Description |
|---------|-----|-------------|
| **API Gateway** | http://localhost:8080 | Central entry point |
| Data Service | http://localhost:8081 | Intent & data management |
| Worker Service | http://localhost:8082 | Runtime operations |
| Packaging Service | http://localhost:8083 | Capsule & quantum drops |
| Orchestrator Service | http://localhost:8084 | Workflow orchestration |
| LLM Service | http://localhost:8085 | AI completion & embeddings |
| Agent Service | http://localhost:8086 | Dynamic agent management |
| Validation Service | http://localhost:8087 | Quality assurance |

## 📊 Key Endpoints

### API Gateway Management
```bash
# Gateway health
curl http://localhost:8080/health

# Service status
curl http://localhost:8080/api/v1/status | jq

# Service configuration
curl http://localhost:8080/api/v1/services | jq
```

### Example API Calls Through Gateway

```bash
# Create an agent
curl -X POST http://localhost:8080/api/v1/tenants/test/agents \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": "demo-001",
    "task_type": "codegen", 
    "task_description": "Create a Hello World app in Go"
  }' | jq

# Get LLM completion
curl -X POST http://localhost:8080/api/v1/tenants/test/completion \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Write a hello world function in Go",
    "model": "mock",
    "max_tokens": 100
  }' | jq

# Validate code
curl -X POST http://localhost:8080/api/v1/tenants/test/validate \
  -H "Content-Type: application/json" \
  -d '{
    "code": "package main\n\nfunc main() {\n    fmt.Println(\"Hello!\")\n}",
    "language": "go",
    "validation_type": "syntax"
  }' | jq
```

## 🔧 Management Commands

```bash
# View service status
./scripts/dev-setup.sh status

# View logs for all services
./scripts/dev-setup.sh logs

# View logs for specific service
./scripts/dev-setup.sh logs api-gateway

# Stop all services
./scripts/dev-setup.sh stop

# Restart services
./scripts/dev-setup.sh restart

# Clean everything (removes all containers and volumes)
./scripts/dev-setup.sh clean
```

## 🎯 Quick Demo Flow

1. **Start everything:**
   ```bash
   ./scripts/dev-setup.sh
   ```

2. **Check health:**
   ```bash
   curl http://localhost:8080/health | jq
   ```

3. **Create an intent:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/tenants/demo/intents \
     -H "Content-Type: application/json" \
     -d '{
       "description": "Create a REST API server",
       "requirements": ["HTTP routing", "JSON handling"],
       "constraints": {"language": "go"}
     }' | jq
   ```

4. **Create an agent:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/tenants/demo/agents \
     -H "Content-Type: application/json" \
     -d '{
       "task_id": "api-server",
       "task_type": "codegen",
       "task_description": "Generate a REST API server in Go"
     }' | jq
   ```

5. **Test LLM completion:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/tenants/demo/completion \
     -H "Content-Type: application/json" \
     -d '{
       "prompt": "Create a Go HTTP server",
       "model": "mock"
     }' | jq
   ```

6. **Run comprehensive tests:**
   ```bash
   ./scripts/test-api.sh
   ```

## 🏗️ Architecture Overview

```
                    ┌─────────────────┐
                    │   API Gateway   │ :8080
                    │   (Rate Limit,  │
                    │ Circuit Breaker,│
                    │   Auth, CORS)   │
                    └─────────┬───────┘
                              │
         ┌────────────────────┼────────────────────┐
         │                    │                    │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │  Data   │         │ Worker  │         │Packaging│
    │Service  │◄────────┤Service  │◄────────┤Service  │
    │  :8081  │         │  :8082  │         │  :8083  │
    └─────────┘         └─────────┘         └─────────┘
         │                    │                    │
         │              ┌────▼────┐               │
         │              │   LLM   │               │
         │              │Service  │               │
         │              │  :8085  │               │
         │              └────┬────┘               │
         │                   │                    │
    ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
    │Orchest- │         │ Agent   │         │Validat- │
    │rator    │◄────────┤Service  │◄────────┤ion      │
    │  :8084  │         │  :8086  │         │Service  │
    └─────────┘         └─────────┘         │  :8087  │
                                            └─────────┘
```

## 🛠️ Development

- **Logs**: `./scripts/dev-setup.sh logs [service-name]`
- **Debug**: Each service exposes health and metrics endpoints
- **Hot reload**: Modify code and run `./scripts/dev-setup.sh restart`
- **Database**: PostgreSQL available at `localhost:5432`

## 🚨 Troubleshooting

**Services won't start?**
```bash
# Check Docker
docker --version
docker-compose --version

# Clean and rebuild
./scripts/dev-setup.sh clean
./scripts/dev-setup.sh start
```

**Health checks failing?**
```bash
# Check individual service logs
./scripts/dev-setup.sh logs [service-name]

# Check service status
docker-compose ps
```

**API tests failing?**
```bash
# Ensure services are healthy first
./scripts/dev-setup.sh health

# Run specific test
./scripts/test-api.sh gateway
```

## 🎉 Success!

You now have a complete microservices architecture running with:
- ✅ 7 microservices with health monitoring
- ✅ API Gateway with authentication, rate limiting, circuit breaking
- ✅ Service-to-service communication
- ✅ PostgreSQL database
- ✅ Comprehensive API testing
- ✅ Production-ready features

**Next Steps:**
- Explore the API endpoints
- Modify service code and see changes
- Add new features to services
- Deploy to Kubernetes/production