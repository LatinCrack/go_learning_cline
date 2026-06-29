package main

// ============================================================
// 🔮 LECCIÓN 19: GENERICS, REFLECT Y METAPROGRAMACIÓN EN GO
// ============================================================
//
// Este programa DEMUESTRA en acción los dos conceptos más
// avanzados de Go: Generics (type parameters) y Reflect.
//
// Parte 1: Colecciones funcionales genéricas (Map, Filter, Reduce...)
// Parte 2: Stack genérico (tipo genérico con métodos)
// Parte 3: Validador de structs con Reflect (lee tags en runtime)
//
// Ejecutar: go run main.go
// ============================================================

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ============================================================
// PARTE 1: FUNCIONES GENÉRICAS DE COLECCIONES
// ============================================================
//
// Estas funciones usan type parameters [T any] para funcionar
// con CUALQUIER tipo. Es como tener un martillo universal que
// acepta cualquier clavo.
//
// Antes de Go 1.18, tendrías que escribir una función por cada
// tipo: MapInt, MapString, MapFloat64... 😫
// Ahora, UNA sola función sirve para todos.

// ------------------------------------------------------------
// Map: Transforma cada elemento de un slice aplicando una función.
//
// Analogía: Es una cinta transportadora donde cada producto pasa
// por una máquina que lo transforma. El producto entra como T
// y sale como U.
//
// Tipo 1 (T): el tipo de los elementos originales
// Tipo 2 (U): el tipo de los elementos transformados
// ------------------------------------------------------------
func Map[T any, U any](slice []T, fn func(T) U) []U {
	// Creamos un slice del mismo tamaño para los resultados.
	// Usamos make con len(slice) para evitar re-allocations.
	resultado := make([]U, len(slice))

	// Recorremos cada elemento del slice original.
	// La función fn transforma T → U.
	for i, v := range slice {
		resultado[i] = fn(v)
	}

	return resultado
}

// ------------------------------------------------------------
// Filter: Selecciona solo los elementos que cumplen una condición.
//
// Analogía: Es un colador de cocina. Solo pasan los elementos
// que caben por los agujeros (cumplen la condición).
//
// Tipo (T): el tipo de los elementos del slice.
// ------------------------------------------------------------
func Filter[T any](slice []T, fn func(T) bool) []T {
	// Empezamos con un slice vacío. No sabemos cuántos elementos
	// pasarán el filtro, así que usamos append (no make con tamaño fijo).
	var resultado []T

	// Recorremos cada elemento. Si fn devuelve true, lo incluimos.
	for _, v := range slice {
		if fn(v) {
			resultado = append(resultado, v)
		}
	}

	return resultado
}

// ------------------------------------------------------------
// Reduce: Combina todos los elementos de un slice en un solo valor.
//
// Analogía: Es una licuadora. Metes muchos ingredientes (elementos)
// y produces un solo batido (valor acumulado).
//
// Tipo 1 (T): el tipo de los elementos del slice.
// Tipo 2 (U): el tipo del valor acumulado (puede ser diferente).
// ------------------------------------------------------------
func Reduce[T any, U any](slice []T, inicial U, fn func(U, T) U) U {
	// Empezamos con el valor inicial proporcionado.
	acumulador := inicial

	// Por cada elemento, aplicamos la función que combina
	// el acumulador actual con el nuevo elemento.
	for _, v := range slice {
		acumulador = fn(acumulador, v)
	}

	return acumulador
}

// ------------------------------------------------------------
// Contains: Verifica si un elemento existe en un slice.
//
// Usamos [T comparable] porque necesitamos el operador ==
// para comparar cada elemento con el objetivo.
//
// Sin comparable, no podríamos usar == (slices, maps, funcs
// no son comparables).
// ------------------------------------------------------------
func Contains[T comparable](slice []T, objetivo T) bool {
	for _, v := range slice {
		if v == objetivo {
			return true // Encontrado
		}
	}
	return false
}

