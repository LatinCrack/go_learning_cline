package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ──────────────────────────────────────────────
// CPU Tests
// ──────────────────────────────────────────────

func TestCPUData_AddDelta(t *testing.T) {
	d := CPUData{
		User: 100, Nice: 50, System: 80, Idle: 700,
		IOWait: 30, IRQ: 10, SoftIRQ: 20, Steal: 5,
	}
	total := d.Total()
	if total != 995 {
		t.Errorf("expected total=995, got %d", total)
	}

	other := CPUData{
		User: 200, Nice: 55, System: 90, Idle: 1300,
		IOWait: 40, IRQ: 15, SoftIRQ: 25, Steal: 8,
	}
	// newer.Delta(older) → other is the newer snapshot.
	delta := other.Delta(d)
	if delta.User != 100 {
		t.Errorf("expected delta.User=100, got %d", delta.User)
	}
	if delta.Idle != 600 {
		t.Errorf("expected delta.Idle=600, got %d", delta.Idle)
	}
}

func TestCPUData_ZeroTotal(t *testing.T) {
	d := CPUData{}
	if d.Total() != 0 {
		t.Errorf("expected zero total, got %d", d.Total())
	}
}

func TestNewCPUCollector(t *testing.T) {
	c := NewCPUCollector(5 * time.Second)
	if c == nil {
		t.Fatal("NewCPUCollector returned nil")
	}
	// Trigger an initial collect to populate the metrics (including Cores).
	c.collect()
	m := c.GetMetrics()
	if m.Cores <= 0 {
		t.Errorf("expected cores > 0, got %d", m.Cores)
	}
}

// ──────────────────────────────────────────────
// Memory Tests
// ──────────────────────────────────────────────

func TestNewMemoryCollector(t *testing.T) {
	c := NewMemoryCollector(5 * time.Second)
	if c == nil {
		t.Fatal("NewMemoryCollector returned nil")
	}
}

func TestMemoryCollector_GetMetrics(t *testing.T) {
	c := NewMemoryCollector(5 * time.Second)
	c.collect()
	m := c.GetMetrics()
	if m.TotalMB == 0 {
		t.Log("Warning: TotalMB is 0 — may be running on a platform without /proc/meminfo")
	}
}

func TestMemoryCollector_ThreadSafety(t *testing.T) {
	c := NewMemoryCollector(5 * time.Second)
	c.collect()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			_ = c.GetMetrics()
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		c.collect()
	}
	<-done
}

// ──────────────────────────────────────────────
// Disk Tests
// ──────────────────────────────────────────────

func TestNewDiskCollector(t *testing.T) {
	c := NewDiskCollector(5 * time.Second)
	if c == nil {
		t.Fatal("NewDiskCollector returned nil")
	}
}

func TestDiskCollector_GetMetrics(t *testing.T) {
	c := NewDiskCollector(5 * time.Second)
	c.collect()
	m := c.GetMetrics()
	if m.Partitions == nil {
		t.Log("Warning: Partitions is nil — may be running on Windows without /proc/mounts")
	}
}

func TestDiskCollector_ThreadSafety(t *testing.T) {
	c := NewDiskCollector(5 * time.Second)
	c.collect()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			_ = c.GetMetrics()
		}
		close(done)
	}()

	for i := 0; i < 50; i++ {
		c.collect()
	}
	<-done
}

// ──────────────────────────────────────────────
// Network Tests
// ──────────────────────────────────────────────

func TestNewNetworkCollector(t *testing.T) {
	c := NewNetworkCollector(5 * time.Second)
	if c == nil {
		t.Fatal("NewNetworkCollector returned nil")
	}
}

func TestNetworkCollector_GetMetrics(t *testing.T) {
	c := NewNetworkCollector(5 * time.Second)
	c.collect()
	m := c.GetMetrics()
	if m.Interfaces == nil {
		t.Log("Warning: Interfaces is nil — may be running on a platform without /proc/net/dev")
	}
}

func TestParseUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"0", 0},
		{"12345", 12345},
		{"99999999999", 99999999999},
		{"abc", 0},
		{"12ab", 12},
		{"", 0},
	}
	for _, tt := range tests {
		result := parseUint64(tt.input)
		if result != tt.expected {
			t.Errorf("parseUint64(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

// ──────────────────────────────────────────────
// Alerter Tests
// ──────────────────────────────────────────────

func TestAlerter_CPUBelowThreshold(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.CPUThreshold = 85.0
	a := NewAlerter(cfg)

	metrics := CPUMetrics{UsagePercent: 50.0}
	a.EvaluateCPU(metrics)
	alerts := a.GetAlerts()
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts for CPU=50%%, got %d", len(alerts))
	}
}

func TestAlerter_CPUAboveThreshold(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.CPUThreshold = 85.0
	a := NewAlerter(cfg)
	// Reset cooldown for testing.
	a.mu.Lock()
	a.cooldown = 0
	a.mu.Unlock()

	metrics := CPUMetrics{UsagePercent: 92.0}
	a.EvaluateCPU(metrics)
	alerts := a.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert for CPU=92%%, got %d", len(alerts))
	}
	if alerts[0].Component != "cpu" {
		t.Errorf("expected component=cpu, got %s", alerts[0].Component)
	}
	if alerts[0].Level != AlertLevelWarning {
		t.Errorf("expected level=WARNING, got %s", alerts[0].Level)
	}
}

func TestAlerter_MemoryAboveThreshold(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.MemThreshold = 90.0
	a := NewAlerter(cfg)
	a.mu.Lock()
	a.cooldown = 0
	a.mu.Unlock()

	a.EvaluateMemory(MemoryMetrics{UsagePercent: 95.0})
	alerts := a.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert for memory=95%%, got %d", len(alerts))
	}
	if alerts[0].Component != "memory" {
		t.Errorf("expected component=memory, got %s", alerts[0].Component)
	}
}

func TestAlerter_DiskAboveThreshold(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.DiskThreshold = 95.0
	a := NewAlerter(cfg)
	a.mu.Lock()
	a.cooldown = 0
	a.mu.Unlock()

	dm := DiskMetrics{
		Partitions: []DiskPartition{
			{MountPoint: "/data", UsagePercent: 97.0},
			{MountPoint: "/boot", UsagePercent: 50.0},
		},
	}
	a.EvaluateDisk(dm)
	alerts := a.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 disk alert, got %d", len(alerts))
	}
	if !strings.Contains(alerts[0].Message, "/data") {
		t.Errorf("expected alert message to contain /data, got: %s", alerts[0].Message)
	}
}

func TestAlerter_SwapAboveThreshold(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.SwapThreshold = 80.0
	a := NewAlerter(cfg)
	a.mu.Lock()
	a.cooldown = 0
	a.mu.Unlock()

	a.EvaluateMemory(MemoryMetrics{SwapTotalMB: 2048, SwapPercent: 85.0, UsagePercent: 50})
	alerts := a.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("expected 1 swap alert, got %d", len(alerts))
	}
	if alerts[0].Component != "swap" {
		t.Errorf("expected component=swap, got %s", alerts[0].Component)
	}
}

func TestAlerter_Cooldown(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.CPUThreshold = 85.0
	a := NewAlerter(cfg)
	a.mu.Lock()
	a.cooldown = 10 * time.Second // long cooldown
	a.mu.Unlock()

	metrics := CPUMetrics{UsagePercent: 95.0}
	a.EvaluateCPU(metrics)
	a.EvaluateCPU(metrics) // second call should be suppressed
	alerts := a.GetAlerts()
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert (cooldown suppressed second), got %d", len(alerts))
	}
}

func TestAlerter_DisabledThreshold(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.CPUThreshold = 0 // disabled
	a := NewAlerter(cfg)
	a.EvaluateCPU(CPUMetrics{UsagePercent: 99.0})
	alerts := a.GetAlerts()
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts when threshold=0, got %d", len(alerts))
	}
}

func TestAlerter_SetConfig(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.CPUThreshold = 85.0
	a := NewAlerter(cfg)

	newCfg := cfg
	newCfg.CPUThreshold = 50.0
	a.SetConfig(newCfg)
	a.mu.Lock()
	a.cooldown = 0
	a.mu.Unlock()

	a.EvaluateCPU(CPUMetrics{UsagePercent: 60.0})
	alerts := a.GetAlerts()
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert with new config, got %d", len(alerts))
	}
}

