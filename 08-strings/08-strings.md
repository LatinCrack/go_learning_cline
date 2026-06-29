# 🧪 Lección 08 — Strings, Runes y el Universo Unicode de Go

```
╔═══════════════════════════════════════════════════════════════════╗
║   🧪 LABORATORIO DE GO — LECCIÓN 08                              ║
║   Strings, Runes y el Universo Unicode de Go                      ║
╚═══════════════════════════════════════════════════════════════════╝
```

## 📋 Índice de la lección

1. [¿Qué son los Strings en Go?](#1-qué-son-los-strings-en-go)
2. [La analogía del tren de vagones](#2-la-analogía-del-tren-de-vagones)
3. [UTF-8: el alfabeto universal](#3-utf-8-el-alfabeto-universal)
4. [El tipo Rune: el verdadero carácter](#4-el-tipo-rune-el-verdadero-carácter)
5. [len() vs []rune(): metros vs pasajeros](#5-len-vs-rune-metros-vs-pasajeros)
6. [Iterar sobre strings: bytes vs runes](#6-iterar-sobre-strings-bytes-vs-runes)
7. [Operaciones comunes con strings](#7-operaciones-comunes-con-strings)
8. [Paquete strings: la navaja suiza](#8-paquete-strings-la-navaja-suiza)
9. [Paquete strconv: conversión de tipos](#9-paquete-strconv-conversión-de-tipos)
10. [Emojis y los grapheme clusters](#10-emojis-y-los-grapheme-clusters)
11. [Ejercicio práctico: Truncador Unicode-Safe y Analizador de Emojis](#11-ejercicio-práctico-truncador-unicode-safe-y-analizador-de-emojis)
12. [Ejercicio Feynman](#12-ejercicio-feynman)

---

## 1. ¿Qué son los Strings en Go?

Un **string** en Go es una secuencia de **bytes** codificada en **UTF-8**. Punto. No es un array de caracteres, no es una lista de letras — es un slice de bytes inmutable.

> 💡 **Concepto clave:** En Go, un string es **inmutable**. Una vez creado, no puedes modificar sus bytes individuales. Si necesitas transformarlo, creas uno nuevo.

### ¿Qué significa "inmutable"?

```go
s := "Hola"
// s[0] = 'h'  ← ❌ COMPILATION ERROR: cannot assign to s[0]
```

Para "modificar" un string, necesitas crear uno nuevo:

```go
s := "Hola"
nuevo := "h" + s[1:]  // "hola" ← nuevo string, el original sigue intacto
```

### ¿Por qué inmutable?

| Ventaja | Explicación |
|---|---|
| **Seguridad** | Puedes compartir un string entre goroutines sin locks |
| **Eficiencia** | Go puede reusar memoria sabiendo que nadie modificará el contenido |
| **Simplicidad** | No hay "efectos colaterales" al pasar strings a funciones |

---

## 2. La analogía del tren de vagones

Imagina que un **string** es un **tren de pasajeros** que viaja por Europa.

### El tren (el string)

El tren tiene vagones dispuestos en orden lineal. Cada vagón ocupa un espacio físico medido en **metros** (bytes).

### Los pasajeros (los caracteres visibles)

Cada vagón puede ser de tamaño diferente según el pasajero que lleve:

```
Tren: "Hola"
┌──────┬──────┬──────┬──────┐
│  H   │  o   │  l   │  a   │
│ 1m   │ 1m   │ 1m   │ 1m   │  ← 4 metros (bytes), 4 pasajeros (runes)
└──────┴──────┴──────┴──────┘

Tren: "café"
┌──────┬──────┬──────┬──────┬──────┐
│  c   │  a   │  f   │  é    │
│ 1m   │ 1m   │ 1m   │ 2m    │     ← 5 metros (bytes), 4 pasajeros (runes)
└──────┴──────┴──────┴──────┴──────┘

Tren: "日本語"
┌──────┬──────┬──────┐
│  日  │  本  │  語  │
│ 3m   │ 3m   │ 3m   │        ← 9 metros (bytes), 3 pasajeros (runes)
└──────┴──────┴──────┘
```

> 🔑 **La clave:** `len()` cuenta los **metros** (bytes), no los **pasajeros** (runes). Si quieres contar pasajeros, necesitas otra forma de medir.

### ¿Por qué los vagones son de diferente tamaño?

Porque Go usa codificación **UTF-8**, donde cada carácter puede ocupar entre **1 y 4 bytes**:

| Caracter | Bytes | Rango Unicode |
|---|---|---|
| `A`, `z`, `0` | **1 byte** | ASCII (0-127) |
| `ñ`, `é`, `ü` | **2 bytes** | Latino extendido |
| `日`, `本`, `語` | **3 bytes** | CJK (chino, japonés, coreano) |
| `😀`, `👨‍👩‍👧‍👦` | **4 bytes** | Emojis y símbolos especiales |

---

## 3. UTF-8: el alfabeto universal

### ¿Qué es UTF-8?

UTF-8 es un sistema de codificación que asigna un **número único** (punto de código) a cada carácter de cualquier idioma del mundo, y lo guarda como una secuencia de bytes.

> 💡 **Analogía:** Piensa en UTF-8 como un sistema de direcciones postal universal. Cada "casa" (carácter) en el mundo tiene una dirección única (punto de código), y esa dirección puede escribirse con 1, 2, 3 o 4 dígitos (bytes).

### ¿Por qué Go eligió UTF-8?

Porque UTF-8 es **retrocompatible con ASCII**: los primeros 128 caracteres (letras inglesas, números, símbolos básicos) ocupan 1 byte, igual que en ASCII. Esto significa que:

```go
s := "Hello"  // 5 bytes — exactamente igual que ASCII
s := "Hola"   // 4 bytes — exactamente igual que ASCII
s := "Hölä"   // 6 bytes — la ö y la ä ocupan 2 bytes cada una
```

### La trampa mortal: `len()` no cuenta lo que crees

```go
fmt.Println(len("Hello"))   // 5 ✅ (5 letras = 5 bytes)
fmt.Println(len("café"))    // 5 ❌ (4 letras = 5 bytes, é = 2 bytes)
fmt.Println(len("日本語"))    // 9 ❌ (3 letras = 9 bytes, cada una = 3 bytes)
fmt.Println(len("😀"))       // 4 ❌ (1 emoji = 4 bytes)
```

> ⚠️ **Trampa:** Si usas `len()` para truncar texto, vas a cortar caracteres a la mitad y producir basura. Esto es uno de los bugs más comunes en software que maneja texto internacional.

---

## 4. El tipo Rune: el verdadero carácter

### ¿Qué es un rune?

Un **rune** es el tipo de Go para representar un **punto de código Unicode** (un carácter). Es un alias de `int32`.

```go
var r rune = 'ñ'       // El punto de código Unicode de ñ
var r2 rune = '日'      // El punto de código Unicode de 日
var r3 rune = '😀'      // El punto de código Unicode de 😀
```

> 💡 **Analogía:** Si un string es un tren, un **rune** es un **pasajero individual**. No importa cuántos metros ocupe su vagón — cada persona cuenta como 1.

### Rune es int32

```go
r := 'A'
fmt.Printf("Valor: %d\n", r)      // 65 (punto de código ASCII)
fmt.Printf("Hex: %x\n", r)        // 41
fmt.Printf("Char: %c\n", r)       // A
fmt.Printf("Tipo: %T\n", r)       // int32
```

### ¿Por qué int32?

Porque el espacio Unicode tiene más de **1.1 millones** de puntos de código posibles (aunque solo se usan ~150,000 actualmente). `int32` puede representar hasta ~2.1 mil millones, más que suficiente.

### Char en comillas simples vs string en comillas dobles

```go
r := 'A'       // rune (int32) — un solo carácter
s := "A"       // string — secuencia de 1 byte

// ¿Cuál es la diferencia?
fmt.Printf("r: %T = %v\n", r, r)   // r: int32 = 65
fmt.Printf("s: %T = %v\n", s, s)   // s: string = A
```

> 🔑 **Regla:** Comillas simples `'` → rune. Comillas dobles `"` → string.

---

## 5. len() vs []rune(): metros vs pasajeros

### Método 1: `len()` — cuenta bytes (metros)

```go
s := "café"
fmt.Println(len(s))  // 5 (bytes)
```

### Método 2: Conversión a `[]rune` — cuenta caracteres (pasajeros)

```go
s := "café"
runes := []rune(s)
fmt.Println(len(runes))  // 4 (caracteres reales)
```

### Método 3: `utf8.RuneCountInString()` — la forma eficiente

```go
import "unicode/utf8"

s := "café"
fmt.Println(utf8.RuneCountInString(s))  // 4 (sin crear un slice nuevo)
```

### Comparación de métodos

| Método | Cuenta | Crea copia en memoria | Eficiencia |
|---|---|---|---|
| `len(s)` | Bytes | No | ⚡ Máxima |
| `len([]rune(s))` | Runes | **Sí** | 🐌 Crea slice nuevo |
| `utf8.RuneCountInString(s)` | Runes | **No** | ⚡ Eficiente |

> 💡 **Regla de oro:** Usa `utf8.RuneCountInString()` cuando solo necesitas contar caracteres. Usa `[]rune(s)` cuando necesitas **acceder** a caracteres individuales por índice.

---

## 6. Iterar sobre strings: bytes vs runes

### Iterar por bytes (con `for` clásico)

```go
s := "café"
for i := 0; i < len(s); i++ {
    fmt.Printf("Posición %d: byte %x = %c\n", i, s[i], s[i])
}
```

**Salida:**
```
Posición 0: byte 63 = c
Posición 1: byte 61 = a
Posición 2: byte 66 = f
Posición 3: byte c3 = Ã   ← ❌ ¡Basura! Parte de 'é' (UTF-8: c3 a9)
Posición 4: byte a9 = ©    ← ❌ ¡Basura! La otra parte de 'é'
```

> ⚠️ Iterar por bytes sobre texto con acentos produce **basura** porque cortas caracteres a la mitad.

### Iterar por runes (con `range`)

```go
s := "café"
for i, r := range s {
    fmt.Printf("Posición %d: rune %c (U+%04X)\n", i, r, r)
}
```

**Salida:**
```
Posición 0: rune c (U+0063)
Posición 1: rune a (U+0061)
Posición 2: rune f (U+0066)
Posición 3: rune é (U+00E9)
```

> 🔑 **Clave:** `range` sobre un string itera por **runes**, no por bytes. El primer valor es el **índice en bytes** (no el índice del rune).

### El índice en bytes es importante

```go
s := "café"
for i, r := range s {
    fmt.Printf("Byte index: %d, Rune: %c, Bytes que ocupa: %d\n",
        i, r, utf8.RuneLen(r))
}
```

**Salida:**
```
Byte index: 0, Rune: c, Bytes que ocupa: 1
Byte index: 1, Rune: a, Bytes que ocupa: 1
Byte index: 2, Rune: f, Bytes que ocupa: 1
Byte index: 3, Rune: é, Bytes que ocupa: 2
```

---

## 7. Operaciones comunes con strings

### Acceso por índice (bytes)

```go
s := "Hello"
b := s[1]         // byte: 101 (letra 'e')
fmt.Printf("%c\n", b)  // e
```

> ⚠️ `s[i]` devuelve un **byte**, no un rune. Solo funciona bien con texto ASCII puro.

### Slicing (substring)

```go
s := "Hello, World!"
sub := s[0:5]     // "Hello"
sub2 := s[7:]     // "World!"
```

> ⚠️ El slicing opera sobre **bytes**. Si cortas en medio de un carácter multi-byte, obtienes basura.

### Concatenación

```go
// Con +
s := "Hola" + " " + "Mundo"

// Con fmt.Sprintf (más legible para plantillas)
s := fmt.Sprintf("Hola %s, tienes %d años", nombre, edad)

// Con strings.Builder (eficiente para muchas concatenaciones)
var b strings.Builder
for i := 0; i < 1000; i++ {
    b.WriteString("item")
}
resultado := b.String()
```

### Comparación

```go
// Comparación directa
if s1 == s2 { ... }

// Case-insensitive (sin importar mayúsculas/minúsculas)
if strings.EqualFold("Go", "go") { ... }  // true

// Comparación lexicográfica
if strings.Compare("abc", "abd") < 0 { ... }  // true ("abc" < "abd")
```

### Conversión entre string y []rune

```go
// String → []rune
s := "café"
runes := []rune(s)    // ['c', 'a', 'f', 'é']

// []rune → string
s2 := string(runes)   // "café"

// Byte → string
b := byte(65)
s3 := string(b)       // "A" ← ¡CUIDADO! Esto NO convierte 65 al string "65"

// Número → string
n := 65
s4 := strconv.Itoa(n) // "65" ← Esta es la forma correcta
```

> ⚠️ **Trampa:** `string(65)` produce `"A"` (el carácter con punto de código 65), NO `"65"`. Para convertir un número a su representación en texto, usa `strconv.Itoa()`.

---

## 8. Paquete strings: la navaja suiza

El paquete `strings` tiene las herramientas más usadas para manipular texto:

### Búsqueda

```go
strings.Contains("pepperoni", "oni")      // true
strings.HasPrefix("Hello", "He")           // true
strings.HasSuffix("Hello", "lo")           // true
strings.Index("Hello", "ll")              // 2 (posición)
strings.Count("banana", "a")              // 3
```

### Transformación

```go
strings.ToUpper("hola")                   // "HOLA"
strings.ToLower("HOLA")                   // "hola"
strings.TrimSpace("  hola  ")             // "hola"
strings.Replace("foo bar foo", "foo", "go", 1)  // "go bar foo"
strings.ReplaceAll("foo bar foo", "foo", "go")  // "go bar go"
```

### División y unión

```go
// Dividir
parts := strings.Split("a,b,c", ",")      // ["a", "b", "c"]
lines := strings.Split("line1\nline2", "\n") // ["line1", "line2"]

// Dividir con límite
parts := strings.SplitN("a,b,c,d", ",", 2)  // ["a", "b,c,d"]

// Unir
s := strings.Join([]string{"Hola", "Mundo"}, " ")  // "Hola Mundo"

// Dividir por espacios (ignora múltiples espacios)
words := strings.Fields("  hola   mundo  ")  // ["hola", "mundo"]
```

### Repetición y construcción

```go
strings.Repeat("=", 20)    // "===================="
strings.Repeat("na", 3)    // "nanana"
```

### Mapa rápido de funciones

| Función | Propósito | Ejemplo |
|---|---|---|
| `Contains` | ¿Contiene substring? | `Contains("pepper", "ep")` → `true` |
| `HasPrefix` | ¿Empieza con...? | `HasPrefix("Go", "G")` → `true` |
| `HasSuffix` | ¿Termina con...? | `HasSuffix("Go", "o")` → `true` |
| `Index` | Posición de substring | `Index("Hello", "ll")` → `2` |
| `Count` | Veces que aparece | `Count("aaa", "a")` → `3` |
| `ToUpper` | Mayúsculas | `ToUpper("go")` → `"GO"` |
| `ToLower` | Minúsculas | `ToLower("GO")` → `"go"` |
| `TrimSpace` | Eliminar espacios | `TrimSpace(" hi ")` → `"hi"` |
| `Split` | Dividir string | `Split("a,b", ",")` → `["a","b"]` |
| `Join` | Unir slice | `Join([]{"a","b"}, "-")` → `"a-b"` |
| `Replace` | Reemplazar | `Replace("aba", "a", "o", 1)` → `"oba"` |
| `Fields` | Dividir por whitespace | `Fields(" a b ")` → `["a","b"]` |

---

## 9. Paquete strconv: conversión de tipos

### String ↔ Número

```go
// String → Int
n, err := strconv.Atoi("42")         // n = 42
n, err := strconv.ParseInt("FF", 16, 64)  // n = 255 (base hexadecimal)

// Int → String
s := strconv.Itoa(42)                 // "42"
s := strconv.FormatInt(255, 16)       // "ff" (base hexadecimal)

// String → Float
f, err := strconv.ParseFloat("3.14", 64)  // f = 3.14

// Float → String
s := strconv.FormatFloat(3.14, 'f', 2, 64)  // "3.14"
s := strconv.FormatFloat(3.14, 'E', -1, 64) // "3.14E+00"
```

### String ↔ Bool

```go
b, err := strconv.ParseBool("true")   // true
b, err := strconv.ParseBool("1")      // true
b, err := strconv.ParseBool("yes")    // error (no es formato válido)

s := strconv.FormatBool(true)         // "true"
```

### Quotes y escapado

```go
// Agregar comillas y escapar caracteres especiales
s := strconv.Quote("He said \"hello\"")
fmt.Println(s)  // "He said \"hello\""

// Quitar comillas y desescapar
unquoted, err := strconv.Unquote(`"He said \"hello\""`)
fmt.Println(unquoted)  // He said "hello"
```

---

## 10. Emojis y los grapheme clusters

### El caso complicado de los emojis

Los emojis son donde Unicode se pone interesante (y peligroso):

```go
fmt.Println(len("😀"))       // 4 (1 emoji = 4 bytes en UTF-8)
fmt.Println(utf8.RuneCountInString("😀"))  // 1 (1 rune)
```

Hasta aquí, todo bien. Pero...

### Los emojis compuestos (familias, banderas, tonos de piel)

```go
// Bandera de Japón: 🇯🇵 = dos runes regionales
fmt.Println(utf8.RuneCountInString("🇯🇵"))  // 2 (dos runes: 🇯 + 🇵)

// Familia: 👨‍👩‍👧‍👦 = múltiples runes unidos por ZWJ
fmt.Println(utf8.RuneCountInString("👨‍👩‍👧‍👦"))  // 7 (7 runes!)
// 👨 + ZWJ + 👩 + ZWJ + 👧 + ZWJ + 👦
```

> 🔑 **Grapheme Cluster:** Lo que tú ves como "un emoji" puede ser múltiples runes unidos por caracteres invisibles (Zero Width Joiner = ZWJ). Esto significa que incluso contando runes, no siempre cuentas "lo que el usuario ve".

### ¿Cómo contar lo que el usuario ve?

Para contar grapheme clusters reales, necesitas una librería externa como `golang.org/x/text/unicode/norm` o `github.com/rivo/uniseg`:

```go
import "github.com/rivo/uniseg"

s := "👨‍👩‍👧‍👦"
g := uniseg.NewGraphemes(s)
count := 0
for g.Next() {
    count++
}
fmt.Println(count)  // 1 ✅ Un solo grapheme visible
```

> ⚠️ Para aplicaciones en producción que manejan emojis (redes sociales, mensajería, etc.), siempre usa una librería de grapheme clusters.

### La tabla de realidad

| Emoji | Bytes | Runes | Graphemes (lo que ves) |
|---|---|---|---|
| `A` | 1 | 1 | 1 |
| `ñ` | 2 | 1 | 1 |
| `日` | 3 | 1 | 1 |
| `😀` | 4 | 1 | 1 |
| `🇯🇵` | 8 | 2 | 1 |
| `👨‍👩‍👧‍👦` | 25 | 7 | 1 |
| `👩🏽‍💻` | 13 | 4 | 1 |

---

## 11. Ejercicio práctico: Truncador Unicode-Safe y Analizador de Emojis

El ejercicio de esta lección es una **herramienta de análisis de texto Unicode** que:
1. Trunca textos a N caracteres visibles sin romper emojis ni acentos
2. Cuenta caracteres visibles (runes) vs bytes
3. Detecta y extrae emojis de un texto
4. Invierte strings respetando Unicode (no al revés de bytes)
5. Genera un reporte estadístico del texto

### Ejecución

```bash
cd 08-strings
go run main.go
```

### Conceptos aplicados en el ejercicio

| Concepto | Uso en el ejercicio |
|---|---|
| `len()` vs `utf8.RuneCountInString()` | Comparar bytes vs caracteres |
| `[]rune(s)` | Acceder a caracteres individuales |
| `range` sobre string | Iterar por runes correctamente |
| `unicode/utf8` | Decodificar runes manualmente |
| `unicode.Is()` | Detectar categorías Unicode (emoji, letra, dígito) |
| `strings.Builder` | Construir strings eficientemente |
| `strings.Fields()` | Dividir texto en palabras |
| `strings.Join()` | Unir partes truncadas |

---

## 12. Ejercicio Feynman

### 🎯 Tu misión: explicar strings y runes como si fueras un profesor

Usando solo palabras simples y analogías (sin jerga técnica), responde:

1. **¿Qué es un string en Go?** Explica con la analogía del tren. ¿Es un array de caracteres? ¿Por qué sí o por qué no?

2. **¿Por qué `len("café")` devuelve 5 y no 4?** Explica la diferencia entre contar metros (bytes) y contar pasajeros (runes) en el tren.

3. **¿Qué es un rune?** Si un string es un tren, ¿qué es un rune? ¿Por qué `rune` es un alias de `int32` y no de `byte`?

4. **¿Por qué `range` sobre un string se comporta diferente que acceder con `s[i]`?** Usa la analogía del tren: `range` te da pasajeros, `s[i]` te da metros.

5. **Los emojis son la prueba de fuego:** Explica por qué `len("👨‍👩‍👧‍👦")` da 25 y `utf8.RuneCountInString("👨‍👩‍👧‍👦")` da 7, pero tú solo ves UN emoji. ¿Qué está pasando "debajo del capó"?

### ✅ Criterio de éxito

Tu explicación debe ser comprensible para alguien que **nunca ha programado**. Si usas palabras como "byte array", "punto de código" o "encoding UTF-8", simplifica aún más. El objetivo es que el concepto quede tan claro que puedas explicarlo de memoria mientras tomas un café ☕.

> 🪶 **Test de Feynman:** Si no puedes explicar por qué `len("日本語")` devuelve 9 y no 3, sin usar la palabra "bytes", necesitas volver a leer la lección.

---

```
══════════════════════════════════════════════════════════════════
   ✅ Fin de la Lección 08 — Strings, Runes y el Universo Unicode
══════════════════════════════════════════════════════════════════