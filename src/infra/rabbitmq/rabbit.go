package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
	"notification-service/src/config"
)

func SetupRabbitMq(cfg *config.Config) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare a topic exchange
	err = ch.ExchangeDeclare(
		"event_exchange",
		"topic",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	bindings := map[string]string{
		"user_signup":   "user.registered",
		"order_created": "order.created",
	}

	for queue, routingKey := range bindings {
		_, err = ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to declare queue: %w", err)
		}

		err = ch.QueueBind(
			queue,
			routingKey,
			"event_exchange",
			false,
			nil,
		)

		if err != nil {
			return nil, nil, fmt.Errorf("failed to bind queue: %w", err)
		}

	}

	return conn, ch, nil
}
