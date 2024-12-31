package main

import (
	"log"
	"net/http"
	"notification-service/src/consumers"
	"os"
	"os/signal"
	"syscall"

	"notification-service/src/config"
	"notification-service/src/handlers"
	"notification-service/src/infra/rabbitmq"
	"notification-service/src/utils"
)

func main() {
	cfg := config.LoadConfig()

	conn, ch, err := rabbitmq.SetupRabbitMq(cfg)
	if err != nil {
		log.Fatalf("failed to setup rabbitmq: %s", err)
	}

	defer conn.Close()
	defer ch.Close()

	go consumers.ConsumeOrderCreated(ch)
	go consumers.ConsumeOrderCreated(ch)

	go utils.HandleBroadcast()

	//for produce message (for simulate)
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		handlers.EventHandler(w, r, ch)
	})

	// for web socket
	http.HandleFunc("/ws", handlers.WSHandler)

	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("Server started on :8080")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	server.Close()
}
