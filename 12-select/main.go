package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ═══════════════════════════════════════════════════════════════
//  🎵 Monitor de Servicios con Timeout y Circuit Breaker
//  Lección 12 — select: El Director de Orquesta Concurrente
// ═══════════════════════════════════════════════════════════════

// ─────────────────────────────────────────────────────────────
//  Tipos y Constantes
// ─────────────────────────────────────────────────────────────

// Estado del Circuit Breaker
type EstadoCB int

const (
	Cerrado    EstadoCB = iota // Funcionando normalmente (0)
	Abierto                    // Circuito abierto: servicio "muerto" (1)
	MedioAbierto               // Probando si el servicio se recuperó (2)
)

// String() convierte el estado a texto legible
func (e EstadoCB) String() string {
	switch e {
	case Cerrado:
		return "🟢 CERRADO"
	case Abierto:
		return "🔴 ABIERTO"
	case MedioAbierto:
		return "🟡 MEDIO-ABIERTO"
	default:
		return "❓ DESCONOCIDO"
	}
}

// ResultadoMonitoreo guarda el resultado de un chequeo de salud
type ResultadoMonitoreo struct {
	Servicio   string
	Estado     string
	Latencia   time.Duration
	StatusCode int
	Error      string
	Timestamp  time.Time
}

// Servicio representa un endpoint HTTP que vamos a monitorear
type Servicio struct {
	Nombre string
	URL    string
}

// CircuitBreaker implementa el patrón circuit breaker para un servicio
type CircuitBreaker struct {
	Nombre         string
	Estado         EstadoCB
	FallosConsecut int
	UmbralFallos   int
	TimeoutReset   time.Duration
	UltimoFallo    time.Time
	mu             sync.Mutex
}

// NuevoCircuitBreaker crea un circuit breaker con configuración inicial
func NuevoCircuitBreaker(nombre string, umbral int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		Nombre:       nombre,
		Estado:       Cerrado,
		UmbralFallos: umbral,
		TimeoutReset: timeout,
	}
}

// RegistrarExito resetea el contador de fallos
func (cb *CircuitBreaker) RegistrarExito() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.FallosConsecut = 0
	cb.Estado = Cerrado
}

// RegistrarFallo incrementa el contador y abre el circuito si se supera el umbral
func (cb *CircuitBreaker) RegistrarFallo() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.FallosConsecut++
	cb.UltimoFallo = time.Now()

	if cb.FallosConsecut >= cb.UmbralFallos {
		cb.Estado = Abierto
		fmt.Printf("    ⚡ Circuit Breaker [%s] ABIERTO — %d fallos consecutivos\n",
			cb.Nombre, cb.FallosConsecut)
	}
}

// PermitirVerificacion decide si se permite un nuevo intento
func (cb *CircuitBreaker) PermitirVerificacion() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.Estado {
	case Cerrado:
		return true // Todo normal, permitir
	case Abierto:
		// ¿Ya pasó el timeout? Si sí, pasar a medio-abierto
		if time.Since(cb.UltimoFallo) > cb.TimeoutReset {
			cb.Estado = MedioAbierto
			fmt.Printf("    🔄 Circuit Breaker [%s] → MEDIO-ABIERTO — probando...\n", cb.Nombre)
			return true
		}
		return false // Todavía en timeout, no permitir
	case MedioAbierto:
		return true // Permitir un intento de prueba
	}
	return false
}

// ─────────────────────────────────────────────────────────────
//  Funciones de Monitoreo
// ─────────────────────────────────────────────────────────────

// verificarServicio hace un HTTP GET con timeout usando context
func verificarServicio(ctx context.Context, servicio Servicio) ResultadoMonitoreo {
	resultado := ResultadoMonitoreo{
		Servicio:  servicio.Nombre,
		Timestamp: time.Now(),
	}

	inicio := time.Now()

	// Creamos un request que respeta el context (para timeout/cancelación)
	req, err := http.NewRequestWithContext(ctx, "GET", servicio.URL, nil)
	if err != nil {
		resultado.Error = err.Error()
		resultado.Latencia = time.Since(inicio)
		return resultado
	}

	// Cliente HTTP con su propio timeout como respaldo
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		resultado.Error = err.Error()
		resultado.Estado = "❌ ERROR"
		resultado.Latencia = time.Since(inicio)
		return resultado
	}
	defer resp.Body.Close()

	resultado.StatusCode = resp.StatusCode
	resultado.Latencia = time.Since(inicio)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		resultado.Estado = "✅ OK"
	} else {
		resultado.Estado = fmt.Sprintf("⚠️ HTTP %d", resp.StatusCode)
	}

	return resultado
}

