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

// CPUMetrics holds the current CPU usage data.
type CPUMetrics struct {
	UsagePercent float64 `json:"usage_percent"`
	UserPercent  float64 `json:"user_percent"`
	SystemPercent float64 `json:"system_percent"`
	IdlePercent  float64 `json:"idle_percent"`
	IOwaitPercent float64 `json:"iowait_percent"`
	Cores        int     `json:"cores"`
	Timestamp    time.Time `json:"timestamp"`
}

// CPUData represents cumulative CPU tick counts from /proc/stat.
type CPUData struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
}

// Total returns the sum of all CPU time fields.
func (d CPUData) Total() uint64 {
	return d.User + d.Nice + d.System + d.Idle + d.IOWait + d.IRQ + d.SoftIRQ + d.Steal
}

// Delta returns the difference between the receiver (newer snapshot) and
// the provided (older) snapshot.  The caller should invoke:
//
//	newer.Delta(older)
//
// to obtain the delta values.  If older fields are larger the result
// wraps around to zero (unsigned arithmetic), which is safe for the
// normal forward-in-time usage.
func (d CPUData) Delta(older CPUData) CPUData {
	return CPUData{
		User:    d.User - older.User,
		Nice:    d.Nice - older.Nice,
		System:  d.System - older.System,
		Idle:    d.Idle - older.Idle,
		IOWait:  d.IOWait - older.IOWait,
		IRQ:     d.IRQ - older.IRQ,
		SoftIRQ: d.SoftIRQ - older.SoftIRQ,
		Steal:   d.Steal - older.Steal,
	}
}

// CPUCollector manages periodic CPU metric collection.
type CPUCollector struct {
	mu          sync.RWMutex
	metrics     CPUMetrics
	interval    time.Duration
	prevTotal   uint64
	prevIdle    uint64
	prevUser    uint64
	prevSystem  uint64
	prevIOWait  uint64
	initialized bool
	stopCh      chan struct{}
}

// NewCPUCollector creates a new CPU metrics collector.
func NewCPUCollector(interval time.Duration) *CPUCollector {
	return &CPUCollector{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// GetMetrics returns a thread-safe copy of current CPU metrics.
func (c *CPUCollector) GetMetrics() CPUMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics
}

// Stop signals the collector goroutine to exit.
func (c *CPUCollector) Stop() {
	close(c.stopCh)
}

// Start begins periodic CPU metric collection in a background goroutine.
func (c *CPUCollector) Start() {
	// Take an initial snapshot to establish baseline deltas.
	c.collect()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.collect()
		}
	}
}

// collect performs a single CPU metric sampling cycle.
func (c *CPUCollector) collect() {
	if runtime.GOOS == "linux" || runtime.GOOS == "android" {
		c.collectLinux()
	} else {
		c.collectFallback()
	}
}

// collectLinux reads /proc/stat and computes CPU percentages from deltas.
func (c *CPUCollector) collectLinux() {
	times, err := readProcStat()
	if err != nil {
		return
	}

	total := times.User + times.Nice + times.System + times.Idle +
		times.IOWait + times.IRQ + times.SoftIRQ + times.Steal

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.initialized {
		// First reading: store baseline, metrics remain at 0%.
		c.prevTotal = total
		c.prevIdle = times.Idle + times.IOWait
		c.prevUser = times.User + times.Nice
		c.prevSystem = times.System + times.IRQ + times.SoftIRQ
		c.prevIOWait = times.IOWait
		c.initialized = true
		c.metrics = CPUMetrics{
			Cores:     runtime.NumCPU(),
			Timestamp: time.Now(),
		}
		return
	}

	deltaTotal := float64(total - c.prevTotal)
	if deltaTotal == 0 {
		return
	}

	deltaIdle := float64((times.Idle+times.IOWait) - c.prevIdle)
	deltaUser := float64((times.User+times.Nice) - c.prevUser)
	deltaSystem := float64((times.System+times.IRQ+times.SoftIRQ) - c.prevSystem)
	deltaIOWait := float64(times.IOWait - c.prevIOWait)

	// Advance baseline for next cycle.
	c.prevTotal = total
	c.prevIdle = times.Idle + times.IOWait
	c.prevUser = times.User + times.Nice
	c.prevSystem = times.System + times.IRQ + times.SoftIRQ
	c.prevIOWait = times.IOWait

	c.metrics = CPUMetrics{
		UsagePercent:   ((deltaTotal - deltaIdle) / deltaTotal) * 100,
		UserPercent:    (deltaUser / deltaTotal) * 100,
		SystemPercent:  (deltaSystem / deltaTotal) * 100,
		IdlePercent:    (deltaIdle / deltaTotal) * 100,
		IOwaitPercent:  (deltaIOWait / deltaTotal) * 100,
		Cores:          runtime.NumCPU(),
		Timestamp:      time.Now(),
	}
}

// readProcStat parses the aggregate "cpu …" line from /proc/stat.
func readProcStat() (CPUData, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return CPUData{}, fmt.Errorf("open /proc/stat: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			return parseCPULine(line)
		}
	}
	return CPUData{}, fmt.Errorf("cpu line not found in /proc/stat")
}

// parseCPULine extracts the numeric fields from a /proc/stat cpu line.
// Expected format: cpu  user nice system idle iowait irq softirq steal [guest guest_nice]
func parseCPULine(line string) (CPUData, error) {
	fields := strings.Fields(line)
	if len(fields) < 9 {
		return CPUData{}, fmt.Errorf("unexpected cpu line format: %s", line)
	}

	vals := make([]uint64, 8)
	for i := 0; i < 8; i++ {
		v, err := strconv.ParseUint(fields[i+1], 10, 64)
		if err != nil {
		return CPUData{}, fmt.Errorf("parse field %d: %w", i, err)
		}
		vals[i] = v
	}

	return CPUData{
		User:    vals[0],
		Nice:    vals[1],
		System:  vals[2],
		Idle:    vals[3],
		IOWait:  vals[4],
		IRQ:     vals[5],
		SoftIRQ: vals[6],
		Steal:   vals[7],
	}, nil
}

// collectFallback is a placeholder for non-Linux platforms.
// On Windows/macOS we still report core count; usage requires elevated APIs.
func (c *CPUCollector) collectFallback() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = CPUMetrics{
		UsagePercent: 0,
		Cores:        runtime.NumCPU(),
		Timestamp:    time.Now(),
	}
}