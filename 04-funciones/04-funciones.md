# 🧬 Lección 04 — Funciones: El ADN de Go

> **Método Feynman:** Si no puedes explicárselo a un niño, no lo entiendes bien.

---

## 🤔 ¿Qué es una función?

Imagina que vas a un restaurante. Le das al mesero tu pedido (ingredientes), él lo lleva a la cocina (procesamiento), y te devuelve un plato listo (resultado). No necesitas saber cómo cocinaron tu comida — solo importa que le diste ingredientes correctos y recibiste lo que pediste.

**Una función es exactamente eso: una receta con ingredientes definidos que produce resultados predecibles.**

En programación, las funciones son **bloques de código reutilizables** que:
- Reciben datos de entrada (**parámetros**)
- Ejecutan una lógica específica
- Devuelven datos de salida (**valores de retorno**)

> **Analogía clave:** Una función es como una **máquina expendedora**. Le insertas monedas (parámetros), la máquina procesa tu selección (código interno), y te entrega una bebida (retorno). Puedes usar la misma máquina miles de veces con diferentes monedas y obtener diferentes bebidas — eso es reutilización.

---

## 🔧 Declaración de funciones en Go

Go tiene una sintaxis de funciones **limpia y explícita**. No hay `function`, no hay `def`, no hay `función` — solo `func`.

```go
func nombreFuncion(parametros) tipoRetorno {
    // cuerpo de la función
}
```

### Ejemplo básico: función sin parámetros

```go
func saludar() {
    fmt.Println("¡Hola! Bienvenido al laboratorio de Go")
}
```

Para llamarla:

```go
saludar()  // Imprime: ¡Hola! Bienvenido al laboratorio de Go
```

**Analogía:** Es como tocar el timbre de una casa. No le entregas nada (sin parámetros), pero suena el timbre (la función hace algo).

---

## 📥 Parámetros: Los ingredientes de la receta

Los parámetros son las **entradas** que una función necesita para trabajar. En Go, cada parámetro tiene su **tipo explicitado** — no hay ambigüedad.

```go
func saludarPersona(nombre string) {
    fmt.Printf("¡Hola, %s! Bienvenido al laboratorio\n", nombre)
}
```

Llamarla:

```go
saludarPersona("Carlos")   // ¡Hola, Carlos! Bienvenido al laboratorio
saludarPersona("María")    // ¡Hola, María! Bienvenido al laboratorio
```

**Analogía:** Los parámetros son como los **ingredientes de una receta de cocina**. Si la receta dice "harina (500g)", sabes exactamente qué poner y en qué cantidad. Si pones azúcar en lugar de harina, el compilador de Go te dice "error: tipo incorrecto" — como un chef que se niega a usar sal cuando piden azúcar.

### Múltiples parámetros

```go
func calcularArea(base float64, altura float64) float64 {
    return base * altura
}
```

**Tip:** Cuando varios parámetros tienen el mismo tipo, puedes abreviar:

```go
func calcularArea(base, altura float64) float64 {
    return base * altura
}
```

Esto es equivalente a decir: "tanto `base` como `altura` son `float64`".

---

## 📤 Valores de retorno: El plato servido

### Un solo valor de retorno

```go
func duplicar(numero int) int {
    return numero * 2
}

resultado := duplicar(21)  // resultado = 42
```

**Analogía:** Una función con un retorno es como una máquina expendedora de un solo producto: insertas monedas y recibes exactamente una bebida.

### Múltiples valores de retorno: El superpoder de Go 🦸

Aquí Go se separa de la mayoría de lenguajes. **Go puede devolver múltiples valores desde una función.** Esto es REVOLUCIONARIO y es el pilar del manejo de errores en Go.

```go
func dividir(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("no se puede dividir por cero")
    }
    return a / b, nil
}
```

Llamarla:

```go
resultado, err := dividir(10, 3)
if err != nil {
    fmt.Println("Error:", err)
    return
}
fmt.Printf("Resultado: %.2f\n", resultado)  // Resultado: 3.33
```

**Analogía:** Es como ir al banco. Le das dinero al cajero (parámetros) y esperas dos cosas: tu recibo (resultado) y que no haya habido un error (error). Si todo sale bien, recibes `(recibo, nil)`. Si algo falla, recibes `(0, mensaje de error)`.

> **Concepto clave:** En Go, `nil` significa "nada". Cuando el error es `nil`, significa "no hubo error". Este patrón `(resultado, error)` es el **ADN del manejo de errores en Go** y lo verás absolutamente en TODOS los programas profesionales.

### Valores de retorno nombrados

Go permite **nombrar los valores de retorno**, lo que hace el código más legible:

```go
func calcularEstadisticas(numeros []float64) (promedio, maximo, minimo float64) {
    suma := 0.0
    maximo = numeros[0]
    minimo = numeros[0]
    
    for _, num := range numeros {
        suma += num
        if num > maximo {
            maximo = num
        }
        if num < minimo {
            minimo = num
        }
    }
    
    promedio = suma / float64(len(numeros))
    return  // "naked return" — devuelve los valores nombrados automáticamente
}
```

**Analogía:** Los retornos nombrados son como los **recibos pre-impresos** de una tienda. En vez de escribir a mano qué te devuelven, el recibo ya tiene los campos etiquetados: "Cambio: $5.00", "Productos: 3", "Total pagado: $20.00".

---

## 🔑 Palabra clave `return`

La palabra `return` hace dos cosas:
1. **Devuelve valores** al llamador
2. **Sale de la función** inmediatamente (cualquier código después de `return` no se ejecuta)

```go
func esMayorDeEdad(edad int) bool {
    if edad >= 18 {
        return true    // Sale aquí si es mayor
    }
    return false       // Solo llega aquí si es menor
}
```

**Analogía:** `return` es como la puerta de salida de un edificio. Una vez que pasas por esa puerta, no sigues caminando por el interior — ya estás afuera.

---

## 🌊 Funciones variádicas: Aceptar muchos ingredientes

A veces no sabes de antemano cuántos parámetros necesitas. Las funciones variádicas aceptan un **número variable de argumentos** usando `...`:

```go
func sumar(numeros ...int) int {
    total := 0
    for _, num := range numeros {
        total += num
    }
    return total
}
```

Llamarla con diferentes cantidades de argumentos:

```go
fmt.Println(sumar(1, 2))            // 3
fmt.Println(sumar(1, 2, 3, 4, 5))   // 15
fmt.Println(sumar())                 // 0
```

**Analogía:** Una función variádica es como una **ensalada de la casa**. Le dices al chef "ponle todo lo que tengas" — puede ser lechuga sola, o lechuga con tomate, cebolla, zanahoria y queso. La cantidad de ingredientes varía, pero el proceso (mezclarlos en un bowl) es siempre el mismo.

### Expandir un slice como argumentos

Si ya tienes un slice y quieres pasarlo como argumentos variádicos, usa `...`:

```go
numeros := []int{10, 20, 30, 40, 50}
total := sumar(numeros...)  // Expande el slice como argumentos individuales
fmt.Println(total)           // 150
```

---

## 🔐 Funciones como ciudadanos de primera clase

En Go, las funciones son **valores**. Esto significa que puedes:

### 1. Asignar funciones a variables

```go
operacion := func(a, b int) int {
    return a + b
}

resultado := operacion(5, 3)  // resultado = 8
```

**Analogía:** Es como guardar una receta en una tarjeta. La tarjeta no es la comida, pero contiene las instrucciones para prepararla. Puedes pasar esa tarjeta a alguien más.

### 2. Pasar funciones como parámetros

```go
func aplicarOperacion(a, b int, operacion func(int, int) int) int {
    return operacion(a, b)
}

suma := func(a, b int) int { return a + b }
resta := func(a, b int) int { return a - b }

fmt.Println(aplicarOperacion(10, 5, suma))   // 15
fmt.Println(aplicarOperacion(10, 5, resta))  // 5
```

**Analogía:** Es como un **control remoto universal**. El control (función `aplicarOperacion`) no sabe si le estás poniendo una TV Samsung o Sony — simplemente ejecuta la acción que le pasas (cambiar canal, subir volumen). La "acción" es la función que le entregas.

### 3. Devolver funciones desde otras funciones

```go
func crearMultiplicador(factor int) func(int) int {
    return func(numero int) int {
        return numero * factor
    }
}

duplicar := crearMultiplicador(2)
triplicar := crearMultiplicador(3)

fmt.Println(duplicar(5))    // 10
fmt.Println(triplicar(5))   // 15
```

**Analogía:** Es como una **fábrica de máquinas**. Le dices a la fábrica "crea una máquina que multiplique por 2" y te entrega esa máquina. Luego le dices "crea una que multiplique por 3" y te entrega otra diferente. Cada máquina fue creada con un factor específico.

---

## 🧩 Closures: Funciones que recuerdan

Un **closure** es una función que "captura" y recuerda las variables de su entorno, incluso después de que ese entorno haya dejado de existir.

```go
func crearContador() func() int {
    cuenta := 0  // Esta variable vive "dentro" del closure
    return func() int {
        cuenta++  // La función recuerda y modifica 'cuenta'
        return cuenta
    }
}

contador := crearContador()
fmt.Println(contador())  // 1
fmt.Println(contador())  // 2
fmt.Println(contador())  // 3
```

