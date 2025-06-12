package engines

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"QLP/internal/logger"
	"QLP/services/packaging-service/pkg/contracts"
)

// QuantumDropsEngine handles creation and management of quantum drops
type QuantumDropsEngine struct {
	outputDir string
}

// NewQuantumDropsEngine creates a new quantum drops engine
func NewQuantumDropsEngine(outputDir string) *QuantumDropsEngine {
	return &QuantumDropsEngine{
		outputDir: outputDir,
	}
}

// CreateQuantumDrops creates categorized quantum drops from tasks and project files
func (qde *QuantumDropsEngine) CreateQuantumDrops(ctx context.Context, tenantID string, req *contracts.CreateQuantumDropRequest) ([]contracts.QuantumDrop, error) {
	logger.WithComponent("quantum-drops-engine").Info("Creating quantum drops",
		zap.String("tenant_id", tenantID),
		zap.String("intent_id", req.IntentID),
		zap.Int("task_count", len(req.Tasks)))

	var drops []contracts.QuantumDrop

	// Determine which drop types to create
	dropTypes := req.DropTypes
	if len(dropTypes) == 0 {
		// Create all types by default
		dropTypes = []contracts.DropType{
			contracts.DropTypeInfrastructure,
			contracts.DropTypeCodebase,
			contracts.DropTypeDocumentation,
			contracts.DropTypeTesting,
			contracts.DropTypeAnalysis,
		}
	}

	// Create drops for each requested type
	for _, dropType := range dropTypes {
		drop, err := qde.createDropByType(dropType, req)
		if err != nil {
			logger.WithComponent("quantum-drops-engine").Error("Failed to create drop",
				zap.String("drop_type", string(dropType)),
				zap.Error(err))
			continue
		}

		if drop != nil {
			drops = append(drops, *drop)
		}
	}

	logger.WithComponent("quantum-drops-engine").Info("Quantum drops created successfully",
		zap.String("intent_id", req.IntentID),
		zap.Int("drops_created", len(drops)))

	return drops, nil
}

// createDropByType creates a specific type of quantum drop
func (qde *QuantumDropsEngine) createDropByType(dropType contracts.DropType, req *contracts.CreateQuantumDropRequest) (*contracts.QuantumDrop, error) {
	switch dropType {
	case contracts.DropTypeInfrastructure:
		return qde.createInfrastructureDrop(req)
	case contracts.DropTypeCodebase:
		return qde.createCodebaseDrop(req)
	case contracts.DropTypeDocumentation:
		return qde.createDocumentationDrop(req)
	case contracts.DropTypeTesting:
		return qde.createTestingDrop(req)
	case contracts.DropTypeAnalysis:
		return qde.createAnalysisDrop(req)
	default:
		return nil, fmt.Errorf("unsupported drop type: %s", dropType)
	}
}

// createInfrastructureDrop creates infrastructure-related files drop
func (qde *QuantumDropsEngine) createInfrastructureDrop(req *contracts.CreateQuantumDropRequest) (*contracts.QuantumDrop, error) {
	files := make(map[string]string)
	structure := make(map[string][]string)
	relatedTasks := []string{}
	technologies := []string{}

	// Collect infrastructure files from tasks
	for _, task := range req.Tasks {
		if task.Type == contracts.TaskTypeInfra {
			relatedTasks = append(relatedTasks, task.ID)
			for fileName, content := range task.Files {
				if qde.isInfrastructureFile(fileName) {
					files[fileName] = content
					qde.categorizeFile(fileName, structure)
				}
			}
		}
	}

	// Collect infrastructure files from project files
	for fileName, content := range req.ProjectFiles {
		if qde.isInfrastructureFile(fileName) {
			files[fileName] = content
			qde.categorizeFile(fileName, structure)
		}
	}

	// Skip if no infrastructure files found
	if len(files) == 0 {
		return nil, nil
	}

	// Detect technologies
	technologies = qde.detectInfrastructureTechnologies(files)

	metadata := contracts.DropMetadata{
		FileCount:       len(files),
		TotalLines:      qde.countTotalLines(files),
		Technologies:    technologies,
		EstimatedEffort: qde.estimateEffort(len(files), qde.countTotalLines(files)),
		Complexity:      qde.estimateComplexity(files),
		Dependencies:    qde.extractDependencies(files),
		CustomFields:    map[string]interface{}{"category": "infrastructure"},
	}

	return &contracts.QuantumDrop{
		ID:          generateDropID(),
		Type:        contracts.DropTypeInfrastructure,
		Name:        "Infrastructure Drop",
		Description: "Infrastructure configuration and deployment files",
		Files:       files,
		Structure:   structure,
		Metadata:    metadata,
		Status:      contracts.DropStatusReady,
		CreatedAt:   time.Now(),
		Tasks:       relatedTasks,
	}, nil
}

