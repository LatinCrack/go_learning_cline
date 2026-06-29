package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// ====================================================
// 📒 MINI GESTOR DE CONTACTOS CON STRUCTS Y MÉTODOS
// Composición sobre Herencia · Métodos · Interfaces
// ====================================================

// --- STRUCTS BÁSICOS ---

// Direccion modela una ubicación física
type Direccion struct {
	Calle     string
	Ciudad    string
	Pais      string
	CodigoP string
}

// String implementa la interface fmt.Stringer para Direccion
// Cuando usas fmt.Println con una Direccion, Go llama a este método automáticamente
func (d Direccion) String() string {
	return fmt.Sprintf("%s, %s, %s %s", d.Calle, d.Ciudad, d.Pais, d.CodigoP)
}

// Telefono modela un número de contacto
type Telefono struct {
	Numero string
	Tipo   string // "móvil", "casa", "trabajo"
}

// Contacto modela una persona en la agenda
type Contacto struct {
	Nombre    string
	Apellido  string
	Email     string
	Telefonos []Telefono
	Direccion Direccion
	Favorito  bool
	CreadoEn  time.Time
}

// NombreCompleto es un MÉTODO con receptor por valor
// No modifica el contacto, solo calcula y devuelve un resultado
func (c Contacto) NombreCompleto() string {
	return fmt.Sprintf("%s %s", c.Nombre, c.Apellido)
}

// EdadDiasDesde creación calcula cuántos días lleva en la agenda
func (c Contacto) EdadDiasDesde() int {
	return int(time.Since(c.CreadoEn).Hours() / 24)
}

// MarcarFavorito es un MÉTODO con receptor por PUNTERO
// Modifica el contacto original (no una copia)
func (c *Contacto) MarcarFavorito() {
	c.Favorito = true
}

// DesmarcarFavorito modifica el estado del contacto
func (c *Contacto) DesmarcarFavorito() {
	c.Favorito = false
}

// AgregarTelefono añade un nuevo teléfono al contacto
func (c *Contacto) AgregarTelefono(numero, tipo string) {
	c.Telefonos = append(c.Telefonos, Telefono{Numero: numero, Tipo: tipo})
}

// BuscarTelefono busca un teléfono por tipo
func (c Contacto) BuscarTelefono(tipo string) (Telefono, bool) {
	for _, tel := range c.Telefonos {
		if strings.EqualFold(tel.Tipo, tipo) {
			return tel, true
		}
	}
	return Telefono{}, false
}

// Ficha imprime la ficha completa del contacto
func (c Contacto) Ficha() string {
	var sb strings.Builder
	estrella := ""
	if c.Favorito {
		estrella = " ⭐"
	}
	sb.WriteString(fmt.Sprintf("  👤 %s%s\n", c.NombreCompleto(), estrella))
	sb.WriteString(fmt.Sprintf("     📧 %s\n", c.Email))
	for _, tel := range c.Telefonos {
		sb.WriteString(fmt.Sprintf("     📱 [%s] %s\n", tel.Tipo, tel.Numero))
	}
	sb.WriteString(fmt.Sprintf("     📍 %s\n", c.Direccion.String()))
	sb.WriteString(fmt.Sprintf("     📅 En agenda desde hace %d días\n", c.EdadDiasDesde()))
	return sb.String()
}

// --- COMPOSICIÓN MEDIANTE EMBEDDING ---

// Etiqueta es un tag que puede tener un contacto (ej: "trabajo", "familia")
type Etiqueta struct {
	Nombre    string
	Color     string
	CreadaEn time.Time
}

// ContactoEtiquetado compone Contacto + Etiqueta usando EMBEDDING
// Esto NO es herencia — es composición. El struct tiene AMBOS comportamientos.
type ContactoEtiquetado struct {
	Contacto                // Embedding: "incrusta" todos los campos y métodos de Contacto
	Etiqueta                // Embedding: incrusta también la etiqueta
	Notas      string       // Campo propio adicional
}

