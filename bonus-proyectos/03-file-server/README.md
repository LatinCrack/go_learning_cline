<div align="center">

# 🟡 Proyecto 03 — Servidor de Archivos Estáticos Concurrente

### *HTTP Server Nativo · Sin http.FileServer · Streaming · Gzip · Seguridad de Rutas*

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![HTTP](https://img.shields.io/badge/Protocol-HTTP/1.1-4ECDC4&style=for-the-badge) ![Security](https://img.shields.io/badge/Security-Anti_Path_Traversal-FF6B6B&style=for-the-badge) ![Compression](https://img.shields.io/badge/Compression-Gzip-FFE66D&style=for-the-badge)

</div>

---

## 📖 Descripción

Servidor HTTP de archivos estáticos construido desde cero en Go, **sin usar `http.FileServer`** como solución directa. Implementa un pipeline completo de middleware con compresión gzip, seguridad contra ataques de path traversal, detección automática de Content-Type, y streaming eficiente mediante `io.Copy`.

---

## 🚀 Inicio Rápido

### Compilar

```bash
go build -o file-server .
```

### Ejecutar

```bash
# Valores por defecto: puerto 8080, carpeta ./public
./file-server

# Especificar puerto y carpeta raíz
./file-server -port 3000 -root ./mi-sitio

# Especificar host, puerto y carpeta
./file-server -host 127.0.0.1 -port 8080 -root /var/www/html

# Ver versión
./file-server -version
```

### Flags Disponibles

| Flag        | Descripción                              | Valor por defecto |
|:------------|:-----------------------------------------|:------------------|
| `-port`     | Puerto TCP para escuchar                 | `8080`            |
| `-root`     | Directorio raíz de archivos estáticos    | `./public`        |
| `-host`     | Dirección de red para vincular           | `0.0.0.0`         |
| `-version`  | Mostrar información de versión           | `false`           |

---

## 🧪 Verificar Compresión Gzip

```bash
# Crear un directorio de prueba
mkdir -p public
echo "<html><body>Hola Mundo</body></html>" > public/index.html

# Compilar y ejecutar
go build -o file-server . && ./file-server -port 8080 -root ./public

# En otra terminal, verificar la cabecera Content-Encoding: gzip
curl -sI -H "Accept-Encoding: gzip" http://localhost:8080/index.html

# Salida esperada (parcial):
#   HTTP/1.1 200 OK
#   Content-Type: text/html; charset=utf-8
#   Content-Encoding: gzip
#   Vary: Accept-Encoding
#   X-Content-Type-Options: nosniff
#   X-Frame-Options: DENY
```

### Sin compresión (sin Accept-Encoding)

```bash
curl -sI http://localhost:8080/index.html

# Salida: NO incluye "Content-Encoding: gzip"
```

---

## 🔒 Seguridad

### Anti-Path Traversal

El servidor bloquea intentos de acceso fuera del directorio raíz:

```bash
# Bloqueado → 403 Forbidden
curl http://localhost:8080/../../../etc/passwd
curl http://localhost:8080/..%2F..%2Fetc/passwd

# Archivos ocultos → 403 Forbidden
curl http://localhost:8080/.env
curl http://localhost:8080/.git/config
```

### Cabeceras de Seguridad

Todas las respuestas incluyen automáticamente:

| Cabecera                     | Valor                             |
|:-----------------------------|:----------------------------------|
| `X-Content-Type-Options`    | `nosniff`                         |
| `X-Frame-Options`           | `DENY`                            |
| `X-XSS-Protection`          | `1; mode=block`                   |
| `Referrer-Policy`            | `strict-origin-when-cross-origin` |

---

## 📁 Estructura del Proyecto

```
03-file-server/
├── README.md          ← Este archivo de documentación
├── go.mod             ← Módulo Go
├── main.go            ← Entry point, CLI con flags, banner
├── server.go          ← Configuración del servidor, lifecycle, graceful shutdown
├── handler.go         ← Manejo de requests, streaming con io.Copy, MIME types
├── security.go        ← Validación de rutas (filepath.Clean, filepath.Rel)
├── middleware.go       ← Gzip, logging, cabeceras de seguridad
└── server_test.go     ← Tests unitarios e integración
```

---

## 🧰 Paquetes de Go Utilizados

| Paquete           | Uso                                             |
|:------------------|:------------------------------------------------|
| `net/http`        | Servidor HTTP nativo, routing, headers           |
| `os` / `io`       | `os.Open` + `io.Copy` para streaming eficiente   |
| `path/filepath`   | `filepath.Clean` y `filepath.Rel` anti-traversal |
| `compress/gzip`   | Compresión al vuelo con `sync.Pool`              |
| `mime`            | Detección dinámica de `Content-Type`             |
| `context`         | Timeout en graceful shutdown                     |
| `sync`            | Pool de writers gzip reutilizables               |
| `log`             | Logging de accesos y eventos de seguridad        |

---

## ▶️ Ejecutar Tests

```bash
go test -v -cover ./...
```

Los tests cubren:
- **Seguridad**: ataques de path traversal, archivos ocultos, null bytes
- **Handler**: servir archivos, detección MIME, directorios, HEAD requests
- **Middleware**: compresión gzip, cabeceras de seguridad, logging
- **Integración**: cadena completa de middleware + handler
- **Servidor**: configuración, validación de directorio raíz

---

<div align="center">

**"No uses `http.FileServer` — constrúyelo tú mismo y entiende cómo funciona HTTP por dentro."**

</div>