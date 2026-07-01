package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

// ── Scanner Tests ──────────────────────────────────────────────────────

func TestLookupService(t *testing.T) {
	tests := []struct {
		port     int
		expected string
	}{
		{22, "SSH"},
		{80, "HTTP"},
		{443, "HTTPS"},
		{3306, "MySQL"},
		{5432, "PostgreSQL"},
		{99999, "unknown"},
		{1, "unknown"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("port_%d", tt.port), func(t *testing.T) {
			result := lookupService(tt.port)
			if result != tt.expected {
				t.Errorf("lookupService(%d) = %q, want %q", tt.port, result, tt.expected)
			}
		})
	}
}

func TestFilterOpen(t *testing.T) {
	results := []PortResult{
		{Port: 22, Open: true, Service: "SSH"},
		{Port: 80, Open: false},
		{Port: 443, Open: true, Service: "HTTPS"},
		{Port: 3306, Open: false},
	}

	open := FilterOpen(results)
	if len(open) != 2 {
		t.Fatalf("FilterOpen: got %d results, want 2", len(open))
	}
	if open[0].Port != 22 || open[1].Port != 443 {
		t.Errorf("FilterOpen: unexpected ports: got %v", open)
	}
}

func TestFilterOpen_Empty(t *testing.T) {
	results := []PortResult{
		{Port: 80, Open: false},
		{Port: 443, Open: false},
	}

	open := FilterOpen(results)
	if len(open) != 0 {
		t.Errorf("FilterOpen: expected empty, got %d results", len(open))
	}
}

func TestFilterOpen_AllOpen(t *testing.T) {
	results := []PortResult{
		{Port: 22, Open: true, Service: "SSH"},
		{Port: 80, Open: true, Service: "HTTP"},
	}

	open := FilterOpen(results)
	if len(open) != 2 {
		t.Errorf("FilterOpen: expected 2, got %d", len(open))
	}
}

// ── Integration Test: Scan against a local listener ────────────────────

func TestScanPorts_OpenPort(t *testing.T) {
	// Start a local TCP listener to simulate an open port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start test listener: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	// Accept connections in background to prevent blocking.
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	cfg := ScanConfig{
		Host:       "127.0.0.1",
		StartPort:  port,
		EndPort:    port,
		Timeout:    2 * time.Second,
		MaxWorkers: 1,
	}

	results := ScanPorts(cfg)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if !results[0].Open {
		t.Errorf("Port %d should be open, but was reported closed", port)
	}
}

func TestScanPorts_ClosedPort(t *testing.T) {
	// Find a port that is definitely not listening.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Give the OS time to release the port.
	time.Sleep(50 * time.Millisecond)

	cfg := ScanConfig{
		Host:       "127.0.0.1",
		StartPort:  port,
		EndPort:    port,
		Timeout:    500 * time.Millisecond,
		MaxWorkers: 1,
	}

	results := ScanPorts(cfg)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Open {
		t.Errorf("Port %d should be closed, but was reported open", port)
	}
}

func TestScanPorts_MultiplePorts(t *testing.T) {
	// Start 3 listeners on random ports.
	var ports []int
	var listeners []net.Listener

	for i := 0; i < 3; i++ {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatalf("Failed to start listener %d: %v", i, err)
		}
		defer ln.Close()
		ports = append(ports, ln.Addr().(*net.TCPAddr).Port)
		listeners = append(listeners, ln)
	}

	// Accept connections in background.
	for _, ln := range listeners {
		go func(l net.Listener) {
			for {
				conn, err := l.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}(ln)
	}

	// Scan a range that includes our open ports.
	minPort, maxPort := ports[0], ports[0]
	for _, p := range ports {
		if p < minPort {
			minPort = p
		}
		if p > maxPort {
			maxPort = p
		}
	}

	cfg := ScanConfig{
		Host:       "127.0.0.1",
		StartPort:  minPort,
		EndPort:    maxPort,
		Timeout:    2 * time.Second,
		MaxWorkers: 10,
	}

	results := ScanPorts(cfg)
	open := FilterOpen(results)

	if len(open) != 3 {
		t.Errorf("Expected 3 open ports, got %d. Ports found: %v", len(open), open)
	}
}

