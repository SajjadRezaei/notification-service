package rabbit

import (
	"context"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"notification-service/src/config"
	"notification-service/src/infra/rabbitmq"
	"notification-service/src/pkg/socket"
)

// ConsumeMessage dynamically consumes messages for all queues in separate goroutines
func ConsumeMessage(cfg *config.RabbitMQConfig, rmq *rabbitmq.RabbitMQ, ctx context.Context) {
	for _, queueName := range cfg.ServiceToQueue {
		go startConsumer(queueName, rmq, ctx)
	}
}

// startConsumer starts a consumer for a specific queue
func startConsumer(queueName string, rmq *rabbitmq.RabbitMQ, ctx context.Context) {
	log.Printf("Starting consumer for queue: %s", queueName)

	messages, err := rmq.Ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to start consuming messages from queue %s: %v", queueName, err)
	}

	for {
		select {
		case msg := <-messages:
			if len(msg.Body) > 0 {
				handleMessage(queueName, msg)
			}
		case <-ctx.Done():
			log.Printf("Stopping consumer for queue: %s", queueName)
			return
		}
	}
}

// handleMessage processes a single message from the queue
func handleMessage(queueName string, msg amqp.Delivery) {
	log.Printf("Received message from queue %s: %s", queueName, string(msg.Body))

	if socket.BroadcastMessage(queueName, msg.Body) {
		if err := msg.Ack(false); err != nil {
			log.Printf("Failed to acknowledge message: %v", err)
		}
	} else {
		// Reject the message and send it to the Dead Letter Queue
		if err := msg.Nack(false, false); err != nil {
			log.Printf("Failed to reject message: %v", err)
		}
	}
}