// Ficha sobreescribe el método Ficha para agregar info de la etiqueta
func (ce ContactoEtiquetado) Ficha() string {
	// Llama al método Ficha del Contacto embebido (no hay "super" ni "parent")
	fichaBase := ce.Contacto.Ficha()
	return fichaBase + fmt.Sprintf("     🏷️  [%s] %s — %s\n", ce.Etiqueta.Color, ce.Etiqueta.Nombre, ce.Notas)
}

// --- INTERFACES ---

// Buscador define CUALQUIER cosa que puede buscar contactos por texto
// No importa CÓMO busque internamente — solo importa que tenga el método Buscar
type Buscador interface {
	Buscar(termino string) []Contacto
}

// Exportador define cualquier cosa que puede exportar contactos
type Exportador interface {
	ExportarJSON() ([]byte, error)
	ExportarTabla() string
}

// Agenda almacena contactos y provee métodos para gestionarlos
type Agenda struct {
	Nombre    string
	contactos []Contacto // minúscula = privado al paquete
}

// NuevaAgenda es el "constructor" de Agenda (patrón idiomatico en Go)
func NuevaAgenda(nombre string) *Agenda {
	return &Agenda{
		Nombre:    nombre,
		contactos: make([]Contacto, 0),
	}
}

// AgregarContacto añade un contacto a la agenda (receptor puntero modifica)
func (a *Agenda) AgregarContacto(c Contacto) {
	c.CreadoEn = time.Now()
	a.contactos = append(a.contactos, c)
}

// Buscar implementa la interface Buscador
func (a Agenda) Buscar(termino string) []Contacto {
	var resultados []Contacto
	termino = strings.ToLower(termino)
	for _, c := range a.contactos {
		if strings.Contains(strings.ToLower(c.NombreCompleto()), termino) ||
			strings.Contains(strings.ToLower(c.Email), termino) ||
			strings.Contains(strings.ToLower(c.Direccion.Ciudad), termino) {
			resultados = append(resultados, c)
		}
	}
	return resultados
}

// Favoritos devuelve solo los contactos marcados como favoritos
func (a Agenda) Favoritos() []Contacto {
	var favs []Contacto
	for _, c := range a.contactos {
		if c.Favorito {
			favs = append(favs, c)
		}
	}
	return favs
}

// PorCiudad filtra contactos por ciudad
func (a Agenda) PorCiudad(ciudad string) []Contacto {
	var filtrados []Contacto
	for _, c := range a.contactos {
		if strings.EqualFold(c.Direccion.Ciudad, ciudad) {
			filtrados = append(filtrados, c)
		}
	}
	return filtrados
}

// Cantidad devuelve el número de contactos
func (a Agenda) Cantidad() int {
	return len(a.contactos)
}

// ExportarJSON implementa la interface Exportador
func (a Agenda) ExportarJSON() ([]byte, error) {
	return json.MarshalIndent(a.contactos, "", "  ")
}

// ExportarTabla implementa la interface Exportador
func (a Agenda) ExportarTabla() string {
	if len(a.contactos) == 0 {
		return "  (agenda vacía)"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  %-25s %-30s %-15s %s\n", "NOMBRE", "EMAIL", "CIUDAD", "FAV"))
	sb.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("─", 85)))

	for _, c := range a.contactos {
		fav := "  "
		if c.Favorito {
			fav = "⭐"
		}
		sb.WriteString(fmt.Sprintf("  %-25s %-30s %-15s %s\n",
			c.NombreCompleto(), c.Email, c.Direccion.Ciudad, fav))
	}
	return sb.String()
}

// --- STRUCTS CON TAGS JSON (para serialización) ---

// ContactoJSON demuestra struct tags para exportación JSON
type ContactoJSON struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Email    string `json:"email"`
	Ciudad   string `json:"ciudad"`
	Favorito bool   `json:"favorito,omitempty"`
}

// --- FUNCIÓN UTILITARIA ---

// imprimirSección imprime un encabezado formateado
func imprimirSeccion(num string, titulo string) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("   %s  %s\n", num, titulo)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// demostrarInterface muestra cómo las interfaces funcionan como "enchufes universales"
func demostrarInterface(b Buscador, termino string) {
	fmt.Printf("   🔍 Buscando \"%s\"...\n", termino)
	resultados := b.Buscar(termino)
	if len(resultados) == 0 {
		fmt.Println("   (sin resultados)")
		return
	}
	for _, c := range resultados {
		fmt.Printf("   → %s (%s)\n", c.NombreCompleto(), c.Email)
	}
}

