package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ╔══════════════════════════════════════════════════════════════╗
// ║  LECCION 16: Paquete os, io y el Mundo de Archivos          ║
// ║  Ejercicio: Herramienta CLI de Backup Incremental           ║
// ╚══════════════════════════════════════════════════════════════╝
//
// Esta herramienta demuestra el poder de los paquetes `os`, `io`,
// `filepath` y `flag` de Go construyendo un sistema de backup
// incremental real y útil.

// ──────────────────────────────────────────────────────────
// ESTRUCTURAS DE DATOS
// ──────────────────────────────────────────────────────────

// FileInfo almacena la información de un archivo para comparación
type FileInfo struct {
	Path     string    // ruta relativa del archivo
	Size     int64     // tamaño en bytes
	ModTime  time.Time // última modificación
	Checksum string    // hash SHA-256 del contenido
}

// BackupManifest es el registro de todos los archivos del último backup
type BackupManifest struct {
	Files    map[string]FileInfo // key = ruta relativa
	BackupAt time.Time           // cuándo se hizo el backup
	SourceDir string             // directorio origen
}

// BackupReport contiene el resumen de una operación de backup
type BackupReport struct {
	Copied    []string // archivos copiados (nuevos o modificados)
	Skipped   []string // archivos sin cambios
	Errors    []string // errores durante el backup
	StartTime time.Time
	EndTime   time.Time
}

// ──────────────────────────────────────────────────────────
// FUNCIONES DE MANEJO DE ARCHIVOS (os + filepath)
// ──────────────────────────────────────────────────────────

// crearDirectorioSeguro crea un directorio si no existe.
// Demuestra: os.MkdirAll, os.IsNotExist
func crearDirectorioSeguro(path string) error {
	// os.Stat retorna info del archivo o un error si no existe
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// os.MkdirAll crea el directorio y todos los padres necesarios
		// Es como `mkdir -p` en la terminal
		return os.MkdirAll(path, 0755)
	}
	return err
}

// escanearDirectorio recorre recursivamente un directorio.
// Demuestra: filepath.WalkDir (la forma moderna y eficiente)
func escanearDirectorio(root string) ([]string, error) {
	var archivos []string

	// filepath.WalkDir es más eficiente que filepath.Walk porque
	// no llama a os.Stat para cada archivo — usa la info del
	// directorio que ya tiene el sistema operativo.
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // propagar errores de permisos, etc.
		}

		// Solo archivos regulares, ignoramos directorios y symlinks
		if !d.IsDir() {
			// filepath.Rel convierte una ruta absoluta a relativa
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			archivos = append(archivos, rel)
		}

		return nil
	})

	return archivos, err
}

// ──────────────────────────────────────────────────────────
// FUNCIONES DE COPIA DE ARCHIVOS (io)
// ──────────────────────────────────────────────────────────

// copiarArchivo copia un archivo de src a dst usando io.Copy.
// Demuestra: os.Open, os.Create, io.Copy, defer
func copiarArchivo(src, dst string) (int64, error) {
	// os.Open abre el archivo en modo lectura
	// Retorna *os.File que implementa io.Reader
	archivoOrigen, err := os.Open(src)
	if err != nil {
		return 0, fmt.Errorf("abriendo origen %s: %w", src, err)
	}
	// defer garantiza que el archivo se cierre cuando la función termine
	// Se ejecuta en orden LIFO (Last In, First Out)
	defer archivoOrigen.Close()

	// os.Create crea un archivo nuevo (o trunca si existe)
	// Retorna *os.File que implementa io.Writer
	archivoDestino, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("creando destino %s: %w", dst, err)
	}
	defer archivoDestino.Close()

	// ⭐ io.Copy es LA función clave de Go para transferir datos.
	// Copia de un Reader a un Writer sin cargar todo en memoria.
	// Internamente usa un buffer de 32KB — eficiente para archivos
	// de cualquier tamaño, desde 1 byte hasta varios GB.
	//
	// Esta es la belleza de io.Reader e io.Writer: no importa si
	// la fuente es un archivo, una conexión HTTP, un string o un
	// buffer en memoria — io.Copy funciona con TODOS.
	bytesCopiados, err := io.Copy(archivoDestino, archivoOrigen)
	if err != nil {
		return 0, fmt.Errorf("copiando %s → %s: %w", src, dst, err)
	}

	return bytesCopiados, nil
}

