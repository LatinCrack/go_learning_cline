package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// CONSTANTES Y TIPOS
// ============================================================================

// contextKey es un tipo personalizado para keys del context.
// Usar un tipo propio evita colisiones con keys de otros paquetes.
// Dos strings iguales de paquetes distintos NO colisionan si tienen tipos distintos.
type contextKey string

const (
	requestIDKey contextKey = "requestID"
	clientIPKey  contextKey = "clientIP"
)

// RateLimiter controla cuántas peticiones puede hacer cada IP por minuto.
// Usa un mapa de IPs a contadores con un mutex para acceso concurrente seguro.
type RateLimiter struct {
	solicitudes map[string][]time.Time
	limit       int
	ventana     time.Duration
	mu          sync.Mutex
}

// NuevoRateLimiter crea un limitador de peticiones.
// limit = máximo de peticiones por ventana de tiempo.
// ventana = duración de la ventana (ej: 1 minuto).
func NuevoRateLimiter(limit int, ventana time.Duration) *RateLimiter {
	return &RateLimiter{
		solicitudes: make(map[string][]time.Time),
		limit:       limit,
		ventana:     ventana,
	}
}

// Permitir verifica si una IP puede hacer una petición.
// Limpia peticiones antiguas que ya están fuera de la ventana.
func (rl *RateLimiter) Permitir(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock() // 🧹 defer garantiza que el mutex se libere

	ahora := time.Now()
	cutoff := ahora.Add(-rl.ventana)

	// Filtrar peticiones dentro de la ventana
	var recientes []time.Time
	for _, t := range rl.solicitudes[ip] {
		if t.After(cutoff) {
			recientes = append(recientes, t)
		}
	}
	rl.solicitudes[ip] = recientes

	if len(recientes) >= rl.limit {
		return false // Rate limit excedido
	}

	rl.solicitudes[ip] = append(rl.solicitudes[ip], ahora)
	return true
}

// ============================================================================
// MIDDLEWARES — context + defer + recover en acción
// ============================================================================

// requestIDMiddleware genera un ID único para cada request y lo
// agrega al context usando WithValue. Así, toda la cadena de llamadas
// puede acceder al request ID sin pasarlo como parámetro explícito.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generar un ID simple (en producción usarías uuid)
		id := fmt.Sprintf("req-%d", time.Now().UnixNano()%100000)

		// Crear un nuevo context con el request ID
		ctx := context.WithValue(r.Context(), requestIDKey, id)

		// Extraer la IP del cliente
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
		}
		ctx = context.WithValue(ctx, clientIPKey, ip)

		// Pasar el nuevo context al siguiente handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// recoveryMiddleware atrapa panics en los handlers y evita que
// el servidor se cae. Funciona como una mampara estanca: si un
// compartimento (handler) se inunda (panic), el barco (servidor) sigue.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Extraer el request ID del context para el log
				reqID := "unknown"
				if id, ok := r.Context().Value(requestIDKey).(string); ok {
					reqID = id
				}

				// Log del panic con contexto
				log.Printf("🔴 PANIC recuperado [%s] en %s %s: %v",
					reqID, r.Method, r.URL.Path, err)

				// Responder al cliente con error 500
				http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware registra cada request con su duración.
// Demuestra el patrón de defer para medir tiempo de ejecución.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()

		// Extraer request ID del context
		reqID := "unknown"
		if id, ok := r.Context().Value(requestIDKey).(string); ok {
			reqID = id
		}

		// defer se ejecuta DESPUÉS de que el handler termine
		defer func() {
			duracion := time.Since(inicio)
			log.Printf("📝 [%s] %s %s → %v",
				reqID, r.Method, r.URL.Path, duracion)
		}()

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// HANDLERS — Ejemplos de context, defer y panic/recover
// ============================================================================

