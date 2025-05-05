// Package config provides functionality to load and manage application configuration
package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Environment represents the running environment
type Environment string

const (
	Development Environment = "development"
	Staging     Environment = "staging"
	Production  Environment = "production"
)

// Config represents the application configuration
type Config struct {
	Environment   Environment
	MongoDB      MongoDBConfig
	Redis        RedisConfig
	RedisCluster RedisClusterConfig
	NATS         NATSConfig
	NATSStreaming NATSStreamingConfig
	JetStream    JetStreamConfig
	General      GeneralConfig
	ProjectRoot  string
}

// MongoDBConfig represents MongoDB configuration
type MongoDBConfig struct {
	URI              string
	Username         string
	Password         string
	Database         string
	ConnectTimeout   time.Duration
	OperationTimeout time.Duration
	MaxPoolSize      uint64
	MinPoolSize      uint64
	RetryWrites      bool
	RetryReads       bool
	Direct           bool
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host               string
	Port               int
	Password           string
	Database           int
	MaxRetries         int
	PoolSize           int
	MinIdleConns       int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolTimeout        time.Duration
	IdleTimeout        time.Duration
	IdleCheckFrequency time.Duration
}

// RedisClusterConfig represents Redis Cluster configuration
type RedisClusterConfig struct {
	Enabled         bool
	Addresses       []string
	Password        string
	RouteByLatency  bool
	RouteRandomly   bool
}

// NATSConfig represents NATS configuration
type NATSConfig struct {
	URL            string
	Username       string
	Password       string
	Token          string
	ConnectTimeout time.Duration
	MaxReconnects  int
	ReconnectWait  time.Duration
}

// NATSStreamingConfig represents NATS Streaming configuration
type NATSStreamingConfig struct {
	ClusterID           string
	ClientID            string
	MaxPubAcksInFlight int
	ConnectWait         time.Duration
	DiscoveryPrefix     string
	MaxPingsOut         int
	PingInterval        time.Duration
}

// JetStreamConfig represents JetStream configuration
type JetStreamConfig struct {
	Enabled         bool
	MaxWait         time.Duration
	MaxMessages     int64
	MaxBytes        int64
	Replicas        int
	MemoryStorage   bool
	RetentionPolicy string
}

// GeneralConfig represents general configuration
type GeneralConfig struct {
	LogLevel        string
	LogFormat       string
	APIPort         int
	MetricsPort     int
	HealthCheckPort int
}

var (
	config *Config
	once   sync.Once
)

