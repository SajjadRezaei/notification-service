package services

import (
	"log"

	"notification-service/src/infra/rabbitmq"
	"notification-service/src/pkg/rabbit"
)

// PublishEvent user for produce message
func PublishEvent(rmq *rabbitmq.RabbitMQ, exchange string, routingKey string, body []byte) error {
	err := rabbit.ProduceMessage(rmq, exchange, routingKey, body)

	if err != nil {
		log.Println("Failed to publish a message:", err)
	}

	return err
}