// monitorearServicio es la goroutine que vigila un servicio continuamente
// Usa SELECT para multiplexar entre: respuesta, timeout, y cancelación
func monitorearServicio(
	ctx context.Context,
	servicio Servicio,
	resultados chan<- ResultadoMonitoreo,
	intervalo time.Duration,
) {
	cb := NuevoCircuitBreaker(servicio.Nombre, 3, 30*time.Second)
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()

	fmt.Printf("    👁️  Monitoreando [%s] cada %v\n", servicio.Nombre, intervalo)

	for {
		// ══════════════════════════════════════════════════
		//  🎵 SELECT — El director de orquesta
		//  Escucha 3 "teléfonos" simultáneamente:
		//    1. El ticker (¿es hora de chequear?)
		//    2. El context (¿nos cancelaron?)
		//    3. El circuit breaker (¿debemos saltar?)
		// ══════════════════════════════════════════════════
		select {
		case <-ticker.C:
			// El ticker sonó — es hora de verificar el servicio

			if !cb.PermitirVerificacion() {
				// Circuit breaker abierto: saltamos este chequeo
				resultados <- ResultadoMonitoreo{
					Servicio:  servicio.Nombre,
					Estado:    "⏭️ SKIP (CB abierto)",
					Timestamp: time.Now(),
				}
				continue
			}

			// Creamos un context con timeout específico para esta verificación
			ctxTimeout, cancel := context.WithTimeout(ctx, 4*time.Second)
			resultado := verificarServicio(ctxTimeout, servicio)
			cancel() // Liberamos recursos del context

			// Actualizamos el circuit breaker según el resultado
			if resultado.Error != "" {
				cb.RegistrarFallo()
			} else {
				cb.RegistrarExito()
			}

			resultados <- resultado

		case <-ctx.Done():
			// Context cancelado — el sistema nos pide parar
			fmt.Printf("    🛑 Monitoreo [%s] detenido: %v\n", servicio.Nombre, ctx.Err())
			return
		}
	}
}

// ─────────────────────────────────────────────────────────────
//  Simulación de servicios (sin necesidad de internet)
// ─────────────────────────────────────────────────────────────

// simularVerificacion simula la verificación de un servicio
// Útil para practicar sin depender de servicios externos
func simularVerificacion(ctx context.Context, servicio Servicio) ResultadoMonitoreo {
	resultado := ResultadoMonitoreo{
		Servicio:  servicio.Nombre,
		Timestamp: time.Now(),
	}

	inicio := time.Now()

	// Simulamos latencia variable (100ms a 3000ms)
	latenciaSimulada := time.Duration(100+rand.Intn(2900)) * time.Millisecond

	// SELECT 1: ¿Termina la simulación o se acaba el timeout?
	done := make(chan struct{})
	go func() {
		time.Sleep(latenciaSimulada)
		close(done)
	}()

	select {
	case <-done:
		// La "verificación" terminó
		resultado.Latencia = time.Since(inicio)

		// Simulamos que ~70% son exitosos, ~20% lentos, ~10% fallan
		azar := rand.Intn(10)
		switch {
		case azar < 7:
			resultado.Estado = "✅ OK"
			resultado.StatusCode = 200
		case azar < 9:
			resultado.Estado = "⚠️ LENTO"
			resultado.StatusCode = 200
			resultado.Latencia = latenciaSimulada
		default:
			resultado.Estado = "❌ ERROR"
			resultado.Error = "connection refused"
			resultado.StatusCode = 503
		}

	case <-ctx.Done():
		// El timeout se agotó antes de que la simulación terminara
		resultado.Estado = "⏰ TIMEOUT"
		resultado.Error = ctx.Err().Error()
		resultado.Latencia = time.Since(inicio)
	}

	return resultado
}

