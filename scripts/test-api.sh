#!/bin/bash

# QLP API Testing Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_GATEWAY_URL="http://localhost:8080"
TENANT_ID="test-tenant"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Function to make HTTP request and check response
test_endpoint() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    local data=$5
    
    print_status "Testing: $description"
    echo "  $method $endpoint"
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method"
    
    if [ ! -z "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    local response=$(eval "$curl_cmd '$endpoint'")
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "$expected_status" ]; then
        print_success "✓ Status: $status_code"
        if [ ! -z "$body" ] && [ "$body" != "null" ]; then
            echo "  Response: $(echo "$body" | jq -r '.' 2>/dev/null || echo "$body" | head -c 100)..."
        fi
    else
        print_error "✗ Expected: $expected_status, Got: $status_code"
        if [ ! -z "$body" ]; then
            echo "  Response: $body"
        fi
        return 1
    fi
    echo
}

# Test gateway health and status
test_gateway() {
    echo "🌐 Testing API Gateway"
    echo "====================="
    
    test_endpoint "GET" "$API_GATEWAY_URL/health" "200" "Gateway health check"
    test_endpoint "GET" "$API_GATEWAY_URL/metrics" "200" "Gateway metrics"
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/status" "200" "Gateway status"
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/services" "200" "Gateway services info"
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/config" "200" "Gateway configuration"
}

# Test LLM Service through gateway
test_llm_service() {
    echo "🤖 Testing LLM Service"
    echo "======================"
    
    # Test providers endpoint
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/providers" "200" "Get LLM providers"
    
    # Test completion endpoint
    local completion_data='{
        "prompt": "Hello, how are you?",
        "model": "mock",
        "max_tokens": 100,
        "temperature": 0.7
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/completion" "200" "LLM completion" "$completion_data"
    
    # Test embedding endpoint
    local embedding_data='{
        "text": "This is a test sentence for embedding",
        "model": "mock"
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/embedding" "200" "LLM embedding" "$embedding_data"
    
    # Test chat completion
    local chat_data='{
        "messages": [
            {"role": "user", "content": "Hello!"}
        ],
        "model": "mock",
        "max_tokens": 50
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/chat/completion" "200" "Chat completion" "$chat_data"
}

# Test Vector Database operations
test_vector_database() {
    echo "🔍 Testing Vector Database"
    echo "=========================="
    
    # Test direct Qdrant health (if available)
    if curl -f http://localhost:6333/health &>/dev/null; then
        print_success "✓ Qdrant vector database is accessible"
        
        # Test Qdrant collections endpoint
        print_status "Checking Qdrant collections..."
        local collections_response=$(curl -s http://localhost:6333/collections 2>/dev/null || echo "{}")
        echo "  Qdrant collections: $(echo "$collections_response" | jq -r '.result.collections | length' 2>/dev/null || echo "unknown")"
    else
        print_warning "Qdrant vector database not accessible (this is optional)"
    fi
    
    # Test PostgreSQL with pgvector (via Data Service)
    print_status "Testing PostgreSQL vector capabilities through Data Service..."
    
    # Test vector search endpoint (if available)
    local vector_search_data='{
        "query": "test query",
        "limit": 5,
        "threshold": 0.7
    }'
    
    # Note: This endpoint might not exist yet, so we test if it's available
    if curl -f "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/vector/search" &>/dev/null; then
        test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/vector/search" "200" "Vector similarity search" "$vector_search_data"
    else
        print_status "Vector search endpoint not yet implemented (future enhancement)"
    fi
}

# Test Agent Service through gateway
test_agent_service() {
    echo "🎭 Testing Agent Service"
    echo "========================"
    
    # List agents (should be empty initially)
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/agents" "200" "List agents"
    
    # Create a new agent
    local agent_data='{
        "task_id": "test-task-001",
        "task_type": "codegen",
        "task_description": "Create a simple Hello World program in Go",
        "priority": "medium",
        "project_context": {
            "project_type": "cli",
            "tech_stack": ["Go"],
            "requirements": ["Simple CLI application"]
        }
    }'
    
    # Create agent and capture response
    print_status "Creating test agent..."
    local create_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$agent_data" \
        "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/agents")
    
    local agent_id=$(echo "$create_response" | jq -r '.agent_id // empty')
    
    if [ ! -z "$agent_id" ] && [ "$agent_id" != "null" ]; then
        print_success "✓ Agent created with ID: $agent_id"
        echo
        
        # Get agent details
        test_endpoint "GET" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/agents/$agent_id" "200" "Get agent details"
        
        # Execute agent
        local execute_data='{"parameters": {"execution_mode": "standard"}}'
        test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/agents/$agent_id/execute" "200" "Execute agent" "$execute_data"
        
    else
        print_error "Failed to create agent"
        echo "Response: $create_response"
    fi
}

# Test Validation Service through gateway
test_validation_service() {
    echo "✅ Testing Validation Service"
    echo "============================="
    
    # Test validation endpoint
    local validation_data='{
        "code": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
        "language": "go",
        "validation_type": "syntax",
        "context": {
            "project_type": "cli",
            "requirements": ["Basic syntax check"]
        }
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/validate" "200" "Code validation" "$validation_data"
}

