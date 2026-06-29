// Package mathutil proporciona funciones matematicas utiles
// que no vienen incluidas en la libreria estandar de Go.
//
// Todas las funciones son puras (sin efectos secundarios)
// y seguras para uso concurrente (stateless).
package mathutil

import (
	"errors"
	"math"
)

// Errores del paquete — exportados para que el importador pueda compararlos.
var (
	ErrEmptySlice    = errors.New("el slice no puede estar vacio")
	ErrNegativeValue = errors.New("el valor no puede ser negativo")
	ErrZeroDivisor   = errors.New("el divisor no puede ser cero")
)

// Suma retorna la suma total de un slice de float64.
// Retorna 0 si el slice esta vacio (no es un error).
func Suma(nums []float64) float64 {
	total := 0.0
	for _, n := range nums {
		total += n
	}
	return total
}

// Promedio calcula el promedio aritmetico de un slice de float64.
// Retorna ErrEmptySlice si el slice esta vacio.
//
// Ejemplo:
//
//	avg, err := mathutil.Promedio([]float64{10, 20, 30})
//	// avg = 20.0, err = nil
func Promedio(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	return Suma(nums) / float64(len(nums)), nil
}

// Maximo retorna el valor maximo de un slice de float64.
// Retorna ErrEmptySlice si el slice esta vacio.
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

// Minimo retorna el valor minimo de un slice de float64.
// Retorna ErrEmptySlice si el slice esta vacio.
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

// Mediana calcula la mediana de un slice de float64.
// Retorna ErrEmptySlice si el slice esta vacio.
func Mediana(nums []float64) (float64, error) {
	if len(nums) == 0 {
		return 0, ErrEmptySlice
	}
	sorted := make([]float64, len(nums))
	copy(sorted, nums)
	// Ordenamiento por insercion (eficiente para slices pequenos)
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2, nil
	}
	return sorted[n/2], nil
}

// DesviacionEstandar calcula la desviacion estandar poblacional.
// Retorna ErrEmptySlice si el slice esta vacio.
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

// Factorial calcula el factorial de n (n!).
// Retorna ErrNegativeValue si n es negativo.
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

// Fibonacci genera los primeros n numeros de la secuencia Fibonacci.
// Retorna ErrNegativeValue si n es negativo.
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

// EsPrimo verifica si un numero entero es primo.
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

// Clamp restringe un valor al rango [min, max].
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
// Retorna ErrZeroDivisor si el total es cero.
func Porcentaje(valor, total float64) (float64, error) {
	if total == 0 {
		return 0, ErrZeroDivisor
	}
	return (valor / total) * 100, nil
}

// ContarPares cuenta cuantos numeros pares hay en el slice.
func ContarPares(nums []int) int {
	count := 0
	for _, n := range nums {
		if n%2 == 0 {
			count++
		}
	}
	return count
}