// Package nats provides a production-ready NATS client library with support for both
// core NATS and JetStream functionality.
package nats

import (
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testServer represents a test NATS server instance
type testServer struct {
	server *server.Server
	addr   string
}

// runTestServer starts a new NATS server instance for testing purposes.
// It configures the server with random port and minimal logging.
func runTestServer(t *testing.T) *testServer {
	opts := &server.Options{
		Host:           "127.0.0.1",
		Port:           -1, // random port
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
		JetStream:      true,        // Enable JetStream
		StoreDir:       t.TempDir(), // Use temp directory for storage
	}

	s, err := server.NewServer(opts)
	require.NoError(t, err)
	go s.Start()

	// Wait for server to be ready
	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatal("NATS server failed to start")
	}

	// Wait for JetStream to be ready
	timeout := time.Now().Add(4 * time.Second)
	for time.Now().Before(timeout) {
		if s.JetStreamEnabled() {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}

	if !s.JetStreamEnabled() {
		t.Fatal("JetStream failed to start")
	}

	// Get the client URL
	clientURL := s.ClientURL()

	return &testServer{
		server: s,
		addr:   clientURL,
	}
}

// setupTestCluster creates a cluster of test NATS servers for testing purposes.
// It returns an array of test server instances.
func setupTestCluster(t *testing.T) []*testServer {
	servers := make([]*testServer, 3)
	for i := 0; i < 3; i++ {
		servers[i] = runTestServer(t)
	}
	return servers
}

// TestMain is the entry point for running all tests in this package.
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	os.Exit(code)
}

// TestNewClient_AllConfigurations tests all possible configuration scenarios
func TestNewClient_AllConfigurations(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errType error
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: false, // Should use default config
		},
		{
			name: "minimal valid config",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{ts.addr},
			},
			wantErr: false,
		},
		{
			name: "full valid config",
			config: &Config{
				Mode:                StandaloneMode,
				Addresses:           []string{ts.addr},
				Username:            "test",
				Password:            "test",
				MaxReconnects:       5,
				ReconnectWait:       time.Second,
				ConnectionName:      "test",
				ConnectTimeout:      time.Second,
				HealthCheckInterval: 2 * time.Second,
				HealthCheckTimeout:  time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			config: &Config{
				Mode:      "invalid",
				Addresses: []string{ts.addr},
			},
			wantErr: true,
			errType: ErrInvalidMode,
		},
		{
			name: "empty addresses",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{},
			},
			wantErr: true,
			errType: ErrInvalidConfig,
		},
		{
			name: "invalid address",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{"invalid://localhost:1234"},
			},
			wantErr: true,
		},
		{
			name: "invalid health check timeout",
			config: &Config{
				Mode:                StandaloneMode,
				Addresses:           []string{ts.addr},
				HealthCheckInterval: time.Second,
				HealthCheckTimeout:  2 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid TLS config",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{ts.addr},
				EnableTLS: true,
			},
			wantErr: true,
		},
		{
			name: "cluster mode config",
			config: &Config{
				Mode:      ClusterMode,
				Addresses: []string{ts.addr, ts.addr}, // Same address for testing
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				tt.config = &Config{
					Mode:      StandaloneMode,
					Addresses: []string{ts.addr},
				}
			}
			client, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.True(t, errors.Is(err, tt.errType))
				}
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
			if client != nil {
				// Test connection
				err = client.Ping(time.Second)
				assert.NoError(t, err)
				client.Close()
			}
		})
	}
}

