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
	Host              string        `mapstructure:"host" yaml:"host" env:"HOST"`
	Port              int           `mapstructure:"port" yaml:"port" env:"PORT"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT"`
	KeepAliveTime     time.Duration `mapstructure:"keepalive_time" yaml:"keepalive_time" env:"KEEPALIVE_TIME"`
	KeepAliveTimeout  time.Duration `mapstructure:"keepalive_timeout" yaml:"keepalive_timeout" env:"KEEPALIVE_TIMEOUT"`
	MaxConnectionIdle time.Duration `mapstructure:"max_connection_idle" yaml:"max_connection_idle" env:"MAX_CONNECTION_IDLE"`
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
	URI         string                          `mapstructure:"uri" yaml:"uri" env:"URI"`
	Database    string                          `mapstructure:"database" yaml:"database" env:"DATABASE"`
	Options     MongoDBOptionsConfig            `mapstructure:"options" yaml:"options"`
	Collections map[string]MongoDBCollectionConfig `mapstructure:"collections" yaml:"collections"`
}

// MongoDBOptionsConfig holds MongoDB connection options
type MongoDBOptionsConfig struct {
	MaxPoolSize              int  `mapstructure:"max_pool_size" yaml:"max_pool_size" env:"MAX_POOL_SIZE"`
	MinPoolSize              int  `mapstructure:"min_pool_size" yaml:"min_pool_size" env:"MIN_POOL_SIZE"`
	MaxIdleTimeMS           int  `mapstructure:"max_idle_time_ms" yaml:"max_idle_time_ms" env:"MAX_IDLE_TIME_MS"`
	ConnectTimeoutMS        int  `mapstructure:"connect_timeout_ms" yaml:"connect_timeout_ms" env:"CONNECT_TIMEOUT_MS"`
	SocketTimeoutMS         int  `mapstructure:"socket_timeout_ms" yaml:"socket_timeout_ms" env:"SOCKET_TIMEOUT_MS"`
	ServerSelectionTimeoutMS int  `mapstructure:"server_selection_timeout_ms" yaml:"server_selection_timeout_ms" env:"SERVER_SELECTION_TIMEOUT_MS"`
	RetryWrites             bool `mapstructure:"retry_writes" yaml:"retry_writes" env:"RETRY_WRITES"`
	RetryReads              bool `mapstructure:"retry_reads" yaml:"retry_reads" env:"RETRY_READS"`
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

// Config holds base configuration for services
type Config struct {
	RootConfig      `mapstructure:",squash" yaml:",inline"`
	Service         ServiceConfig     `mapstructure:"service" yaml:"service"`
	Server          ServerConfig      `mapstructure:"server" yaml:"server"`
	Metrics         MetricsConfig     `mapstructure:"metrics" yaml:"metrics"`
	Logging         LoggingConfig     `mapstructure:"logging" yaml:"logging"`
	MongoDB         MongoDBConfig     `mapstructure:"mongodb" yaml:"mongodb"`
	ServicePorts    ServicePortsConfig `mapstructure:"service_ports" yaml:"service_ports"`
}

// LoadConfig reads configuration from root config and environment variables
func LoadConfig(serviceName string, envPrefix string, configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Configure Viper for environment variables
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv() // read in environment variables that match

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

	// Create a new Viper instance for final config
	finalConfig := viper.New()

	// Load root level settings
	finalConfig.Set("environment", v.Get("environment"))
	finalConfig.Set("region", v.Get("region"))
	finalConfig.Set("project", v.Get("project"))

	// Load shared settings
	finalConfig.Set("metrics", v.Get("metrics"))
	finalConfig.Set("logging", v.Get("logging"))
	finalConfig.Set("mongodb", v.Get("mongodb"))
	finalConfig.Set("service_ports", v.Get("service_ports"))

	// Get service-specific config if it exists
	if v.IsSet(serviceName) {
		serviceConfig := v.Get(serviceName)
		if serviceConfig != nil {
			// Set service-specific configuration
			finalConfig.Set("service", v.Get(fmt.Sprintf("%s.service", serviceName)))
			finalConfig.Set("server", v.Get(fmt.Sprintf("%s.server", serviceName)))
			
			// Set any additional service-specific settings
			serviceMap := v.GetStringMap(serviceName)
			for key, value := range serviceMap {
				if key != "service" && key != "server" {
					finalConfig.Set(key, value)
				}
			}
		}
	} else {
		// If no service-specific config exists, use shared server config
		if v.IsSet("server") {
			finalConfig.Set("server", v.Get("server"))
		}
	}

	// Bind environment variables for each config section
	bindEnvs(finalConfig, Config{})

	var config Config
	if err := finalConfig.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set service name if not already set
	if config.Service.Name == "" {
		config.Service.Name = serviceName
	}

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

// getProjectRoot tries to find the project root directory
func getProjectRoot(startDir string) (string, error) {
	currentDir := startDir
	maxDepth := 5 // Maximum number of parent directories to check

	for i := 0; i < maxDepth; i++ {
		// Check for common project root indicators
		indicators := []string{
			"go.mod",           // Go projects
			".git",             // Git repository
			"package.json",     // Node.js projects
			"README.md",        // Common in project roots
		}

		for _, indicator := range indicators {
			if _, err := os.Stat(filepath.Join(currentDir, indicator)); err == nil {
				return currentDir, nil
			}
		}

		// Get parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// We've reached the root directory
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("could not find project root in parent directories")
}

// bindEnvs recursively binds environment variables to Viper
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

		if fieldv.Kind() == reflect.Struct {
			bindEnvs(v, fieldv.Interface(), path...)
		} else {
			// Bind both mapstructure path and env tag if present
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
	v.SetDefault("server.keepalive_time", "60s")
	v.SetDefault("server.keepalive_timeout", "20s")
	v.SetDefault("server.max_connection_idle", "180s")

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

	// Service ports defaults
	v.SetDefault("service_ports.worker", 50051)
	v.SetDefault("service_ports.gateway", 50052)
	v.SetDefault("service_ports.workflow_manager", 50053)
	v.SetDefault("service_ports.tools", 50054)
} 