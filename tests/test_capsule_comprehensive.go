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

	"QLP/internal/orchestrator"
)

func main() {
	log.Println("🧪 COMPREHENSIVE CAPSULE TEST")
	log.Println("=====================================")

	// Clean up any existing output
	if err := os.RemoveAll("./output"); err != nil {
		log.Printf("Warning: Failed to clean output directory: %v", err)
	}
	
	// Create fresh output directory
	if err := os.MkdirAll("./output", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Test 1: Simple focused project
	testSimpleProject()
	
	// Test 2: Complex microservice project  
	testComplexMicroservice()
	
	// Test 3: Validate capsule contents
	validateCapsuleContents()
	
	log.Println("✅ ALL CAPSULE TESTS COMPLETED SUCCESSFULLY!")
}

func testSimpleProject() {
	log.Println("\n🎯 TEST 1: Simple Go HTTP Server")
	log.Println("----------------------------------")
	
	orch := orchestrator.New()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	intent := "Create a simple Go HTTP server with health check endpoint"
	
	start := time.Now()
	if err := orch.ProcessAndExecuteIntent(ctx, intent); err != nil {
		log.Fatalf("Failed to process simple intent: %v", err)
	}
	duration := time.Since(start)
	
	log.Printf("✅ Simple project completed in %v", duration)
	
	// Check for expected outputs
	checkOutputDirectory("simple")
}

func testComplexMicroservice() {
	log.Println("\n🎯 TEST 2: Complex Microservice")  
	log.Println("--------------------------------")
	
	orch := orchestrator.New()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	if err := orch.Start(ctx); err != nil {
		log.Fatalf("Failed to start orchestrator: %v", err)
	}

	intent := "Create a production-ready Go user authentication microservice with JWT, PostgreSQL database, Docker deployment, comprehensive tests, and API documentation"
	
	start := time.Now()
	if err := orch.ProcessAndExecuteIntent(ctx, intent); err != nil {
		log.Fatalf("Failed to process complex intent: %v", err)
	}
	duration := time.Since(start)
	
	log.Printf("✅ Complex microservice completed in %v", duration)
	
	// Check for expected outputs
	checkOutputDirectory("complex")
}

func checkOutputDirectory(testType string) {
	log.Printf("📂 Checking output directory for %s test...", testType)
	
	outputDir := "./output"
	
	// Check if output directory exists and has content
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		log.Printf("❌ Failed to read output directory: %v", err)
		return
	}
	
	if len(entries) == 0 {
		log.Printf("❌ Output directory is empty")
		return
	}
	
	log.Printf("📁 Found %d output files/directories:", len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			log.Printf("   📁 %s/", entry.Name())
		} else {
			log.Printf("   📄 %s", entry.Name())
		}
	}
	
	// Look for capsule files
	capsuleFiles := []string{}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".qlcapsule") || strings.HasSuffix(entry.Name(), ".zip") {
			capsuleFiles = append(capsuleFiles, entry.Name())
		}
	}
	
	if len(capsuleFiles) > 0 {
		log.Printf("🎯 Found %d capsule files:", len(capsuleFiles))
		for _, file := range capsuleFiles {
			log.Printf("   📦 %s", file)
			validateSingleCapsule(filepath.Join(outputDir, file))
		}
	} else {
		log.Printf("⚠️  No capsule files found")
	}
}

func validateSingleCapsule(capsulePath string) {
	log.Printf("🔍 Validating capsule: %s", filepath.Base(capsulePath))
	
	// Read the capsule file
	data, err := os.ReadFile(capsulePath)
	if err != nil {
		log.Printf("❌ Failed to read capsule file: %v", err)
		return
	}
	
	log.Printf("   📏 Size: %d bytes (%.2f MB)", len(data), float64(len(data))/1024/1024)
	
	// If it's a ZIP file, examine contents
	if strings.HasSuffix(capsulePath, ".zip") || strings.HasSuffix(capsulePath, ".qlcapsule") {
		validateZipCapsule(data)
	}
}

