package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// ═══════════════════════════════════════════════════════════════
//  🧪 Laboratorio de Go — Lección 06
//  Procesador de Datos CSV de Alta Velocidad
//  Arrays · Slices · append · copy · aliasing · capacity
// ═══════════════════════════════════════════════════════════════

// ─────────────────────────────────────────────────────────────
//  Sección 1: Arrays — tamaño fijo, paso por valor
// ─────────────────────────────────────────────────────────────

// Array de 5 temperaturas diarias (tamaño fijo, no se puede cambiar)
var temperaturasSemana [5]float64

// demostrarArrays muestra las propiedades fundamentales de los arrays
func demostrarArrays() {
	fmt.Println("   📌 Arrays: tamaño fijo, paso por valor")
	fmt.Println()

	// Declaración e inicialización
	notas := [4]float64{8.5, 7.2, 9.1, 6.8}
	fmt.Printf("   Array de notas: %v\n", notas)
	fmt.Printf("   Longitud (siempre fija): %d\n", len(notas))
	fmt.Println()

	// Los arrays se COPIAN al pasarlos a funciones
	fmt.Println("   ⚠️  Los arrays se pasan por VALOR (se copian):")
	fmt.Printf("   Antes de modificar: %v\n", notas)
	modificarArray(notas) // Se envía una COPIA
	fmt.Printf("   Después de modificar (original intacto): %v\n", notas)
	fmt.Println()

	// Comparación de arrays (solo si tienen el mismo tamaño)
	a1 := [3]int{1, 2, 3}
	a2 := [3]int{1, 2, 3}
	a3 := [3]int{1, 2, 4}
	fmt.Printf("   %v == %v → %t\n", a1, a2, a1 == a2)
	fmt.Printf("   %v == %v → %t\n", a1, a3, a1 == a3)
	fmt.Println()

	// Iterar con range
	fmt.Println("   Iterar con range (índice y valor):")
	for i, v := range notas {
		fmt.Printf("     Semana %d: %.1f\n", i+1, v)
	}
}

func modificarArray(arr [4]float64) {
	arr[0] = 99.9 // Modifica la COPIA, no el original
}

// ─────────────────────────────────────────────────────────────
//  Sección 2: Slices — ventanas dinámicas sobre arrays
// ─────────────────────────────────────────────────────────────

func demostrarSlices() {
	fmt.Println("   📌 Slices: ventanas dinámicas con puntero, longitud y capacidad")
	fmt.Println()

	// Crear un slice con make: make([]tipo, longitud, capacidad)
	s := make([]int, 3, 5)
	fmt.Printf("   make([]int, 3, 5): %v\n", s)
	fmt.Printf("   len = %d, cap = %d\n", len(s), cap(s))
	fmt.Println()

	// Crear un slice literal
	frutas := []string{"manzana", "banana", "cereza", "durazno"}
	fmt.Printf("   Slice literal: %v\n", frutas)
	fmt.Printf("   len = %d, cap = %d\n", len(frutas), cap(frutas))
	fmt.Println()

	// La diferencia clave: un slice es un DESCRIPTOR de 3 campos
	fmt.Println("   🔍 Un slice internamente tiene 3 componentes:")
	fmt.Println("   ┌─────────────────────────────────────────┐")
	fmt.Println("   │  puntero → apunta al array subyacente   │")
	fmt.Println("   │  len     → cuántos elementos tiene       │")
	fmt.Println("   │  cap     → cuántos puede tener           │")
	fmt.Println("   └─────────────────────────────────────────┘")
	fmt.Println()

	// Slicing: crear sub-slices (comparten memoria = ALIASING)
	fmt.Println("   ⚠️  Slicing y aliasing de memoria:")
	numeros := []int{10, 20, 30, 40, 50, 60}
	sub := numeros[1:4] // Elementos en índice 1, 2, 3
	fmt.Printf("   Original:  %v (len=%d, cap=%d)\n", numeros, len(numeros), cap(numeros))
	fmt.Printf("   Sub-slice: %v (len=%d, cap=%d)\n", sub, len(sub), cap(sub))

	// Modificar el sub-slice afecta al original (comparten el mismo array)
	sub[0] = 999
	fmt.Printf("   Después de sub[0]=999:\n")
	fmt.Printf("     Original:  %v ← ¡CAMBIÓ!\n", numeros)
	fmt.Printf("     Sub-slice: %v\n", sub)
	fmt.Println()

	// Para evitar aliasing: usar copy
	fmt.Println("   ✅ copy() para evitar aliasing:")
	copia := make([]int, len(sub))
	copy(copia, sub)
	copia[0] = 111
	fmt.Printf("   Copia modificada: %v\n", copia)
	fmt.Printf("   Original intacto: %v\n", numeros)
}

