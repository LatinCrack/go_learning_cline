# Proyecto 02 — Analizador de Logs de Seguridad en Tiempo Real

> **Dificultad:** ⭐⭐⭐⭐ Avanzado  
> **Paquetes Go:** `os`, `io`, `regexp`, `sync`, `time`, `flag`, `fmt`, `sort`, `strings`, `strconv`, `sync/atomic`, `os/signal`, `syscall`  
> **Conceptos clave:** Tail de archivos, regex compilados, worker pools con canales, `sync.RWMutex`, concurrencia segura, señales OS

## Descripción

Herramienta CLI que monitorea archivos de log en tiempo real (estilo `tail -f`) y detecta patrones de ataques de seguridad usando expresiones regulares. Implementa un pool de workers concurrentes que analizan las líneas contra múltiples reglas de detección simultáneamente.

## Arquitectura

```
┌─────────────────┐     canal líneas     ┌──────────────────┐     canal alertas     ┌──────────────┐
│   parser.go     │ ──────────────────▶  │   detector.go    │ ──────────────────▶  │  alerter.go  │
│  (Tail -f)      │     chan string       │  (Worker Pool)   │   chan ThreatAlert   │  (Consola)   │
│                 │                       │                  │                      │              │
│  os.Open + Seek │                       │  regexp.Compile  │                      │  ANSI Colors │
│  Ticker poll    │                       │  sync.WaitGroup  │                      │  Resumen     │
└─────────────────┘                       └────────┬─────────┘                      └──────────────┘
                                                   │
                                                   ▼
                                          ┌──────────────────┐
                                          │    stats.go      │
                                          │  (Thread-Safe)   │
                                          │                  │
                                          │  sync.RWMutex    │
                                          │  sync/atomic     │
                                          └──────────────────┘
```

## Estructura de Archivos

| Archivo             | Descripción                                                        |
|---------------------|--------------------------------------------------------------------|
| `main.go`           | CLI, orquestación de componentes, manejo de señales                |
| `parser.go`         | Tail de archivos con `os.Seek` + Ticker, parseo de líneas de log   |
| `detector.go`       | Motor de detección con regex compilados y pool de workers          |
| `alerter.go`        | Presentación de alertas en consola con colores ANSI y resumen      |
| `stats.go`          | Estadísticas thread-safe con `sync.RWMutex` y `sync/atomic`        |
| `detector_test.go`  | Suite completa de tests unitarios y de integración                 |
| `go.mod`            | Definición del módulo Go                                           |

## Expresiones Regulares Utilizadas

El motor de detección compila **15 patrones** organizados por categoría:

### SQL Injection (CRITICAL / HIGH)

| Patrón               | Regex                                                                              | Severidad |
|-----------------------|------------------------------------------------------------------------------------|-----------|
| `SQL_INJECTION_OR`   | `(?i)(\bOR\b\s+\d+\s*=\s*\d+\|\bOR\b\s+['"]?\w+['"]?\s*=\s*['"]?\w+['"]?)`       | CRITICAL  |
| `SQL_INJECTION_UNION`| `(?i)(\bUNION\b\s+\bSELECT\b\|\bUNION\b\s+\bALL\b\s+\bSELECT\b)`                  | CRITICAL  |
| `SQL_INJECTION_COMMENT`| `(?i)(--\s*$\|/\*.*\*/\|;\s*--)`                                                | HIGH      |
| `SQL_INJECTION_DROP` | `(?i)(;\s*\bDROP\b\s+\bTABLE\b\|\bDROP\b\s+\bDATABASE\b)`                         | CRITICAL  |
| `SQL_INJECTION_SLEEP`| `(?i)(\bSLEEP\s*\(\|\bBENCHMARK\s*\(\|\bWAITFOR\b\s+\bDELAY\b)`                   | HIGH      |

### Brute Force / Login (MEDIUM / HIGH)

| Patrón               | Regex                                                                              | Severidad |
|-----------------------|------------------------------------------------------------------------------------|-----------|
| `BRUTE_FORCE_LOGIN`  | `(?i)(POST\s+/?(login\|signin\|auth\|admin/login\|wp-login\|...))`                 | MEDIUM    |
| `BRUTE_FORCE_RAPID`  | `(?i)(POST\s+/?(login\|signin\|auth).*)\s+[45]\d{2}\s`                             | HIGH      |

### Directory Scanning / Path Traversal (MEDIUM → CRITICAL)

| Patrón               | Regex                                                                              | Severidad |
|-----------------------|------------------------------------------------------------------------------------|-----------|
| `DIR_SCAN_ENV`       | `(?i)(GET\|HEAD)\s+/\.(env\|git\|svn\|htaccess\|htpasswd\|DS_Store)`              | HIGH      |
| `DIR_SCAN_ADMIN`     | `(?i)(GET\|HEAD)\s+/(admin\|administrator\|phpmyadmin\|cpanel\|wp-admin\|...)`     | MEDIUM    |
| `DIR_SCAN_CONFIG`    | `(?i)(GET\|HEAD)\s+/((web\.xml\|config\.php\|settings\.py\|...))`                  | HIGH      |
| `PATH_TRAVERSAL`     | `(\.\./\|\.\.\\%7C%2e%2e%2f%7C...)`                                                | CRITICAL  |

