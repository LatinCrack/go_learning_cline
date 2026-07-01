package main

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// ── Tests de Detección de Patrones ─────────────────────────────

func TestSQLInjectionDetection(t *testing.T) {
	stats := NewStats()
	engine := NewDetectorEngine(DetectorConfig{Workers: 1, ChannelBuf: 100}, stats)

	tests := []struct {
		name     string
		line     string
		wantFind bool
		pattern  string
	}{
		{
			name:     "SQL Injection OR 1=1",
			line:     `192.168.1.100 - - [30/Jun/2025:10:15:33 -0500] "GET /login?user=admin' OR 1=1-- HTTP/1.1" 200 1234`,
			wantFind: true,
			pattern:  "SQL_INJECTION_OR",
		},
		{
			name:     "SQL Injection UNION SELECT",
			line:     `10.0.0.1 - - [30/Jun/2025:10:16:00 -0500] "GET /search?q=' UNION SELECT password FROM users-- HTTP/1.1" 200 5678`,
			wantFind: true,
			pattern:  "SQL_INJECTION_UNION",
		},
		{
			name:     "SQL Injection DROP TABLE",
			line:     `172.16.0.5 - - [30/Jun/2025:10:17:00 -0500] "GET /api?id=1; DROP TABLE users-- HTTP/1.1" 500 0`,
			wantFind: true,
			pattern:  "SQL_INJECTION_DROP",
		},
		{
			name:     "SQL Injection SLEEP (Blind)",
			line:     `192.168.1.50 - - [30/Jun/2025:10:18:00 -0500] "GET /api?id=1 AND SLEEP(5) HTTP/1.1" 200 100`,
			wantFind: true,
			pattern:  "SQL_INJECTION_SLEEP",
		},
		{
			name:     "Normal request - no injection",
			line:     `192.168.1.1 - - [30/Jun/2025:10:00:00 -0500] "GET /index.html HTTP/1.1" 200 4567`,
			wantFind: false,
			pattern:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := collectAlerts(engine, []string{tt.line})

			if tt.wantFind {
				if len(alerts) == 0 {
					t.Errorf("esperaba detectar '%s' pero no se generaron alertas", tt.pattern)
					return
				}
				found := false
				for _, a := range alerts {
					if a.Pattern == tt.pattern {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("esperaba patrón '%s', alertas generadas: %v", tt.pattern, alertNames(alerts))
				}
			} else {
				if len(alerts) > 0 {
					t.Errorf("no esperaba alertas para línea normal, pero se generaron: %v", alertNames(alerts))
				}
			}
		})
	}
}

func TestDirectoryScanningDetection(t *testing.T) {
	stats := NewStats()
	engine := NewDetectorEngine(DetectorConfig{Workers: 1, ChannelBuf: 100}, stats)

	tests := []struct {
		name    string
		line    string
		pattern string
	}{
		{
			name:    "Escaneo de .env",
			line:    `45.33.32.156 - - [30/Jun/2025:11:00:00 -0500] "GET /.env HTTP/1.1" 404 196`,
			pattern: "DIR_SCAN_ENV",
		},
		{
			name:    "Escaneo de .git",
			line:    `45.33.32.156 - - [30/Jun/2025:11:00:01 -0500] "GET /.git/config HTTP/1.1" 404 196`,
			pattern: "DIR_SCAN_ENV",
		},
		{
			name:    "Escaneo de phpmyadmin",
			line:    `185.220.101.1 - - [30/Jun/2025:11:01:00 -0500] "GET /phpmyadmin HTTP/1.1" 404 196`,
			pattern: "DIR_SCAN_ADMIN",
		},
		{
			name:    "Escaneo de wp-config.php",
			line:    `185.220.101.1 - - [30/Jun/2025:11:02:00 -0500] "GET /wp-config.php HTTP/1.1" 403 0`,
			pattern: "DIR_SCAN_CONFIG",
		},
		{
			name:    "Path traversal",
			line:    `10.10.10.10 - - [30/Jun/2025:11:03:00 -0500] "GET /../../etc/passwd HTTP/1.1" 400 0`,
			pattern: "PATH_TRAVERSAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := collectAlerts(engine, []string{tt.line})

			if len(alerts) == 0 {
				t.Errorf("esperaba detectar '%s' pero no se generaron alertas", tt.pattern)
				return
			}

			found := false
			for _, a := range alerts {
				if a.Pattern == tt.pattern {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("esperaba patrón '%s', alertas generadas: %v", tt.pattern, alertNames(alerts))
			}
		})
	}
}

