package packaging

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"QLP/internal/models"
)

// QuantumDrop represents a specialized, categorized output that can be reviewed independently
type QuantumDrop struct {
	ID          string                 `json:"id"`
	Type        DropType               `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Files       map[string]string      `json:"files"`
	Structure   map[string][]string    `json:"structure"`
	Metadata    DropMetadata           `json:"metadata"`
	Status      DropStatus             `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	Tasks       []string               `json:"tasks"` // Task IDs that contributed to this drop
}

type DropType string

const (
	DropTypeInfrastructure DropType = "infrastructure"
	DropTypeCodebase       DropType = "codebase"
	DropTypeDocumentation  DropType = "documentation"
	DropTypeTesting        DropType = "testing"
	DropTypeAnalysis       DropType = "analysis"
)

type DropStatus string

const (
	DropStatusPending     DropStatus = "pending"
	DropStatusReady       DropStatus = "ready"
	DropStatusApproved    DropStatus = "approved"
	DropStatusRejected    DropStatus = "rejected"
	DropStatusModified    DropStatus = "modified"
)

type DropMetadata struct {
	FileCount       int               `json:"file_count"`
	TotalLines      int               `json:"total_lines"`
	Technologies    []string          `json:"technologies"`
	Dependencies    []string          `json:"dependencies"`
	QualityScore    int               `json:"quality_score"`
	SecurityScore   int               `json:"security_score"`
	ValidationPassed bool             `json:"validation_passed"`
	HITLRequired    bool              `json:"hitl_required"`
	ReviewNotes     []string          `json:"review_notes,omitempty"`
}

// HITLDecision represents human feedback on a QuantumDrop
type HITLDecision struct {
	DropID      string                 `json:"drop_id"`
	Decision    HITLAction             `json:"decision"`
	Feedback    string                 `json:"feedback,omitempty"`
	Changes     map[string]string      `json:"changes,omitempty"` // file_path -> new_content
	Timestamp   time.Time              `json:"timestamp"`
}

type HITLAction string

const (
	HITLActionContinue HITLAction = "continue"
	HITLActionRedo     HITLAction = "redo"
	HITLActionModify   HITLAction = "modify"
	HITLActionReject   HITLAction = "reject"
)

// QuantumDropGenerator creates specialized QuantumDrops from task results
type QuantumDropGenerator struct {
	fileGenerator *FileGenerator
}

func NewQuantumDropGenerator() *QuantumDropGenerator {
	return &QuantumDropGenerator{
		fileGenerator: NewFileGenerator(),
	}
}

// GenerateQuantumDrops creates categorized drops from task results
func (qdg *QuantumDropGenerator) GenerateQuantumDrops(intent models.Intent, taskResults []TaskExecutionResult) ([]QuantumDrop, error) {
	log.Printf("Generating QuantumDrops from %d task results", len(taskResults))
	
	// Group tasks by type
	taskGroups := qdg.groupTasksByType(taskResults)
	
	var drops []QuantumDrop
	
	// Generate Infrastructure Drop
	if infraTasks, exists := taskGroups[models.TaskTypeInfra]; exists {
		drop, err := qdg.generateInfrastructureDrop(intent, infraTasks)
		if err == nil {
			drops = append(drops, *drop)
		}
	}
	
	// Generate Codebase Drop
	if codeTasks, exists := taskGroups[models.TaskTypeCodegen]; exists {
		drop, err := qdg.generateCodebaseDrop(intent, codeTasks)
		if err == nil {
			drops = append(drops, *drop)
		}
	}
	
	// Generate Documentation Drop
	if docTasks, exists := taskGroups[models.TaskTypeDoc]; exists {
		drop, err := qdg.generateDocumentationDrop(intent, docTasks)
		if err == nil {
			drops = append(drops, *drop)
		}
	}
	
	// Generate Testing Drop
	if testTasks, exists := taskGroups[models.TaskTypeTest]; exists {
		drop, err := qdg.generateTestingDrop(intent, testTasks)
		if err == nil {
			drops = append(drops, *drop)
		}
	}
	
	// Generate Analysis Drop
	if analysisTasks, exists := taskGroups[models.TaskTypeAnalyze]; exists {
		drop, err := qdg.generateAnalysisDrop(intent, analysisTasks)
		if err == nil {
			drops = append(drops, *drop)
		}
	}
	
	log.Printf("Generated %d QuantumDrops", len(drops))
	return drops, nil
}

func (qdg *QuantumDropGenerator) groupTasksByType(taskResults []TaskExecutionResult) map[models.TaskType][]TaskExecutionResult {
	groups := make(map[models.TaskType][]TaskExecutionResult)
	
	for _, result := range taskResults {
		if result.Output != "" && result.Status == models.TaskStatusCompleted {
			groups[result.Task.Type] = append(groups[result.Task.Type], result)
		}
	}
	
	return groups
}