// ─────────────────────────────────────────────────────────────
//  Sección 3: append — el motor del crecimiento de slices
// ─────────────────────────────────────────────────────────────

func demostrarAppend() {
	fmt.Println("   📌 append: crecimiento automático de slices")
	fmt.Println()

	// Empezamos con un slice de capacidad 3
	s := make([]int, 0, 3)
	fmt.Printf("   Slice inicial: len=%d, cap=%d, %v\n", len(s), cap(s), s)

	// Agregar elementos uno por uno y observar el crecimiento
	for i := 1; i <= 10; i++ {
		s = append(s, i*10)
		fmt.Printf("   append(%2d): len=%d, cap=%d, %v\n", i*10, len(s), cap(s), s)
	}
	fmt.Println()

	// Demostración del peligro del aliasing con append
	fmt.Println("   ⚠️  EL PELIGRO: aliasing + append")
	original := []int{1, 2, 3, 4, 5}
	sub := original[1:3] // [2, 3], cap=4 (comparte con original)
	fmt.Printf("   original: %v (cap=%d)\n", original, cap(original))
	fmt.Printf("   sub:      %v (cap=%d)\n", sub, cap(sub))

	// append a sub CON espacio disponible → modifica el array compartido
	sub = append(sub, 99)
	fmt.Printf("   Después de append(sub, 99):\n")
	fmt.Printf("     original: %v ← ¡CAMBIÓ en posición 3!\n", original)
	fmt.Printf("     sub:      %v\n", sub)
	fmt.Println()

	// Solución: usar full slice expression [lo:hi:max]
	fmt.Println("   ✅ Solución: full slice expression [lo:hi:max]")
	safe := original[1:3:3] // limita la cap a 2 (no comparte espacio extra)
	fmt.Printf("   safe: %v (len=%d, cap=%d)\n", safe, len(safe), cap(safe))
	safe = append(safe, 88)
	fmt.Printf("   Después de append(safe, 88):\n")
	fmt.Printf("     original: %v ← ¡INTACTO!\n", original)
	fmt.Printf("     safe:     %v (nuevo array subyacente)\n", safe)
}

// ─────────────────────────────────────────────────────────────
//  Sección 4: Procesador CSV de datos de ventas
// ─────────────────────────────────────────────────────────────

// Venta representa una fila del CSV
type Venta struct {
	Producto  string
	Categoria string
	Cantidad  int
	Precio    float64
	Region    string
	Fecha     string
}

// Estadisticas calculadas por grupo
type Estadisticas struct {
	Grupo       string
	Total       float64
	Promedio    float64
	Mediana     float64
	DesvEstandar float64
	Cantidad    int
	Minimo      float64
	Maximo      float64
}

