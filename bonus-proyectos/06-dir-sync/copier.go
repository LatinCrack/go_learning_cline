package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CopyResult contiene el resultado de la operación de copia de un archivo individual.
type CopyResult struct {
	Decision FileDecision
	Err      error
	Duration time.Duration
}

// CopyStats contiene las estadísticas globales de la operación de copia.
type CopyStats struct {
	CopiedFiles  int
	FailedFiles  int
	TotalBytes   int64
	TotalTime    time.Duration
	Results      []CopyResult
}

// CopyFiles ejecuta la copia concurrente de archivos usando un pool de workers.
// Recibe las decisiones de copia, el directorio destino, la concurrencia máxima
// y retorna estadísticas de la operación.
func CopyFiles(decisions []FileDecision, dstRoot string, concurrency int) *CopyStats {
	if concurrency < 1 {
		concurrency = 1
	}

	stats := &CopyStats{
		Results: make([]CopyResult, 0, len(decisions)),
	}

	if len(decisions) == 0 {
		return stats
	}

	// Canal de trabajo alimentado por el productor.
	jobs := make(chan FileDecision, concurrency*2)
	// Canal de resultados recolectado por el consumidor.
	results := make(chan CopyResult, len(decisions))

	var wg sync.WaitGroup

	// Lanzar pool de workers.
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for decision := range jobs {
				result := copySingleFile(decision, dstRoot)
				results <- result
			}
		}(i)
	}

	// Productor: enviar trabajos al canal.
	go func() {
		for _, d := range decisions {
			jobs <- d
		}
		close(jobs)
	}()

	// Cerrar canal de resultados cuando todos los workers terminen.
	go func() {
		wg.Wait()
		close(results)
	}()

	// Consumir resultados.
	start := time.Now()
	for result := range results {
		stats.Results = append(stats.Results, result)
		if result.Err != nil {
			stats.FailedFiles++
		} else {
			stats.CopiedFiles++
			stats.TotalBytes += result.Decision.SrcFile.Size
		}
	}
	stats.TotalTime = time.Since(start)

	return stats
}

// copySingleFile copia un archivo individual preservando permisos.
// Crea los directorios padre necesarios si no existen.
func copySingleFile(decision FileDecision, dstRoot string) CopyResult {
	start := time.Now()
	result := CopyResult{
		Decision: decision,
	}

	dstPath := filepath.Join(dstRoot, decision.SrcFile.RelPath)

	// Crear directorios padre si no existen.
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		result.Err = fmt.Errorf("error creando directorio %q: %w", dstDir, err)
		result.Duration = time.Since(start)
		return result
	}

	if err := copyFile(decision.SrcFile.Path, dstPath, decision.SrcFile.Mode); err != nil {
		result.Err = fmt.Errorf("error copiando %q → %q: %w", decision.SrcFile.Path, dstPath, err)
		result.Duration = time.Since(start)
		return result
	}

	result.Duration = time.Since(start)
	return result
}

// copyFile realiza la copia byte-a-byte de un archivo usando io.Copy
// y preserva los permisos originales del archivo fuente.
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return fmt.Errorf("abriendo origen: %w", err)
	}
	defer srcFile.Close()

	// Usar archivo temporal para evitar archivos parciales en caso de error.
	tmpPath := dst + ".tmp"
	dstFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("creando destino temporal: %w", err)
	}

	// Copiar contenido usando io.Copy (streaming, sin cargar todo en memoria).
	written, err := io.Copy(dstFile, srcFile)
	if err != nil {
		dstFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("copiando contenido (%d bytes escritos): %w", written, err)
	}

	// Flush buffers al disco.
	if err := dstFile.Sync(); err != nil {
		dstFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sincronizando datos: %w", err)
	}

	if err := dstFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("cerrando archivo destino: %w", err)
	}

	// Renombrar archivo temporal al destino final (operación atómica).
	if err := os.Rename(tmpPath, dst); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renombrando temporal: %w", err)
	}

	// Asegurar que los permisos se apliquen correctamente.
	if err := os.Chmod(dst, mode); err != nil {
		return fmt.Errorf("estableciendo permisos: %w", err)
	}

	return nil
}