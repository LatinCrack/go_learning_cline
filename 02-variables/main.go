package main

import (
	"fmt"
	"os"
	"strconv"
)

// ╔════════════════════════════════════════════════════════════════════╗
// ║  Conversor de Unidades Universal                                  ║
// ║  Lección 02 — Variables, Tipos y el Sistema de Tipos de Go       ║
// ║                                                                   ║
// ║  Este programa demuestra:                                         ║
// ║  - Declaración de variables (var, :=, const)                      ║
// ║  - Tipos primitivos (int, float64, string, bool)                  ║
// ║  - Conversiones explícitas (strconv, fmt.Sprintf)                 ║
// ║  - Zero values                                                    ║
// ║  - Formateo avanzado con fmt.Sprintf                              ║
// ║                                                                   ║
// ║  Uso:                                                             ║
// ║    go run main.go temp 100        (Celsius → °F, K)               ║
// ║    go run main.go dist 42.195     (km → millas, metros)           ║
// ║    go run main.go data 1073741824 (bytes → KB, MB, GB)            ║
// ╚════════════════════════════════════════════════════════════════════╝

func main() {
	// ────────────────────────────────────────────────────────────────
	// Encabezado decorativo con Unicode box-drawing characters
	// Mismo estilo visual que la Lección 01
	// ────────────────────────────────────────────────────────────────
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║   🔄 Conversor de Unidades Universal — Lab Go       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")

	// ────────────────────────────────────────────────────────────────
	// Verificar que el usuario proporcionó suficientes argumentos
	// os.Args es un slice de strings:
	//   os.Args[0] = nombre del programa
	//   os.Args[1] = tipo de conversión (temp, dist, data)
	//   os.Args[2] = valor numérico a convertir
	// ────────────────────────────────────────────────────────────────
	if len(os.Args) < 3 {
		fmt.Println("\n  ⚠️  Uso: go run main.go <tipo> <valor>")
		fmt.Println("  Tipos disponibles: temp, dist, data")
		fmt.Println("  Ejemplo: go run main.go temp 100")
		os.Exit(1)
	}

	// ────────────────────────────────────────────────────────────────
	// Extraer argumentos como strings
	// ────────────────────────────────────────────────────────────────
	tipo := os.Args[1]
	valorStr := os.Args[2]

	// ────────────────────────────────────────────────────────────────
	// Convertir el string del valor a float64
	// strconv.ParseFloat(valorStr, bitSize) intenta parsear el string
	// como un número de punto flotante de precisión bitSize (64 = float64)
	//
	// Devuelve DOS valores:
	//   1. valor → el número convertido (float64)
	//   2. err   → nil si todo salió bien, o un error si falló
	//
	// Este patrón (resultado, error) es fundamental en Go.
	// Lo veremos una y otra vez en lecciones futuras.
	// ────────────────────────────────────────────────────────────────
	valor, err := strconv.ParseFloat(valorStr, 64)
	if err != nil {
		fmt.Printf("\n  ❌ Error: '%s' no es un número válido\n", valorStr)
		fmt.Println("  Ejemplo válido: go run main.go temp 100")
		os.Exit(1)
	}

	// ────────────────────────────────────────────────────────────────
	// Línea en blanco para separar visualmente el encabezado
	// de los resultados
	// ────────────────────────────────────────────────────────────────
	fmt.Println()

	// ────────────────────────────────────────────────────────────────
	// switch evalúa el tipo de conversión solicitado
	// En Go, NO necesita break — cada case se ejecuta y termina
	// automáticamente. Esto elimina el bug clásico del "fall-through"
	// accidental que existe en C, C++ y Java.
	// ────────────────────────────────────────────────────────────────
	switch tipo {
	case "temp":
		convertirTemperatura(valor)
	case "dist":
		convertirDistancia(valor)
	case "data":
		convertirAlmacenamiento(valor)
	default:
		fmt.Printf("  ❌ Tipo desconocido: '%s'\n", tipo)
		fmt.Println("  Tipos disponibles: temp, dist, data")
		os.Exit(1)
	}
}

// ────────────────────────────────────────────────────────────────────────
// convertirTemperatura convierte un valor en grados Celsius a:
//   - Fahrenheit: (°C × 9/5) + 32
//   - Kelvin:      °C + 273.15
//
// IMPORTANTE: Observa que usamos literales 9.0, 5.0, 32.0 (float64)
// Si escribiéramos celsius*9/5, Go calcularía 9/5 = 1 (división entera)
// en vez de 1.8. ¡Este es un bug MUY común para principiantes!
// ────────────────────────────────────────────────────────────────────────
func convertirTemperatura(celsius float64) {
	// ── Variables para almacenar las conversiones ──
	fahrenheit := celsius*9.0/5.0 + 32.0
	kelvin := celsius + 273.15

	// ── Resultado principal con formato alineado ──
	// %10.2f = float, ancho 10 caracteres, 2 decimales
	fmt.Println("  🌡️  Conversión de Temperatura")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Printf("  📥 Celsius     : %10.2f °C\n", celsius)
	fmt.Printf("  📤 Fahrenheit  : %10.2f °F\n", fahrenheit)
	fmt.Printf("  📤 Kelvin      : %10.2f K\n", kelvin)
	fmt.Println("  ─────────────────────────────────────────")

	// ── Tabla de conversiones cruzadas ──
	// Muestra TODAS las relaciones entre las tres escalas
	fmt.Printf("\n  🔄 Tabla de conversiones cruzadas:\n")
	fmt.Printf("     %.2f°F = %.2f°C = %.2fK\n", fahrenheit, celsius, kelvin)
	fmt.Printf("     %.2fK = %.2f°C = %.2f°F\n", kelvin, celsius, fahrenheit)

	// ── Demostración de tipos: const vs var vs := ──
	// Aquí mostramos los diferentes tipos de variables en acción
	var (
		celsiusStr    string  = fmt.Sprintf("%.2f", celsius)     // var explícito
		fahrenheitStr string  = fmt.Sprintf("%.2f", fahrenheit)  // var explícito
		kelvinStr     string  = fmt.Sprintf("%.2f", kelvin)      // var explícito
	)
	_ = celsiusStr    // evitamos "declared and not used"
	_ = fahrenheitStr
	_ = kelvinStr

	fmt.Println("\n  📊 Detalle de tipos utilizados:")
	fmt.Printf("     Tipo de celsius    : %T (valor: %v)\n", celsius, celsius)
	fmt.Printf("     Tipo de fahrenheit : %T (valor: %v)\n", fahrenheit, fahrenheit)
	fmt.Printf("     Tipo de kelvin     : %T (valor: %v)\n", kelvin, kelvin)
}

