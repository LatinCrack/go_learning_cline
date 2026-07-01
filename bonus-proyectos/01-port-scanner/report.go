package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// ScanReport represents the full scan report with metadata and results.
type ScanReport struct {
	Host       string        `json:"host"`
	StartPort  int           `json:"start_port"`
	EndPort    int           `json:"end_port"`
	Duration   string        `json:"duration"`
	TotalPorts int           `json:"total_ports"`
	OpenPorts  int           `json:"open_ports"`
	Results    []PortResult  `json:"results"`
}

// portResultJSON is an auxiliary struct for JSON serialization of PortResult,
// since the Err field is not directly JSON-serializable.
type portResultJSON struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Service string `json:"service,omitempty"`
	Latency string `json:"latency"`
	Error   string `json:"error,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for ScanReport,
// converting time.Duration fields to human-readable strings.
func (r ScanReport) MarshalJSON() ([]byte, error) {
	type Alias ScanReport
	results := make([]portResultJSON, len(r.Results))
	for i, pr := range r.Results {
		results[i] = portResultJSON{
			Port:    pr.Port,
			Open:    pr.Open,
			Service: pr.Service,
			Latency: pr.Latency.String(),
		}
		if pr.Err != nil {
			results[i].Error = pr.Err.Error()
		}
	}
	return json.Marshal(&struct {
		Alias
		Results []portResultJSON `json:"results"`
	}{
		Alias:   (Alias)(r),
		Results: results,
	})
}

// PrintTable writes a formatted ASCII table of open ports to the provided writer.
// Results are sorted by port number for consistent output.
func PrintTable(w io.Writer, results []PortResult, host string, duration time.Duration) {
	open := FilterOpen(results)
	if len(open) == 0 {
		fmt.Fprintf(w, "\n  No open ports found on %s (scanned %d ports in %s)\n\n", host, len(results), duration.Round(time.Millisecond))
		return
	}

	// Sort by port number.
	sort.Slice(open, func(i, j int) bool {
		return open[i].Port < open[j].Port
	})

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  ╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(w, "  ║              PORT SCANNER — SCAN RESULTS                    ║\n")
	fmt.Fprintf(w, "  ╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Fprintf(w, "  ║  Host: %-52s ║\n", host)
	fmt.Fprintf(w, "  ║  Open Ports: %-46d ║\n", len(open))
	fmt.Fprintf(w, "  ║  Scanned: %-49s ║\n", fmt.Sprintf("%d ports in %s", len(results), duration.Round(time.Millisecond)))
	fmt.Fprintf(w, "  ╠══════════════════════════════════════════════════════════════╣\n")
	fmt.Fprintf(w, "  ║  %-8s  %-12s  %-14s  %-18s ║\n", "PORT", "STATE", "SERVICE", "LATENCY")
	fmt.Fprintf(w, "  ╠══════════════════════════════════════════════════════════════╣\n")

	for _, r := range open {
		fmt.Fprintf(w, "  ║  %-8d  %-12s  %-14s  %-18s ║\n",
			r.Port,
			"OPEN",
			r.Service,
			r.Latency.Round(time.Microsecond).String(),
		)
	}

	fmt.Fprintf(w, "  ╚══════════════════════════════════════════════════════════════╝\n")
	fmt.Fprintf(w, "\n")
}

// PrintJSON writes the scan results as formatted JSON to the provided writer.
func PrintJSON(w io.Writer, results []PortResult, cfg ScanConfig, duration time.Duration) error {
	report := ScanReport{
		Host:       cfg.Host,
		StartPort:  cfg.StartPort,
		EndPort:    cfg.EndPort,
		Duration:   duration.String(),
		TotalPorts: len(results),
		OpenPorts:  len(FilterOpen(results)),
		Results:    FilterOpen(results),
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// PrintSummary writes a brief summary line to the provided writer.
func PrintSummary(w io.Writer, results []PortResult, host string, duration time.Duration) {
	open := FilterOpen(results)
	ports := make([]string, len(open))
	for i, r := range open {
		ports[i] = fmt.Sprintf("%d/%s", r.Port, r.Service)
	}

	fmt.Fprintf(w, "\n  Host: %s | Scanned: %d | Open: %d | Duration: %s\n",
		host, len(results), len(open), duration.Round(time.Millisecond))

	if len(open) > 0 {
		fmt.Fprintf(w, "  Open ports: %s\n\n", strings.Join(ports, ", "))
	} else {
		fmt.Fprintf(w, "  No open ports found.\n\n")
	}
}