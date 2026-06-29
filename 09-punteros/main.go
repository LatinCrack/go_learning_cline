package main

import "fmt"

// ============================================================
// 🧭 LECCIÓN 09 — PUNTEROS (POINTERS) EN GO
// ============================================================
// Un puntero es una variable que almacena la DIRECCIÓN DE MEMORIA
// de otra variable, en lugar de almacenar un valor directamente.
// ============================================================

// ─────────────────────────────────────────────────────
// 📌 1. FUNCIÓN QUE MODIFICA UNA COPIA (sin puntero)
// ─────────────────────────────────────────────────────
// Esta función recibe una COPIA del valor original.
// Cualquier cambio aquí NO afecta a la variable externa.
func duplicarSinPuntero(valor int) {
	valor = valor * 2
	fmt.Println("   → Dentro de duplicarSinPuntero:", valor)
	// Este cambio vive y muere aquí dentro. Afuera, nadie se entera.
}

// ─────────────────────────────────────────────────────
// 📌 2. FUNCIÓN QUE MODIFICA EL ORIGINAL (con puntero)
// ─────────────────────────────────────────────────────
// Esta función recibe un PUNTERO a int (*int).
// Esto significa: "dame la dirección donde vive el valor original".
// Así puedo modificar el valor EN SU LUGAR ORIGINAL.
func duplicarConPuntero(valor *int) {
	*valor = *valor * 2
	// *valor = "ve a la dirección que me pasaron y cambia el valor ahí"
	fmt.Println("   → Dentro de duplicarConPuntero:", *valor)
}

// ─────────────────────────────────────────────────────
// 📌 3. ESTRUCTURA Y PUNTEROS
// ─────────────────────────────────────────────────────
type Persona struct {
	Nombre string
	Edad   int
}

// Esta función recibe un puntero a Persona.
// Puede modificar directamente los campos de la persona original.
func cumpleaños(p *Persona) {
	p.Edad++
	// En Go, puedes acceder a campos con p.Edad en vez de (*p).Edad.
	// Es un "atajo" sintáctico que Go ofrece por comodidad.
	fmt.Println("   → ¡Feliz cumpleaños", p.Nombre, "! Ahora tienes", p.Edad, "años.")
}

// ─────────────────────────────────────────────────────
// 📌 4. FUNCIÓN QUE RETORNA UN PUNTERO
// ─────────────────────────────────────────────────────
// Go permite devolver punteros a variables locales.
// El compilador se encarga de que la memoria no se destruya
// mientras alguien siga usándola (escape analysis).
func crearPersona(nombre string, edad int) *Persona {
	p := Persona{Nombre: nombre, Edad: edad}
	return &p
	// &p = "aquí tienes la dirección donde vive esta persona"
}

// ─────────────────────────────────────────────────────
// 📌 5. PUNTERO A PUNTERO (niveles de indirección)
// ─────────────────────────────────────────────────────
func punteroAPuntero() {
	fmt.Println("\n🎯 Ejemplo 5: Puntero a puntero")

	x := 42
	p := &x   // p apunta a x
	pp := &p  // pp apunta a p, que apunta a x

	fmt.Println("   x        =", x)
	fmt.Println("   *p       =", *p)    // Valor de x a través de p
	fmt.Println("   **pp     =", **pp)  // Valor de x a través de pp → p → x

	**pp = 100 // Cambio el valor de x desde pp
	fmt.Println("   Después de **pp = 100:")
	fmt.Println("   x        =", x) // Ahora x es 100
}

// ─────────────────────────────────────────────────────
// 📌 6. SLICES Y PUNTEROS
// ─────────────────────────────────────────────────────
func modificarSlice(nums []int) {
	// Los slices ya contienen internamente un puntero al array subyacente.
	// Por eso, las modificaciones SÍ se reflejan afuera sin necesidad
	// de usar *[]int explícitamente.
	nums[0] = 999
}

func reasignarSlice(nums []int) {
	// Pero si REASIGNO el slice (append que cambia el puntero),
	// el cambio NO se refleja afuera.
	nums = append(nums, 888)
	fmt.Println("   → Dentro de reasignarSlice:", nums)
}

// ─────────────────────────────────────────────────────
// 📌 7. USO PRÁCTICO: INTERCAMBIO DE VALORES (swap)
// ─────────────────────────────────────────────────────
func swap(a, b *int) {
	*a, *b = *b, *a
	// Intercambia los valores en las direcciones de memoria originales.
}

// ─────────────────────────────────────────────────────
// 📌 8. NEW() — OTRA FORMA DE CREAR PUNTEROS
// ─────────────────────────────────────────────────────
func ejemploNew() {
	fmt.Println("\n🎯 Ejemplo 8: new()")

	// new(T) crea una variable anónima de tipo T con valor cero
	// y devuelve un puntero a ella.
	p := new(int)
	fmt.Println("   Valor inicial (*p):", *p) // 0 (valor cero de int)

	*p = 256
	fmt.Println("   Después de *p = 256:", *p)

	// Diferencia con &variable:
	// - new(int) → crea memoria nueva, devuelve *int
	// - &variable → toma la dirección de una variable existente
}

