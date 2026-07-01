package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────────
// Security Tests (security.go)
// ──────────────────────────────────────────────────────────────────────────────

func TestSafePath_NormalPaths(t *testing.T) {
	root := t.TempDir()

	tests := []struct {
		name    string
		urlPath string
		wantErr bool
	}{
		{"root path", "/", false},
		{"simple file", "/index.html", false},
		{"nested file", "/css/style.css", false},
		{"deep nested", "/js/vendor/lib.js", false},
		{"file with spaces", "/my file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SafePath(root, tt.urlPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafePath(%q, %q) error = %v, wantErr %v", root, tt.urlPath, err, tt.wantErr)
			}
		})
	}
}

func TestSafePath_PathTraversal(t *testing.T) {
	root := t.TempDir()

	attacks := []struct {
		name    string
		urlPath string
	}{
		{"simple dot-dot", "/../../../etc/passwd"},
		{"encoded dot-dot", "/..%2F..%2F..%2Fetc/passwd"},
		{"dot-dot in middle", "/public/../../../etc/passwd"},
		{"double dot-dot", "/../../../../../../etc/shadow"},
		{"dot-dot with valid dir", "/images/../../../etc/passwd"},
		{"just dot-dot", "/.."},
		{"dot-dot slash", "/../"},
	}

	for _, atk := range attacks {
		t.Run(atk.name, func(t *testing.T) {
			_, err := SafePath(root, atk.urlPath)
			if err == nil {
				t.Errorf("SafePath should have blocked traversal attack %q, but got no error", atk.urlPath)
			}
			if err != ErrPathTraversal {
				t.Errorf("SafePath(%q) error = %v, want ErrPathTraversal", atk.urlPath, err)
			}
		})
	}
}

func TestSafePath_HiddenFiles(t *testing.T) {
	root := t.TempDir()

	attacks := []struct {
		name    string
		urlPath string
	}{
		{"hidden file", "/.env"},
		{"hidden directory", "/.git/config"},
		{"hidden in nested", "/public/.secret/key.pem"},
	}

	for _, atk := range attacks {
		t.Run(atk.name, func(t *testing.T) {
			_, err := SafePath(root, atk.urlPath)
			if err == nil {
				t.Errorf("SafePath should have blocked hidden file access %q, but got no error", atk.urlPath)
			}
			if err != ErrHiddenFile {
				t.Errorf("SafePath(%q) error = %v, want ErrHiddenFile", atk.urlPath, err)
			}
		})
	}
}

