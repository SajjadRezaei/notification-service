package handlers

import (
	"encoding/json"
	"net/http"

	"notification-service/src/config"
	"notification-service/src/services"

	"github.com/streadway/amqp"
)

// EventHandler give user request and send message to rabbit
func EventHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request, ch *amqp.Channel) {
	var req struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	routingKey, exist := cfg.RabbitMQ.EventToRoutingKey[req.EventType]
	if req.EventType == "" || !exist {
		http.Error(w, "EventType is invalid", http.StatusBadRequest)
		return
	}

	if err := services.PublishEvent(ch, cfg.RabbitMQ.Exchange, routingKey, req.Payload); err != nil {
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("event published"))
	if err != nil {
		return
	}
}
