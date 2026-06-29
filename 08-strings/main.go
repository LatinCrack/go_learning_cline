package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ═══════════════════════════════════════════════════════════════
// LECCIÓN 08 — Truncador Unicode-Safe y Analizador de Texto
// ═══════════════════════════════════════════════════════════════
//
// Este programa demuestra el manejo correcto de strings en Go:
// - Diferencia entre bytes y runes
// - Truncamiento seguro respetando Unicode
// - Detección y extracción de emojis
// - Inversión de strings Unicode-safe
// - Estadísticas completas de texto
//
// Ejecutar: go run main.go

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║  🧪 LECCIÓN 08 — Strings, Runes y el Universo Unicode   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ─────────────────────────────────────────────────────────
	// SECCIÓN 1: Demostración de la trampa de len()
	// ─────────────────────────────────────────────────────────
	demostracionLenVsRunes()

	// ─────────────────────────────────────────────────────────
	// SECCIÓN 2: Truncador Unicode-Safe
	// ─────────────────────────────────────────────────────────
	demostracionTruncador()

	// ─────────────────────────────────────────────────────────
	// SECCIÓN 3: Iteración bytes vs runes
	// ─────────────────────────────────────────────────────────
	demostracionIteracion()

	// ─────────────────────────────────────────────────────────
	// SECCIÓN 4: Detector y extractor de emojis
	// ─────────────────────────────────────────────────────────
	demostracionEmojis()

	// ─────────────────────────────────────────────────────────
	// SECCIÓN 5: Inversión de strings Unicode-Safe
	// ─────────────────────────────────────────────────────────
	demostracionInversion()

	// ─────────────────────────────────────────────────────────
	// SECCIÓN 6: Estadísticas completas de texto
	// ─────────────────────────────────────────────────────────
	demostracionEstadisticas()

	// ─────────────────────────────────────────────────────────
	// SECCIÓN 7: Herramienta práctica — Limitador de texto para tweets
	// ─────────────────────────────────────────────────────────
	demostracionLimitadorTweets()

	fmt.Println("\n═══════════════════════════════════════════════════════════")
	fmt.Println("   ✅ Todas las demostraciones completadas")
	fmt.Println("═══════════════════════════════════════════════════════════")
}

// ═══════════════════════════════════════════════════════════════
// SECCIÓN 1: La gran trampa de len()
// ═══════════════════════════════════════════════════════════════
//
// len() cuenta BYTES, no caracteres. Esta es la fuente #1 de bugs
// en aplicaciones que manejan texto internacional.

func demostracionLenVsRunes() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  📏 SECCIÓN 1: len() vs RuneCountInString()")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Textos de prueba con diferentes tipos de caracteres
	// Cada uno usa un rango Unicode diferente
	textos := []struct {
		nombre string
	 texto  string
	}{
		{"ASCII puro", "Hello World"},                    // Solo caracteres ASCII (1 byte cada uno)
		{"Español con tildes", "café résumé"},           // Caracteres latinos extendidos (2 bytes cada uno)
		{"Chino/Japonés", "日本語 Go語言"},               // CJK (3 bytes cada uno)
		{"Emoji simple", "Hola 😀 mundo 🌍"},           // Emojis (4 bytes cada uno)
		{"Emoji compuesto", "Familia: 👨‍👩‍👧‍👦"},       // Emoji con ZWJ (múltiples runes por emoji)
		{"Mezcla total", "¡Hola! 世界 🎉 café ☕ 123"}, // Todo junto
	}

	// Encabezado de la tabla de comparación
	fmt.Printf("  %-22s │ %-6s │ %-8s │ %-7s │ %s\n",
		"Texto", "len()", "RuneCount", "Runes[]", "Texto")
	fmt.Println("  " + strings.Repeat("─", 22) + "─┼─" +
		strings.Repeat("─", 6) + "─┼─" +
		strings.Repeat("─", 8) + "─┼─" +
		strings.Repeat("─", 7) + "─┼─" +
		strings.Repeat("─", 20))

	for _, t := range textos {
		bytes := len(t.texto)                                  // Cuenta metros (bytes)
		runeCount := utf8.RuneCountInString(t.texto)           // Cuenta pasajeros (runes) eficientemente
		runeSlice := len([]rune(t.texto))                      // Cuenta pasajeros creando un slice nuevo

		fmt.Printf("  %-22s │ %4d   │ %4d      │ %4d    │ %s\n",
			t.nombre, bytes, runeCount, runeSlice, t.texto)
	}

	fmt.Println()

	// La demostración del bug: truncar por bytes en texto con acentos
	texto := "El café de José está aquí"
	fmt.Printf("  Original: %q\n", texto)
	fmt.Printf("  len(): %d bytes\n", len(texto))
	fmt.Printf("  RuneCount: %d caracteres\n", utf8.RuneCountInString(texto))

	// ❌ BUG: truncar por bytes corta caracteres a la mitad
	truncadoBug := texto[:15]
	fmt.Printf("\n  ❌ truncado[:15] = %q  ← ¡Tildes rotas!\n", truncadoBug)

	// ✅ CORRECTO: truncar por runes
	truncadoOK := truncarUnicode(texto, 15)
	fmt.Printf("  ✅ truncarUnicode(15) = %q  ← Perfecto\n", truncadoOK)
}

