// Package mathutil proporciona funciones matematicas utiles.
package mathutil

import (
	"errors"
	"math"
)

// ErrEmptySlice se devuelve cuando se pasa un slice vacio.
var ErrEmptySlice = errors.New("el slice no puede estar vacio")

// ErrNegativeValue se devuelve cuando se pasa un numero negativo.
var ErrNegativeValue = errors.New("el valor no puede ser negativo")

// ErrZeroValue se devuelve cuando se pasa cero a una funcion que no lo acepta.
var ErrZeroValue = errors.New("el valor no puede ser cero")

// Suma devuelve la suma de todos los elementos de un slice de float64.
func Suma(nums []float64) float64 {
	total := 0.0
	for _, n := range nums {
		total += n
	}
	return total
}

// Promedio calcula el promedio aritmetico de un slice de float64.
func Promedio(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	return Suma(nums) / float64(len(nums)), nil
}

// Maximo devuelve el valor maximo de un slice de float64.
func Maximo(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	max := nums[0]
	for _, n := range nums[1:] {
		if n > max {
			max = n
		}
	}
	return max, nil
}

// Minimo devuelve el valor minimo de un slice de float64.
func Minimo(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	min := nums[0]
	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
	}
	return min, nil
}

// DesviacionEstandar calcula la desviacion estandar poblacional.
func DesviacionEstandar(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	prom, _ := Promedio(nums)
	sumaCuadrados := 0.0
	for _, n := range nums {
		diff := n - prom
		sumaCuadrados += diff * diff
	}
	return math.Sqrt(sumaCuadrados / float64(len(nums))), nil
}

// Mediana calcula la mediana de un slice de float64.
func Mediana(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	copia := make([]float64, len(nums))
	copy(copia, nums)
	for i := 0; i < len(copia); i++ {
		for j := i + 1; j < len(copia); j++ {
			if copia[i] > copia[j] {
				copia[i], copia[j] = copia[j], copia[i]
			}
		}
	}
	n := len(copia)
	if n%2 == 1 {
		return copia[n/2], nil
	}
	return (copia[n/2-1] + copia[n/2]) / 2.0, nil
}

// Factorial calcula el factorial de un numero entero no negativo.
func Factorial(n int) (int, error) {
	if n < 0 {
		return 0, ErrNegativeValue
	}
	if n == 0 || n == 1 {
		return 1, nil
	}
	resultado := 1
	for i := 2; i <= n; i++ {
		resultado *= i
	}
	return resultado, nil
}

// Fibonacci devuelve los primeros n numeros de la secuencia Fibonacci.
func Fibonacci(n int) ([]int, error) {
	if n < 0 {
		return nil, ErrNegativeValue
	}
	if n == 0 {
		return []int{}, nil
	}
	if n == 1 {
		return []int{0}, nil
	}
	seq := make([]int, n)
	seq[0] = 0
	seq[1] = 1
	for i := 2; i < n; i++ {
		seq[i] = seq[i-1] + seq[i-2]
	}
	return seq, nil
}

// EsPrimo determina si un numero entero es primo.
func EsPrimo(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i <= int(math.Sqrt(float64(n))); i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// Clamp restringe un valor entre un minimo y un maximo.
func Clamp(valor, min, max int) int {
	if valor < min {
		return min
	}
	if valor > max {
		return max
	}
	return valor
}

// Porcentaje calcula el porcentaje de un valor respecto a un total.
func Porcentaje(valor, total float64) (float64, error) {
	if total == 0 {
		return 0, ErrZeroValue
	}
	return (valor / total) * 100, nil
}