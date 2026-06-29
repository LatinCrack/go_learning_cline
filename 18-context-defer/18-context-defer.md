# 🛡️ Lección 18: Context, Defer y Panic/Recover

## 🎯 Objetivo de la Lección

Dominar los tres mecanismos de **control de flujo avanzado** que separan al desarrollador junior del senior en Go: **`context`** para propagar cancellation y deadlines a través de goroutines, **`defer`** para garantizar la limpieza de recursos sin importar cómo termine una función, y **`panic/recover`** como sistema de emergencia para atrapar errores catastróficos sin tumbar el programa completo.

---

## 🧠 Analogía Fundamental: El Crucero de Seguridad

Imagina que estás en un **crucero de lujo** con cientos de pasajeros (goroutines). Tres sistemas de seguridad mantienen el barco a flote:

### 🛡️ `defer` — Las Mamparas Estancas

En un barco real, cada compartimento tiene **puertas estancas** que se cierran automáticamente si hay una inundación. En Go, `defer` son esas puertas: se ejecutan **siempre** al final de una función, sin importar si terminó bien, con error o con panic.

```go
func operacionPeligrosa() {
    archivo, _ := os.Open("datos.txt")
    defer archivo.Close()  // 🧹 Esta puerta SIEMPRE se cierra
    
    // Si hay un panic aquí, archivo.Close() igual se ejecuta
    procesar(archivo)
}
```

### 📡 `context` — La Central de Comunicaciones

El capitán del barco (main) necesita comunicarse con todos los departamentos (goroutines) simultáneamente. Si decide **cancelar la misión** o hay una **emergencia con tiempo limitado**, la señal debe llegar a TODOS los departamentos al instante.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()  // Al finalizar, avisar a todos que ya no hace falta

// Todas las goroutines reciben el mismo ctx
go consultarBaseDeDatos(ctx)
go llamarAPIExterna(ctx)
go procesarDatos(ctx)
```

### 🧯 `panic/recover` — Los Extintores de Incendios

Si un pasajero grita "¡FUEGO!" (panic), el protocolo dice: alguien con un extintor (recover) debe apagarlo inmediatamente, registrar qué pasó, y **el barco sigue navegando**. Sin extintores, el barco entero se hunde.

```go
defer func() {
    if err := recover(); err != nil {
        log.Printf("Incendio controlado: %v", err)
        // El barco sigue navegando
    }
}()
// Si aquí hay un "fuego" (panic), el extintor lo apaga
```

### 🔑 Los Tres Juntos

```
┌─────────────────────────────────────────────────────────┐
│                    TU FUNCIÓN (el barco)                 │
│                                                         │
│  defer archivo.Close()  ← Mampara estanca (siempre)     │
│  defer cancel()         ← Central de comms (siempre)    │
│  defer recover()        ← Extintor (siempre)            │
│                                                         │
│  ctx, cancel := context.WithTimeout(...)                │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Goroutine│  │ Goroutine│  │ Goroutine│              │
│  │ (pasajero)│  │ (pasajero)│  │ (pasajero)│             │
│  │ lee ctx  │  │ lee ctx  │  │ lee ctx  │              │
│  └──────────┘  └──────────┘  └──────────┘              │
│       ▲              ▲              ▲                    │
│       └──────────────┼──────────────┘                    │
│                      │                                   │
│              ctx.Done() canal                           │
│         (señal de cancelación/timeout)                  │
└─────────────────────────────────────────────────────────┘
```

---

## 📦 Paquetes que Estudiaremos

| Paquete | ¿Qué hace? | Analogía |
|---------|-------------|----------|
| `context` | Propagación de cancellation, deadlines y valores | La **central de comunicaciones** del barco |
| `net/http` | Servidor y cliente HTTP integrado | Las **ventanas** del barco (interactúan con el exterior) |
| `sync` | WaitGroup, Mutex para concurrencia segura | El **sistema de turnos** para recursos compartidos |
| `time` | Timers, timeouts, duraciones | El **reloj del capitán** |
| `log` | Logging con timestamps | El **diario de bitácora** |

---

## 📡 `context` — El Sistema Nervioso de tu Aplicación

### ¿Qué es un Context?

Un `context.Context` es un **objeto que viaja a través de tu programa** llevando tres tipos de información:

1. **Señal de cancelación** → "Deja de trabajar, ya no hace falta"
2. **Deadline** → "Si no terminas antes de las 3:00 PM, para"
3. **Valores** → "El request ID es req-42, la IP es 192.168.1.1"

Piensa en el context como un **expediente médico** que viaja con el paciente por cada departamento del hospital. Todos los doctores leen el mismo expediente y saben si el paciente fue dado de alta (cancelado).

### Las 4 Funciones Fundamentales

```
context.Background()    → Context raíz, nunca se cancela
                        → Es como el Big Bang: el origen de todo

