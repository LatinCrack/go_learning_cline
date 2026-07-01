# Proyecto 08 — Servidor DNS Ligero con Cache y Filtros

## Descripción

Servidor proxy DNS de alto rendimiento escrito en Go puro (sin dependencias externas). Intercepta consultas DNS, aplica listas de bloqueo (anti-ad/malware), almacena respuestas en caché con respeto estricto al TTL del registro y reenvía consultas legítimas a resolvers upstream configurables.

## Arquitectura

```
┌─────────────┐     ┌──────────────────────────────────────────────────┐
│   Cliente    │────▶│  UDP Listener (:5353)                           │
│  (dig/nslook)│◀────│                                                 │
└─────────────┘     │  ┌───────────┐   ┌──────────┐   ┌────────────┐  │
                    │  │ protocol  │──▶│ blocklist │──▶│   cache    │  │
                    │  │ (parsing) │   │ (filtro)  │   │ (TTL+mutex)│  │
                    │  └───────────┘   └──────────┘   └────────────┘  │
                    │       │                            │             │
                    │       ▼                            ▼             │
                    │  ┌───────────┐              ┌──────────────┐    │
                    │  │ resolver  │─────────────▶│  Upstream    │    │
                    │  │ (forward) │◀─────────────│ 8.8.8.8:53  │    │
                    │  └───────────┘              └──────────────┘    │
                    └──────────────────────────────────────────────────┘
```

## Estructura de Archivos

| Archivo          | Descripción                                                    |
|------------------|----------------------------------------------------------------|
| `main.go`        | Punto de entrada, flags CLI, wiring de componentes             |
| `server.go`      | Listener UDP, pipeline de procesamiento de consultas           |
| `protocol.go`    | Parsing/serialización binaria del protocolo DNS (RFC 1035)     |
| `cache.go`       | Caché thread-safe con TTL y limpieza automática                |
| `blocklist.go`   | Carga y matching de dominios bloqueados (exacto + wildcard)    |
| `resolver.go`    | Forwarding a servidores DNS upstream con round-robin           |
| `server_test.go` | Suite completa de tests unitarios, integración y benchmarks    |

## Flags de Configuración

| Flag                 | Alias | Default               | Descripción                                              |
|----------------------|-------|-----------------------|----------------------------------------------------------|
| `--port`             | `-p`  | `5353`                | Puerto UDP de escucha                                    |
| `--upstream`         | `-u`  | `8.8.8.8,1.1.1.1`    | Servidores DNS upstream (separados por coma)             |
| `--blocklist`        | `-b`  | (vacío)               | Ruta al archivo de bloqueo (un dominio por línea)        |
| `--mode`             |       | `nxdomain`            | Modo de bloqueo: `nxdomain` o `sinkhole`                 |
| `--sinkhole`         |       | `127.0.0.1`           | IP para respuestas en modo sinkhole                      |
| `--ttl`              |       | `300`                 | TTL por defecto (segundos) para respuestas sin TTL       |
| `--cache-cleanup`    |       | `60`                  | Intervalo de limpieza de caché (segundos)                |
| `--verbose`          | `-v`  | `false`               | Habilitar logging detallado                              |

## Compilación y Ejecución

```bash
# Compilar
cd bonus-proyectos/08-dns-server
go build -o dns-server .

# Ejecutar directamente
go run . --port 5353 --upstream 8.8.8.8,1.1.1.1

# Con bloqueo de dominios
go run . --port 5353 --upstream 8.8.8.8 --blocklist blocklist.txt

# Modo sinkhole (redirige a 127.0.0.1)
go run . --port 5353 --mode sinkhole --blocklist blocklist.txt

# Puerto 53 (requiere root/admin)
sudo ./dns-server --port 53 --upstream 8.8.8.8 --blocklist blocklist.txt
```

## Pruebas con dig / nslookup

```bash
# Consulta básica A record
dig @127.0.0.1 -p 5353 google.com A

# Consulta con múltiples tipos
dig @127.0.0.1 -p 5353 google.com ANY

# Consulta inversa (PTR)
dig @127.0.0.1 -p 5353 -x 8.8.8.8

# Usando nslookup
nslookup google.com 127.0.0.1 -port=5353

# Verificar bloqueo (debe retornar NXDOMAIN o 127.0.0.1)
dig @127.0.0.1 -p 5353 ads.doubleclick.net A

# Verificar caché (segunda consulta debe ser instantánea)
dig @127.0.0.1 -p 5353 cloudflare.com A
dig @127.0.0.1 -p 5353 cloudflare.com A  # Cache hit
```

## Formato del Archivo de Bloqueo

El archivo de bloqueo es texto plano, un dominio por línea:

```
# Comentarios con # o //
ads.doubleclick.net
*.tracker.com
*.adserver.net
malware.example.com
# Líneas vacías se ignoran
```

- **Match exacto**: `ads.example.com` bloquea solo ese dominio
- **Wildcard**: `*.example.com` bloquea ese dominio y todos sus subdominios
- Las comparaciones son **case-insensitive**

## Tests

```bash
# Ejecutar todos los tests
go test -v

# Solo tests unitarios (sin integración)
go test -v -run "^(TestBuild|TestParse|TestCache|TestBlocklist|TestNormalize|TestType|TestNewResolver|TestRoundTrip|TestParseIPv4|TestParseUpstream)"

# Tests de integración
go test -v -run "TestIntegration"

# Benchmarks
go test -bench=. -benchmem

# Coverage
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Rendimiento

Los benchmarks incluidos miden el rendimiento crítico del pipeline:

```
BenchmarkBuildName         ~50 ns/op    (construcción de nombres DNS)
BenchmarkParseMessage      ~300 ns/op   (parsing de paquetes DNS)
BenchmarkBlocklistLookup   ~100 ns/op   (lookup en tabla de bloqueo)
BenchmarkCacheLookup       ~50 ns/op    (lookup en caché con RLock)
```

## Dependencias

**Ninguna dependencia externa.** El proyecto utiliza únicamente la biblioteca estándar de Go:

- `encoding/binary` — Decodificación binaria de estructuras DNS
- `net` — Socket UDP nativo
- `sync` / `sync/atomic` — Concurrencia segura
- `flag` — Parsing de flags CLI
- `os/signal` — Manejo de señales para shutdown graceful