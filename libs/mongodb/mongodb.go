// Package mongodb provides a production-ready MongoDB client wrapper with support for
// connection management, CRUD operations, and advanced querying capabilities.
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flowpilotx/libs/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	// ErrClientClosed is returned when attempting to use a closed client
	ErrClientClosed = errors.New("mongodb client is closed")
	// ErrInvalidConfig is returned when the configuration is invalid
	ErrInvalidConfig = errors.New("invalid mongodb configuration")
	// ErrNoDocuments is returned when no documents are found
	ErrNoDocuments = mongo.ErrNoDocuments
)

// Config represents MongoDB client configuration
type Config struct {
	URI              string
	Database         string
	Username         string
	Password         string
	ConnectTimeout   time.Duration
	OperationTimeout time.Duration
	MaxPoolSize      uint64
	MinPoolSize      uint64
	RetryWrites      bool
	RetryReads       bool
	Direct           bool
}

// NewConfigFromEnv creates a new Config from environment configuration
func NewConfigFromEnv() (*Config, error) {
	cfg, err := config.LoadConfig(config.DefaultConfigPath())
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Get MongoDB development config since it's the only environment we support now
	mongoConfig := cfg.MongoDB.Development

	return &Config{
		URI:              mongoConfig.URI,
		Database:         mongoConfig.Database,
		Username:         mongoConfig.Username,
		Password:         mongoConfig.Password,
		ConnectTimeout:   mongoConfig.ConnectTimeout,
		OperationTimeout: mongoConfig.OperationTimeout,
		MaxPoolSize:      mongoConfig.MaxPoolSize,
		MinPoolSize:      mongoConfig.MinPoolSize,
		RetryWrites:      mongoConfig.RetryWrites,
		RetryReads:       mongoConfig.RetryReads,
		Direct:           mongoConfig.Direct,
	}, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		URI:              "mongodb://localhost:27017",
		Database:         "test",
		ConnectTimeout:   10 * time.Second,
		OperationTimeout: 5 * time.Second,
		MaxPoolSize:      100,
		MinPoolSize:      10,
		RetryWrites:      true,
		RetryReads:       true,
		Direct:           false,
	}
}

// Client represents a MongoDB client wrapper
type Client struct {
	client   *mongo.Client
	database *mongo.Database
	config   *Config
}

// NewClient creates a new MongoDB client with the given configuration
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		var err error
		cfg, err = NewConfigFromEnv()
		if err != nil {
			cfg = DefaultConfig()
		}
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	opts := options.Client().ApplyURI(cfg.URI)
	opts.SetMaxPoolSize(cfg.MaxPoolSize)
	opts.SetMinPoolSize(cfg.MinPoolSize)
	opts.SetRetryWrites(cfg.RetryWrites)
	opts.SetRetryReads(cfg.RetryReads)
	opts.SetDirect(cfg.Direct)

	if cfg.Username != "" && cfg.Password != "" {
		opts.SetAuth(options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		})
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return &Client{
		client:   client,
		database: client.Database(cfg.Database),
		config:   cfg,
	}, nil
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg.URI == "" {
		return errors.New("URI is required")
	}
	if cfg.Database == "" {
		return errors.New("database is required")
	}
	if cfg.ConnectTimeout <= 0 {
		return errors.New("connect timeout must be positive")
	}
	if cfg.OperationTimeout <= 0 {
		return errors.New("operation timeout must be positive")
	}
	if cfg.MaxPoolSize < cfg.MinPoolSize {
		return errors.New("max pool size must be greater than or equal to min pool size")
	}
	return nil
}

// Close closes the MongoDB client connection
func (c *Client) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnectTimeout)
	defer cancel()
	return c.client.Disconnect(ctx)
}

// Collection returns a handle to a MongoDB collection
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// WithDatabase returns a new Client with the given database
func (c *Client) WithDatabase(name string) *Client {
	return &Client{
		client:   c.client,
		database: c.client.Database(name),
		config:   c.config,
	}
}

// Ping verifies that the client can connect to the MongoDB server
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, readpref.Primary())
}

// Database returns the current database
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Client returns the underlying MongoDB client
func (c *Client) Client() *mongo.Client {
	return c.client
}

// Config returns the current configuration
func (c *Client) Config() *Config {
	return c.config
}
