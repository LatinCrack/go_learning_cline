package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MemoryMetrics holds current memory usage data.
type MemoryMetrics struct {
	TotalMB       uint64  `json:"total_mb"`
	AvailableMB   uint64  `json:"available_mb"`
	UsedMB        uint64  `json:"used_mb"`
	BuffersMB     uint64  `json:"buffers_mb"`
	CachedMB      uint64  `json:"cached_mb"`
	FreeMB        uint64  `json:"free_mb"`
	UsagePercent  float64 `json:"usage_percent"`
	SwapTotalMB   uint64  `json:"swap_total_mb"`
	SwapUsedMB    uint64  `json:"swap_used_mb"`
	SwapFreeMB    uint64  `json:"swap_free_mb"`
	SwapPercent   float64 `json:"swap_percent"`
	Timestamp     time.Time `json:"timestamp"`
}

// MemoryCollector manages periodic memory metric collection.
type MemoryCollector struct {
	mu       sync.RWMutex
	metrics  MemoryMetrics
	interval time.Duration
	stopCh   chan struct{}
}

// NewMemoryCollector creates a new memory metrics collector.
func NewMemoryCollector(interval time.Duration) *MemoryCollector {
	return &MemoryCollector{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// GetMetrics returns a thread-safe copy of current memory metrics.
func (m *MemoryCollector) GetMetrics() MemoryMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.metrics
}

// Stop signals the collector goroutine to exit.
func (m *MemoryCollector) Stop() {
	close(m.stopCh)
}

// Start begins periodic memory metric collection in a background goroutine.
func (m *MemoryCollector) Start() {
	m.collect()
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.collect()
		}
	}
}

// collect performs a single memory metric sampling cycle.
func (m *MemoryCollector) collect() {
	if runtime.GOOS == "linux" || runtime.GOOS == "android" {
		m.collectLinux()
	} else {
		m.collectFallback()
	}
}

// collectLinux reads /proc/meminfo and computes memory usage.
func (m *MemoryCollector) collectLinux() {
	info, err := readProcMeminfo()
	if err != nil {
		return
	}

	totalKB := info["MemTotal"]
	availKB := info["MemAvailable"]
	freeKB := info["MemFree"]
	buffersKB := info["Buffers"]
	cachedKB := info["Cached"]
	swapTotalKB := info["SwapTotal"]
	swapFreeKB := info["SwapFree"]

	// If MemAvailable is not present (kernels < 3.14), approximate it.
	if availKB == 0 {
		availKB = freeKB + buffersKB + cachedKB
	}

	totalMB := totalKB / 1024
	availMB := availKB / 1024
	freeMB := freeKB / 1024
	buffersMB := buffersKB / 1024
	cachedMB := cachedKB / 1024
	usedMB := totalMB - availMB

	var usagePercent float64
	if totalMB > 0 {
		usagePercent = (float64(usedMB) / float64(totalMB)) * 100
	}

	swapTotalMB := swapTotalKB / 1024
	swapFreeMB := swapFreeKB / 1024
	swapUsedMB := swapTotalMB - swapFreeMB

	var swapPercent float64
	if swapTotalMB > 0 {
		swapPercent = (float64(swapUsedMB) / float64(swapTotalMB)) * 100
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = MemoryMetrics{
		TotalMB:      totalMB,
		AvailableMB:  availMB,
		UsedMB:       usedMB,
		BuffersMB:    buffersMB,
		CachedMB:     cachedMB,
		FreeMB:       freeMB,
		UsagePercent: usagePercent,
		SwapTotalMB:  swapTotalMB,
		SwapUsedMB:   swapUsedMB,
		SwapFreeMB:   swapFreeMB,
		SwapPercent:  swapPercent,
		Timestamp:    time.Now(),
	}
}

// readProcMeminfo parses /proc/meminfo into a key→kB map.
func readProcMeminfo() (map[string]uint64, error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("open /proc/meminfo: %w", err)
	}
	defer f.Close()

	result := make(map[string]uint64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Format: "MemTotal:       16384000 kB"
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[1])
		valStr = strings.TrimSuffix(valStr, " kB")
		valStr = strings.TrimSpace(valStr)
		val, err := strconv.ParseUint(valStr, 10, 64)
		if err != nil {
			continue
		}
		result[key] = val
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan /proc/meminfo: %w", err)
	}
	return result, nil
}

// collectFallback uses Go's runtime.MemStats as a rough approximation.
func (m *MemoryCollector) collectFallback() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	// Sys is the total memory obtained from the OS (approximate).
	totalMB := ms.Sys / 1024 / 1024
	heapUsedMB := ms.HeapInuse / 1024 / 1024
	heapIdleMB := ms.HeapIdle / 1024 / 1024

	var usagePercent float64
	if totalMB > 0 {
		usagePercent = (float64(heapUsedMB) / float64(totalMB)) * 100
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics = MemoryMetrics{
		TotalMB:      totalMB,
		UsedMB:       heapUsedMB,
		FreeMB:       heapIdleMB,
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
	}
}