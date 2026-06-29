# 📦 Lección 14 — Paquetes, Módulos y la Organización del Código Go

> **Método Feynman:** Si no puedes explicarlo de forma simple, no lo entiendes bien.
> **Objetivo:** Dominar la arquitectura de paquetes de Go, crear librerías reutilizables con visibilidad controlada, y gestionar dependencias con Go Modules.

---

## 🧩 ¿Qué vamos a aprender?

| Concepto | Analogía Feynman |
|----------|------------------|
| **Paquete** | Una carpeta de herramientas con etiquetas: rojo (público) o gris (privado) |
| **Módulo** | El inventario central que dice qué herramientas tienes y de dónde vinieron |
| **Visibilidad** | Si la etiqueta empieza con mayúscula → exportado. Minúscula → privado |
| **go.mod** | La cédula de identidad de tu proyecto: nombre, versión de Go y dependencias |
| **go.sum** | Los sellos de seguridad que verifican que nadie manipuló tus dependencias |
| **godoc** | El manual de instrucciones que se escribe como comentarios en el código |

**Prerrequisito:** Lecciones 01-13 (fundamentos + tipos + concurrencia)

---

## 📚 Marco Teórico: ¿Por qué necesitamos paquetes?

### El Caos Sin Paquetes

Imagina que trabajas en un proyecto con 500 archivos Go, todos en la misma carpeta:

```
❌ Sin paquetes:
├── main.go
├── user.go
├── auth.go
├── database.go
├── email.go
├── logging.go
├── ... (494 archivos más)
│
├── 😱 10,000 líneas en una sola carpeta
├── 😱 No sabes qué función es pública o privada
├── 😱 Dos archivos definen la misma función "procesar()"
├── 😱 Copias y pegas código entre proyectos
├── 😱 No hay forma de reutilizar nada
```

```
✅ Con paquetes:
├── main.go                  ← Punto de entrada
├── go.mod                   ← Identidad del módulo
├── auth/
│   ├── login.go             ← Lógica de autenticación
│   └── token.go             ← Generación de tokens
├── database/
│   ├── connection.go        ← Pool de conexiones
│   └── queries.go           ← Consultas SQL
├── email/
│   └── sender.go            ← Envío de emails
└── utils/
    ├── math.go              ← Funciones matemáticas
    └── text.go              ← Procesamiento de texto
│
├── ✅ Cada carpeta = un paquete con responsabilidad clara
├── ✅ Minúscula = privado, Mayúscula = exportado
├── ✅ Se reutiliza importando: "mi-proyecto/auth"
└── ✅ El compilador previene ciclos de importación
```

**Los paquetes son la unidad fundamental de organización en Go. Sin ellos, el código es un basurero. Con ellos, es una biblioteca organizada.**

---

## 🏗️ Concepto 1: ¿Qué es un Paquete?

### La Analogía de la Carpeta de Herramientas

Un paquete de Go es como una **carpeta de herramientas** en un taller profesional:

```
┌─────────────────────────────────────────────┐
│           🔧 Paquete "mathutil"             │
│                                             │
│  🔴 Suma()          → Puedes pedir prestada │
│  🔴 Promedio()      → Puedes pedir prestada │
│  🔴 Maximo()        → Puedes pedir prestada │
│  🔴 Factorial()     → Puedes pedir prestada │
│                                             │
│  ⬜ promedioInterno() → Solo uso interno    │
│  ⬜ validar()         → Solo uso interno    │
│  ⬜ redondear()       → Solo uso interno    │
└─────────────────────────────────────────────┘

  🔴 = Mayúscula = EXPORTADO (público, visible desde afuera)
  ⬜ = Minúscula = PRIVADO (solo visible dentro del paquete)
```

### Reglas de Visibilidad

En Go **no existen** `public`, `private`, `protected`, `internal`. Solo hay una regla:

```
╔══════════════════════════════════════════════════════════╗
║  REGLA DE ORO DE LA VISIBILIDAD EN GO                    ║
║                                                          ║
║  Si el nombre empieza con MAYÚSCULA → es EXPORTADO       ║
║    → Visible desde cualquier paquete que lo importe       ║
║                                                          ║
║  Si el nombre empieza con MINÚSCULA → es PRIVADO          ║
║    → Solo visible dentro del MISMO paquete                ║
╚══════════════════════════════════════════════════════════╝
```

Aplica a **todo**: funciones, variables, constantes, tipos, campos de structs, métodos.