context.TODO()          → Placeholder semántico
                        → "Sé que necesito un context, pero aún no sé cuál"

context.WithCancel()    → Puedes cancelarlo manualmente
                        → Como un botón de "STOP" que envías a goroutines

context.WithTimeout()   → Se cancela automáticamente después de N tiempo
                        → Como una alarma que suena sola

context.WithDeadline()  → Se cancela en una fecha/hora específica
                        → Como un vuelo que sale a las 3 PM pase lo que pase

context.WithValue()     → Agrega datos al context (no para cancelación)
                        → Como etiquetas en un paquete que viaja
```

### Diagrama de Árbol de Contexts

```
                    context.Background()
                           │
              ┌────────────┼────────────┐
              │            │            │
         WithCancel   WithTimeout  WithValue
              │            │            │
              ▼            ▼            ▼
           ctx1         ctx2         ctx3
              │            │
              │       ┌────┴────┐
              │       │         │
              ▼       ▼         ▼
           ctx1a    ctx2a     ctx2b

  Si se cancela ctx1 → ctx1a TAMBIÉN se cancela
  Si expira ctx2 → ctx2a y ctx2b TAMBIÉN se cancelan
```

**Regla de oro:** La cancellation se propiga **hacia abajo** en el árbol, nunca hacia arriba.

### context.Background() y context.TODO()

```go
ctx := context.Background()  // El Big Bang: raíz de todo

// Propiedades de un Background:
ctx.Done()   → nil  (nunca se cierra, no hay canal)
ctx.Err()    → nil  (nunca se canceló)
ctx.Deadline() → (time.Time{}, false)  (no tiene deadline)

// ¿Cuándo usar cada uno?
// - Background(): en main(), init(), tests, y como raíz del árbol
// - TODO(): cuando aún no sabes qué context pasar (placeholder temporal)
```

### context.WithCancel — El Botón de STOP

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // 🧹 SIEMPRE defer cancel()

// Lanzar 3 workers
for i := 0; i < 3; i++ {
    go func(id int) {
        for {
            select {
            case <-ctx.Done():  // ¿Llegó la señal de STOP?
                fmt.Printf("Worker %d: cancelado (%v)\n", id, ctx.Err())
                return
            default:
                fmt.Printf("Worker %d: trabajando...\n", id)
                time.Sleep(300 * time.Millisecond)
            }
        }
    }(i)
}

time.Sleep(1 * time.Second)
cancel()  // 🛑 ¡Todos los workers paran!
```

### context.WithTimeout — La Alarma Automática

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()  // 🧹 Siempre defer (libera el timer interno)

resultado := make(chan string, 1)

go func() {
    time.Sleep(5 * time.Second)  // Tarda 5 segundos
    resultado <- "datos de la BD"
}()

select {
case res := <-resultado:
    fmt.Println("✅ Éxito:", res)  // ← Esto NO se ejecuta (tardó 5s)
case <-ctx.Done():
    fmt.Println("⏰ Timeout:", ctx.Err())
    // ctx.Err() == context.DeadlineExceeded
}
```

**La analogía del microondas:** Pones comida por 3 minutos (`WithTimeout`). Si la comida está lista antes, la sacas (`case res`). Si suena la alarma (`case ctx.Done()`), la sacas igual porque el tiempo se agotó. El `defer cancel()` es como apagar el microondas al terminar.

### context.WithValue — Las Etiquetas del Paquete

```go
// Definir tipos de key personalizados (evita colisiones)
type contextKey string
const requestIDKey contextKey = "requestID"

// Middleware agrega datos al context
ctx := context.WithValue(context.Background(), requestIDKey, "req-42")

