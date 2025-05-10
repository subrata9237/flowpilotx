package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	libctx "github.com/flowpilotx/libs/context"
)

// Constants
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	// Debug level for detailed information
	Debug LogLevel = iota
	// Info level for general operational information
	Info
	// Warn level for warning messages
	Warn
	// Error level for error messages
	Error
	// Fatal level for fatal errors that require application termination
	Fatal
)

// StringToLogLevel converts a string log level to LogLevel
func StringToLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return Debug
	case "info":
		return Info
	case "warn", "warning":
		return Warn
	case "error":
		return Error
	case "fatal":
		return Fatal
	default:
		return Info // Default to Info level if unrecognized
	}
}

// LoggerInterface defines the interface for logging operations
type LoggerInterface interface {
	Debug(ctx context.Context, msg string, fields map[string]interface{})
	Info(ctx context.Context, msg string, fields map[string]interface{})
	Warn(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, msg string, fields map[string]interface{})
	Fatal(ctx context.Context, msg string, fields map[string]interface{})
	GetEnvironment() string
	GetApplicationName() string
}


// ServiceLogConfig holds service-specific logging configuration
type ServiceLogConfig struct {
	Level       string
	Format      string
	Environment string
	ServiceName string
}

// Logger wraps the zap logger with additional functionality
type Logger struct {
	*zap.Logger
	environment     string
	applicationName string
	formatter      LogFormatter
	level          LogLevel
	mu             sync.RWMutex
}

// LogFormatter defines how log messages are formatted
type LogFormatter interface {
	Format(level string, msg string, fields map[string]interface{}) string
}

// JSONFormatter formats logs as JSON
type JSONFormatter struct{}

// TextFormatter formats logs as human-readable text
type TextFormatter struct{}

// Constructor Functions

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config ServiceLogConfig) LoggerInterface {
	var formatter LogFormatter
	if config.Format == "json" {
		formatter = &JSONFormatter{}
	} else {
		formatter = &TextFormatter{}
	}

	return &Logger{
		Logger:          zap.NewExample(),
		level:          StringToLogLevel(config.Level),
		environment:    config.Environment,
		applicationName: config.ServiceName,
		formatter:      formatter,
	}
}

// GetEnvironment returns the current environment
func (l *Logger) GetEnvironment() string {
	return l.environment
}

// GetApplicationName returns the current application name
func (l *Logger) GetApplicationName() string {
	return l.applicationName
}

// Debug logs a debug message
func (l *Logger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= Debug {
		l.log(ctx, "DEBUG", msg, fields)
	}
}

// Info logs an info message
func (l *Logger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= Info {
		l.log(ctx, "INFO", msg, fields)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= Warn {
		l.log(ctx, "WARN", msg, fields)
	}
}

// Error logs an error message
func (l *Logger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= Error {
		l.log(ctx, "ERROR", msg, fields)
	}
}

// Fatal logs a fatal message and exits the application
func (l *Logger) Fatal(ctx context.Context, msg string, fields map[string]interface{}) {
	if l.level <= Fatal {
		l.log(ctx, "FATAL", msg, fields)
		os.Exit(1)
	}
}

// getContextFields extracts standard fields from context
func getContextFields(ctx context.Context) map[string]interface{} {
	fields := make(map[string]interface{})
	
	if ctx == nil {
		return fields
	}

	// Extract request ID using the context package's function
	if requestID, err := libctx.GetRequestID(ctx); err == nil {
		fields["request_id"] = requestID
	}

	// Extract execution time using the context package's function
	if execTime, err := libctx.GetExecutionTime(ctx); err == nil {
		fields["execution_time"] = execTime
	}

	return fields
}

// log handles the actual logging
func (l *Logger) log(ctx context.Context, level string, msg string, fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Combine logger fields with message fields
	allFields := make(map[string]interface{})
	
	// Add context fields first (highest priority)
	contextFields := getContextFields(ctx)
	for k, v := range contextFields {
		allFields[k] = v
	}
	
	// Add user-provided fields
	for k, v := range fields {
		// Don't override context fields
		if _, exists := contextFields[k]; !exists {
			allFields[k] = v
		}
	}

	// Add standard fields
	allFields["environment"] = l.environment
	allFields["app_name"] = l.applicationName
	allFields["level"] = level
	allFields["timestamp"] = time.Now().Format(time.RFC3339)
	allFields["message"] = msg

	// Format and write log entry
	logLine := l.formatter.Format(level, msg, allFields)
	fmt.Fprintln(os.Stdout, logLine)
}

// Format formats a log message as JSON
func (f *JSONFormatter) Format(level string, msg string, fields map[string]interface{}) string {
	// Ensure request_id and execution_time are prominently placed in JSON
	logEntry := map[string]interface{}{
		"timestamp": fields["timestamp"],
		"level":     level,
		"message":   msg,
	}
	
	// Add request_id and execution_time first if they exist
	if requestID, ok := fields["request_id"]; ok {
		logEntry["request_id"] = requestID
	}
	if execTime, ok := fields["execution_time"]; ok {
		logEntry["execution_time"] = execTime
	}
	
	// Add remaining fields
	for k, v := range fields {
		if k != "timestamp" && k != "level" && k != "message" && 
		   k != "request_id" && k != "execution_time" {
			logEntry[k] = v
		}
	}

	jsonBytes, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Sprintf("[ERROR] Failed to marshal log entry: %v", err)
	}
	return string(jsonBytes)
}

// Format formats a log message as text
func (f *TextFormatter) Format(level string, msg string, fields map[string]interface{}) string {
	var parts []string
	
	// Start with timestamp and level
	timestamp := fields["timestamp"].(string)
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, fmt.Sprintf("[%s]", level))
	
	// Add request_id and execution_time if they exist
	if requestID, ok := fields["request_id"]; ok {
		parts = append(parts, fmt.Sprintf("request_id=%v", requestID))
	}
	if execTime, ok := fields["execution_time"]; ok {
		parts = append(parts, fmt.Sprintf("execution_time=%v", execTime))
	}
	
	// Add message
	parts = append(parts, fmt.Sprintf("msg=%q", msg))
	
	// Add remaining fields
	for k, v := range fields {
		if k != "timestamp" && k != "level" && k != "message" && 
		   k != "request_id" && k != "execution_time" {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	
	return strings.Join(parts, " ")
}


