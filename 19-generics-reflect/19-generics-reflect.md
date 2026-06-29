# 🔮 Lección 19: Generics, Reflect y Metaprogramación en Go

## 🎯 Objetivo de la Lección

Dominar las dos herramientas más avanzadas de la caja de Go: **Generics** (type parameters), la feature más solicitada en la historia del lenguaje, que permite escribir código que funciona con cualquier tipo que cumpla ciertas reglas; y **Reflect**, el paquete que permite inspeccionar y manipular tipos en tiempo de ejecución — el cuchillo suizo que abre cualquier caja, pero que si lo usas mal te cortas.

---

## 🧠 Analogía Fundamental: La Fábrica de Herramientas Universales

Imagina que tienes una **fábrica de herramientas**. Hay tres niveles de flexibilidad:

### 🔧 Nivel 1: Herramientas Específicas (Sin Generics, Sin Reflect)

Cada herramienta solo funciona con un material. Tienes un martillo para madera, otro para metal, otro para plástico. Si mañana aparece un nuevo material, necesitas fabricar un martillo nuevo.

```go
// Sin generics: duplicar código para cada tipo
func SumarInts(nums []int) int {
    total := 0
    for _, n := range nums { total += n }
    return total
}

func SumarFloats(nums []float64) float64 {
    total := 0.0
    for _, n := range nums { total += n }
    return total
}
// 😫 Misma lógica, dos funciones diferentes
```

### 🧩 Nivel 2: Herramientas Genéricas (Generics)

Un martillo que acepta clavos de **cualquier material** siempre y cuando sean clavos (tengan la forma correcta). No necesitas un martillo por material — uno solo sirve para todos. La restricción (constraint) es: "debe ser un clavo" (debe ser un número).

```go
// Con generics: UNA función para CUALQUIER tipo numérico
func Sumar[T Number](nums []T) T {
    var total T
    for _, n := range nums { total += n }
    return total
}
// ✅ Funciona con int, float32, float64, int64...
```

### 🔮 Nivel 3: La Caja Negra (Reflect)

Una herramienta que acepta **cualquier cosa**, incluso cosas que no son clavos. Puede agarrar un tornillo, una moneda o una piedra e intentar hacer algo con ella. Es extremadamente poderosa pero peligrosa: si intentas clavar algo que no es un clavo, la herramienta no te avisa hasta que estás martillando.

```go
// Con reflect: funciona con LITERALMENTE cualquier cosa
func SumarReflect(v interface{}) float64 {
    val := reflect.ValueOf(v)
    // ¿Es un slice? ¿De qué tipo? ¿Tiene números?
    // No lo sabes hasta ejecutar el programa 😰
}
```

