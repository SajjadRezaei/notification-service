package rabbitmq

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"notification-service/src/config"
)

type RabbitMQ struct {
	Conn *amqp.Connection
	Ch   *amqp.Channel
}

// SetupRabbitMq setup rabbitMq
func SetupRabbitMq(cfg *config.RabbitMQConfig) (*RabbitMQ, error) {

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
		return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
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
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// for handle not processable message (when server get message and send to socket but client not ready for process)
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
		return nil, fmt.Errorf("failed to declare del_exchange: %w", err)
	}

	err = bindQueuesToExchange(cfg, ch)

	if err != nil {
		return nil, err
	}

	return &RabbitMQ{
		Conn: conn,
		Ch:   ch,
	}, nil
}

// OpenChannel check chanel is close try to open chanel
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

func bindQueuesToExchange(cfg *config.RabbitMQConfig, ch *amqp.Channel) error {
	for _, queue := range cfg.ServiceToQueue {

		routingKeys := cfg.QueueToRoutingKeys[queue]

		args := amqp.Table{
			"x-dead-letter-exchange": cfg.DLXExchange,
		}

		dlQueue := fmt.Sprintf("%s_dlq", queue)
		_, err := ch.QueueDeclare(
			dlQueue,
			true,
			false,
			false,
			false,
			nil,
		)

		if err != nil {
			return fmt.Errorf("failed to declare  dlq queue: %w", err)
		}

		// main queue per service
		_, err = ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			args,
		)

		if err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		for _, routingKey := range routingKeys {
			err = ch.QueueBind(
				queue,
				routingKey,
				cfg.Exchange,
				false,
				nil,
			)

			if err != nil {
				return fmt.Errorf("failed to bind main queue: %w", err)
			}

			err = ch.QueueBind(
				dlQueue,
				routingKey,
				cfg.DLXExchange,
				false,
				nil,
			)

			if err != nil {
				return fmt.Errorf("failed to bind queue to dead letter queue %w", err)
			}
		}
	}

	return nil
}

func handelDeadLetterQueue(ch *amqp.Channel, queue string) error {
	dlQueue := fmt.Sprintf("%s_dlq", queue)
	_, err := ch.QueueDeclare(
		dlQueue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare  dlq queue: %w", err)
	}

	return nil
}
