<div align="center">

# 📘 Lección 01

# Hello, Go! — Entorno, Sintaxis y el Pulso del Lenguaje

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Fase I](https://img.shields.io/badge/Fase_I-Fundamentos-4ECDC4?style=for-the-badge) ![Lección 01](https://img.shields.io/badge/Lecci%C3%B3n-01-FF6B6B?style=for-the-badge)

<br>

> *"Go es el lenguaje que diseñarían los ingenieros si empezaran de cero,
> sabiendo todo lo que saben hoy sobre lo que sale mal en los lenguajes existentes."*
> — **Rob Pike**, co-creador de Go

</div>

---

## 🎯 ¿Qué vas a dominar en esta lección?

| Concepto | Descripción |
|:---------|:------------|
| 🧬 **Filosofía de Go** | Por qué Go existe, qué problema resuelve y cuándo NO usarlo |
| 🛠️ **Configuración del entorno** | Instalación, `GOPATH`, `GOROOT`, `go mod init` |
| 📄 **Anatomía de un programa Go** | `package main`, `func main()`, imports, exports |
| 🏗️ **Compilación y ejecución** | `go run`, `go build`, binarios estáticos |
| 🔬 **Runtime de Go** | `runtime.GOOS`, `runtime.GOARCH`, `runtime.Version()` |
| 🧪 **Ejercicio práctico** | Detector de entorno de desarrollo |

---

## 🧬 1. ¿Por qué existe Go?

### La analogía de la navaja suiza

Imagina que eres un cirujano. Necesitas operar un tumor cerebral. Tu asistente te trae una **navaja suiza de 47 herramientas**: tiene cuchillo, sierra, destornillador, lupa, sacacorchos, lima de uñas...

¿La usarías para abrir un cráneo?

**Absolutamente no.** Necesitas un bisturí. Un instrumento diseñado para **exactamente** lo que necesitas, afilado quirúrgicamente, sin piezas que sobren.

**Go es ese bisturí.**

No es un lenguaje multi-paradigma con 47 características. No tiene clases, herencia, excepciones, genéricos (bueno, desde 1.18 tiene algo), pattern matching, macros, ni anotaciones. Fue diseñado con **exactamente** las herramientas que necesitas para construir software de sistemas moderno:

- ✅ **Compilación rapidísima** — segundos, no minutos
- ✅ **Un solo binario estático** — sin dependencias externas
- ✅ **Concurrencia integrada** — goroutines y channels
- ✅ **Tipado estático** — errores en compilación, no en producción a las 3 AM
- ✅ **Garbage collector eficiente** — sin gestión manual de memoria
- ✅ **Formato automático** — `gofmt` elimina los debates de estilo

### 🏭 El contexto histórico: el dolor de Google (2007)

En 2007, tres leyendas de la ingeniería — **Robert Griesemer**, **Rob Pike** y **Ken Thompson** (el creador de UNIX, nada menos) — estaban frustrados. Trabajaban en Google y enfrentaban tres dolores crónicos:

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│   😤 DOLOR 1: Compilaciones lentas                                  │
│      Los proyectos de C++ en Google tardaban MINUTOS en compilar.   │
│      Los desarrolladores perdían horas al día esperando.            │
│                                                                     │
│   😤 DOLOR 2: Concurrencia frágil                                   │
│      Los threads de C++ eran pesados, propensos a deadlocks,        │
│      y escribir código concurrente correcto era un arte negro.      │
│                                                                     │
│   😤 DOLOR 3: Dependencias complejas                                │
│      Los sistemas crecían en complejidad con dependencias           │
│      circulares, headers gigantes, y builds que requerían           │
│      configurar 17 variables de entorno.                            │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

Entonces diseñaron Go desde cero. No como un experimento académico, sino como una **herramienta de ingeniería práctica** para resolver **exactamente** esos tres problemas.

### 🎯 ¿Qué problema resuelve Go?

Go es el lenguaje para cuando necesitas:

| Escenario | Por qué Go brilla |
|:----------|:-------------------|
| 🌐 **Servidores web y APIs REST** | Concurrencia masiva con goroutines ligeras |
| 🐳 **Herramientas CLI** | Un solo binario, sin dependencias, fácil de distribuir |
| 🔧 **Infraestructura y DevOps** | Docker, Kubernetes, Terraform están escritos en Go |
| 📡 **Microservicios** | Compilación rápida, deploy fácil, bajo consumo de memoria |
| 🗄️ **Sistemas de bases de datos** | Rendimiento cercano a C con seguridad de memoria |
| 📦 **Servicios en la nube** | Go es el lenguaje nativo de la nube (CNCF) |

### ⚠️ ¿Cuándo NO deberías usar Go?

Aquí viene lo que separa al que entiende de Go del fanático. Go **NO** es ideal para:

| Escenario | Por qué Go no es la mejor opción |
|:----------|:----------------------------------|
| 🧮 **Computación científica / ML** | Python + NumPy/PyTorch es más productivo para prototipar |
| 🎮 **Desarrollo de videojuegos** | C++/Rust ofrecen control de memoria sin GC pauses |
| 🖥️ **Apps de escritorio con GUI** | No tiene un framework GUI nativo maduro (usarías Flutter, Qt, etc.) |
| 📱 **Desarrollo móvil** | Aunque tiene bindings, Kotlin/Swift son más productivos |
| ⚡ **Sistemas en tiempo real duro** | El garbage collector puede causar pauses (aunque mínimas) |
| 🎭 **Metaprogramación intensiva** | No tiene macros, reflection es lento, y genéricos son limitados |

> 🧠 **Regla de oro:** Si necesitas construir un **sistema backend, una herramienta CLI, o infraestructura de red** → Go es probablemente tu mejor opción. Si necesitas **prototipar rápido con ML o crear una app móvil** → hay mejores herramientas.

---

## 🛠️ 2. Configuración del Entorno

### 📥 Instalación de Go

Antes de escribir una sola línea, necesitas tener Go instalado. Ve a [go.dev/dl](https://go.dev/dl/) y descarga el instalador para tu sistema operativo.

#### Verificación de la instalación

```bash
go version
# Salida esperada: go version go1.22.0 windows/amd64
```

```bash
go env
# Muestra TODAS las variables de entorno de Go
```

### 🗺️ Las variables de entorno clave

Piensa en las variables de entorno de Go como los **controles de un panel de instrumentos** de un avión. Necesitas entender cada una para volar:

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  📍 GOROOT  → ¿Dónde está instalado el compilador Go?              │
│               Normalmente: /usr/local/go (Linux/Mac)                │
│               Normalmente: C:\Go (Windows)                          │
│               ⚠️  NO lo cambies a menos que sepas exactamente       │
│               lo que estás haciendo.                                │
│                                                                     │
│  📂 GOPATH  → ¿Dónde vive tu código y dependencias?                │
│               Por defecto: ~/go                                     │
│               Contiene: src/, bin/, pkg/                            │
│               Con Go Modules, es menos crítico que antes.           │
│                                                                     │
│  🔧 GOBIN   → ¿Dónde instalar los ejecutables compilados?          │
│               Por defecto: $GOPATH/bin                              │
│               Agrega esto a tu PATH para ejecutarlos.               │
│                                                                     │
│  🖥️ GOOS    → Sistema operativo destino (linux, darwin, windows)   │
│  🏗️ GOARCH  → Arquitectura destino (amd64, arm64, 386)             │
│               ¡Puedes compilar para OTRO SO desde tu máquina!       │
│                                                                     │
│  📦 GOMOD   → Ruta al archivo go.mod activo                        │
│               Si estás en un módulo, apunta a tu go.mod             │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 📦 Go Modules: el sistema de dependencias moderno

Antes de Go 1.11, las dependencias se gestionaban con `GOPATH` y era un dolor de cabeza. Hoy usamos **Go Modules**, que es como el `package.json` de Node o el `requirements.txt` de Python, pero mejor:

```bash
# Inicializar un nuevo módulo (ejecutar dentro de tu directorio de proyecto)
go mod init nombre-del-modo

# Esto crea un archivo go.mod:
# module nombre-del-modo
#
# go 1.22

# Agregar una dependencia externa
go get github.com/sirupsen/logrus

# Limpiar dependencias no usadas
go mod tidy

# Descargar todas las dependencias al directorio local
go mod download
```

**La analogía del go.mod:** Imagina que tu proyecto es una receta de cocina. El `go.mod` es la lista de ingredientes con marcas específicas y versiones exactas. Si alguien más quiere cocinar tu receta, le entregas la lista y obtiene los **mismos** ingredientes exactos. Sin sorpresas.

---

## 📄 3. Anatomía de un Programa Go

Veamos la estructura mínima de un programa Go. Cada línea tiene un propósito:

```go
package main    // ← 1. Declaración del paquete

import (        // ← 2. Importación de paquetes
    "fmt"       //    fmt = "format" — entrada/salida formateada
    "os"        //    os = "operating system" — interacción con el SO
    "runtime"   //    runtime = información del runtime de Go
)

func main() {   // ← 3. Punto de entrada del programa
    fmt.Println("¡Hola, Go!")  // ← 4. Imprime en consola
}
```

### 🔍 Desglose línea por línea

#### Línea 1: `package main`

```go
package main
```

En Go, **cada archivo** pertenece a un paquete. Los paquetes son la unidad de organización del código.

- `package main` es **especial**: le dice al compilador que este paquete es un **ejecutable**, no una librería.
- Solo el paquete `main` puede tener una función `main()`.
- Si tu paquete se llama cualquier otra cosa (`package utils`, `package server`), es una **librería** que otros paquetes pueden importar.

> **Analogía:** Piensa en `package main` como la etiqueta "PRODUCTO FINAL" en una fábrica. Los otros paquetes son "COMPONENTES" que se usan para construir el producto final.

#### Línea 2-6: Imports

```go
import (
    "fmt"
    "os"
    "runtime"
)
```

Los imports traen funcionalidad de otros paquetes al código actual. Go tiene una librería estándar riquísima — no necesitas instalar nada para la mayoría de tareas.

| Paquete | ¿Qué hace? | Ejemplo de uso |
|:--------|:------------|:---------------|
| `fmt` | Formateo de strings e I/O | `fmt.Println()`, `fmt.Printf()` |
| `os` | Acceso al sistema operativo | Variables de entorno, archivos, argumentos CLI |
| `runtime` | Información del runtime de Go | OS, arquitectura, CPUs, versión de Go |
| `math` | Funciones matemáticas | `math.Sqrt()`, `math.Pi` |
| `strings` | Manipulación de strings | `strings.ToUpper()`, `strings.Split()` |
| `strconv` | Conversión de tipos | `strconv.Atoi()`, `strconv.Itoa()` |

> ⚠️ **Regla estricta de Go:** Si importas un paquete y NO lo usas, el compilador **rechaza** el programa. Esto es una decisión deliberada — elimina imports muertos que inflan el binario y confunden a los lectores.

#### Línea 8: `func main()`

```go
func main() {
```

- `func` es la palabra reservada para declarar funciones.
- `main()` es **la función especial** que Go ejecuta cuando corres el programa.
- No recibe argumentos y no devuelve nada.
- Solo puede existir en `package main`.

#### Línea 9: `fmt.Println("¡Hola, Go!")`

```go
fmt.Println("¡Hola, Go!")
```

- `fmt` es el paquete que importamos.
- `.Println` es la función (con mayúscula = exportada = pública).
- Los strings van entre comillas dobles `"..."` (no simples — Go no tiene char literals con comillas simples en strings).

### 🔤 Convenciones de nombres: el sistema de visibilidad de Go

Go tiene una convención de nombres **elegante y única**:

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│   Empieza con MAYÚSCULA → Exportado (público)           │
│   Ejemplo: Println, Getenv, Version                     │
│   Visible desde CUALQUIER otro paquete.                 │
│                                                         │
│   Empieza con minúscula → No exportado (privado)        │
│   Ejemplo: println, getenv, version                     │
│   Visible SOLO dentro del mismo paquete.                │
│                                                         │
│   NO hay keywords "public", "private", "protected".     │
│   Solo capitalización. Así de simple.                   │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

> **Analogía:** Imagina que cada paquete es un taller de carpintería. Las herramientas pintadas de **ROJO** (mayúsculas) pueden prestarse a otros talleres. Las pintadas de **GRIS** (minúsculas) son solo para uso interno. No necesitas un sistema de permisos complejo — solo miras el color.

---

## 🏗️ 4. Compilación y Ejecución

### `go run` vs `go build`

Estos son los dos comandos que usarás cientos de veces:

```bash
# go run: compila Y ejecuta en un solo paso
# Perfecto para desarrollo — rápido, directo
go run main.go
# Compila temporalmente, ejecuta, y luego borra el binario temporal

# go build: SOLO compila — genera un binario permanente
# Perfecto para producción — genera el ejecutable
go build -o mi-app main.go
# Crea un archivo "mi-app" (o "mi-app.exe" en Windows)
```

**La analogía del microondas vs el horno:**
- `go run` es como usar el **microondas**: rápido, calienta y comes, pero no queda nada después.
- `go build` es como usar el **horno**: tarda un poco más, pero produces un plato que puedes guardar, enviar a alguien, o servir en un restaurante (producción).

### 🔑 La magia del binario estático

Esta es una de las superpoderes de Go. Cuando haces `go build`:

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  Python:  📄 main.py → Necesita: Python instalado + módulos pip    │
│  Java:    📄 Main.class → Necesita: JVM instalada + classpath      │
│  Node.js: 📄 index.js → Necesita: Node instalado + node_modules    │
│                                                                     │
│  Go:      📄 main.go → → → 📦 main.exe (TODO incluido)            │
│           Solo necesitas EL BINARIO. Nada más.                      │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

El binario de Go incluye:
- ✅ Tu código
- ✅ Las dependencias compiladas
- ✅ El runtime de Go (incluido el garbage collector)
- ✅ Todo lo necesario para ejecutarse

**¿Por qué importa esto?** Porque es la razón por la que Docker, Kubernetes y Terraform son tan fáciles de distribuir. Copias **un solo archivo** a un servidor y funciona. Sin instalar Node, sin pip install, sin configurar Java_HOME. Solo copias y ejecutas.

### 🌍 Cross-compilation: compilar para cualquier plataforma

Go puede compilar binarios para **cualquier combinación de SO y arquitectura** desde tu propia máquina:

```bash
# Compilar para Linux desde Windows/Mac
GOOS=linux GOARCH=amd64 go build -o mi-app-linux main.go

# Compilar para macOS ARM (Apple Silicon) desde Linux
GOOS=darwin GOARCH=arm64 go build -o mi-app-mac main.go

# Compilar para Windows desde cualquier otro OS
GOOS=windows GOARCH=amd64 go build -o mi-app.exe main.go

# Compilar para Raspberry Pi (ARM)
GOOS=linux GOARCH=arm go build -o mi-app-pi main.go
```

> Esto es revolucionario para DevOps: puedes compilar el binario en tu laptop y deployarlo a cualquier servidor sin cruzar los dedos.

---

## 🔬 5. El Runtime de Go

El **runtime** es el motor que corre dentro de tu binario compilado. A diferencia de Java (que necesita la JVM) o Python (que necesita el intérprete), el runtime de Go está **embebido** dentro de cada binario.

### Funciones clave del runtime

```go
package main

import (
    "fmt"
    "runtime"
)

func main() {
    // ¿En qué SO estoy corriendo?
    fmt.Println(runtime.GOOS)      // "linux", "darwin", "windows"

    // ¿En qué arquitectura?
    fmt.Println(runtime.GOARCH)    // "amd64", "arm64", "386"

    // ¿Qué versión de Go compiló esto?
    fmt.Println(runtime.Version()) // "go1.22.0"

    // ¿Cuántas CPUs tengo?
    fmt.Println(runtime.NumCPU())  // 8, 12, 16...

    // ¿Dónde está instalado Go?
    fmt.Println(runtime.GOROOT())  // "/usr/local/go"

    // ¿Cuántas goroutines están activas? (lo veremos en Fase III)
    fmt.Println(runtime.NumGoroutine()) // 1 normalmente
}
```

### `fmt` — Tu herramienta de comunicación

El paquete `fmt` es el que usarás para imprimir en consola. Tiene tres funciones estrella:

```go
// Println: imprime y agrega un salto de línea al final
fmt.Println("Hola", "Mundo")    // "Hola Mundo\n"

// Printf: imprime con formato (como C, pero más seguro)
fmt.Printf("Soy %s y tengo %d años\n", "Go", 15)
// "Soy Go y tengo 15 años"

// Sprintf: NO imprime — devuelve el string formateado
mensaje := fmt.Sprintf("Hola %s", "Mundo")
// mensaje ahora es "Hola Mundo" (no se imprime)
```

### Verbos de formato comunes en `Printf`

| Verbo | Tipo | Ejemplo | Salida |
|:------|:-----|:--------|:-------|
| `%s` | string | `Printf("%s", "Go")` | `Go` |
| `%d` | entero | `Printf("%d", 42)` | `42` |
| `%f` | float | `Printf("%f", 3.14)` | `3.140000` |
| `%t` | boolean | `Printf("%t", true)` | `true` |
| `%v` | cualquier valor | `Printf("%v", x)` | valor por defecto |
| `%T` | tipo del valor | `Printf("%T", 42)` | `int` |
| `%q` | string con comillas | `Printf("%q", "hola")` | `"hola"` |
| `%x` | hexadecimal | `Printf("%x", 255)` | `ff` |
| `%b` | binario | `Printf("%b", 10)` | `1010` |
| `%#v` | sintaxis Go | `Printf("%#v", s)` | representación Go |

---

## 🏋️ 6. Ejercicio Práctico: Detector de Entorno de Desarrollo

Ahora vamos a construir algo **útil**. Antes de construir una casa necesitas verificar que tus cimientos están bien. Este CLI será tu "panel de instrumentos" del laboratorio — lo usarás para diagnosticar problemas a lo largo de todo el curso.

### 📁 Estructura del proyecto

```
01-hello-go/
├── main.go          ← Tu código fuente
└── go.mod           ← La "ficha de identidad" del módulo
```

### 📄 Archivo `go.mod`

```
module 01-hello-go

go 1.22
```

> Este archivo le dice a Go: "Este proyecto se llama `01-hello-go` y usa Go versión 1.22". Es como la cédula de identidad de tu proyecto.

### 📄 Archivo `main.go` — Explicación línea por línea

```go
package main
```

Declara que este archivo pertenece al paquete `main`. Como ya sabemos, esto significa que es un **ejecutable**, no una librería.

---

```go
import (
    "fmt"
    "os"
    "runtime"
)
```

Importamos tres paquetes de la librería estándar:

- **`fmt`** → Para imprimir en consola con formato bonito (`Printf`, `Println`)
- **`os`** → Para acceder al sistema operativo (variables de entorno, directorio actual, PID)
- **`runtime`** → Para obtener información sobre el runtime de Go (OS, arquitectura, CPUs, versión)

---

```go
func main() {
```

La función `main()` es el punto de entrada. Go ejecuta esta función al iniciar el programa. Sin `main()`, no hay programa.

---

```go
fmt.Println("╔══════════════════════════════════════════════╗")
fmt.Println("║   🧪 Detector de Entorno — Laboratorio Go   ║")
fmt.Println("╚══════════════════════════════════════════════╝")
```

Creamos una cabecera visual usando caracteres Unicode (`╔`, `═`, `║`, `╚`). Esto es un panel de instrumentos — queremos que se vea profesional y claro desde el primer momento.

---

```go
osName := runtime.GOOS
```

`runtime.GOOS` devuelve una **constante string** con el nombre del sistema operativo: `"linux"`, `"darwin"` (macOS), `"windows"`, `"freebsd"`, etc.

El operador `:=` es la **declaración corta de variable** de Go. Es equivalente a:
```go
var osName string = runtime.GOOS
```
Pero más conciso. Go infiere el tipo automáticamente — aquí infiere que es `string` porque `runtime.GOOS` devuelve un `string`.

---

```go
arch := runtime.GOARCH
```

`runtime.GOARCH` devuelve la arquitectura del procesador: `"amd64"` (Intel/AMD 64-bit), `"arm64"` (Apple Silicon, Raspberry Pi 4+), `"386"` (32-bit x86), `"arm"` (ARM 32-bit).

---

```go
goVersion := runtime.Version()
```

`runtime.Version()` devuelve la versión exacta de Go con la que se compiló el binario. Esto es útil para debuggear: si alguien ejecuta tu binario con problemas, puedes preguntarle "¿qué dice `goVersion`?" y saber exactamente qué compilador se usó.

---

```go
numCPU := runtime.NumCPU()
```

`runtime.NumCPU()` devuelve el número de CPUs **lógicas** disponibles. Esto es importante porque en la Fase III (Concurrencia), `GOMAXPROCS` (el número máximo de CPUs que Go puede usar simultáneamente) se basa en este valor. Tu laptop con 8 cores puede ejecutar 8 goroutines verdaderamente en paralelo.

---

```go
gopath := os.Getenv("GOPATH")
goroot := os.Getenv("GOROOT")
```

`os.Getenv()` lee una variable de entorno del sistema operativo. Las variables de entorno son como "papelitos pegados en la pared" que el sistema y las aplicaciones pueden leer.

- **`GOPATH`** → Directorio donde Go almacena código fuente y dependencias descargadas
- **`GOROOT`** → Directorio donde está instalado el compilador Go

---

```go
if gopath == "" {
    gopath = "(valor por defecto del sistema — ~/.go o ~/go)"
}
```

Si la variable no está definida, asignamos un valor informativo. En Go, el **zero value** de un `string` es `""` (cadena vacía). Esta es una de las decisiones de diseño más elegantes de Go: **toda variable tiene un valor predeterminado válido**, nunca es `null`/`nil`/`undefined`.

---

```go
wd, err := os.Getwd()
if err != nil {
    wd = "(no se pudo determinar)"
}
```

Aquí vemos el **patrón `(resultado, error)`** — el patrón más importante de Go. `os.Getwd()` devuelve **dos valores**:
1. El directorio actual (`string`)
2. Un posible error (`error`)

En Go, los errores son **valores**, no excepciones. No hay `try/catch`. Si algo puede fallar, la función devuelve un error como segundo valor. **Siempre verificas `err != nil`** — es el equivalente a "¿salió algo mal?".

> 🧠 Este patrón lo verás **millones** de veces en tu carrera Go. Es como respirar: natural, constante, esencial.

---

```go
pid := os.Getpid()
```

Devuelve el **Process ID** del programa en ejecución. Útil para logging y diagnóstico. Cada vez que ejecutes el programa, este número será diferente (asignado por el OS).

---

```go
fmt.Printf("  💻 Sistema Operativo : %s\n", osName)
fmt.Printf("  🏗️  Arquitectura     : %s\n", arch)
fmt.Printf("  🔧 Versión de Go     : %s\n", goVersion)
fmt.Printf("  🧠 CPUs disponibles  : %d\n", numCPU)
fmt.Printf("  📂 GOPATH            : %s\n", gopath)
fmt.Printf("  📂 GOROOT            : %s\n", goroot)
fmt.Printf("  📂 Directorio actual : %s\n", wd)
fmt.Printf("  🆔 PID del proceso   : %d\n", pid)
```

`fmt.Printf` imprime con formato. Usamos `%s` para strings, `%d` para enteros, y `\n` para saltos de línea. Observa que cada línea tiene emojis para identificar visualmente cada métrica — esta es una práctica de UX en CLI que hace los reportes fáciles de escanear.

---

```go
fmt.Println("  ✅ Entorno listo. ¡Tu laboratorio Go está operativo!")
```

El mensaje de confirmación. Si llegaste hasta aquí sin errores, tu entorno está listo para el curso.

### ▶️ Ejecutar el programa

```bash
# Dentro del directorio 01-hello-go/
go run main.go
```

**Salida esperada (tu resultado será diferente según tu sistema):**

```
╔══════════════════════════════════════════════╗
║   🧪 Detector de Entorno — Laboratorio Go   ║
╚══════════════════════════════════════════════╝

  💻 Sistema Operativo : windows
  🏗️  Arquitectura     : amd64
  🔧 Versión de Go     : go1.22.0
  🧠 CPUs disponibles  : 12
  📂 GOPATH            : C:\Users\tu-usuario\go
  📂 GOROOT            : C:\Program Files\Go
  📂 Directorio actual : C:\Cline\GO\01-hello-go
  🆔 PID del proceso   : 12345

  ✅ Entorno listo. ¡Tu laboratorio Go está operativo!
```

### 🔨 Compilar el binario permanente

```bash
go build -o env-detector main.go
# Esto crea un archivo ejecutable "env-detector" (o "env-detector.exe" en Windows)

# Ejecutarlo directamente
./env-detector
# En Windows: env-detector.exe
```

> 💡 **Experimento:** Copia el binario a otra máquina con el mismo OS y arquitectura. ¿Funciona? **SÍ.** Sin instalar Go, sin configurar nada. Solo copias y ejecutas. Esta es la magia de los binarios estáticos de Go.

---

## 🧠 7. Conceptos Clave Resumidos

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  📄 package main                                                    │
│     → Paquete ejecutable (el único que puede tener func main())     │
│                                                                     │
│  📥 import                                                          │
│     → Trae funcionalidad de otros paquetes                          │
│     → Si importas algo que no usas → ERROR DE COMPILACIÓN           │
│                                                                     │
│  🚀 func main()                                                     │
│     → Punto de entrada del programa                                 │
│                                                                     │
│  🏷️ := (declaración corta)                                          │
│     → Declara variable e infiere el tipo automáticamente            │
│                                                                     │
│  ❌ (resultado, error)                                               │
│     → El patrón fundamental de Go para manejo de errores            │
│     → Los errores son VALORES, no excepciones                       │
│                                                                     │
│  📦 go mod init                                                     │
│     → Crea un módulo Go (tu "ficha de identidad" del proyecto)      │
│                                                                     │
│  🏗️ go build                                                        │
│     → Compila a un solo binario estático sin dependencias           │
│                                                                     │
│  ▶️ go run                                                          │
│     → Compila y ejecuta en un solo paso                             │
│                                                                     │
│  🔤 Mayúscula = exportado, minúscula = privado                      │
│     → No hay keywords public/private — solo capitalización           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 📊 8. Comparación: Go vs el Mundo

| Característica | Go | Python | JavaScript | Java |
|:---------------|:---|:-------|:-----------|:-----|
| **Tipado** | Estático con inferencia | Dinámico | Dinámico | Estático |
| **Compilación** | A binario nativo | Interpretado | JIT (V8) | A bytecode (JVM) |
| **Concurrencia** | Goroutines (miles) | GIL limita | Event loop | Threads (pesados) |
| **Dependencias** | go.mod (integrado) | pip + venv | npm + node_modules | Maven/Gradle |
| **Tiempo de compilación** | Segundos | N/A | N/A | Minutos (proyectos grandes) |
| **Distribución** | Un solo binario | Necesita Python | Necesita Node | Necesita JVM |
| **Errores** | Valores (no excepciones) | Excepciones | Excepciones | Excepciones |
| **Formato** | gofmt (automático) | black/opcional | prettier/opcional | IDEs |
| **Curva de aprendizaje** | 🟢 Baja | 🟢 Baja | 🟡 Media | 🔴 Alta |

---

## 🧩 Ejercicio Feynman

> **El reto final de esta lección:** Imagina que un amigo que trabaja en marketing te pregunta: *"¿Para qué necesito otro lenguaje si ya existen Python y JavaScript?"*

### Tu misión: explícale en 5 oraciones

Usa tus propias palabras. Sin jerga técnica innecesaria. Como si le hablaras a alguien inteligente pero sin contexto de programación.

### Las 5 preguntas que debes responder:

1. **¿Para qué fue creado Go?** — ¿Qué dolor específico quería resolver?

2. **¿Qué problema resuelve que Python/JavaScript no resuelven bien?** — Piensa en compilación, distribución y concurrencia.

3. **¿Qué es un binario estático y por qué importa?** — ¿Por qué es más fácil distribuir un programa en Go que uno en Python?

4. **¿Cuándo NO deberías usar Go?** — Si solo ves ventajas, no has pensado lo suficiente.

5. **¿Por qué Go tiene menos características que otros lenguajes y eso es una VENTAJA?** — Usa la analogía de la navaja suiza vs la multi-herramienta de 47 piezas.

### 📝 Autoevaluación

| Criterio | ✅ | ❌ |
|:---------|:---|:---|
| Puedes explicar qué es Go sin mencionar la palabra "compilado" | | |
| Puedes nombrar al menos 3 proyectos famosos escritos en Go | | |
| Puedes explicar qué hace `go run` vs `go build` | | |
| Puedes explicar el patrón `(resultado, error)` con tus palabras | | |
| Puedes explicar por qué `package main` es especial | | |
| Puedes explicar la regla de mayúsculas/minúsculas de Go | | |
| Puedes nombrar 2 situaciones donde NO usarías Go | | |
| Puedes explicar qué es un binario estático a un no-programador | | |

> 🎯 **Si marcaste algún ❌:** Vuelve a la sección correspondiente y léela de nuevo. El Método Feynman funciona así: descubrir que no puedes explicar algo es **exactamente** el momento donde el aprendizaje real comienza.

---

## 🔮 ¿Qué viene en la Lección 02?

> **Variables, Tipos y el Sistema de Tipos de Go** — Dominarás `int`, `float64`, `string`, `bool`, entenderás por qué Go no hace casting automático, y construirás un conversor de unidades del mundo real. Será la lección donde Go deja de ser "un lenguaje nuevo" y empieza a ser "tu lenguaje".

---

<div align="center">

### *"El código se lee mucho más de lo que se escribe. Siendo así, la legibilidad importa mucho."*
### — **Rob Pike**, co-creador de Go

<br>

**¡Lección 01 completada! Avanza a la Lección 02 🚀**

</div>