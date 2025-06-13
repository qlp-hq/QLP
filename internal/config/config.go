package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadEnv loads environment variables from .env file if it exists
func LoadEnv() {
	// Try to load .env file
	file, err := os.Open(".env")
	if err != nil {
		// .env file doesn't exist, which is fine
		return
	}
	defer file.Close()

	fmt.Printf("📝 Loading environment from .env file\n")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set by system environment
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("⚠️  Error reading .env file: %v\\n", err)
	}
}

// GetEnvOrDefault returns environment variable value or default
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetKafkaBrokers returns a slice of Kafka brokers from the environment variable.
func GetKafkaBrokers() []string {
	brokersStr := os.Getenv("KAFKA_BROKERS")
	if brokersStr == "" {
		return []string{}
	}
	return strings.Split(brokersStr, ",")
}
