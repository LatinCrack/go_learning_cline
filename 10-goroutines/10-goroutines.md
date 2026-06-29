# 🦄 Lección 10 — Goroutines: Miles de Hilos por Centavo

## 📖 ¿Qué son las Goroutines?

Imagina que eres el gerente de un restaurante con **10,000 clientes** al mismo tiempo. Si cada cliente necesita su propio mesero físico (un **hilo del sistema operativo**), necesitarías contratar 10,000 meseros. Eso es imposible: cada mesero ocupa espacio, consume recursos, y el gerente (el OS) se vuelve loco coordinando tantos.

Ahora imagina algo diferente: **meseros fantasma ultrarrápidos**. Cada mesero fantasma pesa casi nada (2 kilobytes en vez de 1 megabyte), puede atender miles de mesas simultáneamente, y salta de mesa en mesa cuando un cliente está leyendo el menú (esperando una operación de I/O). Con solo 4 meseros físicos reales (CPUs), puedes tener **100,000 meseros fantasma** atendiendo clientes sin que el restaurante colapse.

> 🦄 **Eso es una goroutine.** Es un hilo verde gestionado por el runtime de Go, no por el sistema operativo. Son extremadamente ligeras y el scheduler de Go las multiplexa sobre un número limitado de threads del OS.

```
  ┌──────────────────────────────────────────────────────┐
  │              SISTEMA OPERATIVO                        │
  │                                                      │
  │   Thread 1    Thread 2    Thread 3    Thread 4        │
  │   ┌─────┐    ┌─────┐    ┌─────┐    ┌─────┐          │
  │   │G│G│G│    │G│G│G│    │G│G│G│    │G│G│G│          │
  │   │G│G│G│    │G│G│G│    │G│G│G│    │G│G│G│          │
  │   │G│G│ │    │G│G│ │    │G│ │ │    │G│ │ │          │
  │   └─────┘    └─────┘    └─────┘    └─────┘          │
  │                                                      │
  │   Miles de goroutines (G) compartiendo pocos threads  │
  └──────────────────────────────────────────────────────┘
```

---

## 🧠 ¿Por qué Go creó las Goroutines?

Go fue diseñado en Google para resolver un problema concreto: **escribir servidores que manejen miles de conexiones simultáneas** sin volverse locos con la complejidad.

| Enfoque | Lenguaje | Problema |
|---------|----------|----------|
| **Hilos del OS** | Java, C++ | Cada hilo pesa ~1MB de stack. 10,000 hilos = 10GB de RAM. Imposible. |
| **Event loop** | Node.js | Un solo hilo. Si una operación bloquea, TODO se detiene. |
| **Goroutines** | Go | ~2KB por goroutine. 100,000 goroutines = ~200MB. El scheduler las gestiona automáticamente. |

> 🧠 **Analogía Feynman**: Los hilos del OS son como taxis físicos: cada taxi cuesta dinero, ocupa espacio en la calle, y hay un límite de taxis que la ciudad puede manejar. Las goroutines son como viajes en Uber que se crean instantáneamente: pesan casi nada, el algoritmo (scheduler) los asigna automáticamente a los taxis disponibles, y puedes tener 100,000 viajes simultáneos con solo 4 taxis físicos.

---

## 🔑 Tu Primera Goroutine

En Go, lanzar una goroutine es ridículamente simple. Solo necesitas la palabra clave `go`:

```go
func decirHola() {
    fmt.Println("¡Hola desde una goroutine!")
}

func main() {
    go decirHola()  // ← Lanza como goroutine
    fmt.Println("Hola desde main")
}
```

### ⚠️ ¡Cuidado! Este programa tiene un bug

Si ejecutas el código anterior, es posible que **nunca veas** "¡Hola desde una goroutine!". ¿Por qué?

Porque cuando `main()` termina, **todas las goroutines se destruyen inmediatamente**. Es como si el restaurante cerrara: los meseros fantasma desaparecen aunque estén atendiendo clientes.

```
  Timeline:
  ──────────────────────────────────────────►
  main() inicia
    │
    ├── go decirHola()  ← Se lanza, pero...
    │
    ├── fmt.Println("Hola desde main")  ← Se ejecuta
    │
    └── main() termina  ← 💀 Todas las goroutines mueren
                           decirHola() nunca corrió :(
```

### ✅ La solución: esperar con `time.Sleep` (por ahora)

```go
func main() {
    go decirHola()
    time.Sleep(1 * time.Second)  // Espera (no es la solución ideal)
    fmt.Println("Hola desde main")
}
```

