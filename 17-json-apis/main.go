package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ============================================================================
// ESTRUCTURAS DE DATOS — PokeAPI
// ============================================================================

// PokemonResponse representa la respuesta JSON de la PokeAPI.
// Cada campo tiene un struct tag json:"..." que le indica al marshaller
// cómo mapear el campo de Go con la key del JSON.
// Ejemplo: el campo "Nombre" en Go se mapea a "name" en el JSON.
type PokemonResponse struct {
	ID      int        `json:"id"`
	Nombre  string     `json:"name"`
	Altura  int        `json:"height"` // en decímetros
	Peso    int        `json:"weight"` // en hectogramos
	Tipos   []TipoSlot `json:"types"`
	Stats   []StatEntry `json:"stats"`
	Sprites Sprites    `json:"sprites"`
}

// TipoSlot representa un slot de tipo del Pokémon.
// La PokeAPI devuelve un array de tipos porque algunos Pokémon
// tienen doble tipo (ej: Charizard es Fuego/Volador).
type TipoSlot struct {
	Slot int  `json:"slot"` // posición del tipo (1 = primario, 2 = secundario)
	Tipo Tipo `json:"type"` // Nota: en el JSON es "type", no "tipo"
}

// Tipo contiene el nombre del tipo (fire, water, grass, etc.)
type Tipo struct {
	Nombre string `json:"name"`
}

// StatEntry representa una estadística base del Pokémon.
// Cada Pokémon tiene 6 stats: hp, attack, defense, etc.
type StatEntry struct {
	BaseStat int  `json:"base_stat"`
	Stat     Stat `json:"stat"`
}

// Stat contiene el nombre de la estadística.
type Stat struct {
	Nombre string `json:"name"`
}

// Sprites contiene URLs de las imágenes del Pokémon.
// Solo extraemos el sprite frontal por defecto.
type Sprites struct {
	FrontDefault string `json:"front_default"`
}

// ============================================================================
// SISTEMA DE CACHE
// ============================================================================

// CacheEntry almacena una respuesta cacheada con su timestamp.
// Esto nos permite saber cuándo se consultó por última vez.
type CacheEntry struct {
	Datos    PokemonResponse `json:"datos"`
	Consulta time.Time       `json:"consulta"`
}

// CacheLocal es un cache persistente en archivo JSON.
// Usa un mapa de strings (nombres de Pokémon) a CacheEntry.
// Se guarda en disco después de cada escritura para persistencia.
type CacheLocal struct {
	Entradas map[string]CacheEntry `json:"entradas"`
	Archivo  string
}

// NuevoCache crea o carga un cache desde un archivo JSON.
// Si el archivo no existe, crea un cache vacío.
func NuevoCache(archivo string) *CacheLocal {
	cache := &CacheLocal{
		Entradas: make(map[string]CacheEntry),
		Archivo:  archivo,
	}

	// Intentar cargar cache existente
	datos, err := os.ReadFile(archivo)
	if err == nil {
		// El archivo existe, deserializar el JSON al struct
		// Nota: pasamos &cache.Entradas (puntero) para que Unmarshal
		// pueda modificar nuestro mapa
		json.Unmarshal(datos, &cache.Entradas)
	}

	return cache
}

// Buscar intenta encontrar un Pokémon en el cache.
// Devuelve los datos y true si lo encontró, o vacío y false si no.
func (c *CacheLocal) Buscar(nombre string) (PokemonResponse, bool) {
	entrada, existe := c.Entradas[nombre]
	if !existe {
		return PokemonResponse{}, false
	}
	return entrada.Datos, true
}

// Guardar almacena una respuesta en el cache y lo persiste a disco.
// Primero agrega la entrada al mapa, luego serializa todo el mapa
// a JSON y lo escribe al archivo.
func (c *CacheLocal) Guardar(nombre string, datos PokemonResponse) {
	// Crear la entrada con timestamp actual
	c.Entradas[nombre] = CacheEntry{
		Datos:    datos,
		Consulta: time.Now(),
	}

	// Serializar el mapa completo a JSON con indentación legible
	jsonDatos, err := json.MarshalIndent(c.Entradas, "", "  ")
	if err != nil {
		fmt.Printf("  ⚠️  Error al serializar cache: %v\n", err)
		return
	}

	// Escribir al archivo con permisos 0644 (lectura para todos, escritura solo owner)
	err = os.WriteFile(c.Archivo, jsonDatos, 0644)
	if err != nil {
		fmt.Printf("  ⚠️  Error al guardar cache: %v\n", err)
	}
}