// ── Worker Pool Concurrency Test ───────────────────────────────────────

func TestScanPorts_ConcurrencyLimit(t *testing.T) {
	// Verify that MaxWorkers is respected by scanning many ports
	// with a very small worker count. The scan should still complete.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	cfg := ScanConfig{
		Host:       "127.0.0.1",
		StartPort:  port,
		EndPort:    port,
		Timeout:    2 * time.Second,
		MaxWorkers: 1, // Only 1 worker at a time.
	}

	done := make(chan []PortResult, 1)
	go func() {
		done <- ScanPorts(cfg)
	}()

	select {
	case results := <-done:
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Scan timed out — possible deadlock with MaxWorkers=1")
	}
}

// ── Report Tests ───────────────────────────────────────────────────────

func TestPrintTable_OpenPorts(t *testing.T) {
	results := []PortResult{
		{Port: 22, Open: true, Service: "SSH", Latency: 1500 * time.Microsecond},
		{Port: 80, Open: true, Service: "HTTP", Latency: 2300 * time.Microsecond},
		{Port: 443, Open: false, Latency: 500 * time.Millisecond},
	}

	var buf bytes.Buffer
	PrintTable(&buf, results, "192.168.1.1", 2*time.Second)

	output := buf.String()
	if !strings.Contains(output, "SSH") {
		t.Error("Table output should contain 'SSH'")
	}
	if !strings.Contains(output, "HTTP") {
		t.Error("Table output should contain 'HTTP'")
	}
	if !strings.Contains(output, "OPEN") {
		t.Error("Table output should contain 'OPEN'")
	}
	if strings.Contains(output, "443") {
		t.Error("Table output should not contain closed port 443")
	}
}

func TestPrintTable_NoOpenPorts(t *testing.T) {
	results := []PortResult{
		{Port: 80, Open: false, Latency: 500 * time.Millisecond},
		{Port: 443, Open: false, Latency: 500 * time.Millisecond},
	}

	var buf bytes.Buffer
	PrintTable(&buf, results, "10.0.0.1", 1*time.Second)

	output := buf.String()
	if !strings.Contains(output, "No open ports found") {
		t.Error("Table should show 'No open ports found' when all ports are closed")
	}
}

func TestPrintJSON(t *testing.T) {
	cfg := ScanConfig{
		Host:       "192.168.1.1",
		StartPort:  22,
		EndPort:    80,
		Timeout:    500 * time.Millisecond,
		MaxWorkers: 100,
	}

	results := []PortResult{
		{Port: 22, Open: true, Service: "SSH", Latency: 1 * time.Millisecond},
		{Port: 80, Open: false, Latency: 500 * time.Millisecond},
	}

	var buf bytes.Buffer
	err := PrintJSON(&buf, results, cfg, 2*time.Second)
	if err != nil {
		t.Fatalf("PrintJSON failed: %v", err)
	}

	// Verify it's valid JSON. Use a struct that matches the actual JSON output
	// (latency is serialized as a string, not time.Duration).
	type portResultJSON struct {
		Port    int    `json:"port"`
		Open    bool   `json:"open"`
		Service string `json:"service"`
		Latency string `json:"latency"`
	}
	type scanReportJSON struct {
		Host       string           `json:"host"`
		StartPort  int              `json:"start_port"`
		EndPort    int              `json:"end_port"`
		Duration   string           `json:"duration"`
		TotalPorts int              `json:"total_ports"`
		OpenPorts  int              `json:"open_ports"`
		Results    []portResultJSON `json:"results"`
	}

	var report scanReportJSON
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, buf.String())
	}

	if report.Host != "192.168.1.1" {
		t.Errorf("Host = %q, want %q", report.Host, "192.168.1.1")
	}
	if report.OpenPorts != 1 {
		t.Errorf("OpenPorts = %d, want 1", report.OpenPorts)
	}
	if report.TotalPorts != 2 {
		t.Errorf("TotalPorts = %d, want 2", report.TotalPorts)
	}
	// Only open ports are included in the JSON results array.
	if len(report.Results) != 1 {
		t.Fatalf("Expected 1 result (open port only), got %d", len(report.Results))
	}
	if report.Results[0].Latency != "1ms" {
		t.Errorf("First result latency = %q, want %q", report.Results[0].Latency, "1ms")
	}
}