```go
// Visible desde otros paquetes
func Promedio(nums []float64) (float64, error) { ... }
var ErrEmptySlice = errors.New("...")
type Calculadora struct { ... }

// Solo visible dentro de este paquete
func calcularPromedio(nums []float64) float64 { ... }
var errEmptySlice = errors.New("...")
type calculadoraInterna struct { ... }
```

### ¿Por qué la capitalización en vez de keywords?

| Enfoque | Lenguajes | Pros | Contras |
|---------|-----------|------|---------|
| Keywords (`public/private`) | Java, C#, C++ | Explícito | Más verboso, más keywords |
| Capitalización | **Go** | Minimalista, una sola regla | Confuso al principio |
| Prefijos (`_`, `__`) | Python, C | Flexible | Feo, inconsistente |

Go eligió la capitalización porque:
1. **Zero keywords extra** → el lenguaje es más simple
2. **Visible en cualquier lugar** → no necesitas buscar el `public` arriba
3. **Consistente** → funciona igual para funciones, tipos, campos y métodos
4. **Forzado por el compilador** → no es una convención, es la ley

---

## 🏗️ Concepto 2: Estructura de un Paquete

### Declaración del Paquete

Cada archivo `.go` debe empezar con un `package` declaration:

```go
// archivo: myutils/mathutil/mathutil.go

package mathutil  // ← TODOS los archivos en esta carpeta deben usar "mathutil"

import (
    "errors"
    "math"
)

// Las funciones exportadas empiezan con mayúscula
func Promedio(nums []float64) (float64, error) {
    if len(nums) == 0 {
        return 0, ErrEmptySlice
    }
    return Suma(nums) / float64(len(nums)), nil
}

// Las privadas empiezan con minúscula
func redondear(v float64) float64 {
    return math.Round(v*100) / 100
}
```

### Reglas del Package Declaration

```
╔══════════════════════════════════════════════════════════╗
║  REGLAS DEL PACKAGE                                      ║
║                                                          ║
║  1. TODOS los archivos en la misma carpeta deben tener   ║
║     el MISMO nombre de paquete                           ║
║                                                          ║
║  2. Solo UN archivo puede tener "package main"           ║
║     → Es el punto de entrada (función main)              ║
║                                                          ║
║  3. Los archivos _test.go se excluyen del build normal   ║
║     → Solo se compilan con "go test"                     ║
║                                                          ║
║  4. No hay "package.json", "setup.py" ni "pom.xml"       ║
║     → Solo la declaración "package" al inicio            ║
╚══════════════════════════════════════════════════════════╝
```

### ¿Por Qué Un Solo Paquete por Carpeta?

Go lo hace así por diseño:

```
mathutil/
├── mathutil.go      ← package mathutil  ✅
├── estadistica.go   ← package mathutil  ✅ (mismo paquete)
├── algebra.go       ← package mathutil  ✅ (mismo paquete)
└── test.go          ← package mathutil  ✅ (mismo paquete)

❌ NO puedes hacer:
mathutil/
├── mathutil.go      ← package mathutil
└── otro.go          ← package otro     ← ERROR: un solo paquete por carpeta
```

**¿Por qué?** Porque la carpeta ES el paquete. Si pudieras mezclar paquetes, la visibilidad (mayúscula/minúscula) perdería sentido.

---

## 📦 Concepto 3: Go Modules — La Cédula de Identidad

### ¿Qué es un Módulo?

Un módulo es un **conjunto de paquetes** agrupados bajo un nombre de ruta único, con un archivo `go.mod` como identidad:

```
mi-proyecto/           ← ESTE es el módulo raíz
├── go.mod             ← Su cédula de identidad
├── go.sum             ← Sus sellos de verificación
├── main.go            ← package main (punto de entrada)
├── auth/              ← package auth (subpaquete)
│   └── login.go
└── utils/             ← package utils (subpaquete)
    └── math.go
```

### El Archivo go.mod

```go
module mi-proyecto          // ← Nombre único del módulo (ruta de importación)

go 1.21                     // ← Versión mínima de Go requerida

require (                   // ← Dependencias externas
    github.com/gin-gonic/gin v1.9.1
    github.com/lib/pq v1.10.9
)
```

**Analogía:** `go.mod` es como la cédula de identidad de tu proyecto:
- **module** → Tu nombre completo (ruta de importación)
- **go** → Tu fecha de nacimiento (versión de Go)
- **require** → Tu lista de contactos (dependencias)

### El Archivo go.sum

