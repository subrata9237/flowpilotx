// Package nats provides a production-ready NATS client library with support for both
// core NATS and JetStream functionality. It includes features such as automatic
// reconnection, connection pooling, health monitoring, and comprehensive error handling.
package nats

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// Mode represents the NATS operation mode
type Mode string

const (
	// StandaloneMode represents a single NATS server
	StandaloneMode Mode = "standalone"
	// ClusterMode represents a NATS cluster
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
	ErrNilConfig = errors.New("nats configuration is nil")
	// ErrInvalidMode is returned when an invalid mode is specified
	ErrInvalidMode = errors.New("invalid nats mode")
	// ErrClientClosed is returned when attempting to use a closed client
	ErrClientClosed = errors.New("nats client is closed")
	// ErrJetStreamNotEnabled is returned when attempting to use JetStream features without enabling it
	ErrJetStreamNotEnabled = errors.New("JetStream not enabled")
	// ErrInvalidConfig is returned when the configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")
)

// Config represents NATS configuration options for establishing and maintaining
// connections to NATS servers. It provides comprehensive control over connection
// behavior, security, and performance characteristics.
type Config struct {
	// Mode specifies whether to use standalone or cluster mode
	Mode Mode

	// Addresses is a list of NATS server URLs
	// For standalone mode, only the first address is used
	// For cluster mode, all addresses are used
	Addresses []string

	// Authentication options
	Username string // Username for authentication
	Password string // Password for authentication
	Token    string // Token for authentication

	// TLS configuration
	EnableTLS     bool   // Enable TLS security
	TLSCertFile   string // Client certificate file
	TLSKeyFile    string // Client key file
	TLSCACertFile string // CA certificate file

	// Connection options
	MaxReconnects    int           // Maximum number of reconnection attempts (-1 for unlimited)
	ReconnectWait    time.Duration // Wait time between reconnection attempts
	ConnectionName   string        // Name of the connection (useful for monitoring)
	ConnectTimeout   time.Duration // Timeout for establishing connection
	PingInterval     time.Duration // Interval between ping messages
	MaxPingOutstand  int           // Maximum number of outstanding ping messages
	ReconnectBufSize int64         // Size of the reconnect buffer

	// JetStream configuration
	EnableJetStream bool         // Enable JetStream functionality
	JetStreamConfig []nats.JSOpt // JetStream specific options

	// Health check settings
	HealthCheckInterval time.Duration // Interval between health checks
	HealthCheckTimeout  time.Duration // Timeout for health check operations
}

// DefaultConfig returns a Config instance initialized with default values
func DefaultConfig() *Config {
	return &Config{
		Mode:                StandaloneMode,
		Addresses:           []string{"nats://localhost:4222"},
		MaxReconnects:       defaultMaxReconnects,
		ReconnectWait:       defaultReconnectWait,
		ConnectTimeout:      defaultConnectTimeout,
		PingInterval:        defaultPingInterval,
		MaxPingOutstand:     defaultMaxPingOutstand,
		ReconnectBufSize:    defaultReconnectBufSize,
		HealthCheckInterval: defaultHealthCheckInterval,
		HealthCheckTimeout:  defaultHealthCheckTimeout / 2, // Ensure timeout is less than interval
	}
}

// Client represents a thread-safe NATS client that provides high-level operations
// for both core NATS and JetStream functionality.
type Client struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	closed bool

	// Health check
	healthCheckTicker *time.Ticker
	healthCheckDone   chan struct{}

	// Connection event handlers
	connectedHandler    func(*nats.Conn)
	disconnectedHandler func(*nats.Conn, error)
	reconnectedHandler  func(*nats.Conn)
	errorHandler        func(*nats.Conn, *nats.Subscription, error)
}

