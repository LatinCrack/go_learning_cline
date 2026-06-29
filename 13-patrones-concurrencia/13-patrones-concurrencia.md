# 🕷️ Lección 13 — Patrones Avanzados de Concurrencia

> **Método Feynman:** Si no puedes explicarle algo a un niño de 6 años, no lo entiendes lo suficiente.
> **Objetivo:** Dominar los 5 patrones de concurrencia más usados en la industria y combinarlos en un crawler web real.

---

## 🧩 ¿Qué vamos a aprender?

| Concepto | Analogía Feynman |
|----------|------------------|
| **Fan-Out** | 1 gerente → 10 empleados haciendo tareas en paralelo |
| **Fan-In** | 10 reportes → 1 sola bandeja de entrada |
| **Worker Pool** | Call center con 5 operadores fijos |
| **Semaphore** | 10 casilleros: si están llenos, esperas |
| **Rate Limiter** | Máquina expendedora de tickets con intervalo |

**Prerrequisito:** Lecciones 10 (Goroutines), 11 (Channels), 12 (Select)

---

## 📚 Marco Teórico: ¿Por qué necesitamos patrones?

### El Problema de la Concurrencia Descontrolada

En las lecciones 10-12 aprendimos a lanzar goroutines y comunicarlas con channels. Pero en la vida real, lanzar goroutines sin control es como abrir una tienda sin planificación:

```
❌ Sin patrones:
- 10,000 goroutines descargando páginas → tu PC se congela
- 1,000 conexiones simultáneas a una API → te bloquean (HTTP 429)
- Miles de goroutines escribiendo al mismo tiempo → condiciones de carrera
- Sin saber cuándo terminaron → memory leaks
```

```
✅ Con patrones:
- 3 workers descargando 10,000 páginas → controlado y eficiente
- Rate limiter: 1 petición cada 200ms → la API nunca te bloquea
- Semaphore: máximo 2 conexiones a la DB → sin sobrecargar el recurso
- WaitGroup + canal cerrado → todo se limpia correctamente
```

**Los patrones de concurrencia son recetas probadas que resuelven problemas comunes de forma elegante y segura.**

---

## 🔀 Patrón 1: Fan-Out — Distribuir el Trabajo

### Concepto

Fan-Out es cuando **un solo canal de trabajo alimenta múltiples goroutines**. Es como un gerente que pone 100 tareas en una mesa compartida y 10 empleados toman la siguiente tarea disponible.

```
                    ┌─── Worker 1 ───→ Resultado 1
                    │
Canal de Trabajo ───┼─── Worker 2 ───→ Resultado 2
                    │
                    ├─── Worker 3 ───→ Resultado 3
                    │
                    └─── Worker 4 ───→ Resultado 4
```

### ¿Cómo funciona en Go?

```go
// Canal de trabajos (compartido entre todos los workers)
trabajos := make(chan URLTrabajo, 100)

// Fan-Out: 4 workers leyendo del MISMO canal
for i := 0; i < 4; i++ {
    go func(workerID int) {
        for trabajo := range trabajos {
            // Cada worker "toma" un trabajo del canal
            // Go garantiza que NUNCA dos workers procesen
            // el mismo trabajo simultáneamente
            procesar(trabajo)
        }
    }(i)
}
```

### ¿Por qué funciona sin conflictos?

Los **canales en Go son thread-safe por naturaleza**. Cuando 4 goroutines leen del mismo canal:

1. El runtime de Go hace **round-robin** internamente
2. Solo UNA goroutine recibe cada mensaje
3. No hay duplicados, no hay condiciones de carrera
4. Es como una fila en el banco: cada cliente es atendido por un solo cajero

### Implementación Completa

```go
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
            defer close(resultadoCh) // ¡Importante! Cerramos al terminar

            for trabajo := range trabajos {
                pagina := descargarPagina(trabajo.URL)
                resultadoCh <- pagina
            }
        }()
    }

    return resultados // Devolvemos N canales de salida
}
```

