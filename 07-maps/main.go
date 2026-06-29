package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// ═══════════════════════════════════════════════════════════════
//  🧪 Laboratorio de Go — Lección 07
//  Contador de Frecuencia de Palabras en Texto Real
//  Maps · Tablas Hash · Comma Ok · Iteración · Ordenamiento
// ═══════════════════════════════════════════════════════════════

// ─────────────────────────────────────────────────────────────
//  Sección 1: Creación y operaciones básicas con maps
// ─────────────────────────────────────────────────────────────

func demostrarCreacionMaps() {
	fmt.Println("   📌 Creación y operaciones básicas con maps")
	fmt.Println()

	// Forma 1: Literal — la más común y legible
	edades := map[string]int{
		"Alice":   30,
		"Bob":     25,
		"Charlie": 35,
	}
	fmt.Printf("   Map literal: %v\n", edades)
	fmt.Printf("   Longitud (pares clave-valor): %d\n", len(edades))
	fmt.Println()

	// Forma 2: make — útil cuando los datos se agregan dinámicamente
	ciudades := make(map[string]string)
	ciudades["PE"] = "Lima"
	ciudades["MX"] = "Ciudad de México"
	ciudades["AR"] = "Buenos Aires"
	ciudades["CO"] = "Bogotá"
	fmt.Printf("   Map con make: %v\n", ciudades)
	fmt.Println()

	// Acceder a valores
	fmt.Println("   🔍 Acceder a valores:")
	fmt.Printf("     edades[\"Alice\"]   = %d\n", edades["Alice"])
	fmt.Printf("     edades[\"Bob\"]     = %d\n", edades["Bob"])
	fmt.Println()

	// ⚠️ El zero value: acceder a una clave que NO existe
	fmt.Println("   ⚠️  Acceder a clave inexistente (el zero value):")
	fmt.Printf("     edades[\"Zoe\"]     = %d  ← ¡No existe, pero NO da error!\n", edades["Zoe"])
	fmt.Printf("     edades[\"Zoe\"]     = %d  ← Devuelve el zero value del tipo (0 para int)\n", edades["Zoe"])
	fmt.Println()

	// El comma ok idiom: distinguir "no existe" de "valor es cero"
	fmt.Println("   🔑 El comma ok idiom (ESCRITURA ESENCIAL):")
	edad, existe := edades["Zoe"]
	if existe {
		fmt.Printf("     Zoe tiene %d años\n", edad)
	} else {
		fmt.Println("     Zoe no está en el mapa ← ¡CORRECTO!")
	}

	edad, existe = edades["Alice"]
	if existe {
		fmt.Printf("     Alice tiene %d años ← ¡EXISTE!\n", edad)
	} else {
		fmt.Println("     Alice no está en el mapa")
	}
	fmt.Println()

	// Agregar y modificar
	fmt.Println("   ✏️  Agregar y modificar:")
	edades["Diana"] = 28                         // Agregar nueva clave
	edades["Alice"] = 31                         // Modificar existente
	fmt.Printf("     Después de agregar Diana y modificar Alice: %v\n", edades)
	fmt.Println()

	// Eliminar
	fmt.Println("   🗑️  Eliminar con delete():")
	delete(edades, "Bob")
	fmt.Printf("     Después de delete(edades, \"Bob\"): %v\n", edades)

	// Eliminar clave inexistente NO causa panic (es seguro)
	delete(edades, "Inexistente")
	fmt.Println("     delete de clave inexistente: NO causa panic (seguro)")
}

// ─────────────────────────────────────────────────────────────
//  Sección 2: Iteración y ordenamiento de maps
// ─────────────────────────────────────────────────────────────

