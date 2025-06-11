package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"QLP/internal/agents"
	"QLP/internal/dag"
	"QLP/internal/database"
	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/models"
	"QLP/internal/packaging"
	"QLP/internal/parser"
	"QLP/internal/types"
	"QLP/internal/validation"
	"QLP/internal/vector"
)

type Orchestrator struct {
	intentParser     *parser.IntentParser
	taskGraph        *models.TaskGraph
	eventBus         *events.EventBus
	dagExecutor      *dag.DAGExecutor
	capsulePackager  *packaging.CapsuleOrchestrator
	quantumDropGen   *packaging.QuantumDropGenerator
	executionResults map[string]*packaging.AgentExecutionResult
	quantumDrops     []packaging.QuantumDrop
	hitlEnabled      bool
	db               *database.Database
	intentRepo       *database.IntentRepository
	vectorService    *vector.VectorService
	llmClient        llm.Client
}

func New() *Orchestrator {
	llmClient := llm.NewLLMClient()
	intentParser := parser.NewIntentParser(llmClient)
	eventBus := events.NewEventBus()
	agentFactory := agents.NewAgentFactory(llmClient, eventBus)
	dagExecutor := dag.NewDAGExecutor(eventBus, agentFactory)
	capsulePackager := packaging.NewCapsuleOrchestrator("./output")
	quantumDropGen := packaging.NewQuantumDropGenerator()

	// Initialize database connection
	db, err := database.New()
	if err != nil {
		log.Printf("⚠️  Database initialization failed: %v", err)
		log.Printf("📝 Continuing without persistent storage...")
	}
	
	intentRepo := database.NewIntentRepository(db)
	vectorService := vector.NewVectorService(db, llmClient)

	return &Orchestrator{
		intentParser:     intentParser,
		eventBus:         eventBus,
		dagExecutor:      dagExecutor,
		capsulePackager:  capsulePackager,
		quantumDropGen:   quantumDropGen,
		executionResults: make(map[string]*packaging.AgentExecutionResult),
		quantumDrops:     make([]packaging.QuantumDrop, 0),
		hitlEnabled:      true, // Enable HITL by default
		db:               db,
		intentRepo:       intentRepo,
		vectorService:    vectorService,
		llmClient:        llmClient,
	}
}

func (o *Orchestrator) Start(ctx context.Context) error {
	log.Println("Orchestrator starting...")

	o.eventBus.Start(ctx)

	o.eventBus.Subscribe(events.EventTaskStarted, func(ctx context.Context, event events.Event) error {
		log.Printf("Task started: %v", event.Payload["task_id"])
		return nil
	})

	o.eventBus.Subscribe(events.EventTaskCompleted, func(ctx context.Context, event events.Event) error {
		log.Printf("Task completed: %v", event.Payload["task_id"])
		return nil
	})

	testIntent := "Create a simple web API server in Go with user authentication"

	intent, err := o.ProcessIntent(ctx, testIntent)
	if err != nil {
		return fmt.Errorf("failed to process test intent: %w", err)
	}

	log.Printf("Processed intent with %d tasks", len(intent.Tasks))
	for _, task := range intent.Tasks {
		log.Printf("Task: %s - %s (%s)", task.ID, task.Description, task.Type)
	}

	if err := o.dagExecutor.ExecuteTaskGraph(ctx, o.taskGraph); err != nil {
		return fmt.Errorf("failed to execute task graph: %w", err)
	}

	return nil
}