// NewClient creates and initializes a new NATS client with the provided configuration.
// If no configuration is provided, default configuration is used.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config:            cfg,
		ctx:               ctx,
		cancel:            cancel,
		healthCheckTicker: time.NewTicker(cfg.HealthCheckInterval),
		healthCheckDone:   make(chan struct{}),
	}

	// Set up connection handlers
	client.setupConnectionHandlers()

	// Connect to NATS
	if err := client.connect(); err != nil {
		client.Close()
		return nil, err
	}

	// Initialize JetStream if enabled
	if cfg.EnableJetStream {
		js, err := client.conn.JetStream(cfg.JetStreamConfig...)
		if err != nil {
			client.Close()
			return nil, fmt.Errorf("failed to initialize JetStream: %w", err)
		}
		client.js = js
	}

	// Start health check routine
	go client.healthCheckLoop()

	return client, nil
}

// validateConfig performs comprehensive validation of the configuration
func validateConfig(cfg *Config) error {
	if cfg.Mode != StandaloneMode && cfg.Mode != ClusterMode {
		return ErrInvalidMode
	}
	if len(cfg.Addresses) == 0 {
		return ErrInvalidConfig
	}
	if cfg.EnableTLS {
		if cfg.TLSCertFile == "" || cfg.TLSKeyFile == "" {
			return errors.New("TLS cert and key files must be provided when TLS is enabled")
		}
	}

	// Set default values if not provided
	if cfg.HealthCheckInterval <= 0 {
		cfg.HealthCheckInterval = defaultHealthCheckInterval
	}
	if cfg.HealthCheckTimeout <= 0 {
		cfg.HealthCheckTimeout = defaultHealthCheckTimeout / 2 // Ensure timeout is less than interval
	}

	// Validate health check timeout is less than interval
	if cfg.HealthCheckTimeout >= cfg.HealthCheckInterval {
		return errors.New("health check timeout must be less than interval")
	}

	// Set other default values if not provided
	if cfg.MaxReconnects == 0 {
		cfg.MaxReconnects = defaultMaxReconnects
	}
	if cfg.ReconnectWait <= 0 {
		cfg.ReconnectWait = defaultReconnectWait
	}
	if cfg.ConnectTimeout <= 0 {
		cfg.ConnectTimeout = defaultConnectTimeout
	}
	if cfg.PingInterval <= 0 {
		cfg.PingInterval = defaultPingInterval
	}
	if cfg.MaxPingOutstand <= 0 {
		cfg.MaxPingOutstand = defaultMaxPingOutstand
	}
	if cfg.ReconnectBufSize <= 0 {
		cfg.ReconnectBufSize = defaultReconnectBufSize
	}

	return nil
}

// setupConnectionHandlers sets up NATS connection event handlers
func (c *Client) setupConnectionHandlers() {
	c.connectedHandler = func(nc *nats.Conn) {
		fmt.Printf("Connected to NATS server: %v\n", nc.ConnectedUrl())
	}

	c.disconnectedHandler = func(nc *nats.Conn, err error) {
		if err != nil {
			fmt.Printf("Disconnected from NATS server with error: %v\n", err)
		} else {
			fmt.Println("Disconnected from NATS server")
		}
	}

	c.reconnectedHandler = func(nc *nats.Conn) {
		fmt.Printf("Reconnected to NATS server: %v\n", nc.ConnectedUrl())
	}

	c.errorHandler = func(nc *nats.Conn, sub *nats.Subscription, err error) {
		if sub != nil {
			fmt.Printf("Error in subscription %q: %v\n", sub.Subject, err)
		} else {
			fmt.Printf("Error in connection: %v\n", err)
		}
	}
}

