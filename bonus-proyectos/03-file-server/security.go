package main

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	// ErrPathTraversal is returned when a request attempts directory traversal.
	ErrPathTraversal = errors.New("path traversal attempt detected")

	// ErrInvalidPath is returned when the resolved path is invalid.
	ErrInvalidPath = errors.New("invalid file path")

	// ErrHiddenFile is returned when a request targets a hidden file or directory.
	ErrHiddenFile = errors.New("access to hidden files is forbidden")
)

// SafePath resolves and validates the requested URL path against the root directory.
// It prevents directory traversal attacks by ensuring the resolved path stays
// within the boundaries of the root directory.
//
// Security checks performed:
//   - filepath.Clean normalizes the path (removes "..", double slashes, etc.)
//   - filepath.Rel computes the relative path from root to the target
//   - Rejects any path that escapes the root (e.g., "../../etc/passwd")
//   - Rejects hidden files/directories (starting with ".")
func SafePath(root, urlPath string) (string, error) {
	// Normalize the URL path: remove leading slash and clean
	urlPath = strings.TrimPrefix(urlPath, "/")
	if urlPath == "" {
		urlPath = "."
	}

	// Clean both paths to normalize separators and remove ".." components
	cleanRoot := filepath.Clean(root)
	cleanURL := filepath.Clean(urlPath)

	// Reject any path containing ".." after cleaning (belt-and-suspenders)
	if strings.Contains(cleanURL, "..") {
		return "", ErrPathTraversal
	}

	// Build the full target path by joining root with the cleaned URL path
	targetPath := filepath.Join(cleanRoot, cleanURL)

	// Clean the resulting joined path again to handle any residual traversal
	targetPath = filepath.Clean(targetPath)

	// Use filepath.Rel to compute the relative path from root to target
	relPath, err := filepath.Rel(cleanRoot, targetPath)
	if err != nil {
		return "", ErrPathTraversal
	}

	// If the relative path starts with "..", it means the target is outside root
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}

	// Reject hidden files and directories (names starting with ".")
	// Split the relative path into components and check each one
	parts := strings.Split(relPath, string(filepath.Separator))
	for _, part := range parts {
		if part != "." && part != "" && strings.HasPrefix(part, ".") {
			return "", ErrHiddenFile
		}
	}

	return targetPath, nil
}

// SanitizePath performs additional sanitization on URL paths.
// It removes null bytes, normalizes slashes, and trims whitespace.
func SanitizePath(urlPath string) string {
	// Remove null byte injection attempts
	urlPath = strings.ReplaceAll(urlPath, "\x00", "")

	// Trim whitespace
	urlPath = strings.TrimSpace(urlPath)

	// Normalize backslashes to forward slashes (Windows compatibility)
	urlPath = strings.ReplaceAll(urlPath, "\\", "/")

	// Collapse multiple consecutive slashes into a single slash
	for strings.Contains(urlPath, "//") {
		urlPath = strings.ReplaceAll(urlPath, "//", "/")
	}

	return urlPath
}