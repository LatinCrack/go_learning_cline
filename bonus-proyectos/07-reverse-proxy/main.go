// Package main implements a simple HTTP/HTTPS reverse proxy with load balancing,
// health checking, and graceful shutdown capabilities.
//
// Usage:
//
//	reverse-proxy --port 8080 --backends http://localhost:9001,http://localhost:9002,http://localhost:9003
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// Config holds the runtime configuration parsed from CLI flags.
type Config struct {
	Port           int
	Backends       []string
	HealthInterval time.Duration
	HealthTimeout  time.Duration
	RequestTimeout time.Duration
}

// parseFlags parses CLI flags and returns a validated Config.
func parseFlags() Config {
	var (
		port           int
		backendsRaw    string
		healthInterval time.Duration
		healthTimeout  time.Duration
		requestTimeout time.Duration
	)

	flag.IntVar(&port, "port", 8080, "Port where the reverse proxy listens")
	flag.StringVar(&backendsRaw, "backends", "", "Comma-separated list of backend URLs (e.g. http://localhost:9001,http://localhost:9002)")
	flag.DurationVar(&healthInterval, "health-interval", 10*time.Second, "Interval between health checks")
	flag.DurationVar(&healthTimeout, "health-timeout", 3*time.Second, "Timeout for each health check request")
	flag.DurationVar(&requestTimeout, "request-timeout", 30*time.Second, "Timeout for proxied requests to backends")

	flag.Parse()

	if backendsRaw == "" {
		fmt.Fprintln(os.Stderr, "ERROR: --backends flag is required")
		flag.Usage()
		os.Exit(1)
	}

	var backends []string
	for _, b := range strings.Split(backendsRaw, ",") {
		b = strings.TrimSpace(b)
		if b != "" {
			backends = append(backends, b)
		}
	}

	if len(backends) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: at least one backend must be specified")
		os.Exit(1)
	}

	return Config{
		Port:           port,
		Backends:       backends,
		HealthInterval: healthInterval,
		HealthTimeout:  healthTimeout,
		RequestTimeout: requestTimeout,
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	cfg := parseFlags()

	// Initialize the round-robin balancer with the configured backends.
	balancer := NewBalancer(cfg.Backends)

	// Initialize health checker that runs in background.
	hc := NewHealthChecker(balancer, cfg.HealthInterval, cfg.HealthTimeout)
	hc.Start()
	defer hc.Stop()

	// Build the reverse proxy handler chain with middleware.
	proxy := NewReverseProxy(balancer, cfg.RequestTimeout)
	handler := LoggingMiddleware(proxy)

	// Configure the HTTP server.
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to receive OS signals for graceful shutdown.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Channel to capture server errors.
	errCh := make(chan error, 1)

	go func() {
		log.Printf("[main] Reverse proxy listening on :%d", cfg.Port)
		log.Printf("[main] Backends: %v", cfg.Backends)
		log.Printf("[main] Health check interval: %s", cfg.HealthInterval)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for either a shutdown signal or a server error.
	select {
	case sig := <-stop:
		log.Printf("[main] Received signal: %s — initiating graceful shutdown", sig)
	case err := <-errCh:
		log.Printf("[main] Server error: %v — initiating graceful shutdown", err)
	}

	// Create a deadline context for the graceful shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[main] Forced shutdown due to error: %v", err)
	}

	log.Println("[main] Server stopped gracefully")
}