// connect establishes connection to NATS server(s)
func (c *Client) connect() error {
	options := []nats.Option{
		nats.Name(c.config.ConnectionName),
		nats.MaxReconnects(c.config.MaxReconnects),
		nats.ReconnectWait(c.config.ReconnectWait),
		nats.Timeout(c.config.ConnectTimeout),
		nats.PingInterval(c.config.PingInterval),
		nats.MaxPingsOutstanding(c.config.MaxPingOutstand),
		nats.ReconnectBufSize(int(c.config.ReconnectBufSize)),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("Disconnected from NATS server with error: %v\n", err)
			}
			if c.disconnectedHandler != nil {
				c.disconnectedHandler(nc, err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("Reconnected to NATS server: %v\n", nc.ConnectedUrl())
			if c.reconnectedHandler != nil {
				c.reconnectedHandler(nc)
			}
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			c.mu.Lock()
			c.closed = true
			c.mu.Unlock()
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			if err != nil {
				fmt.Printf("Error in NATS connection: %v\n", err)
			}
			if c.errorHandler != nil {
				c.errorHandler(nc, sub, err)
			}
		}),
		nats.ConnectHandler(func(nc *nats.Conn) {
			fmt.Printf("Connected to NATS server: %v\n", nc.ConnectedUrl())
			if c.connectedHandler != nil {
				c.connectedHandler(nc)
			}
		}),
		// Ensure we don't give up too quickly in tests
		nats.RetryOnFailedConnect(true),
	}

	// Add authentication options
	if c.config.Username != "" && c.config.Password != "" {
		options = append(options, nats.UserInfo(c.config.Username, c.config.Password))
	} else if c.config.Token != "" {
		options = append(options, nats.Token(c.config.Token))
	}

	// Add TLS options if enabled
	if c.config.EnableTLS {
		if c.config.TLSCertFile != "" && c.config.TLSKeyFile != "" {
			options = append(options, nats.ClientCert(c.config.TLSCertFile, c.config.TLSKeyFile))
		}
		if c.config.TLSCACertFile != "" {
			options = append(options, nats.RootCAs(c.config.TLSCACertFile))
		}
	}

	var err error
	if c.config.Mode == ClusterMode {
		// For cluster mode, try connecting to all servers
		servers := strings.Join(c.config.Addresses, ",")
		c.conn, err = nats.Connect(servers, options...)
	} else {
		// For standalone mode, connect to the first server
		c.conn, err = nats.Connect(c.config.Addresses[0], options...)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Verify connection is established
	if !c.conn.IsConnected() {
		return fmt.Errorf("failed to establish NATS connection")
	}

	return nil
}

// healthCheckLoop performs periodic health checks
func (c *Client) healthCheckLoop() {
	for {
		select {
		case <-c.healthCheckTicker.C:
			if err := c.Ping(c.config.HealthCheckTimeout); err != nil {
				fmt.Printf("NATS health check failed: %v\n", err)
			}
		case <-c.healthCheckDone:
			return
		}
	}
}

// Close closes the NATS client and stops health checks
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	if c.healthCheckTicker != nil {
		c.healthCheckTicker.Stop()
	}
	if c.healthCheckDone != nil {
		close(c.healthCheckDone)
	}
	if c.cancel != nil {
		c.cancel()
	}

	if c.conn != nil {
		return c.conn.Drain()
	}
	return nil
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

// Ping checks the connection to NATS server
func (c *Client) Ping(timeout time.Duration) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.conn.FlushTimeout(timeout)
}

// checkConnection verifies that the client and connection are valid
func (c *Client) checkConnection() error {
	if c == nil {
		return errors.New("client is nil")
	}

	if err := c.checkClosed(); err != nil {
		return err
	}

	if c.conn == nil {
		return errors.New("connection is nil")
	}

	if !c.conn.IsConnected() {
		return errors.New("connection is not established")
	}

	return nil
}

// Publish publishes a message to a subject
func (c *Client) Publish(subject string, data []byte) error {
	if err := c.checkConnection(); err != nil {
		return err
	}
	return c.conn.Publish(subject, data)
}

// PublishMsg publishes a NATS message
func (c *Client) PublishMsg(msg *nats.Msg) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.conn.PublishMsg(msg)
}

// Subscribe creates a subscription
func (c *Client) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if err := c.checkConnection(); err != nil {
		return nil, err
	}
	return c.conn.Subscribe(subject, handler)
}

// QueueSubscribe creates a queue subscription
func (c *Client) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	if err := c.checkConnection(); err != nil {
		return nil, err
	}
	return c.conn.QueueSubscribe(subject, queue, handler)
}