// ═══════════════════════════════════════════════════════════════
// SECCIÓN 2: Truncador Unicode-Safe
// ═══════════════════════════════════════════════════════════════
//
// Esta función trunca un string a N runes (caracteres visibles)
// sin cortar caracteres multi-byte a la mitad.

func truncarUnicode(s string, maxRunes int) string {
	// Si el string tiene menos runes que el límite, devolverlo intacto
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}

	// Usamos un contador de runes para encontrar la posición de corte
	// en bytes. range itera por runes, así que cada iteración nos da
	// un rune completo (no un byte suelto).
	contador := 0
	byteIndex := 0
	for i := range s {
		if contador >= maxRunes {
			// Encontramos la posición en bytes donde cortar
			byteIndex = i
			break
		}
		contador++
		// Si llegamos al final sin encontrar el punto de corte
		// (por si maxRunes es igual al total de runes)
		byteIndex = i + utf8.RuneLen(rune(s[i]))
	}

	// Cortamos en la posición encontrada y añadimos "..."
	return s[:byteIndex] + "..."
}

func demostracionTruncador() {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  ✂️  SECCIÓN 2: Truncador Unicode-Safe")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Diferentes textos para truncar
	pruebas := []struct {
	 texto    string
		maxRunes int
	}{
		{"La programación en Go es increíble", 20},
		{"El café de José está delicioso", 15},
		{"日本語はとても美しい言語です", 6},    // Japonés
		{"¡Hola mundo! 🌍🚀🎉💻", 10},         // Con emojis
		{"Este es un texto muy corto", 50},   // Más corto que el límite
	}

	for _, p := range pruebas {
		resultado := truncarUnicode(p.texto, p.maxRunes)
		bytesOriginal := len(p.texto)
		runesOriginal := utf8.RuneCountInString(p.texto)
		runesResultado := utf8.RuneCountInString(resultado)

		fmt.Printf("  Original (%d bytes, %d runes): %q\n",
			bytesOriginal, runesOriginal, p.texto)
		fmt.Printf("  Truncado a %d runes (%d runes): %q\n\n",
			p.maxRunes, runesResultado, resultado)
	}
}

// ═══════════════════════════════════════════════════════════════
// SECCIÓN 3: Iteración — bytes vs runes
// ═══════════════════════════════════════════════════════════════
//
// range itera por runes, pero s[i] accede por bytes.
// Esta diferencia causa bugs cuando manejas texto no-ASCII.

