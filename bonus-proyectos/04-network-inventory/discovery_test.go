package main

import (
	"bytes"
	"encoding/json"
	"net/netip"
	"strings"
	"testing"
	"time"
)

// ── CIDR Tests ─────────────────────────────────────────────────

func TestExpandCIDR_Slash24(t *testing.T) {
	ips, err := ExpandCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// /24 should yield 254 hosts (256 - network - broadcast).
	if len(ips) != 254 {
		t.Errorf("expected 254 hosts for /24, got %d", len(ips))
	}

	// First host should be 192.168.1.1.
	first := ips[0]
	expected := netip.MustParseAddr("192.168.1.1")
	if first != expected {
		t.Errorf("expected first IP %s, got %s", expected, first)
	}

	// Last host should be 192.168.1.254.
	last := ips[len(ips)-1]
	expectedLast := netip.MustParseAddr("192.168.1.254")
	if last != expectedLast {
		t.Errorf("expected last IP %s, got %s", expectedLast, last)
	}
}

func TestExpandCIDR_Slash32(t *testing.T) {
	ips, err := ExpandCIDR("10.0.0.5/32")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ips) != 1 {
		t.Errorf("expected 1 host for /32, got %d", len(ips))
	}

	expected := netip.MustParseAddr("10.0.0.5")
	if ips[0] != expected {
		t.Errorf("expected %s, got %s", expected, ips[0])
	}
}

func TestExpandCIDR_Slash31(t *testing.T) {
	ips, err := ExpandCIDR("10.0.0.0/31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// /31 should yield 2 addresses (RFC 3021 point-to-point).
	if len(ips) != 2 {
		t.Errorf("expected 2 hosts for /31, got %d", len(ips))
	}
}

func TestExpandCIDR_Slash30(t *testing.T) {
	ips, err := ExpandCIDR("10.0.0.0/30")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// /30 = 4 addresses, minus network and broadcast = 2 usable hosts.
	if len(ips) != 2 {
		t.Errorf("expected 2 hosts for /30, got %d", len(ips))
	}
}

func TestExpandCIDR_InvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"no prefix length", "192.168.1.0"},
		{"invalid IP", "999.999.999.999/24"},
		{"garbage", "not-a-cidr"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExpandCIDR(tt.input)
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tt.input)
			}
		})
	}
}

func TestExpandCIDR_TooLarge(t *testing.T) {
	// /8 expands to 16M+ addresses, should be rejected.
	_, err := ExpandCIDR("10.0.0.0/8")
	if err == nil {
		t.Error("expected error for /8 range exceeding max addresses")
	}
}

// ── ValidateCIDR Tests ─────────────────────────────────────────

func TestValidateCIDR_Valid(t *testing.T) {
	tests := []string{
		"192.168.1.0/24",
		"10.0.0.0/8",
		"172.16.0.0/16",
		"10.0.0.1/32",
	}
	for _, cidr := range tests {
		if err := ValidateCIDR(cidr); err != nil {
			t.Errorf("expected valid CIDR %q, got error: %v", cidr, err)
		}
	}
}

func TestValidateCIDR_Invalid(t *testing.T) {
	tests := []string{
		"",
		"192.168.1.0",
		"not-valid",
		"192.168.1.0/33",
	}
	for _, cidr := range tests {
		if err := ValidateCIDR(cidr); err == nil {
			t.Errorf("expected invalid CIDR %q to return error, got nil", cidr)
		}
	}
}

// ── TTL Parsing Tests ──────────────────────────────────────────

func TestParseTTL_Windows(t *testing.T) {
	output := "Reply from 192.168.1.1: bytes=32 time<1ms TTL=128"
	ttl := parseTTL(output)
	if ttl != 128 {
		t.Errorf("expected TTL 128, got %d", ttl)
	}
}

func TestParseTTL_Linux(t *testing.T) {
	output := "64 bytes from 192.168.1.1: icmp_seq=1 ttl=64 time=0.5 ms"
	ttl := parseTTL(output)
	if ttl != 64 {
		t.Errorf("expected TTL 64, got %d", ttl)
	}
}

