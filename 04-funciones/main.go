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
		"",       // Expresión vacía
		"5 +",    // Termina con operador
		"* 3",    // Empieza con operador
		"5 % 2",  // Carácter inválido
		"10 / 0", // División por cero
		"abc",    // No es una expresión válida
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