func (qdg *QuantumDropGenerator) generateInfrastructureDrop(intent models.Intent, tasks []TaskExecutionResult) (*QuantumDrop, error) {
	drop := &QuantumDrop{
		ID:          fmt.Sprintf("QD-INFRA-%d", time.Now().Unix()),
		Type:        DropTypeInfrastructure,
		Name:        "Infrastructure Configuration",
		Description: "Docker, Kubernetes, and deployment configurations",
		Files:       make(map[string]string),
		Structure:   make(map[string][]string),
		CreatedAt:   time.Now(),
		Status:      DropStatusReady,
	}
	
	var taskIDs []string
	technologies := make(map[string]bool)
	
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.Task.ID)
		
		// Extract LLM output and parse
		llmOutput := qdg.extractLLMOutput(task.Output)
		projectStruct, err := qdg.fileGenerator.ParseLLMOutput(task.Task.ID, string(task.Task.Type), llmOutput)
		if err != nil {
			continue
		}
		
		// Get files from this task
		taskFiles := qdg.fileGenerator.GenerateFileStructure(projectStruct)
		
		// Merge infrastructure files
		for path, content := range taskFiles {
			if qdg.isInfrastructureFile(path) {
				drop.Files[path] = content
				
				// Extract technologies
				if strings.Contains(path, "docker") || strings.Contains(content, "FROM ") {
					technologies["Docker"] = true
				}
				if strings.Contains(path, "kubernetes") || strings.Contains(content, "apiVersion:") {
					technologies["Kubernetes"] = true
				}
				if strings.Contains(path, "terraform") || strings.Contains(content, "resource ") {
					technologies["Terraform"] = true
				}
			}
		}
	}
	
	drop.Tasks = taskIDs
	drop.Metadata = DropMetadata{
		FileCount:       len(drop.Files),
		TotalLines:      qdg.countTotalLines(drop.Files),
		Technologies:    qdg.mapKeysToSlice(technologies),
		QualityScore:    qdg.calculateQualityScore(tasks),
		SecurityScore:   qdg.calculateSecurityScore(tasks),
		ValidationPassed: qdg.checkValidationPassed(tasks),
		HITLRequired:    len(drop.Files) > 3, // Require HITL for complex infrastructure
	}
	
	drop.Structure = qdg.generateDropStructure(drop.Files)
	
	return drop, nil
}

func (qdg *QuantumDropGenerator) generateCodebaseDrop(intent models.Intent, tasks []TaskExecutionResult) (*QuantumDrop, error) {
	drop := &QuantumDrop{
		ID:          fmt.Sprintf("QD-CODE-%d", time.Now().Unix()),
		Type:        DropTypeCodebase,
		Name:        "Application Codebase",
		Description: "Complete source code with proper project structure",
		Files:       make(map[string]string),
		Structure:   make(map[string][]string),
		CreatedAt:   time.Now(),
		Status:      DropStatusReady,
	}
	
	var taskIDs []string
	technologies := make(map[string]bool)
	dependencies := make(map[string]bool)
	
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.Task.ID)
		
		// Extract and merge code files
		llmOutput := qdg.extractLLMOutput(task.Output)
		projectStruct, err := qdg.fileGenerator.ParseLLMOutput(task.Task.ID, string(task.Task.Type), llmOutput)
		if err != nil {
			continue
		}
		
		taskFiles := qdg.fileGenerator.GenerateFileStructure(projectStruct)
		
		for path, content := range taskFiles {
			// Organize Go files into proper structure
			if strings.HasSuffix(path, ".go") {
				newPath := qdg.organizeCodeFile(path, content)
				drop.Files[newPath] = content
				technologies["Go"] = true
				
				// Extract dependencies from imports
				qdg.extractDependencies(content, dependencies)
			} else if path == "go.mod" || path == "go.sum" {
				drop.Files[path] = content
				technologies["Go Modules"] = true
			} else if strings.HasSuffix(path, ".json") && strings.Contains(path, "package") {
				drop.Files[path] = content
				technologies["Node.js"] = true
			}
		}
	}
	
	// Ensure go.mod exists
	if _, exists := drop.Files["go.mod"]; !exists && technologies["Go"] {
		projectName := qdg.generateProjectName(intent.UserInput)
		drop.Files["go.mod"] = fmt.Sprintf("module %s\n\ngo 1.21\n", projectName)
	}
	
	drop.Tasks = taskIDs
	drop.Metadata = DropMetadata{
		FileCount:       len(drop.Files),
		TotalLines:      qdg.countTotalLines(drop.Files),
		Technologies:    qdg.mapKeysToSlice(technologies),
		Dependencies:    qdg.mapKeysToSlice(dependencies),
		QualityScore:    qdg.calculateQualityScore(tasks),
		SecurityScore:   qdg.calculateSecurityScore(tasks),
		ValidationPassed: qdg.checkValidationPassed(tasks),
		HITLRequired:    len(drop.Files) > 5, // Require HITL for complex codebases
	}
	
	drop.Structure = qdg.generateDropStructure(drop.Files)
	
	return drop, nil
}

