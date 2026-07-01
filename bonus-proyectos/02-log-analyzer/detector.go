package main

import (
	"regexp"
	"sync"
)

// Severity representa el nivel de severidad de una amenaza detectada.
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String retorna la etiqueta legible del nivel de severidad.
func (s Severity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ColorCode retorna el código ANSI de color para la severidad.
func (s Severity) ColorCode() string {
	switch s {
	case SeverityLow:
		return "\033[33m" // Amarillo
	case SeverityMedium:
		return "\033[35m" // Magenta
	case SeverityHigh:
		return "\033[31m" // Rojo
	case SeverityCritical:
		return "\033[91;1m" // Rojo brillante
	default:
		return "\033[0m"
	}
}

// AttackPattern define una regla de detección con su regex,
// severidad y descripción del ataque.
type AttackPattern struct {
	Name        string         // Nombre descriptivo del ataque
	Regex       *regexp.Regexp // Patrón compilado
	Severity    Severity       // Nivel de severidad
	Description string         // Descripción de lo que detecta
}

// ThreatAlert representa una alerta generada cuando se detecta una amenaza.
type ThreatAlert struct {
	Pattern   string   // Nombre del patrón que coincidió
	Severity  Severity // Nivel de severidad
	Line      string   // Línea original que disparó la alerta
	IP        string   // IP origen
	Path      string   // Ruta solicitada
	Match     string   // Texto que coincidió con el regex
	LineNum   int64    // Número de línea
	Timestamp int64    // Unix timestamp de la detección
}

// DetectorConfig contiene la configuración del motor de detección.
type DetectorConfig struct {
	Workers    int // Número de workers en el pool
	ChannelBuf int // Tamaño del buffer del canal de líneas
}

// DetectorEngine es el motor principal de detección de amenazas.
// Mantiene un pool de workers que procesan líneas concurrentemente
// usando un conjunto de patrones de ataque compilados.
type DetectorEngine struct {
	patterns []AttackPattern
	config   DetectorConfig
	stats    *Stats
}

// NewDetectorEngine crea un nuevo motor de detección con los patrones
// de ataque predefinidos y la configuración especificada.
func NewDetectorEngine(config DetectorConfig, stats *Stats) *DetectorEngine {
	engine := &DetectorEngine{
		patterns: compilePatterns(),
		config:   config,
		stats:    stats,
	}
	return engine
}

// compilePatterns compila todas las expresiones regulares de detección.
// Cada patrón representa un tipo de ataque conocido.
func compilePatterns() []AttackPattern {
	patterns := []struct {
		name        string
		regex       string
		severity    Severity
		description string
	}{
		// ── SQL Injection ──────────────────────────────────────────
		{
			name:        "SQL_INJECTION_OR",
			regex:       `(?i)(\bOR\b\s+\d+\s*=\s*\d+|\bOR\b\s+['"]?\w+['"]?\s*=\s*['"]?\w+['"]?)`,
			severity:    SeverityCritical,
			description: "Detecta intentos de SQL Injection usando OR 1=1 o variantes",
		},
		{
			name:        "SQL_INJECTION_UNION",
			regex:       `(?i)(\bUNION\b\s+\bSELECT\b|\bUNION\b\s+\bALL\b\s+\bSELECT\b)`,
			severity:    SeverityCritical,
			description: "Detecta SQL Injection con UNION SELECT para extracción de datos",
		},
		{
			name:        "SQL_INJECTION_COMMENT",
			regex:       `(?i)(--\s*$|/\*.*\*/|;\s*--)`,
			severity:    SeverityHigh,
			description: "Detecta SQL Injection con comentarios para truncar queries",
		},
		{
			name:        "SQL_INJECTION_DROP",
			regex:       `(?i)(;\s*\bDROP\b\s+\bTABLE\b|\bDROP\b\s+\bDATABASE\b)`,
			severity:    SeverityCritical,
			description: "Detecta intentos de DROP TABLE/DATABASE destructivos",
		},
		{
			name:        "SQL_INJECTION_SLEEP",
			regex:       `(?i)(\bSLEEP\s*\(|\bBENCHMARK\s*\(|\bWAITFOR\b\s+\bDELAY\b)`,
			severity:    SeverityHigh,
			description: "Detecta SQL Injection basado en tiempo (blind SQLi)",
		},

		// ── Brute Force / Login ────────────────────────────────────
		{
			name:        "BRUTE_FORCE_LOGIN",
			regex:       `(?i)(POST\s+/?(login|signin|auth|admin/login|wp-login|user/login|api/auth))`,
			severity:    SeverityMedium,
			description: "Detecta requests POST a endpoints de autenticación",
		},
		{
			name:        "BRUTE_FORCE_RAPID",
			regex:       `(?i)(POST\s+/?(login|signin|auth).*)\s+[45]\d{2}\s`,
			severity:    SeverityHigh,
			description: "Detecta respuestas 4xx/5xx en endpoints de login (posible fuerza bruta)",
		},

		// ── Directory Scanning / Path Traversal ────────────────────
		{
			name:        "DIR_SCAN_ENV",
			regex:       `(?i)(GET|HEAD)\s+/\.(env|git|svn|htaccess|htpasswd|DS_Store)`,
			severity:    SeverityHigh,
			description: "Detecta escaneo de archivos sensibles (.env, .git, .htaccess)",
		},
		{
			name:        "DIR_SCAN_ADMIN",
			regex:       `(?i)(GET|HEAD)\s+/(admin|administrator|phpmyadmin|cpanel|wp-admin|manager|console)`,
			severity:    SeverityMedium,
			description: "Detecta escaneo de paneles de administración",
		},
		{
			name:        "DIR_SCAN_CONFIG",
			regex:       `(?i)(GET|HEAD)\s+/((web\.xml|config\.php|settings\.py|database\.yml|wp-config\.php|\.ssh|\.aws|server-status))`,
			severity:    SeverityHigh,
			description: "Detecta escaneo de archivos de configuración sensibles",
		},
		{
			name:        "PATH_TRAVERSAL",
			regex:       `(\.\./|\.\.\\|%2e%2e%2f|%2e%2e/|\.\.%2f|%2e%2e%5c)`,
			severity:    SeverityCritical,
			description: "Detecta intentos de path traversal (directory traversal attack)",
		},

		// ── XSS (Cross-Site Scripting) ─────────────────────────────
		{
			name:        "XSS_SCRIPT_TAG",
			regex:       `(?i)(<script[\s>]|javascript\s*:|on\w+\s*=)`,
			severity:    SeverityHigh,
			description: "Detecta intentos de XSS con tags script o event handlers",
		},

		// ── Command Injection ──────────────────────────────────────
		{
			name:        "CMD_INJECTION",
			regex:       `(?i)(;\s*(ls|cat|whoami|id|uname|wget|curl|nc|bash|sh|cmd|powershell)|\|\s*(ls|cat|whoami|id|uname))`,
			severity:    SeverityCritical,
			description: "Detecta intentos de command injection OS",
		},

		// ── Scanner / Bot Detection ────────────────────────────────
		{
			name:        "SCANNER_BOTS",
			regex:       `(?i)(nikto|nmap|sqlmap|dirbuster|gobuster|wfuzz|burp|acunetix|nessus|openvas|masscan)`,
			severity:    SeverityHigh,
			description: "Detecta herramientas de reconocimiento y escaneo conocidas",
		},

		// ── Information Disclosure ─────────────────────────────────
		{
			name:        "INFO_DISCLOSURE",
			regex:       `(?i)(/server-info|/server-status|/phpinfo|/info\.php|/test\.php|/debug|/trace|/metrics|/actuator)`,
			severity:    SeverityMedium,
			description: "Detecta acceso a endpoints de divulgación de información",
		},
	}

	result := make([]AttackPattern, 0, len(patterns))
	for _, p := range patterns {
		compiled, err := regexp.Compile(p.regex)
		if err != nil {
			// Si un regex falla al compilar, lo saltamos (no debería ocurrir)
			continue
		}
		result = append(result, AttackPattern{
			Name:        p.name,
			Regex:       compiled,
			Severity:    p.severity,
			Description: p.description,
		})
	}
	return result
}

// Start lanza el pool de workers que leen líneas del canal de entrada,
// las analizan contra todos los patrones y envían alertas al canal de salida.
func (d *DetectorEngine) Start(
	lines <-chan string,
	alerts chan<- ThreatAlert,
	wg *sync.WaitGroup,
) {
	for i := 0; i < d.config.Workers; i++ {
		wg.Add(1)
		go d.worker(i, lines, alerts, wg)
	}
}

// worker procesa líneas del canal de entrada y genera alertas cuando
// detecta coincidencias con los patrones de ataque.
func (d *DetectorEngine) worker(
	id int,
	lines <-chan string,
	alerts chan<- ThreatAlert,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	lineNum := int64(0)
	for line := range lines {
		lineNum++

		// Parsear la línea para extraer campos
		entry := ParseLogLine(line, lineNum)

		// Analizar contra cada patrón de ataque
		for _, pattern := range d.patterns {
			match := pattern.Regex.FindString(entry.Raw)
			if match != "" {
				alert := ThreatAlert{
					Pattern:  pattern.Name,
					Severity: pattern.Severity,
					Line:     entry.Raw,
					IP:       entry.IP,
					Path:     entry.Path,
					Match:    match,
					LineNum:  entry.LineNum,
				}

				// Actualizar estadísticas de forma thread-safe
				d.stats.RecordAttack(pattern.Name, pattern.Severity, entry.IP)

				// Enviar alerta al canal de salida
				alerts <- alert
			}
		}
	}
}

// GetPatterns retorna la lista de patrones compilados (útil para inspección).
func (d *DetectorEngine) GetPatterns() []AttackPattern {
	return d.patterns
}
