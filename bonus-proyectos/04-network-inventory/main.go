package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

const banner = `
  _   _      _    _____                   _                 
 | \ | |    | |  |_   _|__  __ _ _ __ ___(_) ___  _ __  ___ 
 |  \| | ___| |_   | |/ _ \/ _` + "`" + ` | '__/ __| |/ _ \| '_ \/ __|
 | |\  |/ _ \ __|  | |  __/ (_| | | | (__| | (_) | | | \__ \
 |_| \_|\___/\__|  |_|\___|\__,_|_|  \___|_|\___/|_| |_|___/
                                                             
  Concurrent Network Inventory Scanner v1.0
`

func main() {
	// ── CLI Flags ──────────────────────────────────────────────
	cidr := flag.String("cidr", "", "CIDR range to scan (e.g., '192.168.1.0/24') [required]")
	timeout := flag.Duration("timeout", 3*time.Second, "Ping timeout per host (e.g., 1s, 2s, 3s)")
	workers := flag.Int("workers", 50, "Maximum concurrent ping workers (goroutines)")
	output := flag.String("output", "table", "Output format: 'table', 'json', or 'summary'")
	jsonFile := flag.String("json-file", "", "Path to write JSON report to disk (e.g., 'inventory.json')")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, banner)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  network-inventory -cidr <range> [options]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  network-inventory -cidr 192.168.1.0/24")
		fmt.Fprintln(os.Stderr, "  network-inventory -cidr 10.0.0.0/24 -workers 100 -timeout 1s")
		fmt.Fprintln(os.Stderr, "  network-inventory -cidr 172.16.0.0/16 -output json -json-file inventory.json")
		fmt.Fprintln(os.Stderr, "  network-inventory -cidr 192.168.1.0/24 -output summary")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// ── Handle version flag ────────────────────────────────────
	if *showVersion {
		fmt.Println("network-inventory v1.0.0")
		fmt.Println("Concurrent Network Inventory Scanner — Go Infrastructure Project")
		os.Exit(0)
	}

	// ── Validate required flags ────────────────────────────────
	if *cidr == "" {
		fmt.Fprintln(os.Stderr, "Error: -cidr flag is required.")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	// Validate CIDR format before starting.
	if err := ValidateCIDR(*cidr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// ── Validate workers ───────────────────────────────────────
	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "Error: workers must be at least 1")
		os.Exit(1)
	}
	if *workers > 1000 {
		fmt.Fprintln(os.Stderr, "Warning: capping workers at 1000 to prevent resource exhaustion")
		*workers = 1000
	}

	// ── Validate output format ─────────────────────────────────
	validFormats := map[string]bool{"table": true, "json": true, "summary": true}
	if !validFormats[*output] {
		fmt.Fprintf(os.Stderr, "Error: invalid output format '%s'. Use: table, json, summary\n", *output)
		os.Exit(1)
	}

	// ── Build discovery configuration ──────────────────────────
	cfg := DiscoveryConfig{
		CIDR:       *cidr,
		Timeout:    *timeout,
		MaxWorkers: *workers,
	}

	// ── Print scan banner ──────────────────────────────────────
	fmt.Fprint(os.Stderr, banner)
	fmt.Fprintf(os.Stderr, "  Range:      %s\n", cfg.CIDR)
	fmt.Fprintf(os.Stderr, "  Timeout:    %s\n", cfg.Timeout)
	fmt.Fprintf(os.Stderr, "  Workers:    %d\n", cfg.MaxWorkers)
	fmt.Fprintf(os.Stderr, "  Output:     %s\n", *output)
	fmt.Fprintf(os.Stderr, "\n  Discovery in progress...\n\n")

	// ── Execute discovery ──────────────────────────────────────
	ctx := context.Background()
	start := time.Now()

	results, err := DiscoverHosts(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during discovery: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(start)

	// ── Output results ─────────────────────────────────────────
	switch *output {
	case "table":
		PrintTable(os.Stdout, results, *cidr, duration)
	case "json":
		if err := ExportJSON(os.Stdout, results, *cidr, duration); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(1)
		}
	case "summary":
		PrintSummary(os.Stdout, results, *cidr, duration)
	}

	// ── Export JSON to file if requested ───────────────────────
	if *jsonFile != "" {
		if err := ExportJSONFile(*jsonFile, results, *cidr, duration); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "  ✓ JSON report saved to: %s\n", *jsonFile)
	}

	fmt.Fprintf(os.Stderr, "  ✓ Scan completed in %s\n", duration.Round(time.Millisecond))
}