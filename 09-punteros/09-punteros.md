# 🧭 Lección 09 — Punteros (Pointers) en Go

## 📖 ¿Qué son los punteros?

Imagina que vives en una ciudad enorme. Cada persona tiene una **dirección postal** donde vive. Si quieres enviarle una carta a Ana, necesitas su dirección. Si quieres visitarla, vas a esa dirección.

Un **puntero** es exactamente eso: una **dirección postal** dentro de la memoria de tu computadora. En lugar de guardar el **valor** directamente, guardas **la dirección** donde ese valor vive.

```
   Variable normal:     x = 42          → Guarda el VALOR 42
   Variable puntero:    p = 0xc000014078 → Guarda la DIRECCIÓN donde vive 42
```

### 🤔 ¿Para qué sirven los punteros?

Hay tres razones principales:

| Razón | Sin puntero | Con puntero |
|-------|------------|-------------|
| **Modificar originales** | Las funciones reciben COPIAS. Los cambios no salen de la función. | Las funciones reciben la DIRECCIÓN. Modifican el valor original directamente. |
| **Eficiencia de memoria** | Copiar una struct grande de 100 campos consume memoria y tiempo. | Pasar una dirección (8 bytes en sistemas de 64 bits) es casi gratis. |
| **Señalar "ausencia"** | ¿Cómo indicas que un int "no tiene valor"? ¿Con -1? ¿Con 0? Ambos son valores válidos. | Un puntero `nil` dice claramente: "aquí no hay nada". |

---

## 🔑 Los dos operadores fundamentales

### `&` — "¿Dónde vives?" (Operador de dirección)

El operador `&` obtiene la **dirección de memoria** de una variable.

```go
x := 42
p := &x   // p ahora contiene la dirección donde vive x
fmt.Println(p)  // Algo como: 0xc000014078
```

> 🧠 **Analogía Feynman**: Es como preguntarle a alguien "¿en qué dirección vives?" y anotarla en tu libreta. La libreta no tiene a la persona dentro, tiene la dirección de su casa.

### `*` — "Ve a esa dirección y trae lo que haya" (Operador de desreferencia)

El operador `*` accede al **valor** que está almacenado en la dirección apuntada.

```go
x := 42
p := &x       // p = dirección de x
fmt.Println(*p)  // 42  → "ve a la dirección de p y trae lo que haya"
```

> 🧠 **Analogía Feynman**: Es como ir a la dirección que anotaste en tu libreta y tocar la puerta. Lo que encuentras ahí es el valor.

### `*` en el TIPO vs `*` en el VALOR

Este es un punto donde muchos se confunden. El mismo símbolo `*` tiene **dos significados diferentes** según el contexto:

```go
// * EN EL TIPO → "esto es un puntero a int"
var p *int

// * EN EL VALOR → "accede al valor apuntado"
fmt.Println(*p)
```

| Contexto | Ejemplo | Significado |
|----------|---------|-------------|
| En la **declaración de tipo** | `*int` | "Esto es un puntero a int" |
| En una **expresión** | `*p` | "Ve a la dirección de p y accede al valor" |

---

## 🏗️ Creando punteros

Hay **tres formas** de crear punteros en Go:

### Forma 1: Con `&` (la más común)

```go
x := 42
p := &x   // p es *int, apunta a x
```

### Forma 2: Con `new()` (crea memoria nueva)

```go
p := new(int)  // Crea un int con valor cero (0) y devuelve *int
*p = 42
```

> `new(T)` crea una variable anónima de tipo T con su **valor cero** y devuelve un puntero a ella. No tiene nombre la variable, solo puedes acceder a ella a través del puntero.

### Forma 3: Retornando un puntero de una función

```go
func crearNumero() *int {
    x := 10
    return &x  // Perfectamente válido en Go
}
```

