package main

import (
	"log"
	"time"

	"QLP/internal/models"
	"QLP/internal/packaging"
)

func main() {
	log.Println("💧 QUANTUM DROPS FINAL TEST")
	log.Println("============================")

	// Test QuantumDrops generation
	testQuantumDropsGeneration()
	
	log.Println("✅ QUANTUM DROPS FINAL TEST COMPLETED!")
}

func testQuantumDropsGeneration() {
	log.Println("\n💧 Testing QuantumDrops Generation")
	log.Println("-----------------------------------")
	
	// Create QuantumDrops generator
	generator := packaging.NewQuantumDropGenerator()
	
	// Create sample intent
	intent := models.Intent{
		ID:        "test-intent-drops",
		UserInput: "Create a microservice with authentication, database, tests, and documentation",
		Status:    models.IntentStatusCompleted,
		CreatedAt: time.Now(),
	}
	
	// Create comprehensive task results covering all drop types
	taskResults := []packaging.TaskExecutionResult{
		// Infrastructure tasks
		{
			Task: models.Task{
				ID:          "QL-INF-001",
				Type:        models.TaskTypeInfra,
				Description: "Set up Docker configuration",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"Dockerfile\": \"FROM golang:1.21\", \"docker-compose.yml\": \"version: '3.8'\"}}\n=== SANDBOX EXECUTION ===\nInfrastructure completed.",
			AgentID:       "QLI-AGT-001",
			ExecutionTime: 12 * time.Second,
		},
		{
			Task: models.Task{
				ID:          "QL-INF-002", 
				Type:        models.TaskTypeInfra,
				Description: "Configure Kubernetes deployment",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"k8s/deployment.yaml\": \"apiVersion: apps/v1\", \"k8s/service.yaml\": \"apiVersion: v1\"}}\n=== SANDBOX EXECUTION ===\nKubernetes completed.",
			AgentID:       "QLI-AGT-002",
			ExecutionTime: 15 * time.Second,
		},
		// Code generation tasks
		{
			Task: models.Task{
				ID:          "QL-DEV-003",
				Type:        models.TaskTypeCodegen,
				Description: "Create main Go application",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"go.mod\": \"module auth-microservice\", \"cmd/main.go\": \"package main\"}}\n=== SANDBOX EXECUTION ===\nCode completed.",
			AgentID:       "QLD-AGT-003",
			ExecutionTime: 18 * time.Second,
		},
		{
			Task: models.Task{
				ID:          "QL-DEV-004",
				Type:        models.TaskTypeCodegen,
				Description: "Implement authentication handlers",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"internal/handlers/auth.go\": \"package handlers\", \"internal/middleware/auth.go\": \"package middleware\"}}\n=== SANDBOX EXECUTION ===\nHandlers completed.",
			AgentID:       "QLD-AGT-004", 
			ExecutionTime: 20 * time.Second,
		},
		// Testing tasks
		{
			Task: models.Task{
				ID:          "QL-TST-005",
				Type:        models.TaskTypeTest,
				Description: "Create unit tests",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"tests/handlers_test.go\": \"package tests\"}}\n=== SANDBOX EXECUTION ===\nUnit tests completed.",
			AgentID:       "QLT-AGT-005",
			ExecutionTime: 10 * time.Second,
		},
		{
			Task: models.Task{
				ID:          "QL-TST-006",
				Type:        models.TaskTypeTest,
				Description: "Create integration tests",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"tests/integration_test.go\": \"package tests\"}}\n=== SANDBOX EXECUTION ===\nIntegration tests completed.",
			AgentID:       "QLT-AGT-006",
			ExecutionTime: 14 * time.Second,
		},
		// Documentation tasks
		{
			Task: models.Task{
				ID:          "QL-DOC-007",
				Type:        models.TaskTypeDoc,
				Description: "Create API documentation",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"README.md\": \"# Authentication Microservice\", \"docs/api.md\": \"# API Documentation\"}}\n=== SANDBOX EXECUTION ===\nDocs completed.",
			AgentID:       "QLC-AGT-007",
			ExecutionTime: 8 * time.Second,
		},
		{
			Task: models.Task{
				ID:          "QL-DOC-008",
				Type:        models.TaskTypeDoc,
				Description: "Create setup guide",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"docs/setup.md\": \"# Setup Guide\"}}\n=== SANDBOX EXECUTION ===\nSetup guide completed.",
			AgentID:       "QLC-AGT-008",
			ExecutionTime: 6 * time.Second,
		},
		// Analysis task
		{
			Task: models.Task{
				ID:          "QL-ANL-009",
				Type:        models.TaskTypeAnalyze,
				Description: "Perform security analysis",
			},
			Status:        models.TaskStatusCompleted,
			Output:        "=== LLM OUTPUT ===\n{\"files\": {\"reports/security_analysis.md\": \"# Security Analysis\", \"reports/compliance.md\": \"# Compliance Report\"}}\n=== SANDBOX EXECUTION ===\nSecurity analysis completed.",
			AgentID:       "QLA-AGT-009",
			ExecutionTime: 12 * time.Second,
		},
	}
	
	// Generate QuantumDrops
	drops, err := generator.GenerateQuantumDrops(intent, taskResults)
	if err != nil {
		log.Fatalf("Failed to generate QuantumDrops: %v", err)
	}
	
	log.Printf("💧 Generated %d QuantumDrops:", len(drops))
	
	// Analyze each drop
	totalFiles := 0
	for _, drop := range drops {
		log.Printf("\n   🎯 %s (%s)", drop.Name, drop.Type)
		log.Printf("      📁 Files: %d", drop.Metadata.FileCount)
		log.Printf("      📊 Quality Score: %d/100", drop.Metadata.QualityScore)
		log.Printf("      🔒 Security Score: %d/100", drop.Metadata.SecurityScore)
		log.Printf("      ✅ Validation Passed: %v", drop.Metadata.ValidationPassed)
		log.Printf("      🤔 HITL Required: %v", drop.Metadata.HITLRequired)
		log.Printf("      📋 Status: %s", drop.Status)
		log.Printf("      🏷️  Tasks: %v", drop.Tasks)
		
		if len(drop.Files) > 0 {
			log.Printf("      📂 Sample Files:")
			count := 0
			for filePath := range drop.Files {
				log.Printf("         - %s", filePath)
				count++
				if count >= 3 { // Show only first 3 files
					if len(drop.Files) > 3 {
						log.Printf("         ... and %d more files", len(drop.Files)-3)
					}
					break
				}
			}
		}
		
		totalFiles += drop.Metadata.FileCount
	}
	
	// Summary statistics
	log.Printf("\n📊 QuantumDrops Summary:")
	log.Printf("   💧 Total Drops: %d", len(drops))
	log.Printf("   📁 Total Files: %d", totalFiles)
	
	// Count drops by type
	typeCounts := make(map[packaging.DropType]int)
	hitlRequired := 0
	for _, drop := range drops {
		typeCounts[drop.Type]++
		if drop.Metadata.HITLRequired {
			hitlRequired++
		}
	}
	
	log.Printf("   📈 By Type:")
	for dropType, count := range typeCounts {
		log.Printf("      %s: %d", dropType, count)
	}
	log.Printf("   🤔 HITL Required: %d/%d", hitlRequired, len(drops))
	
	// Simulate HITL decisions
	log.Printf("\n🤔 Simulating HITL Decisions:")
	approvedCount := 0
	for _, drop := range drops {
		decision := simulateHITLDecision(drop)
		log.Printf("   %s: %s - %s", drop.Name, decision, getDecisionReason(drop))
		if decision == "APPROVED" {
			approvedCount++
		}
	}
	
	log.Printf("\n✅ HITL Results: %d/%d drops approved", approvedCount, len(drops))
}

func simulateHITLDecision(drop packaging.QuantumDrop) string {
	// Simple decision logic based on validation and quality
	if drop.Metadata.ValidationPassed && drop.Metadata.QualityScore >= 70 {
		return "APPROVED"
	} else if drop.Metadata.QualityScore >= 50 {
		return "MODIFY"
	} else {
		return "REDO"
	}
}

func getDecisionReason(drop packaging.QuantumDrop) string {
	if drop.Metadata.ValidationPassed && drop.Metadata.QualityScore >= 70 {
		return "High quality, meets all criteria"
	} else if drop.Metadata.QualityScore >= 50 {
		return "Good foundation, needs minor improvements"
	} else {
		return "Quality below threshold, requires rework"
	}
}