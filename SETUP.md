# QuantumLayer Setup Guide

## 🚀 Quick Start

### 1. Environment Configuration

Copy the environment template and configure your API keys:

```bash
cp .env.example .env
```

Edit `.env` with your actual values:

```bash
# Required: Azure OpenAI API Key
AZURE_OPENAI_API_KEY=sk-proj-your-key-here
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/

# Optional: Ollama for local LLM fallback
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=llama3

# Optional: PostgreSQL for persistence
DATABASE_URL=postgres://qlp_user:qlp_password@localhost:5432/qlp_db?sslmode=disable
```

### 2. Build and Run

```bash
# Build the application
go build -o qlp .

# Run with any intent
./qlp "Create a secure REST API for user management"

# Or use interactive mode
./qlp
> 1 (Interactive mode)
> Enter your intent: Create a microservices platform
```

## 🔧 Configuration Options

### LLM Providers (Fallback Chain)

1. **Azure OpenAI** (Primary) - Requires `AZURE_OPENAI_API_KEY`
2. **Ollama** (Fallback) - Local models via `OLLAMA_BASE_URL`
3. **Mock Client** (Final fallback) - Always works for testing

### Database Options

- **PostgreSQL** - Full persistence + vector similarity search
- **File-based** - Graceful fallback when DB unavailable

### Validation Levels

- `standard` - Basic validation (default)
- `enterprise` - Full 3-layer validation with compliance

## 🎯 Example Usage

```bash
# Simple intent
./qlp "Build a TODO API"

# Complex enterprise intent
./qlp "Create a HIPAA-compliant patient management system with JWT auth"

# Interactive mode with suggestions
./qlp
```

## 🔍 Features Available

✅ **Dynamic Intent Processing** - Any natural language input  
✅ **Multi-layer Validation** - Security, quality, compliance  
✅ **Vector Similarity Search** - "Similar to previous projects"  
✅ **HITL Decision Engine** - Automated quality gates  
✅ **QuantumCapsule Generation** - Complete project packages  
✅ **Enterprise Documentation** - Available at [docs](https://qlp-hq.github.io/QLP/)  

## 🛠️ Development Setup

### With PostgreSQL + Vector Search

```bash
# Install PostgreSQL with pgvector
brew install postgresql pgvector

# Create database
createdb qlp_db
psql qlp_db < internal/database/schema.sql
```

### With Ollama Local LLM

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull models
ollama pull llama3
ollama pull codellama
```

## 🔒 Security Notes

- Never commit `.env` files
- Use environment variables in production
- API keys are loaded from environment only
- Database connections use secure defaults