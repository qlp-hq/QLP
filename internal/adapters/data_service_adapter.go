package adapters

import (
	"context"
	"fmt"
	"os"

	"QLP/internal/database"
	"QLP/internal/models"
	"QLP/services/data-service/pkg/client"
	"QLP/services/data-service/pkg/contracts"
)

// DataServiceAdapter provides an interface that can use either the data service or fallback to local DB
type DataServiceAdapter struct {
	dataClient *client.DataClient
	localRepo  *database.IntentRepository // Fallback to existing implementation
	useService bool
}

func NewDataServiceAdapter() *DataServiceAdapter {
	dataServiceURL := os.Getenv("DATA_SERVICE_URL")
	
	adapter := &DataServiceAdapter{
		useService: dataServiceURL != "",
	}
	
	if adapter.useService {
		adapter.dataClient = client.NewDataClient(dataServiceURL)
	}
	
	// TODO: Initialize localRepo as fallback
	// adapter.localRepo = database.NewIntentRepository(db)
	
	return adapter
}

// CreateIntent creates a new intent using the appropriate backend
func (dsa *DataServiceAdapter) CreateIntent(ctx context.Context, tenantID string, userInput string, metadata map[string]string) (*models.Intent, error) {
	if dsa.useService && dsa.dataClient != nil {
		// Use data service
		req := &contracts.CreateIntentRequest{
			UserInput: userInput,
			Metadata:  metadata,
		}
		
		contractIntent, err := dsa.dataClient.CreateIntent(ctx, tenantID, req)
		if err != nil {
			// Fallback to local if service fails
			if dsa.localRepo != nil {
				return dsa.createIntentLocal(ctx, tenantID, userInput, metadata)
			}
			return nil, fmt.Errorf("data service failed and no local fallback: %w", err)
		}
		
		return dsa.contractToModel(contractIntent), nil
	}
	
	// Use local repository
	return dsa.createIntentLocal(ctx, tenantID, userInput, metadata)
}

// GetIntent retrieves an intent by ID
func (dsa *DataServiceAdapter) GetIntent(ctx context.Context, tenantID, intentID string) (*models.Intent, error) {
	if dsa.useService && dsa.dataClient != nil {
		contractIntent, err := dsa.dataClient.GetIntent(ctx, tenantID, intentID)
		if err != nil {
			// Fallback to local if service fails
			if dsa.localRepo != nil {
				return dsa.getIntentLocal(ctx, tenantID, intentID)
			}
			return nil, fmt.Errorf("data service failed and no local fallback: %w", err)
		}
		
		return dsa.contractToModel(contractIntent), nil
	}
	
	// Use local repository
	return dsa.getIntentLocal(ctx, tenantID, intentID)
}

// UpdateIntent updates an existing intent
func (dsa *DataServiceAdapter) UpdateIntent(ctx context.Context, tenantID, intentID string, updates *models.IntentUpdate) (*models.Intent, error) {
	if dsa.useService && dsa.dataClient != nil {
		req := &contracts.UpdateIntentRequest{}
		
		if updates.Status != nil {
			contractStatus := dsa.modelToContractStatus(*updates.Status)
			req.Status = &contractStatus
		}
		
		if updates.Tasks != nil {
			req.ParsedTasks = dsa.modelToContractTasks(updates.Tasks)
		}
		
		if updates.Metadata != nil {
			req.Metadata = updates.Metadata
		}
		
		if updates.OverallScore != nil {
			req.OverallScore = updates.OverallScore
		}
		
		if updates.ExecutionTimeMS != nil {
			req.ExecutionTimeMS = updates.ExecutionTimeMS
		}
		
		contractIntent, err := dsa.dataClient.UpdateIntent(ctx, tenantID, intentID, req)
		if err != nil {
			// Fallback to local if service fails
			if dsa.localRepo != nil {
				return dsa.updateIntentLocal(ctx, tenantID, intentID, updates)
			}
			return nil, fmt.Errorf("data service failed and no local fallback: %w", err)
		}
		
		return dsa.contractToModel(contractIntent), nil
	}
	
	// Use local repository
	return dsa.updateIntentLocal(ctx, tenantID, intentID, updates)
}