// monitorearServicioSimulado usa SELECT con simulación local
func monitorearServicioSimulado(
	ctx context.Context,
	servicio Servicio,
	resultados chan<- ResultadoMonitoreo,
	intervalo time.Duration,
) {
	cb := NuevoCircuitBreaker(servicio.Nombre, 3, 10*time.Second)
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()

	fmt.Printf("    👁️  Monitoreando [%s] cada %v\n", servicio.Nombre, intervalo)

	for {
		select {
		case <-ticker.C:
			if !cb.PermitirVerificacion() {
				resultados <- ResultadoMonitoreo{
					Servicio:  servicio.Nombre,
					Estado:    "⏭️ SKIP (CB abierto)",
					Timestamp: time.Now(),
				}
				continue
			}

			// Timeout de 2 segundos para cada verificación
			ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
			resultado := simularVerificacion(ctxTimeout, servicio)
			cancel()

			if resultado.Error != "" {
				cb.RegistrarFallo()
			} else {
				cb.RegistrarExito()
			}

			resultados <- resultado

		case <-ctx.Done():
			fmt.Printf("    🛑 Monitoreo [%s] detenido\n", servicio.Nombre)
			return
		}
	}
}

// ─────────────────────────────────────────────────────────────
//  Recolector de Resultados
// ─────────────────────────────────────────────────────────────

// recolectarResultados consume el channel de resultados y genera el reporte
func recolectarResultados(resultados <-chan ResultadoMonitoreo, wg *sync.WaitGroup) {
	defer wg.Done()

	estadisticas := make(map[string]int)
	totalChecks := 0

	for resultado := range resultados {
		totalChecks++
		estadisticas[resultado.Estado]++

		emoji := ""
		switch {
		case strings.Contains(resultado.Estado, "OK"):
			emoji = "✅"
		case strings.Contains(resultado.Estado, "ERROR"):
			emoji = "❌"
		case strings.Contains(resultado.Estado, "TIMEOUT"):
			emoji = "⏰"
		case strings.Contains(resultado.Estado, "LENT"):
			emoji = "🐌"
		case strings.Contains(resultado.Estado, "SKIP"):
			emoji = "⏭️"
		default:
			emoji = "❓"
		}

		fmt.Printf("    %s [%s] %s — %v",
			emoji, resultado.Servicio, resultado.Estado, resultado.Latencia)
		if resultado.Error != "" {
			fmt.Printf(" (%s)", resultado.Error)
		}
		fmt.Println()
	}

	// Reporte final
	fmt.Println()
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Println("    📊 REPORTE FINAL DE MONITOREO")
	fmt.Println("    ═══════════════════════════════════════════════════")
	fmt.Printf("    📋 Total de chequeos: %d\n", totalChecks)
	for estado, count := range estadisticas {
		fmt.Printf("       %s: %d\n", estado, count)
	}
	fmt.Println("    ═══════════════════════════════════════════════════")
}

// ─────────────────────────────────────────────────────────────
//  Demostración de SELECT puro
// ─────────────────────────────────────────────────────────────

// demoBasicoSelect muestra los fundamentos de SELECT
func demoBasicoSelect() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🎵 DEMO 1: Select Básico — Múltiples Channels")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	ch1 := make(chan string)
	ch2 := make(chan string)

	// Dos goroutines que envían con diferentes velocidades
	go func() {
		time.Sleep(100 * time.Millisecond)
		ch1 <- "📡 Respuesta del Servidor A"
	}()

	go func() {
		time.Sleep(300 * time.Millisecond)
		ch2 <- "📡 Respuesta del Servidor B"
	}()

	// SELECT: ¿quién responde primero?
	for i := 0; i < 2; i++ {
		select {
		case msg := <-ch1:
			fmt.Printf("    ch1 dice: %s\n", msg)
		case msg := <-ch2:
			fmt.Printf("    ch2 dice: %s\n", msg)
		}
	}
	fmt.Println()
}