// Función más profunda lee los datos
func procesar(ctx context.Context) {
    reqID := ctx.Value(requestIDKey).(string)  // Type assertion
    fmt.Printf("Procesando request %s\n", reqID)
}
```

| Regla | Detalle |
|:------|:--------|
| Usa tipos personalizados para keys | Evita colisiones entre paquetes |
| `WithValue` es SOLO para datos del request | NO uses para pasar parámetros "normales" |
| Los valores NO se pueden borrar | Crea un nuevo context, no intentes modificar |
| Si no existe la key, devuelve `nil` | Siempre verifica el valor retornado |

### ctx.Done() — El Canal de Señales

```go
// ctx.Done() devuelve un canal (<-chan struct{})
// Se CIERRA cuando el context se cancela o expira

select {
case <-ctx.Done():
    // El context fue cancelado o expiró
    fmt.Println(ctx.Err())  // context.Canceled o context.DeadlineExceeded
case resultado := <-otroCanal:
    // Recibimos un resultado normal
    fmt.Println(resultado)
}
```

La belleza de `ctx.Done()` es que funciona con `select`, lo que permite integrar la cancelación con cualquier operación concurrente.

---

## 🧹 `defer` — La Garantía de Limpieza

### ¿Qué es `defer`?

`defer` le dice a Go: "Ejecuta esta función **al final**, sin importar cómo termine la función actual". Es como poner un recordatorio adhesivo en tu pantalla: "Cuando te vayas, apaga la luz".

```go
func ejemplo() {
    fmt.Println("Inicio")
    defer fmt.Println("FIN (ejecutado al final)")
    fmt.Println("Cuerpo")
}
// Salida: Inicio → Cuerpo → FIN
```

### Las 5 Reglas de `defer`

#### Regla 1: Orden LIFO (Last In, First Out)

Los deferred functions se ejecutan en **orden inverso** al que se declararon. Como una pila de platos: el último que pusiste es el primero en salir.

```go
func ejemplo() {
    defer fmt.Println("A")  // Último en apilarse
    defer fmt.Println("B")  // Penúltimo
    fmt.Println("C")        // Cuerpo
}
// Salida: C → B → A
```

**¿Por qué LIFO?** Porque el recurso que abriste **al final** es el que debe cerrarse **primero**. Si abriste archivo, luego conexión, luego transacción, la transacción se cierra primero, luego la conexión, luego el archivo.

```
Orden de apertura:    archivo → conexión → transacción
Orden de cierre (LIFO): transacción → conexión → archivo ✅
```

#### Regla 2: Argumentos se evalúan INMEDIATAMENTE

Los argumentos del deferred function se calculan **cuando se declara el defer**, no cuando se ejecuta.

```go
func trampa() {
    x := 10
    defer fmt.Printf("x = %d\n", x)  // x se evalúa AHORA (10)
    x = 20
    fmt.Printf("x = %d\n", x)
}
// Salida: x = 20 → x = 10
// El defer capturó x = 10, no el valor modificado
```

**La trampa:** Si necesitas el valor **al momento de ejecutarse**, usa un closure:

```go
func solucion() {
    x := 10
    defer func() {
        fmt.Printf("x = %d\n", x)  // Lee x al ejecutarse (20)
    }()
    x = 20
    fmt.Printf("x = %d\n", x)
}
// Salida: x = 20 → x = 20
```

#### Regla 3: defer SIEMPRE se ejecuta

Incluso si la función termina con un **panic** o un **return**, los deferred functions se ejecutan.

```go
func siempre() (resultado int) {
    defer fmt.Println("Me ejecuto aunque haya panic")
    defer fmt.Println("Me ejecuto aunque haya return")
    
    panic("¡auxilio!")
    // Los dos defers se ejecutan ANTES de que el panic se propague
}
```

#### Regla 4: defer puede modificar valores de retorno nombrados

```go
func dividir(a, b int) (resultado int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)  // Modifica el error de retorno
        }
    }()
    return a / b, nil  // Si b es 0 → panic → recover modifica err
}
```

#### Regla 5: SIEMPRE defer `Close()` y `cancel()`

```go
// Archivo
f, err := os.Open("datos.txt")
if err != nil { return err }
defer f.Close()  // 🧹 Garantiza cierre

// Context
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()  // 🧹 Garantiza limpieza del timer