### 🔑 Los Tres Niveles Comparados

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│  🔧 Sin Generics        🧩 Con Generics        🔮 Reflect     │
│  (Específico)            (Genérico con reglas)   (Sin reglas)  │
│                                                                 │
│  ┌──────────┐           ┌──────────────┐       ┌────────────┐ │
│  │Martillo  │           │  Martillo    │       │ Herramienta│ │
│  │Madera    │           │  Universal   │       │  Mágica    │ │
│  └──────────┘           │              │       │            │ │
│  ┌──────────┐           │ "Acepta      │       │ "Acepta    │ │
│  │Martillo  │           │  cualquier   │       │  CUALQUIER │ │
│  │Metal     │           │  clavo"      │       │  cosa"     │ │
│  └──────────┘           │  [T Clavo]   │       │  [any]     │ │
│  ┌──────────┐           └──────────────┘       └────────────┘ │
│  │Martillo  │                                                   │
│  │Plástico  │           ✅ Seguro en compile    ⚠️ Seguro solo │
│  └──────────┘           ✅ Rápido                en runtime     │
│                         ✅ Autocompletado       ❌ Lento        │
│  😫 Duplicación         IDE funciona            ❌ Sin IDE      │
│  ❌ No escalable                                            │
└─────────────────────────────────────────────────────────────────┘
```

---

## 📦 Paquetes y Conceptos que Estudiaremos

| Concepto/Paquete | ¿Qué hace? | Analogía |
|------------------|-------------|----------|
| Type Parameters `[T any]` | Funciones y tipos genéricos | El **martillo universal** |
| Constraints `[T Number]` | Restricciones sobre tipos genéricos | "Solo acepta **clavos**" |
| `comparable` | Constraint built-in para `==`, `!=` | "Debe ser **comparable**" |
| `any` | Alias de `interface{}` | "Acepta **lo que sea**" |
| `reflect.TypeOf()` | Obtiene el tipo en runtime | El **detector de materiales** |
| `reflect.ValueOf()` | Obtiene el valor en runtime | El **brazo robótico** |
| Struct tags con reflect | Lee metadatos de campos | Las **etiquetas adhesivas** |

---

## 🧩 Generics — El Cambio de Era en Go

### ¿Por qué Go tardó tanto en tener Generics?

Go se lanzó en 2009 sin generics. Los creadores (Rob Pike, Ken Thompson, Robert Griesemer) querían un lenguaje **simple**. Los generics de C++ y Java son notoriamente complejos. Durante 12 años, la comunidad debatió: ¿cómo agregar generics sin convertir Go en C++?

En Go 1.18 (marzo 2022), finalmente llegaron. La solución es elegante: **constraints son interfaces**, y los type parameters usan corchetes `[T]` en vez de ángulos `<T>`.

### La Evolución del Problema

```
Go 1.0 - 1.17 (sin generics):
┌──────────────────────────────────────┐
│  interface{} + type assertion        │
│                                      │
│  func Max(a, b interface{}) interface{} { │
│      // ❌ Sin type safety           │
│      // ❌ Runtime panics             │
│      // ❌ Sin autocompletado         │
│  }                                   │
└──────────────────────────────────────┘

Go 1.18+ (con generics):
┌──────────────────────────────────────┐
│  Type parameters + constraints       │
│                                      │
│  func Max[T Ordered](a, b T) T {    │
│      // ✅ Type safety               │
│      // ✅ Compile-time checks        │
│      // ✅ Autocompletado funciona    │
│  }                                   │
└──────────────────────────────────────┘
```

### Sintaxis Básica de Type Parameters

```go
// Función genérica: T es el "type parameter"
func Primero[T any](slice []T) T {
    return slice[0]
}

// Llamada: el tipo se infiere automáticamente
Primero([]int{1, 2, 3})        // T = int, retorna 1
Primero([]string{"a", "b"})    // T = string, retorna "a"

// Llamada explícita (opcional pero permitida)
Primero[int]([]int{1, 2, 3})
```

**La analogía del sobre:** Piensa en `[T any]` como un sobre etiquetado "T". Cuando llamas a la función con `[]int`, Go pone una etiqueta "T = int" en el sobre. Dentro de la función, cada vez que ve "T", sabe que es "int".

### Los 3 Tipos de Constraints

#### 1. `any` — Acepta Cualquier Tipo

```go
func Imprimir[T any](valor T) {
    fmt.Println(valor)
}
// T puede ser int, string, bool, struct, slice, cualquier cosa
// Pero NO puedes hacer operaciones específicas: valor + valor ❌
```

`any` es como decir "acepta cualquier paquete, pero no sabes qué hay dentro". No puedes asumir nada sobre el contenido.

#### 2. `comparable` — Puede Compararse con `==` y `!=`

```go
func Contiene[T comparable](slice []T, objetivo T) bool {
    for _, v := range slice {
        if v == objetivo {
            return true
        }
    }
    return false
}

// Funciona:
Contiene([]int{1, 2, 3}, 2)         // true
Contiene([]string{"a", "b"}, "c")   // false

// NO funciona:
// Contiene([][]int{{1}, {2}}, []int{1})  // ❌ slices no son comparable
```

`comparable` es como decir "acepta cualquier paquete que pueda ponerse en una balanza". Solo los tipos que soportan `==` pasan.

#### 3. Constraints Personalizados — Tú Definir las Reglas

```go
// Un constraint es una INTERFACE con tipos en su lista
type Numero interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~float32 | ~float64
}

func Sumar[T Numero](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}

