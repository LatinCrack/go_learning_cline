package main

// ============================================================
// 🏗️ LECCIÓN 20: PROYECTO FINAL — GoKV (Key-Value Database)
// ============================================================
//
// Este programa implementa una base de datos Key-Value real:
//
// CAPA 1: Motor de almacenamiento en memoria (concurrente)
// CAPA 2: Persistencia a disco (JSON)
// CAPA 3: Servidor TCP para acceso en red
// CAPA 4: Soporte de transacciones (BEGIN/COMMIT/ROLLBACK)
// CAPA 5: CLI interactivo para enviar comandos
//
// Conceptos de TODAS las 19 lecciones anteriores aplicados:
//   L01-Hello       → main(), fmt.Println
//   L02-Variables   → constantes, tipos, iota
//   L03-Tipos       → tipos personalizados, composición
//   L04-Funciones   → closures, funciones como valores
//   L05-Structs     → DB, Transaction, Response
//   L06-Slices      → listas de operaciones, logs
//   L07-Maps        → almacén de datos, índices
//   L08-Strings     → parsing de comandos, formateo
//   L09-Punteros   → receptores *DB, mutación eficiente
//   L10-Goroutines  → handleClient como goroutine
//   L11-Channels    → señal de shutdown
//   L12-Select      → timeout de operaciones
//   L13-Patrones    → fan-out para flush concurrente
//   L14-Paquetes    → organizado por responsabilidades
//   L15-Testing     → estructura testeable
//   L16-OS-IO       → lectura/escritura de archivos
//   L17-JSON-API    → serialización del dump completo
//   L18-Context     → context.Context para cancelación
//   L19-Generics    → colecciones genéricas (Map, Filter)
//
// Ejecutar servidor:  go run main.go
// Ejecutar cliente:   go run main.go --client
// ============================================================

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// ============================================================
// CAPA 1: MOTOR DE ALMACENAMIENTO EN MEMORIA
// ============================================================
//
// Analogía: Imagina un archivador de oficina con cajones.
// Cada cajón tiene una etiqueta (key) y un documento (value).
//
// El motor es el corazón de GoKV: almacena datos en un map
// protegido por un RWMutex para acceso concurrente seguro.
//
// RWMutex permite:
//   - Múltiples lectores simultáneos (RLock)
//   - Un solo escritor a la vez (Lock)
//   - Lectores bloqueados mientras se escribe
//
// Es como una biblioteca: muchos pueden LEER al mismo tiempo,
// pero solo uno puede REORGANIZAR los estantes a la vez.

// Entry representa un par clave-valor almacenado.
type Entry struct {
	Key       string    `json:"key"`        // La etiqueta del cajón
	Value     string    `json:"value"`      // El documento dentro
	CreatedAt time.Time `json:"created_at"` // Cuándo se creó
	UpdatedAt time.Time `json:"updated_at"` // Última modificación
}

// Database es el motor de almacenamiento principal.
type Database struct {
	mu       sync.RWMutex      // Cerrojo de seguridad para concurrencia
	data     map[string]Entry  // El archivador (almacenamiento en memoria)
	dumpPath string            // Ruta del archivo de persistencia (el "respaldo")
	dirty    bool              // ¿Hay cambios sin guardar?
}

// NewDatabase crea una nueva instancia del motor.
// Si existe un dump en disco, lo carga automáticamente.
func NewDatabase(dumpPath string) (*Database, error) {
	db := &Database{
		data:     make(map[string]Entry),
		dumpPath: dumpPath,
		dirty:    false,
	}

	// Intentar cargar datos previos desde disco
	if err := db.loadFromDisk(); err != nil {
		// Si el archivo no existe, no es un error fatal (primera ejecución)
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("error cargando dump: %w", err)
		}
	}

	return db, nil
}