```
github.com/gin-gonic/gin v1.9.1 h1:abc123def456...
github.com/gin-gonic/gin v1.9.1/go.mod h1:xyz789...
```

**¿Qué es?** Es el "sello de seguridad" de cada dependencia. Contiene hashes SHA-256 que verifican que nadie manipuló el código descargado.

**Analogía:** Si `go.mod` es la lista de ingredientes de una receta, `go.sum` es el código de barras de cada ingrediente que verifica que no está adulterado.

### Comandos Esenciales de Modules

```bash
# 1. Crear un nuevo módulo (solo una vez)
go mod init mi-proyecto

# 2. Agregar una dependencia (la descarga y actualiza go.mod)
go get github.com/gin-gonic/gin@v1.9.1

# 3. Limpiar dependencias no usadas y agregar las faltantes
go mod tidy

# 4. Descargar todas las dependencias al cache local
go mod download

# 5. Ver el grafo de dependencias
go mod graph

# 6. Verificar que go.sum coincida con go.mod
go mod verify
```

### Flujo de Trabajo con Módulos

```
┌──────────────────────────────────────────────────────────┐
│                   FLUJO DE TRABAJO                        │
│                                                          │
│  1. go mod init mi-proyecto                              │
│     → Crea go.mod con nombre y versión de Go             │
│                                                          │
│  2. Creas tu código con imports                           │
│     → import "mi-proyecto/auth"                          │
│     → import "github.com/gin-gonic/gin"                  │
│                                                          │
│  3. go mod tidy                                          │
│     → Agrega dependencias faltantes a go.mod             │
│     → Elimina dependencias no usadas                     │
│     → Actualiza go.sum con los hashes                    │
│                                                          │
│  4. go build .                                           │
│     → Compila todo tu módulo                             │
│     → Descarga dependencias del proxy si es necesario    │
│                                                          │
│  5. go run .                                             │
│     → Compila y ejecuta el programa                      │
└──────────────────────────────────────────────────────────┘
```

---

## 🔗 Concepto 4: Importar Paquetes

### Importación Básica

```go
import (
    "fmt"                                    // Paquete estándar
    "math"                                   // Paquete estándar
    "14-paquetes-modulos/myutils/mathutil"   // Tu paquete propio
    "github.com/gin-gonic/gin"              // Paquete externo
)
```

### Reglas de Importación

```
╔══════════════════════════════════════════════════════════╗
║  REGLAS DE IMPORTACIÓN                                   ║
║                                                          ║
║  1. Usas el paquete por su ÚLTIMO segmento               ║
║     import "mi-proyecto/myutils/mathutil"                ║
║     → mathutil.Promedio(...)                             ║
║                                                          ║
║  2. NO puedes importar algo que no usas                   ║
║     → El compilador lo prohíbe (error de build)          ║
║                                                          ║
║  3. Los CICLOS de importación están PROHIBIDOS           ║
║     → A importa B, B importa A → ERROR                   ║
║                                                          ║
║  4. Puedes crear ALIAS si hay colisiones                 ║
║     import m "mi-proyecto/myutils/mathutil"              ║
║     → m.Promedio(...)                                    ║
║                                                          ║
║  5. Puedes importar por efectos secundarios              ║
║     import _ "github.com/lib/pq"                         ║
║     → No usas el paquete directamente                    ║
╚══════════════════════════════════════════════════════════╝
```

### Importación con Alias

```go
import (
    "fmt"
    mathutil "14-paquetes-modulos/myutils/mathutil"   // Alias: usarás "mathutil."
    text "14-paquetes-modulos/myutils/textutil"        // Alias corto: usarás "text."
)

func main() {
    prom, _ := mathutil.Promedio([]float64{1, 2, 3})
    invertido := text.Invertir("hola")
    fmt.Println(prom, invertido)
}
```

### ¿Por Qué Go Prohíbe los Ciclos de Importación?

```
❌ CICLO PROHIBIDO:
  auth/ → importa → database/
  database/ → importa → auth/
  → INFINITO: el compilador no sabe qué compilar primero

✅ DISEÑO CORRECTO:
  auth/ → importa → database/
  main/ → importa → auth/
  main/ → importa → database/
  → Árbol: hay un orden claro de compilación
```

**¿Por qué?** Porque un ciclo de importación significa que dos paquetes están **acoplados circularmente**. Esto es un error de diseño — si A necesita B y B necesita A, deberían ser el mismo paquete o usar una interface para romper el ciclo.

---

## 📖 Concepto 5: Documentación con godoc