func demostrarIteracion() {
	fmt.Println("   📌 Iteración y ordenamiento de maps")
	fmt.Println()

	notas := map[string]float64{
		"Matemáticas": 9.2,
		"Física":      8.5,
		"Química":     7.8,
		"Historia":    9.5,
		"Biología":    8.9,
		"Programación": 10.0,
	}

	// Iterar con range: el orden es ALEATORIO
	fmt.Println("   🔁 Iterar con range (orden NO garantizado):")
	for materia, nota := range notas {
		fmt.Printf("     %-15s → %.1f\n", materia, nota)
	}
	fmt.Println()

	// ⚠️ Go NO garantiza el orden de iteración de maps
	fmt.Println("   ⚠️  Go NO garantiza el orden de iteración.")
	fmt.Println("      Ejecuta este programa varias veces y verás que el orden cambia.")
	fmt.Println("      Esto es INTENCIONAL para evitar que dependas del orden.")
	fmt.Println()

	// ✅ Cómo ordenar un map: obtener las keys, ordenarlas, iterar en orden
	fmt.Println("   ✅ Ordenar un map por claves:")
	keys := make([]string, 0, len(notas)) // Pre-allocar con capacidad exacta
	for k := range notas {
		keys = append(keys, k)
	}
	sort.Strings(keys) // Ordenar alfabéticamente

	for _, k := range keys {
		fmt.Printf("     %-15s → %.1f\n", k, notas[k])
	}
	fmt.Println()

	// Ordenar por valores (más complejo)
	fmt.Println("   ✅ Ordenar un map por valores (descendente):")
	type kv struct {
		Key   string
		Value float64
	}
	var pares []kv
	for k, v := range notas {
		pares = append(pares, kv{k, v})
	}
	sort.Slice(pares, func(i, j int) bool {
		return pares[i].Value > pares[j].Value // Descendente
	})
	for _, p := range pares {
		fmt.Printf("     %-15s → %.1f\n", p.Key, p.Value)
	}
	fmt.Println()

	// Solo claves y solo valores
	fmt.Println("   🔍 Solo claves y solo valores:")
	fmt.Printf("     Claves:  %v\n", keys)
	valores := make([]float64, 0, len(notas))
	for _, v := range notas {
		valores = append(valores, v)
	}
	fmt.Printf("     Valores: %v\n", valores)
}

// ─────────────────────────────────────────────────────────────
//  Sección 3: El zero value, nil maps y maps como conjuntos
// ─────────────────────────────────────────────────────────────

func demostrarZeroValueYConjuntos() {
	fmt.Println("   📌 Zero value, nil maps y maps como conjuntos (sets)")
	fmt.Println()

	// Un map nil es LEYBLE pero no escribible
	fmt.Println("   ⚠️  Un map nil (var m map[string]int):")
	var mapaNil map[string]int
	fmt.Printf("     mapaNil == nil: %t\n", mapaNil == nil)
	fmt.Printf("     len(mapaNil):   %d\n", len(mapaNil))
	fmt.Printf("     mapaNil[\"x\"]:   %d  ← Lectura OK (devuelve zero value)\n", mapaNil["x"])

	// ⚠️ Escribir en un map nil causa PANIC
	// Descomenta la siguiente línea para ver el panic:
	// mapaNil["x"] = 1  ← PANIC: assignment to entry in nil map
	fmt.Println("     mapaNil[\"x\"] = 1 → PANIC ← ¡NUNCA escribas en un map nil!")
	fmt.Println()

	// Mapa como conjunto (SET): solo nos importan las claves
	fmt.Println("   🎯 Maps como conjuntos (SET): solo importan las claves")
	visitados := make(map[string]bool) // valor bool, no nos importa el valor

	// Agregar elementos al "set"
	for _, url := range []string{
		"/api/users",
		"/api/posts",
		"/api/users",   // duplicado
		"/api/comments",
		"/api/posts",   // duplicado
		"/api/health",
	} {
		visitados[url] = true
	}

	fmt.Printf("     URLs únicas visitadas: %d (de 6 intentos)\n", len(visitados))
	for url := range visitados {
		fmt.Printf("       → %s\n", url)
	}
	fmt.Println()

	// Verificar pertenencia al set
	fmt.Println("   🔍 Verificar pertenencia (comma ok):")
	for _, url := range []string{"/api/users", "/api/secret"} {
		if visitados[url] {
			fmt.Printf("     %s → ✅ visitado\n", url)
		} else {
			fmt.Printf("     %s → ❌ NO visitado\n", url)
		}
	}
	fmt.Println()

	// Map[string]struct{}: la forma más eficiente de un set
	// struct{} ocupa 0 bytes en memoria (vs bool que ocupa 1 byte)
	fmt.Println("   💡 Map[string]struct{}: set ultra-eficiente (0 bytes por valor)")
	eficiente := make(map[string]struct{})
	eficiente["go"] = struct{}{}
	eficiente["rust"] = struct{}{}
	eficiente["python"] = struct{}{}
	fmt.Printf("     Lenguajes en el set: %d elementos\n", len(eficiente))
	if _, ok := eficiente["go"]; ok {
		fmt.Println("     \"go\" está en el set ✅")
	}
}

