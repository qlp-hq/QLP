package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"QLP/internal/agents"
	"QLP/internal/deployment/azure"
	"QLP/internal/logger"
	"QLP/internal/packaging"
	"QLP/internal/types"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	
	// Initialize logger
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger.Logger = zapLogger
	
	agentLogger := logger.GetDefaultLogger().WithComponent("azure_deployment_test")
	agentLogger.Info("Starting Azure deployment validation test")
	
	// Check Azure configuration
	if err := checkAzureConfig(); err != nil {
		agentLogger.Error("Azure configuration check failed", zap.Error(err))
		os.Exit(1)
	}
	
	// Create test QuantumDrop
	testDrop := createTestQuantumDrop()
	agentLogger.Info("Created test QuantumDrop", 
		zap.String("id", testDrop.ID),
		zap.Int("file_count", len(testDrop.Files)),
	)
	
	// Configure Azure deployment
	azureConfig := azure.ClientConfig{
		SubscriptionID: getEnvOrDefault("AZURE_SUBSCRIPTION_ID", "test-subscription"),
		Location:       getEnvOrDefault("AZURE_LOCATION", "westeurope"),
		TenantID:       getEnvOrDefault("AZURE_TENANT_ID", ""),
	}
	
	// Configure deployment validator
	deploymentConfig := agents.DeploymentValidatorConfig{
		AzureConfig:           azureConfig,
		CostLimitUSD:         5.00, // $5 limit for testing
		TTL:                  30 * time.Minute, // 30 minute cleanup
		EnableHealthChecks:   true,
		EnableFunctionalTests: true,
		CleanupPolicy:        azure.CleanupPolicy{
			MaxAge:        time.Hour,
			RetryAttempts: 3,
			RetryDelay:    30 * time.Second,
		},
	}
	
	// Create mock LLM client for testing
	mockLLM := &mockLLMClient{}
	
	// Create deployment validator agent
	agentLogger.Info("Creating deployment validator agent")
	agent, err := agents.NewDeploymentValidatorAgent(
		"test-deployment-validator-001",
		mockLLM,
		testDrop,
		deploymentConfig,
	)
	if err != nil {
		agentLogger.Error("Failed to create deployment validator agent", zap.Error(err))
		os.Exit(1)
	}
	
	// Create test task
	task := types.Task{
		ID:          "test-azure-deployment-001",
		Type:        "deployment-validation",
		Description: "Test Azure deployment validation with sample Go web server",
		Priority:    types.TaskPriorityHigh,
		Status:      types.TaskStatusPending,
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"test_mode":     true,
			"cost_limit":    5.00,
			"timeout_minutes": 30,
		},
	}
	
	// Execute deployment validation
	agentLogger.Info("Starting deployment validation execution")
	startTime := time.Now()
	
	result, err := agent.Execute(ctx, task)
	if err != nil {
		agentLogger.Error("Deployment validation failed", zap.Error(err))
		
		// Print partial results if available
		if result != nil {
			printTestResults(result, agentLogger)
		}
		os.Exit(1)
	}
	
	duration := time.Since(startTime)
	agentLogger.Info("Deployment validation completed",
		zap.Duration("total_duration", duration),
		zap.String("result_status", string(result.Status)),
	)
	
	// Print comprehensive test results
	printTestResults(result, agentLogger)
	
	// Test cleanup functionality
	agentLogger.Info("Testing cleanup functionality")
	if err := agent.Cleanup(ctx); err != nil {
		agentLogger.Warn("Cleanup failed", zap.Error(err))
	} else {
		agentLogger.Info("Cleanup completed successfully")
	}
	
	agentLogger.Info("Azure deployment validation test completed successfully!")
}

func checkAzureConfig() error {
	required := []string{
		"AZURE_SUBSCRIPTION_ID",
		"AZURE_TENANT_ID",
	}
	
	optional := []string{
		"AZURE_CLIENT_ID",
		"AZURE_CLIENT_SECRET",
		"AZURE_LOCATION",
	}
	
	fmt.Println("🔍 Checking Azure Configuration...")
	
	missing := []string{}
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		} else {
			fmt.Printf("✅ %s: %s\n", env, maskValue(os.Getenv(env)))
		}
	}
	
	for _, env := range optional {
		if value := os.Getenv(env); value != "" {
			fmt.Printf("✅ %s: %s\n", env, maskValue(value))
		} else {
			fmt.Printf("⚠️  %s: not set (using default)\n", env)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}
	
	fmt.Println("✅ Azure configuration check passed")
	return nil
}

