package statemanager

import (
	"sync"

	"QLP/services/orchestrator-service/internal/dag"
)

// StateManager defines the interface for managing the state of execution graphs (DAGs).
type StateManager interface {
	Get(intentID string) (*dag.DAG, bool)
	Set(intentID string, graph *dag.DAG)
	Delete(intentID string)
}

// InMemoryStateManager is a thread-safe, in-memory implementation of StateManager.
// NOTE: This is suitable for development but not for production. A persistent
// store like Redis or Postgres should be used in a production environment.
type InMemoryStateManager struct {
	graphs map[string]*dag.DAG
	lock   sync.RWMutex
}

// NewInMemoryStateManager creates a new in-memory state manager.
func NewInMemoryStateManager() *InMemoryStateManager {
	return &InMemoryStateManager{
		graphs: make(map[string]*dag.DAG),
	}
}

// Get retrieves a DAG by its intent ID.
func (s *InMemoryStateManager) Get(intentID string) (*dag.DAG, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	graph, found := s.graphs[intentID]
	return graph, found
}

// Set stores a DAG with its corresponding intent ID.
func (s *InMemoryStateManager) Set(intentID string, graph *dag.DAG) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.graphs[intentID] = graph
}

// Delete removes a DAG from the state manager.
func (s *InMemoryStateManager) Delete(intentID string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.graphs, intentID)
}
