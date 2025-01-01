package consumers

import (
	"context"
	"log"

	"notification-service/src/utils"

	"github.com/streadway/amqp"
)

func ConsumeUserRegistered(ch *amqp.Channel, ctx context.Context) {
	msgs, err := ch.Consume(
		"user_signup",
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	for msg := range msgs {
		log.Printf("Received user_signup event: %s", msg.Body)

		// Attempt to broadcast the message to WebSocket clients
		success := utils.BroadcastMessage("user_signup", msg.Body)

		if success {
			// Acknowledge the message if it was successfully sent
			if err = msg.Ack(false); err != nil {
				log.Printf("Failed to acknowledge message: %v", err)
			}
		} else {
			// Reject the message and requeue it for another attempt
			if err = msg.Nack(false, true); err != nil {
				log.Printf("Failed to reject message: %v", err)
			}
		}
	}
}