// handleHome muestra la página principal del servidor de demostración.
func handleHome(w http.ResponseWriter, r *http.Request) {
	// Recuperar datos del context (propagados por middlewares)
	reqID := r.Context().Value(requestIDKey).(string)
	clientIP := r.Context().Value(clientIPKey).(string)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "🛡️ Servidor de Demostración — Lección 18\n")
	fmt.Fprintf(w, "════════════════════════════════════════\n\n")
	fmt.Fprintf(w, "📋 Request ID: %s\n", reqID)
	fmt.Fprintf(w, "🌐 Tu IP: %s\n\n", clientIP)
	fmt.Fprintf(w, "Endpoints disponibles:\n")
	fmt.Fprintf(w, "  GET /              → Esta página\n")
	fmt.Fprintf(w, "  GET /lento         → Simula operación lenta (con timeout)\n")
	fmt.Fprintf(w, "  GET /panic         → Simula un panic (recover lo atrapa)\n")
	fmt.Fprintf(w, "  GET /cancel        → Simula trabajo cancelable\n")
	fmt.Fprintf(w, "  GET /concurrent    → Múltiples goroutines con context\n")
}

// handleLento simula una operación lenta con timeout.
// Demuestra context.WithTimeout y defer cancel().
func handleLento(w http.ResponseWriter, r *http.Request) {
	// Crear un context con timeout de 3 segundos.
	// Si la operación tarda más, el context se cancela automáticamente.
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel() // 🧹 SIEMPRE llamar cancel para liberar recursos del timer

	reqID := ctx.Value(requestIDKey).(string)
	log.Printf("⏳ [%s] Iniciando operación lenta (timeout: 3s)...", reqID)

	// Simular una operación que tarda entre 1 y 6 segundos
	duracion := time.Duration(1+rand.Intn(6)) * time.Second
	log.Printf("⏳ [%s] La operación tardará %v...", reqID, duracion)

	// select: el que termine primero gana
	select {
	case <-time.After(duracion):
		// La operación completó a tiempo
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "✅ [%s] Operación completada en %v\n", reqID, duracion)

	case <-ctx.Done():
		// El timeout se agotó antes de que la operación terminara
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusGatewayTimeout)
		fmt.Fprintf(w, "⏰ [%s] Timeout excedido: %v\n", reqID, ctx.Err())
		fmt.Fprintf(w, "   La operación tardaba %v pero el límite era 3s\n", duracion)
	}
}

// handlePanic simula un panic que es atrapado por el recoveryMiddleware.
// Demuestra que un panic en un handler NO tumbar el servidor.
func handlePanic(w http.ResponseWriter, r *http.Request) {
	reqID := r.Context().Value(requestIDKey).(string)
	log.Printf("💥 [%s] Este handler va a hacer panic...", reqID)

	// Esto causará un panic que será atrapado por recoveryMiddleware
	panic("error catastrófico simulado: división por cero en el módulo de pagos")

	// Esta línea nunca se ejecuta
	fmt.Fprintf(w, "Si ves esto, algo está muy mal")
}

// handleCancel simula trabajo que puede ser cancelado por el cliente.
// Demuestra ctx.Done() y la propagation de cancellation a goroutines hijas.
func handleCancel(w http.ResponseWriter, r *http.Request) {
	// Usar el context del request — se cancela si el cliente cierra la conexión
	ctx := r.Context()
	reqID := ctx.Value(requestIDKey).(string)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "🔄 [%s] Iniciamiento trabajo cancelable...\n", reqID)

	// Flusher permite enviar respuesta parcial antes de que el handler termine
	flusher, ok := w.(http.Flusher)
	if ok {
		flusher.Flush()
	}

	for i := 1; i <= 10; i++ {
		select {
		case <-ctx.Done():
			// El cliente cerró la conexión o el context se canceló
			log.Printf("🛑 [%s] Trabajo cancelado en paso %d: %v", reqID, i, ctx.Err())
			fmt.Fprintf(w, "\n🛑 Trabajo cancelado en paso %d de 10\n", i)
			fmt.Fprintf(w, "   Razón: %v\n", ctx.Err())
			return
		case <-time.After(500 * time.Millisecond):
			// Simular trabajo
			fmt.Fprintf(w, "  ⚙️  Paso %d/10 completado\n", i)
			if ok {
				flusher.Flush() // Enviar cada paso al cliente en tiempo real
			}
		}
	}

	fmt.Fprintf(w, "\n✅ Trabajo completado exitosamente (10/10 pasos)\n")
}