// createCodebaseDrop creates codebase-related files drop
func (qde *QuantumDropsEngine) createCodebaseDrop(req *contracts.CreateQuantumDropRequest) (*contracts.QuantumDrop, error) {
	files := make(map[string]string)
	structure := make(map[string][]string)
	relatedTasks := []string{}
	technologies := []string{}

	// Collect code files from tasks
	for _, task := range req.Tasks {
		if task.Type == contracts.TaskTypeCodegen {
			relatedTasks = append(relatedTasks, task.ID)
			for fileName, content := range task.Files {
				if qde.isCodeFile(fileName) {
					files[fileName] = content
					qde.categorizeFile(fileName, structure)
				}
			}
		}
	}

	// Collect code files from project files
	for fileName, content := range req.ProjectFiles {
		if qde.isCodeFile(fileName) {
			files[fileName] = content
			qde.categorizeFile(fileName, structure)
		}
	}

	// Skip if no code files found
	if len(files) == 0 {
		return nil, nil
	}

	// Detect technologies
	technologies = qde.detectCodeTechnologies(files)

	metadata := contracts.DropMetadata{
		FileCount:       len(files),
		TotalLines:      qde.countTotalLines(files),
		Technologies:    technologies,
		EstimatedEffort: qde.estimateEffort(len(files), qde.countTotalLines(files)),
		Complexity:      qde.estimateComplexity(files),
		Dependencies:    qde.extractDependencies(files),
		CustomFields:    map[string]interface{}{"category": "codebase"},
	}

	return &contracts.QuantumDrop{
		ID:          generateDropID(),
		Type:        contracts.DropTypeCodebase,
		Name:        "Codebase Drop",
		Description: "Application source code and related files",
		Files:       files,
		Structure:   structure,
		Metadata:    metadata,
		Status:      contracts.DropStatusReady,
		CreatedAt:   time.Now(),
		Tasks:       relatedTasks,
	}, nil
}

// createDocumentationDrop creates documentation-related files drop
func (qde *QuantumDropsEngine) createDocumentationDrop(req *contracts.CreateQuantumDropRequest) (*contracts.QuantumDrop, error) {
	files := make(map[string]string)
	structure := make(map[string][]string)
	relatedTasks := []string{}

	// Collect documentation files from tasks
	for _, task := range req.Tasks {
		if task.Type == contracts.TaskTypeDoc {
			relatedTasks = append(relatedTasks, task.ID)
			for fileName, content := range task.Files {
				if qde.isDocumentationFile(fileName) {
					files[fileName] = content
					qde.categorizeFile(fileName, structure)
				}
			}
		}
	}

	// Collect documentation files from project files
	for fileName, content := range req.ProjectFiles {
		if qde.isDocumentationFile(fileName) {
			files[fileName] = content
			qde.categorizeFile(fileName, structure)
		}
	}

	// Skip if no documentation files found
	if len(files) == 0 {
		return nil, nil
	}

	metadata := contracts.DropMetadata{
		FileCount:       len(files),
		TotalLines:      qde.countTotalLines(files),
		Technologies:    []string{"markdown", "documentation"},
		EstimatedEffort: qde.estimateEffort(len(files), qde.countTotalLines(files)),
		Complexity:      "low",
		Dependencies:    []string{},
		CustomFields:    map[string]interface{}{"category": "documentation"},
	}

	return &contracts.QuantumDrop{
		ID:          generateDropID(),
		Type:        contracts.DropTypeDocumentation,
		Name:        "Documentation Drop",
		Description: "Project documentation and guides",
		Files:       files,
		Structure:   structure,
		Metadata:    metadata,
		Status:      contracts.DropStatusReady,
		CreatedAt:   time.Now(),
		Tasks:       relatedTasks,
	}, nil
}

