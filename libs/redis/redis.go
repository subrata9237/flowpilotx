package redis

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Mode represents the Redis operation mode (standalone or cluster)
type Mode string

const (
	// StandaloneMode represents a single Redis instance
	StandaloneMode Mode = "standalone"
	// ClusterMode represents a Redis cluster
	ClusterMode Mode = "cluster"

	// Default configuration values
	defaultMaxReconnects       = 60
	defaultReconnectWait       = 2 * time.Second
	defaultConnectTimeout      = 5 * time.Second
	defaultPingInterval        = 2 * time.Minute
	defaultMaxPingOutstand     = 2
	defaultReconnectBufSize    = 8 * 1024 * 1024 // 8MB
	defaultHealthCheckTimeout  = 2 * time.Second
	defaultHealthCheckInterval = 30 * time.Second
)

var (
	// ErrNilConfig is returned when configuration is nil
	ErrNilConfig = errors.New("redis configuration is nil")
	// ErrInvalidMode is returned when an invalid mode is specified
	ErrInvalidMode = errors.New("invalid redis mode")
	// ErrClientClosed is returned when attempting to use a closed client
	ErrClientClosed = errors.New("redis client is closed")
)

// Config represents Redis configuration options for establishing and maintaining
// connections to Redis servers. It provides comprehensive control over connection
// behavior, security, and performance characteristics.
type Config struct {
	// Mode specifies whether to use standalone or cluster mode
	Mode Mode

	// Addresses is a list of Redis node addresses
	// For standalone mode, only the first address is used
	// For cluster mode, all addresses are used as seed nodes
	Addresses []string

	// Password for authentication
	Password string

	// DB number for standalone mode (ignored in cluster mode)
	DB int

	// Connection pool settings
	PoolSize     int           // Maximum number of socket connections
	MinIdleConns int           // Minimum number of idle connections
	MaxRetries   int           // Maximum number of retries before giving up
	DialTimeout  time.Duration // Timeout for establishing new connections
	ReadTimeout  time.Duration // Timeout for socket reads
	WriteTimeout time.Duration // Timeout for socket writes

	// Health check settings
	HealthCheckInterval time.Duration // Interval between health checks
	HealthCheckTimeout  time.Duration // Timeout for health check operations

	// TLS configuration (optional)
	EnableTLS bool // Enable TLS security
}

// DefaultConfig returns a new Config instance initialized with default values.
// The default configuration is suitable for most use cases but can be customized
// by modifying the returned Config instance.
func DefaultConfig() *Config {
	return &Config{
		Mode:                StandaloneMode,
		Addresses:           []string{"localhost:6379"},
		DB:                  0,
		PoolSize:            10,
		MinIdleConns:        5,
		MaxRetries:          3,
		DialTimeout:         5 * time.Second,
		ReadTimeout:         3 * time.Second,
		WriteTimeout:        3 * time.Second,
		HealthCheckInterval: defaultHealthCheckInterval,
		HealthCheckTimeout:  defaultHealthCheckTimeout,
	}
}

// Client represents a thread-safe Redis client that provides high-level operations
// for both standalone and cluster Redis deployments.
type Client struct {
	// Universal interface for both standalone and cluster clients
	universalClient redis.UniversalClient

	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	closed bool

	// Health check
	healthCheckTicker *time.Ticker
	healthCheckDone   chan struct{}
}

