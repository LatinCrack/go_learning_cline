package textutil

import "testing"

// ============================================================
// TESTS TABLE-DRIVEN para funciones de texto
// ============================================================

func TestInvertir(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"palabra simple", "hola", "aloh"},
		{"palindromo", "ana", "ana"},
		{"vacio", "", ""},
		{"un caracter", "a", "a"},
		{"con espacios", "hola mundo", "odnum aloh"},
		{"unicode", "café", "éfac"},
		{"emoji", "ab🎉", "🎉ba"},
		{"mayusculas", "Go", "oG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Invertir(tt.input)
			if result != tt.expected {
				t.Errorf("Invertir(%q) = %q, se esperaba %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEsPalindromo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"palindromo simple", "ana", true},
		{"palindromo con espacios", "anita lava la tina", true},
		{"no palindromo", "hola", false},
		{"palindromo mayusculas", "Reconocer", true},
		{"vacio", "", true},
		{"un caracter", "a", true},
		{"palindromo numerico tambien es palindromo", "12321", true},
		{"palindromo frase", "A man a plan a canal Panama", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EsPalindromo(tt.input)
			if result != tt.expected {
				t.Errorf("EsPalindromo(%q) = %v, se esperaba %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContarPalabras(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"frase normal", "hola mundo", 2},
		{"una palabra", "hola", 1},
		{"vacio", "", 0},
		{"multiples espacios", "hola   mundo   cruel", 3},
		{"solo espacios", "   ", 0},
		{"tabs y newlines", "hola\tmundo\ngo", 3},
		{"frase larga", "Go es un lenguaje de programacion", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContarPalabras(tt.input)
			if result != tt.expected {
				t.Errorf("ContarPalabras(%q) = %d, se esperaba %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTituloCapital(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"minusculas", "hola mundo", "Hola Mundo"},
		{"mayusculas", "HOLA MUNDO", "Hola Mundo"},
		{"mezcla", "hOLA mUNDO", "Hola Mundo"},
		{"una palabra", "go", "Go"},
		{"vacio", "", ""},
		{"palabras cortas", "a b c", "A B C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TituloCapital(tt.input)
			if result != tt.expected {
				t.Errorf("TituloCapital(%q) = %q, se esperaba %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSoloLetras(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"con numeros", "hola123", "hola"},
		{"con simbolos", "hola!@#$", "hola"},
		{"solo letras", "hola", "hola"},
		{"vacio", "", ""},
		{"solo numeros", "12345", ""},
		{"mezcla completa", "Go 1.21 es genial!", "Goesgenial"},
		{"con acentos", "café résumé", "caférésumé"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SoloLetras(tt.input)
			if result != tt.expected {
				t.Errorf("SoloLetras(%q) = %q, se esperaba %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected string
	}{
		{"sin truncar", "hola", 10, "hola"},
		{"truncar exacto", "hola", 4, "hola"},
		{"truncar corto", "hola mundo", 5, "hola ..."},
		{"truncar uno", "hola", 1, "h..."},
		{"vacio", "", 5, ""},
		{"unicode", "café", 2, "ca..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncar(tt.input, tt.n)
			if result != tt.expected {
				t.Errorf("Truncar(%q, %d) = %q, se esperaba %q", tt.input, tt.n, result, tt.expected)
			}
		})
	}
}

func TestContarVocales(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"palabra normal", "hola", 2},
		{"sin vocales", "bcdfg", 0},
		{"todas vocales", "aeiou", 5},
		{"mayusculas", "AEIOU", 5},
		{"vacio", "", 0},
		{"mezcla", "Programacion en Go", 7},
		{"una vocal", "a", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContarVocales(tt.input)
			if result != tt.expected {
				t.Errorf("ContarVocales(%q) = %d, se esperaba %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReemplazarVocales(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		reemplazo rune
		expected  string
	}{
		{"reemplazar con asterisco", "hola", '*', "h*l*"},
		{"reemplazar con guion", "hola mundo", '-', "h-l- m-nd-"},
		{"sin vocales", "bcdfg", '*', "bcdfg"},
		{"vacio", "", '*', ""},
		{"todas vocales", "aeiou", 'x', "xxxxx"},
		{"mayusculas", "Go", 'x', "Gx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReemplazarVocales(tt.input, tt.reemplazo)
			if result != tt.expected {
				t.Errorf("ReemplazarVocales(%q, %q) = %q, se esperaba %q", tt.input, string(tt.reemplazo), result, tt.expected)
			}
		})
	}
}

// ============================================================
// BENCHMARKS para funciones de texto
// ============================================================

func BenchmarkInvertir(b *testing.B) {
	s := "Esto es una cadena de prueba para benchmarking en Go"
	for i := 0; i < b.N; i++ {
		Invertir(s)
	}
}

func BenchmarkEsPalindromo(b *testing.B) {
	s := "anita lava la tina"
	for i := 0; i < b.N; i++ {
		EsPalindromo(s)
	}
}

func BenchmarkContarPalabras(b *testing.B) {
	s := "Go es un lenguaje de programacion creado por Google"
	for i := 0; i < b.N; i++ {
		ContarPalabras(s)
	}
}

func BenchmarkTituloCapital(b *testing.B) {
	s := "hola mundo desde el lenguaje go"
	for i := 0; i < b.N; i++ {
		TituloCapital(s)
	}
}

func BenchmarkContarVocales(b *testing.B) {
	s := "Esto es una cadena de prueba para benchmarking en Go"
	for i := 0; i < b.N; i++ {
		ContarVocales(s)
	}
}