// createTestingDrop creates testing-related files drop
func (qde *QuantumDropsEngine) createTestingDrop(req *contracts.CreateQuantumDropRequest) (*contracts.QuantumDrop, error) {
	files := make(map[string]string)
	structure := make(map[string][]string)
	relatedTasks := []string{}

	// Collect test files from tasks
	for _, task := range req.Tasks {
		if task.Type == contracts.TaskTypeTest {
			relatedTasks = append(relatedTasks, task.ID)
			for fileName, content := range task.Files {
				if qde.isTestFile(fileName) {
					files[fileName] = content
					qde.categorizeFile(fileName, structure)
				}
			}
		}
	}

	// Collect test files from project files
	for fileName, content := range req.ProjectFiles {
		if qde.isTestFile(fileName) {
			files[fileName] = content
			qde.categorizeFile(fileName, structure)
		}
	}

	// Skip if no test files found
	if len(files) == 0 {
		return nil, nil
	}

	technologies := qde.detectTestTechnologies(files)

	metadata := contracts.DropMetadata{
		FileCount:       len(files),
		TotalLines:      qde.countTotalLines(files),
		Technologies:    technologies,
		EstimatedEffort: qde.estimateEffort(len(files), qde.countTotalLines(files)),
		Complexity:      qde.estimateComplexity(files),
		Dependencies:    qde.extractDependencies(files),
		CustomFields:    map[string]interface{}{"category": "testing"},
	}

	return &contracts.QuantumDrop{
		ID:          generateDropID(),
		Type:        contracts.DropTypeTesting,
		Name:        "Testing Drop",
		Description: "Test files and testing configuration",
		Files:       files,
		Structure:   structure,
		Metadata:    metadata,
		Status:      contracts.DropStatusReady,
		CreatedAt:   time.Now(),
		Tasks:       relatedTasks,
	}, nil
}

// createAnalysisDrop creates analysis-related files drop
func (qde *QuantumDropsEngine) createAnalysisDrop(req *contracts.CreateQuantumDropRequest) (*contracts.QuantumDrop, error) {
	files := make(map[string]string)
	structure := make(map[string][]string)
	relatedTasks := []string{}

	// Collect analysis files from tasks
	for _, task := range req.Tasks {
		if task.Type == contracts.TaskTypeAnalyze {
			relatedTasks = append(relatedTasks, task.ID)
			for fileName, content := range task.Files {
				files[fileName] = content
				qde.categorizeFile(fileName, structure)
			}
		}
	}

	// Skip if no analysis files found
	if len(files) == 0 {
		return nil, nil
	}

	metadata := contracts.DropMetadata{
		FileCount:       len(files),
		TotalLines:      qde.countTotalLines(files),
		Technologies:    []string{"analysis", "reporting"},
		EstimatedEffort: "low",
		Complexity:      "low",
		Dependencies:    []string{},
		CustomFields:    map[string]interface{}{"category": "analysis"},
	}

	return &contracts.QuantumDrop{
		ID:          generateDropID(),
		Type:        contracts.DropTypeAnalysis,
		Name:        "Analysis Drop",
		Description: "Analysis reports and insights",
		Files:       files,
		Structure:   structure,
		Metadata:    metadata,
		Status:      contracts.DropStatusReady,
		CreatedAt:   time.Now(),
		Tasks:       relatedTasks,
	}, nil
}

// File type detection methods

