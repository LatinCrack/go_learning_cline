//go:build linux
// +build linux

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

// collect performs a single disk metric sampling cycle on Linux.
func (d *DiskCollector) collect() {
	mounts, err := parseMounts()
	if err != nil {
		return
	}

	var partitions []DiskPartition
	for _, mp := range mounts {
		var stat syscall.Statfs_t
		if err := syscall.Statfs(mp.mountPoint, &stat); err != nil {
			continue
		}

		totalBytes := stat.Blocks * uint64(stat.Bsize)
		freeBytes := stat.Bfree * uint64(stat.Bsize)
		availBytes := stat.Bavail * uint64(stat.Bsize)
		usedBytes := totalBytes - freeBytes

		totalMB := totalBytes / (1024 * 1024)
		freeMB := freeBytes / (1024 * 1024)
		usedMB := usedBytes / (1024 * 1024)

		var usagePercent float64
		if totalBytes > 0 {
			usagePercent = (1 - float64(availBytes)/float64(totalBytes)) * 100
		}

		var inodesTotal, inodesFree, inodesUsed uint64
		if stat.Files > 0 {
			inodesTotal = stat.Files
			inodesFree = stat.Ffree
			inodesUsed = inodesTotal - inodesFree
		}

		partitions = append(partitions, DiskPartition{
			MountPoint:   mp.mountPoint,
			Device:       mp.device,
			FSType:       mp.fsType,
			TotalMB:      totalMB,
			UsedMB:       usedMB,
			FreeMB:       freeMB,
			UsagePercent: usagePercent,
			InodesTotal:  inodesTotal,
			InodesUsed:   inodesUsed,
			InodesFree:   inodesFree,
		})
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.metrics = DiskMetrics{
		Partitions: partitions,
		Timestamp:  timeNow(),
	}
}

type mountInfo struct {
	device     string
	mountPoint string
	fsType     string
}

// parseMounts reads /proc/mounts and returns real-filesystem mount points.
func parseMounts() ([]mountInfo, error) {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, fmt.Errorf("open /proc/mounts: %w", err)
	}
	defer f.Close()

	skipTypes := map[string]bool{
		"proc": true, "sysfs": true, "devpts": true, "tmpfs": true,
		"cgroup": true, "cgroup2": true, "pstore": true, "debugfs": true,
		"securityfs": true, "devtmpfs": true, "hugetlbfs": true,
		"mqueue": true, "overlay": true, "nsfs": true, "fusectl": true,
		"configfs": true, "tracefs": true, "bpf": true, "binfmt_misc": true,
		"rpc_pipefs": true, "nfsd": true, "squashfs": true, "autofs": true,
		"efivarfs": true, "fuse.gvfsd-fuse": true,
	}

	var mounts []mountInfo
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mp := fields[1]
		fsType := fields[2]

		if skipTypes[fsType] {
			continue
		}
		if !strings.HasPrefix(device, "/dev/") {
			continue
		}
		if seen[mp] {
			continue
		}
		seen[mp] = true

		mounts = append(mounts, mountInfo{
			device:     device,
			mountPoint: mp,
			fsType:     fsType,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan /proc/mounts: %w", err)
	}
	return mounts, nil
}

func timeNow() time.Time {
	return time.Now()
}