**🔑 Puntos clave:**
- `defer close(resultadoCh)` — El worker cierra su canal cuando termina
- Cada worker tiene su **propio canal de resultados**
- Los workers leen de un **canal compartido de trabajo**

---

## 🔀 Patrón 2: Fan-In — Converger los Resultados

### Concepto

Fan-In es lo **opuesto** de Fan-Out: toma N canales de entrada y los converge en **un solo canal de salida**. Como un embudo que junta los reportes de 10 empleados en un solo escritorio.

```
Resultado 1 ───┐
                │
Resultado 2 ───┼──→ Canal Unificado → Consumidor
                │
Resultado 3 ───┤
                │
Resultado 4 ───┘
```

### Implementación

```go
func fanIn(canalesEntrada []chan PaginaWeb) <-chan PaginaWeb {
    canalUnificado := make(chan PaginaWeb)
    var wg sync.WaitGroup

    // Por cada canal de entrada, una goroutine que copia mensajes
    for _, canal := range canalesEntrada {
        wg.Add(1)
        go func(c <-chan PaginaWeb) {
            defer wg.Done()
            for pagina := range c {
                canalUnificado <- pagina  // Copia al canal unificado
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
```

### Fan-Out + Fan-In: El Combo Completo

```go
// 1. Fan-Out: distribuir 8 URLs entre 4 workers
trabajos := make(chan URLTrabajo, 8)
resultadosWorkers := fanOut(trabajos, 4)

// 2. Enviar los trabajos
for _, url := range urls {
    trabajos <- URLTrabajo{URL: url}
}
close(trabajos)

// 3. Fan-In: converger 4 canales en 1
canalUnificado := fanIn(resultadosWorkers)

// 4. Consumir resultados (¡fácil como leer de 1 canal!)
for pagina := range canalUnificado {
    fmt.Println(pagina.URL, pagina.StatusCode)
}
```

---

## 🏭 Patrón 3: Worker Pool — Límite de Concurrencia

### Concepto

Un Worker Pool es un **conjunto fijo de goroutines** que procesan trabajos de un canal compartido. A diferencia del Fan-Out básico, aquí **controlamos exactamente cuántas goroutines están activas**.

```
Canal de Trabajo (20 URLs)
┌─────────────────────────────────────────┐
│ URL1 │ URL2 │ URL3 │ ... │ URL20       │
└──────┴──────┴──────┴─────┴──────────────┘
         │
    ┌────┴────┐
    ▼         ▼         ▼
Worker 1  Worker 2  Worker 3  ← Solo 3 activos
  │         │         │
  ▼         ▼         ▼
Resultado Resultado Resultado → Canal de salida
```

### ¿Por qué es diferente del Fan-Out?

| Característica | Fan-Out básico | Worker Pool |
|---------------|---------------|-------------|
| Número de workers | Variable | **Fijo y controlado** |
| Canal de entrada | Se cierra rápido | **Puede ser infinito** |
| Estadísticas | Difícil de rastrear | **Fácil con closures** |
| Uso típico | Procesar lote conocido | **Cola de trabajo continua** |

### Implementación

```go
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
                // Procesar el trabajo
                pagina := descargarPagina(trabajo.URL)

                // Actualizar estadísticas (thread-safe con atomic)
                atomic.AddInt64(&estadisticas.PaginasDescargadas, 1)

                resultados <- ResultadoCrawl{Pagina: pagina}
            }
        }()
    }

    // Cerrar canal de resultados cuando todos los workers terminen
    go func() {
        wg.Wait()
        close(resultados)
    }()

    return resultados
}
```

### Uso en la práctica

```go
// Configuración
numWorkers := 3      // Solo 3 workers simultáneos
urls := 20           // 20 URLs para procesar

// Crear canal de trabajos
trabajos := make(chan URLTrabajo, len(urls))

// Lanzar el pool
resultados := workerPool(numWorkers, trabajos, estadisticas)

// Enviar trabajos (el canal buffered almacena los que no caben)
for _, url := range urls {
    trabajos <- URLTrabajo{URL: url}
}
close(trabajos)

// Consumir resultados
for resultado := range resultados {
    fmt.Println(resultado.Pagina.URL)
}
```

