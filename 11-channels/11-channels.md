# 🧬 Lección 11 — Channels: El Sistema Nervioso de Go

## 📖 ¿Qué son los Channels?

Imagina un **banco de drive-through**. El cajero (goroutine emisora) pone la cápsula con dinero en la tubería neumática, y el cliente (goroutine receptora) la recibe al otro lado. Si la tubería está llena, el cajero **espera** a que el cliente recoja la cápsula. Si la tubería está vacía, el cliente **espera** a que el cajero envíe algo. No hay pérdida, no hay condición de carrera, no hay confusión.

Eso es un **channel** en Go: una tubería tipada que conecta goroutines.

> 🧬 **Ese es el channel.** Es el mecanismo de comunicación entre goroutines. Son **tipados**, **bloqueantes** por diseño, y siguen la filosofía fundamental de Go: *"Don't communicate by sharing memory; share memory by communicating."*

```
  ┌──────────────┐         Channel         ┌──────────────┐
  │  Goroutine A │ ═══════════════════════▶ │  Goroutine B │
  │  (Emisora)   │    dato → dato → dato   │  (Receptora) │
  │              │ ◀═══════════════════════ │              │
  └──────────────┘    confirmación          └──────────────┘
         │                                        │
         │   "No comparto memoria contigo,        │
         │    te envío datos por la tubería"       │
```

---

## 🧠 ¿Por qué Go creó los Channels?

En la lección anterior vimos que las goroutines comparten memoria. Y compartir memoria es **peligroso**: race conditions, deadlocks, bugs imposibles de reproducir. Los otros lenguajes resuelven esto con **locks** (Mutex, semáforos, etc.), que son como poner una puerta con llave al baño: funciona, pero es frágil y difícil de coordinar.

Go propone algo radical: **no compartas memoria. Comunica.**

| Enfoque | Mecanismo | Problema |
|---------|-----------|----------|
| **Compartir memoria + Locks** | Java, C++, Python | Deadlocks, race conditions, código frágil |
| **Message passing** | Go (channels) | Las goroutines se pasan datos, no comparten estado |

> 🧠 **Analogía Feynman**: Imagina 5 personas que necesitan coordinarse para pintar una casa. El enfoque de "compartir memoria" es como poner un calendario en la pared y que todos intenten escribir al mismo tiempo — chocan codos, se pisan, alguien escribe en la casilla equivocada. El enfoque de "channels" es como un sistema de mensajería: cada persona envía mensajes por un tubo neumático. Si pintas la cocina, envías un mensaje "cocina lista". El que pinta el comedor recibe el mensaje y sabe que puede entrar. **Nunca chocan.**

---

## 🔑 Tu Primer Channel

Crear un channel es trivial. Usarlo requiere entender **una regla de oro**:

> **Un channel bloquea.** El emisor espera hasta que alguien reciba. El receptor espera hasta que alguien envíe.

```go
package main

import "fmt"

func main() {
    // Creamos un channel de strings
    mensajes := make(chan string)

    // Goroutine emisora
    go func() {
        mensajes <- "¡Hola desde la goroutine!" // Envía (bloquea hasta que alguien reciba)
    }()

    // main() recibe
    msg := <-mensajes // Recibe (bloquea hasta que alguien envíe)
    fmt.Println(msg)
}
```

**Salida:**
```
¡Hola desde la goroutine!
```

### 🔍 ¿Qué acaba de pasar?

```
  Timeline:
  ────────────────────────────────────────────────────────►
  
  main() crea el channel
    │
    ├── go func() { mensajes <- "¡Hola!" }  ← Se bloquea (nadie recibe aún)
    │
    ├── msg := <-mensajes                    ← Recibe, desbloquea a la goroutine
    │
    └── fmt.Println(msg)                     ← Imprime "¡Hola!"
```

La goroutine se **bloqueó** en la línea `mensajes <- "¡Hola!"` porque nadie estaba escuchando. En el momento en que `main()` ejecutó `<-mensajes`, el dato fluyó por la tubería y ambas goroutines continuaron.