// HTTP Response
resp, err := http.Get(url)
if err != nil { return err }
defer resp.Body.Close()  // 🧹 Garantiza cierre del body
```

### Casos de Uso de `defer`

| Caso | Ejemplo | ¿Por qué? |
|:-----|:--------|:-----------|
| Cerrar archivos | `defer f.Close()` | Evita resource leaks |
| Cerrar conexiones DB | `defer conn.Close()` | Libera el pool de conexiones |
| Liberar mutex | `defer mu.Unlock()` | Evita deadlocks |
| Cerrar body HTTP | `defer resp.Body.Close()` | Evita connection leaks |
| Cancelar context | `defer cancel()` | Libera timer interno |
| Medir tiempo | `defer func() { time.Since(inicio) }()` | Siempre preciso |
| Rollback transacción | `defer tx.Rollback()` | Si `Commit()` no se llama, se revierte |
| Recover de panics | `defer func() { recover() }()` | Atrapa panics |

---

## 🧯 `panic` y `recover` — El Extintor de Incendios

### `panic` — ¡FUEGO!

`panic` detiene la ejecución normal de la función actual y comienza a **desenrollar el stack** ejecutando todos los deferred functions. Si llega al `main()` sin ser atrapado, el programa termina.

```go
func peligroso() {
    defer fmt.Println("🧹 Limpieza antes de morir")
    panic("¡división por cero!")
    fmt.Println("(esto nunca se ejecuta)")
}
// Salida: 🧹 Limpieza antes de morir → panic: ¡división por cero!
```

**¿Cuándo ocurre un panic?**
- Cuando tú llamas `panic("razón")`
- Acceso a índice fuera de rango: `slice[99]` en slice de 3 elementos
- Enviar a un canal cerrado
- División por cero de enteros
- Type assertion inválida: `valor.(int)` cuando no es int

**Regla general:** Un panic es para **programación incorrecta** (bugs), no para errores esperados (archivos que no existen, red caída).

### `recover` — El Extintor

`recover()` solo funciona dentro de un `defer`. Atrapa el panic, devuelve su valor, y permite que el programa continúe.

```go
func seguro() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("🧯 Incendio controlado: %v\n", r)
        }
    }()
    
    panic("¡fuego!")
    fmt.Println("(nunca se ejecuta)")
}
// Salida: 🧯 Incendio controlado: ¡fuego!
```

### La Regla de Oro de `recover`

`recover()` **solo funciona dentro de un `defer`**. Si lo llamas en cualquier otro lugar, siempre devuelve `nil`:

```go
// ❌ MALO: recover fuera de defer
r := recover()  // r = nil (no atrapa nada)

// ✅ BIEN: recover dentro de defer
defer func() {
    r := recover()  // r = "el panic" (lo atrapa)
}()
```

### Propagación del Panic

El panic **sube por el stack** hasta encontrar un `recover()`. En el camino, ejecuta todos los defers:

```go
func a() {
    defer fmt.Println("defer en a")     // 3️⃣ Se ejecuta
    b()
    fmt.Println("fin de a")             // ❌ Nunca
}

func b() {
    defer fmt.Println("defer en b")     // 2️⃣ Se ejecuta
    c()
    fmt.Println("fin de b")             // ❌ Nunca
}

func c() {
    defer func() {
        recover()                        // 1️⃣ Atrapa el panic
    }()
    panic("¡fuego!")
    defer fmt.Println("defer en c")     // ❌ Nunca (ya panicó)
}
```

```
Stack de ejecución:
  main() → a() → b() → c() → panic("¡fuego!")
                                    │
  recover() ← defer en c ← defer en b ← defer en a
     ▲
     └── Atrapa el panic aquí, todo lo demás sigue normal