// TestClient_ConnectionEvents tests all connection event handlers
func TestClient_ConnectionEvents(t *testing.T) {
	// Start test server with specific port
	port := 14222
	opts := &server.Options{
		Host:           "127.0.0.1",
		Port:           port,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
	}

	s, err := server.NewServer(opts)
	require.NoError(t, err)
	go s.Start()
	defer s.Shutdown()

	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatal("NATS server failed to start")
	}

	clientURL := s.ClientURL()

	client, err := NewClient(&Config{
		Mode:          StandaloneMode,
		Addresses:     []string{clientURL},
		MaxReconnects: 5,
		ReconnectWait: 100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer client.Close()

	// Test disconnection event
	disconnected := make(chan struct{})
	reconnected := make(chan struct{})
	errorOccurred := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)

	client.disconnectedHandler = func(nc *nats.Conn, err error) {
		select {
		case <-disconnected:
			// Channel already closed
		default:
			close(disconnected)
		}
	}

	// Test reconnection event
	client.reconnectedHandler = func(nc *nats.Conn) {
		select {
		case <-reconnected:
			// Channel already closed
		default:
			close(reconnected)
			wg.Done()
		}
	}

	// Test error handler
	client.errorHandler = func(nc *nats.Conn, sub *nats.Subscription, err error) {
		select {
		case <-errorOccurred:
			// Channel already closed
		default:
			close(errorOccurred)
		}
	}

	// Ensure initial connection
	err = client.Ping(time.Second)
	require.NoError(t, err)

	// Simulate server restart
	s.Shutdown()

	// Wait for disconnect
	select {
	case <-disconnected:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Disconnect handler not called")
	}

	// Start new server with same port
	newServer, err := server.NewServer(opts)
	require.NoError(t, err)
	defer newServer.Shutdown()

	go newServer.Start()
	require.True(t, newServer.ReadyForConnections(4*time.Second))

	// Wait for reconnection
	wg.Wait()

	// Verify connection is restored
	err = client.Ping(time.Second)
	assert.NoError(t, err)
}

// TestClient_JetStreamOperations tests all JetStream operations
func TestClient_JetStreamOperations(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:            StandaloneMode,
		Addresses:       []string{ts.addr},
		EnableJetStream: true,
	})
	require.NoError(t, err)
	defer client.Close()

	// Test stream operations
	streamName := "TEST_STREAM"
	streamConfig := &nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{"test.*"},
		Storage:  nats.MemoryStorage,
	}

	// Create stream
	streamInfo, err := client.CreateStream(streamConfig)
	require.NoError(t, err)
	assert.Equal(t, streamName, streamInfo.Config.Name)

	// Update stream
	streamConfig.MaxMsgs = 1000
	streamInfo, err = client.UpdateStream(streamConfig)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), streamInfo.Config.MaxMsgs)

	// Get stream info
	streamInfo, err = client.StreamInfo(streamName)
	require.NoError(t, err)
	assert.Equal(t, streamName, streamInfo.Config.Name)

	// Test publish and subscribe
	msgData := []byte("test message")

	// Publish async
	ack, err := client.PublishAsync("test.msg", msgData)
	require.NoError(t, err)

	select {
	case <-ack.Ok():
		// Success
	case err := <-ack.Err():
		t.Fatalf("Publish error: %v", err)
	case <-time.After(time.Second):
		t.Fatal("Publish timeout")
	}

	// Subscribe
	received := make(chan []byte)
	sub, err := client.JetStreamSubscribe("test.*", func(msg *nats.Msg) {
		received <- msg.Data
		msg.Ack()
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Queue subscribe
	qsub, err := client.JetStreamQueueSubscribe("test.*", "workers", func(msg *nats.Msg) {
		received <- msg.Data
		msg.Ack()
	})
	require.NoError(t, err)
	defer qsub.Unsubscribe()

	// Publish another message
	ack, err = client.PublishAsync("test.msg", msgData)
	require.NoError(t, err)
	<-ack.Ok()

	// Wait for message
	select {
	case msg := <-received:
		assert.Equal(t, msgData, msg)
	case <-time.After(time.Second):
		t.Fatal("Message not received")
	}

	// Delete stream
	err = client.DeleteStream(streamName)
	require.NoError(t, err)

	// Verify stream deletion
	_, err = client.StreamInfo(streamName)
	assert.Error(t, err)
}

// TestClient_ErrorScenarios tests various error scenarios
func TestClient_ErrorScenarios(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)

	// Test operations on closed client
	client.Close()

	tests := []struct {
		name      string
		operation func() error
	}{
		{
			name: "publish after close",
			operation: func() error {
				return client.Publish("test", []byte("msg"))
			},
		},
		{
			name: "subscribe after close",
			operation: func() error {
				_, err := client.Subscribe("test", func(msg *nats.Msg) {})
				return err
			},
		},
		{
			name: "queue subscribe after close",
			operation: func() error {
				_, err := client.QueueSubscribe("test", "queue", func(msg *nats.Msg) {})
				return err
			},
		},
		{
			name: "request after close",
			operation: func() error {
				_, err := client.Request("test", []byte("msg"), time.Second)
				return err
			},
		},
		{
			name: "ping after close",
			operation: func() error {
				return client.Ping(time.Second)
			},
		},
		{
			name: "flush after close",
			operation: func() error {
				return client.Flush(time.Second)
			},
		},
		{
			name: "drain after close",
			operation: func() error {
				return client.DrainConnection(time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			assert.Error(t, err)
			assert.Equal(t, ErrClientClosed, err)
		})
	}
}

// TestClient_RequestReply tests the request-reply pattern with timeouts
func TestClient_RequestReply(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	// Setup reply handler
	sub, err := client.Subscribe("request", func(msg *nats.Msg) {
		client.Publish(msg.Reply, []byte("response"))
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Test successful request
	resp, err := client.Request("request", []byte("request"), time.Second)
	require.NoError(t, err)
	assert.Equal(t, []byte("response"), resp.Data)

	// Test request timeout with non-existent handler
	_, err = client.Request("no_handler", []byte("request"), 100*time.Millisecond)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, nats.ErrTimeout), "Expected timeout error, got: %v", err)

	// Test request with slow handler
	slowSub, err := client.Subscribe("slow_request", func(msg *nats.Msg) {
		time.Sleep(200 * time.Millisecond)
		client.Publish(msg.Reply, []byte("late_response"))
	})
	require.NoError(t, err)
	defer slowSub.Unsubscribe()

	// Request should timeout before response
	_, err = client.Request("slow_request", []byte("request"), 100*time.Millisecond)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, nats.ErrTimeout), "Expected timeout error, got: %v", err)

	// Test request with closed client
	client.Close()
	_, err = client.Request("request", []byte("request"), time.Second)
	assert.Error(t, err)
	assert.Equal(t, ErrClientClosed, err)
}

