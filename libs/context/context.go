package context

import (
	"context"
	"fmt"
	"time"
)

type contextKey string

const (
	// RequestIDKey is the context key for request ID
	requestIDKey contextKey = "request_id"
	// ExecutionTimeKey is the context key for execution time
	executionTimeKey contextKey = "execution_time"
	// ExecutionIDKey is the context key for execution ID
	executionIDKey contextKey = "execution_id"
)

// WithRequestID returns a new context with request ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("nil context")
	}
	
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return "", fmt.Errorf("request ID not found in context")
	}
	return requestID, nil
}

// WithExecutionID returns a new context with execution ID
func WithExecutionID(ctx context.Context, executionID string) context.Context {
	return context.WithValue(ctx, executionIDKey, executionID)
}

// GetExecutionID retrieves execution ID from context
func GetExecutionID(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("nil context")
	}
	
	executionID, ok := ctx.Value(executionIDKey).(string)
	if !ok {
		return "", fmt.Errorf("execution ID not found in context")
	}
	return executionID, nil
}

// WithExecutionTime returns a new context with execution time
func WithExecutionTime(ctx context.Context, executionTime time.Duration) context.Context {
	return context.WithValue(ctx, executionTimeKey, executionTime)
}

// GetExecutionTime retrieves execution time from context
func GetExecutionTime(ctx context.Context) (time.Duration, error) {
	if ctx == nil {
		return 0, fmt.Errorf("nil context")
	}
	
	executionTime, ok := ctx.Value(executionTimeKey).(time.Duration)
	if !ok {
		return 0, fmt.Errorf("execution time not found in context")
	}
	return executionTime, nil
}

// WithRequestAndTime returns a new context with both request ID and execution time
func WithRequestAndTime(ctx context.Context, requestID string, executionTime time.Duration) context.Context {
	ctx = WithRequestID(ctx, requestID)
	return WithExecutionTime(ctx, executionTime)
}

// WithRequestAndExecution returns a new context with request ID and execution ID
func WithRequestAndExecution(ctx context.Context, requestID string, executionID string) context.Context {
	ctx = WithRequestID(ctx, requestID)
	return WithExecutionID(ctx, executionID)
}

// WithAll returns a new context with request ID, execution ID, and execution time
func WithAll(ctx context.Context, requestID string, executionID string, executionTime time.Duration) context.Context {
	ctx = WithRequestID(ctx, requestID)
	ctx = WithExecutionID(ctx, executionID)
	return WithExecutionTime(ctx, executionTime)
}

// GetRequestAndTime retrieves both request ID and execution time from context
func GetRequestAndTime(ctx context.Context) (requestID string, executionTime time.Duration, err error) {
	requestID, err = GetRequestID(ctx)
	if err != nil {
		return "", 0, err
	}

	executionTime, err = GetExecutionTime(ctx)
	if err != nil {
		return requestID, 0, err
	}

	return requestID, executionTime, nil
}

// GetRequestAndExecution retrieves both request ID and execution ID from context
func GetRequestAndExecution(ctx context.Context) (requestID string, executionID string, err error) {
	requestID, err = GetRequestID(ctx)
	if err != nil {
		return "", "", err
	}

	executionID, err = GetExecutionID(ctx)
	if err != nil {
		return requestID, "", err
	}

	return requestID, executionID, nil
}

// GetAll retrieves request ID, execution ID, and execution time from context
func GetAll(ctx context.Context) (requestID string, executionID string, executionTime time.Duration, err error) {
	requestID, err = GetRequestID(ctx)
	if err != nil {
		return "", "", 0, err
	}

	executionID, err = GetExecutionID(ctx)
	if err != nil {
		return requestID, "", 0, err
	}

	executionTime, err = GetExecutionTime(ctx)
	if err != nil {
		return requestID, executionID, 0, err
	}

	return requestID, executionID, executionTime, nil
} 