// ------------------------------------------------------------
// Unique: Elimina elementos duplicados de un slice.
//
// Usamos [T comparable] porque necesitamos == para detectar
// duplicados. Usamos un map como "set" para rastrear qué
// valores ya hemos visto.
// ------------------------------------------------------------
func Unique[T comparable](slice []T) []T {
	// Map vacío para rastrear elementos vistos.
	// map[T]struct{} es el "set" idiomático de Go:
	// usa struct{} como valor porque ocupa 0 bytes de memoria.
	vistos := make(map[T]struct{})

	var resultado []T

	for _, v := range slice {
		// Si NO lo hemos visto antes, lo agregamos al resultado.
		if _, existe := vistos[v]; !existe {
			vistos[v] = struct{}{} // Marcar como visto
			resultado = append(resultado, v)
		}
	}

	return resultado
}

// ------------------------------------------------------------
// SortBy: Ordena un slice usando una función de comparación personalizada.
//
// NO modifica el slice original (trabaja sobre una copia).
// La función 'less' define el criterio: less(a, b) == true
// significa que 'a' debe ir antes que 'b'.
// ------------------------------------------------------------
func SortBy[T any](slice []T, less func(T, T) bool) []T {
	// Creamos una copia para no modificar el original.
	copia := make([]T, len(slice))
	copy(copia, slice)

	// sort.Slice ordena usando un less function.
	// Es el patrón estándar de Go para ordenación personalizada.
	sort.Slice(copia, func(i, j int) bool {
		return less(copia[i], copia[j])
	})

	return copia
}

// ============================================================
// PARTE 2: TIPO GENÉRICO — STACK (PILA)
// ============================================================
//
// Un Stack es una estructura de datos LIFO (Last In, First Out).
// El último elemento que metes es el primero que sacas.
//
// Analogía: Una pila de platos. Siempre sacas el de arriba.
//
// [T any] significa que el stack puede contener CUALQUIER tipo.
// Stack[int] guarda enteros. Stack[string] guarda strings.
// Mismo código, diferentes tipos.
// ------------------------------------------------------------
type Stack[T any] struct {
	// elementos es un slice que almacena los datos.
	// El tipo T se determina cuando creas el Stack.
	elementos []T
}

// Push agrega un elemento al tope de la pila.
func (s *Stack[T]) Push(valor T) {
	s.elementos = append(s.elementos, valor)
}

// Pop remueve y retorna el elemento del tope.
// Retorna (valor, true) si hay elementos, o (zero_value, false) si está vacía.
func (s *Stack[T]) Pop() (T, bool) {
	// var zero T crea el "zero value" del tipo T.
	// Para int es 0, para string es "", para bool es false.
	// Es la forma genérica de decir "no hay valor".
	var zero T

	if len(s.elementos) == 0 {
		return zero, false // Pila vacía
	}

	// Tomamos el último elemento.
	ultimo := s.elementos[len(s.elementos)-1]

	// Reducimos el slice en 1 (quitamos el último).
	s.elementos = s.elementos[:len(s.elementos)-1]

	return ultimo, true
}

// Peek ve el elemento del tope SIN removerlo.
func (s *Stack[T]) Peek() (T, bool) {
	var zero T

	if len(s.elementos) == 0 {
		return zero, false
	}

	// Solo leemos, NO modificamos el slice.
	return s.elementos[len(s.elementos)-1], true
}

// Len retorna la cantidad de elementos en la pila.
func (s *Stack[T]) Len() int {
	return len(s.elementos)
}

// IsEmpty verifica si la pila está vacía.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.elementos) == 0
}

// String implementa fmt.Stringer para impresión bonita.
func (s *Stack[T]) String() string {
	// Convertimos cada elemento a string con fmt.Sprint.
	parts := make([]string, len(s.elementos))
	for i, v := range s.elementos {
		parts[i] = fmt.Sprint(v)
	}
	return "[" + strings.Join(parts, ", ") + "] (tope →)"
}

