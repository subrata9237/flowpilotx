package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Queue QueueConfig
	HTTP  HTTPConfig
}

type QueueConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	QueueName  string
	Exchange   string
	RoutingKey string
}

type HTTPConfig struct {
	TargetURL         string
	TimeoutSeconds    int
	RetryAttempts     int
	RetryDelaySeconds int
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("config")
	v.AddConfigPath(".")

	// Set default values
	v.SetDefault("queue.host", "localhost")
	v.SetDefault("queue.port", 5672)
	v.SetDefault("queue.username", "guest")
	v.SetDefault("queue.password", "guest")
	v.SetDefault("queue.queue_name", "http_requests")
	v.SetDefault("queue.exchange", "")
	v.SetDefault("queue.routing_key", "http_requests")
	v.SetDefault("http.timeout_seconds", 5)
	v.SetDefault("http.retry_attempts", 3)
	v.SetDefault("http.retry_delay_seconds", 1)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}
