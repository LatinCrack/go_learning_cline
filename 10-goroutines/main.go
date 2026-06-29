package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ============================================================
// Lección 10 — Goroutines: Miles de Hilos por Centavo
// Ejercicio práctico: Scanner de Puertos Concurrente
// ============================================================
//
// Este programa escanea puertos TCP de un host usando miles
// de goroutines simultáneas. Demuestra:
//   - Lanzamiento masivo de goroutines
//   - sync.WaitGroup para sincronización
//   - sync.Mutex para proteger datos compartidos
//   - Patrón semáforo con channels buffered para limitar concurrencia
//   - Closures y la trampa de captura de variables
//
// Uso:
//   go run main.go [host] [puerto_inicial] [puerto_final] [max_concurrencia]
//
// Ejemplos:
//   go run main.go localhost
//   go run main.go localhost 80 443
//   go run main.go localhost 1 1024 200
// ============================================================

// ScanResult almacena el resultado de escanear un puerto.
// Guardamos el número de puerto, si está abierto, y cuánto
// tardó en responder la conexión TCP.
type ScanResult struct {
	Port    int
	IsOpen  bool
	Latency time.Duration
}

// escanearPuerto intenta una conexión TCP a un puerto específico.
//
// Parámetros:
//   - host: dirección del host (IP o nombre)
//   - port: número de puerto a escanear
//   - results: channel donde enviamos el resultado
//
// Usa net.DialTimeout para no quedarse colgado indefinidamente.
// Si la conexión tiene éxito, el puerto está ABIERTO.
// Si falla (timeout, conexión rechazada), el puerto está CERRADO.
func escanearPuerto(host string, port int, results chan<- ScanResult) {
	// Construimos la dirección en formato "host:puerto"
	direccion := fmt.Sprintf("%s:%d", host, port)

	// Intentamos conectar con un timeout de 1 segundo.
	// Sin timeout, una conexión a un puerto filtrado podría
	// bloquear durante minutos.
	inicio := time.Now()
	conn, err := net.DialTimeout("tcp", direccion, 1*time.Second)
	latencia := time.Since(inicio)

	resultado := ScanResult{
		Port:    port,
		Latency: latencia,
	}

	if err != nil {
		// No se pudo conectar → puerto cerrado, filtrado o no existe
		resultado.IsOpen = false
	} else {
		// Conexión exitosa → puerto abierto y escuchando
		resultado.IsOpen = true
		conn.Close() // Cerramos inmediatamente, solo queríamos verificar
	}

	// Enviamos el resultado al channel.
	// Si el channel tiene buffer disponible, esto no bloquea.
	// Si el buffer está lleno, esta goroutine espera hasta que
	// haya espacio (lo cual controla la concurrencia).
	results <- resultado
}

// escanearHost orquesta el escaneo concurrente de un rango de puertos.
//
// Parámetros:
//   - host: dirección del host a escanear
//   - startPort: puerto inicial del rango
//   - endPort: puerto final del rango
//   - maxConcurrency: máximo de goroutines ejecutándose simultáneamente
//
// Retorna un slice con los resultados de puertos ABIERTOS.
//
// Usa tres mecanismos de sincronización:
//  1. sync.WaitGroup → para saber cuándo terminaron TODAS las goroutines
//  2. sync.Mutex → para proteger el slice compartido de resultados
//  3. Channel semáforo → para limitar cuántas goroutines corren a la vez
func escanearHost(host string, startPort, endPort int, maxConcurrency int) []ScanResult {
	var (
		allResults []ScanResult // Slice compartido (protegido por Mutex)
		mu         sync.Mutex   // Cerrojo para acceso seguro al slice
		wg         sync.WaitGroup
	)

	// --------------------------------------------------------
	// Patrón SEMÁFORO con channel buffered
	// --------------------------------------------------------
	// Un channel con buffer de tamaño `maxConcurrency` actúa
	// como un semáforo: permite que solo N goroutines "pasen"
	// al mismo tiempo.
	//
	// Analogía: una disco con aforo máximo. Si ya hay 100
	// personas adentro, los demás esperan en la fila hasta
	// que alguien salga.
	semaphore := make(chan struct{}, maxConcurrency)

	// Iteramos sobre cada puerto del rango
	for port := startPort; port <= endPort; port++ {
		wg.Add(1) // Registramos una goroutine pendiente

		// -------------------------------------------------
		// IMPORTANTE: pasamos `port` como parámetro `p`
		// -------------------------------------------------
		// Si no lo hacemos, todas las goroutines capturan
		// la MISMA variable `port` del bucle, y cuando
		// ejecuten, `port` ya habrá avanzado al final.
		// Esta es la trampa clásica de closures en Go.
		go func(p int) {
			defer wg.Done() // Al terminar, decrementamos el WaitGroup

			// Tomamos un "permiso" del semáforo.
			// Si el buffer está lleno (ya hay maxConcurrency
			// goroutines activas), esta línea BLOQUEA hasta
			// que haya un permiso disponible.
			semaphore <- struct{}{}

			// Al terminar la goroutine, devolvemos el permiso
			// para que otra goroutine pueda entrar.
			defer func() { <-semaphore }()

			// Creamos un channel local para recibir el resultado
			// de escanearPuerto. Usamos una función anónima que
			// lanza la goroutine y retorna el channel (patrón Future).
			ch := make(chan ScanResult, 1)
			go escanearPuerto(host, p, ch)

			// Esperamos el resultado del escaneo
			resultado := <-ch

			// Solo guardamos puertos ABIERTOS
			if resultado.IsOpen {
				// -------------------------------------------------
				// Mutex: acceso exclusivo al slice compartido
				// -------------------------------------------------
				// Sin el Mutex, dos goroutines podrían intentar
				// hacer append al mismo tiempo y corromper el slice.
				mu.Lock()
				allResults = append(allResults, resultado)
				mu.Unlock()
			}
		}(port) // ← Pasamos el valor actual de port
	}

	// Bloqueamos hasta que TODAS las goroutines terminen.
	// Sin esto, main() podría retornar antes de que
	// se completen los escaneos.
	wg.Wait()

	return allResults
}