> ⚠️ `time.Sleep` NO es la forma correcta de sincronizar goroutines. Es como cerrar los ojos y esperar 1 segundo a que alguien termine de hablar. En la siguiente lección aprenderemos **channels**, que es la forma profesional de comunicarse entre goroutines. Por ahora, `time.Sleep` nos sirve para aprender.

---

## 🏗️ Cómo Funciona el Scheduler de Go

El scheduler de Go usa un modelo llamado **M:N**:

```
  M goroutines  ←→  N threads del OS
```

### Los tres actores del scheduler

| Actor | Nombre | Descripción |
|-------|--------|-------------|
| **G** | Goroutine | Tu código. Tiene su propio stack (2KB inicial, crece si es necesario). |
| **M** | Machine | Un thread del OS real. Ejecuta goroutines. |
| **P** | Processor | Un "permiso para ejecutar". Por defecto hay `GOMAXPROCS` P's (= número de CPUs). |

```
  ┌─────────────────────────────────────────┐
  │            Runtime de Go                │
  │                                         │
  │   P0              P1                    │
  │   ┌───┐           ┌───┐                │
  │   │ G │           │ G │                │
  │   │ G │ ←runnable │ G │ ←runnable      │
  │   │ G │           │ G │                │
  │   └─┬─┘           └─┬─┘                │
  │     │               │                   │
  │   ┌─▼─┐           ┌─▼─┐                │
  │   │ M0│           │ M1│   ← OS threads │
  │   └───┘           └───┘                │
  └─────────────────────────────────────────┘
```

### ¿Qué es GOMAXPROCS?

`GOMAXPROCS` define **cuántas goroutines pueden ejecutarse en paralelo** (no concurrentemente, sino realmente al mismo tiempo). Por defecto es igual al número de CPUs.

```go
import "runtime"

fmt.Println("CPUs:", runtime.NumCPU())           // Ej: 8
fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0)) // Lee sin cambiar: 8
```

> 🧠 **Analogía Feynman**: Imagina una cocina con 4 hornillas (P = 4). Puedes tener 1,000 platos preparados esperando (goroutines en cola), pero solo 4 se cocinan simultáneamente. Cuando un plato espera a que hierva el agua (I/O bloqueante), el chef retira ese plato y pone otro en la hornilla. Así, con solo 4 hornillas, puedes cocinar 1,000 platos increíblemente rápido.

---

## 🔄 Goroutines en Acción: Múltiples Concurrentes

Veamos el poder real de las goroutines:

```go
func tarea(id int) {
    fmt.Printf("Tarea %d iniciada\n", id)
    time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
    fmt.Printf("Tarea %d completada\n", id)
}

func main() {
    for i := 1; i <= 5; i++ {
        go tarea(i)  // Lanza 5 goroutines
    }
    time.Sleep(2 * time.Second) // Espera que terminen
}
```

**Salida posible** (el orden varía cada ejecución):
```
Tarea 3 iniciada
Tarea 1 iniciada
Tarea 5 iniciada
Tarea 2 iniciada
Tarea 4 iniciada
Tarea 3 completada
Tarea 1 completada
Tarea 5 completada
Tarea 4 completada
Tarea 2 completada
```

> 📌 **Observa**: Las tareas se **inician** en orden (o no), y se **completan** en cualquier orden. Esto es **concurrencia**: múltiples tareas progresan simultáneamente, pero el orden de finalización es impredecible.

---

## 🎯 Consecuencia Crítica: La Race Condition

Aquí es donde las goroutines se ponen peligrosas. Mira este código:

```go
var contador int = 0

func incrementar() {
    for i := 0; i < 1000; i++ {
        contador++  // ⚠️ ¡PELIGRO! Race condition
    }
}

func main() {
    go incrementar()
    go incrementar()
    time.Sleep(2 * time.Second)
    fmt.Println("Contador:", contador)
}
```

**¿Qué esperas?** 2000.

**¿Qué obtienes?** Algo como 1247, 1893, 1562... **Nunca 2000**.

### ¿Por qué ocurre?

```
  Goroutine A                    Goroutine B
  ──────────────                 ──────────────
  Lee contador = 500
                                 Lee contador = 500
  Escribe 500 + 1 = 501
                                 Escribe 500 + 1 = 501
  ────────────────────────────────────────────
  Resultado: 501 (debería ser 502)
  Se perdió una operación :(
```