// El operador ~ significa "este tipo y todos los que lo tienen como base"
type Millas float64
Sumar([]Millas{1.5, 2.5, 3.0})  // ✅ Funciona porque ~float64 incluye Millas
```

**La analogía del club exclusivo:** `Numero` es como un club con lista de miembros. Solo pueden entrar `int`, `float64`, etc. El operador `~` dice "y cualquier tipo que se vista como uno de estos" (tipos definidos por el usuario con el mismo tipo subyacente).

### Diagrama de la Jerarquía de Constraints

```
                    interface{}
                        │
                ┌───────┴───────┐
                │               │
           comparable          any
           (==, !=)     (cualquier cosa)
                │               │
                │         ┌─────┴─────┐
                │         │           │
                │    Numeric       Ordered
                │  (aritmética)  (comparación)
                │         │           │
                │         └─────┬─────┘
                │               │
                │         ┌─────┴─────┐
                │         │           │
                │        int      float64
                │       int8      float32
                │       int16        ...
                │       int32
                │       int64
                │
          ┌─────┴─────┐
          │           │
         int        string
        float       bool
        struct      ...
```

### Funciones Genéricas Fundamentales

#### Map — Transformar Cada Elemento

```go
func Map[T any, U any](slice []T, fn func(T) U) []U {
    resultado := make([]U, len(slice))
    for i, v := range slice {
        resultado[i] = fn(v)
    }
    return resultado
}

// Uso:
nums := []int{1, 2, 3, 4, 5}
strings := Map(nums, strconv.Itoa)  // ["1", "2", "3", "4", "5"]

nombres := []string{"ana", "luis", "maria"}
mayusculas := Map(nombres, strings.ToUpper)  // ["ANA", "LUIS", "MARIA"]
```

#### Filter — Seleccionar Elementos que Cumplen una Condición

```go
func Filter[T any](slice []T, fn func(T) bool) []T {
    var resultado []T
    for _, v := range slice {
        if fn(v) {
            resultado = append(resultado, v)
        }
    }
    return resultado
}

// Uso:
nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
pares := Filter(nums, func(n int) bool { return n%2 == 0 })
// [2, 4, 6, 8, 10]
```

#### Reduce — Combinar Todos los Elementos en Uno

```go
func Reduce[T any, U any](slice []T, inicial U, fn func(U, T) U) U {
    acumulador := inicial
    for _, v := range slice {
        acumulador = fn(acumulador, v)
    }
    return acumulador
}

// Uso:
nums := []int{1, 2, 3, 4, 5}
suma := Reduce(nums, 0, func(acc, n int) int { return acc + n })
// 15

palabras := []string{"Go", "es", "genial"}
frase := Reduce(palabras, "", func(acc, s string) string {
    if acc == "" { return s }
    return acc + " " + s
})
// "Go es genial"
```

### Tipos Genéricos (No Solo Funciones)

```go
// Un Stack genérico
type Stack[T any] struct {
    elementos []T
}

func (s *Stack[T]) Push(valor T) {
    s.elementos = append(s.elementos, valor)
}

func (s *Stack[T]) Pop() (T, bool) {
    var zero T  // Zero value del tipo T
    if len(s.elementos) == 0 {
        return zero, false
    }
    ultimo := s.elementos[len(s.elementos)-1]
    s.elementos = s.elementos[:len(s.elementos)-1]
    return ultimo, true
}

func (s *Stack[T]) Len() int {
    return len(s.elementos)
}

// Uso:
stackInt := &Stack[int]{}
stackInt.Push(10)
stackInt.Push(20)
v, _ := stackInt.Pop()  // 20

stackStr := &Stack[string]{}
stackStr.Push("hello")
stackStr.Push("world")
v2, _ := stackStr.Pop()  // "world"
```

**La analogía del contenedor:** `Stack[T]` es como un contenedor de almacenamiento genérico. Puedes usar el mismo diseño de contenedor para guardar libros (`Stack[Libro]`), herramientas (`Stack[Herramienta]`) o comida (`Stack[Comida]`). La forma del contenedor no cambia, solo cambia lo que metes dentro.

### Restricciones Múltiples (Intersection de Interfaces)

```go
// Un constraint puede ser la intersección de múltiples interfaces
type SerializableNumber interface {
    Numero          // Debe ser un número
    fmt.Stringer    // Debe tener método String()
}