func TestBruteForceDetection(t *testing.T) {
	stats := NewStats()
	engine := NewDetectorEngine(DetectorConfig{Workers: 1, ChannelBuf: 100}, stats)

	lines := []string{
		`192.168.1.100 - - [30/Jun/2025:12:00:00 -0500] "POST /login HTTP/1.1" 401 1234`,
		`192.168.1.100 - - [30/Jun/2025:12:00:01 -0500] "POST /login HTTP/1.1" 401 1234`,
		`192.168.1.100 - - [30/Jun/2025:12:00:02 -0500] "POST /login HTTP/1.1" 401 1234`,
	}

	alerts := collectAlerts(engine, lines)

	if len(alerts) == 0 {
		t.Fatal("esperaba detectar intentos de brute force en login")
	}

	// Verificar que se detectaron los POST a /login
	loginAlerts := 0
	for _, a := range alerts {
		if a.Pattern == "BRUTE_FORCE_LOGIN" {
			loginAlerts++
		}
	}

	if loginAlerts != len(lines) {
		t.Errorf("esperaba %d alertas de brute force, obtuvo %d", len(lines), loginAlerts)
	}
}

func TestXSSDetection(t *testing.T) {
	stats := NewStats()
	engine := NewDetectorEngine(DetectorConfig{Workers: 1, ChannelBuf: 100}, stats)

	tests := []struct {
		name    string
		line    string
		pattern string
	}{
		{
			name:    "Script tag injection",
			line:    `10.0.0.1 - - [30/Jun/2025:13:00:00 -0500] "GET /search?q=<script>alert('xss')</script> HTTP/1.1" 200 100`,
			pattern: "XSS_SCRIPT_TAG",
		},
		{
			name:    "Event handler injection",
			line:    `10.0.0.1 - - [30/Jun/2025:13:01:00 -0500] "GET /profile?name=test onload=alert(1) HTTP/1.1" 200 100`,
			pattern: "XSS_SCRIPT_TAG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := collectAlerts(engine, []string{tt.line})

			if len(alerts) == 0 {
				t.Errorf("esperaba detectar '%s' pero no se generaron alertas", tt.pattern)
				return
			}

			found := false
			for _, a := range alerts {
				if a.Pattern == tt.pattern {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("esperaba patrón '%s', alertas generadas: %v", tt.pattern, alertNames(alerts))
			}
		})
	}
}

func TestCommandInjectionDetection(t *testing.T) {
	stats := NewStats()
	engine := NewDetectorEngine(DetectorConfig{Workers: 1, ChannelBuf: 100}, stats)

	line := `10.0.0.1 - - [30/Jun/2025:14:00:00 -0500] "GET /api?cmd=test;cat /etc/passwd HTTP/1.1" 200 100`
	alerts := collectAlerts(engine, []string{line})

	if len(alerts) == 0 {
		t.Fatal("esperaba detectar command injection")
	}

	found := false
	for _, a := range alerts {
		if a.Pattern == "CMD_INJECTION" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("esperaba patrón CMD_INJECTION, alertas generadas: %v", alertNames(alerts))
	}
}

