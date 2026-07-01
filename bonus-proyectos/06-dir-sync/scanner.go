package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// FileInfo representa la metadata de un archivo escaneado en el árbol de directorios.
type FileInfo struct {
	Path string // Ruta absoluta del archivo
	Size int64  // Tamaño en bytes
	Mode os.FileMode
	ModTime int64 // Unix timestamp de última modificación
	RelPath string // Ruta relativa respecto al directorio raíz escaneado
}

// ScanResult contiene el resultado completo del escaneo de un directorio.
type ScanResult struct {
	RootPath string
	Files    []FileInfo
	TotalSize int64
	Errors    []ScanError
}

// ScanError registra errores encontrados durante el escaneo.
type ScanError struct {
	Path string
	Err  error
}

// ScanDirectory recorre recursivamente el directorio raíz usando filepath.WalkDir
// y retorna un ScanResult con todos los archivos encontrados.
func ScanDirectory(root string) (*ScanResult, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("error resolviendo ruta absoluta de %q: %w", root, err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("error accediendo al directorio %q: %w", absRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("la ruta %q no es un directorio", absRoot)
	}

	result := &ScanResult{
		RootPath: absRoot,
		Files:    make([]FileInfo, 0),
		Errors:   make([]ScanError, 0),
	}

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				Path: path,
				Err:  err,
			})
			// Continuar el walk a pesar del error para no perder otros archivos.
			return nil
		}

		// Ignorar directorios — solo procesar archivos regulares.
		if d.IsDir() {
			return nil
		}

		// Verificar que sea un archivo regular (no symlink, device, etc.)
		fileInfo, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				Path: path,
				Err:  fmt.Errorf("error obteniendo info: %w", err),
			})
			return nil
		}

		if !fileInfo.Mode().IsRegular() {
			return nil
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			result.Errors = append(result.Errors, ScanError{
				Path: path,
				Err:  fmt.Errorf("error calculando ruta relativa: %w", err),
			})
			return nil
		}
		// Normalizar separadores a '/' para consistencia multiplataforma.
		relPath = filepath.ToSlash(relPath)

		fi := FileInfo{
			Path:    path,
			Size:    fileInfo.Size(),
			Mode:    fileInfo.Mode(),
			ModTime: fileInfo.ModTime().Unix(),
			RelPath: relPath,
		}

		result.Files = append(result.Files, fi)
		result.TotalSize += fileInfo.Size()

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error durante el escaneo de %q: %w", absRoot, err)
	}

	return result, nil
}