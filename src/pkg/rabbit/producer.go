package rabbit

import (
	"log"

	"notification-service/src/infra/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ProduceMessage(rmq *rabbitmq.RabbitMQ, exchange string, routingKey string, body []byte) error {

	err := rmq.Ch.Publish(
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
