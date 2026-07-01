package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler serves static files from the configured root directory.
// It implements http.Handler and uses os.Open + io.Copy for efficient streaming
// without loading entire files into memory.
type FileHandler struct {
	Root string
}

// NewFileHandler creates a new FileHandler rooted at the given directory.
func NewFileHandler(root string) *FileHandler {
	return &FileHandler{Root: root}
}

// ServeHTTP handles incoming HTTP requests by resolving the requested path
// against the root directory, validating security constraints, detecting
// the MIME type, and streaming the file contents to the response.
func (fh *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only allow GET and HEAD methods
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Sanitize the URL path
	cleanPath := SanitizePath(r.URL.Path)

	// Resolve the safe path (anti-path traversal)
	targetPath, err := SafePath(fh.Root, cleanPath)
	if err != nil {
		switch err {
		case ErrPathTraversal:
			log.Printf("SECURITY: Path traversal blocked from %s → %s", r.RemoteAddr, r.URL.Path)
			http.Error(w, "Forbidden", http.StatusForbidden)
		case ErrHiddenFile:
			log.Printf("SECURITY: Hidden file access blocked from %s → %s", r.RemoteAddr, r.URL.Path)
			http.Error(w, "Forbidden", http.StatusForbidden)
		default:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}
		return
	}

	// Open the file
	file, err := os.Open(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		if os.IsPermission(err) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("ERROR: opening file %s: %v", targetPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info for size and directory check
	info, err := file.Stat()
	if err != nil {
		log.Printf("ERROR: stating file %s: %v", targetPath, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If it's a directory, look for an index.html
	if info.IsDir() {
		indexPath := filepath.Join(targetPath, "index.html")
		indexFile, err := os.Open(indexPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		indexInfo, err := indexFile.Stat()
		if err != nil || indexInfo.IsDir() {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		// Serve the index.html file
		fh.serveFile(w, r, indexFile, indexInfo, indexPath)
		return
	}

	// Serve the regular file
	fh.serveFile(w, r, file, info, targetPath)
}

// serveFile streams the file contents to the HTTP response using io.Copy.
// It sets appropriate headers including Content-Type (via mime package),
// Content-Length, and Cache-Control.
func (fh *FileHandler) serveFile(w http.ResponseWriter, r *http.Request, file *os.File, info os.FileInfo, path string) {
	// Detect Content-Type based on file extension using the mime package
	contentType := detectContentType(path)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	w.Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))

	// Cache control: allow caching of static assets for 1 hour
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// For HEAD requests, don't send body
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Stream the file directly to the response buffer using io.Copy
	// This avoids loading the entire file into memory — O(1) memory usage
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, file); err != nil {
		// Client may have disconnected; log but don't try to write error headers
		log.Printf("WARN: streaming interrupted for %s: %v", path, err)
	}
}

// detectContentType determines the MIME type of a file based on its extension.
// It uses the standard mime package (mime.TypeByExtension) for reliable mapping.
// Falls back to "application/octet-stream" for unknown extensions.
func detectContentType(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return ""
	}

	// Use Go's standard mime package for type detection
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		// For text types, ensure charset=utf-8 is included
		if strings.HasPrefix(mimeType, "text/") && !strings.Contains(mimeType, "charset") {
			mimeType += "; charset=utf-8"
		}
		return mimeType
	}

	return ""
}
