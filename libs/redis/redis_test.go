package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Test Helpers
// -----------------------------------------------------------------------------

// testRedisHelper represents a test Redis instance using miniredis.
type testRedisHelper struct {
	*miniredis.Miniredis
	t       *testing.T
	origEnv string
}

// newTestRedisHelper creates a new test Redis instance with default configuration.
func newTestRedisHelper(t *testing.T) *testRedisHelper {
	t.Helper()

	// Store original environment
	origEnv := os.Getenv("REDIS_ADDR")

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create test Redis: %v", err)
	}

	// Set environment variable
	if err := os.Setenv("REDIS_ADDR", mr.Addr()); err != nil {
		mr.Close()
		t.Fatalf("Failed to set REDIS_ADDR: %v", err)
	}

	return &testRedisHelper{
		Miniredis: mr,
		t:         t,
		origEnv:   origEnv,
	}
}

// cleanup cleans up the test Redis instance and environment variables.
func (tr *testRedisHelper) cleanup() {
	tr.t.Helper()

	if tr.Miniredis != nil {
		tr.Close()
		tr.Miniredis = nil
	}

	// Restore original environment
	if tr.origEnv != "" {
		if err := os.Setenv("REDIS_ADDR", tr.origEnv); err != nil {
			tr.t.Errorf("Failed to restore REDIS_ADDR: %v", err)
		}
	} else {
		if err := os.Unsetenv("REDIS_ADDR"); err != nil {
			tr.t.Errorf("Failed to unset REDIS_ADDR: %v", err)
		}
	}
}

// getTestConfigHelper returns a test Redis configuration with specified options.
func getTestConfigHelper(addr string) *Config {
	return &Config{
		Mode:               StandaloneMode,
		Addresses:          []string{addr},
		Password:           "",
		DB:                 0,
		PoolSize:           5,
		MinIdleConns:       1,
		MaxRetries:         3,
		DialTimeout:        time.Second,
		ReadTimeout:        time.Second,
		WriteTimeout:       time.Second,
		HealthCheckTimeout: 500 * time.Millisecond,
	}
}

// -----------------------------------------------------------------------------
// Tests
// -----------------------------------------------------------------------------

func TestNewClient_Configuration(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config should use default",
			config:      nil,
			expectError: false,
		},
		{
			name: "valid standalone config",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{mr.Addr()},
			},
			expectError: false,
		},
		{
			name: "valid cluster config",
			config: &Config{
				Mode:      ClusterMode,
				Addresses: []string{mr.Addr(), "localhost:6380", "localhost:6381"},
			},
			expectError: false,
		},
		{
			name: "invalid mode",
			config: &Config{
				Mode:      Mode("invalid"),
				Addresses: []string{mr.Addr()},
			},
			expectError: true,
		},
		{
			name: "no addresses",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{},
			},
			expectError: true,
		},
		{
			name: "multiple addresses in standalone mode",
			config: &Config{
				Mode:      StandaloneMode,
				Addresses: []string{mr.Addr(), "localhost:6380"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				tt.config = getTestConfigHelper(mr.Addr())
			}
			client, err := NewClient(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				if client != nil {
					client.Close()
				}
			}
		})
	}
}