func TestParseTTL_NoTTL(t *testing.T) {
	output := "Request timed out."
	ttl := parseTTL(output)
	if ttl != 0 {
		t.Errorf("expected TTL 0, got %d", ttl)
	}
}

func TestParseTTL_CaseInsensitive(t *testing.T) {
	output := "Reply from 10.0.0.1: bytes=32 time<1ms Ttl=255"
	ttl := parseTTL(output)
	if ttl != 255 {
		t.Errorf("expected TTL 255, got %d", ttl)
	}
}

// ── OS Guessing Tests ──────────────────────────────────────────

func TestGuessOSFromTTL(t *testing.T) {
	tests := []struct {
		ttl      int
		expected string
	}{
		{0, "Unknown"},
		{1, "Linux/Unix"},
		{64, "Linux/Unix"},
		{63, "Linux/Unix"},
		{128, "Windows"},
		{100, "Windows"},
		{255, "Network Device/Solaris"},
		{200, "Network Device/Solaris"},
		{-1, "Unknown"},
	}

	for _, tt := range tests {
		result := guessOSFromTTL(tt.ttl)
		if result != tt.expected {
			t.Errorf("TTL %d: expected %q, got %q", tt.ttl, tt.expected, result)
		}
	}
}

// ── Exporter Tests ──────────────────────────────────────────────

func TestBuildReport(t *testing.T) {
	results := []HostResult{
		{
			IP:       netip.MustParseAddr("192.168.1.1"),
			Alive:    true,
			Latency:  5 * time.Millisecond,
			Hostname: "gateway.lan",
			OSGuess:  "Linux/Unix",
			TTL:      64,
		},
		{
			IP:    netip.MustParseAddr("192.168.1.2"),
			Alive: false,
		},
		{
			IP:       netip.MustParseAddr("192.168.1.10"),
			Alive:    true,
			Latency:  12 * time.Millisecond,
			Hostname: "desktop.lan",
			OSGuess:  "Windows",
			TTL:      128,
		},
	}

	report := buildReport(results, "192.168.1.0/24", 2*time.Second)

	if report.TotalHosts != 3 {
		t.Errorf("expected 3 total hosts, got %d", report.TotalHosts)
	}
	if report.AliveHosts != 2 {
		t.Errorf("expected 2 alive hosts, got %d", report.AliveHosts)
	}
	if report.Range != "192.168.1.0/24" {
		t.Errorf("expected range '192.168.1.0/24', got %q", report.Range)
	}
}

