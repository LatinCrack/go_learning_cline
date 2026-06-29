# 🎵 Lección 12 — Select: El Director de Orquesta Concurrente

## 📖 ¿Qué es `select`?

Imagina un **recepcionista de hotel** que atiende 5 teléfonos al mismo tiempo. Cuando suena uno, atiende ese. Si suenan 3 a la vez, elige uno — no necesariamente el que sonó primero. Si no suena ninguno, espera (a menos que tenga un `default`, que sería como "hacer otra cosa mientras tanto").

Eso es `select` en Go: un mecanismo que permite a una goroutine **esperar en múltiples channels simultáneamente**.

> 🎵 **Ese es `select`.** Es como un `switch`, pero para operaciones de channel. Cuando necesitas que una goroutine responda a múltiples fuentes de eventos al mismo tiempo — timeouts, cancelaciones, datos entrantes — `select` es tu herramienta.

```
  ┌─────────────────────────────────────────────────────────┐
  │                    GOROUTINE                             │
  │                                                         │
  │              ┌──────────────┐                           │
  │   ch1 ──────▶│              │                           │
  │              │    SELECT    │──── Responde al que llegue │
  │   ch2 ──────▶│              │     primero               │
  │              │              │                           │
  │   ch3 ──────▶│              │                           │
  │              └──────────────┘                           │
  │                    │                                    │
  │         "¿Quién suena primero? Yo atiendo ese."        │
  └─────────────────────────────────────────────────────────┘
```

---

## 🧠 ¿Por qué Go necesita `select`?

En la lección anterior aprendimos que los channels son la forma de comunicarse entre goroutines. Pero, ¿qué pasa cuando una goroutine necesita escuchar **múltiples channels a la vez**? Sin `select`, estarías atrapado:

```go
// ❌ Sin select: no puedes esperar en múltiples channels
msg := <-ch1  // Se bloquea aquí. ¿Y si ch2 tiene datos primero?
```

`select` resuelve esto exactamente como un switch-case, pero para channels:

| Situación | Sin `select` | Con `select` |
|-----------|-------------|--------------|
| Esperar en un solo channel | `<-ch` ✅ | `select { case v := <-ch: }` |
| Esperar en 2+ channels | ❌ Imposible | `select { case <-ch1: case <-ch2: }` |
| Timeout | ❌ Imposible | `select { case <-ch: case <-time.After(5s): }` |
| No bloqueante | ❌ Imposible | `select { case <-ch: default: }` |

> 🧠 **Analogía Feynman**: `select` es como un operador de emergencias que recibe llamadas de policía, bombero y ambulancia al mismo tiempo. No puede atender los 3 teléfonos a la vez, pero escucha todos. Cuando suena cualquiera, atiende ese. Si suenan 2 al mismo tiempo, elige uno al azar (para ser justo). Si ninguno suena en 30 segundos, hace una ronda de verificación (timeout).

---

## 🔑 Tu Primer `select`

La sintaxis es casi idéntica a `switch`, pero cada `case` es una operación de channel:

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    // Goroutine A: responde rápido
    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "📡 Respuesta del Servidor A"
    }()

    // Goroutine B: responde lento
    go func() {
        time.Sleep(300 * time.Millisecond)
        ch2 <- "📡 Respuesta del Servidor B"
    }()

    // SELECT: ¿quién responde primero?
    select {
    case msg := <-ch1:
        fmt.Println(msg)
    case msg := <-ch2:
        fmt.Println(msg)
    }
}
```

**Salida:**
```
📡 Respuesta del Servidor A
```

### 🔍 ¿Qué acaba de pasar?

```
  Timeline:
  ────────────────────────────────────────────────────────►

  0ms    Goroutine A y B empiezan a "trabajar"
  │
  100ms  A envía a ch1 ──▶ select DESBLOQUEA con ch1
  │                        Imprime "Servidor A"
  │
  300ms  B envía a ch2 ──▶ (nadie escucha, se queda colgado)
         ⚠️ Goroutine B leak! (memory leak)
```

`select` se bloqueó esperando que **cualquiera** de los dos channels tuviera datos. Como `ch1` fue primero, ejecutó ese `case`. La goroutine B quedó enviando al vacío — esto es un **goroutine leak**, y lo resolveremos más adelante con `context`.

---

## 🎲 La Regla de Oro: Selección Pseudoaleatoria

Cuando **múltiples** channels están listos simultáneamente, `select` elige uno **pseudoaleatoriamente**, no el primero:

```go
ch1 := make(chan string, 1)
ch2 := make(chan string, 1)

