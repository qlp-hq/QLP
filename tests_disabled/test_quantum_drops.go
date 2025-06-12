package main

import (
	"context"
	"log"
	"time"

	"QLP/internal/orchestrator"
)

func main() {
	log.Println("🧪 Testing QuantumDrops → HITL → QuantumCapsule Workflow...")

	// Create orchestrator
	orch := orchestrator.New()

	// Start orchestrator
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	// Create a simple, focused intent for faster testing
	userInput := "Create a simple Go HTTP server with health check endpoint and basic documentation"

	log.Printf("🎯 Testing QuantumDrops workflow with simple intent...")

	// Process and execute the complete QuantumDrops → HITL → QuantumCapsule workflow
	if err := orch.ProcessAndExecuteIntent(ctx, userInput); err != nil {
		log.Fatalf("Failed to process intent with QuantumDrops: %v", err)
	}

	log.Println("✅ QuantumDrops → HITL → QuantumCapsule workflow completed successfully!")
	log.Println("💧 Check output/ directory for categorized QuantumDrops and final unified QuantumCapsule")
}