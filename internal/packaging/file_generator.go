package packaging

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// ProjectStructure represents the JSON structure returned by agents
type ProjectStructure struct {
	ProjectStructure struct {
		ProjectName string `json:"project_name"`
		ProjectType string `json:"project_type"`
		Files       []File `json:"files"`
	} `json:"project_structure"`
}

// File represents a single file in the project structure
type File struct {
	Path    string `json:"path"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

// FileGenerator handles parsing LLM output and generating proper file structures
type FileGenerator struct{}

func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

// ParseLLMOutput attempts to parse the LLM output as structured JSON
// If it fails, it falls back to treating the content as a single file
func (fg *FileGenerator) ParseLLMOutput(taskID, taskType, llmOutput string) (*ProjectStructure, error) {
	// First, try to parse as JSON structure
	var projectStruct ProjectStructure
	
	// Clean the output - remove markdown code blocks if present
	cleanedOutput := fg.cleanMarkdownCodeBlocks(llmOutput)
	
	err := json.Unmarshal([]byte(cleanedOutput), &projectStruct)
	if err == nil && len(projectStruct.ProjectStructure.Files) > 0 {
		return &projectStruct, nil
	}
	
	// Fallback: treat as single file based on task type
	return fg.createFallbackStructure(taskID, taskType, llmOutput), nil
}

// cleanMarkdownCodeBlocks removes ```json and ``` markers if present
func (fg *FileGenerator) cleanMarkdownCodeBlocks(content string) string {
	// Remove ```json prefix
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
	}
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
	}
	
	// Remove ``` suffix
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(content, "```")
	}
	
	return strings.TrimSpace(content)
}

// createFallbackStructure creates a basic structure when JSON parsing fails
func (fg *FileGenerator) createFallbackStructure(taskID, taskType, content string) *ProjectStructure {
	projectStruct := &ProjectStructure{}
	
	// Determine file extension and name based on task type
	fileName, fileType := fg.determineFileInfo(taskType, content)
	
	projectStruct.ProjectStructure.ProjectName = fmt.Sprintf("task-%s", taskID)
	projectStruct.ProjectStructure.ProjectType = taskType
	projectStruct.ProjectStructure.Files = []File{
		{
			Path:    fileName,
			Type:    fileType,
			Content: content,
		},
	}
	
	return projectStruct
}

// determineFileInfo intelligently determines file name and type based on content and task type
func (fg *FileGenerator) determineFileInfo(taskType, content string) (string, string) {
	switch taskType {
	case "codegen":
		if strings.Contains(content, "package main") || strings.Contains(content, "func main") {
			return "main.go", "go"
		}
		if strings.Contains(content, "import (") || strings.Contains(content, "package ") {
			return "code.go", "go"
		}
		if strings.Contains(content, "def ") || strings.Contains(content, "import ") {
			return "code.py", "python"
		}
		if strings.Contains(content, "function ") || strings.Contains(content, "const ") {
			return "code.js", "javascript"
		}
		return "code.txt", "text"
		
	case "infra":
		if strings.Contains(content, "apiVersion:") || strings.Contains(content, "kind:") {
			return "deployment.yaml", "yaml"
		}
		if strings.Contains(content, "resource ") || strings.Contains(content, "provider ") {
			return "main.tf", "terraform"
		}
		if strings.Contains(content, "FROM ") || strings.Contains(content, "RUN ") {
			return "Dockerfile", "dockerfile"
		}
		if strings.Contains(content, "version:") && strings.Contains(content, "services:") {
			return "docker-compose.yml", "yaml"
		}
		return "infrastructure.yaml", "yaml"
		
	case "test":
		if strings.Contains(content, "func Test") || strings.Contains(content, "package ") {
			return "main_test.go", "go"
		}
		if strings.Contains(content, "def test_") || strings.Contains(content, "import pytest") {
			return "test_main.py", "python"
		}
		if strings.Contains(content, "describe(") || strings.Contains(content, "it(") {
			return "test.spec.js", "javascript"
		}
		return "test.txt", "text"
		
	case "doc":
		if strings.Contains(content, "# ") || strings.Contains(content, "## ") {
			return "README.md", "markdown"
		}
		return "documentation.md", "markdown"
		
	case "analyze":
		if strings.Contains(content, "# ") || strings.Contains(content, "## ") {
			return "analysis_report.md", "markdown"
		}
		return "analysis.txt", "text"
		
	default:
		return "output.txt", "text"
	}
}

// GenerateFileStructure creates the final file structure for packaging
func (fg *FileGenerator) GenerateFileStructure(projectStruct *ProjectStructure) map[string]string {
	fileMap := make(map[string]string)
	
	for _, file := range projectStruct.ProjectStructure.Files {
		// Ensure proper directory structure
		cleanPath := filepath.Clean(file.Path)
		fileMap[cleanPath] = file.Content
	}
	
	// Add a project manifest if not present
	if _, exists := fileMap["project.json"]; !exists {
		manifest := map[string]interface{}{
			"name":    projectStruct.ProjectStructure.ProjectName,
			"type":    projectStruct.ProjectStructure.ProjectType,
			"files":   len(projectStruct.ProjectStructure.Files),
			"version": "1.0.0",
		}
		
		manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
		fileMap["project.json"] = string(manifestJSON)
	}
	
	return fileMap
}

// GetFileExtension returns the appropriate file extension for a file type
func (fg *FileGenerator) GetFileExtension(fileType string) string {
	extensions := map[string]string{
		"go":         ".go",
		"python":     ".py",
		"javascript": ".js",
		"typescript": ".ts",
		"yaml":       ".yaml",
		"yml":        ".yml",
		"json":       ".json",
		"markdown":   ".md",
		"text":       ".txt",
		"dockerfile": "",
		"terraform":  ".tf",
		"html":       ".html",
		"css":        ".css",
		"sql":        ".sql",
		"xml":        ".xml",
		"toml":       ".toml",
		"ini":        ".ini",
		"conf":       ".conf",
		"svg":        ".svg",
		"csv":        ".csv",
	}
	
	if ext, exists := extensions[fileType]; exists {
		return ext
	}
	
	return ".txt"
}