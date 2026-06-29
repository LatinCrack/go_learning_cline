# 📦 Lección 17: JSON, Serialización y APIs REST

## 🎯 Objetivo de la Lección

Dominar la **serialización y deserialización JSON** en Go usando el paquete `encoding/json`, entender el sistema de **struct tags**, y construir un **cliente de API REST** que consuma datos de una API pública, los cache localmente y los presente formateados en consola.

---

## 🧠 Analogía Fundamental: Las Etiquetas de la Mudanza

Imagina que te estás **mudando de apartamento**. Tienes cajas con todas tus pertenencias, y cada caja tiene un nombre en tu idioma: "libros de cocina", "ropa de invierno", "electrónicos".

Pero el nuevo apartamento está en otro país, y los movers hablan inglés. Necesitas **etiquetas adhesivas** en cada caja con el nombre en inglés:

| Caja (Campo Go) | Etiqueta adhesiva (JSON tag) | Destino |
|:-----------------|:-----------------------------|:--------|
| `NombreCompleto` | `"name"` | `{ "name": "Carlos" }` |
| `EdadAnos` | `"age"` | `{ "age": 30 }` |
| `EmailPrivado` | `"email,omitempty"` | Se omite si está vacío |

El **marshaller** de Go es como el jefe de los movers: toma tus cajas (structs), lee las etiquetas adhesivas (struct tags), y las coloca en el camión (JSON) con los nombres correctos. Sin etiquetas, usa el nombre original de la caja.

### 🔑 El Viaje de Ida y Vuelta

```
         MARSHAL (ida)                    UNMARSHAL (vuelta)
  ┌─────────────────────┐           ┌─────────────────────┐
  │  Struct en Go        │  ──────►  │  JSON               │
  │  {                   │  json.    │  {                   │
  │    Nombre: "Ana"     │  Marshal  │    "nombre": "Ana"  │
  │    Edad: 25          │           │    "edad": 25        │
  │  }                   │           │  }                   │
  └─────────────────────┘           └─────────────────────┘
            ▲                                  │
            │         json.Unmarshal           │
            └──────────────────────────────────┘
```

- **`json.Marshal`** → Serializa: de Go a JSON (empacas las cajas)
- **`json.Unmarshal`** → Deserializa: de JSON a Go (desempacas las cajas)

---

## 📦 Paquetes que Estudiaremos

| Paquete | ¿Qué hace? | Analogía |
|---------|-------------|----------|
| `encoding/json` | Serialización/deserialización JSON | Los **movers** que empaquetan y desempaquetan |
| `net/http` | Cliente y servidor HTTP | El **teléfono** para hablar con APIs |
| `io` | Interfaces de flujo de datos | La **manguera** de datos |
| `time` | Manejo de fechas y duraciones | El **reloj** del sistema |
| `fmt` | Formateo de strings | La **imprenta** |
| `os` | Sistema operativo, archivos | El **edificio** |

---

## 🔑 Struct Tags: Las Etiquetas Adhesivas

Los **struct tags** son metadatos escritos como strings en los campos de un struct. Van después del tipo, entre backticks:

```go
type Usuario struct {
    Nombre  string  `json:"nombre"`
    Edad    int     `json:"edad"`
    Email   string  `json:"email,omitempty"`
    secreto string  // sin tag = no se exporta (minúscula)
}
```

### Reglas Fundamentales

| Regla | Ejemplo | Resultado |
|:------|:--------|:----------|
| `json:"nombre"` | Campo se serializa como `"nombre"` | Personaliza la key en JSON |
| `json:"-"` | Campo se ignora completamente | Nunca aparece en JSON |
| `json:",omitempty"` | Se omite si el valor es zero value | Solo aparece si tiene valor |
| Sin tag | Se usa el nombre del campo tal cual | `"Nombre"` (con mayúscula) |
| Campo no exportado (minúscula) | Se ignora siempre | No necesita `-` |

### Ejemplos Prácticos de Tags

```go
type Producto struct {
    ID       int     `json:"id"`
    Nombre   string  `json:"nombre"`
    Precio   float64 `json:"precio"`
    Activo   bool    `json:"activo,omitempty"`
    Interno  string  `json:"-"`                    // nunca se serializa
    codigo   string                                 // no exportado, se ignora
}

p := Producto{ID: 1, Nombre: "Laptop", Precio: 999.99}
datos, _ := json.Marshal(p)
// {"id":1,"nombre":"Laptop","precio":999.99}
// Nota: "activo" se omite porque es false (zero value de bool)
// Nota: "interno" se omite por el tag "-"
// Nota: "codigo" se omite porque no está exportado
```

