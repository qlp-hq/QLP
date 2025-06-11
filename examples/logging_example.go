package main

import (
	"time"

	"QLP/internal/logger"
)

func main() {
	// Initialize logger with different configurations
	
	// Example 1: Development mode (console output, colored)
	devConfig := logger.Config{
		Level:      logger.DEBUG,
		Format:     logger.CONSOLE,
		OutputPath: "stdout",
		Caller:     true,
		Stacktrace: true,
	}
	
	logger.InitLogger(devConfig)
	
	// Basic logging
	logger.Logger.Info("Application started")
	logger.Logger.Debug("Debug information")
	logger.Logger.Warn("Warning message")
	
	// Structured logging with context
	logger.WithComponent("orchestrator").Info("Component initialized")
	logger.WithAgent("QLD-AGT-123").Info("Agent created")
	logger.WithTask("QL-DEV-001").Info("Task started")
	
	// Performance logging
	start := time.Now()
	time.Sleep(100 * time.Millisecond) // Simulate work
	duration := time.Since(start).Milliseconds()
	
	logger.LogPerformance("task_execution", duration, true)
	
	// Agent metrics
	logger.LogAgentMetrics("QLD-AGT-123", "QL-DEV-001", duration, 85, true)
	
	// Validation metrics
	logger.LogValidationMetrics("QL-DEV-001", 90, 75, 88, 84, true)
	
	// Error logging
	logger.LogError("database_connection", 
		nil, 
		map[string]interface{}{
			"database": "postgresql",
			"host":     "localhost",
			"port":     5432,
		})
	
	// Production mode example (JSON output)
	prodConfig := logger.Config{
		Level:      logger.INFO,
		Format:     logger.JSON,
		OutputPath: "/tmp/qlp.log",
		Caller:     false,
		Stacktrace: false,
	}
	
	logger.InitLogger(prodConfig)
	
	// Same logs but in JSON format
	logger.Logger.Info("Production logging enabled")
	logger.LogIntentMetrics("QLI-12345", 8, 45000, 92)
	
	logger.Sync()
}