func TestClient_StringOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("Set and Get", func(t *testing.T) {
		err := client.Set("test_key", "test_value", time.Minute)
		assert.NoError(t, err)

		value, err := client.Get("test_key")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("Delete", func(t *testing.T) {
		err := client.Set("delete_key", "value", time.Minute)
		assert.NoError(t, err)

		err = client.Delete("delete_key")
		assert.NoError(t, err)

		_, err = client.Get("delete_key")
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("Exists", func(t *testing.T) {
		err := client.Set("exists_key", "value", time.Minute)
		assert.NoError(t, err)

		exists, err := client.Exists("exists_key")
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = client.Exists("nonexistent_key")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	// Test cleanup
	select {
	case <-ctx.Done():
		t.Fatal("Test timeout")
	default:
		// Test completed within timeout
	}
}

func TestClient_HashOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("HSet and HGet", func(t *testing.T) {
		err := client.HSet("hash_key", "field1", "value1", "field2", "value2")
		assert.NoError(t, err)

		value, err := client.HGet("hash_key", "field1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", value)
	})

	t.Run("HGetAll", func(t *testing.T) {
		err := client.HSet("hash_all", "f1", "v1", "f2", "v2")
		assert.NoError(t, err)

		values, err := client.HGetAll("hash_all")
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"f1": "v1", "f2": "v2"}, values)
	})

	t.Run("HDel", func(t *testing.T) {
		err := client.HSet("hash_del", "f1", "v1", "f2", "v2")
		assert.NoError(t, err)

		err = client.HDel("hash_del", "f1")
		assert.NoError(t, err)

		_, err = client.HGet("hash_del", "f1")
		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})

	// Test cleanup
	select {
	case <-ctx.Done():
		t.Fatal("Test timeout")
	default:
		// Test completed within timeout
	}
}

func TestClient_ListOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("LPush and RPush", func(t *testing.T) {
		err := client.LPush("list_key", "value1", "value2")
		assert.NoError(t, err)

		err = client.RPush("list_key", "value3")
		assert.NoError(t, err)

		values, err := client.LRange("list_key", 0, -1)
		assert.NoError(t, err)
		assert.Equal(t, []string{"value2", "value1", "value3"}, values)
	})

	t.Run("LPop and RPop", func(t *testing.T) {
		err := client.RPush("pop_list", "v1", "v2", "v3")
		assert.NoError(t, err)

		value, err := client.LPop("pop_list")
		assert.NoError(t, err)
		assert.Equal(t, "v1", value)

		value, err = client.RPop("pop_list")
		assert.NoError(t, err)
		assert.Equal(t, "v3", value)
	})
}

func TestClient_SetOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("SAdd and SMembers", func(t *testing.T) {
		err := client.SAdd("set_key", "member1", "member2")
		assert.NoError(t, err)

		members, err := client.SMembers("set_key")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"member1", "member2"}, members)
	})

	t.Run("SIsMember", func(t *testing.T) {
		err := client.SAdd("ismember_set", "member1")
		assert.NoError(t, err)

		exists, err := client.SIsMember("ismember_set", "member1")
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = client.SIsMember("ismember_set", "nonexistent")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("SRem", func(t *testing.T) {
		// Add members to set
		err := client.SAdd("srem_set", "member1", "member2", "member3")
		assert.NoError(t, err)

		// Remove members
		err = client.SRem("srem_set", "member1", "member2")
		assert.NoError(t, err)

		// Verify remaining members
		members, err := client.SMembers("srem_set")
		assert.NoError(t, err)
		assert.Equal(t, []string{"member3"}, members)

		// Try to remove non-existent member
		err = client.SRem("srem_set", "nonexistent")
		assert.NoError(t, err)

		// Try to remove from non-existent set
		err = client.SRem("nonexistent_set", "member")
		assert.NoError(t, err)
	})
}

func TestClient_SortedSetOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("ZAdd and ZRange", func(t *testing.T) {
		err := client.ZAdd("zset_key",
			redis.Z{Score: 1, Member: "member1"},
			redis.Z{Score: 2, Member: "member2"},
		)
		assert.NoError(t, err)

		members, err := client.ZRange("zset_key", 0, -1)
		assert.NoError(t, err)
		assert.Equal(t, []string{"member1", "member2"}, members)
	})

	t.Run("ZRangeByScore", func(t *testing.T) {
		err := client.ZAdd("zscore_key",
			redis.Z{Score: 1, Member: "member1"},
			redis.Z{Score: 2, Member: "member2"},
			redis.Z{Score: 3, Member: "member3"},
		)
		assert.NoError(t, err)

		members, err := client.ZRangeByScore("zscore_key", &redis.ZRangeBy{
			Min: "1",
			Max: "2",
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"member1", "member2"}, members)
	})

	t.Run("ZRem", func(t *testing.T) {
		// Add members to sorted set
		err := client.ZAdd("zrem_set",
			redis.Z{Score: 1, Member: "member1"},
			redis.Z{Score: 2, Member: "member2"},
			redis.Z{Score: 3, Member: "member3"},
		)
		assert.NoError(t, err)

		// Remove members
		err = client.ZRem("zrem_set", "member1", "member2")
		assert.NoError(t, err)

		// Verify remaining members
		members, err := client.ZRange("zrem_set", 0, -1)
		assert.NoError(t, err)
		assert.Equal(t, []string{"member3"}, members)

		// Try to remove non-existent member
		err = client.ZRem("zrem_set", "nonexistent")
		assert.NoError(t, err)

		// Try to remove from non-existent sorted set
		err = client.ZRem("nonexistent_set", "member")
		assert.NoError(t, err)
	})
}