// ============================================================================
// CLIENTE DE API REST
// ============================================================================

const pokeAPIBase = "https://pokeapi.co/api/v2/pokemon/"

// ConsultarPokemon hace un GET a la PokeAPI y deserializa la respuesta.
// Usa json.NewDecoder para streaming directo del body HTTP al struct,
// sin cargar toda la respuesta en memoria como []byte.
func ConsultarPokemon(nombre string, verbose bool) (PokemonResponse, error) {
	// Construir la URL: https://pokeapi.co/api/v2/pokemon/pikachu
	url := pokeAPIBase + strings.ToLower(nombre)

	// Crear un cliente HTTP con timeout de 10 segundos.
	// Sin timeout, el programa podría colgarse indefinidamente
	// si el servidor no responde.
	cliente := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Crear el request. http.NewRequest nos permite personalizar
	// headers, método, etc. Aquí solo necesitamos GET.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return PokemonResponse{}, fmt.Errorf("error creando request: %w", err)
	}

	// Headers: le decimos al servidor que queremos JSON
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GoPokemonCLI/1.0")

	// Ejecutar el request
	resp, err := cliente.Do(req)
	if err != nil {
		return PokemonResponse{}, fmt.Errorf("error en request HTTP: %w", err)
	}
	// defer resp.Body.Close() garantiza que el body se cierre
	// cuando esta función termine, incluso si hay un error.
	// Sin esto, tendríamos "resource leaks" (fugas de recursos).
	defer resp.Body.Close()

	// Verificar código de estado HTTP
	if resp.StatusCode == http.StatusNotFound {
		return PokemonResponse{}, fmt.Errorf("pokemon '%s' no encontrado (404)", nombre)
	}
	if resp.StatusCode != http.StatusOK {
		return PokemonResponse{}, fmt.Errorf("API respondió con código %d", resp.StatusCode)
	}

	// Modo verbose: mostrar el JSON crudo antes de deserializar
	if verbose {
		// Leemos el body completo para mostrarlo
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Println("\n📄 JSON crudo de la PokeAPI:")
		fmt.Println("─────────────────────────────────────")
		// Mostrar solo los primeros 500 caracteres para no inundar la consola
		preview := string(bodyBytes)
		if len(preview) > 500 {
			preview = preview[:500] + "\n  ... (truncado)"
		}
		fmt.Println(preview)
		fmt.Println("─────────────────────────────────────")

		// Deserializar desde los bytes ya leídos
		var pokemon PokemonResponse
		err = json.Unmarshal(bodyBytes, &pokemon)
		if err != nil {
			return PokemonResponse{}, fmt.Errorf("error deserializando JSON: %w", err)
		}
		return pokemon, nil
	}

	// Deserializar directamente del body usando json.NewDecoder.
	// Esto es más eficiente que leer todo el body a []byte primero,
	// especialmente para respuestas grandes.
	var pokemon PokemonResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&pokemon)
	if err != nil {
		return PokemonResponse{}, fmt.Errorf("error deserializando JSON: %w", err)
	}

	return pokemon, nil
}

// ============================================================================
// UTILIDADES DE PRESENTACIÓN
// ============================================================================

// TipoColor devuelve un emoji representativo para cada tipo de Pokémon.
// Esto demuestra cómo usar maps como tablas de lookup.
func TipoColor(tipo string) string {
	colores := map[string]string{
		"normal":   "⬜ Normal",
		"fire":     "🔥 Fuego",
		"water":    "💧 Agua",
		"grass":    "🌿 Planta",
		"electric": "⚡ Eléctrico",
		"ice":      "❄️  Hielo",
		"fighting": "🥊 Lucha",
		"poison":   "☠️  Veneno",
		"ground":   "🌍 Tierra",
		"flying":   "🦅 Volador",
		"psychic":  "🔮 Psíquico",
		"bug":      "🐛 Bicho",
		"rock":     "🪨 Roca",
		"ghost":    "👻 Fantasma",
		"dragon":   "🐉 Dragón",
		"dark":     "🌑 Siniestro",
		"steel":    "⚙️  Acero",
		"fairy":    "🧚 Hada",
	}
	if emoji, ok := colores[tipo]; ok {
		return emoji
	}
	return "❓ " + tipo
}