### Comentarios Exportados

Go tiene una convención poderosa: **los comentarios sobre símbolos exportados son documentación**.

```go
// Promedio calcula el promedio aritmetico de un slice de float64.
// Devuelve error si el slice esta vacio.
//
// Ejemplo:
//
//	avg, err := mathutil.Promedio([]float64{10, 20, 30})
//	// avg = 20.0, err = nil
func Promedio(nums []float64) (float64, error) {
    // ...
}
```

### Reglas de godoc

```
╔══════════════════════════════════════════════════════════╗
║  REGLAS DE GODOC                                         ║
║                                                          ║
║  1. El comentario DEBE empezar con el nombre del         ║
║     símbolo: "// Promedio calcula..."                    ║
║                                                          ║
║  2. Se ejecuta con: go doc mathutil.Promedio             ║
║     O en: pkg.go.dev/mi-proyecto/myutils/mathutil        ║
║                                                          ║
║  3. Los bloques de código se indentan con TAB             ║
║     → godoc los renderiza como código                    ║
║                                                          ║
║  4. El comentario del paquete va ANTES del package        ║
║     declaration y describe el paquete completo           ║
╚══════════════════════════════════════════════════════════╝
```

### Comentario de Paquete

```go
// Package mathutil proporciona funciones matematicas utiles
// que no vienen incluidas en la libreria estandar de Go.
//
// Todas las funciones son puras (sin efectos secundarios)
// y seguras para uso concurrente (stateless).
package mathutil
```

---

## 🏋️ Ejercicio Práctico: Librería de Utilidades Reutilizable

### El Problema

Necesitas funciones comunes que uses en múltiples proyectos: estadísticas, manipulación de texto, validaciones. En vez de copiar y pegar código entre proyectos, crearás una **librería organizada en paquetes** que puedas importar desde cualquier programa Go.

### La Solución: `myutils`

```
14-paquetes-modulos/
├── go.mod
├── main.go                    ← Programa que USA la librería
└── myutils/
    ├── mathutil/
    │   └── mathutil.go        ← Funciones matemáticas
    └── textutil/
        └── textutil.go        ← Funciones de texto
```

### Código: `go.mod` — La Identidad del Módulo

```go
module 14-paquetes-modulos

go 1.21
```

**Línea por línea:**
- `module 14-paquetes-modulos` → Este es el nombre del módulo. Cualquier paquete dentro se importa con esta ruta base.
- `go 1.21` → Requiere Go 1.21 o superior.

---

### Código: `myutils/mathutil/mathutil.go` — El Paquete de Matemáticas

```go
// Package mathutil proporciona funciones matematicas utiles.
package mathutil
```

**¿Por qué `package mathutil`?** Porque la carpeta se llama `mathutil`. Go usa el nombre de la carpeta como nombre del paquete. Cuando otro archivo hace `import "14-paquetes-modulos/myutils/mathutil"`, usa `mathutil.Promedio(...)`.

#### Errores del Paquete

```go
var ErrEmptySlice = errors.New("el slice no puede estar vacio")
var ErrNegativeValue = errors.New("el valor no puede ser negativo")
var ErrZeroValue = errors.New("el valor no puede ser cero")
```

**¿Por qué variables exportadas para errores?** Porque el que importa el paquete necesita poder **comparar** el error:

```go
prom, err := mathutil.Promedio([]float64{})
if err == mathutil.ErrEmptySlice {
    fmt.Println("¡Pasaste un slice vacío!")
}
```

Si el error fuera privado (`errEmptySlice`), nadie afuera podría compararlo. Es como una señal de tránsito que solo los de tu casa pueden ver.

#### Funciones Estadísticas

```go
func Suma(nums []float64) float64 {
    total := 0.0
    for _, n := range nums {
        total += n
    }
    return total
}
```

**`Suma` no retorna error** porque sumar un slice vacío simplemente da 0 — es un resultado válido. No hay nada "mal" con un slice vacío aquí.

```go
func Promedio(nums []float64) (float64, error) {
    if len(nums) == 0 {
        return 0, ErrEmptySlice
    }
    return Suma(nums) / float64(len(nums)), nil
}
```

**`Promedio` sí retorna error** porque dividir por cero es indefinido. El guard clause `if len(nums) == 0` protege contra esto.

**Patrón clave:** Nota cómo `Promedio` llama a `Suma` internamente. Las funciones dentro del mismo paquete pueden llamarse entre sí libremente — tanto las exportadas como las privadas.

