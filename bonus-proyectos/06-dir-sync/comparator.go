package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Action determina la acción a tomar para un archivo durante la sincronización.
type Action int

const (
	ActionCopy  Action = iota // Archivo nuevo o modificado → copiar
	ActionSkip                // Archivo idéntico → omitir
	ActionError               // Error al procesar
)

func (a Action) String() string {
	switch a {
	case ActionCopy:
		return "COPY"
	case ActionSkip:
		return "SKIP"
	case ActionError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// FileDecision representa la decisión tomada para un archivo tras la comparación.
type FileDecision struct {
	SrcFile  FileInfo
	DstFile  FileInfo
	Action   Action
	Reason   string
	SrcHash  string
	DstHash  string
}

// CompareResult contiene el resultado completo de la comparación entre origen y destino.
type CompareResult struct {
	Decisions   []FileDecision
	ToCopy      []FileDecision
	ToSkip      []FileDecision
	Errors      []FileDecision
	TotalSrcSize int64
	SavedSize    int64 // Bytes ahorrados por no copiar archivos idénticos
}

// CompareDirectories compara los archivos del origen contra los del destino.
// Usa una estrategia de dos pasos:
//   1. Comparación rápida por tamaño y timestamp
//   2. Verificación definitiva con hash SHA-256 (solo si es necesario)
func CompareDirectories(srcResult, dstResult *ScanResult) *CompareResult {
	result := &CompareResult{
		Decisions: make([]FileDecision, 0, len(srcResult.Files)),
	}

	// Indexar archivos destino por ruta relativa para búsqueda O(1).
	dstIndex := make(map[string]FileInfo, len(dstResult.Files))
	for _, f := range dstResult.Files {
		dstIndex[f.RelPath] = f
	}

	for _, srcFile := range srcResult.Files {
		result.TotalSrcSize += srcFile.Size

		decision := compareFile(srcFile, dstIndex)

		switch decision.Action {
		case ActionCopy:
			result.ToCopy = append(result.ToCopy, decision)
		case ActionSkip:
			result.ToSkip = append(result.ToSkip, decision)
			result.SavedSize += srcFile.Size
		case ActionError:
			result.Errors = append(result.Errors, decision)
		}

		result.Decisions = append(result.Decisions, decision)
	}

	return result
}

// compareFile evalúa un archivo fuente contra su contraparte en destino.
// Aplica optimizaciones: primero tamaño, luego hash SHA-256 por streaming.
func compareFile(srcFile FileInfo, dstIndex map[string]FileInfo) FileDecision {
	dstFile, exists := dstIndex[srcFile.RelPath]

	decision := FileDecision{
		SrcFile: srcFile,
		Action:  ActionCopy,
	}

	// Caso 1: El archivo no existe en destino → copiar.
	if !exists {
		decision.Reason = "archivo nuevo en origen"
		return decision
	}

	decision.DstFile = dstFile

	// Caso 2: Diferencia de tamaño → copiar sin necesidad de hash.
	if srcFile.Size != dstFile.Size {
		decision.Reason = fmt.Sprintf("tamaño diferente (origen: %d bytes, destino: %d bytes)", srcFile.Size, dstFile.Size)
		return decision
	}

	// Caso 3: Mismo tamaño → verificar hash SHA-256 por streaming.
	srcHash, err := computeFileHash(srcFile.Path)
	if err != nil {
		decision.Action = ActionError
		decision.Reason = fmt.Sprintf("error calculando hash origen: %v", err)
		return decision
	}
	decision.SrcHash = srcHash

	dstHash, err := computeFileHash(dstFile.Path)
	if err != nil {
		decision.Action = ActionError
		decision.Reason = fmt.Sprintf("error calculando hash destino: %v", err)
		return decision
	}
	decision.DstHash = dstHash

	// Comparar hashes.
	if srcHash == dstHash {
		decision.Action = ActionSkip
		decision.Reason = "contenido idéntico (hash SHA-256 coincidente)"
	} else {
		decision.Action = ActionCopy
		decision.Reason = "contenido modificado (hash SHA-256 diferente)"
	}

	return decision
}

// computeFileHash calcula el hash SHA-256 de un archivo usando streaming
// con io.Copy para no cargar el archivo completo en memoria.
func computeFileHash(filePath string) (string, error) {
	f, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return "", fmt.Errorf("abriendo archivo %q: %w", filePath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("leyendo archivo %q para hash: %w", filePath, err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}