// ============================================================
// PARTE 3: VALIDADOR DE STRUCTS CON REFLECT
// ============================================================
//
// Aquí es donde reflect brilla: leer struct tags en runtime
// y validar automáticamente cualquier struct.
//
// Esto es lo que hacen librerías como go-playground/validator
// y encoding/json internamente.
// ------------------------------------------------------------

// ValidationError representa un error de validación encontrado.
type ValidationError struct {
	Field   string // Nombre del campo que falló
	Tag     string // Tag de validación que falló (required, min, max, etc.)
	Message string // Mensaje descriptivo del error
}

// String imprime el error de forma legible.
func (e ValidationError) String() string {
	return fmt.Sprintf("campo '%s': tag '%s' falló → %s", e.Field, e.Tag, e.Message)
}

// Validate inspecciona un struct usando reflect y valida cada campo
// según los tags `validate:"..."`.
//
// Esta función demuestra el poder de reflect: puede trabajar con
// CUALQUIER struct que le pases, sin importar sus campos.
//
// Tags soportados:
//   required  → El campo no puede ser el zero value
//   min=N     → Para strings: longitud mínima N. Para números: valor mínimo N
//   max=N     → Para strings: longitud máxima N. Para números: valor máximo N
//   email     → Debe contener '@' y '.'
//   gt=N      → Greater than: valor debe ser > N
//   lte=N     → Less than or equal: valor debe ser <= N
func Validate(s interface{}) []ValidationError {
	// Slice para acumular errores encontrados.
	var errores []ValidationError

	// reflect.ValueOf obtiene el valor en runtime.
	v := reflect.ValueOf(s)

	// Si es un puntero, desreferenciamos con Elem().
	// Esto es crucial: si pasas &usuario, necesitamos usuario.
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	// Verificamos que sea un struct. Si no lo es, no podemos validar.
	if t.Kind() != reflect.Struct {
		errores = append(errores, ValidationError{
			Field:   "(tipo)",
			Tag:     "struct",
			Message: fmt.Sprintf("se esperaba un struct, se recibió %s", t.Kind()),
		})
		return errores
	}

	// Iteramos sobre TODOS los campos del struct.
	// NumField() retorna cuántos campos tiene el struct.
	for i := 0; i < t.NumField(); i++ {
		// Field(i) retorna la metadata del campo (nombre, tipo, tags).
		campo := t.Field(i)

		// Field(i) sobre el Value retorna el valor actual del campo.
		valor := v.Field(i)

		// Leemos el tag validate de este campo.
		tag := campo.Tag.Get("validate")

		// Si no tiene tag validate, saltamos este campo.
		if tag == "" {
			continue
		}

		// Parseamos las reglas del tag. Ejemplo: "required,min=3,max=50"
		// se convierte en ["required", "min=3", "max=50"]
		reglas := strings.Split(tag, ",")

		// Validamos cada regla contra el valor del campo.
		for _, regla := range reglas {
			errores = append(errores, validarRegla(campo.Name, valor, regla)...)
		}
	}

	return errores
}