// handleConcurrent lanza múltiples goroutines con un context con timeout.
// Demuestra cómo la cancellation se propaga a TODAS las goroutines hijas.
func handleConcurrent(w http.ResponseWriter, r *http.Request) {
	// Timeout global de 2 segundos para todas las goroutines
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel() // 🧹 Libera recursos del timer

	reqID := ctx.Value(requestIDKey).(string)

	// Canal para recolectar resultados
	resultados := make(chan string, 5)
	var wg sync.WaitGroup

	// Lanzar 5 "servicios" concurrentes con tiempos variables
	servicios := []struct {
		nombre  string
		duracion time.Duration
	}{
		{"Base de Datos", time.Duration(500+rand.Intn(1000)) * time.Millisecond},
		{"API Externa", time.Duration(1000+rand.Intn(2000)) * time.Millisecond},
		{"Cache Redis", time.Duration(100+rand.Intn(300)) * time.Millisecond},
		{"Servicio de Email", time.Duration(800+rand.Intn(1500)) * time.Millisecond},
		{"Procesamiento ML", time.Duration(1500+rand.Intn(3000)) * time.Millisecond},
	}

	for _, svc := range servicios {
		wg.Add(1)
		go func(nombre string, dur time.Duration) {
			defer wg.Done() // 🧹 Marcar como completado al terminar

			log.Printf("🧵 [%s] %s iniciando (tardará %v)...", reqID, nombre, dur)

			select {
			case <-time.After(dur):
				// El servicio respondió a tiempo
				resultados <- fmt.Sprintf("  ✅ %s: completado en %v", nombre, dur)
			case <-ctx.Done():
				// El timeout global se agotó antes de que este servicio respondiera
				resultados <- fmt.Sprintf("  ⏰ %s: cancelado (%v)", nombre, ctx.Err())
			}
		}(svc.nombre, svc.duracion)
	}

	// Goroutine para cerrar el canal cuando todas terminen
	go func() {
		wg.Wait()
		close(resultados)
	}()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "🧵 [%s] Ejecutando 5 servicios concurrentes (timeout: 2s)\n\n", reqID)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	// Recolectar resultados a medida que llegan
	for resultado := range resultados {
		fmt.Fprintf(w, "%s\n", resultado)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}

	fmt.Fprintf(w, "\n📊 Todos los servicios reportaron (completados o cancelados)\n")
}

// ============================================================================
// SERVIDOR HTTP — context + defer + recover en producción
// ============================================================================

// IniciarServidor levanta el servidor HTTP con middlewares.
func IniciarServidor(puerto string) {
	// Rate limiter: máximo 30 peticiones por minuto por IP
	limiter := NuevoRateLimiter(30, time.Minute)

	// Crear el mux (router)
	mux := http.NewServeMux()

	// Registrar handlers
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/lento", handleLento)
	mux.HandleFunc("/panic", handlePanic)
	mux.HandleFunc("/cancel", handleCancel)
	mux.HandleFunc("/concurrent", handleConcurrent)

	// Envolver con middlewares (el orden importa: el más externo se ejecuta primero)
	// requestID → recovery → logging → handler
	var handler http.Handler = mux
	handler = loggingMiddleware(handler)
	handler = recoveryMiddleware(handler)
	handler = requestIDMiddleware(handler)

	// Middleware de rate limiting manual (por simplicidad)
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !limiter.Permitir(ip) {
			http.Error(w, "🚫 Rate limit excedido. Intenta en 1 minuto.", http.StatusTooManyRequests)
			return
		}
		handler.ServeHTTP(w, r)
	})

	servidor := &http.Server{
		Addr:         puerto,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("╔══════════════════════════════════════════════════╗\n")
	fmt.Printf("║  🛡️  Servidor HTTP — Lección 18                 ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Puerto: %-39s ║\n", puerto)
	fmt.Printf("║  Endpoints:                                      ║\n")
	fmt.Printf("║    GET /           → Página principal             ║\n")
	fmt.Printf("║    GET /lento      → Timeout demo                 ║\n")
	fmt.Printf("║    GET /panic      → Panic/recover demo           ║\n")
	fmt.Printf("║    GET /cancel     → Context cancellation demo    ║\n")
	fmt.Printf("║    GET /concurrent → Goroutines + context demo    ║\n")
	fmt.Printf("║                                                    ║\n")
	fmt.Printf("║  Presiona Ctrl+C para detener                     ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════╝\n")
	fmt.Println()

	log.Fatal(servidor.ListenAndServe())
}

