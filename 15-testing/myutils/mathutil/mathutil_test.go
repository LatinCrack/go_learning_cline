package mathutil

import (
	"fmt"
	"math"
	"testing"
)

// ============================================================
// TESTS TABLE-DRIVEN: El patron estandar de la industria en Go
// ============================================================

// TestSuma verifica la funcion Suma con multiples casos de prueba.
func TestSuma(t *testing.T) {
	// Tabla de casos: cada fila es un escenario diferente
	tests := []struct {
		name     string    // Nombre descriptivo del caso
		input    []float64 // Datos de entrada
		expected float64   // Resultado esperado
	}{
		{"numeros positivos", []float64{1, 2, 3, 4, 5}, 15.0},
		{"numeros negativos", []float64{-1, -2, -3}, -6.0},
		{"mezcla positivos y negativos", []float64{10, -5, 3, -2}, 6.0},
		{"un solo elemento", []float64{42}, 42.0},
		{"slice vacio", []float64{}, 0.0},
		{"decimales", []float64{0.1, 0.2, 0.3}, 0.6},
		{"numeros grandes", []float64{1e10, 1e10}, 2e10},
		{"ceros", []float64{0, 0, 0}, 0.0},
	}

	for _, tt := range tests {
		// t.Run crea un SUBTEST: cada caso se ejecuta independientemente
		// Si uno falla, los otros continuan
		t.Run(tt.name, func(t *testing.T) {
			result := Suma(tt.input)
			// Usamos una tolerancia para comparar floats (evitar errores de precision)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Suma(%v) = %f, se esperaba %f", tt.input, result, tt.expected)
			}
		})
	}
}