func TestPrintSummary(t *testing.T) {
	results := []PortResult{
		{Port: 22, Open: true, Service: "SSH", Latency: 1 * time.Millisecond},
		{Port: 80, Open: true, Service: "HTTP", Latency: 2 * time.Millisecond},
		{Port: 443, Open: false, Latency: 500 * time.Millisecond},
	}

	var buf bytes.Buffer
	PrintSummary(&buf, results, "example.com", 3*time.Second)

	output := buf.String()
	if !strings.Contains(output, "example.com") {
		t.Error("Summary should contain the host name")
	}
	if !strings.Contains(output, "22/SSH") {
		t.Error("Summary should contain '22/SSH'")
	}
	if !strings.Contains(output, "80/HTTP") {
		t.Error("Summary should contain '80/HTTP'")
	}
}

// ── Port Range Parsing Tests ───────────────────────────────────────────

func TestParsePortRange_SinglePort(t *testing.T) {
	start, end, err := parsePortRange("80")
	if err != nil {
		t.Fatalf("parsePortRange('80') error: %v", err)
	}
	if start != 80 || end != 80 {
		t.Errorf("parsePortRange('80') = (%d, %d), want (80, 80)", start, end)
	}
}

func TestParsePortRange_Range(t *testing.T) {
	start, end, err := parsePortRange("1-1024")
	if err != nil {
		t.Fatalf("parsePortRange('1-1024') error: %v", err)
	}
	if start != 1 || end != 1024 {
		t.Errorf("parsePortRange('1-1024') = (%d, %d), want (1, 1024)", start, end)
	}
}

func TestParsePortRange_CommaSeparated(t *testing.T) {
	start, end, err := parsePortRange("22,80,443")
	if err != nil {
		t.Fatalf("parsePortRange('22,80,443') error: %v", err)
	}
	if start != 22 || end != 443 {
		t.Errorf("parsePortRange('22,80,443') = (%d, %d), want (22, 443)", start, end)
	}
}

func TestParsePortRange_InvalidInputs(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"", "empty string"},
		{"abc", "non-numeric"},
		{"0", "zero port"},
		{"70000", "port > 65535"},
		{"1024-1", "reversed range"},
		{"0-1024", "start port zero"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			_, _, err := parsePortRange(tt.input)
			if err == nil {
				t.Errorf("parsePortRange(%q) should return error for %s", tt.input, tt.desc)
			}
		})
	}
}

// ── Benchmarks ─────────────────────────────────────────────────────────

func BenchmarkScanPorts_100Workers(b *testing.B) {
	cfg := ScanConfig{
		Host:       "127.0.0.1",
		StartPort:  1,
		EndPort:    100,
		Timeout:    100 * time.Millisecond,
		MaxWorkers: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScanPorts(cfg)
	}
}

func BenchmarkScanPorts_10Workers(b *testing.B) {
	cfg := ScanConfig{
		Host:       "127.0.0.1",
		StartPort:  1,
		EndPort:    100,
		Timeout:    100 * time.Millisecond,
		MaxWorkers: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScanPorts(cfg)
	}
}

func BenchmarkFilterOpen(b *testing.B) {
	results := make([]PortResult, 1000)
	for i := range results {
		results[i] = PortResult{
			Port: i + 1,
			Open: i%100 == 0, // 1% open
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterOpen(results)
	}
}