// calcularChecksum genera un hash SHA-256 del archivo.
// Demuestra: io.Copy con un hash.Hash (que implementa io.Writer)
func calcularChecksum(path string) (string, error) {
	archivo, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer archivo.Close()

	// sha256.New() retorna un hash.Hash que implementa io.Writer.
	// Esto significa que podemos pasarle datos con io.Copy,
	// y el hash se calcula "streaming" — sin cargar todo el
	// archivo en memoria. ¡Otra demostración del poder de las
	// interfaces de io!
	hasher := sha256.New()

	if _, err := io.Copy(hasher, archivo); err != nil {
		return "", err
	}

	// Sum(nil) retorna el hash final como []byte
	// hex.EncodeToString lo convierte a string hexadecimal
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ──────────────────────────────────────────────────────────
// FUNCIONES DEL MANIFIESTO (os + encoding implícito)
// ──────────────────────────────────────────────────────────

// guardarManifiesto escribe el manifiesto a un archivo.
// Demuestra: os.Create, fmt.Fprintf (usa io.Writer internamente)
func guardarManifiesto(m BackupManifest, path string) error {
	archivo, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creando manifiesto: %w", err)
	}
	defer archivo.Close()

	// Usamos un buffered writer para escribir eficientemente
	// bufio.Writer implementa io.Writer y acumula datos en un
	// buffer interno antes de escribir al archivo real.
	// Esto reduce las llamadas al sistema operativo.
	writer := bufio.NewWriter(archivo)
	defer writer.Flush()

	// Escribimos encabezado
	fmt.Fprintf(writer, "# Manifiesto de Backup\n")
	fmt.Fprintf(writer, "source_dir=%s\n", m.SourceDir)
	fmt.Fprintf(writer, "backup_at=%s\n", m.BackupAt.Format(time.RFC3339))
	fmt.Fprintf(writer, "file_count=%d\n", len(m.Files))
	fmt.Fprintf(writer, "---\n")

	// Escribimos cada archivo
	for ruta, info := range m.Files {
		fmt.Fprintf(writer, "%s|%d|%s|%s\n",
			ruta,
			info.Size,
			info.ModTime.Format(time.RFC3339),
			info.Checksum,
		)
	}

	return nil
}

// cargarManifiesto lee un manifiesto desde un archivo.
// Demuestra: os.Open, bufio.Scanner, strings.Split
func cargarManifiesto(path string) (BackupManifest, error) {
	m := BackupManifest{
		Files: make(map[string]FileInfo),
	}

	archivo, err := os.Open(path)
	if err != nil {
		return m, fmt.Errorf("abriendo manifiesto: %w", err)
	}
	defer archivo.Close()

	// bufio.Scanner lee línea por línea — mucho más eficiente
	// que leer todo el archivo de una vez para archivos grandes.
	scanner := bufio.NewScanner(archivo)

	// Aumentamos el buffer máximo para archivos con líneas largas
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	enDatos := false
	for scanner.Scan() {
		linea := scanner.Text()

		if linea == "---" {
			enDatos = true
			continue
		}

		if !enDatos {
			// Parsear encabezado
			if strings.HasPrefix(linea, "source_dir=") {
				m.SourceDir = strings.TrimPrefix(linea, "source_dir=")
			} else if strings.HasPrefix(linea, "backup_at=") {
				ts := strings.TrimPrefix(linea, "backup_at=")
				m.BackupAt, _ = time.Parse(time.RFC3339, ts)
			}
			continue
		}

		// Parsear datos de archivos: ruta|tamaño|timestamp|checksum
		partes := strings.Split(linea, "|")
		if len(partes) != 4 {
			continue // línea malformada, la saltamos
		}

		var size int64
		fmt.Sscanf(partes[1], "%d", &size)
		modTime, _ := time.Parse(time.RFC3339, partes[2])

		m.Files[partes[0]] = FileInfo{
			Path:     partes[0],
			Size:     size,
			ModTime:  modTime,
			Checksum: partes[3],
		}
	}

	return m, scanner.Err()
}