```go
func DesviacionEstandar(nums []float64) (float64, error) {
    if len(nums) == 0 {
        return 0, ErrEmptySlice
    }
    prom, _ := Promedio(nums)         // ← Llama a OTRA función del paquete
    sumaCuadrados := 0.0
    for _, n := range nums {
        diff := n - prom
        sumaCuadrados += diff * diff
    }
    return math.Sqrt(sumaCuadrados / float64(len(nums))), nil
}
```

**Llama a `Promedio`** que a su vez llama a `Suma`. Esta es la composición de funciones dentro de un paquete — cada una hace una cosa y se combinan.

#### Funciones Matemáticas Avanzadas

```go
func Factorial(n int) (int, error) {
    if n < 0 {
        return 0, ErrNegativeValue    // ← Error exportado
    }
    if n == 0 || n == 1 {
        return 1, nil
    }
    resultado := 1
    for i := 2; i <= n; i++ {
        resultado *= i
    }
    return resultado, nil
}
```

**Guard clauses en cascada:** Primero validamos negativos, luego el caso base (0 o 1), luego el cálculo. Cada `return` anticipado evita anidar `if-else`.

```go
func Fibonacci(n int) ([]int, error) {
    if n < 0 {
        return nil, ErrNegativeValue
    }
    if n == 0 {
        return []int{}, nil
    }
    if n == 1 {
        return []int{0}, nil
    }
    seq := make([]int, n)
    seq[0] = 0
    seq[1] = 1
    for i := 2; i < n; i++ {
        seq[i] = seq[i-1] + seq[i-2]
    }
    return seq, nil
}
```

**`make([]int, n)`** crea un slice de tamaño exacto. Esto es más eficiente que usar `append` repetidamente porque conocemos el tamaño de antemano.

```go
func EsPrimo(n int) bool {
    if n < 2 {
        return false
    }
    if n == 2 {
        return true
    }
    if n%2 == 0 {
        return false
    }
    for i := 3; i <= int(math.Sqrt(float64(n))); i += 2 {
        if n%i == 0 {
            return false
        }
    }
    return true
}
```

**Optimización clave:** Solo iteramos hasta `√n` y solo por impares. Si `n` no es divisible por ningún número hasta su raíz cuadrada, es primo. Esto reduce la complejidad de O(n) a O(√n).

```go
func Clamp(valor, min, max int) int {
    if valor < min {
        return min
    }
    if valor > max {
        return max
    }
    return valor
}
```

**`Clamp`** es extremadamente útil en la vida real: limitar notas a 0-100, restringir coordenadas a un área visible, normalizar valores de sensores.

---

### Código: `myutils/textutil/textutil.go` — El Paquete de Texto

```go
package textutil

import (
    "strings"
    "unicode"
)
```

**Importa `strings` y `unicode`** de la librería estándar. Los paquetes propios pueden importar paquetes estándar sin restricciones.

```go
func Invertir(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}
```

**¿Por qué `[]rune` y no `[]byte`?** Porque un carácter como `é` ocupa 2 bytes en UTF-8. Si invertimos bytes, separamos el carácter de su acento. Los runes (puntos de código Unicode) mantienen los caracteres intactos.

```go
func EsPalindromo(s string) bool {
    limpio := strings.ToLower(strings.ReplaceAll(s, " ", ""))
    return limpio == Invertir(limpio)
}
```

**Limpieza antes de comparar:** Convertimos a minúsculas y quitamos espacios. "A man a plan a canal Panama" se convierte en "amanaplanacanalpana" — que es igual invertido.

```go
func SoloLetras(s string) string {
    var resultado strings.Builder
    for _, r := range s {
        if unicode.IsLetter(r) {
            resultado.WriteRune(r)
        }
    }
    return resultado.String()
}
```

**`strings.Builder`** es más eficiente que concatenar strings con `+` porque no crea un nuevo string en cada concatenación. Construye el resultado en un buffer interno y solo al final convierte a string.

```go
func Truncar(s string, n int) string {
    runes := []rune(s)
    if len(runes) <= n {
        return s
    }
    return string(runes[:n]) + "..."
}
```

**Truncado Unicode-safe:** Convierte a runes, corta a `n` caracteres visibles, agrega "..." si se truncó. No corta emojis por la mitad ni separa acentos.

---

### Código: `main.go` — El Programa que USA la Librería

```go
import (
    "fmt"
    "math"

    "14-paquetes-modulos/myutils/mathutil"
    "14-paquetes-modulos/myutils/textutil"
)
```