func TestClient_PubSub(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("Publish and Subscribe", func(t *testing.T) {
		pubsub := client.Subscribe("test_channel")
		defer pubsub.Close()

		// Wait for subscription to be ready
		_, err := pubsub.Receive(context.Background())
		require.NoError(t, err)

		err = client.Publish("test_channel", "test_message")
		assert.NoError(t, err)

		msg, err := pubsub.ReceiveMessage(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "test_message", msg.Payload)
	})
}

func TestClient_PipelineOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("Pipeline with Mixed Operations", func(t *testing.T) {
		pipe := client.Pipeline()
		require.NotNil(t, pipe)

		// Queue multiple operations
		pipe.Set(context.Background(), "pipe_key1", "value1", time.Minute)
		pipe.Get(context.Background(), "pipe_key1")
		pipe.Del(context.Background(), "pipe_key1")
		pipe.Exists(context.Background(), "pipe_key1")

		// Execute pipeline
		cmds, err := pipe.Exec(context.Background())
		assert.NoError(t, err)
		assert.Len(t, cmds, 4)

		// Verify results
		assert.NoError(t, cmds[0].Err())                            // Set
		assert.Equal(t, "value1", cmds[1].(*redis.StringCmd).Val()) // Get
		assert.Equal(t, int64(1), cmds[2].(*redis.IntCmd).Val())    // Del
		assert.Equal(t, int64(0), cmds[3].(*redis.IntCmd).Val())    // Exists
	})

	t.Run("Pipeline with Error", func(t *testing.T) {
		pipe := client.Pipeline()
		require.NotNil(t, pipe)

		// Queue operations with an error
		pipe.Set(context.Background(), "pipe_key2", "value2", time.Minute)
		pipe.HGet(context.Background(), "pipe_key2", "field") // This should fail as key is string
		pipe.Get(context.Background(), "pipe_key2")

		// Execute pipeline
		cmds, err := pipe.Exec(context.Background())
		assert.Error(t, err)
		assert.NotNil(t, cmds)
		assert.NoError(t, cmds[0].Err())                    // Set should succeed
		assert.Error(t, cmds[1].Err())                      // HGet should fail
		assert.NoError(t, cmds[2].(*redis.StringCmd).Err()) // Get should succeed
	})

	t.Run("Pipeline with Concurrent Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				pipe := client.Pipeline()

				// Queue multiple operations
				for j := 0; j < 5; j++ {
					key := fmt.Sprintf("concurrent_pipe_%d_%d", i, j)
					pipe.Set(context.Background(), key, j, time.Minute)
					pipe.Get(context.Background(), key)
				}

				// Execute pipeline
				cmds, err := pipe.Exec(context.Background())
				assert.NoError(t, err)
				assert.Len(t, cmds, 10) // 5 sets + 5 gets
			}(i)
		}
		wg.Wait()
	})
}