**Analogía:** Imagina una caja fuerte con un contador dentro. Le das la caja a alguien. Cada vez que abre la caja, el contador sube en 1. La persona no puede ver ni alterar el contador directamente — solo puede abrir la caja y ver el valor. La caja **recuerda** su estado entre aperturas.

> **¿Por qué es útil?** Los closures son la base de patrones como contadores, cachés privadas, callbacks, y middleware en servidores HTTP. Son como "cajitas con memoria" que encapsulan estado privado.

---

## ⚡ Funciones anónimas

Una función anónima es una función **sin nombre**. Se usan cuando necesitas una función pequeña que se usa una sola vez:

```go
// Función anónima que se ejecuta inmediatamente
func() {
    fmt.Println("Ejecutada al instante")
}()

// Función anónima como callback
nums := []int{1, 2, 3, 4, 5}
doble := func(n int) int {
    return n * 2
}

for _, n := range nums {
    fmt.Printf("%d → %d\n", n, doble(n))
}
```

**Analogía:** Una función anónima es como un **post-it**. Le escribes instrucciones rápidas, lo usas una vez, y lo tiras. No necesitas un cuaderno de recetas (función con nombre) para algo que haces una sola vez.

---

## 🛡️ Manejo de errores al estilo Go

Go no tiene excepciones (`try/catch`). En su lugar, los **errores son valores** que se devuelven como parte del resultado de una función.

### El tipo `error`

```go
import "errors"

func dividirSeguro(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("división por cero no permitida")
    }
    return a / b, nil
}
```

### Patrón `if err != nil`

Este es el patrón más repetido en todo el código Go del planeta:

```go
resultado, err := dividirSeguro(10, 0)
if err != nil {
    fmt.Println("❌ Error:", err)
    // Manejar el error: log, devolver, abortar...
    return
}
fmt.Printf("✅ Resultado: %.2f\n", resultado)
```

**Analogía:** Es como ir al banco y esperar dos cosas: (1) tu dinero o (2) un comprobante de error. Si el cajero te dice "cuenta bloqueada" (error), no te da dinero — te da la razón. Tú decides qué hacer con esa información.

### `fmt.Errorf` para errores descriptivos

```go
func verificarEdad(edad int) error {
    if edad < 0 {
        return fmt.Errorf("edad no puede ser negativa: %d", edad)
    }
    if edad > 150 {
        return fmt.Errorf("edad sospechosa: %d años no es realista", edad)
    }
    return nil  // Todo bien, sin error
}
```

### Encadenamiento de errores

```go
func procesarArchivo(nombre string) error {
    archivo, err := os.Open(nombre)
    if err != nil {
        return fmt.Errorf("abriendo archivo %s: %w", nombre, err)
    }
    defer archivo.Close()
    
    // ... procesar
    return nil
}
```

> **Concepto clave `%w`:** El verb `%w` permite "envolver" un error original dentro de un nuevo error con más contexto. Es como las muñecas rusas: el error exterior da contexto ("abriendo archivo config.json"), y el error interior es la causa raíz ("el archivo no existe").

---

## 🚫 `defer`: Limpieza garantizada

La palabra `defer` **pospone** la ejecución de una función hasta que la función que la contiene termine. Se usa para garantizar limpieza de recursos.

```go
func leerArchivo(nombre string) {
    archivo, err := os.Open(nombre)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    defer archivo.Close()  // Se ejecuta AL FINAL, sin importar qué pase
    
    // ... usar el archivo aquí
    fmt.Println("Archivo leído correctamente")
    // Cuando esta función termina, Go ejecuta automáticamente archivo.Close()
}
```

**Analogía:** `defer` es como un **recordatorio automático**. Es como poner un post-it en la puerta de salida que dice "NO OLVIDES CERRAR LA PUERTA". Sin importar cómo salgas de la habitación (normal, emergencia, distracción), ese recordatorio se ejecuta cuando te vas.

**Orden LIFO (Last In, First Out):**

```go
func ejemplo() {
    defer fmt.Println("Primero en salir (último defer)")
    defer fmt.Println("Segundo en salir (penúltimo defer)")
    defer fmt.Println("Tercero en salir (primer defer)")
}
// Salida: Tercero → Segundo → Primero
```

Los `defer` se ejecutan en **orden inverso** (como una pila de platos: el último que pusiste es el primero que sacas).

---

## 🎯 Ejercicio práctico: Calculadora de expresiones con manejo de errores

