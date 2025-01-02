package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"notification-service/src/config"
	"notification-service/src/handlers"
	"notification-service/src/infra/rabbitmq"
	"notification-service/src/pkg/rabbit"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup RabbitMQ connection
	rmq, err := setupRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to setup RabbitMQ: %v", err)
	}

	defer cleanupRabbitMQ(rmq)

	// Monitor RabbitMQ channel and reopen if necessary
	monitorRabbitMQChannel(rmq)

	// Initialize the HTTP server
	initHTTPServer(cfg, rmq, ctx)

	// Start consuming RabbitMQ messages
	rabbit.ConsumeMessage(&cfg.RabbitMQ, rmq, ctx)

	// Wait for shutdown signal
	waitForShutdownSignal(cancel, ctx)
}

// setupRabbitMQ initializes the RabbitMQ connection and channel
func setupRabbitMQ(cfg *config.Config) (*rabbitmq.RabbitMQ, error) {
	rmq, err := rabbitmq.SetupRabbitMq(&cfg.RabbitMQ)
	if err != nil {
		return nil, fmt.Errorf("failed to setup RabbitMQ: %w", err)
	}
	return rmq, nil
}

// cleanupRabbitMQ closes RabbitMQ connections and channels
func cleanupRabbitMQ(rmq *rabbitmq.RabbitMQ) {
	if rmq.Ch != nil {
		rmq.Ch.Close()
	}
	if rmq.Conn != nil {
		rmq.Conn.Close()
	}
	log.Println("RabbitMQ connection and channel closed.")
}

// monitorRabbitMQChannel monitors the RabbitMQ channel and reopens it if it closes
func monitorRabbitMQChannel(rmq *rabbitmq.RabbitMQ) {
	go func() {
		errChan := rmq.Ch.NotifyClose(make(chan *amqp.Error))
		for err := range errChan {
			log.Printf("RabbitMQ channel closed: %v", err)
			newCh, err := rabbitmq.OpenChannel(rmq.Conn)
			if err != nil {
				log.Fatalf("Failed to reopen RabbitMQ channel: %v", err)
			}
			rmq.Ch = newCh
			log.Println("RabbitMQ channel reopened successfully.")
		}
	}()
}

// initHTTPServer initializes the HTTP server and registers handlers
func initHTTPServer(cfg *config.Config, rmq *rabbitmq.RabbitMQ, ctx context.Context) {
	// Register the event handler
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		handlers.EventHandler(cfg, w, r, rmq)
	})

	// Register the WebSocket handler
	http.HandleFunc("/ws", handlers.WSHandler)

	// Start the HTTP server
	server := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Server.Port)}

	go func() {
		log.Printf("HTTP server started on port :%s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()
}

// waitForShutdownSignal handles graceful shutdown when a termination signal is received
func waitForShutdownSignal(cancel context.CancelFunc, ctx context.Context) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Received shutdown signal, canceling context...")
		cancel()
	}()

	<-ctx.Done()
	log.Println("Application shutdown !.")
}