// TestClient_LoadBalancing tests queue subscription load balancing
func TestClient_LoadBalancing(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	// Create multiple queue subscribers
	counts := make([]int, 3)
	var wg sync.WaitGroup
	wg.Add(100) // Total messages to be processed

	for i := 0; i < 3; i++ {
		workerID := i
		sub, err := client.QueueSubscribe("load.test", "workers", func(msg *nats.Msg) {
			counts[workerID]++
			wg.Done()
		})
		require.NoError(t, err)
		defer sub.Unsubscribe()
	}

	// Publish messages
	for i := 0; i < 100; i++ {
		err = client.Publish("load.test", []byte("test"))
		require.NoError(t, err)
	}

	// Wait for all messages to be processed
	wg.Wait()

	// Verify load distribution
	total := 0
	for _, count := range counts {
		assert.True(t, count > 0, "Each worker should receive some messages")
		total += count
	}
	assert.Equal(t, 100, total)
}

// TestClient_HealthCheck tests health check functionality
func TestClient_HealthCheck(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:                StandaloneMode,
		Addresses:           []string{ts.addr},
		HealthCheckInterval: 100 * time.Millisecond,
		HealthCheckTimeout:  50 * time.Millisecond,
	})
	require.NoError(t, err)
	defer client.Close()

	// Wait for some health checks
	time.Sleep(250 * time.Millisecond)

	// Verify client is healthy
	err = client.Ping(time.Second)
	assert.NoError(t, err)

	// Stop server and verify health check fails
	ts.server.Shutdown()
	time.Sleep(150 * time.Millisecond)

	err = client.Ping(time.Second)
	assert.Error(t, err)
}

// TestClient_Monitoring tests monitoring capabilities
func TestClient_Monitoring(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	// Test initial stats
	stats := client.GetStats()
	assert.Equal(t, uint64(0), stats.InMsgs)
	assert.Equal(t, uint64(0), stats.OutMsgs)

	// Test message counts after publishing
	err = client.Publish("test", []byte("message"))
	require.NoError(t, err)

	stats = client.GetStats()
	assert.Equal(t, uint64(0), stats.InMsgs)
	assert.Equal(t, uint64(1), stats.OutMsgs)

	// Test status
	assert.Equal(t, nats.CONNECTED, client.Status())

	// Test last error when none
	assert.NoError(t, client.LastError())
}

