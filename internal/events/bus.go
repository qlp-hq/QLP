package events

import (
	"context"
	"log"
	"sync"
	"time"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
}

type EventType string

const (
	EventTaskCreated   EventType = "task.created"
	EventTaskStarted   EventType = "task.started"
	EventTaskCompleted EventType = "task.completed"
	EventTaskFailed    EventType = "task.failed"
	EventAgentSpawned  EventType = "agent.spawned"
	EventAgentStopped  EventType = "agent.stopped"
)

type Handler func(ctx context.Context, event Event) error

type EventBus struct {
	handlers map[EventType][]Handler
	mu       sync.RWMutex
	events   chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]Handler),
		events:   make(chan Event, 1000),
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler Handler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
	select {
	case eb.events <- event:
	default:
		log.Printf("Event bus full, dropping event: %s", event.ID)
	}
}

func (eb *EventBus) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-eb.events:
				eb.handleEvent(ctx, event)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (eb *EventBus) handleEvent(ctx context.Context, event Event) {
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	for _, handler := range handlers {
		go func(h Handler) {
			if err := h(ctx, event); err != nil {
				log.Printf("Handler error for event %s: %v", event.ID, err)
			}
		}(handler)
	}
}
