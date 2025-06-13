package dag

import (
	"fmt"
	"sync"

	"QLP/internal/models"
)

// Node represents a single node in the DAG
type Node struct {
	Task     models.Task
	outEdges map[string]*Node // Nodes this node points to
	inEdges  map[string]*Node // Nodes pointing to this node
}

// DAG represents a directed acyclic graph of tasks.
type DAG struct {
	nodes map[string]*Node
	lock  sync.RWMutex
}

// NewDAG creates a new, empty DAG.
func NewDAG() *DAG {
	return &DAG{
		nodes: make(map[string]*Node),
	}
}

// AddTask adds a task to the DAG as a node.
func (d *DAG) AddTask(task models.Task) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, exists := d.nodes[task.ID]; !exists {
		d.nodes[task.ID] = &Node{
			Task:     task,
			outEdges: make(map[string]*Node),
			inEdges:  make(map[string]*Node),
		}
	}
}

// AddEdge adds a directed edge between two tasks in the DAG.
func (d *DAG) AddEdge(fromID, toID string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	fromNode, ok := d.nodes[fromID]
	if !ok {
		return fmt.Errorf("from node '%s' not found in DAG", fromID)
	}

	toNode, ok := d.nodes[toID]
	if !ok {
		return fmt.Errorf("to node '%s' not found in DAG", toID)
	}

	fromNode.outEdges[toID] = toNode
	toNode.inEdges[fromID] = fromNode

	return nil
}

// GetReadyTasks returns all tasks that have no incoming dependencies.
// These are the tasks that can be executed immediately.
func (d *DAG) GetReadyTasks() []models.Task {
	d.lock.RLock()
	defer d.lock.RUnlock()

	var readyTasks []models.Task
	for _, node := range d.nodes {
		if len(node.inEdges) == 0 {
			readyTasks = append(readyTasks, node.Task)
		}
	}
	return readyTasks
}

// MarkTaskComplete marks a task as complete and removes it from the graph,
// updating the dependencies of its children.
func (d *DAG) MarkTaskComplete(taskID string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	node, exists := d.nodes[taskID]
	if !exists {
		return // Or log a warning
	}

	// For each child of the completed node, remove the incoming edge from the completed node.
	for _, childNode := range node.outEdges {
		delete(childNode.inEdges, taskID)
	}

	// Remove the completed node from the graph.
	delete(d.nodes, taskID)
}

// IsEmpty returns true if the DAG has no nodes.
func (d *DAG) IsEmpty() bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return len(d.nodes) == 0
}