// Put inserta o actualiza un par clave-valor.
// Si la clave ya existe, actualiza su valor y timestamp.
// Si no existe, crea una nueva entrada.
//
// Analogía: Abres un cajón del archivador y pones un documento nuevo.
// Si el cajón ya tenía algo, lo reemplazas.
func (db *Database) Put(key, value string) {
	db.mu.Lock()         // Cerrojo de ESCRITURA (bloquea a todos)
	defer db.mu.Unlock() // Nos aseguramos de liberar al finalizar

	now := time.Now()

	// Verificar si la clave ya existe para preservar CreatedAt
	if existing, existe := db.data[key]; existe {
		// Actualizar: preservamos la fecha de creación original
		existing.Value = value
		existing.UpdatedAt = now
		db.data[key] = existing
	} else {
		// Insertar nueva entrada
		db.data[key] = Entry{
			Key:       key,
			Value:     value,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	db.dirty = true // Marcar que hay cambios sin persistir
}

// Get recupera el valor asociado a una clave.
// Retorna (entry, true) si existe, o (Entry{}, false) si no.
//
// Analogía: Abres un cajón y lees el documento que hay dentro.
func (db *Database) Get(key string) (Entry, bool) {
	db.mu.RLock()         // Cerrojo de LECTURA (permite múltiples lectores)
	defer db.mu.RUnlock() // Liberar al finalizar

	entry, existe := db.data[key]
	return entry, existe
}

// Delete elimina una clave del almacén.
// Retorna true si existía y fue eliminada, false si no existía.
//
// Analogía: Sacas el documento del cajón y lo tiras a la basura.
func (db *Database) Delete(key string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, existe := db.data[key]
	if existe {
		delete(db.data, key)
		db.dirty = true
		return true
	}
	return false
}

// List retorna todas las claves ordenadas alfabéticamente.
// Útil para ver qué hay en la base de datos.
func (db *Database) List() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	claves := make([]string, 0, len(db.data))
	for k := range db.data {
		claves = append(claves, k)
	}

	// sort.Strings ordena alfabéticamente
	sort.Strings(claves)
	return claves
}

// Count retorna el número total de entradas.
func (db *Database) Count() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.data)
}

// ============================================================
// CAPA 2: PERSISTENCIA A DISCO
// ============================================================
//
// Analogía: Al final del día, el archivador se copia a un
// respaldo de seguridad en otra ubicación. Si el edificio
// se incendia (el programa se cierra), al día siguiente
// puedes reconstruir todo desde el respaldo.
//
// Usamos JSON para serializar el map completo a un archivo.
// Es simple pero efectivo para una base de datos educativa.
// En producción usarías algo como BoltDB o BadgerDB.

// Snapshot es la estructura que se serializa a disco.
type Snapshot struct {
	Entries   []Entry `json:"entries"`    // Todos los datos
	Timestamp string  `json:"timestamp"`  // Cuándo se guardó
	Count     int     `json:"count"`      // Cuántas entradas hay
}

// SaveToDisk persiste el estado actual de la base de datos a un archivo JSON.
// Usa el patrón "write-and-rename" para evitar corrupción:
// 1. Escribe a un archivo temporal (.tmp)
// 2. Renombra el temporal al archivo final (atómico en la mayoría de OS)
func (db *Database) SaveToDisk() error {
	db.mu.RLock()
	// Crear snapshot de los datos actuales
	entries := make([]Entry, 0, len(db.data))
	for _, entry := range db.data {
		entries = append(entries, entry)
	}
	db.mu.RUnlock()

	snapshot := Snapshot{
		Entries:   entries,
		Timestamp: time.Now().Format(time.RFC3339),
		Count:     len(entries),
	}

	// Serializar a JSON con indentación legible
	jsonData, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializando datos: %w", err)
	}

	// Escribir primero a archivo temporal (más seguro)
	tmpPath := db.dumpPath + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
		return fmt.Errorf("error escribiendo archivo temporal: %w", err)
	}

	// Renombrar temporal → final (operación atómica)
	if err := os.Rename(tmpPath, db.dumpPath); err != nil {
		return fmt.Errorf("error renombrando archivo: %w", err)
	}

	db.mu.Lock()
	db.dirty = false // Marcar que todo está persistido
	db.mu.Unlock()

	return nil
}

// loadFromDisk carga el estado de la base de datos desde un archivo JSON.
func (db *Database) loadFromDisk() error {
	data, err := os.ReadFile(db.dumpPath)
	if err != nil {
		return err // Propagamos el error (puede ser os.ErrNotExist)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("error parseando dump: %w", err)
	}

	// Reconstruir el map desde el slice de entries
	db.mu.Lock()
	for _, entry := range snapshot.Entries {
		db.data[entry.Key] = entry
	}
	db.mu.Unlock()

	fmt.Printf("   📂 Cargadas %d entradas desde %s\n", len(snapshot.Entries), db.dumpPath)
	return nil
}

// Close guarda los datos pendientes y cierra limpiamente la base de datos.
// Implementa defer: se llama típicamente con `defer db.Close()`
func (db *Database) Close() error {
	if db.dirty {
		fmt.Println("   💾 Guardando datos pendientes antes de cerrar...")
		return db.SaveToDisk()
	}
	fmt.Println("   ✅ No hay datos pendientes. Cierre limpio.")
	return nil
}