// TestClient_GracefulShutdown tests graceful shutdown capabilities
func TestClient_GracefulShutdown(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)

	// Setup subscriber
	wg := sync.WaitGroup{}
	wg.Add(1)
	sub, err := client.Subscribe("test", func(msg *nats.Msg) {
		time.Sleep(100 * time.Millisecond) // Simulate processing
		wg.Done()
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Publish message
	err = client.Publish("test", []byte("message"))
	require.NoError(t, err)

	// Drain connection
	err = client.DrainConnection(time.Second)
	assert.NoError(t, err)

	// Wait for message processing
	wg.Wait()

	// Verify client is closed
	err = client.Publish("test", []byte("message"))
	assert.Error(t, err)
	assert.Equal(t, ErrClientClosed, err)

	// Verify double drain is safe
	err = client.DrainConnection(time.Second)
	assert.Error(t, err)
	assert.Equal(t, ErrClientClosed, err)

	// Verify double close is safe
	err = client.Close()
	assert.NoError(t, err)
}

// TestClient_Flush tests message flushing capabilities
func TestClient_Flush(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	// Publish messages
	for i := 0; i < 100; i++ {
		err = client.Publish("test", []byte(fmt.Sprintf("msg%d", i)))
		require.NoError(t, err)
	}

	// Flush messages
	err = client.Flush(time.Second)
	assert.NoError(t, err)

	// Verify message count
	stats := client.GetStats()
	assert.Equal(t, uint64(100), stats.OutMsgs)
}

// TestClient_ConcurrentOperations tests thread safety
func TestClient_ConcurrentOperations(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	// Run concurrent operations
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// Publish messages
			for j := 0; j < 100; j++ {
				err := client.Publish("test", []byte(fmt.Sprintf("msg%d-%d", n, j)))
				assert.NoError(t, err)
			}
			// Get stats
			stats := client.GetStats()
			assert.NotNil(t, stats)
			// Check status
			status := client.Status()
			assert.Equal(t, nats.CONNECTED, status)
		}(i)
	}

	wg.Wait()
}

// TestNewClient_Standalone tests the creation and basic functionality
// of a standalone NATS client.
func TestNewClient_Standalone(t *testing.T) {
	// Start test server with auth
	opts := &server.Options{
		Host:           "127.0.0.1",
		Port:           -1, // random port
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
		Username:       "test",
		Password:       "test",
	}

	s, err := server.NewServer(opts)
	require.NoError(t, err)
	go s.Start()
	defer s.Shutdown()

	if !s.ReadyForConnections(4 * time.Second) {
		t.Fatal("NATS server failed to start")
	}

	clientURL := s.ClientURL()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "default config",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{clientURL},
				Username:  "test",
				Password:  "test",
			},
			wantErr: false,
		},
		{
			name: "with auth",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{clientURL},
				Username:  "test",
				Password:  "test",
			},
			wantErr: false,
		},
		{
			name: "invalid auth",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{clientURL},
				Username:  "wrong",
				Password:  "wrong",
			},
			wantErr: true,
		},
		{
			name: "invalid address",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{"nats://invalid:4222"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
			defer client.Close()

			// Test connection
			err = client.Ping(time.Second)
			assert.NoError(t, err)
		})
	}
}

// TestNewClient_Cluster tests the creation and basic functionality
// of a NATS client in cluster mode.
func TestNewClient_Cluster(t *testing.T) {
	// Start test cluster
	servers := setupTestCluster(t)
	defer func() {
		for _, s := range servers {
			s.server.Shutdown()
		}
	}()

	addresses := make([]string, len(servers))
	for i, s := range servers {
		addresses[i] = s.addr
	}

	config := &Config{
		Mode:      ClusterMode,
		Addresses: addresses,
	}

	client, err := NewClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Test connection
	err = client.Ping(time.Second)
	assert.NoError(t, err)
}