func TestSafePath_Sanitization(t *testing.T) {
	root := t.TempDir()

	// Null byte injection
	_, err := SafePath(root, "/file\x00.html")
	if err != nil {
		// After sanitization, this should be treated as "/file.html"
		// It's acceptable if it returns an error (the file won't exist)
		// but it should NOT cause a path traversal
		if err == ErrPathTraversal {
			t.Error("Null byte injection should not trigger path traversal error")
		}
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal path", "/css/style.css", "/css/style.css"},
		{"double slashes", "//css//style.css", "/css/style.css"},
		{"backslashes", "\\css\\style.css", "/css/style.css"},
		{"null byte", "/file\x00.html", "/file.html"},
		{"whitespace", "  /css/style.css  ", "/css/style.css"},
		{"trailing slashes", "/css/style.css//", "/css/style.css/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePath(tt.input)
			if got != tt.want {
				t.Errorf("SanitizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Handler Tests (handler.go)
// ──────────────────────────────────────────────────────────────────────────────

// setupTestRoot creates a temporary directory with test files and returns its path.
func setupTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// Create test files
	files := map[string]string{
		"index.html":              "<html><body>Hello World</body></html>",
		"css/style.css":           "body { margin: 0; }",
		"js/app.js":               "console.log('hello');",
		"data.json":               `{"key": "value"}`,
		"image.png":               "\x89PNG\r\n\x1a\n", // PNG magic bytes
		"text/readme.txt":         "This is a readme file.",
		"deep/nested/file.html":   "<html>Deep</html>",
		"subdir/index.html":       "<html>Subdir Index</html>",
		"image.jpg":               "\xff\xd8\xff\xe0", // JPEG magic bytes
		"document.pdf":            "%PDF-1.4",          // PDF magic bytes
		"style.min.css":           "body{margin:0}",
		"script.min.js":           "var a=1;",
		"robots.txt":              "User-agent: *\nDisallow:",
	}

	for path, content := range files {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %q: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %q: %v", fullPath, err)
		}
	}

	return root
}

func TestFileHandler_ServeFile(t *testing.T) {
	root := setupTestRoot(t)
	handler := NewFileHandler(root)

	tests := []struct {
		name           string
		path           string
		wantStatus     int
		wantContains   string
		wantType       string
		method         string
	}{
		{
			name:         "serve index.html at root",
			path:         "/",
			wantStatus:   http.StatusOK,
			wantContains: "Hello World",
			wantType:     "text/html",
			method:       http.MethodGet,
		},
		{
			name:         "serve css file",
			path:         "/css/style.css",
			wantStatus:   http.StatusOK,
			wantContains: "margin: 0",
			wantType:     "text/css",
			method:       http.MethodGet,
		},
	{
			name:         "serve js file",
			path:         "/js/app.js",
			wantStatus:   http.StatusOK,
			wantContains: "console.log",
			wantType:     "javascript",
			method:       http.MethodGet,
		},
		{
			name:         "serve json file",
			path:         "/data.json",
			wantStatus:   http.StatusOK,
			wantContains: `"key"`,
			wantType:     "application/json",
			method:       http.MethodGet,
		},
		{
			name:         "serve txt file",
			path:         "/text/readme.txt",
			wantStatus:   http.StatusOK,
			wantContains: "readme",
			wantType:     "text/plain",
			method:       http.MethodGet,
		},
		{
			name:         "serve nested html",
			path:         "/deep/nested/file.html",
			wantStatus:   http.StatusOK,
			wantContains: "Deep",
			wantType:     "text/html",
			method:       http.MethodGet,
		},
		{
			name:       "404 for non-existent file",
			path:       "/does-not-exist.html",
			wantStatus: http.StatusNotFound,
			method:     http.MethodGet,
		},
		{
			name:       "403 for path traversal",
			path:       "/../../../etc/passwd",
			wantStatus: http.StatusForbidden,
			method:     http.MethodGet,
		},
		{
			name:       "405 for POST method",
			path:       "/index.html",
			wantStatus: http.StatusMethodNotAllowed,
			method:     http.MethodPost,
		},
		{
			name:       "HEAD request returns headers but no body",
			path:       "/index.html",
			wantStatus: http.StatusOK,
			method:     http.MethodHead,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantContains != "" && !strings.Contains(rec.Body.String(), tt.wantContains) {
				t.Errorf("body does not contain %q, got %q", tt.wantContains, rec.Body.String())
			}

			if tt.wantType != "" {
				ct := rec.Header().Get("Content-Type")
				if !strings.Contains(ct, tt.wantType) {
					t.Errorf("Content-Type = %q, want contains %q", ct, tt.wantType)
				}
			}

			// HEAD should return no body
			if tt.method == http.MethodHead && rec.Body.Len() > 0 {
				t.Errorf("HEAD request should have empty body, got %d bytes", rec.Body.Len())
			}
		})
	}
}

func TestFileHandler_ContentTypeDetection(t *testing.T) {
	root := setupTestRoot(t)
	handler := NewFileHandler(root)

	tests := []struct {
		path         string
		wantType     string
	}{
		{"/css/style.css", "text/css"},
	{"/js/app.js", "javascript"},
		{"/data.json", "application/json"},
		{"/index.html", "text/html"},
		{"/image.png", "image/png"},
		{"/image.jpg", "image/jpeg"},
		{"/document.pdf", "application/pdf"},
		{"/text/readme.txt", "text/plain"},
		{"/robots.txt", "text/plain"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodHead, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, tt.wantType) {
				t.Errorf("%s: Content-Type = %q, want contains %q", tt.path, ct, tt.wantType)
			}
		})
	}
}

