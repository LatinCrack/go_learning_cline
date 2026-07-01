package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const banner = `
  _                _     _                        _                
 | |    ___   __ _(_)___| |_ _ __ ___   __ _ _ __ (_) _____      __
 | |   / _ \ / _' | / __| __| '_ ' _ \ / _' | '_ \| |/ _ \ \ /\ / /
 | |__| (_) | (_| | \__ \ |_| | | | | | (_| | | | | | (_) \ V  V / 
 |_____\___/ \__, |_|___/\__|_| |_| |_|\__,_|_| |_|_|\___/ \_/\_/  
             |___/                                                   
  Real-Time Security Log Analyzer v1.0
`

func main() {
	// ── CLI Flags ──────────────────────────────────────────────
	logFile := flag.String("file", "", "Path to the log file to monitor [required]")
	workers := flag.Int("workers", 4, "Number of concurrent detection workers")
	pollInterval := flag.Duration("poll", 500*time.Millisecond, "Poll interval for tail (e.g., 100ms, 500ms, 1s)")
	minSeverity := flag.String("severity", "low", "Minimum severity to alert: low, medium, high, critical")
	showRaw := flag.Bool("raw", true, "Show raw log line in alerts")
	showVersion := flag.Bool("version", false, "Show version information")
	showPatterns := flag.Bool("patterns", false, "List all detection patterns and exit")
	readExisting := flag.Bool("existing", false, "Also analyze existing file content before tailing")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, banner)
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  log-analyzer -file <logfile> [options]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  log-analyzer -file /var/log/apache2/access.log")
		fmt.Fprintln(os.Stderr, "  log-analyzer -file access.log -workers 8 -severity high")
		fmt.Fprintln(os.Stderr, "  log-analyzer -file access.log -existing -poll 100ms")
		fmt.Fprintln(os.Stderr, "  log-analyzer -file access.log -raw=false -workers 2")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// ── Handle version flag ────────────────────────────────────
	if *showVersion {
		fmt.Println("log-analyzer v1.0.0")
		fmt.Print("Real-Time Security Log Analyzer — Go Infrastructure Project\n")
		os.Exit(0)
	}

	// ── Handle patterns flag ───────────────────────────────────
	if *showPatterns {
		stats := NewStats()
		engine := NewDetectorEngine(DetectorConfig{Workers: 1}, stats)
		fmt.Print("\n  📋 Patrones de detección activos:\n\n")
		for i, p := range engine.GetPatterns() {
			fmt.Printf("  %2d. [%-8s] %-25s %s\n", i+1, p.Severity.String(), p.Name, p.Description)
		}
		fmt.Println()
		os.Exit(0)
	}

	// ── Validate required flags ────────────────────────────────
	if *logFile == "" {
		fmt.Fprintln(os.Stderr, "Error: -file flag is required.")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	// ── Validate file exists ───────────────────────────────────
	info, err := os.Stat(*logFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot access log file '%s': %v\n", *logFile, err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: '%s' is a directory, not a file\n", *logFile)
		os.Exit(1)
	}

	// ── Parse severity ─────────────────────────────────────────
	sev, err := parseSeverity(*minSeverity)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// ── Validate workers ───────────────────────────────────────
	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "Error: workers must be at least 1")
		os.Exit(1)
	}
	if *workers > 64 {
		fmt.Fprintln(os.Stderr, "Warning: capping workers at 64 for log analysis")
		*workers = 64
	}

	// ── Print banner ───────────────────────────────────────────
	fmt.Fprint(os.Stderr, banner)
	fmt.Fprintf(os.Stderr, "  Archivo:     %s\n", *logFile)
	fmt.Fprintf(os.Stderr, "  Workers:     %d\n", *workers)
	fmt.Fprintf(os.Stderr, "  Poll:        %s\n", *pollInterval)
	fmt.Fprintf(os.Stderr, "  Severidad:   ≥ %s\n", sev.String())
	fmt.Fprintf(os.Stderr, "  Modo:        tail -f (monitoreo en tiempo real)\n")
	if *readExisting {
		fmt.Fprintf(os.Stderr, "  Existentes:  sí (analizando contenido previo)\n")
	}
	fmt.Fprintf(os.Stderr, "\n  🔍 Monitoreando... (Ctrl+C para detener)\n\n")

	// ── Initialize components ──────────────────────────────────
	stats := NewStats()

	detectorCfg := DetectorConfig{
		Workers:    *workers,
		ChannelBuf: 1000,
	}
	engine := NewDetectorEngine(detectorCfg, stats)

	alerterCfg := AlerterConfig{
		MinSeverity: sev,
		ShowRaw:     *showRaw,
		Writer:      os.Stdout,
	}
	alerter := NewAlerter(alerterCfg)

	// ── Create channels ────────────────────────────────────────
	linesChan := make(chan string, detectorCfg.ChannelBuf)
	alertsChan := make(chan ThreatAlert, detectorCfg.ChannelBuf)
	var workersWg sync.WaitGroup

	// ── Start alerter (consumes alerts channel) ────────────────
	alerterDone := make(chan struct{})
	go alerter.Start(alertsChan, alerterDone)

	// ── Start detector workers (consume lines, produce alerts) ─
	engine.Start(linesChan, alertsChan, &workersWg)

	// ── Start signal handler for graceful shutdown ─────────────
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine: cerrar canales cuando los workers terminen
	go func() {
		workersWg.Wait()
		close(alertsChan)
	}()

	// ── Read existing content (optional) ───────────────────────
	if *readExisting {
		existingChan := make(chan string, 1000)
		go func() {
			if err := ReadExistingLines(*logFile, existingChan); err != nil {
				fmt.Fprintf(os.Stderr, "  ⚠ Error leyendo contenido existente: %v\n", err)
			}
			close(existingChan)
		}()
		for line := range existingChan {
			stats.IncrementLines()
			linesChan <- line
		}
	}

	// ── Start tailing (producer) ───────────────────────────────
	tailDone := make(chan struct{})
	go func() {
		if err := TailFile(*logFile, linesChan, tailDone, *pollInterval); err != nil {
			fmt.Fprintf(os.Stderr, "  ⚠ Error en tail: %v\n", err)
		}
		// Cuando TailFile termina (por shutdown signal), cerramos el canal
		// de líneas para que los workers completen su procesamiento
		close(linesChan)
	}()

	// ── Wait for shutdown signal ───────────────────────────────
	sig := <-sigChan
	fmt.Fprintf(os.Stderr, "\n\n  ⚡ Señal recibida (%v). Iniciando shutdown graceful...\n", sig)

	// Señalizar a TailFile que pare
	close(tailDone)

	// Esperar a que el alerter termine de procesar alertas pendientes
	<-alerterDone

	// ── Print final summary ────────────────────────────────────
	alerter.PrintSummary(stats, os.Stderr)

	fmt.Fprintln(os.Stderr, "  ✅ Análisis finalizado.")
}

// parseSeverity convierte un string de severidad a su valor Severity.
func parseSeverity(s string) (Severity, error) {
	switch s {
	case "low", "LOW", "Low":
		return SeverityLow, nil
	case "medium", "MEDIUM", "Medium":
		return SeverityMedium, nil
	case "high", "HIGH", "High":
		return SeverityHigh, nil
	case "critical", "CRITICAL", "Critical":
		return SeverityCritical, nil
	default:
		return SeverityLow, fmt.Errorf("severidad inválida '%s'. Usa: low, medium, high, critical", s)
	}
}