ch1 <- "A"
ch2 <- "B"  // Ambos tienen datos al mismo tiempo

// ¿Qué imprime? ¡Puede ser cualquiera!
select {
case msg := <-ch1:
    fmt.Println("ch1:", msg)
case msg := <-ch2:
    fmt.Println("ch2:", msg)
}
```

> 🧠 **¿Por qué aleatorio?** Imagina que `select` siempre eligiera el primer `case`. Si `ch1` tiene datos constantemente, `ch2` **nunca** sería atendido — eso se llama **starvation** (inanición). La selección aleatoria garantiza **justicia**: cada channel tiene probabilidad de ser atendido.

```
  ┌───────────────────────────────────────────────────────┐
  │  CON select aleatorio:     SIN select aleatorio:      │
  │                                                       │
  │  ch1: ██████████          ch1: ██████████████████████ │
  │  ch2: ██████████          ch2: ░                       │
  │       (justo)                   (starvation)           │
  └───────────────────────────────────────────────────────┘
```

---

## ⏱️ `select` con Timeout: `time.After`

El patrón más usado en producción. Si el servicio no responde en N segundos, abortar:

```go
servicioLento := make(chan string)

go func() {
    time.Sleep(2 * time.Second)
    servicioLento <- "Datos del servicio"
}()

select {
case resultado := <-servicioLento:
    fmt.Println("✅ Recibido:", resultado)
case <-time.After(1 * time.Second):
    fmt.Println("⏰ Timeout: el servicio tardó más de 1 segundo")
}
```

**Salida:**
```
⏰ Timeout: el servicio tardó más de 1 segundo
```

### 🔍 ¿Cómo funciona `time.After`?

`time.After(d)` devuelve un `<-chan Time` que recibe exactamente un valor después de la duración `d`. Es un channel que "suena" después de N segundos:

```
  time.After(1s) crea un channel:
  
  0s ───────────────────── 1s ──────────────▶
                            │
                     channel recibe time.Now()
                     select se activa con este case
```

> 📌 **Regla profesional**: Nunca hagas una operación de red sin timeout. Un `select` con `time.After` es tu seguro de vida contra servicios que no responden.

---

## 🏃 `select` No-Bloqueante: `default`

El caso `default` hace que `select` **nunca se bloquee**: si ningún channel está listo, ejecuta `default` inmediatamente:

```go
ch := make(chan string, 1)

// Sin datos en el channel
select {
case msg := <-ch:
    fmt.Println("📬 Mensaje:", msg)
default:
    fmt.Println("📭 No hay mensajes — sigo trabajando")
}
```

**Salida:**
```
📭 No hay mensajes — sigo trabajando
```

### ¿Cuándo es útil `default`?

| Escenario | ¿Usar `default`? | Por qué |
|-----------|-------------------|---------|
| Polling de channel | ✅ Sí | "Miro y si no hay, sigo" |
| Busy loop (¡cuidado!) | ⚠️ Con precaución | `for { select { ... default: } }` consume CPU |
| Espera blocking | ❌ No | Para eso es `select` sin `default` |

> 🧠 **Analogía**: `default` es como mirar el buzón de correo: abres la tapa, si hay cartas las recoges, si no hay, te vas a hacer otra cosa. Sin `default`, te quedas parado frente al buzón esperando a que el cartero venga.

⚠️ **Peligro: Busy Loop**

```go
// ❌ MALO: consume 100% de CPU
for {
    select {
    case msg := <-ch:
        fmt.Println(msg)
    default:
        // Gira en vacío quemando CPU — ¡NUNCA hagas esto!
    }
}
```

Si necesitas polling, agrega un pequeño `time.Sleep` o usa `time.Ticker`.

---

## 🛑 `select` con `context`: Cancelación Profesional

En producción, la pregunta no es solo "¿llegó el dato?" sino también **"¿nos cancelaron?"**. `context.Context` es un channel especial que se cierra cuando hay que parar:

```go
ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
defer cancel()

trabajo := make(chan string)

go func() {
    time.Sleep(2 * time.Second) // Tarea que toma mucho
    trabajo <- "Trabajo completado"
}()

select {
case resultado := <-trabajo:
    fmt.Println("✅", resultado)
case <-ctx.Done():
    fmt.Println("🛑 Cancelado:", ctx.Err())
}
```

**Salida:**
```
🛑 Cancelado: context deadline exceeded
```

```
  Timeline:
  ────────────────────────────────────────────────────────►

  0ms     context.WithTimeout(500ms) ← empieza el reloj
  │
  500ms   ctx.Done() se cierra ──▶ select activa el case
  │        "context deadline exceeded"
  │
  2000ms  trabajo envía "Completado" ← pero nadie escucha (goroutine leak)
