package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"QLP/internal/config"
	"QLP/internal/logger"
	"QLP/internal/orchestrator"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables from .env file
	config.LoadEnv()
	
	// Initialize logger from environment
	if err := logger.InitFromEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	
	logger.Logger.Info("Starting QuantumLayer Universal Agent Orchestration System")
	
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
			logger.Logger.Fatal("Intent processing failed",
				zap.Error(err),
				zap.String("intent", intentText))
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
					logger.Logger.Fatal("Interactive mode failed",
						zap.Error(err))
				}
			case "2":
				if err := runProductionDemo(ctx, orch); err != nil {
					logger.Logger.Fatal("Demo failed",
						zap.Error(err))
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

	logger.WithComponent("demo").Info("Starting production demo",
		zap.Int("total_demos", len(demoIntents)))

	for i, intentText := range demoIntents {
		fmt.Printf("🎯 Demo %d/3: %s\n", i+1, intentText)
		fmt.Println("=" + string(make([]byte, len(intentText)+20)))
		
		startTime := time.Now()
		
		logger.WithComponent("demo").Info("Starting demo intent",
			zap.Int("demo_number", i+1),
			zap.String("intent", intentText))
		
		if err := o.ProcessAndExecuteIntent(ctx, intentText); err != nil {
			logger.WithComponent("demo").Error("Demo intent failed",
				zap.Int("demo_number", i+1),
				zap.String("intent", intentText),
				zap.Error(err))
			return fmt.Errorf("failed to process intent %d: %w", i+1, err)
		}
		
		duration := time.Since(startTime)
		fmt.Printf("⏱️  Completed in %v\n", duration)
		fmt.Println()
		
		logger.LogPerformance("demo_intent", duration.Milliseconds(), true)
		
		// Brief pause between demos
		time.Sleep(2 * time.Second)
	}

	logger.WithComponent("demo").Info("Production demo completed successfully",
		zap.Int("total_demos", len(demoIntents)))
	return nil
}

func processSingleIntent(ctx context.Context, o *orchestrator.Orchestrator, intentText string) error {
	fmt.Printf("🎯 Processing Intent: %s\n", intentText)
	fmt.Println("=" + strings.Repeat("=", len(intentText)+20))
	
	startTime := time.Now()
	
	logger.WithComponent("main").Info("Processing single intent",
		zap.String("intent", intentText))
	
	if err := o.ProcessAndExecuteIntent(ctx, intentText); err != nil {
		logger.WithComponent("main").Error("Intent processing failed",
			zap.String("intent", intentText),
			zap.Error(err))
		return fmt.Errorf("failed to process intent: %w", err)
	}
	
	duration := time.Since(startTime)
	fmt.Printf("⏱️  Completed in %v\n", duration)
	
	logger.LogPerformance("single_intent", duration.Milliseconds(), true)
	logger.WithComponent("main").Info("Intent processing completed",
		zap.String("intent", intentText),
		zap.Duration("duration", duration))
	
	return nil
}

func runInteractiveMode(ctx context.Context, o *orchestrator.Orchestrator) error {
	scanner := bufio.NewScanner(os.Stdin)
	
	logger.WithComponent("interactive").Info("Starting interactive mode")
	
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
			logger.WithComponent("interactive").Warn("Empty intent provided")
			continue
		}
		
		if strings.ToLower(intentText) == "quit" || strings.ToLower(intentText) == "exit" {
			fmt.Println("👋 Exiting interactive mode...")
			logger.WithComponent("interactive").Info("User exited interactive mode")
			break
		}
		
		if err := processSingleIntent(ctx, o, intentText); err != nil {
			fmt.Printf("❌ Error processing intent: %v\n", err)
			fmt.Println("💡 Try again with a different intent...")
			logger.WithComponent("interactive").Error("Interactive intent failed",
				zap.String("intent", intentText),
				zap.Error(err))
			continue
		}
		
		fmt.Println("\n✅ Intent completed successfully!")
		fmt.Println("🔄 Ready for next intent...")
	}
	
	logger.WithComponent("interactive").Info("Interactive mode completed")
	return nil
}
