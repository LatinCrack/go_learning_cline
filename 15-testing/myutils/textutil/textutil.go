// Package textutil proporciona funciones de procesamiento de texto
// que complementan la libreria estandar de Go.
package textutil

import (
	"strings"
	"unicode"
)

// Invertir invierte un string respetando caracteres Unicode (runes).
func Invertir(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// EsPalindromo verifica si un string es un palindromo.
// Ignora espacios y mayusculas/minusculas.
func EsPalindromo(s string) bool {
	limpio := strings.ToLower(strings.ReplaceAll(s, " ", ""))
	return limpio == Invertir(limpio)
}

// ContarPalabras cuenta el numero de palabras en un string.
func ContarPalabras(s string) int {
	return len(strings.Fields(s))
}

// TituloCapital convierte un string a formato titulo
// (primera letra de cada palabra en mayuscula).
func TituloCapital(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// SoloLetras elimina todos los caracteres que no sean letras.
func SoloLetras(s string) string {
	var resultado strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) {
			resultado.WriteRune(r)
		}
	}
	return resultado.String()
}

// Truncar trunca un string a n caracteres visibles (runes).
// Agrega "..." si se trunca.
func Truncar(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// ContarVocales cuenta el numero de vocales en un string.
func ContarVocales(s string) int {
	count := 0
	for _, r := range strings.ToLower(s) {
		switch r {
		case 'a', 'e', 'i', 'o', 'u':
			count++
		}
	}
	return count
}

// ReemplazarVocales reemplaza todas las vocales por un caracter dado.
func ReemplazarVocales(s string, reemplazo rune) string {
	var resultado strings.Builder
	for _, r := range s {
		lower := unicode.ToLower(r)
		if lower == 'a' || lower == 'e' || lower == 'i' || lower == 'o' || lower == 'u' {
			resultado.WriteRune(reemplazo)
		} else {
			resultado.WriteRune(r)
		}
	}
	return resultado.String()
}