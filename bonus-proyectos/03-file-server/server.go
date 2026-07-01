package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Config holds all configuration parameters for the file server.
type Config struct {
	// Host is the address to bind the server to (e.g., "0.0.0.0").
	Host string

	// Port is the TCP port to listen on.
	Port int

	// Root is the absolute or relative path to the directory of static files.
	Root string

	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum time to wait for the next request when keep-alives are enabled.
	IdleTimeout time.Duration

	// ShutdownTimeout is the maximum time to wait for graceful shutdown.
	ShutdownTimeout time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Host:            "0.0.0.0",
		Port:            8080,
		Root:            "./public",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 15 * time.Second,
	}
}

// Server wraps http.Server with additional lifecycle management
// including graceful shutdown on OS signals.
type Server struct {
	httpServer *http.Server
	config     Config
}

// NewServer creates a new Server with the given configuration.
// It assembles the middleware chain: Logging → SecurityHeaders → Gzip → FileHandler.
func NewServer(cfg Config) (*Server, error) {
	// Validate that the root directory exists and is accessible
	rootInfo, err := os.Stat(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("root directory %q: %w", cfg.Root, err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("root path %q is not a directory", cfg.Root)
	}

	// Build the handler chain (innermost to outermost)
	fileHandler := NewFileHandler(cfg.Root)

	// Middleware chain: Logging → SecurityHeaders → Gzip → FileHandler
	var handler http.Handler = fileHandler
	handler = GzipMiddleware(handler)
	handler = SecurityHeadersMiddleware(handler)
	handler = LoggingMiddleware(handler)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &Server{
		httpServer: httpServer,
		config:     cfg,
	}, nil
}

// Start begins listeningAndServe in a goroutine and blocks until an OS signal
// triggers graceful shutdown (SIGINT or SIGTERM).
func (s *Server) Start() error {
	// Channel to receive server errors
	errCh := make(chan error, 1)

	// Start the HTTP server in a goroutine
	go func() {
		log.Printf("🚀 Server listening on http://%s:%d", s.config.Host, s.config.Port)
		log.Printf("📂 Serving files from: %s", s.config.Root)
		log.Printf("🔧 ReadTimeout: %s | WriteTimeout: %s | IdleTimeout: %s",
			s.config.ReadTimeout, s.config.WriteTimeout, s.config.IdleTimeout)
		log.Println("─────────────────────────────────────────────")

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// Listen for OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("\n⏹  Received signal: %s. Initiating graceful shutdown...", sig)
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	return s.Shutdown()
}

// Shutdown gracefully shuts down the server, waiting for in-flight requests
// to complete up to the configured timeout.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	log.Printf("⏳ Waiting up to %s for in-flight requests...", s.config.ShutdownTimeout)

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	log.Println("✅ Server stopped gracefully.")
	return nil
}