**🔑 Insight:** Los 3 workers procesan las 20 URLs. Mientras un worker descarga una página, las otras 17 esperan en el canal. Cuando un worker termina, inmediatamente toma la siguiente URL de la cola.

---

## 🔒 Patrón 4: Semaphore — Controlar Acceso a Recursos

### Concepto

Un Semaphore limita **cuántas goroutines acceden a un recurso simultáneamente**. Se implementa con un **canal buffered**:

```go
// Un semáforo con 2 "permisos"
sem := make(chan struct{}, 2)

// Adquirir permiso (bloquea si no hay disponibles)
sem <- struct{}{}

// ... usar el recurso limitado ...

// Liberar permiso
<-sem
```

### Analogía: Los Casilleros del Gimnasio

```
Gimnasio con 3 casilleros:

Cliente A → [Toma casillero 1] → Entrena
Cliente B → [Toma casillero 2] → Entrena
Cliente C → [Toma casillero 3] → Entrena
Cliente D → [Espera...]         → (casilleros llenos)
Cliente E → [Espera...]         → (casilleros llenos)

Cliente A → [Devuelve casillero 1]
Cliente D → [Toma casillero 1] → Entrena
```

### Implementación como Struct

```go
type Semaphore struct {
    tokens chan struct{}
}

func NuevoSemaphore(n int) *Semaphore {
    return &Semaphore{
        tokens: make(chan struct{}, n),  // N permisos
    }
}

func (s *Semaphore) Adquirir() {
    s.tokens <- struct{}{}  // Toma un token (bloquea si no hay)
}

func (s *Semaphore) Liberar() {
    <-s.tokens  // Devuelve el token
}
```

### ¿Cuándo usar Semaphore vs Worker Pool?

| Situación | Patrón recomendado |
|-----------|-------------------|
| Procesar una cola de trabajos conocida | Worker Pool |
| Limitar conexiones a una base de datos | **Semaphore** |
| Controlar acceso a un archivo | **Semaphore** |
| Procesar requests HTTP con límite | Worker Pool o Semaphore |

**🔑 Regla de oro:** Usa Worker Pool cuando el **trabajo es la cola**. Usa Semaphore cuando el **recurso es el cuello de botella**.

---

## ⏱️ Patrón 5: Rate Limiter — Token Bucket

### Concepto

Un Rate Limitador controla **la velocidad** a la que se ejecutan operaciones. El patrón **Token Bucket** funciona así:

1. Un "bucket" se llena con tokens a intervalos regulares
2. Cada operación necesita tomar un token
3. Si no hay tokens, la operación espera

```
Tiempo →    Token   Token   Token   Token
            ↓       ↓       ↓       ↓
Bucket:     [🪙]    [🪙]    [🪙]    [🪙]
            │
         Request 1 toma token → se ejecuta
            [vacío]
                │
             Request 2 espera...
                    │
                  Token llega → Request 2 se ejecuta
```

### Implementación con Ticker

```go
type RateLimitador struct {
    ticker *time.Ticker
    tokens chan struct{}
}

func NuevoRateLimitador(intervalo time.Duration) *RateLimitador {
    rl := &RateLimitador{
        ticker: time.NewTicker(intervalo),
        tokens: make(chan struct{}, 1),  // Bucket con capacidad 1
    }

    // Goroutine que llena el bucket a intervalos regulares
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

func (rl *RateLimitador) Esperar() {
    <-rl.tokens  // Bloquea hasta que haya un token
}

func (rl *RateLimitador) Detener() {
    rl.ticker.Stop()  // Limpiar recursos
}
```

### Uso Práctico

```go
// 1 petición cada 300ms (~3.3 req/segundo)
limiter := NuevoRateLimitador(300 * time.Millisecond)
defer limiter.Detener()

for _, url := range urls {
    limiter.Esperar()  // Esperar token
    go descargar(url)  // Ejecutar con rate limit
}
```

