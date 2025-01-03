package config

import "os"

type RabbitMQConfig struct {
	Host               string
	Port               string
	Username           string
	Password           string
	Exchange           string
	ExchangeType       string
	DLXExchange        string
	ServiceToQueue     map[string]string
	QueueToRoutingKeys map[string][]string
	EventToRoutingKey  map[string]string
}

type ServerConfig struct {
	Port string
}
type Config struct {
	RabbitMQ RabbitMQConfig
	Server   ServerConfig
}

// LoadConfig get project config
func LoadConfig() *Config {
	return &Config{
		RabbitMQ: RabbitMQConfig{
			Host:        getEnv("RABBITMQ_HOST", "localhost"),
			Port:        getEnv("RABBITMQ_PORT", "5672"),
			Username:    getEnv("RABBITMQ_USERNAME", "guest"),
			Password:    getEnv("RABBITMQ_PASSWORD", "guest"),
			Exchange:    getEnv("RABBITMQ_EXCHANGE", "notification_exchange"),
			DLXExchange: getEnv("RABBITMQ_DLX_EXCHANGE", "dlx_notification_exchange"),

			//todo: refactor this code get map from yaml file and check recreate if need (best solution ....)
			ServiceToQueue: map[string]string{
				"order": "order_notification",
				"user":  "user_notification",
			},
			QueueToRoutingKeys: map[string][]string{
				"user_notification":  {"user.registered"}, // and other event .... user.change_password
				"order_notification": {"order.created"},
			},
			EventToRoutingKey: map[string]string{
				"order_created": "order.created",
				"user_signup":   "user.registered",
			},
			/////////////////////////////////////

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
