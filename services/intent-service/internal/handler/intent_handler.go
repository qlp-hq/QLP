package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"QLP/internal/events"
	"QLP/internal/llm"
	"QLP/services/intent-service/internal/parser"
)

// IntentRequest is the expected structure of a request to the intent endpoint.
type IntentRequest struct {
	Query string `json:"query"`
}

// IntentResponse is the structure of a successful response.
type IntentResponse struct {
	IntentID   string    `json:"intent_id"`
	ReceivedAt time.Time `json:"received_at"`
}

// IntentHandler holds the dependencies for the intent HTTP handler.
type IntentHandler struct {
	parser       *parser.IntentParser
	eventManager events.Manager
}

// NewIntentHandler creates a new IntentHandler with its dependencies.
func NewIntentHandler(llmClient llm.Client, eventManager events.Manager) *IntentHandler {
	return &IntentHandler{
		parser:       parser.NewIntentParser(llmClient),
		eventManager: eventManager,
	}
}

// Handle is the HTTP handler for creating a new intent.
func (h *IntentHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Query == "" {
		http.Error(w, "Query cannot be empty", http.StatusBadRequest)
		return
	}

	// The parsing can be time-consuming, so we use the request context.
	ctx := r.Context()
	intent, err := h.parser.ParseIntent(ctx, req.Query)
	if err != nil {
		log.Printf("ERROR: Failed to parse intent: %v", err)
		http.Error(w, "Failed to parse intent", http.StatusInternalServerError)
		return
	}

	// Create the event payload
	payload, err := json.Marshal(intent)
	if err != nil {
		log.Printf("ERROR: Failed to marshal intent for event: %v", err)
		http.Error(w, "Failed to process intent", http.StatusInternalServerError)
		return
	}

	// Publish the event to Kafka
	event := events.Event{
		ID:        intent.ID,
		Type:      "intent.received",
		Timestamp: time.Now(),
		Source:    "intent-service",
		Payload:   payload,
	}

	if err := h.eventManager.Publish(ctx, event); err != nil {
		log.Printf("ERROR: Failed to publish intent.received event: %v", err)
		http.Error(w, "Failed to publish intent event", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully processed and published intent %s", intent.ID)

	// Respond to the client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(IntentResponse{
		IntentID:   intent.ID,
		ReceivedAt: time.Now(),
	})
}