// ──────────────────────────────────────────────────────────
// FUNCIÓN PRINCIPAL DE BACKUP
// ──────────────────────────────────────────────────────────

// ejecutarBackup orquesta todo el proceso de backup incremental.
// Demuestra: el patrón de composición de funciones de Go
func ejecutarBackup(origen, destino string, usarChecksum bool) BackupReport {
	reporte := BackupReport{
		StartTime: time.Now(),
	}

	// ── Paso 1: Verificar que el directorio origen existe ──
	// os.Stat retorna información sobre un archivo/directorio
	infoOrigen, err := os.Stat(origen)
	if err != nil {
		reporte.Errors = append(reporte.Errors,
			fmt.Sprintf("❌ Directorio origen no existe: %v", err))
		reporte.EndTime = time.Now()
		return reporte
	}
	if !infoOrigen.IsDir() {
		reporte.Errors = append(reporte.Errors,
			fmt.Sprintf("❌ '%s' no es un directorio", origen))
		reporte.EndTime = time.Now()
		return reporte
	}

	// ── Paso 2: Crear directorio destino si no existe ──
	if err := crearDirectorioSeguro(destino); err != nil {
		reporte.Errors = append(reporte.Errors,
			fmt.Sprintf("❌ Error creando destino: %v", err))
		reporte.EndTime = time.Now()
		return reporte
	}

	// ── Paso 3: Cargar manifiesto anterior (si existe) ──
	// filepath.Join une componentes de ruta de forma segura
	// (usa el separador correcto: / en Linux/Mac, \ en Windows)
	manifiestoPath := filepath.Join(destino, ".backup-manifest.txt")
	manifiestoAnterior, err := cargarManifiesto(manifiestoPath)
	if err != nil {
		// Si no existe, es el primer backup — todo es "nuevo"
		manifiestoAnterior = BackupManifest{
			Files: make(map[string]FileInfo),
		}
	}

	// ── Paso 4: Escanear directorio origen ──
	archivos, err := escanearDirectorio(origen)
	if err != nil {
		reporte.Errors = append(reporte.Errors,
			fmt.Sprintf("❌ Error escaneando: %v", err))
		reporte.EndTime = time.Now()
		return reporte
	}

	// ── Paso 5: Comparar y copiar archivos nuevos/modificados ──
	manifiestoNuevo := BackupManifest{
		Files:     make(map[string]FileInfo),
		BackupAt:  time.Now(),
		SourceDir: origen,
	}

	for _, relPath := range archivos {
		srcPath := filepath.Join(origen, relPath)
		dstPath := filepath.Join(destino, relPath)

		// Obtener información actual del archivo
		info, err := os.Stat(srcPath)
		if err != nil {
			reporte.Errors = append(reporte.Errors,
				fmt.Sprintf("⚠️ Error stat %s: %v", relPath, err))
			continue
		}

		// Crear FileInfo actual
		fileInfo := FileInfo{
			Path:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		// ¿Necesitamos copiar este archivo?
		necesitaCopia := false

		anterior, existe := manifiestoAnterior.Files[relPath]
		if !existe {
			// Archivo nuevo — no existía en el backup anterior
			necesitaCopia = true
		} else if anterior.ModTime != fileInfo.ModTime || anterior.Size != fileInfo.Size {
			// Archivo modificado — cambió el timestamp o el tamaño
			necesitaCopia = true
		} else if usarChecksum {
			// Si se pide checksum, verificamos contenido real
			// (un archivo podría tener el mismo tamaño y timestamp
			// pero contenido diferente si fue restaurado)
			checksum, err := calcularChecksum(srcPath)
			if err == nil && checksum != anterior.Checksum {
				necesitaCopia = true
			}
			fileInfo.Checksum = checksum
		}

		if fileInfo.Checksum == "" {
			// Siempre calcular checksum para el manifiesto
			checksum, err := calcularChecksum(srcPath)
			if err == nil {
				fileInfo.Checksum = checksum
			}
		}

		if necesitaCopia {
			// Crear subdirectorios en destino si es necesario
			dirDestino := filepath.Dir(dstPath)
			if err := crearDirectorioSeguro(dirDestino); err != nil {
				reporte.Errors = append(reporte.Errors,
					fmt.Sprintf("⚠️ Error creando dir %s: %v", dirDestino, err))
				continue
			}

			// ⭐ Copiar usando io.Copy (dentro de copiarArchivo)
			bytesCopiados, err := copiarArchivo(srcPath, dstPath)
			if err != nil {
				reporte.Errors = append(reporte.Errors,
					fmt.Sprintf("⚠️ Error copiando %s: %v", relPath, err))
				continue
			}

			reporte.Copied = append(reporte.Copied,
				fmt.Sprintf("  📄 %s (%s)", relPath, formatBytes(bytesCopiados)))
		} else {
			reporte.Skipped = append(reporte.Skipped, relPath)
		}

		manifiestoNuevo.Files[relPath] = fileInfo
	}

	// ── Paso 6: Guardar nuevo manifiesto ──
	if err := guardarManifiesto(manifiestoNuevo, manifiestoPath); err != nil {
		reporte.Errors = append(reporte.Errors,
			fmt.Sprintf("⚠️ Error guardando manifiesto: %v", err))
	}

	// ── Paso 7: Generar log del backup ──
	reporte.EndTime = time.Now()
	generarLog(reporte, destino)

	return reporte
}

// generarLog escribe un archivo de log con los resultados.
// Demuestra: os.Create, fmt.Fprintf, time formatting
func generarLog(reporte BackupReport, destino string) {
	logPath := filepath.Join(destino, "backup-log.txt")

	// Abrimos en modo append (agregar al final) si existe
	// os.OpenFile con O_APPEND | O_CREATE | O_WRONLY es como
	// abrir un archivo para "agregar" sin borrar lo anterior
	archivo, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("⚠️ Error creando log: %v\n", err)
		return
	}
	defer archivo.Close()

	writer := bufio.NewWriter(archivo)
	defer writer.Flush()

	duracion := reporte.EndTime.Sub(reporte.StartTime)

	fmt.Fprintf(writer, "\n%s\n", strings.Repeat("═", 60))
	fmt.Fprintf(writer, "📋 BACKUP LOG — %s\n", reporte.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(writer, "%s\n", strings.Repeat("═", 60))
	fmt.Fprintf(writer, "⏱️  Duración: %v\n", duracion)
	fmt.Fprintf(writer, "✅ Archivos copiados: %d\n", len(reporte.Copied))
	fmt.Fprintf(writer, "⏭️  Archivos sin cambios: %d\n", len(reporte.Skipped))
	fmt.Fprintf(writer, "❌ Errores: %d\n", len(reporte.Errors))

	if len(reporte.Copied) > 0 {
		fmt.Fprintf(writer, "\n📋 Archivos copiados:\n")
		for _, c := range reporte.Copied {
			fmt.Fprintf(writer, "%s\n", c)
		}
	}

	if len(reporte.Errors) > 0 {
		fmt.Fprintf(writer, "\n⚠️ Errores:\n")
		for _, e := range reporte.Errors {
			fmt.Fprintf(writer, "%s\n", e)
		}
	}
}

// ──────────────────────────────────────────────────────────
// FUNCIONES DE ENTORNO (paquete os)
// ──────────────────────────────────────────────────────────

// mostrarEntorno demuestra el acceso a variables de entorno
// y propiedades del sistema operativo.
// Demuestra: os.Getenv, os.Hostname, os.Getwd, os.Environ
func mostrarEntorno() {
	fmt.Println("🌍 INFORMACIÓN DEL ENTORNO")
	fmt.Println(strings.Repeat("─", 50))

	// os.Hostname() retorna el nombre del equipo
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "(desconocido)"
	}
	fmt.Printf("  🖥️  Hostname:      %s\n", hostname)

	// os.Getwd() retorna el directorio de trabajo actual
	wd, err := os.Getwd()
	if err != nil {
		wd = "(desconocido)"
	}
	fmt.Printf("  📁 Directorio:     %s\n", wd)

	// os.Getenv lee una variable de entorno específica
	fmt.Printf("  🏠 HOME/USERPROFILE: %s\n", os.Getenv("USERPROFILE"))
	fmt.Printf("  🔧 PATH:          %s\n", truncar(os.Getenv("PATH"), 80))

	// os.Environ retorna TODAS las variables de entorno
	envVars := os.Environ()
	fmt.Printf("  📊 Variables de entorno: %d\n", len(envVars))
}

