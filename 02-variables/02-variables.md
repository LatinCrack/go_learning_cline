<div align="center">

# 📘 Lección 02

# Variables, Tipos y el Sistema de Tipos de Go

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Fase I](https://img.shields.io/badge/Fase_I-Fundamentos-4ECDC4?style=for-the-badge) ![Lección 02](https://img.shields.io/badge/Lecci%C3%B3n-02-FF6B6B?style=for-the-badge)

<br>

> *"El tipo de una variable es su contrato con el compilador.
> Rompe el contrato y el compilador te detiene antes de que rompas producción."*
> — **Rob Pike**, co-creador de Go

</div>

---

## 🎯 ¿Qué vas a dominar en esta lección?

| Concepto | Descripción |
|:---------|:------------|
| 📦 **Declaración de variables** | `var`, `:=`, `const` y cuándo usar cada uno |
| 🔢 **Tipos primitivos** | `int`, `float64`, `string`, `bool` y sus variantes |
| 🔄 **Conversiones explícitas** | Por qué Go NO hace casting automático y cómo convertir manualmente |
| 🏷️ **Zero values** | El valor por defecto de cada tipo — nunca `null`/`undefined` |
| 🧩 **Tipos compuestos** | `array`, `slice`, `map` (introducción) |
| 📐 **fmt.Sprintf** | Formateo avanzado de strings para construir salidas legibles |
| 🏋️ **Ejercicio práctico** | Conversor de unidades universal (temperatura, distancia, almacenamiento) |

---

## 🧬 1. ¿Por qué los tipos importan?

### La analogía de la cocina profesional

Imagina que eres chef en un restaurante de alta cocina. El sous-chef te grita: *"¡Dame el polvo blanco!"*

¿Le das **harina** o **azúcar**? Ambos son polvos blancos. Ambos se ven iguales a simple vista. Pero si le das azúcar cuando pidió harina, el pollo queda arruinado.

**El compilador de Go es ese sous-chef exigente.** Cuando le dices `var temperatura int`, le estás diciendo al compilador: *"Esto es un número entero, no me des un string, no me des un float, no me des nada que no sea exactamente lo que pedí."*

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│   Python (tipado dinámico):                                         │
│     x = 42          → x es int                                     │
│     x = "hola"      → x ahora es string (¡sin error!)              │
│     x = [1, 2, 3]   → x ahora es lista (¡otra vez sin error!)      │
│                                                                     │
│   Go (tipado estático):                                             │
│     var x int = 42                                                  │
│     x = "hola"      → ❌ ERROR DE COMPILACIÓN                      │
│     "No puedo asignar string a int"                                 │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 🤔 ¿Por qué esto es una VENTAJA?

En Python, esto puede pasar:

```python
# Python — Este código se ejecuta SIN errores...
def calcular_precio(precio, cantidad):
    return precio * cantidad

resultado = calcular_precio("10", 3)
print(resultado)  # "101010" ← ¡Multiplicó un string por un entero!
```

¿Ves el bug? En Python, `"10" * 3` produce `"101010"` (repite el string 3 veces). No hay error. No hay advertencia. El bug llega **silenciosamente** a producción y un cliente recibe `"101010"` como precio.

En Go, esto **jamás** pasaría:

```go
// Go — Este código NI SIQUIERA COMPILA
func calcular_precio(precio string, cantidad int) int {
    return precio * cantidad  // ❌ ERROR: cannot use * operator on string and int
}
```

> 🧠 **Regla de oro:** Go no te deja mezclar harina con azúcar. El compilador te detiene **antes** de que el plato llegue al cliente.

---

## 📦 2. Declaración de Variables

Go tiene **tres formas** de declarar variables. Cada una tiene su momento ideal:

### Forma 1: `var` — La declaración explícita

```go
var nombre string = "Go"
var edad int = 15
var pi float64 = 3.14159
var esGenial bool = true
```

- Usa `var` cuando declaras una variable **fuera de una función** (a nivel de paquete).
- Es la forma más verbosa, pero la más clara.

### Forma 2: `:=` — La declaración corta (la favorita de Go)

```go
nombre := "Go"           // string (inferido)
edad := 15               // int (inferido)
pi := 3.14159            // float64 (inferido)
esGenial := true         // bool (inferido)
```

- El operador `:=` **declara Y asigna** la variable en un solo paso.
- Go **infiere el tipo** automáticamente a partir del valor.
- Solo se puede usar **dentro de funciones** (no a nivel de paquete).
- Es la forma más común y la que verás en el 90% del código Go real.

> **Analogía:** `:=` es como el auto-enfoque de una cámara moderna. Le apuntas al sujeto y la cámara ajusta todo automáticamente. `var` es como el enfoque manual — más control, pero más trabajo.

### Forma 3: `var` sin valor — El zero value

```go
var nombre string    // nombre = ""
var edad int         // edad = 0
var precio float64   // precio = 0.0
var activo bool      // activo = false
```

- Si declaras una variable sin asignarle valor, Go le asigna el **zero value**.
- Esto es **una de las decisiones de diseño más geniales de Go** — nunca tendrás `null`, `undefined`, `nil` (para tipos primitivos), o un valor "basura" como en C.

### 🏷️ Zero Values — El valor seguro por defecto

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│   TIPO              ZERO VALUE         EJEMPLO                     │
│   ─────────────     ──────────────     ─────────────────────────   │
│   int, int64...     0                  var x int           → 0    │
│   float32/64        0.0                var f float64       → 0.0  │
│   string            "" (vacío)         var s string        → ""   │
│   bool              false              var b bool          → false│
│   pointer           nil                var p *int          → nil  │
│   slice             nil                var s []int         → nil  │
│   map               nil                var m map[string]int→ nil  │
│   channel           nil                var ch chan int      → nil  │
│   func              nil                var f func()        → nil  │
│   interface         nil                var i interface{}   → nil  │
│   struct            campos con sus     var p Persona       →      │
│                     zero values             p.Nombre → ""         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

> 🧠 **¿Por qué importa esto?** En Java, si olvidas inicializar un `int`, obtienes un error. En C, obtienes un valor "basura" (lo que sea que hubiera en esa posición de memoria). En JavaScript, obtienes `undefined`. En Go, obtienes un **valor válido y predecible**. Tu variable siempre tiene un estado conocido.

### `var` vs `:=` — ¿Cuándo usar cada uno?

| Situación | Usar | Razón |
|:----------|:-----|:------|
| Dentro de una función, valor inicial conocido | `:=` | Más conciso, inferencia automática |
| Fuera de una función (nivel de paquete) | `var` | `:=` no funciona fuera de funciones |
| Necesitas declarar sin valor inicial | `var` | Go asigna el zero value |
| Necesitas el tipo específico (no el inferido) | `var` | Ej: `var x float64 = 3` (sin `var` sería `int`) |
| Múltiples variables del mismo tipo | `var` o `:=` | `var (a, b, c int)` o `a, b, c := 1, 2, 3` |

### Declaración múltiple

```go
// Con var — bloque de declaraciones
var (
    nombre  string = "Go"
    version int    = 1
    creador string = "Google"
)

// Con := — asignación múltiple
x, y, z := 1, 2, 3

// Variables del mismo tipo en una línea
var a, b, c int
```

---

## 🔢 3. Tipos Primitivos de Go

Go tiene un conjunto **finito y deliberado** de tipos primitivos. No hay docenas de variantes como en Java o C++.

### 🔢 Tipos numéricos enteros

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│   TIPO        TAMAÑO     RANGO                                    │
│   ─────────   ──────     ─────────────────────────────────────     │
│   int8        8 bits     -128 a 127                                │
│   int16       16 bits    -32,768 a 32,767                          │
│   int32       32 bits    -2,147,483,648 a 2,147,483,647            │
│   int64       64 bits    -9.2 × 10¹⁸ a 9.2 × 10¹⁸                 │
│   int         32 o 64*  Depende de la plataforma (64 bits hoy)     │
│                                                                     │
│   uint8       8 bits     0 a 255                                   │
│   uint16      16 bits    0 a 65,535                                │
│   uint32      32 bits    0 a 4,294,967,295                         │
│   uint64      64 bits    0 a 18.4 × 10¹⁸                           │
│   uint        32 o 64*  Sin signo                                  │
│                                                                     │
│   byte        8 bits     Alias de uint8 (representa un byte)       │
│   rune        32 bits    Alias de int32 (representa un code point)  │
│                                                                     │
│   * int y uint tienen el tamaño de la palabra nativa del sistema.   │
│     En máquinas de 64 bits (casi todas hoy), son de 64 bits.        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

> **Regla práctica:** Usa `int` por defecto. Solo usa `int64`, `int32`, etc. cuando necesites un tamaño específico (por ejemplo, para interoperar con APIs o formatos de archivo que lo requieran).

### 🔢 Tipos numéricos de punto flotante

```go
var precio float64 = 19.99    // 64 bits — el estándar, usa SIEMPRE este
var peso float32 = 72.5       // 32 bits — menor precisión, raramente necesario
```

| Tipo | Bits | Precisión | Cuándo usar |
|:-----|:-----|:----------|:------------|
| `float64` | 64 | ~15-16 dígitos decimales | **Siempre por defecto** |
| `float32` | 32 | ~6-7 dígitos decimales | Solo si necesitas ahorrar memoria (ej: gráficos 3D, ML con millones de datos) |

> ⚠️ **Dato crucial:** Cuando escribes `3.14` en Go, el compilador infiere `float64`, NO `float32`. Esto es diferente a `3` que infiere `int`.

### 🔤 Tipo string

```go
var saludo string = "¡Hola, Go!"
nombre := "Gopher"
```

Los strings en Go son:
- **Inmutables** — no puedes cambiar un carácter individual: `nombre[0] = 'B'` es un error
- **UTF-8** — pueden contener caracteres de cualquier idioma y emojis
- **Un slice de bytes internamente** — pero no trates directamente como uno (lo veremos en la Lección 08)

```go
// Operaciones comunes con strings
len("Go")           // 2 (número de bytes, no de caracteres)
len("café")         // 5 (¡la "é" ocupa 2 bytes en UTF-8!)
"Go" + " lang"      // "Go lang" (concatenación)
```

### ✅ Tipo bool

```go
var activo bool = true
esValido := false

// Los booleanos NO son números en Go
// NO puedes hacer: if 1 { ... } como en C
// Debes usar: if activo { ... }
```

### 🧮 Tipo `byte` y `rune`

```go
var b byte = 'A'      // byte es alias de uint8 (0-255)
var r rune = 'ñ'      // rune es alias de int32 (cualquier Unicode)
```

- **`byte`** → Un solo byte (0-255). Útil para datos binarios.
- **`rune`** → Un "carácter" Unicode (puede ser de 1 a 4 bytes). Útil para texto.

> **Analogía:** Piensa en `byte` como una moneda de un centavo (unidad mínima de dinero) y `rune` como una "letra" real — una "ñ" es una letra (rune), pero ocupa dos centavos (bytes) en el disco.

### 📝 Tipos de constantes — `const`

```go
const pi = 3.14159
const nombreApp = "ConversorGo"
const maxIntentos = 3

// Con tipo explícito
const version float64 = 2.0

// Bloque de constantes
const (
    statusOK    = 200
    statusError = 500
)
```

Las constantes en Go:
- Se evalúan en **tiempo de compilación**, no en tiempo de ejecución
- No pueden ser cambiadas después de su declaración
- No pueden ser declaradas con `:=` (solo con `const`)

---

## 🔄 4. Conversiones Explícitas — Go No Hace Magia Negra

Esta es una de las diferencias más importantes entre Go y los lenguajes de tipado dinámico. **Go NO convierte tipos automáticamente.** Nunca. Jamás.

### ❌ Lo que NO funciona en Go

```go
var edad int = 25
var precio float64 = 9.99

resultado := edad + precio  
// ❌ ERROR: invalid operation: mismatched types int and float64

var x int = 65
letra := string(x)
// ⚠️ Esto SÍ compila, pero ¿qué imprime? No "65" — imprime "A" (code point 65 = 'A')
```

### ✅ La conversión explícita

Para operar con tipos diferentes, debes convertir **explícitamente**:

```go
var edad int = 25
var precio float64 = 9.99

// Opción 1: Convertir int a float64
resultado := float64(edad) + precio    // ✅ 34.99

// Opción 2: Convertir float64 a int (¡PIERDE la parte decimal!)
entero := int(precio)                   // ✅ 9 (trunca, no redondea)

// Convertir int a string (¡cuidado! da el carácter Unicode, no el número)
letra := string(65)                     // "A" (no "65")

// Convertir número a string correctamente
numero := fmt.Sprintf("%d", 65)         // "65" ✅

// Convertir string a número
import "strconv"
num, err := strconv.Atoi("42")          // num = 42, err = nil
fl, err := strconv.ParseFloat("3.14", 64) // fl = 3.14, err = nil
```

### 🧠 ¿Por qué Go prohíbe las conversiones implícitas?

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│   C++ (conversiones implícitas):                                    │
│     int x = 3.9;     // x vale 3 ← truncamiento silencioso         │
│     bool b = x;      // b vale true ← ¿querías esto?               │
│                                                                     │
│   Go (conversiones explícitas):                                     │
│     var x int = 3.9  // ❌ ERROR: no puedo asignar float64 a int    │
│     x := int(3.9)    // ✅ x = 3 — tú decidiste truncar            │
│                                                                     │
│   En C++, el compilador hizo una decisión por ti silenciosamente.   │
│   En Go, TÚ tomas la decisión y escribes la intención explícita.    │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

> **Analogía:** Las conversiones implícitas son como un traductor automático que decide por ti cuándo traducir del español al inglés. A veces acierta, a veces traduce "estoy embarazada" como "I am embarrassed" (significado completamente diferente). Go te obliga a **revisar cada traducción manualmente**.

---

## 🧩 5. Tipos Compuestos — Introducción

Los tipos compuestos combinan tipos primitivos para crear estructuras más complejas. Aquí los veremos brevemente; las lecciones 06, 07 y 08 los profundizan.

### 📦 Arrays — Tamaño fijo

```go
// Un array de exactamente 5 enteros
var numeros [5]int
numeros[0] = 10
numeros[1] = 20

// Declaración con valores iniciales
colores := [3]string{"rojo", "verde", "azul"}

// Go puede contar los elementos por ti
frutas := [...]string{"manzana", "plátano", "naranja"} // len = 3
```

> ⚠️ Los arrays en Go son de **tamaño fijo** y se pasan **por valor** (se copian). Si creas `[5]int` y lo pasas a una función, se copian los 5 elementos. Por eso casi nadie usa arrays directamente — usan **slices**.

### 🔪 Slices — Arrays dinámicos (el 90% del tiempo)

```go
// Un slice de enteros (tamaño dinámico)
notas := []int{9, 8, 7, 10, 6}

// Agregar elementos
notas = append(notas, 5)

// Crear un slice vacío con capacidad inicial
puntajes := make([]int, 0, 10) // len=0, cap=10

// Acceder a elementos
primera := notas[0]    // 9
ultima := notas[len(notas)-1] // 5 (el último elemento)
```

> **Analogía:** Un array es un estacionamiento con exactamente 10 espacios — si necesitas 11, debes construir otro. Un slice es un auto con un remolque expandible: el runtime de Go agranda el remolque automáticamente cuando necesitas más espacio.

### 🗺️ Maps — Diccionarios (tablas hash)

```go
// Un map de string a int (como un diccionario)
edades := map[string]int{
    "Ana":  30,
    "Luis": 25,
    "María": 28,
}

// Agregar un nuevo par
edades["Pedro"] = 35

// Buscar un valor (comma ok idiom)
edad, existe := edades["Ana"]
if existe {
    fmt.Printf("Ana tiene %d años\n", edad)
}

// Eliminar una clave
delete(edades, "Luis")
```

> **Analogía:** Un map es como el índice de un libro. En vez de leer todo el libro para encontrar "algoritmo", vas al índice y en O(1) encuentras la página.

---

## 📐 6. Formateo Avanzado con `fmt.Sprintf`

Ya conoces `fmt.Printf` de la Lección 01. Ahora vamos con `fmt.Sprintf` — que no imprime, sino que **devuelve** el string formateado. Es como un `Printf` que escribe en un buffer invisible en vez de la pantalla.

```go
// Printf → imprime directamente en consola
fmt.Printf("Hola %s\n", "Go")

// Sprintf → devuelve el string (NO imprime)
mensaje := fmt.Sprintf("Hola %s", "Go")
// mensaje ahora es "Hola Go"
```

### Verbos de formato avanzados para números

| Verbo | Descripción | Ejemplo | Salida |
|:------|:------------|:--------|:-------|
| `%d` | Entero decimal | `Sprintf("%d", 42)` | `42` |
| `%f` | Float (6 decimales por defecto) | `Sprintf("%f", 3.14)` | `3.140000` |
| `%.2f` | Float con 2 decimales | `Sprintf("%.2f", 3.14)` | `3.14` |
| `%e` | Notación científica | `Sprintf("%e", 123456.789)` | `1.234568e+05` |
| `%s` | String | `Sprintf("%s", "Go")` | `Go` |
| `%v` | Cualquier valor (formato por defecto) | `Sprintf("%v", true)` | `true` |
| `%T` | Tipo del valor | `Sprintf("%T", 42)` | `int` |
| `%b` | Binario | `Sprintf("%b", 10)` | `1010` |
| `%x` | Hexadecimal | `Sprintf("%x", 255)` | `ff` |
| `%05d` | Entero con ceros a la izquierda (ancho 5) | `Sprintf("%05d", 42)` | `00042` |
| `%10s` | String a la derecha (ancho 10) | `Sprintf("%10s", "Go")` | `        Go` |
| `%-10s` | String a la izquierda (ancho 10) | `Sprintf("%-10s", "Go")` | `Go        ` |

> 🧠 **`Sprintf` es tu mejor amigo** para construir salidas legibles. En el ejercicio práctico, lo usarás extensivamente para formatear resultados de conversiones.

---

## 🏋️ 7. Ejercicio Práctico: Conversor de Unidades Universal

Ahora vamos a construir algo **realmente útil**. Un conversor de unidades que convierte:

- 🌡️ **Temperatura:** °C ↔ °F ↔ K
- 📏 **Distancia:** km ↔ millas ↔ metros
- 💾 **Almacenamiento:** bytes ↔ KB ↔ MB ↔ GB

La herramienta leerá argumentos desde la línea de comandos y mostrará **todas las conversiones posibles**.

### 📁 Estructura del proyecto

```
02-variables/
├── main.go          ← Tu código fuente
└── go.mod           ← La "ficha de identidad" del módulo
```

### 📄 Archivo `go.mod`

```
module 02-variables

go 1.22
```

### 📄 Archivo `main.go` — Explicación línea por línea

```go
package main
```

Declara que este archivo pertenece al paquete `main` — es un ejecutable. Ya lo conoces de la Lección 01.

---

```go
import (
    "fmt"
    "os"
    "strconv"
)
```

Importamos tres paquetes:

- **`fmt`** → Para imprimir con formato
- **`os`** → Para leer argumentos de la línea de comandos (`os.Args`)
- **`strconv`** → Para convertir strings a números (`strconv.ParseFloat`)

---

```go
func main() {
    // Mostrar encabezado
    fmt.Println("╔══════════════════════════════════════════════════════╗")
    fmt.Println("║   🔄 Conversor de Unidades Universal — Lab Go       ║")
    fmt.Println("╚══════════════════════════════════════════════════════╝")
```

`Println` imprime el encabezado con caracteres Unicode decorativos. Misma técnica que la Lección 01.

---

```go
    // Verificar argumentos
    if len(os.Args) < 3 {
        fmt.Println("\n  ⚠️  Uso: go run main.go <tipo> <valor>")
        fmt.Println("  Tipos disponibles: temp, dist, data")
        fmt.Println("  Ejemplo: go run main.go temp 100")
        os.Exit(1)
    }
```

`os.Args` es un **slice de strings** que contiene los argumentos de la línea de comandos. `os.Args[0]` es siempre el nombre del programa. Necesitamos al menos 3 elementos: `[programa, tipo, valor]`.

`len(os.Args)` devuelve la longitud del slice. Si el usuario no proporciona suficientes argumentos, mostramos ayuda y salimos con `os.Exit(1)` (código de salida 1 = error).

---

```go
    tipo := os.Args[1]
    valorStr := os.Args[2]
```

Extraemos el tipo de conversión y el valor como strings. `tipo` será `"temp"`, `"dist"` o `"data"`. `valorStr` será el número escrito como texto (por ejemplo, `"100"`).

---

```go
    valor, err := strconv.ParseFloat(valorStr, 64)
    if err != nil {
        fmt.Printf("\n  ❌ Error: '%s' no es un número válido\n", valorStr)
        os.Exit(1)
    }
```

Aquí ocurre la **magia de la conversión de tipos**. `strconv.ParseFloat(valorStr, 64)` intenta convertir el string `"100"` al float64 `100.0`.

Devuelve **dos valores** (el patrón `(resultado, error)` que vimos en la Lección 01):
1. `valor` → el número convertido (float64)
2. `err` → `nil` si todo salió bien, o un error si el string no es un número válido

El segundo argumento `64` le dice a `ParseFloat` que genere un resultado de precisión de 64 bits (es decir, `float64`).

Si el usuario escribe `"abc"` en vez de `"100"`, el error no es `nil` y mostramos un mensaje amigable.

---

```go
    fmt.Println() // Línea en blanco para separar visualmente
```

---

```go
    switch tipo {
    case "temp":
        convertirTemperatura(valor)
    case "dist":
        convertirDistancia(valor)
    case "data":
        convertirAlmacenamiento(valor)
    default:
        fmt.Printf("  ❌ Tipo desconocido: '%s'\n", tipo)
        fmt.Println("  Tipos disponibles: temp, dist, data")
        os.Exit(1)
    }
```

Un `switch` en Go evalúa el tipo de conversión solicitado. Observa que:
- No necesita `break` (a diferencia de C/Java) — cada `case` se ejecuta y termina automáticamente.
- El `default` captura cualquier tipo no reconocido.

Cada `case` llama a una función especializada que hace las conversiones. Esto sigue el principio de **separación de responsabilidades**: `main()` se encarga de la lógica de entrada/salida, las funciones se encargan de las matemáticas.

---

#### Función `convertirTemperatura`

```go
func convertirTemperatura(celsius float64) {
```

Recibe un valor en grados Celsius (el valor que el usuario ingresó). Desde aquí convertimos a Fahrenheit y Kelvin.

---

```go
    fahrenheit := celsius*9.0/5.0 + 32.0
    kelvin := celsius + 273.15
```

Las fórmulas de conversión:

- **Celsius → Fahrenheit:** `(°C × 9/5) + 32`
- **Celsius → Kelvin:** `°C + 273.15`

Observa que operamos con `float64` en todos los operandos. Los literales `9.0`, `5.0`, `32.0` y `273.15` son `float64` implícitos. Si escribiéramos `9/5` con enteros, Go calcularía `1` (división entera) en vez de `1.8`. **Esto es un bug común en Go para principiantes.**

---

```go
    fmt.Println("  🌡️  Conversión de Temperatura")
    fmt.Println("  ─────────────────────────────────────────")
    fmt.Printf("  📥 Celsius     : %10.2f °C\n", celsius)
    fmt.Printf("  📤 Fahrenheit  : %10.2f °F\n", fahrenheit)
    fmt.Printf("  📤 Kelvin      : %10.2f K\n", kelvin)
    fmt.Println("  ─────────────────────────────────────────")
```

`%10.2f` significa: número de punto flotante, con un ancho total de 10 caracteres y exactamente 2 decimales. Esto alinea perfectamente los resultados en columnas.

---

```go
    // Conversiones adicionales entre unidades
    fmt.Printf("\n  🔄 Tabla de conversiones cruzadas:\n")
    fmt.Printf("     %.2f°F = %.2f°C = %.2fK\n", fahrenheit, celsius, kelvin)
    fmt.Printf("     %.2fK = %.2f°C = %.2f°F\n", kelvin, celsius, fahrenheit)
}
```

Mostramos una tabla de conversiones cruzadas para que el usuario vea todas las relaciones posibles entre las tres escalas.

---

#### Función `convertirDistancia`

```go
func convertirDistancia(km float64) {
    millas := km * 0.621371
    metros := km * 1000.0
```

Conversiones de distancia:
- **Kilómetros → Millas:** `km × 0.621371`
- **Kilómetros → Metros:** `km × 1000`

---

```go
    fmt.Println("  📏 Conversión de Distancia")
    fmt.Println("  ─────────────────────────────────────────")
    fmt.Printf("  📥 Kilómetros  : %12.4f km\n", km)
    fmt.Printf("  📤 Millas      : %12.4f mi\n", millas)
    fmt.Printf("  📤 Metros      : %12.4f m\n", metros)
    fmt.Println("  ─────────────────────────────────────────")

    fmt.Printf("\n  🔄 Tabla de conversiones cruzadas:\n")
    fmt.Printf("     %.4fkm = %.4fmi = %.4fm\n", km, millas, metros)
    fmt.Printf("     %.4fmi = %.4fkm = %.4fm\n", millas, km, metros)
}
```

Mismo patrón de presentación que la temperatura, con `%12.4f` para mostrar 4 decimales y alinear las columnas.

---

#### Función `convertirAlmacenamiento`

```go
func convertirAlmacenamiento(bytes float64) {
    kb := bytes / 1024.0
    mb := bytes / (1024.0 * 1024.0)
    gb := bytes / (1024.0 * 1024.0 * 1024.0)
```

Conversiones de almacenamiento:
- **Bytes → KB:** dividir por 1024 (no por 1000 — en informática se usa base 2)
- **Bytes → MB:** dividir por 1024² (1,048,576)
- **Bytes → GB:** dividir por 1024³ (1,073,741,824)

> 🧠 **Dato importante:** 1 KB = 1024 bytes, no 1000. Esto es porque los computadores trabajan en base 2. El prefijo correcto es "kibibyte" (KiB), pero la industria dice "kilobyte" informalmente. Si escribes `bytes / 1000.0` aquí, todo tu conversor estará mal.

---

```go
    fmt.Println("  💾 Conversión de Almacenamiento")
    fmt.Println("  ─────────────────────────────────────────")
    fmt.Printf("  📥 Bytes       : %15.0f B\n", bytes)
    fmt.Printf("  📤 Kilobytes   : %15.4f KB\n", kb)
    fmt.Printf("  📤 Megabytes   : %15.4f MB\n", mb)
    fmt.Printf("  📤 Gigabytes   : %15.4f GB\n", gb)
    fmt.Println("  ─────────────────────────────────────────")

    fmt.Printf("\n  🔄 Tabla de conversiones cruzadas:\n")
    fmt.Printf("     %.0fB = %.4fKB = %.4fMB = %.6fGB\n", bytes, kb, mb, gb)
}
```

Para bytes usamos `%15.0f` — 15 caracteres de ancho, 0 decimales (son bytes, no tiene sentido mostrar decimales para el valor base). Para KB, MB, GB mostramos 4 decimales, y para GB mostramos 6 porque son valores muy pequeños cuando el input son bytes.

---

### ▶️ Ejecutar el programa

```bash
# Dentro del directorio 02-variables/
go run main.go temp 100
```

**Salida esperada:**

```
╔══════════════════════════════════════════════════════╗
║   🔄 Conversor de Unidades Universal — Lab Go       ║
╚══════════════════════════════════════════════════════╝

  🌡️  Conversión de Temperatura
  ─────────────────────────────────────────
  📥 Celsius     :     100.00 °C
  📤 Fahrenheit  :     212.00 °F
  📤 Kelvin      :     373.15 K
  ─────────────────────────────────────────

  🔄 Tabla de conversiones cruzadas:
     212.00°F = 100.00°C = 373.15K
     373.15K = 100.00°C = 212.00°F
```

```bash
go run main.go dist 42.195
```

**Salida esperada (maratón):**

```
╔══════════════════════════════════════════════════════╗
║   🔄 Conversor de Unidades Universal — Lab Go       ║
╚══════════════════════════════════════════════════════╝

  📏 Conversión de Distancia
  ─────────────────────────────────────────
  📥 Kilómetros  :      42.1950 km
  📤 Millas      :      26.2188 mi
  📤 Metros      :   42195.0000 m
  ─────────────────────────────────────────

  🔄 Tabla de conversiones cruzadas:
     42.1950km = 26.2188mi = 42195.0000m
     26.2188mi = 42.1950km = 42195.0000m
```

```bash
go run main.go data 1073741824
```

**Salida esperada (1 GB en bytes):**

```
╔══════════════════════════════════════════════════════╗
║   🔄 Conversor de Unidades Universal — Lab Go       ║
╚══════════════════════════════════════════════════════╝

  💾 Conversión de Almacenamiento
  ─────────────────────────────────────────
  📥 Bytes       :      1073741824 B
  📤 Kilobytes   :    1048576.0000 KB
  📤 Megabytes   :       1024.0000 MB
  📤 Gigabytes   :          1.0000 GB
  ─────────────────────────────────────────

  🔄 Tabla de conversiones cruzadas:
     1073741824B = 1048576.0000KB = 1024.0000MB = 1.000000GB
```

### 🔨 Compilar el binario permanente

```bash
go build -o converter main.go
# En Windows: converter.exe

# Ejecutarlo directamente
./converter temp 37
# En Windows: converter.exe temp 37
```

> 💡 **Experimento:** Intenta ejecutar `go run main.go temp abc`. ¿Qué pasa? El programa debería mostrar un error amigable porque `strconv.ParseFloat` falla al intentar convertir `"abc"` a número. Esto es el **patrón `(resultado, error)`** en acción.

---

## 🧠 8. Conceptos Clave Resumidos

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  📦 var x int = 42                                                  │
│     → Declaración explícita de variable                             │
│     → Se puede usar fuera de funciones                              │
│                                                                     │
│  🏷️ x := 42                                                         │
│     → Declaración corta con inferencia de tipo                      │
│     → Solo dentro de funciones                                      │
│     → Es la forma preferida en el 90% de los casos                  │
│                                                                     │
│  🔢 Tipos primitivos:                                               │
│     int, float64, string, bool                                      │
│     → int por defecto para enteros                                  │
│     → float64 por defecto para decimales                            │
│     → strings son inmutables y UTF-8                                │
│     → bool NO se puede usar como número                             │
│                                                                     │
│  🏷️ Zero values                                                     │
│     → int → 0, float → 0.0, string → "", bool → false              │
│     → Toda variable tiene un valor válido siempre                   │
│     → Nunca hay null/undefined para tipos primitivos                │
│                                                                     │
│  🔄 Conversiones explícitas                                         │
│     → float64(intVar) — no hay casting automático                   │
│     → strconv.Atoi("42") — string a int                            │
│     → strconv.ParseFloat("3.14", 64) — string a float64            │
│     → fmt.Sprintf("%d", 42) — int a string                         │
│                                                                     │
│  🔢 const pi = 3.14                                                 │
│     → Constante — no puede cambiar después de declararla            │
│     → Se evalúa en tiempo de compilación                            │
│                                                                     │
│  🧩 Tipos compuestos (introducción):                                │
│     → [5]int  → Array fijo de 5 enteros                             │
│     → []int   → Slice dinámico (el que más usarás)                  │
│     → map[string]int → Diccionario (tabla hash)                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 📊 9. Comparación: Conversión de Tipos en Diferentes Lenguajes

| Escenario | Go | Python | JavaScript | Java |
|:----------|:---|:-------|:-----------|:-----|
| `int + float` | ❌ Error de compilación | ✅ Funciona (float) | ✅ Funciona (float) | ❌ Error (sin cast) |
| `int + string` | ❌ Error de compilación | Depende | `"5" + 3 = "53"` | ❌ Error (sin cast) |
| `bool = int` | ❌ Error de compilación | `1` es truthy | `1` es truthy | ❌ Error |
| `string → int` | `strconv.Atoi()` explícito | `int("42")` | `parseInt("42")` | `Integer.parseInt()` |
| `float → int` | `int(3.9)` → `3` (trunca) | `int(3.9)` → `3` | `Math.floor(3.9)` | `(int) 3.9` → `3` |
| **Filosofía** | **Tú decides siempre** | Lo que sea más fácil | Lo que sea más fácil | Contrato estricto |

> 🧠 **Observa el patrón:** Go es el único que te obliga a **declarar cada conversión explícitamente**. Ni Python, ni JavaScript, ni Java (totalmente) hacen esto. Es más código, pero cero sorpresas en producción.

---

## 🧩 Ejercicio Feynman

> **El reto final de esta lección:** Go tiene tipos estáticos como Java, pero la gente dice que es más fácil. ¿Por qué?

### Tu misión: explícalo con tus propias palabras

Imagina que le explicas a un desarrollador de JavaScript por qué los tipos de Go son una **ventaja**, no una carga.

### Las 5 preguntas que debes responder:

1. **¿Qué significa que Go tenga inferencia de tipos (`:=`)?** — ¿Por qué no necesitas escribir `var x int = 42` si Go ya sabe que `42` es un `int`?

2. **¿Por qué Go NO permite conversiones implícitas (`int` a `float64`)?** — ¿Qué bug previene esto? ¿Puedes imaginar un escenario real donde una conversión implícita causa un desastre?

3. **¿Qué pasaría si el compilador te dejara sumar un `string` con un `int`?** — Imagina el escenario: tienes una variable `precio` que en algún punto del código se convierte en string `"10"` por error. Ahora `precio * 3` da `"101010"` en vez de `30`. Dibuja mentalmente este escenario y explica por qué es peor que un error de compilación.

4. **¿Qué es un zero value y por qué es mejor que `null`/`undefined`?** — Explica el concepto de que en Go, **toda variable siempre tiene un valor válido**. ¿Cuántos bugs de `NullPointerException` has visto en Java? ¿Cuántos `undefined is not a function` en JavaScript?

5. **¿Por qué `var x float64 = 3` es diferente de `x := 3`?** — Si no puedes responder esto con claridad, necesitas repasar la sección de inferencia de tipos. La respuesta tiene que ver con qué tipo asigna Go al literal `3` vs `3.0`.

### 📝 Autoevaluación

| Criterio | ✅ | ❌ |
|:---------|:---|:---|
| Puedes explicar la diferencia entre `var`, `:=` y `const` | | |
| Puedes nombrar los 4 tipos primitivos principales de Go | | |
| Puedes explicar qué es un zero value y cuál es el de cada tipo | | |
| Puedes convertir un string a int con `strconv.Atoi` sin mirar documentación | | |
| Puedes explicar por qué Go prohíbe la suma de `int` con `float64` | | |
| Puedes explicar la diferencia entre `float64(3)` y `3.0` | | |
| Puedes explicar por qué `string(65)` no produce `"65"` | | |
| Puedes explicar por qué los zero values eliminan la necesidad de `null` | | |
| Puedes ejecutar el conversor de unidades con 3 tipos diferentes | | |

> 🎯 **Si marcaste algún ❌:** Vuelve a la sección correspondiente y léela de nuevo. El Método Feynman funciona así: descubrir que no puedes explicar algo es **exactamente** el momento donde el aprendizaje real comienza.

---

## 🔮 ¿Qué viene en la Lección 03?

> **Control de Flujo: `if`, `switch`, `for` y la Elegancia del Minimalismo** — Go solo tiene `for` como bucle (no hay `while`, no hay `do-while`). Entenderás por qué esta decisión radical de diseño hace el código más legible, y construirás un analizador de logs del sistema.

---

<div align="center">

### *"El código se lee mucho más de lo que se escribe. Siendo así, la legibilidad importa mucho."*
### — **Rob Pike**, co-creador de Go

<br>

**¡Lección 02 completada! Avanza a la Lección 03 🚀**

</div>