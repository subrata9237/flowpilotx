package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// RootConfig holds root level configuration
type RootConfig struct {
	Environment string `mapstructure:"environment" yaml:"environment" env:"ENVIRONMENT"`
	Region      string `mapstructure:"region" yaml:"region" env:"REGION"`
	Project     string `mapstructure:"project" yaml:"project" env:"PROJECT"`
}

// ServiceConfig holds service-specific configuration
type ServiceConfig struct {
	Name        string `mapstructure:"name" yaml:"name" env:"NAME"`
	Version     string `mapstructure:"version" yaml:"version" env:"VERSION"`
	Environment string `mapstructure:"environment" yaml:"environment" env:"ENVIRONMENT"`
}

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host" yaml:"host" env:"HOST"`
	Port            int           `mapstructure:"port" yaml:"port" env:"PORT"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" yaml:"read_timeout" env:"READ_TIMEOUT"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" yaml:"write_timeout" env:"WRITE_TIMEOUT"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" yaml:"idle_timeout" env:"IDLE_TIMEOUT"`
}

// WebSocketConfig holds WebSocket-related configuration
type WebSocketConfig struct {
	Path             string        `mapstructure:"path" yaml:"path" env:"PATH"`
	ReadBufferSize   int           `mapstructure:"read_buffer_size" yaml:"read_buffer_size" env:"READ_BUFFER_SIZE"`
	WriteBufferSize  int           `mapstructure:"write_buffer_size" yaml:"write_buffer_size" env:"WRITE_BUFFER_SIZE"`
	HandshakeTimeout time.Duration `mapstructure:"handshake_timeout" yaml:"handshake_timeout" env:"HANDSHAKE_TIMEOUT"`
	PingInterval     time.Duration `mapstructure:"ping_interval" yaml:"ping_interval" env:"PING_INTERVAL"`
	PongWait         time.Duration `mapstructure:"pong_wait" yaml:"pong_wait" env:"PONG_WAIT"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	Version         string `mapstructure:"version" yaml:"version" env:"VERSION"`
	BasePath        string `mapstructure:"base_path" yaml:"base_path" env:"BASE_PATH"`
	HealthCheckPath string `mapstructure:"health_check_path" yaml:"health_check_path" env:"HEALTH_CHECK_PATH"`
	VersionPath     string `mapstructure:"version_path" yaml:"version_path" env:"VERSION_PATH"`
}

// CORSConfig holds CORS-related configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins" yaml:"allowed_origins" env:"ALLOWED_ORIGINS"`
	AllowedMethods []string `mapstructure:"allowed_methods" yaml:"allowed_methods" env:"ALLOWED_METHODS"`
	AllowedHeaders []string `mapstructure:"allowed_headers" yaml:"allowed_headers" env:"ALLOWED_HEADERS"`
	MaxAge         int      `mapstructure:"max_age" yaml:"max_age" env:"MAX_AGE"`
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled" yaml:"enabled" env:"ENABLED"`
	RequestsPerSecond int  `mapstructure:"requests_per_second" yaml:"requests_per_second" env:"REQUESTS_PER_SECOND"`
	Burst             int  `mapstructure:"burst" yaml:"burst" env:"BURST"`
}

// WorkerConfig holds worker-specific configuration
type WorkerConfig struct {
	NumWorkers         int      `mapstructure:"num_workers" yaml:"num_workers" env:"NUM_WORKERS"`
	MaxQueuesPerWorker int      `mapstructure:"max_queues_per_worker" yaml:"max_queues_per_worker" env:"MAX_QUEUES_PER_WORKER"`
	InitialQueues      []string `mapstructure:"initial_queues" yaml:"initial_queues" env:"INITIAL_QUEUES"`
	HeartbeatInterval  int      `mapstructure:"heartbeat_interval" yaml:"heartbeat_interval" env:"HEARTBEAT_INTERVAL"`
}