// ventasSimuladas genera datos de ejemplo para el procesamiento
func ventasSimuladas() []Venta {
	return []Venta{
		{Producto: "Laptop Pro", Categoria: "Electrónica", Cantidad: 5, Precio: 1200.00, Region: "Norte", Fecha: "2024-01-15"},
		{Producto: "Mouse USB", Categoria: "Accesorios", Cantidad: 50, Precio: 15.99, Region: "Sur", Fecha: "2024-01-16"},
		{Producto: "Teclado Mec", Categoria: "Accesorios", Cantidad: 30, Precio: 75.00, Region: "Norte", Fecha: "2024-01-16"},
		{Producto: "Monitor 27", Categoria: "Electrónica", Cantidad: 10, Precio: 350.00, Region: "Este", Fecha: "2024-01-17"},
		{Producto: "Webcam HD", Categoria: "Accesorios", Cantidad: 25, Precio: 45.00, Region: "Oeste", Fecha: "2024-01-17"},
		{Producto: "Laptop Air", Categoria: "Electrónica", Cantidad: 8, Precio: 950.00, Region: "Sur", Fecha: "2024-01-18"},
		{Producto: "USB 64GB", Categoria: "Almacenamiento", Cantidad: 100, Precio: 8.50, Region: "Norte", Fecha: "2024-01-18"},
		{Producto: "Disco SSD", Categoria: "Almacenamiento", Cantidad: 20, Precio: 89.99, Region: "Este", Fecha: "2024-01-19"},
		{Producto: "Audífonos", Categoria: "Accesorios", Cantidad: 40, Precio: 35.00, Region: "Oeste", Fecha: "2024-01-19"},
		{Producto: "Tablet 10", Categoria: "Electrónica", Cantidad: 12, Precio: 450.00, Region: "Sur", Fecha: "2024-01-20"},
		{Producto: "Mouse Pad", Categoria: "Accesorios", Cantidad: 60, Precio: 12.00, Region: "Norte", Fecha: "2024-01-20"},
		{Producto: "Cable HDMI", Categoria: "Accesorios", Cantidad: 80, Precio: 9.99, Region: "Este", Fecha: "2024-01-21"},
		{Producto: "Laptop Gaming", Categoria: "Electrónica", Cantidad: 3, Precio: 2500.00, Region: "Oeste", Fecha: "2024-01-21"},
		{Producto: "Hub USB", Categoria: "Accesorios", Cantidad: 35, Precio: 25.00, Region: "Norte", Fecha: "2024-01-22"},
		{Producto: "Disco HDD", Categoria: "Almacenamiento", Cantidad: 15, Precio: 55.00, Region: "Sur", Fecha: "2024-01-22"},
		{Producto: "Monitor 24", Categoria: "Electrónica", Cantidad: 18, Precio: 220.00, Region: "Este", Fecha: "2024-01-23"},
		{Producto: "Webcam 4K", Categoria: "Accesorios", Cantidad: 10, Precio: 89.00, Region: "Oeste", Fecha: "2024-01-23"},
		{Producto: "RAM 16GB", Categoria: "Almacenamiento", Cantidad: 25, Precio: 65.00, Region: "Norte", Fecha: "2024-01-24"},
		{Producto: "Teclado RGB", Categoria: "Accesorios", Cantidad: 22, Precio: 95.00, Region: "Sur", Fecha: "2024-01-24"},
		{Producto: "Laptop Ultrabook", Categoria: "Electrónica", Cantidad: 6, Precio: 1400.00, Region: "Este", Fecha: "2024-01-25"},
	}
}

// filtrarVentas usa slicing para crear subconjuntos sin modificar el original
func filtrarVentas(ventas []Venta, filtro func(Venta) bool) []Venta {
	resultado := make([]Venta, 0) // Slice vacío con cap=0
	for _, v := range ventas {
		if filtro(v) {
			resultado = append(resultado, v)
		}
	}
	return resultado
}

// ingresoTotal calcula el ingreso (cantidad * precio) de un slice de ventas
func ingresoTotal(ventas []Venta) float64 {
	total := 0.0
	for _, v := range ventas {
		total += float64(v.Cantidad) * v.Precio
	}
	return total
}

// agruparPorCampo agrupa ventas usando un map de slices
func agruparPorCampo(ventas []Venta, campo func(Venta) string) map[string][]Venta {
	grupos := make(map[string][]Venta)
	for _, v := range ventas {
		key := campo(v)
		grupos[key] = append(grupos[key], v) // append al slice del map
	}
	return grupos
}

// calcularEstadisticas calcula stats de un grupo de valores
func calcularEstadisticas(nombre string, valores []float64) Estadisticas {
	if len(valores) == 0 {
		return Estadisticas{Grupo: nombre}
	}

	// Ordenar para calcular mediana
	sorted := make([]float64, len(valores))
	copy(sorted, valores) // copy para no modificar el original
	sort.Float64s(sorted)

	total := 0.0
	min := sorted[0]
	max := sorted[len(sorted)-1]

	for _, v := range sorted {
		total += v
	}

	promedio := total / float64(len(sorted))

	// Mediana
	var mediana float64
	n := len(sorted)
	if n%2 == 0 {
		mediana = (sorted[n/2-1] + sorted[n/2]) / 2
	} else {
		mediana = sorted[n/2]
	}

	// Desviación estándar
	sumaCuadrados := 0.0
	for _, v := range sorted {
		diff := v - promedio
		sumaCuadrados += diff * diff
	}
	desvEstandar := math.Sqrt(sumaCuadrados / float64(len(sorted)))

	return Estadisticas{
		Grupo:        nombre,
		Total:        total,
		Promedio:     promedio,
		Mediana:      mediana,
		DesvEstandar: desvEstandar,
		Cantidad:     len(sorted),
		Minimo:       min,
		Maximo:       max,
	}
}

// generarReporteCSV genera un CSV de reporte como string
func generarReporteCSV(ventas []Venta) string {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Encabezado
	writer.Write([]string{"Producto", "Categoria", "Cantidad", "Ingreso", "Region", "Fecha"})

	for _, v := range ventas {
		ingreso := float64(v.Cantidad) * v.Precio
		writer.Write([]string{
			v.Producto,
			v.Categoria,
			strconv.Itoa(v.Cantidad),
			fmt.Sprintf("%.2f", ingreso),
			v.Region,
			v.Fecha,
		})
	}
	writer.Flush()
	return sb.String()
}

