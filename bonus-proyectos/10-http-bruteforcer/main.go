package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// HTTP Bruteforcer — A pentesting tool for auditing HTTP authentication endpoints.
//
// This tool performs dictionary-based credential testing against HTTP login forms
// and Basic Auth endpoints with configurable rate limiting, proxy rotation, and
// automatic login form detection.
//
// ⚠️  ETHICAL USE ONLY: This tool is for authorized security testing only.
//     Unauthorized access to computer systems is illegal.
//
// Usage:
//
//	go run . -config config.json
//	go run . -url http://target/login -users users.txt -passwords passwords.txt
//	go run . -config config.json -v -workers 20 -delay 1000
func main() {
	// CLI flags — these override config file values when specified.
	configFile := flag.String("config", "", "Path to JSON configuration file")
	urlFlag := flag.String("url", "", "Target URL (overrides config)")
	usersFile := flag.String("users", "", "Path to users dictionary file (overrides config)")
	passwordsFile := flag.String("passwords", "", "Path to passwords dictionary file (overrides config)")
	workers := flag.Int("workers", 0, "Number of concurrent workers (overrides config)")
	delay := flag.Int("delay", 0, "Delay between requests in milliseconds (overrides config)")
	burstSize := flag.Int("burst", 0, "Burst size before throttling (overrides config)")
	burstDelay := flag.Int("burst-delay", 0, "Delay after burst exhaustion in ms (overrides config)")
	timeout := flag.Int("timeout", 0, "HTTP request timeout in seconds (overrides config)")
	retries := flag.Int("retries", 0, "Max retries per request on error (overrides config)")
	verbose := flag.Bool("v", false, "Verbose output (show all attempts)")
	proxyList := flag.String("proxies", "", "Comma-separated proxy URLs (overrides config)")
	outputFile := flag.String("output", "", "Path to JSON report file (overrides config)")
	detectForm := flag.Bool("detect", false, "Auto-detect login form fields before attack")
	successIndicator := flag.String("success", "", "String in response body indicating success (overrides config)")
	failureIndicator := flag.String("failure", "", "String in response body indicating failure (overrides config)")
	successStatus := flag.Int("success-status", 0, "HTTP status code indicating success (overrides config)")
	userAgent := flag.String("ua", "", "Custom User-Agent string (overrides config)")

	flag.Parse()

	// Load configuration.
	var cfg *Config
	if *configFile != "" {
		var err error
		cfg, err = LoadConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg = DefaultConfig()
	}

	// Apply CLI overrides.
	if *urlFlag != "" {
		cfg.TargetURL = *urlFlag
	}
	if *usersFile != "" {
		cfg.UsersFile = *usersFile
	}
	if *passwordsFile != "" {
		cfg.PasswordsFile = *passwordsFile
	}
	if *workers > 0 {
		cfg.Workers = *workers
	}
	if *delay > 0 {
		cfg.Delay = *delay
	}
	if *burstSize > 0 {
		cfg.BurstSize = *burstSize
	}
	if *burstDelay > 0 {
		cfg.BurstDelay = *burstDelay
	}
	if *timeout > 0 {
		cfg.Timeout = *timeout
	}
	if *retries > 0 {
		cfg.MaxRetries = *retries
	}
	if *verbose {
		cfg.Verbose = true
	}
	if *proxyList != "" {
		cfg.Proxies = parseCommaSeparated(*proxyList)
	}
	if *outputFile != "" {
		cfg.OutputFile = *outputFile
	}
	if *successIndicator != "" {
		cfg.SuccessIndicator = *successIndicator
	}
	if *failureIndicator != "" {
		cfg.FailureIndicator = *failureIndicator
	}
	if *successStatus > 0 {
		cfg.SuccessStatus = *successStatus
	}
	if *userAgent != "" {
		cfg.UserAgent = *userAgent
	}

	// Validate configuration.
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nUse -h for help.\n")
		os.Exit(1)
	}

	// Auto-detect login form if requested.
	if *detectForm && cfg.TargetType == "form" {
		fmt.Println("  🔍 Auto-detecting login form...")
		detection, err := detectFormFields(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Form detection failed: %v\n", err)
		} else if detection.Found {
			fmt.Printf("  %s\n", detection.String())
			if detection.UsernameField != "" {
				cfg.UsernameField = detection.UsernameField
			}
			if detection.PasswordField != "" {
				cfg.PasswordField = detection.PasswordField
			}
			if detection.FormAction != "" && cfg.TargetURL == "" {
				cfg.TargetURL = detection.FormAction
			}
			fmt.Println()
		} else {
			fmt.Println("  ⚠ No login form detected. Using configured field names.")
			fmt.Println()
		}
	}

	// Load wordlists.
	fmt.Printf("  Loading wordlists...\n")
	users, err := LoadWordlist("users", cfg.UsersFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	passwords, err := LoadWordlist("passwords", cfg.PasswordsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	totalCombinations := users.Len() * passwords.Len()
	fmt.Printf("  Users loaded:      %d\n", users.Len())
	fmt.Printf("  Passwords loaded:  %d\n", passwords.Len())
	fmt.Printf("  Combinations:      %d\n", totalCombinations)
	fmt.Println()

	// Create bruteforcer.
	bf, err := NewBruteforcer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer bf.Close()

	// Setup context with signal handling (Ctrl+C graceful shutdown).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Fprintf(os.Stderr, "\n\n  ⚠ Received signal %v — shutting down gracefully...\n", sig)
		cancel()
	}()

	// Generate credential pairs and run the attack.
	pairs := GeneratePairs(users, passwords)
	found := bf.Run(ctx, pairs)

	// Exit with appropriate code.
	if len(found) > 0 {
		os.Exit(0) // Credentials found.
	}
	os.Exit(2) // No credentials found.
}

// detectFormFields fetches the login page and attempts to detect form fields.
func detectFormFields(cfg *Config) (FormDetection, error) {
	client, err := NewHTTPClient(cfg)
	if err != nil {
		return FormDetection{}, fmt.Errorf("create HTTP client: %w", err)
	}
	defer client.Close()

	// GET the login page.
	resp := client.AttemptLogin("", "")
	if resp.Error != nil {
		return FormDetection{}, fmt.Errorf("fetch login page: %w", resp.Error)
	}

	return DetectLoginForm(resp.Body, cfg.TargetURL), nil
}

// parseCommaSeparated splits a comma-separated string into a slice.
func parseCommaSeparated(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else if c != ' ' && c != '\t' {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}