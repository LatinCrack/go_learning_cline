package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// ═══════════════════════════════════════════════════════════════
//  🍕 Crawler Web Concurrente con Rate Limiting
//  Lección 13 — Patrones Avanzados de Concurrencia
// ═══════════════════════════════════════════════════════════════

// ─────────────────────────────────────────────────────────────
//  Tipos y Constantes
// ─────────────────────────────────────────────────────────────

// PaginaWeb simula una página descargada del internet
type PaginaWeb struct {
	URL        string
	Contenido  string
	Tamano     int        // Tamaño en bytes (simulado)
	Latencia   time.Duration
	StatusCode int
	WorkerID   int        // Qué worker la procesó
	Error      string
}

// ResultadoCrawl agrupa la página procesada con su profundidad
type ResultadoCrawl struct {
	Pagina     PaginaWeb
	Profundidad int
	Procesado   time.Time
}

// Estadisticas del crawler (usando atomic para seguridad concurrente)
type EstadisticasCrawl struct {
	PaginasDescargadas int64
	PaginasError       int64
	BytesTotal         int64
	TiempoTotal        time.Duration
}

// URLTrabajo representa un trabajo pendiente para el worker pool
type URLTrabajo struct {
	URL        string
	Profundidad int
}

// ─────────────────────────────────────────────────────────────
//  1️⃣  FAN-OUT: Distribuir trabajo a múltiples goroutines
// ─────────────────────────────────────────────────────────────

// fanOut lanza N goroutines que leen del mismo channel de trabajos.
// Cada goroutine "toma" un trabajo del canal — el runtime distribuye
// las tareas entre los workers automáticamente.
//
// ANLOGÍA: Como un gerente que pone 100 tareas en una mesa compartida,
// y 10 empleados toman la siguiente tarea libre cada vez que terminan.
func fanOut(
	trabajos <-chan URLTrabajo,
	numWorkers int,
) []chan PaginaWeb {
	// Creamos un channel de resultados por cada worker
	resultados := make([]chan PaginaWeb, numWorkers)

	for i := 0; i < numWorkers; i++ {
		resultados[i] = make(chan PaginaWeb)
	}

	for i := 0; i < numWorkers; i++ {
		workerID := i + 1
		resultadoCh := resultados[i]

		go func() {
			defer close(resultadoCh)

			// Cada worker toma trabajos del channel compartido
			for trabajo := range trabajos {
				pagina := descargarPaginaSimulada(workerID, trabajo.URL)
				resultadoCh <- pagina
			}
		}()
	}

	return resultados
}

// ─────────────────────────────────────────────────────────────
//  2️⃣  FAN-IN: Converger múltiples channels en uno solo
// ─────────────────────────────────────────────────────────────

// fanIn toma N channels de entrada y los converge en un solo channel
// de salida. Usa un WaitGroup para saber cuándo cerrar el output.
//
// ANALOGÍA: Como un embudo que junta los reportes de 10 empleados
// en un solo escritorio donde el gerente los lee todos.
func fanIn(
	canalesEntrada []chan PaginaWeb,
) <-chan PaginaWeb {
	canalUnificado := make(chan PaginaWeb)
	var wg sync.WaitGroup

	// Por cada canal de entrada, lanzamos una goroutine que lee
	// y copia al canal unificado
	for _, canal := range canalesEntrada {
		wg.Add(1)
		go func(c <-chan PaginaWeb) {
			defer wg.Done()
			for pagina := range c {
				canalUnificado <- pagina
			}
		}(canal)
	}

	// Cuando TODOS los canales de entrada se cierran,
	// cerramos el canal unificado
	go func() {
		wg.Wait()
		close(canalUnificado)
	}()

	return canalUnificado
}

// ─────────────────────────────────────────────────────────────
//  3️⃣  WORKER POOL: Limitar concurrencia activa
// ─────────────────────────────────────────────────────────────

