package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

// Reporter handles output formatting and real-time progress display.
type Reporter struct {
	config     *Config
	lastUpdate atomic.Int64
}

// NewReporter creates a new Reporter for the given config.
func NewReporter(cfg *Config) *Reporter {
	return &Reporter{config: cfg}
}

// PrintBanner displays the attack configuration banner.
func (r *Reporter) PrintBanner(cfg *Config) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║           HTTP BRUTEFORCER — Security Audit Tool            ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  Target:       %s\n", cfg.TargetURL)
	fmt.Printf("  Type:         %s\n", cfg.TargetType)
	if cfg.TargetType == "form" {
		fmt.Printf("  Username:     %s\n", cfg.UsernameField)
		fmt.Printf("  Password:     %s\n", cfg.PasswordField)
	}
	fmt.Printf("  Workers:      %d\n", cfg.Workers)
	fmt.Printf("  Delay:        %dms\n", cfg.Delay)
	fmt.Printf("  Burst size:   %d\n", cfg.BurstSize)
	fmt.Printf("  Burst delay:  %dms\n", cfg.BurstDelay)
	fmt.Printf("  Jitter:       %v (factor=%.2f)\n", cfg.JitterEnabled, cfg.JitterFactor)
	fmt.Printf("  Timeout:      %ds\n", cfg.Timeout)
	fmt.Printf("  Retries:      %d\n", cfg.MaxRetries)
	if len(cfg.Proxies) > 0 {
		fmt.Printf("  Proxies:      %d configured\n", len(cfg.Proxies))
	} else {
		fmt.Printf("  Proxies:      none (direct)\n")
	}
	fmt.Println()
	fmt.Println("  ────────────────────────────────────────────────────────────")
	fmt.Println()
}

// Start consumes results from the channel and prints real-time progress.
func (r *Reporter) Start(results <-chan AttemptResult, stats *Stats, startTime time.Time) {
	for result := range results {
		r.printResult(result, stats, startTime)
	}
}

// printResult displays a single attempt result.
func (r *Reporter) printResult(result AttemptResult, stats *Stats, startTime time.Time) {
	total := stats.TotalAttempts.Load()
	elapsed := time.Since(startTime)

	var statusStr string
	switch result.Status {
	case "success":
		statusStr = "\033[32m✓ SUCCESS\033[0m" // Green
	case "failure":
		if r.config.Verbose {
			statusStr = "\033[31m✗ failed\033[0m" // Red
		} else {
			// In non-verbose mode, only print successes, errors, and blocks.
			return
		}
	case "blocked":
		statusStr = "\033[33m⚠ BLOCKED\033[0m" // Yellow
	case "error":
		statusStr = "\033[35m✖ ERROR\033[0m" // Magenta
	default:
		statusStr = "? unknown"
	}

	// Calculate rate.
	var rate float64
	if elapsed.Seconds() > 0 {
		rate = float64(total) / elapsed.Seconds()
	}

	// Format the output line.
	fmt.Printf("  [%s] %s:%s → %s  (HTTP %d | %v)  [attempts: %d | %.1f req/s]\n",
		result.Timestamp.Format("15:04:05"),
		result.Username,
		truncateString(result.Password, 16),
		statusStr,
		result.HTTPCode,
		result.Duration.Round(time.Millisecond),
		total,
		rate,
	)

	if result.Error != nil && r.config.Verbose {
		fmt.Printf("           └─ Error: %v\n", result.Error)
	}
}

