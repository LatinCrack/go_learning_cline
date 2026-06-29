# 🏗️ Lección 20: Proyecto Final — GoKV (Base de Datos Key-Value)

> **"El conocimiento que no se aplica, se evapora."**
> — Metodología Feynman

---

## 🎯 Objetivo de la Lección

Construir **GoKV**, una base de datos Key-Value real y funcional que integra **TODOS** los conceptos aprendidos en las 19 lecciones anteriores. Este proyecto es tu examen final: demuestra que puedes combinar structs, concurrencia, persistencia, redes y más en un solo programa cohesivo.

```
┌─────────────────────────────────────────────────────┐
│                   GOKV - Arquitectura                │
│                                                      │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐       │
│  │  Cliente  │◄──►│ Servidor │◄──►│ Database │       │
│  │   CLI     │TCP │   TCP    │    │  (Map)   │       │
│  └──────────┘    └──────────┘    └────┬─────┘       │
│                                       │              │
│                                  ┌────▼─────┐       │
│                                  │   Disco   │       │
│                                  │  (JSON)   │       │
│                                  └──────────┘       │
└─────────────────────────────────────────────────────┘
```

---

## 📖 Teoría: ¿Qué Construiremos?

### Analogía del Archivador Completo 🗄️

Imagina que eres el gerente de una oficina y necesitas un sistema para guardar documentos. Tu sistema tiene **5 capas**:

| Capa | Componente | Analogía de Oficina |
|------|-----------|---------------------|
| **1** | Motor en Memoria | El archivador con cajones donde guardas papeles |
| **2** | Persistencia a Disco | El respaldo fotocopiado que guardas en otra bóveda |
| **3** | Servidor TCP | El mostrador de atención al público |
| **4** | Transacciones | Los borradores que puedes desechar o publicar |
| **5** | Cliente CLI | El formulario que llenan los visitantes |

Cada capa usa conceptos específicos de Go que hemos aprendido:

### Mapa de Conceptos por Capa 🗺️

```
CAPA 1: Motor en Memoria
├── sync.RWMutex     → L10 (Goroutines) + L11 (Channels)
├── map[string]Entry → L07 (Maps)
├── Entry struct     → L05 (Structs)
└── Punteros (*DB)   → L09 (Punteros)

CAPA 2: Persistencia
├── os.ReadFile      → L16 (OS/IO)
├── json.Marshal     → L17 (JSON/APIs)
├── Snapshot struct   → L05 (Structs)
└── defer db.Close() → L18 (Context/Defer)

CAPA 3: Servidor TCP
├── net.Listen       → L10 (Goroutines)
├── go handleClient()→ L10 (Goroutines)
├── bufio.Scanner    → L08 (Strings)
├── context.Context  → L18 (Context)
└── select {}        → L12 (Select)

CAPA 4: Transacciones
├── Slice de ops     → L06 (Slices)
├── switch/case      → L03 (Tipos)
└── Atomicidad       → L13 (Patrones)

CAPA 5: Cliente CLI
├── net.Dial         → L10 (Goroutines)
├── strings.Fields   → L08 (Strings)
└── fmt.Printf       → L01 (Hello Go)
```

---

## 🔬 Desglose del Código

### CAPA 1: Motor de Almacenamiento en Memoria

El corazón de GoKV es un `map[string]Entry` protegido por un `sync.RWMutex`.

**¿Por qué RWMutex y no Mutex?**

```
Mutex simple (como un baño con una sola llave):
  ┌────────┐     ┌────────┐     ┌────────┐
  │ Lector1 │────►│ BLOQUEO │     │ Lector2 │ (espera...)
  └────────┘     └────────┘     └────────┘

RWMutex (como una sala de lectura con reglas):
  ┌────────┐     ┌──────────┐     ┌────────┐
  │ Lector1 │────►│          │◄────│ Lector2 │ (¡simultáneos!)
  └────────┘     │ RLock()  │     └────────┘
  ┌────────┐     │          │     ┌────────┐
  │ Lector3 │────►│          │     │ Escritor│ (espera a que todos terminen)
  └────────┘     └──────────┘     └────────┘
```

- `RLock()` → Múltiples lectores pueden leer AL MISMO TIEMPO
- `Lock()` → Solo UN escritor, y BLOQUEA a todos los demás

```go
type Entry struct {
    Key       string    `json:"key"`        // La etiqueta del cajón
    Value     string    `json:"value"`      // El documento dentro
    CreatedAt time.Time `json:"created_at"` // Cuándo se creó
    UpdatedAt time.Time `json:"updated_at"` // Última modificación
}
```

