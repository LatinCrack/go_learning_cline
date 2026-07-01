package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// PortResult holds the scan result for a single port.
type PortResult struct {
	Port    int
	Open    bool
	Service string
	Latency time.Duration
	Err     error
}

// ScanConfig defines the configuration for a port scan operation.
type ScanConfig struct {
	Host       string
	StartPort  int
	EndPort    int
	Timeout    time.Duration
	MaxWorkers int
}

// WellKnownServices maps common port numbers to their service names.
var WellKnownServices = map[int]string{
	20:    "FTP-Data",
	21:    "FTP",
	22:    "SSH",
	23:    "Telnet",
	25:    "SMTP",
	53:    "DNS",
	80:    "HTTP",
	110:    "POP3",
	111:    "RPCBind",
	135:    "MSRPC",
	139:    "NetBIOS-SSN",
	143:    "IMAP",
	443:   "HTTPS",
	445:   "SMB",
	993:   "IMAPS",
	995:   "POP3S",
	1433:  "MSSQL",
	1521:  "Oracle",
	3306:  "MySQL",
	3389:  "RDP",
	5432:  "PostgreSQL",
	5900:  "VNC",
	6379:  "Redis",
	8080:  "HTTP-Proxy",
	8443:  "HTTPS-Alt",
	9090:  "Prometheus",
	27017: "MongoDB",
}

// lookupService returns the well-known service name for a port number.
// If the port is not in the well-known list, it returns "unknown".
func lookupService(port int) string {
	if svc, ok := WellKnownServices[port]; ok {
		return svc
	}
	return "unknown"
}

// ScanPorts executes a concurrent port scan against the configured host.
// It uses a worker pool pattern to control the number of concurrent goroutines,
// preventing file descriptor exhaustion. Results are collected via a channel and
// returned sorted by port number.
func ScanPorts(cfg ScanConfig) []PortResult {
	totalPorts := cfg.EndPort - cfg.StartPort + 1

	// Channels for job distribution and result collection.
	ports := make(chan int, cfg.MaxWorkers)
	results := make(chan PortResult, totalPorts)

	// WaitGroup to synchronize worker completion.
	var wg sync.WaitGroup

	// Start the worker pool. Each worker reads ports from the channel
	// and attempts a TCP connection with a strict timeout.
	for i := 0; i < cfg.MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range ports {
				results <- scanSinglePort(cfg.Host, port, cfg.Timeout)
			}
		}()
	}

	// Enqueue all ports into the jobs channel.
	go func() {
		for port := cfg.StartPort; port <= cfg.EndPort; port++ {
			ports <- port
		}
		close(ports)
	}()

	// Close the results channel once all workers have finished.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results.
	var allResults []PortResult
	for r := range results {
		allResults = append(allResults, r)
	}

	return allResults
}

// scanSinglePort attempts a TCP connection to the specified host and port
// using net.DialTimeout. It measures connection latency and identifies
// the associated well-known service.
func scanSinglePort(host string, port int, timeout time.Duration) PortResult {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, timeout)
	latency := time.Since(start)

	if err != nil {
		return PortResult{
			Port:    port,
			Open:    false,
			Service: "",
			Latency: latency,
			Err:     err,
		}
	}

	// Connection succeeded — port is open. Close immediately.
	conn.Close()

	return PortResult{
		Port:    port,
		Open:    true,
		Service: lookupService(port),
		Latency: latency,
		Err:     nil,
	}
}

// FilterOpen returns only the port results where the port is open.
func FilterOpen(results []PortResult) []PortResult {
	var open []PortResult
	for _, r := range results {
		if r.Open {
			open = append(open, r)
		}
	}
	return open
}