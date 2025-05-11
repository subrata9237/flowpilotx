package logger

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	libctx "github.com/flowpilotx/libs/context"
)

// Add this at the top of the file with other imports

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
	level           LogLevel
	mu              sync.RWMutex
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(config ServiceLogConfig) LoggerInterface {
	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core with appropriate encoder
	var core zapcore.Core
	if config.Format == "json" {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			getZapLevel(config.Level),
		)
	} else {
		// Text format with custom console encoder
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			getZapLevel(config.Level),
		)
	}

	// Create logger with the core
	logger := zap.New(core, zap.AddCaller())

	return &Logger{
		Logger:          logger,
		level:           StringToLogLevel(config.Level),
		environment:     config.Environment,
		applicationName: config.ServiceName,
	}
}

// getZapLevel converts our LogLevel to zapcore.Level
func getZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
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

	// Convert fields to zap fields
	zapFields := make([]zap.Field, 0, len(allFields))
	for k, v := range allFields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	//Add color to the message based on level
	color := getColorForLevel(level)
	coloredMsg := fmt.Sprintf("%s%s%s", color, msg, colorReset)
	// Log using appropriate Zap level
	switch level {
	case "DEBUG":
		l.Logger.Debug(coloredMsg, zapFields...)
	case "INFO":
		l.Logger.Info(coloredMsg, zapFields...)
	case "WARN":
		l.Logger.Warn(coloredMsg, zapFields...)
	case "ERROR":
		l.Logger.Error(coloredMsg, zapFields...)
	case "FATAL":
		l.Logger.Fatal(coloredMsg, zapFields...)
	default:
		l.Logger.Info(coloredMsg, zapFields...)
	}
}

// Add these color constants at the top with other constants
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
)

// Add this function to color messages based on level
func getColorForLevel(level string) string {
	switch level {
	case "DEBUG":
		return colorBlue
	case "INFO":
		return colorGreen
	case "WARN":
		return colorYellow
	case "ERROR":
		return colorRed
	case "FATAL":
		return colorPurple
	default:
		return colorReset
	}
}