```

> 📌 **Patrón profesional**: Siempre pasa `ctx context.Context` como primer parámetro de funciones que pueden tardar. Esto permite que el llamador cancele la operación en cualquier momento. Es el estándar de la industria en Go.

---

## 🔄 Los 5 Patrones Fundamentales de `select`

### Patrón 1: Multiplexación (el más básico)

```go
select {
case msg := <-ch1:
    // Manejar mensaje de ch1
case msg := <-ch2:
    // Manejar mensaje de ch2
case msg := <-ch3:
    // Manejar mensaje de ch3
}
```

> Espera en múltiples canales. El que primero tenga datos, gana.

### Patrón 2: Timeout

```go
select {
case data := <-resultado:
    // Procesar datos
case <-time.After(5 * time.Second):
    // Timeout — abortar operación
}
```

> Nunca esperes indefinidamente. Ponle un límite de tiempo a todo.

### Patrón 3: Cancelación con Context

```go
select {
case data := <-resultado:
    // Procesar datos
case <-ctx.Done():
    // Cancelado — limpiar y salir
    return ctx.Err()
}
```

> El estándar de la industria para propagar cancelaciones a través de goroutines.

### Patrón 4: No-bloqueante con Default

```go
select {
case msg := <-ch:
    // Procesar mensaje
default:
    // No hay nada — seguir trabajando
}
```

> "Miro y si no hay, sigo." Útil para polling.

### Patrón 5: Heartbeat / Ticker

```go
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        // Hacer algo cada 5 segundos (heartbeat, métricas, limpieza)
    case data := <-ch:
        // Procesar datos entrantes
    case <-ctx.Done():
        return
    }
}
```

> Combina trabajo periódico con procesamiento de eventos. Es el patrón base de todo servidor Go profesional.

---

## ⚡ `select` en el `for`: El Bucle Infinito de Eventos

La forma más común de usar `select` es dentro de un `for` — esto crea un **bucle de eventos** que responde continuamente:

```go
for {
    select {
    case msg := <-mensajes:
        fmt.Println("📩", msg)
    case err := <-errores:
        fmt.Println("❌", err)
    case <-ticker.C:
        fmt.Println("💓 Heartbeat cada 5s")
    case <-ctx.Done():
        fmt.Println("🛑 Shutting down...")
        return
    }
}
```

```
  ┌─────────────────────────────────────────────────────┐
  │              BUCLE DE EVENTOS INFINITO               │
  │                                                     │
  │    for {                                            │
  │      select {         ┌─────────┐                   │
  │        <-mensajes ───▶│         │                   │
  │        <-errores  ───▶│ ¿Quién  │──▶ Ejecutar case │
  │        <-ticker   ───▶│ llega?  │                   │
  │        <-ctx.Done ───▶│         │──▶ return (salir) │
  │                       └─────────┘                   │
  │    }                                                │
  │              "El hotel nunca cierra,                 │
  │               excepto cuando el gerente              │
  │               dice que sí (ctx.Done)"                │
  └─────────────────────────────────────────────────────┘
```

> 🧠 **Este es el patrón fundamental del servidor Go.** Todo servidor HTTP, todo worker, todo microservicio en Go tiene este bucle en algún lugar. Si lo dominas, dominas la concurrencia en Go.

---

## 🚫 Errores Comunes con `select`

### Error 1: Goroutine Leak (fuga de goroutines)

```go
// ❌ MALO: la goroutine interna nunca termina
func hacerAlgo() <-chan string {
    ch := make(chan string)
    go func() {
        time.Sleep(10 * time.Second) // Trabajo largo
        ch <- "resultado"
    }()
    return ch
}

func main() {
    select {
    case r := <-hacerAlgo():
        fmt.Println(r)
    case <-time.After(1 * time.Second):
        fmt.Println("Timeout") // ← La goroutine de hacerAlgo() nunca se entera
    }
}
```

**Solución**: Usar `context` para cancelar:

```go
// ✅ BUENO: la goroutine escucha ctx.Done()
func hacerAlgo(ctx context.Context) <-chan string {
    ch := make(chan string)
    go func() {
        select {
        case <-time.After(10 * time.Second):
            ch <- "resultado"
        case <-ctx.Done():
            return // Se detiene limpiamente
        }
    }()
    return ch
}
```

### Error 2: Empty select (deadlock)

```go
// ❌ MALO: deadlock inmediato
select {}
// panic: all goroutines are asleep - deadlock!
```

Un `select` sin `cases` bloquea para siempre. Nunca lo hagas.

### Error 3: Send en select (confusión de dirección)

```go
// ⚠️ CUIDADO: enviar en select se bloquea si el buffer está lleno
select {
case ch <- dato:  // Se bloquea si nadie recibe
    fmt.Println("Enviado")
case <-ctx.Done():
    fmt.Println("Cancelado")
}
```

Esto es **correcto**, pero ten cuidado: si `ch` es unbuffered y nadie recibe, este `case` se bloquea igual que un `<-` normal. El `ctx.Done` te salva de quedarte colgado.

---

## 🏋️ Ejercicio Práctico: Monitor de Servicios con Timeout y Circuit Breaker

Vamos a construir algo real y útil: un **monitor de servicios** que vigila múltiples endpoints HTTP simultáneamente, maneja timeouts, y implementa un **Circuit Breaker** que deja de monitorear un servicio después de N fallos consecutivos.

### ¿Por qué un monitor de servicios?

Porque es el escenario perfecto donde `select` demuestra todo su poder:
- Cada verificación puede tardar o fallar (timeout)
- El sistema debe responder a cancelaciones (context)
- Un servicio degradado no debe desperdiciar recursos (circuit breaker)
- Múltiples fuentes de eventos convergen en una goroutine (multiplexación)

```
  ┌─────────────────────────────────────────────────────────────┐
  │              MONITOR DE SERVICIOS                            │
  │                                                             │
  │  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐    │
  │  │  API-Users   │   │  API-Pagos   │   │ API-Notif.   │    │
  │  │  (goroutine) │   │  (goroutine) │   │ (goroutine)  │    │
  │  └──────┬───────┘   └──────┬───────┘   └──────┬───────┘    │
  │         │                  │                   │            │
  │         │    resultados chan (fan-in)           │            │
  │         └──────────────────┼───────────────────┘            │
  │                            ▼                                │
  │                   ┌────────────────┐                        │
  │                   │  Recolector    │                        │
  │                   │  (main gorout) │                        │
  │                   └────────────────┘                        │
  │                                                             │
  │  Cada goroutine usa SELECT para:                            │
  │    1. Escuchar el ticker (¿toca verificar?)                │
  │    2. Escuchar el context (¿nos cancelan?)                 │
  │    3. Manejar timeouts individuales por verificación       │
  └─────────────────────────────────────────────────────────────┘
```

### El código explicado línea por línea

```go
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
```

> Importamos `context` para cancelación, `math/rand` para simulación de latencias, y `sync` para coordinar el cierre del recolector.

---

```go
// Estado del Circuit Breaker
type EstadoCB int

const (
    Cerrado    EstadoCB = iota // Funcionando normalmente (0)
    Abierto                    // Circuito abierto: servicio "muerto" (1)
    MedioAbierto               // Probando si el servicio se recuperó (2)
)
```

> Un `iota` enumera los estados. El Circuit Breaker tiene 3 estados como un semáforo: verde (funciona), rojo (caído), amarillo (probando).

---

```go
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
```

> Implementamos `String()` para que los estados se impriman con emojis legibles. Go llama a `String()` automáticamente cuando usas `fmt.Println`.

---

```go
type ResultadoMonitoreo struct {
    Servicio   string
    Estado     string
    Latencia   time.Duration
    StatusCode int
    Error      string
    Timestamp  time.Time
}

type Servicio struct {
    Nombre string
    URL    string
}
```

> Dos structs: uno para los datos del resultado (viaja por los channels), otro para la configuración del servicio.

---

```go
type CircuitBreaker struct {
    Nombre         string
    Estado         EstadoCB
    FallosConsecut int
    UmbralFallos   int
    TimeoutReset   time.Duration
    UltimoFallo    time.Time
    mu             sync.Mutex
}

func NuevoCircuitBreaker(nombre string, umbral int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        Nombre:       nombre,
        Estado:       Cerrado,
        UmbralFallos: umbral,
        TimeoutReset: timeout,
    }
}
```

> El Circuit Breaker tiene un `sync.Mutex` porque puede ser accedido desde múltiples goroutines. El constructor inicializa el estado en `Cerrado` (normal).

---

```go
func (cb *CircuitBreaker) RegistrarExito() {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    cb.FallosConsecut = 0
    cb.Estado = Cerrado
}

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
```

> Un éxito resetea todo. Un fallo incrementa el contador. Si se supera el umbral (3 fallos), el circuito se **abre** y deja de monitorear el servicio.

---

```go
func (cb *CircuitBreaker) PermitirVerificacion() bool {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    switch cb.Estado {
    case Cerrado:
        return true
    case Abierto:
        if time.Since(cb.UltimoFallo) > cb.TimeoutReset {
            cb.Estado = MedioAbierto
            return true
        }
        return false
    case MedioAbierto:
        return true
    }
    return false
}
```

> Esta es la función de decisión. Si el circuito está cerrado → permite. Si está abierto → solo permite si pasó el timeout (cambia a medio-abierto). Si está medio-abierto → permite un intento de prueba.

```
  Diagrama de estados del Circuit Breaker:

  ┌─────────┐  N fallos   ┌─────────┐  timeout   ┌──────────────┐
  │ CERRADO │────────────▶│ ABIERTO │───────────▶│ MEDIO-ABIERTO│
  │ (normal)│             │ (muerto)│            │  (probando)  │
  └─────────┘             └─────────┘            └──────────────┘
       ▲                                               │
       │          éxito                                │
       └───────────────────────────────────────────────┘
       │          fallo                                │
       │                                               ▼
       └─────────────────────────────────────────── ABIERTO