// ─────────────────────────────────────────────────────────────
//  Sección 4: Maps anidados y map[string][]string
// ─────────────────────────────────────────────────────────────

func demostrarMapsAnidados() {
	fmt.Println("   📌 Maps anidados y map[string][]T (agrupación)")
	fmt.Println()

	// Map anidado: inventario por tienda y producto
	inventario := map[string]map[string]int{
		"Tienda Centro": {
			"Laptop":  15,
			"Mouse":   200,
			"Teclado": 80,
		},
		"Tienda Norte": {
			"Laptop":  8,
			"Monitor": 50,
			"Mouse":   150,
		},
		"Tienda Sur": {
			"Laptop":  22,
			"Tablet":  100,
			"Teclado": 60,
			"Mouse":   300,
		},
	}

	fmt.Println("   📦 Inventario por tienda:")
	tiendas := make([]string, 0, len(inventario))
	for t := range inventario {
		tiendas = append(tiendas, t)
	}
	sort.Strings(tiendas)

	for _, tienda := range tiendas {
		productos := inventario[tienda]
		total := 0
		for _, cant := range productos {
			total += cant
		}
		fmt.Printf("     %-16s → %d productos (total: %d unidades)\n", tienda, len(productos), total)
		for prod, cant := range productos {
			fmt.Printf("       %-14s → %d unidades\n", prod, cant)
		}
	}
	fmt.Println()

	// Map[string][]string: agrupación (patrón más usado en Go)
	fmt.Println("   🗂️  map[string][]T: agrupación dinámica (patrón estrella)")
	fmt.Println()

	// Simular un sistema de archivos: directorio → archivos
	archivos := []string{
		"src/main.go",
		"src/utils.go",
		"src/api/handler.go",
		"src/api/middleware.go",
		"tests/main_test.go",
		"tests/api_test.go",
		"docs/README.md",
		"docs/CHANGELOG.md",
	}

	// Agrupar archivos por directorio usando el patrón map[string][]string
	porDirectorio := make(map[string][]string)
	for _, archivo := range archivos {
		// Extraer el directorio (todo antes del último /)
		dir := archivo[:strings.LastIndex(archivo, "/")]
		porDirectorio[dir] = append(porDirectorio[dir], archivo)
	}

	fmt.Println("   Archivos agrupados por directorio:")
	dirs := make([]string, 0, len(porDirectorio))
	for d := range porDirectorio {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	for _, dir := range dirs {
		files := porDirectorio[dir]
		fmt.Printf("     📁 %s/ (%d archivos)\n", dir, len(files))
		for _, f := range files {
			nombre := f[strings.LastIndex(f, "/")+1:]
			fmt.Printf("       📄 %s\n", nombre)
		}
	}
}

// ─────────────────────────────────────────────────────────────
//  Sección 5: Contador de frecuencia de palabras (EJERCICIO)
// ─────────────────────────────────────────────────────────────

// stopwords es un SET de palabras comunes que ignoramos en el análisis
var stopwords = map[string]struct{}{
	// Español
	"de": {}, "la": {}, "el": {}, "en": {}, "y": {}, "a": {},
	"que": {}, "los": {}, "del": {}, "las": {}, "un": {}, "una": {},
	"por": {}, "con": {}, "para": {}, "se": {}, "su": {},
	"al": {}, "lo": {}, "como": {}, "o": {}, "pero": {},
	"es": {}, "son": {}, "no": {}, "ha": {}, "le": {},
	// Inglés
	"the": {}, "an": {}, "of": {}, "and": {}, "to": {},
	"in": {}, "is": {}, "it": {}, "for": {}, "on": {},
	"with": {}, "that": {}, "this": {}, "are": {}, "was": {},
	"be": {}, "at": {}, "by": {}, "from": {}, "or": {},
	"as": {}, "into": {}, "has": {}, "its": {}, "can": {},
	"will": {}, "do": {}, "not": {}, "we": {}, "he": {},
	"she": {}, "they": {}, "you": {}, "i": {}, "my": {},
	"our": {}, "his": {}, "her": {}, "their": {}, "your": {},
	"me": {}, "him": {}, "us": {}, "them": {}, "so": {},
	"if": {}, "all": {}, "would": {}, "which": {}, "been": {},
	"have": {}, "had": {}, "were": {},
}

// PalabraFrecuencia representa una palabra con su conteo
type PalabraFrecuencia struct {
	Palabra    string
	Frecuencia int
	Porcentaje float64
}

// limpiarTexto normaliza el texto: minúsculas, solo letras y espacios
func limpiarTexto(texto string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(texto) {
		if unicode.IsLetter(r) || unicode.IsSpace(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ') // Reemplazar puntuación con espacio
		}
	}
	return sb.String()
}

// contarPalabras cuenta la frecuencia de cada palabra ignorando stopwords
func contarPalabras(texto string) map[string]int {
	textoLimpio := limpiarTexto(texto)
	palabras := strings.Fields(textoLimpio) // Split por whitespace

	frecuencia := make(map[string]int)
	for _, palabra := range palabras {
		// Ignorar stopwords
		if _, esStopword := stopwords[palabra]; esStopword {
			continue
		}
		// Ignorar palabras muy cortas (1-2 caracteres)
		if len(palabra) < 3 {
			continue
		}
		frecuencia[palabra]++
	}

	return frecuencia
}

// topN devuelve las N palabras más frecuentes ordenadas descendente
func topN(frecuencia map[string]int, n int) []PalabraFrecuencia {
	// Calcular el total de palabras (para porcentajes)
	total := 0
	for _, count := range frecuencia {
		total += count
	}

	// Convertir el map a un slice para ordenar
	var palabras []PalabraFrecuencia
	for palabra, count := range frecuencia {
		pct := float64(count) / float64(total) * 100
		palabras = append(palabras, PalabraFrecuencia{
			Palabra:    palabra,
			Frecuencia: count,
			Porcentaje: pct,
		})
	}

	// Ordenar por frecuencia descendente, luego alfabético
	sort.Slice(palabras, func(i, j int) bool {
		if palabras[i].Frecuencia == palabras[j].Frecuencia {
			return palabras[i].Palabra < palabras[j].Palabra
		}
		return palabras[i].Frecuencia > palabras[j].Frecuencia
	})

	// Limitar a N resultados
	if n > len(palabras) {
		n = len(palabras)
	}
	return palabras[:n]
}

// mostrarBarra crea una barra visual de frecuencia
func mostrarBarra(valor, maximo int, ancho int) string {
	if maximo == 0 {
		return ""
	}
	barras := int(float64(valor) / float64(maximo) * float64(ancho))
	return strings.Repeat("█", barras) + strings.Repeat("░", ancho-barras)
}

// buscarPalabra busca una palabra específica en el mapa de frecuencias
func buscarPalabra(frecuencia map[string]int, palabra string) {
	palabra = strings.ToLower(palabra)
	count, existe := frecuencia[palabra]
	if existe {
		fmt.Printf("     🔎 \"%s\" aparece %d veces en el texto\n", palabra, count)
	} else {
		fmt.Printf("     🔎 \"%s\" NO aparece en el texto\n", palabra)
	}
}

// analizarBigrams cuenta pares de palabras consecutivas
func analizarBigrams(texto string) map[string]int {
	textoLimpio := limpiarTexto(texto)
	palabras := strings.Fields(textoLimpio)

	bigrams := make(map[string]int)
	for i := 0; i < len(palabras)-1; i++ {
		// Solo incluir bigrams donde ninguna palabra es stopword
		if _, sw := stopwords[palabras[i]]; sw {
			continue
		}
		if _, sw := stopwords[palabras[i+1]]; sw {
			continue
		}
		if len(palabras[i]) < 3 || len(palabras[i+1]) < 3 {
			continue
		}
		bigram := palabras[i] + " " + palabras[i+1]
		bigrams[bigram]++
	}
	return bigrams
}

// estadisticasMap calcula estadísticas sobre el mapa de frecuencias
func estadisticasMap(frecuencia map[string]int) (totalPalabras, palabrasUnicas, hapaxLegomena int) {
	totalPalabras = 0
	hapaxLegomena = 0 // Palabras que aparecen solo una vez

	for _, count := range frecuencia {
		totalPalabras += count
		if count == 1 {
			hapaxLegomena++
		}
	}
	palabrasUnicas = len(frecuencia)
	return
}

// ─────────────────────────────────────────────────────────────
//  FUNCIÓN PRINCIPAL
// ─────────────────────────────────────────────────────────────

func main() {
	// ── Banner ──
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   🧪 LABORATORIO DE GO — LECCIÓN 07                          ║")
	fmt.Println("║   Maps: Diccionarios, Tablas Hash y el Arte del Acceso O(1)   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ── 1. CREACIÓN Y OPERACIONES BÁSICAS ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   1️⃣  Creación y operaciones básicas con maps")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarCreacionMaps()

	// ── 2. ITERACIÓN Y ORDENAMIENTO ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   2️⃣  Iteración y ordenamiento de maps")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarIteracion()

	// ── 3. ZERO VALUE, NIL MAPS Y SETS ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   3️⃣  Zero value, nil maps y maps como conjuntos (sets)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarZeroValueYConjuntos()

	// ── 4. MAPS ANIDADOS Y AGRUPACIÓN ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   4️⃣  Maps anidados y map[string][]T (agrupación)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	demostrarMapsAnidados()

	// ── 5. CONTADOR DE FRECUENCIA DE PALABRAS ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   5️⃣  Contador de frecuencia de palabras (EJERCICIO PRINCIPAL)")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Texto de ejemplo: fragmento de un discurso sobre tecnología
	texto := `La programación en Go ha revolucionado el mundo del desarrollo de software.
	Go fue creado en Google por Robert Griesemer, Rob Pike y Ken Thompson con el objetivo
	de resolver los problemas de compilación lenta y complejidad de los lenguajes existentes.
	La filosofía de Go se basa en la simplicidad: menos es más, y la claridad del código es
	más importante que la cleverness del programador. Go ofrece concurrencia nativa a través
	de goroutines y channels, lo que permite construir sistemas de alta disponibilidad y
	escalabilidad. Los desarrolladores de Go valoran el código limpio, la documentación clara
	y los tests automatizados. En la industria, Go es utilizado por empresas como Google,
	Docker, Kubernetes, Uber, Twitch y Cloudflare para construir servicios de infraestructura
	crítica. La comunidad de Go crece cada día porque el lenguaje prioriza la productividad
	del desarrollador sin sacrificar el rendimiento del sistema. Go compila a binarios nativos
	sin dependencias externas, lo que facilita el despliegue en contenedores y sistemas
	distribuidos. El futuro de Go es brillante porque resuelve problemas reales de la
	ingeniería de software moderna.`

	fmt.Println("   📄 Texto analizado (fragmento sobre tecnología y Go):")
	// Mostrar texto truncado
	lineas := strings.Split(texto, "\n")
	for i, linea := range lineas {
		trimmed := strings.TrimSpace(linea)
		if len(trimmed) > 70 {
			trimmed = trimmed[:67] + "..."
		}
		if i < 4 {
			fmt.Printf("     \"%s\"\n", trimmed)
		}
	}
	fmt.Println("     ...")
	fmt.Println()

	// Contar palabras
	frecuencia := contarPalabras(texto)
	totalPalabras, palabrasUnicas, hapax := estadisticasMap(frecuencia)

	fmt.Println("   📊 Estadísticas generales:")
	fmt.Printf("     Total de palabras (sin stopwords): %d\n", totalPalabras)
	fmt.Printf("     Palabras únicas:                   %d\n", palabrasUnicas)
	fmt.Printf("     Hapax legómena (aparecen 1 vez):   %d\n", hapax)
	fmt.Printf("     Diversidad léxica:                 %.1f%%\n",
		float64(palabrasUnicas)/float64(totalPalabras)*100)
	fmt.Println()

	// Top 15 palabras
	top := topN(frecuencia, 15)
	maxFreq := top[0].Frecuencia

	fmt.Println("   🏆 Top 15 palabras más frecuentes:")
	fmt.Printf("     %-4s %-18s %6s %8s  %s\n", "#", "PALABRA", "FREQ", "%", "DISTRIBUCIÓN")
	fmt.Println("     " + strings.Repeat("─", 70))

	for i, pf := range top {
		barra := mostrarBarra(pf.Frecuencia, maxFreq, 20)
		fmt.Printf("     %-4d %-18s %6d %7.1f%%  %s\n",
			i+1, pf.Palabra, pf.Frecuencia, pf.Porcentaje, barra)
	}
	fmt.Println()

	// Buscar palabras específicas usando comma ok
	fmt.Println("   🔎 Búsquedas individuales (comma ok idiom):")
	buscarPalabra(frecuencia, "go")
	buscarPalabra(frecuencia, "google")
	buscarPalabra(frecuencia, "python")
	buscarPalabra(frecuencia, "goroutines")
	buscarPalabra(frecuencia, "rust")
	fmt.Println()

	// Bigrams (pares de palabras frecuentes)
	fmt.Println("   🔗 Top 10 bigrams (pares de palabras consecutivas):")
	bigrams := analizarBigrams(texto)
	topBigrams := topN(bigrams, 10)
	if len(topBigrams) > 0 {
		maxBigramFreq := topBigrams[0].Frecuencia
		for i, bg := range topBigrams {
			barra := mostrarBarra(bg.Frecuencia, maxBigramFreq, 15)
			fmt.Printf("     %-4d %-25s %3d  %s\n",
				i+1, "\""+bg.Palabra+"\"", bg.Frecuencia, barra)
		}
	}
	fmt.Println()

	// Agrupar palabras por longitud usando map[int][]string
	fmt.Println("   📏 Palabras agrupadas por longitud (map[int][]string):")
	porLongitud := make(map[int][]string)
	for palabra := range frecuencia {
		porLongitud[len(palabra)] = append(porLongitud[len(palabra)], palabra)
	}

	longitudes := make([]int, 0, len(porLongitud))
	for l := range porLongitud {
		longitudes = append(longitudes, l)
	}
	sort.Ints(longitudes)

	for _, l := range longitudes {
		pals := porLongitud[l]
		sort.Strings(pals)
		muestra := pals
		if len(muestra) > 6 {
			muestra = muestra[:6]
		}
		sufijo := ""
		if len(pals) > 6 {
			sufijo = fmt.Sprintf(" ... (+%d más)", len(pals)-6)
		}
		fmt.Printf("     %2d letras (%2d palabras): %v%s\n", l, len(pals), muestra, sufijo)
	}
	fmt.Println()

	// ── RESUMEN ──
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   📊 Resumen de la demostración")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("   ✅ Creación de maps:        literal y make()\n")
	fmt.Printf("   ✅ Acceso y modificación:   lectura, escritura, delete\n")
	fmt.Printf("   ✅ Comma ok idiom:          distinguir \"no existe\" de zero value\n")
	fmt.Printf("   ✅ Iteración:               range sobre maps (orden aleatorio)\n")
	fmt.Printf("   ✅ Ordenamiento:            por claves y por valores\n")
	fmt.Printf("   ✅ Nil maps:                lectura OK, escritura PANIC\n")
	fmt.Printf("   ✅ Maps como SET:           map[string]bool y map[string]struct{}\n")
	fmt.Printf("   ✅ Maps anidados:           map[string]map[string]int\n")
	fmt.Printf("   ✅ Agrupación dinámica:     map[string][]T (patrón estrella)\n")
	fmt.Printf("   ✅ Contador de frecuencia:  análisis de texto real con top-N\n")
	fmt.Printf("   ✅ Bigrams:                 pares de palabras consecutivas\n")
	fmt.Printf("   ✅ Agrupación por longitud: map[int][]string\n")
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Println("   ✅ Todos los conceptos de maps ejecutados correctamente")
	fmt.Println("══════════════════════════════════════════════════════════════════")
	fmt.Println()
}