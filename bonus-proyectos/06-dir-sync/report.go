package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SyncReport encapsula todos los datos necesarios para generar el reporte final.
type SyncReport struct {
	SourcePath  string
	TargetPath  string
	Compare     *CompareResult
	CopyStats   *CopyStats
	TotalTime   time.Duration
	Concurrency int
}

// PrintConsoleReport imprime un resumen limpio y estructurado en la consola.
func PrintConsoleReport(report *SyncReport) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║            📋  REPORTE DE SINCRONIZACIÓN INCREMENTAL           ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Información general.
	fmt.Printf("  📂 Origen:       %s\n", report.SourcePath)
	fmt.Printf("  📂 Destino:      %s\n", report.TargetPath)
	fmt.Printf("  ⚙️  Concurrencia: %d workers\n", report.Concurrency)
	fmt.Println()

	// Estadísticas de escaneo.
	fmt.Println("  ── Escaneo ──────────────────────────────────────────────────────")
	totalFiles := len(report.Compare.Decisions)
	fmt.Printf("  📊 Total archivos analizados: %d\n", totalFiles)
	fmt.Printf("  📊 Tamaño total origen:      %s\n", formatBytes(report.Compare.TotalSrcSize))
	if len(report.Compare.Errors) > 0 {
		fmt.Printf("  ⚠️  Errores durante análisis: %d\n", len(report.Compare.Errors))
	}
	fmt.Println()

	// Decisiones de comparación.
	fmt.Println("  ── Comparación ──────────────────────────────────────────────────")
	fmt.Printf("  ✅ Archivos idénticos (omitidos):  %d  (%s ahorrados)\n",
		len(report.Compare.ToSkip), formatBytes(report.Compare.SavedSize))
	fmt.Printf("  📋 Archivos a copiar (nuevos/mod): %d\n", len(report.Compare.ToCopy))
	fmt.Println()

	// Detalle de la copia.
	if report.CopyStats != nil {
		fmt.Println("  ── Copia ────────────────────────────────────────────────────────")
		fmt.Printf("  📤 Archivos copiados exitosamente: %d\n", report.CopyStats.CopiedFiles)
		if report.CopyStats.FailedFiles > 0 {
			fmt.Printf("  ❌ Archivos con error:            %d\n", report.CopyStats.FailedFiles)
		}
		fmt.Printf("  📦 Total bytes transferidos:       %s\n", formatBytes(report.CopyStats.TotalBytes))
		fmt.Printf("  ⏱️  Tiempo de copia:               %s\n", report.CopyStats.TotalTime.Round(time.Millisecond))
		fmt.Println()
	}

	// Resumen de eficiencia.
	fmt.Println("  ── Eficiencia ───────────────────────────────────────────────────")
	fmt.Printf("  ⏱️  Tiempo total:                   %s\n", report.TotalTime.Round(time.Millisecond))
	if report.Compare.SavedSize > 0 {
		fmt.Printf("  💾 Ancho de banda ahorrado:        %s\n", formatBytes(report.Compare.SavedSize))
		savedPercent := float64(report.Compare.SavedSize) / float64(report.Compare.TotalSrcSize) * 100
		fmt.Printf("  📈 Eficiencia incremental:         %.1f%%\n", savedPercent)
	}
	fmt.Println()

	// Errores si los hay.
	if len(report.Compare.Errors) > 0 {
		fmt.Println("  ── Errores ──────────────────────────────────────────────────────")
		for _, e := range report.Compare.Errors {
			fmt.Printf("  ⚠️  %s: %s\n", e.SrcFile.RelPath, e.Reason)
		}
		fmt.Println()
	}

	if report.CopyStats != nil {
		for _, r := range report.CopyStats.Results {
			if r.Err != nil {
				fmt.Printf("  ❌ %s: %v\n", r.Decision.SrcFile.RelPath, r.Err)
			}
		}
		if report.CopyStats.FailedFiles > 0 {
			fmt.Println()
		}
	}

	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ✅  SINCRONIZACIÓN COMPLETADA                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// WriteLogFile genera un archivo de log detallado con todas las acciones tomadas.