func demostracionIteracion() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  🔍 SECCIÓN 3: Iteración bytes vs runes")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	texto := "café"
	fmt.Printf("  Texto: %q (%d bytes, %d runes)\n\n",
		texto, len(texto), utf8.RuneCountInString(texto))

	// ❌ MÉTODO INCORRECTO: iterar por bytes con for clásico
	fmt.Println("  ❌ Iteración por BYTES (for clásico):")
	fmt.Println("  ┌─────────┬───────────┬──────────────┐")
	fmt.Println("  │ Índice  │ Byte (hex)│ Carácter     │")
	fmt.Println("  ├─────────┼───────────┼──────────────┤")
	for i := 0; i < len(texto); i++ {
		b := texto[i]
		// Si el byte no es un carácter ASCII válido (>= 128),
		// mostrar como byte parcial
		caracter := string(rune(b))
		if b > 127 {
			caracter = "⚠️ PARTE DE É"
		}
		fmt.Printf("  │ %4d     │   0x%02x    │ %-12s │\n", i, b, caracter)
	}
	fmt.Println("  └─────────┴───────────┴──────────────┘")
	fmt.Println()

	// ✅ MÉTODO CORRECTO: iterar por runes con range
	fmt.Println("  ✅ Iteración por RUNES (range):")
	fmt.Println("  ┌─────────┬───────────┬──────────────────┬─────────────┐")
	fmt.Println("  │ Byte Idx│ Rune      │ Punto de código  │ Bytes rune  │")
	fmt.Println("  ├─────────┼───────────┼──────────────────┼─────────────┤")
	for i, r := range texto {
		fmt.Printf("  │ %4d     │    %c      │    U+%04X        │     %d       │\n",
			i, r, r, utf8.RuneLen(r))
	}
	fmt.Println("  └─────────┴───────────┴──────────────────┴─────────────┘")
	fmt.Println()

	// Ejemplo más extremo con japonés
	textoExtremo := "Go言語"
	fmt.Printf("  Texto extremo: %q\n\n", textoExtremo)
	for i, r := range textoExtremo {
		fmt.Printf("  Byte index: %2d | Rune: %c | Código: U+%04X | Bytes: %d\n",
			i, r, r, utf8.RuneLen(r))
	}
}

// ═══════════════════════════════════════════════════════════════
// SECCIÓN 4: Detector y extractor de emojis
// ═══════════════════════════════════════════════════════════════
//
// Los emojis están en rangos específicos de Unicode.
// Go no tiene una función built-in para detectarlos,
// pero podemos usar los rangos conocidos.

func esEmoji(r rune) bool {
	// Rangos principales de emojis en Unicode
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticonos (😀-🙏)
		(r >= 0x1F300 && r <= 0x1F5FF) || // Símbolos y pictogramas (🌀-🗿)
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transporte y mapas (🚀-🛿)
		(r >= 0x1F1E0 && r <= 0x1F1FF) || // Banderas regionales (🇦-🇿)
		(r >= 0x2600 && r <= 0x26FF) ||   // Símbolos misc (☀-⛿)
		(r >= 0x2700 && r <= 0x27BF) ||   // Dingbats (✀-➿)
		(r >= 0xFE00 && r <= 0xFE0F) ||   // Selectores de variación
		(r >= 0x200D && r <= 0x200D) ||   // Zero Width Joiner (ZWJ)
		(r >= 0x1F900 && r <= 0x1F9FF) || // Suplemento de emoticonos (🤀-🧿)
		(r >= 0x1FA00 && r <= 0x1FA6F) || // Ajedrez y símbolos extendidos
		(r >= 0x1FA70 && r <= 0x1FAFF)   // Símbolos extendidos-A
}

func esParteDeEmoji(r rune) bool {
	// Incluye emojis + selectores de variación + ZWJ
	return esEmoji(r) ||
		r == 0xFE0F || // Selector de variación (hace que el emoji sea colorido)
		r == 0x200D || // Zero Width Joiner (une emojis)
		r == 0xFE0E   // Selector de presentación de texto
}

func extraerEmojis(texto string) []string {
	var emojis []string
	var buffer strings.Builder
	dentroDeEmoji := false

	for _, r := range texto {
		if esParteDeEmoji(r) {
			// Estamos dentro de una secuencia de emoji
			buffer.WriteRune(r)
			dentroDeEmoji = true
		} else {
			if dentroDeEmoji {
				// Terminó la secuencia de emoji, guardar lo acumulado
				emojis = append(emojis, buffer.String())
				buffer.Reset()
				dentroDeEmoji = false
			}
			// Si no es emoji, simplemente lo ignoramos
		}
	}

	// Guardar el último emoji si el string termina con uno
	if dentroDeEmoji && buffer.Len() > 0 {
		emojis = append(emojis, buffer.String())
	}

	return emojis
}

