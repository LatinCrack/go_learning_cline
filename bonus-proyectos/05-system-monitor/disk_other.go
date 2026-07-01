//go:build !linux
// +build !linux

package main

import "time"

// collect provides basic disk info for non-Linux platforms (Windows, macOS).
func (d *DiskCollector) collect() {
	partitions := []DiskPartition{
		{
			MountPoint: "/",
			Device:     "N/A",
			FSType:     "N/A",
		},
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = DiskMetrics{
		Partitions: partitions,
		Timestamp:  time.Now(),
	}
}