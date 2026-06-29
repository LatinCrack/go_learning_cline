# 🧬 Lección 03 — Tipos de Datos en Go

> **Método Feynman:** Si no puedes explicárselo a un niño, no lo entiendes bien.

---

## 🤔 ¿Qué son los tipos de datos?

Imagina que vas al supermercado. Compras:

- 🥛 Leche → viene en **litros** (medida líquida)
- 🍞 Pan → viene en **unidades** (se cuenta por piezas)
- 🧀 Queso → viene en **gramos** (medida de peso)

Cada cosa tiene un **tipo de medición** distinto. No puedes medir el pan en litros ni la leche en unidades. **El tipo define cómo se almacena, cuánto espacio ocupa y qué operaciones puedes hacer con él.**

En programación pasa exactamente igual. Los **tipos de datos** son las "formas de medir" la información que tu programa maneja.

> **Analogía clave:** El tipo de dato es como el **tamaño del contenedor** en una mudanza. No guardas un sofá en una caja de zapatos, ni pones un anillo en una caja de nevera. Cada cosa necesita su contenedor adecuado.

---

## 📦 Los tipos primitivos de Go

Go es un lenguaje **fuertemente tipado**. Esto significa que una vez que le dices "aquí hay un número entero", Go no te deja meterle un texto de forma accidental. Es como un almacén con estantes etiquetados: el estante de "enteros" solo acepta enteros.

### 🔢 Tipos numéricos enteros

| Tipo       | Tamaño    | Rango aproximado                        | Analogía               |
|------------|-----------|-----------------------------------------|------------------------|
| `int8`     | 8 bits    | -128 a 127                              | 🪙 Monedas en el bolsillo |
| `int16`    | 16 bits   | -32,768 a 32,767                        | 📖 Páginas de un cuaderno |
| `int32`    | 32 bits   | ≈ -2 mil millones a 2 mil millones      | 🏟️ Personas en una ciudad |
| `int64`    | 64 bits   | ≈ -9.2 quintillones a 9.2 quintillones  | 🌌 Estrellas en la galaxia |
| `int`      | Plataforma| 32 o 64 bits según el sistema           | 📏 La regla "estándar" |
| `uint`     | Plataforma| Solo positivos (0 en adelante)          | 🚫📏 Regla sin negativos |

**`int` vs `uint`:** La diferencia es simple. `int` permite negativos (como un termómetro: -5°C). `uint` solo permite positivos (como una báscula: no puedes pesar -3 kg).

> **Consejo Feynman:** En el 99% de los casos, usa `int`. Solo necesitas tipos específicos cuando trabajas con protocolos de red, archivos binarios o optimización de memoria extrema.

### 🔢 Tipos numéricos de punto flotante (decimales)

| Tipo        | Precisión       | Analogía                          |
|-------------|-----------------|-----------------------------------|
| `float32`   | ~7 dígitos      | 📐 Regla escolar (precisión básica) |
| `float64`   | ~15 dígitos     | 🔬 Microscopio (precisión alta)   |

**¿Por qué dos tipos de decimales?** Piensa en una calculadora de juguete vs una científica. La de juguete (`float32`) redondea antes: `3.14159265` podría verse como `3.141593`. La científica (`float64`) mantiene más cifras.

> **Regla de oro:** Siempre usa `float64` por defecto. Go usa `float64` como predeterminado cuando escribes `3.14`.

### 📝 Tipo texto (string)

```go
var saludo string = "Hola, mundo"
```

Un `string` en Go es una **secuencia inmutable de bytes** (generalmente codificados en UTF-8).

**Analogía:** Piensa en un `string` como una **tira de papel con letras pegadas**. Puedes leerla, pero no puedes cambiar una letra sin crear una tira nueva.

**Concepto importante — Runa (`rune`):**
En Go existe el tipo `rune`, que equivale a `int32` y representa un **único carácter Unicode**.

- `'A'` → rune que vale 65 (código ASCII)
- `'ñ'` → rune que vale 241
- `'😀'` → rune que vale 128512

> **¿Por qué importa?** Porque en español usamos tildes y eñes. Go maneja esto perfectamente gracias a UTF-8, pero necesitas entender que un "carácter visible" puede ocupar más de un byte.

### ✅ Tipo booleano