func Formatear[T SerializableNumber](nums []T) []string {
    resultado := make([]string, len(nums))
    for i, n := range nums {
        resultado[i] = n.String()  // ✅ Sabemos que tiene String()
    }
    return resultado
}
```

---

## 🔮 Reflect — El Espejo de la Realidad

### ¿Qué es Reflect?

`reflect` es el paquete que permite **inspeccionar tipos y valores en tiempo de ejecución**. Es como tener un escáner de rayos X que te dice exactamente qué hay dentro de cualquier variable.

```go
import "reflect"

x := 42
t := reflect.TypeOf(x)   // int
v := reflect.ValueOf(x)  // 42

fmt.Println(t)           // "int"
fmt.Println(v)           // "42"
fmt.Println(v.Kind())    // "int" (la categoría del tipo)
fmt.Println(v.Int())     // 42 (el valor como int64)
```

**La analogía del escáner:** `reflect.TypeOf()` es como un escáner que te dice "esto es una caja de madera". `reflect.ValueOf()` te abre la caja y te dice "dentro hay 42 manzanas". `Kind()` te dice la categoría general: "es fruta" (número entero).

### Las Dos Columnas de Reflect: Type y Value

```
┌──────────────────────┬──────────────────────┐
│   reflect.TypeOf()   │  reflect.ValueOf()   │
│   (¿QUÉ es?)         │  (¿CUÁNTO vale?)     │
├──────────────────────┼──────────────────────┤
│   .Name()            │  .Int()              │
│   .Kind()            │  .Float()            │
│   .NumField()        │  .String()           │
│   .Field(i)          │  .Bool()             │
│   .NumMethod()       │  .Len()              │
│   .Implements()      │  .Index(i)           │
│   .AssignableTo()    │  .MapKeys()          │
│                      │  .FieldByName()      │
│                      │  .Interface()        │
└──────────────────────┴──────────────────────┘
```

### reflect.Kind — Las Categorías Fundamentales

```go
reflect.TypeOf(42).Kind()          // reflect.Int
reflect.TypeOf(3.14).Kind()        // reflect.Float64
reflect.TypeOf("hello").Kind()     // reflect.String
reflect.TypeOf(true).Kind()        // reflect.Bool
reflect.TypeOf([]int{}).Kind()     // reflect.Slice
reflect.TypeOf(map[string]int{}).Kind() // reflect.Map
reflect.TypeOf(struct{}{}).Kind()  // reflect.Struct
reflect.TypeOf((*int)(nil)).Kind() // reflect.Ptr
```

### Inspeccionando Structs con Reflect

```go
type Usuario struct {
    Nombre  string `json:"nombre" validate:"required"`
    Edad    int    `json:"edad" validate:"min=0,max=150"`
    Email   string `json:"email" validate:"required,email"`
    Activo  bool   `json:"activo"`
}

u := Usuario{Nombre: "Ana", Edad: 30, Email: "ana@mail.com", Activo: true}

t := reflect.TypeOf(u)
v := reflect.ValueOf(u)

fmt.Printf("Tipo: %s\n", t.Name())        // "Usuario"
fmt.Printf("Campos: %d\n", t.NumField())  // 4

for i := 0; i < t.NumField(); i++ {
    campo := t.Field(i)
    valor := v.Field(i)
    
    fmt.Printf("  %s (%s) = %v | json: %s | validate: %s\n",
        campo.Name,
        campo.Type,
        valor,
        campo.Tag.Get("json"),
        campo.Tag.Get("validate"),
    )
}
// Salida:
//   Nombre (string) = Ana | json: nombre | validate: required
//   Edad (int) = 30 | json: edad | validate: min=0,max=150
//   Email (string) = ana@mail.com | json: email | validate: required,email
//   Activo (bool) = true | json: activo | validate:
```

### El Poder de los Struct Tags con Reflect

Los struct tags son **metadatos** que viven en los campos de un struct. `encoding/json` los usa para serializar JSON, `gorm` los usa para mapear a base de datos, y `go-playground/validator` los usa para validar datos. Todos usan `reflect` internamente.

```go
// Los tags son strings, pero reflect los parsea por key
type Producto struct {
    Nombre string `json:"nombre" db:"product_name" validate:"required,min=2"`
    Precio float64 `json:"precio" db:"price" validate:"gt=0"`
}