**¿Por qué guardamos `CreatedAt` y `UpdatedAt`?**

Porque en una base de datos real, saber **CUÁNDO** se creó o modificó un dato es tan importante como el dato mismo. Es como el sello de fecha en un documento oficial.

```go
type Database struct {
    mu       sync.RWMutex      // Cerrojo de seguridad
    data     map[string]Entry  // El archivador
    dumpPath string            // Ruta del respaldo en disco
    dirty    bool              // ¿Hay cambios sin guardar?
}
```

**El campo `dirty`** es como una bandera en tu escritorio que dice "hay papeles sin archivar". Si el `dirty` es `true`, sabemos que debemos guardar antes de cerrar.

#### Put: Insertar o Actualizar

```go
func (db *Database) Put(key, value string) {
    db.mu.Lock()         // Cerrojo de ESCRITURA
    defer db.mu.Unlock() // Liberar al finalizar (¡siempre!)

    now := time.Now()

    if existing, existe := db.data[key]; existe {
        // Ya existe → actualizar, PERO preservar CreatedAt
        existing.Value = value
        existing.UpdatedAt = now
        db.data[key] = existing
    } else {
        // No existe → crear nueva entrada
        db.data[key] = Entry{
            Key:       key,
            Value:     value,
            CreatedAt: now,
            UpdatedAt: now,
        }
    }

    db.dirty = true // Marcar que hay cambios sin persistir
}
```

**Puntos clave:**
- `defer db.mu.Unlock()` → Si olvidas esto, el programa se **cuelga para siempre** (deadlock)
- Preservamos `CreatedAt` al actualizar → La fecha de creación NUNCA cambia
- `db.dirty = true` → Como poner una nota "recordar guardar"

#### Get: Consultar

```go
func (db *Database) Get(key string) (Entry, bool) {
    db.mu.RLock()         // Solo LECTURA (permite otros lectores)
    defer db.mu.RUnlock()

    entry, existe := db.data[key]
    return entry, existe
}
```

Usa `RLock` (no `Lock`) porque solo estamos leyendo. Esto permite que múltiples clientes consulten simultáneamente sin bloquearse entre sí.

#### Delete: Eliminar

```go
func (db *Database) Delete(key string) bool {
    db.mu.Lock()
    defer db.mu.Unlock()

    _, existe := db.data[key]
    if existe {
        delete(db.data, key)  // Función built-in de Go
        db.dirty = true
        return true
    }
    return false
}
```

`delete()` es una función built-in de Go para eliminar entradas de un map. No retorna nada; si la clave no existe, no hace nada (no es un error).

---

### CAPA 2: Persistencia a Disco

**El problema:** Cuando el programa se cierra, la memoria RAM se borra. Todos los datos se pierden.

**La solución:** Serializar el map a JSON y guardarlo en un archivo.

**Analogía:** Es como fotocopiar todos los documentos del archivador al final del día y guardar las copias en una bóveda.

#### El Patrón "Write-and-Rename" ✍️

```
Paso 1: Escribir a archivo TEMPORAL
  datos.json.tmp  ← (escritura nueva)

Paso 2: Renombrar temporal → final
  datos.json.tmp  →  datos.json  (atómico)

¿Por qué? Si se corta la luz durante la escritura:
  - Sin patrón: datos.json queda CORRUPTO 💀
  - Con patrón: datos.json intacto, .tmp parcial (lo ignoramos) ✅
```

```go
func (db *Database) SaveToDisk() error {
    db.mu.RLock()
    // Crear snapshot de los datos
    entries := make([]Entry, 0, len(db.data))
    for _, entry := range db.data {
        entries = append(entries, entry)
    }
    db.mu.RUnlock()

    snapshot := Snapshot{
        Entries:   entries,
        Timestamp: time.Now().Format(time.RFC3339),
        Count:     len(entries),
    }

    jsonData, err := json.MarshalIndent(snapshot, "", "  ")
    if err != nil {
        return fmt.Errorf("error serializando datos: %w", err)
    }

    // Write-and-rename: patrón seguro contra corrupción
    tmpPath := db.dumpPath + ".tmp"
    if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
        return fmt.Errorf("error escribiendo temporal: %w", err)
    }

    if err := os.Rename(tmpPath, db.dumpPath); err != nil {
        return fmt.Errorf("error renombrando: %w", err)
    }

    db.mu.Lock()
    db.dirty = false
    db.mu.Unlock()

    return nil
}
```

**Conceptos de lecciones anteriores aplicados:**
- `json.MarshalIndent` → L17 (JSON)
- `os.WriteFile` → L16 (OS/IO)
- `fmt.Errorf("...: %w", err)` → L04 (Funciones, error wrapping)
- `defer`, `sync.RWMutex` → L18 (Context/Defer), L10 (Concurrencia)

