package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"notification-service/src/config"
	"notification-service/src/infra/rabbitmq"
	"notification-service/src/services"
)

// EventHandler give user request and send message to rabbit
func EventHandler(cfg *config.Config, w http.ResponseWriter, r *http.Request, rmq *rabbitmq.RabbitMQ) {
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

	reqBytes := new(bytes.Buffer)
	json.NewEncoder(reqBytes).Encode(req)

	if err := services.PublishEvent(rmq, cfg.RabbitMQ.Exchange, routingKey, reqBytes.Bytes()); err != nil {
		http.Error(w, "Failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("event published"))
	if err != nil {
		return
	}
}