// PrintSummary displays the final attack summary and writes the JSON report.
func (r *Reporter) PrintSummary(stats *Stats, found []AttemptResult, totalDuration time.Duration) {
	total := stats.TotalAttempts.Load()
	success := stats.SuccessAttempts.Load()
	failed := stats.FailedAttempts.Load()
	errors := stats.ErrorAttempts.Load()
	blocked := stats.BlockedAttempts.Load()

	fmt.Println()
	fmt.Println("  ═══════════════════════════════════════════════════════════")
	fmt.Println("  ATTACK SUMMARY")
	fmt.Println("  ═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Total attempts:  %d\n", total)
	fmt.Printf("  Successes:       \033[32m%d\033[0m\n", success)
	fmt.Printf("  Failures:        %d\n", failed)
	fmt.Printf("  Errors:          %d\n", errors)
	fmt.Printf("  Blocked:         %d\n", blocked)
	fmt.Println()

	var rate float64
	if totalDuration.Seconds() > 0 {
		rate = float64(total) / totalDuration.Seconds()
	}
	fmt.Printf("  Duration:        %s\n", formatDuration(totalDuration))
	fmt.Printf("  Average rate:    %.1f req/s\n", rate)
	fmt.Println()

	if len(found) > 0 {
		fmt.Println("  ┌──────────────────────────────────────────────────────────┐")
		fmt.Println("  │                    🔓 CREDENTIALS FOUND                  │")
		fmt.Println("  ├──────────────────────────────────────────────────────────┤")
		for _, f := range found {
			fmt.Printf("  │  Username: %-20s  Password: %-18s │\n", f.Username, truncateString(f.Password, 18))
		}
		fmt.Println("  └──────────────────────────────────────────────────────────┘")
	} else {
		fmt.Println("  No valid credentials found.")
	}
	fmt.Println()

	// Write JSON report.
	if r.config.OutputFile != "" {
		if err := r.writeJSONReport(stats, found, totalDuration); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Warning: failed to write report: %v\n", err)
		} else {
			fmt.Printf("  Report saved to: %s\n", r.config.OutputFile)
		}
	}
}

// JSONReport is the structure for the exported report.
type JSONReport struct {
	Summary  ReportSummary    `json:"summary"`
	Config   ReportConfig     `json:"config"`
	Findings []ReportFinding  `json:"findings"`
}

// ReportSummary contains attack statistics.
type ReportSummary struct {
	TotalAttempts int64   `json:"total_attempts"`
	Successes     int64   `json:"successes"`
	Failures      int64   `json:"failures"`
	Errors        int64   `json:"errors"`
	Blocked       int64   `json:"blocked"`
	Duration      string  `json:"duration"`
	RatePerSecond float64 `json:"rate_per_second"`
}

// ReportConfig records the attack parameters used.
type ReportConfig struct {
	TargetURL  string   `json:"target_url"`
	TargetType string   `json:"target_type"`
	Workers    int      `json:"workers"`
	Delay      int      `json:"delay_ms"`
	BurstSize  int      `json:"burst_size"`
	Proxies    int      `json:"proxy_count"`
}

// ReportFinding is a single successful credential pair.
type ReportFinding struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	HTTPCode  int    `json:"http_status_code"`
	Timestamp string `json:"timestamp"`
}

// writeJSONReport writes the attack report to a JSON file.
func (r *Reporter) writeJSONReport(stats *Stats, found []AttemptResult, totalDuration time.Duration) error {
	total := stats.TotalAttempts.Load()
	var rate float64
	if totalDuration.Seconds() > 0 {
		rate = float64(total) / totalDuration.Seconds()
	}

	report := JSONReport{
		Summary: ReportSummary{
			TotalAttempts: total,
			Successes:     stats.SuccessAttempts.Load(),
			Failures:      stats.FailedAttempts.Load(),
			Errors:        stats.ErrorAttempts.Load(),
			Blocked:       stats.BlockedAttempts.Load(),
			Duration:      formatDuration(totalDuration),
			RatePerSecond: rate,
		},
		Config: ReportConfig{
			TargetURL:  r.config.TargetURL,
			TargetType: r.config.TargetType,
			Workers:    r.config.Workers,
			Delay:      r.config.Delay,
			BurstSize:  r.config.BurstSize,
			Proxies:    len(r.config.Proxies),
		},
	}

	for _, f := range found {
		report.Findings = append(report.Findings, ReportFinding{
			Username:  f.Username,
			Password:  f.Password,
			HTTPCode:  f.HTTPCode,
			Timestamp: f.Timestamp.Format(time.RFC3339),
		})
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	return os.WriteFile(r.config.OutputFile, data, 0644)
}