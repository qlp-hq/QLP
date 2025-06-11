package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"QLP/internal/orchestrator"
)

func main() {
	fmt.Println("🚀 QuantumLayer Universal Agent Orchestration System")
	fmt.Println("============================================")
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	orch := orchestrator.New()

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down QuantumLayer...")
		cancel()
	}()

	// Run production-grade demo
	fmt.Println("📋 Running production-grade end-to-end demo...")
	fmt.Println()

	if err := runProductionDemo(ctx, orch); err != nil {
		log.Fatalf("❌ Demo failed: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Demo completed successfully!")
	fmt.Println("🔄 QuantumLayer is now ready for interactive use...")
	fmt.Println()

	<-ctx.Done()
}

func runProductionDemo(ctx context.Context, o *orchestrator.Orchestrator) error {
	demoIntents := []string{
		"Create a secure REST API for user management with JWT authentication",
		"Build infrastructure for a microservices deployment on Kubernetes",
		"Analyze the performance of a Go web application and generate optimization recommendations",
	}

	for i, intentText := range demoIntents {
		fmt.Printf("🎯 Demo %d/3: %s\n", i+1, intentText)
		fmt.Println("=" + string(make([]byte, len(intentText)+20)))
		
		startTime := time.Now()
		
		if err := o.ProcessAndExecuteIntent(ctx, intentText); err != nil {
			return fmt.Errorf("failed to process intent %d: %w", i+1, err)
		}
		
		duration := time.Since(startTime)
		fmt.Printf("⏱️  Completed in %v\n", duration)
		fmt.Println()
		
		// Brief pause between demos
		time.Sleep(2 * time.Second)
	}

	return nil
}