// ============================================================
// CAPA 3: SERVIDOR TCP
// ============================================================
//
// Analogía: El servidor TCP es como el mostrador de atención
// al cliente de una biblioteca. Cada persona que llega
// (conexión TCP) recibe un asistente dedicado (goroutine)
// que procesa sus solicitudes.
//
// Protocolo de texto simple (línea por línea):
//   PUT <key> <value>     → Guardar dato
//   GET <key>             → Consultar dato
//   DEL <key>             → Eliminar dato
//   LIST                  → Ver todas las claves
//   COUNT                 → Contar entradas
//   FLUSH                 → Forzar guardado a disco
//   BEGIN                 → Iniciar transacción
//   COMMIT                → Confirmar transacción
//   ROLLBACK              → Deshacer transacción
//   QUIT                  → Cerrar conexión
//   PING                  → Verificar que el servidor vive

// Response es la respuesta que el servidor envía al cliente.
type Response struct {
	Status  string `json:"status"`  // "ok", "error", "not_found"
	Message string `json:"message"` // Mensaje descriptivo
	Data    string `json:"data"`    // Datos opcionales
}

// Servidor mantiene el estado del servidor TCP.
type Servidor struct {
	db       *Database       // La base de datos que sirve
	listener net.Listener    // El "teléfono" que escucha conexiones
	quit     chan struct{}   // Canal para señal de apagado
	wg       sync.WaitGroup // Contador de goroutines activas
}

// NewServidor crea un servidor TCP que sirve la base de datos.
func NewServidor(db *Database, addr string) (*Servidor, error) {
	// net.Listen crea el socket y empieza a escuchar conexiones
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error iniciando servidor en %s: %w", addr, err)
	}

	return &Servidor{
		db:       db,
		listener: listener,
		quit:     make(chan struct{}),
	}, nil
}

// Start pone el servidor en marcha. Acepta conexiones en un loop.
// Cada conexión se maneja en su propia goroutine (concurrencia).
func (s *Servidor) Start(ctx context.Context) error {
	fmt.Printf("   🌐 Servidor escuchando en %s\n", s.listener.Addr())
	fmt.Println("   📡 Esperando conexiones...")
	fmt.Println()

	for {
		// Accept() bloquea hasta que un cliente se conecta
		conn, err := s.listener.Accept()
		if err != nil {
			// Verificar si nos cerraron el servidor
			select {
			case <-s.quit:
				return nil // Apagado limpio
			default:
				return fmt.Errorf("error aceptando conexión: %w", err)
			}
		}

		// Cada cliente recibe su propia goroutine
		// Como un asistente dedicado por cada visitante
		s.wg.Add(1)
		go s.handleClient(ctx, conn)
	}
}

