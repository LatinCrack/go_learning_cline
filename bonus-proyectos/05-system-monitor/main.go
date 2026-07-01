package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// Version is set at build time via ldflags.
var Version = "dev"

// Monitor is the top-level orchestrator that ties all collectors, alerter, and exporter together.
type Monitor struct {
	cpu     *CPUCollector
	mem     *MemoryCollector
	disk    *DiskCollector
	net     *NetworkCollector
	alerter *Alerter
	host    string
}

// NewMonitor creates and wires all subsystems.
func NewMonitor(interval time.Duration, alertCfg AlertConfig) *Monitor {
	hostname, _ := os.Hostname()
	return &Monitor{
		cpu:     NewCPUCollector(interval),
		mem:     NewMemoryCollector(interval),
		disk:    NewDiskCollector(interval),
		net:     NewNetworkCollector(interval),
		alerter: NewAlerter(alertCfg),
		host:    hostname,
	}
}

// GetSnapshot implements the MetricsProvider interface used by Exporter.
func (m *Monitor) GetSnapshot() Snapshot {
	cpuM := m.cpu.GetMetrics()
	memM := m.mem.GetMetrics()
	diskM := m.disk.GetMetrics()
	netM := m.net.GetMetrics()

	// Evaluate alert thresholds on every snapshot request.
	m.alerter.EvaluateCPU(cpuM)
	m.alerter.EvaluateMemory(memM)
	m.alerter.EvaluateDisk(diskM)

	return Snapshot{
		CPU:     cpuM,
		Memory:  memM,
		Disk:    diskM,
		Network: netM,
		Alerts:  m.alerter.GetAlerts(),
		Host:    m.host,
	}
}

// Start launches all collector goroutines.
func (m *Monitor) Start() {
	go m.cpu.Start()
	go m.mem.Start()
	go m.disk.Start()
	go m.net.Start()
	log.Printf("monitor started — host=%s cores=%d os=%s/%s",
		m.host, runtime.NumCPU(), runtime.GOOS, runtime.GOARCH)
}

// Stop signals all collector goroutines to exit.
func (m *Monitor) Stop() {
	m.cpu.Stop()
	m.mem.Stop()
	m.disk.Stop()
	m.net.Stop()
	log.Println("monitor stopped")
}

func main() {
	// ── CLI flags ──────────────────────────────────────────────
	interval := flag.Duration("interval", 5*time.Second,
		"Metric collection interval (e.g. 5s, 10s, 1m)")
	cpuAlert := flag.Float64("cpu-alert", 85.0,
		"CPU usage alert threshold percentage (0 to disable)")
	memAlert := flag.Float64("mem-alert", 90.0,
		"Memory usage alert threshold percentage (0 to disable)")
	diskAlert := flag.Float64("disk-alert", 95.0,
		"Disk usage alert threshold percentage (0 to disable)")
	swapAlert := flag.Float64("swap-alert", 80.0,
		"Swap usage alert threshold percentage (0 to disable)")
	listenAddr := flag.String("listen", ":9100",
		"HTTP listen address for the metrics exporter")
	showVersion := flag.Bool("version", false,
		"Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("system-monitor %s (go %s, %s/%s)\n",
			Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	// ── Configuration ─────────────────────────────────────────
	alertCfg := AlertConfig{
		CPUThreshold:  *cpuAlert,
		MemThreshold:  *memAlert,
		DiskThreshold: *diskAlert,
		SwapThreshold: *swapAlert,
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lmsgprefix)
	log.SetPrefix("[monitor] ")

	// ── Bootstrap ─────────────────────────────────────────────
	mon := NewMonitor(*interval, alertCfg)
	mon.Start()

	exporter := NewExporter(mon)
	go func() {
		log.Printf("metrics exporter listening on %s", *listenAddr)
		log.Printf("  → Prometheus: http://localhost%s/metrics", *listenAddr)
		log.Printf("  → JSON:       http://localhost%s/metrics/json", *listenAddr)
		log.Printf("  → Health:     http://localhost%s/health", *listenAddr)
		if err := exporter.ListenAndServe(*listenAddr); err != nil {
			log.Fatalf("exporter error: %v", err)
		}
	}()

	// ── Console banner ────────────────────────────────────────
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           System Resource Monitor v" + Version + "                ║")
	fmt.Println("╠══════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Interval    : %-40s ║\n", *interval)
	fmt.Printf("║  CPU Alert   : %-38s ║\n", formatThreshold(*cpuAlert))
	fmt.Printf("║  Mem Alert   : %-38s ║\n", formatThreshold(*memAlert))
	fmt.Printf("║  Disk Alert  : %-38s ║\n", formatThreshold(*diskAlert))
	fmt.Printf("║  Swap Alert  : %-38s ║\n", formatThreshold(*swapAlert))
	fmt.Printf("║  Listen      : %-40s ║\n", *listenAddr)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// ── Graceful shutdown ─────────────────────────────────────
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("received signal %v, shutting down...", sig)
	mon.Stop()
}

// formatThreshold renders a threshold value; 0 means disabled.
func formatThreshold(v float64) string {
	if v <= 0 {
		return "disabled"
	}
	return fmt.Sprintf("%.1f%%", v)
}