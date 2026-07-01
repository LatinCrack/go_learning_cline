package main

import (
	"log"
	"sync"
	"time"
)

// AlertLevel represents the severity of an alert.
type AlertLevel string

const (
	AlertLevelWarning  AlertLevel = "WARNING"
	AlertLevelCritical AlertLevel = "CRITICAL"
)

// Alert represents a single triggered alert.
type Alert struct {
	Level     AlertLevel `json:"level"`
	Component string     `json:"component"`
	Message   string     `json:"message"`
	Value     float64    `json:"value"`
	Threshold float64    `json:"threshold"`
	Timestamp time.Time  `json:"timestamp"`
}

// AlertConfig holds configurable thresholds for each monitored resource.
type AlertConfig struct {
	CPUThreshold    float64 // e.g. 85.0 means 85%
	MemThreshold    float64 // e.g. 90.0 means 90%
	DiskThreshold   float64 // e.g. 95.0 means 95%
	SwapThreshold   float64 // e.g. 80.0 means 80%
}

// DefaultAlertConfig returns sensible default thresholds.
func DefaultAlertConfig() AlertConfig {
	return AlertConfig{
		CPUThreshold:  85.0,
		MemThreshold:  90.0,
		DiskThreshold: 95.0,
		SwapThreshold: 80.0,
	}
}

// Alerter evaluates metrics against thresholds and fires log warnings.
type Alerter struct {
	mu      sync.RWMutex
	config  AlertConfig
	alerts  []Alert // recent alert history (ring buffer)
	maxHist int
	// Cooldown tracking: avoid spamming the same alert every interval.
	lastFired map[string]time.Time
	cooldown  time.Duration
}

// NewAlerter creates a new Alerter with the given configuration.
func NewAlerter(cfg AlertConfig) *Alerter {
	return &Alerter{
		config:    cfg,
		alerts:    make([]Alert, 0, 64),
		maxHist:   100,
		lastFired: make(map[string]time.Time),
		cooldown:  60 * time.Second, // minimum 60s between identical alerts
	}
}

// SetConfig allows updating thresholds at runtime.
func (a *Alerter) SetConfig(cfg AlertConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.config = cfg
}

// GetAlerts returns a copy of recent alerts.
func (a *Alerter) GetAlerts() []Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]Alert, len(a.alerts))
	copy(out, a.alerts)
	return out
}

// EvaluateCPU checks CPU usage against the configured threshold.
func (a *Alerter) EvaluateCPU(metrics CPUMetrics) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.config.CPUThreshold <= 0 {
		return
	}
	if metrics.UsagePercent >= a.config.CPUThreshold {
		a.fire(Alert{
			Level:     AlertLevelWarning,
			Component: "cpu",
			Message:   "CPU usage above threshold",
			Value:     metrics.UsagePercent,
			Threshold: a.config.CPUThreshold,
			Timestamp: time.Now(),
		})
	}
}

// EvaluateMemory checks memory usage against the configured threshold.
func (a *Alerter) EvaluateMemory(metrics MemoryMetrics) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.config.MemThreshold > 0 && metrics.UsagePercent >= a.config.MemThreshold {
		a.fire(Alert{
			Level:     AlertLevelWarning,
			Component: "memory",
			Message:   "Memory usage above threshold",
			Value:     metrics.UsagePercent,
			Threshold: a.config.MemThreshold,
			Timestamp: time.Now(),
		})
	}
	if a.config.SwapThreshold > 0 && metrics.SwapTotalMB > 0 && metrics.SwapPercent >= a.config.SwapThreshold {
		a.fire(Alert{
			Level:     AlertLevelWarning,
			Component: "swap",
			Message:   "Swap usage above threshold",
			Value:     metrics.SwapPercent,
			Threshold: a.config.SwapThreshold,
			Timestamp: time.Now(),
		})
	}
}

// EvaluateDisk checks each disk partition usage against the configured threshold.
func (a *Alerter) EvaluateDisk(metrics DiskMetrics) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.config.DiskThreshold <= 0 {
		return
	}
	for _, p := range metrics.Partitions {
		if p.UsagePercent >= a.config.DiskThreshold {
			a.fire(Alert{
				Level:     AlertLevelCritical,
				Component: "disk",
				Message:   "Disk usage above threshold on " + p.MountPoint,
				Value:     p.UsagePercent,
				Threshold: a.config.DiskThreshold,
				Timestamp: time.Now(),
			})
		}
	}
}

// fire appends the alert to history and logs it, respecting cooldown.
func (a *Alerter) fire(alert Alert) {
	key := alert.Component + ":" + alert.Message
	if last, ok := a.lastFired[key]; ok {
		if time.Since(last) < a.cooldown {
			return // still in cooldown
		}
	}
	a.lastFired[key] = alert.Timestamp

	// Append to ring buffer.
	if len(a.alerts) >= a.maxHist {
		// Shift oldest out.
		copy(a.alerts, a.alerts[1:])
		a.alerts = a.alerts[:a.maxHist-1]
	}
	a.alerts = append(a.alerts, alert)

	// Log to stderr with structured format.
	log.Printf("[ALERT][%s] %s: %.2f%% (threshold: %.2f%%)",
		alert.Level, alert.Component, alert.Value, alert.Threshold)
}