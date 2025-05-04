package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flowpilotx/queue-sidecar/internal/config"
	"github.com/flowpilotx/queue-sidecar/internal/queue"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	queueClient *queue.Client
	httpClient  *http.Client
	config      config.HTTPConfig
	logger      *logrus.Logger
}

func NewWorker(
	queueClient *queue.Client,
	httpClient *http.Client,
	cfg config.HTTPConfig,
	logger *logrus.Logger,
) *Worker {
	return &Worker{
		queueClient: queueClient,
		httpClient:  httpClient,
		config:      cfg,
		logger:      logger,
	}
}

func (w *Worker) Start(ctx context.Context) {
	msgs, err := w.queueClient.Consume()
	if err != nil {
		w.logger.WithError(err).Fatal("Failed to start consuming messages")
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping worker")
			return
		case msg, ok := <-msgs:
			if !ok {
				w.logger.Error("Message channel closed")
				return
			}

			w.processMessage(msg)
		}
	}
}

func (w *Worker) processMessage(msg queue.Delivery) {
	// Create a copy of the message body for retries
	body := make([]byte, len(msg.Body))
	copy(body, msg.Body)

	// Try to forward the message with retries
	for attempt := 0; attempt < w.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(w.config.RetryDelaySeconds) * time.Second)
		}

		err := w.forwardMessage(body)
		if err == nil {
			// Message successfully forwarded, acknowledge it
			if err := msg.Ack(false); err != nil {
				w.logger.WithError(err).Error("Failed to acknowledge message")
			}
			return
		}

		w.logger.WithError(err).WithField("attempt", attempt+1).Error("Failed to forward message")
	}

	// All retries failed, reject the message
	if err := msg.Nack(false, true); err != nil {
		w.logger.WithError(err).Error("Failed to reject message")
	}
}

func (w *Worker) forwardMessage(body []byte) error {
	req, err := http.NewRequest("POST", w.config.TargetURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
