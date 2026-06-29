package main

import "fmt"

func main() {

	// ====================================================
	// 🏪 SISTEMA DE INVENTARIO - FERRETERÍA "DON CLODO"
	// ====================================================

	// --- Datos del producto 1: Martillo ---

	var nombreProducto1 string = "Martillo de uña"
	var precioProducto1 float64 = 25.50
	var cantidadProducto1 int = 15
	var enPromocion1 bool = true

	// --- Datos del producto 2: Destornillador ---

	var nombreProducto2 string = "Destornillador Phillips"
	var precioProducto2 float64 = 12.75
	var cantidadProducto2 int = 40
	var enPromocion2 bool = false

	// --- Datos del producto 3: Cinta métrica ---

	nombreProducto3 := "Cinta métrica 5m" // Inferencia: string
	precioProducto3 := 18.90              // Inferencia: float64
	cantidadProducto3 := 22               // Inferencia: int
	enPromocion3 := true                  // Inferencia: bool

	// ====================================================
	// 📊 MOSTRAR INVENTARIO
	// ====================================================

	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║   🏪 INVENTARIO - FERRETERÍA DON CLODO      ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	// Producto 1 con fmt.Printf (formato controlado)
	fmt.Printf("📦 Producto:    %s\n", nombreProducto1)
	fmt.Printf("   💰 Precio:    $%.2f\n", precioProducto1)
	fmt.Printf("   📦 Cantidad:  %d unidades\n", cantidadProducto1)
	fmt.Printf("   🏷️  Promoción: %t\n", enPromocion1)
	fmt.Println("   ─────────────────────────────")

	// Producto 2
	fmt.Printf("📦 Producto:    %s\n", nombreProducto2)
	fmt.Printf("   💰 Precio:    $%.2f\n", precioProducto2)
	fmt.Printf("   📦 Cantidad:  %d unidades\n", cantidadProducto2)
	fmt.Printf("   🏷️  Promoción: %t\n", enPromocion2)
	fmt.Println("   ─────────────────────────────")

	// Producto 3
	fmt.Printf("📦 Producto:    %s\n", nombreProducto3)
	fmt.Printf("   💰 Precio:    $%.2f\n", precioProducto3)
	fmt.Printf("   📦 Cantidad:  %d unidades\n", cantidadProducto3)
	fmt.Printf("   🏷️  Promoción: %t\n", enPromocion3)
	fmt.Println("   ─────────────────────────────")

	// ====================================================
	// 🔄 CONVERSIÓN DE TIPOS
	// ====================================================

	fmt.Println()
	fmt.Println("🔄 Demostración de conversión de tipos:")
	fmt.Println()

	// Tenemos la cantidad como int, pero necesitamos hacer cálculos decimales
	var cantidadFloat float64 = float64(cantidadProducto1)
	iva := cantidadFloat * 0.18
	fmt.Printf("   IVA sobre %d martillos: $%.2f\n", cantidadProducto1, iva)

	// Tenemos un precio float64, pero necesitamos redondear a entero para etiquetas
	var precioEntero int = int(precioProducto1)
	fmt.Printf("   Precio para etiqueta (sin decimales): $%d\n", precioEntero)

	// ====================================================
	// 🎯 VALORES CERO
	// ====================================================

	fmt.Println()
	fmt.Println("🎯 Valores cero (variables sin inicializar):")
	fmt.Println()

	var sinEntero int
	var sinDecimal float64
	var sinTexto string
	var sinBooleano bool

	fmt.Printf("   int sin valor:    %d\n", sinEntero)
	fmt.Printf("   float64 sin valor: %f\n", sinDecimal)
	fmt.Printf("   string sin valor:  [%s]\n", sinTexto)
	fmt.Printf("   bool sin valor:    %t\n", sinBooleano)

	// ====================================================
	// 🔠 RUNAS (CARACTERES UNICODE)
	// ====================================================

	fmt.Println()
	fmt.Println("🔤 Runas (caracteres Unicode):")
	fmt.Println()

	var letraA rune = 'A'
	var letraEne rune = 'ñ'
	var emoji rune = '😀'

	fmt.Printf("   'A'   → valor numérico: %d\n", letraA)
	fmt.Printf("   'ñ'   → valor numérico: %d\n", letraEne)
	fmt.Printf("   '😀'  → valor numérico: %d\n", emoji)

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("   ✅ Inventario cargado correctamente")
	fmt.Println("═══════════════════════════════════════════════")
}