// LoadConfig loads the configuration from environment variables and .dotnetenv files
func LoadConfig(envFiles ...string) (*Config, error) {
	var err error
	
	once.Do(func() {
		// Get project root with error handling
		projectRoot, err := os.Getwd()
		if err != nil {
			log.Printf("Warning: Failed to get working directory: %v", err)
			projectRoot = "." // Fallback to current directory
		}
		
		// Try loading .dotnetenv file first
		dotnetEnvPath := filepath.Join(projectRoot, ".dotnetenv")
		if err = LoadDotNetEnv(dotnetEnvPath); err != nil {
			log.Printf("Note: .dotnetenv file not found at %s", dotnetEnvPath)
		}

		// Load additional env files if provided
		if len(envFiles) > 0 {
			for _, file := range envFiles {
				if err := LoadDotNetEnv(file); err != nil {
					log.Printf("Note: Could not load env file %s: %v", file, err)
				}
			}
		}

		config = &Config{
			ProjectRoot: projectRoot,
			Environment: Environment(strings.ToLower(getEnvOrDefault("FLOWPILOT_ENV", "development"))),
		}
		
		// MongoDB configuration
		config.MongoDB = MongoDBConfig{
			URI:              getEnvOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
			Username:         getEnvOrDefault("MONGODB_USERNAME", ""),
			Password:         getEnvOrDefault("MONGODB_PASSWORD", ""),
			Database:         getEnvOrDefault("MONGODB_DATABASE", "flowpilot"),
			ConnectTimeout:   getDurationOrDefault("MONGODB_CONNECT_TIMEOUT", 10*time.Second),
			OperationTimeout: getDurationOrDefault("MONGODB_OPERATION_TIMEOUT", 5*time.Second),
			MaxPoolSize:      getUint64OrDefault("MONGODB_MAX_POOL_SIZE", 10),
			MinPoolSize:      getUint64OrDefault("MONGODB_MIN_POOL_SIZE", 1),
			RetryWrites:      getBoolOrDefault("MONGODB_RETRY_WRITES", true),
			RetryReads:       getBoolOrDefault("MONGODB_RETRY_READS", true),
			Direct:           getBoolOrDefault("MONGODB_DIRECT", false),
		}

		// Redis configuration
		config.Redis = RedisConfig{
			Host:               getEnvOrDefault("REDIS_HOST", "localhost"),
			Port:               getIntOrDefault("REDIS_PORT", 6379),
			Password:           getEnvOrDefault("REDIS_PASSWORD", ""),
			Database:           getIntOrDefault("REDIS_DATABASE", 0),
			MaxRetries:         getIntOrDefault("REDIS_MAX_RETRIES", 3),
			PoolSize:           getIntOrDefault("REDIS_POOL_SIZE", 10),
			MinIdleConns:       getIntOrDefault("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:        getDurationOrDefault("REDIS_DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:        getDurationOrDefault("REDIS_READ_TIMEOUT", 3*time.Second),
			WriteTimeout:       getDurationOrDefault("REDIS_WRITE_TIMEOUT", 3*time.Second),
			PoolTimeout:        getDurationOrDefault("REDIS_POOL_TIMEOUT", 4*time.Second),
			IdleTimeout:        getDurationOrDefault("REDIS_IDLE_TIMEOUT", 300*time.Second),
			IdleCheckFrequency: getDurationOrDefault("REDIS_IDLE_CHECK_FREQUENCY", 60*time.Second),
		}

		// Redis Cluster configuration
		config.RedisCluster = RedisClusterConfig{
			Enabled:         getBoolOrDefault("REDIS_CLUSTER_ENABLED", false),
			Addresses:       getStringSliceOrDefault("REDIS_CLUSTER_ADDRESSES", []string{}),
			Password:        getEnvOrDefault("REDIS_CLUSTER_PASSWORD", ""),
			RouteByLatency:  getBoolOrDefault("REDIS_CLUSTER_ROUTE_BY_LATENCY", false),
			RouteRandomly:   getBoolOrDefault("REDIS_CLUSTER_ROUTE_RANDOMLY", false),
		}

		// NATS configuration
		config.NATS = NATSConfig{
			URL:            getEnvOrDefault("NATS_URL", "nats://localhost:4222"),
			Username:       getEnvOrDefault("NATS_USERNAME", ""),
			Password:       getEnvOrDefault("NATS_PASSWORD", ""),
			Token:          getEnvOrDefault("NATS_TOKEN", ""),
			ConnectTimeout: getDurationOrDefault("NATS_CONNECT_TIMEOUT", 2*time.Second),
			MaxReconnects:  getIntOrDefault("NATS_MAX_RECONNECTS", -1),
			ReconnectWait:  getDurationOrDefault("NATS_RECONNECT_WAIT", 2*time.Second),
		}

		// General configuration
		config.General = GeneralConfig{
			LogLevel:        getEnvOrDefault("LOG_LEVEL", "info"),
			LogFormat:       getEnvOrDefault("LOG_FORMAT", "text"),
			APIPort:         getIntOrDefault("API_PORT", 8080),
			MetricsPort:     getIntOrDefault("METRICS_PORT", 9090),
			HealthCheckPort: getIntOrDefault("HEALTH_CHECK_PORT", 8081),
		}
	})

	return config, err
}

// GetConfig returns the singleton config instance
func GetConfig() *Config {
	if config == nil {
		_, err := LoadConfig()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}
	return config
}

// Helper functions for type conversion
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

func getBoolOrDefault(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func getUint64OrDefault(key string, defaultValue uint64) uint64 {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.ParseUint(value, 10, 64); err == nil {
			return i
		}
	}
	return defaultValue
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func getStringSliceOrDefault(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// GetMongoURI returns the MongoDB URI with credentials and database name
func (c *Config) GetMongoURI() string {
	uri := c.MongoDB.URI
	if c.MongoDB.Username != "" && c.MongoDB.Password != "" {
		// Add credentials if provided
		uri = fmt.Sprintf("mongodb://%s:%s@%s",
			c.MongoDB.Username,
			c.MongoDB.Password,
			uri[10:]) // Remove "mongodb://" prefix
	}
	return fmt.Sprintf("%s/%s", uri, c.MongoDB.Database)
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == Development
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == Production
}

// IsStaging returns true if the application is running in staging mode
func (c *Config) IsStaging() bool {
	return c.Environment == Staging
}

// GetRedisAddress returns the full Redis address (host:port)
func (c *Config) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
} 