// Request sends a request and waits for a response
func (c *Client) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	if err := c.checkConnection(); err != nil {
		return nil, err
	}
	return c.conn.Request(subject, data, timeout)
}

// JetStream Operations

// JetStream returns the JetStream context
func (c *Client) JetStream() (nats.JetStreamContext, error) {
	if err := c.checkConnection(); err != nil {
		return nil, err
	}
	if c.js == nil {
		return nil, ErrJetStreamNotEnabled
	}
	return c.js, nil
}

// CreateStream creates a new stream
func (c *Client) CreateStream(cfg *nats.StreamConfig) (*nats.StreamInfo, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if c.js == nil {
		return nil, errors.New("JetStream not enabled")
	}
	return c.js.AddStream(cfg)
}

// UpdateStream updates an existing stream
func (c *Client) UpdateStream(cfg *nats.StreamConfig) (*nats.StreamInfo, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if c.js == nil {
		return nil, errors.New("JetStream not enabled")
	}
	return c.js.UpdateStream(cfg)
}

// DeleteStream deletes a stream
func (c *Client) DeleteStream(name string) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	if c.js == nil {
		return errors.New("JetStream not enabled")
	}
	return c.js.DeleteStream(name)
}

// StreamInfo gets information about a stream
func (c *Client) StreamInfo(name string) (*nats.StreamInfo, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if c.js == nil {
		return nil, errors.New("JetStream not enabled")
	}
	return c.js.StreamInfo(name)
}

// PublishAsync publishes a message using JetStream asynchronously
func (c *Client) PublishAsync(subject string, data []byte) (nats.PubAckFuture, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if c.js == nil {
		return nil, errors.New("JetStream not enabled")
	}
	return c.js.PublishAsync(subject, data)
}

// Subscribe creates a JetStream subscription
func (c *Client) JetStreamSubscribe(subject string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if c.js == nil {
		return nil, errors.New("JetStream not enabled")
	}
	return c.js.Subscribe(subject, handler, opts...)
}

// QueueSubscribe creates a JetStream queue subscription
func (c *Client) JetStreamQueueSubscribe(subject, queue string, handler nats.MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	if err := c.checkClosed(); err != nil {
		return nil, err
	}
	if c.js == nil {
		return nil, errors.New("JetStream not enabled")
	}
	return c.js.QueueSubscribe(subject, queue, handler, opts...)
}

// GetConnection returns the underlying NATS connection
// This should be used with caution as it bypasses the wrapper's safety checks
func (c *Client) GetConnection() *nats.Conn {
	return c.conn
}

// GetStats returns current statistics about the NATS connection
func (c *Client) GetStats() nats.Statistics {
	if c.conn == nil {
		return nats.Statistics{}
	}
	return c.conn.Stats()
}

// DrainConnection drains the connection, allowing in-flight messages to complete
func (c *Client) DrainConnection(timeout time.Duration) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.conn.Drain()
}

// Flush flushes the connection, ensuring all published messages have been sent
func (c *Client) Flush(timeout time.Duration) error {
	if err := c.checkClosed(); err != nil {
		return err
	}
	return c.conn.FlushTimeout(timeout)
}

// Status returns the current status of the NATS connection
func (c *Client) Status() nats.Status {
	if c.conn == nil {
		return nats.CLOSED
	}
	return c.conn.Status()
}

// InMsgCount returns the count of incoming messages
func (c *Client) InMsgCount() uint64 {
	if c.conn == nil {
		return 0
	}
	return c.conn.InMsgs
}

// OutMsgCount returns the count of outgoing messages
func (c *Client) OutMsgCount() uint64 {
	if c.conn == nil {
		return 0
	}
	return c.conn.OutMsgs
}

// LastError returns the last error encountered by the connection
func (c *Client) LastError() error {
	if c.conn == nil {
		return ErrClientClosed
	}
	return c.conn.LastError()
}
