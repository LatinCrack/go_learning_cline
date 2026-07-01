package main

import (
	"flag"
	"fmt"
	"os"
)

// Mini-Docker: A simplified container orchestrator with automatic restart,
// health checks, centralized logging, and a REST control API.
//
// Usage:
//
//	go run . -config containers.json
//	go run . -config containers.json -port 9090
func main() {
	configPath := flag.String("config", "containers.json", "Path to the JSON configuration file")
	apiPort := flag.Int("port", 0, "API server port (overrides config file)")
	logFile := flag.String("log", "", "Log file path (overrides config file)")
	flag.Parse()

	// Validate config file exists.
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error: config file not found: %s\n", *configPath)
		fmt.Fprintf(os.Stderr, "use -config <path> to specify the configuration file\n")
		os.Exit(1)
	}

	// Create the logger. We start with a default, then override if config specifies one.
	logger, err := NewLogger("orchestrator.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Log("main", "info", "=== Mini-Docker Orchestrator starting ===")

	// Create orchestrator and load configuration.
	orchestrator := NewOrchestrator(logger)
	cfg, err := orchestrator.LoadConfig(*configPath)
	if err != nil {
		logger.Log("main", "error", fmt.Sprintf("failed to load config: %v", err))
		os.Exit(1)
	}

	// Re-initialize logger if config specifies a different log file.
	if *logFile != "" {
		cfg.LogFile = *logFile
	}
	if cfg.LogFile != "orchestrator.log" {
		if newLogger, err := NewLogger(cfg.LogFile); err == nil {
			orchestrator.logger = newLogger
			orchestrator.health = NewHealthChecker(newLogger)
			orchestrator.restart = NewRestartSupervisor(newLogger)
			// Re-register containers with the new supervisor.
			for _, c := range orchestrator.containers {
				orchestrator.restart.Register(c)
			}
			logger.Close()
			logger = newLogger
		} else {
			logger.Log("main", "warn", fmt.Sprintf("could not open configured log file %s: %v", cfg.LogFile, err))
		}
	}

	// Determine API port.
	port := cfg.APIPort
	if *apiPort > 0 {
		port = *apiPort
	}
	if port == 0 {
		port = 8090
	}

	logger.Log("main", "info", fmt.Sprintf("managing %d container(s)", len(cfg.Containers)))
	for _, cc := range cfg.Containers {
		logger.Log("main", "info", fmt.Sprintf("  → %s: %s %v (policy=%s)", cc.Name, cc.Command, cc.Args, cc.RestartPolicy))
	}

	// Start the REST API server in a goroutine.
	api := NewAPIServer(orchestrator, logger, port)
	go func() {
		if err := api.Start(); err != nil {
			logger.Log("api", "error", fmt.Sprintf("API server error: %v", err))
		}
	}()

	// Start all containers.
	orchestrator.StartAll()

	logger.Log("main", "info", fmt.Sprintf("orchestrator ready — API at http://localhost:%d", port))
	logger.Log("main", "info", "press Ctrl+C or send SIGTERM to shut down")

	// Block until signal.
	orchestrator.WaitForSignal(15)
}