package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	libctx "github.com/flowpilotx/libs/context"
)

// Environment constants
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// Config holds the logger configuration
type Config struct {
	// Level is the minimum enabled logging level
	Level string `yaml:"level" json:"level"`
	// Format specifies the output format ("json" or "console")
	Format string `yaml:"format" json:"format"`
	// OutputPaths is a list of URLs or file paths to write logging output to
	OutputPaths []string `yaml:"output_paths" json:"output_paths"`
	// ErrorOutputPaths is a list of URLs to write internal logger errors to
	ErrorOutputPaths []string `yaml:"error_output_paths" json:"error_output_paths"`
	// Development puts the logger in development mode
	Development bool `yaml:"development" json:"development"`
	// DisableCaller stops annotating logs with the calling function's file name and line number
	DisableCaller bool `yaml:"disable_caller" json:"disable_caller"`
	// DisableStacktrace disables automatic stacktrace capturing on error level and above
	DisableStacktrace bool `yaml:"disable_stacktrace" json:"disable_stacktrace"`
	// Sampling configures sampling of duplicate log statements
	Sampling *SamplingConfig `yaml:"sampling" json:"sampling"`
	// Environment specifies the running environment (development, staging, production)
	Environment string `yaml:"environment" json:"environment"`
	// ApplicationName specifies the name of the application
	ApplicationName string `yaml:"application_name" json:"application_name"`
}

// SamplingConfig sets a sampling strategy for the logger
type SamplingConfig struct {
	Initial    int `yaml:"initial" json:"initial"`
	Thereafter int `yaml:"thereafter" json:"thereafter"`
}

// Logger wraps the zap logger with additional functionality
type Logger struct {
	*zap.Logger
	config *Config
	fields map[string]interface{}
	mu     sync.RWMutex
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp      time.Time     `json:"timestamp"`
	RequestID      string        `json:"request_id"`
	ExecutionTime  time.Duration `json:"execution_time"`
	Environment    string        `json:"environment"`
	ApplicationName string       `json:"application_name"`
	Message        string        `json:"message"`
	Level          string        `json:"level"`
	Data           interface{}   `json:"data,omitempty"`
}

// Logger interface defines the logging methods
type LoggerInterface interface {
	// Log methods with mandatory context (can be nil if not needed)
	Info(ctx context.Context, message string, fields ...map[string]interface{})
	Error(ctx context.Context, message string, fields ...map[string]interface{})
	Debug(ctx context.Context, message string, fields ...map[string]interface{})
	Warn(ctx context.Context, message string, fields ...map[string]interface{})
}

// New creates a new logger instance with the given configuration
func New(cfg *Config) (*Logger, error) {
	if cfg == nil {
		cfg = &Config{
			Level:           "info",
			Format:          "json",
			OutputPaths:     []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			Environment:     EnvDevelopment,
			ApplicationName: "default-app",
		}
	}

	// Set default application name if not provided
	if cfg.ApplicationName == "" {
		cfg.ApplicationName = "default-app"
	}

	// Set environment
	if cfg.Environment == "" {
		cfg.Environment = EnvDevelopment
	}
	
	// Validate and normalize environment
	switch strings.ToLower(cfg.Environment) {
	case EnvDevelopment, "dev":
		cfg.Environment = EnvDevelopment
	case EnvStaging, "stage":
		cfg.Environment = EnvStaging
	case EnvProduction, "prod":
		cfg.Environment = EnvProduction
	default:
		cfg.Environment = EnvDevelopment
	}

	// Convert log level string to zapcore.Level
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		return nil, fmt.Errorf("invalid log level: %v", err)
	}

	// Create basic encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Configure zap logger
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       cfg.Development,
		DisableCaller:    cfg.DisableCaller,
		DisableStacktrace: cfg.DisableStacktrace,
		Sampling:         nil, // Configured below if needed
		Encoding:         cfg.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      cfg.OutputPaths,
		ErrorOutputPaths: cfg.ErrorOutputPaths,
	}

	// Configure sampling if specified
	if cfg.Sampling != nil {
		zapConfig.Sampling = &zap.SamplingConfig{
			Initial:    cfg.Sampling.Initial,
			Thereafter: cfg.Sampling.Thereafter,
		}
	}

	// Build the logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.Fields(
			zap.String("environment", cfg.Environment),
			zap.String("application_name", cfg.ApplicationName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %v", err)
	}

	return &Logger{
		Logger: logger,
		config: cfg,
		fields: make(map[string]interface{}),
	}, nil
}

// With creates a new Logger with the given fields added to its context
func (l *Logger) With(fields map[string]interface{}) *Logger {
	if len(fields) == 0 {
		return l
	}

	l.mu.RLock()
	newFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	l.mu.RUnlock()

	for k, v := range fields {
		newFields[k] = v
	}

	return &Logger{
		Logger: l.Logger.With(fieldsToZapFields(fields)...),
		config: l.config,
		fields: newFields,
	}
}

// Debug logs a message at debug level
func (l *Logger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) {
	zapFields := l.getZapFields(ctx, fields...)
	l.Logger.Debug(message, zapFields...)
}

// Info logs a message at info level
func (l *Logger) Info(ctx context.Context, message string, fields ...map[string]interface{}) {
	zapFields := l.getZapFields(ctx, fields...)
	l.Logger.Info(message, zapFields...)
}

// Warn logs a message at warn level
func (l *Logger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) {
	zapFields := l.getZapFields(ctx, fields...)
	l.Logger.Warn(message, zapFields...)
}

