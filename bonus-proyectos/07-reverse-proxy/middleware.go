package main

import (
	"log"
	"net/http"
	"time"
)

// responseRecorder wraps http.ResponseWriter to capture the status code
// written by the downstream handler for logging purposes.
type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	bytes       int
	wroteHeader bool
}

// newResponseRecorder creates a ResponseRecorder that defaults to 200 OK.
func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code before delegating to the underlying writer.
func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.wroteHeader {
		rr.statusCode = code
		rr.wroteHeader = true
	}
	rr.ResponseWriter.WriteHeader(code)
}

// Write delegates to the underlying writer and tracks bytes written.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.wroteHeader {
		rr.wroteHeader = true
	}
	n, err := rr.ResponseWriter.Write(b)
	rr.bytes += n
	return n, err
}

// Flush supports http.Flusher if the underlying writer implements it.
func (rr *responseRecorder) Flush() {
	if f, ok := rr.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// LoggingMiddleware wraps an http.Handler and logs each request with
// method, path, status code, byte count, client IP, and duration.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := newResponseRecorder(w)

		// Call the next handler in the chain.
		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		log.Printf(
			"[access] %s %s %s -> %d (%d bytes, %s)",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			rec.statusCode,
			rec.bytes,
			duration.Round(time.Microsecond),
		)
	})
}

// RecoveryMiddleware wraps an http.Handler and recovers from panics,
// returning a 500 Internal Server Error to the client instead of crashing.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[recovery] panic recovered: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Chain composes multiple middleware functions around a handler.
// Middleware is applied in the order provided (first argument = outermost).
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}