t := reflect.TypeOf(Producto{})
campo, _ := t.FieldByName("Nombre")

campo.Tag.Get("json")     // "nombre"
campo.Tag.Get("db")       // "product_name"
campo.Tag.Get("validate") // "required,min=2"
```

### Modificar Valores con Reflect

Para modificar un valor con reflect, necesitas pasar un **puntero**:

```go
x := 10
v := reflect.ValueOf(&x)  // ⚠️ Puntero, NO el valor directo
v.Elem().SetInt(42)       // Elem() desreferencia el puntero
fmt.Println(x)            // 42 ✅

// ❌ Esto NO funciona (panic):
v2 := reflect.ValueOf(x)  // Sin puntero
v2.SetInt(42)             // PANIC: reflect.Value.SetInt using unaddressable value
```

**La analogía del control remoto:** Pasar el valor directo es como intentar cambiar el canal tocando la pantalla de la TV — no puedes. Pasar un puntero es como usar el control remoto: puedes modificar el valor a distancia.

### Funciones Útiles con Reflect

#### Verificar si un Tipo Implementa una Interface

```go
tipoArchivo := reflect.TypeOf((*os.File)(nil)).Elem()
tipoReader := reflect.TypeOf((*io.Reader)(nil)).Elem()

// ¿os.File implementa io.Reader?
fmt.Println(tipoArchivo.Implements(tipoReader))  // true
```

#### Crear Instancias Dinámicamente

```go
t := reflect.TypeOf(Usuario{})
v := reflect.New(t)  // Crea un *Usuario nuevo (como new(Usuario))

// Acceder al struct apuntado
elem := v.Elem()
elem.FieldByName("Nombre").SetString("Carlos")
elem.FieldByName("Edad").SetInt(25)

usuario := v.Interface().(*Usuario)
fmt.Println(usuario)  // &{Carlos 25  false}
```

#### Convertir un Struct a Map[string]interface{}

```go
func StructAMap(obj interface{}) map[string]interface{} {
    resultado := make(map[string]interface{})
    v := reflect.ValueOf(obj)
    t := v.Type()
    
    if t.Kind() == reflect.Ptr {
        v = v.Elem()
        t = v.Type()
    }
    
    for i := 0; i < t.NumField(); i++ {
        campo := t.Field(i)
        valor := v.Field(i)
        
        key := campo.Tag.Get("json")
        if key == "" {
            key = strings.ToLower(campo.Name)
        }
        resultado[key] = valor.Interface()
    }
    return resultado
}
```

---

## ⚖️ Generics vs Reflect: ¿Cuándo Usar Cada Uno?

```
┌─────────────────────────────────────────────────────────────────┐
│                    ÁRBOL DE DECISIÓN                            │
│                                                                 │
│  ¿Conoces los tipos EN TIEMPO DE COMPILACIÓN?                  │
│         │                                                       │
│    ┌────┴────┐                                                  │
│    │         │                                                  │
│   SÍ        NO                                                  │
│    │         │                                                  │
│    ▼         ▼                                                  │
│ ¿Necesitas  Usa reflect                                         │
│ trabajar    (serializadores, ORMs,                              │
│ con varios  validadores, dependency injection)                  │
│ tipos?                                                          │
│    │                                                             │
│ ┌──┴──┐                                                         │
│ │     │                                                         │
│ SÍ    NO                                                        │
│ │     │                                                         │
│ ▼     ▼                                                         │
│ GENE- Usa tipos                                                 │
│ RICS  concretos                                                 │
│                                                                 │
│ ┌─────────────────────────────────────────────────────────┐     │
│ │ REGLA DE ORO:                                           │     │
│ │                                                         │     │
│ │ 🧩 Generics → Cuando el TIPO se conoce al compilar      │     │
│ │ 🔮 Reflect  → Cuando el TIPO se conoce solo en runtime  │     │
│ │                                                         │     │
│ │ Si puedes usar generics, USA GENERICS.                  │     │
│ │ Reflect es el último recurso.                           │     │
│ └─────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