**🔑 ¿Por qué `select` con `default`?** Sin el `default`, si el bucket ya tiene un token, el ticker se bloquearía y dejaría de generar tokens. Con `default`, si el bucket está lleno, simplemente descartamos el token extra.

---

## 🕷️ Proyecto Final: Crawler Web Concurrente

Ahora combinamos **todos los patrones** en un crawler web real:

```
URLs Semilla
    │
    ▼
┌─────────────────────────────────────┐
│           CANAL DE TRABAJOS         │
│  URL1 │ URL2 │ URL3 │ ... │ URL12   │
└───────┴──────┴──────┴─────┴─────────┘
    │
    ▼ (Fan-Out: 3 workers)
┌──────────┐ ┌──────────┐ ┌──────────┐
│ Worker 1 │ │ Worker 2 │ │ Worker 3 │
│          │ │          │ │          │
│ ┌──────┐ │ │ ┌──────┐ │ │ ┌──────┐ │
│ │ Rate │ │ │ │ Rate │ │ │ │ Rate │ │  ← Rate Limiter
│ │Limit │ │ │ │Limit │ │ │ │Limit │ │
│ └──────┘ │ │ └──────┘ │ │ └──────┘ │
│ ┌──────┐ │ │ ┌──────┐ │ │ ┌──────┐ │
│ │Semaph│ │ │ │Semaph│ │ │ │Semaph│ │  ← Semaphore
│ └──────┘ │ │ └──────┘ │ │ └──────┘ │
└──────────┘ └──────────┘ └──────────┘
    │              │              │
    ▼              ▼              ▼
┌─────────────────────────────────────┐
│     CANAL DE RESULTADOS (Fan-In)    │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│         RECOLECTOR DE RESULTADOS    │
│   • Imprime cada página descargada  │
│   • Actualiza estadísticas           │
└─────────────────────────────────────┘
```

### Código del Crawler (Línea por Línea)

#### Estructuras de Datos

```go
// PaginaWeb representa una página descargada
type PaginaWeb struct {
    URL        string
    Contenido  string
    Tamano     int
    Latencia   time.Duration
    StatusCode int
    WorkerID   int    // Qué worker la procesó
    Error      string
}

// URLTrabajo representa un trabajo pendiente
type URLTrabajo struct {
    URL        string
    Profundidad int
}

// EstadisticasCrawl rastrea métricas del crawler
type EstadisticasCrawl struct {
    PaginasDescargadas int64
    PaginasError       int64
    BytesTotal         int64
    TiempoTotal        time.Duration
}
```

**¿Por qué `int64` en las estadísticas?** Porque `atomic.AddInt64` requiere `int64`. Los operadores atómicos son la forma más segura de compartir datos entre goroutines sin usar `Mutex`.

#### Configuración del Crawler

```go
maxWorkers := 3                      // 3 workers simultáneos
maxConexiones := 2                   // Semaphore: max 2 conexiones a la vez
intervaloRateLimit := 200 * time.Millisecond  // 1 req / 200ms

sem := NuevoSemaphore(maxConexiones)
limiter := NuevoRateLimitador(intervaloRateLimit)
defer limiter.Detener()
```

#### Workers con Rate Limiting y Semaphore

```go
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

            // 3. Descargar la página (sección crítica)
            pagina := descargarPagina(trabajo.URL)

            // 4. Liberar semaphore (¡inmediatamente después!)
            sem.Liberar()

            // 5. Actualizar estadísticas (thread-safe)
            if pagina.Error != "" {
                atomic.AddInt64(&estadisticas.PaginasError, 1)
            } else {
                atomic.AddInt64(&estadisticas.PaginasDescargadas, 1)
                atomic.AddInt64(&estadisticas.BytesTotal, int64(pagina.Tamano))
            }

            // 6. Enviar resultado al canal unificado
            resultados <- ResultadoCrawl{
                Pagina:      pagina,
                Profundidad: trabajo.Profundidad,
                Procesado:   time.Now(),
            }
        }
    }()
}
```