Vamos a construir una calculadora que parsea expresiones matemáticas simples desde texto, las evalúa respetando precedencia de operadores, y maneja errores elegantemente: división por cero, caracteres inválidos, y expresiones mal formadas.

**Crea el archivo `main.go` dentro de la carpeta `04-funciones/`:**

```go
package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ====================================================
// 🧮 CALCULADORA DE EXPRESIONES MATEMÁTICAS
// Manejo de errores al estilo Go
// ====================================================

// --- TIPOS PERSONALIZADOS ---

// Token representa un elemento de la expresión (número u operador)
type Token struct {
	Valor    string
	EsNumero bool
}

// --- FUNCIONES DE TOKENIZACIÓN ---

// tokenizar convierte una cadena de texto en una lista de tokens
func tokenizar(expresion string) ([]Token, error) {
	// Eliminar espacios en blanco
	expresion = strings.ReplaceAll(expresion, " ", "")

	if len(expresion) == 0 {
		return nil, errors.New("la expresión está vacía")
	}

	var tokens []Token
	var numeroActual strings.Builder

	for _, char := range expresion {
		if unicode.IsDigit(char) || char == '.' {
			// Acumular dígitos del número
			numeroActual.WriteRune(char)
		} else if esOperador(char) {
			// Si teníamos un número acumulado, lo guardamos
			if numeroActual.Len() > 0 {
				tokens = append(tokens, Token{
					Valor:    numeroActual.String(),
					EsNumero: true,
				})
				numeroActual.Reset()
			}
			// Guardar el operador
			tokens = append(tokens, Token{
				Valor:    string(char),
				EsNumero: false,
			})
		} else {
			return nil, fmt.Errorf("carácter inválido: '%c'", char)
		}
	}

	// Guardar el último número si existe
	if numeroActual.Len() > 0 {
		tokens = append(tokens, Token{
			Valor:    numeroActual.String(),
			EsNumero: true,
		})
	}

	return tokens, nil
}

// esOperador verifica si un carácter es un operador matemático
func esOperador(char rune) bool {
	return char == '+' || char == '-' || char == '*' || char == '/'
}

// --- FUNCIONES DE EVALUACIÓN ---

// evaluarTokens procesa los tokens y calcula el resultado
// Primero procesa multiplicación y división, luego suma y resta
func evaluarTokens(tokens []Token) (float64, error) {
	if len(tokens) == 0 {
		return 0, errors.New("no hay tokens para evaluar")
	}

	// Validar que la expresión no empiece ni termine con operador
	if !tokens[0].EsNumero {
		return 0, fmt.Errorf("la expresión no puede empezar con '%s'", tokens[0].Valor)
	}
	if !tokens[len(tokens)-1].EsNumero {
		return 0, fmt.Errorf("la expresión no puede terminar con '%s'", tokens[len(tokens)-1].Valor)
	}

	// Paso 1: Convertir tokens numéricos a valores float64
	var numeros []float64
	var operadores []string

	for _, token := range tokens {
		if token.EsNumero {
			valor, err := strconv.ParseFloat(token.Valor, 64)
			if err != nil {
				return 0, fmt.Errorf("'%s' no es un número válido: %w", token.Valor, err)
			}
			numeros = append(numeros, valor)
		} else {
			operadores = append(operadores, token.Valor)
		}
	}

	// Validar que hay un operador entre cada par de números
	if len(numeros) != len(operadores)+1 {
		return 0, fmt.Errorf("expresión mal formada: %d números pero %d operadores",
			len(numeros), len(operadores))
	}

	// Paso 2: Procesar multiplicación (*) y división (/) primero (mayor precedencia)
	for i := 0; i < len(operadores); {
		if operadores[i] == "*" || operadores[i] == "/" {
			resultado, err := aplicarOperacion(numeros[i], numeros[i+1], operadores[i])
			if err != nil {
				return 0, err
			}
			// Reemplazar los dos números con el resultado
			numeros[i] = resultado
			numeros = append(numeros[:i+1], numeros[i+2:]...)
			operadores = append(operadores[:i], operadores[i+1:]...)
		} else {
			i++
		}
	}

	// Paso 3: Procesar suma (+) y resta (-) de izquierda a derecha
	for len(operadores) > 0 {
		resultado, err := aplicarOperacion(numeros[0], numeros[1], operadores[0])
		if err != nil {
			return 0, err
		}
		numeros[0] = resultado
		numeros = append(numeros[:1], numeros[2:]...)
		operadores = operadores[1:]
	}

	return numeros[0], nil
}

// aplicarOperacion ejecuta una operación entre dos números
// Devuelve (resultado, error) — el patrón clásico de Go
func aplicarOperacion(a, b float64, operador string) (float64, error) {
	switch operador {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, errors.New("❌ error: división por cero")
		}
		return a / b, nil
	default:
		return 0, fmt.Errorf("operador desconocido: '%s'", operador)
	}
}

// --- FUNCIONES DE ALTO NIVEL (COMPOSICIÓN) ---

// calcular es la función principal que orquesta todo el proceso
func calcular(expresion string) (float64, error) {
	// Paso 1: Tokenizar
	tokens, err := tokenizar(expresion)
	if err != nil {
		return 0, fmt.Errorf("tokenizando: %w", err)
	}

	// Paso 2: Evaluar
	resultado, err := evaluarTokens(tokens)
	if err != nil {
		return 0, fmt.Errorf("evaluando: %w", err)
	}

	return resultado, nil
}

// --- FUNCIONES COMO VALORES ---

// crearFormateador devuelve una función que formatea resultados
// con un número específico de decimales (closure)
func crearFormateador(decimales int) func(float64) string {
	plantilla := fmt.Sprintf("%%.%df", decimales)
	return func(valor float64) string {
		return fmt.Sprintf(plantilla, valor)
	}
}

// --- FUNCIÓN PRINCIPAL ---

func main() {

	// ====================================================
	// 🧮 CALCULADORA DE EXPRESIONES MATEMÁTICAS
	// ====================================================

	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║   🧮 CALCULADORA DE EXPRESIONES MATEMÁTICAS         ║")
	fmt.Println("║   Manejo de errores al estilo Go                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// ====================================================
	// 1️⃣ DEMOSTRACIÓN BÁSICA: Funciones simples
	// ====================================================

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   1️⃣  Funciones básicas y retorno múltiple")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Expresiones válidas
	expresiones := []string{
		"3 + 5",
		"10 - 3 * 2",
		"100 / 4 + 5",
		"2 * 3 + 4 * 5",
		"10 + 20 * 3 - 15 / 5",
	}

	for _, expr := range expresiones {
		resultado, err := calcular(expr)
		if err != nil {
			fmt.Printf("   ❌ \"%s\" → Error: %s\n", expr, err)
		} else {
			fmt.Printf("   ✅ \"%s\" = %.2f\n", expr, resultado)
		}
	}

	// ====================================================
	// 2️⃣ MANEJO DE ERRORES: El patrón (resultado, error)
	// ====================================================

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   2️⃣  Manejo de errores con el patrón (resultado, error)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Expresiones con errores (cada una falla por una razón diferente)
	expresionesErroneas := []string{
		"",           // Expresión vacía
		"5 +",        // Termina con operador
		"* 3",        // Empieza con operador
		"5 % 2",      // Carácter inválido
		"10 / 0",     // División por cero
		"abc",        // No es una expresión válida
	}

	for _, expr := range expresionesErroneas {
		mostrar := expr
		if mostrar == "" {
			mostrar = "(vacía)"
		}
		resultado, err := calcular(expr)
		if err != nil {
			fmt.Printf("   ❌ \"%s\" → %s\n", mostrar, err)
		} else {
			fmt.Printf("   ✅ \"%s\" = %.2f\n", mostrar, resultado)
		}
	}

	// ====================================================
	// 3️⃣ FUNCIONES COMO VALORES: Closure de formateo
	// ====================================================

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   3️⃣  Funciones como valores (closures)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Creamos diferentes formateadores con closures
	formatearEntero := crearFormateador(0)
	formatear2Dec := crearFormateador(2)
	formatear4Dec := crearFormateador(4)

	resultado, _ := calcular("22 / 7")
	fmt.Printf("   Resultado de \"22 / 7\":\n")
	fmt.Printf("   📏 Entero:     %s\n", formatearEntero(resultado))
	fmt.Printf("   📏 2 decimales: %s\n", formatear2Dec(resultado))
	fmt.Printf("   📏 4 decimales: %s\n", formatear4Dec(resultado))

	// ====================================================
	// 4️⃣ FUNCIONES VARIÁDICAS
	// ====================================================

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   4️⃣  Funciones variádicas")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Función variádica: promedio de cualquier cantidad de números
	promedio := promedioNumeros(85, 90, 78, 92, 88)
	fmt.Printf("   📊 Promedio de notas [85, 90, 78, 92, 88]: %.2f\n", promedio)

	notas := []float64{70, 75, 80, 85, 90}
	promedio2 := promedioNumeros(notas...)
	fmt.Printf("   📊 Promedio de notas [70, 75, 80, 85, 90]: %.2f\n", promedio2)

	// Función variádica con parámetro fijo + variádico
	mejorNota := buscarMejor("Matemáticas", 85, 90, 78, 92, 88)
	fmt.Printf("   🏆 %s\n", mejorNota)

	// ====================================================
	// 5️⃣ DEFER: Limpieza garantizada
	// ====================================================

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   5️⃣  Demostración de defer (LIFO)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	demostrarDefer()

	// ====================================================
	// 6️⃣ FUNCIÓN QUE DEVUELVE FUNCIÓN (Factory)
	// ====================================================

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   6️⃣  Funciones que crean funciones (Factory pattern)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Creamos funciones especializadas con closures
	sumar10 := crearOperadorConstante("+", 10)
	multiplicar3 := crearOperadorConstante("*", 3)
	dividir2 := crearOperadorConstante("/", 2)

	fmt.Printf("   ➕ sumar10(25) = %.0f\n", sumar10(25))
	fmt.Printf("   ✖️  multiplicar3(7) = %.0f\n", multiplicar3(7))
	fmt.Printf("   ➗ dividir2(100) = %.0f\n", dividir2(100))

	// ====================================================
	// 7️⃣ RESUMEN: EL PATRÓN (resultado, error) EN ACCIÓN
	// ====================================================

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   7️⃣  Uso práctico completo")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Simulamos procesar un lote de expresiones
	lote := []string{
		"50 + 50",
		"200 - 55",
		"12 * 12",
		"144 / 12",
		"10 + 20 * 3",
		"100 / 0",
	}

	resultados, errores := procesarLote(lote)

	exitosos := 0
	fallidos := 0
	for i, expr := range lote {
		if errores[i] != nil {
			fmt.Printf("   ❌ \"%s\" → %s\n", expr, errores[i])
			fallidos++
		} else {
			fmt.Printf("   ✅ \"%s\" = %.2f\n", expr, resultados[i])
			exitosos++
		}
	}

	fmt.Println()
	fmt.Printf("   📊 Resumen: %d exitosas, %d con error (%d total)\n",
		exitosos, fallidos, len(lote))

	// ====================================================
	// 8️⃣ CONTADOR CON CLOSURE
	// ====================================================

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   8️⃣  Contador con closure (estado privado)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	contadorA := crearContador("Operaciones exitosas")
	contadorB := crearContador("Operaciones fallidas")

	// Simular conteo
	contadorA()
	contadorA()
	contadorA()
	contadorA()
	contadorA()
	contadorB()

	fmt.Printf("   %s\n", contadorA())
	fmt.Printf("   %s\n", contadorB())

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("   ✅ Todas las funciones ejecutadas correctamente")
	fmt.Println("═══════════════════════════════════════════════════════")
}

// --- FUNCIONES AUXILIARES ---

// promedioNumeros calcula el promedio de N números (función variádica)
func promedioNumeros(numeros ...float64) float64 {
	if len(numeros) == 0 {
		return 0
	}
	total := 0.0
	for _, n := range numeros {
		total += n
	}
	return total / float64(len(numeros))
}

// buscarMejor recibe una materia y N notas, devuelve la nota más alta
// Demuestra parámetro fijo + variádico
func buscarMejor(materia string, notas ...float64) string {
	if len(notas) == 0 {
		return fmt.Sprintf("No hay notas para %s", materia)
	}
	mejor := notas[0]
	for _, n := range notas[1:] {
		if n > mejor {
			mejor = n
		}
	}
	return fmt.Sprintf("Mejor nota en %s: %.1f", materia, mejor)
}

// demostrarDefer muestra el orden LIFO de los defers
func demostrarDefer() {
	fmt.Println("   Inicio de la función...")
	defer fmt.Println("   🔓 1er defer (se ejecuta ÚLTIMO)")
	defer fmt.Println("   🔓 2do defer (se ejecuta PENÚLTIMO)")
	defer fmt.Println("   🔓 3er defer (se ejecuta PRIMERO)")
	fmt.Println("   Fin de la función...")
	fmt.Println("   ⬇️  Ahora Go ejecuta los defers en orden inverso:")
}

// crearOperadorConstante es una factory que devuelve funciones
// que aplican una operación con un valor constante
func crearOperadorConstante(operador string, constante float64) func(float64) float64 {
	switch operador {
	case "+":
		return func(x float64) float64 {
			resultado, _ := aplicarOperacion(x, constante, "+")
			return resultado
		}
	case "-":
		return func(x float64) float64 {
			resultado, _ := aplicarOperacion(x, constante, "-")
			return resultado
		}
	case "*":
		return func(x float64) float64 {
			resultado, _ := aplicarOperacion(x, constante, "*")
			return resultado
		}
	case "/":
		return func(x float64) float64 {
			resultado, _ := aplicarOperacion(x, constante, "/")
			return resultado
		}
	default:
		return func(x float64) float64 {
			return x
		}
	}
}

// procesarLote procesa múltiples expresiones y devuelve resultados y errores
// Demuestra retorno múltiple con slices
func procesarLote(expresiones []string) ([]float64, []error) {
	resultados := make([]float64, len(expresiones))
	errores := make([]error, len(expresiones))

	for i, expr := range expresiones {
		resultado, err := calcular(expr)
		resultados[i] = resultado
		errores[i] = err
	}

	return resultados, errores
}

// crearContador crea un closure que cuenta y tiene un nombre descriptivo
func crearContador(nombre string) func() string {
	cuenta := 0
	return func() string {
		cuenta++
		return fmt.Sprintf("%s: %d", nombre, cuenta)
	}
}
```

