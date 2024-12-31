package consumers

import (
	"log"
	"notification-service/src/utils"

	"github.com/streadway/amqp"
)

func ConsumeOrderCreated(ch *amqp.Channel) {
	msgs, err := ch.Consume(
		"order_created",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to register a consumer: %s", err)
	}

	for msg := range msgs {
		log.Printf("user register event: %s", msg.Body)
		utils.BroadcastMessage("created_order", msg.Body)
	}
}
