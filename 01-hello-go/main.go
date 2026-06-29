package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	// ═══════════════════════════════════════════════════════════
	// 🔬 LABORATORIO GO — Detector de Entorno de Desarrollo
	// ═══════════════════════════════════════════════════════════
	// Este CLI mínimo actúa como tu "panel de instrumentos":
	// muestra toda la información crítica del entorno donde
	// se ejecuta tu código Go. Úsalo para diagnosticar
	// problemas a lo largo de todo el curso.
	// ═══════════════════════════════════════════════════════════

	// ── Cabecera visual ──────────────────────────────────────
	fmt.Println("╔══════════════════════════════════════════════╗")
	fmt.Println("║   🧪 Detector de Entorno — Laboratorio Go   ║")
	fmt.Println("╚══════════════════════════════════════════════╝")
	fmt.Println()

	// ── 1. Sistema Operativo ─────────────────────────────────
	// runtime.GOOS devuelve una cadena con el nombre del OS
	// donde se está ejecutando el binario compilado.
	// Valores comunes: "linux", "darwin", "windows", "freebsd"
	osName := runtime.GOOS

	// ── 2. Arquitectura del procesador ───────────────────────
	// runtime.GOARCH devuelve la arquitectura de la CPU.
	// Valores comunes: "amd64" (64-bit x86), "arm64" (Apple Silicon/M1),
	// "386" (32-bit x86), "arm" (ARM 32-bit)
	arch := runtime.GOARCH

	// ── 3. Versión de Go ─────────────────────────────────────
	// runtime.Version() devuelve la versión del runtime de Go
	// con la que fue compilado el binario.
	// Ejemplo: "go1.22.0"
	goVersion := runtime.Version()

	// ── 4. Número de CPUs disponibles ────────────────────────
	// runtime.NumCPU() devuelve el número de CPUs lógicas
	// que el sistema tiene disponibles. Este valor influye
	// directamente en GOMAXPROCS (lo veremos en concurrencia).
	numCPU := runtime.NumCPU()

	// ── 5. Variables de entorno Go ───────────────────────────
	// os.Getenv lee variables de entorno del sistema operativo.
	// GOPATH: directorio de trabajo donde Go descarga dependencias
	// GOROOT: directorio de instalación del compilador Go
	// GOFLAGS: flags globales para los comandos go
	gopath := os.Getenv("GOPATH")
	goroot := os.Getenv("GOROOT")

	// Si GOPATH no está definido, Go usa un valor por defecto
	if gopath == "" {
		gopath = "(valor por defecto del sistema — ~/.go o ~/go)"
	}

	// Si GOROOT no está definido, intentamos con runtime.GOROOT()
	if goroot == "" {
		goroot = runtime.GOROOT()
	}

	// ── 6. Directorio de trabajo actual ──────────────────────
	// os.Getwd() devuelve el directorio de trabajo actual
	// del proceso. Es útil para debuggear problemas de paths.
	wd, err := os.Getwd()
	if err != nil {
		wd = "(no se pudo determinar)"
	}

	// ── 7. PID del proceso actual ────────────────────────────
	// os.Getpid() devuelve el Process ID del programa en ejecución.
	// Útil para logging y diagnóstico de procesos.
	pid := os.Getpid()

	// ── Impresión formateada del reporte ─────────────────────
	fmt.Printf("  💻 Sistema Operativo : %s\n", osName)
	fmt.Printf("  🏗️  Arquitectura     : %s\n", arch)
	fmt.Printf("  🔧 Versión de Go     : %s\n", goVersion)
	fmt.Printf("  🧠 CPUs disponibles  : %d\n", numCPU)
	fmt.Printf("  📂 GOPATH            : %s\n", gopath)
	fmt.Printf("  📂 GOROOT            : %s\n", goroot)
	fmt.Printf("  📂 Directorio actual : %s\n", wd)
	fmt.Printf("  🆔 PID del proceso   : %d\n", pid)
	fmt.Println()
	fmt.Println("  ✅ Entorno listo. ¡Tu laboratorio Go está operativo!")
}