// workerPool ejecuta trabajos con un número FIJO de workers.
// Si llegan más trabajos que workers, los trabajos esperan en el canal.
//
// ANALOGÍA: Un call center con exactamente 5 operadores.
// Si llaman 100 personas, 95 esperan en cola.
// Los 5 operadores atienden llamadas secuencialmente.
func workerPool(
	numWorkers int,
	trabajos <-chan URLTrabajo,
	estadisticas *EstadisticasCrawl,
) <-chan ResultadoCrawl {
	resultados := make(chan ResultadoCrawl)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		workerID := i + 1

		go func() {
			defer wg.Done()

			for trabajo := range trabajos {
				inicio := time.Now()

				// Simulamos la descarga
				pagina := descargarPaginaSimulada(workerID, trabajo.URL)
				latencia := time.Since(inicio)

				pagina.Latencia = latencia

				// Actualizamos estadísticas de forma segura (atomic)
				if pagina.Error != "" {
					atomic.AddInt64(&estadisticas.PaginasError, 1)
				} else {
					atomic.AddInt64(&estadisticas.PaginasDescargadas, 1)
					atomic.AddInt64(&estadisticas.BytesTotal, int64(pagina.Tamano))
				}

				resultados <- ResultadoCrawl{
					Pagina:      pagina,
					Profundidad: trabajo.Profundidad,
					Procesado:   time.Now(),
				}
			}
		}()
	}

	// Cerramos el canal de resultados cuando todos los workers terminen
	go func() {
		wg.Wait()
		close(resultados)
	}()

	return resultados
}

// ─────────────────────────────────────────────────────────────
//  4️⃣  SEMAPHORE: Controlar acceso a recursos limitados
// ─────────────────────────────────────────────────────────────

// Semaphore limita cuántas goroutines acceden a un recurso simultáneamente.
// Se implementa como un channel buffered: solo N goroutines pueden tener
// un "token" al mismo tiempo.
//
// ANALOGÍA: Como los casilleros de un gimnasio. Solo hay 10 casilleros.
// Si los 10 están ocupados, el siguiente cliente espera a que uno se libere.
type Semaphore struct {
	tokens chan struct{}
}

// NuevoSemaphore crea un semáforo con N "permisos" disponibles
func NuevoSemaphore(n int) *Semaphore {
	return &Semaphore{
		tokens: make(chan struct{}, n),
	}
}

// Adquirir toma un token (bloquea si no hay disponibles)
func (s *Semaphore) Adquirir() {
	s.tokens <- struct{}{}
}

// Liberar devuelve un token
func (s *Semaphore) Liberar() {
	<-s.tokens
}

// ─────────────────────────────────────────────────────────────
//  5️⃣  RATE LIMITER: Token Bucket Pattern
// ─────────────────────────────────────────────────────────────

// RateLimitador implementa el patrón Token Bucket.
// Permite N operaciones por período de tiempo.
//
// ANALOGÍA: Como una máquina expendedora que suelta 1 ticket cada 200ms.
// Si llegan 10 personas a la vez, la primera toma el ticket y las otras
// esperan a que la máquina suelte el siguiente.
type RateLimitador struct {
	ticker *time.Ticker
	tokens chan struct{}
}

// NuevoRateLimitador crea un limitador que permite 1 operación cada "intervalo"
func NuevoRateLimitador(intervalo time.Duration) *RateLimitador {
	rl := &RateLimitador{
		ticker: time.NewTicker(intervalo),
		tokens: make(chan struct{}, 1),
	}

	// Goroutine que llena el bucket con tokens a intervalos regulares
	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
				// Token puesto en el bucket
			default:
				// Bucket lleno — no acumulamos tokens
			}
		}
	}()

	return rl
}

// Esperar bloquea hasta que haya un token disponible
func (rl *RateLimitador) Esperar() {
	<-rl.tokens
}

// Detener limpia los recursos del rate limiter
func (rl *RateLimitador) Detener() {
	rl.ticker.Stop()
}

// ─────────────────────────────────────────────────────────────
//  Simulación de descarga de páginas web
// ─────────────────────────────────────────────────────────────

