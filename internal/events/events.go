package events

import (
	"context"
	"encoding/json"
	"time"
)

// EventType is a string that defines the type of an event.
type EventType string

// Event represents a single, discrete event in the system.
type Event struct {
	ID        string          `json:"id"`
	Type      EventType       `json:"type"`
	Source    string          `json:"source"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// EventHandler is a function that processes an event.
type EventHandler func(ctx context.Context, event Event) error

// Manager is the interface for an event bus system.
// It allows for publishing events and subscribing to topics.
type Manager interface {
	// Publish sends an event to the event bus.
	Publish(ctx context.Context, event Event) error

	// Subscribe listens for events of a specific type.
	// The handler is called for each event received.
	// The underlying implementation runs the listener in a background goroutine.
	Subscribe(ctx context.Context, eventType EventType, handler EventHandler) error

	// Close gracefully shuts down the connection to the event bus.
	Close() error
}