> 🧠 **Analogía Feynman**: Imagina que dos personas están editando el mismo documento de Word al mismo tiempo en la misma computadora. Ambas leen "Tengo 500 manzanas". Ambas deciden agregar 1. Ambas escriben "Tengo 501 manzanas". Se perdió una manzana. Esto es una **race condition**.

### 🛡️ Solución temporal: `sync.Mutex`

```go
var (
    contador int
    mu       sync.Mutex
)

func incrementar() {
    for i := 0; i < 1000; i++ {
        mu.Lock()
        contador++
        mu.Unlock()
    }
}
```

> El `Mutex` (Mutual Exclusion) es como un baño con llave: solo una persona puede entrar a la vez. La goroutine que tiene la llave puede modificar `contador`, las demás esperan afuera.

---

## ⏱️ `sync.WaitGroup`: La Forma Correcta de Esperar

`time.Sleep` es una mala práctica para sincronizar goroutines. `sync.WaitGroup` es la herramienta correcta:

```go
func tarea(id int, wg *sync.WaitGroup) {
    defer wg.Done()  // Decrementa el contador al terminar

    fmt.Printf("Tarea %d iniciada\n", id)
    time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
    fmt.Printf("Tarea %d completada\n", id)
}

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 5; i++ {
        wg.Add(1)       // Incrementa el contador
        go tarea(i, &wg)
    }

    wg.Wait()  // Bloquea hasta que el contador llegue a 0
    fmt.Println("✅ Todas las tareas completadas")
}
```

### ¿Cómo funciona WaitGroup?

```
  WaitGroup counter
  ┌─────────────────────────────────────┐
  │  Start: 0                           │
  │  wg.Add(1) → 1                      │
  │  wg.Add(1) → 2                      │
  │  wg.Add(1) → 3                      │
  │  wg.Add(1) → 4                      │
  │  wg.Add(1) → 5                      │
  │                                     │
  │  wg.Wait() ← bloquea aquí          │
  │                                     │
  │  Tarea 1 termina → wg.Done() → 4   │
  │  Tarea 2 termina → wg.Done() → 3   │
  │  Tarea 3 termina → wg.Done() → 2   │
  │  Tarea 4 termina → wg.Done() → 1   │
  │  Tarea 5 termina → wg.Done() → 0   │
  │                                     │
  │  Counter = 0 → wg.Wait() desbloquea│
  └─────────────────────────────────────┘
```

| Método | Qué hace |
|--------|----------|
| `wg.Add(n)` | Suma n al contador. Llama ANTES de lanzar la goroutine. |
| `wg.Done()` | Resta 1 al contador. Equivalente a `wg.Add(-1)`. Llama al final de la goroutine (idealmente con `defer`). |
| `wg.Wait()` | Bloquea hasta que el contador sea 0. |

---

## 🏋️ Ejercicio Práctico: Scanner de Puertos Concurrente

Vamos a construir algo real y útil: un **escáner de puertos TCP** que lance miles de goroutines para verificar qué puertos están abiertos en un host.

### ¿Por qué un escáner de puertos?

Porque es el ejemplo perfecto de **I/O-bound concurrency**:
- Cada conexión es **independiente**
- La mayoría del tiempo se **espera la respuesta de red** (I/O)
- Miles de goroutines pueden ejecutarse simultáneamente sin consumir CPU significativa
- Es una herramienta que los ingenieros de seguridad usan **todos los días**

### El código completo explicado línea por línea

```go
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
```

> Importamos `net` para conexiones TCP, `sync` para WaitGroup y Mutex, `sort` para ordenar resultados, y los demás para formateo y parsing.

---

```go
// ScanResult almacena el resultado de escanear un puerto
type ScanResult struct {
	Port    int  // Número del puerto
	IsOpen  bool // ¿Está abierto?
	Latency time.Duration
}
```

> Un struct simple para encapsular el resultado de cada escaneo. Guardamos el puerto, si está abierto, y cuánto tardó la conexión.

---

```go
// escanearPuerto intenta conectarse a un puerto específico
func escanearPuerto(host string, port int, results chan<- ScanResult) {
	// Construimos la dirección: "host:puerto"
 direccion := fmt.Sprintf("%s:%d", host, port)

 // Intentamos conectar con timeout de 1 segundo
 inicio := time.Now()
 conn, err := net.DialTimeout("tcp", direccion, 1*time.Second)
 latencia := time.Since(inicio)

 resultado := ScanResult{
 	Port:    port,
 	Latency: latencia,
 }

 if err != nil {
 	// No se pudo conectar → puerto cerrado o filtrado
 	resultado.IsOpen = false
 } else {
 	// Conexión exitosa → puerto abierto
 	resultado.IsOpen = true
 	conn.Close() // Cerramos la conexión inmediatamente
 }

 // Enviamos el resultado al channel
 results <- resultado
}
```

