package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Snapshot aggregates all metric families into a single exportable structure.
type Snapshot struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Disk    DiskMetrics    `json:"disk"`
	Network NetworkMetrics `json:"network"`
	Alerts  []Alert        `json:"alerts"`
	Host    string         `json:"host"`
}

// MetricsProvider is the interface the main monitor implements to supply
// the latest snapshot of all collected metrics.
type MetricsProvider interface {
	GetSnapshot() Snapshot
}

// Exporter serves metrics over HTTP in JSON and Prometheus text format.
type Exporter struct {
	provider MetricsProvider
}

// NewExporter creates a new HTTP metrics exporter.
func NewExporter(provider MetricsProvider) *Exporter {
	return &Exporter{provider: provider}
}

// ListenAndServe starts the HTTP server on the given address (e.g. ":9100").
func (e *Exporter) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", e.handlePrometheus)
	mux.HandleFunc("/metrics/json", e.handleJSON)
	mux.HandleFunc("/health", e.handleHealth)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return srv.ListenAndServe()
}

// handleHealth returns a simple 200 OK for liveness probes.
func (e *Exporter) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// handleJSON serializes the full snapshot as JSON.
func (e *Exporter) handleJSON(w http.ResponseWriter, r *http.Request) {
	snap := e.provider.GetSnapshot()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Metrics-Timestamp", snap.CPU.Timestamp.Format(time.RFC3339))
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(snap); err != nil {
		http.Error(w, `{"error":"encoding failed"}`, http.StatusInternalServerError)
	}
}

