package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// InterfaceMetrics holds network counters for a single interface.
type InterfaceMetrics struct {
	Name         string `json:"name"`
	RxBytes      uint64 `json:"rx_bytes"`
	TxBytes      uint64 `json:"tx_bytes"`
	RxPackets    uint64 `json:"rx_packets"`
	TxPackets    uint64 `json:"tx_packets"`
	RxErrors     uint64 `json:"rx_errors"`
	TxErrors     uint64 `json:"tx_errors"`
	RxDropped    uint64 `json:"rx_dropped"`
	TxDropped    uint64 `json:"tx_dropped"`
	RxRateBytesS uint64 `json:"rx_rate_bytes_s"`
	TxRateBytesS uint64 `json:"tx_rate_bytes_s"`
}

// NetworkMetrics holds current network usage data for all interfaces.
type NetworkMetrics struct {
	Interfaces []InterfaceMetrics `json:"interfaces"`
	Timestamp  time.Time          `json:"timestamp"`
}

// NetworkCollector manages periodic network metric collection.
type NetworkCollector struct {
	mu       sync.RWMutex
	metrics  NetworkMetrics
	interval time.Duration
	prevData map[string]InterfaceMetrics // previous snapshot for rate calculation
	stopCh   chan struct{}
}

// NewNetworkCollector creates a new network metrics collector.
func NewNetworkCollector(interval time.Duration) *NetworkCollector {
	return &NetworkCollector{
		interval: interval,
		prevData: make(map[string]InterfaceMetrics),
		stopCh:   make(chan struct{}),
	}
}

// GetMetrics returns a thread-safe copy of current network metrics.
func (n *NetworkCollector) GetMetrics() NetworkMetrics {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.metrics
}

// Stop signals the collector goroutine to exit.
func (n *NetworkCollector) Stop() {
	close(n.stopCh)
}

// Start begins periodic network metric collection in a background goroutine.
func (n *NetworkCollector) Start() {
	n.collect()
	ticker := time.NewTicker(n.interval)
	defer ticker.Stop()
	for {
		select {
		case <-n.stopCh:
			return
		case <-ticker.C:
			n.collect()
		}
	}
}

// collect performs a single network metric sampling cycle.
func (n *NetworkCollector) collect() {
	ifaces, err := readNetDev()
	if err != nil {
		return
	}

	intervalSec := n.interval.Seconds()
	if intervalSec == 0 {
		intervalSec = 1
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	for i := range ifaces {
		iface := &ifaces[i]
		if prev, ok := n.prevData[iface.Name]; ok {
			if iface.RxBytes >= prev.RxBytes {
				iface.RxRateBytesS = uint64(float64(iface.RxBytes-prev.RxBytes) / intervalSec)
			}
			if iface.TxBytes >= prev.TxBytes {
				iface.TxRateBytesS = uint64(float64(iface.TxBytes-prev.TxBytes) / intervalSec)
			}
		}
		// Store current snapshot for next rate calculation.
		n.prevData[iface.Name] = *iface
	}

	n.metrics = NetworkMetrics{
		Interfaces: ifaces,
		Timestamp:  time.Now(),
	}
}

// readNetDev parses /proc/net/dev for per-interface byte/packet counters.
func readNetDev() ([]InterfaceMetrics, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("open /proc/net/dev: %w", err)
	}
	defer f.Close()

	var ifaces []InterfaceMetrics
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		// First two lines are headers.
		if lineNum <= 2 {
			continue
		}

		// Format: iface: rx_bytes rx_packets rx_errs rx_drop ... tx_bytes tx_packets ...
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		name := strings.TrimSpace(line[:colonIdx])
		if name == "lo" {
			continue // skip loopback
		}
		fields := strings.Fields(line[colonIdx+1:])
		if len(fields) < 16 {
			continue
		}

		rxBytes := parseUint64(fields[0])
		rxPackets := parseUint64(fields[1])
		rxErrors := parseUint64(fields[2])
		rxDropped := parseUint64(fields[3])
		txBytes := parseUint64(fields[8])
		txPackets := parseUint64(fields[9])
		txErrors := parseUint64(fields[10])
		txDropped := parseUint64(fields[11])

		ifaces = append(ifaces, InterfaceMetrics{
			Name:      name,
			RxBytes:   rxBytes,
			TxBytes:   txBytes,
			RxPackets: rxPackets,
			TxPackets: txPackets,
			RxErrors:  rxErrors,
			TxErrors:  txErrors,
			RxDropped: rxDropped,
			TxDropped: txDropped,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan /proc/net/dev: %w", err)
	}
	return ifaces, nil
}

// parseUint64 is a fast decimal-to-uint64 parser without error return.
func parseUint64(s string) uint64 {
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + uint64(c-'0')
	}
	return n
}