func TestScannerBotDetection(t *testing.T) {
	stats := NewStats()
	engine := NewDetectorEngine(DetectorConfig{Workers: 1, ChannelBuf: 100}, stats)

	lines := []string{
		`192.168.1.1 - - [30/Jun/2025:15:00:00 -0500] "GET / HTTP/1.1" 200 1000 "nikto/2.1.6"`,
		`192.168.1.2 - - [30/Jun/2025:15:01:00 -0500] "GET / HTTP/1.1" 200 1000 "sqlmap/1.5"`,
		`192.168.1.3 - - [30/Jun/2025:15:02:00 -0500] "GET / HTTP/1.1" 200 1000 "nmap NSE"`,
	}

	alerts := collectAlerts(engine, lines)

	botAlerts := 0
	for _, a := range alerts {
		if a.Pattern == "SCANNER_BOTS" {
			botAlerts++
		}
	}

	if botAlerts != len(lines) {
		t.Errorf("esperaba %d detecciones de scanner bots, obtuvo %d", len(lines), botAlerts)
	}
}

// ── Tests de Parser ────────────────────────────────────────────

func TestParseLogLine(t *testing.T) {
	line := `192.168.1.100 - - [30/Jun/2025:10:15:33 -0500] "GET /login HTTP/1.1" 200 1234`
	entry := ParseLogLine(line, 1)

	if entry.IP != "192.168.1.100" {
		t.Errorf("IP: esperaba '192.168.1.100', obtuvo '%s'", entry.IP)
	}
	if entry.Method != "GET" {
		t.Errorf("Method: esperaba 'GET', obtuvo '%s'", entry.Method)
	}
	if entry.Path != "/login" {
		t.Errorf("Path: esperaba '/login', obtuvo '%s'", entry.Path)
	}
	if entry.Status != "200" {
		t.Errorf("Status: esperaba '200', obtuvo '%s'", entry.Status)
	}
	if entry.LineNum != 1 {
		t.Errorf("LineNum: esperaba 1, obtuvo %d", entry.LineNum)
	}
}

func TestParseLogLineMalformed(t *testing.T) {
	line := "esto no es un log válido"
	entry := ParseLogLine(line, 42)

	if entry.Raw != line {
		t.Errorf("Raw: esperaba '%s', obtuvo '%s'", line, entry.Raw)
	}
	if entry.LineNum != 42 {
		t.Errorf("LineNum: esperaba 42, obtuvo %d", entry.LineNum)
	}
}

// ── Tests de Stats (thread-safety) ─────────────────────────────

func TestStatsConcurrency(t *testing.T) {
	stats := NewStats()
	var wg sync.WaitGroup

	// Lanzar 100 goroutines escribiendo concurrentemente
	numGoroutines := 100
	attacksPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < attacksPerGoroutine; j++ {
				stats.IncrementLines()
				stats.RecordAttack("TEST_ATTACK", SeverityHigh, "10.0.0.1")
			}
		}(i)
	}

	wg.Wait()

	expectedLines := int64(numGoroutines * attacksPerGoroutine)
	if stats.GetTotalLines() != expectedLines {
		t.Errorf("TotalLines: esperaba %d, obtuvo %d", expectedLines, stats.GetTotalLines())
	}
	if stats.GetTotalAttacks() != expectedLines {
		t.Errorf("TotalAttacks: esperaba %d, obtuvo %d", expectedLines, stats.GetTotalAttacks())
	}
	if stats.GetAttackCount("TEST_ATTACK") != expectedLines {
		t.Errorf("AttackCount: esperaba %d, obtuvo %d", expectedLines, stats.GetAttackCount("TEST_ATTACK"))
	}
	if stats.GetSeverityCount(SeverityHigh) != expectedLines {
		t.Errorf("SeverityHigh: esperaba %d, obtuvo %d", expectedLines, stats.GetSeverityCount(SeverityHigh))
	}
	if stats.GetIPCount("10.0.0.1") != expectedLines {
		t.Errorf("IPCount: esperaba %d, obtuvo %d", expectedLines, stats.GetIPCount("10.0.0.1"))
	}
}