// StatAbreviatura devuelve el nombre corto y legible de cada stat.
func StatAbreviatura(stat string) string {
	abrevs := map[string]string{
		"hp":              "❤️  HP",
		"attack":          "⚔️  Ataque",
		"defense":         "🛡️  Defensa",
		"special-attack":  "✨ At. Esp.",
		"special-defense": "🔮 Def. Esp.",
		"speed":           "💨 Velocidad",
	}
	if abr, ok := abrevs[stat]; ok {
		return abr
	}
	return stat
}

// BarraStat genera una barra visual proporcional al stat.
// El máximo posible de un stat base es ~255 (Blissey tiene 255 HP).
func BarraStat(valor int) string {
	// Escalar: 255 = 20 caracteres
	largo := valor * 20 / 255
	if largo < 1 {
		largo = 1
	}
	return strings.Repeat("█", largo) + strings.Repeat("░", 20-largo)
}

// MostrarPokemon muestra la información de un Pokémon formateada en consola.
func MostrarPokemon(p PokemonResponse) {
	fmt.Println()
	fmt.Printf("  ╔══════════════════════════════════════════════════╗\n")
	fmt.Printf("  ║  #%03d  %-42s  ║\n", p.ID, strings.ToUpper(p.Nombre))
	fmt.Printf("  ╠══════════════════════════════════════════════════╣\n")

	// Tipos
	tipos := make([]string, len(p.Tipos))
	for i, t := range p.Tipos {
		tipos[i] = TipoColor(t.Tipo.Nombre)
	}
	fmt.Printf("  ║  Tipo: %-42s ║\n", strings.Join(tipos, " / "))

	// Físico: altura en decímetros → metros, peso en hectogramos → kg
	alturaM := float64(p.Altura) / 10.0
	pesoKg := float64(p.Peso) / 10.0
	fmt.Printf("  ║  Altura: %.1fm  |  Peso: %.1fkg%-18s ║\n",
		alturaM, pesoKg, "")

	fmt.Printf("  ╠══════════════════════════════════════════════════╣\n")
	fmt.Printf("  ║  📊 Estadísticas Base                          ║\n")
	fmt.Printf("  ╠══════════════════════════════════════════════════╣\n")

	// Stats con barras visuales
	total := 0
	for _, s := range p.Stats {
		nombre := StatAbreviatura(s.Stat.Nombre)
		barra := BarraStat(s.BaseStat)
		fmt.Printf("  ║  %-10s %s %3d  ║\n", nombre, barra, s.BaseStat)
		total += s.BaseStat
	}

	fmt.Printf("  ╠══════════════════════════════════════════════════╣\n")
	fmt.Printf("  ║  📈 Total: %-38d ║\n", total)
	fmt.Printf("  ╚══════════════════════════════════════════════════╝\n")

	if p.Sprites.FrontDefault != "" {
		fmt.Printf("  🖼️  Sprite: %s\n", p.Sprites.FrontDefault)
	}
	fmt.Println()
}

// ============================================================================
// MODO DEMOSTRACIÓN
// ============================================================================