// demoTimeoutSelect muestra el patrón timeout con select
func demoTimeoutSelect() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  ⏱️  DEMO 2: Select con Timeout")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	servicioLento := make(chan string)

	// Simulamos un servicio que tarda 2 segundos
	go func() {
		time.Sleep(2 * time.Second)
		servicioLento <- "Datos del servicio"
	}()

	// SELECT con timeout de 1 segundo
	select {
	case resultado := <-servicioLento:
		fmt.Printf("    ✅ Recibido: %s\n", resultado)
	case <-time.After(1 * time.Second):
		fmt.Println("    ⏰ Timeout: el servicio tardó más de 1 segundo")
	}
	fmt.Println()
}

// demoDefaultSelect muestra el select no-bloqueante con default
func demoDefaultSelect() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🏃 DEMO 3: Select No-Bloqueante (default)")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	ch := make(chan string, 1)

	// Intentamos recibir sin bloquear
	select {
	case msg := <-ch:
		fmt.Printf("    📬 Mensaje recibido: %s\n", msg)
	default:
		fmt.Println("    📭 No hay mensajes — continúo haciendo otra cosa")
	}

	// Ahora enviamos algo
	ch <- "¡Hola!"

	select {
	case msg := <-ch:
		fmt.Printf("    📬 Mensaje recibido: %s\n", msg)
	default:
		fmt.Println("    📭 No hay mensajes")
	}
	fmt.Println()
}

// demoCancelacionSelect muestra cómo cancelar operaciones con context
func demoCancelacionSelect() {
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🛑 DEMO 4: Select con Cancelación (context)")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	trabajo := make(chan string)

	// Simulamos trabajo que toma 2 segundos
	go func() {
		time.Sleep(2 * time.Second)
		trabajo <- "Trabajo completado"
	}()

	// SELECT: ¿el trabajo termina o nos cancelan?
	select {
	case resultado := <-trabajo:
		fmt.Printf("    ✅ %s\n", resultado)
	case <-ctx.Done():
		fmt.Printf("    🛑 Cancelado: %v\n", ctx.Err())
	}
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────
//  Main
// ─────────────────────────────────────────────────────────────

func main() {
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🎵 SELECT: El Director de Orquesta Concurrente")
	fmt.Println("  📚 Lección 12 — Laboratorio de Go")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// ── Parte 1: Demos de select ──────────────────────────
	demoBasicoSelect()
	demoTimeoutSelect()
	demoDefaultSelect()
	demoCancelacionSelect()

	// ── Parte 2: Monitor de Servicios con Circuit Breaker ─
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  🔍 DEMO 5: Monitor de Servicios + Circuit Breaker")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	// Servicios a monitorear (simulados)
	servicios := []Servicio{
		{Nombre: "API-Usuarios", URL: "https://example.com/users"},
		{Nombre: "API-Pagos", URL: "https://example.com/payments"},
		{Nombre: "API-Notificaciones", URL: "https://example.com/notifications"},
	}

	// Context principal con cancelación — 15 segundos de monitoreo
	duracionMonitoreo := 15 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), duracionMonitoreo)
	defer cancel()

	// Channel donde convergen todos los resultados (fan-in)
	resultados := make(chan ResultadoMonitoreo)

	// Lanzamos una goroutine monitoreadora por servicio
	intervalo := 2 * time.Second
	for _, svc := range servicios {
		go monitorearServicioSimulado(ctx, svc, resultados, intervalo)
	}

	// Recolector de resultados en el main
	var wg sync.WaitGroup
	wg.Add(1)
	go recolectarResultados(resultados, &wg)

	// Esperamos a que el context se cancele (timeout de 15s)
	<-ctx.Done()
	fmt.Println()
	fmt.Println("    ⏱️  Tiempo de monitoreo agotado, cerrando...")
	close(resultados)
	wg.Wait()

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("  ✅ Monitor de Servicios finalizado")
	fmt.Println("  📚 Conceptos demostrados:")
	fmt.Println("     • select multiplexando channels")
	fmt.Println("     • time.After para timeouts")
	fmt.Println("     • context.WithTimeout para cancelación")
	fmt.Println("     • Circuit Breaker para resiliencia")
	fmt.Println("     • Fan-in de resultados concurrentes")
	fmt.Println("═══════════════════════════════════════════════════════════")
}