// RetryConfig holds retry-related configuration
type RetryConfig struct {
	MaxAttempts     int           `mapstructure:"max_attempts" yaml:"max_attempts" env:"MAX_ATTEMPTS"`
	InitialInterval time.Duration `mapstructure:"initial_interval" yaml:"initial_interval" env:"INITIAL_INTERVAL"`
	MaxInterval     time.Duration `mapstructure:"max_interval" yaml:"max_interval" env:"MAX_INTERVAL"`
}

// GatewayConfig holds gateway service configuration
type GatewayConfig struct {
	Service   ServiceConfig   `mapstructure:"service" yaml:"service"`
	Server    ServerConfig    `mapstructure:"server" yaml:"server"`
	WebSocket WebSocketConfig `mapstructure:"websocket" yaml:"websocket"`
	API       APIConfig       `mapstructure:"api" yaml:"api"`
	CORS      CORSConfig      `mapstructure:"cors" yaml:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit" yaml:"rate_limit"`
	MongoDB   MongoDBConfig   `mapstructure:"mongodb" yaml:"mongodb"`
	Temporal  TemporalConfig  `mapstructure:"temporal" yaml:"temporal"`
}

// TemporalConfig holds Temporal-related configuration
type TemporalConfig struct {
	HostPort  string                 `mapstructure:"host_port" yaml:"host_port" env:"HOST_PORT"`
	Namespace string                 `mapstructure:"namespace" yaml:"namespace" env:"NAMESPACE"`
	TaskQueue string                 `mapstructure:"task_queue" yaml:"task_queue" env:"TASK_QUEUE"`
	Client    TemporalClientConfig   `mapstructure:"client" yaml:"client"`
	Workflow  TemporalWorkflowConfig `mapstructure:"workflow" yaml:"workflow"`
}

// TemporalClientConfig holds Temporal client configuration
type TemporalClientConfig struct {
	ConnectionTimeout  string  `mapstructure:"connection_timeout" yaml:"connection_timeout" env:"CONNECTION_TIMEOUT"`
	RPCTimeout         string  `mapstructure:"rpc_timeout" yaml:"rpc_timeout" env:"RPC_TIMEOUT"`
	RPCLongPollTimeout string  `mapstructure:"rpc_long_poll_timeout" yaml:"rpc_long_poll_timeout" env:"RPC_LONG_POLL_TIMEOUT"`
	RPCQueryTimeout    string  `mapstructure:"rpc_query_timeout" yaml:"rpc_query_timeout" env:"RPC_QUERY_TIMEOUT"`
	MaxRetryAttempts   int     `mapstructure:"max_retry_attempts" yaml:"max_retry_attempts" env:"MAX_RETRY_ATTEMPTS"`
	InitialInterval    string  `mapstructure:"initial_interval" yaml:"initial_interval" env:"INITIAL_INTERVAL"`
	MaxInterval        string  `mapstructure:"max_interval" yaml:"max_interval" env:"MAX_INTERVAL"`
	BackoffCoefficient float64 `mapstructure:"backoff_coefficient" yaml:"backoff_coefficient" env:"BACKOFF_COEFFICIENT"`
	Identity           string  `mapstructure:"identity" yaml:"identity" env:"IDENTITY"`
}

// TemporalWorkflowConfig holds Temporal workflow configuration
type TemporalWorkflowConfig struct {
	ExecutionTimeout      string                    `mapstructure:"execution_timeout" yaml:"execution_timeout" env:"EXECUTION_TIMEOUT"`
	RunTimeout            string                    `mapstructure:"run_timeout" yaml:"run_timeout" env:"RUN_TIMEOUT"`
	TaskTimeout           string                    `mapstructure:"task_timeout" yaml:"task_timeout" env:"TASK_TIMEOUT"`
	WorkflowIDReusePolicy string                    `mapstructure:"workflow_id_reuse_policy" yaml:"workflow_id_reuse_policy" env:"WORKFLOW_ID_REUSE_POLICY"`
	RetryPolicy           TemporalRetryPolicyConfig `mapstructure:"retry_policy" yaml:"retry_policy"`
}

