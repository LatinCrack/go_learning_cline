// Package main implements a domain blocklist for DNS query filtering.
//
// It loads blocked domains from a plain-text file (one domain per line)
// and provides O(1) lookup using a thread-safe map. When a query matches
// a blocked domain, the server can immediately respond with NXDOMAIN or
// redirect to a sinkhole address (e.g., 127.0.0.1) without contacting
// the upstream resolver.
//
// File format (one domain per line):
//
//	# Lines starting with '#' are comments
//	ads.example.com
//	malware.tracker.net
//	*.doubleclick.net    ← wildcard prefix matching
//	// empty lines are ignored
package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
)

// ────────────────────────────────────────────────────────
// Blocklist Mode
// ────────────────────────────────────────────────────────

// BlockMode defines how blocked queries are answered.
type BlockMode int

const (
	// BlockModeNXDOMAIN returns a Name Error (rcode=3) for blocked domains.
	// This tells the client the domain does not exist at all.
	BlockModeNXDOMAIN BlockMode = iota

	// BlockModeSinkhole returns a type-A record pointing to a configurable
	// IP address (default 127.0.0.1), useful for ad-blocking dashboards.
	BlockModeSinkhole
)

// ────────────────────────────────────────────────────────
// Blocklist
// ────────────────────────────────────────────────────────

// Blocklist holds a set of blocked domain names and provides fast
// thread-safe lookup for DNS query filtering.
type Blocklist struct {
	mu       sync.RWMutex
	domains  map[string]struct{} // Exact match: "ads.example.com"
	prefixes []string            // Wildcard prefixes: "doubleclick.net" for "*.doubleclick.net"

	// SinkholeIP is the IPv4 address returned when BlockMode is Sinkhole.
	SinkholeIP [4]byte
	// Mode controls the response type for blocked domains.
	Mode BlockMode
	// Enabled allows toggling the blocklist without unloading it.
	Enabled bool
}

// NewBlocklist creates an empty blocklist with default settings.
// Blocked queries will be answered with NXDOMAIN (mode=BlockModeNXDOMAIN).
func NewBlocklist() *Blocklist {
	return &Blocklist{
		domains:    make(map[string]struct{}),
		SinkholeIP: [4]byte{127, 0, 0, 1}, // Default sinkhole: localhost
		Mode:       BlockModeNXDOMAIN,
		Enabled:    true,
	}
}

// ────────────────────────────────────────────────────────
// Loading
// ────────────────────────────────────────────────────────

// LoadFromFile reads a blocklist file and populates the in-memory set.
// The file should contain one domain per line. Lines starting with '#'
// are treated as comments. Empty lines are ignored.
//
// Wildcard entries like "*.example.com" will match any subdomain of
// example.com (e.g., "sub.example.com", "a.b.example.com").
//
// Example file:
//
//	# Ad networks
//	ads.doubleclick.net
//	*.adsense.google.com
//	tracker.analytics.com
//
//	# Malware domains
//	evil.malware.net
func (b *Blocklist) LoadFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	domains := make(map[string]struct{})
	var prefixes []string

	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Normalize to lowercase for case-insensitive matching.
		line = strings.ToLower(line)

		// Wildcard prefix: "*.example.com" → store "example.com" prefix.
		if strings.HasPrefix(line, "*.") {
			prefix := line[2:] // Remove "*."
			prefix = strings.TrimSuffix(prefix, ".")
			if prefix != "" {
				prefixes = append(prefixes, prefix)
			}
			continue
		}

		// Exact domain entry.
		line = strings.TrimSuffix(line, ".")
		if line != "" {
			domains[line] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	b.mu.Lock()
	b.domains = domains
	b.prefixes = prefixes
	b.mu.Unlock()

	log.Printf("[blocklist] Loaded %d exact domains and %d wildcard prefixes from %s",
		len(domains), len(prefixes), path)
	return nil
}

// ────────────────────────────────────────────────────────
// Lookup
// ────────────────────────────────────────────────────────

// IsBlocked checks if a domain name is in the blocklist.
// The domain should be in lowercase, dot-trailing format (e.g., "ads.example.com.").
//
// Matching rules:
//   - Exact match: "ads.example.com" matches "ads.example.com"
//   - Wildcard prefix: "*.example.com" matches "sub.example.com", "a.b.example.com"
//   - Case-insensitive (domains are normalized to lowercase)
func (b *Blocklist) IsBlocked(domain string) bool {
	if !b.Enabled {
		return false
	}

	// Normalize: lowercase, remove trailing dot.
	normalized := strings.ToLower(domain)
	normalized = strings.TrimSuffix(normalized, ".")

	b.mu.RLock()
	defer b.mu.RUnlock()

	// Check exact match.
	if _, blocked := b.domains[normalized]; blocked {
		return true
	}

	// Check wildcard prefix match.
	// For domain "sub.ads.example.com", check suffixes:
	//   "sub.ads.example.com" → no match
	//   "ads.example.com"     → match against prefix "ads.example.com"
	//   "example.com"         → match against prefix "example.com"
	for _, prefix := range b.prefixes {
		if normalized == prefix || strings.HasSuffix(normalized, "."+prefix) {
			return true
		}
	}

	return false
}

// ────────────────────────────────────────────────────────
// Management
// ────────────────────────────────────────────────────────

// Add adds a single domain to the blocklist at runtime.
// Wildcard entries (e.g., "*.example.com") are stored as prefix patterns.
func (b *Blocklist) Add(domain string) {
	normalized := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))

	b.mu.Lock()
	defer b.mu.Unlock()

	// Wildcard prefix: "*.example.com" → store "example.com" prefix.
	if strings.HasPrefix(normalized, "*.") {
		prefix := normalized[2:]
		if prefix != "" {
			b.prefixes = append(b.prefixes, prefix)
		}
		return
	}

	b.domains[normalized] = struct{}{}
}

// Remove removes a single domain from the blocklist at runtime.
func (b *Blocklist) Remove(domain string) {
	normalized := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))

	b.mu.Lock()
	delete(b.domains, normalized)
	b.mu.Unlock()
}

// Size returns the total number of blocked domains (exact + wildcard).
func (b *Blocklist) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.domains) + len(b.prefixes)
}

// Clear removes all entries from the blocklist.
func (b *Blocklist) Clear() {
	b.mu.Lock()
	b.domains = make(map[string]struct{})
	b.prefixes = nil
	b.mu.Unlock()
}