// mostrarEstadisticas imprime las estadísticas formateadas
func mostrarEstadisticas(stats []Estadisticas) {
	fmt.Printf("     %-15s %10s %10s %10s %12s %5s %10s %10s\n",
		"GRUPO", "TOTAL", "PROMEDIO", "MEDIANA", "DESV.STD", "N", "MIN", "MAX")
	fmt.Println("     " + strings.Repeat("─", 95))

	for _, s := range stats {
		fmt.Printf("     %-15s %10.2f %10.2f %10.2f %12.2f %5d %10.2f %10.2f\n",
			s.Grupo, s.Total, s.Promedio, s.Mediana, s.DesvEstandar, s.Cantidad, s.Minimo, s.Maximo)
	}
}

// ─────────────────────────────────────────────────────────────
//  Sección 5: Demostración de operaciones avanzadas con slices
// ─────────────────────────────────────────────────────────────

func demostrarOperacionesAvanzadas() {
	fmt.Println("   📌 Operaciones avanzadas con slices")
	fmt.Println()

	// insertar: insertar un elemento en posición específica
	fmt.Println("   🔧 Insertar elemento en posición 2:")
	numeros := []int{1, 2, 4, 5}
	fmt.Printf("     Antes:  %v\n", numeros)

	// Técnica: append con slicing para crear espacio
	numeros = append(numeros[:2+1], numeros[2:]...) // Abrir espacio en índice 2
	numeros[2] = 3                                   // Insertar el valor
	fmt.Printf("     Después: %v\n", numeros)
	fmt.Println()

	// eliminar: quitar elemento en posición 1
	fmt.Println("   🔧 Eliminar elemento en posición 1:")
	letras := []string{"a", "b", "c", "d", "e"}
	fmt.Printf("     Antes:  %v\n", letras)
	letras = append(letras[:1], letras[2:]...) // Saltar índice 1
	fmt.Printf("     Después: %v\n", letras)
	fmt.Println()

	// eliminar sin mantener orden (más eficiente)
	fmt.Println("   🔧 Eliminar sin mantener orden (O(1)):")
	items := []string{"x", "y", "z", "w", "v"}
	fmt.Printf("     Antes:  %v\n", items)
	// Reemplazar el elemento a eliminar con el último
	items[1] = items[len(items)-1]
	items = items[:len(items)-1]
	fmt.Printf("     Después (eliminar 'y'): %v\n", items)
	fmt.Println()

	// filtrar in-place (sin crear nuevo slice)
	fmt.Println("   🔧 Filtrar in-place (evita allocations):")
	valores := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Printf("     Original:          %v\n", valores)

	n := 0
	for _, v := range valores {
		if v%2 == 0 { // Solo pares
			valores[n] = v
			n++
		}
	}
	valores = valores[:n]
	fmt.Printf("     Solo pares (in-place): %v\n", valores)
	fmt.Println()

	// Revertir un slice in-place
	fmt.Println("   🔧 Revertir slice in-place:")
	abc := []string{"A", "B", "C", "D", "E"}
	fmt.Printf("     Original: %v\n", abc)
	for i, j := 0, len(abc)-1; i < j; i, j = i+1, j-1 {
		abc[i], abc[j] = abc[j], abc[i]
	}
	fmt.Printf("     Revertido: %v\n", abc)
}

// ─────────────────────────────────────────────────────────────
//  FUNCIÓN PRINCIPAL
// ─────────────────────────────────────────────────────────────