// handleClient maneja la comunicación con UN cliente.
// Ejecuta en su propia goroutine (concurrencia).
//
// Flujo: leer comando → procesar → enviar respuesta → repetir
// Hasta que el cliente envíe QUIT o se desconecte.
func (s *Servidor) handleClient(ctx context.Context, conn net.Conn) {
	defer s.wg.Done()       // Decrementar contador de goroutines
	defer conn.Close()      // Cerrar conexión al terminar

	addr := conn.RemoteAddr().String()
	fmt.Printf("   🔌 Cliente conectado: %s\n", addr)

	// Cada cliente puede tener su propia transacción
	var tx *Transaction

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		// Ver si el contexto fue cancelado (shutdown del servidor)
		select {
		case <-ctx.Done():
			fmt.Printf("   ⚠️  Cliente %s: servidor cerrando\n", addr)
			sendResponse(conn, Response{Status: "error", Message: "servidor cerrando"})
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parsear el comando y sus argumentos
		parts := parseCommand(line)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToUpper(parts[0])
		args := parts[1:]

		var resp Response

		switch cmd {

		// --- PING: Verificar que el servidor está vivo ---
		case "PING":
			resp = Response{Status: "ok", Message: "PONG 🏓"}

		// --- PUT: Guardar un par clave-valor ---
		case "PUT":
			if len(args) < 2 {
				resp = Response{Status: "error", Message: "uso: PUT <key> <value>"}
				break
			}
			key, value := args[0], strings.Join(args[1:], " ")

			// Si hay transacción activa, registrar en ella
			if tx != nil && tx.IsActive {
				tx.AddOperation("PUT", key, value)
				resp = Response{Status: "ok", Message: fmt.Sprintf("PUT '%s' registrado en transacción", key)}
			} else {
				s.db.Put(key, value)
				resp = Response{Status: "ok", Message: fmt.Sprintf("'%s' guardado", key)}
			}

		// --- GET: Consultar el valor de una clave ---
		case "GET":
			if len(args) < 1 {
				resp = Response{Status: "error", Message: "uso: GET <key>"}
				break
			}
			key := args[0]
			entry, existe := s.db.Get(key)
			if !existe {
				resp = Response{Status: "not_found", Message: fmt.Sprintf("'%s' no existe", key)}
			} else {
				resp = Response{
					Status:  "ok",
					Message: fmt.Sprintf("encontrado (creado: %s)", entry.CreatedAt.Format(time.RFC3339)),
					Data:    entry.Value,
				}
			}

		// --- DEL: Eliminar una clave ---
		case "DEL", "DELETE":
			if len(args) < 1 {
				resp = Response{Status: "error", Message: "uso: DEL <key>"}
				break
			}
			key := args[0]

			if tx != nil && tx.IsActive {
				tx.AddOperation("DEL", key, "")
				resp = Response{Status: "ok", Message: fmt.Sprintf("DEL '%s' registrado en transacción", key)}
			} else {
				if s.db.Delete(key) {
					resp = Response{Status: "ok", Message: fmt.Sprintf("'%s' eliminado", key)}
				} else {
					resp = Response{Status: "not_found", Message: fmt.Sprintf("'%s' no existía", key)}
				}
			}

		// --- LIST: Mostrar todas las claves ---
		case "LIST":
			claves := s.db.List()
			if len(claves) == 0 {
				resp = Response{Status: "ok", Message: "base de datos vacía"}
			} else {
				resp = Response{
					Status:  "ok",
					Message: fmt.Sprintf("%d claves encontradas", len(claves)),
					Data:    strings.Join(claves, ", "),
				}
			}

		// --- COUNT: Contar entradas ---
		case "COUNT":
			resp = Response{
				Status:  "ok",
				Message: fmt.Sprintf("%d entradas en la base de datos", s.db.Count()),
			}

		// --- FLUSH: Forzar guardado a disco ---
		case "FLUSH":
			if err := s.db.SaveToDisk(); err != nil {
				resp = Response{Status: "error", Message: fmt.Sprintf("error guardando: %v", err)}
			} else {
				resp = Response{Status: "ok", Message: "datos guardados en disco 💾"}
			}

		// --- BEGIN: Iniciar transacción ---
		case "BEGIN":
			if tx != nil && tx.IsActive {
				resp = Response{Status: "error", Message: "ya hay una transacción activa. COMMIT o ROLLBACK primero"}
				break
			}
			tx = NewTransaction(s.db)
			resp = Response{Status: "ok", Message: "transacción iniciada 📝"}

		// --- COMMIT: Confirmar transacción ---
		case "COMMIT":
			if tx == nil || !tx.IsActive {
				resp = Response{Status: "error", Message: "no hay transacción activa"}
				break
			}
			result := tx.Commit()
			resp = Response{Status: "ok", Message: fmt.Sprintf("transacción confirmada ✅ (%s)", result)}
			tx = nil

		// --- ROLLBACK: Deshacer transacción ---
		case "ROLLBACK":
			if tx == nil || !tx.IsActive {
				resp = Response{Status: "error", Message: "no hay transacción activa"}
				break
			}
			count := tx.Rollback()
			resp = Response{Status: "ok", Message: fmt.Sprintf("transacción desecha 🔄 (%d operaciones descartadas)", count)}
			tx = nil

		// --- STATS: Información del servidor ---
		case "STATS":
			resp = Response{
				Status: "ok",
				Data: fmt.Sprintf("entradas=%d, dump=%s, dirty=%v",
					s.db.Count(), s.db.dumpPath, s.db.dirty),
			}

		// --- QUIT: Cerrar conexión del cliente ---
		case "QUIT":
			sendResponse(conn, Response{Status: "ok", Message: "adiós 👋"})
			fmt.Printf("   🔌 Cliente desconectado: %s\n", addr)
			return

		// --- Comando desconocido ---
		default:
			resp = Response{
				Status:  "error",
				Message: fmt.Sprintf("comando desconocido: '%s'. Comandos: PUT, GET, DEL, LIST, COUNT, FLUSH, BEGIN, COMMIT, ROLLBACK, STATS, PING, QUIT", cmd),
			}
		}

		sendResponse(conn, resp)
	}

	// Si salimos del loop, el cliente se desconectó
	fmt.Printf("   🔌 Cliente desconectado: %s\n", addr)
}