// mostrarProcesos demuestra cómo obtener el PID y información del proceso.
// Demuestra: os.Getpid, os.Getppid, os.Getuid
func mostrarProcesos() {
	fmt.Println()
	fmt.Println("⚙️  INFORMACIÓN DEL PROCESO")
	fmt.Println(strings.Repeat("─", 50))

	// os.Getpid() retorna el Process ID actual
	fmt.Printf("  🔢 PID:            %d\n", os.Getpid())

	// os.Getppid() retorna el Parent Process ID
	fmt.Printf("  👨 PPID (padre):   %d\n", os.Getppid())

	// os.Getuid() retorna el User ID (solo en Unix)
	// En Windows retorna -1, eso es normal
	uid := os.Getuid()
	if uid >= 0 {
		fmt.Printf("  👤 UID:            %d\n", uid)
	} else {
		fmt.Printf("  👤 UID:            N/A (Windows)\n")
	}
}

// explorarArchivo demuestra cómo obtener información detallada de un archivo.
// Demuestra: os.Stat, os.FileInfo (la interfaz), os.Mode
func explorarArchivo(path string) {
	fmt.Println()
	fmt.Printf("📄 ANÁLISIS DE ARCHIVO: %s\n", path)
	fmt.Println(strings.Repeat("─", 50))

	// os.Stat retorna un os.FileInfo (interface) con toda la info
	// del archivo SIN abrirlo — es una operación del sistema operativo
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}

	// os.FileInfo es una interface con estos métodos:
	// Name() string      → nombre del archivo
	// Size() int64       → tamaño en bytes
	// Mode() FileMode    → permisos y tipo
	// ModTime() time.Time → última modificación
	// IsDir() bool       → ¿es directorio?
	fmt.Printf("  📛 Nombre:         %s\n", info.Name())
	fmt.Printf("  📏 Tamaño:         %s (%d bytes)\n", formatBytes(info.Size()), info.Size())
	fmt.Printf("  🕐 Última modif:   %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	fmt.Printf("  📂 ¿Es directorio? %v\n", info.IsDir())
	fmt.Printf("  🔐 Permisos:       %s (%v)\n", info.Mode().String(), info.Mode())

	// filepath.Ext extrae la extensión del archivo
	fmt.Printf("  📎 Extensión:      %s\n", filepath.Ext(path))

	// filepath.Base extrae el nombre del archivo
	fmt.Printf("  📛 Base:           %s\n", filepath.Base(path))

	// filepath.Dir extrae el directorio padre
	fmt.Printf("  📁 Directorio:     %s\n", filepath.Dir(path))
}

