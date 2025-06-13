package agents

import (
	"context"

	"QLP/internal/models"
)

// Agent is the interface for all specialized agents in the system.
// Each agent is responsible for executing a specific type of task.
type Agent interface {
	Execute(ctx context.Context, task models.Task) (*models.Artifact, error)
}