// El archivo se crea en el directorio destino con timestamp en el nombre.
func WriteLogFile(report *SyncReport) (string, error) {
	logDir := filepath.Join(report.TargetPath, ".dir-sync")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", fmt.Errorf("error creando directorio de logs: %w", err)
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("sync-%s.log", time.Now().Format("2006-01-02-150405")))
	f, err := os.Create(logFile)
	if err != nil {
		return "", fmt.Errorf("error creando archivo de log: %w", err)
	}
	defer f.Close()

	w := func(format string, args ...interface{}) {
		fmt.Fprintf(f, format+"\n", args...)
	}

	w("═══════════════════════════════════════════════════════════════")
	w("  DIR-SYNC — Log de Sincronización Incremental")
	w("═══════════════════════════════════════════════════════════════")
	w("")
	w("Fecha:         %s", time.Now().Format("2006-01-02 15:04:05"))
	w("Origen:        %s", report.SourcePath)
	w("Destino:       %s", report.TargetPath)
	w("Concurrencia:  %d workers", report.Concurrency)
	w("")

	// Archivos omitidos.
	w("── ARCHIVOS OMITIDOS (sin cambios) ──────────────────────────")
	if len(report.Compare.ToSkip) == 0 {
		w("  (ninguno)")
	} else {
		for _, d := range report.Compare.ToSkip {
			w("  [SKIP] %s — %s", d.SrcFile.RelPath, d.Reason)
		}
	}
	w("  Total omitidos: %d", len(report.Compare.ToSkip))
	w("")

	// Archivos copiados.
	w("── ARCHIVOS COPIADOS ────────────────────────────────────────")
	if report.CopyStats == nil || len(report.CopyStats.Results) == 0 {
		if len(report.Compare.ToCopy) == 0 {
			w("  (ninguno)")
		} else {
			for _, d := range report.Compare.ToCopy {
				w("  [COPY] %s — %s (%d bytes)", d.SrcFile.RelPath, d.Reason, d.SrcFile.Size)
			}
		}
	} else {
		for _, r := range report.CopyStats.Results {
			status := "OK"
			if r.Err != nil {
				status = fmt.Sprintf("ERROR: %v", r.Err)
			}
			w("  [COPY] %s — %s (%d bytes) [%s] [%s]",
				r.Decision.SrcFile.RelPath, r.Decision.Reason,
				r.Decision.SrcFile.Size, r.Duration.Round(time.Millisecond), status)
		}
	}
	w("")

	// Errores.
	w("── ERRORES ──────────────────────────────────────────────────")
	if len(report.Compare.Errors) == 0 {
		w("  (ninguno)")
	} else {
		for _, e := range report.Compare.Errors {
			w("  [ERROR] %s — %s", e.SrcFile.RelPath, e.Reason)
		}
	}
	w("")

	// Estadísticas.
	w("── RESUMEN ──────────────────────────────────────────────────")
	w("  Total archivos analizados: %d", len(report.Compare.Decisions))
	w("  Tamaño total origen:      %s", formatBytes(report.Compare.TotalSrcSize))
	w("  Archivos omitidos:        %d", len(report.Compare.ToSkip))
	w("  Archivos copiados:        %d", len(report.Compare.ToCopy))
	w("  Archivos con error:       %d", len(report.Compare.Errors))
	w("  Ancho de banda ahorrado:  %s", formatBytes(report.Compare.SavedSize))
	if report.CopyStats != nil {
		w("  Bytes transferidos:       %s", formatBytes(report.CopyStats.TotalBytes))
	}
	w("  Tiempo total:             %s", report.TotalTime.Round(time.Millisecond))
	w("")
	w("═══════════════════════════════════════════════════════════════")
	w("  FIN DEL LOG")
	w("═══════════════════════════════════════════════════════════════")

	return logFile, nil
}

// formatBytes convierte bytes a una representación legible.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// PrintFileList imprime una lista tabulada de archivos con su acción asignada.
func PrintFileList(decisions []FileDecision, title string) {
	if len(decisions) == 0 {
		return
	}

	fmt.Printf("\n  %s:\n", title)
	fmt.Println("  " + strings.Repeat("─", 66))

	for _, d := range decisions {
		icon := "📋"
		switch d.Action {
		case ActionSkip:
			icon = "✅"
		case ActionError:
			icon = "⚠️"
		}
		fmt.Printf("  %s %-48s %s\n", icon, d.SrcFile.RelPath, formatBytes(d.SrcFile.Size))
	}
}