// demostrarManipulacionPaths demuestra las funciones de filepath.
// Demuestra: filepath.Join, filepath.Ext, filepath.Base, filepath.Dir,
// filepath.Abs, filepath.Match, filepath.Glob
func demostrarManipulacionPaths() {
	fmt.Println()
	fmt.Println("🛤️  MANIPULACIÓN DE RUTAS (filepath)")
	fmt.Println(strings.Repeat("─", 50))

	// filepath.Join une componentes de ruta de forma segura
	// Usa el separador correcto según el OS (/ o \)
	ruta := filepath.Join("home", "usuario", "proyecto", "main.go")
	fmt.Printf("  filepath.Join:           %s\n", ruta)

	// filepath.Ext extrae la extensión
	fmt.Printf("  filepath.Ext:            %s\n", filepath.Ext(ruta))

	// filepath.Base extrae el último componente
	fmt.Printf("  filepath.Base:           %s\n", filepath.Base(ruta))

	// filepath.Dir retorna todo excepto el último componente
	fmt.Printf("  filepath.Dir:            %s\n", filepath.Dir(ruta))

	// filepath.Abs convierte a ruta absoluta
	abs, _ := filepath.Abs(".")
	fmt.Printf("  filepath.Abs(\".\"):       %s\n", abs)

	// filepath.Match prueba un patrón glob
	matched, _ := filepath.Match("*.go", "main.go")
	fmt.Printf("  filepath.Match(*.go, main.go): %v\n", matched)

	matched, _ = filepath.Match("*.go", "main.txt")
	fmt.Printf("  filepath.Match(*.go, main.txt): %v\n", matched)

	// filepath.Glob busca archivos que coincidan con un patrón
	fmt.Println()
	fmt.Println("  📂 filepath.Glob(\"*.go\"):")
	matches, _ := filepath.Glob("*.go")
	if matches != nil {
		for _, m := range matches {
			fmt.Printf("     → %s\n", m)
		}
	} else {
		fmt.Println("     → (no hay coincidencias en el directorio actual)")
	}

	// filepath.Clean limpia una ruta (elimina //, ./, ../ redundantes)
	rutaSucia := "home//usuario/../usuario/./proyecto///main.go"
	fmt.Printf("\n  filepath.Clean(\"%s\")\n", rutaSucia)
	fmt.Printf("  → \"%s\"\n", filepath.Clean(rutaSucia))
}

