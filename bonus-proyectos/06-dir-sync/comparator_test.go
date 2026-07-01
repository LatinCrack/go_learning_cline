package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// --- Helpers para tests ---

// createTempFile crea un archivo temporal con contenido específico y retorna su path.
func createTempFile(t *testing.T, dir, name, content string, perm os.FileMode) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("error creando directorio para %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatalf("error creando archivo temporal %q: %v", path, err)
	}
	return path
}

// setupTestDirs crea directorios de origen y destino temporales con archivos de prueba.
func setupTestDirs(t *testing.T) (srcDir, dstDir string, cleanup func()) {
	t.Helper()
	srcDir = t.TempDir()
	dstDir = t.TempDir()
	return srcDir, dstDir, func() {} // t.TempDir() se limpia automáticamente
}

// --- Tests de ScanDirectory ---

func TestScanDirectory_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	result, err := ScanDirectory(dir)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(result.Files) != 0 {
		t.Errorf("esperaba 0 archivos, obtuvo %d", len(result.Files))
	}
	if result.TotalSize != 0 {
		t.Errorf("esperaba tamaño 0, obtuvo %d", result.TotalSize)
	}
}

func TestScanDirectory_WithFiles(t *testing.T) {
	srcDir, _, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "file1.txt", "hello", 0644)
	createTempFile(t, srcDir, "subdir/file2.txt", "world", 0644)
	createTempFile(t, srcDir, "subdir/deep/file3.txt", "deep content", 0644)

	result, err := ScanDirectory(srcDir)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(result.Files) != 3 {
		t.Errorf("esperaba 3 archivos, obtuvo %d", len(result.Files))
	}

	// Verificar que los archivos fueron encontrados.
	relPaths := make(map[string]bool)
	for _, f := range result.Files {
		relPaths[f.RelPath] = true
	}

	expected := []string{"file1.txt", "subdir/file2.txt", "subdir/deep/file3.txt"}
	for _, e := range expected {
		if !relPaths[e] {
			t.Errorf("archivo esperado %q no encontrado", e)
		}
	}
}

func TestScanDirectory_NonExistentDir(t *testing.T) {
	_, err := ScanDirectory("/ruta/que/no/existe/en/ningun/lado")
	if err == nil {
		t.Fatal("esperaba error para directorio inexistente, obtuvo nil")
	}
}

func TestScanDirectory_NotADirectory(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	_, err := ScanDirectory(tmpFile)
	if err == nil {
		t.Fatal("esperaba error para archivo (no directorio), obtuvo nil")
	}
}

func TestScanDirectory_PreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permisos Unix no aplican en Windows")
	}
	dir := t.TempDir()
	createTempFile(t, dir, "script.sh", "#!/bin/bash\necho ok", 0755)
	createTempFile(t, dir, "data.txt", "data", 0644)

	result, err := ScanDirectory(dir)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	for _, f := range result.Files {
		switch f.RelPath {
		case "script.sh":
			if f.Mode.Perm() != 0755 {
				t.Errorf("esperaba permisos 0755, obtuvo %o", f.Mode.Perm())
			}
		case "data.txt":
			if f.Mode.Perm() != 0644 {
				t.Errorf("esperaba permisos 0644, obtuvo %o", f.Mode.Perm())
			}
		}
	}
}

// --- Tests de computeFileHash ---

func TestComputeFileHash_SameContent(t *testing.T) {
	dir := t.TempDir()
	createTempFile(t, dir, "a.txt", "contenido identico", 0644)
	createTempFile(t, dir, "b.txt", "contenido identico", 0644)

	hashA, err := computeFileHash(filepath.Join(dir, "a.txt"))
	if err != nil {
		t.Fatalf("error calculando hash a: %v", err)
	}
	hashB, err := computeFileHash(filepath.Join(dir, "b.txt"))
	if err != nil {
		t.Fatalf("error calculando hash b: %v", err)
	}

	if hashA != hashB {
		t.Errorf("hashes deberían ser iguales: %s != %s", hashA, hashB)
	}
}

func TestComputeFileHash_DifferentContent(t *testing.T) {
	dir := t.TempDir()
	createTempFile(t, dir, "a.txt", "contenido A", 0644)
	createTempFile(t, dir, "b.txt", "contenido B", 0644)

	hashA, err := computeFileHash(filepath.Join(dir, "a.txt"))
	if err != nil {
		t.Fatalf("error calculando hash a: %v", err)
	}
	hashB, err := computeFileHash(filepath.Join(dir, "b.txt"))
	if err != nil {
		t.Fatalf("error calculando hash b: %v", err)
	}

	if hashA == hashB {
		t.Errorf("hashes deberían ser diferentes, ambos son: %s", hashA)
	}
}

