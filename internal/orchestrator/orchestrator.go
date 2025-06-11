package orchestrator

import (
	"context"
	"fmt"
	"time"

	"QLP/internal/agents"
	"QLP/internal/dag"
	"QLP/internal/database"
	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/internal/logger"
	"QLP/internal/models"
	"QLP/internal/packaging"
	"QLP/internal/parser"
	"QLP/internal/types"
	"QLP/internal/vector"
	"go.uber.org/zap"
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
		logger.Logger.Warn("Database initialization failed",
			zap.Error(err))
		logger.Logger.Info("Continuing without persistent storage")
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
	logger.WithComponent("orchestrator").Info("Orchestrator starting")

	o.eventBus.Start(ctx)

	o.eventBus.Subscribe(events.EventTaskStarted, func(ctx context.Context, event events.Event) error {
		logger.WithComponent("orchestrator").Info("Task started",
			zap.Any("task_id", event.Payload["task_id"]))
		return nil
	})

	o.eventBus.Subscribe(events.EventTaskCompleted, func(ctx context.Context, event events.Event) error {
		logger.WithComponent("orchestrator").Info("Task completed",
			zap.Any("task_id", event.Payload["task_id"]))
		return nil
	})

	testIntent := "Create a simple web API server in Go with user authentication"

	intent, err := o.ProcessIntent(ctx, testIntent)
	if err != nil {
		return fmt.Errorf("failed to process test intent: %w", err)
	}

	logger.WithComponent("orchestrator").Info("Processed intent",
		zap.Int("task_count", len(intent.Tasks)))
	for _, task := range intent.Tasks {
		logger.WithComponent("orchestrator").Info("Task parsed",
			zap.String("task_id", task.ID),
			zap.String("description", task.Description),
			zap.String("type", string(task.Type)))
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
	logger.WithComponent("orchestrator").Info("Processing intent",
		zap.String("intent_text", intentText))
	
	startTime := time.Now()
	
	// Step 1: Parse intent
	intent, err := o.intentParser.ParseIntent(ctx, intentText)
	if err != nil {
		return fmt.Errorf("failed to parse intent: %w", err)
	}
	
	// Step 1.1: Check for similar intents first
	suggestions, err := o.vectorService.GetIntentSuggestions(ctx, intentText)
	if err != nil {
		logger.WithComponent("orchestrator").Warn("Failed to get intent suggestions",
			zap.Error(err))
	} else if len(suggestions) > 0 {
		logger.WithComponent("orchestrator").Info("Found similar intents",
			zap.Strings("suggestions", suggestions))
	}
	
	// Step 1.2: Persist intent to database
	intent.Status = models.IntentStatusProcessing
	intent.UpdatedAt = time.Now()
	if err := o.intentRepo.Create(intent); err != nil {
		logger.WithComponent("orchestrator").Warn("Failed to save intent to database",
			zap.Error(err))
		// Continue execution even if database save fails
	} else {
		logger.WithComponent("orchestrator").Info("Intent saved to database",
			zap.String("intent_id", intent.ID))
	}
	
	// Step 1.3: Generate and store intent embedding
	if err := o.vectorService.StoreIntentEmbedding(ctx, intent.ID, intentText); err != nil {
		logger.WithComponent("orchestrator").Warn("Failed to store intent embedding",
			zap.Error(err))
		// Continue execution even if embedding storage fails
	} else {
		logger.WithComponent("orchestrator").Info("Stored embedding for intent",
			zap.String("intent_id", intent.ID))
	}
	
	logger.WithComponent("orchestrator").Info("Parsed tasks from intent",
		zap.Int("task_count", len(intent.Tasks)))
	for _, task := range intent.Tasks {
		logger.WithComponent("orchestrator").Info("Task parsed",
			zap.String("task_id", task.ID),
			zap.String("description", task.Description),
			zap.String("type", string(task.Type)))
	}

	// Step 2: Build task graph
	taskGraph, err := o.buildTaskGraph(intent.Tasks)
	if err != nil {
		return fmt.Errorf("failed to build task graph: %w", err)
	}
	o.taskGraph = taskGraph

	// Step 3: Execute task graph with real agents
	logger.WithComponent("orchestrator").Info("Executing task graph with real agents",
		zap.Int("agent_count", len(taskGraph.Tasks)),
		zap.Int("task_count", len(taskGraph.Tasks)))
	
	if err := o.dagExecutor.ExecuteTaskGraph(ctx, taskGraph); err != nil {
		return fmt.Errorf("failed to execute task graph: %w", err)
	}
	
	// Collect real execution results from agents
	o.executionResults = o.collectAgentResults(taskGraph.Tasks)

	// Step 4: Generate QuantumDrops
	logger.WithComponent("orchestrator").Info("Generating QuantumDrops for HITL workflow")
	
	taskResults := o.convertToTaskExecutionResults(taskGraph.Tasks)
	quantumDrops, err := o.quantumDropGen.GenerateQuantumDrops(*intent, taskResults)
	if err != nil {
		return fmt.Errorf("failed to generate QuantumDrops: %w", err)
	}
	
	o.quantumDrops = quantumDrops
	logger.WithComponent("orchestrator").Info("Generated QuantumDrops",
		zap.Int("drop_count", len(quantumDrops)))
	for _, drop := range quantumDrops {
		logger.WithComponent("orchestrator").Info("QuantumDrop created",
			zap.String("name", drop.Name),
			zap.String("type", string(drop.Type)),
			zap.Int("file_count", drop.Metadata.FileCount),
			zap.Bool("hitl_required", drop.Metadata.HITLRequired))
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
	logger.WithComponent("orchestrator").Info("Generating final QuantumCapsule from approved QuantumDrops")
	
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
		logger.WithComponent("orchestrator").Warn("Failed to update intent completion in database",
			zap.Error(err))
	} else {
		logger.WithComponent("orchestrator").Info("Intent completion saved to database")
	}
	
	// Step 8: Display results
	logger.WithComponent("orchestrator").Info("QuantumCapsule generated",
		zap.String("capsule_id", capsule.Metadata.CapsuleID),
		zap.Int("overall_score", capsule.Metadata.OverallScore),
		zap.Int("successful_tasks", capsule.Metadata.SuccessfulTasks),
		zap.Int("total_tasks", capsule.Metadata.TotalTasks),
		zap.String("security_risk", string(capsule.SecurityReport.OverallRiskLevel)),
		zap.Int("quality_score", capsule.QualityReport.OverallQualityScore),
		zap.Duration("execution_time", capsule.Metadata.Duration))

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

// convertValidationResult is now a passthrough since types are aligned
func (o *Orchestrator) convertValidationResult(valResult *types.ValidationResult) *types.ValidationResult {
	return valResult
}

// processHITLDecisions handles the human-in-the-loop decision workflow
func (o *Orchestrator) processHITLDecisions(ctx context.Context, intent models.Intent) error {
	_ = ctx    // Context available for future HTTP/gRPC HITL interfaces
	_ = intent // Intent available for context-aware decisions
	logger.WithComponent("orchestrator").Info("Processing HITL decisions",
		zap.Int("quantum_drop_count", len(o.quantumDrops)))
	
	for i := range o.quantumDrops {
		drop := &o.quantumDrops[i]
		
		if !drop.Metadata.HITLRequired {
			// Auto-approve drops that don't require HITL
			drop.Status = packaging.DropStatusApproved
			logger.WithComponent("orchestrator").Info("Auto-approved QuantumDrop",
				zap.String("name", drop.Name),
				zap.String("type", string(drop.Type)))
			continue
		}
		
		// For production, this would interface with actual UI/CLI for human input
		// For now, simulate intelligent auto-decision based on validation scores
		decision := o.simulateHITLDecision(*drop)
		
		switch decision.Decision {
		case packaging.HITLActionContinue:
			drop.Status = packaging.DropStatusApproved
			logger.WithComponent("orchestrator").Info("HITL Approved",
				zap.String("name", drop.Name),
				zap.String("type", string(drop.Type)),
				zap.String("feedback", decision.Feedback))
			
		case packaging.HITLActionRedo:
			drop.Status = packaging.DropStatusRejected
			drop.Metadata.ReviewNotes = append(drop.Metadata.ReviewNotes, decision.Feedback)
			logger.WithComponent("orchestrator").Warn("HITL Redo",
				zap.String("name", drop.Name),
				zap.String("type", string(drop.Type)),
				zap.String("feedback", decision.Feedback))
			
		case packaging.HITLActionModify:
			drop.Status = packaging.DropStatusModified
			// Apply modifications from decision.Changes
			for filePath, newContent := range decision.Changes {
				drop.Files[filePath] = newContent
			}
			drop.Metadata.ReviewNotes = append(drop.Metadata.ReviewNotes, decision.Feedback)
			logger.WithComponent("orchestrator").Info("HITL Modified",
				zap.String("name", drop.Name),
				zap.String("type", string(drop.Type)),
				zap.String("feedback", decision.Feedback))
			
		case packaging.HITLActionReject:
			drop.Status = packaging.DropStatusRejected
			drop.Metadata.ReviewNotes = append(drop.Metadata.ReviewNotes, decision.Feedback)
			logger.WithComponent("orchestrator").Warn("HITL Rejected",
				zap.String("name", drop.Name),
				zap.String("type", string(drop.Type)),
				zap.String("feedback", decision.Feedback))
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
	logger.WithComponent("orchestrator").Info("Auto-approved all QuantumDrops",
		zap.Int("quantum_drop_count", len(o.quantumDrops)),
		zap.Bool("hitl_disabled", true))
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
	
	logger.WithComponent("orchestrator").Info("Merging approved QuantumDrops into final capsule",
		zap.Int("approved_drops", len(approvedDrops)))
	
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
