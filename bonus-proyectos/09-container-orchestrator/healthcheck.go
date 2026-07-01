package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// HealthStatus represents the result of a health check probe.
type HealthStatus struct {
	Container string    `json:"container"`
	Healthy   bool      `json:"healthy"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthChecker runs periodic health check probes on containers that have
// health check configurations. Each container's checks run in their own
// goroutine and report status via channels.
type HealthChecker struct {
	logger   *Logger
	statusCh chan HealthStatus
}

// NewHealthChecker creates a HealthChecker with a buffered status channel.
func NewHealthChecker(logger *Logger) *HealthChecker {
	return &HealthChecker{
		logger:   logger,
		statusCh: make(chan HealthStatus, 128),
	}
}

// StatusChannel returns the read-only channel for receiving health status updates.
func (hc *HealthChecker) StatusChannel() <-chan HealthStatus {
	return hc.statusCh
}

// StartCheck launches a health check goroutine for the given container.
// It periodically runs the configured health check command and reports
// the result. The goroutine stops when the context is cancelled.
func (hc *HealthChecker) StartCheck(ctx context.Context, c *Container) {
	if c.Config.HealthCheck == nil {
		hc.logger.Log(c.Config.Name, "info", "no health check configured, skipping")
		return
	}

	cfg := c.Config.HealthCheck
	interval := time.Duration(cfg.IntervalSec) * time.Second
	timeout := time.Duration(cfg.TimeoutSec) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	hc.logger.Log(c.Config.Name, "info", fmt.Sprintf("health check started: interval=%v, timeout=%v, retries=%d", interval, timeout, cfg.Retries))

	go func() {
		consecutiveFailures := 0

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				hc.logger.Log(c.Config.Name, "info", "health check stopped")
				return
			case <-ticker.C:
				// Only run health checks if the container is running.
				if c.GetState() != StateRunning {
					continue
				}

				status := hc.probe(c, timeout)

				if status.Healthy {
					consecutiveFailures = 0
					c.SetHealthy(true)
				} else {
					consecutiveFailures++
					hc.logger.Log(c.Config.Name, "warn", fmt.Sprintf("health check failed (%d/%d): %s",
						consecutiveFailures, cfg.Retries, status.Message))

					if consecutiveFailures >= cfg.Retries {
						c.SetHealthy(false)
						hc.logger.Log(c.Config.Name, "error", fmt.Sprintf("marked unhealthy after %d consecutive failures", consecutiveFailures))
					}
				}

				// Send status to channel for API consumption.
				select {
				case hc.statusCh <- status:
				default:
					// Drop if channel is full to avoid blocking.
				}
			}
		}
	}()
}

// probe executes the health check command for a container with a timeout.
func (hc *HealthChecker) probe(c *Container, timeout time.Duration) HealthStatus {
	cfg := c.Config.HealthCheck
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	output, err := cmd.CombinedOutput()

	status := HealthStatus{
		Container: c.Config.Name,
		Timestamp: time.Now(),
	}

	if err != nil {
		status.Healthy = false
		status.Message = fmt.Sprintf("command failed: %v (output: %s)", err, string(output))
	} else {
		status.Healthy = true
		status.Message = "healthy"
	}

	return status
}

// GetLatestStatus returns the most recent health status for all containers.
func (hc *HealthChecker) GetLatestStatus(containers map[string]*Container) []HealthStatus {
	var statuses []HealthStatus
	for _, c := range containers {
		status := HealthStatus{
			Container: c.Config.Name,
			Healthy:   c.IsHealthy(),
			Timestamp: time.Now(),
		}
		if c.IsHealthy() {
			status.Message = "healthy"
		} else {
			status.Message = "unhealthy"
		}
		statuses = append(statuses, status)
	}
	return statuses
}