// TemporalRetryPolicyConfig holds Temporal retry policy configuration
type TemporalRetryPolicyConfig struct {
	InitialInterval        string   `mapstructure:"initial_interval" yaml:"initial_interval" env:"INITIAL_INTERVAL"`
	MaxInterval            string   `mapstructure:"max_interval" yaml:"max_interval" env:"MAX_INTERVAL"`
	MaxAttempts            int      `mapstructure:"max_attempts" yaml:"max_attempts" env:"MAX_ATTEMPTS"`
	BackoffCoefficient     float64  `mapstructure:"backoff_coefficient" yaml:"backoff_coefficient" env:"BACKOFF_COEFFICIENT"`
	NonRetryableErrorTypes []string `mapstructure:"non_retryable_error_types" yaml:"non_retryable_error_types" env:"NON_RETRYABLE_ERROR_TYPES"`
}

// WorkerServiceConfig holds worker service configuration
type WorkerServiceConfig struct {
	Service ServiceConfig `mapstructure:"service" yaml:"service"`
	Server  ServerConfig  `mapstructure:"server" yaml:"server"`
	Worker  WorkerConfig  `mapstructure:"worker" yaml:"worker"`
}

// MetricsConfig holds metrics-related configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled" env:"ENABLED"`
	Host    string `mapstructure:"host" yaml:"host" env:"HOST"`
	Port    int    `mapstructure:"port" yaml:"port" env:"PORT"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level" yaml:"level" env:"LEVEL"`
	Format     string `mapstructure:"format" yaml:"format" env:"FORMAT"`
	OutputPath string `mapstructure:"output_path" yaml:"output_path" env:"OUTPUT_PATH"`
}

// MongoDBConfig holds MongoDB-related configuration
type MongoDBConfig struct {
	URI         string                             `mapstructure:"uri" yaml:"uri" env:"URI"`
	Database    string                             `mapstructure:"database" yaml:"database" env:"DATABASE"`
	Username    string                             `mapstructure:"username" yaml:"username" env:"USERNAME"`
	Password    string                             `mapstructure:"password" yaml:"password" env:"PASSWORD"`
	Options     MongoDBOptionsConfig               `mapstructure:"options" yaml:"options"`
	Collections map[string]MongoDBCollectionConfig `mapstructure:"collections" yaml:"collections"`
}

// MongoDBOptionsConfig holds MongoDB connection options
type MongoDBOptionsConfig struct {
	MaxPoolSize              int  `mapstructure:"max_pool_size" yaml:"max_pool_size" env:"MAX_POOL_SIZE"`
	MinPoolSize              int  `mapstructure:"min_pool_size" yaml:"min_pool_size" env:"MIN_POOL_SIZE"`
	MaxIdleTimeMS            int  `mapstructure:"max_idle_time_ms" yaml:"max_idle_time_ms" env:"MAX_IDLE_TIME_MS"`
	ConnectTimeoutMS         int  `mapstructure:"connect_timeout_ms" yaml:"connect_timeout_ms" env:"CONNECT_TIMEOUT_MS"`
	SocketTimeoutMS          int  `mapstructure:"socket_timeout_ms" yaml:"socket_timeout_ms" env:"SOCKET_TIMEOUT_MS"`
	ServerSelectionTimeoutMS int  `mapstructure:"server_selection_timeout_ms" yaml:"server_selection_timeout_ms" env:"SERVER_SELECTION_TIMEOUT_MS"`
	RetryWrites              bool `mapstructure:"retry_writes" yaml:"retry_writes" env:"RETRY_WRITES"`
	RetryReads               bool `mapstructure:"retry_reads" yaml:"retry_reads" env:"RETRY_READS"`
}

// MongoDBCollectionConfig holds collection-specific configuration
type MongoDBCollectionConfig struct {
	Indexes []MongoDBIndexConfig `mapstructure:"indexes" yaml:"indexes"`
}