func TestComputeFileHash_NonExistentFile(t *testing.T) {
	_, err := computeFileHash("/archivo/que/no/existe.txt")
	if err == nil {
		t.Fatal("esperaba error para archivo inexistente, obtuvo nil")
	}
}

func TestComputeFileHash_MatchesSHA256(t *testing.T) {
	dir := t.TempDir()
	content := "test content for sha256 verification"
	createTempFile(t, dir, "test.txt", content, 0644)

	hash, err := computeFileHash(filepath.Join(dir, "test.txt"))
	if err != nil {
		t.Fatalf("error calculando hash: %v", err)
	}

	// Calcular hash esperado manualmente.
	expected := sha256.Sum256([]byte(content))
	expectedStr := fmt.Sprintf("%x", expected)

	if hash != expectedStr {
		t.Errorf("hash no coincide con SHA-256 esperado:\n  obtenido:  %s\n  esperado:  %s", hash, expectedStr)
	}
}

// --- Tests de CompareDirectories ---

func TestCompareDirectories_NewFiles(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "new1.txt", "new file 1", 0644)
	createTempFile(t, srcDir, "new2.txt", "new file 2", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if len(result.ToCopy) != 2 {
		t.Errorf("esperaba 2 archivos para copiar, obtuvo %d", len(result.ToCopy))
	}
	if len(result.ToSkip) != 0 {
		t.Errorf("esperaba 0 archivos omitidos, obtuvo %d", len(result.ToSkip))
	}
}

func TestCompareDirectories_IdenticalFiles(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "same.txt", "same content", 0644)
	createTempFile(t, dstDir, "same.txt", "same content", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if len(result.ToSkip) != 1 {
		t.Errorf("esperaba 1 archivo omitido, obtuvo %d", len(result.ToSkip))
	}
	if len(result.ToCopy) != 0 {
		t.Errorf("esperaba 0 archivos para copiar, obtuvo %d", len(result.ToCopy))
	}
}

func TestCompareDirectories_ModifiedFiles(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "modified.txt", "new content", 0644)
	createTempFile(t, dstDir, "modified.txt", "old content", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if len(result.ToCopy) != 1 {
		t.Errorf("esperaba 1 archivo para copiar, obtuvo %d", len(result.ToCopy))
	}
	if result.ToCopy[0].Reason == "" {
		t.Error("esperaba razón de copia no vacía")
	}
}

func TestCompareDirectories_SizeDifferenceSkipsHash(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "diffsize.txt", "short", 0644)
	createTempFile(t, dstDir, "diffsize.txt", "much longer content here", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if len(result.ToCopy) != 1 {
		t.Errorf("esperaba 1 archivo para copiar, obtuvo %d", len(result.ToCopy))
	}
	// El reason debe mencionar tamaño, no hash.
	if result.ToCopy[0].SrcHash != "" {
		t.Error("no debería calcular hash cuando los tamaños difieren")
	}
}

