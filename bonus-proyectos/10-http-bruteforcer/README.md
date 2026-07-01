# Proyecto 10 — HTTP Bruteforcer con Rate Limiting

## Descripción

Herramienta de pentesting para auditoría de seguridad en endpoints de autenticación HTTP. Implementa un motor de fuerza bruta concurrente con control de velocidad avanzado (rate limiting), rotación de proxies, detección automática de formularios de login, y generación de reportes en JSON.

> ⚠️ **USO ÉTICO ÚNICAMENTE**: Esta herramienta está diseñada exclusivamente para pruebas de seguridad autorizadas. El acceso no autorizado a sistemas informáticos es ilegal.

---

## Estructura del Proyecto

```
10-http-bruteforcer/
├── main.go                # Punto de entrada, CLI y orquestación principal
├── config.go              # Estructura Config, validación, carga/guardado JSON
├── wordlist.go            # Gestión de diccionarios (usuarios y contraseñas)
├── ratelimiter.go         # Token Bucket Rate Limiter con burst y jitter
├── httpclient.go          # Cliente HTTP con rotación de proxies y cookies
├── detector.go            # Detección de éxito/fallo y auto-detección de formularios
├── bruteforcer.go         # Motor concurrente con workers, retry y backoff
├── reporter.go            # Salida en terminal (colores) y exportación JSON
├── util.go                # Utilidades: jitter, contains, truncate, formatDuration
├── bruteforcer_test.go    # Tests unitarios e integración con httptest
├── config.json            # Archivo de configuración de ejemplo
├── go.mod                 # Módulo Go (sin dependencias externas)
└── wordlists/
    ├── users.txt          # Diccionario de usuarios comunes (30 entradas)
    └── passwords.txt      # Diccionario de contraseñas comunes (100 entradas)
```

### Archivos Principales

| Archivo | Responsabilidad |
|---|---|
| `main.go` | CLI con `flag`, carga de config, signal handling (Ctrl+C), orquestación |
| `config.go` | Struct `Config` con tags JSON, validación, carga desde archivo, defaults |
| `wordlist.go` | Parsing de diccionarios (ignora comentarios `#`, líneas vacías), generación de pares |
| `ratelimiter.go` | Token bucket con burst configurable, jitter aleatorio, y delayed burst |
| `httpclient.go` | `http.Client` con proxy rotativo, cookies, custom headers, User-Agent |
| `detector.go` | Análisis de respuestas HTTP: body, status code, redirects, detección de formularios |
| `bruteforcer.go` | Workers goroutines, retry con backoff, estadísticas atómicas, stop-on-success |
| `reporter.go` | Output formateado con colores ANSI, resumen final, exportación JSON |
| `util.go` | Funciones puras: `applyJitter`, `contains`, `truncateString`, `formatDuration` |

---

## Configuración (`config.json`)

El archivo `config.json` permite configurar todos los parámetros del ataque:

```json
{
  "target_url": "http://192.168.1.100/login",
  "target_type": "form",
  "username_field": "username",
  "password_field": "password",
  "success_indicator": "",
  "failure_indicator": "Invalid credentials",
  "success_status": 302,
  "users_file": "wordlists/users.txt",
  "passwords_file": "wordlists/passwords.txt",
  "workers": 10,
  "delay_ms": 500,
  "burst_size": 20,
  "burst_delay_ms": 5000,
  "max_retries": 3,
  "timeout_sec": 10,
  "jitter": true,
  "jitter_factor": 0.3,
  "proxies": [],
  "output_file": "report.json",
  "verbose": false,
  "custom_headers": {},
  "cookies": {},
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
}
```

### Parámetros Clave

| Parámetro | Tipo | Descripción |
|---|---|---|
| `target_url` | string | URL del endpoint de login (requerido) |
| `target_type` | string | `"form"` (POST con campos) o `"basic"` (HTTP Basic Auth) |
| `username_field` | string | Nombre del campo HTML del usuario (solo para `form`) |
| `password_field` | string | Nombre del campo HTML de la contraseña (solo para `form`) |
| `success_indicator` | string | Texto en el body de respuesta que indica éxito |
| `failure_indicator` | string | Texto en el body que indica fallo (default: `"invalid"`) |
| `success_status` | int | Código HTTP que indica éxito (ej: `302` para redirects) |
| `workers` | int | Número de goroutines concurrentes (1-200, default: `10`) |
| `delay_ms` | int | Delay base entre requests en milisegundos |
| `burst_size` | int | Requests permitidos en ráfaga antes de throttling |
| `burst_delay_ms` | int | Delay tras agotar la ráfaga (milisegundos) |
| `jitter` | bool | Activar variación aleatoria en los delays |
| `jitter_factor` | float64 | Factor de jitter 0.0-1.0 (0.3 = ±30%) |
| `proxies` | []string | Lista de URLs de proxy para rotación automática |
| `max_retries` | int | Reintentos por request en error de red |
| `custom_headers` | map | Headers HTTP adicionales para cada request |
| `cookies` | map | Cookies predefinidas para cada request |

---

## Uso

### Con archivo de configuración

```bash
go run . -config config.json
```

### Directamente por CLI

```bash
go run . -url http://target/login \
         -users wordlists/users.txt \
         -passwords wordlists/passwords.txt \
         -workers 20 \
         -delay 1000 \
         -burst 10 \
         -failure "Invalid credentials"
```

### Flags de CLI (override config)