> 🧠 **Punto clave**: Esta función es una goroutine independiente. Cada puerto se escanea en su propia goroutine. El resultado se envía a un **channel** (lo veremos en la Lección 11), que actúa como un buzón de correo donde todas las goroutines depositan sus resultados.

---

```go
// escanearHost escanea un rango de puertos con concurrencia limitada
func escanearHost(host string, startPort, endPort int, maxConcurrency int) []ScanResult {
	var (
		allResults []ScanResult
		mu         sync.Mutex
		wg         sync.WaitGroup
	)
```

> Usamos un `Mutex` para proteger el slice `allResults` (múltiples goroutines escriben en él), y un `WaitGroup` para saber cuándo terminaron todas.

---

```go
	// Channel semáforo para limitar concurrencia
	semaphore := make(chan struct{}, maxConcurrency)
```

> 🧠 **Patrón del semáforo**: Un channel con buffer actúa como un semáforo. Tiene `maxConcurrency` "permisos". Cada goroutine debe "tomar un permiso" antes de ejecutarse y "devolverlo" al terminar. Si no hay permisos disponibles, la goroutine espera.

```
  Semaphore (buffer = 100)
  ┌──────────────────────────────┐
  │ 🟢🟢🟢🟢🟢🟢🟢🟢🟢🟢         │ ← 10 permisos usados
  │ ⚪⚪⚪⚪⚪⚪⚪⚪⚪⚪⚪⚪⚪⚪⚪... │ ← 90 permisos libres
  └──────────────────────────────┘
  Si todos los permisos están en uso,
  la siguiente goroutine espera.
```

---

```go
	for port := startPort; port <= endPort; port++ {
		wg.Add(1) // Contamos esta tarea

		go func(p int) {
			defer wg.Done()    // Al terminar, decrementamos el contador
			semaphore <- struct{}{} // Tomamos un permiso del semáforo
			defer func() { <-semaphore }() // Al terminar, devolvemos el permiso

			// Escaneamos el puerto
			resultado := <-func() chan ScanResult {
				ch := make(chan ScanResult, 1)
				go escanearPuerto(host, p, ch)
				return ch
			}()

			// Si el puerto está abierto, lo guardamos
			if resultado.IsOpen {
				mu.Lock()
				allResults = append(allResults, resultado)
				mu.Unlock()
			}
		}(port) // ← ¡Importante! Pasamos `port` como parámetro
	}

	wg.Wait() // Esperamos a que todas las goroutines terminen
	return allResults
}
```

> ⚠️ **Trampa clásica de closures**: Si no pasamos `port` como parámetro a la goroutine, todas las goroutines capturarían la misma variable `port` y escanearían el mismo puerto. Siempre pasa las variables del bucle como argumentos a las goroutines.

---

```go
func main() {
	// Valores por defecto
	host := "localhost"
	startPort := 1
	endPort := 1024
	maxConcurrency := 100

	// Parsear argumentos de línea de comandos
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
```

> Aceptamos argumentos opcionales: host, puerto inicial, puerto final, y concurrencia máxima.

---

```go
	totalPorts := endPort - startPort + 1
	fmt.Printf("🔍 Escaneando %s — puertos %d a %d (%d puertos)\n",
		host, startPort, endPort, totalPorts)
	fmt.Printf("⚡ Concurrencia máxima: %d goroutines simultáneas\n", maxConcurrency)
	fmt.Println(strings.Repeat("─", 50))

	inicio := time.Now()
	resultados := escanearHost(host, startPort, endPort, maxConcurrency)
	duracion := time.Since(inicio)

	// Ordenar resultados por número de puerto
	sort.Slice(resultados, func(i, j int) bool {
		return resultados[i].Port < resultados[j].Port
	})

	// Mostrar resultados
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

	fmt.Printf("\n⏱️  Tiempo total: %v\n", duracion)
	fmt.Printf("📊 Puertos escaneados: %d\n", totalPorts)
	throughput := float64(totalPorts) / duracion.Seconds()
	fmt.Printf("🚀 Throughput: %.0f puertos/segundo\n", throughput)
}
```

