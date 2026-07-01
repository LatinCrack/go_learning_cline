<div align="center">

# 🌐 CLI de Inventario de Red Concurrente

### *Proyecto 04 — Bonus Proyectos de Sistemas*

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Concurrency](https://img.shields.io/badge/Concurrency-Goroutines-4ECDC4?style=for-the-badge) ![Network](https://img.shields.io/badge/Type-Network_Scanner-FF6B6B?style=for-the-badge)

<br>

</div>

---

## 🎯 Objetivo

Herramienta CLI de inventario de red que escanea un rango de IPs en formato CIDR, realiza pings concurrentes para descubrir hosts activos, resuelve nombres DNS inversos, estima el sistema operativo basándose en el TTL de las respuestas, y exporta el inventario completo en formato JSON y tabla ASCII.

---

## 🧰 Stack Técnico

| Componente | Implementación |
|:-----------|:---------------|
| **`net/netip`** | Parsing CIDR y manipulación de IPs (Go 1.18+) |
| **Goroutines** | Pool de workers concurrentes para pings masivos |
| **Channels** | Semáforo para limitar concurrencia + recolección de resultados |
| **`sync.WaitGroup`** | Sincronización del pool de descubrimiento |
| **`os/exec`** | Ping nativo del SO (Windows/Linux/macOS) |
| **`net` package** | `net.LookupAddr()` para DNS reverso |
| **`encoding/json`** | Exportación estructurada del inventario |
| **`flag` package** | CLI con flags configurables |

---

## 📁 Estructura de Archivos

```
04-network-inventory/
├── README.md             ← Este archivo (documentación)
├── go.mod                ← Módulo Go
├── main.go               ← Entry point y configuración CLI
├── cidr.go               ← Parsing y expansión de rangos CIDR
├── discovery.go          ← Pool de descubrimiento concurrente (ping sweep)
├── resolver.go           ← Resolución DNS reversa y detección de OS
├── exporter.go           ← Exportación a JSON y tabla formateada
└── discovery_test.go     ← Tests unitarios
```

---

## ⚙️ Flags Disponibles

| Flag | Tipo | Default | Descripción |
|:-----|:-----|:--------|:------------|
| `--cidr` | `string` | *(requerido)* | Rango CIDR a escanear (ej: `192.168.1.0/24`) |
| `--timeout` | `duration` | `3s` | Timeout de ping por host (ej: `1s`, `2s`, `500ms`) |
| `--workers` | `int` | `50` | Número máximo de workers concurrentes (goroutines) |
| `--output` | `string` | `table` | Formato de salida: `table`, `json`, o `summary` |
| `--json-file` | `string` | *(vacío)* | Ruta para guardar el reporte JSON en disco |
| `--version` | `bool` | `false` | Mostrar información de versión |

---

## 🚀 Uso

### Compilar el binario

```bash
cd bonus-proyectos/04-network-inventory
go build -o network-inventory .
```

### Escanear la red local (tabla)

```bash
./network-inventory -cidr 192.168.1.0/24
```

### Escanear con más workers y timeout corto

```bash
./network-inventory -cidr 10.0.0.0/24 -workers 100 -timeout 1s
```

### Exportar a JSON en consola

```bash
./network-inventory -cidr 192.168.1.0/24 -output json
```

### Guardar reporte JSON en archivo

```bash
./network-inventory -cidr 172.16.0.0/24 -output table -json-file inventory.json
```

### Ver resumen rápido

```bash
./network-inventory -cidr 192.168.1.0/24 -output summary
```

---

## 🔍 Ejemplo de Salida (Tabla)

```
  ╔══════════════════════════════════════════════════════════════════════════════════════╗
  ║                         🌐 NETWORK INVENTORY REPORT                                ║
  ╚══════════════════════════════════════════════════════════════════════════════════════╝

  Range:      192.168.1.0/24
  Duration:   12.4s
  Hosts up:   5 / 254

  IP ADDRESS         HOSTNAME                     OS GUESS             TTL    LATENCY   
  ──────────────────────────────────────────────────────────────────────────────────────────
  192.168.1.1        gateway.lan                  Linux/Unix           64     3ms       
  192.168.1.10       desktop-pc                   Windows              128    12ms      
  192.168.1.15       —                            Linux/Unix           61     8ms       
  192.168.1.20       nas.local                    Linux/Unix           64     5ms       
  192.168.1.100      printer.local                Network Device...    255    45ms      
```

---

## 🔍 Ejemplo de Salida (JSON)

```json
{
  "scan_duration": "12.4s",
  "total_hosts": 254,
  "alive_hosts": 5,
  "cidr_range": "192.168.1.0/24",
  "hosts": [
    {
      "ip": "192.168.1.1",
      "hostname": "gateway.lan",
      "os_guess": "Linux/Unix",
      "ttl": 64,
      "latency": "3ms",
      "alive": true
    },
    {
      "ip": "192.168.1.2",
      "ttl": 0,
      "latency": "0s",
      "alive": false
    }
  ]
}
```

---

## 🏗️ Arquitectura Interna

### Flujo de Descubrimiento

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│  CIDR Input  │────▶│  ExpandCIDR  │────▶│  Worker Pool     │────▶│   Results    │
│  (flag)      │     │  (net/netip) │     │  (goroutines +   │     │   Channel    │
└─────────────┘     └──────────────┘     │   semaphore)     │     └──────┬───────┘
                                          └────────┬─────────┘            │
                                                   │                      ▼
                                                   │              ┌──────────────┐
                                                   ▼              │  Enrichment  │
                                          ┌─────────────────┐     │  (DNS + TTL) │
                                          │  os/exec ping   │     └──────┬───────┘
                                          │  (platform-     │            │
                                          │   specific)     │            ▼
                                          └─────────────────┘     ┌──────────────┐
                                                                  │   Exporter   │
                                                                  │  (JSON/TABLE)│
                                                                  └──────────────┘
```

### Detección de OS por TTL

| TTL Inicial Probable | Sistema Operativo |
|:---------------------|:------------------|
| **64** | Linux, macOS, Android, FreeBSD |
| **128** | Windows (todas las versiones) |
| **255** | Cisco IOS, Solaris, dispositivos de red |

> ⚠️ La detección por TTL es una **heurística** — el TTL decrece con cada hop de router. Un host a 3 hops con TTL inicial 64 se verá como TTL 61.

---

## 🧪 Ejecutar Tests

```bash
cd bonus-proyectos/04-network-inventory
go test -v ./...
```

Los tests cubren:
- Expansión correcta de CIDR (/24, /30, /31, /32)
- Validación de inputs inválidos
- Parsing de TTL desde salidas de ping (Windows y Linux)
- Estimación de OS por TTL
- Generación de reportes JSON válidos
- Filtrado de hosts activos
- Formateo de tablas
- Round-trip de conversión IP ↔ uint32

---

## 🔒 Consideraciones de Seguridad

- **Sin raw sockets**: Usa `os/exec` para ejecutar el `ping` nativo del SO, evitando la necesidad de privilegios root/admin para ICMP.
- **Límite de concurrencia**: El pool de workers está acotado por un canal semáforo para evitar agotamiento de recursos.
- **Safety cap**: Rangos CIDR mayores a /12 (~1M hosts) son rechazados para prevenir consumo excesivo de memoria.
- **Context-aware**: Soporta `context.Context` para cancelación graceful.

---

## 📋 Requisitos

- **Go 1.21+**
- Conectividad de red activa
- El comando `ping` disponible en el PATH del sistema

---

<div align="center">

*"Conocer tu red es el primer paso para protegerla."*

</div>