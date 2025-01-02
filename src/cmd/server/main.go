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

	cfg := config.LoadConfig()
	ctx, cancel := context.WithCancel(context.Background())

	rmq, err := rabbitmq.SetupRabbitMq(&cfg.RabbitMQ)
	if err != nil {
		log.Fatalf("failed to setup rabbitmq: %s", err)
	}

	defer rmq.Conn.Close()
	defer rmq.Ch.Close()

	reopenChanelIfNecessary(rmq)

	// init server
	initServer(cfg, rmq, ctx)

	//start
	rabbit.ConsumeMessage(&cfg.RabbitMQ, rmq, ctx)

	waitForShutdownSignal(cancel, ctx)
}

// waitForShutdownSignal for handle Graceful Shutdown
func waitForShutdownSignal(cancel context.CancelFunc, ctx context.Context) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Received shutdown signal, canceling context...")
		cancel()
	}()
	<-ctx.Done()
}

// initServer initialize http server for get event and handle socket web server
func initServer(cfg *config.Config, rmq *rabbitmq.RabbitMQ, ctx context.Context) {
	//for simulate notify event for example (user_signup, order_created,and ....)
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		handlers.EventHandler(cfg, w, r, rmq)
	})

	// for web socket
	go http.HandleFunc("/ws", handlers.WSHandler)

	server := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Server.Port)}

	go func() {
		log.Printf("Server started on :%s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil {
			ctx.Done()
			log.Fatalf("Server failed: %v", err)
		}
	}()
}

// reopenChanelIfNecessary monitor channel closure and recreate if necessary
func reopenChanelIfNecessary(rmq *rabbitmq.RabbitMQ) {
	go func() {
		errChan := rmq.Ch.NotifyClose(make(chan *amqp.Error))
		for err := range errChan {
			log.Printf("RabbitMQ channel closed: %v", err)
			newCh, err := rabbitmq.OpenChannel(rmq.Conn)
			if err != nil {
				log.Fatalf("Failed to recreate RabbitMQ channel: %v", err)
			}
			rmq.Ch = newCh
		}
	}()
}