// MongoDBIndexConfig holds index configuration
type MongoDBIndexConfig struct {
	Keys    []string               `mapstructure:"keys" yaml:"keys"`
	Options map[string]interface{} `mapstructure:"options" yaml:"options"`
}

// ServicePortsConfig holds service port mappings
type ServicePortsConfig struct {
	Worker          int `mapstructure:"worker" yaml:"worker" env:"WORKER_PORT"`
	Gateway         int `mapstructure:"gateway" yaml:"gateway" env:"GATEWAY_PORT"`
	WorkflowManager int `mapstructure:"workflow_manager" yaml:"workflow_manager" env:"WORKFLOW_MANAGER_PORT"`
	Tools           int `mapstructure:"tools" yaml:"tools" env:"TOOLS_PORT"`
}

// ActivityConfig holds activity-related configuration
type ActivityConfig struct {
	Name           string        `mapstructure:"name" yaml:"name"`
	Type           string        `mapstructure:"type" yaml:"type"`
	Queue          string        `mapstructure:"queue" yaml:"queue"`
	RetryCount     int           `mapstructure:"retry_count" yaml:"retry_count"`
	RetryInterval  time.Duration `mapstructure:"retry_interval" yaml:"retry_interval"`
	TimeoutSeconds int           `mapstructure:"timeout_seconds" yaml:"timeout_seconds"`
	Description    string        `mapstructure:"description" yaml:"description"`
	InputParams    []ParamConfig `mapstructure:"input_params" yaml:"input_params"`
}

// ParamConfig holds parameter configuration
type ParamConfig struct {
	Name string `mapstructure:"name" yaml:"name"`
	Type string `mapstructure:"type" yaml:"type"`
}

// ActivitiesConfig holds all activity configurations
type ActivitiesConfig struct {
	Activities []ActivityConfig `mapstructure:"activities" yaml:"activities"`
}

// Config holds base configuration for services
type Config struct {
	RootConfig           `mapstructure:",squash" yaml:",inline"`
	FlowpilotxGateway    GatewayConfig       `mapstructure:"flowpilotx_gateway" yaml:"flowpilotx_gateway"`
	FlowpilotxWorker     WorkerServiceConfig `mapstructure:"flowpilotx_worker" yaml:"flowpilotx_worker"`
	Service              ServiceConfig       `mapstructure:"service" yaml:"service"`
	Server               ServerConfig        `mapstructure:"server" yaml:"server"`
	Metrics              MetricsConfig       `mapstructure:"metrics" yaml:"metrics"`
	Logging              LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	MongoDB              MongoDBConfig       `mapstructure:"mongodb" yaml:"mongodb"`
	ServicePorts         ServicePortsConfig  `mapstructure:"service_ports" yaml:"service_ports"`
	FlowpilotxActivities ActivitiesConfig    `mapstructure:"flowpilotx_activities" yaml:"flowpilotx_activities"`
}

// LoadConfig reads configuration from root config and environment variables
func LoadConfig(serviceName string, envPrefix string, configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)
	// Configure Viper
	v.SetConfigType("yaml")       // Set the config type explicitly
	v.SetTypeByDefaultValue(true) // Infer type from default values

	// Configure Viper for environment variables
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv() // read in environment variables that match

	// Enable environment variable override
	v.AllowEmptyEnv(true)
	v.SetEnvPrefix(strings.ToUpper(envPrefix))

	// Configure array delimiter for environment variables
	v.SetTypeByDefaultValue(true)
	v.SetEnvKeyReplacer(strings.NewReplacer(
		".", "_",
		"[", "_",
		"]", "",
	))

	var configFile string

	// If configPath is provided, use it directly
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			configFile = configPath
		} else {
			return nil, fmt.Errorf("specified config file not found at %s: %w", configPath, err)
		}
	} else {
		// Get current working directory
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		// Try to find config.yaml by traversing up the directory tree
		configFile = findConfigFile(currentDir)
	}

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error reading config file %s: %w", configFile, err)
		}
		fmt.Printf("Using config file: %s\n", configFile)
	} else {
		fmt.Println("No config file found, using environment variables and defaults")
	}

	// Bind environment variables for each config section
	bindEnvs(v, Config{})

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set service name if not already set
	if config.Service.Name == "" {
		config.Service.Name = serviceName
	}

	// Debug print the CORS configuration
	fmt.Printf("Loaded CORS config: %+v\n", config.FlowpilotxGateway.CORS)

	return &config, nil
}