// validarRegla aplica UNA regla de validación a un campo específico.
//
// Esta función usa reflect.Value para inspeccionar el tipo y valor
// del campo en runtime. No sabemos si es string, int, bool...
// reflect nos permite manejar todos los tipos uniformemente.
func validarRegla(nombreCampo string, valor reflect.Value, regla string) []ValidationError {
	var errores []ValidationError

	// Separamos la regla de su parámetro. Ejemplo: "min=3" → ["min", "3"]
	partes := strings.SplitN(regla, "=", 2)
	tag := partes[0]
	param := ""
	if len(partes) > 1 {
		param = partes[1]
	}

	switch tag {

	// -------------------------------------------------------
	// REQUIRED: El campo no puede ser el zero value.
	// Para strings: no puede ser ""
	// Para ints: no puede ser 0
	// Para bools: no puede ser false
	// Para slices: no puede ser nil o len 0
	// -------------------------------------------------------
	case "required":
		esVacio := false

		// reflect.Value.IsZero() verifica si es el zero value del tipo.
		// Funciona para CUALQUIER tipo: int, string, bool, struct, slice...
		if valor.IsZero() {
			esVacio = true
		}

		// Para strings, también verificamos que no sea solo espacios.
		if valor.Kind() == reflect.String && strings.TrimSpace(valor.String()) == "" {
			esVacio = true
		}

		if esVacio {
			errores = append(errores, ValidationError{
				Field:   nombreCampo,
				Tag:     "required",
				Message: "el valor no puede estar vacío",
			})
		}

	// -------------------------------------------------------
	// MIN: Para strings → longitud mínima. Para números → valor mínimo.
	// -------------------------------------------------------
	case "min":
		minVal, err := strconv.ParseFloat(param, 64)
		if err != nil {
			break
		}

		switch valor.Kind() {
		case reflect.String:
			// Para strings, comparamos la longitud.
			if float64(len(valor.String())) < minVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "min",
					Message: fmt.Sprintf("longitud %d < mínimo %s", len(valor.String()), param),
				})
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// Para enteros, comparamos el valor directamente.
			if float64(valor.Int()) < minVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "min",
					Message: fmt.Sprintf("%d < mínimo %s", valor.Int(), param),
				})
			}
		case reflect.Float32, reflect.Float64:
			// Para floats, comparamos el valor directamente.
			if valor.Float() < minVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "min",
					Message: fmt.Sprintf("%.2f < mínimo %s", valor.Float(), param),
				})
			}
		}

	// -------------------------------------------------------
	// MAX: Para strings → longitud máxima. Para números → valor máximo.
	// -------------------------------------------------------
	case "max":
		maxVal, err := strconv.ParseFloat(param, 64)
		if err != nil {
			break
		}

		switch valor.Kind() {
		case reflect.String:
			if float64(len(valor.String())) > maxVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "max",
					Message: fmt.Sprintf("longitud %d > máximo %s", len(valor.String()), param),
				})
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if float64(valor.Int()) > maxVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "max",
					Message: fmt.Sprintf("%d > máximo %s", valor.Int(), param),
				})
			}
		case reflect.Float32, reflect.Float64:
			if valor.Float() > maxVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "max",
					Message: fmt.Sprintf("%.2f > máximo %s", valor.Float(), param),
				})
			}
		}

	// -------------------------------------------------------
	// GT: Greater Than (mayor que). Solo para números.
	// -------------------------------------------------------
	case "gt":
		gtVal, err := strconv.ParseFloat(param, 64)
		if err != nil {
			break
		}

		switch valor.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if float64(valor.Int()) <= gtVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "gt",
					Message: fmt.Sprintf("%d no es mayor que %s", valor.Int(), param),
				})
			}
		case reflect.Float32, reflect.Float64:
			if valor.Float() <= gtVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "gt",
					Message: fmt.Sprintf("%.2f no es mayor que %s", valor.Float(), param),
				})
			}
		}

	// -------------------------------------------------------
	// LTE: Less Than or Equal (menor o igual que). Solo para números.
	// -------------------------------------------------------
	case "lte":
		lteVal, err := strconv.ParseFloat(param, 64)
		if err != nil {
			break
		}

		switch valor.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if float64(valor.Int()) > lteVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "lte",
					Message: fmt.Sprintf("%d excede el máximo %s", valor.Int(), param),
				})
			}
		case reflect.Float32, reflect.Float64:
			if valor.Float() > lteVal {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "lte",
					Message: fmt.Sprintf("%.2f excede el máximo %s", valor.Float(), param),
				})
			}
		}

	// -------------------------------------------------------
	// EMAIL: Verificación simple de que parece un email.
	// No es un regex perfecto, pero cubre el 99% de casos.
	// -------------------------------------------------------
	case "email":
		if valor.Kind() == reflect.String {
			email := valor.String()
			if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
				errores = append(errores, ValidationError{
					Field:   nombreCampo,
					Tag:     "email",
					Message: fmt.Sprintf("'%s' no parece un email válido", email),
				})
			}
		}
	}

	return errores
}

