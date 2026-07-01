package main

import (
	"context"
	"fmt"
	"net/netip"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// HostResult represents the outcome of a single host discovery probe.
type HostResult struct {
	IP       netip.Addr
	Alive    bool
	Latency  time.Duration
	Hostname string
	OSGuess  string
	TTL      int
}

// DiscoveryConfig holds the configuration for a network discovery scan.
type DiscoveryConfig struct {
	CIDR       string
	Timeout    time.Duration
	MaxWorkers int
}

// DiscoverHosts performs a concurrent ping sweep of all hosts in the given CIDR range.
// It uses a bounded worker pool (controlled by a semaphore channel) to limit concurrency,
// and sync.WaitGroup to synchronize completion of all probes.
//
// Results are collected via a channel and returned as a slice. The enrichment step
// (DNS reverse lookup and OS detection) is performed inline for each alive host
// to maximize throughput.
func DiscoverHosts(ctx context.Context, cfg DiscoveryConfig) ([]HostResult, error) {
	// Expand CIDR to individual IP addresses.
	ips, err := ExpandCIDR(cfg.CIDR)
	if err != nil {
		return nil, fmt.Errorf("failed to expand CIDR: %w", err)
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("no hosts to scan in range %s", cfg.CIDR)
	}

	// Semaphore channel to limit concurrent workers.
	sem := make(chan struct{}, cfg.MaxWorkers)
	// Results channel to collect discovery outcomes.
	resultsCh := make(chan HostResult, len(ips))

	var wg sync.WaitGroup

	for _, ip := range ips {
		wg.Add(1)
		go func(addr netip.Addr) {
			defer wg.Done()

			// Acquire semaphore slot (blocks if pool is full).
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }() // Release slot when done.
			case <-ctx.Done():
				return // Context cancelled; exit early.
			}

			result := probeHost(ctx, addr, cfg.Timeout)
			resultsCh <- result
		}(ip)
	}

	// Close results channel once all workers complete.
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect all results.
	var results []HostResult
	for r := range resultsCh {
		results = append(results, r)
	}

	return results, nil
}

// probeHost sends a single ICMP ping to the given address and enriches the result
// with DNS reverse lookup and OS guessing if the host responds.
func probeHost(ctx context.Context, addr netip.Addr, timeout time.Duration) HostResult {
	result := HostResult{
		IP:    addr,
		Alive: false,
	}

	// Execute platform-specific ping command.
	alive, latency, ttl := pingHost(ctx, addr.String(), timeout)
	result.Alive = alive
	result.Latency = latency
	result.TTL = ttl

	// Enrich alive hosts with DNS and OS information.
	if alive {
		result.Hostname = reverseLookup(addr.String())
		result.OSGuess = guessOSFromTTL(ttl)
	}

	return result
}

// pingHost executes the system's native ping command via os/exec for cross-platform
// compatibility (avoids raw socket / ICMP privilege issues).
//
// On Windows it uses: ping -n 1 -w <timeout_ms> <ip>
// On Linux/macOS it uses: ping -c 1 -W <timeout_secs> <ip>
//
// Returns: alive (bool), latency (time.Duration), ttl (int).
func pingHost(ctx context.Context, ip string, timeout time.Duration) (bool, time.Duration, int) {
	var cmd *exec.Cmd

	timeoutMs := timeout.Milliseconds()
	timeoutSecs := int(timeout.Seconds())
	if timeoutSecs < 1 {
		timeoutSecs = 1
	}

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", fmt.Sprintf("%d", timeoutMs), ip)
	} else {
		cmd = exec.CommandContext(ctx, "ping", "-c", "1", "-W", fmt.Sprintf("%d", timeoutSecs), ip)
	}

	start := time.Now()
	output, err := cmd.CombinedOutput()
	elapsed := time.Since(start)

	if err != nil {
		return false, 0, 0
	}

	outputStr := string(output)

	// Parse TTL from ping output.
	ttl := parseTTL(outputStr)

	// If we got output without error, the host is alive.
	return true, elapsed, ttl
}

// parseTTL extracts the TTL value from the ping command output.
// Handles both Windows and Linux/macOS output formats:
//
//	Windows: "TTL=64" or "TTL=128"
//	Linux:   "ttl=64" or "ttl=128"
func parseTTL(output string) int {
	lower := strings.ToLower(output)

	// Look for "ttl=" or "ttl=" patterns in the output.
	for _, line := range strings.Split(lower, "\n") {
		if idx := strings.Index(line, "ttl="); idx != -1 {
			remainder := line[idx+4:]
			// Extract the numeric value after "ttl=".
			var ttl int
			_, err := fmt.Sscanf(remainder, "%d", &ttl)
			if err == nil && ttl > 0 && ttl <= 255 {
				return ttl
			}
		}
	}

	return 0
}

// guessOSFromTTL provides a rough OS estimation based on the TTL (Time To Live)
// value observed in the ICMP response. This is a heuristic — not definitive —
// because TTL values can be modified by intermediate routers and administrators.
//
// Common default TTL values:
//
//	Linux/Unix:  64
//	Windows:     128
//	Cisco/Solaris: 255
//	macOS:       64
func guessOSFromTTL(ttl int) string {
	if ttl <= 0 {
		return "Unknown"
	}

	switch {
	case ttl <= 64:
		// Could be Linux, macOS, or another Unix-like system (default TTL 64,
		// decremented by each hop).
		return "Linux/Unix"
	case ttl <= 128:
		// Windows systems typically start with TTL 128.
		return "Windows"
	case ttl <= 255:
		// Network devices (Cisco routers, Solaris, etc.) often use TTL 255.
		return "Network Device/Solaris"
	default:
		return "Unknown"
	}
}