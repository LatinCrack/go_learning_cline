<div align="center">

# 🧪 Laboratorio de Go — Lección 05

## **Estructuras, Métodos y la Filosofía "Composición sobre Herencia"**

<br>

![Lección 05](https://img.shields.io/badge/Lección-05-4ECDC4?style=for-the-badge) ![Método Feynman](https://img.shields.io/badge/Método-Feynman-FF6B6B?style=for-the-badge) ![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)

<br>

> *"La composición es el arte de construir castillos con LEGO: cada pieza es independiente, reemplazable, y no necesitas heredar el defecto del caballo para tener un barco."*
> — **Principio de Diseño Go**

</div>

---

## 🧠 Primero: El Mapa Mental

Antes de escribir una sola línea de código, necesitas entender POR QUÉ Go no tiene clases, y por qué eso es una **ventaja**, no una limitación.

### 🎭 El Problema que Go Resolvió

En los años 90 y 2000, los lenguajes orientados a objetos (Java, C++, C#) promovieron la **herencia** como la forma "correcta" de reutilizar código. La idea era seductora:

```
Animal (clase base)
├── Mamífero extends Animal
│   ├── Perro extends Mamífero
│   └── Gato extends Mamífero
└── Ave extends Animal
    └── Loro extends Ave
```

Pero en la práctica, esto generó **jerarquías frágiles**:

1. **El problema del diamante:** ¿Qué pasa si `Pato` hereda de `Volador` Y de `Nadador`? Ambos tienen un método `mover()`. ¿Cuál se ejecuta?
2. **El efecto cascada:** Cambias `Animal` y rompes 47 clases que heredan de él.
3. **El acoplamiento invisible:** Tu clase `Perro` arrastra todo el comportamiento de `Mamífero`, incluso lo que no necesita.

**Los creadores de Go (Rob Pike, Ken Thompson, Robert Griesemer) vieron estos problemas en Google** — con millones de líneas de código C++ y Java — y decidieron que Go NO tendría herencia.

En su lugar, Go tiene **structs** (para datos) y **composición** (para reutilización). Y eso es TODO lo que necesitas.

---

## 🏗️ ¿Qué es un Struct?

> **Analogía:** Un struct es como un **formulario de registro** en blanco. Tiene casillas etiquetadas: "Nombre", "Apellido", "Edad". Cada casilla tiene un tipo de dato (texto, número, fecha). Cuando llenas el formulario, creas una **instancia** de ese struct.

### 📝 Declaración e Instanciación

En Go, un struct se define con la palabra `type` seguida del nombre y la palabra `struct`:

```go
// Definición del "formulario" (el molde)
type Contacto struct {
    Nombre    string
    Apellido  string
    Email     string
    Edad      int
    Activo    bool
}
```

Esto es como diseñar el formulario. Aún no hay datos — solo la **estructura**.

### 🔧 Tres Formas de Crear una Instancia

```go
// Forma 1: Literal con nombres de campos (RECOMENDADA)
c1 := Contacto{
    Nombre:   "Carlos",
    Apellido: "Mendoza",
    Email:    "carlos@ejemplo.com",
    Edad:     30,
    Activo:   true,
}

// Forma 2: Usando new() — devuelve un PUNTERO al struct
c2 := new(Contacto)
c2.Nombre = "María"     // Se accede con punto, igual que un valor normal
c2.Apellido = "García"

// Forma 3: Función constructor (patrón idiomatico)
func NuevoContacto(nombre, apellido, email string) *Contacto {
    return &Contacto{
        Nombre:   nombre,
        Apellido: apellido,
        Email:    email,
        Activo:   true,  // Valor por defecto
    }
}
```

### 🎯 ¿Cuál usar?

| Forma | Cuándo usarla |
|:------|:--------------|
| **Literal con nombres** | Cuando creas el struct completo de una vez. Siempre incluye los nombres de los campos. |
| **`new()`** | Cuando quieres un puntero y vas a llenar los campos después. |
| **Función constructor** | Cuando quieres lógica de inicialización (validar campos, poner defaults). **Es el patrón preferido en Go.** |

### ⚠️ Regla de oro: Siempre nombra los campos

```go
// ❌ PELIGROSO — depende del orden
c := Contacto{"Carlos", "Mendoza", "carlos@ejemplo.com", 30, true}

// ✅ SEGURO — explícito y legible
c := Contacto{
    Nombre:   "Carlos",
    Apellido: "Mendoza",
    Email:    "carlos@ejemplo.com",
    Edad:     30,
    Activo:   true,
}
```

¿Por qué? Si alguien agrega un campo al struct, la primera forma se rompe silenciosamente. La segunda es inmune a cambios.

---

## 🔧 Métodos: Comportamiento para tus Structs

> **Analogía:** Si el struct es el cuerpo de un robot, los métodos son sus **habilidades**. El robot `Contacto` puede "hablar su nombre completo", "marcarse como favorito", o "agregar un teléfono". Cada habilidad es un método.

Un método en Go es una función que tiene un **receptor** — el tipo al que pertenece. Se declara así:

```go
func (receptor TipoReceptor) NombreMetodo(parametros) retorno {
    // cuerpo
}
```

### 📖 Receptor por Valor vs Receptor por Puntero

Esta es una de las decisiones más importantes que tomarás en Go. Hay dos formas de declarar el receptor:

```go
// Receptor por VALOR: trabaja con una COPIA del struct
func (c Contacto) NombreCompleto() string {
    return c.Nombre + " " + c.Apellido
}

// Receptor por PUNTERO: trabaja con el struct ORIGINAL
func (c *Contacto) MarcarFavorito() {
    c.Favorito = true
}
```

### 🧩 ¿Cuándo usar cada uno?

| Receptor | Cuándo usarlo | ¿Modifica el original? |
|:---------|:--------------|:-----------------------|
| **Valor** `(c Contacto)` | Solo LEE datos. No cambia nada. Ejemplo: calcular, formatear, comparar. | ❌ No |
| **Puntero** `(c *Contacto)` | MODIFICA el struct. Ejemplo: cambiar un campo, agregar a un slice, actualizar estado. | ✅ Sí |

**Regla simple:** Cuando dudes, usa puntero. Es más consistente y no copia structs innecesariamente.

### 🏭 Ejemplo Completo de Métodos

```go
type Contacto struct {
    Nombre    string
    Apellido  string
    Favorito  bool
    Telefonos []string
}

// Receptor VALOR → solo lectura
func (c Contacto) NombreCompleto() string {
    return c.Nombre + " " + c.Apellido
}

// Receptor PUNTERO → modifica el struct
func (c *Contacto) MarcarFavorito() {
    c.Favorito = true
}

// Receptor PUNTERO → modifica el slice
func (c *Contacto) AgregarTelefono(tel string) {
    c.Telefonos = append(c.Telefonos, tel)
}
```

**Uso:**

```go
c := Contacto{Nombre: "Carlos", Apellido: "Mendoza"}

fmt.Println(c.NombreCompleto()) // "Carlos Mendoza" — receptor valor, no cambia nada

c.MarcarFavorito()              // receptor puntero, MODIFICA c
c.AgregarTelefono("+51 999")   // receptor puntero, MODIFICA c

fmt.Println(c.Favorito)        // true — el cambio persiste
fmt.Println(c.Telefonos)       // ["+51 999"]
```

### 🎯 El Poder de `fmt.Stringer`

Go tiene una interface especial llamada `fmt.Stringer`:

```go
type Stringer interface {
    String() string
}
```

Si tu struct implementa este método, `fmt.Println` lo llama automáticamente:

```go
func (c Contacto) String() string {
    return fmt.Sprintf("%s <%s>", c.NombreCompleto(), c.Email)
}

c := Contacto{Nombre: "Carlos", Apellido: "Mendoza", Email: "carlos@ej.com"}
fmt.Println(c) // Imprime: "Carlos Mendoza <carlos@ej.com>"
```

---

## 🧩 Composición mediante Embedding

> **Analogía:** Imagina que estás armando un equipo de fútbol. En la herencia (Java), si quieres un "Jugador que también es Entrenador", necesitas una jerarquía compleja. En la composición (Go), simplemente le pegas una gorra de entrenador al jugador: `JugadorConGorra struct { Jugador; Gorra }`. El jugador sigue siendo jugador, pero ahora también tiene la gorra.

### 🏗️ ¿Qué es el Embedding?

El embedding (incrustación) es la forma en Go de componer structs. En vez de heredar de una clase padre, **incluyes** otro struct directamente:

```go
// Struct base: una etiqueta
type Etiqueta struct {
    Nombre string
    Color  string
}

// Struct compuesto: Contacto + Etiqueta
type ContactoEtiquetado struct {
    Contacto             // ← Sin nombre de campo = EMBEDDING
    Etiqueta             // ← Otro embedding
    Notas    string      // ← Campo propio adicional
}
```

### 🔑 ¿Qué obtienes con el Embedding?

1. **Acceso directo a los campos del struct embebido:**
   ```go
   ce := ContactoEtiquetado{
       Contacto: Contacto{Nombre: "Carlos", Apellido: "Mendoza"},
       Etiqueta: Etiqueta{Nombre: "trabajo", Color: "azul"},
   }

   fmt.Println(ce.Nombre)   // "Carlos" — campo de Contacto, accesible directamente
   fmt.Println(ce.Color)    // "azul"   — campo de Etiqueta, accesible directamente
   ```

2. **Los métodos del struct embebido "se heredan" (pero no es herencia):**
   ```go
   ce.MarcarFavorito()      // Método de Contacto, disponible en ContactoEtiquetado
   ce.NombreCompleto()      // Método de Contacto, disponible también
   ```

3. **Puedes sobrescribir métodos:**
   ```go
   func (ce ContactoEtiquetado) Ficha() string {
       // Llama al Ficha original del Contacto embebido
       fichaBase := ce.Contacto.Ficha()
       // Agrega información de la etiqueta
       return fichaBase + fmt.Sprintf("Etiqueta: %s (%s)", ce.Etiqueta.Nombre, ce.Color)
   }
   ```

### 🚫 ¿Qué pasa si dos structs embebidos tienen el mismo campo?

```go
type Persona struct { Nombre string }
type Producto struct { Nombre string }

type Combo struct {
    Persona
    Producto
}

c := Combo{}
c.Nombre = "Carlos" // ❌ ERROR: ambiguous selector c.Nombre
```

**Solución:** Especifica cuál quieres:

```go
c.Persona.Nombre = "Carlos"    // ✅ Desambiguado
c.Producto.Nombre = "Laptop"   // ✅ Desambiguado
```

### 💡 Composición vs Herencia: La Diferencia Clave

| Herencia (Java/C++) | Composición (Go) |
|:--------------------|:-----------------|
| `Perro extends Animal` — el perro **ES** un animal | `ContactoEtiquetado` **CONTIENE** un Contacto |
| Cambiar `Animal` rompe `Perro` | Cambiar `Contacto` no rompe `ContactoEtiquetado` |
| Problema del diamante | No existe el problema del diamante |
| Jerarquía rígida | Composición flexible y reconfigurable |
| Acoplamiento fuerte | Acoplamiento débil |

---

## 🔌 Interfaces: El Superpoder Oculto de Go

> **Analogía:** Una interface es como un **enchufe universal**. No le importa si enchufas un cargador de iPhone o de Samsung — mientras tenga el conector USB-C, funciona. La interface no sabe NADA del dispositivo, solo sabe que tiene el conector correcto.

### 📐 Declaración

```go
// Define QUÉ puede hacer, no CÓMO lo hace
type Buscador interface {
    Buscar(termino string) []Contacto
}

type Exportador interface {
    ExportarJSON() ([]byte, error)
    ExportarTabla() string
}
```

### 🪄 La Magia: Implementación Implícita

En Java, necesitas decir `class MiClase implements Buscador`. En Go, **NO**. Si tu tipo tiene los métodos que la interface pide, **automáticamente** implementa esa interface:

```go
// Agenda tiene un método Buscar(termino string) []Contacto
// → ¡Automáticamente implementa la interface Buscador!
type Agenda struct {
    contactos []Contacto
}

func (a Agenda) Buscar(termino string) []Contacto {
    // ... implementación
}
```

**¿Por qué esto es genial?** Porque permite **desacoplamiento total**:

- El paquete que define `Buscador` no necesita importar el paquete que define `Agenda`.
- El paquete que define `Agenda` no necesita importar el paquete que define `Buscador`.
- Cualquier struct futuro puede implementar `Buscador` sin modificar el código original.

### 🎯 Interfaces como Parámetros

```go
// Esta función acepta CUALQUIER cosa que implemente Buscador
func mostrarResultados(b Buscador, termino string) {
    resultados := b.Buscar(termino)
    for _, r := range resultados {
        fmt.Println(r)
    }
}

// Puedes pasar una Agenda, un BuscadorWeb, un BuscadorDeEmails...
// La función no sabe qué tipo es, solo sabe que puede Buscar
mostrarResultados(miAgenda, "Carlos")
```

### 📏 La Interface Vacía: `interface{}` / `any`

```go
// Acepta CUALQUIER cosa — como Object en Java
func mostrar(valor interface{}) {
    fmt.Println(valor)
}

// En Go 1.18+, puedes usar `any` (alias de interface{})
func mostrar(valor any) {
    fmt.Println(valor)
}
```

Úsala con precaución: pierdes la seguridad de tipos.

---

## 🏷️ Struct Tags: Metadatos para Serialización

> **Analogía:** Los struct tags son como **etiquetas adhesivas** en una mudanza. Le dicen a los movers (el marshaller JSON): "esta caja se llama 'nombre' en el nuevo apartamento" (`json:"nombre"`). Sin etiquetas, los movers usan el nombre original de la caja.

### 📋 Sintaxis

```go
type ContactoJSON struct {
    Nombre   string `json:"nombre"`
    Apellido string `json:"apellido"`
    Email    string `json:"email"`
    Favorito bool   `json:"favorito,omitempty"`
    Interno  string `json:"-"` // ← Nunca se serializa
}
```

### 🎯 Tags Más Comunes

| Tag | Efecto |
|:----|:-------|
| `json:"nombre"` | Usa "nombre" como key en JSON (en vez de "Nombre") |
| `json:"-"` | **Nunca** incluir este campo en JSON |
| `json:"nombre,omitempty"` | Omitir si el valor es cero (`""`, `false`, `0`, `nil`) |
| `xml:"nombre"` | Similar, pero para XML |
| `validate:"required"` | Validación (requiere librería externa) |

### 🔧 Ejemplo de Serialización

```go
c := ContactoJSON{
    Nombre:   "Carlos",
    Apellido: "Mendoza",
    Email:    "carlos@ej.com",
    Favorito: false,       // omitempty → no aparece en JSON
    Interno:  "secreto",   // json:"-" → nunca aparece
}

jsonData, _ := json.MarshalIndent(c, "", "  ")
fmt.Println(string(jsonData))
```

**Salida:**
```json
{
  "nombre": "Carlos",
  "apellido": "Mendoza",
  "email": "carlos@ej.com"
}
```

Observa: `favorito` no aparece (es `false` + `omitempty`), e `interno` no aparece (`"-"`).

---

## 🧱 Patrón Constructor: `NuevoXxx()`

> **Analogía:** Si el struct es un coche recién salido de la fábrica, el constructor es el **taller de personalización** que le pone las llantas correctas, llena el tanque, y ajusta los espejos antes de entregártelo.

Go no tiene constructores como Java (`constructor()`). En su lugar, usa **funciones que devuelven punteros**:

```go
// Patrón idiomatico: función NuevaXxx que devuelve *Xxx
func NuevaAgenda(nombre string) *Agenda {
    return &Agenda{
        Nombre:    nombre,
        contactos: make([]Contacto, 0), // Inicializa el slice
    }
}

// Uso
agenda := NuevaAgenda("Mi Agenda")
```

**¿Por qué un puntero?** Porque el constructor crea algo y quiere que trabajes con **la misma instancia**, no con una copia.

### 🎯 Ventajas del Patrón Constructor

1. **Validación en la creación:**
   ```go
   func NuevoContacto(nombre, email string) (*Contacto, error) {
       if nombre == "" {
           return nil, fmt.Errorf("el nombre no puede estar vacío")
       }
       if !strings.Contains(email, "@") {
           return nil, fmt.Errorf("email inválido: %s", email)
       }
       return &Contacto{Nombre: nombre, Email: email, Activo: true}, nil
   }
   ```

2. **Valores por defecto sensatos:**
   ```go
   func NuevaAgenda(nombre string) *Agenda {
       return &Agenda{
           Nombre:    nombre,
           contactos: make([]Contacto, 0), // ← Evita nil slice
       }
   }
   ```

3. **Encapsulamiento:** El campo `contactos` (minúscula) es privado. Solo puedes acceder a través de métodos.

---

## 📦 Structs con Slices y Maps Internos

Los structs pueden contener cualquier tipo, incluyendo slices y maps:

```go
type Agenda struct {
    Nombre    string
    contactos []Contacto                    // Slice de structs
    porCiudad map[string][]Contacto         // Map: ciudad → contactos
}

func NuevaAgenda(nombre string) *Agenda {
    return &Agenda{
        Nombre:    nombre,
        contactos: make([]Contacto, 0),
        porCiudad: make(map[string][]Contacto),
    }
}
```

**⚠️ Importante:** Siempre inicializa los slices y maps en el constructor. Un slice o map `nil` causa panic al usar `append` o asignar valores.

---

## 🔄 Ordenar Structs con `sort.Slice`

Go no tiene `sort` integrado en los structs. Usa `sort.Slice`:

```go
contactos := []Contacto{
    {Nombre: "Carlos", Apellido: "Mendoza"},
    {Nombre: "Ana", Apellido: "Torres"},
    {Nombre: "Bruno", Apellido: "García"},
}

// Ordenar por apellido
sort.Slice(contactos, func(i, j int) bool {
    return contactos[i].Apellido < contactos[j].Apellido
})

// Resultado: García, Mendoza, Torres
```

---

## 🧩 Resumen Visual

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   TYPE (definición)        INSTANCIA (valor)                │
│                                                             │
│   type Contacto struct {   c := Contacto{                   │
│       Nombre string            Nombre: "Carlos",            │
│       Edad   int               Edad:   30,                  │
│   }                          }                              │
│                                                             │
│   MÉTODO (receptor)        EMBEDDING (composición)          │
│                                                             │
│   func (c Contacto)        type ContactoEtiquetado struct { │
│       NombreCompleto()         Contacto   ← embebido        │
│       string {                 Etiqueta   ← embebido        │
│       return c.Nombre         Notas string                  │
│   }                         }                               │
│                                                             │
│   INTERFACE (contrato)     STRUCT TAGS (metadatos)          │
│                                                             │
│   type Buscador interface  type ContactoJSON struct {       │
│       Buscar(string)           Nombre string `json:"nom"`   │
│       []Contacto               Edad int    `json:"-"`       │
│   }                          }                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🏋️ Ejercicio Práctico: Mini Gestor de Contactos en Archivo

### 🎯 Objetivo

Construir un gestor de contactos CLI que demuestre **TODOS** los conceptos de structs: creación, métodos con ambos receptores, composición mediante embedding, interfaces, struct tags, y exportación JSON.

### 📁 Estructura del Proyecto

```
05-structs/
├── main.go     ← Código completo del ejercicio
└── go.mod      ← Módulo Go
```

### 📄 Código Línea por Línea

El archivo `main.go` está dividido en **9 secciones** que demuestran cada concepto progresivamente. A continuación, un recorrido exhaustivo por la lógica de cada parte:

---

#### **Sección 1: Definición de Structs Básicos**

```go
type Direccion struct {
    Calle     string
    Ciudad    string
    Pais      string
    CodigoP string
}
```

> 📌 **`Direccion`** modela una ubicación física. Nota que todos los campos empiezan con **mayúscula** → son exportados (visibles fuera del paquete).

```go
func (d Direccion) String() string {
    return fmt.Sprintf("%s, %s, %s %s", d.Calle, d.Ciudad, d.Pais, d.CodigoP)
}
```

> 📌 Implementa la interface `fmt.Stringer`. Cuando haces `fmt.Println(d)`, Go llama a `String()` automáticamente. Esto es **polimorfismo** sin herencia.

```go
type Contacto struct {
    Nombre    string
    Apellido  string
    Email     string
    Telefonos []Telefono
    Direccion Direccion      // ← Un struct dentro de otro (composición simple)
    Favorito  bool
    CreadoEn  time.Time
}
```

> 📌 `Contacto` contiene un `Direccion` y un slice de `Telefono`. Esto es **composición** — el contacto TIENE una dirección, no ES una dirección.

---

#### **Sección 2: Métodos con Receptor Valor**

```go
func (c Contacto) NombreCompleto() string {
    return fmt.Sprintf("%s %s", c.Nombre, c.Apellido)
}
```

> 📌 **Receptor por valor** `(c Contacto)`: Go copia el struct completo para esta operación. Es seguro para operaciones de solo lectura. La copia es desechable — no afecta al original.

```go
func (c Contacto) EdadDiasDesde() int {
    return int(time.Since(c.CreadoEn).Hours() / 24)
}
```

> 📌 Calcula cuántos días lleva el contacto en la agenda. Solo lee `CreadoEn`, no modifica nada → receptor valor.

---

#### **Sección 3: Métodos con Receptor Puntero**

```go
func (c *Contacto) MarcarFavorito() {
    c.Favorito = true
}
```

> 📌 **Receptor por puntero** `*Contacto`: trabaja con el struct **original**, no con una copia. El cambio `c.Favorito = true` persiste después de que el método termina. Si fuera receptor por valor, el cambio se perdería.

```go
func (c *Contacto) AgregarTelefono(numero, tipo string) {
    c.Telefonos = append(c.Telefonos, Telefono{Numero: numero, Tipo: tipo})
}
```

> 📌 Modifica el slice `Telefonos` del contacto original. `append` puede crear un nuevo array subyacente, pero el puntero `c` siempre apunta al struct correcto.

---

#### **Sección 4: Composición mediante Embedding**

```go
type ContactoEtiquetado struct {
    Contacto             // ← Embedding: sin nombre = "incrustado"
    Etiqueta             // ← Otro embedding
    Notas      string    // ← Campo propio
}
```

> 📌 `ContactoEtiquetado` **compone** `Contacto` y `Etiqueta`. No es herencia: puedes quitar `Etiqueta` sin romper `ContactoEtiquetado`. Los campos de `Contacto` son accesibles directamente: `ce.Nombre` (si no hay ambigüedad).

```go
func (ce ContactoEtiquetado) Ficha() string {
    fichaBase := ce.Contacto.Ficha()
    return fichaBase + fmt.Sprintf("     🏷️  [%s] %s — %s\n", ce.Etiqueta.Color, ce.Etiqueta.Nombre, ce.Notas)
}
```

> 📌 **Sobrescritura de método:** `ContactoEtiquetado` tiene su propio `Ficha()` que extiende el del `Contacto` embebido. Para llamar al original, usas `ce.Contacto.Ficha()` explícitamente (no hay `super` ni `parent`).

---

#### **Sección 5: Interfaces**

```go
type Buscador interface {
    Buscar(termino string) []Contacto
}

type Exportador interface {
    ExportarJSON() ([]byte, error)
    ExportarTabla() string
}
```

> 📌 Estas interfaces definen **comportamiento**, no datos. No importa qué struct las implemente — mientras tenga los métodos correctos, Go lo acepta.

```go
func (a Agenda) Buscar(termino string) []Contacto { ... }
func (a Agenda) ExportarJSON() ([]byte, error) { ... }
func (a Agenda) ExportarTabla() string { ... }
```

> 📌 `Agenda` implementa `Buscador` y `Exportador` **implícitamente**. No hay `implements` keyword. Si mañana creas un `BuscadorWeb` con un método `Buscar`, también implementará `Buscador` — sin modificar nada.

```go
func demostrarInterface(b Buscador, termino string) {
    resultados := b.Buscar(termino)
    // ...
}
```

> 📌 La función `demostrarInterface` acepta **cualquier cosa** que implemente `Buscador`. Esto es el poder del desacoplamiento: la función no sabe nada sobre `Agenda`.

---

#### **Sección 6: Struct Tags y JSON**

```go
type ContactoJSON struct {
    Nombre   string `json:"nombre"`
    Apellido string `json:"apellido"`
    Email    string `json:"email"`
    Ciudad   string `json:"ciudad"`
    Favorito bool   `json:"favorito,omitempty"`
}
```

> 📌 Los tags `json:"..."` controlan cómo `encoding/json` serializa cada campo. `omitempty` omite campos con valor cero. `json:"-"` excluiría el campo completamente.

---

#### **Sección 7: Constructor y Agregación**

```go
func NuevaAgenda(nombre string) *Agenda {
    return &Agenda{
        Nombre:    nombre,
        contactos: make([]Contacto, 0),
    }
}
```

> 📌 **Patrón constructor:** devuelve un puntero (`*Agenda`) y inicializa el slice con `make`. Esto evita el error clásico de hacer `append` a un slice `nil` (que funciona, pero es confuso) y garantiza que todos los usuarios trabajen con la misma instancia.

---

#### **Sección 8: Ordenamiento con `sort.Slice`**

```go
sort.Slice(contactosOrdenados, func(i, j int) bool {
    return contactosOrdenados[i].Apellido < contactosOrdenados[j].Apellido
})
```

> 📌 Go no tiene `Comparable` ni `Sortable` como Java. Usa funciones anónimas (closures) para definir el criterio de ordenamiento. `sort.Slice` es la forma más simple de ordenar slices de structs.

---

#### **Sección 9: Demostración Final**

La última sección consolida todo: crea la agenda, agrega contactos, busca por texto, filtra por ciudad, cuenta favoritos, y exporta a tabla y JSON.

---

### ▶️ Ejecución

```bash
cd 05-structs
go run main.go
```

**Salida esperada:**

```
╔══════════════════════════════════════════════════════╗
║   📒 MINI GESTOR DE CONTACTOS                       ║
║   Structs · Métodos · Composición · Interfaces       ║
╚══════════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   1️⃣  Crear Structs: instanciación y literales
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   ✅ Tres contactos creados con diferentes métodos:

   Forma 1 (literal): Carlos Mendoza — carlos@ejemplo.com
   Forma 2 (new):     María García — maria@ejemplo.com
   Forma 3 (campos):  Ana Torres — ana@ejemplo.com

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   2️⃣  Métodos: receptor valor vs receptor puntero
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📋 Ficha de contacto1 ANTES de marcar favorito:
     👤 Carlos Mendoza
        📧 carlos@ejemplo.com
        📍 Av. Principal 123, Lima, Perú 15001
        📅 En agenda desde hace 0 días
   📋 Ficha de contacto1 DESPUÉS de modificar:
     👤 Carlos Mendoza ⭐
        📧 carlos@ejemplo.com
        📱 [móvil] +51 999 888 777
        📱 [trabajo] +51 01 234 5678
        📍 Av. Principal 123, Lima, Perú 15001
        📅 En agenda desde hace 0 días

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   3️⃣  Composición mediante embedding (NO herencia)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   Nombre: Roberto
   Etiqueta: trabajo
   Email: roberto@empresa.com
   Notas: Gerente de proyecto del equipo backend

   📋 Ficha compuesta (Contacto + Etiqueta):
     👤 Roberto Díaz ⭐
        📧 roberto@empresa.com
        📱 [móvil] +51 999 111 222
        📍 Av. Industrial 500, Lima, Perú 15033
        📅 En agenda desde hace 0 días
        🏷️  [🔵] trabajo — Gerente de proyecto del equipo backend

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   4️⃣  Interfaces: duck typing verificado por el compilador
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📒 Agenda "Mi Agenda Personal" creada con 4 contactos

   🔌 Probando interface Buscador:
   🔍 Buscando "Carlos"...
   → Carlos Mendoza (carlos@ejemplo.com)

   🔍 Buscando "Lima"...
   → Carlos Mendoza (carlos@ejemplo.com)
   → Roberto Díaz (roberto@empresa.com)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   5️⃣  Exportación con interfaces (Exportador)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📊 Tabla de contactos:
     NOMBRE                    EMAIL                          CIUDAD          FAV
   ─────────────────────────────────────────────────────────────────────────────────
     Carlos Mendoza            carlos@ejemplo.com             Lima            ⭐
     María García              maria@ejemplo.com              Bogotá
     Ana Torres                ana@ejemplo.com                São Paulo
     Roberto Díaz              roberto@empresa.com            Lima            ⭐

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   6️⃣  Struct tags: controlando la serialización JSON
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   Con Favorito=false y tag `omitempty`:
   {
     "nombre": "Laura",
     "apellido": "Sánchez",
     "email": "laura@ejemplo.com",
     "ciudad": "Ciudad de México"
   }

   Con Favorito=true:
   {
     "nombre": "Laura",
     "apellido": "Sánchez",
     "email": "laura@ejemplo.com",
     "ciudad": "Ciudad de México",
     "favorito": true
   }

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   9️⃣  Demostración final: agenda completa
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📒 Agenda: Mi Agenda Personal
   📊 Total contactos: 4
   ⭐ Favoritos: 2
   🌎 En Lima: 2
   🌎 En Bogotá: 1

═══════════════════════════════════════════════════════
   ✅ Todos los conceptos de structs ejecutados correctamente
═══════════════════════════════════════════════════════
```

---

## 🏋️ Ejercicio Feynman

> **Instrucción:** Toma una hoja en blanco o abre un archivo de texto vacío. Sin consultar esta lección, intenta responder cada pregunta **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado.

### 📝 Preguntas para explicar desde cero:

1. **¿Qué es un struct?** Explica con una analogía que NO sea de cajones, formularios ni bases de datos. Crea la tuya propia.

2. **¿Por qué Go no tiene clases ni herencia?** ¿Qué problemas de la herencia querían evitar los creadores de Go? Usa el ejemplo del "problema del diamante".

3. **Un amigo Java te dice:** *"Sin herencia no puedo reutilizar código."* Respóndele explicando la composición con embedding. ¿Por qué es MÁS flexible que la herencia?

4. **Explica la diferencia entre receptor por valor y receptor por puntero sin usar la palabra "copia".** Usa una analogía de la vida real.

5. **¿Qué son las interfaces implícitas de Go y por qué son más poderosas que las explícitas de Java?** Piensa en un escenario donde un código escrito hace 5 años funciona con un tipo que se creó ayer.

6. **Un colega te dice:** *"Las interfaces de Go son como duck typing de Python."* ¿Es correcto? ¿Cuál es la diferencia fundamental?

7. **Predice el resultado de este código SIN ejecutarlo:**
   ```go
   type Base struct {
       Valor int
   }

   func (b Base) Obtener() int {
       return b.Valor
   }

   type Compuesto struct {
       Base
       Extra int
   }

   c := Compuesto{Base: Base{Valor: 10}, Extra: 5}
   fmt.Println(c.Obtener())
   fmt.Println(c.Valor)
   fmt.Println(c.Extra)
   ```

### ✅ Criterio de autoevaluación:

| Criterio                                                          | ¿Lo lograste? |
|-------------------------------------------------------------------|:-------------:|
| Explicaste structs sin mirar la lección                           | ⬜ Sí / ⬜ No |
| Creaste tu propia analogía original                               | ⬜ Sí / ⬜ No |
| Entiendes por qué Go eligió composición sobre herencia           | ⬜ Sí / ⬜ No |
| Puedes explicar receptor valor vs puntero con una analogía       | ⬜ Sí / ⬜ No |
| Entiendes las interfaces implícitas de Go                        | ⬜ Sí / ⬜ No |
| Distingues composición de herencia claramente                    | ⬜ Sí / ⬜ No |
| Predijiste correctamente la salida del código del ejercicio 7    | ⬜ Sí / ⬜ No |

---

## 🗺️ Próxima lección

En la **Lección 06** exploraremos **Arrays, Slices y el Secreto del Runtime de Go**: cómo Go gestiona colecciones de datos dinámicas, qué pasa internamente cuando haces `append`, y por qué entender el aliasing de memoria te salvará de bugs silenciosos en producción.

> *"Un struct bien diseñado es como una celda perfecta de un panal: contiene exactamente lo que necesita, nada más, y se conecta con las demás sin fricción."* — Principio Feynman