func TestExportJSON_ValidOutput(t *testing.T) {
	results := []HostResult{
		{
			IP:       netip.MustParseAddr("10.0.0.1"),
			Alive:    true,
			Latency:  3 * time.Millisecond,
			Hostname: "router.local",
			OSGuess:  "Linux/Unix",
			TTL:      64,
		},
	}

	var buf bytes.Buffer
	err := ExportJSON(&buf, results, "10.0.0.0/30", 1*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the output is valid JSON.
	var report InventoryReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if report.TotalHosts != 1 {
		t.Errorf("expected 1 total host, got %d", report.TotalHosts)
	}
	if report.AliveHosts != 1 {
		t.Errorf("expected 1 alive host, got %d", report.AliveHosts)
	}
}

// ── FilterAlive Tests ──────────────────────────────────────────

func TestFilterAlive(t *testing.T) {
	results := []HostResult{
		{IP: netip.MustParseAddr("10.0.0.1"), Alive: true},
		{IP: netip.MustParseAddr("10.0.0.2"), Alive: false},
		{IP: netip.MustParseAddr("10.0.0.3"), Alive: true},
		{IP: netip.MustParseAddr("10.0.0.4"), Alive: false},
	}

	alive := filterAlive(results)
	if len(alive) != 2 {
		t.Errorf("expected 2 alive hosts, got %d", len(alive))
	}
}

func TestFilterAlive_Empty(t *testing.T) {
	results := []HostResult{}
	alive := filterAlive(results)
	if len(alive) != 0 {
		t.Errorf("expected 0 alive hosts, got %d", len(alive))
	}
}

// ── Truncate Tests ─────────────────────────────────────────────

func TestTruncate(t *testing.T) {
	t.Run("short string not truncated", func(t *testing.T) {
		result := truncate("hello", 10)
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})

	t.Run("exact length not truncated", func(t *testing.T) {
		result := truncate("hello", 5)
		if result != "hello" {
			t.Errorf("expected 'hello', got %q", result)
		}
	})

	t.Run("truncated string respects maxLen", func(t *testing.T) {
		result := truncate("hello world", 8)
		if len(result) > 8 {
			t.Errorf("result length %d exceeds maxLen 8: %q", len(result), result)
		}
		// With maxLen=8, we get s[:5]+"..." = "hello..." (8 chars).
		if !strings.HasPrefix(result, "hello") {
			t.Errorf("expected prefix 'hello', got %q", result)
		}
		if !strings.HasSuffix(result, "...") {
			t.Errorf("expected suffix '...', got %q", result)
		}
	})

	t.Run("small maxLen with truncation", func(t *testing.T) {
		result := truncate("hi", 1)
		if len(result) > 1 {
			t.Errorf("result length %d exceeds maxLen 1: %q", len(result), result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result := truncate("", 5)
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})

	t.Run("zero maxLen", func(t *testing.T) {
		result := truncate("test", 0)
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})
}

// ── PrintTable Tests ───────────────────────────────────────────

func TestPrintTable_WithAliveHosts(t *testing.T) {
	results := []HostResult{
		{
			IP:      netip.MustParseAddr("192.168.1.1"),
			Alive:   true,
			Latency: 5 * time.Millisecond,
			OSGuess: "Linux/Unix",
			TTL:     64,
		},
	}

	var buf bytes.Buffer
	PrintTable(&buf, results, "192.168.1.0/24", 1*time.Second)

	output := buf.String()
	if !strings.Contains(output, "192.168.1.1") {
		t.Error("table output should contain the alive host IP")
	}
	if !strings.Contains(output, "Linux/Unix") {
		t.Error("table output should contain the OS guess")
	}
	if !strings.Contains(output, "NETWORK INVENTORY REPORT") {
		t.Error("table output should contain the report header")
	}
}

func TestPrintTable_NoAliveHosts(t *testing.T) {
	results := []HostResult{
		{IP: netip.MustParseAddr("192.168.1.1"), Alive: false},
	}

	var buf bytes.Buffer
	PrintTable(&buf, results, "192.168.1.0/24", 1*time.Second)

	output := buf.String()
	if !strings.Contains(output, "No active hosts found") {
		t.Error("table output should indicate no active hosts")
	}
}

// ── PrintSummary Tests ─────────────────────────────────────────

func TestPrintSummary(t *testing.T) {
	results := []HostResult{
		{
			IP:       netip.MustParseAddr("10.0.0.1"),
			Alive:    true,
			Hostname: "server.local",
			OSGuess:  "Windows",
		},
		{IP: netip.MustParseAddr("10.0.0.2"), Alive: false},
	}

	var buf bytes.Buffer
	PrintSummary(&buf, results, "10.0.0.0/24", 500*time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "Alive:") {
		t.Error("summary should contain alive count")
	}
	if !strings.Contains(output, "server.local") {
		t.Error("summary should contain the resolved hostname")
	}
	if !strings.Contains(output, "10.0.0.1") {
		t.Error("summary should contain the alive host IP")
	}
}

// ── Unit Conversion Tests ──────────────────────────────────────

func TestAddrUint32RoundTrip(t *testing.T) {
	tests := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.100",
		"255.255.255.255",
		"0.0.0.0",
	}

	for _, ip := range tests {
		addr := netip.MustParseAddr(ip)
		n := addrToUint32(addr)
		result := uint32ToAddr(n)
		if result != addr {
			t.Errorf("round trip failed for %s: got %s (uint32=%d)", ip, result, n)
		}
	}
}