// ============================================================================
// MODO DEMOSTRACIÓN
// ============================================================================

// DemoModo ejecuta ejemplos interactivos de cada concepto.
func DemoModo() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   📚 DEMOSTRACIÓN: context, defer, panic/recover            ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")

	// --- Ejemplo 1: context.Background() y context.TODO() ---
	fmt.Println("\n━━━ 1. context.Background() y context.TODO() ━━━")
	fmt.Println("   Background() = raíz del árbol de contexts (nunca se cancela)")
	fmt.Println("   TODO() = placeholder semántico para contexts no implementados\n")

	ctx := context.Background()
	fmt.Printf("   ctx = context.Background()\n")
	fmt.Printf("   ctx.Done() = %v (nil = nunca se cierra)\n", ctx.Done())
	fmt.Printf("   ctx.Err()  = %v (nil = nunca se canceló)\n", ctx.Err())
	deadline, ok := ctx.Deadline()
	fmt.Printf("   ctx.Deadline() = %v, ok = %v (no tiene deadline)\n\n", deadline, ok)

	ctxTODO := context.TODO()
	fmt.Printf("   ctx = context.TODO()\n")
	fmt.Printf("   ¿Son iguales? %v (funcionalmente sí, semánticamente no)\n\n",
		ctx.Done() == ctxTODO.Done())

	// --- Ejemplo 2: context.WithCancel ---
	fmt.Println("━━━ 2. context.WithCancel — Cancelación manual ━━━")
	fmt.Println("   Creamos 3 workers que se cancelan con un solo cancel()\n")

	ctx2, cancel2 := context.WithCancel(context.Background())
	var wg2 sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done() // 🧹 defer: se ejecuta cuando el worker termina
			paso := 0
			for {
				select {
				case <-ctx2.Done():
					fmt.Printf("   Worker %d: cancelado en paso %d (%v)\n",
						id, paso, ctx2.Err())
					return
				default:
					paso++
					fmt.Printf("   Worker %d: paso %d\n", id, paso)
					time.Sleep(300 * time.Millisecond)
				}
			}
		}(i)
	}

	time.Sleep(1 * time.Second)
	fmt.Println("\n   🛑 Llamando cancel()...")
	cancel2()
	wg2.Wait()

	// --- Ejemplo 3: context.WithTimeout ---
	fmt.Println("\n━━━ 3. context.WithTimeout — Cancelación automática ━━━")
	fmt.Println("   Simulamos una operación que tarda 3 segundos con timeout de 1.5s\n")

	ctx3, cancel3 := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel3() // 🧹 SIEMPRE defer cancel()

	resultado := make(chan string, 1)
	go func() {
		time.Sleep(3 * time.Second) // Tarda 3 segundos
		resultado <- "datos de la base de datos"
	}()

	select {
	case res := <-resultado:
		fmt.Printf("   ✅ Resultado: %s\n", res)
	case <-ctx3.Done():
		fmt.Printf("   ⏰ Timeout: %v\n", ctx3.Err())
		fmt.Printf("   ctx.Err() = context.DeadlineExceeded? %v\n",
			ctx3.Err() == context.DeadlineExceeded)
	}

	// --- Ejemplo 4: context.WithValue ---
	fmt.Println("\n━━━ 4. context.WithValue — Datos del request ━━━")
	fmt.Println("   Propagamos requestID y userID a través del context\n")

	ctx4 := context.WithValue(context.Background(), requestIDKey, "req-42")
	ctx4 = context.WithValue(ctx4, clientIPKey, "192.168.1.100")

	// Simular una función más profunda que accede al context
	func(ctx context.Context) {
		reqID := ctx.Value(requestIDKey).(string)
		clientIP := ctx.Value(clientIPKey).(string)
		fmt.Printf("   Dentro de función anidada:\n")
		fmt.Printf("     Request ID = %s\n", reqID)
		fmt.Printf("     Client IP  = %s\n", clientIP)
		fmt.Printf("     ctx.Value(\"noExiste\") = %v (nil si no existe)\n",
			ctx.Value("noExiste"))
	}(ctx4)

	// --- Ejemplo 5: defer LIFO ---
	fmt.Println("\n━━━ 5. defer — Orden LIFO ━━━")
	fmt.Println("   Los deferred functions se ejecutan en orden inverso\n")

	fmt.Println("   Código:")
	fmt.Println("     defer fmt.Println('A')")
	fmt.Println("     defer fmt.Println('B')")
	fmt.Println("     fmt.Println('C')")
	fmt.Println()
	fmt.Println("   Salida:")
	fmt.Println("     C") // Primero el cuerpo de la función
	func() {
		defer fmt.Println("     A") // Último en deferirse → último en ejecutarse
		defer fmt.Println("     B") // Segundo en deferirse
		fmt.Println("     C")       // Cuerpo de la función
	}()

	// --- Ejemplo 6: defer con argumentos evaluados inmediatamente ---
	fmt.Println("\n━━━ 6. defer — Trampa de argumentos ━━━")
	fmt.Println("   Los argumentos se evalúan AHORA, no al ejecutarse\n")

	fmt.Println("   Ejemplo CON argumento directo (captura valor inicial):")
	func() {
		x := 10
		defer fmt.Printf("     defer: x = %d (valor capturado en el defer)\n", x)
		x = 20
		fmt.Printf("     main:  x = %d (valor modificado)\n", x)
	}()
	time.Sleep(100 * time.Millisecond) // Esperar a que defer se ejecute

	fmt.Println("\n   Ejemplo CON closure (captura referencia):")
	func() {
		x := 10
		defer func() {
			fmt.Printf("     defer: x = %d (lee el valor al ejecutarse)\n", x)
		}()
		x = 20
		fmt.Printf("     main:  x = %d\n", x)
	}()
	time.Sleep(100 * time.Millisecond)

	// --- Ejemplo 7: defer para medir tiempo ---
	fmt.Println("\n━━━ 7. defer — Medir tiempo de ejecución ━━━")

	func() {
		inicio := time.Now()
		defer func() {
			fmt.Printf("   ⏱️  Duración total: %v\n", time.Since(inicio))
		}()

		fmt.Println("   Ejecutando operación...")
		time.Sleep(800 * time.Millisecond)
		fmt.Println("   ✅ Operación completada")
	}()

	// --- Ejemplo 8: panic y recover ---
	fmt.Println("\n━━━ 8. panic y recover — El extintor de incendios ━━━")
	fmt.Println("   Un panic en c() se propaga hasta que recover() lo atrapa\n")

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("   🧯 recover() atrapó: %v\n", r)
				fmt.Println("   ✅ El programa continúa después del recover")
			}
		}()

		func() {
			defer fmt.Println("   🧹 defer en función intermedia (se ejecuta)")
			func() {
				defer fmt.Println("   🧹 defer en función que panic (se ejecuta)")
				fmt.Println("   💥 Haciendo panic...")
				panic("error catastrófico simulado")
			}()
			fmt.Println("   (esto nunca se ejecuta)")
		}()
		fmt.Println("   (esto tampoco se ejecuta)")
	}()

	// --- Ejemplo 9: recover convierte panic en error ---
	fmt.Println("\n━━━ 9. Patrón: convertir panic en error ━━━")

	func() {
		err := OperacionQuePodriaPanic(true)
		if err != nil {
			fmt.Printf("   Error manejado: %v\n", err)
		} else {
			fmt.Println("   ✅ Sin errores")
		}

		err = OperacionQuePodriaPanic(false)
		if err != nil {
			fmt.Printf("   Error manejado: %v\n", err)
		} else {
			fmt.Println("   ✅ Sin errores (no hubo panic)")
		}
	}()

	// --- Ejemplo 10: Context tree ---
	fmt.Println("\n━━━ 10. Árbol de contexts — Padre e hijos ━━━")
	fmt.Println("   Si se cancela el padre, TODOS los hijos se cancelan\n")

	padre, cancelPadre := context.WithCancel(context.Background())

	hijo1, _ := context.WithTimeout(padre, 10*time.Second)
	hijo2, _ := context.WithTimeout(padre, 10*time.Second)
	nieto1, _ := context.WithTimeout(hijo1, 10*time.Second)

	fmt.Printf("   Padre.Err()   = %v\n", padre.Err())
	fmt.Printf("   Hijo1.Err()   = %v\n", hijo1.Err())
	fmt.Printf("   Hijo2.Err()   = %v\n", hijo2.Err())
	fmt.Printf("   Nieto1.Err()  = %v\n", nieto1.Err())

	fmt.Println("\n   🛑 Cancelando el padre...")
	cancelPadre()

	// Pequeña espera para que la cancellation se propague
	time.Sleep(50 * time.Millisecond)

	fmt.Printf("   Padre.Err()   = %v\n", padre.Err())
	fmt.Printf("   Hijo1.Err()   = %v\n", hijo1.Err())
	fmt.Printf("   Hijo2.Err()   = %v\n", hijo2.Err())
	fmt.Printf("   Nieto1.Err()  = %v\n", nieto1.Err())
	fmt.Println("   → ¡Todos los descendientes se cancelaron!")

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   ✅ Fin de la demostración.                                  ║")
	fmt.Println("║   Ejecuta con -server para ver el servidor HTTP en acción.   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// OperacionQuePodriaPanic convierte un panic en un error usando defer+recover.
// Esto permite que el llamador maneje el error de forma normal.
func OperacionQuePodriaPanic(debePanic bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recuperado: %v", r)
		}
	}()

	if debePanic {
		panic("¡división por cero!")
	}

	return nil
}