func main() {
	// ── Banner ──
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   🧪 LABORATORIO DE GO — LECCIÓN 06                          ║")
	fmt.Println("║   Arrays, Slices y el Secreto del Runtime de Go               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ── 1. ARRAYS ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   1️⃣  Arrays: tamaño fijo, paso por valor")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarArrays()

	// ── 2. SLICES ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   2️⃣  Slices: ventanas dinámicas sobre arrays")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarSlices()

	// ── 3. APPEND ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   3️⃣  append: el motor del crecimiento de slices")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarAppend()

	// ── 4. PROCESADOR CSV ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   4️⃣  Procesador de Datos CSV: slices en acción")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	ventas := ventasSimuladas()
	fmt.Printf("   📊 %d ventas cargadas para procesamiento\n\n", len(ventas))

	// 4a. Filtrar por categoría
	fmt.Println("   🔍 Filtrar ventas de Electrónica:")
	electrónica := filtrarVentas(ventas, func(v Venta) bool {
		return v.Categoria == "Electrónica"
	})
	for _, v := range electrónica {
		ingreso := float64(v.Cantidad) * v.Precio
		fmt.Printf("     → %-20s  %d uds × $%.2f = $%10.2f\n", v.Producto, v.Cantidad, v.Precio, ingreso)
	}
	fmt.Printf("   💰 Ingreso total Electrónica: $%.2f\n\n", ingresoTotal(electrónica))

	// 4b. Agrupar por categoría y calcular estadísticas
	fmt.Println("   📈 Estadísticas por CATEGORÍA:")
	porCategoria := agruparPorCampo(ventas, func(v Venta) string { return v.Categoria })
	var statsCategoria []Estadisticas
	for cat, vs := range porCategoria {
		ingresos := make([]float64, len(vs))
		for i, v := range vs {
			ingresos[i] = float64(v.Cantidad) * v.Precio
		}
		statsCategoria = append(statsCategoria, calcularEstadisticas(cat, ingresos))
	}
	// Ordenar por total descendente
	sort.Slice(statsCategoria, func(i, j int) bool {
		return statsCategoria[i].Total > statsCategoria[j].Total
	})
	mostrarEstadisticas(statsCategoria)
	fmt.Println()

	// 4c. Agrupar por región
	fmt.Println("   📈 Estadísticas por REGIÓN:")
	porRegion := agruparPorCampo(ventas, func(v Venta) string { return v.Region })
	var statsRegion []Estadisticas
	for reg, vs := range porRegion {
		ingresos := make([]float64, len(vs))
		for i, v := range vs {
			ingresos[i] = float64(v.Cantidad) * v.Precio
		}
		statsRegion = append(statsRegion, calcularEstadisticas(reg, ingresos))
	}
	sort.Slice(statsRegion, func(i, j int) bool {
		return statsRegion[i].Total > statsRegion[j].Total
	})
	mostrarEstadisticas(statsRegion)
	fmt.Println()

	// 4d. Top 5 ventas por ingreso
	fmt.Println("   🏆 Top 5 ventas por ingreso:")
	// Usar copy para no modificar el slice original
	ventasCopia := make([]Venta, len(ventas))
	copy(ventasCopia, ventas)
	sort.Slice(ventasCopia, func(i, j int) bool {
		ingresoI := float64(ventasCopia[i].Cantidad) * ventasCopia[i].Precio
		ingresoJ := float64(ventasCopia[j].Cantidad) * ventasCopia[j].Precio
		return ingresoI > ingresoJ
	})
	for i := 0; i < 5 && i < len(ventasCopia); i++ {
		v := ventasCopia[i]
		ingreso := float64(v.Cantidad) * v.Precio
		fmt.Printf("     %d. %-20s  %s  → $%10.2f\n", i+1, v.Producto, v.Region, ingreso)
	}
	fmt.Println()

	// 4e. Generar CSV de reporte
	fmt.Println("   📄 CSV de reporte generado:")
	csvReport := generarReporteCSV(ventas)
	lineas := strings.Split(csvReport, "\n")
	for i, linea := range lineas {
		if i < 8 { // Mostrar primeras 8 líneas
			fmt.Printf("     %s\n", linea)
		}
	}
	if len(lineas) > 8 {
		fmt.Printf("     ... (%d líneas más)\n", len(lineas)-8)
	}
	fmt.Println()

	// ── 5. OPERACIONES AVANZADAS ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   5️⃣  Operaciones avanzadas: insertar, eliminar, filtrar, revertir")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarOperacionesAvanzadas()

	// ── 6. RESUMEN ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   📊 Resumen de la demostración")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("   ✅ Arrays demostrados:       tamaño fijo, paso por valor, comparación\n")
	fmt.Printf("   ✅ Slices demostrados:       make, literal, slicing, len, cap\n")
	fmt.Printf("   ✅ Aliasing demostrado:      sub-slices comparten memoria\n")
	fmt.Printf("   ✅ copy() demostrado:        copia independiente de slices\n")
	fmt.Printf("   ✅ append() demostrado:      crecimiento automático y full slice expression\n")
	fmt.Printf("   ✅ Procesador CSV:           filtrar, agrupar, estadísticas, ordenar\n")
	fmt.Printf("   ✅ Operaciones avanzadas:    insertar, eliminar, filtrar in-place, revertir\n")
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Println("   ✅ Todos los conceptos de arrays y slices ejecutados correctamente")
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Println()
}