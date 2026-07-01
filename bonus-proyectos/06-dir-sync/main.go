package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

const (
	appName    = "dir-sync"
	appVersion = "1.0.0"
)

func main() {
	// ─── Definición de flags ────────────────────────────────────────────
	source := flag.String("source", "", "Ruta del directorio origen (requerido)")
	target := flag.String("target", "", "Ruta del directorio destino (requerido)")
	concurrency := flag.Int("concurrency", 4, "Número de workers de copia concurrentes")
	dryRun := flag.Bool("dry-run", false, "Solo mostrar qué se haría sin copiar archivos")
	verbose := flag.Bool("verbose", false, "Mostrar detalle de cada archivo analizado")
	showVersion := flag.Bool("version", false, "Mostrar versión y salir")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n%s v%s — Clonador de Directorios con Sincronización Incremental\n\n", appName, appVersion)
		fmt.Fprintf(os.Stderr, "Uso:\n")
		fmt.Fprintf(os.Stderr, "  %s --source <dir> --target <dir> [opciones]\n\n", appName)
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEjemplo:\n")
		fmt.Fprintf(os.Stderr, "  %s --source /var/log --target /backup/logs --concurrency 8\n", appName)
		fmt.Fprintf(os.Stderr, "  %s --source ./scripts --target ./backup/scripts --dry-run --verbose\n\n", appName)
	}

	flag.Parse()

	// ─── Validaciones ───────────────────────────────────────────────────
	if *showVersion {
		fmt.Printf("%s v%s\n", appName, appVersion)
		os.Exit(0)
	}

	if *source == "" || *target == "" {
		fmt.Fprintf(os.Stderr, "Error: --source y --target son requeridos.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if *concurrency < 1 {
		fmt.Fprintf(os.Stderr, "Error: --concurrency debe ser >= 1.\n")
		os.Exit(1)
	}

	fmt.Printf("\n🔄 %s v%s — Iniciando sincronización incremental...\n\n", appName, appVersion)

	totalStart := time.Now()

	// ─── Fase 1: Escaneo del directorio origen ─────────────────────────
	fmt.Printf("📂 Escaneando directorio origen: %s\n", *source)
	srcResult, err := ScanDirectory(*source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error escaneando origen: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   → %d archivos encontrados (%s)\n", len(srcResult.Files), formatBytes(srcResult.TotalSize))

	// ─── Fase 2: Escaneo del directorio destino ────────────────────────
	fmt.Printf("📂 Escaneando directorio destino: %s\n", *target)
	dstResult, err := ScanDirectory(*target)
	if err != nil {
		// Si el destino no existe, tratarlo como directorio vacío.
		fmt.Printf("   → Directorio destino no existe, se creará.\n")
		dstResult = &ScanResult{
			RootPath: *target,
			Files:    make([]FileInfo, 0),
			Errors:   make([]ScanError, 0),
		}
	} else {
		fmt.Printf("   → %d archivos encontrados (%s)\n", len(dstResult.Files), formatBytes(dstResult.TotalSize))
	}

	// ─── Fase 3: Comparación criptográfica ─────────────────────────────
	fmt.Printf("\n🔍 Comparando archivos (SHA-256 streaming)...\n")
	compareResult := CompareDirectories(srcResult, dstResult)
	fmt.Printf("   → %d idénticos, %d para copiar, %d errores\n",
		len(compareResult.ToSkip), len(compareResult.ToCopy), len(compareResult.Errors))

	// Modo verbose: mostrar detalle de cada archivo.
	if *verbose {
		PrintFileList(compareResult.Decisions, "Detalle de decisiones")
	}

	// ─── Fase 4: Copia concurrente ─────────────────────────────────────
	var copyStats *CopyStats
	if len(compareResult.ToCopy) > 0 {
		if *dryRun {
			fmt.Printf("\n🏜️  Modo dry-run: no se copiarán archivos.\n")
		} else {
			fmt.Printf("\n📋 Copiando %d archivos con %d workers...\n", len(compareResult.ToCopy), *concurrency)
			copyStats = CopyFiles(compareResult.ToCopy, *target, *concurrency)
		}
	} else {
		fmt.Printf("\n✅ Todos los archivos están sincronizados. Nada que copiar.\n")
	}

	totalTime := time.Since(totalStart)

	// ─── Fase 5: Reporte ───────────────────────────────────────────────
	report := &SyncReport{
		SourcePath:  *source,
		TargetPath:  *target,
		Compare:     compareResult,
		CopyStats:   copyStats,
		TotalTime:   totalTime,
		Concurrency: *concurrency,
	}

	PrintConsoleReport(report)

	// Generar log de texto (excepto en dry-run).
	if !*dryRun {
		logPath, err := WriteLogFile(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Error escribiendo log: %v\n", err)
		} else {
			fmt.Printf("📝 Log detallado: %s\n\n", logPath)
		}
	}

	// Reportar errores de escaneo.
	if len(srcResult.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "⚠️  %d errores durante el escaneo del origen:\n", len(srcResult.Errors))
		for _, e := range srcResult.Errors {
			fmt.Fprintf(os.Stderr, "   - %s: %v\n", e.Path, e.Err)
		}
	}

	if len(compareResult.Errors) > 0 {
		os.Exit(2)
	}
	if copyStats != nil && copyStats.FailedFiles > 0 {
		os.Exit(3)
	}
}