func demostracionEmojis() {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  😄 SECCIÓN 4: Detector y Extractor de Emojis")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Textos con diferentes tipos de emojis
	textos := []string{
		"Hola mundo 😀😂🎉",
		"Me encanta Go 🚀",
		"Familia: 👨‍👩‍👧‍👦 y mascota 🐕",
		"Países: 🇵🇪🇲🇽🇪🇸🇯🇵",
		"Sin emojis aquí",
		"Mezcla: café ☕ + pizza 🍕 = 😋",
	}

	for _, texto := range textos {
		emojis := extraerEmojis(texto)
		bytes := len(texto)
		runes := utf8.RuneCountInString(texto)

		fmt.Printf("  Texto: %q\n", texto)
		fmt.Printf("  Bytes: %d | Runes: %d | Emojis encontrados: %d\n",
			bytes, runes, len(emojis))

		if len(emojis) > 0 {
			for i, emoji := range emojis {
				runeCount := utf8.RuneCountInString(emoji)
				fmt.Printf("    Emoji %d: %s (%d runes, %d bytes)\n",
					i+1, emoji, runeCount, len(emoji))
			}
		} else {
			fmt.Println("    (ninguno)")
		}
		fmt.Println()
	}

	// Análisis detallado de un emoji compuesto
	fmt.Println("  🔬 Análisis detallado del emoji 👨‍👩‍👧‍👦:")
	emoji := "👨‍👩‍👧‍👦"
	fmt.Printf("    Bytes totales: %d\n", len(emoji))
	fmt.Printf("    Runes totales: %d\n", utf8.RuneCountInString(emoji))
	fmt.Printf("    Lo que ves: UNA familia\n")
	fmt.Printf("    Lo que hay dentro:\n")

	for i, r := range emoji {
		tipo := "Carácter"
		if r == 0x200D {
			tipo = "🔗 ZWJ (Zero Width Joiner)"
		} else if r == 0xFE0F {
			tipo = "🎨 Selector de variación"
		} else if esEmoji(r) {
			tipo = "😄 Emoji"
		}
		fmt.Printf("      [%d] %c → U+%04X (%s)\n", i, r, r, tipo)
	}
}

// ═══════════════════════════════════════════════════════════════
// SECCIÓN 5: Inversión de strings Unicode-Safe
// ═══════════════════════════════════════════════════════════════
//
// Invertir un string byte por byte DESTRUYE caracteres multi-byte.
// La forma correcta es convertir a []rune, invertir, y volver a string.

func invertirRunes(s string) string {
	runes := []rune(s)
	// Algoritmo de inversión de slice: swap desde los extremos
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func demostracionInversion() {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  🔄 SECCIÓN 5: Inversión de Strings Unicode-Safe")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	textos := []string{
		"Hello",
		"café",
		"日本語",
		"Go es genial",
	}

	for _, texto := range textos {
		// ❌ Inversión INCORRECTA: por bytes
		bytesInvertido := invertirPorBytes(texto)

		// ✅ Inversión CORRECTA: por runes
		runesInvertido := invertirRunes(texto)

		fmt.Printf("  Original:    %q\n", texto)
		fmt.Printf("  ❌ Por bytes: %q", bytesInvertido)
		if bytesInvertido != runesInvertido {
			fmt.Printf("  ← ¡ROTO!")
		}
		fmt.Println()
		fmt.Printf("  ✅ Por runes: %q\n", runesInvertido)
		fmt.Println()
	}

	// Demostración visual del problema
	fmt.Println("  🔍 ¿Por qué invertir por bytes rompe los caracteres?")
	ejemplo := "café"
	fmt.Printf("  Original: %q → bytes: %v\n", ejemplo, []byte(ejemplo))

	bytesOriginales := []byte(ejemplo)
	fmt.Printf("  Bytes invertidos: %v\n", reverseBytes(bytesOriginales))
	fmt.Printf("  Como string: %q ← ¡Basura!\n", string(bytesOriginales))
	fmt.Println()
	fmt.Printf("  Runes originales: %v\n", []rune(ejemplo))
	runesOriginales := []rune(ejemplo)
	for i, j := 0, len(runesOriginales)-1; i < j; i, j = i+1, j-1 {
		runesOriginales[i], runesOriginales[j] = runesOriginales[j], runesOriginales[i]
	}
	fmt.Printf("  Runes invertidos: %v\n", runesOriginales)
	fmt.Printf("  Como string: %q ← ¡Perfecto!\n", string(runesOriginales))
}

// invertirPorBytes invierte un string byte por byte (INCORRECTO para Unicode)
func invertirPorBytes(s string) string {
	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return string(b)
}

