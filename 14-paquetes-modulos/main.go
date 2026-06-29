// ============================================================
//  Leccion 14 - Paquetes y Modulos en Go
// ============================================================
//  Este programa demuestra como crear, organizar e importar
//  paquetes propios dentro de un modulo Go.
//
//  Ejecutar:  go run main.go
// ============================================================
package main

import (
	"fmt"
	"math"

	"14-paquetes-modulos/myutils/mathutil"
	"14-paquetes-modulos/myutils/textutil"
)

// ─────────────────────────────────────────────────────────────
//  Funcion auxiliar para mostrar separadores visuales
// ─────────────────────────────────────────────────────────────
func separador(titulo string) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  %s\n", titulo)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func main() {
	fmt.Println("🚀 LECCION 14: PAQUETES Y MODULOS EN GO")
	fmt.Println("   Organizando codigo como un profesional")

	// ========================================================
	//  BLOQUE 1: mathutil - Funciones estadisticas basicas
	// ========================================================
	separador("📊 mathutil: Estadistica Basica")

	notas := []float64{85.5, 92.0, 78.3, 96.1, 88.7, 73.4, 91.0}

	fmt.Printf("\n  Notas del grupo: %v\n", notas)

	// Usamos Suma (no retorna error, es seguro con slices vacios)
	total := mathutil.Suma(notas)
	fmt.Printf("  Suma total:      %.1f\n", total)

	// Usamos Promedio (retorna error si el slice esta vacio)
	promedio, err := mathutil.Promedio(notas)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  Promedio:        %.2f\n", promedio)
	}

	// Usamos Maximo y Minimo
	maximo, _ := mathutil.Maximo(notas)
	minimo, _ := mathutil.Minimo(notas)
	fmt.Printf("  Nota mas alta:   %.1f\n", maximo)
	fmt.Printf("  Nota mas baja:   %.1f\n", minimo)

	// Usamos DesviacionEstandar
	desv, _ := mathutil.DesviacionEstandar(notas)
	fmt.Printf("  Desv. estandar:  %.2f\n", desv)

	// Usamos Mediana
	mediana, _ := mathutil.Mediana(notas)
	fmt.Printf("  Mediana:         %.1f\n", mediana)

	// ========================================================
	//  BLOQUE 2: mathutil - Funciones matematicas avanzadas
	// ========================================================
	separador("🔢 mathutil: Matematicas Avanzadas")

	// Factorial
	fact, err := mathutil.Factorial(10)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  10! = %d\n", fact)
	}

	// Fibonacci
	fib, _ := mathutil.Fibonacci(12)
	fmt.Printf("  Fibonacci(12) = %v\n", fib)

	// Numeros primos
	fmt.Print("  Primos del 1 al 30: ")
	for i := 1; i <= 30; i++ {
		if mathutil.EsPrimo(i) {
			fmt.Printf("%d ", i)
		}
	}
	fmt.Println()

	// Clamp (util para limitar valores en rangos)
	fmt.Printf("\n  Clamp(105, 0, 100) = %d  (nota maxima 100)\n", mathutil.Clamp(105, 0, 100))
	fmt.Printf("  Clamp(-5, 0, 100)  = %d  (nota minima 0)\n", mathutil.Clamp(-5, 0, 100))
	fmt.Printf("  Clamp(87, 0, 100)  = %d  (nota dentro del rango)\n", mathutil.Clamp(87, 0, 100))

	// Porcentaje
	p, err := mathutil.Porcentaje(85.5, 100)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
	} else {
		fmt.Printf("  85.5 es el %.1f%% de 100\n", p)
	}

	// ========================================================
	//  BLOQUE 3: Manejo de errores de mathutil
	// ========================================================
	separador("⚠️ mathutil: Manejo de Errores")

	// Caso: slice vacio
	_, err = mathutil.Promedio([]float64{})
	if err != nil {
		fmt.Printf("  Promedio([])       → ❌ %v\n", err)
	}

	// Caso: factorial negativo
	_, err = mathutil.Factorial(-5)
	if err != nil {
		fmt.Printf("  Factorial(-5)      → ❌ %v\n", err)
	}

	// Caso: fibonacci negativo
	_, err = mathutil.Fibonacci(-3)
	if err != nil {
		fmt.Printf("  Fibonacci(-3)      → ❌ %v\n", err)
	}

	// Caso: division por cero
	_, err = mathutil.Porcentaje(50, 0)
	if err != nil {
		fmt.Printf("  Porcentaje(50, 0)  → ❌ %v\n", err)
	}

	// ========================================================
	//  BLOQUE 4: textutil - Manipulacion de texto
	// ========================================================
	separador("📝 textutil: Manipulacion de Texto")

	frase := "A man a plan a canal Panama"
	fmt.Printf("\n  Frase original:    \"%s\"\n", frase)
	fmt.Printf("  Es palindromo?     %v\n", textutil.EsPalindromo(frase))
	fmt.Printf("  Cantidad palabras: %d\n", textutil.ContarPalabras(frase))
	fmt.Printf("  Invertida:         \"%s\"\n", textutil.Invertir(frase))

	// Truncar texto largo
	textoLargo := "Este es un texto muy largo que necesitamos truncar para mostrar en una tarjeta"
	fmt.Printf("\n  Texto original (%d chars):\n    \"%s\"\n", len([]rune(textoLargo)), textoLargo)
	fmt.Printf("  Truncado a 30 chars:\n    \"%s\"\n", textutil.Truncar(textoLargo, 30))

	// SoloLetras (limpiar caracteres especiales)
	sucio := "Go 2.0 es genial!!! @#$%"
	limpio := textutil.SoloLetras(sucio)
	fmt.Printf("\n  Original:  \"%s\"\n", sucio)
	fmt.Printf("  Solo letras: \"%s\"\n", limpio)

	// ========================================================
	//  BLOQUE 5: Combinando paquetes - Ejemplo real
	// ========================================================
	separador("🔬 Ejemplo Real: Analisis de Calificaciones")

	// Simulamos calificaciones de 3 materias
	materias := map[string][]float64{
		"Matematica":   {85, 90, 78, 92, 88},
		"Programacion": {95, 88, 92, 97, 91},
		"Fisica":       {70, 65, 80, 75, 72},
	}

	type resumenMateria struct {
		nombre   string
		promedio float64
		maximo   float64
		minimo   float64
		desv     float64
	}

	var mejores resumenMateria
	mejores.promedio = 0

	for nombre, notas := range materias {
		prom, _ := mathutil.Promedio(notas)
		max, _ := mathutil.Maximo(notas)
		min, _ := mathutil.Minimo(notas)
		desv, _ := mathutil.DesviacionEstandar(notas)

		// Formateamos el nombre a titulo usando textutil
		nombreFormateado := textutil.Truncar(nombre, 15)

		fmt.Printf("\n  📚 %s\n", nombreFormateado)
		fmt.Printf("     Promedio: %.1f | Max: %.0f | Min: %.0f | Desv: %.1f\n",
			prom, max, min, desv)

		// Rastreamos la mejor materia
		if prom > mejores.promedio {
			mejores = resumenMateria{nombre, prom, max, min, desv}
		}
	}

	fmt.Printf("\n  🏆 Mejor materia: %s (promedio: %.1f)\n", mejores.nombre, mejores.promedio)

	// Usamos Clamp para normalizar notas a escala 0-100
	fmt.Println("\n  📐 Notas normalizadas (Clamp 0-100):")
	for nombre, notas := range materias {
		fmt.Printf("     %s: ", textutil.Truncar(nombre, 14))
		for _, n := range notas {
			fmt.Printf("[%d→%d] ", int(n), mathutil.Clamp(int(n), 0, 100))
		}
		fmt.Println()
	}

	// ========================================================
	//  BLOQUE 6: Verificacion de paquete importado
	// ========================================================
	separador("✅ Verificacion Final")

	fmt.Println("\n  Paquetes utilizados en esta leccion:")
	fmt.Println("    ✦ 14-paquetes-modulos/myutils/mathutil")
	fmt.Println("      → Suma, Promedio, Maximo, Minimo, Mediana")
	fmt.Println("      → DesviacionEstandar, Factorial, Fibonacci")
	fmt.Println("      → EsPrimo, Clamp, Porcentaje")
	fmt.Println()
	fmt.Println("    ✦ 14-paquetes-modulos/myutils/textutil")
	fmt.Println("      → Invertir, EsPalindromo, ContarPalabras")
	fmt.Println("      → TituloCapital, SoloLetras, Truncar")
	fmt.Println()

	// Demostramos que math funciona desde el import estandar
	fmt.Printf("  Constante math.Pi = %.15f\n", math.Pi)
	fmt.Println("\n  🎯 ¡Paquetes y modulos dominados!")
}