```

### Patrón: Convertir Panic en Error

En Go, es común usar `defer + recover` como **red de seguridad** que convierte panics en errores normales:

```go
func OperacionQuePodriaPanic() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic recuperado: %v", r)
        }
    }()
    
    // Código que podría hacer panic...
    resultado := datos[999]  // Si datos tiene menos de 999 elementos → panic
    
    return nil
}
```

**¿Por qué es útil?** Porque convierte un crash del programa en un error manejable. El llamador puede decidir qué hacer: reintentar, loggear, devolver un error al usuario.

---

## 🌐 `net/http` con Context — Servidores Robustos

### Middlewares y Context

En un servidor HTTP, el **context del request** es el vehículo perfecto para transportar datos y controlar timeouts:

```go
func miMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Agregar datos al context
        ctx := context.WithValue(r.Context(), requestIDKey, "req-123")
        
        // Pasar el context modificado al siguiente handler
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Timeout por Request

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Timeout de 3 segundos para esta operación
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    
    select {
    case resultado := <-consultarServicio(ctx):
        fmt.Fprintf(w, "Resultado: %s", resultado)
    case <-ctx.Done():
        http.Error(w, "Timeout", http.StatusGatewayTimeout)
    }
}
```

### Recovery Middleware

El middleware de recovery atrapa panics en **cualquier handler** sin que el servidor se caiga:

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("PANIC: %v", err)
                http.Error(w, "Error interno", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

---

## 📝 Ejercicio Práctico: Servidor HTTP con Context, Defer y Recover

### ¿Qué construimos?

Un **servidor HTTP completo** que demuestra los tres conceptos en acción real:

1. **Context** → Propaga request ID, maneja timeouts, cancela operaciones largas
2. **Defer** → Limpia mutexes, mide tiempo, garantiza cierre de recursos
3. **Panic/Recover** → Middleware que atrapa panics sin tumbar el servidor

### Arquitectura del Código

```
📁 18-context-defer/
├── go.mod
└── main.go          ← Todo el código en un solo archivo
```

### Componentes Principales

#### 1. RateLimiter con defer

```go
func (rl *RateLimiter) Permitir(ip string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()  // 🧹 Garantiza que el mutex se libere
    
    // Si hay un panic aquí, el mutex igual se libera
    // Sin defer, un panic dejaría el mutex bloqueado para siempre (deadlock)
    ...
}
```

**¿Por qué `defer` aquí?** Porque sin él, si ocurre un panic entre `Lock()` y `Unlock()`, el mutex queda bloqueado **para siempre**. Todas las demás goroutines que intenten acceder se quedarán colgadas. `defer` garantiza que el `Unlock()` se ejecute sin importar qué.

#### 2. Middlewares encadenados

```go
var handler http.Handler = mux
handler = loggingMiddleware(handler)     // Mide tiempo con defer
handler = recoveryMiddleware(handler)    // Atrapa panics con defer+recover
handler = requestIDMiddleware(handler)   // Agrega datos al context
```

El orden importa: el middleware más externo se ejecuta primero. Como capas de una cebolla:

```
Request → [RateLimit] → [RequestID] → [Recovery] → [Logging] → Handler
                                    ← Response ←
```

#### 3. Handler con Timeout (`/lento`)

```go
func handleLento(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    
    duracion := time.Duration(1+rand.Intn(6)) * time.Second
    
    select {
    case <-time.After(duracion):
        fmt.Fprintf(w, "✅ Operación completada en %v", duracion)
    case <-ctx.Done():
        fmt.Fprintf(w, "⏰ Timeout: %v", ctx.Err())
    }
}
```

La operación tarda entre 1 y 6 segundos, pero el timeout es de 3. Si tarda más de 3, el context se cancela automáticamente y respondemos con error 504.

#### 4. Handler con Cancelación (`/cancel`)

```go
func handleCancel(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()  // Se cancela si el cliente cierra la conexión
    
    for i := 1; i <= 10; i++ {
        select {
        case <-ctx.Done():
            fmt.Fprintf(w, "🛑 Cancelado en paso %d", i)
            return
        case <-time.After(500 * time.Millisecond):
            fmt.Fprintf(w, "  ⚙️  Paso %d/10\n", i)
        }
    }
}
```

**Prueba esto:** Abre `/cancel` en el navegador y presiona ESC o cierra la pestaña. Verás en los logs del servidor que el trabajo se cancela inmediatamente.

#### 5. Goroutines con Context (`/concurrent`)

```go
func handleConcurrent(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
    defer cancel()
    
    for _, svc := range servicios {
        go func(nombre string, dur time.Duration) {
            select {
            case <-time.After(dur):
                resultados <- fmt.Sprintf("✅ %s completado", nombre)
            case <-ctx.Done():
                resultados <- fmt.Sprintf("⏰ %s cancelado", nombre)
            }
        }(svc.nombre, svc.duracion)
    }
}
```

Cada servicio (DB, API, Cache, Email, ML) tiene un tiempo diferente. El timeout global de 2 segundos corta los que tardan demasiado. Los que terminan a tiempo reportan éxito.

### Ejecución

```bash
# Modo demostración (explica cada concepto)
go run main.go -demo

# Modo servidor (levantar el servidor HTTP)
go run main.go -server

# Modo ejemplos (snippets de código)
go run main.go -ejemplos