```

---

```go
func simularVerificacion(ctx context.Context, servicio Servicio) ResultadoMonitoreo {
    resultado := ResultadoMonitoreo{
        Servicio:  servicio.Nombre,
        Timestamp: time.Now(),
    }
    inicio := time.Now()

    latenciaSimulada := time.Duration(100+rand.Intn(2900)) * time.Millisecond

    done := make(chan struct{})
    go func() {
        time.Sleep(latenciaSimulada)
        close(done)
    }()

    select {
    case <-done:
        resultado.Latencia = time.Since(inicio)
        azar := rand.Intn(10)
        switch {
        case azar < 7:
            resultado.Estado = "✅ OK"
            resultado.StatusCode = 200
        case azar < 9:
            resultado.Estado = "⚠️ LENTO"
            resultado.StatusCode = 200
        default:
            resultado.Estado = "❌ ERROR"
            resultado.Error = "connection refused"
            resultado.StatusCode = 503
        }
    case <-ctx.Done():
        resultado.Estado = "⏰ TIMEOUT"
        resultado.Error = ctx.Err().Error()
        resultado.Latencia = time.Since(inicio)
    }
    return resultado
}
```

> 🧠 **Aquí está la magia**: el `select` dentro de `simularVerificacion` compite entre "la simulación terminó" (`done`) y "el timeout se agotó" (`ctx.Done()`). Si la latencia simulada es de 3 segundos pero el timeout es de 2, el `select` elige `ctx.Done()` y reporta TIMEOUT. **Esto es exactamente lo que harías con un HTTP request real.**

---

```go
func monitorearServicioSimulado(
    ctx context.Context,
    servicio Servicio,
    resultados chan<- ResultadoMonitoreo,
    intervalo time.Duration,
) {
    cb := NuevoCircuitBreaker(servicio.Nombre, 3, 10*time.Second)
    ticker := time.NewTicker(intervalo)
    defer ticker.Stop()

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
            return
        }
    }
}
```

> 🧠 **El `select` de alto nivel** escucha 2 señales:
> 1. **`ticker.C`**: "Es hora de verificar" → llama a `simularVerificacion` (que tiene su propio `select` interno para timeout)
> 2. **`ctx.Done()`**: "El sistema dice basta" → `return` detiene la goroutine
>
> Cada goroutine de monitoreo es **independiente**. Un servicio caído no afecta a los demás. Esto es desacoplamiento total gracias a los channels.

---

```go
func recolectarResultados(resultados <-chan ResultadoMonitoreo, wg *sync.WaitGroup) {
    defer wg.Done()
    for resultado := range resultados {
        // Imprime cada resultado en tiempo real
    }
    // Reporte final
}
```

> El recolector consume el channel `resultados` usando `range`. Se detiene automáticamente cuando el channel se cierra. El `WaitGroup` le dice a `main()` que espere hasta que procese todos los resultados pendientes.

---

```go
func main() {
    // Demos de select
    demoBasicoSelect()
    demoTimeoutSelect()
    demoDefaultSelect()
    demoCancelacionSelect()

    // Monitor real
    servicios := []Servicio{
        {Nombre: "API-Usuarios", URL: "https://example.com/users"},
        {Nombre: "API-Pagos", URL: "https://example.com/payments"},
        {Nombre: "API-Notificaciones", URL: "https://example.com/notifications"},
    }

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    resultados := make(chan ResultadoMonitoreo)

    for _, svc := range servicios {
        go monitorearServicioSimulado(ctx, svc, resultados, 2*time.Second)
    }

    var wg sync.WaitGroup
    wg.Add(1)
    go recolectarResultados(resultados, &wg)

    <-ctx.Done()        // Espera 15 segundos
    close(resultados)   // Cierra channel → range se detiene
    wg.Wait()           // Espera al recolector
}
```

> 📌 **La orquestación completa**:
> 1. `context.WithTimeout(15s)` → el reloj global
> 2. Tres goroutines monitorean servicios independientemente
> 3. Todas envían a `resultados` (fan-in por channel)
> 4. El recolector imprime en tiempo real
> 5. `<-ctx.Done()` espera 15 segundos
> 6. `close(resultados)` le dice al recolector "ya no hay más"
> 7. `wg.Wait()` asegura que el recolector terminó

---

## 🧪 Ejecutando el Monitor

```bash
cd 12-select
go run main.go
```

**Salida de ejemplo:**
```
═══════════════════════════════════════════════════════════
  🎵 SELECT: El Director de Orquesta Concurrente
  📚 Lección 12 — Laboratorio de Go
