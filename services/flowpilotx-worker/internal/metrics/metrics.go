package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/flowpilotx/libs/config"
	"github.com/flowpilotx/libs/logger"
)

// Metrics handles service metrics collection and reporting
type Metrics struct {
	logger          logger.LoggerInterface
	inFlightRequests atomic.Int64
	requestCounter   *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	server          *http.Server
}

// New creates a new metrics instance
func New(log logger.LoggerInterface) *Metrics {
	m := &Metrics{
		logger: log,
		requestCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "worker_requests_total",
				Help: "Total number of requests by type and status",
			},
			[]string{"type", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "worker_request_duration_seconds",
				Help:    "Request duration in seconds by type",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"type"},
		),
	}

	// Register metrics with Prometheus
	prometheus.MustRegister(m.requestCounter)
	prometheus.MustRegister(m.requestDuration)

	return m
}

// StartServer starts the metrics server
func (m *Metrics) StartServer(cfg *config.MetricsConfig) error {
	if !cfg.Enabled {
		m.logger.Info(nil, "Metrics server disabled")
		return nil
	}

	m.logger.Info(nil, "Starting metrics server", map[string]interface{}{
		"port": cfg.Port,
	})

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	m.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			m.logger.Error(nil, "Metrics server failed", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	return nil
}

// StopServer gracefully stops the metrics server
func (m *Metrics) StopServer(ctx context.Context) error {
	if m.server == nil {
		return nil
	}

	return m.server.Shutdown(ctx)
}

// IncInFlightRequests increments the number of in-flight requests
func (m *Metrics) IncInFlightRequests() {
	m.inFlightRequests.Add(1)
}

// DecInFlightRequests decrements the number of in-flight requests
func (m *Metrics) DecInFlightRequests() {
	m.inFlightRequests.Add(-1)
}

// IncRequests increments the request counter for a given type and status
func (m *Metrics) IncRequests(requestType, status string) {
	m.requestCounter.WithLabelValues(requestType, status).Inc()
}

// ObserveRequestDuration records the duration of a request
func (m *Metrics) ObserveRequestDuration(requestType string, duration float64) {
	m.requestDuration.WithLabelValues(requestType).Observe(duration)
} 