**🔑 El orden importa:**
1. `limiter.Esperar()` → Primero respetamos el rate limit
2. `sem.Adquirir()` → Luego tomamos el permiso de conexión
3. Descargamos → Operación de red
4. `sem.Liberar()` → Liberamos el permiso **inmediatamente**
5. Actualizamos estadísticas → Operación atómica
6. Enviamos resultado → Al canal de Fan-In

### Salida Esperada

```
═══════════════════════════════════════════════════════════
  🕷️  DEMO 5: Crawler Web Completo
  Fan-Out + Worker Pool + Semaphore + Rate Limiter
═══════════════════════════════════════════════════════════

    🏭 Workers:         3
    🔒 Max conexiones: 2 (semaphore)
    ⏱️  Rate limit:      1 req / 200ms

    ✅ W1 | https://ejemplo.com/                    | HTTP 200 |   3842 bytes | 156ms
    ✅ W2 | https://ejemplo.com/productos            | HTTP 200 |   7231 bytes | 234ms
    ❌ W3 | https://ejemplo.com/blog                 | HTTP 503 |      0 bytes | 512ms
    ✅ W1 | https://ejemplo.com/contacto             | HTTP 200 |   1567 bytes | 189ms
    ...

    ═══════════════════════════════════════════════════
    📊 REPORTE FINAL DEL CRAWLER
    ═══════════════════════════════════════════════════
    🕷️  URLs procesadas:       12
    🏭 Workers utilizados:    3
    🔒 Conexiones máximas:    2 (semaphore)
    ⏱️  Rate limit aplicado:   1 req / 200ms
    ✅ Descargas exitosas:    10
    ❌ Descargas fallidas:    2
    💾 Bytes totales:         45678
    ⏱️  Tiempo total:          3.456s
    ═══════════════════════════════════════════════════
```

---

## 🧠 Tabla de Decisión: ¿Qué Patrón Usar?

| Situación | Patrón | Ejemplo |
|-----------|--------|---------|
| Procesar N items en paralelo | **Fan-Out** | Descargar 100 archivos |
| Consolidar resultados de N goroutines | **Fan-In** | Juntar logs de 5 servidores |
| Límite fijo de goroutines activas | **Worker Pool** | Procesar cola de emails |
| Limitar acceso a recurso compartido | **Semaphore** | Max 5 conexiones a DB |
| Controlar velocidad de peticiones | **Rate Limiter** | API con límite de 100 req/min |
| Combinar todo | **Crawler** | Web scraper profesional |

---

## 🔑 Atomic vs Mutex: ¿Cuándo usar cada uno?

### Atomic (lo que usamos en el proyecto)

```go
// Operaciones simples: incrementar, leer, escribir
atomic.AddInt64(&contador, 1)     // +1 thread-safe
valor := atomic.LoadInt64(&contador) // Leer thread-safe
```

**Ventaja:** Ultra rápido (1 instrucción de CPU)
**Limitación:** Solo para operaciones simples (int64, float64, pointer)

### Mutex (para operaciones complejas)

```go
var mu sync.Mutex
var datos []string

mu.Lock()
datos = append(datos, "nuevo")  // Operación compleja
mu.Unlock()
```

**Ventaja:** Puede proteger cualquier operación
**Limitación:** Más lento que atomic, posible deadlock

### ¿Cuándo usar cada uno?

| Operación | Usa |
|-----------|-----|
| Incrementar un contador | `atomic.AddInt64` |
| Leer/escribir un bool | `atomic.LoadBool` / `StoreBool` |
| Modificar un slice | `sync.Mutex` |
| Actualizar un map | `sync.Mutex` o `sync.Map` |
| Operaciones compuestas | `sync.Mutex` |

---

## 🎯 Ejercicio Feynman

### Instrucciones

1. **Explica en voz alta** (o escribe) qué hace cada patrón con tus propias palabras
2. **Dibuja el diagrama** de flujo de datos para cada patrón
3. **Completa los ejercicios** a continuación