// TestClient_PubSub tests basic publish/subscribe functionality
// including message delivery and subscription handling.
func TestClient_PubSub(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	// Test basic pub/sub
	subject := "test.pubsub"
	message := []byte("hello world")
	received := make(chan []byte, 1)

	// Subscribe
	sub, err := client.Subscribe(subject, func(msg *nats.Msg) {
		received <- msg.Data
	})
	require.NoError(t, err)
	defer sub.Unsubscribe()

	// Publish
	err = client.Publish(subject, message)
	require.NoError(t, err)

	// Wait for message
	select {
	case msg := <-received:
		assert.Equal(t, message, msg)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

// TestClient_QueueSubscribe tests queue subscription functionality
// including load balancing between multiple subscribers.
func TestClient_QueueSubscribe(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	subject := "test.queue"
	queue := "workers"
	messageCount := 10
	received1 := make(chan int, messageCount)
	received2 := make(chan int, messageCount)

	// Create two queue subscribers
	sub1, err := client.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		received1 <- 1
	})
	require.NoError(t, err)
	defer sub1.Unsubscribe()

	sub2, err := client.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		received2 <- 1
	})
	require.NoError(t, err)
	defer sub2.Unsubscribe()

	// Publish messages
	for i := 0; i < messageCount; i++ {
		err = client.Publish(subject, []byte(fmt.Sprintf("msg%d", i)))
		require.NoError(t, err)
	}

	// Wait for all messages
	timeout := time.After(time.Second)
	count1, count2 := 0, 0
	for count1+count2 < messageCount {
		select {
		case <-received1:
			count1++
		case <-received2:
			count2++
		case <-timeout:
			t.Fatal("timeout waiting for messages")
		}
	}

	// Verify load balancing
	assert.True(t, count1 > 0, "subscriber 1 received no messages")
	assert.True(t, count2 > 0, "subscriber 2 received no messages")
	assert.Equal(t, messageCount, count1+count2)
}

// TestClient_Reconnect tests the client's ability to handle
// server disconnections and automatic reconnection.
func TestClient_Reconnect(t *testing.T) {
	ts := runTestServer(t)

	client, err := NewClient(&Config{
		Mode:          StandaloneMode,
		Addresses:     []string{ts.addr},
		MaxReconnects: 5,
		ReconnectWait: 100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer client.Close()

	// Ensure initial connection
	err = client.Ping(time.Second)
	require.NoError(t, err)

	// Setup reconnection handler
	reconnected := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	client.reconnectedHandler = func(nc *nats.Conn) {
		select {
		case <-reconnected:
			// Channel already closed
		default:
			close(reconnected)
			wg.Done()
		}
	}

	// Get the server's port
	oldServer := ts.server
	port := oldServer.Addr().(*net.TCPAddr).Port

	// Shutdown server
	oldServer.Shutdown()
	time.Sleep(100 * time.Millisecond)

	// Start new server on same port
	opts := &server.Options{
		Host:           "127.0.0.1",
		Port:           port,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
	}

	newServer, err := server.NewServer(opts)
	require.NoError(t, err)
	defer newServer.Shutdown()

	go newServer.Start()
	require.True(t, newServer.ReadyForConnections(4*time.Second))

	// Wait for reconnection
	wg.Wait()

	// Verify connection is restored
	err = client.Ping(time.Second)
	assert.NoError(t, err)

	// Test operations after reconnection
	err = client.Publish("test", []byte("message"))
	assert.NoError(t, err)

	// Test subscription after reconnection
	sub, err := client.Subscribe("test", func(msg *nats.Msg) {})
	assert.NoError(t, err)
	defer sub.Unsubscribe()
}

// TestClient_Context tests the client's context-aware operations
// including cancellation handling.
func TestClient_Context(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)
	defer client.Close()

	// Test basic operation
	err = client.Publish("test", []byte("message"))
	assert.NoError(t, err)
}

// TestClient_Close tests proper client cleanup and resource release
// including handling of operations after closure.
func TestClient_Close(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	client, err := NewClient(&Config{
		Mode:      StandaloneMode,
		Addresses: []string{ts.addr},
	})
	require.NoError(t, err)

	// Close client
	err = client.Close()
	assert.NoError(t, err)

	// Verify operations fail after close
	err = client.Publish("test", []byte("message"))
	assert.Error(t, err)
	assert.Equal(t, ErrClientClosed, err)

	// Verify double close is safe
	err = client.Close()
	assert.NoError(t, err)
}

// TestNewClient_InvalidConfigurations tests invalid configuration scenarios
func TestNewClient_InvalidConfigurations(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr string
	}{
		{
			name: "health check timeout greater than interval",
			config: &Config{
				Mode:                StandaloneMode,
				Addresses:           []string{"nats://localhost:4222"},
				HealthCheckInterval: time.Second,
				HealthCheckTimeout:  2 * time.Second,
			},
			wantErr: "health check timeout must be less than interval",
		},
		{
			name: "invalid connection URL",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{"invalid://localhost:4222"},
			},
			wantErr: "failed to establish NATS connection",
		},
		{
			name: "empty addresses",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{},
			},
			wantErr: "no addresses provided",
		},
		{
			name: "invalid mode",
			config: &Config{
				Mode:      "invalid",
				Addresses: []string{"nats://localhost:4222"},
			},
			wantErr: "invalid mode",
		},
		{
			name: "invalid TLS config",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{"nats://localhost:4222"},
				EnableTLS: true,
			},
			wantErr: "TLS configuration error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.config)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestClient_NilHandling tests nil client handling