// NewClient creates and initializes a new Redis client with the provided configuration.
// If no configuration is provided, default configuration is used.
//
// Example:
//
//	config := &redis.Config{
//	    Mode:      redis.StandaloneMode,
//	    Addresses: []string{"localhost:6379"},
//	    Password:  "secret",
//	}
//	client, err := redis.NewClient(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var universalClient redis.UniversalClient

	switch cfg.Mode {
	case StandaloneMode:
		universalClient = redis.NewClient(&redis.Options{
			Addr:         cfg.Addresses[0],
			Password:     cfg.Password,
			DB:           cfg.DB,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
	case ClusterMode:
		universalClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        cfg.Addresses,
			Password:     cfg.Password,
			PoolSize:     cfg.PoolSize,
			MinIdleConns: cfg.MinIdleConns,
			MaxRetries:   cfg.MaxRetries,
			DialTimeout:  cfg.DialTimeout,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
		})
	}

	// Ensure health check interval is positive
	if cfg.HealthCheckInterval <= 0 {
		cfg.HealthCheckInterval = defaultHealthCheckInterval
	}

	client := &Client{
		universalClient:   universalClient,
		config:            cfg,
		ctx:               ctx,
		cancel:            cancel,
		healthCheckTicker: time.NewTicker(cfg.HealthCheckInterval),
		healthCheckDone:   make(chan struct{}),
	}

	// Start health check routine
	go client.healthCheckLoop()

	// Test initial connection
	if err := client.Ping(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// validateConfig performs comprehensive validation of the Redis configuration.
// It checks for valid mode, addresses, and sets default values where needed.
func validateConfig(cfg *Config) error {
	if cfg.Mode != StandaloneMode && cfg.Mode != ClusterMode {
		return ErrInvalidMode
	}
	if len(cfg.Addresses) == 0 {
		return errors.New("no Redis addresses provided")
	}
	if cfg.Mode == StandaloneMode && len(cfg.Addresses) > 1 {
		return errors.New("multiple addresses provided for standalone mode")
	}

	// Set default values if not provided
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 10
	}
	if cfg.MinIdleConns <= 0 {
		cfg.MinIdleConns = 5
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = defaultConnectTimeout
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 3 * time.Second
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = 3 * time.Second
	}
	if cfg.HealthCheckInterval <= 0 {
		cfg.HealthCheckInterval = defaultHealthCheckInterval
	}
	if cfg.HealthCheckTimeout <= 0 {
		cfg.HealthCheckTimeout = defaultHealthCheckTimeout
	}

	// Validate health check timeout is less than interval
	if cfg.HealthCheckTimeout >= cfg.HealthCheckInterval {
		return errors.New("health check timeout must be less than interval")
	}

	return nil
}

// healthCheckLoop performs periodic health checks
func (c *Client) healthCheckLoop() {
	for {
		select {
		case <-c.healthCheckTicker.C:
			ctx, cancel := context.WithTimeout(c.ctx, c.config.HealthCheckTimeout)
			if err := c.PingContext(ctx); err != nil {
				// Log error or emit metric for monitoring
				fmt.Printf("Redis health check failed: %v\n", err)
			}
			cancel()
		case <-c.healthCheckDone:
			return
		}
	}
}

// Close closes the Redis client and stops health checks.
// It is safe to call Close multiple times.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.healthCheckTicker.Stop()
	close(c.healthCheckDone)
	c.cancel()
	return c.universalClient.Close()
}

// checkClosed checks if the client is closed
func (c *Client) checkClosed() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.closed {
		return ErrClientClosed
	}
	return nil
}

// WithContext returns a new Client with the given context.
// This is useful for setting timeouts or deadlines for operations.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	clientWithTimeout := client.WithContext(ctx)
//	value, err := clientWithTimeout.Get("key")
func (c *Client) WithContext(ctx context.Context) *Client {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client := *c
	client.ctx = ctx
	return &client
}

// Ping checks the connection to Redis.
// It returns an error if the connection is not healthy.
func (c *Client) Ping() error {
	return c.PingContext(c.ctx)
}

// PingContext checks the connection to Redis with context
func (c *Client) PingContext(ctx context.Context) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.Ping(ctx).Err()
}

// String Operations

// Set sets a key-value pair with optional expiration.
// If expiration is 0, the key will not expire.
//
// Example:
//
//	err := client.Set("key", "value", time.Hour)
func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.Set(c.ctx, key, value, expiration).Err()
}

// Get retrieves the value of a key.
// It returns redis.Nil if the key does not exist.
//
// Example:
//
//	value, err := client.Get("key")
//	if err == redis.Nil {
//	    // key does not exist
//	}
func (c *Client) Get(key string) (string, error) {
	if err := c.checkClosed(); err != nil {
		return "", err
	}
	return c.universalClient.Get(c.ctx, key).Result()
}