// ──────────────────────────────────────────────────────────
// DEMOSTRACIONES ADICIONALES DE io
// ──────────────────────────────────────────────────────────

// demostrarIOInterfaces demuestra el poder de io.Reader y io.Writer.
// Demuestra: io.Copy, io.MultiWriter, io.TeeReader, strings.NewReader
func demostrarIOInterfaces() {
	fmt.Println()
	fmt.Println("🔌 INTERFACES IO: Reader y Writer")
	fmt.Println(strings.Repeat("─", 50))

	// strings.NewReader crea un io.Reader desde un string.
	// Es útil para pasar strings a funciones que esperan io.Reader
	textoOriginal := "¡Hola, Go es increíble!"
	fuente := strings.NewReader(textoOriginal)
	fmt.Printf("  📥 Fuente: \"%s\"\n", textoOriginal)

	// io.Copy copia de Reader a Writer
	// Aquí usamos os.Stdout (la consola) como Writer
	fmt.Print("  📤 io.Copy → stdout: \"")
	bytesCopiados, _ := io.Copy(os.Stdout, fuente)
	fmt.Printf("\" (%d bytes)\n", bytesCopiados)

	fmt.Println()
	fmt.Println("  💡 Concepto clave: io.Reader e io.Writer son interfaces.")
	fmt.Println("     Cualquier tipo que implemente Read() o Write()")
	fmt.Println("     puede ser usado con io.Copy, io.MultiWriter, etc.")
	fmt.Println()
	fmt.Println("  Tipos que implementan io.Reader:")
	fmt.Println("     • *os.File        (archivos)")
	fmt.Println("     • *bytes.Buffer   (buffers en memoria)")
	fmt.Println("     • *strings.Reader (strings)")
	fmt.Println("     • *net.Conn       (conexiones de red)")
	fmt.Println("     • *gzip.Reader    (archivos comprimidos)")
	fmt.Println()
	fmt.Println("  Tipos que implementan io.Writer:")
	fmt.Println("     • *os.File        (archivos)")
	fmt.Println("     • *bytes.Buffer   (buffers en memoria)")
	fmt.Println("     • *net.Conn       (conexiones de red)")
	fmt.Println("     • *gzip.Writer    (compresión)")
	fmt.Println("     • hash.Hash       (SHA-256, MD5, etc.)")
}