// ============================================================
// 🏁 FUNCIÓN PRINCIPAL
// ============================================================
func main() {
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println("🧭 LECCIÓN 09 — PUNTEROS (POINTERS) EN GO")
	fmt.Println("═══════════════════════════════════════════════")

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 1: Sin puntero — no se modifica el original
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 1: Sin puntero (copia)")

	numero := 10
	fmt.Println("   Antes de llamar a duplicarSinPuntero:", numero)
	duplicarSinPuntero(numero)
	fmt.Println("   Después de llamar a duplicarSinPuntero:", numero)
	// numero sigue siendo 10. La función trabajó con una copia.

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 2: Con puntero — SÍ se modifica el original
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 2: Con puntero (modifica original)")

	numero2 := 10
	fmt.Println("   Antes de llamar a duplicarConPuntero:", numero2)
	duplicarConPuntero(&numero2)
	// &numero2 = "pásale la dirección donde vive numero2"
	fmt.Println("   Después de llamar a duplicarConPuntero:", numero2)
	// numero2 ahora es 20. ¡La función lo modificó directamente!

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 3: Operador & (dirección) y * (desreferencia)
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 3: Operadores & y *")

	x := 42
	p := &x // p almacena la dirección de memoria de x

	fmt.Println("   x         =", x)
	fmt.Println("   &x        =", &x)  // Dirección de memoria de x
	fmt.Println("   p         =", p)   // Misma dirección
	fmt.Println("   *p        =", *p)  // Valor almacenado en esa dirección = 42

	*p = 100 // Modifica el valor en la dirección apuntada
	fmt.Println("   Después de *p = 100:")
	fmt.Println("   x         =", x)   // Ahora x es 100
	fmt.Println("   *p        =", *p)  // También 100

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 4: Structs con punteros
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 4: Structs con punteros")

	persona := Persona{Nombre: "Ana", Edad: 28}
	fmt.Println("   Antes:", persona)

	cumpleaños(&persona)
	// Le pasamos la dirección de la persona, no una copia.
	fmt.Println("   Después:", persona)

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 5: Puntero a puntero
	// ─────────────────────────────────────────────────────
	punteroAPuntero()

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 6: Slices y punteros
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 6: Slices y punteros")

	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("   Original:", nums)

	modificarSlice(nums)
	// Los slices internamente contienen un puntero al array,
	// así que las modificaciones a elementos SÍ se reflejan afuera.
	fmt.Println("   Después de modificarSlice:", nums)

	reasignarSlice(nums)
	// Pero si el slice se reasigna internamente (append que crece),
	// el cambio NO se refleja afuera.
	fmt.Println("   Después de reasignarSlice:", nums)

	// Si quieres que reasignarSlice SÍ afecte afuera,
	// necesitarías pasar *[]int o retornar el slice modificado.

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 7: Swap de valores
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 7: Swap de valores")

	a := 100
	b := 200
	fmt.Println("   Antes  → a:", a, "b:", b)

	swap(&a, &b)
	fmt.Println("   Después → a:", a, "b:", b)

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 8: new()
	// ─────────────────────────────────────────────────────
	ejemploNew()

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 9: Nil — el valor cero de los punteros
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 9: Nil — valor cero de punteros")

	var pNil *int // Sin inicializar → nil (no apunta a nada)
	fmt.Println("   pNil es nil:", pNil == nil)

	// ⚠️ IMPORTANTE: Desreferenciar un puntero nil causa un PANIC (crash).
	// fmt.Println(*pNil) // ← ¡NO HACER ESTO! Panic: runtime error.

	// Antes de usar un puntero, siempre verifica que no sea nil:
	if pNil != nil {
		fmt.Println("   Valor:", *pNil)
	} else {
		fmt.Println("   El puntero es nil, no apunta a ningún valor.")
	}

	// ─────────────────────────────────────────────────────
	// 📌 EJEMPLO 10: Puntero como señal de "opcional"
	// ─────────────────────────────────────────────────────
	fmt.Println("\n🎯 Ejemplo 10: Puntero como valor opcional")

	// En Go no existe "null" como en otros lenguajes.
	// Un puntero nil se usa como indicador de "no hay valor".
	busqueda := buscarUsuario("Carlos")
	if busqueda != nil {
		fmt.Println("   Usuario encontrado:", busqueda.Nombre)
	} else {
		fmt.Println("   Usuario no encontrado.")
	}

	busqueda2 := buscarUsuario("Ana")
	if busqueda2 != nil {
		fmt.Println("   Usuario encontrado:", busqueda2.Nombre)
	} else {
		fmt.Println("   Usuario no encontrado.")
	}

	fmt.Println("\n═══════════════════════════════════════════════")
	fmt.Println("🏁 FIN DE LA LECCIÓN 09 — PUNTEROS")
	fmt.Println("═══════════════════════════════════════════════")
}

// ─────────────────────────────────────────────────────
// 📌 FUNCIÓN AUXILIAR: Simula buscar un usuario en una "base de datos"
// ─────────────────────────────────────────────────────
func buscarUsuario(nombre string) *Persona {
	// "Base de datos" simulada
	usuarios := []Persona{
		{Nombre: "Ana", Edad: 28},
		{Nombre: "Luis", Edad: 35},
		{Nombre: "María", Edad: 42},
	}

	for _, u := range usuarios {
		if u.Nombre == nombre {
			return &u
			// Retornamos un puntero a la persona encontrada.
		}
	}

	// Si no encontramos al usuario, retornamos nil.
	// Esto es el equivalente Go de "no encontrado".
	return nil
}