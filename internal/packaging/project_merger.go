package packaging

import (
	"fmt"
	"path/filepath"
	"strings"

	"QLP/internal/models"
)

// ProjectMerger handles merging multiple task outputs into a single coherent project
type ProjectMerger struct {
	fileGenerator *FileGenerator
}

func NewProjectMerger() *ProjectMerger {
	return &ProjectMerger{
		fileGenerator: NewFileGenerator(),
	}
}

// MergeTasksIntoProject combines all task outputs into a single unified project structure
func (pm *ProjectMerger) MergeTasksIntoProject(intent models.Intent, taskResults []TaskExecutionResult) (*UnifiedProject, error) {
	// Determine overall project type and name from intent
	projectName := pm.generateProjectName(intent.UserInput)
	projectType := pm.determineProjectType(taskResults)
	
	unifiedProject := &UnifiedProject{
		Name:        projectName,
		Type:        projectType,
		Description: intent.UserInput,
		Files:       make(map[string]string),
		Structure:   make(map[string][]string),
	}
	
	// Collect all files from all tasks
	allFiles := make(map[string]string)
	
	for _, taskResult := range taskResults {
		if taskResult.Output == "" {
			continue
		}
		
		// Extract LLM output
		llmOutput := pm.extractLLMOutput(taskResult.Output)
		
		// Parse task output
		projectStruct, err := pm.fileGenerator.ParseLLMOutput(taskResult.Task.ID, string(taskResult.Task.Type), llmOutput)
		if err != nil {
			continue // Skip tasks that can't be parsed
		}
		
		// Get files from this task
		taskFiles := pm.fileGenerator.GenerateFileStructure(projectStruct)
		
		// Merge files intelligently based on task type
		pm.mergeTaskFiles(taskResult.Task, taskFiles, allFiles)
	}
	
	// Organize files into proper project structure
	unifiedProject.Files = pm.organizeProjectFiles(allFiles, projectType)
	unifiedProject.Structure = pm.generateProjectStructure(unifiedProject.Files)
	
	return unifiedProject, nil
}

// mergeTaskFiles intelligently merges files from different tasks
func (pm *ProjectMerger) mergeTaskFiles(task models.Task, taskFiles map[string]string, allFiles map[string]string) {
	switch task.Type {
	case models.TaskTypeCodegen:
		// Merge Go code files
		for path, content := range taskFiles {
			if strings.HasSuffix(path, ".go") {
				// Organize Go files into proper structure
				newPath := pm.organizeGoFile(path, content)
				allFiles[newPath] = content
			} else {
				allFiles[path] = content
			}
		}
		
	case models.TaskTypeInfra:
		// Merge infrastructure files at root level
		for path, content := range taskFiles {
			if strings.Contains(path, "Dockerfile") || strings.Contains(path, "docker-compose") {
				allFiles[filepath.Base(path)] = content
			} else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				allFiles[path] = content
			} else {
				allFiles[path] = content
			}
		}
		
	case models.TaskTypeTest:
		// Put all tests in tests/ directory
		for path, content := range taskFiles {
			if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, ".test.go") {
				newPath := fmt.Sprintf("tests/%s", filepath.Base(path))
				allFiles[newPath] = content
			} else {
				newPath := fmt.Sprintf("tests/%s", path)
				allFiles[newPath] = content
			}
		}
		
	case models.TaskTypeDoc:
		// Merge documentation
		for path, content := range taskFiles {
			if strings.Contains(path, "README") {
				// Append to main README or create if doesn't exist
				if existing, exists := allFiles["README.md"]; exists {
					allFiles["README.md"] = existing + "\n\n" + content
				} else {
					allFiles["README.md"] = content
				}
			} else if strings.Contains(path, "docs/") {
				allFiles[path] = content
			} else {
				allFiles[fmt.Sprintf("docs/%s", path)] = content
			}
		}
		
	default:
		// Default: add files as-is
		for path, content := range taskFiles {
			allFiles[path] = content
		}
	}
}