| Criterio | Generics | Reflect |
|----------|----------|---------|
| **Seguridad** | ✅ Compile-time | ❌ Runtime (panics) |
| **Velocidad** | ✅ Misma que código sin generics | ❌ 10-100x más lento |
| **Autocompletado IDE** | ✅ Funciona perfectamente | ❌ No sabe los tipos |
| **Legibilidad** | ✅ Clara y concisa | ❌ Verbosa y confusa |
| **¿Cuándo usar?** | Colecciones, utilidades, algoritmos | Serializadores, ORMs, validadores |

### Ejemplo Comparativo: El Mismo Problema, Dos Soluciones

```go
// ✅ SOLUCIÓN CON GENERICS (preferida)
func Contains[T comparable](s []T, v T) bool {
    for _, item := range s {
        if item == v {
            return true
        }
    }
    return false
}

// ❌ SOLUCIÓN CON REFLECT (cuando no queda otra)
func ContainsReflect(s interface{}, v interface{}) bool {
    sv := reflect.ValueOf(s)
    if sv.Kind() != reflect.Slice {
        panic("primer argumento debe ser un slice")
    }
    for i := 0; i < sv.Len(); i++ {
        if sv.Index(i).Interface() == v {
            return true
        }
    }
    return false
}

// Comparación de rendimiento:
// Contains[int]    → ~2ns por operación
// ContainsReflect  → ~200ns por operación (100x más lento)
```

---

## 📝 Ejercicio Práctico: Librería Genérica de Colecciones + Validador de Structs

### ¿Qué construimos?

**Dos componentes** en un solo programa:

1. **Librería genérica de colecciones funcionales** → `Map`, `Filter`, `Reduce`, `Contains`, `Unique`, `Sort` para slices de cualquier tipo.
2. **Validador de structs con reflect** → Lee tags `validate:"required,min=3,max=50"` y valida automáticamente cualquier struct.

### Arquitectura del Código

```
📁 19-generics-reflect/
├── go.mod
└── main.go          ← Todo el código en un solo archivo
```

### Componentes Principales

#### 1. Funciones Genéricas de Colecciones

```go
// Map transforma cada elemento de un slice
func Map[T any, U any](slice []T, fn func(T) U) []U

// Filter selecciona elementos que cumplen una condición
func Filter[T any](slice []T, fn func(T) bool) []T

// Reduce combina todos los elementos en un solo valor
func Reduce[T any, U any](slice []T, inicial U, fn func(U, T) U) U

// Contains verifica si un elemento existe en el slice
func Contains[T comparable](slice []T, objetivo T) bool

// Unique elimina duplicados
func Unique[T comparable](slice []T) []T

// SortBy ordena con una función de comparación personalizada
func SortBy[T any](slice []T, less func(T, T) bool) []T
```

#### 2. Stack Genérico

```go
type Stack[T any] struct {
    elementos []T
}

func (s *Stack[T]) Push(v T)          // Apilar
func (s *Stack[T]) Pop() (T, bool)    // Desapilar
func (s *Stack[T]) Peek() (T, bool)   // Ver sin sacar
func (s *Stack[T]) Len() int          // Tamaño
func (s *Stack[T]) IsEmpty() bool     // ¿Vacío?
```

#### 3. Validador de Structs con Reflect

```go
type ValidationError struct {
    Field   string
    Tag     string
    Message string
}

func Validate(s interface{}) []ValidationError

// Lee tags como:
// validate:"required,min=3,max=50"
// validate:"required,email"
// validate:"gt=0,lte=100"
// validate:"required,min=1"
```

### Ejecución

```bash
# Ejecutar todo el programa (demostraciones + ejemplos)
go run main.go
```

### Qué Verás al Ejecutar