// LoadConfigWithPath is a convenience function that loads config from a specific path
func LoadConfigWithPath(configPath string, serviceName string, envPrefix string) (*Config, error) {
	return LoadConfig(serviceName, envPrefix, configPath)
}

// LoadConfigAuto is a convenience function that automatically finds the config file
func LoadConfigAuto(serviceName string, envPrefix string) (*Config, error) {
	return LoadConfig(serviceName, envPrefix, "")
}

// findConfigFile searches for config.yaml by traversing up the directory tree
func findConfigFile(startDir string) string {
	currentDir := startDir
	maxDepth := 5 // Maximum number of parent directories to check

	for i := 0; i < maxDepth; i++ {
		// Try current directory
		configPath := filepath.Join(currentDir, "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			relPath, err := filepath.Rel(startDir, configPath)
			if err == nil {
				fmt.Printf("Found config file at: %s (relative: %s)\n", configPath, relPath)
			}
			return configPath
		}

		// Get parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We've reached the root directory
			break
		}
		currentDir = parentDir
	}

	return ""
}

// handleArrayEnvVars processes environment variables for array types
func handleArrayEnvVars(v *viper.Viper, key string, envKey string) {
	// Check if environment variable exists
	if val := os.Getenv(envKey); val != "" {
		// Split the value by comma and trim spaces
		values := strings.Split(val, ",")
		for i := range values {
			values[i] = strings.TrimSpace(values[i])
		}
		v.Set(key, values)
	}
}

func bindEnvs(v *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		fieldv := ifv.Field(i)
		t := ift.Field(i)
		name := t.Tag.Get("mapstructure")
		env := t.Tag.Get("env")

		// Skip if no mapstructure tag
		if name == "" {
			continue
		}

		path := append(parts, name)
		key := strings.Join(path, ".")

		switch fieldv.Kind() {
		case reflect.Struct:
			bindEnvs(v, fieldv.Interface(), path...)
		case reflect.Slice, reflect.Array:
			// For slice types, bind both the entire slice and individual elements
			v.BindEnv(key)
			if env != "" {
				envKey := fmt.Sprintf("%s_%s", strings.ToUpper(strings.Join(parts, "_")), env)
				v.BindEnv(key, envKey)
				// Handle array environment variables
				handleArrayEnvVars(v, key, envKey)
			}
		default:
			// For all other types
			v.BindEnv(key)
			if env != "" {
				envKey := fmt.Sprintf("%s_%s", strings.ToUpper(strings.Join(parts, "_")), env)
				v.BindEnv(key, envKey)
			}
		}
	}
}