```go
var esVerdad bool = true
var esMentira bool = false
```

Solo dos valores posibles: `true` o `false`. Como un interruptor de luz: encendido o apagado.

**Analogía:** El booleano es la **pregunta más simple del universo**. ¿Está lloviendo? Sí o no. ¿Tienes hambre? Sí o no. No hay término medio.

---

## 🏗️ Cero valores: Go te regala un inicio

Go tiene una característica única y elegante: **toda variable tiene un "valor cero" por defecto** cuando no le asignas nada.

| Tipo      | Valor cero | Analogía                          |
|-----------|------------|-----------------------------------|
| `int`     | `0`        | Termómetro en cero grados         |
| `float64` | `0.0`      | Báscula sin nada encima           |
| `string`  | `""`       | Pizarra borrada                   |
| `bool`    | `false`    | Interruptor apagado               |

> **¿Por qué es importante?** En otros lenguajes como C o Java, una variable sin inicializar puede contener "basura" (datos aleatorios de memoria). Go **nunca** permite esto. Siempre hay un valor predecible. Es como un hotel que siempre deja las habitaciones limpias antes de recibir al siguiente huésped.

---

## 🔬 Inferencia de tipos con `:=`

Cuando usas `:=`, Go **adivina** el tipo basándose en el valor que le asignas:

```go
edad := 25           // Go dice: "¡Esto es un int!"
precio := 19.99      // Go dice: "¡Esto es un float64!"
nombre := "Carlos"   // Go dice: "¡Esto es un string!"
activo := true       // Go dice: "¡Esto es un bool!"
```

**Analogía:** Es como cuando llegas a un restaurante y dices "quiero lo de siempre". El mesero (Go) **ya sabe** qué tipo de plato esperas según tu historial. No necesitas ser explícito cada vez.

---

## 🔄 Conversión de tipos (Type Casting)

Go **NO** convierte tipos automáticamente. Tú debes hacerlo explícitamente:

```go
var entero int = 42
var decimal float64 = float64(entero)  // Convierto int → float64

var pi float64 = 3.14
var truncado int = int(pi)              // Convierto float64 → int (pierde decimales: queda 3)
```

**Analogía:** Es como cambiar de moneda. No puedes pagar en euros directamente en Perú. Tienes que ir al banco (la función de conversión) y cambiar euros a soles. Go te obliga a ir al banco.

> **⚠️ Advertencia:** Al convertir `float64` a `int`, la parte decimal se **trunca** (se corta, NO se redondea). `int(3.99)` da `3`, no `4`.

---

## 🎯 Ejercicio práctico: El inventario de la ferretería

Vamos a construir un programa que gestiona información de productos en una ferretería. Este ejercicio usa **todos los tipos de datos** que hemos aprendido.

**Crea el archivo `main.go` dentro de la carpeta `03-tipos-de-datos/`:**

