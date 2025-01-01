package config

import "os"

type RabbitMQConfig struct {
	Host          string
	Port          string
	Username      string
	Password      string
	Exchange      string
	RoutingKeyMap map[string]string
}

type ServerConfig struct {
	Port string
}
type Config struct {
	RabbitMQ RabbitMQConfig
	Server   ServerConfig
}

func LoadConfig() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
			Username: getEnv("RABBITMQ_USERNAME", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
			Exchange: getEnv("RABBITMQ_EXCHANGE", "event_exchange"),
			RoutingKeyMap: map[string]string{ //todo: refactor this code
				"user_signup":   "user.registered",
				"order_created": "order.created",
			},
		},
		Server: ServerConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