func TestFileHandler_DirectoryIndex(t *testing.T) {
	root := setupTestRoot(t)
	handler := NewFileHandler(root)

	// Root directory should serve index.html
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("root directory: status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Hello World") {
		t.Error("root directory should serve index.html")
	}

	// Subdirectory with index.html should serve it
	req = httptest.NewRequest(http.MethodGet, "/subdir/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("subdir: status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Subdir Index") {
		t.Error("subdir should serve its index.html")
	}

	// Directory without index.html should 404
	req = httptest.NewRequest(http.MethodGet, "/css/", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("dir without index: status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestFileHandler_HiddenFileBlocking(t *testing.T) {
	root := setupTestRoot(t)

	// Create a hidden file
	hiddenPath := filepath.Join(root, ".env")
	if err := os.WriteFile(hiddenPath, []byte("SECRET=abc123"), 0644); err != nil {
		t.Fatalf("failed to create hidden file: %v", err)
	}

	handler := NewFileHandler(root)

	req := httptest.NewRequest(http.MethodGet, "/.env", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("hidden file: status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if strings.Contains(rec.Body.String(), "SECRET") {
		t.Error("hidden file content should not be exposed")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Middleware Tests (middleware.go)
// ──────────────────────────────────────────────────────────────────────────────

func TestGzipMiddleware_CompressesResponse(t *testing.T) {
	// Create a simple handler that writes a known body
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Hello World</body></html>"))
	})

	handler := GzipMiddleware(inner)

	// Request WITH Accept-Encoding: gzip
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Verify Content-Encoding header
	ce := rec.Header().Get("Content-Encoding")
	if ce != "gzip" {
		t.Errorf("Content-Encoding = %q, want %q", ce, "gzip")
	}

	// Verify Vary header
	vary := rec.Header().Get("Vary")
	if vary != "Accept-Encoding" {
		t.Errorf("Vary = %q, want %q", vary, "Accept-Encoding")
	}

	// Verify the body is actually gzip-compressed
	body := rec.Body.Bytes()
	if len(body) == 0 {
		t.Fatal("response body is empty")
	}

	// Decompress and verify content
	reader, err := gzip.NewReader(strings.NewReader(string(body)))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if !strings.Contains(string(decompressed), "Hello World") {
		t.Errorf("decompressed body does not contain expected content, got %q", string(decompressed))
	}
}

func TestGzipMiddleware_NoCompressionWithoutAcceptEncoding(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello World"))
	})

	handler := GzipMiddleware(inner)

	// Request WITHOUT Accept-Encoding
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Should NOT have Content-Encoding: gzip
	ce := rec.Header().Get("Content-Encoding")
	if ce == "gzip" {
		t.Error("should not compress when client does not accept gzip")
	}

	// Body should be plain text
	if rec.Body.String() != "Hello World" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "Hello World")
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := SecurityHeadersMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"X-XSS-Protection":       "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
	}

	for header, expected := range expectedHeaders {
		got := rec.Header().Get(header)
		if got != expected {
			t.Errorf("%s = %q, want %q", header, got, expected)
		}
	}
}

func TestLoggingMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := LoggingMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Verify the response still works (logging doesn't break anything)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "OK" {
		t.Errorf("body = %q, want %q", rec.Body.String(), "OK")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Integration Tests (full middleware chain)
// ──────────────────────────────────────────────────────────────────────────────

func TestIntegration_FullMiddlewareChain(t *testing.T) {
	root := setupTestRoot(t)
	fileHandler := NewFileHandler(root)

	// Build the same middleware chain as the real server
	var handler http.Handler = fileHandler
	handler = GzipMiddleware(handler)
	handler = SecurityHeadersMiddleware(handler)
	handler = LoggingMiddleware(handler)

	t.Run("serves file with all middleware", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/index.html", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		// Should have security headers
		if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Error("missing security header X-Content-Type-Options")
		}

		// Should have gzip compression
		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Error("missing gzip Content-Encoding header")
		}

		// Should have correct Content-Type
		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("Content-Type = %q, want contains text/html", ct)
		}
	})

	t.Run("blocks path traversal through full chain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/../../../etc/passwd", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
		if strings.Contains(rec.Body.String(), "root:") {
			t.Error("path traversal should not expose /etc/passwd content")
		}
	})

	t.Run("serves CSS with correct mime type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/css/style.css", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/css") {
			t.Errorf("Content-Type = %q, want text/css", ct)
		}
	})

	t.Run("returns 404 for missing files", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nonexistent.txt", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("HEAD request returns headers only", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodHead, "/index.html", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if rec.Body.Len() > 0 {
			t.Errorf("HEAD response should have empty body, got %d bytes", rec.Body.Len())
		}
		// Should still have Content-Length
		if rec.Header().Get("Content-Length") == "" {
			t.Error("HEAD response should include Content-Length")
		}
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Server Tests (server.go)
// ──────────────────────────────────────────────────────────────────────────────

func TestNewServer_InvalidRoot(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Root = "/nonexistent/directory/that/does/not/exist"

	_, err := NewServer(cfg)
	if err == nil {
		t.Error("NewServer should fail with non-existent root directory")
	}
}

func TestNewServer_FileAsRoot(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cfg := DefaultConfig()
	cfg.Root = tmpFile.Name()

	_, err = NewServer(cfg)
	if err == nil {
		t.Error("NewServer should fail when root is a file, not a directory")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "0.0.0.0" {
		t.Errorf("DefaultConfig().Host = %q, want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != 8080 {
		t.Errorf("DefaultConfig().Port = %d, want %d", cfg.Port, 8080)
	}
	if cfg.Root != "./public" {
		t.Errorf("DefaultConfig().Root = %q, want %q", cfg.Root, "./public")
	}
	if cfg.ReadTimeout == 0 {
		t.Error("DefaultConfig().ReadTimeout should not be zero")
	}
	if cfg.WriteTimeout == 0 {
		t.Error("DefaultConfig().WriteTimeout should not be zero")
	}
	if cfg.IdleTimeout == 0 {
		t.Error("DefaultConfig().IdleTimeout should not be zero")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1073741824, "1.0GB"},
		{5368709120, "5.0GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}