### 🧠 Explicación línea por línea del código clave

**Función con retorno múltiple `(float64, error)`:**

```go
func aplicarOperacion(a, b float64, operador string) (float64, error) {
```

Esta función devuelve DOS valores: el resultado numérico y un posible error. Si todo sale bien, el error es `nil`. Si algo falla (como dividir por cero), el resultado es `0` y el error contiene la explicación.

**Verificación de errores con `if err != nil`:**

```go
resultado, err := calcular(expr)
if err != nil {
    fmt.Printf("   ❌ \"%s\" → Error: %s\n", expr, err)
} else {
    fmt.Printf("   ✅ \"%s\" = %.2f\n", expr, resultado)
}
```

Este es el **patrón más importante de Go**. Cada vez que una función devuelve un error, lo verificamos inmediatamente. No hay `try/catch` — el error es un valor más que debemos manejar.

**Función variádica con `...`:**

```go
func promedioNumeros(numeros ...float64) float64 {
```

Los tres puntos `...` le dicen a Go: "esta función acepta cualquier cantidad de argumentos de tipo `float64`". Internamente, `numeros` es un slice.

**Closure — función que recuerda su entorno:**

```go
func crearContador(nombre string) func() string {
    cuenta := 0
    return func() string {
        cuenta++
        return fmt.Sprintf("%s: %d", nombre, cuenta)
    }
}
```