func TestClient_TransactionOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("Transaction with Multiple Operations", func(t *testing.T) {
		// Set initial values
		err := client.Set("tx_key1", "value1", 0)
		assert.NoError(t, err)
		err = client.Set("tx_key2", "value2", 0)
		assert.NoError(t, err)

		err = client.Watch(func(tx *redis.Tx) error {
			// Get current values
			val1, err := tx.Get(context.Background(), "tx_key1").Result()
			assert.NoError(t, err)
			val2, err := tx.Get(context.Background(), "tx_key2").Result()
			assert.NoError(t, err)

			// Update both keys in transaction
			_, err = tx.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
				pipe.Set(context.Background(), "tx_key1", val1+"_updated", 0)
				pipe.Set(context.Background(), "tx_key2", val2+"_updated", 0)
				return nil
			})
			return err
		}, "tx_key1", "tx_key2")
		assert.NoError(t, err)

		// Verify updates
		val1, err := client.Get("tx_key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1_updated", val1)
		val2, err := client.Get("tx_key2")
		assert.NoError(t, err)
		assert.Equal(t, "value2_updated", val2)
	})

	t.Run("Transaction Retry on Conflict", func(t *testing.T) {
		key := "retry_key"
		err := client.Set(key, "initial", 0)
		assert.NoError(t, err)

		// Start a goroutine that will modify the key
		done := make(chan bool)
		go func() {
			for i := 0; i < 10; i++ { // Increase modifications to ensure conflict
				time.Sleep(10 * time.Millisecond)
				err := client.Set(key, fmt.Sprintf("concurrent_%d", i), 0)
				assert.NoError(t, err)
			}
			done <- true
		}()

		// Try transaction with retry
		maxRetries := 3
		var retryCount int
		for i := 0; i < maxRetries; i++ {
			retryCount++
			err = client.Watch(func(tx *redis.Tx) error {
				// Get current value
				_, err := tx.Get(context.Background(), key).Result()
				if err != nil {
					return err
				}

				// Simulate some processing time
				time.Sleep(50 * time.Millisecond)

				// Try to update
				_, err = tx.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
					pipe.Set(context.Background(), key, "transaction_update", 0)
					return nil
				})
				if err != nil {
					return err
				}

				// Verify if our value was actually set
				val, err := client.Get(key)
				if err != nil {
					return err
				}
				if val != "transaction_update" {
					return redis.TxFailedErr
				}
				return nil
			}, key)
			if err == nil {
				break
			}
		}
		<-done

		assert.True(t, retryCount > 1, "Transaction should have been retried")
		assert.Error(t, err, "Transaction should fail due to concurrent modification")
	})

	t.Run("Transaction with Invalid Operation", func(t *testing.T) {
		err := client.Watch(func(tx *redis.Tx) error {
			// Try to perform an invalid operation
			_, err := tx.TxPipelined(context.Background(), func(pipe redis.Pipeliner) error {
				pipe.HGet(context.Background(), "string_key", "field") // Invalid operation on string key
				return nil
			})
			return err
		}, "string_key")
		assert.Error(t, err)
	})
}

func TestClient_HealthCheck(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	config := getTestConfigHelper(mr.Addr())
	config.HealthCheckInterval = time.Second
	config.HealthCheckTimeout = 500 * time.Millisecond

	client, err := NewClient(config)
	require.NoError(t, err)
	defer client.Close()

	t.Run("Health Check", func(t *testing.T) {
		err := client.Ping()
		assert.NoError(t, err)

		// Wait for at least one health check cycle
		time.Sleep(2 * time.Second)
	})
}

