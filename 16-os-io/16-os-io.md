# 📂 Lección 16: Paquete `os`, `io` y el Mundo de los Archivos y Procesos

## 🎯 Objetivo de la Lección

Aprender a interactuar con el **sistema operativo** y el **sistema de archivos** usando los paquetes estándar de Go: `os`, `io`, `filepath`, `bufio` y `flag`. Construiremos una **herramienta CLI de backup incremental** que demuestre cómo leer, escribir, copiar archivos y recorrer directorios de forma eficiente.

---

## 🧠 Analogía Fundamental: El Archivero Universal

Imagina que eres el **director de una oficina** y necesitas:

- **Abrir cajones** para leer documentos → `os.Open`
- **Crear archivos nuevos** en carpetas → `os.Create`
- **Fotocopiar documentos** de un lugar a otro → `io.Copy`
- **Recorrer todas las carpetas** de un archivador → `filepath.WalkDir`
- **Verificar la huella digital** de cada documento → `sha256` con `io.Copy`

En este mundo, hay dos **interfaces mágicas** que todo documento respeta:

> 🔵 **`io.Reader`** → "Puedo ser leído" (como un documento que sacas de un cajón)
> 
> 🔴 **`io.Writer`** → "Puedo ser escrito" (como una hoja en blanco donde anotas)

La genialidad de Go es que **no importa si el documento viene de un archivo, de internet, de un string o de una tubería comprimida** — si puede leerse, es un `Reader`. Si puede escribirse, es un `Writer`. Y `io.Copy` conecta cualquier Reader con cualquier Writer como una fotocopiadora universal.

---

## 📦 Paquetes que Estudiaremos

| Paquete | ¿Qué hace? | Analogía |
|---------|-------------|----------|
| `os` | Interactúa con el SO: archivos, env vars, procesos | El **edificio** de la oficina |
| `io` | Interfaces para flujo de datos: Reader, Writer, Copy | La **fotocopiadora universal** |
| `bufio` | Lectura/escritura con buffer (eficiente) | El **asistente** que acumula hojas antes de enviarlas |
| `filepath` | Manipulación de rutas multiplataforma | El **GPS** que sabe las calles de Linux y Windows |
| `flag` | Parsing de argumentos de línea de comandos | El **recepcionista** que interpreta tus pedidos |
| `crypto/sha256` | Hashing criptográfico | El **notario** que certifica documentos |

---

## 🔑 Interfaces Clave: `io.Reader` y `io.Writer`

### io.Reader

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

**Cualquier tipo** que implemente el método `Read()` es un `io.Reader`. Esto incluye:
- `*os.File` — archivos en disco
- `*strings.Reader` — strings en memoria
- `*bytes.Buffer` — buffers en memoria
- `*net.Conn` — conexiones de red
- `*gzip.Reader` — archivos comprimidos

### io.Writer

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

**Cualquier tipo** que implemente el método `Write()` es un `io.Writer`:
- `*os.File` — archivos en disco
- `*bytes.Buffer` — buffers en memoria
- `*net.Conn` — conexiones de red
- `*gzip.Writer` — compresión
- `hash.Hash` — cálculo de hashes (SHA-256, MD5)

### ⭐ io.Copy — La Fotocopiadora Universal

```go
func copy(dst Writer, src Reader) (written int64, err error)
```

`io.Copy` toma un **Reader** y un **Writer**, y transfiere datos del uno al otro. Usa un buffer interno de 32KB, así que es eficiente para archivos de cualquier tamaño, desde 1 byte hasta varios gigabytes, **sin cargar todo en memoria**.

---

## 📁 Paquete `os` — Interactuando con el Sistema Operativo

### Archivos: Lectura y Escritura

**Lectura completa (archivos pequeños):**
```go
contenido, err := os.ReadFile("config.json")
// contenido es []byte con TODO el archivo
```

**Escritura completa (archivos pequeños):**
```go
datos := []byte("Hola mundo\n")
err := os.WriteFile("salida.txt", datos, 0644)
```