La función interna (anónima) captura dos variables: `nombre` (inmutable) y `cuenta` (mutable). Cada vez que llamas al contador, `cuenta` incrementa. La variable `cuenta` vive dentro del closure — nadie más puede modificarla.

**Error wrapping con `%w`:**

```go
return 0, fmt.Errorf("tokenizando: %w", err)
```

El verb `%w` "envuelve" el error original dentro de uno nuevo con más contexto. Esto permite rastrear el camino del error: "evaluando → tokenizando → carácter inválido: '%'". Es como una cadena de contexto.

**`defer` — limpieza garantizada:**

```go
defer fmt.Println("   🔓 1er defer (se ejecuta ÚLTIMO)")
```

`defer` pospone la ejecución hasta que la función termina. Múltiples defers se ejecutan en **orden inverso** (LIFO). Es el mecanismo de Go para garantizar limpieza de recursos (cerrar archivos, conexiones, etc.).

**Factory de funciones con closures:**

```go
func crearOperadorConstante(operador string, constante float64) func(float64) float64 {
```

Esta función NO retorna un número — retorna OTRA FUNCIÓN. Cada función retornada "recuerda" el operador y la constante con los que fue creada. Es como una fábrica que produce máquinas personalizadas.

---

### ▶️ Ejecución

```bash
cd 04-funciones
go run main.go
```

