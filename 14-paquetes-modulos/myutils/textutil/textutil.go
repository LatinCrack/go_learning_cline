// Package textutil proporciona funciones de manipulacion de texto.
package textutil

import (
	"strings"
	"unicode"
)

// Invertir devuelve una cadena con los caracteres en orden inverso.
func Invertir(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// EsPalindromo determina si una cadena es un palindromo.
func EsPalindromo(s string) bool {
	limpio := strings.ToLower(strings.ReplaceAll(s, " ", ""))
	return limpio == Invertir(limpio)
}

// ContarPalabras cuenta el numero de palabras en una cadena.
func ContarPalabras(s string) int {
	return len(strings.Fields(s))
}

// TituloCapital convierte una cadena a formato titulo.
func TituloCapital(s string) string {
	return strings.Title(strings.ToLower(s))
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

// Truncar corta una cadena a un maximo de n caracteres, agregando "..." si se trunca.
func Truncar(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}