func TestClient_ErrorHandling(t *testing.T) {
	t.Run("Closed Client", func(t *testing.T) {
		mr := newTestRedisHelper(t)
		defer mr.cleanup()

		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)

		client.Close()

		// Test all operations after closing
		_, err = client.Get("any_key")
		assert.Equal(t, ErrClientClosed, err)

		err = client.Set("any_key", "value", 0)
		assert.Equal(t, ErrClientClosed, err)

		err = client.Delete("any_key")
		assert.Equal(t, ErrClientClosed, err)

		_, err = client.Exists("any_key")
		assert.Equal(t, ErrClientClosed, err)

		err = client.HSet("hash", "field", "value")
		assert.Equal(t, ErrClientClosed, err)

		err = client.LPush("list", "value")
		assert.Equal(t, ErrClientClosed, err)

		err = client.SAdd("set", "member")
		assert.Equal(t, ErrClientClosed, err)

		err = client.ZAdd("zset", redis.Z{Score: 1, Member: "member"})
		assert.Equal(t, ErrClientClosed, err)

		err = client.Publish("channel", "message")
		assert.Equal(t, ErrClientClosed, err)

		err = client.FlushDB()
		assert.Equal(t, ErrClientClosed, err)
	})

	t.Run("Invalid Key Type", func(t *testing.T) {
		mr := newTestRedisHelper(t)
		defer mr.cleanup()

		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)
		defer client.Close()

		// Set string key
		err = client.Set("string_key", "value", time.Minute)
		assert.NoError(t, err)

		// Try hash operation on string key
		_, err = client.HGet("string_key", "field")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WRONGTYPE")

		// Set hash key
		err = client.HSet("hash_key", "field", "value")
		assert.NoError(t, err)

		// Try string operation on hash key
		_, err = client.Get("hash_key")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WRONGTYPE")

		// Set list key
		err = client.LPush("list_key", "value")
		assert.NoError(t, err)

		// Try set operation on list key
		err = client.SAdd("list_key", "member")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "WRONGTYPE")
	})

	t.Run("Invalid Arguments", func(t *testing.T) {
		mr := newTestRedisHelper(t)
		defer mr.cleanup()

		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)
		defer client.Close()

		// Test with empty key (Redis allows empty keys)
		err = client.Set("", "value", 0)
		assert.NoError(t, err)

		// Test with nil value (Redis allows nil values)
		err = client.Set("key", nil, 0)
		assert.NoError(t, err)

		// Test with negative score in sorted set
		err = client.ZAdd("zset", redis.Z{Score: -1, Member: "member"})
		assert.NoError(t, err) // Redis allows negative scores

		// Test with empty member in sorted set
		err = client.ZAdd("zset", redis.Z{Score: 1, Member: ""})
		assert.NoError(t, err) // Redis allows empty members
	})
}

func TestClient_ConnectionHandling(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	t.Run("Connection Pool", func(t *testing.T) {
		config := getTestConfigHelper(mr.Addr())
		config.PoolSize = 2
		config.MinIdleConns = 1

		client, err := NewClient(config)
		require.NoError(t, err)
		defer client.Close()

		// Test concurrent operations
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				key := fmt.Sprintf("key_%d", i)
				err := client.Set(key, i, time.Minute)
				assert.NoError(t, err)

				val, err := client.Get(key)
				assert.NoError(t, err)
				assert.Equal(t, strconv.Itoa(i), val)
			}(i)
		}
		wg.Wait()
	})

	t.Run("Health Check", func(t *testing.T) {
		config := getTestConfigHelper(mr.Addr())
		config.HealthCheckInterval = 100 * time.Millisecond
		config.HealthCheckTimeout = 50 * time.Millisecond

		client, err := NewClient(config)
		require.NoError(t, err)
		defer client.Close()

		// Wait for health checks
		time.Sleep(250 * time.Millisecond)

		// Verify client is still operational
		err = client.Set("health_check", "value", time.Minute)
		assert.NoError(t, err)
	})
}

func TestClient_PubSubOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("Multiple Subscribers", func(t *testing.T) {
		// Create multiple subscribers
		sub1 := client.Subscribe("channel1")
		defer sub1.Close()
		sub2 := client.Subscribe("channel1", "channel2")
		defer sub2.Close()

		// Wait for subscriptions
		_, err = sub1.Receive(context.Background())
		assert.NoError(t, err)
		_, err = sub2.Receive(context.Background())
		assert.NoError(t, err)

		// Publish messages
		err = client.Publish("channel1", "msg1")
		assert.NoError(t, err)
		err = client.Publish("channel2", "msg2")
		assert.NoError(t, err)

		// Check messages on first subscriber
		msg, err := sub1.ReceiveMessage(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "msg1", msg.Payload)

		// Check messages on second subscriber
		msg, err = sub2.ReceiveMessage(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "msg1", msg.Payload)
		msg, err = sub2.ReceiveMessage(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "msg2", msg.Payload)
	})
}