func TestAlerter_RingBufferOverflow(t *testing.T) {
	cfg := DefaultAlertConfig()
	cfg.CPUThreshold = 1.0
	a := NewAlerter(cfg)
	a.mu.Lock()
	a.cooldown = 0
	a.maxHist = 5
	a.mu.Unlock()

	for i := 0; i < 10; i++ {
		a.EvaluateCPU(CPUMetrics{UsagePercent: float64(50 + i)})
	}
	alerts := a.GetAlerts()
	if len(alerts) > 5 {
		t.Errorf("expected alerts capped at 5, got %d", len(alerts))
	}
}

// ──────────────────────────────────────────────
// Exporter Tests
// ──────────────────────────────────────────────

type mockProvider struct{}

func (m mockProvider) GetSnapshot() Snapshot {
	return Snapshot{
		CPU: CPUMetrics{
			UsagePercent: 42.5,
			UserPercent:  30.0,
			SystemPercent: 12.5,
			Cores:        4,
			Timestamp:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Memory: MemoryMetrics{
			TotalMB:      16384,
			UsedMB:       8192,
			FreeMB:       8192,
			UsagePercent: 50.0,
		},
		Disk: DiskMetrics{
			Partitions: []DiskPartition{
				{MountPoint: "/", TotalMB: 500000, UsagePercent: 45.0},
			},
		},
		Network: NetworkMetrics{
			Interfaces: []InterfaceMetrics{
				{Name: "eth0", RxBytes: 1000, TxBytes: 500},
			},
		},
		Host: "test-host",
	}
}

func TestExporter_JSONEndpoint(t *testing.T) {
	e := NewExporter(mockProvider{})
	req := httptest.NewRequest(http.MethodGet, "/metrics/json", nil)
	w := httptest.NewRecorder()
	e.handleJSON(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected JSON content type, got %s", ct)
	}

	var snap Snapshot
	if err := json.NewDecoder(w.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if snap.CPU.UsagePercent != 42.5 {
		t.Errorf("expected CPU=42.5, got %f", snap.CPU.UsagePercent)
	}
	if snap.Memory.TotalMB != 16384 {
		t.Errorf("expected MemTotal=16384, got %d", snap.Memory.TotalMB)
	}
}

func TestExporter_PrometheusEndpoint(t *testing.T) {
	e := NewExporter(mockProvider{})
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	e.handlePrometheus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "system_cpu_usage_percent") {
		t.Error("missing system_cpu_usage_percent metric")
	}
	if !strings.Contains(body, "system_memory_usage_percent") {
		t.Error("missing system_memory_usage_percent metric")
	}
	if !strings.Contains(body, "system_disk_usage_percent") {
		t.Error("missing system_disk_usage_percent metric")
	}
	if !strings.Contains(body, "system_net_rx_bytes") {
		t.Error("missing system_net_rx_bytes metric")
	}
	if !strings.Contains(body, "system_alerts_active") {
		t.Error("missing system_alerts_active metric")
	}
	// Verify HELP and TYPE annotations.
	if !strings.Contains(body, "# HELP system_cpu_usage_percent") {
		t.Error("missing HELP annotation for CPU metric")
	}
	if !strings.Contains(body, "# TYPE system_cpu_usage_percent gauge") {
		t.Error("missing TYPE annotation for CPU metric")
	}
}

func TestExporter_HealthEndpoint(t *testing.T) {
	e := NewExporter(mockProvider{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	e.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `"status":"ok"`) {
		t.Errorf("expected ok status, got: %s", body)
	}
}

// ──────────────────────────────────────────────
// Utility Tests
// ──────────────────────────────────────────────

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{1099511627776, "1.00 TB"},
		{5368709120, "5.00 GB"},
	}
	for _, tt := range tests {
		result := FormatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDefaultAlertConfig(t *testing.T) {
	cfg := DefaultAlertConfig()
	if cfg.CPUThreshold != 85.0 {
		t.Errorf("expected CPU threshold 85, got %f", cfg.CPUThreshold)
	}
	if cfg.MemThreshold != 90.0 {
		t.Errorf("expected Mem threshold 90, got %f", cfg.MemThreshold)
	}
	if cfg.DiskThreshold != 95.0 {
		t.Errorf("expected Disk threshold 95, got %f", cfg.DiskThreshold)
	}
	if cfg.SwapThreshold != 80.0 {
		t.Errorf("expected Swap threshold 80, got %f", cfg.SwapThreshold)
	}
}