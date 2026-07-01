package main

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"
)

const resetColor = "\033[0m"

// AlerterConfig contiene la configuración del sistema de alertas.
type AlerterConfig struct {
	MinSeverity Severity  // Severidad mínima para mostrar alerta
	ShowRaw     bool      // Mostrar línea cruda del log
	Writer      io.Writer // Destino de salida (os.Stdout por defecto)
}

// Alerter es el sistema de alertas que recibe amenazas detectadas
// y las presenta en consola con formato visual y por severidad.
type Alerter struct {
	config    AlerterConfig
	counter   atomic.Int64 // Contador total de alertas mostradas
	startTime time.Time
}

// NewAlerter crea una nueva instancia del sistema de alertas.
func NewAlerter(config AlerterConfig) *Alerter {
	if config.Writer == nil {
		config.Writer = io.Discard
	}
	return &Alerter{
		config:    config,
		startTime: time.Now(),
	}
}

// Start inicia la escucha del canal de alertas y las procesa
// en el horno principal de presentación.
func (a *Alerter) Start(alerts <-chan ThreatAlert, done chan<- struct{}) {
	for alert := range alerts {
		if alert.Severity >= a.config.MinSeverity {
			a.printAlert(alert)
			a.counter.Add(1)
		}
	}
	// Canal cerrado: señalizar que terminamos
	close(done)
}

// printAlert imprime una alerta formateada en consola con colores ANSI.
func (a *Alerter) printAlert(alert ThreatAlert) {
	w := a.config.Writer
	color := alert.Severity.ColorCode()

	// Cabecera de la alerta
	fmt.Fprintf(w, "\n%s╔══════════════════════════════════════════════════════════════╗%s\n", color, resetColor)
	fmt.Fprintf(w, "%s║  🚨 ALERTA DE SEGURIDAD DETECTADA                           ║%s\n", color, resetColor)
	fmt.Fprintf(w, "%s╚══════════════════════════════════════════════════════════════╝%s\n", color, resetColor)

	// Detalles de la alerta
	fmt.Fprintf(w, "  %-14s %s[%s]%s %s\n",
		"Severidad:",
		color, alert.Severity.String(), resetColor,
		severityLabel(alert.Severity),
	)
	fmt.Fprintf(w, "  %-14s %s\n", "Patrón:", alert.Pattern)
	fmt.Fprintf(w, "  %-14s %s\n", "Coincidencia:", alert.Match)

	if alert.IP != "" {
		fmt.Fprintf(w, "  %-14s %s\n", "IP Origen:", alert.IP)
	}
	if alert.Path != "" {
		fmt.Fprintf(w, "  %-14s %s\n", "Ruta:", alert.Path)
	}
	fmt.Fprintf(w, "  %-14s %d\n", "Línea:", alert.LineNum)

	// Línea cruda (opcional)
	if a.config.ShowRaw {
		// Truncar líneas muy largas
		raw := alert.Line
		if len(raw) > 120 {
			raw = raw[:120] + "..."
		}
		fmt.Fprintf(w, "  %-14s %s\n", "Log:", raw)
	}

	fmt.Fprintf(w, "  %s\n", strings.Repeat("─", 60))
}

// severityLabel retorna una descripción legible de la severidad.
func severityLabel(s Severity) string {
	switch s {
	case SeverityLow:
		return "Actividad sospechosa menor"
	case SeverityMedium:
		return "Requiere monitoreo activo"
	case SeverityHigh:
		return "Amenaza activa — investigar"
	case SeverityCritical:
		return "ALERTA CRÍTICA — acción inmediata"
	default:
		return ""
	}
}

// GetAlertCount retorna el total de alertas mostradas.
func (a *Alerter) GetAlertCount() int64 {
	return a.counter.Load()
}

// PrintSummary imprime un resumen final del análisis de seguridad.
func (a *Alerter) PrintSummary(stats *Stats, writer io.Writer) {
	duration := time.Since(a.startTime)
	snapshot := stats.Snapshot()

	fmt.Fprintf(writer, "\n%s", strings.Repeat("═", 64))
	fmt.Fprintf(writer, "\n  📊 RESUMEN DEL ANÁLISIS DE SEGURIDAD\n")
	fmt.Fprintf(writer, "%s\n", strings.Repeat("═", 64))

	fmt.Fprintf(writer, "  %-24s %s\n", "Duración:", duration.Round(time.Millisecond))
	fmt.Fprintf(writer, "  %-24s %d\n", "Total de líneas:", snapshot.TotalLines)
	fmt.Fprintf(writer, "  %-24s %d\n", "Total de alertas:", a.counter.Load())
	fmt.Fprintf(writer, "  %-24s %d\n", "IPs únicas involucradas:", snapshot.UniqueIPs)

	// Distribución por severidad
	fmt.Fprintf(writer, "\n  📈 Distribución por severidad:\n")
	fmt.Fprintf(writer, "  %-18s %s\n", "CRITICAL:", severityBar(snapshot.BySeverity[SeverityCritical], snapshot.TotalAttacks))
	fmt.Fprintf(writer, "  %-18s %s\n", "HIGH:", severityBar(snapshot.BySeverity[SeverityHigh], snapshot.TotalAttacks))
	fmt.Fprintf(writer, "  %-18s %s\n", "MEDIUM:", severityBar(snapshot.BySeverity[SeverityMedium], snapshot.TotalAttacks))
	fmt.Fprintf(writer, "  %-18s %s\n", "LOW:", severityBar(snapshot.BySeverity[SeverityLow], snapshot.TotalAttacks))

	// Top ataques
	if len(snapshot.TopAttacks) > 0 {
		fmt.Fprintf(writer, "\n  🎯 Top ataques detectados:\n")
		limit := 5
		if len(snapshot.TopAttacks) < limit {
			limit = len(snapshot.TopAttacks)
		}
		for i := 0; i < limit; i++ {
			attack := snapshot.TopAttacks[i]
			fmt.Fprintf(writer, "  %d. %-30s %d detecciones\n", i+1, attack.Name, attack.Count)
		}
	}

	// Top IPs atacantes
	if len(snapshot.TopIPs) > 0 {
		fmt.Fprintf(writer, "\n  🌐 Top IPs atacantes:\n")
		limit := 5
		if len(snapshot.TopIPs) < limit {
			limit = len(snapshot.TopIPs)
		}
		for i := 0; i < limit; i++ {
			ip := snapshot.TopIPs[i]
			fmt.Fprintf(writer, "  %d. %-30s %d ataques\n", i+1, ip.IP, ip.Count)
		}
	}

	fmt.Fprintf(writer, "\n%s\n", strings.Repeat("═", 64))
}

// severityBar genera una barra visual proporcional para el resumen.
func severityBar(count, total int64) string {
	if total == 0 {
		return "0"
	}
	barLen := int(count * 20 / total)
	if barLen == 0 && count > 0 {
		barLen = 1
	}
	bar := strings.Repeat("█", barLen)
	return fmt.Sprintf("%-20s %d", bar, count)
}
