# Proyecto 05 — Monitor de Recursos del Sistema (CPU / Memoria / Disco / Red)

> **Serie:** bonus-proyectos · **Dificultad:** ██████████ 10/10 (SRE / Systems Engineering)  
> **Tema central:** Recolección nativa de métricas OS, concurrencia con tickers, exclusión mutua, exportación Prometheus/JSON.

---

## 🎯 Objetivo

Construir un **monitor de recursos del sistema** escrito 100% en Go puro (sin CGO, sin dependencias externas) que:

1. Lee métricas de **CPU**, **memoria**, **disco** y **red** directamente de `/proc` y `syscall` en Linux.
2. Expone los datos vía **HTTP** en formato **JSON** y **Prometheus text exposition**.
3. Dispara **alertas** cuando los recursos superan umbrales configurables.
4. Utiliza **goroutines** con `time.NewTicker` para recolección paralela y `sync.RWMutex` para acceso seguro.

---

## 📂 Estructura de Archivos

```
05-system-monitor/
├── main.go            # Punto de entrada, flags CLI, orquestación
├── cpu.go             # Recolector CPU — parsing nativo de /proc/stat
├── memory.go          # Recolector Memoria — parsing nativo de /proc/meminfo
├── disk.go            # Tipos y estructura base del recolector de disco
├── disk_linux.go      # Implementación Linux: syscall.Statfs + /proc/mounts
├── disk_other.go      # Fallback para Windows/macOS
├── network.go         # Recolector Red — parsing nativo de /proc/net/dev
├── alerter.go         # Motor de alertas con umbrales configurables
├── exporter.go        # Servidor HTTP (Prometheus + JSON + health)
├── monitor_test.go    # Suite de tests unitarios
├── go.mod             # Definición del módulo Go
└── README.md          # Este archivo
```

---

## 🚀 Uso Rápido

### Compilar

```bash
cd bonus-proyectos/05-system-monitor
go build -o system-monitor .
```

### Ejecutar con valores por defecto

```bash
./system-monitor
```

Salida esperada:
```
╔══════════════════════════════════════════════════════════╗
║           System Resource Monitor vdev                  ║
╠══════════════════════════════════════════════════════════╣
║  Interval    : 5s                                       ║
║  CPU Alert   : 85.0%                                    ║
║  Mem Alert   : 90.0%                                    ║
║  Disk Alert  : 95.0%                                    ║
║  Swap Alert  : 80.0%                                    ║
║  Listen      : :9100                                    ║
╚══════════════════════════════════════════════════════════╝
```

### Ejecutar con flags personalizados

```bash
./system-monitor \
  --interval 10s \
  --cpu-alert 90 \
  --mem-alert 85 \
  --disk-alert 98 \
  --swap-alert 75 \
  --listen :9200
```

---

## ⚙️ Flags de Configuración

| Flag           | Tipo       | Default | Descripción                                                  |
|----------------|------------|---------|--------------------------------------------------------------|
| `--interval`   | `duration` | `5s`    | Intervalo de recolección de métricas. Acepta `5s`, `1m`, etc. |
| `--cpu-alert`  | `float64`  | `85.0`  | Umbral de alerta para uso de CPU (%). `0` desactiva.         |
| `--mem-alert`  | `float64`  | `90.0`  | Umbral de alerta para uso de memoria RAM (%). `0` desactiva. |
| `--disk-alert` | `float64`  | `95.0`  | Umbral de alerta para uso de disco (%). `0` desactiva.       |
| `--swap-alert` | `float64`  | `80.0`  | Umbral de alerta para uso de swap (%). `0` desactiva.        |
| `--listen`     | `string`   | `:9100` | Dirección HTTP para el endpoint de métricas.                 |
| `--version`    | `bool`     | `false` | Imprime la versión y sale.                                   |

---

## 📡 Endpoints HTTP

### `GET /metrics` — Formato Prometheus

Devuelve todas las métricas en formato **Prometheus text exposition** (`text/plain; version=0.0.4`).

Compatible con `prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'system-monitor'
    static_configs:
      - targets: ['localhost:9100']
```

