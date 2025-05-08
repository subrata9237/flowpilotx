package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
	"log"
)

// GRPCLogging returns a gRPC unary interceptor for logging
func GRPCLogging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		// Log request
		log.Printf("gRPC Request - Method: %s", info.FullMethod)
		
		// Execute handler
		resp, err := handler(ctx, req)
		
		// Log response
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf("gRPC Response - Method: %s, Duration: %v, Status: %s, Error: %v",
				info.FullMethod, duration, st.Code(), st.Message())
		} else {
			log.Printf("gRPC Response - Method: %s, Duration: %v, Status: %s",
				info.FullMethod, duration, codes.OK)
		}
		
		return resp, err
	}
}

// GRPCStreamLogging returns a gRPC stream interceptor for logging
func GRPCStreamLogging() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		
		// Log stream start
		log.Printf("gRPC Stream Started - Method: %s", info.FullMethod)
		
		// Execute handler
		err := handler(srv, ss)
		
		// Log stream end
		duration := time.Since(start)
		if err != nil {
			st, _ := status.FromError(err)
			log.Printf("gRPC Stream Ended - Method: %s, Duration: %v, Status: %s, Error: %v",
				info.FullMethod, duration, st.Code(), st.Message())
		} else {
			log.Printf("gRPC Stream Ended - Method: %s, Duration: %v, Status: %s",
				info.FullMethod, duration, codes.OK)
		}
		
		return err
	}
}

// HTTPLogging returns an HTTP middleware for logging REST requests
func HTTPLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Log request
		log.Printf("HTTP Request - Method: %s, Path: %s", r.Method, r.URL.Path)
		
		// Create custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Execute handler
		h.ServeHTTP(rw, r)
		
		// Log response
		duration := time.Since(start)
		log.Printf("HTTP Response - Method: %s, Path: %s, Status: %d, Duration: %v",
			r.Method, r.URL.Path, rw.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
} 