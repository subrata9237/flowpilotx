package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := []byte(`
mongodb:
  development:
    uri: "mongodb://localhost:27017"
    database: "devdb"
    connect_timeout: 10s
    operation_timeout: 5s
    max_pool_size: 10
    min_pool_size: 1
    retry_writes: true
    retry_reads: true
    direct: false
redis:
  development:
    host: "localhost"
    port: 6379
    database: 0
    pool_size: 10
    min_idle_conns: 5
    max_retries: 3
    dial_timeout: 5s
    read_timeout: 3s
    write_timeout: 3s
    pool_timeout: 4s
    idle_timeout: 300s
    idle_check_frequency: 60s
`)

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	// Set environment variables
	os.Setenv("FLOWPILOT_ENV", "development")
	os.Setenv("FLOWPILOT_MONGODB_USERNAME", "devuser")
	os.Setenv("FLOWPILOT_MONGODB_PASSWORD", "devpass")
	defer func() {
		os.Unsetenv("FLOWPILOT_ENV")
		os.Unsetenv("FLOWPILOT_MONGODB_USERNAME")
		os.Unsetenv("FLOWPILOT_MONGODB_PASSWORD")
	}()

	// Test loading config
	config, err := LoadConfig(tmpfile.Name())
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test MongoDB config
	assert.Equal(t, "mongodb://development:27017", config.MongoDB.URI)
	assert.Equal(t, "devuser", config.MongoDB.Username)
	assert.Equal(t, "devpass", config.MongoDB.Password)
	assert.Equal(t, "flowpilot", config.MongoDB.Database)
	assert.Equal(t, 10*time.Second, config.MongoDB.ConnectTimeout)
	assert.Equal(t, uint64(10), config.MongoDB.MaxPoolSize)
	assert.False(t, config.MongoDB.Direct)

	// Test Redis config
	assert.Equal(t, "redis-development", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)
	assert.Equal(t, 0, config.Redis.Database)
	assert.Equal(t, 10, config.Redis.PoolSize)
	assert.Equal(t, 5, config.Redis.MinIdleConns)
	assert.Equal(t, 5*time.Second, config.Redis.DialTimeout)
}

func TestEnvironmentOverrides(t *testing.T) {
	// Create a temporary config file with default values
	content := []byte(`
mongodb:
  development:
    uri: "mongodb://localhost:27017"
    database: "devdb"
redis:
  development:
    host: "localhost"
    port: 6379
    database: 0
`)

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	// Set environment variables
	envVars := map[string]string{
		"FLOWPILOT_ENV":              "development",
		"FLOWPILOT_MONGODB_URI":      "mongodb://dev-override:27017",
		"FLOWPILOT_MONGODB_USERNAME": "override-user",
		"FLOWPILOT_MONGODB_PASSWORD": "override-pass",
		"FLOWPILOT_MONGODB_DATABASE": "override-db",
		"FLOWPILOT_REDIS_HOST":       "redis-override",
		"FLOWPILOT_REDIS_PORT":       "6380",
		"FLOWPILOT_REDIS_DATABASE":   "1",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	// Load config and test overrides
	config, err := LoadConfig(tmpfile.Name())
	require.NoError(t, err)

	// Test MongoDB overrides
	assert.Equal(t, "mongodb://dev-override:27017", config.MongoDB.URI)
	assert.Equal(t, "override-user", config.MongoDB.Username)
	assert.Equal(t, "override-pass", config.MongoDB.Password)
	assert.Equal(t, "override-db", config.MongoDB.Database)

	// Test Redis overrides
	assert.Equal(t, "redis-override", config.Redis.Host)
	assert.Equal(t, 6379, config.Redis.Port)
	assert.Equal(t, 0, config.Redis.Database)
}

func TestGetCurrentEnvironment(t *testing.T) {
	// Test default environment
	os.Unsetenv("FLOWPILOT_ENV")
	env := getCurrentEnvironment()
	assert.Equal(t, DefaultEnvironment, env)

	// Test custom environment
	os.Setenv("FLOWPILOT_ENV", "development")
	defer os.Unsetenv("FLOWPILOT_ENV")
	env = getCurrentEnvironment()
	assert.Equal(t, Development, env)
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	assert.Error(t, err)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := []byte(`invalid: yaml: content`)
	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	_, err = LoadConfig(tmpfile.Name())
	assert.Error(t, err)
}

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	assert.NotEmpty(t, path)
}

func TestMustLoadConfig_Panic(t *testing.T) {
	assert.Panics(t, func() {
		MustLoadConfig("nonexistent.yaml")
	})
}

func TestDefaultValues(t *testing.T) {
	// Create a minimal config file
	content := []byte(`
mongodb:
  development:
    uri: "mongodb://localhost:27017"
redis:
  development:
    host: "localhost"
    port: 6379
`)

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	os.Setenv("FLOWPILOT_ENV", "development")
	defer os.Unsetenv("FLOWPILOT_ENV")

	// Load config and test defaults
	config, err := LoadConfig(tmpfile.Name())
	require.NoError(t, err)

	// Test MongoDB defaults
	assert.Equal(t, "flowpilot", config.MongoDB.Database)
	assert.Equal(t, uint64(0), config.MongoDB.MaxPoolSize)
	assert.Equal(t, uint64(0), config.MongoDB.MinPoolSize)
	assert.False(t, config.MongoDB.RetryWrites)
	assert.False(t, config.MongoDB.RetryReads)

	// Test Redis defaults
	assert.Equal(t, 0, config.Redis.Database)
	assert.Equal(t, 0, config.Redis.PoolSize)
	assert.Equal(t, 0, config.Redis.MinIdleConns)
	assert.Equal(t, 0, config.Redis.MaxRetries)
	assert.Equal(t, time.Duration(0), config.Redis.DialTimeout)
} 