---

### CAPA 3: Servidor TCP

**¿Qué es TCP?** Es el protocolo que usan las computadoras para conversar de forma confiable.

**Analogía:** TCP es como una llamada de teléfono:
1. Marcas el número → `net.Dial()`
2. La otra persona contesta → `net.Listen()` + `Accept()`
3. Hablan ida y vuelta → `Read()` / `Write()`
4. Cuelan → `Close()`

```
CLIENTE                          SERVIDOR
   │                                │
   │──── Conexión TCP ─────────────►│  net.Listen() + Accept()
   │                                │
   │──── "PUT nombre Ana" ─────────►│  scanner.Scan()
   │                                │  db.Put("nombre", "Ana")
   │◄─── "ok | 'nombre' guardado"──│  fmt.Fprintln()
   │                                │
   │──── "GET nombre" ─────────────►│
   │◄─── "ok | encontrado | Ana" ──│
   │                                │
   │──── "QUIT" ────────────────────►│
   │◄─── "ok | adiós 👋" ──────────│
   │                                │
```

#### El Servidor: Un Goroutine por Cliente

```go
func (s *Servidor) Start(ctx context.Context) error {
    for {
        conn, err := s.listener.Accept()  // Esperar cliente
        if err != nil { /* manejar error */ }

        // Cada cliente recibe su PROPIA goroutine
        s.wg.Add(1)
        go s.handleClient(ctx, conn)  // ← ¡Concurrencia!
    }
}
```

**Esto es el patrón fan-out de L13:** Un goroutine principal acepta conexiones y distribuye el trabajo a goroutines dedicadas.

```
                    ┌──► handleClient(goroutine) ──► Cliente A
                    │
listener.Accept() ──┼──► handleClient(goroutine) ──► Cliente B
                    │
                    └──► handleClient(goroutine) ──► Cliente C
```

#### Protocolo de Texto Simple

Usamos un protocolo de texto línea por línea (no binario):

```
CLIENTE ENVÍA:           SERVIDOR RESPONDE:
─────────────────        ─────────────────────
PUT nombre Ana           ok | 'nombre' guardado
GET nombre               ok | encontrado | Ana
GET xyz                  not_found | 'xyz' no existe
DEL nombre               ok | 'nombre' eliminado
LIST                     ok | 3 claves | a, b, c
PING                     ok | PONG 🏓
LOL                      error | comando desconocido
```

**Formato de respuesta:** `STATUS | message [| data]`

#### Parser de Comandos con Soporte de Comillas

```go
func parseCommand(line string) []string {
    var parts []string
    var current strings.Builder
    inQuotes := false

    for _, ch := range line {
        switch {
        case ch == '"':
            inQuotes = !inQuotes
        case ch == ' ' && !inQuotes:
            if current.Len() > 0 {
                parts = append(parts, current.String())
                current.Reset()
            }
        default:
            current.WriteRune(ch)
        }
    }

    if current.Len() > 0 {
        parts = append(parts, current.String())
    }

    return parts
}
```

**¿Por qué necesitamos comillas?**

```
Sin comillas:
  PUT nombre Juan Pérez  →  ["PUT", "nombre", "Juan", "Pérez"]  ❌
                           (¿"Pérez" es parte del valor o un error?)

Con comillas:
  PUT nombre "Juan Pérez" → ["PUT", "nombre", "Juan Pérez"]     ✅
                           (Todo el nombre es un solo valor)
```

---

### CAPA 4: Transacciones

**El problema:** ¿Qué pasa si quieres hacer 3 cambios a la vez, pero solo quieres que se apliquen TODOS o NINGUNO?

**La solución:** Transacciones.

**Analogía del borrador:** 📝

```
Sin transacción (peligroso):
  1. PUT salario 5000     ← ¡Ya se aplicó!
  2. PUT departamento IT  ← ¡Ya se aplicó!
  3. PUT nivel senior     ← ERROR 💥
  Resultado: salario y departamento cambiaron, nivel NO
  → ¡Datos inconsistentes!

Con transacción (seguro):
  BEGIN
  1. PUT salario 5000     ← Solo en el borrador
  2. PUT departamento IT  ← Solo en el borrador
  3. PUT nivel senior     ← ERROR 💥
  ROLLBACK                ← Tiramos el borrador entero
  Resultado: NADA cambió. Datos intactos. ✅

  BEGIN
  1. PUT salario 5000     ← Solo en el borrador
  2. PUT departamento IT  ← Solo en el borrador
  3. PUT nivel senior     ← OK
  COMMIT                  ← Publicamos todos los cambios
  Resultado: Los 3 cambios aplicados juntos. ✅
```

