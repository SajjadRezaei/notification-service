package main

import (
	"context"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"notification-service/src/config"
	"os"
	"os/signal"
	"syscall"

	//"notification-service/src/config"
	"notification-service/src/handlers"
	"notification-service/src/infra/rabbitmq"
)

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	conn, ch, err := rabbitmq.SetupRabbitMq(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("failed to setup rabbitmq: %s", err)
	}

	defer conn.Close()
	defer ch.Close()

	//go utils.HandleBroadcast()

	initServer(cfg, ch, ctx)

	// Run the consumer in a goroutine
	//go func() {
	//	log.Println("Starting message consumer...")
	//	err := consumers.ConsumeOrderCreated(ch, ctx)
	//	if err != nil {
	//		log.Printf("Error consuming messages: %v", err)
	//	}
	//}()

	// Handle OS signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Received shutdown signal, canceling context...")
		cancel()
	}()
	<-ctx.Done()
	//
}

func initServer(cfg *config.Config, ch *amqp.Channel, ctx context.Context) {
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		handlers.EventHandler(cfg, w, r, ch)
	})

	// for web socket
	http.HandleFunc("/ws", handlers.WSHandler)

	server := &http.Server{Addr: ":8080"}

	go func() {
		log.Println("Server started on :8080")
		if err := server.ListenAndServe(); err != nil {
			ctx.Done()
			log.Fatalf("Server failed: %v", err)
		}
	}()
}

func getConfig() {

}
