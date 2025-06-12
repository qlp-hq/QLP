#!/bin/bash

# QLP Development Environment Setup Script
set -e

echo "🚀 Setting up QuantumLayer Platform Development Environment"
echo "============================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check Go
    if ! command -v go &> /dev/null; then
        print_warning "Go is not installed. You'll need Go for local development."
    else
        print_success "Go $(go version | awk '{print $3}') is available"
    fi
    
    # Check curl
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed. Please install curl for API testing."
        exit 1
    fi
    
    print_success "Prerequisites check completed"
}

# Clean up existing containers and volumes
cleanup() {
    print_status "Cleaning up existing containers and volumes..."
    
    # Stop and remove containers
    docker-compose down -v --remove-orphans 2>/dev/null || true
    
    # Remove dangling images
    docker image prune -f 2>/dev/null || true
    
    print_success "Cleanup completed"
}

# Build all services
build_services() {
    print_status "Building all microservices..."
    
    # Build with no cache to ensure fresh builds
    docker-compose build --no-cache
    
    print_success "All services built successfully"
}

# Start services
start_services() {
    print_status "Starting all services..."
    
    # Start database first
    print_status "Starting PostgreSQL database..."
    docker-compose up -d postgres
    
    # Wait for database to be ready
    print_status "Waiting for database to be ready..."
    timeout=60
    while ! docker-compose exec -T postgres pg_isready -U qlp_user -d qlp_db &>/dev/null; do
        if [ $timeout -le 0 ]; then
            print_error "Database failed to start within 60 seconds"
            exit 1
        fi
        sleep 2
        timeout=$((timeout-2))
        echo -n "."
    done
    echo
    print_success "Database is ready"
    
    # Start core services
    print_status "Starting core services..."
    docker-compose up -d llm-service validation-service
    
    # Wait a bit for core services
    sleep 10
    
    # Start remaining services
    print_status "Starting remaining services..."
    docker-compose up -d data-service worker-service packaging-service agent-service orchestrator-service
    
    # Wait for services to be ready
    sleep 15
    
    # Start API Gateway last
    print_status "Starting API Gateway..."
    docker-compose up -d api-gateway
    
    print_success "All services started"
}

# Health check
health_check() {
    print_status "Performing health checks..."
    
    services=(
        "data-service:8081"
        "worker-service:8082" 
        "packaging-service:8083"
        "orchestrator-service:8084"
        "llm-service:8085"
        "agent-service:8086"
        "validation-service:8087"
        "api-gateway:8080"
    )
    
    for service in "${services[@]}"; do
        name=$(echo $service | cut -d: -f1)
        port=$(echo $service | cut -d: -f2)
        
        print_status "Checking $name on port $port..."
        
        # Wait up to 30 seconds for service to be ready
        timeout=30
        while ! curl -f http://localhost:$port/health &>/dev/null; do
            if [ $timeout -le 0 ]; then
                print_error "$name health check failed"
                return 1
            fi
            sleep 2
            timeout=$((timeout-2))
            echo -n "."
        done
        echo
        print_success "$name is healthy"
    done
    
    print_success "All services are healthy"
}

# Show service status
show_status() {
    print_status "Service Status:"
    echo
    docker-compose ps
    echo
    
    print_status "Service URLs:"
    echo "🌐 API Gateway:        http://localhost:8080"
    echo "🗄️  Data Service:       http://localhost:8081"
    echo "⚙️  Worker Service:     http://localhost:8082"
    echo "📦 Packaging Service:  http://localhost:8083"
    echo "🔄 Orchestrator:       http://localhost:8084"
    echo "🤖 LLM Service:        http://localhost:8085"
    echo "🎭 Agent Service:      http://localhost:8086"
    echo "✅ Validation Service: http://localhost:8087"
    echo
    print_status "Database URLs:"
    echo "🐘 PostgreSQL:         localhost:5432"
    echo "🔍 Qdrant Vector DB:   http://localhost:6333"
    echo "📊 Redis Cache:        localhost:6379 (optional)"
    echo "🦙 Ollama LLM:         http://localhost:11434 (optional)"
    echo
    
    print_status "API Gateway Endpoints:"
    echo "📊 Health:   GET  http://localhost:8080/health"
    echo "📈 Metrics:  GET  http://localhost:8080/metrics"
    echo "ℹ️  Status:   GET  http://localhost:8080/api/v1/status"
    echo "🔧 Config:   GET  http://localhost:8080/api/v1/config"
    echo "🌍 Services: GET  http://localhost:8080/api/v1/services"
    echo
}

# Show logs
show_logs() {
    if [ $# -eq 0 ]; then
        print_status "Showing logs for all services..."
        docker-compose logs -f
    else
        print_status "Showing logs for $1..."
        docker-compose logs -f $1
    fi
}

# Stop services
stop_services() {
    print_status "Stopping all services..."
    docker-compose down
    print_success "All services stopped"
}

# Main execution
case "${1:-start}" in
    "start")
        check_prerequisites
        cleanup
        build_services
        start_services
        health_check
        show_status
        
        print_success "🎉 QuantumLayer Platform is running!"
        echo
        print_status "Next steps:"
        echo "• Test the API: ./scripts/test-api.sh"
        echo "• View logs: ./scripts/dev-setup.sh logs [service-name]"
        echo "• Stop services: ./scripts/dev-setup.sh stop"
        ;;
        
    "stop")
        stop_services
        ;;
        
    "status")
        show_status
        ;;
        
    "health")
        health_check
        ;;
        
    "logs")
        shift
        show_logs $@
        ;;
        
    "restart")
        stop_services
        start_services
        health_check
        show_status
        ;;
        
    "clean")
        print_status "Performing deep cleanup..."
        docker-compose down -v --remove-orphans
        docker system prune -f
        docker volume prune -f
        print_success "Deep cleanup completed"
        ;;
        
    "help"|*)
        echo "QuantumLayer Platform Development Environment"
        echo
        echo "Usage: $0 [command]"
        echo
        echo "Commands:"
        echo "  start     Start all services (default)"
        echo "  stop      Stop all services"
        echo "  restart   Restart all services"
        echo "  status    Show service status"
        echo "  health    Check service health"
        echo "  logs      Show logs for all services"
        echo "  logs [service]  Show logs for specific service"
        echo "  clean     Deep cleanup (removes all containers and volumes)"
        echo "  help      Show this help message"
        echo
        echo "Examples:"
        echo "  $0                    # Start all services"
        echo "  $0 logs api-gateway   # Show API gateway logs"
        echo "  $0 health             # Check all service health"
        ;;
esac