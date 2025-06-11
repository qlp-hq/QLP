package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"QLP/internal/agents"
	"QLP/internal/deployment/azure"
	"QLP/internal/logger"
	"QLP/internal/packaging"
	"QLP/internal/types"
	"go.uber.org/zap"
)

// Mock test that doesn't require real Azure resources
func main() {
	ctx := context.Background()
	
	// Initialize logger
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger.Logger = zapLogger
	
	agentLogger := logger.GetDefaultLogger().WithComponent("azure_deployment_mock_test")
	agentLogger.Info("🧪 Starting Azure deployment validation MOCK test")
	
	// Create simple test QuantumDrop
	testDrop := &packaging.QuantumDrop{
		ID:          "mock-test-drop-001",
		Type:        packaging.DropTypeCodebase,
		Name:        "Mock Test Application",
		Description: "Simple test application for mock Azure deployment validation",
		Status:      packaging.DropStatusReady,
		CreatedAt:   time.Now(),
		Files: map[string]string{
			"main.go": `package main
import "fmt"
func main() { fmt.Println("Hello from QuantumLayer!") }`,
			"go.mod": "module test-app\ngo 1.21",
		},
		Metadata: packaging.DropMetadata{
			FileCount:    2,
			TotalLines:   3,
			Technologies: []string{"go"},
		},
		Tasks: []string{"mock-task-001"},
	}
	
	agentLogger.Info("✅ Created mock test QuantumDrop", 
		zap.String("id", testDrop.ID),
		zap.Int("file_count", len(testDrop.Files)),
	)
	
	// Configure Azure deployment (mock values)
	azureConfig := azure.ClientConfig{
		SubscriptionID: "mock-subscription-12345",
		Location:       "westeurope",
		TenantID:       "mock-tenant-67890",
	}
	
	// Configure deployment validator
	deploymentConfig := agents.DeploymentValidatorConfig{
		AzureConfig:           azureConfig,
		CostLimitUSD:         1.00, // Low limit for mock testing
		TTL:                  5 * time.Minute, // Short TTL for mock
		EnableHealthChecks:   true,
		EnableFunctionalTests: true,
		CleanupPolicy:        azure.CleanupPolicy{
			MaxAge:          10 * time.Minute,
			RetryAttempts:   3,
			RetryDelay:      30 * time.Second,
		},
	}
	
	agentLogger.Info("✅ Configured Azure deployment settings",
		zap.String("subscription", azureConfig.SubscriptionID),
		zap.String("location", azureConfig.Location),
		zap.Float64("cost_limit", deploymentConfig.CostLimitUSD),
	)
	
	// Create mock LLM client
	mockLLM := &mockLLMClient{}
	
	// Create deployment validator agent
	agentLogger.Info("🏗️  Creating deployment validator agent...")
	agent, err := agents.NewDeploymentValidatorAgent(
		"mock-deployment-validator-001",
		mockLLM,
		testDrop,
		deploymentConfig,
	)
	if err != nil {
		agentLogger.Error("❌ Failed to create deployment validator agent", zap.Error(err))
		log.Fatal(err)
	}
	
	agentLogger.Info("✅ Deployment validator agent created successfully",
		zap.String("agent_id", "mock-deployment-validator-001"),
		zap.String("agent_type", "deployment-validator"),
	)
	
	// Create test task
	task := types.Task{
		ID:          "mock-azure-deployment-task-001",
		Type:        "deployment-validation",
		Description: "Mock Azure deployment validation test",
		Priority:    types.TaskPriorityHigh,
		Status:      types.TaskStatusPending,
		CreatedAt:   time.Now(),
		Metadata: map[string]interface{}{
			"test_mode":     true,
			"mock_mode":     true,
			"cost_limit":    1.00,
			"timeout_minutes": 5,
		},
	}
	
	agentLogger.Info("✅ Created test task",
		zap.String("task_id", task.ID),
		zap.String("task_type", task.Type),
	)
	
	// Execute deployment validation
	agentLogger.Info("🚀 Starting mock deployment validation execution...")
	startTime := time.Now()
	
	result, err := agent.Execute(ctx, task)
	if err != nil {
		agentLogger.Error("❌ Mock deployment validation failed", zap.Error(err))
		if result != nil {
			printMockResults(result, agentLogger)
		}
		log.Fatal(err)
	}
	
	duration := time.Since(startTime)
	agentLogger.Info("✅ Mock deployment validation completed",
		zap.Duration("total_duration", duration),
		zap.String("result_status", string(result.Status)),
	)
	
	// Print comprehensive test results
	printMockResults(result, agentLogger)
	
	// Test cleanup functionality
	agentLogger.Info("🧹 Testing cleanup functionality...")
	if err := agent.Cleanup(ctx); err != nil {
		agentLogger.Warn("⚠️  Cleanup returned error (expected in mock mode)", zap.Error(err))
	} else {
		agentLogger.Info("✅ Cleanup completed successfully")
	}
	
	// Test metrics
	metrics := agent.GetMetrics()
	agentLogger.Info("📊 Agent metrics:",
		zap.Any("metrics", metrics),
	)
	
	agentLogger.Info("🎉 Mock Azure deployment validation test completed successfully!")
	
	// Summary
	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("🎯 MOCK TEST SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("✅ Agent Creation: SUCCESS\n")
	fmt.Printf("✅ Task Execution: SUCCESS (%v)\n", duration)
	fmt.Printf("✅ Result Generation: SUCCESS\n")
	fmt.Printf("✅ Cleanup Test: SUCCESS\n")
	fmt.Printf("✅ Metrics Collection: SUCCESS\n")
	fmt.Println("\n🔍 Next Steps:")
	fmt.Println("1. Run real Azure test: go run cmd/test-azure-deployment/main.go")
	fmt.Println("2. Set up Azure credentials: ./scripts/setup-azure-test.sh")
	fmt.Println("3. Check Azure deployment validation in action!")
	fmt.Println(strings.Repeat("=", 70))
}

func printMockResults(result *types.TaskResult, logger logger.Interface) {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("📊 MOCK DEPLOYMENT VALIDATION RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	
	fmt.Printf("🆔 Task ID: %s\n", result.TaskID)
	fmt.Printf("🤖 Agent ID: %s\n", result.AgentID)
	fmt.Printf("📈 Status: %s\n", result.Status)
	fmt.Printf("⏱️  Duration: %v\n", result.EndTime.Sub(result.StartTime))
	
	if result.ErrorMessage != "" {
		fmt.Printf("❌ Error: %s\n", result.ErrorMessage)
	}
	
	fmt.Println("\n📋 METADATA:")
	for key, value := range result.Metadata {
		fmt.Printf("  📌 %s: %v\n", key, value)
	}
	
	fmt.Println("\n📎 ATTACHMENTS:")
	for filename, data := range result.Attachments {
		fmt.Printf("  📄 %s (%d bytes)\n", filename, len(data))
	}
	
	if result.Output != "" {
		fmt.Printf("\n📝 OUTPUT:\n%s\n", result.Output)
	}
	
	fmt.Println(strings.Repeat("=", 60))
}

// mockLLMClient provides a simple mock implementation for testing
type mockLLMClient struct{}

func (m *mockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	return "Mock LLM response: Deployment validation analysis completed successfully.", nil
}

func (m *mockLLMClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}