func TestClient_AdditionalOperations(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	t.Run("Info_with_Error", func(t *testing.T) {
		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)
		defer client.Close()

		// Test Info
		info, err := client.Info()
		assert.NoError(t, err)
		assert.NotEmpty(t, info)

		// Test error case
		client.Close()
		_, err = client.Info()
		assert.Error(t, err)
	})

	t.Run("Empty_Value_Operations", func(t *testing.T) {
		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)
		defer client.Close()

		// Test empty string value
		err = client.Set("empty_key", "", 0)
		assert.NoError(t, err)

		value, err := client.Get("empty_key")
		assert.NoError(t, err)
		assert.Equal(t, "", value)

		// Test nil value
		err = client.Set("nil_key", nil, 0)
		assert.NoError(t, err)

		value, err = client.Get("nil_key")
		assert.NoError(t, err)
		assert.Equal(t, "", value)
	})

	t.Run("Special_Characters_in_Keys", func(t *testing.T) {
		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)
		defer client.Close()

		specialKeys := []string{
			"key:with:colon",
			"key with space",
			"key_with_underscore",
			"key-with-dash",
			"key.with.dot",
			"!@#$%^&*()",
		}

		for _, key := range specialKeys {
			err := client.Set(key, "value", 0)
			assert.NoError(t, err)

			value, err := client.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, "value", value)

			err = client.Delete(key)
			assert.NoError(t, err)
		}
	})

	t.Run("Large_Values", func(t *testing.T) {
		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)
		defer client.Close()

		// Create a large value (1MB)
		largeValue := make([]byte, 1024*1024)
		for i := range largeValue {
			largeValue[i] = byte(i % 256)
		}
		largeStr := string(largeValue)

		err = client.Set("large_key", largeStr, 0)
		assert.NoError(t, err)

		value, err := client.Get("large_key")
		assert.NoError(t, err)
		assert.Equal(t, largeStr, value)

		err = client.Delete("large_key")
		assert.NoError(t, err)
	})

	t.Run("Concurrent_Key_Access", func(t *testing.T) {
		client, err := NewClient(getTestConfigHelper(mr.Addr()))
		require.NoError(t, err)
		defer client.Close()

		key := "concurrent_key"
		var wg sync.WaitGroup
		numGoroutines := 10

		// Set initial value
		err = client.Set(key, "0", 0)
		assert.NoError(t, err)

		// Channel to collect errors from goroutines
		errChan := make(chan error, numGoroutines)

		// Mutex to protect concurrent access
		var mu sync.Mutex

		// Concurrent increments
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				// Create a separate client for each goroutine
				concurrentClient, err := NewClient(getTestConfigHelper(mr.Addr()))
				if err != nil {
					errChan <- fmt.Errorf("failed to create client: %v", err)
					return
				}
				defer concurrentClient.Close()

				mu.Lock()
				defer mu.Unlock()

				// Get current value
				val, err := concurrentClient.Get(key)
				if err != nil {
					errChan <- fmt.Errorf("failed to get value: %v", err)
					return
				}

				current, err := strconv.Atoi(val)
				if err != nil {
					errChan <- fmt.Errorf("failed to parse value: %v", err)
					return
				}

				// Increment
				err = concurrentClient.Set(key, strconv.Itoa(current+1), 0)
				if err != nil {
					errChan <- fmt.Errorf("failed to set value: %v", err)
					return
				}
			}()
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errChan)

		// Check for any errors from goroutines
		var errors []error
		for err := range errChan {
			errors = append(errors, err)
		}
		assert.Empty(t, errors, "concurrent operations produced errors: %v", errors)

		// Verify final value
		value, err := client.Get(key)
		assert.NoError(t, err)
		finalValue, err := strconv.Atoi(value)
		assert.NoError(t, err)
		assert.True(t, finalValue > 0, "Value should have been incremented")
		assert.Equal(t, numGoroutines, finalValue, "Value should equal number of goroutines")
	})
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotNil(t, config)
	assert.Equal(t, StandaloneMode, config.Mode)
	assert.Equal(t, []string{"localhost:6379"}, config.Addresses)
	assert.Equal(t, "", config.Password)
	assert.Equal(t, 0, config.DB)
	assert.Equal(t, 10, config.PoolSize)
	assert.Equal(t, 5, config.MinIdleConns)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 5*time.Second, config.DialTimeout)
	assert.Equal(t, 3*time.Second, config.ReadTimeout)
	assert.Equal(t, 3*time.Second, config.WriteTimeout)
	assert.Equal(t, 2*time.Second, config.HealthCheckTimeout)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
}

