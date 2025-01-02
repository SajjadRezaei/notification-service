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

const (
	MainExchangeType = "topic"
	DedExchangeType  = "direct"
)

// SetupRabbitMq initializes and sets up RabbitMQ with retries for connection and channel creation.
func SetupRabbitMq(cfg *config.RabbitMQConfig) (*RabbitMQ, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	log.Printf("Connecting to RabbitMQ at %s", url)

	// Establish connection with retry logic
	conn, err := connectWithRetry(url, 10, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchanges
	if err := declareExchanges(cfg, ch); err != nil {
		conn.Close()
		return nil, err
	}

	// Bind queues to exchanges
	if err := bindQueuesToExchange(cfg, ch); err != nil {
		conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		Conn: conn,
		Ch:   ch,
	}, nil
}

// OpenChannel ensures a channel is open, retrying if necessary.
func OpenChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	var ch *amqp.Channel
	var err error
	for i := 0; i < 5; i++ {
		ch, err = conn.Channel()
		if err == nil {
			return ch, nil
		}
		log.Printf("Failed to open channel (attempt %d): %s", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to open channel after retries")
}

// connectWithRetry attempts to connect to RabbitMQ with retries.
func connectWithRetry(url string, maxRetries int, delay time.Duration) (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			return conn, nil
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %s", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("unable to connect to RabbitMQ after %d retries: %w", maxRetries, err)
}

// declareExchanges declares the necessary exchanges for RabbitMQ.
func declareExchanges(cfg *config.RabbitMQConfig, ch *amqp.Channel) error {
	// Declare main exchange
	if err := ch.ExchangeDeclare(
		cfg.Exchange,
		MainExchangeType,
		false,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare main exchange '%s': %w", cfg.Exchange, err)
	}

	// Declare dead-letter exchange
	if err := ch.ExchangeDeclare(
		cfg.DLXExchange,
		DedExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare dead-letter exchange '%s': %w", cfg.DLXExchange, err)
	}

	return nil
}

// bindQueuesToExchange binds queues to the main and dead-letter exchanges.
func bindQueuesToExchange(cfg *config.RabbitMQConfig, ch *amqp.Channel) error {
	for _, queue := range cfg.ServiceToQueue {
		if err := setupQueueBindings(cfg, ch, queue); err != nil {
			return err
		}
	}
	return nil
}

// setupQueueBindings sets up the main queue and its dead-letter queue with bindings.
func setupQueueBindings(cfg *config.RabbitMQConfig, ch *amqp.Channel, queue string) error {
	// Dead-letter queue
	dlQueue := fmt.Sprintf("%s_dlq", queue)
	if err := declareQueue(ch, dlQueue, nil); err != nil {
		return fmt.Errorf("failed to declare dead-letter queue '%s': %w", dlQueue, err)
	}

	// Main queue with dead-letter exchange binding
	args := amqp.Table{"x-dead-letter-exchange": cfg.DLXExchange}
	if err := declareQueue(ch, queue, args); err != nil {
		return fmt.Errorf("failed to declare main queue '%s': %w", queue, err)
	}

	// Bind queues to exchanges
	routingKeys := cfg.QueueToRoutingKeys[queue]
	for _, routingKey := range routingKeys {
		if err := bindQueue(ch, queue, routingKey, cfg.Exchange); err != nil {
			return fmt.Errorf("failed to bind main queue '%s' to exchange '%s': %w", queue, cfg.Exchange, err)
		}
		if err := bindQueue(ch, dlQueue, routingKey, cfg.DLXExchange); err != nil {
			return fmt.Errorf("failed to bind dead-letter queue '%s' to exchange '%s': %w", dlQueue, cfg.DLXExchange, err)
		}
	}

	return nil
}

// declareQueue declares a queue with optional arguments.
func declareQueue(ch *amqp.Channel, queue string, args amqp.Table) error {
	_, err := ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		args,
	)
	return err
}

// bindQueue binds a queue to an exchange with a routing key.
func bindQueue(ch *amqp.Channel, queue, routingKey, exchange string) error {
	return ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false,
		nil,
	)
}
