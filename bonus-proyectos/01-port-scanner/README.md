<div align="center">

# 🔴 Proyecto 01 — Port Scanner Concurrente de Alta Velocidad

### *Escáner TCP de puertos con pool de workers, CLI profesional y reportes estructurados*

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Concurrencia](https://img.shields.io/badge/Concurrencia-Goroutines-4ECDC4?style=for-the-badge) ![Tests](https://img.shields.io/badge/Tests-19+-FF6B6B?style=for-the-badge) ![Arquitectura](https://img.shields.io/badge/Arquitectura-Worker_Pool-FFE66D&style=for-the-badge)

</div>

---

## 🎯 Objetivo

Construir un **escáner de puertos TCP concurrente** de alta velocidad capaz de escanear miles de puertos en segundos usando goroutines. La herramienta detecta puertos abiertos, identifica servicios conocidos (HTTP, SSH, HTTPS, MySQL, etc.) y genera reportes estructurados en múltiples formatos.

**Inspirado en:** `nmap`, `masscan`, `zmap`

---

## 🏗️ Arquitectura del Pipeline de Escaneo

El escáner implementa el patrón **Worker Pool** (pool de trabajadores) para gestionar la concurrencia de forma controlada, evitando la saturación de file descriptors del sistema operativo.

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PIPELINE DE ESCANEO                              │
│                                                                     │
│   ┌──────────┐     ┌──────────────┐     ┌──────────────────────┐   │
│   │  MAIN    │     │  JOB CHANNEL │     │  WORKER POOL         │   │
│   │          │     │              │     │                      │   │
│   │  CLI     │────▶│  ports       │────▶│  ┌─────────────┐    │   │
│   │  Flags   │     │  (buffered)  │     │  │ Worker 1    │    │   │
│   │          │     │              │     │  │ DialTimeout │──┐ │   │
│   └──────────┘     └──────────────┘     │  └─────────────┘  │ │   │
│                                          │  ┌─────────────┐  │ │   │
│   ┌──────────┐     ┌──────────────┐     │  │ Worker 2    │  │ │   │
│   │  REPORT  │     │  RESULTS     │     │  │ DialTimeout │──┤ │   │
│   │          │◀────│  CHANNEL     │◀────│  └─────────────┘  │ │   │
│   │  Table   │     │  (buffered)  │     │  ...              │ │   │
│   │  JSON    │     │              │     │  ┌─────────────┐  │ │   │
│   │  Summary │     │              │     │  │ Worker N    │  │ │   │
│   └──────────┘     └──────────────┘     │  │ DialTimeout │──┘ │   │
│                                          │  └─────────────┘    │   │
│                                          └──────────────────────┘   │
│                                                                     │
│   Sincronización: sync.WaitGroup                                    │
│   Timeout:        net.DialTimeout("tcp", addr, timeout)             │
│   Concurrencia:   Canal buffered como semáforo (MaxWorkers)         │
└─────────────────────────────────────────────────────────────────────┘
```

### Flujo Detallado

| Fase | Descripción | Componente |
|:-----|:------------|:-----------|
| **1. Inicialización** | Se parsean los flags de CLI (`-host`, `-ports`, `-timeout`, `-workers`) y se valida la configuración | `main.go` |
| **2. Creación de Canales** | Se crean dos canales buffered: `ports` (jobs) y `results` (output). El canal `ports` tiene tamaño `MaxWorkers` para actuar como semáforo | `scanner.go` |
| **3. Pool de Workers** | Se lanzan `N` goroutines (workers). Cada una lee puertos del canal `ports` y ejecuta `net.DialTimeout` para probar la conexión TCP | `scanner.go` |
| **4. Enqueue de Jobs** | Una goroutine separada itera sobre el rango de puertos y los envía al canal `ports`. Al terminar, cierra el canal | `scanner.go` |
| **5. Sincronización** | `sync.WaitGroup` garantiza que todos los workers terminen antes de cerrar el canal de resultados | `scanner.go` |
| **6. Conexión TCP** | Cada worker usa `net.DialTimeout("tcp", addr, timeout)` con timeout estricto. Si la conexión es exitosa, el puerto está ABIERTO. Se mide la latencia | `scanner.go` |
| **7. Recolección** | El goroutine principal recoge todos los `PortResult` del canal `results` | `scanner.go` |
| **8. Reporte** | Los resultados se formatean en tabla ASCII, JSON o resumen según el flag `-output` | `report.go` |

---

## 📁 Estructura de Archivos

```
01-port-scanner/
├── README.md          ← Este archivo de documentación
├── go.mod             ← Definición del módulo Go
├── main.go            ← Entry point, CLI con flag package, validación
├── scanner.go         ← Lógica de escaneo concurrente (worker pool)
├── scanner_test.go    ← Tests unitarios, de integración y benchmarks
└── report.go          ← Formateo de resultados (tabla, JSON, resumen)
```

### Responsabilidades por Archivo

| Archivo | Responsabilidad |
|:--------|:----------------|
| `main.go` | Punto de entrada. Define flags CLI (`-host`, `-ports`, `-timeout`, `-workers`, `-output`, `-all`, `-version`). Valida inputs. Orquesta el flujo completo. |
| `scanner.go` | Motor de escaneo. Contiene `ScanPorts()` (worker pool), `scanSinglePort()` (conexión TCP individual), `FilterOpen()`, el mapa `WellKnownServices` y `lookupService()`. |
| `report.go` | Generación de reportes. `PrintTable()` (tabla ASCII con bordes Unicode), `PrintJSON()` (JSON estructurado con serialización custom), `PrintSummary()` (una línea). |
| `scanner_test.go` | Suite de tests completa: tests unitarios, tests de integración con listeners TCP reales, tests de concurrencia, y benchmarks de rendimiento. |

---

## 🚀 Instalación y Ejecución

### Requisitos

- **Go 1.21+** instalado
- Permisos de red para conexiones TCP salientes

### Compilar

```bash
cd bonus-proyectos/01-port-scanner
go build -o port-scanner .
```

### Ejecutar Directamente

```bash
go run . -host <target> [opciones]
```

---

## 📘 Uso

```
Usage:
  port-scanner -host <target> [options]

Options:
  -host string       Target host (IP or hostname) [required]
  -ports string      Port range to scan (default "1-1024")
  -timeout duration  Connection timeout per port (default 500ms)
  -workers int       Maximum concurrent workers (default 100)
  -output string     Output format: 'table', 'json', or 'summary' (default "table")
  -all               Show all ports including closed ones
  -version           Show version information
```

### Formatos de `-ports`

| Formato | Ejemplo | Descripción |
|:--------|:--------|:------------|
| Puerto único | `80` | Escanea solo el puerto 80 |
| Rango | `1-1024` | Escanea puertos del 1 al 1024 |
| Lista | `22,80,443,3306` | Escanea puertos específicos |

---

## 💡 Ejemplos de Uso

### Escaneo básico (puertos well-known)

```bash
go run . -host 192.168.1.1
```

### Escaneo de rango extendido con más workers

```bash
go run . -host scanme.nmap.org -ports 1-1024 -workers 200
```

### Escaneo de puertos específicos con timeout alto

```bash
go run . -host 10.0.0.1 -ports 22,80,443,3306,5432,8080 -timeout 2s
```

### Escaneo completo con salida JSON

```bash
go run . -host 192.168.1.1 -ports 1-65535 -workers 1000 -output json > results.json
```

### Escaneo rápido con resumen

```bash
go run . -host example.com -ports 80-443 -timeout 200ms -workers 500 -output summary
```

### Ver todos los puertos (abiertos y cerrados)

```bash
go run . -host 192.168.1.1 -ports 20-25 -all
```

---

## 📊 Formatos de Salida

### Tabla (default)

```
  ╔══════════════════════════════════════════════════════════════╗
  ║              PORT SCANNER — SCAN RESULTS                    ║
  ╠══════════════════════════════════════════════════════════════╣
  ║  Host: scanme.nmap.org                                       ║
  ║  Open Ports: 2                                               ║
  ║  Scanned: 1024 ports in 5.234s                               ║
  ╠══════════════════════════════════════════════════════════════╣
  ║  PORT      STATE         SERVICE        LATENCY             ║
  ╠══════════════════════════════════════════════════════════════╣
  ║  22        OPEN          SSH            45.123ms            ║
  ║  80        OPEN          HTTP           32.456ms            ║
  ╚══════════════════════════════════════════════════════════════╝
```

### JSON

```json
{
  "host": "scanme.nmap.org",
  "start_port": 1,
  "end_port": 1024,
  "duration": "5.234s",
  "total_ports": 1024,
  "open_ports": 2,
  "results": [
    {
      "port": 22,
      "open": true,
      "service": "SSH",
      "latency": "45.123ms"
    },
    {
      "port": 80,
      "open": true,
      "service": "HTTP",
      "latency": "32.456ms"
    }
  ]
}
```

### Resumen

```
  Host: scanme.nmap.org | Scanned: 1024 | Open: 2 | Duration: 5.234s
  Open ports: 22/SSH, 80/HTTP
```

---

## 🧪 Tests

### Ejecutar todos los tests

```bash
go test -v ./...
```

### Ejecutar con race detector (recomendado)

```bash
go test -v -race ./...
```

### Ejecutar benchmarks

```bash
go test -bench=. -benchmem ./...
```

### Cobertura de tests

```bash
go test -cover ./...
```

### Suite de Tests Incluida

| Categoría | Tests | Descripción |
|:----------|:------|:------------|
| **Unitarios** | `TestLookupService`, `TestFilterOpen` | Lógica pura sin red |
| **Integración** | `TestScanPorts_OpenPort`, `TestScanPorts_ClosedPort`, `TestScanPorts_MultiplePorts` | Usan listeners TCP reales en localhost |
| **Concurrencia** | `TestScanPorts_ConcurrencyLimit` | Verifica que el pool de workers funciona sin deadlocks |
| **Reportes** | `TestPrintTable_*`, `TestPrintJSON`, `TestPrintSummary` | Validan la generación de cada formato de salida |
| **Parsing** | `TestParsePortRange_*` | Validan el parsing de rangos de puertos |
| **Benchmarks** | `BenchmarkScanPorts_*`, `BenchmarkFilterOpen` | Mediciones de rendimiento |

---

## 🛡️ Servicios Conocidos

El escáner identifica automáticamente los siguientes servicios por número de puerto:

| Puerto | Servicio | Puerto | Servicio |
|:-------|:---------|:-------|:---------|
| 20 | FTP-Data | 993 | IMAPS |
| 21 | FTP | 995 | POP3S |
| 22 | SSH | 1433 | MSSQL |
| 23 | Telnet | 1521 | Oracle |
| 25 | SMTP | 3306 | MySQL |
| 53 | DNS | 3389 | RDP |
| 80 | HTTP | 5432 | PostgreSQL |
| 110 | POP3 | 5900 | VNC |
| 111 | RPCBind | 6379 | Redis |
| 135 | MSRPC | 8080 | HTTP-Proxy |
| 139 | NetBIOS-SSN | 8443 | HTTPS-Alt |
| 143 | IMAP | 9090 | Prometheus |
| 443 | HTTPS | 27017 | MongoDB |
| 445 | SMB | | |

---

## 🔧 Decisiones de Diseño

### ¿Por qué Worker Pool y no fan-out ilimitado?

Lanzar una goroutine por cada puerto (fan-out ilimitado) es tentador pero peligroso:
- **File descriptor exhaustion**: El OS tiene un límite de descriptores de archivo (~1024 por defecto en Linux, ~8192 en Windows). Miles de conexiones simultáneas agotan este recurso.
- **Consumo de memoria**: Cada goroutine reserva ~2-8KB de stack. Con 65,535 puertos, serían ~130MB-524MB solo en stacks.
- **Saturación de red**: Demasiadas conexiones simultáneas pueden provocar paquetes RST del kernel.

El **worker pool con canal buffered como semáforo** controla estos problemas:
```go
// El canal con capacidad MaxWorkers actúa como semáforo.
// Solo MaxWorkers goroutines pueden estar activas simultáneamente.
ports := make(chan int, cfg.MaxWorkers)
```

### ¿Por qué `net.DialTimeout` y no `net.Dial`?

`net.Dial` usa el timeout del sistema operativo (~2 minutos en Linux). `net.DialTimeout` permite especificar un timeout estricto:
```go
conn, err := net.DialTimeout("tcp", addr, timeout)  // Timeout explícito
```

### ¿Por qué `sync.WaitGroup` además de channels?

El WaitGroup sincroniza la finalización de todos los workers para poder cerrar el canal de resultados de forma segura:
```go
go func() {
    wg.Wait()      // Esperar a que TODOS los workers terminen
    close(results) // Solo entonces cerrar el canal
}()
```

---

## 📜 Licencia

Este proyecto es parte del curso **"Go desde Cero: Laboratorio Completo"** — Fase Bonus: Proyectos de Sistemas y Redes.