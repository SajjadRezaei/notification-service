package services

import (
	"github.com/streadway/amqp"
	"log"
)

func PublishEvent(ch *amqp.Channel, routingKey string, body []byte) error {
	err := ch.Publish(
		"event_exchange",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Println("Failed to publish a message:", err)
	}

	return err
}