func (qdg *QuantumDropGenerator) generateDocumentationDrop(intent models.Intent, tasks []TaskExecutionResult) (*QuantumDrop, error) {
	drop := &QuantumDrop{
		ID:          fmt.Sprintf("QD-DOCS-%d", time.Now().Unix()),
		Type:        DropTypeDocumentation,
		Name:        "Project Documentation",
		Description: "API documentation, setup guides, and usage instructions",
		Files:       make(map[string]string),
		Structure:   make(map[string][]string),
		CreatedAt:   time.Now(),
		Status:      DropStatusReady,
	}
	
	var taskIDs []string
	
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.Task.ID)
		
		llmOutput := qdg.extractLLMOutput(task.Output)
		projectStruct, err := qdg.fileGenerator.ParseLLMOutput(task.Task.ID, string(task.Task.Type), llmOutput)
		if err != nil {
			continue
		}
		
		taskFiles := qdg.fileGenerator.GenerateFileStructure(projectStruct)
		
		for path, content := range taskFiles {
			if strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".txt") {
				if strings.Contains(path, "README") {
					drop.Files["README.md"] = content
				} else {
					drop.Files[fmt.Sprintf("docs/%s", filepath.Base(path))] = content
				}
			}
		}
	}
	
	drop.Tasks = taskIDs
	drop.Metadata = DropMetadata{
		FileCount:       len(drop.Files),
		TotalLines:      qdg.countTotalLines(drop.Files),
		Technologies:    []string{"Markdown"},
		QualityScore:    qdg.calculateQualityScore(tasks),
		ValidationPassed: qdg.checkValidationPassed(tasks),
		HITLRequired:    false, // Documentation usually doesn't need HITL
	}
	
	drop.Structure = qdg.generateDropStructure(drop.Files)
	
	return drop, nil
}

func (qdg *QuantumDropGenerator) generateTestingDrop(intent models.Intent, tasks []TaskExecutionResult) (*QuantumDrop, error) {
	drop := &QuantumDrop{
		ID:          fmt.Sprintf("QD-TEST-%d", time.Now().Unix()),
		Type:        DropTypeTesting,
		Name:        "Test Suite",
		Description: "Unit tests, integration tests, and test data",
		Files:       make(map[string]string),
		Structure:   make(map[string][]string),
		CreatedAt:   time.Now(),
		Status:      DropStatusReady,
	}
	
	var taskIDs []string
	
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.Task.ID)
		
		llmOutput := qdg.extractLLMOutput(task.Output)
		projectStruct, err := qdg.fileGenerator.ParseLLMOutput(task.Task.ID, string(task.Task.Type), llmOutput)
		if err != nil {
			continue
		}
		
		taskFiles := qdg.fileGenerator.GenerateFileStructure(projectStruct)
		
		for path, content := range taskFiles {
			if strings.Contains(path, "test") || strings.HasSuffix(path, "_test.go") {
				testPath := fmt.Sprintf("tests/%s", filepath.Base(path))
				drop.Files[testPath] = content
			}
		}
	}
	
	drop.Tasks = taskIDs
	drop.Metadata = DropMetadata{
		FileCount:       len(drop.Files),
		TotalLines:      qdg.countTotalLines(drop.Files),
		Technologies:    []string{"Go Testing"},
		QualityScore:    qdg.calculateQualityScore(tasks),
		ValidationPassed: qdg.checkValidationPassed(tasks),
		HITLRequired:    false, // Tests usually don't need HITL
	}
	
	drop.Structure = qdg.generateDropStructure(drop.Files)
	
	return drop, nil
}