// ============================================================================
// MODO EJEMPLOS — Snippets ejecutables aislados
// ============================================================================

// EjemplosModo ejecuta snippets de código que ilustran cada concepto.
func EjemplosModo() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   🧩 EJEMPLOS PRÁCTICOS: context, defer, panic/recover      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")

	// --- Ejemplo A: Goroutine con timeout ---
	fmt.Println("\n━━━ A. Goroutine con timeout (worker pool) ━━━")

	ejemploGoroutineTimeout()

	// --- Ejemplo B: defer como cleanup de múltiples recursos ---
	fmt.Println("\n━━━ B. defer para cleanup de múltiples recursos ━━━")

	ejemploDeferCleanup()

	// --- Ejemplo C: recover en un servidor de eco TCP simulado ---
	fmt.Println("\n━━━ C. recover protege cada goroutine ━━━")

	ejemploRecoverGoroutine()

	// --- Ejemplo D: Context con valores para tracing ---
	fmt.Println("\n━━━ D. Context values para request tracing ━━━")

	ejemploContextTracing()

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   ✅ Fin de los ejemplos.                                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func ejemploGoroutineTimeout() {
	// Crear un context con timeout de 2 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Canal para resultados
	resultados := make(chan string, 3)

	// Simular 3 trabajos con duraciones diferentes
	trabajos := []struct {
		nombre   string
		duracion time.Duration
	}{
		{"Rápido", 500 * time.Millisecond},
		{"Medio", 1500 * time.Millisecond},
		{"Lento", 4 * time.Second},
	}

	var wg sync.WaitGroup
	for _, t := range trabajos {
		wg.Add(1)
		go func(nombre string, dur time.Duration) {
			defer wg.Done()
			select {
			case <-time.After(dur):
				resultados <- fmt.Sprintf("  ✅ %s completado (%v)", nombre, dur)
			case <-ctx.Done():
				resultados <- fmt.Sprintf("  ⏰ %s cancelado (%v)", nombre, ctx.Err())
			}
		}(t.nombre, t.duracion)
	}

	go func() {
		wg.Wait()
		close(resultados)
	}()

	for r := range resultados {
		fmt.Println(r)
	}
}

