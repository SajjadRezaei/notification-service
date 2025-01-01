package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"

	"notification-service/src/config"
)

func SetupRabbitMq(cfg *config.RabbitMQConfig) (*amqp.Connection, *amqp.Channel, error) {

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	fmt.Println(url)

	var conn *amqp.Connection
	var err error
	// Retry logic
	for i := 0; i < 10; i++ { // Retry 10 times with a delay
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d): %s", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare a topic exchange
	err = ch.ExchangeDeclare(
		cfg.Exchange,
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

	err = bindQueuesToExchange(cfg, err, ch)

	if err != nil {
		return nil, nil, err
	}

	return conn, ch, nil
}

func bindQueuesToExchange(cfg *config.RabbitMQConfig, err error, ch *amqp.Channel) error {
	for queue, routingKey := range cfg.RoutingKeyMap {
		_, err = ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			fmt.Errorf("failed to declare queue: %w", err)
			return err
		}

		err = ch.QueueBind(
			queue,
			routingKey,
			cfg.Exchange,
			false,
			nil,
		)

		return err

	}

	return nil
}
