package handlers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"notification-service/src/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SubscribeRequest struct {
	EventType string `json:"event_type"`
	Action    string `json:"action"`
}

func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	utils.RegisterClient(conn)

	for {
		var req SubscribeRequest
		err := conn.ReadJSON(&req)
		if err != nil {
			log.Println("Failed to read message from client", err)
			utils.UnRegisterClient(conn)
			break
		}

		switch req.Action {
		case "subscribe":
			utils.SubscribeToEvent(conn, req.EventType)
		case "unsubscribe":
			utils.UnSubscribeFromEvent(conn, req.EventType)
		}

	}
}