═══════════════════════════════════════════════════════════

═══════════════════════════════════════════════════════════
  🎵 DEMO 1: Select Básico — Múltiples Channels
═══════════════════════════════════════════════════════════

    ch1 dice: 📡 Respuesta del Servidor A
    ch2 dice: 📡 Respuesta del Servidor B

═══════════════════════════════════════════════════════════
  ⏱️  DEMO 2: Select con Timeout
═══════════════════════════════════════════════════════════

    ⏰ Timeout: el servicio tardó más de 1 segundo

═══════════════════════════════════════════════════════════
  🏃 DEMO 3: Select No-Bloqueante (default)
═══════════════════════════════════════════════════════════

    📭 No hay mensajes — continúo haciendo otra cosa
    📬 Mensaje recibido: ¡Hola!

═══════════════════════════════════════════════════════════
  🛑 DEMO 4: Select con Cancelación (context)
═══════════════════════════════════════════════════════════

    🛑 Cancelado: context deadline exceeded

═══════════════════════════════════════════════════════════
  🔍 DEMO 5: Monitor de Servicios + Circuit Breaker
═══════════════════════════════════════════════════════════

    👁️  Monitoreando [API-Usuarios] cada 2s
    👁️  Monitoreando [API-Pagos] cada 2s
    👁️  Monitoreando [API-Notificaciones] cada 2s

    ✅ [API-Pagos] ✅ OK — 245ms
    ✅ [API-Usuarios] ✅ OK — 1.2s
    ❌ [API-Notificaciones] ❌ ERROR — 150ms (connection refused)
    ...

    ⏱️  Tiempo de monitoreo agotado, cerrando...