// TestPromedio verifica la funcion Promedio incluyendo el caso de error.
func TestPromedio(t *testing.T) {
	tests := []struct {
		name      string
		input     []float64
		expected  float64
		expectErr error // Si no es nil, esperamos que retorne este error
	}{
		{"numeros positivos", []float64{10, 20, 30}, 20.0, nil},
		{"un solo elemento", []float64{100}, 100.0, nil},
		{"negativos", []float64{-10, -20, -30}, -20.0, nil},
		{"slice vacio", []float64{}, 0, ErrEmptySlice},
		{"decimales", []float64{1.5, 2.5, 3.5}, 2.5, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Promedio(tt.input)

			// Verificamos si el error es el esperado
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("Promedio(%v) error = %v, se esperaba %v", tt.input, err, tt.expectErr)
				}
				return // Si esperabamos error, no verificamos el resultado
			}

			if err != nil {
				t.Errorf("Promedio(%v) retorno error inesperado: %v", tt.input, err)
				return
			}

			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Promedio(%v) = %f, se esperaba %f", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMaximo verifica la funcion Maximo.
func TestMaximo(t *testing.T) {
	tests := []struct {
		name      string
		input     []float64
		expected  float64
		expectErr error
	}{
		{"numeros positivos", []float64{3, 1, 4, 1, 5, 9}, 9.0, nil},
		{"numeros negativos", []float64{-5, -3, -1}, -1.0, nil},
		{"un solo elemento", []float64{42}, 42.0, nil},
		{"slice vacio", []float64{}, 0, ErrEmptySlice},
		{"todos iguales", []float64{7, 7, 7}, 7.0, nil},
		{"decimales", []float64{1.1, 2.9, 2.8}, 2.9, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Maximo(tt.input)
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("Maximo(%v) error = %v, se esperaba %v", tt.input, err, tt.expectErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Maximo(%v) error inesperado: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Maximo(%v) = %f, se esperaba %f", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMinimo verifica la funcion Minimo.
func TestMinimo(t *testing.T) {
	tests := []struct {
		name      string
		input     []float64
		expected  float64
		expectErr error
	}{
		{"numeros positivos", []float64{3, 1, 4, 1, 5, 9}, 1.0, nil},
		{"numeros negativos", []float64{-5, -3, -1}, -5.0, nil},
		{"slice vacio", []float64{}, 0, ErrEmptySlice},
		{"un solo elemento", []float64{42}, 42.0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Minimo(tt.input)
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("Minimo(%v) error = %v, se esperaba %v", tt.input, err, tt.expectErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Minimo(%v) error inesperado: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Minimo(%v) = %f, se esperaba %f", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMediana verifica la funcion Mediana.
func TestMediana(t *testing.T) {
	tests := []struct {
		name      string
		input     []float64
		expected  float64
		expectErr error
	}{
		{"impar elementos", []float64{3, 1, 2}, 2.0, nil},
		{"par elementos", []float64{1, 2, 3, 4}, 2.5, nil},
		{"un solo elemento", []float64{5}, 5.0, nil},
		{"slice vacio", []float64{}, 0, ErrEmptySlice},
		{"numeros desordenados", []float64{5, 1, 3, 2, 4}, 3.0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Mediana(tt.input)
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("Mediana(%v) error = %v, se esperaba %v", tt.input, err, tt.expectErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Mediana(%v) error inesperado: %v", tt.input, err)
				return
			}
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Mediana(%v) = %f, se esperaba %f", tt.input, result, tt.expected)
			}
		})
	}
}

// TestDesviacionEstandar verifica la funcion DesviacionEstandar.
func TestDesviacionEstandar(t *testing.T) {
	tests := []struct {
		name      string
		input     []float64
		expected  float64
		expectErr error
	}{
		{"todos iguales", []float64{5, 5, 5, 5}, 0.0, nil},
		{"numeros conocidos", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 2.0, nil},
		{"slice vacio", []float64{}, 0, ErrEmptySlice},
		{"un solo elemento", []float64{10}, 0.0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DesviacionEstandar(tt.input)
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("DesviacionEstandar(%v) error = %v, se esperaba %v", tt.input, err, tt.expectErr)
				}
				return
			}
			if err != nil {
				t.Errorf("DesviacionEstandar(%v) error inesperado: %v", tt.input, err)
				return
			}
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("DesviacionEstandar(%v) = %f, se esperaba %f", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFactorial verifica la funcion Factorial.
func TestFactorial(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		expected  int
		expectErr error
	}{
		{"factorial de 0", 0, 1, nil},
		{"factorial de 1", 1, 1, nil},
		{"factorial de 5", 5, 120, nil},
		{"factorial de 10", 10, 3628800, nil},
		{"factorial de 20", 20, 2432902008176640000, nil},
		{"numero negativo", -1, 0, ErrNegativeValue},
		{"numero negativo grande", -100, 0, ErrNegativeValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Factorial(tt.input)
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("Factorial(%d) error = %v, se esperaba %v", tt.input, err, tt.expectErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Factorial(%d) error inesperado: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Factorial(%d) = %d, se esperaba %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFibonacci verifica la funcion Fibonacci.
func TestFibonacci(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		expected  []int
		expectErr error
	}{
		{"fibonacci de 0", 0, []int{}, nil},
		{"fibonacci de 1", 1, []int{0}, nil},
		{"fibonacci de 2", 2, []int{0, 1}, nil},
		{"fibonacci de 8", 8, []int{0, 1, 1, 2, 3, 5, 8, 13}, nil},
		{"fibonacci de 12", 12, []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89}, nil},
		{"numero negativo", -5, nil, ErrNegativeValue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Fibonacci(tt.input)
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("Fibonacci(%d) error = %v, se esperaba %v", tt.input, err, tt.expectErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Fibonacci(%d) error inesperado: %v", tt.input, err)
				return
			}
			if len(result) != len(tt.expected) {
				t.Errorf("Fibonacci(%d) longitud = %d, se esperaba %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Fibonacci(%d)[%d] = %d, se esperaba %d", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

// TestEsPrimo verifica la funcion EsPrimo.
func TestEsPrimo(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected bool
	}{
		{"cero no es primo", 0, false},
		{"uno no es primo", 1, false},
		{"dos es primo", 2, true},
		{"tres es primo", 3, true},
		{"cuatro no es primo", 4, false},
		{"cinco es primo", 5, true},
		{"nueve no es primo", 9, false},
		{"once es primo", 11, true},
		{"quince no es primo", 15, false},
		{"diecisiete es primo", 17, true},
		{"cien no es primo", 100, false},
		{"ciento tres es primo", 103, true},
		{"negativo no es primo", -7, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EsPrimo(tt.input)
			if result != tt.expected {
				t.Errorf("EsPrimo(%d) = %v, se esperaba %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestClamp verifica la funcion Clamp.
func TestClamp(t *testing.T) {
	tests := []struct {
		name     string
		valor    int
		min      int
		max      int
		expected int
	}{
		{"dentro del rango", 50, 0, 100, 50},
		{"debajo del minimo", -5, 0, 100, 0},
		{"encima del maximo", 105, 0, 100, 100},
		{"en el limite inferior", 0, 0, 100, 0},
		{"en el limite superior", 100, 0, 100, 100},
		{"rango negativo", -50, -100, 0, -50},
		{"min igual a max", 5, 10, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clamp(tt.valor, tt.min, tt.max)
			if result != tt.expected {
				t.Errorf("Clamp(%d, %d, %d) = %d, se esperaba %d", tt.valor, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

// TestPorcentaje verifica la funcion Porcentaje.
func TestPorcentaje(t *testing.T) {
	tests := []struct {
		name      string
		valor     float64
		total     float64
		expected  float64
		expectErr error
	}{
		{"cincuenta por ciento", 50, 100, 50.0, nil},
		{"cien por ciento", 100, 100, 100.0, nil},
		{"cero por ciento", 0, 100, 0.0, nil},
		{"division por cero", 50, 0, 0, ErrZeroDivisor},
		{"porcentaje pequeno", 1, 3, 33.333333, nil},
		{"valor mayor que total", 150, 100, 150.0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Porcentaje(tt.valor, tt.total)
			if tt.expectErr != nil {
				if err != tt.expectErr {
					t.Errorf("Porcentaje(%f, %f) error = %v, se esperaba %v", tt.valor, tt.total, err, tt.expectErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Porcentaje(%f, %f) error inesperado: %v", tt.valor, tt.total, err)
				return
			}
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("Porcentaje(%f, %f) = %f, se esperaba %f", tt.valor, tt.total, result, tt.expected)
			}
		})
	}
}

// TestContarPares verifica la funcion ContarPares.
func TestContarPares(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected int
	}{
		{"todos pares", []int{2, 4, 6, 8}, 4},
		{"todos impares", []int{1, 3, 5, 7}, 0},
		{"mezcla", []int{1, 2, 3, 4, 5, 6}, 3},
		{"slice vacio", []int{}, 0},
		{"un solo par", []int{2}, 1},
		{"cero es par", []int{0, 1, 2}, 2},
		{"negativos pares", []int{-2, -4, -3}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContarPares(tt.input)
			if result != tt.expected {
				t.Errorf("ContarPares(%v) = %d, se esperaba %d", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================
// BENCHMARKS: Midiendo el rendimiento de las funciones
// ============================================================

// BenchmarkSuma mide el rendimiento de Suma con 1000 elementos.
func BenchmarkSuma(b *testing.B) {
	nums := make([]float64, 1000)
	for i := range nums {
		nums[i] = float64(i)
	}
	// b.N se ajusta automaticamente para obtener una medicion estable
	for i := 0; i < b.N; i++ {
		Suma(nums)
	}
}

// BenchmarkPromedio mide el rendimiento de Promedio.
func BenchmarkPromedio(b *testing.B) {
	nums := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < b.N; i++ {
		Promedio(nums)
	}
}

// BenchmarkFactorial mide el rendimiento de Factorial con n=20.
func BenchmarkFactorial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Factorial(20)
	}
}

// BenchmarkEsPrimo mide el rendimiento de EsPrimo con un numero grande.
func BenchmarkEsPrimo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EsPrimo(104729) // Numero primo grande
	}
}

// BenchmarkFibonacci mide el rendimiento de Fibonacci con n=50.
func BenchmarkFibonacci(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fibonacci(50)
	}
}

// ============================================================
// EXAMPLES: Ejemplos que tambien son tests (y documentacion)
// ============================================================

// ExampleSuma demuestra como usar la funcion Suma.
// La salida esperada se verifica automaticamente.
func ExampleSuma() {
	result := Suma([]float64{1, 2, 3, 4, 5})
	fmt.Println(result)
	// Output: 15
}

// ExamplePromedio demuestra como usar la funcion Promedio.
func ExamplePromedio() {
	avg, err := Promedio([]float64{10, 20, 30})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%.1f\n", avg)
	// Output: 20.0
}

// ExampleEsPrimo demuestra como usar la funcion EsPrimo.
func ExampleEsPrimo() {
	fmt.Println(EsPrimo(7))
	fmt.Println(EsPrimo(10))
	// Output:
	// true
	// false
}

// ExampleFactorial demuestra como usar la funcion Factorial.
func ExampleFactorial() {
	result, _ := Factorial(5)
	fmt.Println(result)
	// Output: 120
}