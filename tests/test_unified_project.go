package main

import (
	"context"
	"log"
	"time"

	"QLP/internal/orchestrator"
)

func main() {
	log.Println("🚀 Testing Unified Project Generation...")

	// Create orchestrator
	orch := orchestrator.New()

	// Start orchestrator
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	// Create a focused intent for a complete microservice
	userInput := "Create a Go JWT authentication microservice with user registration, login endpoints, middleware, tests, and Docker configuration"

	log.Printf("📋 Generating unified microservice project...")

	// Process and execute the complete workflow
	if err := orch.ProcessAndExecuteIntent(ctx, userInput); err != nil {
		log.Fatalf("Failed to process and execute intent: %v", err)
	}

	log.Println("✅ Unified project generation completed!")
	log.Println("📦 Check output/ directory for single cohesive project structure")
}