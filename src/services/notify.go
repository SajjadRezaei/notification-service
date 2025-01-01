package services

import (
	"github.com/streadway/amqp"
	"log"
)

func PublishEvent(ch *amqp.Channel, exchange string, routingKey string, body []byte) error {
	err := ch.Publish(
		exchange,
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