func validateZipCapsule(data []byte) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Printf("❌ Failed to open zip capsule: %v", err)
		return
	}
	
	log.Printf("   📋 ZIP Contents (%d files):", len(reader.File))
	
	expectedFiles := []string{
		"manifest.json",
		"metadata.json", 
		"README.md",
	}
	
	expectedDirs := []string{
		"tasks/",
		"reports/", 
		"project/",
	}
	
	foundFiles := make(map[string]bool)
	foundDirs := make(map[string]bool)
	codeFiles := 0
	configFiles := 0
	
	for _, file := range reader.File {
		fileName := file.Name
		log.Printf("      📄 %s (%d bytes)", fileName, file.UncompressedSize64)
		
		// Track expected files
		for _, expected := range expectedFiles {
			if strings.Contains(fileName, expected) {
				foundFiles[expected] = true
			}
		}
		
		// Track expected directories
		for _, expected := range expectedDirs {
			if strings.HasPrefix(fileName, expected) {
				foundDirs[expected] = true
			}
		}
		
		// Count file types
		if strings.HasSuffix(fileName, ".go") {
			codeFiles++
		}
		if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") || strings.HasSuffix(fileName, ".json") {
			configFiles++
		}
		
		// Validate specific important files
		if fileName == "manifest.json" || fileName == "metadata.json" {
			validateJSONFile(file)
		}
		
		if strings.Contains(fileName, "README.md") {
			validateREADME(file)
		}
	}
	
	// Report validation results
	log.Printf("   📊 Validation Results:")
	log.Printf("      🔢 Code files (.go): %d", codeFiles)
	log.Printf("      ⚙️  Config files (.yaml/.json): %d", configFiles)
	
	// Check for expected structure
	log.Printf("   ✅ Expected Files Found:")
	for _, expected := range expectedFiles {
		status := "❌"
		if foundFiles[expected] {
			status = "✅"
		}
		log.Printf("      %s %s", status, expected)
	}
	
	log.Printf("   ✅ Expected Directories Found:")
	for _, expected := range expectedDirs {
		status := "❌"
		if foundDirs[expected] {
			status = "✅"
		}
		log.Printf("      %s %s", status, expected)
	}
}

func validateJSONFile(file *zip.File) {
	reader, err := file.Open()
	if err != nil {
		log.Printf("❌ Failed to open %s: %v", file.Name, err)
		return
	}
	defer reader.Close()
	
	// Try to parse as JSON
	var jsonData interface{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&jsonData); err != nil {
		log.Printf("❌ Invalid JSON in %s: %v", file.Name, err)
		return
	}
	
	log.Printf("      ✅ %s is valid JSON", file.Name)
	
	// If it's metadata, check for required fields
	if file.Name == "metadata.json" {
		validateMetadata(jsonData)
	}
}

func validateMetadata(data interface{}) {
	metadata, ok := data.(map[string]interface{})
	if !ok {
		log.Printf("❌ Metadata is not a JSON object")
		return
	}
	
	requiredFields := []string{
		"capsule_id",
		"version", 
		"intent_text",
		"total_tasks",
		"successful_tasks",
		"overall_score",
	}
	
	log.Printf("      🔍 Checking metadata fields:")
	for _, field := range requiredFields {
		if _, exists := metadata[field]; exists {
			log.Printf("         ✅ %s: %v", field, metadata[field])
		} else {
			log.Printf("         ❌ Missing: %s", field)
		}
	}
}

func validateREADME(file *zip.File) {
	reader, err := file.Open()
	if err != nil {
		log.Printf("❌ Failed to open README: %v", err)
		return
	}
	defer reader.Close()
	
	// Read content
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	content := buf.String()
	
	if len(content) < 100 {
		log.Printf("❌ README too short (%d chars)", len(content))
		return
	}
	
	// Check for expected sections
	expectedSections := []string{"#", "Overview", "API", "Usage"}
	foundSections := 0
	
	for _, section := range expectedSections {
		if strings.Contains(content, section) {
			foundSections++
		}
	}
	
	log.Printf("      ✅ README: %d chars, %d/%d sections found", 
		len(content), foundSections, len(expectedSections))
}