func TestStatsSnapshot(t *testing.T) {
	stats := NewStats()

	stats.IncrementLinesBy(100)
	stats.RecordAttack("SQL_INJECTION_OR", SeverityCritical, "192.168.1.1")
	stats.RecordAttack("SQL_INJECTION_OR", SeverityCritical, "192.168.1.1")
	stats.RecordAttack("DIR_SCAN_ENV", SeverityHigh, "10.0.0.1")
	stats.RecordAttack("BRUTE_FORCE_LOGIN", SeverityMedium, "192.168.1.1")

	snapshot := stats.Snapshot()

	if snapshot.TotalLines != 100 {
		t.Errorf("Snapshot.TotalLines: esperaba 100, obtuvo %d", snapshot.TotalLines)
	}
	if snapshot.TotalAttacks != 4 {
		t.Errorf("Snapshot.TotalAttacks: esperaba 4, obtuvo %d", snapshot.TotalAttacks)
	}
	if snapshot.UniqueIPs != 2 {
		t.Errorf("Snapshot.UniqueIPs: esperaba 2, obtuvo %d", snapshot.UniqueIPs)
	}
	if len(snapshot.TopAttacks) == 0 {
		t.Error("Snapshot.TopAttacks no debería estar vacío")
	}

	// El primer ataque debería ser SQL_INJECTION_OR (2 detecciones)
	if snapshot.TopAttacks[0].Name != "SQL_INJECTION_OR" || snapshot.TopAttacks[0].Count != 2 {
		t.Errorf("TopAttacks[0]: esperaba SQL_INJECTION_OR con 2, obtuvo %s con %d",
			snapshot.TopAttacks[0].Name, snapshot.TopAttacks[0].Count)
	}
}

// ── Tests de Severity ──────────────────────────────────────────