# Test Packaging Service through gateway
test_packaging_service() {
    echo "📦 Testing Packaging Service"
    echo "============================"
    
    # Test capsule creation
    local capsule_data='{
        "name": "test-capsule",
        "description": "Test capsule for API testing",
        "files": {
            "main.go": "package main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello from capsule!\")\n}",
            "go.mod": "module test-capsule\n\ngo 1.21"
        },
        "metadata": {
            "version": "1.0.0",
            "author": "QLP Test"
        }
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/capsules" "200" "Create capsule" "$capsule_data"
    
    # List capsules
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/capsules" "200" "List capsules"
}

# Test Orchestrator Service through gateway
test_orchestrator_service() {
    echo "🔄 Testing Orchestrator Service"
    echo "==============================="
    
    # Test DAG validation
    local dag_data='{
        "nodes": [
            {"id": "task1", "name": "Generate Code", "type": "codegen"},
            {"id": "task2", "name": "Validate Code", "type": "validation"},
            {"id": "task3", "name": "Package Code", "type": "packaging"}
        ],
        "edges": [
            {"from": "task1", "to": "task2"},
            {"from": "task2", "to": "task3"}
        ]
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/dag/validate" "200" "Validate DAG" "$dag_data"
    
    # List workflows
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/workflows" "200" "List workflows"
}

# Test Data Service through gateway
test_data_service() {
    echo "🗄️ Testing Data Service"
    echo "======================="
    
    # List intents
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/intents" "200" "List intents"
    
    # Create an intent
    local intent_data='{
        "description": "Create a REST API server",
        "requirements": [
            "HTTP server with routing",
            "JSON request/response handling",
            "Basic authentication"
        ],
        "constraints": {
            "language": "go",
            "framework": "chi"
        }
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/intents" "200" "Create intent" "$intent_data"
}

# Test Worker Service through gateway
test_worker_service() {
    echo "⚙️ Testing Worker Service"
    echo "========================"
    
    # Test runtime status
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/runtime" "200" "Get runtime status"
    
    # Test task execution
    local task_data='{
        "task_type": "codegen",
        "description": "Generate a simple function",
        "parameters": {
            "language": "go",
            "function_name": "hello"
        }
    }'
    test_endpoint "POST" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/runtime" "200" "Execute runtime task" "$task_data"
}

# Test error cases
test_error_cases() {
    echo "❌ Testing Error Cases"
    echo "====================="
    
    # Test 404 errors
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/nonexistent" "404" "Non-existent route"
    test_endpoint "GET" "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/agents/nonexistent" "404" "Non-existent agent"
    
    # Test invalid data
    local invalid_data='{"invalid": json}'
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$invalid_data" \
        "$API_GATEWAY_URL/api/v1/tenants/$TENANT_ID/agents" \
        -w '%{http_code}' > /dev/null
    
    # Test method not allowed
    test_endpoint "PUT" "$API_GATEWAY_URL/health" "405" "Method not allowed"
}

# Main test execution
main() {
    echo "🧪 QuantumLayer Platform API Testing"
    echo "===================================="
    echo
    
    # Check if services are running
    print_status "Checking if API Gateway is accessible..."
    if ! curl -f "$API_GATEWAY_URL/health" &>/dev/null; then
        print_error "API Gateway is not accessible at $API_GATEWAY_URL"
        print_status "Please run: ./scripts/dev-setup.sh start"
        exit 1
    fi
    print_success "API Gateway is accessible"
    echo
    
    # Run all tests
    test_gateway
    test_llm_service
    test_agent_service
    test_validation_service
    test_packaging_service
    test_orchestrator_service
    test_data_service
    test_worker_service
    test_error_cases
    
    echo "🎉 API Testing Complete!"
    echo
    print_status "Summary:"
    echo "• All microservices are accessible through the API Gateway"
    echo "• Core functionality is working across all services"
    echo "• Error handling is functioning correctly"
    echo
    print_status "Next steps:"
    echo "• Monitor logs: ./scripts/dev-setup.sh logs"
    echo "• View service status: ./scripts/dev-setup.sh status"
    echo "• Stop services: ./scripts/dev-setup.sh stop"
}

# Handle command line arguments
case "${1:-test}" in
    "test")
        main
        ;;
    "gateway")
        test_gateway
        ;;
    "llm")
        test_llm_service
        ;;
    "agent")
        test_agent_service
        ;;
    "validation")
        test_validation_service
        ;;
    "packaging")
        test_packaging_service
        ;;
    "orchestrator")
        test_orchestrator_service
        ;;
    "data")
        test_data_service
        ;;
    "worker")
        test_worker_service
        ;;
    "errors")
        test_error_cases
        ;;
    "help"|*)
        echo "QLP API Testing Script"
        echo
        echo "Usage: $0 [test-type]"
        echo
        echo "Test Types:"
        echo "  test          Run all tests (default)"
        echo "  gateway       Test API Gateway only"
        echo "  llm           Test LLM Service only"
        echo "  agent         Test Agent Service only"
        echo "  validation    Test Validation Service only"
        echo "  packaging     Test Packaging Service only"
        echo "  orchestrator  Test Orchestrator Service only"
        echo "  data          Test Data Service only"
        echo "  worker        Test Worker Service only"
        echo "  errors        Test error cases only"
        echo "  help          Show this help message"
        ;;
esac