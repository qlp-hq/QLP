# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Run
```bash
# Build the main application
go build -o qlp ./main.go

# Run with a specific intent
./qlp "Create a secure REST API with enterprise validation"

# Run in interactive mode
./qlp

# Run in demo mode  
./qlp
# Then select option 2 for demo mode

# Run comprehensive test suite
./comprehensive_test.sh
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific package tests
go test -v ./internal/agents
go test -v ./internal/orchestrator

# Run integration tests
go test -v ./integration_test.go

# Run performance benchmarks
go test -bench=. -benchmem ./...

# Test specific demos
go run cmd/demo-e2e-simple/main.go
go run cmd/demo-e2e-real-azure/main.go  # Creates REAL Azure resources
go run cmd/quantumlayer-e2e/main.go
```

### Development
```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...

# Clean dependencies
go mod tidy
go mod verify

# Build all packages
go build ./...
```

## High-Level Architecture

### Core System Flow
The system follows an **Intent → Tasks → Agents → Validation → Packaging** pipeline:

1. **Intent Parser** (`internal/parser/intent_parser.go`) converts natural language to structured tasks using LLM
2. **Orchestrator** (`internal/orchestrator/orchestrator.go`) coordinates the entire workflow
3. **DAG Executor** (`internal/dag/executor.go`) executes tasks in parallel based on dependencies  
4. **Dynamic Agents** (`internal/agents/dynamic_agent.go`) are created per task with LLM integration
5. **Validation Engine** (`internal/validation/core/unified_validator.go`) performs multi-layer validation
6. **Packaging System** generates QuantumDrops → HITL workflow → final QLCapsule

### Key Design Patterns

**Agent Factory Pattern**: `internal/agents/factory.go` creates specialized agents (Dynamic, DeploymentValidator) with shared context and LLM clients.

**Event-Driven Architecture**: `internal/events/bus.go` coordinates async communication between orchestrator, agents, and DAG executor.

**Fallback Chain Pattern**: LLM client (`internal/llm/client.go`) tries Azure OpenAI → Ollama → Mock in sequence.

**Validation Pipeline**: Three validation layers (LLM-based static → Dynamic sandbox → Enterprise compliance) with configurable thresholds.

**Real/Mock Toggle**: Azure deployment system (`internal/deployment/azure/client.go`) can switch between real Azure API calls and mock implementations via `azure.SetImplementationMode()`.

### Critical Components

**Orchestrator State Machine**: The orchestrator manages complex state transitions from intent processing through final capsule generation, including HITL decision points and real Azure validation.

**QuantumDrops Workflow**: Intermediate packaging format that allows human-in-the-loop decisions before final capsule generation. Each drop contains files, metadata, and validation results for specific task groups.

**Unified Project Merger**: `internal/packaging/project_merger.go` intelligently combines outputs from multiple agents into a coherent project structure with proper file organization and dependency resolution.

**Vector Similarity Search**: `internal/vector/service.go` stores and retrieves intent embeddings for suggesting similar previous work and improving response quality.

### Environment Configuration

**LLM Provider Priority**: 
1. Azure OpenAI (requires `AZURE_OPENAI_API_KEY` and `AZURE_OPENAI_ENDPOINT`)
2. Ollama local models (requires `OLLAMA_BASE_URL`) 
3. Mock client (always available)

**Database Options**:
- PostgreSQL with pgvector for full persistence and similarity search (`DATABASE_URL`)
- Graceful file-based fallback when database unavailable

**Azure Integration**:
- Real deployment validation requires Azure CLI login (`az login`)
- Uses Azure credential chain: environment variables → managed identity → Azure CLI → interactive
- Resource groups auto-cleanup with TTL tags to prevent cost accumulation

### Output Structure

All generated artifacts go to `./output/` directory:
- `manifest.json` - Capsule metadata and file structure
- `metadata.json` - Execution summary and metrics  
- `tasks/` - Individual task execution results
- `project/` - Unified project structure from merged tasks
- `reports/` - Validation, security, and quality reports

### Testing Strategy

**Unit Tests**: Focus on individual components (agents, parser, validation)
**Integration Tests**: Test full orchestrator workflow with mock LLM
**End-to-End Tests**: Real Azure deployment demos that create actual cloud resources
**Performance Tests**: Benchmark intent parsing and agent execution times
**Scenario Tests**: Complex real-world intents with multiple task dependencies

The comprehensive test suite (`comprehensive_test.sh`) validates the entire pipeline from environment setup through real-world scenario testing.