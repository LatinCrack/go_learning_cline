package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Orchestrator is the central engine that manages all containers, their lifecycle,
// health checks, restart policies, and provides status information.
type Orchestrator struct {
	containers map[string]*Container
	logger     *Logger
	health     *HealthChecker
	restart    *RestartSupervisor
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	mu         sync.RWMutex
}

// OrchestratorConfig represents the top-level configuration file structure.
type OrchestratorConfig struct {
	Containers []ContainerConfig `json:"containers"`
	LogFile    string            `json:"log_file,omitempty"`
	APIPort    int               `json:"api_port,omitempty"`
}

// NewOrchestrator creates a new Orchestrator with all subsystems initialized.
func NewOrchestrator(logger *Logger) *Orchestrator {
	ctx, cancel := context.WithCancel(context.Background())

	health := NewHealthChecker(logger)
	restart := NewRestartSupervisor(logger)

	o := &Orchestrator{
		containers: make(map[string]*Container),
		logger:     logger,
		health:     health,
		restart:    restart,
		ctx:        ctx,
		cancel:     cancel,
	}

	return o
}

// LoadConfig reads and parses the JSON configuration file, creating containers.
func (o *Orchestrator) LoadConfig(path string) (*OrchestratorConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg OrchestratorConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults.
	if cfg.LogFile == "" {
		cfg.LogFile = "orchestrator.log"
	}
	if cfg.APIPort == 0 {
		cfg.APIPort = 8090
	}

	for _, cc := range cfg.Containers {
		if cc.Command == "" {
			return nil, fmt.Errorf("container %s has no command", cc.Name)
		}
		if cc.RestartPolicy == "" {
			cc.RestartPolicy = "never"
		}
		if cc.MaxRestarts == 0 {
			cc.MaxRestarts = 10
		}
		c := NewContainer(cc)
		o.containers[cc.Name] = c
		o.restart.Register(c)
		o.logger.Log(cc.Name, "info", fmt.Sprintf("loaded from config: cmd=%s, policy=%s", cc.Command, cc.RestartPolicy))
	}

	return &cfg, nil
}

// StartContainer starts a container by name. It is idempotent for already-running containers.
func (o *Orchestrator) StartContainer(name string) error {
	o.mu.RLock()
	c, ok := o.containers[name]
	o.mu.RUnlock()

	if !ok {
		return fmt.Errorf("container %s not found", name)
	}

	if err := c.Start(o.ctx, o.logger); err != nil {
		return err
	}

	// Launch health check if configured.
	o.health.StartCheck(o.ctx, c)

	// Monitor for process exit to feed the restart supervisor.
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		o.monitorContainer(c)
	}()

	return nil
}

// StopContainer gracefully stops a container by name.
func (o *Orchestrator) StopContainer(name string) error {
	o.mu.RLock()
	c, ok := o.containers[name]
	o.mu.RUnlock()

	if !ok {
		return fmt.Errorf("container %s not found", name)
	}

	return c.Stop(o.logger, 10*time.Second)
}

// GetContainer returns a container by name (read-only reference).
func (o *Orchestrator) GetContainer(name string) (*Container, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	c, ok := o.containers[name]
	return c, ok
}

// GetAllContainers returns a snapshot of all managed containers.
func (o *Orchestrator) GetAllContainers() map[string]*Container {
	o.mu.RLock()
	defer o.mu.RUnlock()
	result := make(map[string]*Container, len(o.containers))
	for k, v := range o.containers {
		result[k] = v
	}
	return result
}

// monitorContainer watches a container until it exits, then feeds the restart supervisor.
func (o *Orchestrator) monitorContainer(c *Container) {
	// Poll until the container is no longer running.
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			state := c.GetState()
			if state == StateFailed || state == StateStopped {
				o.logger.Log(c.Config.Name, "info", fmt.Sprintf("container exited (state=%s, code=%d), notifying restart supervisor", state, c.ExitCode))
				// Send restart event.
				o.restart.Events() <- RestartEvent{
					Container: c,
					ExitCode:  c.ExitCode,
				}
				return
			}
		}
	}
}

// StartAll starts all registered containers and their subsystems.
func (o *Orchestrator) StartAll() {
	o.logger.Log("orchestrator", "info", "starting all containers...")

	// Start the restart supervisor.
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		o.restart.Run(o.ctx, o)
	}()

	// Drain the health status channel in background.
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		for {
			select {
			case <-o.ctx.Done():
				return
			case <-o.health.StatusChannel():
				// Status updates are consumed; API reads from GetLatestStatus.
			}
		}
	}()

	// Start each container.
	for name := range o.containers {
		if err := o.StartContainer(name); err != nil {
			o.logger.Log(name, "error", fmt.Sprintf("failed to start: %v", err))
		}
	}
}

// Shutdown gracefully stops all containers and waits for subsystems to finish.
func (o *Orchestrator) Shutdown(timeout time.Duration) {
	o.logger.Log("orchestrator", "info", "initiating graceful shutdown...")

	// Stop all running containers.
	var wg sync.WaitGroup
	for name, c := range o.containers {
		if c.GetState() == StateRunning {
			wg.Add(1)
			go func(n string, cont *Container) {
				defer wg.Done()
				if err := cont.Stop(o.logger, timeout); err != nil {
					o.logger.Log(n, "error", fmt.Sprintf("shutdown stop failed: %v", err))
				}
			}(name, c)
		}
	}
	wg.Wait()

	// Cancel context to stop all goroutines.
	o.cancel()
	o.wg.Wait()

	o.logger.Log("orchestrator", "info", "all containers stopped, shutdown complete")
}

// WaitForSignal blocks until an OS signal (SIGTERM, SIGINT) is received,
// then initiates graceful shutdown.
func (o *Orchestrator) WaitForSignal(timeout time.Duration) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	sig := <-sigCh
	o.logger.Log("orchestrator", "info", fmt.Sprintf("received signal: %v", sig))
	o.Shutdown(timeout)
}