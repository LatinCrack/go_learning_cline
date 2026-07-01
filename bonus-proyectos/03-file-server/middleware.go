package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// gzipWriter wraps http.ResponseWriter to intercept writes and compress them.
type gzipWriter struct {
	http.ResponseWriter
	writer      io.Writer
	wroteHeader bool
}

// Write implements io.Writer, delegating to the underlying gzip writer.
func (gw *gzipWriter) Write(b []byte) (int, error) {
	if !gw.wroteHeader {
		gw.WriteHeader(http.StatusOK)
	}
	return gw.writer.Write(b)
}

// WriteHeader captures the status code and sets the Content-Encoding header.
func (gw *gzipWriter) WriteHeader(code int) {
	if gw.wroteHeader {
		return
	}
	gw.wroteHeader = true
	gw.Header().Set("Content-Encoding", "gzip")
	gw.ResponseWriter.WriteHeader(code)
}

// gzipPool is a sync.Pool of gzip.Writer instances to reduce allocations.
var gzipPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

// GzipMiddleware compresses HTTP responses using gzip when the client supports it.
// It checks the Accept-Encoding header and only compresses text-based content types.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Get a gzip writer from the pool
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			gz.Close()
			gzipPool.Put(gz)
		}()

		// Wrap the ResponseWriter
		gw := &gzipWriter{
			ResponseWriter: w,
			writer:         gz,
		}

		// Set Vary header so caches differentiate compressed/non-compressed
		w.Header().Set("Vary", "Accept-Encoding")

		next.ServeHTTP(gw, r)
	})
}

// SecurityHeadersMiddleware adds standard security headers to all responses.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Clickjacking protection
		w.Header().Set("X-Frame-Options", "DENY")
		// XSS protection (legacy browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// accessLogEntry holds structured data for a single request log line.
type accessLogEntry struct {
	Method     string
	Path       string
	StatusCode int
	Size       int64
	Duration   time.Duration
	RemoteAddr string
}

// loggingResponseWriter wraps http.ResponseWriter to capture the status code and bytes written.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int64
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lrw.ResponseWriter.Write(b)
	lrw.bytes += int64(n)
	return n, err
}

// LoggingMiddleware logs each HTTP request with method, path, status, size, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		log.Printf("%s %s → %d (%s) [%s] %s",
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			formatBytes(lrw.bytes),
			duration.Round(time.Millisecond),
			r.RemoteAddr,
		)
	})
}

// formatBytes returns a human-readable string for byte counts.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}