**Observa los imports:**
- `"fmt"` y `"math"` → Paquetes estándar (sin ruta de módulo)
- `"14-paquetes-modulos/myutils/mathutil"` → Tu paquete propio (usa la ruta del módulo)
- `"14-paquetes-modulos/myutils/textutil"` → Otro paquete propio

**El compilador prohíbe imports no usados.** Si importas `math` pero no lo usas, el build falla. Esto elimina imports basura que inflan el binario.

#### Usando mathutil

```go
notas := []float64{85.5, 92.0, 78.3, 96.1, 88.7, 73.4, 91.0}

total := mathutil.Suma(notas)              // Función que no retorna error
promedio, err := mathutil.Promedio(notas)  // Función que retorna error
maximo, _ := mathutil.Maximo(notas)        // Ignoramos el error con _
desv, _ := mathutil.DesviacionEstandar(notas)
mediana, _ := mathutil.Mediana(notas)
```

**Patrón `(valor, error)`:** Las funciones que pueden fallar devuelven dos valores. El segundo siempre es `error`. Si es `nil`, todo bien. Si no, debes manejarlo.

**El `_` (blank identifier):** Cuando estás SEGURO de que no habrá error (o no te importa), puedes ignorarlo. Pero en producción, **nunca ignores errores**.

#### Manejo de Errores

```go
_, err = mathutil.Promedio([]float64{})
if err != nil {
    fmt.Printf("Promedio([]) → %v\n", err)
    // Salida: Promedio([]) → el slice no puede estar vacio
}

_, err = mathutil.Factorial(-5)
if err != nil {
    fmt.Printf("Factorial(-5) → %v\n", err)
    // Salida: Factorial(-5) → el valor no puede ser negativo
}
```

**Cada error es un valor** que puedes inspeccionar, comparar, loggear o propagar. No hay stack unwinding, no hay try/catch, no hay sorpresas.

#### Combinando Paquetes

```go
materias := map[string][]float64{
    "Matematica":   {85, 90, 78, 92, 88},
    "Programacion": {95, 88, 92, 97, 91},
    "Fisica":       {70, 65, 80, 75, 72},
}

for nombre, notas := range materias {
    prom, _ := mathutil.Promedio(notas)
    max, _ := mathutil.Maximo(notas)
    desv, _ := mathutil.DesviacionEstandar(notas)
    nombreFormateado := textutil.Truncar(nombre, 15)
    fmt.Printf("📚 %s → Promedio: %.1f | Max: %.0f | Desv: %.1f\n",
        nombreFormateado, prom, max, desv)
}
```

**Un solo bloque usa AMBOS paquetes:** `mathutil` para estadísticas, `textutil` para formatear texto. Esto es el poder de la organización por paquetes: cada uno hace una cosa bien.

### Salida del Programa

```
🚀 LECCION 14: PAQUETES Y MODULOS EN GO
   Organizando codigo como un profesional

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  📊 mathutil: Estadistica Basica
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Notas del grupo: [85.5 92 78.3 96.1 88.7 73.4 91]
  Suma total:      605.0
  Promedio:        86.43
  Nota mas alta:   96.1
  Nota mas baja:   73.4
  Desv. estandar:  7.44
  Mediana:         88.7

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🔢 mathutil: Matematicas Avanzadas
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  10! = 3628800
  Fibonacci(12) = [0 1 1 2 3 5 8 13 21 34 55 89]
  Primos del 1 al 30: 2 3 5 7 11 13 17 19 23 29

  Clamp(105, 0, 100) = 100  (nota maxima 100)
  Clamp(-5, 0, 100)  = 0    (nota minima 0)
  Clamp(87, 0, 100)  = 87   (nota dentro del rango)
  85.5 es el 85.5% de 100

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ⚠️ mathutil: Manejo de Errores
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Promedio([])       → ❌ el slice no puede estar vacio
  Factorial(-5)      → ❌ el valor no puede ser negativo
  Fibonacci(-3)      → ❌ el valor no puede ser negativo
  Porcentaje(50, 0)  → ❌ el valor no puede ser cero

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  📝 textutil: Manipulacion de Texto
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Frase original:    "A man a plan a canal Panama"
  Es palindromo?     true
  Cantidad palabras: 7
  Invertida:         "amanaP lanac a nalp a nam A"

  Texto truncado:    "Este es un texto muy largo que..."
  Solo letras:       "Goesgenial" (de "Go 2.0 es genial!!! @#$%")

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🔬 Ejemplo Real: Analisis de Calificaciones
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  📚 Programacion → Promedio: 92.6 | Max: 97 | Desv: 3.1
  📚 Fisica       → Promedio: 72.4 | Max: 80 | Desv: 5.0
  📚 Matematica   → Promedio: 86.6 | Max: 92 | Desv: 4.9

  🏆 Mejor materia: Programacion (promedio: 92.6)
```