func ejemploDeferCleanup() {
	// Simular la apertura de múltiples recursos con defer LIFO
	fmt.Println("  Abriendo recursos en orden: archivo → conexión → transacción")

	func() {
		fmt.Println("  📄 Abriendo archivo...")
		defer fmt.Println("  📄 Cerrando archivo (último defer → último en limpiarse)")

		fmt.Println("  🔌 Abriendo conexión a DB...")
		defer fmt.Println("  🔌 Cerrando conexión a DB")

		fmt.Println("  💳 Iniciando transacción...")
		defer fmt.Println("  💳 Rollback de transacción (primer defer → primero en limpiarse)")

		fmt.Println("  ⚙️  Ejecutando operaciones...")
		fmt.Println("  ✅ Operaciones completadas")
	}()

	fmt.Println("\n  Orden de limpieza (LIFO): transacción → conexión → archivo")
}

func ejemploRecoverGoroutine() {
	// Sin recover: un panic en una goroutine tumbar TODO el programa
	// Con recover: cada goroutine es independiente

	var wg sync.WaitGroup

	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("  🧯 Goroutine %d: panic recuperado: %v\n", id, r)
				}
			}()

			if id == 2 {
				panic(fmt.Sprintf("goroutine %d falló", id))
			}

			fmt.Printf("  ✅ Goroutine %d: completada exitosamente\n", id)
		}(i)
	}

	wg.Wait()
	fmt.Println("  📊 Todas las goroutines terminaron (incluso la que panic)")
}

