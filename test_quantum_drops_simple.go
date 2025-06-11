package main

import (
	"log"
	"time"

	"QLP/internal/models"
	"QLP/internal/packaging"
)

func main() {
	log.Println("💧 QUANTUM DROPS GENERATION TEST")
	log.Println("=================================")

	// Test QuantumDrops generation
	testQuantumDropsGeneration()
	
	log.Println("✅ QUANTUM DROPS TEST COMPLETED!")
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
			Output:        createInfraOutput(),
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
			Output:        createKubernetesOutput(),
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
			Output:        createMainAppOutput(),
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
			Output:        createAuthHandlersOutput(),
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
			Output:        createUnitTestsOutput(),
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
			Output:        createIntegrationTestsOutput(),
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
			Output:        createAPIDocsOutput(),
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
			Output:        createSetupGuideOutput(),
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
			Output:        createSecurityAnalysisOutput(),
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
	for i, drop := range drops {
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

func createInfraOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "Dockerfile": "FROM golang:1.21-alpine AS builder\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN go build -o main .\n\nFROM alpine:latest\nRUN apk --no-cache add ca-certificates\nWORKDIR /root/\nCOPY --from=builder /app/main .\nEXPOSE 8080\nCMD [\"./main\"]",
    "docker-compose.yml": "version: '3.8'\nservices:\n  app:\n    build: .\n    ports:\n      - \"8080:8080\"\n    depends_on:\n      - postgres\n  postgres:\n    image: postgres:15\n    environment:\n      POSTGRES_DB: myapp\n      POSTGRES_USER: user\n      POSTGRES_PASSWORD: password\n    ports:\n      - \"5432:5432\""
  }
}
=== SANDBOX EXECUTION ===
Infrastructure setup completed successfully.`
}

func createKubernetesOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "k8s/deployment.yaml": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: auth-service\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: auth-service\n  template:\n    metadata:\n      labels:\n        app: auth-service\n    spec:\n      containers:\n      - name: auth-service\n        image: auth-service:latest\n        ports:\n        - containerPort: 8080",
    "k8s/service.yaml": "apiVersion: v1\nkind: Service\nmetadata:\n  name: auth-service\nspec:\n  selector:\n    app: auth-service\n  ports:\n  - protocol: TCP\n    port: 80\n    targetPort: 8080\n  type: LoadBalancer"
  }
}
=== SANDBOX EXECUTION ===
Kubernetes configuration created successfully.`
}

func createMainAppOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "go.mod": "module auth-microservice\n\ngo 1.21\n\nrequire (\n\tgithub.com/golang-jwt/jwt/v5 v5.0.0\n\tgithub.com/gorilla/mux v1.8.0\n)",
    "cmd/main.go": "package main\n\nimport (\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n\n\t\"auth-microservice/internal/handlers\"\n\t\"github.com/gorilla/mux\"\n)\n\nfunc main() {\n\tr := mux.NewRouter()\n\tr.HandleFunc(\"/health\", handlers.HealthCheck).Methods(\"GET\")\n\tr.HandleFunc(\"/register\", handlers.Register).Methods(\"POST\")\n\tr.HandleFunc(\"/login\", handlers.Login).Methods(\"POST\")\n\n\tport := os.Getenv(\"PORT\")\n\tif port == \"\" {\n\t\tport = \"8080\"\n\t}\n\n\tlog.Printf(\"Server starting on port %s\", port)\n\tlog.Fatal(http.ListenAndServe(\":\"+port, r))\n}"
  }
}
=== SANDBOX EXECUTION ===
Main application created successfully.`
}

func createAuthHandlersOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "internal/handlers/auth.go": "package handlers\n\nimport (\n\t\"encoding/json\"\n\t\"net/http\"\n)\n\nfunc HealthCheck(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode(map[string]string{\"status\": \"healthy\"})\n}\n\nfunc Register(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tw.WriteHeader(http.StatusCreated)\n\tjson.NewEncoder(w).Encode(map[string]string{\"message\": \"User registered\"})\n}\n\nfunc Login(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode(map[string]string{\"token\": \"sample-jwt-token\"})\n}",
    "internal/middleware/auth.go": "package middleware\n\nimport (\n\t\"net/http\"\n)\n\nfunc AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {\n\treturn func(w http.ResponseWriter, r *http.Request) {\n\t\tauthHeader := r.Header.Get(\"Authorization\")\n\t\tif authHeader == \"\" {\n\t\t\thttp.Error(w, \"Authorization required\", http.StatusUnauthorized)\n\t\t\treturn\n\t\t}\n\t\tnext(w, r)\n\t}\n}"
  }
}
=== SANDBOX EXECUTION ===
Authentication handlers created successfully.`
}

func createUnitTestsOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "tests/handlers_test.go": "package tests\n\nimport (\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"testing\"\n\n\t\"auth-microservice/internal/handlers\"\n)\n\nfunc TestHealthCheck(t *testing.T) {\n\treq, err := http.NewRequest(\"GET\", \"/health\", nil)\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\trr := httptest.NewRecorder()\n\thandler := http.HandlerFunc(handlers.HealthCheck)\n\thandler.ServeHTTP(rr, req)\n\n\tif status := rr.Code; status != http.StatusOK {\n\t\tt.Errorf(\"Wrong status: got %v want %v\", status, http.StatusOK)\n\t}\n}\n\nfunc TestRegister(t *testing.T) {\n\treq, err := http.NewRequest(\"POST\", \"/register\", nil)\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\trr := httptest.NewRecorder()\n\thandler := http.HandlerFunc(handlers.Register)\n\thandler.ServeHTTP(rr, req)\n\n\tif status := rr.Code; status != http.StatusCreated {\n\t\tt.Errorf(\"Wrong status: got %v want %v\", status, http.StatusCreated)\n\t}\n}"
  }
}
=== SANDBOX EXECUTION ===
Unit tests created successfully.`
}

func createIntegrationTestsOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "tests/integration_test.go": "package tests\n\nimport (\n\t\"testing\"\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"bytes\"\n\t\"encoding/json\"\n\n\t\"auth-microservice/internal/handlers\"\n)\n\nfunc TestAuthFlow(t *testing.T) {\n\tt.Run(\"Complete Auth Flow\", func(t *testing.T) {\n\t\t// Register user\n\t\tuser := map[string]string{\"username\": \"test\", \"password\": \"pass\"}\n\t\tjsonData, _ := json.Marshal(user)\n\t\treq := httptest.NewRequest(\"POST\", \"/register\", bytes.NewBuffer(jsonData))\n\t\tw := httptest.NewRecorder()\n\t\thandlers.Register(w, req)\n\n\t\tif w.Code != http.StatusCreated {\n\t\t\tt.Fatalf(\"Registration failed: %d\", w.Code)\n\t\t}\n\n\t\t// Login user\n\t\treq = httptest.NewRequest(\"POST\", \"/login\", bytes.NewBuffer(jsonData))\n\t\tw = httptest.NewRecorder()\n\t\thandlers.Login(w, req)\n\n\t\tif w.Code != http.StatusOK {\n\t\t\tt.Fatalf(\"Login failed: %d\", w.Code)\n\t\t}\n\t})\n}"
  }
}
=== SANDBOX EXECUTION ===
Integration tests created successfully.`
}

func createAPIDocsOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "README.md": "# Authentication Microservice\n\nA production-ready Go microservice for user authentication.\n\n## Features\n\n- User registration and login\n- JWT token authentication\n- Health check endpoint\n- Docker containerization\n\n## API Endpoints\n\n### Health Check\n```\nGET /health\n```\n\n### User Registration\n```\nPOST /register\n```\n\n### User Login\n```\nPOST /login\n```\n\n## Quick Start\n\n1. `docker-compose up`\n2. Access at `http://localhost:8080`",
    "docs/api.md": "# API Documentation\n\n## Authentication Endpoints\n\n### POST /register\n\nRegisters a new user.\n\n**Request:**\n```json\n{\n  \"username\": \"string\",\n  \"password\": \"string\"\n}\n```\n\n**Response:**\n- Status: 201 Created\n\n### POST /login\n\nAuthenticates user and returns JWT token.\n\n**Response:**\n- Status: 200 OK\n- Body: `{\"token\": \"jwt_token\"}`"
  }
}
=== SANDBOX EXECUTION ===
API documentation created successfully.`
}

func createSetupGuideOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "docs/setup.md": "# Setup Guide\n\n## Prerequisites\n\n- Go 1.21+\n- Docker\n- PostgreSQL\n\n## Installation\n\n1. Clone repository\n2. Install dependencies: `go mod download`\n3. Set environment variables\n4. Run: `go run cmd/main.go`\n\n## Docker Setup\n\n1. Build: `docker build -t auth-service .`\n2. Run: `docker-compose up`\n\n## Environment Variables\n\n- `PORT`: Server port (default: 8080)\n- `DB_HOST`: Database host\n- `JWT_SECRET`: JWT signing secret"
  }
}
=== SANDBOX EXECUTION ===
Setup guide created successfully.`
}

func createSecurityAnalysisOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "reports/security_analysis.md": "# Security Analysis Report\n\n## Overview\n\nComprehensive security analysis of the authentication microservice.\n\n## Findings\n\n### High Priority\n- JWT secret should be configurable\n- Password hashing not implemented\n\n### Medium Priority\n- Rate limiting not configured\n- CORS headers missing\n\n### Low Priority\n- Request logging recommended\n\n## Recommendations\n\n1. Implement bcrypt password hashing\n2. Add environment-based JWT secret\n3. Configure rate limiting\n4. Add CORS middleware\n\n## Security Score: 65/100\n\nThe service has basic security measures but requires improvements before production deployment.",
    "reports/compliance.md": "# Compliance Report\n\n## Standards Checked\n\n- OWASP Top 10\n- JWT Best Practices\n- Go Security Guidelines\n\n## Compliance Score: 70%\n\n### Compliant\n- ✅ Input validation\n- ✅ JWT token usage\n- ✅ HTTPS ready\n\n### Non-Compliant\n- ❌ Password storage\n- ❌ Rate limiting\n- ❌ Audit logging"
  }
}
=== SANDBOX EXECUTION ===
Security analysis completed successfully.`
}