// ============================================================
// PARTE 4: FUNCIÓN AUXILIAR — StructAMap
// ============================================================
//
// Convierte cualquier struct a un map[string]interface{} usando
// los tags json como keys. Esto demuestra otra aplicación
// poderosa de reflect: serialización personalizada.
// ------------------------------------------------------------
func StructAMap(obj interface{}) map[string]interface{} {
	resultado := make(map[string]interface{})

	v := reflect.ValueOf(obj)
	t := v.Type()

	// Desreferenciar puntero si es necesario
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	// Iterar sobre cada campo del struct
	for i := 0; i < t.NumField(); i++ {
		campo := t.Field(i)
		valor := v.Field(i)

		// Usar el tag json como key. Si no tiene, usar el nombre en minúsculas.
		key := campo.Tag.Get("json")
		if key == "" || key == "-" {
			key = strings.ToLower(campo.Name)
		}

		// Ignorar campos con omitempty si son zero value
		if strings.Contains(key, ",omitempty") {
			key = strings.Split(key, ",")[0]
			if valor.IsZero() {
				continue
			}
		}

		// Interface() convierte reflect.Value de vuelta a interface{}
		// para poder guardarlo en el map.
		resultado[key] = valor.Interface()
	}

	return resultado
}

// ============================================================
// PARTE 5: DEMOSTRACIÓN DE REFLECT — Introspección de Tipos
// ============================================================
//
// Función que demuestra cómo inspect funciona en runtime.
// ------------------------------------------------------------
func demoReflectInspeccion(valor interface{}, nombre string) {
	fmt.Printf("   🔍 Inspeccionando: %s = %v\n", nombre, valor)

	// reflect.TypeOf: ¿QUÉ tipo es?
	t := reflect.TypeOf(valor)
	fmt.Printf("      TypeOf  → %v\n", t)

	// reflect.ValueOf: ¿CUÁNTO vale?
	v := reflect.ValueOf(valor)
	fmt.Printf("      ValueOf → %v\n", v)

	// Kind: ¿CATEGORÍA del tipo?
	// Kind() agrupa: int, int8, int16... todos son reflect.Int, reflect.Int8, etc.
	fmt.Printf("      Kind    → %v\n", v.Kind())

	// ¿Es addressable? (puedo modificarlo con Set?)
	fmt.Printf("      Addr.   → %v\n", v.CanAddr())

	fmt.Println()
}

