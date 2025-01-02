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

type SubscribeRequest struct {
	EventType string `json:"event_type,omitempty"`
	Action    string `json:"action"`
}

// WSHandler handle socket client request
func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	socket.RegisterClient(conn)

	for {
		var req SubscribeRequest
		err = conn.ReadJSON(&req)
		if err != nil {
			log.Println("Failed to read message from client", err)
			socket.UnRegisterClient(conn)
			break
		}

		switch req.Action {
		case "subscribe":
			socket.SubscribeToEvent(conn, req.EventType)
		case "unsubscribe":
			socket.UnSubscribeFromEvent(conn, req.EventType)
		case "ping":
			socket.DirectMessageToClient(conn, []byte("pong!"))
		default:
			log.Printf("Invalid action: %s\n", req.Action)
			err = conn.WriteMessage(websocket.TextMessage, []byte("Invalid action: must be 'subscribe' or 'unsubscribe'"))
			if err != nil {
				log.Printf("Failed to send error message to client: %v\n", err)
			}
		}
	}
}