### XSS / Command Injection / Scanners (HIGH → CRITICAL)

| Patrón               | Regex                                                                              | Severidad |
|-----------------------|------------------------------------------------------------------------------------|-----------|
| `XSS_SCRIPT_TAG`     | `(?i)(<script[\s>]\|javascript\s*:\|on\w+\s*=)`                                    | HIGH      |
| `CMD_INJECTION`      | `(?i)(;\s*(ls\|cat\|whoami\|id\|uname\|wget\|curl\|nc\|bash\|sh\|cmd\|...))`      | CRITICAL  |
| `SCANNER_BOTS`       | `(?i)(nikto\|nmap\|sqlmap\|dirbuster\|gobuster\|wfuzz\|burp\|acunetix\|nessus\|...)` | HIGH   |
| `INFO_DISCLOSURE`    | `(?i)(/server-info\|/server-status\|/phpinfo\|/info\.php\|/debug\|/trace\|...)`    | MEDIUM    |

## Cómo Compilar y Ejecutar

```bash
cd bonus-proyectos/02-log-analyzer

# Compilar
go build -o log-analyzer .

# Ejecutar monitoreando un archivo de log
./log-analyzer -file access.log

# Ver patrones disponibles
./log-analyzer -patterns

# Con opciones avanzadas
./log-analyzer -file access.log -workers 8 -severity high -poll 100ms -existing
```

## Cómo Probar Inyectando Líneas Falsas al Log

### Paso 1: Crear el archivo de log de prueba

```bash
# Crear archivo con tráfico legítimo
cat > test-access.log << 'EOF'
192.168.1.1 - - [30/Jun/2025:10:00:00 -0500] "GET /index.html HTTP/1.1" 200 4567
192.168.1.2 - - [30/Jun/2025:10:00:01 -0500] "GET /about HTTP/1.1" 200 2345
192.168.1.3 - - [30/Jun/2025:10:00:02 -0500] "GET /contact HTTP/1.1" 200 1234
EOF
```

### Paso 2: Iniciar el analizador

```bash
./log-analyzer -file test-access.log -severity low
```

### Paso 3: Inyectar líneas de ataque (en otra terminal)

```bash
# SQL Injection
echo '192.168.1.100 - - [30/Jun/2025:10:15:33 -0500] "GET /login?user=admin'"'"' OR 1=1-- HTTP/1.1" 200 1234' >> test-access.log

# UNION SELECT
echo '10.0.0.1 - - [30/Jun/2025:10:16:00 -0500] "GET /search?q='"'"' UNION SELECT password FROM users-- HTTP/1.1" 200 5678' >> test-access.log

# Escaneo de .env
echo '45.33.32.156 - - [30/Jun/2025:11:00:00 -0500] "GET /.env HTTP/1.1" 404 196' >> test-access.log

# Escaneo de .git
echo '45.33.32.156 - - [30/Jun/2025:11:00:01 -0500] "GET /.git/config HTTP/1.1" 404 196' >> test-access.log

# Brute force login
echo '192.168.1.100 - - [30/Jun/2025:12:00:00 -0500] "POST /login HTTP/1.1" 401 1234' >> test-access.log
echo '192.168.1.100 - - [30/Jun/2025:12:00:01 -0500] "POST /login HTTP/1.1" 401 1234' >> test-access.log
echo '192.168.1.100 - - [30/Jun/2025:12:00:02 -0500] "POST /login HTTP/1.1" 401 1234' >> test-access.log

# Path traversal
echo '10.10.10.10 - - [30/Jun/2025:11:03:00 -0500] "GET /../../etc/passwd HTTP/1.1" 400 0' >> test-access.log

# XSS
echo '10.0.0.1 - - [30/Jun/2025:13:00:00 -0500] "GET /search?q=<script>alert(1)</script> HTTP/1.1" 200 100' >> test-access.log

# Command injection
echo '10.0.0.1 - - [30/Jun/2025:14:00:00 -0500] "GET /api?cmd=test;cat /etc/passwd HTTP/1.1" 200 100' >> test-access.log

# Scanner bot
echo '192.168.1.1 - - [30/Jun/2025:15:00:00 -0500] "GET / HTTP/1.1" 200 1000 "nikto/2.1.6"' >> test-access.log

# DROP TABLE
echo '172.16.0.5 - - [30/Jun/2025:10:17:00 -0500] "GET /api?id=1; DROP TABLE users-- HTTP/1.1" 500 0' >> test-access.log
```

### Script rápido de simulación masiva

