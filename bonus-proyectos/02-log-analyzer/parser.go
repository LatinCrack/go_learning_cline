package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

// LogEntry representa una línea parseada del log.
// Contiene la línea cruda y los campos extraídos relevantes.
type LogEntry struct {
	Raw       string    // Línea original sin procesar
	Timestamp time.Time // Timestamp del evento (si se puede parsear)
	IP        string    // IP origen del request
	Method    string    // Método HTTP (GET, POST, etc.)
	Path      string    // Ruta solicitada
	Status    string    // Código de respuesta HTTP
	LineNum   int64     // Número de línea en el archivo
}

// TailFile abre un archivo de log y monitorea en tiempo real nuevas líneas
// usando os.Seek para posicionarse al final del archivo y un Ticker para
// leer periódicamente el contenido nuevo. Esto evita cargar todo el archivo
// en memoria RAM.
//
// Las líneas nuevas se envían al canal `out`. El monitoreo se detiene cuando
// se cierra el canal `done`.
func TailFile(path string, out chan<- string, done <-chan struct{}, pollInterval time.Duration) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("abriendo archivo de log '%s': %w", path, err)
	}
	defer file.Close()

	// Posicionarse al final del archivo para leer solo líneas nuevas.
	// Esto simula el comportamiento de `tail -f`.
	_, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("seeking al final del archivo: %w", err)
	}

	reader := bufio.NewReader(file)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			// Señal de shutdown: leer cualquier línea pendiente antes de salir
			drainRemaining(reader, out)
			return nil
		case <-ticker.C:
			// Cada tick intenta leer nuevas líneas disponibles
			if err := readNewLines(reader, out); err != nil {
				if err == io.EOF {
					// No hay nuevas líneas, esperar al siguiente tick
					continue
				}
				return fmt.Errorf("leyendo nuevas líneas: %w", err)
			}
		}
	}
}

// readNewLines lee todas las líneas completas disponibles en el reader
// y las envía al canal. Retorna io.EOF si no hay datos nuevos.
func readNewLines(reader *bufio.Reader, out chan<- string) error {
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			// Enviar la línea al canal (sin el newline final)
			if line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			// También quitar \r si el archivo usa CRLF (Windows)
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 {
				out <- line
			}
		}
		if err != nil {
			return err
		}
	}
}

// drainRemaining lee todas las líneas restantes del reader sin bloquearse.
// Se usa durante el shutdown para no perder líneas pendientes.
func drainRemaining(reader *bufio.Reader, out chan<- string) {
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 {
				out <- line
			}
		}
		if err != nil {
			return
		}
	}
}

// ParseLogLine intenta extraer campos estructurados de una línea de log
// en formato combinado de Apache/Nginx. Ejemplo de línea:
//
//	192.168.1.100 - - [30/Jun/2025:10:15:33 -0500] "GET /login HTTP/1.1" 200 1234
//
// Si el parseo falla, retorna una LogEntry con solo la línea cruda.
func ParseLogLine(line string, lineNum int64) LogEntry {
	entry := LogEntry{
		Raw:     line,
		LineNum: lineNum,
	}

	// Intentar parsear formato combinado de Apache/Nginx
	// Formato: IP - - [timestamp] "METHOD PATH PROTOCOL" STATUS SIZE
	fields := splitLogLine(line)

	if len(fields) >= 1 {
		entry.IP = fields[0]
	}
	if len(fields) >= 4 {
		entry.Method = fields[1]
		entry.Path = fields[2]
	}
	if len(fields) >= 5 {
		entry.Status = fields[4]
	}

	return entry
}

// splitLogLine extrae los campos clave de una línea de log combinado.
// Retorna: [IP, method, path, protocol, status]
func splitLogLine(line string) []string {
	const (
		stateInit = iota
		stateInQuote
		stateAfterQuote
	)

	result := make([]string, 0, 4)
	current := make([]byte, 0, 64)
	state := stateInit

	for i := 0; i < len(line); i++ {
		c := line[i]

		switch state {
		case stateInit:
			if c == ' ' {
				if len(current) > 0 {
					result = append(result, string(current))
					current = current[:0]
					// Después de la IP, saltamos hasta la comilla de la request
					// Buscamos el primer '"' que indica el inicio del request
					for i < len(line) && line[i] != '"' {
						i++
					}
					if i < len(line) {
						state = stateInQuote
					}
				}
			} else {
				current = append(current, c)
			}

		case stateInQuote:
			if c == '"' {
				// Fin del request "METHOD PATH PROTOCOL"
				state = stateAfterQuote
				// Parsear "METHOD PATH PROTOCOL"
				req := string(current)
				current = current[:0]
				parts := splitRequest(req)
				result = append(result, parts...)
			} else {
				current = append(current, c)
			}

		case stateAfterQuote:
			if c == ' ' {
				if len(current) > 0 {
					result = append(result, string(current))
					current = current[:0]
					return result
				}
			} else {
				current = append(current, c)
			}
		}
	}

	// Capturar último campo si no terminó con espacio
	if len(current) > 0 && state == stateAfterQuote {
		result = append(result, string(current))
	}

	return result
}

// splitRequest divide "METHOD PATH PROTOCOL" en [method, path]
func splitRequest(req string) []string {
	result := make([]string, 0, 2)
	start := 0

	for i := 0; i < len(req); i++ {
		if req[i] == ' ' {
			if i > start {
				result = append(result, req[start:i])
			}
			start = i + 1
		}
	}
	// Capturar el path (todo después del METHOD hasta el final)
	if start < len(req) {
		// Extraer solo el path, sin el protocolo
		pathPart := req[start:]
		for i := 0; i < len(pathPart); i++ {
			if pathPart[i] == ' ' {
				pathPart = pathPart[:i]
				break
			}
		}
		result = append(result, pathPart)
	}

	return result
}

// ReadExistingLines lee un archivo completo desde el inicio hasta el final
// y envía cada línea al canal. Útil para procesar contenido existente
// antes de cambiar al modo tail.
func ReadExistingLines(path string, out chan<- string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("abriendo archivo de log '%s': %w", path, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 {
				out <- line
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("leyendo líneas existentes: %w", err)
		}
	}
}