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
	log.Println("🎯 SIMPLE CAPSULE GENERATION TEST")
	log.Println("==================================")

	// Test the capsule packaging directly
	testCapsulePackaging()
	
	log.Println("✅ CAPSULE TEST COMPLETED!")
}

func testCapsulePackaging() {
	log.Println("\n📦 Direct Capsule Packaging Test")
	log.Println("---------------------------------")
	
	// Create a capsule packager
	packager := packaging.NewCapsulePackager("./output")
	
	// Create a sample intent
	intent := models.Intent{
		ID:        "test-intent-001",
		UserInput: "Create a simple Go HTTP server",
		Status:    models.IntentStatusCompleted,
		CreatedAt: time.Now(),
	}
	
	// Create sample task execution results
	taskResults := []packaging.TaskExecutionResult{
		{
			Task: models.Task{
				ID:          "QL-INF-001",
				Type:        models.TaskTypeInfra,
				Description: "Set up Docker configuration",
			},
			Status:        models.TaskStatusCompleted,
			Output:        createSampleInfraOutput(),
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
			Output:        createSampleCodeOutput(),
			AgentID:       "QLD-AGT-002",
			ExecutionTime: 15 * time.Second,
		},
	}
	
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
	if err := os.MkdirAll("./output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	
	capsuleFile := filepath.Join("./output", capsule.Metadata.CapsuleID+".qlcapsule")
	if err := os.WriteFile(capsuleFile, zipData, 0644); err != nil {
		log.Fatalf("Failed to save capsule file: %v", err)
	}
	
	log.Printf("💾 Capsule saved: %s (%d bytes)", capsuleFile, len(zipData))
	
	// Validate the exported capsule
	validateCapsule(capsuleFile, zipData)
}

func createSampleInfraOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "Dockerfile": "FROM golang:1.21-alpine\nWORKDIR /app\nCOPY . .\nRUN go build -o main .\nEXPOSE 8080\nCMD [\"./main\"]",
    "docker-compose.yml": "version: '3.8'\nservices:\n  app:\n    build: .\n    ports:\n      - \"8080:8080\""
  }
}
=== SANDBOX EXECUTION ===
Infrastructure setup completed.`
}

func createSampleCodeOutput() string {
	return `=== LLM OUTPUT ===
{
  "files": {
    "go.mod": "module simple-server\n\ngo 1.21",
    "main.go": "package main\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n\t\"log\"\n)\n\nfunc main() {\n\thttp.HandleFunc(\"/health\", healthHandler)\n\tlog.Println(\"Server starting on :8080\")\n\tlog.Fatal(http.ListenAndServe(\":8080\", nil))\n}\n\nfunc healthHandler(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprintf(w, \"OK\")\n}"
  }
}
=== SANDBOX EXECUTION ===
Code generation completed.`
}

func validateCapsule(capsuleFile string, data []byte) {
	log.Printf("🔍 Validating capsule: %s", filepath.Base(capsuleFile))
	log.Printf("   📏 Size: %d bytes (%.2f KB)", len(data), float64(len(data))/1024)
	
	// Parse as ZIP
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Printf("❌ Failed to parse ZIP: %v", err)
		return
	}
	
	log.Printf("   📂 Contents (%d files):", len(reader.File))
	
	// Validation counters
	var (
		hasManifest      = false
		hasMetadata      = false
		hasREADME        = false
		hasProjectFiles  = false
		goFileCount      = 0
		dockerFileCount  = 0
	)
	
	// Examine each file
	for _, file := range reader.File {
		fileName := file.Name
		log.Printf("      📄 %s (%d bytes)", fileName, file.UncompressedSize64)
		
		// Check for key files
		if fileName == "manifest.json" {
			hasManifest = true
			validateJSONFile(file, "manifest")
		}
		if fileName == "metadata.json" {
			hasMetadata = true
			validateJSONFile(file, "metadata")
		}
		if strings.Contains(fileName, "README.md") {
			hasREADME = true
		}
		if strings.HasPrefix(fileName, "project/") {
			hasProjectFiles = true
		}
		
		// Count file types
		if strings.HasSuffix(fileName, ".go") {
			goFileCount++
		}
		if strings.Contains(fileName, "Dockerfile") || strings.Contains(fileName, "docker-compose") {
			dockerFileCount++
		}
	}
	
	// Report validation results
	log.Printf("   📊 Validation Summary:")
	log.Printf("      ✅ Manifest JSON: %v", hasManifest)
	log.Printf("      ✅ Metadata JSON: %v", hasMetadata)
	log.Printf("      ✅ README file: %v", hasREADME)
	log.Printf("      ✅ Project files: %v", hasProjectFiles)
	log.Printf("      📁 Go files: %d", goFileCount)
	log.Printf("      🐳 Docker files: %d", dockerFileCount)
	
	// Calculate quality score
	score := 0
	if hasManifest { score += 20 }
	if hasMetadata { score += 20 }
	if hasREADME { score += 20 }
	if hasProjectFiles { score += 20 }
	if goFileCount > 0 { score += 20 }
	
	log.Printf("   🎯 Capsule Quality Score: %d/100", score)
	
	if score >= 80 {
		log.Printf("   ✅ EXCELLENT - Capsule is production ready!")
	} else if score >= 60 {
		log.Printf("   ⚠️  GOOD - Capsule meets basic requirements")
	} else {
		log.Printf("   ❌ POOR - Capsule needs improvement")
	}
}

func validateJSONFile(file *zip.File, fileType string) {
	reader, err := file.Open()
	if err != nil {
		log.Printf("         ❌ Failed to open %s: %v", fileType, err)
		return
	}
	defer reader.Close()
	
	// Try to parse as JSON
	var jsonData interface{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&jsonData); err != nil {
		log.Printf("         ❌ Invalid JSON in %s: %v", fileType, err)
		return
	}
	
	log.Printf("         ✅ Valid %s JSON", fileType)
	
	// For metadata, check specific fields
	if fileType == "metadata" {
		if metadata, ok := jsonData.(map[string]interface{}); ok {
			if capsuleID, exists := metadata["capsule_id"]; exists {
				log.Printf("         📝 Capsule ID: %v", capsuleID)
			}
			if totalTasks, exists := metadata["total_tasks"]; exists {
				log.Printf("         📝 Total Tasks: %v", totalTasks)
			}
		}
	}
}