**Salida esperada:**

```
╔══════════════════════════════════════════════════════╗
║   🧮 CALCULADORA DE EXPRESIONES MATEMÁTICAS         ║
║   Manejo de errores al estilo Go                    ║
╚══════════════════════════════════════════════════════╝

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   1️⃣  Funciones básicas y retorno múltiple
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   ✅ "3 + 5" = 8.00
   ✅ "10 - 3 * 2" = 4.00
   ✅ "100 / 4 + 5" = 30.00
   ✅ "2 * 3 + 4 * 5" = 26.00
   ✅ "10 + 20 * 3 - 15 / 5" = 67.00

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   2️⃣  Manejo de errores con el patrón (resultado, error)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   ❌ "(vacía)" → tokenizando: la expresión está vacía
   ❌ "5 +" → evaluando: la expresión no puede terminar con '+'
   ❌ "* 3" → evaluando: la expresión no puede empezar con '*'
   ❌ "5 % 2" → tokenizando: carácter inválido: '%'
   ❌ "10 / 0" → evaluando: ❌ error: división por cero
   ❌ "abc" → tokenizando: carácter inválido: 'a'

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   3️⃣  Funciones como valores (closures)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   Resultado de "22 / 7":
   📏 Entero:     3
   📏 2 decimales: 3.14
   📏 4 decimales: 3.1429

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   4️⃣  Funciones variádicas
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   📊 Promedio de notas [85, 90, 78, 92, 88]: 86.60
   📊 Promedio de notas [70, 75, 80, 85, 90]: 80.00
   🏆 Mejor nota en Matemáticas: 92.0

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   5️⃣  Demostración de defer (LIFO)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   Inicio de la función...
   Fin de la función...
   ⬇️  Ahora Go ejecuta los defers en orden inverso:
   🔓 3er defer (se ejecuta PRIMERO)
   🔓 2do defer (se ejecuta PENÚLTIMO)
   🔓 1er defer (se ejecuta ÚLTIMO)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   6️⃣  Funciones que crean funciones (Factory pattern)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   ➕ sumar10(25) = 35
   ✖️  multiplicar3(7) = 21
   ➗ dividir2(100) = 50

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   7️⃣  Uso práctico completo
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   ✅ "50 + 50" = 100.00
   ✅ "200 - 55" = 145.00
   ✅ "12 * 12" = 144.00
   ✅ "144 / 12" = 12.00
   ✅ "10 + 20 * 3" = 70.00
   ❌ "100 / 0" → evaluando: ❌ error: división por cero

   📊 Resumen: 5 exitosas, 1 con error (6 total)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
   8️⃣  Contador con closure (estado privado)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   Operaciones exitosas: 6
   Operaciones fallidas: 2

═══════════════════════════════════════════════════════
   ✅ Todas las funciones ejecutadas correctamente
═══════════════════════════════════════════════════════
```