### Ejercicio 1: Fan-Out con Colores

Modifica el código para que cada worker use un emoji diferente:

```
Worker 1: 🔴
Worker 2: 🟡
Worker 3: 🟢
Worker 4: 🔵
```

Cada resultado debe mostrar el emoji correspondiente al worker que lo procesó.

### Ejercicio 2: Worker Pool con Tiempo Máximo

Agrega un **timeout de 2 segundos** al Worker Pool. Si un trabajo tarda más de 2 segundos, debe ser cancelado y marcado como "timeout".

**Pista:** Usa `select` con `time.After()`:

```go
select {
case resultado := <-canal:
    // Procesar resultado
case <-time.After(2 * time.Second):
    // Timeout
}
```

### Ejercicio 3: Semaphore Dinámico

Crea un semaphore que pueda **cambiar su capacidad** en tiempo de ejecución:

```go
sem := NuevoSemaphoreDinamico(2)   // Empezar con 2
sem.Ajustar(5)                      // Cambiar a 5 en caliente
```

**Pista:** Usa un `sync.RWMutex` para proteger la capacidad actual.

### Ejercicio 4: Rate Limiter con Burst

Modifica el Rate Limiter para permitir **ráfagas** (burst): que pueda ejecutar N operaciones instantáneamente, y luego aplicar el rate limit.

```go
limiter := NuevoRateLimitadorConBurst(
    200 * time.Millisecond,  // Intervalo entre tokens
    5,                        // Burst: 5 tokens iniciales
)
```

### Ejercicio 5: Crawler con Profundidad

Modifica el crawler para que, cuando descargue una página exitosamente, **extraiga los links** y los agregue como nuevos trabajos al canal:

```
Nivel 0: https://ejemplo.com/         → encuentra 3 links
Nivel 1: /productos, /blog, /about   → cada uno encuentra 2 links
Nivel 2: 6 links nuevos               → (max profundidad = 2, se detiene)
```

**Pista:** Usa un contador de profundidad en `URLTrabajo` y solo agrega links si `profundidad < maxProfundidad`.

---

## 📝 Resumen de la Lección

```
┌─────────────────────────────────────────────────────────────┐
│                  PATRONES DE CONCURRENCIA                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Fan-Out     → 1 canal → N goroutines (distribuir trabajo)  │
│  Fan-In      → N canales → 1 canal (converger resultados)   │
│  Worker Pool → N workers fijos procesando cola continua      │
│  Semaphore   → Limitar acceso a recurso con N tokens         │
│  Rate Limit  → 1 operación cada T tiempo (token bucket)      │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                  THREAD-SAFETY                              │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  atomic.AddInt64  → Contadores rápidos                       │
│  sync.Mutex       → Operaciones complejas                    │
│  sync.WaitGroup   → Esperar a que terminen N goroutines      │
│  chan struct{}     → Señalización (semáforo, done)            │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                  REGLAS DE ORO                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. SIEMPRE cerrar los channels cuando el productor termina  │
│  2. SIEMPRE usar WaitGroup para esperar goroutines           │
│  3. SIEMPRE usar defer wg.Done() al inicio de la goroutine   │
│  4. SIEMPRE usar defer close(ch) al inicio de la goroutine   │
│  5. NUNCA enviar a un channel cerrado (panic)                │
│  6. NUNCA usar variables compartidas sin atomic/mutex        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 Siguiente Lección

En la **Lección 14** aprenderemos sobre **Context en Go** (`context.Context`):
- Cómo cancelar goroutines de forma limpia
- Pasar valores a través de la cadena de llamadas
- Timeouts automáticos en operaciones
- `context.WithCancel`, `context.WithTimeout`, `context.WithValue`

El `context.Context` es la pieza que falta para que nuestros patrones de concurrencia sean **completamente profesionales**.

---

> *"La mejor forma de aprender concurrencia es practicándola. No leas este código — ejecútalo, modifíalo, rómpelo, arréglalo."*
> — Filosofía Feynman