Métricas exportadas:
- `system_cpu_usage_percent`, `system_cpu_user_percent`, `system_cpu_system_percent`, `system_cpu_iowait_percent`, `system_cpu_cores`
- `system_memory_usage_percent`, `system_memory_total_mb`, `system_memory_used_mb`, `system_memory_free_mb`, `system_memory_buffers_mb`, `system_memory_cached_mb`
- `system_swap_usage_percent`, `system_swap_total_mb`, `system_swap_used_mb`
- `system_disk_usage_percent{mount,fstype}`, `system_disk_total_mb{mount}`, `system_disk_used_mb{mount}`, `system_disk_free_mb{mount}`, `system_disk_inodes_used{mount}`, `system_disk_inodes_free{mount}`
- `system_net_rx_bytes{iface}`, `system_net_tx_bytes{iface}`, `system_net_rx_rate_bytes_s{iface}`, `system_net_tx_rate_bytes_s{iface}`, `system_net_rx_errors{iface}`, `system_net_tx_errors{iface}`
- `system_alerts_active`

### `GET /metrics/json` — Formato JSON

Devuelve el snapshot completo como JSON estructurado:

```json
{
  "cpu": { "usage_percent": 23.5, "cores": 4, "..." },
  "memory": { "total_mb": 16384, "used_mb": 8192, "..." },
  "disk": { "partitions": [ { "mount_point": "/", "..." } ] },
  "network": { "interfaces": [ { "name": "eth0", "..." } ] },
  "alerts": [],
  "host": "lab-server-01"
}
```

### `GET /health` — Health Check

```json
{"status":"ok","timestamp":"2025-01-01T00:00:00Z"}
```

---

## 🏗️ Arquitectura Interna

### Recolectores Concurrentes

Cada recurso tiene su propio **collector** que corre en una goroutine independiente:

```
main goroutine
  ├── go cpu.Start()     ← time.NewTicker(cpu.interval)
  ├── go mem.Start()     ← time.NewTicker(mem.interval)
  ├── go disk.Start()    ← time.NewTicker(disk.interval)
  └── go net.Start()     ← time.NewTicker(net.interval)
```

### Acceso Seguro con sync.RWMutex

Cada collector protege sus métricas con un `sync.RWMutex`:
- **Escritura** (`Lock`): solo dentro de `collect()`, en cada tick.
- **Lectura** (`RLock`): desde `GetMetrics()`, llamado por el exporter HTTP.

### Cálculo de CPU por Deltas

La CPU se calcula matemáticamente a partir de **dos lecturas consecutivas** de `/proc/stat`:

```
usage% = (deltaTotal - deltaIdle) / deltaTotal × 100
```

La primera lectura solo almacena el baseline; los porcentajes se reportan a partir del segundo tick.

### Parsing Nativo de /proc

| Archivo           | Contenido                                  |
|-------------------|--------------------------------------------|
| `/proc/stat`      | Contadores acumulados de CPU (user, system, idle, iowait...) |
| `/proc/meminfo`   | MemTotal, MemFree, Buffers, Cached, SwapTotal, SwapFree      |
| `/proc/mounts`    | Puntos de montaje reales para `syscall.Statfs`               |
| `/proc/net/dev`   | Contadores de bytes/paquetes por interfaz de red             |

---

## 🧪 Tests

```bash
go test -v -race -count=1 ./...
```

Los tests cubren:
- Cálculo de deltas de CPU (`CPUData.Total`, `CPUData.Delta`)
- Parsing de `parseUint64`
- Umbrales de alerta (por debajo, por encima, cooldown, desactivado)
- Endpoints HTTP (JSON, Prometheus, Health)
- Thread-safety concurrente de collectors
- Formateo de bytes humanos
- Ring buffer de alertas

---

## 🔧 Compilación Cruzada (para Linux target)

```bash
# Desde Windows/macOS, compilar para Linux:
GOOS=linux GOARCH=amd64 go build -o system-monitor-linux .

# Con versión inyectada:
go build -ldflags "-s -w -X main.Version=1.0.0" -o system-monitor .
```

---

## 📄 Licencia

Proyecto educativo dentro de la serie **bonus-proyectos** del curso de Go.