```go
package main

import "fmt"

func main() {

	// ====================================================
	// 🏪 SISTEMA DE INVENTARIO - FERRETERÍA "DON CLODO"
	// ====================================================

	// --- Datos del producto 1: Martillo ---

	var nombreProducto1 string = "Martillo de uña"
	var precioProducto1 float64 = 25.50
	var cantidadProducto1 int = 15
	var enPromocion1 bool = true

	// --- Datos del producto 2: Destornillador ---

	var nombreProducto2 string = "Destornillador Phillips"
	var precioProducto2 float64 = 12.75
	var cantidadProducto2 int = 40
	var enPromocion2 bool = false

	// --- Datos del producto 3: Cinta métrica ---

	nombreProducto3 := "Cinta métrica 5m"      // Inferencia: string
	precioProducto3 := 18.90                    // Inferencia: float64
	cantidadProducto3 := 22                     // Inferencia: int
	enPromocion3 := true                        // Inferencia: bool

	// ====================================================
	// 📊 MOSTRAR INVENTARIO
	// ====================================================

	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║   🏪 INVENTARIO - FERRETERÍA DON CLODO      ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	// Producto 1 con fmt.Printf (formato controlado)
	fmt.Printf("📦 Producto:    %s\n", nombreProducto1)
	fmt.Printf("   💰 Precio:    $%.2f\n", precioProducto1)
	fmt.Printf("   📦 Cantidad:  %d unidades\n", cantidadProducto1)
	fmt.Printf("   🏷️  Promoción: %t\n", enPromocion1)
	fmt.Println("   ─────────────────────────────")

	// Producto 2
	fmt.Printf("📦 Producto:    %s\n", nombreProducto2)
	fmt.Printf("   💰 Precio:    $%.2f\n", precioProducto2)
	fmt.Printf("   📦 Cantidad:  %d unidades\n", cantidadProducto2)
	fmt.Printf("   🏷️  Promoción: %t\n", enPromocion2)
	fmt.Println("   ─────────────────────────────")

	// Producto 3
	fmt.Printf("📦 Producto:    %s\n", nombreProducto3)
	fmt.Printf("   💰 Precio:    $%.2f\n", precioProducto3)
	fmt.Printf("   📦 Cantidad:  %d unidades\n", cantidadProducto3)
	fmt.Printf("   🏷️  Promoción: %t\n", enPromocion3)
	fmt.Println("   ─────────────────────────────")

	// ====================================================
	// 🔄 CONVERSIÓN DE TIPOS
	// ====================================================

	fmt.Println()
	fmt.Println("🔄 Demostración de conversión de tipos:")
	fmt.Println()

	// Tenemos la cantidad como int, pero necesitamos hacer cálculos decimales
	var cantidadFloat float64 = float64(cantidadProducto1)
	iva := cantidadFloat * 0.18
	fmt.Printf("   IVA sobre %d martillos: $%.2f\n", cantidadProducto1, iva)

	// Tenemos un precio float64, pero necesitamos redondear a entero para etiquetas
	var precioEntero int = int(precioProducto1)
	fmt.Printf("   Precio para etiqueta (sin decimales): $%d\n", precioEntero)

	// ====================================================
	// 🎯 VALORES CERO
	// ====================================================

	fmt.Println()
	fmt.Println("🎯 Valores cero (variables sin inicializar):")
	fmt.Println()

	var sinEntero int
	var sinDecimal float64
	var sinTexto string
	var sinBooleano bool

	fmt.Printf("   int sin valor:    %d\n", sinEntero)
	fmt.Printf("   float64 sin valor: %f\n", sinDecimal)
	fmt.Printf("   string sin valor:  [%s]\n", sinTexto)
	fmt.Printf("   bool sin valor:    %t\n", sinBooleano)

	// ====================================================
	// 🔠 RUNAS (CARACTERES UNICODE)
	// ====================================================

	fmt.Println()
	fmt.Println("🔤 Runas (caracteres Unicode):")
	fmt.Println()

	var letraA rune = 'A'
	var letraEne rune = 'ñ'
	var emoji rune = '😀'

	fmt.Printf("   'A'   → valor numérico: %d\n", letraA)
	fmt.Printf("   'ñ'   → valor numérico: %d\n", letraEne)
	fmt.Printf("   '😀'  → valor numérico: %d\n", emoji)

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("   ✅ Inventario cargado correctamente")
	fmt.Println("═══════════════════════════════════════════════")
}
```

### 🧠 Explicación línea por línea del código clave

**Declaraciones con `var` explícito:**

```go
var nombreProducto1 string = "Martillo de uña"
```

Le decimos a Go: "Crea una variable llamada `nombreProducto1`, de tipo `string`, con el valor `Martillo de uña`". Es como llenar una etiqueta de almacén: nombre, tipo de contenedor y contenido.

**Inferencia con `:=`:**

```go
nombreProducto3 := "Cinta métrica 5m"
```

Go mira el valor entre comillas y deduce: "Esto es un `string`". No necesitamos escribir el tipo. Es la forma **más común** de declarar variables en Go.

**`fmt.Printf` y los verbos de formato:**

```go
fmt.Printf("💰 Precio: $%.2f\n", precioProducto1)
```

- `%s` → imprime un string
- `%d` → imprime un entero
- `%f` → imprime un decimal (`%.2f` = 2 decimales)
- `%t` → imprime un booleano (`true`/`false`)
- `\n` → salto de línea

**Conversión explícita:**

```go
var cantidadFloat float64 = float64(cantidadProducto1)
```

Tomamos un `int` (15) y lo convertimos a `float64` (15.0) para poder multiplicarlo por el IVA (0.18). Sin esta conversión, Go lanza error.