// handlePrometheus renders metrics in Prometheus text exposition format.
// Format reference: https://prometheus.io/docs/instrumenting/exposition_formats/
func (e *Exporter) handlePrometheus(w http.ResponseWriter, r *http.Request) {
	snap := e.provider.GetSnapshot()
	var sb strings.Builder

	ts := snap.CPU.Timestamp.Unix()

	// --- CPU ---
	sb.WriteString("# HELP system_cpu_usage_percent Current CPU usage percentage.\n")
	sb.WriteString("# TYPE system_cpu_usage_percent gauge\n")
	sb.WriteString(fmt.Sprintf("system_cpu_usage_percent %.2f %d\n", snap.CPU.UsagePercent, ts))

	sb.WriteString("# HELP system_cpu_user_percent CPU user-space percentage.\n")
	sb.WriteString("# TYPE system_cpu_user_percent gauge\n")
	sb.WriteString(fmt.Sprintf("system_cpu_user_percent %.2f %d\n", snap.CPU.UserPercent, ts))

	sb.WriteString("# HELP system_cpu_system_percent CPU kernel-space percentage.\n")
	sb.WriteString("# TYPE system_cpu_system_percent gauge\n")
	sb.WriteString(fmt.Sprintf("system_cpu_system_percent %.2f %d\n", snap.CPU.SystemPercent, ts))

	sb.WriteString("# HELP system_cpu_iowait_percent CPU I/O wait percentage.\n")
	sb.WriteString("# TYPE system_cpu_iowait_percent gauge\n")
	sb.WriteString(fmt.Sprintf("system_cpu_iowait_percent %.2f %d\n", snap.CPU.IOwaitPercent, ts))

	sb.WriteString("# HELP system_cpu_cores Number of logical CPU cores.\n")
	sb.WriteString("# TYPE system_cpu_cores gauge\n")
	sb.WriteString(fmt.Sprintf("system_cpu_cores %d %d\n", snap.CPU.Cores, ts))

	// --- Memory ---
	sb.WriteString("# HELP system_memory_usage_percent Memory usage percentage.\n")
	sb.WriteString("# TYPE system_memory_usage_percent gauge\n")
	sb.WriteString(fmt.Sprintf("system_memory_usage_percent %.2f %d\n", snap.Memory.UsagePercent, ts))

	sb.WriteString("# HELP system_memory_total_mb Total memory in MB.\n")
	sb.WriteString("# TYPE system_memory_total_mb gauge\n")
	sb.WriteString(fmt.Sprintf("system_memory_total_mb %d %d\n", snap.Memory.TotalMB, ts))

	sb.WriteString("# HELP system_memory_used_mb Used memory in MB.\n")
	sb.WriteString("# TYPE system_memory_used_mb gauge\n")
	sb.WriteString(fmt.Sprintf("system_memory_used_mb %d %d\n", snap.Memory.UsedMB, ts))

	sb.WriteString("# HELP system_memory_free_mb Free memory in MB.\n")
	sb.WriteString("# TYPE system_memory_free_mb gauge\n")
	sb.WriteString(fmt.Sprintf("system_memory_free_mb %d %d\n", snap.Memory.FreeMB, ts))

	sb.WriteString("# HELP system_memory_buffers_mb Buffer memory in MB.\n")
	sb.WriteString("# TYPE system_memory_buffers_mb gauge\n")
	sb.WriteString(fmt.Sprintf("system_memory_buffers_mb %d %d\n", snap.Memory.BuffersMB, ts))

	sb.WriteString("# HELP system_memory_cached_mb Cached memory in MB.\n")
	sb.WriteString("# TYPE system_memory_cached_mb gauge\n")
	sb.WriteString(fmt.Sprintf("system_memory_cached_mb %d %d\n", snap.Memory.CachedMB, ts))

	sb.WriteString("# HELP system_swap_usage_percent Swap usage percentage.\n")
	sb.WriteString("# TYPE system_swap_usage_percent gauge\n")
	sb.WriteString(fmt.Sprintf("system_swap_usage_percent %.2f %d\n", snap.Memory.SwapPercent, ts))

	sb.WriteString("# HELP system_swap_total_mb Total swap in MB.\n")
	sb.WriteString("# TYPE system_swap_total_mb gauge\n")
	sb.WriteString(fmt.Sprintf("system_swap_total_mb %d %d\n", snap.Memory.SwapTotalMB, ts))

	sb.WriteString("# HELP system_swap_used_mb Used swap in MB.\n")
	sb.WriteString("# TYPE system_swap_used_mb gauge\n")
	sb.WriteString(fmt.Sprintf("system_swap_used_mb %d %d\n", snap.Memory.SwapUsedMB, ts))

	// --- Disk ---
	sb.WriteString("# HELP system_disk_usage_percent Disk usage percentage per partition.\n")
	sb.WriteString("# TYPE system_disk_usage_percent gauge\n")
	for _, p := range snap.Disk.Partitions {
		sb.WriteString(fmt.Sprintf("system_disk_usage_percent{mount=\"%s\",device=\"%s\",fstype=\"%s\"} %.2f %d\n",
			p.MountPoint, p.Device, p.FSType, p.UsagePercent, ts))
	}

	sb.WriteString("# HELP system_disk_total_mb Total disk space in MB per partition.\n")
	sb.WriteString("# TYPE system_disk_total_mb gauge\n")
	for _, p := range snap.Disk.Partitions {
		sb.WriteString(fmt.Sprintf("system_disk_total_mb{mount=\"%s\"} %d %d\n",
			p.MountPoint, p.TotalMB, ts))
	}

	sb.WriteString("# HELP system_disk_used_mb Used disk space in MB per partition.\n")
	sb.WriteString("# TYPE system_disk_used_mb gauge\n")
	for _, p := range snap.Disk.Partitions {
		sb.WriteString(fmt.Sprintf("system_disk_used_mb{mount=\"%s\"} %d %d\n",
			p.MountPoint, p.UsedMB, ts))
	}

	sb.WriteString("# HELP system_disk_free_mb Free disk space in MB per partition.\n")
	sb.WriteString("# TYPE system_disk_free_mb gauge\n")
	for _, p := range snap.Disk.Partitions {
		sb.WriteString(fmt.Sprintf("system_disk_free_mb{mount=\"%s\"} %d %d\n",
			p.MountPoint, p.FreeMB, ts))
	}

	sb.WriteString("# HELP system_disk_inodes_used Used inodes per partition.\n")
	sb.WriteString("# TYPE system_disk_inodes_used gauge\n")
	for _, p := range snap.Disk.Partitions {
		sb.WriteString(fmt.Sprintf("system_disk_inodes_used{mount=\"%s\"} %d %d\n",
			p.MountPoint, p.InodesUsed, ts))
	}

	sb.WriteString("# HELP system_disk_inodes_free Free inodes per partition.\n")
	sb.WriteString("# TYPE system_disk_inodes_free gauge\n")
	for _, p := range snap.Disk.Partitions {
		sb.WriteString(fmt.Sprintf("system_disk_inodes_free{mount=\"%s\"} %d %d\n",
			p.MountPoint, p.InodesFree, ts))
	}

	// --- Network ---
	sb.WriteString("# HELP system_net_rx_bytes Total received bytes per interface.\n")
	sb.WriteString("# TYPE system_net_rx_bytes counter\n")
	for _, iface := range snap.Network.Interfaces {
		sb.WriteString(fmt.Sprintf("system_net_rx_bytes{iface=\"%s\"} %d %d\n",
			iface.Name, iface.RxBytes, ts))
	}

	sb.WriteString("# HELP system_net_tx_bytes Total transmitted bytes per interface.\n")
	sb.WriteString("# TYPE system_net_tx_bytes counter\n")
	for _, iface := range snap.Network.Interfaces {
		sb.WriteString(fmt.Sprintf("system_net_tx_bytes{iface=\"%s\"} %d %d\n",
			iface.Name, iface.TxBytes, ts))
	}

	sb.WriteString("# HELP system_net_rx_rate_bytes_s Receive rate in bytes/sec.\n")
	sb.WriteString("# TYPE system_net_rx_rate_bytes_s gauge\n")
	for _, iface := range snap.Network.Interfaces {
		sb.WriteString(fmt.Sprintf("system_net_rx_rate_bytes_s{iface=\"%s\"} %d %d\n",
			iface.Name, iface.RxRateBytesS, ts))
	}

	sb.WriteString("# HELP system_net_tx_rate_bytes_s Transmit rate in bytes/sec.\n")
	sb.WriteString("# TYPE system_net_tx_rate_bytes_s gauge\n")
	for _, iface := range snap.Network.Interfaces {
		sb.WriteString(fmt.Sprintf("system_net_tx_rate_bytes_s{iface=\"%s\"} %d %d\n",
			iface.Name, iface.TxRateBytesS, ts))
	}

	sb.WriteString("# HELP system_net_rx_errors Total receive errors per interface.\n")
	sb.WriteString("# TYPE system_net_rx_errors counter\n")
	for _, iface := range snap.Network.Interfaces {
		sb.WriteString(fmt.Sprintf("system_net_rx_errors{iface=\"%s\"} %d %d\n",
			iface.Name, iface.RxErrors, ts))
	}

	sb.WriteString("# HELP system_net_tx_errors Total transmit errors per interface.\n")
	sb.WriteString("# TYPE system_net_tx_errors counter\n")
	for _, iface := range snap.Network.Interfaces {
		sb.WriteString(fmt.Sprintf("system_net_tx_errors{iface=\"%s\"} %d %d\n",
			iface.Name, iface.TxErrors, ts))
	}

	// --- Alerts ---
	sb.WriteString("# HELP system_alerts_active Number of active alerts.\n")
	sb.WriteString("# TYPE system_alerts_active gauge\n")
	sb.WriteString(fmt.Sprintf("system_alerts_active %d %d\n", len(snap.Alerts), ts))

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Write([]byte(sb.String()))
}

// FormatBytes is a utility to render byte counts in human-readable form.
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}