```
🧩 === GENERICS: Colecciones Funcionales ===

📊 Map: Convertir enteros a strings
   [1 2 3 4 5] → ["1" "2" "3" "4" "5"]

🔍 Filter: Números pares
   [1 2 3 4 5 6 7 8 9 10] → [2 4 6 8 10]

📦 Reduce: Sumar todos los elementos
   [1 2 3 4 5] → 15

🎯 Contains: ¿Existe el 3?
   [1 2 3 4 5] contiene 3? true

✨ Unique: Eliminar duplicados
   [1 2 2 3 3 3 4 4 4 4] → [1 2 3 4]

📐 SortBy: Ordenar por longitud de string
   ["banana" "kiwi" "manzana" "uva"] → ["uva" "kiwi" "banana" "manzana"]

📚 Stack Genérico:
   Push: Go, Python, Rust
   Pop: Rust → Python → Go

🔮 === REFLECT: Validador de Structs ===

✅ Usuario válido: {Ana ana@mail.com 25}
   Validación: ✅ Sin errores

❌ Usuario inválido: {"" "bad-email" -5}
   Errores:
   - Nombre: tag 'required' falló: valor vacío
   - Email: tag 'email' falló: no es un email válido
   - Edad: tag 'min' falló: -5 < 0
```

---

## 🔍 Conceptos Clave Explicados

### ¿Por qué `~int` en un constraint?

```go
type MiInt int  // MiInt es un tipo distinto de int

type Numero interface {
    int  // Solo int exacto, NO MiInt
}

type NumeroFlexible interface {
    ~int  // int Y cualquier tipo con int como base (incluye MiInt)
}
```

El operador `~` significa "este tipo subyacente y todos los tipos definidos basados en él". Sin `~`, `MiInt` no pasaría el constraint aunque internamente sea un `int`.

### ¿Por qué Reflect es tan lento?

```go
// Código normal: el compilador sabe que x es int
x := 42
y := x + 10  // El compilador genera: MOV $42, reg; ADD $10, reg

// Código con reflect: el compilador NO sabe qué es x
x := 42
v := reflect.ValueOf(x)
v.Int()  // En runtime: ¿es int? ¿int8? ¿int64? ¿float64?
         // Tiene que: buscar el tipo, verificar que sea numérico,
         // extraer el valor, convertir a int64, devolver
```

Reflect es 10-100x más lento porque **el compilador no puede optimizar**. Cada operación requiere verificaciones en runtime que el código normal hace en compile-time.

### ¿Cuándo `reflect.ValueOf` devuelve un puntero vs valor?

```go
x := 42

v1 := reflect.ValueOf(x)     // Valor: 42 (Kind = Int)
v2 := reflect.ValueOf(&x)    // Puntero: *int (Kind = Ptr)

v1.SetInt(42)                 // ❌ PANIC: no es addressable
v2.Elem().SetInt(42)          // ✅ Funciona: Elem() desreferencia

// Regla: para MODIFICAR, siempre pasa puntero
// Regla: Elem() desreferencia punteros e interfaces
```

### La Trampa de `any` vs Constraints Específicos

```go
// ❌ MALO: any no te deja hacer nada útil
func Sumar[T any](a, b T) T {
    return a + b  // ❌ Compile error: operator + not defined on T
}

// ✅ BIEN: constraint específico permite operaciones
type Sumable interface {
    ~int | ~float64 | ~string  // + está definido para estos tipos
}

func Sumar[T Sumable](a, b T) T {
    return a + a  // ✅ Funciona: sabemos que T soporta +
}
```

---

## 🏋️ Ejercicio Feynman

### Instrucciones
Usando la **Técnica Feynman**, explica estos conceptos **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado. Usa analogías de la vida cotidiana.

---

### Ejercicio 1: Generics como Recetas Universales
> Imagina que tienes una **receta de cocina** que dice "toma N ingredientes del mismo tipo y mézclalos". La receta funciona para frutas (batido de frutas), verduras (ensalada) o carnes (estofado). Explica qué es "N" (el type parameter), qué es "del mismo tipo" (el constraint), y por qué la receta NO te deja mezclar frutas con tornillos (`any` vs `Numero`).

---

### Ejercicio 2: El Constraint como Club Exclusivo
> Explica un constraint personalizado como un **club nocturno** con lista de miembros. `[Numero]` es el club "Solo Números" — entran `int`, `float64`, etc. `~int` significa "y todos los que se visten como `int`" (tipos derivados). `any` es la puerta abierta de una fiesta pública — entra cualquiera pero no puedes hacer nada específico. ¿Por qué `comparable` es un club donde solo entran personas que pueden ponerse en una balanza?

---