**Lectura streaming (archivos grandes):**
```go
archivo, err := os.Open("bigfile.dat")
defer archivo.Close()

scanner := bufio.NewScanner(archivo)
for scanner.Scan() {
    linea := scanner.Text()
    // procesar línea...
}
```

**Escritura streaming con buffer:**
```go
archivo, _ := os.Create("output.txt")
defer archivo.Close()

writer := bufio.NewWriter(archivo)
defer writer.Flush()

fmt.Fprintf(writer, "Línea %d\n", 1)
```

### Información de Archivos (`os.Stat`)

```go
info, err := os.Stat("archivo.txt")
// info.Name()    → "archivo.txt"
// info.Size()    → 1234 (bytes)
// info.ModTime() → time.Time
// info.IsDir()   → false
// info.Mode()    → -rw-r--r-- (permisos)
```

### Variables de Entorno

```go
home := os.Getenv("USERPROFILE")  // Leer una variable
todas := os.Environ()             // Todas las variables
```

### Información del Proceso

```go
pid := os.Getpid()    // Process ID actual
ppid := os.Getppid()  // Parent Process ID
host, _ := os.Hostname() // Nombre del equipo
dir, _ := os.Getwd()     // Directorio actual
```

---

## 🛤️ Paquete `filepath` — Rutas Multiplataforma

```go
filepath.Join("home", "user", "file.go")  // home/user/file.go (Linux)
                                            // home\user\file.go (Windows)
filepath.Ext("archivo.tar.gz")   // ".gz"
filepath.Base("a/b/c/file.go")   // "file.go"
filepath.Dir("a/b/c/file.go")   // "a/b/c"
filepath.Abs(".")                // /ruta/absoluta/actual
filepath.Clean("a//b/../b/./c") // "a/b/c"
filepath.Match("*.go", "main.go") // true
filepath.Glob("src/*.go")        // [src/main.go src/util.go]
```

### filepath.WalkDir — Recursión de Directorios

```go
filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
    if !d.IsDir() {
        fmt.Println(path) // es un archivo
    }
    return nil
})
```

Es más eficiente que `filepath.Walk` porque no hace `os.Stat` adicional para cada entrada.

---

## 🚩 Paquete `flag` — Argumentos de Línea de Comandos

```go
source := flag.String("source", "", "directorio origen")
dest := flag.String("dest", "", "directorio destino")
verbose := flag.Bool("verbose", false, "modo detallado")
flag.Parse()

// Uso: go run main.go -source ./src -dest ./backup -verbose
```

---

## 📝 Ejercicio Práctico: Herramienta CLI de Backup Incremental

### ¿Qué construimos?

Una herramienta de línea de comandos que:
1. **Escanea** recursivamente un directorio origen
2. **Compara** con el manifiesto del último backup
3. **Copia** solo archivos nuevos o modificados (incremental)
4. **Calcula checksums** SHA-256 para verificación de integridad
5. **Registra** logs detallados de cada operación

### Arquitectura del Código

```
📁 16-os-io/
├── go.mod
└── main.go          ← Todo el código en un solo archivo
```

### Estructuras de Datos

```go
type FileInfo struct {
    Path     string
    Size     int64
    ModTime  time.Time
    Checksum string
}

type BackupManifest struct {
    Files     map[string]FileInfo
    BackupAt  time.Time
    SourceDir string
}
```

El **manifiesto** es un archivo de texto que registra qué archivos existían en el último backup, con su tamaño, fecha de modificación y hash. Al comparar el estado actual con el manifiesto, sabemos exactamente qué archivos copiar.

### Ejecución

```bash
# 1. Crear datos de prueba
go run main.go -setup

# 2. Ejecutar primer backup (copia todo)
go run main.go -source test-data\origen -dest test-data\backup

# 3. Ejecutar segundo backup (solo copia cambios)
go run main.go -source test-data\origen -dest test-data\backup

# 4. Backup con verificación de checksum
go run main.go -source test-data\origen -dest test-data\backup -checksum

# 5. Ver demostraciones de os/io/filepath
go run main.go -demo
```

---

## 🔍 Conceptos Clave Explicados

### ¿Por qué `defer` es crucial?

```go
archivo, _ := os.Open("datos.txt")
defer archivo.Close() // Se ejecuta cuando la función termina
```