// demostrarLecturaEscritura demuestra diferentes formas de leer/escribir.
// Demuestra: os.ReadFile, os.WriteFile, os.Open, os.Create
func demostrarLecturaEscritura() {
	fmt.Println()
	fmt.Println("📖 LECTURA Y ESCRITURA DE ARCHIVOS")
	fmt.Println(strings.Repeat("─", 50))

	// ── Forma 1: os.ReadFile (lectura completa, simple) ──
	// Ideal para archivos pequeños que quepan en memoria
	contenido, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Printf("  ⚠️ No se pudo leer go.mod: %v\n", err)
		return
	}
	fmt.Printf("  📄 os.ReadFile(\"go.mod\"):\n")
	fmt.Printf("     %d bytes leídos\n", len(contenido))
	// Mostrar las primeras líneas
	lineas := strings.Split(string(contenido), "\n")
	for i, linea := range lineas {
		if i >= 3 {
			fmt.Println("     ...")
			break
		}
		fmt.Printf("     %s\n", linea)
	}

	// ── Forma 2: os.WriteFile (escritura completa, simple) ──
	datos := []byte("Este archivo fue creado por la Lección 16\n")
	err = os.WriteFile("demo-archivo.txt", datos, 0644)
	if err != nil {
		fmt.Printf("  ⚠️ Error escribiendo: %v\n", err)
		return
	}
	fmt.Println()
	fmt.Println("  📝 os.WriteFile(\"demo-archivo.txt\"):")
	fmt.Println("     Archivo creado exitosamente")

	// ── Forma 3: Lectura línea por línea con bufio.Scanner ──
	fmt.Println()
	fmt.Println("  📖 Lectura línea por línea con bufio.Scanner:")
	archivo, err := os.Open("go.mod")
	if err != nil {
		fmt.Printf("  ⚠️ Error: %v\n", err)
		return
	}
	defer archivo.Close()

	scanner := bufio.NewScanner(archivo)
	numLinea := 1
	for scanner.Scan() {
		if numLinea > 3 {
			fmt.Println("     ...")
			break
		}
		fmt.Printf("     Línea %d: %s\n", numLinea, scanner.Text())
		numLinea++
	}

	// Limpiar archivo demo
	os.Remove("demo-archivo.txt")
}

// ──────────────────────────────────────────────────────────
// FUNCIONES AUXILIARES
// ──────────────────────────────────────────────────────────