func TestCompareDirectories_MixedScenarios(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	// Archivo nuevo.
	createTempFile(t, srcDir, "new.txt", "new", 0644)
	// Archivo idéntico.
	createTempFile(t, srcDir, "same.txt", "same", 0644)
	createTempFile(t, dstDir, "same.txt", "same", 0644)
	// Archivo modificado.
	createTempFile(t, srcDir, "changed.txt", "changed v2", 0644)
	createTempFile(t, dstDir, "changed.txt", "changed v1", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if len(result.ToCopy) != 2 {
		t.Errorf("esperaba 2 archivos para copiar (new + changed), obtuvo %d", len(result.ToCopy))
	}
	if len(result.ToSkip) != 1 {
		t.Errorf("esperaba 1 archivo omitido (same), obtuvo %d", len(result.ToSkip))
	}
}

func TestCompareDirectories_Subdirectories(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "sub/deep/file.txt", "nested content", 0644)
	createTempFile(t, dstDir, "sub/deep/file.txt", "nested content", 0644)
	createTempFile(t, srcDir, "sub/new.txt", "new nested", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if len(result.ToSkip) != 1 {
		t.Errorf("esperaba 1 omitido, obtuvo %d", len(result.ToSkip))
	}
	if len(result.ToCopy) != 1 {
		t.Errorf("esperaba 1 para copiar, obtuvo %d", len(result.ToCopy))
	}
}

func TestCompareDirectories_SavedSizeCalculation(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	content := "a]content that is identical and has some length"
	createTempFile(t, srcDir, "a.txt", content, 0644)
	createTempFile(t, dstDir, "a.txt", content, 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if result.SavedSize != int64(len(content)) {
		t.Errorf("esperaba SavedSize=%d, obtuvo %d", len(content), result.SavedSize)
	}
}

func TestCompareDirectories_EmptySource(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, dstDir, "orphan.txt", "exists in dst only", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)

	if len(result.Decisions) != 0 {
		t.Errorf("esperaba 0 decisiones con origen vacío, obtuvo %d", len(result.Decisions))
	}
}

// --- Tests de CopyFiles ---

func TestCopyFiles_BasicCopy(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "file.txt", "hello world", 0644)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)
	stats := CopyFiles(result.ToCopy, dstDir, 2)

	if stats.CopiedFiles != 1 {
		t.Errorf("esperaba 1 archivo copiado, obtuvo %d", stats.CopiedFiles)
	}
	if stats.FailedFiles != 0 {
		t.Errorf("esperaba 0 fallos, obtuvo %d", stats.FailedFiles)
	}

	// Verificar que el archivo existe en destino con el contenido correcto.
	dstContent, err := os.ReadFile(filepath.Join(dstDir, "file.txt"))
	if err != nil {
		t.Fatalf("error leyendo archivo copiado: %v", err)
	}
	if string(dstContent) != "hello world" {
		t.Errorf("contenido incorrecto: %q", string(dstContent))
	}
}

func TestCopyFiles_PreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permisos Unix no aplican en Windows")
	}
	srcDir, dstDir, _ := setupTestDirs(t)

	createTempFile(t, srcDir, "exec.sh", "#!/bin/bash", 0755)

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)
	CopyFiles(result.ToCopy, dstDir, 1)

	dstInfo, err := os.Stat(filepath.Join(dstDir, "exec.sh"))
	if err != nil {
		t.Fatalf("error stat archivo copiado: %v", err)
	}
	if dstInfo.Mode().Perm() != 0755 {
		t.Errorf("permisos no preservados: esperaba 0755, obtuvo %o", dstInfo.Mode().Perm())
	}
}

func TestCopyFiles_EmptyDecisionList(t *testing.T) {
	dstDir := t.TempDir()
	stats := CopyFiles([]FileDecision{}, dstDir, 4)

	if stats.CopiedFiles != 0 {
		t.Errorf("esperaba 0 copiados, obtuvo %d", stats.CopiedFiles)
	}
}

func TestCopyFiles_Concurrency(t *testing.T) {
	srcDir, dstDir, _ := setupTestDirs(t)

	// Crear múltiples archivos.
	for i := 0; i < 10; i++ {
		createTempFile(t, srcDir, fmt.Sprintf("file_%02d.txt", i), fmt.Sprintf("content %d", i), 0644)
	}

	srcResult, _ := ScanDirectory(srcDir)
	dstResult, _ := ScanDirectory(dstDir)

	result := CompareDirectories(srcResult, dstResult)
	stats := CopyFiles(result.ToCopy, dstDir, 4)

	if stats.CopiedFiles != 10 {
		t.Errorf("esperaba 10 archivos copiados, obtuvo %d", stats.CopiedFiles)
	}

	// Verificar que todos los archivos existen.
	for i := 0; i < 10; i++ {
		path := filepath.Join(dstDir, fmt.Sprintf("file_%02d.txt", i))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("archivo %q no fue copiado", path)
		}
	}
}

// --- Tests de Action String ---

func TestActionString(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionCopy, "COPY"},
		{ActionSkip, "SKIP"},
		{ActionError, "ERROR"},
		{Action(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.action.String(); got != tt.expected {
			t.Errorf("Action(%d).String() = %q, esperaba %q", tt.action, got, tt.expected)
		}
	}
}

// --- Tests de formatBytes ---

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 bytes"},
		{512, "512 bytes"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
		{5368709120, "5.00 GB"},
	}

	for _, tt := range tests {
		got := formatBytes(tt.bytes)
		if got != tt.expected {
			t.Errorf("formatBytes(%d) = %q, esperaba %q", tt.bytes, got, tt.expected)
		}
	}
}