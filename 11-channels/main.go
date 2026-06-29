package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// =============================================================
// Pipeline de Descarga y Procesamiento de Archivos
// Ejercicio práctico — Lección 11: Channels
// =============================================================

// DownloadTask representa una tarea de descarga
type DownloadTask struct {
	ID  int
	URL string
}

// DownloadResult representa el resultado de una descarga
type DownloadResult struct {
	TaskID    int
	URL       string
	Success   bool
	SizeBytes int
	HashSHA   string
	Latency   time.Duration
	Error     string
}

// =============================================================
// ETAPA 1: Generador de URLs (Productor)
// =============================================================
// generaURLs emite URLs al channel de tareas.
// Cuando termina, CIERRA el channel para señalar que no hay más trabajo.
func generaURLs(urls []string) <-chan DownloadTask {
	out := make(chan DownloadTask) // Channel unbuffered: sincronizado

	go func() {
		for i, url := range urls {
			out <- DownloadTask{
				ID:  i + 1,
				URL: url,
			}
		}
		close(out) // 🔑 Señal: "no hay más tareas"
	}()

	return out // Devolvemos el channel de solo-lectura
}

// =============================================================
// ETAPA 2: Downloader Concurrente (Worker)
// =============================================================
// descargaURLs recibe tareas, las descarga con concurrencia limitada,
// y envía los resultados al channel de salida.
func descargaURLs(
	tasks <-chan DownloadTask, // Entrada: solo lectura
	maxConcurrency int, // Límite de goroutines simultáneas
) <-chan DownloadResult {
	out := make(chan DownloadResult)

	var wg sync.WaitGroup

	// Lanzamos N workers (goroutines)
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Cada worker consume tareas del mismo channel
			for task := range tasks {
				result := descargarArchivo(task)
				out <- result
			}
		}(i + 1)
	}

	// Cerramos el channel de salida cuando TODOS los workers terminen
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// descargarArchivo realiza la descarga HTTP real de una URL
func descargarArchivo(task DownloadTask) DownloadResult {
	inicio := time.Now()

	result := DownloadResult{
		TaskID: task.ID,
		URL:    task.URL,
	}

	// Realizamos la petición HTTP con timeout
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(task.URL)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Latency = time.Since(inicio)
		return result
	}
	defer resp.Body.Close()

	// Leemos el cuerpo de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Latency = time.Since(inicio)
		return result
	}

	// Calculamos el hash SHA-256 del contenido
	hash := sha256.Sum256(body)

	result.Success = true
	result.SizeBytes = len(body)
	result.HashSHA = fmt.Sprintf("%x", hash[:8]) // Primeros 8 bytes del hash
	result.Latency = time.Since(inicio)

	return result
}

// =============================================================
// ETAPA 3: Procesador de Resultados (Consumidor)
// =============================================================
// procesarResultos recoge todos los resultados y genera un reporte.
func procesarResultos(results <-chan DownloadResult) []DownloadResult {
	var todos []DownloadResult

	for result := range results { // Se desbloquea cuando el channel se cierra
		todos = append(todos, result)

		// Mostramos progreso en tiempo real
		estado := "✅"
		if !result.Success {
			estado = "❌"
		}
		fmt.Printf("  %s [%d] %-50s %v\n", estado, result.TaskID, truncarURL(result.URL, 50), result.Latency)
	}

	return todos
}

// truncarURL acorta una URL para que quepa en la tabla
func truncarURL(url string, maxLen int) string {
	if len(url) <= maxLen {
		return url
	}
	return url[:maxLen-3] + "..."
}

// =============================================================
// FUNCIÓN PRINCIPAL: Orquesta el Pipeline
// =============================================================
func main() {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("   🔄 Pipeline de Descarga y Procesamiento — Lección 11")
	fmt.Println("   📡 Channels como sistema nervioso de Go")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// URLs de prueba (endpoints públicos que responden rápido)
	urls := []string{
		"https://httpbin.org/get",
		"https://httpbin.org/ip",
		"https://httpbin.org/user-agent",
		"https://httpbin.org/headers",
		"https://httpbin.org/uuid",
		"https://httpbin.org/base64/SExPIEdv",
		"https://httpbin.org/bytes/256",
		"https://httpbin.org/delay/1",
		"https://httpbin.org/status/200",
		"https://httpbin.org/gzip",
	}

	maxWorkers := 3 // Concurrencia máxima

	fmt.Printf("📋 Tareas: %d URLs\n", len(urls))
	fmt.Printf("⚡ Workers: %d goroutines concurrentes\n", maxWorkers)
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("🔄 Descargando...")
	fmt.Println()

	// ┌─────────────┐    ┌─────────────┐    ┌──────────────┐
	// │  ETAPA 1:   │───▶│  ETAPA 2:   │───▶│  ETAPA 3:    │
	// │  Generador  │    │  Downloaders│    │  Procesador  │
	// │  (1 gorout) │    │  (N gorouts)│    │  (1 gorout)  │
	// └─────────────┘    └─────────────┘    └──────────────┘

	inicio := time.Now()

	// Etapa 1: Generar tareas
	tasks := generaURLs(urls)

	// Etapa 2: Descargar con concurrencia limitada
	results := descargaURLs(tasks, maxWorkers)

	// Etapa 3: Procesar y mostrar resultados
	resultados := procesarResultos(results)

	duracion := time.Since(inicio)

	// ─── Reporte Final ───
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("   📊 REPORTE FINAL")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	exitosos := 0
	fallidos := 0
	totalBytes := 0

	for _, r := range resultados {
		if r.Success {
			exitosos++
			totalBytes += r.SizeBytes
		} else {
			fallidos++
		}
	}

	fmt.Printf("   ✅ Exitosos:  %d\n", exitosos)
	fmt.Printf("   ❌ Fallidos:  %d\n", fallidos)
	fmt.Printf("   📦 Total:     %d bytes (%.2f KB)\n", totalBytes, float64(totalBytes)/1024)
	fmt.Printf("   ⏱️  Duración:  %v\n", duracion)
	fmt.Println()

	// Tabla detallada de resultados exitosos
	resultadosOK := make([]DownloadResult, 0)
	for _, r := range resultados {
		if r.Success {
			resultadosOK = append(resultadosOK, r)
		}
	}

	if len(resultadosOK) > 0 {
		// Ordenar por latencia
		sort.Slice(resultadosOK, func(i, j int) bool {
			return resultadosOK[i].Latency < resultadosOK[j].Latency
		})

		fmt.Println("   📋 Detalle de descargas exitosas (ordenadas por latencia):")
		fmt.Println()
		fmt.Printf("   %-5s %-8s %-16s %-52s\n", "TAREA", "TAMAÑO", "LATENCIA", "HASH SHA-256")
		fmt.Println("   " + strings.Repeat("─", 85))

		for _, r := range resultadosOK {
			tamano := fmt.Sprintf("%.1f KB", float64(r.SizeBytes)/1024)
			fmt.Printf("   #%-4d %-8s %-16v %-52s\n", r.TaskID, tamano, r.Latency, r.HashSHA)
		}
	}

	// Mostrar errores si los hay
	for _, r := range resultados {
		if !r.Success {
			fmt.Printf("\n   ❌ Error en #%d (%s): %s\n", r.TaskID, r.URL, r.Error)
		}
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("   🧠 Lección: Los channels sincronizan las 3 etapas del")
	fmt.Println("   pipeline sin Mutex ni WaitGroup. El close() es la señal.")
	fmt.Println("═══════════════════════════════════════════════════════════════")
}