func ejemploContextTracing() {
	// Simular un request con múltiples capas de tracing
	fmt.Println("  Simulando request con trace ID y user ID en el context\n")

	// Capa 1: middleware (agrega datos al context)
	ctx := context.WithValue(context.Background(), requestIDKey, "trace-789")
	ctx = context.WithValue(ctx, clientIPKey, "10.0.0.5")

	// Capa 2: handler
	func(ctx context.Context) {
		reqID := ctx.Value(requestIDKey).(string)
		fmt.Printf("  Handler [%s]: procesando request...\n", reqID)

		// Capa 3: servicio
		func(ctx context.Context) {
			reqID := ctx.Value(requestIDKey).(string)
			ip := ctx.Value(clientIPKey).(string)
			fmt.Printf("  Servicio [%s]: consultando para IP %s...\n", reqID, ip)

			// Capa 4: repositorio
			func(ctx context.Context) {
				reqID := ctx.Value(requestIDKey).(string)
				fmt.Printf("  Repositorio [%s]: ejecutando query...\n", reqID)
			}(ctx)

			fmt.Printf("  Servicio [%s]: datos obtenidos\n", reqID)
		}(ctx)

		fmt.Printf("  Handler [%s]: respuesta enviada\n", reqID)
	}(ctx)

	fmt.Println("\n  → El requestID viaja a través de TODAS las capas sin ser parámetro")
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	// Definir flags
	demo := flag.Bool("demo", false, "Modo demostración: explica context, defer y panic/recover")
	server := flag.Bool("server", false, "Modo servidor: levanta el servidor HTTP de demostración")
	ejemplos := flag.Bool("ejemplos", false, "Modo ejemplos: ejecuta snippets de código prácticos")
	puerto := flag.String("puerto", ":8080", "Puerto para el servidor HTTP (solo con -server)")
	flag.Parse()

	switch {
	case *demo:
		DemoModo()
	case *server:
		IniciarServidor(*puerto)
	case *ejemplos:
		EjemplosModo()
	default:
		fmt.Println("╔══════════════════════════════════════════════════╗")
		fmt.Println("║  🛡️  context, defer, panic/recover — Lección 18  ║")
		fmt.Println("╠══════════════════════════════════════════════════╣")
		fmt.Println("║                                                  ║")
		fmt.Println("║  Uso:                                            ║")
		fmt.Println("║    go run main.go -demo          Explicación     ║")
		fmt.Println("║    go run main.go -server        Servidor HTTP   ║")
		fmt.Println("║    go run main.go -ejemplos      Snippets        ║")
		fmt.Println("║                                                  ║")
		fmt.Println("║  Flags:                                          ║")
		fmt.Println("║    -puerto :9090    Puerto custom (default :8080) ║")
		fmt.Println("║                                                  ║")
		fmt.Println("╚══════════════════════════════════════════════════╝")
	}
}