---

## 📦 Buffered vs Unbuffered: Dos Tamaños de Tubería

### Unbuffered (sin buffer) — El handshake garantizado

```go
ch := make(chan int) // Sin buffer
```

Es como una **llamada telefónica**: el emisor espera hasta que el receptor contesta. No puedes dejar un mensaje y colgar — si nadie contesta, te quedas esperando.

```
  Emisor                    Channel (vacío)               Receptor
  ────────                  ─────────────                 ─────────
  ch <- 42   ──bloquea──▶   ┃ 42 ┃  ──desbloquea──▶      <-ch
                            (1 slot)
```

### Buffered (con buffer) — El buzón de correo

```go
ch := make(chan int, 3) // Buffer de 3 elementos
```

Es como un **buzón de correo**: puedes dejar hasta 3 cartas sin que nadie las recoja. Solo te bloqueas si el buzón está lleno.

```
  Emisor              Channel (buffer = 3)
  ────────            ┌───┬───┬───┐
  ch <- 1             │ 1 │   │   │  ← No bloquea (hay espacio)
  ch <- 2             │ 1 │ 2 │   │  ← No bloquea (hay espacio)
  ch <- 3             │ 1 │ 2 │ 3 │  ← No bloquea (hay espacio)
  ch <- 4             │ 1 │ 2 │ 3 │  ← ¡BLOQUEA! Buffer lleno
                      └───┴───┴───┘     Debe esperar a que alguien
                                        lea un elemento
```

### Tabla comparativa

| Característica | Unbuffered | Buffered |
|----------------|-----------|----------|
| Creación | `make(chan T)` | `make(chan T, n)` |
| Emisor se bloquea cuando... | Nadie ha recibido | Buffer está lleno |
| Receptor se bloquea cuando... | No hay datos | Buffer está vacío |
| Uso ideal | Sincronización estricta | Desacoplamiento temporal |
| Analogía | Llamada telefónica | Buzón de correo |

> 📌 **Regla práctica**: Empieza con channels sin buffer. Solo usa buffer cuando tengas una razón concreta (rendimiento, desacoplamiento de productor/consumidor). Un buffer de 1 es como un buzón con espacio para una sola carta.

---

## 🚫 Los Tres Panics de los Channels

Los channels tienen 3 comportamientos que causan **panic** (crash del programa). Memorízalos:

```
  ┌──────────────────────────────────────────────────────────────┐
  │              OPERACIÓN         CHANNEL       RESULTADO       │
  ├──────────────────────────────────────────────────────────────┤
  │  ch <- dato                    nil           💥 PANIC        │
  │  <-ch                          nil           💥 PANIC        │
  │  ch <- dato                    cerrado       💥 PANIC        │
  │  <-ch                          cerrado       zero value ✓   │
  │  close(ch)                     cerrado       💥 PANIC        │
  │  close(ch)                     nil           💥 PANIC        │
  └──────────────────────────────────────────────────────────────┘
```

> 🧠 **Analogía Feynman**: Enviar a un channel `nil` es como intentar meter una carta en una puerta que no tiene buzón. Enviar a un channel cerrado es como meter una carta en un buzón con un cartel de "CERRADO PERMANENTEMENTE". **Ambas acciones son errores de programación** — el programa se rompe a propósito para que lo arregles.

---

## 🔒 `close()`: La Señal de "No Hay Más Datos"

Cuando una goroutine termina de producir datos, debe **cerrar el channel** para avisar a los receptores:

```go
func producir(ch chan<- int) {
    for i := 1; i <= 5; i++ {
        ch <- i
    }
    close(ch) // 🔑 "Ya no hay más datos"
}

func main() {
    ch := make(chan int)
    go producir(ch)

    // range se detiene automáticamente cuando el channel se cierra
    for valor := range ch {
        fmt.Println(valor)
    }
    fmt.Println("✅ Canal cerrado, terminamos")
}
```

**Salida:**
```
1
2
3
4
5
✅ Canal cerrado, terminamos
```

