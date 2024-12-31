package consumers

import (
	"github.com/streadway/amqp"
	"log"
	"notification-service/src/utils"
)

func ConsumeUserRegistered(ch *amqp.Channel) {
	msgs, err := ch.Consume(
		"user_signup,",
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
		utils.BroadcastMessage("", msg.Body)
	}
}