---

## 🔧 Funciones del Paquete `encoding/json`

### json.Marshal — Empaquetar

```go
datos, err := json.Marshal(valor)
// datos es []byte con el JSON
```

- Devuelve `[]byte`, no `string` (por eficiencia)
- Si hay error de serialización, devuelve `error`
- Los campos no exportados (minúscula) **siempre se ignoran**

### json.MarshalIndent — Empaquetar con formato

```go
datos, err := json.MarshalIndent(valor, "", "  ")
// El tercer parámetro es la indentación
// "" = prefijo de cada línea, "  " = 2 espacios por nivel
```

Ideal para mostrar JSON legible en consola o guardar en archivos de configuración.

### json.Unmarshal — Desempaquetar

```go
var usuario Usuario
err := json.Unmarshal(jsonBytes, &usuario)
// Nota: pasamos PUNTERO (&) para que pueda modificar el struct
```

- Necesita un **puntero** al struct destino
- Si el JSON tiene campos que no existen en el struct, **se ignoran** (no es error)
- Si el struct tiene campos que no están en el JSON, **se quedan con su zero value**

### json.NewEncoder / json.NewDecoder — Streaming

```go
// Escribir JSON directamente a un Writer (archivo, red, stdout)
encoder := json.NewEncoder(os.Stdout)
encoder.Encode(usuario)

// Leer JSON directamente de un Reader (archivo, red, stdin)
decoder := json.NewDecoder(resp.Body)
decoder.Decode(&usuario)
```

**Ventaja sobre Marshal/Unmarshal:** No necesitan cargar todo el JSON en memoria. Ideales para APIs con respuestas grandes.

---

## 🌐 Paquete `net/http` — El Teléfono Universal

Go tiene un **cliente y servidor HTTP integrados** en la librería estándar. No necesitas frameworks externos para tareas básicas.

### GET Request

```go
resp, err := http.Get("https://api.ejemplo.com/datos")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close() // ¡Siempre cerrar el Body!

// Leer el body
body, err := io.ReadAll(resp.Body)
```

### POST Request con JSON

```go
datos := []byte(`{"nombre": "Ana", "edad": 25}`)

resp, err := http.Post(
    "https://api.ejemplo.com/usuarios",
    "application/json",           // Content-Type
    bytes.NewBuffer(datos),       // Body como Reader
)
defer resp.Body.Close()
```

### Cliente Personalizado (con timeout)

```go
cliente := &http.Client{
    Timeout: 10 * time.Second,
}

req, _ := http.NewRequest("GET", url, nil)
req.Header.Set("Authorization", "Bearer token123")
req.Header.Set("Accept", "application/json")

resp, err := cliente.Do(req)
```

### Códigos de Estado HTTP

| Código | Significado | Analogía |
|:-------|:------------|:---------|
| `200` | OK | "Aquí tienes lo que pediste" |
| `201` | Created | "Creé lo que me pediste" |
| `400` | Bad Request | "No entiendo lo que me pides" |
| `401` | Unauthorized | "No te conozco, identifícate" |
| `403` | Forbidden | "Te conozco, pero no tienes permiso" |
| `404` | Not Found | "Eso que buscas no existe aquí" |
| `500` | Internal Server Error | "Se me rompió algo por dentro" |

---

## 📐 JSON Anidado y Arrays

Los JSON reales suelen tener **estructuras anidadas** y **arrays**. Veamos cómo modelarlos:

```json
{
  "nombre": "Tech Corp",
  "fundada": 2010,
  "direccion": {
    "ciudad": "Lima",
    "pais": "Peru"
  },
  "empleados": [
    {"nombre": "Ana", "rol": "dev"},
    {"nombre": "Luis", "rol": "devops"}
  ]
}
```

```go
type Direccion struct {
    Ciudad string `json:"ciudad"`
    Pais   string `json:"pais"`
}

type Empleado struct {
    Nombre string `json:"nombre"`
    Rol    string `json:"rol"`
}

type Empresa struct {
    Nombre    string      `json:"nombre"`
    Fundada   int         `json:"fundada"`
    Direccion Direccion   `json:"direccion"`
    Empleados []Empleado  `json:"empleados"`
}
```

**La clave:** Cada nivel de anidación es un struct separado que se compone. Arrays JSON se mapean a slices de Go.

---

## 🧩 Custom Unmarshaling — Cuando Necesitas Control Total

A veces el JSON no coincide con lo que necesitas en Go. Por ejemplo, una API devuelve fechas como strings:

```json
{"nombre": "Ana", "nacimiento": "1995-03-15T00:00:00Z"}
```

