package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Logger Tests ---

func TestNewLogger(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	// Verify the file was created.
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Fatal("log file was not created")
	}
}

func TestLoggerLog(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	logger.Log("test-container", "info", "hello world")

	entries := logger.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Container != "test-container" {
		t.Errorf("expected container 'test-container', got '%s'", entries[0].Container)
	}
	if entries[0].Level != "info" {
		t.Errorf("expected level 'info', got '%s'", entries[0].Level)
	}
	if entries[0].Message != "hello world" {
		t.Errorf("expected message 'hello world', got '%s'", entries[0].Message)
	}
}

func TestLoggerWriter(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")

	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	w := logger.Writer("my-container", "stdout")
	data := []byte("output line 1\noutput line 2\n")
	n, err := w.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected to write %d bytes, wrote %d", len(data), n)
	}

	entries := logger.GetEntries()
	if len(entries) < 2 {
		t.Fatalf("expected at least 2 entries, got %d", len(entries))
	}
}

// --- Restart Policy Tests ---

func TestRestartManager_AlwaysPolicy(t *testing.T) {
	rm := NewRestartManager(5)

	// Always should restart regardless of exit code.
	should, _ := rm.ShouldRestart(PolicyAlways, 0, 0)
	if !should {
		t.Error("always policy should restart on exit code 0")
	}

	should, _ = rm.ShouldRestart(PolicyAlways, 1, 0)
	if !should {
		t.Error("always policy should restart on exit code 1")
	}

	// Should not restart after max is reached.
	should, reason := rm.ShouldRestart(PolicyAlways, 1, 5)
	if should {
		t.Errorf("should not restart after max reached, got: %s", reason)
	}
}

func TestRestartManager_OnFailurePolicy(t *testing.T) {
	rm := NewRestartManager(5)

	// Should NOT restart on exit code 0.
	should, _ := rm.ShouldRestart(PolicyOnFailure, 0, 0)
	if should {
		t.Error("on-failure policy should not restart on exit code 0")
	}

	// Should restart on non-zero exit code.
	should, _ = rm.ShouldRestart(PolicyOnFailure, 1, 0)
	if !should {
		t.Error("on-failure policy should restart on exit code 1")
	}

	// Should not restart after max is reached.
	should, _ = rm.ShouldRestart(PolicyOnFailure, 1, 5)
	if should {
		t.Error("should not restart after max reached")
	}
}

func TestRestartManager_NeverPolicy(t *testing.T) {
	rm := NewRestartManager(0)

	should, _ := rm.ShouldRestart(PolicyNever, 1, 0)
	if should {
		t.Error("never policy should never restart")
	}
}

func TestRestartManager_NoLimit(t *testing.T) {
	rm := NewRestartManager(0) // 0 means no limit.

	should, _ := rm.ShouldRestart(PolicyAlways, 1, 1000)
	if !should {
		t.Error("unlimited restarts should allow restart at count 1000")
	}
}

// --- Container Tests ---

func TestNewContainer(t *testing.T) {
	cfg := ContainerConfig{
		Name:          "test-svc",
		Command:       "echo",
		Args:          []string{"hello"},
		RestartPolicy: "never",
	}

	c := NewContainer(cfg)
	if c.Config.Name != "test-svc" {
		t.Errorf("expected name 'test-svc', got '%s'", c.Config.Name)
	}
	if c.GetState() != StateCreated {
		t.Errorf("expected state 'created', got '%s'", c.GetState())
	}
	if !c.IsHealthy() {
		t.Error("new container should be healthy by default")
	}
}

func TestContainerStateThreadSafety(t *testing.T) {
	c := NewContainer(ContainerConfig{Name: "ts-test", Command: "echo"})

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			c.SetState(StateRunning)
		}
		close(done)
	}()

	for i := 0; i < 1000; i++ {
		_ = c.GetState()
	}
	<-done
}