// organizeGoFile determines the proper location for a Go source file
func (pm *ProjectMerger) organizeGoFile(originalPath, content string) string {
	// Analyze content to determine proper location
	if strings.Contains(content, "func main()") {
		return "cmd/main.go"
	}
	
	if strings.Contains(content, "func Test") || strings.Contains(content, "testing") {
		return fmt.Sprintf("tests/%s", filepath.Base(originalPath))
	}
	
	// Extract package name from content
	lines := strings.Split(content, "\n")
	packageName := "main"
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			packageName = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "package "))
			break
		}
	}
	
	// Organize by package name
	switch packageName {
	case "main":
		return "cmd/main.go"
	case "handlers", "handler":
		return "internal/handlers/" + filepath.Base(originalPath)
	case "auth", "authentication":
		return "internal/auth/" + filepath.Base(originalPath)
	case "middleware":
		return "internal/middleware/" + filepath.Base(originalPath)
	case "models", "model":
		return "internal/models/" + filepath.Base(originalPath)
	case "user", "users":
		return "internal/users/" + filepath.Base(originalPath)
	case "api":
		return "internal/api/" + filepath.Base(originalPath)
	default:
		return fmt.Sprintf("internal/%s/%s", packageName, filepath.Base(originalPath))
	}
}

// organizeProjectFiles creates a clean, organized file structure
func (pm *ProjectMerger) organizeProjectFiles(files map[string]string, projectType string) map[string]string {
	organized := make(map[string]string)
	
	// Ensure go.mod exists for Go projects
	if projectType == "go-api" || projectType == "go-microservice" {
		if _, exists := files["go.mod"]; !exists {
			// Create default go.mod
			projectName := pm.extractProjectNameFromFiles(files)
			organized["go.mod"] = fmt.Sprintf("module %s\n\ngo 1.21\n", projectName)
		}
	}
	
	// Copy all files to organized structure
	for path, content := range files {
		// Clean up paths
		cleanPath := filepath.Clean(path)
		organized[cleanPath] = content
	}
	
	return organized
}

// generateProjectStructure creates a directory structure map
func (pm *ProjectMerger) generateProjectStructure(files map[string]string) map[string][]string {
	structure := make(map[string][]string)
	
	for path := range files {
		dir := filepath.Dir(path)
		if dir == "." {
			dir = "/"
		}
		filename := filepath.Base(path)
		structure[dir] = append(structure[dir], filename)
	}
	
	return structure
}

// Helper functions
func (pm *ProjectMerger) generateProjectName(userInput string) string {
	// Extract project name from user input
	input := strings.ToLower(userInput)
	
	if strings.Contains(input, "microservice") {
		if strings.Contains(input, "auth") {
			return "auth-microservice"
		}
		if strings.Contains(input, "user") {
			return "user-microservice"
		}
		return "api-microservice"
	}
	
	if strings.Contains(input, "api") {
		if strings.Contains(input, "user") {
			return "user-api"
		}
		if strings.Contains(input, "auth") {
			return "auth-api"
		}
		return "rest-api"
	}
	
	return "project"
}

func (pm *ProjectMerger) determineProjectType(taskResults []TaskExecutionResult) string {
	hasGo := false
	hasDocker := false
	hasAPI := false
	
	for _, result := range taskResults {
		if strings.Contains(result.Output, "package main") || strings.Contains(result.Output, "go.mod") {
			hasGo = true
		}
		if strings.Contains(result.Output, "Dockerfile") || strings.Contains(result.Output, "docker") {
			hasDocker = true
		}
		if strings.Contains(result.Output, "http") || strings.Contains(result.Output, "API") || strings.Contains(result.Output, "endpoint") {
			hasAPI = true
		}
	}
	
	if hasGo && hasAPI && hasDocker {
		return "go-microservice"
	} else if hasGo && hasAPI {
		return "go-api"
	} else if hasGo {
		return "go-application"
	}
	
	return "application"
}

func (pm *ProjectMerger) extractProjectNameFromFiles(files map[string]string) string {
	// Try to extract project name from go.mod
	if goMod, exists := files["go.mod"]; exists {
		lines := strings.Split(goMod, "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "module ") {
				return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "module "))
			}
		}
	}
	
	return "project"
}

func (pm *ProjectMerger) extractLLMOutput(agentOutput string) string {
	// Same as in capsule.go
	lines := strings.Split(agentOutput, "\n")
	var llmOutput []string
	inLLMSection := false
	
	for _, line := range lines {
		if strings.Contains(line, "=== LLM OUTPUT ===") {
			inLLMSection = true
			continue
		}
		if strings.Contains(line, "=== SANDBOX EXECUTION ===") {
			break
		}
		if inLLMSection {
			llmOutput = append(llmOutput, line)
		}
	}
	
	return strings.Join(llmOutput, "\n")
}

// UnifiedProject represents a single coherent project structure
type UnifiedProject struct {
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Files       map[string]string   `json:"files"`
	Structure   map[string][]string `json:"structure"`
}