func (qdg *QuantumDropGenerator) generateAnalysisDrop(intent models.Intent, tasks []TaskExecutionResult) (*QuantumDrop, error) {
	drop := &QuantumDrop{
		ID:          fmt.Sprintf("QD-ANALYSIS-%d", time.Now().Unix()),
		Type:        DropTypeAnalysis,
		Name:        "Analysis Reports",
		Description: "Security analysis, performance reports, and recommendations",
		Files:       make(map[string]string),
		Structure:   make(map[string][]string),
		CreatedAt:   time.Now(),
		Status:      DropStatusReady,
	}
	
	var taskIDs []string
	
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.Task.ID)
		
		llmOutput := qdg.extractLLMOutput(task.Output)
		projectStruct, err := qdg.fileGenerator.ParseLLMOutput(task.Task.ID, string(task.Task.Type), llmOutput)
		if err != nil {
			continue
		}
		
		taskFiles := qdg.fileGenerator.GenerateFileStructure(projectStruct)
		
		for path, content := range taskFiles {
			if strings.Contains(path, "analysis") || strings.Contains(path, "report") {
				drop.Files[fmt.Sprintf("reports/%s", filepath.Base(path))] = content
			}
		}
	}
	
	drop.Tasks = taskIDs
	drop.Metadata = DropMetadata{
		FileCount:       len(drop.Files),
		TotalLines:      qdg.countTotalLines(drop.Files),
		Technologies:    []string{"Analysis"},
		QualityScore:    qdg.calculateQualityScore(tasks),
		ValidationPassed: qdg.checkValidationPassed(tasks),
		HITLRequired:    true, // Analysis reports should be reviewed
	}
	
	drop.Structure = qdg.generateDropStructure(drop.Files)
	
	return drop, nil
}

// Helper methods
func (qdg *QuantumDropGenerator) isInfrastructureFile(path string) bool {
	infraFiles := []string{"dockerfile", "docker-compose", ".yaml", ".yml", ".tf", ".hcl"}
	pathLower := strings.ToLower(path)
	
	for _, suffix := range infraFiles {
		if strings.Contains(pathLower, suffix) {
			return true
		}
	}
	
	return false
}

func (qdg *QuantumDropGenerator) organizeCodeFile(originalPath, content string) string {
	if strings.Contains(content, "func main()") {
		return "cmd/main.go"
	}
	
	if strings.Contains(content, "func Test") {
		return fmt.Sprintf("tests/%s", filepath.Base(originalPath))
	}
	
	// Extract package name
	lines := strings.Split(content, "\n")
	packageName := "main"
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			packageName = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "package "))
			break
		}
	}
	
	// Organize by package
	switch packageName {
	case "main":
		return "cmd/main.go"
	case "handlers", "handler":
		return fmt.Sprintf("internal/handlers/%s", filepath.Base(originalPath))
	case "auth", "authentication":
		return fmt.Sprintf("internal/auth/%s", filepath.Base(originalPath))
	case "middleware":
		return fmt.Sprintf("internal/middleware/%s", filepath.Base(originalPath))
	default:
		return fmt.Sprintf("internal/%s/%s", packageName, filepath.Base(originalPath))
	}
}

func (qdg *QuantumDropGenerator) extractDependencies(content string, dependencies map[string]bool) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "github.com/") || strings.HasPrefix(trimmed, "golang.org/") {
			parts := strings.Fields(trimmed)
			if len(parts) > 0 {
				dependencies[parts[0]] = true
			}
		}
	}
}

func (qdg *QuantumDropGenerator) countTotalLines(files map[string]string) int {
	total := 0
	for _, content := range files {
		total += len(strings.Split(content, "\n"))
	}
	return total
}

func (qdg *QuantumDropGenerator) mapKeysToSlice(m map[string]bool) []string {
	var result []string
	for key := range m {
		result = append(result, key)
	}
	return result
}

func (qdg *QuantumDropGenerator) calculateQualityScore(tasks []TaskExecutionResult) int {
	if len(tasks) == 0 {
		return 0
	}
	
	total := 0
	for _, task := range tasks {
		if task.ValidationResult != nil {
			total += task.ValidationResult.OverallScore
		}
	}
	
	return total / len(tasks)
}

func (qdg *QuantumDropGenerator) calculateSecurityScore(tasks []TaskExecutionResult) int {
	if len(tasks) == 0 {
		return 0
	}
	
	total := 0
	count := 0
	for _, task := range tasks {
		if task.SandboxResult != nil {
			total += task.SandboxResult.SecurityScore
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	
	return total / count
}

func (qdg *QuantumDropGenerator) checkValidationPassed(tasks []TaskExecutionResult) bool {
	for _, task := range tasks {
		if task.ValidationResult != nil && !task.ValidationResult.Passed {
			return false
		}
	}
	return true
}

func (qdg *QuantumDropGenerator) generateDropStructure(files map[string]string) map[string][]string {
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

func (qdg *QuantumDropGenerator) generateProjectName(userInput string) string {
	input := strings.ToLower(userInput)
	
	if strings.Contains(input, "microservice") && strings.Contains(input, "auth") {
		return "auth-microservice"
	}
	if strings.Contains(input, "api") && strings.Contains(input, "user") {
		return "user-api"
	}
	
	return "quantum-project"
}

func (qdg *QuantumDropGenerator) extractLLMOutput(agentOutput string) string {
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