---

## 🏋️ Ejercicio Feynman

> **Instrucción:** Toma una hoja en blanco o abre un archivo de texto vacío. Sin consultar esta lección, intenta responder cada pregunta **con tus propias palabras**, como si se lo explicaras a alguien que nunca ha programado.

### 📝 Preguntas para explicar desde cero:

1. **¿Qué es una función?** Explica con una analogía que NO sea de cocina ni de máquinas. Crea la tuya propia.

2. **¿Por qué Go devuelve errores como valores (`resultado, error`) en vez de usar `try/catch` como Java o Python?** ¿Qué ventaja tiene? ¿Hay alguna desventaja?

3. **Un amigo te dice:** *"¿Para qué sirve que una función devuelva dos valores? Con uno basta."* Respóndele con un ejemplo concreto de la vida real donde necesitas dos piezas de información de una operación.

4. **Explica qué es un closure sin usar la palabra "variable" ni "capturar".** Usa una analogía de la vida real.

5. **¿Qué diferencia hay entre pasar una función como parámetro y simplemente llamarla dentro de otra función?** ¿Cuándo es útil cada enfoque?

6. **¿Qué es `defer` y por qué el orden LIFO es importante?** Imagina que estás abriendo 3 cajas anidadas — ¿por qué tiene sentido cerrarlas en orden inverso?

7. **Predice la salida de este código SIN ejecutarlo:**
   ```go
   func crearSumador(inicio int) func(int) int {
       total := inicio
       return func(n int) int {
           total += n
           return total
       }
   }

   s := crearSumador(10)
   fmt.Println(s(5))
   fmt.Println(s(3))
   fmt.Println(s(1))
   ```

### ✅ Criterio de autoevaluación:

| Criterio                                           | ¿Lo lograste? |
|----------------------------------------------------|:-------------:|
| Explicaste funciones sin mirar la lección          | ⬜ Sí / ⬜ No |
| Creaste tu propia analogía original                | ⬜ Sí / ⬜ No |
| Entendes el patrón (resultado, error) de Go        | ⬜ Sí / ⬜ No |
| Puedes explicar closures con una analogía          | ⬜ Sí / ⬜ No |
| Entiendes por qué defer usa orden LIFO             | ⬜ Sí / ⬜ No |
| Predijiste correctamente la salida del código      | ⬜ Sí / ⬜ No |
| Diferencias funciones como valores de llamadas     | ⬜ Sí / ⬜ No |

---

## 🗺️ Próxima lección

En la **Lección 05** exploraremos **Estructuras (structs), Métodos y la filosofía "Composición sobre Herencia"**: cómo Go reemplaza las clases de Java/Python con algo más elegante, flexible y testeable.

> *"Una función bien diseñada es como una herramienta perfecta: hace exactamente una cosa, la hace bien, y cualquiera puede usarla sin instrucciones."* — Principio Feynman