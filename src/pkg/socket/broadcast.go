package socket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	clients = make(map[*websocket.Conn]map[string]bool)
	mu      sync.Mutex
)

// BroadcastRequest represents the structure of a broadcast message
type BroadcastRequest struct {
	EventType string `json:"event_type"`
	Message   []byte `json:"message"`
}

// RegisterClient registers a new WebSocket client
func RegisterClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	clients[conn] = make(map[string]bool)
	log.Printf("New WebSocket client connected: %s", conn.RemoteAddr())
}

// UnRegisterClient unregisters a WebSocket client and closes its connection
func UnRegisterClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	delete(clients, conn)
	conn.Close()
	log.Printf("WebSocket client disconnected: %s", conn.RemoteAddr())
}

// SubscribeToEvent subscribes a client to a specific event type
func SubscribeToEvent(conn *websocket.Conn, eventType string) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := clients[conn]; !exists {
		clients[conn] = make(map[string]bool)
	}

	clients[conn][eventType] = true
	log.Printf("Client subscribed to event: %s", eventType)
}

// UnSubscribeFromEvent unsubscribes a client from a specific event type
func UnSubscribeFromEvent(conn *websocket.Conn, eventType string) {
	mu.Lock()
	defer mu.Unlock()

	if subscriptions, exists := clients[conn]; exists {
		delete(subscriptions, eventType)

		if len(subscriptions) == 0 {
			delete(clients, conn)
		}
		log.Printf("Client unsubscribed from event: %s", eventType)
	}
}

// BroadcastMessage sends a message to all clients subscribed to a specific event type
func BroadcastMessage(topic string, message []byte) bool {
	mu.Lock()
	defer mu.Unlock()

	log.Printf("Broadcasting message to clients subscribed to event: %s", topic)

	var clientMessage struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(message, &clientMessage); err != nil {
		log.Printf("Failed to decode message: %v", err)
		return false
	}

	success := false

	for client, subscriptions := range clients {
		if _, subscribed := subscriptions[clientMessage.EventType]; subscribed {
			if err := SendMessageToClient(client, clientMessage.Payload); err != nil {
				log.Printf("Failed to send message to client: %v", err)
				UnRegisterClient(client)
			} else {
				success = true
			}
		}
	}

	return success
}

// SendMessageToClient sends a direct message to a specific client
func SendMessageToClient(client *websocket.Conn, message []byte) error {
	err := client.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}