// descargarPaginaSimulada simula descargar una página web
// En un crawler real, esto haría http.Get() y leería el body
func descargarPaginaSimulada(workerID int, url string) PaginaWeb {
	// Simulamos latencia de red variable (50ms a 800ms)
	latencia := time.Duration(50+rand.Intn(750)) * time.Millisecond
	time.Sleep(latencia)

	// Simulamos que ~80% descargas exitosas, ~15% lentas, ~5% fallan
	azar := rand.Intn(100)

	pagina := PaginaWeb{
		URL:      url,
		WorkerID: workerID,
		Latencia: latencia,
	}

	switch {
	case azar < 80:
		// Éxito normal
		pagina.StatusCode = 200
		pagina.Contenido = fmt.Sprintf("<html>Contenido de %s (procesado por worker %d)</html>", url, workerID)
		pagina.Tamano = 1000 + rand.Intn(9000) // 1KB a 10KB
	case azar < 95:
		// Respuesta lenta pero exitosa
		time.Sleep(500 * time.Millisecond) // Latencia extra
		pagina.StatusCode = 200
		pagina.Contenido = fmt.Sprintf("<html>Contenido LENTO de %s</html>", url)
		pagina.Tamano = 500 + rand.Intn(2000)
	default:
		// Error
		pagina.StatusCode = 503
		pagina.Error = "service unavailable"
	}

	return pagina
}

// ─────────────────────────────────────────────────────────────
//  DEMO 1: Fan-Out / Fan-In Básico
// ─────────────────────────────────────────────────────────────

func demoFanOutFanIn() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🔀 DEMO 1: Fan-Out / Fan-In — Distribuir y Converger")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	urls := []string{
		"https://api.ejemplo.com/usuarios",
		"https://api.ejemplo.com/productos",
		"https://api.ejemplo.com/pedidos",
		"https://api.ejemplo.com/inventario",
		"https://api.ejemplo.com/reportes",
		"https://api.ejemplo.com/notificaciones",
		"https://api.ejemplo.com/configuracion",
		"https://api.ejemplo.com/auditoria",
	}

	fmt.Println("    📋 URLs a procesar:", len(urls))
	fmt.Println()

	// ── PASO 1: Fan-Out ──────────────────────────────────
	// Distribuimos las URLs entre 4 workers
	numWorkers := 4
	fmt.Printf("    🔀 Fan-Out: distribuyendo %d URLs entre %d workers\n\n", len(urls), numWorkers)

	// Creamos el channel de trabajos
	trabajos := make(chan URLTrabajo, len(urls))

	// Lanzamos los workers (fan-out)
	resultadosWorkers := fanOut(trabajos, numWorkers)

	// Enviamos los trabajos
	for _, url := range urls {
		trabajos <- URLTrabajo{URL: url, Profundidad: 0}
	}
	close(trabajos)

	// ── PASO 2: Fan-In ───────────────────────────────────
	// Convergemos los 4 channels en uno solo
	fmt.Printf("    🔀 Fan-In: convergiendo %d canales en 1\n\n", numWorkers)

	canalUnificado := fanIn(resultadosWorkers)

	// Consumimos los resultados
	var totalPaginas int
	var exitosos, fallidos int
	for p := range canalUnificado {
		totalPaginas++

		if p.Error != "" {
			fallidos++
			fmt.Printf("    ❌ [%s] Error: %s (worker %d, %v)\n",
				p.URL, p.Error, p.WorkerID, p.Latencia)
		} else {
			exitosos++
			fmt.Printf("    ✅ [%s] HTTP %d — %d bytes (worker %d, %v)\n",
				p.URL, p.StatusCode, p.Tamano, p.WorkerID, p.Latencia)
		}
	}

	fmt.Println()
	fmt.Printf("    📊 Resumen: %d procesadas | ✅ %d exitosas | ❌ %d fallidas\n",
		totalPaginas, exitosos, fallidos)
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────
//  DEMO 2: Worker Pool con Límite de Concurrencia
// ─────────────────────────────────────────────────────────────

