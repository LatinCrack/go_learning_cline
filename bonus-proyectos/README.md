<div align="center">

# 🛡️ Fase Bonus: Proyectos de Sistemas y Redes

### *10 Proyectos Prácticos · Go Aplicado · Infraestructura Real*

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Proyectos](https://img.shields.io/badge/Proyectos-10-FF6B6B?style=for-the-badge) ![Nivel](https://img.shields.io/badge/Nivel-Profesional-4ECDC4?style=for-the-badge) ![Enfoque](https://img.shields.io/badge/Enfoque-Sistemas_%26_Redes-FFE66D&style=for-the-badge)

<br>

> *"La mejor forma de aprender un lenguaje es construir herramientas que resolverían problemas reales de tu día a día como ingeniero."*

</div>

---

## 📖 Acerca de Esta Fase Bonus

Has completado las **20 lecciones** del laboratorio de Go. Ahora es momento de **consolidar** ese conocimiento construyendo herramientas de infraestructura que usarías en tu trabajo diario como Ingeniero de Software, DevOps o especialista en Ciberseguridad.

### 🎯 Filosofía de los Proyectos

| Principio | Descripción |
|:----------|:------------|
| 🔧 **Utilidad real** | Cada proyecto resuelve un problema que los ingenieros enfrentan en producción |
| 🏗️ **Arquitectura profesional** | Código organizado, con tests, documentación y patrones probados |
| ⚡ **Concurrencia nativa** | Aprovechamos goroutines y channels para rendimiento real |
| 🌐 **Sistemas y redes** | Enfoque en infraestructura: red, seguridad, automatización, monitoreo |
| 🧱 **Progresión técnica** | Los proyectos van de menor a mayor complejidad, construyendo sobre los anteriores |

---

## 🗺️ Índice de Proyectos

<br>

---

### 🔴 [Proyecto 01 — Port Scanner Concurrente de Alta Velocidad](./01-port-scanner/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir un escáner de puertos TCP concurrente que pueda escanear miles de puertos en segundos usando goroutines. La herramienta detectará puertos abiertos, identificará servicios conocidos (HTTP, SSH, HTTPS, MySQL, etc.) y generará un reporte estructurado. Inspirado en herramientas como `nmap` y `masscan`.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Lanzar miles de escaneos de puertos en paralelo |
| **Channels** | Recopilar resultados de escaneo de forma segura |
| **`net` package** | `net.DialTimeout()` para conexiones TCP con timeout |
| **`sync.WaitGroup`** | Sincronizar la espera de todas las goroutines |
| **`flag` package** | CLI flags para host, rango de puertos, timeout, concurrencia |
| **`fmt` / `os`** | Formateo de salida y manejo de errores |
| **`time` package** | Medir latencia y timeouts de conexión |

**🌍 Utilidad Real**

- **Auditoría de seguridad**: Identificar puertos abiertos no autorizados en servidores
- **Inventarios de red**: Descubrir qué servicios corre cada máquina en una infraestructura
- **Diagnóstico**: Verificar conectividad y reglas de firewall
- **Pentesting**: Primera herramienta que usa cualquier auditor de seguridad

**📁 Estructura Propuesta**

```
01-port-scanner/
├── README.md          ← Documentación detallada
├── go.mod
├── main.go            ← Entry point y CLI
├── scanner.go         ← Lógica de escaneo concurrente
├── scanner_test.go    ← Tests y benchmarks
└── report.go          ← Formateo de resultados
```

</details>

---

### 🟠 [Proyecto 02 — Analizador de Logs de Seguridad en Tiempo Real](./02-log-analyzer/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir un analizador de logs que procese archivos de log en tiempo real (estilo `tail -f`), detectando patrones de ataques comunes: intentos de inyección SQL, fuerza bruta en endpoints de login, escaneo de directorios, y anomalías de tráfico. Generará alertas con nivel de severidad y estadísticas en tiempo real.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Procesamiento concurrente de múltiples archivos de log |
| **Channels** | Pipeline de detección: lectura → parsing → análisis → alertas |
| **`os` / `io` package** | Lectura de archivos, `os.Seek` para tail en tiempo real |
| **`regexp` package** | Detección de patrones de ataque con expresiones regulares |
| **`strings` package** | Parsing de líneas de log, extracción de campos |
| **`sync.RWMutex`** | Acceso concurrente seguro al mapa de estadísticas |
| **`time` package** | Ventanas de tiempo para detectar ráfagas de requests |
| **`encoding/json`** | Exportación de alertas y reportes en formato JSON |

**🌍 Utilidad Real**

- **SOC (Security Operations Center)**: Monitoreo de seguridad en tiempo real
- **WAF (Web Application Firewall)**: Detección de ataques contra aplicaciones web
- **Compliance**: Registro y auditoría de eventos de seguridad
- **DevSecOps**: Integración en pipelines de CI/CD para análisis de logs post-deploy

**📁 Estructura Propuesta**

```
02-log-analyzer/
├── README.md
├── go.mod
├── main.go            ← Entry point y CLI
├── parser.go          ← Parsing de formatos de log (Apache, Nginx, custom)
├── detector.go        ← Reglas de detección de ataques
├── detector_test.go   ← Tests de detección
├── alerter.go         ← Sistema de alertas por severidad
└── stats.go           ← Estadísticas en tiempo real
```

</details>

---

### 🟡 [Proyecto 03 — Servidor de Archivos Estáticos Concurrente](./03-file-server/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir un servidor HTTP de archivos estáticos concurrente desde cero (sin usar `http.FileServer` de la librería estándar como muleta). Implementará control de rutas, seguridad de directorios (prevención de path traversal), soporte para rangos de bytes (descargas parciales), logging de accesos, y compresión gzip.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Cada conexión HTTP se maneja en su propia goroutine (nativo de `net/http`) |
| **`net/http` package** | Servidor HTTP, routing manual, headers HTTP |
| **`os` / `io` package** | Lectura de archivos, `io.Copy` para streaming eficiente |
| **`path/filepath`** | Sanitización de rutas, prevención de path traversal |
| **`compress/gzip`** | Compresión de respuestas para reducir ancho de banda |
| **`sync` package** | Cache de archivos en memoria con `sync.RWMutex` |
| **`mime` package** | Detección automática de Content-Type |
| **`context`** | Timeouts en conexiones y graceful shutdown |

**🌍 Utilidad Real**

- **CDN interno**: Servir assets estáticos en infraestructura privada
- **Desarrollo local**: Alternativa rápida a nginx para servir archivos
- **Edge computing**: Servidor ligero para entornos con recursos limitados
- **Aprendizaje**: Entender cómo funciona HTTP internamente sin frameworks

**📁 Estructura Propuesta**

```
03-file-server/
├── README.md
├── go.mod
├── main.go            ← Entry point y configuración del servidor
├── server.go          ← Lógica del servidor HTTP
├── server_test.go     ← Tests del servidor
├── handler.go         ← Manejo de requests y respuestas
├── security.go        ← Validación de rutas y seguridad
└── middleware.go       ← Logging, gzip, headers de seguridad
```

</details>

---

### 🟢 [Proyecto 04 — CLI de Inventario de Red Concurrente](./04-network-inventory/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir una herramienta CLI de inventario de red que escanee un rango de IPs (CIDR), haga ping concurrente para descubrir hosts activos, resuelva nombres DNS, detecte el sistema operativo aproximado (por TTL), y exporte el inventario completo en formato JSON y tabla ASCII.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Ping concurrente a cientos/miles de hosts simultáneamente |
| **Channels** | Recopilar resultados de discovery sin condiciones de carrera |
| **`net` package** | `net.LookupAddr()` para DNS reverso, `net.Interfaces()` para info local |
| **`os/exec` package** | Ejecutar `ping` del sistema operativo |
| **`sync.WaitGroup`** | Sincronizar el pool de escaneo |
| **`encoding/json`** | Exportar inventario a JSON |
| **`flag` package** | CLI flags para rango CIDR, timeout, concurrencia |
| **`net/netip`** | Parsing y manipulación de rangos CIDR (Go 1.18+) |

**🌍 Utilidad Real**

- **DevOps**: Inventario automático de infraestructura en data centers
- **Seguridad**: Descubrir hosts no autorizados en la red
- **Operaciones**: Documentar topología de red de forma automática
- **Compliance**: Verificar que solo hosts autorizados están activos

**📁 Estructura Propuesta**

```
04-network-inventory/
├── README.md
├── go.mod
├── main.go            ← Entry point y CLI
├── discovery.go       ← Lógica de descubrimiento de hosts
├── discovery_test.go  ← Tests
├── resolver.go        ← Resolución DNS y detección de OS
├── exporter.go        ← Exportación a JSON y tabla
└── cidr.go            ← Parsing y generación de rangos CIDR
```

</details>

---

### 🔵 [Proyecto 05 — Monitor de Recursos del Sistema (CPU / Memoria / Disco)](./05-system-monitor/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir un monitor de recursos del sistema que muestre en tiempo real el uso de CPU, memoria RAM, disco y red. Funcionará como un dashboard en terminal (TUI) con actualización periódica, alertas configurables por umbrales, y exportación de métricas en formato compatible con Prometheus.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Recolectores paralelos de métricas de CPU, memoria, disco y red |
| **Channels** | Pipeline de recolección → procesamiento → visualización |
| **`os` package** | Lectura de `/proc/stat`, `/proc/meminfo` en Linux, o APIs de Windows |
| **`time` package** | Tickers para actualización periódica de métricas |
| **`encoding/json`** | Exportación de métricas en formato JSON/Prometheus |
| **`sync` package** | Acceso seguro a las métricas compartidas |
| **`math` package** | Cálculos de porcentajes y promedios |
| **`os/exec`** | Ejecutar comandos del sistema como `df`, `wmic` |

**🌍 Utilidad Real**

- **SRE/DevOps**: Monitoreo de servidores sin dependencias externas
- **Diagnóstico**: Identificar cuellos de botella de rendimiento
- **Alertas**: Notificar cuando un servidor se queda sin recursos
- **Capacity planning**: Recolectar métricas históricas para planificar escalamiento

**📁 Estructura Propuesta**

```
05-system-monitor/
├── README.md
├── go.mod
├── main.go            ← Entry point y CLI
├── cpu.go             ← Recolector de métricas de CPU
├── memory.go          ← Recolector de métricas de memoria
├── disk.go            ← Recolector de métricas de disco
├── network.go         ← Recolector de métricas de red
├── monitor_test.go    ← Tests
├── alerter.go         ← Sistema de alertas por umbrales
└── exporter.go        ← Exportación a JSON y Prometheus
```

</details>

---

### 🟣 [Proyecto 06 — Clonador de Directorios con Sincronización Incremental](./06-dir-sync/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir una herramienta de sincronización de directorios (estilo `rsync`) que copie solo los archivos nuevos o modificados, usando hashes MD5/SHA256 para detección de cambios. Soportará copia concurrente, preservación de permisos, exclusiones por patrón, y generación de un log detallado de cambios.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Copia concurrente de múltiples archivos en paralelo |
| **Channels** | Pipeline de escaneo → comparación → copia → verificación |
| **`os` / `io` package** | Recorrido recursivo de directorios, copia de archivos |
| **`path/filepath`** | `filepath.Walk` para recorrer árboles de directorios |
| **`crypto/md5` / `crypto/sha256`** | Cálculo de hashes para detectar cambios |
| **`sync.WaitGroup`** | Sincronizar workers de copia |
| **`flag` package** | CLI flags para origen, destino, exclusiones, concurrencia |
| **`io.Copy` / `io.MultiWriter`** | Copia eficiente y cálculo de hash simultáneo |

**🌍 Utilidad Real**

- **Backups**: Sincronización incremental de datos críticos
- **Deployments**: Distribución de archivos a múltiples servidores
- **Disaster recovery**: Replica de directorios entre data centers
- **CI/CD**: Sincronización de artefactos de build

**📁 Estructura Propuesta**

```
06-dir-sync/
├── README.md
├── go.mod
├── main.go            ← Entry point y CLI
├── scanner.go         ← Escaneo recursivo de directorios
├── comparator.go      ← Comparación por hashes y timestamps
├── comparator_test.go ← Tests
├── copier.go          ← Copia concurrente de archivos
└── report.go          ← Generación de reporte de cambios
```

</details>

---

### 🟤 [Proyecto 07 — Proxy HTTP/HTTPS Reverso Simple](./07-reverse-proxy/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir un proxy HTTP/HTTPS reverso que distribuya tráfico entre múltiples backends usando algoritmos de balanceo de carga (Round Robin, Least Connections, Weighted). Implementará health checks automáticos, headers de proxy (`X-Forwarded-For`), logging de requests, y graceful shutdown.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Health checks periódicos en background, cada request en su goroutine |
| **`net/http/httputil`** | `ReverseProxy` como base para el forwarding |
| **`net/http` package** | Servidor HTTP, middleware, manipulación de headers |
| **`net/url` package** | Parsing y reescritura de URLs de backend |
| **`context`** | Timeouts y cancellation de requests al backend |
| **`sync` package** | Atomic operations para contadores de Round Robin |
| **`time` package** | Health checks periódicos, timeouts de conexión |
| **`crypto/tls`** | Soporte HTTPS con certificados TLS |

**🌍 Utilidad Real**

- **Arquitectura de microservicios**: Punto de entrada único para múltiples servicios
- **High Availability**: Distribución de carga para evitar puntos únicos de fallo
- **Desarrollo**: Simular entornos de producción localmente
- **Seguridad**: Ocultar la topología interna de la infraestructura

**📁 Estructura Propuesta**

```
07-reverse-proxy/
├── README.md
├── go.mod
├── main.go            ← Entry point y configuración
├── proxy.go           ← Lógica del proxy reverso
├── proxy_test.go      ← Tests
├── balancer.go        ← Algoritmos de balanceo de carga
├── healthcheck.go     ← Health checks de backends
└── middleware.go       ← Logging, headers, rate limiting
```

</details>

---

### ⚫ [Proyecto 08 — Servidor DNS Ligero con Cache y Filtros](./08-dns-server/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir un servidor DNS recursivo/cache que resuelva consultas DNS, implemente cache con TTL configurable, soporte listas de bloqueo (ad-blocking, malware domains), y registre todas las consultas para análisis posterior. Inspirado en `Pi-hole` y `dnsmasq`.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Manejo concurrente de consultas DNS entrantes |
| **Channels** | Pipeline de consulta → cache lookup → resolución → respuesta |
| **`net` package** | Conexiones UDP/TCP para protocolo DNS |
| **`encoding/binary`** | Parsing del protocolo DNS binario (headers, questions, answers) |
| **`sync.RWMutex`** | Cache DNS thread-safe |
| **`time` package** | TTL management, expiración de cache |
| **`os` package** | Carga de listas de bloqueo desde archivos |
| **`context`** | Timeouts en resoluciones DNS upstream |

**🌍 Utilidad Real**

- **Seguridad de red**: Bloquear dominios de malware y phishing a nivel de red
- **Privacidad**: Evitar que el ISP vea tu historial DNS
- **Performance**: Cache local reduce latencia de resolución DNS
- **Control parental**: Filtrar contenido por categorías de dominio

**📁 Estructura Propuesta**

```
08-dns-server/
├── README.md
├── go.mod
├── main.go            ← Entry point y configuración
├── server.go          ← Servidor DNS UDP/TCP
├── server_test.go     ← Tests
├── protocol.go        ← Parsing y construcción de paquetes DNS
├── cache.go           ← Cache con TTL y eviction
├── resolver.go        ← Resolución DNS recursiva upstream
└── blocklist.go       ← Sistema de filtrado por listas
```

</details>

---

### 🔶 [Proyecto 09 — Orquestador de Contenedores Simplificado (Mini-Docker)](./09-container-orchestrator/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir un orquestador de procesos que gestione "contenedores" (procesos aislados) con reinicio automático, health checks, logging centralizado, y balanceo de carga entre instancias. Usará namespaces de Linux y cgroups para aislamiento básico, simulando los conceptos fundamentales de Docker/Kubernetes.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Monitoreo concurrente de múltiples procesos/contenedores |
| **Channels** | Señales de salud, eventos de muerte de proceso, logs centralizados |
| **`os/exec` package** | Lanzamiento y gestión de procesos hijos |
| **`os` / `syscall`** | Namespaces, signals (`SIGTERM`, `SIGKILL`), process groups |
| **`encoding/json`** | Configuración de servicios en formato JSON/YAML |
| **`sync` package** | Estado compartido de los contenedores |
| **`context`** | Lifecycle management con cancellation |
| **`net/http`** | API REST para gestionar contenedores |

**🌍 Utilidad Real**

- **DevOps**: Entender los fundamentos de containerización y orquestación
- **Supervisión de procesos**: Alternativa a `systemd`, `supervisord` para servicios Go
- **Microservicios**: Gestionar pools de workers con reinicio automático
- **Educación**: Comprender cómo funciona Docker/Kubernetes por dentro

**📁 Estructura Propuesta**

```
09-container-orchestrator/
├── README.md
├── go.mod
├── main.go            ← Entry point y CLI/API
├── orchestrator.go    ← Lógica de orquestación
├── orchestrator_test.go ← Tests
├── container.go       ← Modelo de contenedor/proceso
├── healthcheck.go     ← Health checks de contenedores
├── restart.go         ← Políticas de reinicio (always, on-failure, never)
├── logger.go          ← Logging centralizado
└── api.go             ← API REST de gestión
```

</details>

---

### 🔴 [Proyecto 10 — Herramienta de Pentesting: Bruteforce HTTP con Rate Limiting](./10-http-bruteforcer/)

<details>
<summary><strong>🔍 Expandir detalles del proyecto</strong></summary>

<br>

**🎯 Objetivo**

Construir una herramienta de auditoría de seguridad que realice ataques de fuerza bruta contra endpoints HTTP (login forms, APIs con autenticación básica). Implementará diccionarios de usuarios/contraseñas, detección automática de campos de login, rate limiting configurable para evitar detección, proxy rotation, y generación de reportes de hallazgos.

**🧰 Componentes de Go a Usar**

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Workers concurrentes para enviar requests de login en paralelo |
| **Channels** | Pool de workers con control de concurrencia, canal de resultados |
| **`net/http` package** | Requests HTTP, manejo de cookies y sesiones |
| **`net/http/cookiejar`** | Persistencia de cookies entre requests de sesión |
| **`html` / `regexp`** | Parsing del HTML para detectar formularios de login |
| **`os` package** | Carga de diccionarios de usuarios y contraseñas |
| **`sync` package** | Contadores atómicos de intentos y éxitos |
| **`time` package** | Rate limiting, delays configurables entre intentos |
| **`context`** | Timeout global y cancellation del ataque |

**🌍 Utilidad Real**

- **Pentesting**: Auditoría de seguridad de aplicaciones web
- **Red Team**: Evaluación de la fortaleza de credenciales
- **Compliance**: Verificar que los sistemas cumplen políticas de contraseñas
- **Blue Team**: Entender técnicas de ataque para mejorar defensas

**⚠️ Nota Ética**

> Esta herramienta es **exclusivamente para uso legítimo en auditorías de seguridad autorizadas**. Úsala solo en sistemas que poseas o para los que tengas autorización escrita. El acceso no autorizado a sistemas informáticos es ilegal.

**📁 Estructura Propuesta**

```
10-http-bruteforcer/
├── README.md
├── go.mod
├── main.go            ← Entry point y CLI
├── bruteforcer.go     ← Lógica de fuerza bruta concurrente
├── bruteforcer_test.go ← Tests
├── detector.go        ← Detección automática de formularios de login
├── httpclient.go      ← Cliente HTTP con proxy rotation y cookies
├── ratelimiter.go     ← Rate limiter por tiempo y por request
├── wordlist.go        ← Carga y combinación de diccionarios
└── reporter.go        ← Generación de reportes de hallazgos
```

</details>

---

## 📊 Resumen de Proyectos

```
┌──────────────────────────────────────────────────────────────────────┐
│                                                                      │
│  🔴 Proyecto 01: Port Scanner Concurrente         Red · Seguridad   │
│  🟠 Proyecto 02: Analizador de Logs               Seguridad · SIEM  │
│  🟡 Proyecto 03: Servidor de Archivos             Red · HTTP        │
│  🟢 Proyecto 04: Inventario de Red                Red · Automatizar │
│  🔵 Proyecto 05: Monitor de Sistema               Sistemas · OS     │
│  🟣 Proyecto 06: Clonador de Directorios          Sistemas · IO     │
│  🟤 Proyecto 07: Proxy Reverso                    Red · Arquitectura│
│  ⚫ Proyecto 08: Servidor DNS                     Red · Seguridad   │
│  🔶 Proyecto 09: Orquestador de Contenedores      Sistemas · DevOps │
│  🔴 Proyecto 10: HTTP Bruteforcer                 Seguridad · Red   │
│                                                                      │
│  Conceptos transversales en TODOS los proyectos:                     │
│  ✅ Goroutines y Channels          ✅ Tests y Benchmarks             │
│  ✅ Manejo de errores idiomático   ✅ CLI con flags                  │
│  ✅ Estructura de paquetes Go      ✅ Documentación                  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 🗓️ Plan de Desarrollo

Los proyectos se desarrollarán **secuencialmente** (uno a la vez) para garantizar máxima profundidad técnica:

| Paso | Proyecto | Estado |
|:-----|:---------|:-------|
| 1 | 🔴 Port Scanner Concurrente | ⏳ Pendiente |
| 2 | 🟠 Analizador de Logs de Seguridad | ⏳ Pendiente |
| 3 | 🟡 Servidor de Archivos Estáticos | ⏳ Pendiente |
| 4 | 🟢 CLI de Inventario de Red | ⏳ Pendiente |
| 5 | 🔵 Monitor de Recursos del Sistema | ⏳ Pendiente |
| 6 | 🟣 Clonador de Directorios | ⏳ Pendiente |
| 7 | 🟤 Proxy HTTP/HTTPS Reverso | ⏳ Pendiente |
| 8 | ⚫ Servidor DNS Ligero | ⏳ Pendiente |
| 9 | 🔶 Orquestador de Contenedores | ⏳ Pendiente |
| 10 | 🔴 HTTP Bruteforcer | ⏳ Pendiente |

---

## ⚡ Requisitos Previos

- Haber completado las **20 lecciones** del laboratorio de Go
- **Go 1.21+** instalado
- Conocimientos básicos de redes TCP/IP
- Terminal con permisos de administrador (algunos proyectos requieren acceso a puertos privilegiados)

---

<div align="center">

### *"En teoría, no hay diferencia entre la teoría y la práctica. En práctica, sí la hay."*
### — **Yogi Berra**

<br>

**¡Comienza con el Proyecto 01 y construye herramientas reales! 🛠️**

</div>