package services

import "github.com/streadway/amqp"

func PublishEvent(ch *amqp.Channel, routingKey string, body []byte) error {
	return ch.Publish(
		"events_exchange",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