func TestClient_NilHandling(t *testing.T) {
	var client *Client

	// Test nil client operations
	err := client.Publish("test", []byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is nil")

	_, err = client.Subscribe("test", func(msg *nats.Msg) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is nil")

	_, err = client.QueueSubscribe("test", "queue", func(msg *nats.Msg) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is nil")

	_, err = client.Request("test", []byte("test"), time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is nil")

	_, err = client.JetStream()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is nil")
}

// TestClient_ConnectionStateHandling tests connection state handling
func TestClient_ConnectionStateHandling(t *testing.T) {
	ts := runTestServer(t)

	client, err := NewClient(&Config{
		Mode:          StandaloneMode,
		Addresses:     []string{ts.addr},
		MaxReconnects: 5,
		ReconnectWait: 100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer client.Close()

	// Test initial connection state
	assert.True(t, client.conn.IsConnected())
	assert.Equal(t, nats.CONNECTED, client.Status())

	// Setup connection event handlers
	disconnected := make(chan struct{})
	reconnected := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	client.disconnectedHandler = func(nc *nats.Conn, err error) {
		select {
		case <-disconnected:
			// Channel already closed
		default:
			close(disconnected)
		}
	}

	client.reconnectedHandler = func(nc *nats.Conn) {
		select {
		case <-reconnected:
			// Channel already closed
		default:
			close(reconnected)
			wg.Done()
		}
	}

	// Get the server's port
	oldServer := ts.server
	port := oldServer.Addr().(*net.TCPAddr).Port

	// Test operations before disconnect
	err = client.Publish("test", []byte("test"))
	assert.NoError(t, err)

	// Shutdown server
	oldServer.Shutdown()

	// Wait for disconnect
	select {
	case <-disconnected:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Disconnect handler not called")
	}

	// Test operations during disconnect
	err = client.Publish("test", []byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection lost")

	// Start new server on same port
	opts := &server.Options{
		Host:           "127.0.0.1",
		Port:           port,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
	}

	newServer, err := server.NewServer(opts)
	require.NoError(t, err)
	defer newServer.Shutdown()

	go newServer.Start()
	require.True(t, newServer.ReadyForConnections(4*time.Second))

	// Wait for reconnection
	wg.Wait()

	// Test operations after reconnection
	err = client.Publish("test", []byte("test"))
	assert.NoError(t, err)

	// Test subscription after reconnection
	sub, err := client.Subscribe("test", func(msg *nats.Msg) {})
	assert.NoError(t, err)
	defer sub.Unsubscribe()

	// Test client closure
	client.Close()
	assert.Equal(t, nats.CLOSED, client.Status())

	// Test operations after closure
	err = client.Publish("test", []byte("test"))
	assert.Error(t, err)
	assert.Equal(t, ErrClientClosed, err)
}

// TestClient_HealthCheckConfiguration tests health check configuration
func TestClient_HealthCheckConfiguration(t *testing.T) {
	ts := runTestServer(t)
	defer ts.server.Shutdown()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "default health check values",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{ts.addr},
			},
			wantErr: false,
		},
		{
			name: "custom valid health check values",
			config: &Config{
				Mode:                StandaloneMode,
				Addresses:           []string{ts.addr},
				HealthCheckInterval: 5 * time.Second,
				HealthCheckTimeout:  1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid health check values",
			config: &Config{
				Mode:                StandaloneMode,
				Addresses:           []string{ts.addr},
				HealthCheckInterval: 1 * time.Second,
				HealthCheckTimeout:  2 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, client)
			if client != nil {
				client.Close()
			}
		})
	}
}