// Helper methods for local fallback (these would use the existing database code)
func (dsa *DataServiceAdapter) createIntentLocal(ctx context.Context, tenantID, userInput string, metadata map[string]string) (*models.Intent, error) {
	// TODO: Implement using existing database.IntentRepository
	return nil, fmt.Errorf("local intent creation not implemented")
}

func (dsa *DataServiceAdapter) getIntentLocal(ctx context.Context, tenantID, intentID string) (*models.Intent, error) {
	// TODO: Implement using existing database.IntentRepository
	return nil, fmt.Errorf("local intent retrieval not implemented")
}

func (dsa *DataServiceAdapter) updateIntentLocal(ctx context.Context, tenantID, intentID string, updates *models.IntentUpdate) (*models.Intent, error) {
	// TODO: Implement using existing database.IntentRepository
	return nil, fmt.Errorf("local intent update not implemented")
}

// Conversion helpers between models and contracts
func (dsa *DataServiceAdapter) contractToModel(contract *contracts.Intent) *models.Intent {
	model := &models.Intent{
		ID:              contract.ID,
		UserInput:       contract.UserInput,
		Tasks:           make([]models.Task, len(contract.ParsedTasks)),
		Metadata:        contract.Metadata,
		Status:          dsa.contractToModelStatus(contract.Status),
		OverallScore:    contract.OverallScore,
		ExecutionTimeMS: contract.ExecutionTimeMS,
		CreatedAt:       contract.CreatedAt,
		UpdatedAt:       contract.UpdatedAt,
		CompletedAt:     contract.CompletedAt,
	}
	
	// Convert tasks
	for i, contractTask := range contract.ParsedTasks {
		model.Tasks[i] = models.Task{
			ID:           contractTask.ID,
			Type:         dsa.contractToModelTaskType(contractTask.Type),
			Description:  contractTask.Description,
			Dependencies: contractTask.Dependencies,
			Priority:     dsa.contractToModelPriority(contractTask.Priority),
			Metadata:     contractTask.Metadata,
			Status:       dsa.contractToModelTaskStatus(contractTask.Status),
			AgentID:      contractTask.AgentID,
			CreatedAt:    contractTask.CreatedAt,
			CompletedAt:  contractTask.CompletedAt,
		}
	}
	
	return model
}

func (dsa *DataServiceAdapter) modelToContractTasks(tasks []models.Task) []contracts.Task {
	contractTasks := make([]contracts.Task, len(tasks))
	
	for i, task := range tasks {
		contractTasks[i] = contracts.Task{
			ID:           task.ID,
			Type:         dsa.modelToContractTaskType(task.Type),
			Description:  task.Description,
			Dependencies: task.Dependencies,
			Priority:     dsa.modelToContractPriority(task.Priority),
			Metadata:     task.Metadata,
			Status:       dsa.modelToContractTaskStatus(task.Status),
			AgentID:      task.AgentID,
			CreatedAt:    task.CreatedAt,
			CompletedAt:  task.CompletedAt,
		}
	}
	
	return contractTasks
}

// Status conversions
func (dsa *DataServiceAdapter) contractToModelStatus(status contracts.IntentStatus) models.IntentStatus {
	switch status {
	case contracts.IntentStatusPending:
		return models.IntentStatusPending
	case contracts.IntentStatusProcessing:
		return models.IntentStatusProcessing
	case contracts.IntentStatusCompleted:
		return models.IntentStatusCompleted
	case contracts.IntentStatusFailed:
		return models.IntentStatusFailed
	default:
		return models.IntentStatusPending
	}
}