---

## 📂 Resumen Visual de la Estructura

```
14-paquetes-modulos/
│
├── go.mod                          ← "Soy el módulo 14-paquetes-modulos, Go 1.21"
│
├── main.go                         ← package main (programa principal)
│   ├── import mathutil             ← Usa funciones matemáticas
│   └── import textutil             ← Usa funciones de texto
│
└── myutils/                        ← Carpeta contenedora (no es paquete por sí misma)
    │
    ├── mathutil/                   ← package mathutil (10 funciones exportadas)
    │   ├── Suma()                  ← 🔴 Exportada
    │   ├── Promedio()              ← 🔴 Exportada (retorna error)
    │   ├── Maximo()                ← 🔴 Exportada
    │   ├── Minimo()                ← 🔴 Exportada
    │   ├── Mediana()               ← 🔴 Exportada
    │   ├── DesviacionEstandar()    ← 🔴 Exportada
    │   ├── Factorial()             ← 🔴 Exportada
    │   ├── Fibonacci()             ← 🔴 Exportada
    │   ├── EsPrimo()               ← 🔴 Exportada
    │   ├── Clamp()                 ← 🔴 Exportada
    │   ├── Porcentaje()            ← 🔴 Exportada
    │   ├── ErrEmptySlice           ← 🔴 Error exportado
    │   ├── ErrNegativeValue        ← 🔴 Error exportado
    │   └── ErrZeroValue            ← 🔴 Error exportado
    │
    └── textutil/                   ← package textutil (6 funciones exportadas)
        ├── Invertir()              ← 🔴 Exportada
        ├── EsPalindromo()          ← 🔴 Exportada
        ├── ContarPalabras()        ← 🔴 Exportada
        ├── TituloCapital()         ← 🔴 Exportada
        ├── SoloLetras()            ← 🔴 Exportada
        └── Truncar()               ← 🔴 Exportada
```

---

## 🧠 Tabla de Decisión: Conceptos Clave

| Concepto | Qué es | Ejemplo |
|----------|--------|---------|
| **Paquete** | Carpeta con código Go relacionado | `package mathutil` |
| **Módulo** | Grupo de paquetes con go.mod | `module mi-proyecto` |
| **Exportado** | Nombre con mayúscula inicial | `func Promedio()` |
| **Privado** | Nombre con minúscula inicial | `func calcular()` |
| **go.mod** | Identidad del módulo + dependencias | `go mod init` |
| **go.sum** | Hashes de verificación de dependencias | Automático |
| **go mod tidy** | Sincroniza dependencias | `go mod tidy` |
| **go get** | Agrega/actualiza una dependencia | `go get pkg@v1.0` |
| **godoc** | Documentación en comentarios | `// Funcion hace X` |
| **Import path** | Ruta completa para importar | `"mi-modulo/paquete"` |
| **Ciclo de import** | A importa B importa A → PROHIBIDO | Error de compilación |

---

## 🔑 Reglas de Oro

```
╔══════════════════════════════════════════════════════════╗
║                  REGLAS DE ORO                            ║
║                                                          ║
║  1. Un paquete = una carpeta. Sin excepciones.           ║
║                                                          ║
║  2. Mayúscula = exportado. Minúscula = privado.          ║
║     No hay keywords, no hay excepciones.                 ║
║                                                          ║
║  3. Cada archivo .go DEBE empezar con package xxx.       ║
║     Todos los archivos de la misma carpeta = mismo pkg.  ║
║                                                          ║
║  4. Los imports no usados son ERROR de compilación.      ║
║     El compilador es tu editor de código limpio.         ║
║                                                          ║
║  5. Los ciclos de importación están PROHIBIDOS.          ║
║     Si A necesita B y B necesita A, es un error de      ║
║     diseño. Usa interfaces para romper el ciclo.         ║
║                                                          ║
║  6. Siempre ejecuta "go mod tidy" antes de commitear.    ║
║     Limpia dependencias no usadas y agrega las faltantes.║
║                                                          ║
║  7. Documenta TODO símbolo exportado con // Comentario.  ║
║     Si no lo documentas, nadie sabe cómo usarlo.         ║
║                                                          ║
║  8. Los errores del paquete deben ser variables           ║
║     exportadas para que el importador pueda compararlos.  ║
╚══════════════════════════════════════════════════════════╝
```