**Valores cero:**

```go
var sinEntero int
```

Declaramos un `int` sin asignar valor. Go le pone `0` automáticamente. Esto demuestra que **Go nunca deja variables con "basura" en memoria**.

**Runas:**

```go
var letraA rune = 'A'
```

Las comillas simples `' '` denotan una runa (un solo carácter). El valor almacenado es el **código numérico Unicode**: 'A' = 65, 'ñ' = 241, '😀' = 128512.

---

### ▶️ Ejecución

```bash
cd 03-tipos-de-datos
go run main.go
```

**Salida esperada:**

```
╔══════════════════════════════════════════════╗
║   🏪 INVENTARIO - FERRETERÍA DON CLODO      ║
╚══════════════════════════════════════════════╝

📦 Producto:    Martillo de uña
   💰 Precio:    $25.50
   📦 Cantidad:  15 unidades
   🏷️  Promoción: true
   ─────────────────────────────
📦 Producto:    Destornillador Phillips
   💰 Precio:    $12.75
   📦 Cantidad:  40 unidades
   🏷️  Promoción: false
   ─────────────────────────────
📦 Producto:    Cinta métrica 5m
   💰 Precio:    $18.90
   📦 Cantidad:  22 unidades
   🏷️  Promoción: true
   ─────────────────────────────

🔄 Demostración de conversión de tipos:

   IVA sobre 15 martillos: $2.70
   Precio para etiqueta (sin decimales): $25

🎯 Valores cero (variables sin inicializar):

   int sin valor:    0
   float64 sin valor: 0.000000
   string sin valor:  []
   bool sin valor:    false

🔤 Runas (caracteres Unicode):

   'A'   → valor numérico: 65
   'ñ'   → valor numérico: 241
   '😀'  → valor numérico: 128512

═══════════════════════════════════════════════
   ✅ Inventario cargado correctamente
═══════════════════════════════════════════════
```

---

## 🏋️ Ejercicio Feynman

> **Instrucción:** Toma una hoja en blanco o abre un archivo de texto vacío. Sin consultar esta lección, intenta responder cada pregunta **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado.

### 📝 Preguntas para explicar desde cero:

1. **¿Qué es un tipo de dato?** Explica con una analogía cotidiana (no uses la analogía del supermercado de esta lección, crea la tuya propia).

2. **¿Por qué `int` y `float64` son tipos diferentes si ambos son "números"?** ¿Qué pasaría si Go los tratara como iguales?

3. **Explica qué es el "valor cero" de Go con un ejemplo de la vida real.** ¿Por qué es una ventaja sobre otros lenguajes?

4. **Un amigo te dice:** *"Declaré `var edad = "25"` y luego intenté sumarle 5, pero Go me da error. ¿Por qué?"* Respóndele como si fuera tu alumno.

5. **¿Qué diferencia hay entre `:=` y `var ... = ...`?** ¿Cuándo usarías cada una?

6. **Escribe en papel qué imprime este código SIN ejecutarlo** (predice la salida):
   ```go
   var a int
   var b float64
   var c string
   var d bool
   fmt.Println(a, b, c, d)
   ```

### ✅ Criterio de autoevaluación:

| Criterio                                           | ¿Lo lograste? |
|----------------------------------------------------|:-------------:|
| Explicaste tipos de datos sin mirar la lección     | ⬜ Sí / ⬜ No |
| Creaste tu propia analogía original                | ⬜ Sí / ⬜ No |
| Entendes por qué Go no convierte tipos solo       | ⬜ Sí / ⬜ No |
| Puedes predecir los valores cero sin ayuda         | ⬜ Sí / ⬜ No |
| Respondiste la pregunta del amigo correctamente    | ⬜ Sí / ⬜ No |
| Diferencias `:=` de `var` con claridad             | ⬜ Sí / ⬜ No |

---

## 🗺️ Próxima lección

En la **Lección 04** exploraremos las **constantes** en Go: valores que no cambian, como la velocidad de la luz o el número PI. Veremos `const`, `iota` para enumeraciones, y por qué Go las maneja de forma diferente a otros lenguajes.

> *"El tipo de dato es la primera decisión que tomas sobre cada pieza de información. Si la decisión es correcta, todo lo demás fluye."* — Principio Feynman