> 📌 **Observa el throughput**: Con 100 goroutines, escanear 1024 puertos toma ~2 segundos. Secuencialmente tomaría ~1024 segundos (17 minutos). Eso es el poder de la concurrencia.

---

## 🧪 Ejecutando el Scanner

```bash
# Escanear localhost, puertos 1-1024
go run main.go localhost

# Escanear un host específico, puertos 80-443
go run main.go example.com 80 443

# Escanear con 200 goroutines simultáneas
go run main.go localhost 1 1000 200
```

**Salida de ejemplo:**
```
🔍 Escaneando localhost — puertos 1 a 1024 (1024 puertos)
⚡ Concurrencia máxima: 100 goroutines simultáneas
──────────────────────────────────────────────────

✅ Puertos abiertos encontrados: 3

PUERTO   ESTADO       LATENCIA
───────────────────────────────────
80       🟢 ABIERTO   1.2ms
443      🟢 ABIERTO   0.8ms
8080     🟢 ABIERTO   1.5ms

⏱️  Tiempo total: 2.34s
📊 Puertos escaneados: 1024
🚀 Throughput: 437 puertos/segundo
```

---

## 📊 Comparación: Secuencial vs Concurrente

| Métrica | Secuencial | 100 Goroutines |
|---------|-----------|----------------|
| Tiempo (1024 puertos) | ~17 minutos | ~2 segundos |
| Memoria | ~1MB | ~2MB |
| Líneas de código | ~30 | ~50 |
| Complejidad | Simple | Moderada |

> 🧠 **Lección clave**: Las goroutines no son mágicas — solo son eficientes cuando las tareas son **I/O-bound** (pasan más tiempo esperando que procesando). Si las tareas fueran **CPU-bound** (cálculos matemáticos intensivos), las goroutines no te darían más velocidad que `GOMAXPROCS` (número de CPUs).

---

## 📋 Resumen Visual

```
  ┌─────────────────────────────────────────────────────┐
  │               GOROUTINES EN GO                      │
  ├─────────────────────────────────────────────────────┤
  │                                                     │
  │  go func()  →  Lanza una goroutine                  │
  │                                                     │
  │  Tamaño: ~2KB (vs ~1MB de un hilo de OS)            │
  │                                                     │
  │  Scheduler M:N:                                     │
  │    G (Goroutine) → P (Processor) → M (Machine/OS)   │
  │                                                     │
  │  GOMAXPROCS = número de CPUs (por defecto)          │
  │                                                     │
  │  ⚠️  Race condition: goroutines comparten memoria    │
  │  🛡️  sync.Mutex: protege acceso concurrente         │
  │  ✅  sync.WaitGroup: espera a que terminen           │
  │                                                     │
  │  ⚠️  main() termina → todas las goroutines mueren    │
  │  ⚠️  Closures en bucles → pasa variables como args   │
  │                                                     │
  └─────────────────────────────────────────────────────┘
```

---

## 🧩 Cómo Ejecutar Este Ejercicio

```bash
cd 10-goroutines
go run main.go localhost
```

---

## 🧪 Ejercicio Feynman

> **Instrucción**: Explica estos conceptos con tus propias palabras, como si le enseñaras a alguien que nunca ha programado. Usa tus propias analogías. Si no puedes explicarlo de forma simple, es que no lo entendiste lo suficiente.

### ✍️ Ejercicio 1: Explica con una analogía
> "¿Qué es una goroutine y por qué es mejor que un hilo del sistema operativo?"

**Tu respuesta** (escribe aquí):
```
Tu analogía aquí...

```

---

### ✍️ Ejercicio 2: ¿Qué imprime este código?
```go
func main() {
    for i := 0; i < 3; i++ {
        go func() {
            fmt.Println(i)
        }()
    }
    time.Sleep(time.Second)
}
```

**Tu respuesta** (escribe aquí):
```
Imprime:
¿Es predecible? ¿Por qué?

```

<details>
<summary>🔍 Ver solución</summary>

**Imprime algo como:**
```
3
3
3
```

**Explicación:** La goroutine closure captura la **variable** `i`, no su **valor**. Cuando las goroutines ejecutan, el bucle `for` ya terminó y `i` vale 3. Todas imprimen 3.

**La solución correcta:**
```go
for i := 0; i < 3; i++ {
    go func(valor int) {
        fmt.Println(valor)
    }(i)  // Pasamos el valor actual de i
}
```

Ahora imprime 0, 1, 2 (en orden impredecible).
</details>