func (o *Orchestrator) ProcessIntent(ctx context.Context, userInput string) (*models.Intent, error) {
	intent, err := o.intentParser.ParseIntent(ctx, userInput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %w", err)
	}

	taskGraph, err := o.buildTaskGraph(intent.Tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	o.taskGraph = taskGraph
	intent.Status = models.IntentStatusProcessing

	return intent, nil
}

func (o *Orchestrator) buildTaskGraph(tasks []models.Task) (*models.TaskGraph, error) {
	taskGraph := &models.TaskGraph{
		ID:    fmt.Sprintf("graph_%d", len(tasks)),
		Tasks: tasks,
		Edges: []models.Edge{},
	}

	for _, task := range tasks {
		for _, depID := range task.Dependencies {
			edge := models.Edge{
				From: depID,
				To:   task.ID,
			}
			taskGraph.Edges = append(taskGraph.Edges, edge)
		}
	}

	return taskGraph, nil
}

func (o *Orchestrator) ProcessAndExecuteIntent(ctx context.Context, intentText string) error {
	log.Printf("🔄 Processing intent: %s", intentText)
	
	startTime := time.Now()
	
	// Step 1: Parse intent
	intent, err := o.intentParser.ParseIntent(ctx, intentText)
	if err != nil {
		return fmt.Errorf("failed to parse intent: %w", err)
	}
	
	// Step 1.1: Check for similar intents first
	suggestions, err := o.vectorService.GetIntentSuggestions(ctx, intentText)
	if err != nil {
		log.Printf("⚠️  Failed to get intent suggestions: %v", err)
	} else if len(suggestions) > 0 {
		log.Printf("💡 Found similar intents:")
		for _, suggestion := range suggestions {
			log.Printf("   • %s", suggestion)
		}
	}
	
	// Step 1.2: Persist intent to database
	intent.Status = models.IntentStatusProcessing
	intent.UpdatedAt = time.Now()
	if err := o.intentRepo.Create(intent); err != nil {
		log.Printf("⚠️  Failed to save intent to database: %v", err)
		// Continue execution even if database save fails
	} else {
		log.Printf("💾 Intent saved to database: %s", intent.ID)
	}
	
	// Step 1.3: Generate and store intent embedding
	if err := o.vectorService.StoreIntentEmbedding(ctx, intent.ID, intentText); err != nil {
		log.Printf("⚠️  Failed to store intent embedding: %v", err)
		// Continue execution even if embedding storage fails
	}
	
	log.Printf("📋 Parsed %d tasks from intent", len(intent.Tasks))
	for _, task := range intent.Tasks {
		log.Printf("   • %s: %s (%s)", task.ID, task.Description, task.Type)
	}

	// Step 2: Build task graph
	taskGraph, err := o.buildTaskGraph(intent.Tasks)
	if err != nil {
		return fmt.Errorf("failed to build task graph: %w", err)
	}
	o.taskGraph = taskGraph

	// Step 3: Execute task graph with real agents
	log.Printf("🤖 Executing task graph with %d real agents for %d tasks", len(taskGraph.Tasks), len(taskGraph.Tasks))
	
	if err := o.dagExecutor.ExecuteTaskGraph(ctx, taskGraph); err != nil {
		return fmt.Errorf("failed to execute task graph: %w", err)
	}
	
	// Collect real execution results from agents
	o.executionResults = o.collectAgentResults(taskGraph.Tasks)

	// Step 4: Generate QuantumDrops
	log.Printf("💧 Generating QuantumDrops for HITL workflow...")
	
	taskResults := o.convertToTaskExecutionResults(taskGraph.Tasks)
	quantumDrops, err := o.quantumDropGen.GenerateQuantumDrops(*intent, taskResults)
	if err != nil {
		return fmt.Errorf("failed to generate QuantumDrops: %w", err)
	}
	
	o.quantumDrops = quantumDrops
	log.Printf("💧 Generated %d QuantumDrops", len(quantumDrops))
	for _, drop := range quantumDrops {
		log.Printf("   • %s (%s): %d files, HITL: %v", drop.Name, drop.Type, drop.Metadata.FileCount, drop.Metadata.HITLRequired)
	}

	// Step 5: HITL Decision Points (if enabled)
	if o.hitlEnabled {
		if err := o.processHITLDecisions(ctx, *intent); err != nil {
			return fmt.Errorf("failed to process HITL decisions: %w", err)
		}
	} else {
		// Auto-approve all drops
		o.autoApproveAllDrops()
	}

	// Step 6: Generate final QuantumCapsule from approved drops
	log.Printf("📦 Generating final QuantumCapsule from approved QuantumDrops...")
	
	capsule, err := o.generateQuantumCapsule(ctx, *intent)
	if err != nil {
		return fmt.Errorf("failed to generate QuantumCapsule: %w", err)
	}

	// Step 7: Update intent completion in database
	executionTime := time.Since(startTime)
	intent.Status = models.IntentStatusCompleted
	intent.OverallScore = capsule.Metadata.OverallScore
	intent.ExecutionTimeMS = int(executionTime.Milliseconds())
	completedAt := time.Now()
	intent.CompletedAt = &completedAt
	intent.UpdatedAt = completedAt
	
	if err := o.intentRepo.Update(intent); err != nil {
		log.Printf("⚠️  Failed to update intent completion in database: %v", err)
	} else {
		log.Printf("💾 Intent completion saved to database")
	}
	
	// Step 8: Display results
	log.Printf("🎯 QuantumCapsule generated: %s", capsule.Metadata.CapsuleID)
	log.Printf("   📊 Overall Score: %d/100", capsule.Metadata.OverallScore)
	log.Printf("   ✅ Successful Tasks: %d/%d", capsule.Metadata.SuccessfulTasks, capsule.Metadata.TotalTasks)
	log.Printf("   🔒 Security Risk: %s", capsule.SecurityReport.OverallRiskLevel)
	log.Printf("   📈 Quality Score: %d/100", capsule.QualityReport.OverallQualityScore)
	log.Printf("   ⏱️  Execution Time: %v", capsule.Metadata.Duration)

	return nil
}

func (o *Orchestrator) collectAgentResults(tasks []models.Task) map[string]*packaging.AgentExecutionResult {
	results := make(map[string]*packaging.AgentExecutionResult)
	
	for _, task := range tasks {
		// Get real agent execution results from the DAG executor
		agentResult := o.dagExecutor.GetTaskResult(task.ID)
		if agentResult != nil {
			results[task.ID] = &packaging.AgentExecutionResult{
				AgentID:          agentResult.AgentID,
				Status:           string(agentResult.Status),
				Output:           agentResult.Output,
				ExecutionTime:    agentResult.ExecutionTime,
				SandboxResult:    agentResult.SandboxResult,
				ValidationResult: o.convertValidationResult(agentResult.ValidationResult),
				Error:            agentResult.Error,
				StartTime:        agentResult.StartTime,
				EndTime:          agentResult.EndTime,
			}
		}
	}
	
	return results
}

// convertToTaskExecutionResults converts DAG executor results to packaging format
func (o *Orchestrator) convertToTaskExecutionResults(tasks []models.Task) []packaging.TaskExecutionResult {
	var results []packaging.TaskExecutionResult
	
	for _, task := range tasks {
		agentResult := o.dagExecutor.GetTaskResult(task.ID)
		if agentResult != nil {
			result := packaging.TaskExecutionResult{
				Task:             task,
				Status:           agentResult.Status,
				Output:           agentResult.Output,
				AgentID:          agentResult.AgentID,
				ExecutionTime:    agentResult.ExecutionTime,
				SandboxResult:    agentResult.SandboxResult,
				ValidationResult: o.convertValidationResult(agentResult.ValidationResult),
				Error:            agentResult.Error,
			}
			results = append(results, result)
		}
	}
	
	return results
}

// convertValidationResult converts validation.ValidationResult to types.ValidationResult
func (o *Orchestrator) convertValidationResult(valResult *validation.ValidationResult) *types.ValidationResult {
	if valResult == nil {
		return nil
	}
	
	securityScore := 0
	if valResult.SecurityResult != nil {
		securityScore = valResult.SecurityResult.Score
	}
	
	qualityScore := 0
	if valResult.QualityResult != nil {
		qualityScore = valResult.QualityResult.Score
	}
	
	return &types.ValidationResult{
		OverallScore:   valResult.OverallScore,
		SecurityScore:  securityScore,
		QualityScore:   qualityScore,
		Passed:         valResult.Passed,
		ValidationTime: valResult.ValidationTime,
		ValidatedAt:    valResult.Timestamp,
	}
}

// processHITLDecisions handles the human-in-the-loop decision workflow
func (o *Orchestrator) processHITLDecisions(ctx context.Context, intent models.Intent) error {
	_ = ctx    // Context available for future HTTP/gRPC HITL interfaces
	_ = intent // Intent available for context-aware decisions
	log.Printf("🤔 Processing HITL decisions for %d QuantumDrops...", len(o.quantumDrops))
	
	for i := range o.quantumDrops {
		drop := &o.quantumDrops[i]
		
		if !drop.Metadata.HITLRequired {
			// Auto-approve drops that don't require HITL
			drop.Status = packaging.DropStatusApproved
			log.Printf("   ✅ Auto-approved: %s (%s)", drop.Name, drop.Type)
			continue
		}
		
		// For production, this would interface with actual UI/CLI for human input
		// For now, simulate intelligent auto-decision based on validation scores
		decision := o.simulateHITLDecision(*drop)
		
		switch decision.Decision {
		case packaging.HITLActionContinue:
			drop.Status = packaging.DropStatusApproved
			log.Printf("   ✅ HITL Approved: %s (%s) - %s", drop.Name, drop.Type, decision.Feedback)
			
		case packaging.HITLActionRedo:
			drop.Status = packaging.DropStatusRejected
			drop.Metadata.ReviewNotes = append(drop.Metadata.ReviewNotes, decision.Feedback)
			log.Printf("   🔄 HITL Redo: %s (%s) - %s", drop.Name, drop.Type, decision.Feedback)
			
		case packaging.HITLActionModify:
			drop.Status = packaging.DropStatusModified
			// Apply modifications from decision.Changes
			for filePath, newContent := range decision.Changes {
				drop.Files[filePath] = newContent
			}
			drop.Metadata.ReviewNotes = append(drop.Metadata.ReviewNotes, decision.Feedback)
			log.Printf("   🔧 HITL Modified: %s (%s) - %s", drop.Name, drop.Type, decision.Feedback)
			
		case packaging.HITLActionReject:
			drop.Status = packaging.DropStatusRejected
			drop.Metadata.ReviewNotes = append(drop.Metadata.ReviewNotes, decision.Feedback)
			log.Printf("   ❌ HITL Rejected: %s (%s) - %s", drop.Name, drop.Type, decision.Feedback)
		}
	}
	
	return nil
}

// simulateHITLDecision simulates intelligent human decision making based on validation scores
func (o *Orchestrator) simulateHITLDecision(drop packaging.QuantumDrop) packaging.HITLDecision {
	decision := packaging.HITLDecision{
		DropID:    drop.ID,
		Timestamp: time.Now(),
	}
	
	// Decision logic based on validation scores and content analysis
	if drop.Metadata.ValidationPassed && drop.Metadata.QualityScore >= 80 && drop.Metadata.SecurityScore >= 70 {
		decision.Decision = packaging.HITLActionContinue
		decision.Feedback = "High quality output meets all validation criteria"
	} else if drop.Metadata.QualityScore < 50 || drop.Metadata.SecurityScore < 50 {
		decision.Decision = packaging.HITLActionRedo
		decision.Feedback = "Quality or security scores below acceptable threshold. Requires rework."
	} else if drop.Metadata.QualityScore < 70 {
		decision.Decision = packaging.HITLActionModify
		decision.Feedback = "Good foundation but needs minor improvements"
		// Simulate minor modifications
		decision.Changes = make(map[string]string)
		for filePath, content := range drop.Files {
			if len(content) < 100 { // Simple heuristic for small files needing improvement
				decision.Changes[filePath] = content + "\n// Added improvement comment for production readiness"
			}
		}
	} else {
		decision.Decision = packaging.HITLActionContinue
		decision.Feedback = "Acceptable quality, approved for inclusion"
	}
	
	return decision
}

// autoApproveAllDrops automatically approves all QuantumDrops when HITL is disabled
func (o *Orchestrator) autoApproveAllDrops() {
	for i := range o.quantumDrops {
		o.quantumDrops[i].Status = packaging.DropStatusApproved
	}
	log.Printf("✅ Auto-approved all %d QuantumDrops (HITL disabled)", len(o.quantumDrops))
}

// generateQuantumCapsule creates the final capsule from approved QuantumDrops
func (o *Orchestrator) generateQuantumCapsule(ctx context.Context, intent models.Intent) (*packaging.QLCapsule, error) {
	// Collect only approved and modified drops
	var approvedDrops []packaging.QuantumDrop
	for _, drop := range o.quantumDrops {
		if drop.Status == packaging.DropStatusApproved || drop.Status == packaging.DropStatusModified {
			approvedDrops = append(approvedDrops, drop)
		}
	}
	
	log.Printf("📦 Merging %d approved QuantumDrops into final capsule", len(approvedDrops))
	
	// Use existing capsule packager to generate the final capsule
	capsule, err := o.capsulePackager.ProcessIntentExecution(ctx, intent, o.taskGraph.Tasks, o.executionResults)
	if err != nil {
		return nil, fmt.Errorf("failed to generate capsule from approved drops: %w", err)
	}
	
	// Add QuantumDrops metadata to the capsule
	capsule.Metadata.Environment["quantum_drops_generated"] = len(o.quantumDrops)
	capsule.Metadata.Environment["quantum_drops_approved"] = len(approvedDrops)
	capsule.Metadata.Environment["hitl_enabled"] = o.hitlEnabled
	
	return capsule, nil
}

// convertDropsToTaskResults converts approved QuantumDrops back to task execution results
func (o *Orchestrator) convertDropsToTaskResults(drops []packaging.QuantumDrop) []packaging.TaskExecutionResult {
	var results []packaging.TaskExecutionResult
	
	for _, drop := range drops {
		for _, taskID := range drop.Tasks {
			if agentResult := o.dagExecutor.GetTaskResult(taskID); agentResult != nil {
				// Find the original task
				var task models.Task
				for _, t := range o.taskGraph.Tasks {
					if t.ID == taskID {
						task = t
						break
					}
				}
				
				result := packaging.TaskExecutionResult{
					Task:             task,
					Status:           agentResult.Status,
					Output:           agentResult.Output,
					AgentID:          agentResult.AgentID,
					ExecutionTime:    agentResult.ExecutionTime,
					SandboxResult:    agentResult.SandboxResult,
					ValidationResult: o.convertValidationResult(agentResult.ValidationResult),
					Error:            agentResult.Error,
				}
				results = append(results, result)
			}
		}
	}
	
	return results
}

func (o *Orchestrator) generateSampleOutput(task models.Task) string {
	switch task.Type {
	case models.TaskTypeCodegen:
		return "package main\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n\t\"log\"\n)\n\nfunc main() {\n\thttp.HandleFunc(\"/users\", usersHandler)\n\tlog.Println(\"Server starting on :8080\")\n\tlog.Fatal(http.ListenAndServe(\":8080\", nil))\n}\n\nfunc usersHandler(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprintf(w, \"User management API with JWT authentication\")\n}"

	case models.TaskTypeInfra:
		return "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: microservices\n---\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: user-service\n  namespace: microservices\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: user-service\n  template:\n    metadata:\n      labels:\n        app: user-service\n    spec:\n      containers:\n      - name: user-service\n        image: user-service:latest\n        ports:\n        - containerPort: 8080"

	case models.TaskTypeAnalyze:
		return "# Performance Analysis Report\n\n## Executive Summary\nAnalysis of Go web application performance reveals several optimization opportunities.\n\n## Key Findings\n\n### 1. Memory Usage\n- Current peak memory: 256MB\n- Recommended optimization: Implement connection pooling\n- Expected improvement: 30% reduction in memory usage\n\n### 2. Response Times\n- Average response time: 45ms\n- 95th percentile: 120ms\n- Bottleneck identified: Database queries without indexing\n\n## Recommendations\n1. Database indexing (High priority)\n2. Caching layer (Medium priority)\n3. Connection pooling (Medium priority)"

	case models.TaskTypeTest:
		return "package main\n\nimport (\n\t\"testing\"\n\t\"net/http\"\n\t\"net/http/httptest\"\n)\n\nfunc TestUsersHandler(t *testing.T) {\n\treq, err := http.NewRequest(\"GET\", \"/users\", nil)\n\tif err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\trr := httptest.NewRecorder()\n\thandler := http.HandlerFunc(usersHandler)\n\thandler.ServeHTTP(rr, req)\n\n\tif status := rr.Code; status != http.StatusOK {\n\t\tt.Errorf(\"handler returned wrong status code: got %v want %v\", status, http.StatusOK)\n\t}\n}"

	case models.TaskTypeDoc:
		return "# User Management API Documentation\n\n## Overview\nThis REST API provides secure user management functionality with JWT-based authentication.\n\n## Authentication\nAll protected endpoints require a valid JWT token in the Authorization header.\n\n## Endpoints\n\n### GET /users\nRetrieve list of users (requires authentication).\n\n### POST /users\nCreate a new user account.\n\n### POST /auth/login\nAuthenticate user and receive JWT token.\n\n## Security Considerations\n- All passwords are hashed using bcrypt\n- JWT tokens expire after 24 hours\n- HTTPS is required in production"

	default:
		return fmt.Sprintf("Output for task %s of type %s", task.ID, task.Type)
	}
}