// reverseBytes invierte un slice de bytes in-place
func reverseBytes(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

// ═══════════════════════════════════════════════════════════════
// SECCIÓN 6: Estadísticas completas de texto
// ═══════════════════════════════════════════════════════════════
//
// Analiza un texto y genera estadísticas detalladas usando
// todo lo aprendido sobre strings, runes y Unicode.

type EstadisticasTexto struct {
	Texto           string
	Bytes           int
	Runes           int
	Palabras        int
	Líneas          int
	Emojis          int
	EmojisLista     []string
	Dígitos         int
	Letras          int
	Espacios        int
	Puntuación      int
	Símbolos        int
	RangoMasComún   rune
	FrecuenciaMax   int
	RunesUnicos     int
	TamañoPromedio  float64 // bytes por rune
}

func analizarTexto(texto string) EstadisticasTexto {
	stats := EstadisticasTexto{
		Texto: texto,
		Bytes: len(texto),
		Runes: utf8.RuneCountInString(texto),
	}

	// Contar palabras usando strings.Fields (divide por whitespace)
	stats.Palabras = len(strings.Fields(texto))

	// Contar líneas
	stats.Líneas = strings.Count(texto, "\n") + 1

	// Extraer emojis
	stats.EmojisLista = extraerEmojis(texto)
	stats.Emojis = len(stats.EmojisLista)

	// Mapa para contar frecuencia de runes
	frecuencia := make(map[rune]int)
	maxFreq := 0
	var runeMasComún rune

	// Analizar cada rune
	for _, r := range texto {
		// Clasificar el rune
		switch {
		case unicode.IsLetter(r):
			stats.Letras++
		case unicode.IsDigit(r):
			stats.Dígitos++
		case unicode.IsSpace(r):
			stats.Espacios++
		case unicode.IsPunct(r):
			stats.Puntuación++
		case unicode.IsSymbol(r):
			stats.Símbolos++
		}

		// Contar frecuencia (solo para letras y dígitos, ignorando case)
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			normalizado := unicode.ToLower(r)
			frecuencia[normalizado]++
			if frecuencia[normalizado] > maxFreq {
				maxFreq = frecuencia[normalizado]
				runeMasComún = normalizado
			}
		}
	}

	stats.RangoMasComún = runeMasComún
	stats.FrecuenciaMax = maxFreq
	stats.RunesUnicos = len(frecuencia)

	// Tamaño promedio en bytes por rune
	if stats.Runes > 0 {
		stats.TamañoPromedio = float64(stats.Bytes) / float64(stats.Runes)
	}

	return stats
}

func demostracionEstadisticas() {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  📊 SECCIÓN 6: Estadísticas Completas de Texto")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	textos := []string{
		"Hello World! This is Go programming at its finest. 🚀",
		"El café de José está delicioso. ¿Quieres más? ☕😄",
		"日本語のテキスト分析は面白いです 🔍",
		"Mix: café 日本語 🎉 123 Hello! 🌍👨‍👩‍👧‍👦",
	}

	for _, texto := range textos {
		stats := analizarTexto(texto)

		fmt.Printf("  ╔══ Texto: %s\n", truncarUnicode(texto, 40))
		fmt.Printf("  ║\n")
		fmt.Printf("  ║  📏 Dimensiones:\n")
		fmt.Printf("  ║     Bytes:           %d\n", stats.Bytes)
		fmt.Printf("  ║     Runes:           %d\n", stats.Runes)
		fmt.Printf("  ║     Promedio:        %.2f bytes/rune\n", stats.TamañoPromedio)
		fmt.Printf("  ║\n")
		fmt.Printf("  ║  📝 Contenido:\n")
		fmt.Printf("  ║     Palabras:        %d\n", stats.Palabras)
		fmt.Printf("  ║     Líneas:          %d\n", stats.Líneas)
		fmt.Printf("  ║     Letras:          %d\n", stats.Letras)
		fmt.Printf("  ║     Dígitos:         %d\n", stats.Dígitos)
		fmt.Printf("  ║     Espacios:        %d\n", stats.Espacios)
		fmt.Printf("  ║     Puntuación:      %d\n", stats.Puntuación)
		fmt.Printf("  ║     Símbolos:        %d\n", stats.Símbolos)
		fmt.Printf("  ║     Emojis:          %d\n", stats.Emojis)
		fmt.Printf("  ║\n")
		fmt.Printf("  ║  🔤 Análisis:\n")
		fmt.Printf("  ║     Runes únicos:    %d\n", stats.RunesUnicos)
		if stats.FrecuenciaMax > 0 {
			fmt.Printf("  ║     Más frecuente:   '%c' (%d veces)\n",
				stats.RangoMasComún, stats.FrecuenciaMax)
		}
		if stats.Emojis > 0 {
			fmt.Printf("  ║     Emojis:          %s\n",
				strings.Join(stats.EmojisLista, " "))
		}
		fmt.Printf("  ╚══════════════════════════════════════\n\n")
	}
}

