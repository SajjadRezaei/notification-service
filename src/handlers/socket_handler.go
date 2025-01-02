package handlers

import (
	"log"
	"net/http"

	"notification-service/src/pkg/socket"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SubscribeRequest represents the structure of a client's subscription request
type SubscribeRequest struct {
	EventType string `json:"event_type,omitempty"`
	Action    string `json:"action"`
}

// WSHandler handles WebSocket client requests
func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	defer conn.Close()

	socket.RegisterClient(conn)

	handleClientMessages(conn)
}

// handleClientMessages processes messages received from the WebSocket client
func handleClientMessages(conn *websocket.Conn) {
	for {
		var req SubscribeRequest

		// Read JSON message from the client
		err := conn.ReadJSON(&req)
		if err != nil {
			log.Printf("Failed to read message from client: %v", err)
			socket.UnRegisterClient(conn)
			break
		}

		// Handle the client's action
		if err := handleClientAction(conn, req); err != nil {
			log.Printf("Error handling client action: %v", err)
			break
		}
	}
}

// handleClientAction processes a single client action based on the request
func handleClientAction(conn *websocket.Conn, req SubscribeRequest) error {
	switch req.Action {
	case "subscribe":
		socket.SubscribeToEvent(conn, req.EventType)
		log.Printf("Client subscribed to event: %s", req.EventType)
	case "unsubscribe":
		socket.UnSubscribeFromEvent(conn, req.EventType)
		log.Printf("Client unsubscribed from event: %s", req.EventType)
	case "ping":
		directMessageToClient(conn, []byte("pong!"))
		return nil
	default:
		log.Printf("Invalid action received: %s", req.Action)
		directMessageToClient(conn, []byte("Invalid action: must be 'subscribe', 'unsubscribe', or 'ping      x"))

	}
	return nil
}

// directMessageToClient sends a "pong" response to the client for a "ping" action
func directMessageToClient(conn *websocket.Conn, message []byte) {
	err := socket.DirectMessageToClient(conn, message)
	if !err {
		log.Printf("Failed to send pong response: %v", err)
	}
}
