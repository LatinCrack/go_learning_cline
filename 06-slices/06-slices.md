<div align="center">

# 🧪 Laboratorio de Go — Lección 06

## **Arrays, Slices y el Secreto del Runtime de Go**

<br>

![Lección 06](https://img.shields.io/badge/Lección-06-4ECDC4?style=for-the-badge) ![Método Feynman](https://img.shields.io/badge/Método-Feynman-FF6B6B?style=for-the-badge) ![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)

<br>

> *"Un slice es como un visor de cámara: no copias la escena completa, solo encuadras una porción. Pero si alguien mueve un mueble dentro del encuadre, la escena original también cambia."*
> — **Principio de Diseño Go**

</div>

---

## 🧠 Primero: El Mapa Mental

Antes de escribir una sola línea de código, necesitas entender POR QUÉ Go tiene **dos** formas de manejar colecciones — arrays y slices — y por qué casi siempre usarás slices.

### 🎯 El Problema que Go Resolvió

En C, los arrays son punteros disfrazados. Pasas un array a una función y el puntero se copia, pero **los datos no**. Esto crea bugs silenciosos donde una función modifica datos que "no le pertenecen". En C++, `std::vector` resuelve esto pero con complejidad de templates. En Java, `ArrayList` es un objeto con métodos, no una primitiva del lenguaje.

Los creadores de Go (Rob Pike, Ken Thompson, Robert Griesemer) diseñaron una solución elegante con **dos capas**:

```
ARRAY (base fija, rígida)
│
│  → Contenedor de tamaño fijo, conocido en compilación
│  → Se pasa por VALOR (se copia completo)
│  → Comparable con == (si tienen el mismo tamaño)
│  → Útil para datos de tamaño conocido y fijo
│
└── SLICE (ventana dinámica, flexible)
    │
    │  → Descriptor de 3 campos: puntero + longitud + capacidad
    │  → Se pasa por REFERENCIA (comparte el array subyacente)
    │  → Crece automáticamente con append()
    │  → Es lo que usarás el 99% del tiempo
```

### 🏭 ¿Por qué separarlos?

| Característica | Array | Slice |
|:---------------|:------|:------|
| Tamaño | Fijo, conocido en compilación | Dinámico, puede cambiar en ejecución |
| Paso a funciones | Por **valor** (se copia todo) | Por **referencia** (se copia el descriptor) |
| Comparación | `==` funciona directamente | Solo con `slices.Equal()` (Go 1.21+) |
| Uso típico | Raros casos especiales | **Todo lo demás** |

**Regla simple:** Si no tienes una razón específica para usar un array, usa un slice.

---

## 📦 Arrays: El Bloque de Lego Fijo

> **Analogía:** Un array es como un **tren de vagones soldados**. Tiene exactamente N vagones — no puedes agregar ni quitar. Si le pasas el tren a otra persona, le entregas una **réplica exacta** (por valor). Lo que haga con su réplica no afecta al tuyo.

### 📝 Declaración

```go
// Array: el tamaño es PARTE del tipo
var temperaturas [7]float64         // 7 días de la semana
notas := [4]float64{8.5, 7.2, 9.1, 6.8}  // Inicialización literal

// El compilador cuenta los elementos con [...]
colores := [...]string{"rojo", "verde", "azul"}  // [3]string
```

### 🔑 Lo que necesitas saber sobre arrays

```go
// 1. El tamaño es parte del tipo: [3]int ≠ [4]int
var a [3]int
var b [4]int
// a = b  ← ❌ ERROR: tipos diferentes

// 2. Se pasan por VALOR (se copian completamente)
func modificar(arr [3]int) {
    arr[0] = 999  // Modifica la COPIA, no el original
}

nums := [3]int{1, 2, 3}
modificar(nums)
fmt.Println(nums) // [1, 2, 3] ← ¡INTACTO!

// 3. Son comparables (si tienen el mismo tipo/tamaño)
a1 := [3]int{1, 2, 3}
a2 := [3]int{1, 2, 3}
fmt.Println(a1 == a2) // true

// 4. Se pueden iterar con range
for i, v := range notas {
    fmt.Printf("Índice %d: %.1f\n", i, v)
}
```

### 🚫 ¿Cuándo usar un array?

Solo cuando:
1. El tamaño es **absolutamente fijo** y conocido en compilación (ej: coordenadas RGB = 3 valores).
2. Necesitas comparar dos colecciones con `==`.
3. Estás trabajando con código C interop (cgo).

Para TODO lo demás, usa slices.

---

## 🔍 Slices: El Secreto Mejor Guardado del Runtime de Go

> **Analogía:** Imagina que tienes una **hoja de cálculo** con 1000 filas. Un slice no copia las filas — es como un **filtro de vista** que dice: "muéstrame las filas 10 a 50". El filtro tiene tres datos: dónde empieza (puntero), cuántas filas muestra (longitud), y cuántas filas cabrían antes de necesitar más espacio (capacidad).

### 🏗️ La Estructura Interna: El Slice Header

Un slice es un **descriptor** de 24 bytes (en sistemas de 64 bits) que contiene exactamente 3 campos:

```
┌─────────────────────────────────────────────────────────┐
│                    SLICE HEADER                          │
│                                                          │
│   ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│   │ puntero  │  │   len    │  │   cap    │              │
│   │ (8 bytes)│  │ (8 bytes)│  │ (8 bytes)│              │
│   └────┬─────┘  └──────────┘  └──────────┘              │
│        │                                                 │
│        ▼                                                 │
│   ┌────┬────┬────┬────┬────┬────┬────┐                  │
│   │ 10 │ 20 │ 30 │ 40 │ 50 │ 60 │    │ ← array         │
│   └────┴────┴────┴────┴────┴────┴────┘   subyacente     │
│        ▲                                                 │
│        │                                                 │
│        └── puntero apunta aquí (después de copiar)       │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

**Lo crucial:** Cuando pasas un slice a una función, se copia el **slice header** (24 bytes), NO el array subyacente. Ambos headers apuntan al **mismo array** en memoria.

### 📝 Tres Formas de Crear un Slice

```go
// Forma 1: Literal — la más común
frutas := []string{"manzana", "banana", "cereza"}
// len=3, cap=3

// Forma 2: make([]tipo, len, cap)
s := make([]int, 3, 10)
// len=3 (3 ceros), cap=10 (espacio para 10)

// Forma 3: make sin cap (cap = len)
s2 := make([]int, 5)
// len=5, cap=5
```

### 🎯 ¿Cuál usar?

| Forma | Cuándo usarla |
|:------|:--------------|
| **Literal** | Cuando conoces los valores iniciales. |
| **make con cap** | Cuando sabes aproximadamente cuántos elementos necesitas (evita reallocations). |
| **make sin cap** | Cuando solo necesitas un slice del tamaño exacto. |

---

## ✂️ Slicing: Crear Sub-slices (y el Peligro del Aliasing)

> **Analogía:** El slicing es como poner un **marco fotográfico** sobre una pintura grande. El marco no copia la pintura — solo muestra una porción. Si alguien pinta un bigote en la porción visible, la pintura original TAMBIÉN se modifica. Esto es el **aliasing de memoria**.

### 📝 Sintaxis del Slicing

```go
original := []int{10, 20, 30, 40, 50, 60}

// Slicing: original[inicio:fin]  (fin es exclusivo)
sub := original[1:4]  // Elementos en índice 1, 2, 3 → [20, 30, 40]

fmt.Println(sub)          // [20, 30, 40]
fmt.Println(len(sub))     // 3  (4 - 1)
fmt.Println(cap(sub))     // 5  (6 - 1) ← ¡Importante!
```

### ⚠️ El Aliasing: Cuando Modificar Uno Modifica al Otro

```go
original := []int{10, 20, 30, 40, 50, 60}
sub := original[1:4]  // [20, 30, 40]

sub[0] = 999  // Modifica el array subyacente COMPARTIDO

fmt.Println(original) // [10, 999, 30, 40, 50, 60] ← ¡CAMBIÓ!
fmt.Println(sub)       // [999, 30, 40]
```

**¿Por qué?** Porque `sub` y `original` comparten el mismo array subyacente. `sub[0]` apunta a `original[1]` en memoria.

### ✅ La Solución: `copy()`

```go
original := []int{10, 20, 30, 40, 50, 60}
sub := original[1:4]  // [20, 30, 40]

// copy crea un NUEVO array independiente
copia := make([]int, len(sub))
copy(copia, sub)

copia[0] = 111

fmt.Println(original) // [10, 20, 30, 40, 50, 60] ← ¡INTACTO!
fmt.Println(sub)       // [20, 30, 40]              ← ¡INTACTO!
fmt.Println(copia)     // [111, 30, 40]              ← Solo la copia cambió
```

### 🧩 Diagrama del Aliasing

```
ALIASING (peligroso):               COPY (seguro):

original ──┐                         original ──→ [10, 20, 30, 40, 50, 60]
            │                              sub ──→ [20, 30, 40]
sub ────────┤                         
            ▼                         copia ───→ [20, 30, 40]  ← array propio
[10, 20, 30, 40, 50, 60]

sub[0]=999 → modifica AMBOS          copia[0]=111 → solo copia cambia
```

---

## 🚀 `append`: El Motor del Crecimiento

> **Analogía:** `append` es como un **mudador inteligente**. Si tu departamento (array subyacente) tiene espacio para más muebles, simplemente los coloca. Si no cabe, consigue un departamento MÁS GRANDE, mueve TODO, y coloca el nuevo mueble. El problema: tu "dirección" (puntero del slice) puede cambiar.

### 📝 Cómo funciona `append`

```go
s := make([]int, 0, 3)  // len=0, cap=3

s = append(s, 1)   // len=1, cap=3  → cabe
s = append(s, 2)   // len=2, cap=3  → cabe
s = append(s, 3)   // len=3, cap=3  → ¡lleno!
s = append(s, 4)   // len=4, cap=6  → ¡nuevo array! (cap se duplica)
```

### 📈 El Patrón de Crecimiento del Runtime de Go

El runtime de Go usa una **estrategia de crecimiento adaptativo**:

```
Capacidad inicial    →    Crecimiento
─────────────────────────────────────────
     0 a 1024       →    se duplica (×2)
   1024 a ...       →    crece ~25% (×1.25)
```

Esto significa que `append` amortiza el costo de copiar: no copia en cada llamada, solo cuando se agota la capacidad.

### ⚠️ La Regla de Oro de `append`

```go
// ❌ PELIGROSO: olvidar reasignar
append(s, elemento)  // Devuelve un NUEVO slice, pero lo ignoras

// ✅ CORRECTO: siempre reasignar
s = append(s, elemento)  // Guardas el resultado
```

**¿Por qué?** Si `append` necesita crear un nuevo array (porque el actual está lleno), devuelve un slice header con un puntero DIFERENTE. Si no reasignas, sigues apuntando al viejo array.

---

## 🔥 El Peligro Oculto: Aliasing + append

Este es uno de los bugs más sutiles y difíciles de detectar en Go:

### 🐛 El Bug

```go
original := []int{1, 2, 3, 4, 5}
sub := original[1:3]  // [2, 3], cap=4 ← ¡Tiene espacio extra!

// append a sub: como cap=4 > len=2, NO crea nuevo array
sub = append(sub, 99)

fmt.Println(original) // [1, 2, 3, 99, 5] ← ¡SILENCIOSAMENTE MODIFICADO!
fmt.Println(sub)       // [2, 3, 99]
```

**¿Qué pasó?** `sub` tenía capacidad para 4 elementos (original[1:] tiene 5 posiciones). `append` usó esa posición libre para colocar `99`, pero esa posición **pertenece a `original`**.

### ✅ La Solución: Full Slice Expression

```go
original := []int{1, 2, 3, 4, 5}

// [inicio:fin:max] limita la CAPACIDAD
sub := original[1:3:3]  // len=2, cap=2 (limitado a posición 3)

sub = append(sub, 99)  // cap agotada → NUEVO array independiente

fmt.Println(original) // [1, 2, 3, 4, 5] ← ¡INTACTO!
fmt.Println(sub)       // [2, 3, 99]       ← array propio
```

**`[1:3:3]` significa:** "desde índice 1 hasta índice 3 (exclusivo), con capacidad máxima en índice 3". Esto **cierra la ventana** para que `append` no pueda usar espacio de `original`.

---

## 🧩 Operaciones Fundamentales con Slices

### 📝 Insertar un Elemento en Posición Específica

```go
numeros := []int{1, 2, 4, 5}

// Insertar 3 en posición 2
// Paso 1: Abrir espacio moviendo elementos a la derecha
numeros = append(numeros[:2+1], numeros[2:]...)
// Paso 2: Colocar el valor
numeros[2] = 3

fmt.Println(numeros) // [1, 2, 3, 4, 5]
```

### 📝 Eliminar un Elemento (manteniendo orden)

```go
letras := []string{"a", "b", "c", "d", "e"}

// Eliminar elemento en índice 1 ("b")
letras = append(letras[:1], letras[2:]...)

fmt.Println(letras) // ["a", "c", "d", "e"]
```

### 📝 Eliminar sin Mantener Orden (O(1))

```go
items := []string{"x", "y", "z", "w", "v"}

// Eliminar "y" (índice 1) sin mantener orden
items[1] = items[len(items)-1]  // Reemplazar con el último
items = items[:len(items)-1]    // Acortar

fmt.Println(items) // ["x", "v", "z", "w"]
```

### 📝 Filtrar In-Place (sin allocations)

```go
valores := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

n := 0
for _, v := range valores {
    if v%2 == 0 {  // Solo pares
        valores[n] = v
        n++
    }
}
valores = valores[:n]

fmt.Println(valores) // [2, 4, 6, 8, 10]
```

### 📝 Revertir un Slice In-Place

```go
abc := []string{"A", "B", "C", "D", "E"}

for i, j := 0, len(abc)-1; i < j; i, j = i+1, j-1 {
    abc[i], abc[j] = abc[j], abc[i]
}

fmt.Println(abc) // ["E", "D", "C", "B", "A"]
```

---

## 🏗️ Slices con `maps`: Agrupación Dinámica

> **Analogía:** Un `map[string][]Venta` es como un **archivador de oficina**. Cada cajón (key) tiene una carpeta (slice) con documentos (elementos). Puedes agregar documentos a cualquier carpeta con `append`, y si el cajón no existe, lo creas automáticamente.

```go
// Agrupar ventas por categoría
porCategoria := make(map[string][]Venta)

for _, v := range ventas {
    porCategoria[v.Categoria] = append(porCategoria[v.Categoria], v)
}

// Acceder al grupo
electrónica := porCategoria["Electrónica"]
fmt.Printf("Hay %d ventas de Electrónica\n", len(electrónica))
```

**La línea mágica:** `porCategoria[v.Categoria] = append(porCategoria[v.Categoria], v)`

- Si la key no existe, `porCategoria[v.Categoria]` devuelve `nil`.
- `append(nil, v)` crea un nuevo slice con `v`.
- El resultado se asigna a la key, creando la entrada en el map.

Todo en una sola línea. Esto es **idiomático Go**.

---

## 📊 `len()`, `cap()` y el Cero Valor

```go
var s []int           // Slice nil

fmt.Println(s == nil) // true
fmt.Println(len(s))   // 0
fmt.Println(cap(s))   // 0

// append a nil slice funciona perfectamente
s = append(s, 1, 2, 3)
fmt.Println(s)        // [1, 2, 3]
fmt.Println(s == nil) // false
```

**Regla importante:** Un slice `nil` es completamente seguro para `len()`, `cap()`, `range` y `append`. Solo es peligroso hacer indexing directo (`s[0]`) en un slice nil o vacío.

---

## 🧩 Resumen Visual

```
┌───────────────────────────────────────────────────────────────────┐
│                                                                   │
│   ARRAY [3]int                      SLICE []int                   │
│                                                                   │
│   ┌───┬───┬───┐                     ┌─────────────┐              │
│   │ 1 │ 2 │ 3 │                     │ puntero ─────┼──┐          │
│   └───┴───┴───┘                     │ len = 3     │  │          │
│   Tamaño FIJO                       │ cap = 5     │  │          │
│   Se COPIA al pasar                 └─────────────┘  │          │
│   Comparable con ==                   │              │          │
│                                       ▼              │          │
│                                   ┌───┬───┬───┬───┬───┐        │
│                                   │ 1 │ 2 │ 3 │   │   │        │
│                                   └───┴───┴───┴───┴───┘        │
│                                    ▲         ▲     ▲           │
│                                    │         │     │           │
│                                  len=3     cap=5   libre       │
│                                                                   │
│   SLICING original[1:3]             APPEND                        │
│                                                                   │
│   ┌───┬───┬───┬───┬───┐            s = append(s, 4)             │
│   │   │ ↙ │ ↙ │   │   │            Si cap alcanzada:            │
│   └───┴───┴───┴───┴───┘              → NUEVO array              │
│         ▲     ▲                       → COPIA elementos         │
│         │     │                       → NUEVO puntero            │
│       sub[0] sub[1]                                               │
│       len=2  cap=4                                                │
│       (¡comparte array!)                                         │
│                                                                   │
│   FULL SLICE [1:3:3]               COPY                           │
│                                                                   │
│   cap limitada a 3                  copia := make([]int, len(s)) │
│   → append crea array propio        copy(copia, s)               │
│   → ¡SIN aliasing!                  → array INDEPENDIENTE        │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

---

## 🏋️ Ejercicio Práctico: Procesador de Datos CSV de Alta Velocidad

### 🎯 Objetivo

Construir un procesador de datos CSV que demuestre **TODOS** los conceptos de arrays y slices: creación, slicing, aliasing, copy, append, crecimiento de capacidad, full slice expression, y operaciones avanzadas (insertar, eliminar, filtrar in-place, revertir).

### 📁 Estructura del Proyecto

```
06-slices/
├── main.go     ← Código completo del ejercicio
└── go.mod      ← Módulo Go
```

### 📄 Código Línea por Línea

El archivo `main.go` está dividido en **5 secciones** que demuestran cada concepto progresivamente. A continuación, un recorrido exhaustivo por la lógica de cada parte:

---

#### **Sección 1: Arrays — Tamaño Fijo, Paso por Valor**

```go
func demostrarArrays() {
    notas := [4]float64{8.5, 7.2, 9.1, 6.8}
    fmt.Printf("Array de notas: %v\n", notas)
    fmt.Printf("Longitud (siempre fija): %d\n", len(notas))
}
```

> 📌 **Arrays en Go:** El tamaño (`[4]`) es parte del tipo. Esto significa que `[3]float64` y `[4]float64` son **tipos diferentes** y no se pueden comparar ni asignar entre sí.

```go
func modificarArray(arr [4]float64) {
    arr[0] = 99.9 // Modifica la COPIA, no el original
}
```

> 📌 **Paso por valor:** Cuando pasas un array a una función, Go copia **todos los elementos**. Para un array de 4 float64 (32 bytes), esto es barato. Para un array de 10,000 elementos, sería muy costoso — por eso existen los slices.

```go
a1 := [3]int{1, 2, 3}
a2 := [3]int{1, 2, 3}
fmt.Println(a1 == a2) // true — comparación directa
```

> 📌 **Comparabilidad:** Los arrays son comparables con `==` (si tienen el mismo tipo/elementos). Los slices NO son comparables directamente — solo con `slices.Equal()` desde Go 1.21.

---

#### **Sección 2: Slices — Ventanas Dinámicas**

```go
s := make([]int, 3, 5)
// len = 3 → 3 ceros inicializados
// cap = 5 → espacio para 5 sin realloc
```

> 📌 **`make` con capacidad:** Especificar la capacidad por separado es una **optimización**. Si sabes que vas a tener ~100 elementos, haz `make([]T, 0, 100)` para evitar las múltiples copias que `append` haría al crecer desde 0.

```go
numeros := []int{10, 20, 30, 40, 50, 60}
sub := numeros[1:4] // Elementos en índice 1, 2, 3
```

> 📌 **Slicing:** `numeros[1:4]` crea un nuevo slice header que apunta al **mismo array subyacente**, pero con puntero desplazado a `numeros[1]`, len=3, y cap=5 (6-1). Este es el origen del aliasing.

```go
sub[0] = 999
fmt.Printf("Original: %v ← ¡CAMBIÓ!\n", numeros)
```

> 📌 **Demostración de aliasing:** Modificar `sub[0]` cambia `numeros[1]` porque ambos apuntan a la misma posición en memoria. Este es el bug más común con slices en Go.

```go
copia := make([]int, len(sub))
copy(copia, sub)
```

> 📌 **`copy()`:** Crea un array subyacente **independiente**. `copy` necesita que el destino tenga al menos `len(src)` elementos. Por eso primero hacemos `make([]int, len(sub))`.

---

#### **Sección 3: append — El Motor del Crecimiento**

```go
s := make([]int, 0, 3)
for i := 1; i <= 10; i++ {
    s = append(s, i*10)
    fmt.Printf("append(%2d): len=%d, cap=%d\n", i*10, len(s), cap(s))
}
```

> 📌 **Observar el crecimiento:** La salida muestra el patrón de duplicación: cap pasa de 3→6→12. Después de 1024 elementos, el crecimiento se vuelve ~25% en vez de 100%. Esto es una optimización del runtime para no desperdiciar memoria con slices gigantes.

```go
original := []int{1, 2, 3, 4, 5}
sub := original[1:3] // [2, 3], cap=4

sub = append(sub, 99)
// sub tiene cap=4, len=2 → hay espacio → usa posición de original[3]
fmt.Println(original) // [1, 2, 3, 99, 5] ← ¡MODIFICADO!
```

> 📌 **El peligro de aliasing + append:** Este es el bug que más sorprende a los desarrolladores. `append` NO crea un nuevo array si hay espacio disponible en el array subyacente. Silenciosamente sobreescribe datos del slice "original".

```go
safe := original[1:3:3] // len=2, cap=2 (limitado)
safe = append(safe, 88)
// cap=2, len=2 → NO hay espacio → NUEVO array
fmt.Println(original) // [1, 2, 3, 4, 5] ← ¡INTACTO!
```

> 📌 **Full slice expression `[1:3:3]`:** El tercer número (`3`) limita la capacidad. Al limitar cap a 2 (igual que len), `append` se ve **forzado** a crear un nuevo array subyacente. Esto **rompe el aliasing** de forma segura.

---

#### **Sección 4: Procesador CSV — Slices en Acción**

**Definición de datos:**

```go
type Venta struct {
    Producto  string
    Categoria string
    Cantidad  int
    Precio    float64
    Region    string
    Fecha     string
}
```

> 📌 **Struct para datos tabulares:** Cada fila del CSV se modela como un struct. Un slice de structs (`[]Venta`) es la estructura más eficiente para procesamiento secuencial de datos en Go — mejor que un map o una interfaz.

**Filtrado con funciones de alto orden:**

```go
func filtrarVentas(ventas []Venta, filtro func(Venta) bool) []Venta {
    resultado := make([]Venta, 0)
    for _, v := range ventas {
        if filtro(v) {
            resultado = append(resultado, v)
        }
    }
    return resultado
}
```

> 📌 **Slices + funciones como parámetros:** La función `filtro` es un **closure** — una función anónima que se pasa como argumento. Esto permite reutilizar `filtrarVentas` con cualquier criterio sin duplicar código.

**Agrupación con maps de slices:**

```go
func agruparPorCampo(ventas []Venta, campo func(Venta) string) map[string][]Venta {
    grupos := make(map[string][]Venta)
    for _, v := range ventas {
        key := campo(v)
        grupos[key] = append(grupos[key], v)
    }
    return grupos
}
```

> 📌 **El patrón `map[K][]V`:** Es el patrón más usado en Go para agrupar datos. La línea `grupos[key] = append(grupos[key], v)` funciona porque: (1) si la key no existe, devuelve `nil`; (2) `append(nil, v)` crea un slice válido; (3) el resultado se asigna de vuelta al map.

**Estadísticas con `copy` para proteger datos:**

```go
sorted := make([]float64, len(valores))
copy(sorted, valores) // copy para no modificar el original
sort.Float64s(sorted)
```

> 📌 **`copy` para protección:** Antes de ordenar, copiamos el slice. `sort.Float64s` modifica **in-place** (sin crear nuevo slice). Si no copiáramos, el slice original del llamador quedaría alterado silenciosamente.

**Ordenamiento con `copy` de structs:**

```go
ventasCopia := make([]Venta, len(ventas))
copy(ventasCopia, ventas)
sort.Slice(ventasCopia, func(i, j int) bool { ... })
```

> 📌 **`copy` con slices de structs:** `copy` copia **elementos**, no punteros. Cada `Venta` en `ventasCopia` es una copia independiente. Modificar `ventasCopia[i]` no afecta a `ventas[i]`.

---

#### **Sección 5: Operaciones Avanzadas**

**Insertar en posición:**

```go
numeros = append(numeros[:2+1], numeros[2:]...)
numeros[2] = 3
```

> 📌 **Técnica de inserción:** `append(numeros[:3], numeros[2:]...)` "abre" un espacio en índice 2 desplazando elementos a la derecha. Los `...` expanden el slice como argumentos variádicos. Luego asignamos el valor en la posición vacía.

**Eliminar manteniendo orden:**

```go
letras = append(letras[:1], letras[2:]...)
```

> 📌 **Técnica de eliminación:** `letras[:1]` mantiene todo antes del índice 1, `letras[2:]` incluye todo después del índice 1. `append` los concatena, "saltando" el elemento en índice 1. **Advertencia:** esto modifica el array subyacente.

**Eliminar sin mantener orden (O(1)):**

```go
items[1] = items[len(items)-1]
items = items[:len(items)-1]
```

> 📌 **Eliminación O(1):** Reemplazar el elemento a eliminar con el último y acortar el slice. Es O(1) porque no mueve elementos. Solo funciona cuando el orden no importa.

**Filtrar in-place:**

```go
n := 0
for _, v := range valores {
    if v%2 == 0 {
        valores[n] = v
        n++
    }
}
valores = valores[:n]
```

> 📌 **Filtrado sin allocations:** Dos punteros (`range` y `n`) recorren el mismo slice. Los elementos que pasan el filtro se compactan al inicio. Al final, `valores[:n]` acorta el slice. Cero allocations, cero copias — máxima eficiencia.

---

### ▶️ Ejecución

```bash
cd 06-slices
go run main.go
```

**Salida esperada:**

```
╔════════════════════════════════════════════════════════════════╗
║   🧪 LABORATORIO DE GO — LECCIÓN 06                          ║
║   Arrays, Slices y el Secreto del Runtime de Go               ║
╚════════════════════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   1️⃣  Arrays: tamaño fijo, paso por valor
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📌 Arrays: tamaño fijo, paso por valor

   Array de notas: [8.5 7.2 9.1 6.8]
   Longitud (siempre fija): 4

   ⚠️  Los arrays se pasan por VALOR (se copian):
   Antes de modificar: [8.5 7.2 9.1 6.8]
   Después de modificar (original intacto): [8.5 7.2 9.1 6.8]

   [1 2 3] == [1 2 3] → true
   [1 2 3] == [1 2 4] → false

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   2️⃣  Slices: ventanas dinámicas sobre arrays
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📌 Slices: ventanas dinámicas con puntero, longitud y capacidad

   make([]int, 3, 5): [0 0 0]
   len = 3, cap = 5

   🔍 Un slice internamente tiene 3 componentes:
   ┌─────────────────────────────────────────┐
   │  puntero → apunta al array subyacente   │
   │  len     → cuántos elementos tiene       │
   │  cap     → cuántos puede tener           │
   └─────────────────────────────────────────┘

   ⚠️  Slicing y aliasing de memoria:
   Original:  [10 20 30 40 50 60] (len=6, cap=6)
   Sub-slice: [20 30 40] (len=3, cap=5)
   Después de sub[0]=999:
     Original:  [10 999 30 40 50 60] ← ¡CAMBIÓ!

   ✅ copy() para evitar aliasing:
   Copia modificada: [111 30 40]
   Original intacto: [10 999 30 40 50 60]

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   3️⃣  append: el motor del crecimiento de slices
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📌 append: crecimiento automático de slices

   Slice inicial: len=0, cap=3, []
   append(10): len=1, cap=3, [10]
   ...
   ⚠️  EL PELIGRO: aliasing + append
   original: [1 2 3 4 5] (cap=5)
   sub:      [2 3] (cap=4)
   Después de append(sub, 99):
     original: [1 2 3 99 5] ← ¡CAMBIÓ en posición 3!

   ✅ Solución: full slice expression [lo:hi:max]
   safe: [2 3] (len=2, cap=2)
   Después de append(safe, 88):
     original: [1 2 3 4 5] ← ¡INTACTO!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   4️⃣  Procesador de Datos CSV: slices en acción
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📊 20 ventas cargadas para procesamiento

   🔍 Filtrar ventas de Electrónica:
     → Laptop Pro          5 uds × $1200.00 = $  6000.00
     → Monitor 27         10 uds × $350.00  = $  3500.00
     → Laptop Air          8 uds × $950.00  = $  7600.00
     → Tablet 10          12 uds × $450.00  = $  5400.00
     → Laptop Gaming       3 uds × $2500.00 = $  7500.00
     → Monitor 24         18 uds × $220.00  = $  3960.00
     → Laptop Ultrabook    6 uds × $1400.00 = $  8400.00
   💰 Ingreso total Electrónica: $42360.00

   📈 Estadísticas por CATEGORÍA:
     GRUPO              TOTAL   PROMEDIO    MEDIANA    DESV.STD     N        MIN        MAX
     ───────────────────────────────────────────────────────────────────────────────────────
     Electrónica     42360.00    6051.43    6000.00     1799.79     7    3500.00    8400.00
     ...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   5️⃣  Operaciones avanzadas: insertar, eliminar, filtrar, revertir
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   🔧 Insertar elemento en posición 2:
     Antes:  [1 2 4 5]
     Después: [1 2 3 4 5]

   🔧 Eliminar elemento en posición 1:
     Antes:  [a b c d e]
     Después: [a c d e]

   🔧 Revertir slice in-place:
     Original: [A B C D E]
     Revertido: [E D C B A]

══════════════════════════════════════════════════════════════════
   ✅ Todos los conceptos de arrays y slices ejecutados correctamente
══════════════════════════════════════════════════════════════════
```

---

## 🏋️ Ejercicio Feynman

> **Instrucción:** Toma una hoja en blanco o abre un archivo de texto vacío. Sin consultar esta lección, intenta responder cada pregunta **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado.

### 📝 Preguntas para explicar desde cero:

1. **¿Qué es un slice?** Explica con una analogía que NO sea de ventanas, marcos ni visores de cámara. Crea la tuya propia.

2. **¿Por qué Go tiene arrays Y slices?** ¿Por qué no simplemente tener un solo tipo de colección dinámica como Python o JavaScript? ¿Qué ganas con el array como tipo base?

3. **Un amigo JavaScript te dice:** *"En JS todo es dinámico, no necesito preocuparme por tamaño ni capacidad."* Respóndele explicando por qué el modelo de Go (slice header + array subyacente) es más eficiente en memoria y rendimiento.

4. **Explica el aliasing de memoria con una analogía de la vida real** que NO involucre computadoras. ¿En qué situaciones cotidianas compartimos "espacio" y un cambio en una vista afecta a otra?

5. **¿Qué pasa cuando `append` se queda sin espacio?** Explica el proceso completo paso a paso: qué hace con el viejo array, qué crea, cómo copia, y por qué el slice header cambia.

6. **Un colega te dice:** *"Yo siempre hago `sub := original[i:j]` y nunca he tenido problemas."* ¿Por qué puede estar teniendo bugs silenciosos que no ha detectado aún? ¿Qué debería hacer en su lugar?

7. **Predice el resultado de este código SIN ejecutarlo:**
   ```go
   a := []int{1, 2, 3, 4, 5}
   b := a[1:3]
   c := a[1:3:3]
   b = append(b, 10)
   c = append(c, 20)
   fmt.Println(a)
   fmt.Println(b)
   fmt.Println(c)
   ```

### ✅ Criterio de autoevaluación:

| Criterio                                                          | ¿Lo lograste? |
|-------------------------------------------------------------------|:-------------:|
| Explicaste slices sin mirar la lección                            | ⬜ Sí / ⬜ No |
| Creaste tu propia analogía original                               | ⬜ Sí / ⬜ No |
| Entiendes la diferencia interna entre array y slice               | ⬜ Sí / ⬜ No |
| Puedes explicar el aliasing con una analogía cotidiana            | ⬜ Sí / ⬜ No |
| Entiendes cuándo `append` crea un nuevo array vs reutiliza        | ⬜ Sí / ⬜ No |
| Sabes cuándo y por qué usar `copy()`                              | ⬜ Sí / ⬜ No |
| Predijiste correctamente la salida del código del ejercicio 7     | ⬜ Sí / ⬜ No |

---

## 🗺️ Próxima lección

En la **Lección 07** exploraremos **Mapas y Manejo de Errores**: cómo Go gestiona colecciones clave-valor con `map`, el patrón `error` como valor (no excepciones), y por qué el manejo explícito de errores es una de las decisiones de diseño más inteligentes de Go.

> *"Un slice bien comprendido es como un bisturí en manos de un cirujano: preciso, eficiente, y peligroso si no sabes cómo funciona el filo."* — Principio Feynman