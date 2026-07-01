package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds all configuration for a bruteforce attack session.
type Config struct {
	// Target settings.
	TargetURL  string `json:"target_url"`  // URL of the login endpoint.
	TargetType string `json:"target_type"` // "form" or "basic" (HTTP Basic Auth).

	// Form field names (for form-based attacks).
	UsernameField string `json:"username_field"` // e.g. "username", "user", "email".
	PasswordField string `json:"password_field"` // e.g. "password", "pass".

	// Detection of success/failure.
	SuccessIndicator string `json:"success_indicator"` // String or pattern in response that indicates success.
	FailureIndicator string `json:"failure_indicator"` // String or pattern in response that indicates failure.
	SuccessStatus    int    `json:"success_status"`     // HTTP status code that indicates success (e.g. 302 redirect).

	// Wordlist paths.
	UsersFile     string `json:"users_file"`     // Path to usernames dictionary.
	PasswordsFile string `json:"passwords_file"` // Path to passwords dictionary.

	// Rate limiting and concurrency.
	Workers       int     `json:"workers"`        // Number of concurrent workers.
	Delay         int     `json:"delay_ms"`       // Delay between requests per worker (milliseconds).
	BurstSize     int     `json:"burst_size"`     // Requests allowed in burst before throttling.
	BurstDelay    int     `json:"burst_delay_ms"` // Delay after burst is exhausted (milliseconds).
	MaxRetries    int     `json:"max_retries"`    // Max retries per request on network error.
	Timeout       int     `json:"timeout_sec"`    // HTTP request timeout in seconds.
	JitterEnabled bool    `json:"jitter"`         // Add random jitter to delays.
	JitterFactor  float64 `json:"jitter_factor"`  // Jitter factor (0.0-1.0) as fraction of delay.

	// Proxy settings.
	Proxies []string `json:"proxies"` // List of proxy URLs (http://host:port).

	// Output settings.
	OutputFile string `json:"output_file"` // Path to JSON report output.
	Verbose    bool   `json:"verbose"`     // Enable verbose output.

	// HTTP headers (custom).
	CustomHeaders map[string]string `json:"custom_headers"` // Extra headers to send with each request.
	Cookies       map[string]string `json:"cookies"`        // Pre-set cookies.
	UserAgent     string            `json:"user_agent"`     // Custom User-Agent string.
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		TargetURL:        "",
		TargetType:       "form",
		UsernameField:    "username",
		PasswordField:    "password",
		SuccessIndicator: "",
		FailureIndicator: "invalid",
		SuccessStatus:    0,
		UsersFile:        "",
		PasswordsFile:    "",
		Workers:          10,
		Delay:            500,
		BurstSize:        20,
		BurstDelay:       5000,
		MaxRetries:       3,
		Timeout:          10,
		JitterEnabled:    true,
		JitterFactor:     0.3,
		Proxies:          nil,
		OutputFile:       "report.json",
		Verbose:          false,
		CustomHeaders:    make(map[string]string),
		Cookies:          make(map[string]string),
		UserAgent:        "HTTP-Bruteforcer/1.0 (Security Audit Tool)",
	}
}

// LoadConfig reads a JSON configuration file and merges it with defaults.
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}

// Validate checks that the configuration is valid and complete.
func (c *Config) Validate() error {
	if c.TargetURL == "" {
		return fmt.Errorf("target_url is required")
	}
	if c.TargetType != "form" && c.TargetType != "basic" {
		return fmt.Errorf("target_type must be 'form' or 'basic', got '%s'", c.TargetType)
	}
	if c.UsersFile == "" {
		return fmt.Errorf("users_file is required")
	}
	if c.PasswordsFile == "" {
		return fmt.Errorf("passwords_file is required")
	}
	if c.Workers < 1 {
		return fmt.Errorf("workers must be >= 1, got %d", c.Workers)
	}
	if c.Workers > 200 {
		return fmt.Errorf("workers must be <= 200, got %d", c.Workers)
	}
	if c.Delay < 0 {
		return fmt.Errorf("delay_ms must be >= 0, got %d", c.Delay)
	}
	if c.BurstSize < 1 {
		return fmt.Errorf("burst_size must be >= 1, got %d", c.BurstSize)
	}
	if c.Timeout < 1 {
		return fmt.Errorf("timeout_sec must be >= 1, got %d", c.Timeout)
	}
	if c.JitterFactor < 0 || c.JitterFactor > 1 {
		return fmt.Errorf("jitter_factor must be between 0.0 and 1.0, got %f", c.JitterFactor)
	}
	if c.SuccessIndicator == "" && c.FailureIndicator == "" && c.SuccessStatus == 0 {
		return fmt.Errorf("at least one of success_indicator, failure_indicator, or success_status must be set")
	}
	return nil
}

// SaveConfig writes the configuration to a JSON file.
func SaveConfig(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// AttackDelay returns the base delay with optional jitter applied.
func (c *Config) AttackDelay() time.Duration {
	base := time.Duration(c.Delay) * time.Millisecond
	if !c.JitterEnabled || c.JitterFactor == 0 {
		return base
	}
	return applyJitter(base, c.JitterFactor)
}