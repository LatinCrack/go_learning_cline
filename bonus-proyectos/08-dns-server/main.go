// DNS Server Ligero con Cache y Filtros
//
// A lightweight, high-performance DNS proxy server written in pure Go.
// It intercepts DNS queries, applies domain blocklists (anti-ad/malware),
// serves cached responses with strict TTL enforcement, and forwards
// legitimate queries to configurable upstream resolvers.
//
// Usage:
//
//	go run . --port 5353 --upstream 8.8.8.8,1.1.1.1 --blocklist blocklist.txt
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	defaultPort         = 5353
	defaultUpstream     = "8.8.8.8,1.1.1.1"
	defaultTTL          = 300 // 5 minutes fallback TTL
	defaultCacheCleanup = 60  // seconds between cache cleanup runs
)

// ────────────────────────────────────────────────────────
// Banner
// ────────────────────────────────────────────────────────

const banner = `
╔══════════════════════════════════════════════════════╗
║           DNS Server Ligero con Cache                ║
║           Cache + Blocklist + Upstream                ║
╚══════════════════════════════════════════════════════╝
`

// ────────────────────────────────────────────────────────
// CLI Flags
// ────────────────────────────────────────────────────────

var (
	port         int
	upstream     string
	blocklistFile string
	sinkholeIP   string
	blockMode    string
	ttl          int
	cacheCleanup int
	verbose      bool
)

func parseFlags() {
	flag.IntVar(&port, "port", defaultPort, "UDP port to listen on (use 53 for production, high port for testing)")
	flag.IntVar(&port, "p", defaultPort, "Shorthand for --port")

	flag.StringVar(&upstream, "upstream", defaultUpstream, "Comma-separated upstream DNS servers (e.g., 8.8.8.8,1.1.1.1)")
	flag.StringVar(&upstream, "u", defaultUpstream, "Shorthand for --upstream")

	flag.StringVar(&blocklistFile, "blocklist", "", "Path to blocklist file (one domain per line)")
	flag.StringVar(&blocklistFile, "b", "", "Shorthand for --blocklist")

	flag.StringVar(&sinkholeIP, "sinkhole", "127.0.0.1", "IP address for sinkhole responses (used with --mode sinkhole)")
	flag.StringVar(&blockMode, "mode", "nxdomain", "Block response mode: 'nxdomain' or 'sinkhole'")
	flag.IntVar(&ttl, "ttl", defaultTTL, "Default TTL in seconds for cached responses without TTL")
	flag.IntVar(&cacheCleanup, "cache-cleanup", defaultCacheCleanup, "Cache cleanup interval in seconds")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&verbose, "v", false, "Shorthand for --verbose")

	flag.Usage = printUsage
	flag.Parse()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "%s\n", banner)
	fmt.Fprintf(os.Stderr, "Usage: dns-server [options]\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  --port, -p <int>         UDP port to listen on (default: %d)\n", defaultPort)
	fmt.Fprintf(os.Stderr, "  --upstream, -u <servers>  Comma-separated upstream DNS servers (default: %s)\n", defaultUpstream)
	fmt.Fprintf(os.Stderr, "  --blocklist, -b <path>    Path to blocklist file (one domain per line)\n")
	fmt.Fprintf(os.Stderr, "  --sinkhole <ip>           Sinkhole IP address (default: 127.0.0.1)\n")
	fmt.Fprintf(os.Stderr, "  --mode <mode>             Block mode: 'nxdomain' or 'sinkhole' (default: nxdomain)\n")
	fmt.Fprintf(os.Stderr, "  --ttl <int>               Default cache TTL in seconds (default: %d)\n", defaultTTL)
	fmt.Fprintf(os.Stderr, "  --cache-cleanup <int>     Cache cleanup interval in seconds (default: %d)\n", defaultCacheCleanup)
	fmt.Fprintf(os.Stderr, "  --verbose, -v             Enable verbose logging\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  # Basic usage on high port (no root needed)\n")
	fmt.Fprintf(os.Stderr, "  go run . --port 5353 --upstream 8.8.8.8\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  # With blocklist\n")
	fmt.Fprintf(os.Stderr, "  go run . --port 5353 --blocklist blocklist.txt\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  # Production on port 53 (requires root/admin)\n")
	fmt.Fprintf(os.Stderr, "  sudo go run . --port 53 --upstream 8.8.8.8,1.1.1.1 --blocklist blocklist.txt\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  # Test with dig:\n")
	fmt.Fprintf(os.Stderr, "  dig @127.0.0.1 -p 5353 google.com A\n")
}

// ────────────────────────────────────────────────────────
// Main
// ────────────────────────────────────────────────────────

func main() {
	parseFlags()

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	fmt.Print(banner)

	// ── Parse upstream servers ──
	upstreams := parseUpstreamList(upstream)

	// ── Parse sinkhole IP ──
	sinkIP, err := parseIPv4(sinkholeIP)
	if err != nil {
		log.Fatalf("Invalid sinkhole IP %q: %v", sinkholeIP, err)
	}

	// ── Create blocklist ──
	bl := NewBlocklist()
	bl.SinkholeIP = sinkIP

	switch strings.ToLower(blockMode) {
	case "sinkhole":
		bl.Mode = BlockModeSinkhole
	default:
		bl.Mode = BlockModeNXDOMAIN
	}

	if blocklistFile != "" {
		if err := bl.LoadFromFile(blocklistFile); err != nil {
			log.Fatalf("Failed to load blocklist from %q: %v", blocklistFile, err)
		}
	}

	// ── Create cache ──
	cache := NewDNSCache(time.Duration(cacheCleanup) * time.Second)
	defer cache.Stop()

	// ── Create resolver ──
	resolver := NewResolver(upstreams, 5*time.Second)

	// ── Create and start server ──
	addr := fmt.Sprintf(":%d", port)
	server := NewDNSServer(addr, resolver, cache, bl, uint32(ttl))

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start DNS server: %v", err)
	}

	// ── Wait for shutdown signal ──
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("[main] Received signal: %v", sig)

	server.Stop()

	// Print final metrics.
	m := server.Metrics()
	log.Printf("[main] Final metrics:")
	log.Printf("  Queries processed : %d", m.Queries)
	log.Printf("  Queries blocked   : %d", m.Blocked)
	log.Printf("  Cache hits        : %d", m.CacheHits)
	log.Printf("  Errors            : %d", m.Errors)
	log.Printf("  Cache hit rate    : %.1f%%", m.HitRate)
}

// ────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────

// parseUpstreamList splits a comma-separated upstream string and
// appends ":53" to entries that don't have a port.
func parseUpstreamList(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Append default port if not specified.
		if !strings.Contains(p, ":") {
			p = p + ":53"
		}
		result = append(result, p)
	}
	return result
}

// parseIPv4 parses a dotted-quad IPv4 address into a [4]byte.
func parseIPv4(s string) ([4]byte, error) {
	var ip [4]byte
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return ip, fmt.Errorf("invalid IPv4 address: %s", s)
	}
	for i, p := range parts {
		var n int
		_, err := fmt.Sscanf(p, "%d", &n)
		if err != nil || n < 0 || n > 255 {
			return ip, fmt.Errorf("invalid octet %q in IPv4 address: %s", p, s)
		}
		ip[i] = byte(n)
	}
	return ip, nil
}