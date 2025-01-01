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

	// for handle not processable message (when server get message and send to rabbit but client not ready for process)
	err = ch.ExchangeDeclare(
		cfg.DLXExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to declare del_exchange: %w", err)
	}

	err = bindQueuesToExchange(cfg, err, ch)

	if err != nil {
		return nil, nil, err
	}

	return conn, ch, nil
}

func OpenChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	for i := 0; i < 5; i++ {
		ch, err := conn.Channel()
		if err == nil {
			return ch, nil
		}
		log.Printf("Failed to open channel (attempt %d): %s", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to open channel after retries")
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

		args := amqp.Table{
			"x-dead-letter-exchange":    cfg.DLXExchange,     // Dead Letter Exchange
			"x-dead-letter-routing-key": routingKey + "_dlq", // Routing key for the DLQ
			"x-message-ttl":             10000,               // Optional: Time-to-live for retries (in milliseconds)
		}

		_, err = ch.QueueDeclare(
			queue+"_dlq",
			true,
			false,
			false,
			false,
			args,
		)

		if err != nil {
			fmt.Errorf("failed to declare  dlq queue: %w", err)
			return err
		}

		err = ch.QueueBind(
			queue,
			routingKey,
			cfg.Exchange,
			false,
			nil,
		)

		err = ch.QueueBind(
			queue+"_dlq",
			routingKey+"_dlq",
			cfg.DLXExchange,
			false,
			nil,
		)

		return err

	}

	return nil
}
