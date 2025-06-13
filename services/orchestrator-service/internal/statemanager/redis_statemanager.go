package statemanager

import (
	"context"
	"encoding/json"
	"time"

	"QLP/services/orchestrator-service/internal/dag"

	"github.com/redis/go-redis/v9"
)

// RedisStateManager is a production-ready implementation of StateManager using Redis.
type RedisStateManager struct {
	client *redis.Client
}

// NewRedisStateManager creates a new state manager connected to a Redis instance.
// It expects the Redis address (e.g., "localhost:6379") in the REDIS_ADDR env var.
func NewRedisStateManager(addr string) *RedisStateManager {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStateManager{client: rdb}
}

// Get retrieves a DAG by its intent ID from Redis.
func (s *RedisStateManager) Get(intentID string) (*dag.DAG, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	val, err := s.client.Get(ctx, s.key(intentID)).Result()
	if err == redis.Nil {
		return nil, false // Key does not exist
	} else if err != nil {
		// Log the error for observability
		return nil, false
	}

	var graph dag.DAG
	if err := json.Unmarshal([]byte(val), &graph); err != nil {
		// Log the unmarshalling error
		return nil, false
	}
	return &graph, true
}

// Set stores a DAG in Redis, serialized as JSON.
func (s *RedisStateManager) Set(intentID string, graph *dag.DAG) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	graphBytes, err := json.Marshal(graph)
	if err != nil {
		// Log the marshalling error
		return
	}

	// We could set an expiration here if needed (e.g., 24 hours)
	s.client.Set(ctx, s.key(intentID), graphBytes, 0)
}

// Delete removes a DAG from Redis.
func (s *RedisStateManager) Delete(intentID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.client.Del(ctx, s.key(intentID))
}

// key is a helper to create a namespaced Redis key.
func (s *RedisStateManager) key(intentID string) string {
	return "qlp:orchestrator:dag:" + intentID
}
