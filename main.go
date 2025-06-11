package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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

	// Check if intent provided as command line argument
	if len(os.Args) > 1 {
		// Use provided intent
		intentText := strings.Join(os.Args[1:], " ")
		if err := processSingleIntent(ctx, orch, intentText); err != nil {
			log.Fatalf("❌ Intent processing failed: %v", err)
		}
		return
	}

	// Check for demo mode
	if len(os.Args) == 1 {
		fmt.Println("🎯 Choose mode:")
		fmt.Println("1. Interactive mode (enter your intent)")
		fmt.Println("2. Demo mode (run predefined examples)")
		fmt.Println("3. Exit")
		fmt.Print("\nEnter choice (1-3): ")
		
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			choice := strings.TrimSpace(scanner.Text())
			
			switch choice {
			case "1":
				if err := runInteractiveMode(ctx, orch); err != nil {
					log.Fatalf("❌ Interactive mode failed: %v", err)
				}
			case "2":
				if err := runProductionDemo(ctx, orch); err != nil {
					log.Fatalf("❌ Demo failed: %v", err)
				}
			case "3":
				fmt.Println("👋 Goodbye!")
				return
			default:
				fmt.Println("❌ Invalid choice. Exiting...")
				return
			}
		}
	}

	fmt.Println("\n✅ QuantumLayer session completed!")
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

func processSingleIntent(ctx context.Context, o *orchestrator.Orchestrator, intentText string) error {
	fmt.Printf("🎯 Processing Intent: %s\n", intentText)
	fmt.Println("=" + strings.Repeat("=", len(intentText)+20))
	
	startTime := time.Now()
	
	if err := o.ProcessAndExecuteIntent(ctx, intentText); err != nil {
		return fmt.Errorf("failed to process intent: %w", err)
	}
	
	duration := time.Since(startTime)
	fmt.Printf("⏱️  Completed in %v\n", duration)
	
	return nil
}

func runInteractiveMode(ctx context.Context, o *orchestrator.Orchestrator) error {
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Println("\n🎯 Interactive Mode")
		fmt.Println("Enter your intent (or 'quit' to exit):")
		fmt.Print("> ")
		
		if !scanner.Scan() {
			break
		}
		
		intentText := strings.TrimSpace(scanner.Text())
		
		if intentText == "" {
			fmt.Println("❌ Please enter a valid intent")
			continue
		}
		
		if strings.ToLower(intentText) == "quit" || strings.ToLower(intentText) == "exit" {
			fmt.Println("👋 Exiting interactive mode...")
			break
		}
		
		if err := processSingleIntent(ctx, o, intentText); err != nil {
			fmt.Printf("❌ Error processing intent: %v\n", err)
			fmt.Println("💡 Try again with a different intent...")
			continue
		}
		
		fmt.Println("\n✅ Intent completed successfully!")
		fmt.Println("🔄 Ready for next intent...")
	}
	
	return nil
}
