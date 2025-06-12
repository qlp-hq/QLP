package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"QLP/internal/models"
	"QLP/internal/packaging"
)

func main() {
	log.Println("🎯 FOCUSED CAPSULE GENERATION TEST")
	log.Println("===================================")

	// Test the capsule packaging directly with simulated task results
	testCapsulePackagingDirectly()
	
	// Test QuantumDrops generation
	testQuantumDropsGeneration()
	
	log.Println("✅ FOCUSED CAPSULE TESTS COMPLETED!")
}

func testCapsulePackagingDirectly() {
	log.Println("\n📦 TEST 1: Direct Capsule Packaging")
	log.Println("------------------------------------")
	
	// Create a capsule packager
	packager := packaging.NewCapsulePackager("./output")
	
	// Create a sample intent
	intent := models.Intent{
		ID:        "test-intent-001",
		UserInput: "Create a simple Go HTTP server with authentication",
		Status:    models.IntentStatusCompleted,
		CreatedAt: time.Now(),
	}
	
	// Create sample task execution results
	taskResults := createSampleTaskResults()
	
	// Package the capsule
	ctx := context.Background()
	capsule, err := packager.PackageCapsule(ctx, intent, taskResults)
	if err != nil {
		log.Fatalf("Failed to package capsule: %v", err)
	}
	
	log.Printf("✅ Capsule created: %s", capsule.Metadata.CapsuleID)
	log.Printf("   📊 Overall Score: %d/100", capsule.Metadata.OverallScore)
	log.Printf("   ✅ Tasks: %d successful, %d failed", capsule.Metadata.SuccessfulTasks, capsule.Metadata.FailedTasks)
	
	// Export as ZIP
	zipData, err := packager.ExportCapsule(ctx, capsule, "zip")
	if err != nil {
		log.Fatalf("Failed to export capsule: %v", err)
	}
	
	// Save to file
	capsuleFile := filepath.Join("./output", capsule.Metadata.CapsuleID+".qlcapsule")
	if err := os.MkdirAll("./output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	
	if err := os.WriteFile(capsuleFile, zipData, 0644); err != nil {
		log.Fatalf("Failed to save capsule file: %v", err)
	}
	
	log.Printf("💾 Capsule saved: %s (%d bytes)", capsuleFile, len(zipData))
	
	// Validate the exported capsule
	validateExportedCapsule(capsuleFile)
}

func testQuantumDropsGeneration() {
	log.Println("\n💧 TEST 2: QuantumDrops Generation")
	log.Println("-----------------------------------")
	
	// Create QuantumDrops generator
	generator := packaging.NewQuantumDropGenerator()
	
	// Create sample intent and task results
	intent := models.Intent{
		ID:        "test-intent-002",
		UserInput: "Create a microservice with authentication, database, and tests",
		Status:    models.IntentStatusCompleted,
		CreatedAt: time.Now(),
	}
	
	taskResults := createSampleTaskResults()
	
	// Generate QuantumDrops
	drops, err := generator.GenerateQuantumDrops(intent, taskResults)
	if err != nil {
		log.Fatalf("Failed to generate QuantumDrops: %v", err)
	}
	
	log.Printf("💧 Generated %d QuantumDrops:", len(drops))
	
	for _, drop := range drops {
		log.Printf("   • %s (%s)", drop.Name, drop.Type)
		log.Printf("     Files: %d, Quality: %d, Security: %d", 
			drop.Metadata.FileCount, drop.Metadata.QualityScore, drop.Metadata.SecurityScore)
		log.Printf("     HITL Required: %v, Status: %s", 
			drop.Metadata.HITLRequired, drop.Status)
		
		// Validate drop has proper file structure
		if len(drop.Files) == 0 {
			log.Printf("     ⚠️  WARNING: Drop has no files")
		} else {
			log.Printf("     ✅ Drop contains files:")
			for path := range drop.Files {
				log.Printf("        - %s", path)
				if len(drop.Files) > 5 { // Don't spam output
					log.Printf("        ... and %d more files", len(drop.Files)-5)
					break
				}
			}
		}
	}
}

func createSampleTaskResults() []packaging.TaskExecutionResult {
	return []packaging.TaskExecutionResult{
		{
			Task: models.Task{
				ID:          "QL-INF-001",
				Type:        models.TaskTypeInfra,
				Description: "Set up Docker configuration",
			},
			Status:        models.TaskStatusCompleted,
			Output:        createInfrastructureOutput(),
			AgentID:       "QLI-AGT-001",
			ExecutionTime: 10 * time.Second,
		},
		{
			Task: models.Task{
				ID:          "QL-DEV-002",
				Type:        models.TaskTypeCodegen,
				Description: "Create main Go application",
			},
			Status:        models.TaskStatusCompleted,
			Output:        createCodegenOutput(),
			AgentID:       "QLD-AGT-002",
			ExecutionTime: 15 * time.Second,
		},
		{
			Task: models.Task{
				ID:          "QL-TST-003",
				Type:        models.TaskTypeTest,
				Description: "Write unit tests",
			},
			Status:        models.TaskStatusCompleted,
			Output:        createTestOutput(),
			AgentID:       "QLT-AGT-003",
			ExecutionTime: 8 * time.Second,
		},
		{
			Task: models.Task{
				ID:          "QL-DOC-004",
				Type:        models.TaskTypeDoc,
				Description: "Generate API documentation",
			},
			Status:        models.TaskStatusCompleted,
			Output:        createDocumentationOutput(),
			AgentID:       "QLC-AGT-004",
			ExecutionTime: 5 * time.Second,
		},
	}
}

func createInfrastructureOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "Dockerfile": "FROM golang:1.21-alpine AS builder\nWORKDIR /app\nCOPY go.mod go.sum ./\nRUN go mod download\nCOPY . .\nRUN go build -o main .\n\nFROM alpine:latest\nRUN apk --no-cache add ca-certificates\nWORKDIR /root/\nCOPY --from=builder /app/main .\nEXPOSE 8080\nCMD [\"./main\"]",
    "docker-compose.yml": "version: '3.8'\nservices:\n  app:\n    build: .\n    ports:\n      - \"8080:8080\"\n    depends_on:\n      - postgres\n  postgres:\n    image: postgres:15\n    environment:\n      POSTGRES_DB: myapp\n      POSTGRES_USER: user\n      POSTGRES_PASSWORD: password\n    ports:\n      - \"5432:5432\"",
    "k8s/deployment.yaml": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: auth-service\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: auth-service\n  template:\n    metadata:\n      labels:\n        app: auth-service\n    spec:\n      containers:\n      - name: auth-service\n        image: auth-service:latest\n        ports:\n        - containerPort: 8080"
  }
}
=== SANDBOX EXECUTION ===
Infrastructure setup completed successfully.`
}

func createCodegenOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "go.mod": "module auth-microservice\n\ngo 1.21\n\nrequire (\n\tgithub.com/golang-jwt/jwt/v5 v5.0.0\n\tgithub.com/gorilla/mux v1.8.0\n\tgithub.com/lib/pq v1.10.9\n)",
    "cmd/main.go": "package main\n\nimport (\n\t\"log\"\n\t\"net/http\"\n\t\"os\"\n\n\t\"auth-microservice/internal/handlers\"\n\t\"github.com/gorilla/mux\"\n)\n\nfunc main() {\n\tr := mux.NewRouter()\n\tr.HandleFunc(\"/health\", handlers.HealthCheck).Methods(\"GET\")\n\tr.HandleFunc(\"/register\", handlers.Register).Methods(\"POST\")\n\tr.HandleFunc(\"/login\", handlers.Login).Methods(\"POST\")\n\n\tport := os.Getenv(\"PORT\")\n\tif port == \"\" {\n\t\tport = \"8080\"\n\t}\n\n\tlog.Printf(\"Server starting on port %s\", port)\n\tlog.Fatal(http.ListenAndServe(\":\"+port, r))\n}",
    "internal/handlers/auth.go": "package handlers\n\nimport (\n\t\"encoding/json\"\n\t\"net/http\"\n\t\"time\"\n\n\t\"github.com/golang-jwt/jwt/v5\"\n)\n\ntype User struct {\n\tID       int    `json:\"id\"`\n\tUsername string `json:\"username\"`\n\tEmail    string `json:\"email\"`\n}\n\ntype LoginRequest struct {\n\tUsername string `json:\"username\"`\n\tPassword string `json:\"password\"`\n}\n\nfunc HealthCheck(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode(map[string]string{\"status\": \"healthy\"})\n}\n\nfunc Register(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tw.WriteHeader(http.StatusCreated)\n\tjson.NewEncoder(w).Encode(map[string]string{\"message\": \"User registered successfully\"})\n}\n\nfunc Login(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\ttoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{\n\t\t\"sub\": \"1234567890\",\n\t\t\"exp\": time.Now().Add(24 * time.Hour).Unix(),\n\t})\n\ttokenString, _ := token.SignedString([]byte(\"secret\"))\n\tjson.NewEncoder(w).Encode(map[string]string{\"token\": tokenString})\n}",
    "internal/middleware/auth.go": "package middleware\n\nimport (\n\t\"net/http\"\n\t\"strings\"\n\n\t\"github.com/golang-jwt/jwt/v5\"\n)\n\nfunc AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {\n\treturn func(w http.ResponseWriter, r *http.Request) {\n\t\tauthHeader := r.Header.Get(\"Authorization\")\n\t\tif authHeader == \"\" {\n\t\t\thttp.Error(w, \"Authorization header required\", http.StatusUnauthorized)\n\t\t\treturn\n\t\t}\n\n\t\ttokenString := strings.TrimPrefix(authHeader, \"Bearer \")\n\t\ttoken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {\n\t\t\treturn []byte(\"secret\"), nil\n\t\t})\n\n\t\tif err != nil || !token.Valid {\n\t\t\thttp.Error(w, \"Invalid token\", http.StatusUnauthorized)\n\t\t\treturn\n\t\t}\n\n\t\tnext(w, r)\n\t}\n}"
  }
}
=== SANDBOX EXECUTION ===
Code generation completed successfully.`
}

func createTestOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "tests/auth_test.go": "package tests\n\nimport (\n\t\"bytes\"\n\t\"encoding/json\"\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"testing\"\n\n\t\"auth-microservice/internal/handlers\"\n)\n\nfunc TestHealthCheck(t *testing.T) {\n\treq, err := http.NewRequest(\"GET\", \"/health\", nil)\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\trr := httptest.NewRecorder()\n\thandler := http.HandlerFunc(handlers.HealthCheck)\n\thandler.ServeHTTP(rr, req)\n\n\tif status := rr.Code; status != http.StatusOK {\n\t\tt.Errorf(\"handler returned wrong status code: got %v want %v\", status, http.StatusOK)\n\t}\n}\n\nfunc TestRegister(t *testing.T) {\n\tuser := map[string]string{\n\t\t\"username\": \"testuser\",\n\t\t\"email\": \"test@example.com\",\n\t\t\"password\": \"password123\",\n\t}\n\n\tjsonData, _ := json.Marshal(user)\n\treq, err := http.NewRequest(\"POST\", \"/register\", bytes.NewBuffer(jsonData))\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\trr := httptest.NewRecorder()\n\thandler := http.HandlerFunc(handlers.Register)\n\thandler.ServeHTTP(rr, req)\n\n\tif status := rr.Code; status != http.StatusCreated {\n\t\tt.Errorf(\"handler returned wrong status code: got %v want %v\", status, http.StatusCreated)\n\t}\n}",
    "tests/integration_test.go": "package tests\n\nimport (\n\t\"testing\"\n\t\"net/http\"\n\t\"net/http/httptest\"\n\t\"bytes\"\n\t\"encoding/json\"\n\n\t\"auth-microservice/internal/handlers\"\n)\n\nfunc TestAuthFlow(t *testing.T) {\n\t// Test complete authentication flow\n\tt.Run(\"Complete Auth Flow\", func(t *testing.T) {\n\t\t// 1. Register user\n\t\tuser := map[string]string{\"username\": \"testuser\", \"password\": \"password123\"}\n\t\tjsonData, _ := json.Marshal(user)\n\t\treq := httptest.NewRequest(\"POST\", \"/register\", bytes.NewBuffer(jsonData))\n\t\tw := httptest.NewRecorder()\n\t\thandlers.Register(w, req)\n\n\t\tif w.Code != http.StatusCreated {\n\t\t\tt.Fatalf(\"Registration failed: %d\", w.Code)\n\t\t}\n\n\t\t// 2. Login user\n\t\treq = httptest.NewRequest(\"POST\", \"/login\", bytes.NewBuffer(jsonData))\n\t\tw = httptest.NewRecorder()\n\t\thandlers.Login(w, req)\n\n\t\tif w.Code != http.StatusOK {\n\t\t\tt.Fatalf(\"Login failed: %d\", w.Code)\n\t\t}\n\t})\n}"
  }
}
=== SANDBOX EXECUTION ===
Tests generated successfully.`
}

func createDocumentationOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "README.md": "# Authentication Microservice\n\nA production-ready Go microservice for user authentication using JWT tokens.\n\n## Features\n\n- User registration and login\n- JWT token-based authentication\n- Health check endpoint\n- PostgreSQL database integration\n- Docker containerization\n- Comprehensive test suite\n\n## API Endpoints\n\n### Health Check\n```\nGET /health\n```\nReturns service health status.\n\n### User Registration\n```\nPOST /register\nContent-Type: application/json\n\n{\n  \"username\": \"string\",\n  \"email\": \"string\",\n  \"password\": \"string\"\n}\n```\n\n### User Login\n```\nPOST /login\nContent-Type: application/json\n\n{\n  \"username\": \"string\",\n  \"password\": \"string\"\n}\n```\n\nReturns JWT token for authentication.\n\n## Quick Start\n\n1. Clone the repository\n2. Run `docker-compose up`\n3. Service will be available at `http://localhost:8080`\n\n## Environment Variables\n\n- `PORT`: Server port (default: 8080)\n- `DB_HOST`: Database host\n- `DB_NAME`: Database name\n- `JWT_SECRET`: JWT signing secret",
    "docs/api.md": "# API Documentation\n\n## Authentication Endpoints\n\n### POST /register\n\nRegisters a new user in the system.\n\n**Request Body:**\n```json\n{\n  \"username\": \"string\",\n  \"email\": \"string\", \n  \"password\": \"string\"\n}\n```\n\n**Response:**\n- Status: 201 Created\n- Body: `{\"message\": \"User registered successfully\"}`\n\n### POST /login\n\nAuthenticates a user and returns a JWT token.\n\n**Request Body:**\n```json\n{\n  \"username\": \"string\",\n  \"password\": \"string\"\n}\n```\n\n**Response:**\n- Status: 200 OK\n- Body: `{\"token\": \"jwt_token_here\"}`\n\n### GET /health\n\nHealth check endpoint.\n\n**Response:**\n- Status: 200 OK\n- Body: `{\"status\": \"healthy\"}`\n\n## Authentication\n\nProtected endpoints require the JWT token in the Authorization header:\n\n```\nAuthorization: Bearer <jwt_token>\n```"
  }
}
=== SANDBOX EXECUTION ===
Documentation generated successfully.`
}

func validateExportedCapsule(capsulePath string) {
	log.Printf("🔍 Validating exported capsule: %s", filepath.Base(capsulePath))
	
	// Read the file
	data, err := os.ReadFile(capsulePath)
	if err != nil {
		log.Printf("❌ Failed to read capsule: %v", err)
		return
	}
	
	log.Printf("   📏 File size: %d bytes (%.2f KB)", len(data), float64(len(data))/1024)
	
	// Parse as ZIP
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Printf("❌ Failed to parse ZIP: %v", err)
		return
	}
	
	log.Printf("   📂 ZIP contains %d files:", len(reader.File))
	
	// Track what we find
	hasManifest := false
	hasMetadata := false
	hasREADME := false
	hasProjectFiles := false
	goFileCount := 0
	yamlFileCount := 0
	
	for _, file := range reader.File {
		log.Printf("      📄 %s (%d bytes)", file.Name, file.UncompressedSize64)
		
		if file.Name == "manifest.json" {
			hasManifest = true
			validateManifestFile(file)
		}
		if file.Name == "metadata.json" {
			hasMetadata = true
			validateMetadataFile(file)
		}
		if strings.Contains(file.Name, "README.md") {
			hasREADME = true
		}
		if strings.HasPrefix(file.Name, "project/") {
			hasProjectFiles = true
		}
		if strings.HasSuffix(file.Name, ".go") {
			goFileCount++
		}
		if strings.HasSuffix(file.Name, ".yaml") || strings.HasSuffix(file.Name, ".yml") {
			yamlFileCount++
		}
	}
	
	// Report validation results
	log.Printf("   📊 Validation Results:")
	log.Printf("      ✅ Manifest: %v", hasManifest)
	log.Printf("      ✅ Metadata: %v", hasMetadata) 
	log.Printf("      ✅ README: %v", hasREADME)
	log.Printf("      ✅ Project Files: %v", hasProjectFiles)
	log.Printf("      📁 Go Files: %d", goFileCount)
	log.Printf("      📁 YAML Files: %d", yamlFileCount)
	
	score := 0
	if hasManifest { score += 20 }
	if hasMetadata { score += 20 }
	if hasREADME { score += 20 }
	if hasProjectFiles { score += 20 }
	if goFileCount > 0 { score += 20 }
	
	log.Printf("   🎯 Capsule Quality Score: %d/100", score)
	
	if score >= 80 {
		log.Printf("   ✅ EXCELLENT capsule quality")
	} else if score >= 60 {
		log.Printf("   ⚠️  GOOD capsule quality")
	} else {
		log.Printf("   ❌ POOR capsule quality")
	}
}

func validateManifestFile(file *zip.File) {
	reader, err := file.Open()
	if err != nil {
		log.Printf("         ❌ Failed to open manifest: %v", err)
		return
	}
	defer reader.Close()
	
	var manifest map[string]interface{}
	if err := json.NewDecoder(reader).Decode(&manifest); err != nil {
		log.Printf("         ❌ Invalid JSON in manifest: %v", err)
		return
	}
	
	log.Printf("         ✅ Valid JSON manifest")
}

func validateMetadataFile(file *zip.File) {
	reader, err := file.Open()
	if err != nil {
		log.Printf("         ❌ Failed to open metadata: %v", err)
		return
	}
	defer reader.Close()
	
	var metadata map[string]interface{}
	if err := json.NewDecoder(reader).Decode(&metadata); err != nil {
		log.Printf("         ❌ Invalid JSON in metadata: %v", err)
		return
	}
	
	// Check for key fields
	if capsuleID, ok := metadata["capsule_id"]; ok {
		log.Printf("         ✅ Capsule ID: %v", capsuleID)
	}
	if version, ok := metadata["version"]; ok {
		log.Printf("         ✅ Version: %v", version)
	}
	if totalTasks, ok := metadata["total_tasks"]; ok {
		log.Printf("         ✅ Total Tasks: %v", totalTasks)
	}
}