func (qde *QuantumDropsEngine) isInfrastructureFile(fileName string) bool {
	infraPatterns := []string{
		"dockerfile", "docker-compose", ".yaml", ".yml",
		"terraform", ".tf", "kubernetes", "k8s",
		"helm", "ansible", "vagrant", "jenkins",
		"pipeline", "deploy", "infrastructure",
	}

	lowerName := strings.ToLower(fileName)
	for _, pattern := range infraPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

func (qde *QuantumDropsEngine) isCodeFile(fileName string) bool {
	codeExtensions := []string{
		".go", ".js", ".ts", ".py", ".java", ".rb", ".php",
		".cs", ".cpp", ".c", ".h", ".swift", ".kt", ".rs",
		".scala", ".clj", ".hs", ".lua", ".r", ".pl",
	}

	lowerName := strings.ToLower(fileName)
	for _, ext := range codeExtensions {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	return false
}

func (qde *QuantumDropsEngine) isDocumentationFile(fileName string) bool {
	docPatterns := []string{
		".md", ".txt", ".doc", ".pdf", "readme", "changelog",
		"license", "contributing", "docs/", "documentation/",
	}

	lowerName := strings.ToLower(fileName)
	for _, pattern := range docPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

func (qde *QuantumDropsEngine) isTestFile(fileName string) bool {
	testPatterns := []string{
		"test", "spec", "_test.", ".test.", "tests/",
		"__tests__", "testing", "e2e", "integration",
	}

	lowerName := strings.ToLower(fileName)
	for _, pattern := range testPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

// Technology detection methods

func (qde *QuantumDropsEngine) detectInfrastructureTechnologies(files map[string]string) []string {
	technologies := []string{}
	techMap := make(map[string]bool)

	for fileName := range files {
		lowerName := strings.ToLower(fileName)
		if strings.Contains(lowerName, "docker") {
			techMap["docker"] = true
		}
		if strings.Contains(lowerName, "kubernetes") || strings.Contains(lowerName, "k8s") {
			techMap["kubernetes"] = true
		}
		if strings.Contains(lowerName, "terraform") {
			techMap["terraform"] = true
		}
		if strings.Contains(lowerName, "ansible") {
			techMap["ansible"] = true
		}
		if strings.Contains(lowerName, "helm") {
			techMap["helm"] = true
		}
	}

	for tech := range techMap {
		technologies = append(technologies, tech)
	}

	return technologies
}

func (qde *QuantumDropsEngine) detectCodeTechnologies(files map[string]string) []string {
	technologies := []string{}
	techMap := make(map[string]bool)

	for fileName := range files {
		ext := strings.ToLower(filepath.Ext(fileName))
		switch ext {
		case ".go":
			techMap["go"] = true
		case ".js":
			techMap["javascript"] = true
		case ".ts":
			techMap["typescript"] = true
		case ".py":
			techMap["python"] = true
		case ".java":
			techMap["java"] = true
		case ".rb":
			techMap["ruby"] = true
		case ".php":
			techMap["php"] = true
		case ".cs":
			techMap["csharp"] = true
		}
	}

	for tech := range techMap {
		technologies = append(technologies, tech)
	}

	return technologies
}

func (qde *QuantumDropsEngine) detectTestTechnologies(files map[string]string) []string {
	technologies := []string{"testing"}
	techMap := make(map[string]bool)

	for fileName, content := range files {
		lowerContent := strings.ToLower(content)
		if strings.Contains(lowerContent, "jest") {
			techMap["jest"] = true
		}
		if strings.Contains(lowerContent, "mocha") {
			techMap["mocha"] = true
		}
		if strings.Contains(lowerContent, "pytest") {
			techMap["pytest"] = true
		}
		if strings.Contains(lowerContent, "junit") {
			techMap["junit"] = true
		}
		if strings.Contains(fileName, "_test.go") {
			techMap["go-testing"] = true
		}
	}

	for tech := range techMap {
		technologies = append(technologies, tech)
	}

	return technologies
}

// Utility methods

func (qde *QuantumDropsEngine) categorizeFile(fileName string, structure map[string][]string) {
	dir := filepath.Dir(fileName)
	if dir == "." {
		dir = "root"
	}
	structure[dir] = append(structure[dir], filepath.Base(fileName))
}

func (qde *QuantumDropsEngine) countTotalLines(files map[string]string) int {
	totalLines := 0
	for _, content := range files {
		totalLines += len(strings.Split(content, "\n"))
	}
	return totalLines
}

func (qde *QuantumDropsEngine) estimateEffort(fileCount, lineCount int) string {
	if lineCount < 100 {
		return "low"
	} else if lineCount < 1000 {
		return "medium"
	} else {
		return "high"
	}
}

func (qde *QuantumDropsEngine) estimateComplexity(files map[string]string) string {
	totalComplexity := 0
	
	for _, content := range files {
		// Simple complexity heuristic based on control structures
		complexity := strings.Count(content, "if ") +
			strings.Count(content, "for ") +
			strings.Count(content, "while ") +
			strings.Count(content, "switch ") +
			strings.Count(content, "case ")
		totalComplexity += complexity
	}

	if totalComplexity < 10 {
		return "low"
	} else if totalComplexity < 50 {
		return "medium"
	} else {
		return "high"
	}
}

func (qde *QuantumDropsEngine) extractDependencies(files map[string]string) []string {
	dependencies := []string{}
	depMap := make(map[string]bool)

	for fileName, content := range files {
		if strings.Contains(fileName, "package.json") {
			// Extract npm dependencies (simplified)
			if strings.Contains(content, "dependencies") {
				depMap["npm"] = true
			}
		}
		if strings.Contains(fileName, "go.mod") {
			depMap["go-modules"] = true
		}
		if strings.Contains(fileName, "requirements.txt") {
			depMap["pip"] = true
		}
		if strings.Contains(fileName, "pom.xml") {
			depMap["maven"] = true
		}
	}

	for dep := range depMap {
		dependencies = append(dependencies, dep)
	}

	return dependencies
}

func generateDropID() string {
	return fmt.Sprintf("QD-%s", strings.ReplaceAll(uuid.New().String(), "-", "")[:12])
}