---

### ✍️ Ejercicio 3: Race Condition — ¿Por qué este código falla?
```go
var total int

func sumar(nums []int, wg *sync.WaitGroup) {
    defer wg.Done()
    for _, n := range nums {
        total += n  // ⚠️ Problema aquí
    }
}

func main() {
    var wg sync.WaitGroup
    datos := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

    wg.Add(2)
    go sumar(datos[:5], &wg)
    go sumar(datos[5:], &wg)
    wg.Wait()

    fmt.Println("Total:", total)
}
```

**Tu respuesta** (escribe aquí):
```
¿Cuál es el problema?
¿Cuál sería la solución?

```

<details>
<summary>🔍 Ver solución</summary>

**Problema:** Dos goroutines modifican `total` simultáneamente sin protección. Esto es una **race condition**. El resultado puede ser cualquier número entre 30 y 55.

**Solución con Mutex:**
```go
var (
    total int
    mu    sync.Mutex
)

func sumar(nums []int, wg *sync.WaitGroup) {
    defer wg.Done()
    for _, n := range nums {
        mu.Lock()
        total += n
        mu.Unlock()
    }
}
```

**Solución mejorada** (cada goroutine calcula su subtotal localmente):
```go
func sumar(nums []int, wg *sync.WaitGroup, resultado *int) {
    defer wg.Done()
    subtotal := 0
    for _, n := range nums {
        subtotal += n
    }
    *resultado = subtotal
}
```
</details>

---

### ✍️ Ejercicio 4: Implementa tu propio código
> Crea un programa que lance 10 goroutines, cada una calculando el factorial de un número del 1 al 10. Usa `sync.WaitGroup` para esperar a todas y muestra los resultados ordenados.

**Tu respuesta** (escribe aquí):
```go
// Escribe tu código aquí

```

<details>
<summary>🔍 Ver solución sugerida</summary>

```go
package main

import (
	"fmt"
	"sync"
)

func factorial(n int, wg *sync.WaitGroup, resultados chan string) {
	defer wg.Done()

	resultado := 1
	for i := 2; i <= n; i++ {
		resultado *= i
	}

	resultados <- fmt.Sprintf("%d! = %d", n, resultado)
}

func main() {
	var wg sync.WaitGroup
	resultados := make(chan string, 10)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go factorial(i, &wg, resultados)
	}

	// Cerramos el channel cuando todas las goroutines terminen
	go func() {
		wg.Wait()
		close(resultados)
	}()

	// Leemos los resultados
	for r := range resultados {
		fmt.Println(r)
	}
}
```
</details>

---

### ✍️ Ejercicio 5: Explica el concepto clave
> "¿Por qué Go eligió goroutines en vez de usar hilos del OS como Java? ¿Cuándo las goroutines NO son la solución ideal?"

**Tu respuesta** (escribe aquí):
```
¿Por qué goroutines en vez de hilos del OS?
¿Cuándo NO son ideales?

```

<details>
<summary>🔍 Ver respuesta sugerida</summary>

**¿Por qué goroutines?**
- Los hilos del OS pesan ~1MB cada uno. Crear 100,000 hilos = 100GB de RAM. Imposible.
- Las goroutines pesan ~2KB. Crear 100,000 goroutines = ~200MB. Factible.
- El scheduler de Go gestiona las goroutines automáticamente. El programador no necesita crear thread pools ni gestionar manualmente la concurrencia.
- Las goroutines se comunican con channels (Lección 11), que son más seguros que compartir memoria con locks.

**¿Cuándo NO son ideales?**
- **CPU-bound puro**: Si tus tareas son cálculos matemáticos intensivos sin I/O, las goroutines no te dan más rendimiento que GOMAXPROCS CPUs. Necesitas paralelismo real, no concurrencia.
- **Interoperabilidad con C**: Las goroutines no pueden llamarse desde código C directamente.
- **Sistemas embebidos con RAM extrema limitada**: Aunque 2KB es poco, si tienes un microcontrolador con 4KB de RAM, 100 goroutines ya son demasiado.
</details>

---

## 🚀 Próxima lección

En la **Lección 11** exploraremos los **Channels: El Sistema Nervioso de Go** — el mecanismo que permite a las goroutines comunicarse entre sí de forma segura. Sin channels, las goroutines son como personas en cuartos separados sin teléfono. Los channels son el teléfono.

---

> *"No comuniques compartiendo memoria; comparte memoria comunicando."*
> — **Proverbio de Go** 🦄