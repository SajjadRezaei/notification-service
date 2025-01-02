package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/streadway/amqp"

	"notification-service/src/config"
	"notification-service/src/handlers"
	"notification-service/src/infra/rabbitmq"
	"notification-service/src/pkg/rabbit"
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

	reopenChanelIfNecessary(ch, conn)

	// init server
	initServer(cfg, ch, ctx)

	//start
	rabbit.ConsumeMessage(&cfg.RabbitMQ, ch, ctx)

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
func initServer(cfg *config.Config, ch *amqp.Channel, ctx context.Context) {
	//for simulate notify event for example (user_signup, order_created,and ....)
	http.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		handlers.EventHandler(cfg, w, r, ch)
	})

	// for web socket
	http.HandleFunc("/ws", handlers.WSHandler)

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
func reopenChanelIfNecessary(ch *amqp.Channel, conn *amqp.Connection) {
	go func() {
		errChan := ch.NotifyClose(make(chan *amqp.Error))
		for err := range errChan {
			log.Printf("RabbitMQ channel closed: %v", err)
			newCh, err := rabbitmq.OpenChannel(conn)
			if err != nil {
				log.Fatalf("Failed to recreate RabbitMQ channel: %v", err)
			}
			ch = newCh
		}
	}()
}