func TestContainerStart(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	cfg := ContainerConfig{
		Name:          "echo-test",
		Command:       "echo",
		Args:          []string{"hello from test"},
		RestartPolicy: "never",
	}
	c := NewContainer(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := c.Start(ctx, logger); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if c.GetState() != StateRunning {
		t.Errorf("expected state 'running', got '%s'", c.GetState())
	}

	if c.PID == 0 {
		t.Error("PID should not be 0 after start")
	}

	// Wait for the process to finish (echo exits immediately).
	time.Sleep(1 * time.Second)

	state := c.GetState()
	if state != StateStopped && state != StateFailed {
		t.Errorf("expected stopped or failed after echo, got '%s'", state)
	}

	if c.ExitCode != 0 {
		t.Errorf("echo should exit with 0, got %d", c.ExitCode)
	}
}

func TestContainerStopNotRunning(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	c := NewContainer(ContainerConfig{Name: "idle", Command: "echo"})
	err = c.Stop(logger, 1*time.Second)
	if err == nil {
		t.Error("stopping a non-running container should return an error")
	}
}

func TestContainerHealthThreadSafety(t *testing.T) {
	c := NewContainer(ContainerConfig{Name: "health-ts", Command: "echo"})

	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			c.SetHealthy(i%2 == 0)
		}
		close(done)
	}()

	for i := 0; i < 1000; i++ {
		_ = c.IsHealthy()
	}
	<-done
}

// --- Health Checker Tests ---

func TestHealthChecker_GetLatestStatus(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	hc := NewHealthChecker(logger)

	containers := map[string]*Container{
		"svc-1": {Config: ContainerConfig{Name: "svc-1"}, Healthy: true},
		"svc-2": {Config: ContainerConfig{Name: "svc-2"}, Healthy: false},
	}

	statuses := hc.GetLatestStatus(containers)
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}

	for _, s := range statuses {
		switch s.Container {
		case "svc-1":
			if !s.Healthy {
				t.Error("svc-1 should be healthy")
			}
		case "svc-2":
			if s.Healthy {
				t.Error("svc-2 should be unhealthy")
			}
		}
	}
}

// --- Config Parsing Tests ---

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	configContent := `{
		"containers": [
			{
				"name": "web-server",
				"command": "echo",
				"args": ["hello"],
				"restart_policy": "on-failure",
				"max_restarts": 3,
				"health_check": {
					"command": "echo",
					"args": ["ok"],
					"interval_sec": 5,
					"timeout_sec": 2,
					"retries": 3
				}
			},
			{
				"name": "worker",
				"command": "echo",
				"args": ["work"],
				"restart_policy": "always"
			}
		],
		"log_file": "test-orchestrator.log",
		"api_port": 9090
	}`

	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	o := NewOrchestrator(logger)
	cfg, err := o.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Containers) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(cfg.Containers))
	}
	if cfg.APIPort != 9090 {
		t.Errorf("expected API port 9090, got %d", cfg.APIPort)
	}
	if cfg.LogFile != "test-orchestrator.log" {
		t.Errorf("expected log file 'test-orchestrator.log', got '%s'", cfg.LogFile)
	}

	// Verify containers were registered.
	c1, ok := o.GetContainer("web-server")
	if !ok {
		t.Fatal("web-server container not found")
	}
	if c1.Config.HealthCheck == nil {
		t.Error("web-server should have a health check")
	}
	if c1.Config.HealthCheck.IntervalSec != 5 {
		t.Errorf("expected health check interval 5, got %d", c1.Config.HealthCheck.IntervalSec)
	}

	c2, ok := o.GetContainer("worker")
	if !ok {
		t.Fatal("worker container not found")
	}
	if c2.Config.RestartPolicy != "always" {
		t.Errorf("expected worker restart policy 'always', got '%s'", c2.Config.RestartPolicy)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	configPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	o := NewOrchestrator(logger)
	_, err = o.LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig should fail on invalid JSON")
	}
}

func TestLoadConfigMissingCommand(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "test.log")
	logger, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("NewLogger failed: %v", err)
	}
	defer logger.Close()

	configContent := `{"containers": [{"name": "bad"}]}`
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	o := NewOrchestrator(logger)
	_, err = o.LoadConfig(configPath)
	if err == nil {
		t.Error("LoadConfig should fail when container has no command")
	}
}

// --- Parse Policy Tests ---

func TestParsePolicy(t *testing.T) {
	tests := []struct {
		input    string
		expected RestartPolicy
	}{
		{"always", PolicyAlways},
		{"on-failure", PolicyOnFailure},
		{"never", PolicyNever},
		{"unknown", PolicyNever},
		{"", PolicyNever},
	}

	for _, tt := range tests {
		result := parsePolicy(tt.input)
		if result != tt.expected {
			t.Errorf("parsePolicy(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}