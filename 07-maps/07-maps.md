# 🧪 Lección 07 — Maps: Diccionarios, Tablas Hash y el Arte del Acceso O(1)

```
╔════════════════════════════════════════════════════════════════╗
║   🧪 LABORATORIO DE GO — LECCIÓN 07                          ║
║   Maps: Diccionarios, Tablas Hash y el Arte del Acceso O(1)   ║
╚════════════════════════════════════════════════════════════════╝
```

## 📋 Índice de la lección

1. [¿Qué son los Maps?](#1-qué-son-los-maps)
2. [La analogía del diccionario](#2-la-analogía-del-diccionario)
3. [Creación y operaciones básicas](#3-creación-y-operaciones-básicas)
4. [El comma ok idiom](#4-el-comma-ok-idiom)
5. [Iteración y ordenamiento](#5-iteración-y-ordenamiento)
6. [El zero value y los nil maps](#6-el-zero-value-y-los-nil-maps)
7. [Maps como conjuntos (Sets)](#7-maps-como-conjuntos-sets)
8. [Maps anidados y agrupación](#8-maps-anidados-y-agrupación)
9. [Ejercicio práctico: Contador de frecuencia de palabras](#9-ejercicio-práctico-contador-de-frecuencia-de-palabras)
10. [Ejercicio Feynman](#10-ejercicio-feynman)

---

## 1. ¿Qué son los Maps?

Un **map** en Go es una estructura de datos que almacena pares **clave → valor**, donde puedes buscar el valor asociado a una clave de forma casi instantánea.

> 💡 **Concepto clave:** Un map es como un diccionario real: buscas una palabra (clave) y encuentras su definición (valor) sin tener que leer todo el diccionario desde el principio.

### Diferencia con los slices

| Característica | Slice | Map |
|---|---|---|
| **Acceso** | Por índice numérico (0, 1, 2...) | Por clave (string, int, etc.) |
| **Orden** | Ordenado por índice | **NO garantizado** |
| **Búsqueda** | O(n) — recorrer todo | **O(1)** — acceso directo |
| **Zero value** | `nil` (legible y escribible) | `nil` (solo legible ⚠️) |
| **Comparación** | Solo con `==` si mismo tamaño | **NO comparable** directamente |

### ¿Cuándo usar un map?

Usa un map cuando necesites:
- **Buscar datos por una clave** que no es un índice numérico
- **Contar frecuencias** (cuántas veces aparece algo)
- **Agrupar elementos** por categoría
- **Verificar existencia** rápidamente (¿está o no está?)
- **Eliminar duplicados** (como un conjunto/set)

---

## 2. La analogía del diccionario

Imagina que eres un profesor con 1,000 estudiantes y necesitas saber las notas de cada uno.

### ❌ Sin maps (como buscar en una lista desordenada)

Tendrías una lista de tuplas:
```
[(Ana, 9.2), (Bob, 7.5), (Carlos, 8.1), ..., (Zoe, 9.8)]
```

Para encontrar la nota de "Zoe" tendrías que recorrer **toda la lista** hasta encontrarla. En el peor caso, revisas los 1,000 elementos → **O(n)**.

### ✅ Con maps (como un diccionario real)

Tienes un diccionario donde cada nombre lleva directamente a su nota:
```
Ana → 9.2
Bob → 7.5
Carlos → 8.1
...
Zoe → 9.8
```

Para encontrar la nota de "Zoe", el sistema va **directamente** a ella → **O(1)**.

> 🔑 **¿Cómo logra O(1)?** Go usa una **tabla hash** internamente. Cuando dás una clave, Go la pasa por una función hash que produce una dirección de memoria. Es como si el diccionario tuviera una pestaña alfabética: abres directamente en la "Z" y encuentras "Zoe" sin buscar en la "A", "B", "C"...

---

## 3. Creación y operaciones básicas

### Forma 1: Literal (la más común)

```go
edades := map[string]int{
    "Alice":   30,
    "Bob":     25,
    "Charlie": 35,
}
```

### Forma 2: Con make (para datos dinámicos)

```go
ciudades := make(map[string]string)
ciudades["PE"] = "Lima"
ciudades["MX"] = "Ciudad de México"
```

### Forma 3: make con tamaño inicial (optimización)

```go
// Si sabes que tendrás ~100 elementos, reservar espacio evita rehashing
datos := make(map[string]int, 100)
```

### Lectura, escritura y eliminación

```go
// Leer
edad := edades["Alice"]   // 30

// Escribir (agregar o modificar)
edades["Diana"] = 28      // Agrega nueva clave
edades["Alice"] = 31      // Modifica existente

// Eliminar
delete(edades, "Bob")     // Elimina la clave "Bob"

// Longitud
n := len(edades)          // Número de pares clave-valor
```

### ⚠️ Clave inexistente: el zero value

```go
edades := map[string]int{"Alice": 30}
fmt.Println(edades["Zoe"])  // 0 ← ¡No da error! Devuelve el zero value de int
```

Esto es un **peligro sutil**: ¿cómo distinguir entre "Zoe tiene 0 años" y "Zoe no existe"?

---

## 4. El comma ok idiom

Este es uno de los patrones **más importantes** en Go:

```go
edad, existe := edades["Zoe"]
if existe {
    fmt.Printf("Zoe tiene %d años\n", edad)
} else {
    fmt.Println("Zoe no está en el mapa")
}
```

> 💡 **Regla de oro:** Siempre usa `comma ok` cuando el zero value de tu tipo sea un valor válido. Para `int` (0 es válido), para `string` ("" es válido), para `bool` (false es válido).

### Casos donde SÍ necesitas comma ok:

```go
// ¿Existe este usuario?
usuario, ok := usuarios[id]
if !ok {
    return errors.New("usuario no encontrado")
}

// ¿Tiene este producto stock?
stock, ok := inventario[producto]
if !ok || stock == 0 {
    return errors.New("producto sin stock")
}
```

### Casos donde NO lo necesitas:

```go
// Si sabes que la clave siempre existe (o el zero value es aceptable)
fmt.Println(config["debug"])  // Si no existe, "" es aceptable como default
```

---

## 5. Iteración y ordenamiento

### Iterar con range

```go
for clave, valor := range miMap {
    fmt.Printf("%s → %d\n", clave, valor)
}
```

### ⚠️ El orden es ALEATORIO

Go **intencionalmente** randomiza el orden de iteración de los maps. Esto es para que tu código **no dependa** del orden (que podría cambiar entre versiones del lenguaje).

### Ordenar por claves

```go
keys := make([]string, 0, len(miMap))
for k := range miMap {
    keys = append(keys, k)
}
sort.Strings(keys)

for _, k := range keys {
    fmt.Printf("%s → %d\n", k, miMap[k])
}
```

### Ordenar por valores

```go
type kv struct {
    Key   string
    Value int
}
var pares []kv
for k, v := range miMap {
    pares = append(pares, kv{k, v})
}
sort.Slice(pares, func(i, j int) bool {
    return pares[i].Value > pares[j].Value
})
```

---

## 6. El zero value y los nil maps

```go
var m map[string]int  // m es nil

// ✅ Lectura OK (devuelve zero value)
fmt.Println(m["x"])   // 0

// ❌ Escritura PANIC
m["x"] = 1            // PANIC: assignment to entry in nil map
```

> ⚠️ **Regla:** Siempre inicializa un map con `make()` o un literal antes de escribir en él.

---

## 7. Maps como conjuntos (Sets)

Go no tiene un tipo `Set` nativo, pero los maps lo resuelven perfectamente:

### Con bool (simple)

```go
visitados := make(map[string]bool)
visitados["/api/users"] = true
visitados["/api/posts"] = true

if visitados["/api/users"] {
    fmt.Println("Ya visitamos /api/users")
}
```

### Con struct{} (eficiente, 0 bytes por valor)

```go
set := make(map[string]struct{})
set["go"] = struct{}{}
set["rust"] = struct{}{}

if _, ok := set["go"]; ok {
    fmt.Println("go está en el set")
}
```

> 💡 `struct{}` ocupa **0 bytes** en memoria. Es la forma más eficiente de implementar un set cuando solo te importa la existencia de la clave.

---

## 8. Maps anidados y agrupación

### Maps anidados: `map[string]map[string]int`

```go
inventario := map[string]map[string]int{
    "Tienda Centro": {"Laptop": 15, "Mouse": 200},
    "Tienda Norte":  {"Laptop": 8, "Monitor": 50},
}
```

### El patrón estrella: `map[string][]T` (agrupación dinámica)

Este es el patrón **más usado** en Go para agrupar elementos:

```go
porCategoria := make(map[string][]Venta)
for _, v := range ventas {
    porCategoria[v.Categoria] = append(porCategoria[v.Categoria], v)
}
```

> 💡 **¿Por qué es el patrón estrella?** Porque `append` en un slice que es zero value devuelve un slice válido. Si la clave no existe, `porCategoria["Nueva"]` devuelve `nil`, y `append(nil, v)` crea un nuevo slice automáticamente. ¡No necesitas verificar si la clave existe!

---

## 9. Ejercicio práctico: Contador de frecuencia de palabras

El ejercicio de esta lección es un **analizador de texto** que:
1. Limpia y normaliza el texto (minúsculas, sin puntuación)
2. Filtra stopwords (palabras comunes sin significado)
3. Cuenta la frecuencia de cada palabra
4. Muestra el Top-N con barras visuales
5. Analiza bigrams (pares de palabras consecutivas)
6. Agrupa palabras por longitud

### Ejecución

```bash
cd 07-maps
go run main.go
```

### Conceptos aplicados en el ejercicio

| Concepto | Uso en el ejercicio |
|---|---|
| `map[string]int` | Frecuencia de cada palabra |
| `map[string]struct{}` | Set de stopwords (0 bytes por valor) |
| `map[int][]string` | Agrupar palabras por longitud |
| `map[string][]string` | Agrupar archivos por directorio |
| `comma ok` | Verificar si una palabra es stopword |
| `sort.Slice` | Ordenar resultados por frecuencia |
| `range` | Iterar sobre todos los maps |

---

## 10. Ejercicio Feynman

### 🎯 Tu misión: explicar maps como si fueras un profesor

Usando solo palabras simples y analogías (sin jerga técnica), responde:

1. **¿Qué es un map?** Explica con la analogía del diccionario o de una libreta de contactos.

2. **¿Por qué un map es más rápido que un slice para buscar?** Usa la analogía de buscar en una lista vs. un diccionario con pestañas.

3. **¿Qué es el "comma ok"?** Imagina que tienes una máquina expendedora: si pides un producto que no existe, ¿qué pasa? ¿Cómo sabes si no existe o si el precio es $0?

4. **¿Por qué no puedes usar un map nil?** Usa la analogía de un estante vacío: ¿puedes guardar algo en un estante que no existe?

5. **¿Cómo implementarías un "set" con un map?** Usa la analogía de una lista de invitados a una fiesta: solo importa si el nombre está o no en la lista.

### ✅ Criterio de éxito

Tu explicación debe ser comprensible para alguien que **nunca ha programado**. Si usas palabras como "hash", "bucket" o "key-value pair", simplifica aún más. El objetivo es que el concepto quede tan claro que puedas explicarlo de memoria en una conversación casual.

---

```
══════════════════════════════════════════════════════════════════
   ✅ Fin de la Lección 07 — Maps completados
══════════════════════════════════════════════════════════════════