═══════════════════════════════════════════════════════════
  ✅ Monitor de Servicios finalizado
  📚 Conceptos demostrados:
     • select multiplexando channels
     • time.After para timeouts
     • context.WithTimeout para cancelación
     • Circuit Breaker para resiliencia
     • Fan-in de resultados concurrentes
═══════════════════════════════════════════════════════════
```

---

## 📊 Patrón Select + Context: El Diagrama Completo

```
  ┌─────────────────────────────────────────────────────────────┐
  │           SELECT + CONTEXT + TIMEOUT                        │
  │                                                             │
  │   GOROUTINE MONITOR                                         │
  │   ┌─────────────────────────────────────────┐               │
  │   │ for {                                   │               │
  │   │   select {                              │               │
  │   │     │                                   │               │
  │   │     ├── <-ticker.C  ──▶ Verificar       │               │
  │   │     │                   servicio        │               │
  │   │     │                   │               │               │
  │   │     │              select {             │               │
  │   │     │                <-done: OK ✅      │               │
  │   │     │                <-ctx.Done: TO ⏰  │               │
  │   │     │              }                    │               │
  │   │     │                   │               │               │
  │   │     │              Circuit Breaker:     │               │
  │   │     │              ┌─ Cerrado → OK     │               │
  │   │     │              ├─ Abierto → SKIP   │               │
  │   │     │              └─ MedioAb → test   │               │
  │   │     │                   │               │               │
  │   │     │              resultados <- res    │               │
  │   │     │                                   │               │
  │   │     └── <-ctx.Done  ──▶ return 🛑       │               │
  │   │        (cancel global)                  │               │
  │   │ }                                       │               │
  │   └─────────────────────────────────────────┘               │
  │                                                             │
  │   "El select anidado es el pan de cada día en Go"           │
  └─────────────────────────────────────────────────────────────┘
```

### Ventajas de este patrón

| Ventaja | Explicación |
|---------|-------------|
| **Timeout por operación** | Cada verificación tiene su propio timeout (`context.WithTimeout`) |
| **Cancelación global** | Un solo `ctx.Done()` apaga TODAS las goroutines |
| **Resiliencia** | El Circuit Breaker evita bombardear un servicio caído |
| **Observabilidad** | Cada resultado viaja por el channel con timestamp y latencia |
| **Sin locks** | Solo los Circuit Breakers usan Mutex (porque actualizan estado interno) |

---

## 📋 Resumen Visual

```
  ┌─────────────────────────────────────────────────────────────┐
  │                    SELECT EN GO                             │
  ├─────────────────────────────────────────────────────────────┤
  │                                                             │
  │  select {                                                   │
  │    case msg := <-ch:      ← Espera en múltiples channels   │
  │    case <-time.After(d):  ← Timeout                         │
  │    case <-ctx.Done():     ← Cancelación                     │
  │    default:               ← No-bloqueante                   │
  │  }                                                          │
  │                                                             │
  │  📌 Si múltiples cases listos → elige pseudoaleatorio       │
  │  📌 Sin default → se bloquea hasta que uno se active        │
  │  📌 Con default → no bloquea nunca                          │
  │                                                             │
  │  🎵 Patrones:                                               │
  │     1. Multiplexación  (escuchar N channels)               │
  │     2. Timeout         (time.After)                         │
  │     3. Cancelación     (ctx.Done)                           │
  │     4. No-bloqueante   (default)                            │
  │     5. Heartbeat       (ticker + select)                    │
  │                                                             │
  │  ⚠️  select{} sin cases → deadlock                         │
  │  ⚠️  Goroutine leak → siempre usa context                  │
  │  ⚠️  Busy loop con default → agrega time.Sleep             │
  │                                                             │
  │  🧠 "select es el switch de los channels"                   │
  │                                                             │
  └─────────────────────────────────────────────────────────────┘
```

---

## 🧩 Cómo Ejecutar Este Ejercicio

```bash
cd 12-select
go run main.go
```

---

## 🧪 Ejercicio Feynman

> **Instrucción**: Explica estos conceptos con tus propias palabras, como si le enseñaras a alguien que nunca ha programado. Usa tus propias analogías. Si no puedes explicarlo de forma simple, es que no lo entendiste lo suficiente.

### ✍️ Ejercicio 1: Explica con una analogía
> "¿Qué es `select` y por qué es como un recepcionista de hotel que atiende varios teléfonos?"

**Tu respuesta** (escribe aquí):
```
Tu analogía aquí...

```

---

### ✍️ Ejercicio 2: ¿Qué imprime este código?
```go
func main() {
    ch1 := make(chan string, 1)
    ch2 := make(chan string, 1)

    ch1 <- "A"
    ch2 <- "B"

    select {
    case msg := <-ch1:
        fmt.Println("ch1:", msg)
    case msg := <-ch2:
        fmt.Println("ch2:", msg)
    }
}
```

**Tu respuesta** (escribe aquí):
```
¿Siempre imprime lo mismo? ¿Por qué sí o por qué no?

```

<details>
<summary>🔍 Ver solución</summary>

**No siempre imprime lo mismo.** Puede imprimir `ch1: A` o `ch2: B`.

**¿Por qué?** Ambos channels tienen datos al mismo tiempo (ya se envió antes del `select`). Cuando múltiples `cases` están listos, `select` elige **pseudoaleatoriamente**. Esto es intencional: previene starvation (que un channel siempre gane).

**Ejecuta el programa varias veces y observa diferentes resultados.**
</details>

---

### ✍️ Ejercicio 3: Timeout — ¿Qué pasa aquí?
```go
func main() {
    ch := make(chan string)

    go func() {
        time.Sleep(3 * time.Second)
        ch <- "datos"
    }()

    select {
    case msg := <-ch:
        fmt.Println("✅", msg)
    case <-time.After(1 * time.Second):
        fmt.Println("⏰ Timeout")
    }
    fmt.Println("Fin")
}
```

**Tu respuesta** (escribe aquí):
```
¿Qué imprime? ¿La goroutine interna termina? ¿Hay un goroutine leak?

```

<details>
<summary>🔍 Ver solución</summary>

**Imprime:**
```
⏰ Timeout
Fin
```

**¿La goroutine interna termina?** Sí, después de 3 segundos envía `"datos"` al channel, pero nadie lo recibe. Sin embargo, como el channel es unbuffered, la goroutine queda **bloqueada intentando enviar** para siempre. **Sí hay un goroutine leak.**

**Solución:** Usar `context.WithCancel` para que la goroutine escuche la cancelación:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    select {
    case <-time.After(3 * time.Second):
        ch <- "datos"
    case <-ctx.Done():
        return // Se detiene limpiamente
    }
}()
```
</details>

---

### ✍️ Ejercicio 4: Implementa un Select No-Bloqueante
> Escribe un programa que intente recibir de un channel sin bloquearse. Si hay datos, imprímelos. Si no hay, imprime "trabajando en otra cosa...". Usa `default` en el `select`.

**Tu respuesta** (escribe aquí):
```go
// Escribe tu código aquí

```

<details>
<summary>🔍 Ver solución sugerida</summary>

```go
package main

import "fmt"

func main() {
    ch := make(chan string, 1)

    // Intento 1: channel vacío
    select {
    case msg := <-ch:
        fmt.Println("📬 Mensaje:", msg)
    default:
        fmt.Println("📭 Sin mensajes — trabajando en otra cosa...")
    }

    // Enviamos algo
    ch <- "¡Hola mundo!"

    // Intento 2: channel con datos
    select {
    case msg := <-ch:
        fmt.Println("📬 Mensaje:", msg)
    default:
        fmt.Println("📭 Sin mensajes")
    }
}
```

**Salida:**
```
📭 Sin mensajes — trabajando en otra cosa...
📬 Mensaje: ¡Hola mundo!
```

**Observa:** `default` hace que `select` no espere. Si hay datos → los toma. Si no hay → pasa al `default` inmediatamente.
</details>

---

### ✍️ Ejercicio 5: Explica el concepto clave
> "¿Por qué `select` elige aleatoriamente entre múltiples channels listos, en vez de siempre el primero? ¿Qué problema de starvation se resolvería si eligiera siempre el primero? Luego explica el caso `default` — ¿por qué hace que `select` sea no-bloqueante, y en qué escenario eso es útil vs peligroso (busy loop)?"

**Tu respuesta** (escribe aquí):
```
¿Por qué aleatorio?
¿Qué es starvation?
¿Cuándo default es útil?
¿Cuándo default es peligroso?

```

<details>
<summary>🔍 Ver respuesta sugerida</summary>

**¿Por qué aleatorio?**
Si `select` siempre eligiera el primer `case`, un channel que tiene datos constantemente monopolizaría la atención. Los otros channels nunca serían atendidos. Esto se llama **starvation** (inanición). La selección aleatoria garantiza justicia estadística: a largo plazo, cada channel tiene la misma probabilidad de ser atendido.

**¿Qué es starvation?**
Es cuando un recurso nunca es asignado a un proceso porque otro proceso siempre tiene prioridad. Ejemplo real: un servidor que procesa pedidos urgentes de forma que los pedidos normales nunca se atienden.

**¿Cuándo default es útil?**
- **Polling de channels:** "Miro si hay datos, y si no hay, sigo con mi trabajo"
- **Intentar enviar sin bloquear:** "Si puedo enviar, envío. Si no, hago otra cosa"
- **Implementar timeouts personalizados:** checkear múltiples cosas sin quedarse atascado

**¿Cuándo default es peligroso?**
Un `for { select { ... default: {} } }` sin pausa consume 100% de CPU (busy loop). El `default` hace que el `select` retorne inmediatamente, así que el `for` gira millones de veces por segundo sin hacer nada útil.

**Solución:** Agregar `time.Sleep(100*time.Millisecond)` o usar `time.Ticker` para no quemar CPU.

**Regla:** `default` es para "intentar una vez". Si necesitas intentar muchas veces, usa un `ticker` o un `time.Sleep`.
</details>

---

## 🚀 Próxima lección

En la **Lección 13** exploraremos **Patrones Avanzados de Concurrencia: Fan-Out, Fan-In y Worker Pools** — las recetas probadas que convierten tus goroutines y channels en sistemas de producción escalables. Aprenderás a distribuir trabajo a múltiples goroutines (fan-out), converger resultados en un solo channel (fan-in), y limitar la concurrencia con worker pools. Es el cierre de la fase de concurrencia de Go.

---

> *"No es la herramienta la que hace al maestro, sino saber cuándo usarla."*
> — **Select en Go** 🎵