func setDefaults(v *viper.Viper) {
	// Root level defaults
	v.SetDefault("environment", "development")
	v.SetDefault("region", "us-west-1")
	v.SetDefault("project", "flowpilotx")

	// Service defaults
	v.SetDefault("service.name", "")
	v.SetDefault("service.version", "1.0.0")
	v.SetDefault("service.environment", "development")

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 50051)
	v.SetDefault("server.shutdown_timeout", "30s")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "180s")

	// MongoDB defaults
	v.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	v.SetDefault("mongodb.database", "flowpilotx")
	v.SetDefault("mongodb.options.max_pool_size", 100)
	v.SetDefault("mongodb.options.min_pool_size", 10)
	v.SetDefault("mongodb.options.max_idle_time_ms", 30000)
	v.SetDefault("mongodb.options.connect_timeout_ms", 5000)
	v.SetDefault("mongodb.options.socket_timeout_ms", 10000)
	v.SetDefault("mongodb.options.server_selection_timeout_ms", 5000)
	v.SetDefault("mongodb.options.retry_writes", true)
	v.SetDefault("mongodb.options.retry_reads", true)

	// Metrics defaults
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.host", "0.0.0.0")
	v.SetDefault("metrics.port", 9090)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output_path", "stdout")

	// Gateway MongoDB defaults
	v.SetDefault("flowpilotx_gateway.mongodb.uri", "mongodb://localhost:27017")
	v.SetDefault("flowpilotx_gateway.mongodb.database", "flowpilotx")
	v.SetDefault("flowpilotx_gateway.mongodb.username", "")
	v.SetDefault("flowpilotx_gateway.mongodb.password", "")
	v.SetDefault("flowpilotx_gateway.mongodb.options.max_pool_size", 100)
	v.SetDefault("flowpilotx_gateway.mongodb.options.min_pool_size", 10)
	v.SetDefault("flowpilotx_gateway.mongodb.options.max_idle_time_ms", 30000)
	v.SetDefault("flowpilotx_gateway.mongodb.options.connect_timeout_ms", 5000)
	v.SetDefault("flowpilotx_gateway.mongodb.options.socket_timeout_ms", 10000)
	v.SetDefault("flowpilotx_gateway.mongodb.options.server_selection_timeout_ms", 5000)
	v.SetDefault("flowpilotx_gateway.mongodb.options.retry_writes", true)
	v.SetDefault("flowpilotx_gateway.mongodb.options.retry_reads", true)

	// Gateway Temporal defaults
	v.SetDefault("flowpilotx_gateway.temporal.host_port", "localhost:7233")

	//add all flowpilotx_gateway config
	v.SetDefault("flowpilotx_gateway.service.name", "flowpilotx-gateway")
	v.SetDefault("flowpilotx_gateway.service.version", "1.0.0")
	v.SetDefault("flowpilotx_gateway.service.environment", "development")
	v.SetDefault("flowpilotx_gateway.server.host", "0.0.0.0")
	v.SetDefault("flowpilotx_gateway.server.port", 8080)
	v.SetDefault("flowpilotx_gateway.server.shutdown_timeout", "30s")
	v.SetDefault("flowpilotx_gateway.server.read_timeout", "15s")
	v.SetDefault("flowpilotx_gateway.server.write_timeout", "15s")
	v.SetDefault("flowpilotx_gateway.server.idle_timeout", "60s")
	v.SetDefault("flowpilotx_gateway.websocket.path", "/ws/flowpilotx-gateway")
	v.SetDefault("flowpilotx_gateway.websocket.read_buffer_size", 1024)
	v.SetDefault("flowpilotx_gateway.websocket.write_buffer_size", 1024)
	v.SetDefault("flowpilotx_gateway.websocket.handshake_timeout", "10s")
	v.SetDefault("flowpilotx_gateway.websocket.ping_interval", "30s")
	v.SetDefault("flowpilotx_gateway.websocket.pong_wait", "60s")
	v.SetDefault("flowpilotx_gateway.api.version", "v1")
	v.SetDefault("flowpilotx_gateway.api.base_path", "/api/flowpilotx-gateway")
	v.SetDefault("flowpilotx_gateway.api.health_check_path", "/health")
	v.SetDefault("flowpilotx_gateway.api.version_path", "/version")

	// CORS defaults
	v.SetDefault("flowpilotx_gateway.cors.allowed_origins", []string{"*"})
	v.SetDefault("flowpilotx_gateway.cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("flowpilotx_gateway.cors.allowed_headers", []string{"Content-Type", "Authorization"})
	v.SetDefault("flowpilotx_gateway.cors.max_age", 300)

	// Service ports defaults
	v.SetDefault("service_ports.worker", 50051)
	v.SetDefault("service_ports.gateway", 50052)
	v.SetDefault("service_ports.workflow_manager", 50053)
	v.SetDefault("service_ports.tools", 50054)
}