// Delete deletes one or more keys.
// It returns nil if the keys were deleted or did not exist.
//
// Example:
//
//	err := client.Delete("key1", "key2")
func (c *Client) Delete(keys ...string) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.Del(c.ctx, keys...).Err()
}

// Exists checks if one or more keys exist.
// It returns true if at least one key exists.
//
// Example:
//
//	exists, err := client.Exists("key1", "key2")
func (c *Client) Exists(keys ...string) (bool, error) {
	if err := c.checkClosed(); err != nil {
		return false, err
	}
	n, err := c.universalClient.Exists(c.ctx, keys...).Result()
	return n > 0, err
}

// Expire sets a key's time to live in seconds
func (c *Client) Expire(key string, expiration time.Duration) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.Expire(c.ctx, key, expiration).Err()
}

// Hash Operations

// HSet sets field in the hash stored at key to value.
// If key does not exist, a new key holding a hash is created.
//
// Example:
//
//	err := client.HSet("hash", "field1", "value1", "field2", "value2")
func (c *Client) HSet(key string, values ...interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.HSet(c.ctx, key, values...).Err()
}

// HGet returns the value associated with field in the hash stored at key.
// It returns redis.Nil if the key or field do not exist.
//
// Example:
//
//	value, err := client.HGet("hash", "field")
func (c *Client) HGet(key, field string) (string, error) {
	if err := c.checkClosed(); err != nil {
		return "", err
	}
	return c.universalClient.HGet(c.ctx, key, field).Result()
}

// HGetAll returns all fields and values of the hash stored at key
func (c *Client) HGetAll(key string) (map[string]string, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	return c.universalClient.HGetAll(c.ctx, key).Result()
}

// HDel removes one or more specified fields from the hash stored at key
func (c *Client) HDel(key string, fields ...string) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.HDel(c.ctx, key, fields...).Err()
}

// List Operations

// LPush inserts all the specified values at the head of the list stored at key.
// If key does not exist, it is created as empty list before performing the push operations.
//
// Example:
//
//	err := client.LPush("list", "value1", "value2")
func (c *Client) LPush(key string, values ...interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.LPush(c.ctx, key, values...).Err()
}

// RPush inserts all the specified values at the tail of the list stored at key.
// If key does not exist, it is created as empty list before performing the push operation.
//
// Example:
//
//	err := client.RPush("list", "value1", "value2")
func (c *Client) RPush(key string, values ...interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.RPush(c.ctx, key, values...).Err()
}

// LPop removes and returns the first element of the list stored at key
func (c *Client) LPop(key string) (string, error) {
	if err := c.checkClosed(); err != nil {
		return "", err
	}
	return c.universalClient.LPop(c.ctx, key).Result()
}

// RPop removes and returns the last element of the list stored at key
func (c *Client) RPop(key string) (string, error) {
	if err := c.checkClosed(); err != nil {
		return "", err
	}
	return c.universalClient.RPop(c.ctx, key).Result()
}

// LRange returns the specified elements of the list stored at key
func (c *Client) LRange(key string, start, stop int64) ([]string, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	return c.universalClient.LRange(c.ctx, key, start, stop).Result()
}

// Set Operations

// SAdd adds one or more members to a set.
// If key does not exist, a new set is created before adding the specified members.
//
// Example:
//
//	err := client.SAdd("set", "member1", "member2")
func (c *Client) SAdd(key string, members ...interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.SAdd(c.ctx, key, members...).Err()
}

// SMembers returns all the members of the set value stored at key
func (c *Client) SMembers(key string) ([]string, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	return c.universalClient.SMembers(c.ctx, key).Result()
}

// SRem removes one or more members from a set
func (c *Client) SRem(key string, members ...interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.SRem(c.ctx, key, members...).Err()
}

