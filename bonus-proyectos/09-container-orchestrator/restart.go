package main

import (
	"context"
	"fmt"
	"time"
)

// RestartPolicy defines the type of restart behavior for a container.
type RestartPolicy string

const (
	PolicyAlways    RestartPolicy = "always"
	PolicyOnFailure RestartPolicy = "on-failure"
	PolicyNever     RestartPolicy = "never"
)

// RestartManager handles automatic restart decisions based on the container's
// configured restart policy. It tracks restart counts and enforces maximum limits.
type RestartManager struct {
	maxRestarts int
}

// NewRestartManager creates a RestartManager with the given maximum restart limit.
func NewRestartManager(maxRestarts int) *RestartManager {
	return &RestartManager{
		maxRestarts: maxRestarts,
	}
}

// ShouldRestart evaluates whether a container should be restarted based on its
// restart policy, exit code, and current restart count. Returns true if the
// container should be restarted, false otherwise.
func (rm *RestartManager) ShouldRestart(policy RestartPolicy, exitCode int, currentRestarts int) (bool, string) {
	switch policy {
	case PolicyAlways:
		if rm.maxRestarts > 0 && currentRestarts >= rm.maxRestarts {
			return false, fmt.Sprintf("max restart limit (%d) reached", rm.maxRestarts)
		}
		return true, "policy=always: restarting unconditionally"

	case PolicyOnFailure:
		if exitCode == 0 {
			return false, "policy=on-failure: exited cleanly, no restart needed"
		}
		if rm.maxRestarts > 0 && currentRestarts >= rm.maxRestarts {
			return false, fmt.Sprintf("max restart limit (%d) reached", rm.maxRestarts)
		}
		return true, fmt.Sprintf("policy=on-failure: exit code %d != 0, restarting", exitCode)

	case PolicyNever:
		return false, "policy=never: container will not be restarted"

	default:
		return false, fmt.Sprintf("unknown policy: %s", policy)
	}
}

// RestartSupervisor watches containers and applies restart policies when they exit.
// It runs as a goroutine, receiving container exit events and triggering restarts.
type RestartSupervisor struct {
	logger   *Logger
	managers map[string]*RestartManager // keyed by container name
	events   chan RestartEvent
}

// RestartEvent is sent when a container exits and needs policy evaluation.
type RestartEvent struct {
	Container *Container
	ExitCode  int
}

// NewRestartSupervisor creates a new supervisor that processes restart events.
func NewRestartSupervisor(logger *Logger) *RestartSupervisor {
	return &RestartSupervisor{
		logger:   logger,
		managers: make(map[string]*RestartManager),
		events:   make(chan RestartEvent, 64),
	}
}

// Register adds a container to be monitored by the restart supervisor.
func (rs *RestartSupervisor) Register(c *Container) {
	policy := parsePolicy(c.Config.RestartPolicy)
	rm := NewRestartManager(c.Config.MaxRestarts)
	rs.managers[c.Config.Name] = rm
	rs.logger.Log(c.Config.Name, "info", fmt.Sprintf("registered with policy=%s, max_restarts=%d", policy, c.Config.MaxRestarts))
}

// Events returns the channel for sending restart events.
func (rs *RestartSupervisor) Events() chan<- RestartEvent {
	return rs.events
}

// Run starts the restart supervisor loop. It listens for container exit events
// and applies the appropriate restart policy. This should be called as a goroutine.
func (rs *RestartSupervisor) Run(ctx context.Context, orchestrator *Orchestrator) {
	for {
		select {
		case <-ctx.Done():
			rs.logger.Log("restart-supervisor", "info", "shutting down")
			return
		case event := <-rs.events:
			c := event.Container
			name := c.Config.Name
			policy := parsePolicy(c.Config.RestartPolicy)

			rm, ok := rs.managers[name]
			if !ok {
				rs.logger.Log(name, "warn", "no restart manager found, skipping")
				continue
			}

			shouldRestart, reason := rm.ShouldRestart(policy, event.ExitCode, c.StartCount)
			rs.logger.Log(name, "info", fmt.Sprintf("restart decision: %s", reason))

			if shouldRestart {
				// Exponential backoff: 1s, 2s, 4s, 8s... capped at 30s.
				backoff := time.Duration(1<<uint(c.StartCount-1)) * time.Second
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}

				rs.logger.Log(name, "info", fmt.Sprintf("restarting in %v (attempt %d)", backoff, c.StartCount+1))
				time.Sleep(backoff)

				if err := orchestrator.StartContainer(name); err != nil {
					rs.logger.Log(name, "error", fmt.Sprintf("restart failed: %v", err))
				}
			}
		}
	}
}

// parsePolicy converts a string restart policy to the RestartPolicy type.
func parsePolicy(s string) RestartPolicy {
	switch s {
	case "always":
		return PolicyAlways
	case "on-failure":
		return PolicyOnFailure
	case "never":
		return PolicyNever
	default:
		return PolicyNever
	}
}