func (dsa *DataServiceAdapter) modelToContractStatus(status models.IntentStatus) contracts.IntentStatus {
	switch status {
	case models.IntentStatusPending:
		return contracts.IntentStatusPending
	case models.IntentStatusProcessing:
		return contracts.IntentStatusProcessing
	case models.IntentStatusCompleted:
		return contracts.IntentStatusCompleted
	case models.IntentStatusFailed:
		return contracts.IntentStatusFailed
	default:
		return contracts.IntentStatusPending
	}
}

// Task type conversions
func (dsa *DataServiceAdapter) contractToModelTaskType(taskType contracts.TaskType) models.TaskType {
	switch taskType {
	case contracts.TaskTypeCodegen:
		return models.TaskTypeCodegen
	case contracts.TaskTypeInfra:
		return models.TaskTypeInfra
	case contracts.TaskTypeDoc:
		return models.TaskTypeDoc
	case contracts.TaskTypeTest:
		return models.TaskTypeTest
	case contracts.TaskTypeAnalyze:
		return models.TaskTypeAnalyze
	default:
		return models.TaskTypeCodegen
	}
}

func (dsa *DataServiceAdapter) modelToContractTaskType(taskType models.TaskType) contracts.TaskType {
	switch taskType {
	case models.TaskTypeCodegen:
		return contracts.TaskTypeCodegen
	case models.TaskTypeInfra:
		return contracts.TaskTypeInfra
	case models.TaskTypeDoc:
		return contracts.TaskTypeDoc
	case models.TaskTypeTest:
		return contracts.TaskTypeTest
	case models.TaskTypeAnalyze:
		return contracts.TaskTypeAnalyze
	default:
		return contracts.TaskTypeCodegen
	}
}

// Task status conversions
func (dsa *DataServiceAdapter) contractToModelTaskStatus(status contracts.TaskStatus) models.TaskStatus {
	switch status {
	case contracts.TaskStatusPending:
		return models.TaskStatusPending
	case contracts.TaskStatusInProgress:
		return models.TaskStatusInProgress
	case contracts.TaskStatusCompleted:
		return models.TaskStatusCompleted
	case contracts.TaskStatusFailed:
		return models.TaskStatusFailed
	case contracts.TaskStatusSkipped:
		return models.TaskStatusSkipped
	default:
		return models.TaskStatusPending
	}
}

func (dsa *DataServiceAdapter) modelToContractTaskStatus(status models.TaskStatus) contracts.TaskStatus {
	switch status {
	case models.TaskStatusPending:
		return contracts.TaskStatusPending
	case models.TaskStatusInProgress:
		return contracts.TaskStatusInProgress
	case models.TaskStatusCompleted:
		return contracts.TaskStatusCompleted
	case models.TaskStatusFailed:
		return contracts.TaskStatusFailed
	case models.TaskStatusSkipped:
		return contracts.TaskStatusSkipped
	default:
		return contracts.TaskStatusPending
	}
}

// Priority conversions
func (dsa *DataServiceAdapter) contractToModelPriority(priority contracts.Priority) models.Priority {
	switch priority {
	case contracts.PriorityHigh:
		return models.PriorityHigh
	case contracts.PriorityMedium:
		return models.PriorityMedium
	case contracts.PriorityLow:
		return models.PriorityLow
	default:
		return models.PriorityMedium
	}
}

func (dsa *DataServiceAdapter) modelToContractPriority(priority models.Priority) contracts.Priority {
	switch priority {
	case models.PriorityHigh:
		return contracts.PriorityHigh
	case models.PriorityMedium:
		return contracts.PriorityMedium
	case models.PriorityLow:
		return contracts.PriorityLow
	default:
		return contracts.PriorityMedium
	}
}

// IntentUpdate represents updates to an intent (this would go in models package)
type IntentUpdate struct {
	Status          *models.IntentStatus      `json:"status,omitempty"`
	ParsedTasks     []models.Task             `json:"parsed_tasks,omitempty"`
	Metadata        map[string]string         `json:"metadata,omitempty"`
	OverallScore    *int                      `json:"overall_score,omitempty"`
	ExecutionTimeMS *int                      `json:"execution_time_ms,omitempty"`
}