// sendResponse envía una respuesta formateada al cliente.
// Formato: STATUS | message [| data]
func sendResponse(conn net.Conn, resp Response) {
	line := fmt.Sprintf("%s | %s", resp.Status, resp.Message)
	if resp.Data != "" {
		line += " | " + resp.Data
	}
	fmt.Fprintln(conn, line)
}

// parseCommand divide una línea en comando y argumentos,
// respetando comillas para valores con espacios.
//
// Ejemplo:
//
//	PUT nombre "Juan Pérez" → ["PUT", "nombre", "Juan Pérez"]
//	GET nombre              → ["GET", "nombre"]
func parseCommand(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for _, ch := range line {
		switch {
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// ============================================================
// CAPA 4: TRANSACCIONES
// ============================================================
//
// Analogía: Una transacción es como hacer cambios en un borrador
// antes de publicar el documento final. Si algo sale mal,
// tiras el borrador a la basura (ROLLBACK). Si todo está bien,
// publicas los cambios (COMMIT).
//
// Las transacciones garantizan atomicidad: o se aplican TODAS
// las operaciones, o NO se aplica ninguna.

// TxOperation representa una operación dentro de una transacción.
type TxOperation struct {
	Action string // "PUT" o "DEL"
	Key    string
	Value  string
}

// Transaction mantiene el estado de una transacción activa.
type Transaction struct {
	db         *Database     // Referencia a la base de datos
	operations []TxOperation // Operaciones acumuladas (el "borrador")
	IsActive   bool          // ¿La transacción está activa?
}

// NewTransaction crea una nueva transacción vacía.
func NewTransaction(db *Database) *Transaction {
	return &Transaction{
		db:         db,
		operations: make([]TxOperation, 0),
		IsActive:   true,
	}
}

// AddOperation registra una operación en la transacción (sin aplicarla).
func (tx *Transaction) AddOperation(action, key, value string) {
	tx.operations = append(tx.operations, TxOperation{
		Action: action,
		Key:    key,
		Value:  value,
	})
}

// Commit aplica TODAS las operaciones acumuladas a la base de datos.
// Retorna un resumen de lo que se hizo.
func (tx *Transaction) Commit() string {
	tx.IsActive = false

	puts := 0
	dels := 0

	for _, op := range tx.operations {
		switch op.Action {
		case "PUT":
			tx.db.Put(op.Key, op.Value)
			puts++
		case "DEL":
			tx.db.Delete(op.Key)
			dels++
		}
	}

    return fmt.Sprintf("%d PUTs, %d DELs", puts, dels)
}

// Rollback descarta TODAS las operaciones sin aplicarlas.
// Retorna cuántas operaciones se descartaron.
func (tx *Transaction) Rollback() int {
	tx.IsActive = false
	count := len(tx.operations)
	tx.operations = nil // Liberar memoria
	return count
}

// ============================================================
// CAPA 5: CLIENTE CLI INTERACTIVO
// ============================================================
//
// Analogía: El cliente es como un control remoto para la
// base de datos. Te permite enviar comandos desde la
// terminal y ver las respuestas en tiempo real.

// ClienteTCP es un cliente que se conecta al servidor GoKV.
type ClienteTCP struct {
	conn net.Conn // La conexión TCP al servidor
}

// NewClienteTCP crea un cliente y se conecta al servidor.
func NewClienteTCP(addr string) (*ClienteTCP, error) {
	// net.Dial se conecta a un servidor TCP
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error conectando a %s: %w", addr, err)
	}
	return &ClienteTCP{conn: conn}, nil
}

// Send envía un comando al servidor y retorna la respuesta.
func (c *ClienteTCP) Send(comando string) (string, error) {
	// Enviar comando
	fmt.Fprintln(c.conn, comando)

	// Leer respuesta
	reader := bufio.NewReader(c.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error leyendo respuesta: %w", err)
	}

	return strings.TrimSpace(line), nil
}

// Close cierra la conexión del cliente.
func (c *ClienteTCP) Close() error {
	return c.conn.Close()
}

// RunCLI inicia el loop interactivo del cliente.
func (c *ClienteTCP) RunCLI() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("   Comandos: PUT, GET, DEL, LIST, COUNT, FLUSH, BEGIN, COMMIT, ROLLBACK, STATS, PING, QUIT")
	fmt.Println()

	for {
		fmt.Print("gokv> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Enviar al servidor y recibir respuesta
		resp, err := c.Send(line)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			break
		}

		// Parsear y mostrar respuesta con formato
		mostrarRespuesta(resp)

		// Si el comando fue QUIT, salir del loop
		if strings.ToUpper(strings.Fields(line)[0]) == "QUIT" {
			break
		}
	}
}

// mostrarRespuesta formatea la respuesta del servidor para el usuario.
func mostrarRespuesta(resp string) {
	parts := strings.SplitN(resp, " | ", 3)
	status := parts[0]

	emoji := "✅"
	if status == "error" {
		emoji = "❌"
	} else if status == "not_found" {
		emoji = "🔍"
	}

	msg := ""
	if len(parts) > 1 {
		msg = parts[1]
	}
	data := ""
	if len(parts) > 2 {
		data = parts[2]
	}

	fmt.Printf("   %s %s\n", emoji, msg)
	if data != "" {
		fmt.Printf("   📦 %s\n", data)
	}
}

// ============================================================
// MAIN — El punto de entrada que orquesta todo
// ============================================================
func main() {
	// ========================================================
	// MODO DETECCIÓN: ¿Servidor o Cliente?
	// ========================================================
	if len(os.Args) > 1 && os.Args[1] == "--client" {
		runCliente()
		return
	}

	runServidor()
}

// runServidor inicia el servidor GoKV.
func runServidor() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🏗️  GOKV — Base de Datos Key-Value en Go")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// ========================================================
	// PASO 1: Crear la base de datos (motor de almacenamiento)
	// ========================================================
	fmt.Println("📦 Inicializando motor de almacenamiento...")
	db, err := NewDatabase("gokv-dump.json")
	if err != nil {
		fmt.Printf("❌ Error fatal: %v\n", err)
		os.Exit(1)
	}

	// defer db.Close() se ejecuta cuando main() termine.
	// Garantiza que los datos se guarden al cerrar.
	defer db.Close()

	fmt.Printf("   ✅ Motor listo (%d entradas en memoria)\n", db.Count())
	fmt.Println()

	// ========================================================
	// PASO 2: Crear contexto con cancelación
	// ========================================================
	// context.WithCancel permite cancelar el servidor limpiamente.
	// Como el botón de apagado de una computadora: señaliza a
	// todas las goroutines que deben terminar.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ========================================================
	// PASO 3: Crear y arrancar el servidor TCP
	// ========================================================
	srv, err := NewServidor(db, ":6969")
	if err != nil {
		fmt.Printf("❌ Error creando servidor: %v\n", err)
		os.Exit(1)
	}

	// Goroutine que guarda periódicamente (auto-save cada 30s)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if db.dirty {
					if err := db.SaveToDisk(); err != nil {
						fmt.Printf("   ⚠️  Error en auto-save: %v\n", err)
					} else {
						fmt.Println("   💾 Auto-save completado")
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	fmt.Println("🚀 Servidor GoKV iniciado")
	fmt.Println("   📡 Puerto: 6969")
	fmt.Println("   💾 Dump: gokv-dump.json")
	fmt.Println("   ⏱️  Auto-save: cada 30 segundos")
	fmt.Println("   🔗 Cliente: go run main.go --client")
	fmt.Println()

	// Iniciar el servidor (bloquea hasta que se cierre)
	if err := srv.Start(ctx); err != nil {
		fmt.Printf("❌ Error del servidor: %v\n", err)
	}

	// Al llegar aquí, el servidor se cerró
	fmt.Println("🛑 Servidor detenido.")
}

// runCliente inicia el cliente CLI interactivo.
func runCliente() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🔗 GOKV — Cliente Interactivo")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("   Conectando a localhost:6969...")
	cliente, err := NewClienteTCP("localhost:6969")
	if err != nil {
		fmt.Printf("❌ No se pudo conectar: %v\n", err)
		fmt.Println("   ¿Está corriendo el servidor? (go run main.go)")
		os.Exit(1)
	}
	defer cliente.Close()

	// Verificar conexión con PING
	resp, err := cliente.Send("PING")
	if err != nil {
		fmt.Printf("❌ Error en PING: %v\n", err)
		os.Exit(1)
	}
	mostrarRespuesta(resp)
	fmt.Println()

	// Iniciar loop interactivo
	cliente.RunCLI()

	fmt.Println("\n👋 Desconectado. ¡Hasta luego!")
}