// ────────────────────────────────────────────────────────────────────────
// convertirDistancia convierte un valor en kilómetros a:
//   - Millas: km × 0.621371
//   - Metros: km × 1000
//
// Usa float64 para todas las operaciones, lo que garantiza
// precisión en los cálculos de punto flotante.
// ────────────────────────────────────────────────────────────────────────
func convertirDistancia(km float64) {
	// ── Constantes de conversión (inmutables) ──
	// const se evalúa en tiempo de compilación
	const kmAMillas float64 = 0.621371
	const kmAMetros float64 = 1000.0

	// ── Variables de resultado con := (inferencia de tipo) ──
	millas := km * kmAMillas
	metros := km * kmAMetros

	// ── Resultado principal con formato alineado ──
	// %12.4f = float, ancho 12 caracteres, 4 decimales
	fmt.Println("  📏 Conversión de Distancia")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Printf("  📥 Kilómetros  : %12.4f km\n", km)
	fmt.Printf("  📤 Millas      : %12.4f mi\n", millas)
	fmt.Printf("  📤 Metros      : %12.4f m\n", metros)
	fmt.Println("  ─────────────────────────────────────────")

	// ── Tabla de conversiones cruzadas ──
	fmt.Printf("\n  🔄 Tabla de conversiones cruzadas:\n")
	fmt.Printf("     %.4fkm = %.4fmi = %.4fm\n", km, millas, metros)
	fmt.Printf("     %.4fmi = %.4fkm = %.4fm\n", millas, km, metros)

	// ── Tipos en acción ──
	fmt.Println("\n  📊 Detalle de tipos utilizados:")
	fmt.Printf("     Tipo de km     : %T (valor: %v)\n", km, km)
	fmt.Printf("     Tipo de millas : %T (valor: %v)\n", millas, millas)
	fmt.Printf("     Tipo de metros : %T (valor: %v)\n", metros, metros)
	fmt.Printf("     Tipo de constante kmAMillas: %T (valor: %v)\n", kmAMillas, kmAMillas)
}

// ────────────────────────────────────────────────────────────────────────
// convertirAlmacenamiento convierte un valor en bytes a:
//   - Kilobytes (KB): bytes / 1024
//   - Megabytes (MB): bytes / 1024²
//   - Gigabytes (GB): bytes / 1024³
//
// NOTA: En informática, 1 KB = 1024 bytes (base 2), NO 1000.
// El prefijo correcto sería "kibibyte" (KiB), pero la industria
// usa "kilobyte" informalmente.
//
// Si escribiéramos bytes / 1000.0, todo el conversor estaría mal.
// ────────────────────────────────────────────────────────────────────────
func convertirAlmacenamiento(bytes float64) {
	// ── Constantes para las conversiones ──
	const (
		bytesPorKB = 1024.0
		bytesPorMB = 1024.0 * 1024.0          // 1,048,576
		bytesPorGB = 1024.0 * 1024.0 * 1024.0 // 1,073,741,824
	)

	// ── Variables de resultado con := (inferencia de tipo) ──
	kb := bytes / bytesPorKB
	mb := bytes / bytesPorMB
	gb := bytes / bytesPorGB

	// ── Resultado principal con formato alineado ──
	// %15.0f = float, ancho 15 caracteres, 0 decimales (para bytes)
	// %15.4f = float, ancho 15 caracteres, 4 decimales (para KB, MB, GB)
	fmt.Println("  💾 Conversión de Almacenamiento")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Printf("  📥 Bytes       : %15.0f B\n", bytes)
	fmt.Printf("  📤 Kilobytes   : %15.4f KB\n", kb)
	fmt.Printf("  📤 Megabytes   : %15.4f MB\n", mb)
	fmt.Printf("  📤 Gigabytes   : %15.4f GB\n", gb)
	fmt.Println("  ─────────────────────────────────────────")

	// ── Tabla de conversiones cruzadas ──
	// Para GB usamos 6 decimales porque son valores muy pequeños
	fmt.Printf("\n  🔄 Tabla de conversiones cruzadas:\n")
	fmt.Printf("     %.0fB = %.4fKB = %.4fMB = %.6fGB\n", bytes, kb, mb, gb)

	// ── Tipos en acción ──
	fmt.Println("\n  📊 Detalle de tipos utilizados:")
	fmt.Printf("     Tipo de bytes : %T (valor: %v)\n", bytes, bytes)
	fmt.Printf("     Tipo de kb    : %T (valor: %v)\n", kb, kb)
	fmt.Printf("     Tipo de mb    : %T (valor: %v)\n", mb, mb)
	fmt.Printf("     Tipo de gb    : %T (valor: %v)\n", gb, gb)
}