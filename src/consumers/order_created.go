package consumers

import (
	"context"
	"fmt"
	"log"
	"notification-service/src/utils"

	"github.com/streadway/amqp"
)

func ConsumeOrderCreated(ch *amqp.Channel, ctx context.Context) error {

	// Set prefetch count to limit unacknowledged messages
	err := ch.Qos(
		10,    // Prefetch count
		0,     // Prefetch size
		false, // Apply to the entire channel
	)

	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Register the consumer
	msgs, err := ch.Consume(
		"order_created", // Queue name
		"",              // Consumer tag (empty for auto-generated)
		true,            // Auto-acknowledge
		false,           // Exclusive
		false,           // No-local (deprecated, usually false)
		false,           // No-wait
		nil,             // Arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Create a channel to handle messages
	msgChan := make(chan amqp.Delivery)

	// Start a goroutine to process messages
	go func() {
		for msg := range msgs {
			select {
			case <-ctx.Done():
				log.Println("Context canceled, stopping message processing")
				close(msgChan)
				return
			case msgChan <- msg:
			}
		}
	}()

	// Process messages
	for {
		select {
		case <-ctx.Done():
			log.Println("Context canceled, stopping consumer")
			return nil
		case msg, ok := <-msgChan:
			if !ok {
				log.Println("Message channel closed, stopping consumer")
				return nil
			}

			// Process the message
			log.Printf("Received order_created event: %s", msg.Body)
			utils.BroadcastMessage("order_created", msg.Body)

			// Add additional processing logic here if needed
		}
	}
}
