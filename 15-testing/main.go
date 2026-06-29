// =============================================================
// LECCION 15: TESTING EN GO — De la confianza ciega a la certeza absoluta
// =============================================================
// Este programa principal demuestra las funciones que luego
// testearamos exhaustivamente con el framework de testing de Go.
// =============================================================
package main

import (
	"fmt"
	"strings"

	"15-testing/myutils/mathutil"
	"15-testing/myutils/textutil"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  LECCION 15: TESTING EN GO")
	fmt.Println("  De la confianza ciega a la certeza")
	fmt.Println("========================================")
	fmt.Println()

	// ── Demostracion de mathutil ──────────────────────────────
	fmt.Println("🔬 MATHUTIL - Funciones matematicas")
	fmt.Println(strings.Repeat("─", 45))

	nums := []float64{10, 20, 30, 40, 50}
	fmt.Printf("  Datos: %v\n", nums)

	suma := mathutil.Suma(nums)
	fmt.Printf("  Suma:           %.2f\n", suma)

	prom, _ := mathutil.Promedio(nums)
	fmt.Printf("  Promedio:       %.2f\n", prom)

	max, _ := mathutil.Maximo(nums)
	fmt.Printf("  Maximo:         %.2f\n", max)

	min, _ := mathutil.Minimo(nums)
	fmt.Printf("  Minimo:         %.2f\n", min)

	med, _ := mathutil.Mediana(nums)
	fmt.Printf("  Mediana:        %.2f\n", med)

	desv, _ := mathutil.DesviacionEstandar(nums)
	fmt.Printf("  Desv. Estandar: %.2f\n", desv)

	pct, _ := mathutil.Porcentaje(30, 80)
	fmt.Printf("  30 de 80:       %.1f%%\n", pct)

	fact, _ := mathutil.Factorial(6)
	fmt.Printf("  6! =            %d\n", fact)

	fib, _ := mathutil.Fibonacci(10)
	fmt.Printf("  Fibonacci(10):  %v\n", fib)

	fmt.Printf("  ¿Es 17 primo?   %v\n", mathutil.EsPrimo(17))
	fmt.Printf("  ¿Es 100 primo?  %v\n", mathutil.EsPrimo(100))
	fmt.Printf("  Clamp(150,0,100): %d\n", mathutil.Clamp(150, 0, 100))
	fmt.Printf("  Contar pares [1..10]: %d\n", mathutil.ContarPares([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}))

	// ── Manejo de errores ─────────────────────────────────────
	fmt.Println()
	fmt.Println("⚠️  MATHUTIL - Casos de error")
	fmt.Println(strings.Repeat("─", 45))

	_, err := mathutil.Promedio([]float64{})
	if err != nil {
		fmt.Printf("  Promedio([]):    ❌ %v\n", err)
	}

	_, err = mathutil.Factorial(-5)
	if err != nil {
		fmt.Printf("  Factorial(-5):   ❌ %v\n", err)
	}

	_, err = mathutil.Porcentaje(50, 0)
	if err != nil {
		fmt.Printf("  Porcentaje(50,0):❌ %v\n", err)
	}

	// ── Demostracion de textutil ──────────────────────────────
	fmt.Println()
	fmt.Println("📝 TEXTUTIL - Funciones de texto")
	fmt.Println(strings.Repeat("─", 45))

	fmt.Printf("  Invertir(\"hola\"):         %q\n", textutil.Invertir("hola"))
	fmt.Printf("  ¿Es \"ana\" palindromo?     %v\n", textutil.EsPalindromo("ana"))
	fmt.Printf("  ¿Es \"hola\" palindromo?    %v\n", textutil.EsPalindromo("hola"))
	fmt.Printf("  Palabras en \"Go es genial\":%d\n", textutil.ContarPalabras("Go es genial"))
	fmt.Printf("  Titulo(\"hola mundo\"):     %q\n", textutil.TituloCapital("hola mundo"))
	fmt.Printf("  SoloLetras(\"Go 1.21!\"):   %q\n", textutil.SoloLetras("Go 1.21!"))
	fmt.Printf("  Truncar(\"hola mundo\",5):   %q\n", textutil.Truncar("hola mundo", 5))
	fmt.Printf("  Vocales en \"programacion\": %d\n", textutil.ContarVocales("programacion"))
	fmt.Printf("  Reemplazar vocales \"hola*\":%q\n", textutil.ReemplazarVocales("hola", '*'))

	// ── Mensaje final ─────────────────────────────────────────
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  ✅ Ejecuta los tests con:")
	fmt.Println("     go test ./... -v")
	fmt.Println()
	fmt.Println("  ⚡ Ejecuta los benchmarks con:")
	fmt.Println("     go test ./... -bench=.")
	fmt.Println()
	fmt.Println("  📊 Cobertura de codigo:")
	fmt.Println("     go test ./... -cover")
	fmt.Println("========================================")
}