# Servidor en puerto custom
go run main.go -server -puerto :9090
```

### Endpoints del Servidor

| Endpoint | ¿Qué demuestra? | Concepto clave |
|:---------|:-----------------|:---------------|
| `GET /` | Página con request ID y tu IP | `context.WithValue` |
| `GET /lento` | Operación con timeout de 3s | `context.WithTimeout` + `defer cancel()` |
| `GET /panic` | Handler que hace panic | `recover()` en middleware |
| `GET /cancel` | Trabajo paso a paso cancelable | `ctx.Done()` + `http.Flusher` |
| `GET /concurrent` | 5 servicios con timeout global | Goroutines + context tree |

### Prueba en tu navegador

1. Abre `http://localhost:8080/` → Ve tu request ID y IP
2. Abre `http://localhost:8080/lento` → Algunas veces verás "completado", otras "timeout"
3. Abre `http://localhost:8080/panic` → Verás "Error interno" pero el servidor sigue vivo
4. Abre `http://localhost:8080/concurrent` → Ve qué servicios completan y cuáles se cancelan

---

## 🔍 Conceptos Clave Explicados

### ¿Por qué `defer cancel()` es obligatorio?

```go
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
// Si no llamas cancel(), el timer interno sigue vivo en memoria
// hasta que el timeout se cumpla o el parent se cancele

defer cancel()  // Libera los recursos INMEDIATAMENTE cuando la función termina
```

Es como dejar la alarma del despertador puesta cuando ya te levantaste. No pasa nada grave, pero desperdicia batería. `defer cancel()` apaga la alarma.

### ¿Por qué `recover()` solo funciona en `defer`?

```go
// ❌ Esto NO funciona:
func malo() {
    r := recover()  // r = nil, no atrapa nada
    panic("fuego")
}

// ✅ Esto SÍ funciona:
func bueno() {
    defer func() {
        r := recover()  // r = "fuego", atrapa el panic
    }()
    panic("fuego")
}
```

La razón es mecánica: cuando ocurre un `panic`, Go ejecuta los deferred functions del stack. `recover()` interrumpe el desenrollado del stack y retoma el control. Si no estás en un `defer`, no hay nada que interrumpir.

### ¿Por qué `defer` es LIFO y no FIFO?

```go
func ejemploLIFO() {
    archivo, _ := os.Open("datos.txt")
    defer archivo.Close()       // 1er defer
    
    conexion, _ := sql.Open(...)
    defer conexion.Close()      // 2do defer
    
    tx, _ := conexion.Begin()
    defer tx.Rollback()         // 3er defer
    
    tx.Commit()  // Si esto funciona, Rollback no hace nada
}
```

LIFO garantiza que los recursos se cierren en **orden inverso** al que se abrieron. Esto es importante porque:
- La transacción depende de la conexión
- La conexión depende del driver
- Cerrar la conexión antes de la transacción causaría un error

### Diferencia entre `context.Canceled` y `context.DeadlineExceeded`

```go
ctx1, cancel := context.WithCancel(parent)
cancel()
ctx1.Err()  // → context.Canceled (cancelación manual)

ctx2, _ := context.WithTimeout(parent, 1*time.Second)
time.Sleep(2 * time.Second)
ctx2.Err()  // → context.DeadlineExceeded (se agotó el tiempo)
```

| Error | Significado | Analogía |
|:------|:------------|:---------|
| `context.Canceled` | Alguien llamó `cancel()` | "Para, ya no necesito el resultado" |
| `context.DeadlineExceeded` | El timeout/deadline se agotó | "Se acabó el tiempo, para" |

### ¿Por qué los types personalizados para keys de context?

```go
// ❌ MALO: dos paquetes podrían usar la misma string
ctx = context.WithValue(ctx, "requestID", "123")

// ✅ BIEN: tipos distintos = keys distintas
type contextKey string
const requestIDKey contextKey = "requestID"
ctx = context.WithValue(ctx, requestIDKey, "123")
```

Si dos paquetes diferentes usan `string` como tipo de key, podrían colisionar accidentalmente. Con tipos personalizados (`contextKey` vs `otroContextKey`), Go los trata como completamente diferentes.

---

## 🏋️ Ejercicio Feynman

### Instrucciones
Usando la **Técnica Feynman**, explica estos conceptos **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado. Usa analogías de la vida cotidiana.

---