```
  Goroutine emisora               Channel                  main()
  ─────────────────               ───────                  ──────
  ch <- 1                         [1]
  ch <- 2                         [1,2]
  ch <- 3                         [1,2,3]
  ch <- 4                         [1,2,3,4]
  ch <- 5                         [1,2,3,4,5]
  close(ch)  ──señal de fin──▶    [CERRADO]  ──range termina──▶  print
```

> 📌 **Regla de oro**: Solo el **emisor** debe cerrar un channel. **Nunca** el receptor. Enviar a un channel cerrado causa panic. Recibir de un channel cerrado devuelve el zero value.

---

## 🔄 `for range` sobre Channels

El `range` sobre un channel lee valores hasta que el channel se cierra:

```go
ch := make(chan int, 5)
ch <- 10
ch <- 20
ch <- 30
close(ch)

for v := range ch {
    fmt.Println(v) // Imprime 10, 20, 30
}
```

> 🧠 **Analogía**: Es como poner un contador de personas en la puerta de un cine. Cada persona que sale (dato recibido) decrementa el contador. Cuando el cine cierra (channel cerrado), el contador para.

---

## 🏗️ Direcciones de Channel: Solo Lectura vs Solo Escritura

Go permite restringir la dirección de un channel en los tipos de funciones. Esto es **poderoso** para prevenir errores:

```go
// Solo puede ENVIAR a este channel
func productor(ch chan<- int) {
    ch <- 42
    // v := <-ch  ← Error de compilación
}

// Solo puede RECIBIR de este channel
func consumidor(ch <-chan int) {
    v := <-ch
    fmt.Println(v)
    // ch <- 42  ← Error de compilación
}

func main() {
    ch := make(chan int)
    go productor(ch)   // chan<- int  (solo escritura)
    consumidor(ch)      // <-chan int (solo lectura)
}
```

| Tipo | Significado | Operaciones permitidas |
|------|------------|----------------------|
| `chan T` | Bidireccional | Enviar y recibir |
| `chan<- T` | Solo escritura | Enviar y cerrar |
| `<-chan T` | Solo lectura | Recibir (no puede cerrar) |

> 📌 **Patrón profesional**: Siempre declara los channels con la restricción más estricta posible en los parámetros de funciones. Esto previene que un worker por accidente cierre un channel que no le pertenece.

---

## 🏋️ Ejercicio Práctico: Pipeline de Descarga y Procesamiento de Archivos

Vamos a construir algo real y útil: un **pipeline de 3 etapas** que descarga archivos de internet, calcula hashes SHA-256, y genera un reporte — todo comunicándose exclusivamente por channels.

### ¿Por qué un pipeline?

Porque el patrón pipeline es el **"Hello World" de la concurrencia en producción**:
- Cada etapa es una goroutine independiente
- Se comunican solo por channels (sin Mutex, sin shared state)
- Puedes escalar cada etapa independientemente
- Es el mismo patrón que usan los sistemas de streaming de datos

```
  ┌─────────────┐    ┌─────────────┐    ┌──────────────┐
  │  ETAPA 1:   │───▶│  ETAPA 2:   │───▶│  ETAPA 3:    │
  │  Generador  │    │  Downloaders│    │  Procesador  │
  │  de URLs    │    │  (3 workers)│    │  de Resultados│
  │  (1 gorout) │    │  (N gorouts)│    │  (1 gorout)  │
  └─────────────┘    └─────────────┘    └──────────────┘
   tasks chan ──────▶ results chan ──────▶ reporte final
```

### El código completo explicado línea por línea

```go
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
```

> Importamos `crypto/sha256` para calcular hashes de los archivos descargados, `net/http` para las descargas HTTP, y `sync` para coordinar los workers internamente.

---

```go
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
```

> Dos structs simples: uno para la tarea (lo que hay que hacer), otro para el resultado (lo que se obtuvo). Los structs viajan por los channels como paquetes en una cinta transportadora.

---