### Ejercicio 3: Reflect como un Detective
> Explica `reflect` como un **detective** que investiga una caja cerrada. `TypeOf` es el detective que dice "esta caja es de madera, mide 30x20x15, tiene etiqueta 'FRÁGIL'". `ValueOf` abre la caja y dice "dentro hay 5 manzanas". `Kind()` dice "es una caja de frutas" (la categoría). ¿Por qué investigar cajas cerradas (reflect) es más lento que leer la etiqueta de fábrica (generics)?

---

### Ejercicio 4: Generics vs Reflect — La Fábrica vs el Detective
> Imagina una **fábrica de juguetes**. Sin generics, tienes una máquina para hacer coches de madera, otra para coches de metal, otra para coches de plástico. Con generics, tienes una máquina que hace coches de CUALQUIER material siempre que sea "material rígido" (constraint). Reflect es un detective que examina lo que salió de la máquina en runtime. ¿Por qué la fábrica con generics es más rápida que tener un detective verificar cada juguete?

---

### Ejercicio 5: Struct Tags y Reflect como Etiquetas de Mudanza
> Explica los struct tags como **etiquetas adhesivas** en cajas de mudanza. La caja se llama `Nombre` en tu casa antigua, pero tiene una etiqueta que dice `json:"nombre"` (así se llamará en la casa nueva) y otra que dice `validate:"required"` (esta caja NO puede estar vacía). `reflect` es el coordinador de la mudanza que lee TODAS las etiquetas de TODAS las cajas y verifica que todo esté en orden. ¿Por qué no puedes leer etiquetas con generics?

---

### Ejercicio 6: `~int` y la Ropa de Etiqueta
> Explica el operador `~` en constraints usando la analogía de la **ropa**. `int` es como un código de vestimenta que dice "solo camisa blanca". `~int` dice "cualquier camisa de color claro". Si defines `type MiCamisa int`, `MiCamisa` es una camisa azul — cumple con `~int` pero NO con `int` exacto. ¿Por qué esta distinción es importante para crear tipos personalizados seguros?

---

## 📋 Resumen de Funciones Clave

| Función/Concepto | Paquete/Feature | ¿Qué hace? |
|------------------|-----------------|-------------|
| `[T any]` | Generics | Declara un type parameter que acepta cualquier tipo |
| `[T comparable]` | Generics | Type parameter que soporta `==` y `!=` |
| `[T ~int \| ~float64]` | Generics | Type parameter con tipos específicos y derivados |
| `func F[T any](...)` | Generics | Función genérica con un type parameter |
| `type S[T any] struct{...}` | Generics | Struct genérico con un type parameter |
| `reflect.TypeOf(x)` | reflect | Devuelve el `reflect.Type` de x |
| `reflect.ValueOf(x)` | reflect | Devuelve el `reflect.Value` de x |
| `.Kind()` | reflect | Devuelve la categoría del tipo (Int, String, Struct...) |
| `.NumField()` | reflect | Número de campos de un struct |
| `.Field(i)` | reflect | Accede al campo i de un struct |
| `.FieldByName("x")` | reflect | Busca un campo por nombre |
| `.Tag.Get("json")` | reflect | Lee el valor de un struct tag |
| `.SetInt(v)` | reflect | Modifica un valor int vía reflect (necesita puntero) |
| `.Elem()` | reflect | Desreferencia un puntero o interface |
| `.Interface()` | reflect | Convierte un reflect.Value de vuelta a interface{} |
| `.Implements(t)` | reflect | Verifica si un tipo implementa una interface |
| `reflect.New(t)` | reflect | Crea una nueva instancia del tipo t |

---

## 🔗 ¿Qué sigue?

En la **Lección 20 — Proyecto Final**, aplicarás TODO lo aprendido en las 19 lecciones para construir **GoKV**, una base de datos key-value con persistencia en disco, transacciones, y un servidor TCP estilo Redis. Será tu examen final donde generics, reflect, goroutines, channels, interfaces, y todo lo demás convergen en un solo proyecto épico.

---

> 💡 *"Los generics eliminan la duplicación. Reflect elimina los límites. La sabiduría está en saber cuál usar: la navaja suiza o el cuchillo multiusos que corta todo (incluyéndote)."*