`defer` garantiza que el archivo se cierre **incluso si hay un error** entre la apertura y el cierre. Sin `defer`, podrías tener archivos abiertos que consumen recursos del sistema.

### ¿Por qué `io.Copy` en lugar de leer todo a memoria?

```go
// ❌ MALO para archivos grandes:
contenido, _ := os.ReadFile("video.mp4") // ¡4GB en memoria!

// ✅ BIEN para cualquier tamaño:
io.Copy(destino, origen) // Solo 32KB en memoria
```

### ¿Por qué `filepath.Join` en lugar de concatenar strings?

```go
// ❌ MALO:
ruta := dir + "/" + archivo  // Falla en Windows (\) vs Linux (/)

// ✅ BIEN:
ruta := filepath.Join(dir, archivo) // Usa el separador correcto
```

---

## 🏋️ Ejercicio Feynman

### Instrucciones
Usando la **Técnica Feynman**, explica estos conceptos **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado. Usa analogías de la vida cotidiana.

---

### Ejercicio 1: `io.Reader` e `io.Writer`
> Explica qué son `io.Reader` e `io.Writer` usando la analogía de una **cadena de producción** en una fábrica. ¿Qué representa cada rol?

---

### Ejercicio 2: `io.Copy`
> Imagina que `io.Copy` es una **manguera de agua**. ¿De dónde sale el agua (Reader) y a dónde va (Writer)? ¿Por qué es mejor usar la manguera que llenar cubetas (leer todo a memoria)?

---

### Ejercicio 3: `defer`
> Explica `defer` usando la analogía de un **mesero en un restaurante**. ¿Qué pasa si el mesero tiene que cerrar la puerta de la cocina al final de cada servicio?

---

### Ejercicio 4: `filepath.WalkDir`
> Si `filepath.WalkDir` fuera un **detective investigando una casa**, ¿cómo recorrería todas las habitaciones y archivos? ¿Qué le preguntaría a cada puerta?

---

### Ejercicio 5: Backup Incremental
> Explica la diferencia entre un backup **completo** y uno **incremental** usando la analogía de **fotocopiar un libro**. ¿Cuándo fotocopias todo el libro y cuándo solo las páginas que cambiaron?

---

### Ejercicio 6: SHA-256 Checksum
> Imagina que cada archivo tiene una **huella digital única**. ¿Por qué es útil verificar la huella antes de decidir si copiarlo? ¿Qué problemas evita?

---

## 📋 Resumen de Funciones Clave

| Función | Paquete | ¿Qué hace? |
|---------|---------|-------------|
| `os.Open(path)` | `os` | Abre archivo para lectura |
| `os.Create(path)` | `os` | Crea archivo para escritura |
| `os.ReadFile(path)` | `os` | Lee archivo completo a `[]byte` |
| `os.WriteFile(path, data, perm)` | `os` | Escribe `[]byte` a archivo |
| `os.Stat(path)` | `os` | Info del archivo sin abrirlo |
| `os.MkdirAll(path, perm)` | `os` | Crea directorio y padres |
| `os.Getenv(key)` | `os` | Lee variable de entorno |
| `os.Getpid()` | `os` | PID del proceso actual |
| `io.Copy(dst, src)` | `io` | Copia de Reader a Writer |
| `bufio.NewScanner(file)` | `bufio` | Lee línea por línea |
| `bufio.NewWriter(file)` | `bufio` | Escritura con buffer |
| `filepath.Join(parts...)` | `filepath` | Une rutas de forma segura |
| `filepath.WalkDir(root, fn)` | `filepath` | Recorre directorio recursivo |
| `filepath.Ext(path)` | `filepath` | Extrae extensión |
| `flag.String(name, default, help)` | `flag` | Define flag de CLI |

---

## 🔗 ¿Qué sigue?

En la siguiente lección exploraremos **archivos de configuración** (JSON, YAML, TOML), **logging estructurado** y cómo construir aplicaciones CLI más sofisticadas usando `cobra` o `urfave/cli`.

---

> 💡 *"El que domina los archivos y los flujos de datos, domina el sistema operativo."*