> 🧠 **¿No se destruye x al salir de la función?** En otros lenguajes como C, esto sería un error (dangling pointer). Pero Go tiene un sistema llamado **escape analysis**: si el compilador detecta que una variable local sigue siendo referenciada después de que la función termina, **mueve esa variable al heap** automáticamente. Tú no tienes que preocuparte por esto.

---

## 🎯 Ejemplo práctico: ¿Por qué necesito punteros?

### El problema: las funciones reciben copias

```go
func intentarDuplicar(valor int) {
    valor = valor * 2  // Solo modifica la COPIA
}

numero := 10
intentarDuplicar(numero)
fmt.Println(numero)  // 10 ← ¡No cambió!
```

> En Go, **todos los argumentos se pasan por valor**. Esto significa que la función siempre recibe una **copia** del dato original.

### La solución: pasar un puntero

```go
func duplicar(valor *int) {
    *valor = *valor * 2  // Modifica el ORIGINAL
}

numero := 10
duplicar(&numero)
fmt.Println(numero)  // 20 ← ¡Sí cambió!
```

> 🧠 **Analogía Feynman**: Imagina que tienes un libro en tu casa.
> - **Sin puntero**: Le das una FOTOCOPIA del libro a un amigo. Él puede rayarla, romperla, lo que quiera. Tu libro original no cambia.
> - **Con puntero**: Le das a tu amigo la DIRECCIÓN de tu casa. Él va y raya TU libro original. Cuando vuelves, tu libro tiene las marcas.

---

## 🧱 Punteros con Structs

Los punteros se usan **constantemente** con structs en Go. Es la forma estándar de trabajar con estructuras de datos.

```go
type Persona struct {
    Nombre string
    Edad   int
}

func cumpleaños(p *Persona) {
    p.Edad++  // Go permite p.Edad en vez de (*p).Edad
}

ana := Persona{Nombre: "Ana", Edad: 28}
cumpleaños(&ana)
fmt.Println(ana.Edad)  // 29
```

### 📌 Nota importante: Atajo sintáctico

Go permite escribir `p.Edad` en lugar de `(*p).Edad`. Es un **atajo sintáctico** automático. Ambos significan lo mismo, pero `p.Edad` es mucho más legible.

---

## 🚫 Nil — El valor cero de los punteros

El valor cero de un puntero es `nil`, que significa "no apunta a nada".

```go
var p *int       // p es nil (no apunta a ningún lugar)
fmt.Println(p == nil)  // true
```

### ⚠️ Peligro: Desreferenciar un puntero nil

```go
var p *int
fmt.Println(*p)  // PANIC: runtime error: invalid memory address
```

> 🧠 **Regla de oro**: Antes de usar `*p`, siempre verifica que `p != nil`.

```go
if p != nil {
    fmt.Println(*p)  // Seguro
}
```

---

## 📦 Punteros y Slices

Los slices en Go ya son, internamente, una estructura que contiene un **puntero** a un array subyacente. Por eso:

```go
func modificar(nums []int) {
    nums[0] = 999  // SÍ se refleja afuera
}

func reasignar(nums []int) {
    nums = append(nums, 888)  // NO se refleja afuera
}
```

| Operación | ¿Se refleja afuera? | ¿Por qué? |
|-----------|---------------------|------------|
| Modificar un elemento existente | ✅ Sí | El puntero interno del slice sigue apuntando al mismo array |
| `append` que NO crece el array | ✅ Sí | Se modifica el mismo array subyacente |
| `append` que SÍ crece el array | ❌ No | Go crea un NUEVO array y el puntero interno cambia, pero solo dentro de la función |

> Si necesitas que un `append` se refleje afuera, pasa un puntero al slice (`*[]int`) o retorna el slice modificado.

---

## 🔄 Uso práctico: Swap (intercambio de valores)

```go
func swap(a, b *int) {
    *a, *b = *b, *a
}

x := 100
y := 200
swap(&x, &y)
fmt.Println(x, y)  // 200 100
```

> Sin punteros, esto sería imposible en Go (no hay forma de "devolver" dos valores modificados sin usar tuplas de retorno).