func demoWorkerPool() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🏭 DEMO 2: Worker Pool — Límite de Concurrencia")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// 20 URLs para crawlear
	urls := make([]string, 20)
	for i := range urls {
		urls[i] = fmt.Sprintf("https://ejemplo.com/pagina-%d", i+1)
	}

	numWorkers := 3 // Solo 3 workers simultáneos
	fmt.Printf("    📋 URLs a procesar: %d\n", len(urls))
	fmt.Printf("    🏭 Workers disponibles: %d\n", numWorkers)
	fmt.Printf("    ⏳ Las URLs restantes esperan en cola...\n\n")

	estadisticas := &EstadisticasCrawl{}
	inicio := time.Now()

	// Creamos el channel de trabajos (tamaño = num trabajos)
	trabajos := make(chan URLTrabajo, len(urls))

	// Lanzamos el worker pool
	resultados := workerPool(numWorkers, trabajos, estadisticas)

	// Enviamos los trabajos
	for _, url := range urls {
		trabajos <- URLTrabajo{URL: url, Profundidad: 0}
	}
	close(trabajos)

	// Consumimos resultados e imprimimos en tiempo real
	for resultado := range resultados {
		p := resultado.Pagina
		status := "✅"
		if p.Error != "" {
			status = "❌"
		}
		fmt.Printf("    %s Worker %d | %s | HTTP %d | %d bytes | %v\n",
			status, p.WorkerID, p.URL, p.StatusCode, p.Tamano, p.Latencia)
	}

	estadisticas.TiempoTotal = time.Since(inicio)

	fmt.Println()
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Println("    📊 ESTADÍSTICAS DEL WORKER POOL")
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Printf("    🏭 Workers utilizados:    %d\n", numWorkers)
	fmt.Printf("    📋 Páginas descargadas:   %d\n", atomic.LoadInt64(&estadisticas.PaginasDescargadas))
	fmt.Printf("    ❌ Páginas con error:     %d\n", atomic.LoadInt64(&estadisticas.PaginasError))
	fmt.Printf("    💾 Bytes totales:         %d\n", atomic.LoadInt64(&estadisticas.BytesTotal))
	fmt.Printf("    ⏱️  Tiempo total:          %v\n", estadisticas.TiempoTotal)
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────
//  DEMO 3: Semaphore — Controlar acceso concurrente
// ─────────────────────────────────────────────────────────────

func demoSemaphore() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🔒 DEMO 3: Semaphore — Controlar Acceso a Recursos")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Simulamos una base de datos que solo soporta 2 conexiones simultáneas
	maxConexiones := 2
	sem := NuevoSemaphore(maxConexiones)
	fmt.Printf("    🗄️  Base de datos soporta máximo %d conexiones simultáneas\n\n", maxConexiones)

	var wg sync.WaitGroup
	var conexionesActivas int64
	var maxConcurrentes int64

	// Lanzamos 8 "consultas" que intentan conectarse
	numConsultas := 8
	for i := 1; i <= numConsultas; i++ {
		wg.Add(1)
		consultaID := i

		go func() {
			defer wg.Done()

			// Intentar adquirir un "token" (conexión a la DB)
			fmt.Printf("    📋 Consulta %d: esperando conexión...\n", consultaID)
			sem.Adquirir()

			// Dentro de la sección crítica
			actual := atomic.AddInt64(&conexionesActivas, 1)

			// Registramos el máximo de conexiones concurrentes
			for {
				max := atomic.LoadInt64(&maxConcurrentes)
				if actual <= max || atomic.CompareAndSwapInt64(&maxConcurrentes, max, actual) {
					break
				}
			}

			fmt.Printf("    🔗 Consulta %d: CONECTADA (activas: %d)\n", consultaID, actual)

			// Simulamos trabajo de base de datos
			time.Sleep(time.Duration(200+rand.Intn(800)) * time.Millisecond)

			atomic.AddInt64(&conexionesActivas, -1)
			fmt.Printf("    ✅ Consulta %d: completada — desconectando\n", consultaID)

			// Liberar el token (cerrar conexión)
			sem.Liberar()
		}()
	}

	wg.Wait()

	fmt.Println()
	fmt.Printf("    📊 Conexiones simultáneas máximas: %d (límite: %d)\n",
		atomic.LoadInt64(&maxConcurrentes), maxConexiones)
	fmt.Printf("    📊 El semáforo nunca permitió más de %d conexiones a la vez\n", maxConexiones)
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────
//  DEMO 4: Rate Limiter — Token Bucket
// ─────────────────────────────────────────────────────────────

