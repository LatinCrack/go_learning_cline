<div align="center">

# 🟤 Proyecto 07 — Proxy HTTP/HTTPS Reverso Simple

### *Balanceo de Carga · Health Checks · Graceful Shutdown · Go Nativo*

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Net/http](https://img.shields.io/badge/Lib-Net/http-00ADD8?style=for-the-badge) ![Concurrent](https://img.shields.io/badge/Concurrent-Goroutines-4ECDC4&style=for-the-badge) ![Tests](https://img.shields.io/badge/Tests-✅-FF6B6B?style=for-the-badge)

<br>

> *"Un proxy reverso es la puerta de entrada a tu infraestructura: un único punto que distribuye, protege y escala tu tráfico."*

</div>

---

## 🎯 Objetivo

Construir un **proxy HTTP/HTTPS reverso** de grado producción que distribuya tráfico entre múltiples backends utilizando el algoritmo **Round Robin**. Implementa health checks automáticos en background, inyección de headers de proxy (`X-Forwarded-For`, `X-Real-IP`), logging de accesos, manejo de timeouts estrictos con `context`, y graceful shutdown interceptando señales del sistema operativo.

---

## 🏗️ Arquitectura

```
                    ┌─────────────────────────────────┐
                    │        Reverse Proxy (:8080)     │
                    │                                   │
   Cliente ───────►│  LoggingMiddleware                │
                    │       │                           │
                    │       ▼                           │
                    │  ReverseProxyHandler              │
                    │       │                           │
                    │       ▼                           │
                    │  Balancer (Round Robin)           │
                    │       │                           │
                    │       ├──► Backend 1 (9001)       │
                    │       ├──► Backend 2 (9002)       │
                    │       └──► Backend 3 (9003)       │
                    │                                   │
                    │  HealthChecker ──── goroutine     │
                    │  (Ticker periódico, marca UP/DOWN)│
                    └─────────────────────────────────┘
```

---

## 📁 Estructura de Archivos

```
07-reverse-proxy/
├── README.md            ← Este archivo de documentación
├── go.mod               ← Definición del módulo Go
├── main.go              ← Entry point, CLI flags, graceful shutdown
├── proxy.go             ← Lógica del reverse proxy con httputil
├── balancer.go          ← Algoritmo Round Robin con sync/atomic
├── healthcheck.go       ← Rutina de health checks en background
├── middleware.go        ← Logging y recovery middleware
└── proxy_test.go        ← Tests unitarios e integración
```

---

## 🧰 Componentes de Go Utilizados

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **`net/http/httputil`** | `ReverseProxy` como base para el forwarding de requests |
| **`sync/atomic`** | Contador atómico para Round Robin lock-free, estado Alive sin mutexes |
| **`sync.RWMutex`** | Protección del slice de backends durante mutaciones del health checker |
| **`context`** | Timeouts estrictos en conexiones hacia backends y shutdown controlado |
| **`time.Ticker`** | Disparador periódico del health checker en su propia goroutine |
| **`os/signal`** | Intercepta SIGTERM/SIGINT para graceful shutdown |
| **`net/http`** | Servidor HTTP, manipulación de headers, middleware pattern |
| **`net/url`** | Parsing de URLs de backends |

---

## 🚀 Flags de Configuración

| Flag | Tipo | Default | Descripción |
|:-----|:-----|:--------|:------------|
| `--port` | `int` | `8080` | Puerto donde escucha el proxy |
| `--backends` | `string` | *(requerido)* | Lista separada por comas de URLs de backends |
| `--health-interval` | `duration` | `10s` | Intervalo entre health checks |
| `--health-timeout` | `duration` | `3s` | Timeout de cada health check individual |
| `--request-timeout` | `duration` | `30s` | Timeout para requests proxyeados hacia backends |

---

## ▶️ Compilación y Ejecución

### Compilar

```bash
cd bonus-proyectos/07-reverse-proxy
go build -o reverse-proxy .
```

### Ejecutar con 3 backends

```bash
./reverse-proxy \
  --port 8080 \
  --backends http://localhost:9001,http://localhost:9002,http://localhost:9003
```

### Ejecutar directamente con `go run`

```bash
go run . --port 8080 --backends http://localhost:9001,http://localhost:9002,http://localhost:9003
```

---

## 🧪 Probar el Balanceo de Carga

### Paso 1: Iniciar servidores backend simulados

Abre **3 terminales** y ejecuta un servidor HTTP simple en cada una:

```bash
# Terminal 1
python3 -m http.server 9001 --directory /tmp/backend1

# Terminal 2
python3 -m http.server 9002 --directory /tmp/backend2

# Terminal 3
python3 -m http.server 9003 --directory /tmp/backend3
```

O alternativamente, con Go:

```bash
# Terminal 1
go run -exec 'echo' . 2>/dev/null || python3 -c "
import http.server, socketserver
handler = http.server.SimpleHTTPRequestHandler
with socketserver.TCPServer(('', 9001), handler) as s:
    print('Backend 1 on :9001'); s.serve_forever()
"
```

### Paso 2: Iniciar el proxy

```bash
# Terminal 4
cd bonus-proyectos/07-reverse-proxy
go run . --port 8080 \
  --backends http://localhost:9001,http://localhost:9002,http://localhost:9003
```

### Paso 3: Enviar peticiones y observar el balanceo

```bash
# Enviar 6 peticiones y observar los logs del proxy
for i in {1..6}; do
  echo "=== Request $i ==="
  curl -s http://localhost:8080/ | head -5
  echo ""
done
```

### Paso 4: Verificar headers de proxy inyectados

```bash
# Usar un backend que responda con los headers recibidos
curl -v http://localhost:8080/ 2>&1 | grep -i "x-forwarded\|x-real-ip"
```

### Paso 5: Probar health checks

```bash
# Detener uno de los backends (Ctrl+C en Terminal 2)
# Observar en los logs del proxy:
#   [healthcheck] backend http://localhost:9002 is unreachable: ...
#   [balancer] backend marked DOWN: http://localhost:9002

# Las siguientes peticiones solo irán a los backends activos:
for i in {1..4}; do curl -s http://localhost:8080/ | head -3; done

# Reiniciar el backend 2 — el health checker lo detectará:
#   [healthcheck] backend http://localhost:9002 recovered (HTTP 200)
#   [balancer] backend marked UP: http://localhost:9002
```

### Paso 6: Probar graceful shutdown

```bash
# Enviar SIGTERM al proceso del proxy
kill -SIGTERM $(pgrep reverse-proxy)

# Logs esperados:
#   [main] Received signal: terminated — initiating graceful shutdown
#   [healthcheck] stopped
#   [main] Server stopped gracefully
```

---

## 🔍 Ejemplo de Salida del Proxy

```
2026/06/30 09:35:00.123456 main.go:83: [main] Reverse proxy listening on :8080
2026/06/30 09:35:00.123789 main.go:84: [main] Backends: [http://localhost:9001 http://localhost:9002 http://localhost:9003]
2026/06/30 09:35:00.124012 healthcheck.go:42: [healthcheck] started — interval: 10s, timeout: 3s
2026/06/30 09:35:01.500123 middleware.go:64: [access] GET / 127.0.0.1:52340 → 200 (1256 bytes, 1.2ms)
2026/06/30 09:35:02.100456 middleware.go:64: [access] GET / 127.0.0.1:52341 → 200 (1256 bytes, 890µs)
2026/06/30 09:35:02.700789 middleware.go:64: [access] GET / 127.0.0.1:52342 → 200 (1256 bytes, 1.1ms)
2026/06/30 09:35:10.125000 healthcheck.go:82: [healthcheck] backend http://localhost:9002 is unreachable: ...
2026/06/30 09:35:10.125345 balancer.go:72: [balancer] backend marked DOWN: http://localhost:9002
```

---

## 🧪 Ejecutar Tests

```bash
cd bonus-proyectos/07-reverse-proxy
go test -v ./...
```

### Cobertura de Tests

| Test | Descripción |
|:-----|:------------|
| `TestBalancerRoundRobin` | Verifica que el balanceador cicla por los backends en orden |
| `TestBalancerSkipsUnhealthy` | Confirma que los backends DOWN son saltados |
| `TestBalancerAllUnhealthy` | Valida que retorna nil cuando no hay backends sanos |
| `TestBalancerHealthyCount` | Verifica el conteo de backends activos |
| `TestBalancerMarkUpDown` | Prueba las operaciones MarkUp y MarkDown |
| `TestBalancerConcurrency` | Test de concurrencia con 100 goroutines simultáneas |
| `TestReverseProxyForwarding` | Verifica el forwarding de requests y respuesta del backend |
| `TestReverseProxyNoHealthyBackends` | Valida respuesta 503 cuando no hay backends |
| `TestReverseProxyHeaderInjection` | Confirma inyección de X-Forwarded-For, X-Real-IP, etc. |
| `TestLoggingMiddleware` | Verifica que el middleware no altera el response |
| `TestRecoveryMiddleware` | Confirma que panics se recuperan con 500 |
| `TestChain` | Valida el orden de ejecución de la cadena de middleware |
| `TestIntegrationRoundRobinProxy` | Test end-to-end: 9 requests distribuidas equitativamente entre 3 backends |
| `TestClientIP*` | Tests de extracción de IP del cliente |
| `TestSchemeOf*` | Tests de detección de esquema HTTP/HTTPS |

---

## 🌍 Utilidad Real

| Uso | Descripción |
|:----|:------------|
| **Microservicios** | Punto de entrada único para distribuir tráfico entre múltiples servicios |
| **High Availability** | Elimina puntos únicos de fallo con balanceo automático |
| **Desarrollo Local** | Simula infraestructura de producción en tu máquina |
| **Seguridad** | Oculta la topología interna de los backends |
| **Observabilidad** | Logging centralizado de todas las peticiones que pasan por el proxy |

---

<div align="center">

### *"La simplicidad es la sofisticación suprema."*
### — **Leonardo da Vinci**

</div>