// ============================================================
// MAIN — El punto de entrada que ejecuta todas las demos
// ============================================================
func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🔮 LECCIÓN 19: GENERICS, REFLECT Y METAPROGRAMACIÓN")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// ========================================================
	// DEMO 1: Map — Transformar cada elemento
	// ========================================================
	fmt.Println("🧩 === GENERICS: Colecciones Funcionales ===")
	fmt.Println()

	fmt.Println("📊 Map: Convertir enteros a strings")
	nums := []int{1, 2, 3, 4, 5}
	// strconv.Itoa convierte int → string.
	// Map aplica esa función a cada elemento del slice.
	comoStrings := Map(nums, strconv.Itoa)
	fmt.Printf("   %v → %v\n", nums, comoStrings)
	fmt.Println()

	fmt.Println("📊 Map: Doblar cada número")
	dobles := Map(nums, func(n int) int { return n * 2 })
	fmt.Printf("   %v → %v\n", nums, dobles)
	fmt.Println()

	fmt.Println("📊 Map: Extraer longitud de cada string")
	palabras := []string{"Go", "Python", "Rust", "Java"}
	longitudes := Map(palabras, func(s string) int { return len(s) })
	fmt.Printf("   %v → %v\n", palabras, longitudes)
	fmt.Println()

	// ========================================================
	// DEMO 2: Filter — Seleccionar elementos
	// ========================================================
	fmt.Println("🔍 Filter: Números pares")
	grandes := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	pares := Filter(grandes, func(n int) bool { return n%2 == 0 })
	fmt.Printf("   %v → %v\n", grandes, pares)
	fmt.Println()

	fmt.Println("🔍 Filter: Palabras con más de 3 caracteres")
	largas := Filter(palabras, func(s string) bool { return len(s) > 3 })
	fmt.Printf("   %v → %v\n", palabras, largas)
	fmt.Println()

	// ========================================================
	// DEMO 3: Reduce — Combinar en un solo valor
	// ========================================================
	fmt.Println("📦 Reduce: Sumar todos los números")
	suma := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("   %v → %d\n", nums, suma)
	fmt.Println()

	fmt.Println("📦 Reduce: Concatenar strings con espacio")
	oracion := Reduce(palabras, "", func(acc, s string) string {
		if acc == "" {
			return s
		}
		return acc + " " + s
	})
	fmt.Printf("   %v → \"%s\"\n", palabras, oracion)
	fmt.Println()

	fmt.Println("📦 Reduce: Encontrar el número máximo")
	max := Reduce(nums[1:], nums[0], func(acc, n int) int {
		if n > acc {
			return n
		}
		return acc
	})
	fmt.Printf("   %v → %d\n", nums, max)
	fmt.Println()

	// ========================================================
	// DEMO 4: Contains — ¿Existe el elemento?
	// ========================================================
	fmt.Println("🎯 Contains: ¿Existe el 3?")
	fmt.Printf("   %v contiene 3? %v\n", nums, Contains(nums, 3))
	fmt.Printf("   %v contiene 99? %v\n", nums, Contains(nums, 99))
	fmt.Println()

	fmt.Println("🎯 Contains: ¿Existe \"Rust\"?")
	fmt.Printf("   %v contiene \"Rust\"? %v\n", palabras, Contains(palabras, "Rust"))
	fmt.Printf("   %v contiene \"C++\"? %v\n", palabras, Contains(palabras, "C++"))
	fmt.Println()

	// ========================================================
	// DEMO 5: Unique — Eliminar duplicados
	// ========================================================
	fmt.Println("✨ Unique: Eliminar duplicados de enteros")
	conDuplicados := []int{1, 2, 2, 3, 3, 3, 4, 4, 4, 4}
	sinDuplicados := Unique(conDuplicados)
	fmt.Printf("   %v → %v\n", conDuplicados, sinDuplicados)
	fmt.Println()

	fmt.Println("✨ Unique: Eliminar duplicados de strings")
	palabrasDup := []string{"Go", "Python", "Go", "Rust", "Python", "Go"}
	fmt.Printf("   %v → %v\n", palabrasDup, Unique(palabrasDup))
	fmt.Println()

	// ========================================================
	// DEMO 6: SortBy — Ordenar con criterio personalizado
	// ========================================================
	fmt.Println("📐 SortBy: Ordenar strings por longitud (corto → largo)")
	ordenPorLen := SortBy(palabras, func(a, b string) bool {
		return len(a) < len(b)
	})
	fmt.Printf("   %v → %v\n", palabras, ordenPorLen)
	fmt.Println()

	fmt.Println("📐 SortBy: Ordenar enteros descendente")
	ordenDesc := SortBy(nums, func(a, b int) bool {
		return a > b // Nota: invertimos para descendente
	})
	fmt.Printf("   %v → %v\n", nums, ordenDesc)
	fmt.Println()

	// ========================================================
	// DEMO 7: Stack Genérico — Estructura de datos genérica
	// ========================================================
	fmt.Println("📚 === STACK GENÉRICO ===")
	fmt.Println()

	// Stack de enteros
	fmt.Println("📚 Stack[int]: Push 10, 20, 30")
	stackInt := &Stack[int]{}
	stackInt.Push(10)
	stackInt.Push(20)
	stackInt.Push(30)
	fmt.Printf("   Estado: %s\n", stackInt)
	fmt.Printf("   Tamaño: %d\n", stackInt.Len())

	top, ok := stackInt.Peek()
	fmt.Printf("   Peek: %d (existe: %v)\n", top, ok)

	for !stackInt.IsEmpty() {
		val, _ := stackInt.Pop()
		fmt.Printf("   Pop: %d → Estado: %s\n", val, stackInt)
	}
	fmt.Println()

	// Stack de strings
	fmt.Println("📚 Stack[string]: Push \"Go\", \"Python\", \"Rust\"")
	stackStr := &Stack[string]{}
	stackStr.Push("Go")
	stackStr.Push("Python")
	stackStr.Push("Rust")
	fmt.Printf("   Estado: %s\n", stackStr)

	for !stackStr.IsEmpty() {
		val, _ := stackStr.Pop()
		fmt.Printf("   Pop: \"%s\"\n", val)
	}
	fmt.Println()

	// Stack de floats (demostrando que el mismo código funciona con cualquier tipo)
	fmt.Println("📚 Stack[float64]: Push 3.14, 2.71, 1.41")
	stackFloat := &Stack[float64]{}
	stackFloat.Push(3.14)
	stackFloat.Push(2.71)
	stackFloat.Push(1.41)
	fmt.Printf("   Estado: %s\n", stackFloat)
	fmt.Println()

	// ========================================================
	// DEMO 8: Reflect — Inspección de tipos
	// ========================================================
	fmt.Println("🔮 === REFLECT: Inspección de Tipos ===")
	fmt.Println()

	demoReflectInspeccion(42, "entero")
	demoReflectInspeccion(3.14, "float")
	demoReflectInspeccion("Go", "string")
	demoReflectInspeccion(true, "bool")
	demoReflectInspeccion([]int{1, 2, 3}, "slice")

	// ========================================================
	// DEMO 9: Reflect — Validación de Structs
	// ========================================================
	fmt.Println("🔮 === REFLECT: Validador de Structs ===")
	fmt.Println()

	// Struct con tags de validación
	type Usuario struct {
		Nombre string `json:"nombre" validate:"required,min=2,max=50"`
		Email  string `json:"email"  validate:"required,email"`
		Edad   int    `json:"edad"   validate:"min=0,max=150"`
		Activo bool   `json:"activo"`
	}

	// --- CASO 1: Usuario válido ---
	fmt.Println("✅ Usuario válido:")
	valido := Usuario{
		Nombre: "Ana García",
		Email:  "ana@mail.com",
		Edad:   25,
		Activo: true,
	}
	fmt.Printf("   %+v\n", valido)
	errores := Validate(valido)
	if len(errores) == 0 {
		fmt.Println("   ✅ Validación: SIN ERRORES")
	} else {
		for _, e := range errores {
			fmt.Printf("   ❌ %s\n", e)
		}
	}
	fmt.Println()

	// --- CASO 2: Usuario inválido (múltiples errores) ---
	fmt.Println("❌ Usuario inválido:")
	invalido := Usuario{
		Nombre: "",         // Falla: required
		Email:  "bad-email", // Falla: email
		Edad:   -5,          // Falla: min=0
		Activo: false,       // No tiene tag validate, no falla
	}
	fmt.Printf("   %+v\n", invalido)
	errores = Validate(invalido)
	if len(errores) == 0 {
		fmt.Println("   ✅ Validación: SIN ERRORES")
	} else {
		fmt.Printf("   ❌ %d errores encontrados:\n", len(errores))
		for _, e := range errores {
			fmt.Printf("      • %s\n", e)
		}
	}
	fmt.Println()

	// --- CASO 3: Otro struct diferente (misma función Validate!) ---
	fmt.Println("✅ Producto (otro struct, misma función Validate):")
	type Producto struct {
		Nombre  string  `json:"nombre"  validate:"required,min=2"`
		Precio  float64 `json:"precio"  validate:"gt=0"`
		Stock   int     `json:"stock"   validate:"min=0,lte=10000"`
	}

	producto := Producto{
		Nombre: "Laptop",
		Precio: 999.99,
		Stock:  50,
	}
	fmt.Printf("   %+v\n", producto)
	errores = Validate(producto)
	if len(errores) == 0 {
		fmt.Println("   ✅ Validación: SIN ERRORES")
	} else {
		for _, e := range errores {
			fmt.Printf("   ❌ %s\n", e)
		}
	}
	fmt.Println()

	// --- CASO 4: Producto inválido ---
	fmt.Println("❌ Producto inválido:")
	productoMalo := Producto{
		Nombre: "",      // Falla: required
		Precio: -10.5,   // Falla: gt=0
		Stock:  99999,   // Falla: lte=10000
	}
	fmt.Printf("   %+v\n", productoMalo)
	errores = Validate(productoMalo)
	if len(errores) == 0 {
		fmt.Println("   ✅ Validación: SIN ERRORES")
	} else {
		fmt.Printf("   ❌ %d errores encontrados:\n", len(errores))
		for _, e := range errores {
			fmt.Printf("      • %s\n", e)
		}
	}
	fmt.Println()

	// ========================================================
	// DEMO 10: Reflect — Struct a Map
	// ========================================================
	fmt.Println("🗺️  === STRUCT A MAP (con Reflect) ===")
	fmt.Println()

	fmt.Println("🗺️  Convirtiendo Usuario a map[string]interface{}:")
	mapa := StructAMap(valido)
	for k, v := range mapa {
		fmt.Printf("   %s → %v (%s)\n", k, v, reflect.TypeOf(v))
	}
	fmt.Println()

	fmt.Println("🗺️  Convirtiendo Producto a map[string]interface{}:")
	mapaProd := StructAMap(producto)
	for k, v := range mapaProd {
		fmt.Printf("   %s → %v (%s)\n", k, v, reflect.TypeOf(v))
	}
	fmt.Println()

	// ========================================================
	// DEMO 11: Combinando Generics + Reflect
	// ========================================================
	fmt.Println("🧩🔮 === GENERICS + REFLECT COMBINADOS ===")
	fmt.Println()

	// Usamos Map genérico para transformar los errores de validación en strings.
	erroresStrings := Map(errores, func(e ValidationError) string {
		return e.String()
	})
	fmt.Println("📝 Errores del producto inválido (formateados con Map genérico):")
	for _, s := range erroresStrings {
		fmt.Printf("   • %s\n", s)
	}
	fmt.Println()

	// Usamos Filter genérico para filtrar errores de un tag específico.
	erroresDeMin := Filter(errores, func(e ValidationError) bool {
		return e.Tag == "min" || e.Tag == "gt" || e.Tag == "lte"
	})
	fmt.Printf("🔍 Solo errores numéricos (min/gt/lte): %d encontrados\n", len(erroresDeMin))
	for _, e := range erroresDeMin {
		fmt.Printf("   • %s\n", e)
	}
	fmt.Println()

	// ========================================================
	// RESUMEN FINAL
	// ========================================================
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("📋 RESUMEN DE LA LECCIÓN")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("🧩 GENERICS aprendido:")
	fmt.Println("   • Type parameters: [T any], [T comparable]")
	fmt.Println("   • Constraints personalizados: interface { ~int | ~float64 }")
	fmt.Println("   • Funciones genéricas: Map, Filter, Reduce, Contains, Unique")
	fmt.Println("   • Tipos genéricos: Stack[T] con métodos genéricos")
	fmt.Println("   • Múltiples type params: Map[T any, U any]")
	fmt.Println()
	fmt.Println("🔮 REFLECT aprendido:")
	fmt.Println("   • reflect.TypeOf() → inspeccionar tipos en runtime")
	fmt.Println("   • reflect.ValueOf() → acceder a valores en runtime")
	fmt.Println("   • .Kind(), .NumField(), .Field(i) → navegar structs")
	fmt.Println("   • .Tag.Get(\"json\") → leer struct tags")
	fmt.Println("   • Validate() → validador automático de structs")
	fmt.Println("   • StructAMap() → conversión struct → map")
	fmt.Println()
	fmt.Println("⚖️ REGLA DE ORO:")
	fmt.Println("   🧩 Generics → cuando conoces los tipos al compilar")
	fmt.Println("   🔮 Reflect  → cuando solo los conoces en runtime")
	fmt.Println("   Si puedes usar generics → USA GENERICS.")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
}