```
-config          Ruta al archivo JSON de configuración
-url             URL objetivo (override config)
-users           Archivo de diccionario de usuarios
-passwords       Archivo de diccionario de contraseñas
-workers         Número de workers concurrentes
-delay           Delay entre requests (ms)
-burst           Tamaño de ráfaga
-burst-delay     Delay post-ráfaga (ms)
-timeout         Timeout HTTP (segundos)
-retries         Reintentos por request
-v               Modo verbose (muestra todos los intentos)
-proxies         URLs de proxy separadas por coma
-output          Ruta del reporte JSON
-detect          Auto-detectar campos del formulario de login
-success         Texto del body que indica éxito
-failure         Texto del body que indica fallo
-success-status  Código HTTP de éxito
-ua              User-Agent personalizado
```

### Ejemplos de Uso

```bash
# Auditoría con detección automática de formulario
go run . -config config.json -detect -v

# Basic Auth con proxies
go run . -url http://api.internal/admin \
         -users users.txt -passwords passwords.txt \
         -success-status 200 \
         -proxies "http://proxy1:8080,http://proxy2:8080"

# Rate limiting agresivo para evitar WAF
go run . -config config.json -delay 3000 -burst 5 -burst-delay 30000

# Ejecutar tests
go test -v ./...
```

---

## Motor de Rate Limiting

El sistema de rate limiting implementa un **Token Bucket** con las siguientes características:

1. **Burst Control**: Permite N requests rápidos (burst) antes de activar el throttling
2. **Base Delay**: Aplica un delay configurable entre cada request individual
3. **Delayed Burst**: Cuando se agota la ráfaga, aplica un delay extendido para "recargar" el bucket
4. **Jitter Aleatorio**: Añade variación aleatoria (±factor) a todos los delays para evitar patrones predecibles
5. **Backoff Adaptativo**: Si el servidor responde con 429/403/503, aplica backoff triple automáticamente

```
Timeline ejemplo (burst=10, delay=500ms, burst_delay=5000ms):

[req][req][req]...[req]  ← 10 requests en burst (rápidos)
      ↓ burst agotado
[==== 5000ms delay ====]
      ↓ bucket recargado
[req][req]...[req]  ← siguiente burst
```

---

## Detección de Login

### Detección Automática de Formularios (`-detect`)

La herramienta puede analizar automáticamente el HTML de la página de login para detectar:
- Campos de usuario (busca `type="text"`, `type="email"`, nombres comunes)
- Campo de contraseña (`type="password"`)
- Action del formulario

### Estrategias de Detección de Éxito/Fallo

1. **Status code** — Código HTTP configurado como exitoso (ej: 302)
2. **Body indicator (success)** — Texto presente en la respuesta exitosa
3. **Body indicator (failure)** — Texto presente en la respuesta fallida
4. **Redirect analysis** — Redirecciones fuera de la página de login = éxito
5. **Status heuristics** — 200 en página de login generalmente = fallo (formulario re-renderizado)

---

## Arquitectura de Concurrencia

```
┌──────────────┐
│  Generator   │──→ channel pairs ──→ ┌──────────┐
│ (users × pw) │                      │ Worker 1 │
└──────────────┘                      │ Worker 2 │
                                      │ Worker 3 │
                                      │ ...      │
                                      │ Worker N │
                                      └────┬─────┘
                                           │
                                      channel results
                                           │
                                      ┌────▼─────┐
                                      │ Reporter  │→ terminal + JSON
                                      └──────────┘
```

- **Generator**: Produce pares (user, password) en un canal con buffer
- **Workers**: Goroutines que consumen pares, aplican rate limit, ejecutan HTTP requests
- **Reporter**: Goroutine que consume resultados y actualiza la terminal en tiempo real
- **Stats**: Contadores atómicos (`atomic.Int64`) para thread-safe statistics
- **Graceful Shutdown**: `context.Context` + `signal.Notify` para Ctrl+C limpio

---

## Dependencias

**Cero dependencias externas** — implementado 100% con la librería estándar de Go:

- `net/http` — Cliente HTTP
- `context` — Cancelación y timeouts
- `sync` — WaitGroup, Mutex
- `sync/atomic` — Contadores thread-safe
- `encoding/json` — Configuración y reportes
- `crypto/tls` — Control de certificados
- `flag` — CLI
- `os/signal` — Signal handling
- `math/rand` — Jitter aleatorio
- `strings` — Parsing HTML
- `time` — Delays y medición

---

## Ejemplo de Salida

```
╔══════════════════════════════════════════════════════════════╗
║           HTTP BRUTEFORCER — Security Audit Tool            ║
╚══════════════════════════════════════════════════════════════╝

  Target:       http://192.168.1.100/login
  Type:         form
  Workers:      10
  Delay:        500ms
  Burst size:   20

  ────────────────────────────────────────────────────────────

  [14:23:01] root:123456 → ✗ failed  (HTTP 200 | 45ms)  [attempts: 1 | 22.2 req/s]
  [14:23:01] root:password → ✗ failed  (HTTP 200 | 38ms)  [attempts: 2 | 25.0 req/s]
  ...
  [14:23:15] admin:admin123 → ✓ SUCCESS  (HTTP 302 | 52ms)  [attempts: 847 | 56.3 req/s]

  ═══════════════════════════════════════════════════════════
  ATTACK SUMMARY
  ═══════════════════════════════════════════════════════════

  Total attempts:  847
  Successes:       1
  Duration:        15s
  Average rate:    56.5 req/s

  ┌──────────────────────────────────────────────────────────┐
  │                    🔓 CREDENTIALS FOUND                  │
  ├──────────────────────────────────────────────────────────┤
  │  Username: admin                 Password: admin123      │
  └──────────────────────────────────────────────────────────┘

  Report saved to: report.json
```

---

## Código de Salida

| Código | Significado |
|---|---|
| `0` | Credenciales encontradas |
| `1` | Error de configuración o ejecución |
| `2` | Ejecución completada sin encontrar credenciales |