func TestClient_WithContext(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test WithContext
	clientWithCtx := client.WithContext(ctx)
	assert.NotNil(t, clientWithCtx)

	// Wait for context to expire
	time.Sleep(200 * time.Millisecond)

	// Try operation with expired context
	err = clientWithCtx.Set("key", "value", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClient_KeyManagement(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("Expire and TTL", func(t *testing.T) {
		// Clean up any existing keys
		err := client.FlushAll()
		assert.NoError(t, err)

		// Set key with expiration
		err = client.Set("expire_key", "value", 0)
		assert.NoError(t, err)

		// Set expiration
		err = client.Expire("expire_key", time.Minute)
		assert.NoError(t, err)

		// Check TTL
		ttl, err := client.TTL("expire_key")
		assert.NoError(t, err)
		assert.True(t, ttl > 0)
		assert.True(t, ttl <= time.Minute)

		// Check TTL on non-existent key
		ttl, err = client.TTL("nonexistent_key")
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-2), ttl)

		// Check TTL on key without expiration
		err = client.Set("no_expire_key", "value", 0)
		assert.NoError(t, err)
		ttl, err = client.TTL("no_expire_key")
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(-1), ttl)
	})

	t.Run("Keys Pattern", func(t *testing.T) {
		// Clean up any existing keys
		err := client.FlushAll()
		assert.NoError(t, err)

		// Set multiple keys
		patterns := []string{"test:1", "test:2", "other:1", "other:2"}
		for _, key := range patterns {
			err := client.Set(key, "value", 0)
			assert.NoError(t, err)
		}

		// Get keys by pattern
		keys, err := client.Keys("test:*")
		assert.NoError(t, err)
		assert.Len(t, keys, 2)
		assert.Contains(t, keys, "test:1")
		assert.Contains(t, keys, "test:2")

		// Get all keys
		keys, err = client.Keys("*")
		assert.NoError(t, err)
		assert.Len(t, keys, 4)

		// Get keys with non-matching pattern
		keys, err = client.Keys("nonexistent:*")
		assert.NoError(t, err)
		assert.Empty(t, keys)
	})
}

func TestClient_DatabaseManagement(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("FlushAll", func(t *testing.T) {
		// Set some keys
		err := client.Set("key1", "value1", 0)
		assert.NoError(t, err)
		err = client.Set("key2", "value2", 0)
		assert.NoError(t, err)

		// Flush all DBs
		err = client.FlushAll()
		assert.NoError(t, err)

		// Verify keys are gone
		exists, err := client.Exists("key1")
		assert.NoError(t, err)
		assert.False(t, exists)
		exists, err = client.Exists("key2")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestClient_GetClient(t *testing.T) {
	mr := newTestRedisHelper(t)
	defer mr.cleanup()

	client, err := NewClient(getTestConfigHelper(mr.Addr()))
	require.NoError(t, err)
	defer client.Close()

	t.Run("GetClient", func(t *testing.T) {
		// Get underlying client
		redisClient := client.GetClient()
		assert.NotNil(t, redisClient)

		// Test the client works
		err := redisClient.Set(context.Background(), "test_key", "test_value", 0).Err()
		assert.NoError(t, err)

		val, err := redisClient.Get(context.Background(), "test_key").Result()
		assert.NoError(t, err)
		assert.Equal(t, "test_value", val)

		// Test client is the same instance
		redisClient2 := client.GetClient()
		assert.Same(t, redisClient, redisClient2)
	})
}