---

## 🎯 Ejercicio Feynman

### Instrucciones

1. **Explica en voz alta** (o escribe) qué es un paquete y un módulo con tus propias palabras
2. **Dibuja el diagrama** de cómo se importan los paquetes en el ejercicio
3. **Completa los ejercicios** a continuación

### Ejercicio 1: ¿Exportado o Privado?

Dada esta función, ¿es visible desde otro paquete?

```go
func CalcularImpuesto(monto float64) float64 {
    return monto * 0.18
}
```

**Respuesta:** ____________

Ahora esta:

```go
func calcularImpuesto(monto float64) float64 {
    return monto * 0.18
}
```

**Respuesta:** ____________

### Ejercicio 2: Diseña tu Propio Paquete

Diseña un paquete llamado `validador` que tenga:
- 3 funciones exportadas
- 2 funciones privadas (de apoyo)
- 1 error exportado

Escribe la firma de cada función (solo la primera línea, sin cuerpo):

```go
package validador

// Errores
var ...

// Exportadas
func ...
func ...
func ...

// Privadas
func ...
func ...
```

### Ejercicio 3: Rompe un Ciclo de Importación

Tienes esta situación:
```
paquete "email" necesita "usuario" para obtener el email del usuario
paquete "usuario" necesita "email" para enviar un email de bienvenida
```

Esto crea un ciclo. ¿Cómo lo resuelves? Pista: crea un tercer paquete o usa interfaces.

**Escribe tu solución:** ____________

### Ejercicio 4: Explica go mod tidy

Un colega dice: *"Yo nunca ejecuto go mod tidy, solo agrego dependencias con go get"*. ¿Qué problemas podría tener? Lista al menos 3.

**Respuesta:** ____________

### Ejercicio 5: Errores como Variables Exportadas

¿Por qué los errores en un paquete deben ser `var ErrAlgo = errors.New("...")` (exportado) en vez de `var errAlgo = errors.New("...")` (privado)?

Escribe un ejemplo de código que DEMUESTRE por qué necesitas que sea exportado.

**Respuesta:** ____________

---

## 📝 Resumen de la Lección

```
┌─────────────────────────────────────────────────────────────┐
│                  PAQUETES Y MÓDULOS                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Paquete   → Carpeta con código Go relacionado              │
│  Módulo    → Grupo de paquetes con go.mod como raíz         │
│  go.mod    → Identidad: nombre, versión Go, dependencias    │
│  go.sum    → Hashes de seguridad de dependencias            │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                  VISIBILIDAD                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Mayúscula → EXPORTADO (visible desde otros paquetes)       │
│  Minúscula → PRIVADO (solo visible dentro del paquete)      │
│  Aplica a: funciones, variables, tipos, campos, métodos     │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                  COMANDOS ESENCIALES                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  go mod init X     → Crear módulo con nombre X              │
│  go mod tidy       → Sincronizar dependencias               │
│  go get pkg@v      → Agregar/actualizar dependencia         │
│  go build .        → Compilar el módulo                     │
│  go doc pkg.Func   → Ver documentación de una función       │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│                  REGLAS DE ORO                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  1. Un paquete = una carpeta                                │
│  2. Mayúscula exporta, minúscula priva                      │
│  3. Imports no usados → error de compilación                │
│  4. Ciclos de importación → prohibidos                      │
│  5. Documenta todo símbolo exportado                        │
│  6. Errores del paquete → variables exportadas              │
│  7. "go mod tidy" antes de cada commit                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🚀 Siguiente Lección

En la **Lección 15** aprenderemos sobre **Testing en Go** (`testing` package):
- Table-driven tests: el patrón estándar de la industria
- Subtests para organizar casos de prueba
- Benchmarks para medir rendimiento
- Examples que también sirven como documentación
- Cobertura de código con `go test -cover`

Los tests son lo que separa al código "que funciona" del código "que funciona en producción". Y Go tiene el mejor soporte de testing integrado de cualquier lenguaje.

---

> *"Si no puedes explicar la visibilidad de Go con la regla de la mayúscula, no has entendido paquetes. Si no sabes qué hace go mod tidy, no has entendido módulos. Vuelve, simplifica, vuelve a intentar."*
> — Filosofía Feynman