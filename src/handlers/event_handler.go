package handlers

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"net/http"
	"notification-service/src/services"
)

func EventHandler(w http.ResponseWriter, r *http.Request, ch *amqp.Channel) {
	var req struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.EventType == "" {
		http.Error(w, "EventType is required", http.StatusBadRequest)
		return
	}

	if err := services.PublishEvent(ch, req.EventType, req.Payload); err != nil {
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("event published"))
	if err != nil {
		return
	}
}