func main() {
	// --------------------------------------------------------
	// Valores por defecto
	// --------------------------------------------------------
	host := "localhost"
	startPort := 1
	endPort := 1024
	maxConcurrency := 100

	// --------------------------------------------------------
	// Parseo de argumentos de línea de comandos
	// --------------------------------------------------------
	// os.Args[0] es el nombre del programa, los demás son argumentos
	if len(os.Args) >= 2 {
		host = os.Args[1]
	}
	if len(os.Args) >= 3 {
		p, err := strconv.Atoi(os.Args[2])
		if err == nil {
			startPort = p
		}
	}
	if len(os.Args) >= 4 {
		p, err := strconv.Atoi(os.Args[3])
		if err == nil {
			endPort = p
		}
	}
	if len(os.Args) >= 5 {
		p, err := strconv.Atoi(os.Args[4])
		if err == nil {
			maxConcurrency = p
		}
	}

	// --------------------------------------------------------
	// Banner informativo
	// --------------------------------------------------------
	totalPorts := endPort - startPort + 1
	fmt.Printf("🔍 Escaneando %s — puertos %d a %d (%d puertos)\n",
		host, startPort, endPort, totalPorts)
	fmt.Printf("⚡ Concurrencia máxima: %d goroutines simultáneas\n", maxConcurrency)
	fmt.Println(strings.Repeat("─", 50))

	// --------------------------------------------------------
	// Ejecutamos el escaneo y medimos el tiempo
	// --------------------------------------------------------
	inicio := time.Now()
	resultados := escanearHost(host, startPort, endPort, maxConcurrency)
	duracion := time.Since(inicio)

	// --------------------------------------------------------
	// Ordenamos los resultados por número de puerto
	// --------------------------------------------------------
	// sort.Slice usa una función de comparación para ordenar
	sort.Slice(resultados, func(i, j int) bool {
		return resultados[i].Port < resultados[j].Port
	})

	// --------------------------------------------------------
	// Mostramos los resultados
	// --------------------------------------------------------
	fmt.Println()
	if len(resultados) == 0 {
		fmt.Println("❌ No se encontraron puertos abiertos.")
	} else {
		fmt.Printf("✅ Puertos abiertos encontrados: %d\n\n", len(resultados))
		fmt.Printf("%-8s %-12s %s\n", "PUERTO", "ESTADO", "LATENCIA")
		fmt.Println(strings.Repeat("─", 35))
		for _, r := range resultados {
			fmt.Printf("%-8d %-12s %v\n", r.Port, "🟢 ABIERTO", r.Latency)
		}
	}

	// --------------------------------------------------------
	// Estadísticas de rendimiento
	// --------------------------------------------------------
	fmt.Printf("\n⏱️  Tiempo total: %v\n", duracion)
	fmt.Printf("📊 Puertos escaneados: %d\n", totalPorts)
	throughput := float64(totalPorts) / duracion.Seconds()
	fmt.Printf("🚀 Throughput: %.0f puertos/segundo\n", throughput)
}