---

## 🧭 Puntero a puntero

Un puntero puede apuntar a otro puntero, creando niveles de indirección:

```go
x := 42
p := &x    // *int      → apunta a x
pp := &p   // **int     → apunta a p, que apunta a x

fmt.Println(**pp)  // 42  → dereferencia dos veces para llegar a x
```

> 🧠 **Analogía**: Es como una cadena de direcciones:
> - `x` vive en la Calle 1
> - `p` es una libreta que dice "ve a la Calle 1"
> - `pp` es otra libreta que dice "busca la libreta de `p`, y en ella encontrarás la dirección de `x`"

En la práctica, raramente necesitas más de un nivel de indirección (`*` o `**`). Más de dos niveles suele ser señal de un diseño confuso.

---

## 🆚 `new()` vs `&variable`

| Característica | `new(T)` | `&variable` |
|----------------|----------|-------------|
| ¿Qué hace? | Crea una variable **anónima** con valor cero | Toma la dirección de una variable **existente** |
| Valor inicial | Cero del tipo (0, "", false, nil...) | El valor que ya tenía la variable |
| ¿Tiene nombre? | No, solo accedes por el puntero | Sí |
| Uso típico | Cuando solo necesitas un puntero rápido | Cuando ya tienes una variable |

```go
// Con new()
p := new(int)    // *int apuntando a un 0
*p = 42          // Ahora apunta a 42

// Con &
x := 42
p := &x          // *int apuntando a x que vale 42
```

---

## 🧩 Punteros como valores "opcionales"

Go no tiene `null` como Java o JavaScript. Pero a veces necesitas representar la idea de "no hay valor". Los punteros nil son la solución estándar:

```go
func buscarUsuario(nombre string) *Persona {
    // ... buscar en base de datos ...
    if encontrado {
        return &persona  // Retornamos puntero a la persona
    }
    return nil  // "No encontrado"
}

resultado := buscarUsuario("Ana")
if resultado != nil {
    fmt.Println("Encontrado:", resultado.Nombre)
} else {
    fmt.Println("No existe ese usuario.")
}
```

> Este patrón es **extremadamente común** en Go. Lo verás en bibliotecas estándar, APIs, bases de datos, etc.

---

## 📋 Resumen visual

```
  ┌─────────────────────────────────────────────────────┐
  │                  PUNTEROS EN GO                      │
  ├─────────────────────────────────────────────────────┤
  │                                                     │
  │  &variable  →  obtiene la dirección de memoria      │
  │  *puntero   →  accede al valor en esa dirección     │
  │  *Tipo      →  declara un tipo puntero              │
  │                                                     │
  │  Valor cero de un puntero: nil                      │
  │                                                     │
  │  ⚠️  Desreferenciar nil → PANIC (crash)             │
  │  ✅  Siempre verifica nil antes de usar *p           │
  │                                                     │
  │  Pasar puntero a función → modifica el original     │
  │  Retornar puntero de función → escape analysis      │
  │                                                     │
  │  new(T) → crea puntero a T con valor cero           │
  │  &x     → puntero a la variable x existente         │
  │                                                     │
  └─────────────────────────────────────────────────────┘
```

---

## 🧪 Ejercicio Feynman

> **Instrucción**: Explica estos conceptos con tus propias palabras, como si le enseñarás a alguien que nunca ha programado. Usa tus propias analogías. Si no puedes explicarlo de forma simple, es que no lo entendiste lo suficiente.

### ✍️ Ejercicio 1: Explica con una analogía
> "¿Qué es un puntero y por qué es útil?"

**Tu respuesta** (escribe aquí):
```
Tu analogía aquí...

```

---

### ✍️ Ejercicio 2: ¿Qué imprime este código?
```go
func misterio(a *int, b *int) {
    *a = *a + *b
    *b = *a - *b
}

x := 5
y := 3
misterio(&x, &y)
fmt.Println(x, y)
```