func demoRateLimiter() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  ⏱️  DEMO 4: Rate Limiter — Patrón Token Bucket")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Rate limit: 1 request cada 300ms (~3.3 req/segundo)
	intervalo := 300 * time.Millisecond
	limiter := NuevoRateLimitador(intervalo)
	defer limiter.Detener()

	fmt.Printf("    ⏱️  Rate limit: 1 request cada %v\n", intervalo)
	fmt.Printf("    📡 Simulando 10 requests a una API...\n\n")

	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		requestID := i

		go func() {
			defer wg.Done()

			inicio := time.Now()

			// Esperar a que haya un token disponible
			limiter.Esperar()

			tiempoEspera := time.Since(inicio)

			// Simular el request
			fmt.Printf("    📡 Request %d: enviado (esperó %v para obtener token)\n",
				requestID, tiempoEspera)
		}()
	}

	wg.Wait()

	fmt.Println()
	fmt.Println("    📊 Observa: cada request respeta el intervalo mínimo.")
	fmt.Println("    📊 Esto evita saturar APIs y recibir errores HTTP 429.")
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────
//  DEMO 5: Crawler Completo (todos los patrones combinados)
// ─────────────────────────────────────────────────────────────

func demoCrawlerCompleto() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🕷️  DEMO 5: Crawler Web Completo")
	fmt.Println("  Fan-Out + Worker Pool + Semaphore + Rate Limiter")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// ── Configuración ────────────────────────────────────
	maxWorkers := 3
	maxConexiones := 2
	intervaloRateLimit := 200 * time.Millisecond

	fmt.Printf("    🏭 Workers:         %d\n", maxWorkers)
	fmt.Printf("    🔒 Max conexiones: %d (semaphore)\n", maxConexiones)
	fmt.Printf("    ⏱️  Rate limit:      1 req / %v\n", intervaloRateLimit)
	fmt.Println()

	// ── Semáforo y Rate Limiter ──────────────────────────
	sem := NuevoSemaphore(maxConexiones)
	limiter := NuevoRateLimitador(intervaloRateLimit)
	defer limiter.Detener()

	// ── URLs semilla ─────────────────────────────────────
	urlsSemilla := []string{
		"https://ejemplo.com/",
		"https://ejemplo.com/productos",
		"https://ejemplo.com/blog",
		"https://ejemplo.com/contacto",
		"https://ejemplo.com/api/v1/usuarios",
		"https://ejemplo.com/api/v1/pedidos",
		"https://ejemplo.com/docs",
		"https://ejemplo.com/pricing",
		"https://ejemplo.com/about",
		"https://ejemplo.com/careers",
		"https://ejemplo.com/status",
		"https://ejemplo.com/changelog",
	}

	estadisticas := &EstadisticasCrawl{}
	inicioTotal := time.Now()

	// ── Canal de trabajos ────────────────────────────────
	trabajos := make(chan URLTrabajo, len(urlsSemilla))

	// ── Fan-Out: Workers con semaphore + rate limiter ────
	resultados := make(chan ResultadoCrawl)
	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		workerID := i + 1

		go func() {
			defer wg.Done()

			for trabajo := range trabajos {
				// 1. Rate limiter: respetar límites de la API
				limiter.Esperar()

				// 2. Semaphore: limitar conexiones simultáneas
				sem.Adquirir()

				// 3. Descargar la página
				inicio := time.Now()
				pagina := descargarPaginaSimulada(workerID, trabajo.URL)
				pagina.Latencia = time.Since(inicio)

				// 4. Liberar semaphore
				sem.Liberar()

				// 5. Actualizar estadísticas
				if pagina.Error != "" {
					atomic.AddInt64(&estadisticas.PaginasError, 1)
				} else {
					atomic.AddInt64(&estadisticas.PaginasDescargadas, 1)
					atomic.AddInt64(&estadisticas.BytesTotal, int64(pagina.Tamano))
				}

				// 6. Enviar resultado al canal unificado (fan-in implícito)
				resultados <- ResultadoCrawl{
					Pagina:      pagina,
					Profundidad: trabajo.Profundidad,
					Procesado:   time.Now(),
				}
			}
		}()
	}

	// ── Fan-In: Recolector de resultados ─────────────────
	var wgRecolector sync.WaitGroup
	wgRecolector.Add(1)

	go func() {
		defer wgRecolector.Done()
		for resultado := range resultados {
			p := resultado.Pagina
			emoji := "✅"
			if p.Error != "" {
				emoji = "❌"
			}
			fmt.Printf("    %s W%d | %-40s | HTTP %d | %6d bytes | %v\n",
				emoji, p.WorkerID, p.URL, p.StatusCode, p.Tamano, p.Latencia)
		}
	}()

	// ── Enviar trabajos ──────────────────────────────────
	for _, url := range urlsSemilla {
		trabajos <- URLTrabajo{URL: url, Profundidad: 0}
	}
	close(trabajos)

	// ── Esperar a que los workers terminen ────────────────
	wg.Wait()
	close(resultados)
	wgRecolector.Wait()

	estadisticas.TiempoTotal = time.Since(inicioTotal)

	// ── Reporte final ────────────────────────────────────
	fmt.Println()
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Println("    📊 REPORTE FINAL DEL CRAWLER")
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Printf("    🕷️  URLs procesadas:       %d\n", len(urlsSemilla))
	fmt.Printf("    🏭 Workers utilizados:    %d\n", maxWorkers)
	fmt.Printf("    🔒 Conexiones máximas:    %d (semaphore)\n", maxConexiones)
	fmt.Printf("    ⏱️  Rate limit aplicado:   1 req / %v\n", intervaloRateLimit)
	fmt.Printf("    ✅ Descargas exitosas:    %d\n", atomic.LoadInt64(&estadisticas.PaginasDescargadas))
	fmt.Printf("    ❌ Descargas fallidas:    %d\n", atomic.LoadInt64(&estadisticas.PaginasError))
	fmt.Printf("    💾 Bytes totales:         %d\n", atomic.LoadInt64(&estadisticas.BytesTotal))
	fmt.Printf("    ⏱️  Tiempo total:          %v\n", estadisticas.TiempoTotal)
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────
//  Main
// ─────────────────────────────────────────────────────────────

func main() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🍕 PATRONES AVANZADOS DE CONCURRENCIA")
	fmt.Println("  📚 Lección 13 — Laboratorio de Go")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Ejecutar cada demo secuencialmente
	demoFanOutFanIn()
	demoWorkerPool()
	demoSemaphore()
	demoRateLimiter()
	demoCrawlerCompleto()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  ✅ Todas las demos completadas")
	fmt.Println("  📚 Patrones demostrados:")
	fmt.Println("     • Fan-Out: distribuir trabajo a N goroutines")
	fmt.Println("     • Fan-In: converger N channels en 1")
	fmt.Println("     • Worker Pool: limitar goroutines activas")
	fmt.Println("     • Semaphore: controlar acceso a recursos")
	fmt.Println("     • Rate Limiter: token bucket pattern")
	fmt.Println("     • Crawler completo: todos los patrones combinados")
	fmt.Println("═══════════════════════════════════════════════════════════")
}