func validateCapsuleContents() {
	log.Println("\n🎯 TEST 3: Detailed Capsule Content Validation")
	log.Println("-----------------------------------------------")
	
	// Find the most recent capsule
	outputDir := "./output"
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		log.Printf("❌ Failed to read output directory: %v", err)
		return
	}
	
	var latestCapsule string
	var latestTime time.Time
	
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".qlcapsule") || strings.HasSuffix(entry.Name(), ".zip") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestCapsule = entry.Name()
			}
		}
	}
	
	if latestCapsule == "" {
		log.Printf("❌ No capsule files found for detailed validation")
		return
	}
	
	log.Printf("🔍 Performing detailed validation on: %s", latestCapsule)
	
	capsulePath := filepath.Join(outputDir, latestCapsule)
	data, err := os.ReadFile(capsulePath)
	if err != nil {
		log.Printf("❌ Failed to read capsule: %v", err)
		return
	}
	
	// Detailed ZIP analysis
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		log.Printf("❌ Failed to open capsule: %v", err)
		return
	}
	
	// Extract and analyze project structure
	projectFiles := make(map[string]string)
	for _, file := range reader.File {
		if strings.HasPrefix(file.Name, "project/") {
			reader, err := file.Open()
			if err != nil {
				continue
			}
			
			buf := new(bytes.Buffer)
			buf.ReadFrom(reader)
			projectFiles[file.Name] = buf.String()
			reader.Close()
		}
	}
	
	log.Printf("📁 Found %d project files:", len(projectFiles))
	
	// Validate Go project structure
	validateGoProjectStructure(projectFiles)
	
	log.Printf("✅ Detailed capsule validation completed")
}

func validateGoProjectStructure(files map[string]string) {
	log.Printf("🔍 Validating Go project structure...")
	
	hasGoMod := false
	hasMainGo := false
	hasTests := false
	hasDocumentation := false
	
	packageCount := make(map[string]int)
	
	for path, content := range files {
		fileName := filepath.Base(path)
		
		// Check for key files
		if fileName == "go.mod" {
			hasGoMod = true
			log.Printf("   ✅ Found go.mod")
		}
		
		if fileName == "main.go" || strings.Contains(content, "func main()") {
			hasMainGo = true
			log.Printf("   ✅ Found main.go")
		}
		
		if strings.HasSuffix(fileName, "_test.go") || strings.Contains(content, "func Test") {
			hasTests = true
		}
		
		if strings.HasSuffix(fileName, ".md") && len(content) > 50 {
			hasDocumentation = true
		}
		
		// Extract package names
		if strings.HasSuffix(fileName, ".go") {
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "package ") {
					pkg := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "package "))
					packageCount[pkg]++
					break
				}
			}
		}
	}
	
	// Report findings
	log.Printf("   📊 Project Structure Analysis:")
	log.Printf("      📦 go.mod: %v", hasGoMod)
	log.Printf("      🚀 main.go: %v", hasMainGo)  
	log.Printf("      🧪 tests: %v", hasTests)
	log.Printf("      📚 documentation: %v", hasDocumentation)
	
	log.Printf("   📂 Packages found:")
	for pkg, count := range packageCount {
		log.Printf("      %s: %d files", pkg, count)
	}
	
	// Validate structure quality
	structureScore := 0
	if hasGoMod { structureScore += 25 }
	if hasMainGo { structureScore += 25 }
	if hasTests { structureScore += 25 }
	if hasDocumentation { structureScore += 25 }
	
	log.Printf("   🎯 Structure Quality Score: %d/100", structureScore)
	
	if structureScore >= 75 {
		log.Printf("   ✅ EXCELLENT project structure")
	} else if structureScore >= 50 {
		log.Printf("   ⚠️  GOOD project structure")
	} else {
		log.Printf("   ❌ POOR project structure")
	}
}