```go
// ETAPA 1: Generador de URLs (Productor)
func generaURLs(urls []string) <-chan DownloadTask {
    out := make(chan DownloadTask) // Channel unbuffered

    go func() {
        for i, url := range urls {
            out <- DownloadTask{
                ID:  i + 1,
                URL: url,
            }
        }
        close(out) // 🔑 Señal: "no hay más tareas"
    }()

    return out // Devolvemos solo-lectura
}
```

> 🧠 **Punto clave**: Esta función **lanza su propia goroutine** y devuelve el channel inmediatamente. El productor trabaja en segundo plano, emitiendo tareas una por una. Cuando termina, `close(out)` le dice a los consumidores "ya no hay más trabajo". Esto es el patrón **Fan-Out** en su forma más pura.

---

```go
// ETAPA 2: Downloader Concurrente (Worker)
func descargaURLs(
    tasks <-chan DownloadTask,    // Entrada: solo lectura
    maxConcurrency int,          // Límite de goroutines simultáneas
) <-chan DownloadResult {
    out := make(chan DownloadResult)

    var wg sync.WaitGroup

    // Lanzamos N workers
    for i := 0; i < maxConcurrency; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()

            // Cada worker consume tareas del MISMO channel
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
```

> 🧠 **Punto clave**: Múltiples goroutines **leen del mismo channel** (`tasks`). Go garantiza que cada tarea sea entregada a **exactamente un** worker — no hay duplicados, no hay locks. Esto es **Fan-In** natural. El `WaitGroup` coordina el cierre del channel de salida: solo se cierra cuando TODOS los workers terminaron de procesar.

---

```go
// descargarArchivo realiza la descarga HTTP real
func descargarArchivo(task DownloadTask) DownloadResult {
    inicio := time.Now()
    result := DownloadResult{TaskID: task.ID, URL: task.URL}

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(task.URL)
    if err != nil {
        result.Success = false
        result.Error = err.Error()
        result.Latency = time.Since(inicio)
        return result
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        result.Success = false
        result.Error = err.Error()
        result.Latency = time.Since(inicio)
        return result
    }

    hash := sha256.Sum256(body)
    result.Success = true
    result.SizeBytes = len(body)
    result.HashSHA = fmt.Sprintf("%x", hash[:8])
    result.Latency = time.Since(inicio)

    return result
}
```

> Cada descarga es independiente: HTTP GET, leer el body, calcular SHA-256. Los errores se capturan como valores (patrón Go), no como excepciones.

---

```go
// ETAPA 3: Procesador de Resultados (Consumidor)
func procesarResultos(results <-chan DownloadResult) []DownloadResult {
    var todos []DownloadResult

    for result := range results { // Se detiene cuando el channel se cierra
        todos = append(todos, result)

        estado := "✅"
        if !result.Success {
            estado = "❌"
        }
        fmt.Printf("  %s [%d] %s — %v\n", estado, result.TaskID, result.URL, result.Latency)
    }

    return todos
}
```

> El `range` sobre el channel de resultados se ejecuta indefinidamente hasta que `close(out)` se ejecuta en la Etapa 2. Esa es la señal de que **todos los workers terminaron**. No necesitamos WaitGroup aquí — el channel cerrado ES la señal.

---

```go
func main() {
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

    maxWorkers := 3
    inicio := time.Now()

    // Etapa 1 → Etapa 2 → Etapa 3
    tasks := generaURLs(urls)
    results := descargaURLs(tasks, maxWorkers)
    resultados := procesarResultos(results)

    duracion := time.Since(inicio)
    fmt.Printf("\n✅ %d archivos procesados en %v\n", len(resultados), duracion)
}
```

> 📌 **Observa la elegancia**: Tres líneas conectan las tres etapas. Los channels fluyen como tuberías de agua: la Etapa 1 llena la tubería `tasks`, la Etapa 2 la vacía y llena `results`, la Etapa 3 vacía `results`. Sin Mutex, sin shared state, sin locks.

---

## 🧪 Ejecutando el Pipeline

```bash
cd 11-channels
go run main.go
```