```bash
#!/bin/bash
# simulate-attack.sh — Simula un ataque completo al log
LOG="test-access.log"

echo "[*] Iniciando simulación de ataque..."
sleep 2

# Ráfaga SQL Injection
for i in $(seq 1 5); do
  echo "192.168.1.$((100+i)) - - [30/Jun/2025:10:$((15+i)):00 -0500] \"GET /login?user=admin' OR 1=1-- HTTP/1.1\" 200 1234" >> $LOG
  sleep 0.3
done

# Escaneo de directorios
for path in .env .git/config .htaccess wp-config.php phpmyadmin; do
  echo "45.33.32.156 - - [30/Jun/2025:11:00:00 -0500] \"GET /$path HTTP/1.1\" 404 196" >> $LOG
  sleep 0.2
done

# Brute force
for i in $(seq 1 10); do
  echo "10.0.0.99 - - [30/Jun/2025:12:00:$i -0500] \"POST /login HTTP/1.1\" 401 1234" >> $LOG
  sleep 0.1
done

# Path traversal + XSS + Command injection
echo '10.10.10.10 - - [30/Jun/2025:13:00:00 -0500] "GET /../../etc/passwd HTTP/1.1" 400 0' >> $LOG
echo '10.0.0.1 - - [30/Jun/2025:13:01:00 -0500] "GET /search?q=<script>alert(1)</script> HTTP/1.1" 200 100' >> $LOG
echo '10.0.0.1 - - [30/Jun/2025:13:02:00 -0500] "GET /api?cmd=test;cat /etc/passwd HTTP/1.1" 200 100' >> $LOG

# Scanner bots
echo '192.168.1.1 - - [30/Jun/2025:15:00:00 -0500] "GET / HTTP/1.1" 200 1000 "sqlmap/1.5"' >> $LOG
echo '192.168.1.2 - - [30/Jun/2025:15:01:00 -0500] "GET / HTTP/1.1" 200 1000 "nmap NSE"' >> $LOG

echo "[✓] Simulación completada. Revisa la consola del analizador."
```

## Ejecutar Tests

```bash
cd bonus-proyectos/02-log-analyzer

# Ejecutar todos los tests
go test -v

# Ejecutar tests con detector de race conditions
go test -race -v

# Ejecutar solo tests de detección SQL
go test -v -run TestSQLInjection

# Ejecutar solo tests de concurrencia
go test -v -run TestStatsConcurrency
```

## Opciones CLI

| Flag         | Default   | Descripción                                              |
|--------------|-----------|----------------------------------------------------------|
| `-file`      | (requerido) | Ruta al archivo de log a monitorear                   |
| `-workers`   | `4`       | Número de workers de detección concurrentes              |
| `-poll`      | `500ms`   | Intervalo de polling para detectar nuevas líneas         |
| `-severity`  | `low`     | Severidad mínima para mostrar alertas                    |
| `-raw`       | `true`    | Mostrar línea cruda del log en la alerta                 |
| `-existing`  | `false`   | Analizar contenido existente antes de tail               |
| `-patterns`  | `false`   | Listar todos los patrones de detección y salir           |
| `-version`   | `false`   | Mostrar información de versión                           |

## Detención Graceful

Presiona `Ctrl+C` para detener el analizador de forma segura. El sistema:

1. Recibe la señal `SIGINT`/`SIGTERM`
2. Detiene el lector de archivos (tail)
3. Los workers procesan las líneas pendientes en el canal
4. Se genera el reporte resumen final con estadísticas completas
5. Se imprime el resumen con distribución por severidad, top ataques y top IPs

## Diseño de Concurrencia

```
         ┌─────────────────────────────────────────────┐
         │              main goroutine                  │
         │  1. Inicia TailFile (producer)               │
         │  2. Inicia N workers (consumers)             │
         │  3. Inicia Alerter (consumer de alertas)     │
         │  4. Espera señal SIGINT                      │
         └─────────────────────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
   ┌──────────┐     ┌──────────┐     ┌──────────┐
   │ Worker 1 │     │ Worker 2 │     │ Worker N │
   │          │     │          │     │          │
   │ Lee de   │     │ Lee de   │     │ Lee de   │
   │ linesChan│     │ linesChan│     │ linesChan│
   │          │     │          │     │          │
   │ Regex ×15│     │ Regex ×15│     │ Regex ×15│
   │          │     │          │     │          │
   │ stats.mu │     │ stats.mu │     │ stats.mu │
   │ .Lock()  │     │ .Lock()  │     │ .Lock()  │
   │ Write    │     │ Write    │     │ Write    │
   │ .Unlock()│     │ .Unlock()│     │ .Unlock()│
   └────┬─────┘     └────┬─────┘     └────┬─────┘
        │                │                │
        └────────────────┼────────────────┘
                         ▼
                  ┌──────────────┐
                  │  alertsChan  │
                  └──────┬───────┘
                         ▼
                  ┌──────────────┐
                  │   Alerter    │
                  │  (1 goroutina│
                  │   consol)   │
                  └──────────────┘
```

**Seguridad contra Race Conditions:**
- `Stats.totalLines` → `sync/atomic.Int64` (sin bloqueo)
- `Stats.totalAttacks` → `sync/atomic.Int64` (sin bloqueo)
- `Stats.attackCounts/ipCounts/bySeverity` → `sync.RWMutex` (write lock para updates, read lock para snapshots)
- Canales con buffer para desacoplar productores y consumidores