```go
type TxOperation struct {
    Action string // "PUT" o "DEL"
    Key    string
    Value  string
}

type Transaction struct {
    db         *Database
    operations []TxOperation  // Slice de operaciones pendientes
    IsActive   bool
}
```

**Commit aplica todo:**
```go
func (tx *Transaction) Commit() string {
    tx.IsActive = false

    puts := 0
    dels := 0

    for _, op := range tx.operations {
        switch op.Action {
        case "PUT":
            tx.db.Put(op.Key, op.Value)
            puts++
        case "DEL":
            tx.db.Delete(op.Key)
            dels++
        }
    }

    return fmt.Sprintf("%d PUTs, %d DELs", puts, dels)
}
```

**Rollback descarta todo:**
```go
func (tx *Transaction) Rollback() int {
    tx.IsActive = false
    count := len(tx.operations)
    tx.operations = nil  // Liberar memoria (el "borrador" a la basura)
    return count
}
```

---

### CAPA 5: Cliente CLI Interactivo

El cliente es la interfaz de usuario. Se conecta al servidor vía TCP y permite enviar comandos interactivamente.

```go
func (c *ClienteTCP) RunCLI() {
    scanner := bufio.NewScanner(os.Stdin)

    for {
        fmt.Print("gokv> ")
        if !scanner.Scan() {
            break
        }

        line := strings.TrimSpace(scanner.Text())
        resp, err := c.Send(line)  // Enviar al servidor
        // ... mostrar respuesta ...
    }
}
```

**Flujo completo del cliente:**
```
1. go run main.go --client
2. Se conecta a localhost:6969 (net.Dial)
3. Envía PING para verificar conexión
4. Loop: Leer comando → Enviar → Mostrar respuesta
5. QUIT → Cerrar conexión
```

---

## 🧪 Ejercicio Práctico: Prueba GoKV

### Paso 1: Iniciar el Servidor

```bash
cd 20-proyecto-final
go run main.go
```

Verás:
```
============================================================
🏗️  GOKV — Base de Datos Key-Value en Go
============================================================

📦 Inicializando motor de almacenamiento...
   ✅ Motor listo (0 entradas en memoria)

🚀 Servidor GoKV iniciado
   📡 Puerto: 6969
```

### Paso 2: Conectar el Cliente (otra terminal)

```bash
cd 20-proyecto-final
go run main.go --client
```

### Paso 3: Experimentar con Comandos

```
gokv> PING
   ✅ PONG 🏓

gokv> PUT nombre "Ana García"
   ✅ 'nombre' guardado

gokv> PUT email "ana@mail.com"
   ✅ 'email' guardado

gokv> PUT edad 25
   ✅ 'edad' guardado

gokv> GET nombre
   ✅ encontrado (creado: 2024-...)
   📦 Ana García

gokv> LIST
   ✅ 3 claves encontradas
   📦 edad, email, nombre

gokv> COUNT
   ✅ 3 entradas en la base de datos

gokv> DEL edad
   ✅ 'edad' eliminado

gokv> BEGIN
   ✅ transacción iniciada 📝

gokv> PUT telefono "999-888-777"
   ✅ PUT 'telefono' registrado en transacción

gokv> PUT direccion "Av. Siempre Viva 742"
   ✅ PUT 'direccion' registrado en transacción

gokv> COMMIT
   ✅ transacción confirmada ✅ (2 PUTs, 0 DELs)

gokv> LIST
   ✅ 4 claves encontradas
   📦 direccion, email, nombre, telefono

gokv> QUIT
   ✅ adiós 👋
```

### Paso 4: Verificar Persistencia

Detén el servidor (Ctrl+C) y revisa el archivo:

```bash
cat gokv-dump.json
```

Verás todos los datos en JSON. Reinicia el servidor y los datos seguirán ahí.

---

## 🧠 Ejercicio Feynman: Explica GoKV a Otro

### Instrucciones

Toma un cuaderno o abre un editor de texto. **Explica cada capa de GoKV como si le hablaras a alguien que sabe Go pero no ha visto el código.** Usa tus propias palabras y analogías.

### Preguntas Guía

#### Nivel 1: Motor de Memoria 🗄️

1. **¿Por qué usamos `sync.RWMutex` y no `sync.Mutex`?**
   - Pista: Piensa en una biblioteca. ¿Cuántas personas pueden leer el mismo libro a la vez?

2. **¿Qué pasa si olvidas `defer db.mu.Unlock()`?**
   - Pista: ¿Qué pasa en una puerta de baño si nadie devuelve la llave?