Quieres que `nacimiento` sea un `time.Time` en Go, no un string:

```go
type FechaAPI struct {
    time.Time
}

func (f *FechaAPI) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    t, err := time.Parse(time.RFC3339, s)
    if err != nil {
        return err
    }
    f.Time = t
    return nil
}

type Persona struct {
    Nombre     string   `json:"nombre"`
    Nacimiento FechaAPI `json:"nacimiento"`
}
```

**¿Por qué es útil?** Porque las APIs del mundo real no siempre devuelven datos en el formato que necesitas. El custom unmarshaling te da control total sobre la transformación.

---

## 📝 Ejercicio Práctico: Cliente de API REST con Cache Local

### ¿Qué construimos?

Un CLI que consume la **PokeAPI** (https://pokeapi.co/), una API pública gratuita con datos de Pokémon. Nuestra herramienta:

1. **Consulta** datos de Pokémon por nombre
2. **Deserializa** la respuesta JSON en structs tipados
3. **Cachea** respuestas en un archivo JSON local (evita requests repetidos)
4. **Muestra** los datos formateados en una tabla bonita en consola

### Arquitectura del Código

```
📁 17-json-apis/
├── go.mod
└── main.go          ← Todo el código en un solo archivo
```

### Estructuras de Datos (PokeAPI)

La PokeAPI devuelve JSON anidado. Modelamos solo los campos que nos interesan:

```go
type PokemonResponse struct {
    ID       int            `json:"id"`
    Nombre   string         `json:"name"`
    Altura   int            `json:"height"`
    Peso     int            `json:"weight"`
    Tipos    []TipoSlot     `json:"types"`
    Stats    []StatEntry    `json:"stats"`
    Sprites  Sprites        `json:"sprites"`
}

type TipoSlot struct {
    Slot int  `json:"slot"`
    Tipo Tipo `json:"tipo"`
}

type Tipo struct {
    Nombre string `json:"name"`
}

type StatEntry struct {
    BaseStat int  `json:"base_stat"`
    Stat     Stat `json:"stat"`
}

type Stat struct {
    Nombre string `json:"name"`
}

type Sprites struct {
    FrontDefault string `json:"front_default"`
}
```

### Sistema de Cache

```go
type CacheEntry struct {
    Datos     PokemonResponse `json:"datos"`
    Consulta  time.Time       `json:"consulta"`
}

type CacheLocal struct {
    Entradas map[string]CacheEntry `json:"entradas"`
    Archivo  string
}
```

El cache guarda respuestas en un archivo JSON. Antes de hacer un request HTTP, verifica si ya tenemos los datos cacheados. Esto ahorra ancho de banda y es más rápido.

### Ejecución

```bash
# Buscar un Pokémon
go run main.go pikachu

# Buscar otro
go run main.go charizard

# Buscar de nuevo (viene del cache, sin request HTTP)
go run main.go pikachu

# Modo demostración (explica JSON y tags)
go run main.go -demo

# Modo verbose (muestra el JSON crudo)
go run main.go -verbose bulbasaur
```

---

## 🔍 Conceptos Clave Explicados

### ¿Por qué `json.Unmarshal` necesita puntero?

```go
var u Usuario
json.Unmarshal(data, u)    // ❌ MALO: copia, no modifica
json.Unmarshal(data, &u)   // ✅ BIEN: modifica el original
```

Es como pedirle a alguien que reorganice tu maleta. Si le das la maleta directamente, organiza una copia. Si le das la **ubicación** de tu maleta (&), organiza la original.

### ¿Por qué los campos deben empezar con mayúscula?

```go
type Malo struct {
    nombre string  // ❌ No exportado = invisible para json
}

type Bueno struct {
    Nombre string  // ✅ Exportado = visible para json
}
```

El paquete `encoding/json` usa **reflection** para leer los campos del struct. En Go, la reflection solo puede ver campos **exportados** (mayúscula). Los campos privados son invisibles — ni siquiera necesitan el tag `json:"-"`.

### ¿Qué es el zero value y cómo afecta a `omitempty`?

```go
type Config struct {
    Puerto  int    `json:"puerto,omitempty"`
    Host    string `json:"host,omitempty"`
    Debug   bool   `json:"debug,omitempty"`
}

c := Config{} // Todos los campos en zero value
data, _ := json.Marshal(c)
// Resultado: {}  — todos se omiten

c2 := Config{Puerto: 8080}
data2, _ := json.Marshal(c2)
// Resultado: {"puerto":8080}  — solo aparece puerto
```

| Tipo | Zero Value | `omitempty` lo omite? |
|:-----|:-----------|:----------------------|
| `string` | `""` | ✅ Sí |
| `int`, `float64` | `0` | ✅ Sí |
| `bool` | `false` | ✅ Sí |
| `slice`, `map` | `nil` | ✅ Sí |
| `pointer` | `nil` | ✅ Sí |
| `struct` | (sus campos) | ❌ **NO** se omite nunca |

> ⚠️ **Trampa común:** `omitempty` en un campo de tipo `struct` **nunca lo omite**, porque un struct no tiene zero value "vacío". Para campos opcionales de tipo struct, usa un **puntero** (`*Direccion`).

---

## 🏋️ Ejercicio Feynman

### Instrucciones
Usando la **Técnica Feynman**, explica estos conceptos **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado. Usa analogías de la vida cotidiana.

---

### Ejercicio 1: JSON y Serialización
> Explica qué es JSON y qué significa "serializar" usando la analogía de **traducir una carta** de español a inglés. ¿Por qué necesitamos traducir los datos? ¿Qué se pierde en la traducción?

---

### Ejercicio 2: Struct Tags
> Imagina que los struct tags son **etiquetas en maletas de equipaje** en un aeropuerto. ¿Qué pasa si no pones etiqueta? ¿Qué pasa si pones `json:"-"`? ¿Qué pasa con `omitempty`? Dibuja mentalmente el proceso.

---

### Ejercicio 3: Marshal vs Unmarshal
> Explica `json.Marshal` y `json.Unmarshal` con la analogía de **empacar y desempacar una maleta**. ¿Cuál es el ida? ¿Cuál es el vuelta? ¿Por qué `Unmarshal` necesita un puntero (la dirección de la maleta)?

---

### Ejercicio 4: API REST y HTTP GET
> Si una API REST fuera un **restaurante con menú**, ¿qué representa cada parte? ¿El URL es el menú? ¿El método GET es pedir? ¿La respuesta JSON es el plato? ¿Qué significa un error 404?

---

### Ejercicio 5: Cache Local
> Explica el cache usando la analogía de **anotar la respuesta de una pregunta frecuente en un post-it**. ¿Cuándo es útil el post-it? ¿Cuándo debes preguntar de nuevo? ¿Qué pasa si la respuesta cambió y tu post-it está desactualizado?

---

### Ejercicio 6: Campos no exportados y Reflection
> Imagina que el marshaller es un **mudador ciego** que solo puede leer etiquetas escritas con letras GRANDES. Si la etiqueta está escrita en letras pequeñas (campo no exportado), el mudador no la ve. ¿Por qué Go eligió esta convención? ¿Qué ventaja tiene sobre keywords como `private`?

---

## 📋 Resumen de Funciones Clave

| Función | Paquete | ¿Qué hace? |
|---------|---------|-------------|
| `json.Marshal(v)` | `encoding/json` | Serializa a `[]byte` JSON |
| `json.MarshalIndent(v, prefix, indent)` | `encoding/json` | Serializa con formato legible |
| `json.Unmarshal(data, &v)` | `encoding/json` | Deserializa JSON a struct |
| `json.NewEncoder(w).Encode(v)` | `encoding/json` | Serializa directamente a Writer |
| `json.NewDecoder(r).Decode(&v)` | `encoding/json` | Deserializa directamente de Reader |
| `http.Get(url)` | `net/http` | GET request simple |
| `http.Post(url, contentType, body)` | `net/http` | POST request simple |
| `http.NewRequest(method, url, body)` | `net/http` | Request personalizado |
| `http.Client.Do(req)` | `net/http` | Ejecuta request con cliente custom |
| `io.ReadAll(r)` | `io` | Lee todo un Reader a `[]byte` |
| `resp.Body.Close()` | `net/http` | Cierra el body (¡siempre con defer!) |
| `time.Parse(layout, value)` | `time` | Parsea string a time.Time |
| `time.Format(layout)` | `time` | Formatea time.Time a string |
| `os.WriteFile(path, data, perm)` | `os` | Escribe bytes a archivo |
| `os.ReadFile(path)` | `os` | Lee archivo completo |

---

## 🔗 ¿Qué sigue?

En la siguiente lección exploraremos **`context`, `defer`, `panic/recover`** — las herramientas de control de flujo avanzado que separan al desarrollador junior del senior. Aprenderás a propagar cancellation a través de goroutines, manejar timeouts en llamadas a APIs, y usar `defer` para garantizar la limpieza de recursos.

---

> 💡 *"JSON es el idioma universal de internet. Si hablas JSON, hablas con cualquier sistema."*