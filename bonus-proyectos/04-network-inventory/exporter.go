package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

// InventoryReport is the top-level structure exported to JSON.
type InventoryReport struct {
	ScanTime   time.Duration   `json:"scan_duration"`
	TotalHosts int             `json:"total_hosts"`
	AliveHosts int             `json:"alive_hosts"`
	Range      string          `json:"cidr_range"`
	Hosts      []HostJSONEntry `json:"hosts"`
}

// HostJSONEntry is the JSON-serializable representation of a discovered host.
type HostJSONEntry struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname,omitempty"`
	OSGuess  string `json:"os_guess,omitempty"`
	TTL      int    `json:"ttl"`
	Latency  string `json:"latency"`
	Alive    bool   `json:"alive"`
}

// ExportJSON writes the inventory report as formatted JSON to the given writer.
// It also writes the same report to a file on disk for persistence.
func ExportJSON(w io.Writer, results []HostResult, cidr string, duration time.Duration) error {
	report := buildReport(results, cidr, duration)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// ExportJSONFile writes the inventory report to a JSON file on disk.
func ExportJSONFile(filename string, results []HostResult, cidr string, duration time.Duration) error {
	report := buildReport(results, cidr, duration)

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", filename, err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// buildReport constructs the InventoryReport from raw discovery results.
func buildReport(results []HostResult, cidr string, duration time.Duration) InventoryReport {
	// Sort results by IP for consistent output.
	sort.Slice(results, func(i, j int) bool {
		return results[i].IP.Less(results[j].IP)
	})

	entries := make([]HostJSONEntry, 0, len(results))
	aliveCount := 0

	for _, r := range results {
		if r.Alive {
			aliveCount++
		}
		entries = append(entries, HostJSONEntry{
			IP:       r.IP.String(),
			Hostname: r.Hostname,
			OSGuess:  r.OSGuess,
			TTL:      r.TTL,
			Latency:  r.Latency.Round(time.Millisecond).String(),
			Alive:    r.Alive,
		})
	}

	return InventoryReport{
		ScanTime:   duration.Round(time.Millisecond),
		TotalHosts: len(results),
		AliveHosts: aliveCount,
		Range:      cidr,
		Hosts:      entries,
	}
}

// PrintTable writes a formatted ASCII table of alive hosts to the given writer.
// Dead hosts are omitted from the table for clarity; the JSON export includes all hosts.
func PrintTable(w io.Writer, results []HostResult, cidr string, duration time.Duration) {
	// Filter alive hosts and sort by IP.
	alive := filterAlive(results)
	sort.Slice(alive, func(i, j int) bool {
		return alive[i].IP.Less(alive[j].IP)
	})

	fmt.Fprintf(w, "\n  ╔══════════════════════════════════════════════════════════════════════════════════════╗\n")
	fmt.Fprintf(w, "  ║                         🌐 NETWORK INVENTORY REPORT                                ║\n")
	fmt.Fprintf(w, "  ╚══════════════════════════════════════════════════════════════════════════════════════╝\n\n")
	fmt.Fprintf(w, "  Range:      %s\n", cidr)
	fmt.Fprintf(w, "  Duration:   %s\n", duration.Round(time.Millisecond))
	fmt.Fprintf(w, "  Hosts up:   %d / %d\n", len(alive), len(results))
	fmt.Fprintf(w, "\n")

	if len(alive) == 0 {
		fmt.Fprintf(w, "  ⚠  No active hosts found in the specified range.\n\n")
		return
	}

	// Column headers.
	fmt.Fprintf(w, "  %-18s %-28s %-20s %-6s %-10s\n",
		"IP ADDRESS", "HOSTNAME", "OS GUESS", "TTL", "LATENCY")
	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", 86))

	for _, r := range alive {
		hostname := r.Hostname
		if hostname == "" {
			hostname = "—"
		}
		osGuess := r.OSGuess
		if osGuess == "" {
			osGuess = "Unknown"
		}

		fmt.Fprintf(w, "  %-18s %-28s %-20s %-6d %-10s\n",
			r.IP.String(),
			truncate(hostname, 27),
			osGuess,
			r.TTL,
			r.Latency.Round(time.Millisecond).String(),
		)
	}

	fmt.Fprintf(w, "\n")
}

// PrintSummary writes a brief summary of the scan results.
func PrintSummary(w io.Writer, results []HostResult, cidr string, duration time.Duration) {
	alive := filterAlive(results)

	fmt.Fprintf(w, "\n  Network Inventory Summary\n")
	fmt.Fprintf(w, "  ─────────────────────────\n")
	fmt.Fprintf(w, "  Range:     %s\n", cidr)
	fmt.Fprintf(w, "  Scanned:   %d hosts\n", len(results))
	fmt.Fprintf(w, "  Alive:     %d hosts\n", len(alive))
	fmt.Fprintf(w, "  Duration:  %s\n", duration.Round(time.Millisecond))

	if len(alive) > 0 {
		fmt.Fprintf(w, "\n  Active hosts:\n")
		for _, r := range alive {
			hostname := r.Hostname
			if hostname != "" {
				hostname = " (" + hostname + ")"
			}
			fmt.Fprintf(w, "    ✓ %s%s — %s\n", r.IP.String(), hostname, r.OSGuess)
		}
	}

	fmt.Fprintf(w, "\n")
}

// filterAlive returns only the HostResult entries where Alive is true.
func filterAlive(results []HostResult) []HostResult {
	alive := make([]HostResult, 0, len(results))
	for _, r := range results {
		if r.Alive {
			alive = append(alive, r)
		}
	}
	return alive
}

// truncate shortens a string to the given max length, appending "..." if truncated.
// The returned string will never exceed maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		// Not enough room for "..." suffix; hard-truncate.
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}