3. **¿Por qué preservamos `CreatedAt` al hacer Put sobre una clave existente?**
   - Pista: Cuando editas un documento de Word, ¿cambia la fecha de creación?

#### Nivel 2: Persistencia 💾

4. **¿Por qué escribimos a un archivo `.tmp` primero y luego renombramos?**
   - Pista: ¿Qué pasa si se corta la luz mientras escribes? ¿El archivo queda corrupto?

5. **¿Por qué usamos `json.MarshalIndent` y no `json.Marshal`?**
   - Pista: ¿Qué es más fácil de leer: todo pegado o con espacios?

#### Nivel 3: Servidor TCP 🌐

6. **¿Por qué cada cliente se maneja en una goroutine separada?**
   - Pista: ¿Prefieres una tienda con 1 cajero o con 1 cajero por cliente?

7. **¿Qué hace `context.WithCancel` y por qué es importante?**
   - Pista: ¿Cómo le dices a 10 goroutines "paren todo, nos vamos"?

8. **¿Por qué el parser de comandos necesita manejar comillas?**
   - Pista: `PUT nombre Juan Pérez` → ¿"Pérez" es parte del valor o un segundo argumento?

#### Nivel 4: Transacciones 📝

9. **¿Por qué las operaciones de una transacción no se aplican inmediatamente?**
   - Pista: ¿Publicas un borrador antes de revisarlo?

10. **¿Qué diferencia hay entre Commit y Rollback?**
    - Pista: ¿Qué haces con un borrador que tiene errores?

#### Nivel 5: Integración Total 🔗

11. **Enumera al menos 10 lecciones anteriores y explica qué concepto de cada una usamos en GoKV.**

12. **Si quisieras agregar un comando `KEYS <pattern>` que busque claves con comodines (ej: `KEYS user_*`), ¿qué cambiarías?**
    - Pista: Piensa en `strings.HasPrefix` o `strings.Contains`.

13. **Dibuja el flujo completo de un comando `PUT nombre "Ana"` desde que el cliente lo escribe hasta que se guarda en disco.**
    - Incluye: Cliente → TCP → Servidor → handleClient → Database.Put → dirty flag → SaveToDisk

---

## 📋 Resumen de la Lección

```
┌─────────────────────────────────────────────────────────────┐
│                   GOKV - Resumen Final                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  🗄️  Motor de Memoria                                       │
│     • map[string]Entry con RWMutex                           │
│     • Put, Get, Delete, List, Count                          │
│     • Concurrencia segura con RLock/Lock                     │
│                                                              │
│  💾  Persistencia                                            │
│     • Snapshot JSON en disco                                 │
│     • Patrón write-and-rename (anti-corrupción)              │
│     • Auto-save periódico con ticker                         │
│                                                              │
│  🌐  Servidor TCP                                            │
│     • Un goroutine por cliente (fan-out)                     │
│     • Protocolo de texto línea por línea                     │
│     • Context para cancelación limpia                        │
│                                                              │
│  📝  Transacciones                                           │
│     • BEGIN / COMMIT / ROLLBACK                              │
│     • Operaciones acumuladas en slice                        │
│     • Atomicidad: todo o nada                                │
│                                                              │
│  🔗  Cliente CLI                                             │
│     • Conexión con net.Dial                                  │
│     • Loop interactivo con bufio.Scanner                     │
│     • Respuestas formateadas con emojis                      │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│  📚  19 lecciones integradas en un solo proyecto             │
│  🎯  De "Hello World" a base de datos funcional              │
│  🚀  Listo para proyectos reales en Go                       │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎓 Felicidades: Has Completado el Laboratorio Feynman de Go

```
  L01 ─► L02 ─► L03 ─► L04 ─► L05 ─► L06 ─► L07 ─► L08 ─► L09
   │      │      │      │      │      │      │      │      │
   └──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┘
                           │
                        GOKV 🏗️
                           │
  L10 ─► L11 ─► L12 ─► L13 ─► L14 ─► L15 ─► L16 ─► L17 ─► L18 ─► L19
   │      │      │      │      │      │      │      │      │      │
   └──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┴──────┘
                           │
                    PROYECTO FINAL ✅
```

**De aquí en adelante, ya no eres un principiante en Go.** Tienes las herramientas para construir:
- Servidores web
- APIs REST
- Microservicios
- Herramientas de línea de comandos
- Bases de datos
- Sistemas distribuidos

> **"Si no puedes explicarlo de forma simple, no lo entiendes lo suficiente."**
> — Richard Feynman

¡Ahora ve y construye algo increíble! 🚀

---