// ═══════════════════════════════════════════════════════════════
// SECCIÓN 7: Limitador de texto para tweets/redes sociales
// ═══════════════════════════════════════════════════════════════
//
// Ejercicio práctico real: limitar texto para Twitter/X (280 caracteres)
// respetando Unicode y no cortando palabras a la mitad.

type ResultadoLimitador struct {
	TextoOriginal   string
	TextoLimitado   string
	BytesOriginal   int
	RunesOriginal   int
	RunesLimitado   int
	FueTruncado     bool
	PalabrasCortada bool
}

func limitarParaTweet(texto string, maxCaracteres int) ResultadoLimitador {
	resultado := ResultadoLimitador{
		TextoOriginal: texto,
		BytesOriginal: len(texto),
		RunesOriginal: utf8.RuneCountInString(texto),
	}

	// Si el texto cabe completo, devolverlo tal cual
	runeCount := utf8.RuneCountInString(texto)
	if runeCount <= maxCaracteres {
		resultado.TextoLimitado = texto
		resultado.RunesLimitado = runeCount
		resultado.FueTruncado = false
		return resultado
	}

	resultado.FueTruncado = true

	// Convertir a runes para truncar correctamente
	runes := []rune(texto)

	// Reservar espacio para "..." (3 runes)
	limite := maxCaracteres - 3

	// Buscar el último espacio para no cortar una palabra a la mitad
	// Esto es lo que hace un buen limitador de texto
	corteEnEspacio := false
	for i := limite - 1; i >= 0; i-- {
		if unicode.IsSpace(runes[i]) {
			limite = i
			corteEnEspacio = true
			break
		}
	}

	resultado.PalabrasCortada = !corteEnEspacio

	// Construir el resultado
	truncado := string(runes[:limite])

	// Eliminar espacio final si existe
	truncado = strings.TrimRight(truncado, " ")

	resultado.TextoLimitado = truncado + "..."
	resultado.RunesLimitado = utf8.RuneCountInString(resultado.TextoLimitado)

	return resultado
}

func demostracionLimitadorTweets() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  🐦 SECCIÓN 7: Limitador de Texto para Tweets")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	textos := []string{
		"Acabo de aprender Go y es increíble. La forma en que maneja la concurrencia con goroutines y channels es revolucionaria. Definitivamente es el lenguaje del futuro para sistemas distribuidos y microservicios. ¡Lo recomiendo a todos los desarrolladores!",
		"Hoy visité Tokyo 🇯🇵 y la comida 日本料理 es increíble 🍣🍱. Los templos son hermosos ⛩️ y la gente muy amable. ¡Volveré pronto! ✈️",
		"Corto",
		"Go es un lenguaje de programación creado por Google que se enfoca en la simplicidad, la eficiencia y la concurrencia. Fue diseñado por Rob Pike, Ken Thompson y Robert Griesemer.",
	}

	for _, texto := range textos {
		r := limitarParaTweet(texto, 280)

		fmt.Printf("  Original (%d runes):\n", r.RunesOriginal)
		fmt.Printf("  %q\n\n", truncarUnicode(texto, 80))

		fmt.Printf("  Limitado (%d runes):\n", r.RunesLimitado)
		if r.FueTruncado {
			fmt.Printf("  %q\n", truncarUnicode(r.TextoLimitado, 80))
			fmt.Printf("  ⚠️  Fue truncado")
			if r.PalabrasCortada {
				fmt.Printf(" (cortó una palabra)")
			} else {
				fmt.Printf(" (cortó en espacio)")
			}
			fmt.Println()
		} else {
			fmt.Printf("  ✅ Cabe completo sin truncar\n")
		}
		fmt.Println("  " + strings.Repeat("─", 55))
		fmt.Println()
	}

	// Ejemplo con límite más corto para ver el efecto
	fmt.Println("  📱 Ejemplo con límite de 30 caracteres (estilo tweet corto):")
	textoCorto := "¡Hola mundo! Esto es una prueba del limitador de texto Unicode-Safe en Go 🚀"
	r := limitarParaTweet(textoCorto, 30)
	fmt.Printf("  Original (%d runes):  %s\n", r.RunesOriginal, textoCorto)
	fmt.Printf("  Limitado (%d runes):  %s\n", r.RunesLimitado, r.TextoLimitado)
}