### Ejercicio 1: context como Expediente Médico
> Imagina que el `context` es un **expediente médico** que viaja con un paciente por todos los departamentos del hospital. Explica qué significa cada parte del expediente: el nombre del paciente (`WithValue`), la hora límite de la cirugía (`WithTimeout`), y la orden de "detener todo" (`WithCancel`). ¿Qué pasa si un doctor ignora el expediente?

---

### Ejercicio 2: defer como Lista de Pendientes
> Explica `defer` usando la analogía de una **lista de tareas al salir de casa**: "Apagar la luz", "Cerrar con llave", "Apagar el gas". ¿Por qué la lista se ejecuta en orden inverso? ¿Qué pasa si te vas de la casa corriendo por una emergencia (panic)? ¿Se ejecuta la lista igual?

---

### Ejercicio 3: panic/recover como Protocolo de Emergencia
> Imagina que `panic` es alguien gritando "¡FUEGO!" en un edificio, y `recover` es el guardia de seguridad que dice "Tranquilo, ya lo apagué". Explica por qué el guardia debe estar en su puesto (defer) antes de que alguien grite. ¿Qué pasa si no hay guardia? ¿Qué pasa si hay guardia en cada piso (goroutine)?

---

### Ejercicio 4: context.WithTimeout y el Microondas
> Explica `context.WithTimeout` con la analogía de un **microondas**. Pones comida por 3 minutos. ¿Cuáles son los dos escenarios posibles en el `select`? ¿Qué es `defer cancel()` en esta analogía? ¿Qué pasa si nunca apagas el microondas (no llamas cancel)?

---

### Ejercicio 5: defer y el Orden LIFO
> Imagina que estás **desempacando cajas** de una mudanza. La última caja que metiste en el camión fue la de "cosas frágiles". ¿Por qué tiene sentido sacarla primero? Relaciona esto con: abrir archivo → abrir conexión → iniciar transacción. ¿Por qué se cierran en orden inverso?

---

### Ejercicio 6: El Árbol de Contexts
> Si el `context` es un **árbol genealógico**, explica qué pasa cuando el abuelo (`Background`) crea un hijo (`WithCancel`), y ese hijo crea nietos (`WithTimeout`). ¿Qué pasa si el padre muere (se cancela)? ¿Los nietos sobreviven? ¿Por qué esta propagación es útil en un servidor web con 1000 requests simultáneos?

---

## 📋 Resumen de Funciones Clave

| Función | Paquete | ¿Qué hace? |
|---------|---------|-------------|
| `context.Background()` | `context` | Crea el context raíz (nunca se cancela) |
| `context.TODO()` | `context` | Placeholder para context pendiente |
| `context.WithCancel(ctx)` | `context` | Crea un context cancelable manualmente |
| `context.WithTimeout(ctx, dur)` | `context` | Se cancela automáticamente después de dur |
| `context.WithDeadline(ctx, t)` | `context` | Se cancela en el instante t |
| `context.WithValue(ctx, key, val)` | `context` | Agrega datos al context |
| `ctx.Done()` | `context` | Canal que se cierra al cancelar |
| `ctx.Err()` | `context` | Devuelve el error de cancelación |
| `ctx.Value(key)` | `context` | Recupera un valor del context |
| `defer fn()` | builtin | Ejecuta fn al final de la función (LIFO) |
| `panic(msg)` | builtin` | Detiene ejecución y desenrolla el stack |
| `recover()` | builtin | Atrapa un panic (solo en defer) |
| `mu.Lock()` / `mu.Unlock()` | `sync` | Bloquea/desbloquea un mutex |
| `wg.Add(n)` | `sync` | Registra n goroutines en el WaitGroup |
| `wg.Done()` | `sync` | Marca una goroutine como completada |
| `wg.Wait()` | `sync` | Espera a que todas las goroutines terminen |
| `http.ListenAndServe(addr, h)` | `net/http` | Inicia el servidor HTTP |
| `r.Context()` | `net/http` | Obtiene el context del request |
| `w.(http.Flusher)` | `net/http` | Verifica si el writer soporta flush |

---

## 🔗 ¿Qué sigue?

En la siguiente lección exploraremos **interfaces y polimorfismo** — el mecanismo que hace que Go sea tan elegante sin necesidad de clases ni herencia. Aprenderás a definir comportamientos compartidos, crear mocks para tests, y entender el patrón de composición que hace que el código Go sea tan mantenible.

---

> 💡 *"El código que maneja errores es aburrido. El código que ignora errores es peligroso. El código que maneja panics es profesional."*