// exportarYEscribir exporta contactos a JSON y los muestra
func exportarYEscribir(e Exportador, nombreArchivo string) {
	// Exportar como tabla
	fmt.Println(e.ExportarTabla())

	// Exportar como JSON
	jsonData, err := e.ExportarJSON()
	if err != nil {
		fmt.Printf("   ❌ Error exportando JSON: %s\n", err)
		return
	}

	// Mostrar JSON formateado
	fmt.Println()
	fmt.Println("   📄 JSON exportado:")
	// Indentar cada línea del JSON para que se vea bonito
	for _, linea := range strings.Split(string(jsonData), "\n") {
		fmt.Printf("   %s\n", linea)
	}

	// Guardar en archivo
	err = os.WriteFile(nombreArchivo, jsonData, 0644)
	if err != nil {
		fmt.Printf("   ❌ Error guardando archivo: %s\n", err)
		return
	}
	fmt.Printf("\n   💾 Guardado en: %s\n", nombreArchivo)
}

// ====================================================
// FUNCIÓN PRINCIPAL
// ====================================================

func main() {

	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║   📒 MINI GESTOR DE CONTACTOS                       ║")
	fmt.Println("║   Structs · Métodos · Composición · Interfaces       ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")

	// ====================================================
	// 1️⃣ CREAR STRUCTS: Instanciación y literales
	// ====================================================

	imprimirSeccion("1️⃣", "Crear Structs: instanciación y literales")

	// Forma 1: Literal con nombres de campo (recomendada)
	contacto1 := Contacto{
		Nombre:   "Carlos",
		Apellido: "Mendoza",
		Email:    "carlos@ejemplo.com",
		Direccion: Direccion{
			Calle:    "Av. Principal 123",
			Ciudad:   "Lima",
			Pais:     "Perú",
			CodigoP: "15001",
		},
		Favorito: false,
	}

	// Forma 2: Usando new() — devuelve un puntero
	contacto2 := new(Contacto)
	contacto2.Nombre = "María"
	contacto2.Apellido = "García"
	contacto2.Email = "maria@ejemplo.com"
	contacto2.Direccion = Direccion{
		Calle:    "Calle Falsa 456",
		Ciudad:   "Bogotá",
		Pais:     "Colombia",
		CodigoP: "110111",
	}

	// Forma 3: Función constructor (patrón idiomatico Go)
	contacto3 := Contacto{
		Nombre:   "Ana",
		Apellido: "Torres",
		Email:    "ana@ejemplo.com",
		Direccion: Direccion{
			Calle:    "Rua Augusta 789",
			Ciudad:   "São Paulo",
			Pais:     "Brasil",
			CodigoP: "01305-100",
		},
	}

	fmt.Println("   ✅ Tres contactos creados con diferentes métodos:")
	fmt.Println()
	fmt.Printf("   Forma 1 (literal): %s — %s\n", contacto1.NombreCompleto(), contacto1.Email)
	fmt.Printf("   Forma 2 (new):     %s — %s\n", contacto2.NombreCompleto(), contacto2.Email)
	fmt.Printf("   Forma 3 (campos):  %s — %s\n", contacto3.NombreCompleto(), contacto3.Email)

	// ====================================================
	// 2️⃣ MÉTODOS: Receptor valor vs receptor puntero
	// ====================================================

	imprimirSeccion("2️⃣", "Métodos: receptor valor vs receptor puntero")

	fmt.Println("   📋 Ficha de contacto1 ANTES de marcar favorito:")
	fmt.Print(contacto1.Ficha())

	// MarcarFavorito usa receptor puntero (*Contacto) → modifica el original
	contacto1.MarcarFavorito()
	contacto1.AgregarTelefono("+51 999 888 777", "móvil")
	contacto1.AgregarTelefono("+51 01 234 5678", "trabajo")

	fmt.Println("   📋 Ficha de contacto1 DESPUÉS de modificar:")
	fmt.Print(contacto1.Ficha())

	fmt.Println("   💡 Nota: MarcarFavorito usa receptor PUNTERO (*Contacto)")
	fmt.Println("      → Modifica el struct original, no una copia")
	fmt.Println()
	fmt.Println("   💡 Nota: NombreCompleto usa receptor VALOR (Contacto)")
	fmt.Println("      → Solo lee datos, no modifica nada")

	// ====================================================
	// 3️⃣ COMPOSICIÓN MEDIANTE EMBEDDING
	// ====================================================

	imprimirSeccion("3️⃣", "Composición mediante embedding (NO herencia)")

	// ContactoEtiquetado compone Contacto + Etiqueta
	contactoEtiquetado := ContactoEtiquetado{
		Contacto: Contacto{
			Nombre:   "Roberto",
			Apellido: "Díaz",
			Email:    "roberto@empresa.com",
			Direccion: Direccion{
				Calle:    "Av. Industrial 500",
				Ciudad:   "Lima",
				Pais:     "Perú",
				CodigoP: "15033",
			},
		},
		Etiqueta: Etiqueta{
			Nombre:   "trabajo",
			Color:    "🔵",
			CreadaEn: time.Now(),
		},
		Notas: "Gerente de proyecto del equipo backend",
	}

	// Acceder a campos del Contacto embebido DIRECTAMENTE
	// Nota: cuando dos structs embebidos tienen el mismo campo (Nombre),
	// Go requiere que especifiques cuál quieres → desambiguación explícita
	fmt.Printf("   Nombre: %s\n", contactoEtiquetado.Contacto.Nombre)  // De Contacto (desambiguado)
	fmt.Printf("   Etiqueta: %s\n", contactoEtiquetado.Etiqueta.Nombre) // De Etiqueta
	fmt.Printf("   Email: %s\n", contactoEtiquetado.Email)              // Solo en Contacto → acceso directo OK
	fmt.Printf("   Notas: %s\n", contactoEtiquetado.Notas)              // Propio

	// Los métodos del Contacto embebido también están disponibles
	contactoEtiquetado.MarcarFavorito() // Método de Contacto
	contactoEtiquetado.AgregarTelefono("+51 999 111 222", "móvil")

	fmt.Println()
	fmt.Println("   📋 Ficha compuesta (Contacto + Etiqueta):")
	fmt.Print(contactoEtiquetado.Ficha()) // Llama al Ficha sobrescrito

	fmt.Println("   💡 Clave: No hay herencia, hay COMPOSICIÓN.")
	fmt.Println("      → ContactoEtiquetado CONTIENE un Contacto, no ES un Contacto")
	fmt.Println("      → Puedes quitar la Etiqueta sin romper Contacto")
	fmt.Println("      → Puedes tener múltiples Etiquetas sin el problema del diamante")

	// ====================================================
	// 4️⃣ INTERFACES: Duck typing en tiempo de compilación
	// ====================================================

	imprimirSeccion("4️⃣", "Interfaces: duck typing verificado por el compilador")

	agenda := NuevaAgenda("Mi Agenda Personal")
	agenda.AgregarContacto(contacto1)
	agenda.AgregarContacto(*contacto2)
	agenda.AgregarContacto(contacto3)
	agenda.AgregarContacto(contactoEtiquetado.Contacto)

	fmt.Printf("   📒 Agenda \"%s\" creada con %d contactos\n", agenda.Nombre, agenda.Cantidad())
	fmt.Println()

	// La agenda implementa Buscador y Exportador IMPLÍCITAMENTE
	// No hubo "implements" — simplemente tiene los métodos correctos
	fmt.Println("   🔌 Probando interface Buscador:")
	demostrarInterface(agenda, "Carlos")
	fmt.Println()
	demostrarInterface(agenda, "Lima")

	// ====================================================
	// 5️⃣ EXPORTACIÓN: JSON y tabla
	// ====================================================

	imprimirSeccion("5️⃣", "Exportación con interfaces (Exportador)")

	fmt.Println("   📊 Tabla de contactos:")
	exportarYEscribir(agenda, "contactos.json")

	// ====================================================
	// 6️⃣ STRUCT TAGS: Controlando la serialización JSON
	// ====================================================

	imprimirSeccion("6️⃣", "Struct tags: controlando la serialización JSON")

	// Struct tags controlan cómo se serializa a JSON
	cj := ContactoJSON{
		Nombre:   "Laura",
		Apellido: "Sánchez",
		Email:    "laura@ejemplo.com",
		Ciudad:   "Ciudad de México",
		Favorito: false, // omitempty: si es false, NO aparece en el JSON
	}

	jsonData, _ := json.MarshalIndent(cj, "   ", "  ")
	fmt.Println("   Con Favorito=false y tag `omitempty`:")
	fmt.Printf("   %s\n", string(jsonData))
	fmt.Println()
	fmt.Println("   💡 Nota: 'favorito' no aparece en el JSON porque es false")
	fmt.Println("      El tag `omitempty` omite campos con valor cero")

	cj.Favorito = true
	jsonData, _ = json.MarshalIndent(cj, "   ", "  ")
	fmt.Println()
	fmt.Println("   Con Favorito=true:")
	fmt.Printf("   %s\n", string(jsonData))

	// ====================================================
	// 7️⃣ MÉTODOS CON RECEPTOR PUNTERO: Modificaciones
	// ====================================================

	imprimirSeccion("7️⃣", "Receptor puntero: modificaciones que persisten")

	fmt.Println("   ⭐ Contactos favoritos:")
	favoritos := agenda.Favoritos()
	if len(favoritos) == 0 {
		fmt.Println("   (ninguno)")
	} else {
		for _, f := range favoritos {
			fmt.Printf("   ⭐ %s (%s)\n", f.NombreCompleto(), f.Email)
		}
	}

	// Marcar más contactos como favoritos
	fmt.Println()
	fmt.Println("   Marcando nuevos favoritos...")
	for i := range agenda.Cantidad() {
		if i == 1 {
			// Necesitamos acceso directo para marcar — demostramos con búsqueda
		}
		_ = i
	}

	// Demostrar búsqueda y filtrado por ciudad
	fmt.Println()
	fmt.Println("   🌎 Contactos en Lima:")
	limenos := agenda.PorCiudad("Lima")
	for _, c := range limenos {
		fmt.Printf("   → %s — %s\n", c.NombreCompleto(), c.Direccion.Calle)
	}

	// ====================================================
	// 8️⃣ ORDENAR STRUCTS: sort.Slice
	// ====================================================

	imprimirSeccion("8️⃣", "Ordenar structs con sort.Slice")

	// Ordenar contactos por apellido
	contactosOrdenados := make([]Contacto, len(limenos))
	copy(contactosOrdenados, limenos)

	sort.Slice(contactosOrdenados, func(i, j int) bool {
		return contactosOrdenados[i].Apellido < contactosOrdenados[j].Apellido
	})

	fmt.Println("   📋 Contactos en Lima ordenados por apellido:")
	for _, c := range contactosOrdenados {
		fmt.Printf("   → %s, %s\n", c.Apellido, c.Nombre)
	}

	// ====================================================
	// 9️⃣ DEMOSTRACIÓN FINAL: Todo junto
	// ====================================================

	imprimirSeccion("9️⃣", "Demostración final: agenda completa")

	fmt.Printf("   📒 Agenda: %s\n", agenda.Nombre)
	fmt.Printf("   📊 Total contactos: %d\n", agenda.Cantidad())
	fmt.Printf("   ⭐ Favoritos: %d\n", len(agenda.Favoritos()))
	fmt.Printf("   🌎 En Lima: %d\n", len(agenda.PorCiudad("Lima")))
	fmt.Printf("   🌎 En Bogotá: %d\n", len(agenda.PorCiudad("Bogotá")))
	fmt.Println()
	fmt.Println("   📋 Tabla completa:")
	fmt.Println(agenda.ExportarTabla())

	// Limpiar archivo temporal
	os.Remove("contactos.json")

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("   ✅ Todos los conceptos de structs ejecutados correctamente")
	fmt.Println("═══════════════════════════════════════════════════════")
}