func createTestQuantumDrop() *packaging.QuantumDrop {
	return &packaging.QuantumDrop{
		ID:          "test-quantum-drop-001",
		Type:        packaging.DropTypeCodebase,
		Name:        "Azure Test Web Server",
		Description: "Simple Go web server for Azure deployment validation testing",
		Status:      packaging.DropStatusReady,
		CreatedAt:   time.Now(),
		Files: map[string]string{
			"main.go": `package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status    string    ` + "`json:\"status\"`" + `
	Timestamp time.Time ` + "`json:\"timestamp\"`" + `
	Version   string    ` + "`json:\"version\"`" + `
	Uptime    string    ` + "`json:\"uptime\"`" + `
}

var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/metrics", metricsHandler)

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "🚀 QuantumLayer Test Server - Azure Deployment Validation\\n")
	fmt.Fprintf(w, "Uptime: %v\\n", time.Since(startTime))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    time.Since(startTime).String(),
	}
	
	json.NewEncoder(w).Encode(response)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "# HELP uptime_seconds Server uptime in seconds\\n")
	fmt.Fprintf(w, "# TYPE uptime_seconds counter\\n")
	fmt.Fprintf(w, "uptime_seconds %d\\n", int(time.Since(startTime).Seconds()))
}`,
			"go.mod": `module azure-test-server

go 1.21`,
			"Dockerfile": `FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]`,
			"infrastructure/main.tf": `terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_container_group" "test_app" {
  name                = "quantum-test-app"
  location            = var.location
  resource_group_name = var.resource_group_name
  ip_address_type     = "Public"
  dns_name_label      = "quantum-test-${random_string.suffix.result}"
  os_type             = "Linux"

  container {
    name   = "web-server"
    image  = "nginx:alpine"
    cpu    = "0.5"
    memory = "1.0"

    ports {
      port     = 8080
      protocol = "TCP"
    }
  }

  tags = {
    Environment = "test"
    Purpose     = "quantum-validation"
  }
}

resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "West Europe"
}

variable "resource_group_name" {
  description = "Resource group name"
  type        = string
}

output "app_url" {
  value = "http://${azurerm_container_group.test_app.fqdn}:8080"
}

output "health_check_url" {
  value = "http://${azurerm_container_group.test_app.fqdn}:8080/health"
}`,
			"tests/integration_test.go": `package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	baseURL := "http://localhost:8080"
	
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}
	
	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
}

func TestMetricsEndpoint(t *testing.T) {
	baseURL := "http://localhost:8080"
	
	resp, err := http.Get(baseURL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to call metrics endpoint: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}`,
			"README.md": `# Azure Test Web Server

Simple Go web server for testing QuantumLayer Azure deployment validation.

## Endpoints

- GET / - Home page with uptime information
- GET /health - Health check endpoint (JSON response)
- GET /metrics - Prometheus-style metrics

## Deployment

This application is designed to be deployed to Azure Container Instances via Terraform.

## Testing

Run integration tests:
` + "```bash\ngo test ./tests/\n```" + `,`,
		},
		Metadata: packaging.DropMetadata{
			FileCount:    7,
			TotalLines:   200,
			Technologies: []string{"go", "terraform", "azure-container-instances"},
		},
		Tasks: []string{"test-task-001"},
	}
}

func printTestResults(result *types.TaskResult, logger logger.Interface) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 AZURE DEPLOYMENT VALIDATION RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Printf("Task ID: %s\n", result.TaskID)
	fmt.Printf("Agent ID: %s\n", result.AgentID)
	fmt.Printf("Status: %s\n", result.Status)
	fmt.Printf("Duration: %v\n", result.EndTime.Sub(result.StartTime))
	
	if result.ErrorMessage != "" {
		fmt.Printf("❌ Error: %s\n", result.ErrorMessage)
	}
	
	fmt.Println("\n📈 METADATA:")
	for key, value := range result.Metadata {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	fmt.Println("\n📎 ATTACHMENTS:")
	for filename, data := range result.Attachments {
		fmt.Printf("  %s (%d bytes)\n", filename, len(data))
		
		// Print JSON attachments in readable format
		if filename == "deployment_result.json" && len(data) > 0 {
			var deploymentResult map[string]interface{}
			if err := json.Unmarshal(data, &deploymentResult); err == nil {
				fmt.Println("    🏗️  Deployment Details:")
				if status, ok := deploymentResult["status"]; ok {
					fmt.Printf("      Status: %v\n", status)
				}
				if cost, ok := deploymentResult["cost_estimate"]; ok {
					fmt.Printf("      Cost Estimate: %v\n", cost)
				}
			}
		}
	}
	
	if result.Output != "" {
		fmt.Println("\n📄 OUTPUT:")
		fmt.Println(result.Output)
	}
	
	fmt.Println(strings.Repeat("=", 60))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func maskValue(value string) string {
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}

// mockLLMClient provides a simple mock implementation for testing
type mockLLMClient struct{}

func (m *mockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	return "Mock LLM response for testing", nil
}

func (m *mockLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}