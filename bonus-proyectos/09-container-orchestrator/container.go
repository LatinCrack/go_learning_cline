package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// ContainerState represents the current state of a container/process.
type ContainerState string

const (
	StateCreated    ContainerState = "created"
	StateRunning    ContainerState = "running"
	StateStopped    ContainerState = "stopped"
	StateFailed     ContainerState = "failed"
	StateRestarting ContainerState = "restarting"
)

// ContainerConfig holds the configuration for a single container.
type ContainerConfig struct {
	Name          string    `json:"name"`
	Command       string    `json:"command"`
	Args          []string  `json:"args"`
	WorkDir       string    `json:"work_dir,omitempty"`
	Env           []string  `json:"env,omitempty"`
	RestartPolicy string    `json:"restart_policy"` // "always", "on-failure", "never"
	HealthCheck   *HCConfig `json:"health_check,omitempty"`
	MaxRestarts   int       `json:"max_restarts,omitempty"`
}

// HCConfig holds health check configuration for a container.
type HCConfig struct {
	Command     string   `json:"command"`
	Args        []string `json:"args,omitempty"`
	IntervalSec int      `json:"interval_sec"`
	TimeoutSec  int      `json:"timeout_sec"`
	Retries     int      `json:"retries"`
}

// Container represents a managed process with its lifecycle state.
type Container struct {
	Config      ContainerConfig
	State       ContainerState
	PID         int
	ExitCode    int
	StartCount  int
	LastStarted time.Time
	LastStopped time.Time
	Healthy     bool
	Cmd         *exec.Cmd
	cancel      context.CancelFunc
	mu          sync.RWMutex
}

// NewContainer creates a new Container instance from a configuration.
func NewContainer(cfg ContainerConfig) *Container {
	return &Container{
		Config:  cfg,
		State:   StateCreated,
		Healthy: true,
	}
}

// GetState returns the current state in a thread-safe manner.
func (c *Container) GetState() ContainerState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.State
}

// SetState updates the container state in a thread-safe manner.
func (c *Container) SetState(s ContainerState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.State = s
}

// IsHealthy returns the health status in a thread-safe manner.
func (c *Container) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Healthy
}

// SetHealthy updates the health status in a thread-safe manner.
func (c *Container) SetHealthy(h bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Healthy = h
}

// Start launches the container's process using os/exec.
// It captures stdout/stderr via the centralized logger and manages process groups
// for proper signal propagation.
func (c *Container) Start(ctx context.Context, logger *Logger) error {
	c.mu.Lock()
	if c.State == StateRunning {
		c.mu.Unlock()
		return fmt.Errorf("container %s is already running", c.Config.Name)
	}

	ctx, c.cancel = context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, c.Config.Command, c.Config.Args...)
	cmd.Dir = c.Config.WorkDir
	if len(c.Config.Env) > 0 {
		cmd.Env = append(os.Environ(), c.Config.Env...)
	}

	// Create a new process group (platform-specific).
	SetProcessGroup(cmd)

	// Wire stdout/stderr to the centralized logger.
	cmd.Stdout = logger.Writer(c.Config.Name, "stdout")
	cmd.Stderr = logger.Writer(c.Config.Name, "stderr")

	c.Cmd = cmd
	c.mu.Unlock()

	if err := cmd.Start(); err != nil {
		c.SetState(StateFailed)
		return fmt.Errorf("failed to start container %s: %w", c.Config.Name, err)
	}

	c.mu.Lock()
	c.PID = cmd.Process.Pid
	c.StartCount++
	c.LastStarted = time.Now()
	c.State = StateRunning
	c.Healthy = true
	c.mu.Unlock()

	logger.Log(c.Config.Name, "info", fmt.Sprintf("started with PID %d", c.PID))

	// Monitor the process in a separate goroutine.
	go func() {
		err := cmd.Wait()
		c.mu.Lock()
		c.LastStopped = time.Now()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				c.ExitCode = exitErr.ExitCode()
			} else {
				c.ExitCode = -1
			}
			c.State = StateFailed
		} else {
			c.ExitCode = 0
			c.State = StateStopped
		}
		c.mu.Unlock()

		logger.Log(c.Config.Name, "info", fmt.Sprintf("exited with code %d", c.ExitCode))
	}()

	return nil
}

// Stop sends SIGTERM to the container's process group and waits for it to exit.
// If it does not exit within the timeout, SIGKILL is sent.
func (c *Container) Stop(logger *Logger, timeout time.Duration) error {
	c.mu.RLock()
	if c.State != StateRunning {
		c.mu.RUnlock()
		return fmt.Errorf("container %s is not running (state: %s)", c.Config.Name, c.State)
	}
	cmd := c.Cmd
	cancel := c.cancel
	pid := c.PID
	c.mu.RUnlock()

	logger.Log(c.Config.Name, "info", "sending SIGTERM to process group")

	// Cancel the context first (stops exec.CommandContext monitoring).
	if cancel != nil {
		cancel()
	}

	// Send SIGTERM to the entire process group (platform-specific).
	if cmd.Process != nil {
		if err := KillProcessGroup(pid, syscall.SIGTERM); err != nil {
			logger.Log(c.Config.Name, "warn", fmt.Sprintf("SIGTERM failed: %v", err))
		}
	}

	// Wait for the process to exit or force kill after timeout.
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Log(c.Config.Name, "info", "process exited gracefully")
	case <-time.After(timeout):
		logger.Log(c.Config.Name, "warn", "timeout reached, sending SIGKILL")
		if cmd.Process != nil {
			_ = KillProcessGroup(pid, syscall.SIGKILL)
		}
		<-done
	}

	c.SetState(StateStopped)
	return nil
}

// Restart stops (if running) and then starts the container again.
func (c *Container) Restart(ctx context.Context, logger *Logger, timeout time.Duration) error {
	c.SetState(StateRestarting)
	logger.Log(c.Config.Name, "info", "restarting container")

	if c.Cmd != nil {
		_ = c.Stop(logger, timeout)
	}

	return c.Start(ctx, logger)
}
