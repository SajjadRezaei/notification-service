package utils

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

var (
	clients   = make(map[*websocket.Conn]map[string]bool)
	broadcast = make(chan []byte)
	mu        = new(sync.Mutex)
)

type BroadcastRequest struct {
	EventType string `json:"event_type"`
	Message   []byte `json:"message"`
}

func RegisterClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	clients[conn] = make(map[string]bool)
	log.Println("New WebSocket Client Connected:", conn.RemoteAddr())
}

func UnRegisterClient(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	delete(clients, conn)
	conn.Close()
	log.Println("WebSocket Client disconnected:")
}

func SubscribeToEvent(conn *websocket.Conn, eventType string) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := clients[conn]; !exists {
		clients[conn][eventType] = true
		log.Printf("Clients Subscribe To Event  %s", eventType)
	}
}

func UnSubscribeFromEvent(conn *websocket.Conn, eventType string) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := clients[conn]; !exists {
		delete(clients[conn], eventType)
		log.Printf("Client Unsubscribe from Event %s", eventType)
	}
}

func BroadcastMessage(eventType string, message []byte) {
	mu.Lock()
	defer mu.Unlock()

	for client, subscription := range clients {
		if _, subscribed := subscription[eventType]; subscribed {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to send message to client: %v", err)
				UnRegisterClient(client)
			}
		}
	}

	broadcast <- message
}

func HandleBroadcast() {
	for {
		msg := <-broadcast
		mu.Lock()
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("Failed to broadcast message to clients: %v\n", err)
				UnRegisterClient(client)
			}
		}
		mu.Unlock()
	}
}