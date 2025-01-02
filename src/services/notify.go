package services

import (
	"log"

	"github.com/streadway/amqp"
	"notification-service/src/pkg/rabbit"
)

// PublishEvent user for produce message
func PublishEvent(ch *amqp.Channel, exchange string, routingKey string, body []byte) error {
	err := rabbit.ProduceMessage(ch, exchange, routingKey, body)

	if err != nil {
		log.Println("Failed to publish a message:", err)
	}

	return err
}