**Salida de ejemplo:**
```
═══════════════════════════════════════════════════════════════
   🔄 Pipeline de Descarga y Procesamiento — Lección 11
   📡 Channels como sistema nervioso de Go
═══════════════════════════════════════════════════════════════

📋 Tareas: 10 URLs
⚡ Workers: 3 goroutines concurrentes
───────────────────────────────────────────────────────────────

🔄 Descargando...

  ✅ [1] https://httpbin.org/get — 342ms
  ✅ [2] https://httpbin.org/ip — 281ms
  ✅ [3] https://httpbin.org/user-agent — 295ms
  ...

═══════════════════════════════════════════════════════════════
   📊 REPORTE FINAL
═══════════════════════════════════════════════════════════════
   ✅ Exitosos:  10
   ❌ Fallidos:  0
   📦 Total:     4523 bytes (4.42 KB)
   ⏱️  Duración:  2.34s
═══════════════════════════════════════════════════════════════
```

---

## 📊 Patrón Pipeline: El Diagrama Completo

```
  ┌─────────────────────────────────────────────────────────────────┐
  │                    PIPELINE EN GO                               │
  │                                                                 │
  │   ETAPA 1 (Generador)                                           │
  │   ┌──────────────────┐                                          │
  │   │ for _, url :=    │                                          │
  │   │   range urls {   │──── tasks chan ────▶                      │
  │   │   ch <- task     │   (DownloadTask)                          │
  │   │ }                │                                          │
  │   │ close(ch)        │  ← Señal: no hay más tareas              │
  │   └──────────────────┘                                          │
  │                                                                 │
  │   ETAPA 2 (Workers × N)                                         │
  │   ┌──────────────────┐                                          │
  │   │ for task :=      │◀── tasks chan                             │
  │   │   range tasks {  │                                          │
  │   │   result = proc()│                                          │
  │   │   ch <- result   │──── results chan ────▶                    │
  │   │ }                │   (DownloadResult)                        │
  │   └──────────────────┘                                          │
  │        ↑ wg.Wait() → close(results)  ← Señal: workers listos   │
  │   ETAPA 3 (Consumidor)                                          │
  │   ┌──────────────────┐                                          │
  │   │ for result :=    │◀── results chan                           │
  │   │   range results {│                                          │
  │   │   procesar()     │                                          │
  │   │ }                │  ← Se detiene cuando results se cierra   │
  │   └──────────────────┘                                          │
  └─────────────────────────────────────────────────────────────────┘
```

### ¿Por qué este patrón es tan poderoso?

| Ventaja | Explicación |
|---------|-------------|
| **Sin locks** | Los channels son thread-safe por naturaleza. No necesitas Mutex. |
| **Desacoplamiento** | Cada etapa no sabe nada de las demás. Solo conoce su channel. |
| **Escalable** | ¿Necesitas más descarga? Aumenta `maxWorkers`. ¿Necesitas más procesamiento? Agrega otra etapa. |
| **Composable** | Puedes encadenar etapas infinitas: `etapa1 → etapa2 → etapa3 → etapa4 → ...` |
| **Backpressure** | Si la Etapa 3 es lenta, los channels se llenan y la Etapa 2 automáticamente frena. |

---

## 🧩 Ejemplo Adicional: Channel como Señal de "Done"

A veces no necesitas enviar datos — solo necesitas **señalar** que algo terminó:

```go
func tareaPesada(done chan<- bool) {
    fmt.Println("⏳ Trabajando...")
    time.Sleep(2 * time.Second)
    fmt.Println("✅ Listo!")
    done <- true // Señal: "terminé"
}

func main() {
    done := make(chan bool)
    go tareaPesada(done)

    fmt.Println("main() haciendo otras cosas...")

    <-done // Espera la señal (bloquea hasta recibir)
    fmt.Println("🎉 Tarea completada")
}
```

**Salida:**
```
main() haciendo otras cosas...
⏳ Trabajando...
✅ Listo!
🎉 Tarea completada
```

