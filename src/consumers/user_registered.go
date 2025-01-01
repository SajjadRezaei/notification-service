package consumers

import (
	"context"
	"log"

	"notification-service/src/config"
	"notification-service/src/utils"

	"github.com/streadway/amqp"
)

// ConsumeMessage dynamically consumes messages for all queues in the RoutingKeyMap
func ConsumeMessage(cfg *config.RabbitMQConfig, ch *amqp.Channel, ctx context.Context) {
	for topic, routingKey := range cfg.RoutingKeyMap {
		go func(queue, routingKey string) {
			log.Printf("Starting consumer for queue: %s, routing key: %s", queue, routingKey)

			// Start consuming messages from the queue
			mags, err := ch.Consume(
				topic, // Queue name
				"",    // Consumer tag (auto-generated if empty)
				false, // Auto-ack (set to false for manual ack)
				false, // Exclusive
				false, // No-local
				false, // No-wait
				nil,   // Arguments
			)

			if err != nil {
				log.Fatalf("Failed to start consuming messages from queue %s: %v", topic, err)
			}

			for {
				select {
				case msg := <-mags:
					processMessage(topic, msg, err)
				case <-ctx.Done():
					log.Printf("Stopping consumer for queue: %s", topic)
					return
				}
			}
		}(topic, routingKey)
	}
}

// processMessage handles the processing of a single message
func processMessage(topic string, msg amqp.Delivery, err error) {
	if len(msg.Body) > 0 {
		log.Printf("Received message from topic %s: %s", topic, string(msg.Body))

		success := utils.BroadcastMessage(topic, msg.Body)

		if success {
			// Acknowledge the message if it was successfully sent
			if err = msg.Ack(false); err != nil {
				log.Printf("Failed to acknowledge message: %v", err)
			}
		} else {
			if err = msg.Nack(false, false); err != nil {
				log.Printf("Failed to reject message: %v", err)
			}
		}
	}
}
