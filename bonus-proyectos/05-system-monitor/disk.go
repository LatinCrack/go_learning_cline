package main

import (
	"sync"
	"time"
)

// DiskPartition holds metrics for a single mount point.
type DiskPartition struct {
	MountPoint   string  `json:"mount_point"`
	Device       string  `json:"device"`
	FSType       string  `json:"fs_type"`
	TotalMB      uint64  `json:"total_mb"`
	UsedMB       uint64  `json:"used_mb"`
	FreeMB       uint64  `json:"free_mb"`
	UsagePercent float64 `json:"usage_percent"`
	InodesTotal  uint64  `json:"inodes_total"`
	InodesUsed   uint64  `json:"inodes_used"`
	InodesFree   uint64  `json:"inodes_free"`
}

// DiskMetrics holds current disk usage data for all monitored partitions.
type DiskMetrics struct {
	Partitions []DiskPartition `json:"partitions"`
	Timestamp  time.Time       `json:"timestamp"`
}

// DiskCollector manages periodic disk metric collection.
type DiskCollector struct {
	mu       sync.RWMutex
	metrics  DiskMetrics
	interval time.Duration
	stopCh   chan struct{}
}

// NewDiskCollector creates a new disk metrics collector.
func NewDiskCollector(interval time.Duration) *DiskCollector {
	return &DiskCollector{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// GetMetrics returns a thread-safe copy of current disk metrics.
func (d *DiskCollector) GetMetrics() DiskMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.metrics
}

// Stop signals the collector goroutine to exit.
func (d *DiskCollector) Stop() {
	close(d.stopCh)
}

// Start begins periodic disk metric collection in a background goroutine.
func (d *DiskCollector) Start() {
	d.collect()
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.collect()
		}
	}
}