func TestSeverityString(t *testing.T) {
	tests := []struct {
		sev  Severity
		want string
	}{
		{SeverityLow, "LOW"},
		{SeverityMedium, "MEDIUM"},
		{SeverityHigh, "HIGH"},
		{SeverityCritical, "CRITICAL"},
		{Severity(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.sev.String(); got != tt.want {
			t.Errorf("Severity(%d).String(): esperaba '%s', obtuvo '%s'", tt.sev, tt.want, got)
		}
	}
}

// ── Tests de Tail File ─────────────────────────────────────────

func TestTailFile(t *testing.T) {
	// Crear archivo temporal con contenido inicial
	tmpFile, err := os.CreateTemp("", "log-analyzer-test-*.log")
	if err != nil {
		t.Fatalf("error creando archivo temporal: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Abrir en modo append y escribir líneas iniciales
	f, _ := os.OpenFile(tmpFile.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("línea inicial 1\n")
	f.WriteString("línea inicial 2\n")
	f.Close()

	// Canal y done para tail
	lines := make(chan string, 100)
	done := make(chan struct{})

	// Iniciar tail en background
	go func() {
		TailFile(tmpFile.Name(), lines, done, 50*time.Millisecond)
	}()

	// Dar tiempo a que inicie el tail
	time.Sleep(200 * time.Millisecond)

	// Agregar nuevas líneas al archivo
	f, _ = os.OpenFile(tmpFile.Name(), os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("nueva línea ataque\n")
	f.WriteString("otra nueva línea\n")
	f.Close()

	// Esperar a que se lean las nuevas líneas
	timeout := time.After(3 * time.Second)
	received := 0
loop:
	for {
		select {
		case <-lines:
			received++
			if received >= 2 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	// Cerrar tail
	close(done)

	if received < 2 {
		t.Errorf("esperaba recibir al menos 2 líneas nuevas, recibió %d", received)
	}
}

func TestReadExistingLines(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "log-existing-*.log")
	if err != nil {
		t.Fatalf("error creando archivo temporal: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString("línea 1\nlínea 2\nlínea 3\n")
	tmpFile.Close()

	lines := make(chan string, 100)
	go func() {
		ReadExistingLines(tmpFile.Name(), lines)
		close(lines)
	}()

	collected := make([]string, 0)
	for line := range lines {
		collected = append(collected, line)
	}

	if len(collected) != 3 {
		t.Errorf("esperaba 3 líneas, obtuvo %d", len(collected))
	}
	if collected[0] != "línea 1" {
		t.Errorf("primera línea: esperaba 'línea 1', obtuvo '%s'", collected[0])
	}
}

// ── Tests de Alerter ───────────────────────────────────────────

func TestAlerterSeverityFilter(t *testing.T) {
	// Alerter que solo muestra HIGH y CRITICAL
	var buf strings.Builder
	alerter := NewAlerter(AlerterConfig{
		MinSeverity: SeverityHigh,
		ShowRaw:     true,
		Writer:      &buf,
	})

	alertsChan := make(chan ThreatAlert, 10)
	done := make(chan struct{})

	go alerter.Start(alertsChan, done)

	// Enviar alertas de diferentes severidades
	alertsChan <- ThreatAlert{Pattern: "LOW_TEST", Severity: SeverityLow, Line: "low", LineNum: 1}
	alertsChan <- ThreatAlert{Pattern: "MEDIUM_TEST", Severity: SeverityMedium, Line: "medium", LineNum: 2}
	alertsChan <- ThreatAlert{Pattern: "HIGH_TEST", Severity: SeverityHigh, Line: "high", LineNum: 3}
	alertsChan <- ThreatAlert{Pattern: "CRITICAL_TEST", Severity: SeverityCritical, Line: "critical", LineNum: 4}

	close(alertsChan)
	<-done

	output := buf.String()

	// Solo HIGH y CRITICAL deberían aparecer
	if !strings.Contains(output, "HIGH_TEST") {
		t.Error("esperaba que HIGH_TEST apareciera en la salida")
	}
	if !strings.Contains(output, "CRITICAL_TEST") {
		t.Error("esperaba que CRITICAL_TEST apareciera en la salida")
	}

	// LOW y MEDIUM NO deberían aparecer
	if strings.Contains(output, "LOW_TEST") {
		t.Error("LOW_TEST no debería aparecer con filtro MinSeverity=HIGH")
	}
	if strings.Contains(output, "MEDIUM_TEST") {
		t.Error("MEDIUM_TEST no debería aparecer con filtro MinSeverity=HIGH")
	}

	if alerter.GetAlertCount() != 2 {
		t.Errorf("esperaba 2 alertas, obtuvo %d", alerter.GetAlertCount())
	}
}

// ── Test de Parse Severity ─────────────────────────────────────

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input   string
		want    Severity
		wantErr bool
	}{
		{"low", SeverityLow, false},
		{"medium", SeverityMedium, false},
		{"high", SeverityHigh, false},
		{"critical", SeverityCritical, false},
		{"CRITICAL", SeverityCritical, false},
		{"invalido", SeverityLow, true},
	}

	for _, tt := range tests {
		got, err := parseSeverity(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseSeverity('%s'): error = %v, wantErr = %v", tt.input, err, tt.wantErr)
		}
		if got != tt.want {
			t.Errorf("parseSeverity('%s'): esperaba %v, obtuvo %v", tt.input, tt.want, got)
		}
	}
}

// ── Helpers de testing ─────────────────────────────────────────

// collectAlerts envía líneas al motor de detección y recoge todas las alertas.
func collectAlerts(engine *DetectorEngine, lines []string) []ThreatAlert {
	linesChan := make(chan string, len(lines))
	alertsChan := make(chan ThreatAlert, len(lines)*10) // Buffer generoso para múltiples patrones
	var wg sync.WaitGroup

	// Iniciar workers
	engine.Start(linesChan, alertsChan, &wg)

	// Enviar líneas
	for _, line := range lines {
		linesChan <- line
	}
	close(linesChan)

	// Esperar workers
	wg.Wait()
	close(alertsChan)

	// Recoger alertas
	var alerts []ThreatAlert
	for a := range alertsChan {
		alerts = append(alerts, a)
	}
	return alerts
}

// alertNames retorna los nombres de patrones de una lista de alertas.
func alertNames(alerts []ThreatAlert) []string {
	names := make([]string, len(alerts))
	for i, a := range alerts {
		names[i] = a.Pattern
	}
	return names
}