// Error logs a message at error level
func (l *Logger) Error(ctx context.Context, message string, fields ...map[string]interface{}) {
	zapFields := l.getZapFields(ctx, fields...)
	l.Logger.Error(message, zapFields...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1)
func (l *Logger) Fatal(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	l.Logger.Fatal(message, fieldsToZapFields(fields)...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// fieldsToZapFields converts a map of fields to a slice of zap.Field
func fieldsToZapFields(fields map[string]interface{}) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

// getZapFields combines context and fields into zap fields
func (l *Logger) getZapFields(ctx context.Context, fields ...map[string]interface{}) []zap.Field {
	mergedFields := make(map[string]interface{})

	// Add context information if context is not nil
	if ctx != nil {
		if requestID, err := libctx.GetRequestID(ctx); err == nil {
			mergedFields["request_id"] = requestID
		}
		if execTime, err := libctx.GetExecutionTime(ctx); err == nil {
			mergedFields["execution_time"] = execTime
		}
	}

	// Merge all fields
	for _, f := range fields {
		for k, v := range f {
			mergedFields[k] = v
		}
	}

	return fieldsToZapFields(mergedFields)
}

type defaultLogger struct {
	environment     string
	applicationName string
}

// NewLogger creates a new logger instance
func NewLogger() LoggerInterface {
	return &defaultLogger{
		environment:     EnvDevelopment,
		applicationName: "default-app",
	}
}

// GetEnvironment returns the current environment
func (l *defaultLogger) GetEnvironment() string {
	return l.environment
}

// GetApplicationName returns the current application name
func (l *defaultLogger) GetApplicationName() string {
	return l.applicationName
}

func (l *defaultLogger) Info(ctx context.Context, message string, fields ...map[string]interface{}) {
	l.log(ctx, "INFO", message, fields...)
}

func (l *defaultLogger) Error(ctx context.Context, message string, fields ...map[string]interface{}) {
	l.log(ctx, "ERROR", message, fields...)
}

func (l *defaultLogger) Debug(ctx context.Context, message string, fields ...map[string]interface{}) {
	l.log(ctx, "DEBUG", message, fields...)
}

func (l *defaultLogger) Warn(ctx context.Context, message string, fields ...map[string]interface{}) {
	l.log(ctx, "WARN", message, fields...)
}

func (l *defaultLogger) log(ctx context.Context, level string, message string, fields ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp:      time.Now(),
		Message:        message,
		Level:         level,
		Environment:   l.GetEnvironment(),
		ApplicationName: l.GetApplicationName(),
	}

	// Add context information if context is not nil
	if ctx != nil {
		if requestID, err := libctx.GetRequestID(ctx); err == nil {
			entry.RequestID = requestID
		}
		if execTime, err := libctx.GetExecutionTime(ctx); err == nil {
			entry.ExecutionTime = execTime
		}
	}

	// Merge all fields
	if len(fields) > 0 {
		mergedFields := make(map[string]interface{})
		for _, f := range fields {
			for k, v := range f {
				mergedFields[k] = v
			}
		}
		entry.Data = mergedFields
	}

	// Convert to JSON and log
	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error marshaling log entry: %v", err)
		return
	}

	log.Println(string(jsonEntry))
}

// ServiceLogConfig holds service-specific logging configuration
type ServiceLogConfig struct {
	Level       string
	Format      string
	Environment string
	ServiceName string
}

// SetupServiceLogger creates a new logger instance with standard configuration for services
func SetupServiceLogger(cfg ServiceLogConfig) (*Logger, error) {
	// Validate and normalize environment
	env := strings.ToLower(cfg.Environment)
	switch env {
	case EnvDevelopment, "dev":
		env = EnvDevelopment
	case EnvStaging, "stage":
		env = EnvStaging
	case EnvProduction, "prod":
		env = EnvProduction
	default:
		env = EnvDevelopment
	}

	// Define log directory based on environment
	var logDir string
	if env == EnvProduction {
		logDir = "/var/log/flowpilotx"
	} else {
		// Use local logs directory for non-prod environments
		logDir = filepath.Join(".", "logs", "flowpilotx")
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Fallback to current directory if can't create log directory
		logDir = "."
	}

	// Define log files
	mainLogFile := filepath.Join(logDir, fmt.Sprintf("%s.log", cfg.ServiceName))
	errorLogFile := filepath.Join(logDir, fmt.Sprintf("%s-error.log", cfg.ServiceName))

	// Initialize logger configuration
	logConfig := &Config{
		// Basic settings
		Level:           cfg.Level,
		Format:          cfg.Format,
		Environment:     env,
		ApplicationName: cfg.ServiceName,

		// Output configuration
		OutputPaths: []string{
			"stdout",
			mainLogFile,
		},
		ErrorOutputPaths: []string{
			"stderr",
			errorLogFile,
		},

		// Development mode settings
		Development:       env != EnvProduction,
		DisableCaller:    false, // Enable caller information
		DisableStacktrace: false, // Enable stack traces

		// Sampling configuration to prevent log flooding
		Sampling: &SamplingConfig{
			Initial:    100, // Log first 100 entries
			Thereafter: 100, // Then log every 100th entry
		},
	}

	// Create logger
	logger, err := New(logConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	// Log initialization
	logger.Info(nil, "Logger initialized", map[string]interface{}{
		"service":     cfg.ServiceName,
		"environment": env,
		"log_level":   cfg.Level,
		"log_format":  cfg.Format,
		"main_log":    mainLogFile,
		"error_log":   errorLogFile,
	})

	return logger, nil
}

// GetEnvironment returns the current environment from the logger config
func (l *Logger) GetEnvironment() string {
	return l.config.Environment
}

// GetApplicationName returns the current application name from the logger config
func (l *Logger) GetApplicationName() string {
	return l.config.ApplicationName
}