> Este patrón reemplaza a `sync.WaitGroup` para casos simples. ¿Cuándo usar cada uno? `WaitGroup` cuando esperas múltiples goroutines. Un channel de señal cuando esperas una sola y además quieres pasar algún dato.

---

## 📋 Resumen Visual

```
  ┌─────────────────────────────────────────────────────────────┐
  │                 CHANNELS EN GO                              │
  ├─────────────────────────────────────────────────────────────┤
  │                                                             │
  │  ch := make(chan T)       ← Unbuffered (sincronizado)      │
  │  ch := make(chan T, n)    ← Buffered (hasta n elementos)   │
  │                                                             │
  │  ch <- dato               ← Enviar (bloquea si lleno)      │
  │  dato := <-ch             ← Recibir (bloquea si vacío)     │
  │  close(ch)                ← Señal de fin de datos           │
  │                                                             │
  │  for v := range ch {}     ← Lee hasta que cierren           │
  │                                                             │
  │  chan<- T                  ← Solo escritura                  │
  │  <-chan T                  ← Solo lectura                    │
  │                                                             │
  │  ⚠️  Enviar a nil  → PANIC                                 │
  │  ⚠️  Recibir de nil → PANIC                                 │
  │  ⚠️  Enviar a cerrado → PANIC                               │
  │  ✅  Recibir de cerrado → zero value                        │
  │  ⚠️  Cerrar nil → PANIC                                     │
  │  ⚠️  Cerrar dos veces → PANIC                               │
  │                                                             │
  │  🧠 Filosofía: No compartas memoria, comunica por channels  │
  │                                                             │
  └─────────────────────────────────────────────────────────────┘
```

---

## 🧩 Cómo Ejecutar Este Ejercicio

```bash
cd 11-channels
go run main.go
```

---

## 🧪 Ejercicio Feynman

> **Instrucción**: Explica estos conceptos con tus propias palabras, como si le enseñaras a alguien que nunca ha programado. Usa tus propias analogías. Si no puedes explicarlo de forma simple, es que no lo entendiste lo suficiente.

### ✍️ Ejercicio 1: Explica con una analogía
> "¿Qué es un channel y por qué es mejor que compartir memoria con locks?"

**Tu respuesta** (escribe aquí):
```
Tu analogía aquí...

```

---

### ✍️ Ejercicio 2: ¿Qué imprime este código?
```go
func main() {
    ch := make(chan int, 2)
    ch <- 10
    ch <- 20
    close(ch)

    fmt.Println(<-ch)
    fmt.Println(<-ch)
    fmt.Println(<-ch)
}
```

**Tu respuesta** (escribe aquí):
```
Imprime:
¿El tercer receive funciona? ¿Por qué?

```

<details>
<summary>🔍 Ver solución</summary>

**Imprime:**
```
10
20
0
```

**Explicación:** Los dos primeros `<-ch` reciben los valores enviados (10 y 20). El tercer `<-ch` recibe del channel cerrado: devuelve el **zero value** del tipo `int`, que es `0`. No causa panic porque recibir de un channel cerrado es válido — solo devuelve zero value.

**La trampa:** Si intentas `ch <- 30` después de `close(ch)`, ahí sí hay panic.
</details>

---

### ✍️ Ejercicio 3: Buffered vs Unbuffered — ¿Qué pasa aquí?
```go
func main() {
    ch := make(chan string) // Sin buffer
    ch <- "hola"
    fmt.Println(<-ch)
}
```

**Tu respuesta** (escribe aquí):
```
¿Este programa funciona? ¿Por qué sí o por qué no?
¿Qué pasa exactamente?

```

<details>
<summary>🔍 Ver solución</summary>

**El programa hace DEADLOCK (se cuelga para siempre).**

**¿Por qué?** Un channel unbuffered requiere que el emisor y receptor estén listos al mismo tiempo (handshake). En este programa:
1. `main()` es la única goroutine
2. `ch <- "hola"` intenta enviar, pero **nadie está recibiendo** todavía
3. `main()` se bloquea esperando un receptor que nunca llegará
4. Go detecta que todas las goroutines están dormidas → **fatal error: all goroutines are asleep - deadlock!**