// DemoModo muestra una explicación interactiva de JSON y struct tags.
func DemoModo() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   📚 DEMOSTRACIÓN: JSON, Struct Tags y Serialización        ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")

	// --- Ejemplo 1: Marshal básico ---
	fmt.Println("\n━━━ 1. json.Marshal — De Go a JSON ━━━")
	fmt.Println("   Creamos un struct y lo serializamos a JSON:\n")

	type EjemploBasico struct {
		Nombre string `json:"nombre"`
		Edad   int    `json:"edad"`
		Activo bool   `json:"activo"`
	}

	ej := EjemploBasico{Nombre: "Ana", Edad: 25, Activo: true}
	fmt.Printf("   Go struct: %+v\n\n", ej)

	jsonBytes, _ := json.Marshal(ej)
	fmt.Printf("   JSON:      %s\n", string(jsonBytes))

	// --- Ejemplo 2: MarshalIndent ---
	fmt.Println("\n━━━ 2. json.MarshalIndent — JSON con formato ━━━")

	jsonPretty, _ := json.MarshalIndent(ej, "   ", "  ")
	fmt.Printf("   JSON formateado:\n\n%s\n", string(jsonPretty))

	// --- Ejemplo 3: omitempty ---
	fmt.Println("\n━━━ 3. omitempty — Campos opcionales ━━━")
	fmt.Println("   Si un campo está en su 'zero value', omitempty lo omite:\n")

	type Config struct {
		Puerto int    `json:"puerto,omitempty"`
		Host   string `json:"host,omitempty"`
		Debug  bool   `json:"debug,omitempty"`
		Nombre string `json:"nombre"`
	}

	c1 := Config{Puerto: 0, Host: "", Debug: false, Nombre: "app"}
	json1, _ := json.Marshal(c1)
	fmt.Printf("   Config vacío:     %s\n", string(json1))
	fmt.Println("   → Puerto(0), Host(\"\"), Debug(false) se omiten")

	c2 := Config{Puerto: 8080, Host: "localhost", Debug: true, Nombre: "app"}
	json2, _ := json.Marshal(c2)
	fmt.Printf("   Config con datos: %s\n", string(json2))

	// --- Ejemplo 4: json:"-"" ---
	fmt.Println("\n━━━ 4. json:\"-\" — Campos ignorados ━━━")

	type Seguro struct {
		Usuario string `json:"usuario"`
		Token   string `json:"-"` // NUNCA se serializa
		secreto string            // no exportado, tampoco se serializa
	}

	s := Seguro{Usuario: "admin", Token: "abc123", secreto: "password"}
	json3, _ := json.Marshal(s)
	fmt.Printf("   Struct: {Usuario: admin, Token: abc123, secreto: password}\n")
	fmt.Printf("   JSON:   %s\n", string(json3))
	fmt.Println("   → Token y secreto NO aparecen en el JSON")

	// --- Ejemplo 5: Unmarshal ---
	fmt.Println("\n━━━ 5. json.Unmarshal — De JSON a Go ━━━")
	fmt.Println("   Deserializamos un JSON a un struct:\n")

	jsonInput := []byte(`{"name":"Pikachu","id":25,"weight":60}`)
	fmt.Printf("   JSON de entrada: %s\n\n", string(jsonInput))

	var pokemon PokemonResponse
	err := json.Unmarshal(jsonInput, &pokemon)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
		return
	}
	fmt.Printf("   Struct Go: ID=%d, Nombre=%s, Peso=%d\n",
		pokemon.ID, pokemon.Nombre, pokemon.Peso)
	fmt.Println("   → Altura, Tipos, Stats quedan en zero value (no estaban en el JSON)")

	// --- Ejemplo 6: JSON anidado ---
	fmt.Println("\n━━━ 6. JSON Anidado — Structs dentro de structs ━━━")

	type Habilidad struct {
		Nombre string `json:"nombre"`
		Nivel  int    `json:"nivel"`
	}

	type Personaje struct {
		Nombre      string     `json:"nombre"`
		Habilidades []Habilidad `json:"habilidades"`
	}

	p := Personaje{
		Nombre: "Guerrero",
		Habilidades: []Habilidad{
			{Nombre: "Espadazo", Nivel: 5},
			{Nombre: "Escudo", Nivel: 3},
		},
	}
	json4, _ := json.MarshalIndent(p, "   ", "  ")
	fmt.Printf("   Go → JSON:\n\n%s\n", string(json4))

	// --- Ejemplo 7: NuevoEncoder (streaming) ---
	fmt.Println("\n━━━ 7. json.NewEncoder — Streaming a Writer ━━━")
	fmt.Println("   Encode escribe JSON directamente a cualquier io.Writer:")
	fmt.Println("   (aquí escribimos a stdout)\n")
	fmt.Print("   ")
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(map[string]interface{}{
		"mensaje": "¡Hola desde json.NewEncoder!",
		"valor":   42,
	})

	// --- Ejemplo 8: Decode con campos extra ---
	fmt.Println("\n━━━ 8. Campos extra en JSON — Tolerancia ━━━")
	fmt.Println("   Si el JSON tiene campos que no están en el struct,")
	fmt.Println("   Go los IGNORA silenciosamente (no es error):\n")

	jsonConExtras := []byte(`{
		"id": 25,
		"name": "Pikachu",
		"species": "Mouse Pokémon",
		"habitat": "forest",
		"no_existe_en_go": true
	}`)
	fmt.Printf("   JSON con 5 campos:\n   %s\n", string(jsonConExtras))

	var p2 PokemonResponse
	json.Unmarshal(jsonConExtras, &p2)
	fmt.Printf("\n   Struct solo captura: ID=%d, Nombre=%s\n", p2.ID, p2.Nombre)
	fmt.Println("   → Los campos 'species', 'habitat' y 'no_existe_en_go' se ignoran")

	fmt.Println("\n╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║   ✅ Fin de la demostración. Ejecuta con un nombre de       ║")
	fmt.Println("║   Pokémon para ver el ejercicio práctico en acción.         ║")
	fmt.Println("║   Ejemplo: go run main.go pikachu                           ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	// Definir flags de línea de comandos usando el paquete flag.
	// flag es como un recepcionista que interpreta tus pedidos.
	demo := flag.Bool("demo", false, "Modo demostración: explica JSON y struct tags")
	verbose := flag.Bool("verbose", false, "Modo detallado: muestra el JSON crudo de la API")
	flag.Parse()

	// Si se activa el modo demo, mostrar explicaciones y salir
	if *demo {
		DemoModo()
		return
	}

	// Obtener el nombre del Pokémon de los argumentos posicionales.
	// flag.Args() devuelve los argumentos que no son flags.
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("╔══════════════════════════════════════════════════╗")
		fmt.Println("║  🔍 Pokémon CLI — Cliente de PokeAPI con Cache  ║")
		fmt.Println("╠══════════════════════════════════════════════════╣")
		fmt.Println("║                                                  ║")
		fmt.Println("║  Uso:                                            ║")
		fmt.Println("║    go run main.go <nombre_pokemon>               ║")
		fmt.Println("║    go run main.go -demo                          ║")
		fmt.Println("║    go run main.go -verbose <nombre_pokemon>      ║")
		fmt.Println("║                                                  ║")
		fmt.Println("║  Ejemplos:                                       ║")
		fmt.Println("║    go run main.go pikachu                        ║")
		fmt.Println("║    go run main.go charizard                      ║")
		fmt.Println("║    go run main.go mewtwo                         ║")
		fmt.Println("║    go run main.go -demo                          ║")
		fmt.Println("║                                                  ║")
		fmt.Println("╚══════════════════════════════════════════════════╝")
		return
	}

	nombre := strings.ToLower(args[0])

	// Inicializar el cache local.
	// El archivo pokecache.json se crea en el directorio actual.
	cache := NuevoCache("pokecache.json")

	// Paso 1: Buscar en cache
	pokemon, encontrado := cache.Buscar(nombre)
	if encontrado {
		fmt.Printf("  ✅ ¡Cache hit! '%s' encontrado en cache local.\n", nombre)
		fmt.Printf("     Consultado: %s\n",
			pokemonTipos(cache.Entradas[nombre].Consulta))
		MostrarPokemon(pokemon)
		return
	}

	// Paso 2: No estaba en cache, consultar la API
	fmt.Printf("  🌐 Consultando PokeAPI por '%s'...\n", nombre)

	pokemon, err := ConsultarPokemon(nombre, *verbose)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Paso 3: Guardar en cache para futuras consultas
	cache.Guardar(nombre, pokemon)
	fmt.Printf("  💾 Guardado en cache local (pokecache.json)\n")

	// Paso 4: Mostrar resultado
	MostrarPokemon(pokemon)
}

// pokemonTipos formatea un time.Time de forma legible.
// Es un wrapper simple para no repetir el layout en todo el código.
func pokemonTipos(t time.Time) string {
	return t.Format("02/01/2006 15:04:05")
}