// SIsMember returns if member is a member of the set stored at key
func (c *Client) SIsMember(key string, member interface{}) (bool, error) {
	if err := c.checkClosed(); err != nil {
		return false, err
	}
	return c.universalClient.SIsMember(c.ctx, key, member).Result()
}

// Sorted Set Operations

// ZAdd adds one or more members to a sorted set.
// If key does not exist, a new sorted set is created before adding the specified members.
//
// Example:
//
//	err := client.ZAdd("zset",
//	    redis.Z{Score: 1, Member: "member1"},
//	    redis.Z{Score: 2, Member: "member2"},
//	)
func (c *Client) ZAdd(key string, members ...redis.Z) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.ZAdd(c.ctx, key, members...).Err()
}

// ZRange returns a range of members in a sorted set, by index
func (c *Client) ZRange(key string, start, stop int64) ([]string, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	return c.universalClient.ZRange(c.ctx, key, start, stop).Result()
}

// ZRangeByScore returns a range of members in a sorted set, by score
func (c *Client) ZRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	return c.universalClient.ZRangeByScore(c.ctx, key, opt).Result()
}

// ZRem removes one or more members from a sorted set
func (c *Client) ZRem(key string, members ...interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.ZRem(c.ctx, key, members...).Err()
}

// Pipeline returns a new pipeline which can be used to execute multiple commands
// in a single round-trip.
//
// Example:
//
//	pipe := client.Pipeline()
//	pipe.Set(ctx, "key1", "value1", time.Hour)
//	pipe.Set(ctx, "key2", "value2", time.Hour)
//	cmds, err := pipe.Exec(ctx)
func (c *Client) Pipeline() redis.Pipeliner {
	if err := c.checkClosed(); err != nil {
		return nil
	}
	return c.universalClient.Pipeline()
}

// Transaction Operations

// Watch watches the given keys to determine execution of the MULTI/EXEC block
func (c *Client) Watch(fn func(*redis.Tx) error, keys ...string) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.Watch(c.ctx, fn, keys...)
}

// Pub/Sub Operations

// Subscribe subscribes to the specified channels.
// It returns a *redis.PubSub that can be used to receive messages.
//
// Example:
//
//	pubsub := client.Subscribe("channel1", "channel2")
//	defer pubsub.Close()
//
//	for {
//	    msg, err := pubsub.ReceiveMessage(context.Background())
//	    if err != nil {
//	        break
//	    }
//	    fmt.Println(msg.Channel, msg.Payload)
//	}
func (c *Client) Subscribe(channels ...string) *redis.PubSub {
	if err := c.checkClosed(); err != nil {
		return nil
	}
	return c.universalClient.Subscribe(c.ctx, channels...)
}

// Publish posts a message to the specified channel
func (c *Client) Publish(channel string, message interface{}) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.Publish(c.ctx, channel, message).Err()
}

// Utility Operations

// FlushDB deletes all keys in the current DB
func (c *Client) FlushDB() error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.FlushDB(c.ctx).Err()
}

// FlushAll deletes all keys from all DBs
func (c *Client) FlushAll() error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.universalClient.FlushAll(c.ctx).Err()
}

// Keys returns all keys matching pattern
func (c *Client) Keys(pattern string) ([]string, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	return c.universalClient.Keys(c.ctx, pattern).Result()
}

// TTL returns the remaining time to live of a key
func (c *Client) TTL(key string) (time.Duration, error) {
	if err := c.checkClosed(); err != nil {
		return 0, err
	}
	return c.universalClient.TTL(c.ctx, key).Result()
}

// Info returns information and statistics about the server
func (c *Client) Info(section ...string) (string, error) {
	if err := c.checkClosed(); err != nil {
		return "", err
	}
	return c.universalClient.Info(c.ctx, section...).Result()
}

// GetClient returns the underlying Redis client.
// This should be used with caution as it bypasses the wrapper's safety checks.
//
// Warning: Direct usage of the underlying client may bypass safety checks and
// connection management features provided by this wrapper.
func (c *Client) GetClient() redis.UniversalClient {
	return c.universalClient
}
