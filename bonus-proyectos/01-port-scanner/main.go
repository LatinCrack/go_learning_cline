package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const banner = `
  ____            _     ____                                 
 |  _ \ ___  _ __| |_  / ___|  ___ __ _ _ __  _ __   ___ _ __ 
 | |_) / _ \| '__| __| \___ \ / __/ _' | '_ \| '_ \ / _ \ '__|
 |  __/ (_) | |  | |_   ___) | (_| (_| | | | | | | |  __/ |   
 |_|   \___/|_|   \__| |____/ \___\__,_|_| |_|_| |_|\___|_|   
                                                               
  Concurrent TCP Port Scanner v1.0
`

func main() {
	// ── CLI Flags ──────────────────────────────────────────────
	host := flag.String("host", "", "Target host (IP or hostname) [required]")
	portRange := flag.String("ports", "1-1024", "Port range to scan (e.g., '80', '1-1024', '22,80,443')")
	timeout := flag.Duration("timeout", 500*time.Millisecond, "Connection timeout per port (e.g., 500ms, 1s, 2s)")
	workers := flag.Int("workers", 100, "Maximum concurrent workers (goroutines)")
	output := flag.String("output", "table", "Output format: 'table', 'json', or 'summary'")
	showAll := flag.Bool("all", false, "Show all ports including closed ones (table format only)")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, banner)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  port-scanner -host <target> [options]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  port-scanner -host 192.168.1.1")
		fmt.Fprintln(os.Stderr, "  port-scanner -host scanme.nmap.org -ports 1-1024 -workers 200")
		fmt.Fprintln(os.Stderr, "  port-scanner -host 10.0.0.1 -ports 22,80,443,3306 -timeout 2s")
		fmt.Fprintln(os.Stderr, "  port-scanner -host 192.168.1.1 -ports 1-65535 -workers 1000 -output json")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// ── Handle version flag ────────────────────────────────────
	if *showVersion {
		fmt.Println("port-scanner v1.0.0")
		fmt.Println("Concurrent TCP Port Scanner — Go Infrastructure Project")
		os.Exit(0)
	}

	// ── Validate required flags ────────────────────────────────
	if *host == "" {
		fmt.Fprintln(os.Stderr, "Error: -host flag is required.")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	// ── Parse port range ───────────────────────────────────────
	startPort, endPort, err := parsePortRange(*portRange)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid port range '%s': %v\n", *portRange, err)
		os.Exit(1)
	}

	// ── Validate workers ───────────────────────────────────────
	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "Error: workers must be at least 1")
		os.Exit(1)
	}
	if *workers > 10000 {
		fmt.Fprintln(os.Stderr, "Warning: capping workers at 10000 to prevent resource exhaustion")
		*workers = 10000
	}

	// ── Validate output format ─────────────────────────────────
	validFormats := map[string]bool{"table": true, "json": true, "summary": true}
	if !validFormats[*output] {
		fmt.Fprintf(os.Stderr, "Error: invalid output format '%s'. Use: table, json, summary\n", *output)
		os.Exit(1)
	}

	// ── Build scan configuration ───────────────────────────────
	cfg := ScanConfig{
		Host:       *host,
		StartPort:  startPort,
		EndPort:    endPort,
		Timeout:    *timeout,
		MaxWorkers: *workers,
	}

	// ── Print scan banner ──────────────────────────────────────
	fmt.Fprint(os.Stderr, banner)
	fmt.Fprintf(os.Stderr, "  Target:     %s\n", cfg.Host)
	fmt.Fprintf(os.Stderr, "  Ports:      %d - %d (%d ports)\n", cfg.StartPort, cfg.EndPort, endPort-startPort+1)
	fmt.Fprintf(os.Stderr, "  Timeout:    %s\n", cfg.Timeout)
	fmt.Fprintf(os.Stderr, "  Workers:    %d\n", cfg.MaxWorkers)
	fmt.Fprintf(os.Stderr, "  Output:     %s\n", *output)
	fmt.Fprintf(os.Stderr, "\n  Scanning in progress...\n")

	// ── Execute scan ───────────────────────────────────────────
	start := time.Now()
	results := ScanPorts(cfg)
	duration := time.Since(start)

	// ── Output results ─────────────────────────────────────────
	switch *output {
	case "table":
		if *showAll {
			printFullTable(os.Stdout, results, cfg.Host, duration)
		} else {
			PrintTable(os.Stdout, results, cfg.Host, duration)
		}
	case "json":
		if err := PrintJSON(os.Stdout, results, cfg, duration); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(1)
		}
	case "summary":
		PrintSummary(os.Stdout, results, cfg.Host, duration)
	}
}

// parsePortRange parses a port range string into start and end ports.
// Supported formats:
//   - "80"        → single port (80, 80)
//   - "1-1024"    → range (1, 1024)
//   - "22,80,443" → comma-separated → expands to min and max
func parsePortRange(s string) (int, int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, 0, fmt.Errorf("empty port range")
	}

	// Check if it's a range (e.g., "1-1024")
	if strings.Contains(s, "-") && !strings.Contains(s, ",") {
		parts := strings.SplitN(s, "-", 2)
		var start, end int
		_, err := fmt.Sscanf(parts[0], "%d", &start)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid start port '%s': %w", parts[0], err)
		}
		_, err = fmt.Sscanf(parts[1], "%d", &end)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid end port '%s': %w", parts[1], err)
		}
		if start < 1 || start > 65535 {
			return 0, 0, fmt.Errorf("start port must be between 1 and 65535, got %d", start)
		}
		if end < 1 || end > 65535 {
			return 0, 0, fmt.Errorf("end port must be between 1 and 65535, got %d", end)
		}
		if start > end {
			return 0, 0, fmt.Errorf("start port (%d) cannot be greater than end port (%d)", start, end)
		}
		return start, end, nil
	}

	// Check if it's comma-separated (e.g., "22,80,443")
	if strings.Contains(s, ",") {
		parts := strings.Split(s, ",")
		minPort, maxPort := 65535, 0
		for _, p := range parts {
			var port int
			_, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &port)
			if err != nil {
				return 0, 0, fmt.Errorf("invalid port '%s': %w", p, err)
			}
			if port < 1 || port > 65535 {
				return 0, 0, fmt.Errorf("port must be between 1 and 65535, got %d", port)
			}
			if port < minPort {
				minPort = port
			}
			if port > maxPort {
				maxPort = port
			}
		}
		return minPort, maxPort, nil
	}

	// Single port
	var port int
	_, err := fmt.Sscanf(s, "%d", &port)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid port '%s': %w", s, err)
	}
	if port < 1 || port > 65535 {
		return 0, 0, fmt.Errorf("port must be between 1 and 65535, got %d", port)
	}
	return port, port, nil
}

// printFullTable writes a formatted ASCII table showing all ports (open and closed).
func printFullTable(w *os.File, results []PortResult, host string, duration time.Duration) {
	fmt.Fprintf(w, "\n  Full scan results for %s (%d ports scanned in %s):\n\n", host, len(results), duration.Round(time.Millisecond))
	fmt.Fprintf(w, "  %-8s  %-12s  %-14s  %-18s\n", "PORT", "STATE", "SERVICE", "LATENCY")
	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", 58))

	for _, r := range results {
		state := "CLOSED"
		if r.Open {
			state = "OPEN"
		}
		svc := r.Service
		if svc == "" {
			svc = "—"
		}
		fmt.Fprintf(w, "  %-8d  %-12s  %-14s  %-18s\n",
			r.Port,
			state,
			svc,
			r.Latency.Round(time.Microsecond).String(),
		)
	}
	fmt.Fprintln(w)
}