**Soluciones:**
```go
// Opción 1: Usar buffer
ch := make(chan string, 1)  // Buffer de 1: no bloquea

// Opción 2: Lanzar una goroutine para recibir
go func() { fmt.Println(<-ch) }()
ch <- "hola"

// Opción 3: Recibir primero en otra goroutine
go func() { ch <- "hola" }()
fmt.Println(<-ch)
```
</details>

---

### ✍️ Ejercicio 4: Implementa tu propio Pipeline
> Crea un pipeline de 2 etapas: (1) un generador que emita los números del 1 al 20, (2) un procesador que reciba cada número y calcule su cuadrado. Muestra los resultados en consola. Usa channels para conectar las etapas.

**Tu respuesta** (escribe aquí):
```go
// Escribe tu código aquí

```

<details>
<summary>🔍 Ver solución sugerida</summary>

```go
package main

import "fmt"

// Etapa 1: Generador
func generar(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Etapa 2: Procesador (calcula cuadrados)
func cuadrados(in <-chan int) <-chan string {
    out := make(chan string)
    go func() {
        for n := range in {
            out <- fmt.Sprintf("%d² = %d", n, n*n)
        }
        close(out)
    }()
    return out
}

func main() {
    // Pipeline: generar → cuadrados → imprimir
    numeros := generar(1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
                       11, 12, 13, 14, 15, 16, 17, 18, 19, 20)
    resultados := cuadrados(numeros)

    for r := range resultados {
        fmt.Println(r)
    }
}
```

**Salida:**
```
1² = 1
2² = 4
3² = 9
...
20² = 400
```

**Observa:** Dos etapas, un channel entre ellas, cero locks, cero shared state.
</details>

---

### ✍️ Ejercicio 5: Explica el concepto clave
> "¿Por qué Go dice 'No compartas memoria, comunica'? ¿Cuál es la diferencia práctica entre usar un Mutex y usar un channel para coordinar dos goroutines?"

**Tu respuesta** (escribe aquí):
```
¿Por qué los channels son más seguros?
¿Cuándo un Mutex sería mejor?
¿Qué es un deadlock y cómo lo previenes?

```

<details>
<summary>🔍 Ver respuesta sugerida</summary>

**¿Por qué "comunica, no compartas"?**
- Con Mutex, ambas goroutines acceden a la **misma variable**. El Mutex es una "puerta con llave" que serializa el acceso. Pero si olvidas usar el Mutex en algún lugar, hay una race condition silenciosa.
- Con channels, los datos **fluyen de una goroutine a otra**. Solo una goroutine posee el dato a la vez. No hay "olvidar el lock" porque el channel fuerza la transferencia.

**¿Cuándo usar Mutex?**
- Cuando proteges acceso a un **dato compartido simple** (contador, cache, flag)
- Cuando el overhead de crear/copiar structs para enviarlos por channel es alto
- Cuando el patrón es "leer/escribir misma variable" y no "pasar mensajes"

**¿Qué es un deadlock?**
- Dos goroutines esperándose mutuamente y nunca avanzan. Ejemplo: Goroutine A espera lock en Mutex 1 (que tiene Goroutine B), y Goroutine B espera lock en Mutex 2 (que tiene Goroutine A). Ambas esperan eternamente.

**Prevención:** Nunca adquieras dos locks en orden diferente en goroutines diferentes. Con channels, el deadlock ocurre cuando todos los channels están bloqueados y no hay goroutines activas — Go lo detecta automáticamente y paniquea.
</details>

---

## 🚀 Próxima lección

En la **Lección 12** exploraremos **`select`: El Director de Orquesta Concurrente** — la herramienta que permite a una goroutine esperar en múltiples channels simultáneamente. Con `select`, podrás implementar timeouts, cancelación, y multiplexación de eventos. Es el `switch` de los channels, y es lo que transforma tus pipelines de simples a robustos.

---

> *"No comuniques compartiendo memoria; comparte memoria comunicando."*
> — **Proverbio de Go** 🧬