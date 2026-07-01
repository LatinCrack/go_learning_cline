<div align="center">

# 🟣 Proyecto 06 — Clonador de Directorios con Sincronización Incremental

### *Backup inteligente · SHA-256 streaming · Copia concurrente*

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Concepto](https://img.shields.io/badge/Concepto-Sistemas_%26_IO-FF6B6B?style=for-the-badge) ![Concurrencia](https://img.shields.io/badge/Concurrencia-Goroutines-4ECDC4&style=for-the-badge)

<br>

> *"No copies todo de nuevo. Solo lo que cambió."*

</div>

---

## 🎯 Objetivo

Construir una herramienta de sincronización de directorios (estilo `rsync`) que copie **solo los archivos nuevos o modificados**, usando hashes SHA-256 para detección de cambios. Soporta copia concurrente, preservación de permisos, y generación de un log detallado de cambios.

---

## 🧰 Stack Técnico

| Componente | Uso en el Proyecto |
|:-----------|:-------------------|
| **Goroutines** | Pool de workers de copia concurrente |
| **Channels** | Pipeline de escaneo → comparación → copia → reporte |
| **`os` / `io` package** | Escaneo recursivo de directorios, copia de archivos |
| **`path/filepath`** | `filepath.WalkDir` para recorrer árboles de directorios |
| **`crypto/sha256`** | Cálculo de hashes para detectar cambios |
| **`io.Copy`** | Streaming eficiente: hash y copia sin cargar archivos en memoria |
| **`sync.WaitGroup`** | Sincronización de workers de copia |
| **`flag` package** | CLI flags para origen, destino y concurrencia |

---

## 📁 Estructura del Proyecto

```
06-dir-sync/
├── README.md              ← Este archivo de documentación
├── go.mod                 ← Módulo Go
├── main.go                ← Entry point y orquestación del pipeline
├── scanner.go             ← Escaneo recursivo de directorios (filepath.WalkDir)
├── comparator.go          ← Comparación criptográfica SHA-256 por streaming
├── copier.go              ← Pipeline de copia concurrente con workers
├── report.go              ← Generación de reporte en consola y log de texto
└── comparator_test.go     ← Tests unitarios completos (scanner, hash, comparación, copia)
```

---

## ⚙️ Flags Disponibles

| Flag | Tipo | Default | Descripción |
|:-----|:-----|:--------|:------------|
| `--source` | `string` | `""` (requerido) | Ruta del directorio origen a sincronizar |
| `--target` | `string` | `""` (requerido) | Ruta del directorio destino (backup) |
| `--concurrency` | `int` | `4` | Número de workers de copia concurrentes |
| `--dry-run` | `bool` | `false` | Mostrar qué se haría sin copiar archivos |
| `--verbose` | `bool` | `false` | Mostrar detalle de cada archivo analizado |
| `--version` | `bool` | `false` | Mostrar versión y salir |

---

## 🚀 Ejemplos de Uso

### Backup incremental de logs

```bash
go run . --source /var/log --target /backup/logs --concurrency 8
```

### Sincronizar scripts con modo verbose

```bash
go run . --source ./scripts --target ./backup/scripts --verbose
```

### Preview sin copiar (dry-run)

```bash
go run . --source ./proyecto --target ./backup/proyecto --dry-run --verbose
```

### Copia secuencial (1 worker)

```bash
go run . --source ./datos --target /mnt/nfs/datos --concurrency 1
```

---

## 🔍 Flujo Interno del Pipeline

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────┐
│   Escaneo   │────▶│  Comparación │────▶│     Copia    │────▶│ Reporte  │
│  (scanner)  │     │ (comparator) │     │   (copier)   │     │ (report) │
└─────────────┘     └──────────────┘     └──────────────┘     └──────────┘
     │                    │                     │                    │
 filepath.WalkDir    SHA-256 streaming     Pool de workers      Consola +
     │              io.Copy → hash       sync.WaitGroup         Log file
     ▼                    ▼                     ▼                    ▼
  []FileInfo      []FileDecision          CopyStats           SyncReport
```

### Fase 1 — Escaneo (`scanner.go`)
- Recorre recursivamente origen y destino con `filepath.WalkDir`.
- Extrae: ruta relativa, tamaño, permisos, timestamp de modificación.

### Fase 2 — Comparación (`comparator.go`)
- **Paso rápido**: Si el tamaño difiere → marcar para copiar (sin hash).
- **Paso definitivo**: Si el tamaño es igual → calcular SHA-256 por streaming con `io.Copy` → `sha256.New()`.
- Si los hashes coinciden → omitir archivo (ahorro de ancho de banda).

### Fase 3 — Copia (`copier.go`)
- Pool de N goroutines alimentado por canal de trabajos.
- Copia atómica: escribe a `.tmp` → `Sync()` → `Rename()`.
- Preserva permisos originales con `os.Chmod()`.

### Fase 4 — Reporte (`report.go`)
- Resumen en consola: archivos copiados, omitidos, eficiencia incremental.
- Log detallado en `.dir-sync/sync-YYYY-MM-DD-HHMMSS.log`.

---

## 🧪 Tests

Ejecutar la suite completa de tests:

```bash
cd bonus-proyectos/06-dir-sync
go test -v ./...
```

Los tests cubren:
- ✅ Escaneo de directorios vacíos, con archivos y subdirectorios anidados
- ✅ Manejo de errores (rutas inexistentes, archivos regulares)
- ✅ Preservación de permisos de archivos
- ✅ Cálculo de hash SHA-256 correcto (contra valor esperado)
- ✅ Hash de contenido idéntico vs diferente
- ✅ Comparación: archivos nuevos, idénticos, modificados
- ✅ Optimización: salto de hash cuando el tamaño difiere
- ✅ Cálculo de bytes ahorrados (SavedSize)
- ✅ Copia concurrente con múltiples workers
- ✅ Preservación de permisos en destino
- ✅ Formateo de bytes legible (bytes, KB, MB, GB)

---

## 🌍 Utilidad Real

| Escenario | Uso |
|:----------|:----|
| **Backups** | Sincronización incremental de datos críticos a disco externo o NAS |
| **Deployments** | Distribuir solo archivos modificados a servidores |
| **Disaster Recovery** | Replicar directorios entre data centers minimizando transferencia |
| **CI/CD** | Sincronizar artefactos de build entre etapas del pipeline |
| **Logs** | Respaldar directorios de logs consumiendo mínimo ancho de banda |

---

<div align="center">

**⬅️ [Proyecto 05 — Monitor de Sistema](../05-system-monitor/) · · · · · [Proyecto 07 — Proxy Reverso](../07-reverse-proxy/) ➡️**

</div>