// formatBytes convierte bytes a formato legible (KB, MB, GB)
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// truncar corta un string a n caracteres con "..."
func truncar(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// imprimirReporte muestra el reporte de backup en consola
func imprimirReporte(reporte BackupReport) {
	duracion := reporte.EndTime.Sub(reporte.StartTime)

	fmt.Println()
	fmt.Println("📊 REPORTE DE BACKUP")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Printf("  ⏱️  Duración:           %v\n", duracion)
	fmt.Printf("  ✅ Archivos copiados:   %d\n", len(reporte.Copied))
	fmt.Printf("  ⏭️  Sin cambios:         %d\n", len(reporte.Skipped))
	fmt.Printf("  ❌ Errores:             %d\n", len(reporte.Errors))

	if len(reporte.Copied) > 0 {
		fmt.Println()
		fmt.Println("  📋 Archivos copiados:")
		for _, c := range reporte.Copied {
			fmt.Println(c)
		}
	}

	if len(reporte.Errors) > 0 {
		fmt.Println()
		fmt.Println("  ⚠️ Errores:")
		for _, e := range reporte.Errors {
			fmt.Println(e)
		}
	}

	fmt.Println(strings.Repeat("═", 60))
}

// crearDatosDePrueba crea archivos de ejemplo para el backup
func crearDatosDePrueba(dir string) error {
	// Crear estructura de directorios
	dirs := []string{
		filepath.Join(dir, "docs"),
		filepath.Join(dir, "src"),
		filepath.Join(dir, "src", "utils"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	// Crear archivos de ejemplo
	archivos := map[string]string{
		filepath.Join(dir, "README.md"):               "# Mi Proyecto\n\nEste es un proyecto de ejemplo.\n",
		filepath.Join(dir, "docs", "guia.md"):          "# Guía de Uso\n\nInstrucciones detalladas...\n",
		filepath.Join(dir, "src", "main.go"):           "package main\n\nfunc main() {\n\tprintln(\"Hola\")\n}\n",
		filepath.Join(dir, "src", "utils", "helpers.go"): "package utils\n\n// Suma dos números\nfunc Suma(a, b int) int {\n\treturn a + b\n}\n",
		filepath.Join(dir, "src", "config.json"):       "{\n  \"port\": 8080,\n  \"debug\": true\n}\n",
	}

	for path, contenido := range archivos {
		if err := os.WriteFile(path, []byte(contenido), 0644); err != nil {
			return err
		}
	}

	return nil
}

// ──────────────────────────────────────────────────────────
// FUNCIÓN MAIN
// ──────────────────────────────────────────────────────────

func main() {
	// ── Definir flags de línea de comandos ──
	// El paquete `flag` parsea argumentos como --origen, -destino, etc.
	source := flag.String("source", "", "📁 Directorio origen (requerido)")
	dest := flag.String("dest", "", "📁 Directorio destino (requerido)")
	checksum := flag.Bool("checksum", false, "🔐 Verificar contenido con SHA-256")
	demo := flag.Bool("demo", false, "🧪 Ejecutar demostraciones de os/io")
	setup := flag.Bool("setup", false, "🛠️  Crear datos de prueba")

	// flag.Parse() procesa los argumentos de la línea de comandos
	flag.Parse()

	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║  LECCION 16: Paquete os, io y el Mundo de Archivos      ║")
	fmt.Println("║  Herramienta CLI de Backup Incremental                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ── Modo demo: mostrar todas las demostraciones ──
	if *demo {
		fmt.Println("🧪 MODO DEMOSTRACIÓN")
		fmt.Println(strings.Repeat("═", 60))

		mostrarEntorno()
		mostrarProcesos()
		demostrarManipulacionPaths()
		demostrarLecturaEscritura()
		demostrarIOInterfaces()

		// Explorar el propio archivo main.go
		explorarArchivo("main.go")

		fmt.Println()
		fmt.Println(strings.Repeat("═", 60))
		fmt.Println("✅ Demostración completada. Ejecuta con --source y --dest")
		fmt.Println("   para hacer un backup real.")
		return
	}

	// ── Modo setup: crear datos de prueba ──
	if *setup {
		fmt.Println("🛠️  Creando datos de prueba en ./test-data/origen ...")
		if err := crearDatosDePrueba("test-data" + string(filepath.Separator) + "origen"); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Datos creados. Ahora ejecuta:")
		fmt.Println("   go run main.go -source test-data/origen -dest test-data/backup")
		return
	}

	// ── Modo backup: ejecutar backup incremental ──
	if *source == "" || *dest == "" {
		fmt.Println("⚠️  Uso: go run main.go -source <origen> -dest <destino>")
		fmt.Println()
		fmt.Println("Opciones:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Ejemplos:")
		fmt.Println("  go run main.go -setup                                     # Crear datos de prueba")
		fmt.Println("  go run main.go -source test-data/origen -dest test-data/backup")
		fmt.Println("  go run main.go -source test-data/origen -dest test-data/backup -checksum")
		fmt.Println("  go run main.go -demo                                      # Ver demostraciones")
		return
	}

	fmt.Printf("📁 Origen:  %s\n", *source)
	fmt.Printf("📁 Destino: %s\n", *dest)
	fmt.Printf("🔐 Checksum: %v\n", *checksum)
	fmt.Println()

	// Ejecutar el backup
	reporte := ejecutarBackup(*source, *dest, *checksum)
	imprimirReporte(reporte)

	// Si hubo errores pero también éxitos, código 0
	// Si TODO falló, código 1
	if len(reporte.Copied) == 0 && len(reporte.Errors) > 0 {
		os.Exit(1)
	}
}