**Tu respuesta** (escribe aquí):
```
x = ???
y = ???
Explica por qué:

```

<details>
<summary>🔍 Ver solución</summary>

```
x = 8
y = 5
```

**Explicación paso a paso:**
1. `*a = *a + *b` → `*a = 5 + 3 = 8` → x ahora es 8
2. `*b = *a - *b` → `*b = 8 - 3 = 5` → y ahora es 5

Nota: Como `*a` ya cambió en el paso 1, en el paso 2 se usa el nuevo valor (8).
</details>

---

### ✍️ Ejercicio 3: ¿Por qué este código NO funciona como esperas?
```go
func agregarNumero(nums []int, num int) []int {
    nums = append(nums, num)
    return nums
}

lista := []int{1, 2, 3}
agregarNumero(lista, 4)
fmt.Println(lista)  // ¿Qué imprime?
```

**Tu respuesta** (escribe aquí):
```
Imprime: ???
¿Por qué no incluye el 4?

```

<details>
<summary>🔍 Ver solución</summary>

```
Imprime: [1 2 3]
```

**Explicación**: Aunque los slices contienen un puntero interno, `append` puede crear un **nuevo array** si no hay capacidad suficiente. Pero más importante: el retorno del `append` **no se está guardando**. La solución correcta sería:

```go
lista = agregarNumero(lista, 4)
```

O retornar el slice y reasignarlo. El puntero interno del slice original no cambió.

</details>

---

### ✍️ Ejercicio 4: Implementa tu propio código
> Crea una función `actualizarPerfil(p *Persona, nuevoNombre string, nuevaEdad int)` que modifique directamente los campos de una Persona. Pruébala en `main()`.

**Tu respuesta** (escribe aquí):
```go
// Escribe tu código aquí

```

<details>
<summary>🔍 Ver solución sugerida</summary>

```go
type Persona struct {
    Nombre string
    Edad   int
}

func actualizarPerfil(p *Persona, nuevoNombre string, nuevaEdad int) {
    p.Nombre = nuevoNombre  // Atajo para (*p).Nombre
    p.Edad = nuevaEdad
}

func main() {
    persona := Persona{Nombre: "Ana", Edad: 28}
    fmt.Println("Antes:", persona)
    actualizarPerfil(&persona, "María", 30)
    fmt.Println("Después:", persona)
    // Output: Después: {María 30}
}
```

</details>

---

### ✍️ Ejercicio 5: Explica el concepto clave
> "¿Cuál es la diferencia entre pasar un valor y pasar un puntero a una función? ¿Cuándo usarías cada uno?"

**Tu respuesta** (escribe aquí):
```
Diferencia:
¿Cuándo usar valor?: 
¿Cuándo usar puntero?: 

```

<details>
<summary>🔍 Ver respuesta sugerida</summary>

**Diferencia:**
- **Pasar valor**: La función recibe una COPIA. Los cambios no afectan al original.
- **Pasar puntero**: La función recibe la DIRECCIÓN. Los cambios SÍ afectan al original.

**¿Cuándo usar valor?**
- Cuando son tipos pequeños (int, bool, float64, string)
- Cuando NO necesitas modificar el original
- Cuando quieres que la función trabaje con su propia copia aislada

**¿Cuándo usar puntero?**
- Cuando necesitas modificar el valor original
- Cuando el tipo es grande (structs con muchos campos) y copiarlo sería ineficiente
- Cuando necesitas representar "ausencia de valor" (nil)
- Cuando implementas métodos que modifican el receptor (lo veremos en la lección de interfaces)

</details>

---

## 🚀 Próxima lección

En la **Lección 10** exploraremos los **Métodos y la Programación Orientada a Objetos en Go**: cómo asociar funciones a tipos, receptores por valor vs por puntero, y cómo Go implementa la orientación a objetos sin clases.

---

> *"Si no puedes explicarlo de forma simple, es que no lo entendiste lo suficiente."*
> — Richard Feynman 🧠