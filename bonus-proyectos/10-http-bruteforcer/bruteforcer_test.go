package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ─────────────────────────────────────────────────────────────
// Config Tests
// ─────────────────────────────────────────────────────────────

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Workers != 10 {
		t.Errorf("expected Workers=10, got %d", cfg.Workers)
	}
	if cfg.Delay != 500 {
		t.Errorf("expected Delay=500, got %d", cfg.Delay)
	}
	if cfg.Timeout != 10 {
		t.Errorf("expected Timeout=10, got %d", cfg.Timeout)
	}
	if cfg.UserAgent == "" {
		t.Error("expected non-empty UserAgent")
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "empty target URL",
			cfg:     Config{TargetURL: ""},
			wantErr: true,
		},
		{
			name:    "invalid target type",
			cfg:     Config{TargetURL: "http://x", TargetType: "invalid"},
			wantErr: true,
		},
		{
			name: "form type without fields",
			cfg: Config{
				TargetURL:     "http://x",
				TargetType:    "form",
				UsernameField: "",
				PasswordField: "",
			},
			wantErr: true,
		},
		{
			name: "basic auth valid",
			cfg: Config{
				TargetURL:        "http://x",
				TargetType:       "basic",
				Workers:          10,
				BurstSize:        20,
				Timeout:          10,
				UsersFile:        "users.txt",
				PasswordsFile:    "passwords.txt",
				SuccessIndicator: "Welcome",
			},
			wantErr: false,
		},
		{
			name: "form valid",
			cfg: Config{
				TargetURL:        "http://x",
				TargetType:       "form",
				Workers:          10,
				BurstSize:        20,
				Timeout:          10,
				UsernameField:    "user",
				PasswordField:    "pass",
				UsersFile:        "users.txt",
				PasswordsFile:    "passwords.txt",
				FailureIndicator: "error",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigSaveLoad(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test_config.json")

	cfg := DefaultConfig()
	cfg.TargetURL = "http://test.local/login"
	cfg.TargetType = "form"
	cfg.UsernameField = "email"
	cfg.PasswordField = "pwd"

	if err := SaveConfig(cfg, path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.TargetURL != cfg.TargetURL {
		t.Errorf("TargetURL mismatch: %s vs %s", loaded.TargetURL, cfg.TargetURL)
	}
	if loaded.UsernameField != cfg.UsernameField {
		t.Errorf("UsernameField mismatch: %s vs %s", loaded.UsernameField, cfg.UsernameField)
	}
	if loaded.Workers != cfg.Workers {
		t.Errorf("Workers mismatch: %d vs %d", loaded.Workers, cfg.Workers)
	}
}

// ─────────────────────────────────────────────────────────────
// Wordlist Tests
// ─────────────────────────────────────────────────────────────

func TestLoadWordlist(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "test.txt")
	content := "admin\nroot\n# comment line\n\nuser1\n  user2  \n"
	os.WriteFile(path, []byte(content), 0644)

	wl, err := LoadWordlist("test", path)
	if err != nil {
		t.Fatalf("LoadWordlist failed: %v", err)
	}

	if wl.Len() != 4 {
		t.Errorf("expected 4 entries, got %d", wl.Len())
	}
	if wl.Entries[0] != "admin" {
		t.Errorf("expected 'admin', got '%s'", wl.Entries[0])
	}
	if wl.Entries[3] != "user2" {
		t.Errorf("expected 'user2' (trimmed), got '%s'", wl.Entries[3])
	}
}

func TestLoadWordlistEmpty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.txt")
	os.WriteFile(path, []byte("# only comments\n\n"), 0644)

	_, err := LoadWordlist("empty", path)
	if err == nil {
		t.Error("expected error for empty wordlist")
	}
}

func TestGeneratePairs(t *testing.T) {
	users := &Wordlist{Name: "u", Entries: []string{"a", "b"}}
	passwords := &Wordlist{Name: "p", Entries: []string{"1", "2", "3"}}

	pairs := GeneratePairs(users, passwords)
	var collected []EntryPair
	for p := range pairs {
		collected = append(collected, p)
	}

	if len(collected) != 6 {
		t.Errorf("expected 6 pairs, got %d", len(collected))
	}
}

// ─────────────────────────────────────────────────────────────
// Rate Limiter Tests
// ─────────────────────────────────────────────────────────────

func TestRateLimiterBasic(t *testing.T) {
	rl := NewRateLimiter(10*time.Millisecond, 5, 0, 0)

	start := time.Now()
	for i := 0; i < 5; i++ {
		rl.Wait()
	}
	elapsed := time.Since(start)

	// First 5 should be fast (burst).
	if elapsed > 200*time.Millisecond {
		t.Errorf("burst requests took too long: %v", elapsed)
	}
}

func TestRateLimiterBurstExhaustion(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 2, 100*time.Millisecond, 0)

	// Exhaust burst.
	rl.Wait()
	rl.Wait()

	// Third should be delayed.
	start := time.Now()
	rl.Wait()
	elapsed := time.Since(start)

	if elapsed < 50*time.Millisecond {
		t.Errorf("expected delay after burst, got %v", elapsed)
	}
}

// ─────────────────────────────────────────────────────────────
// Detector Tests
// ─────────────────────────────────────────────────────────────

func TestDetectorSuccess(t *testing.T) {
	cfg := &Config{SuccessIndicator: "Welcome"}
	d := NewDetector(cfg)

	resp := &LoginResponse{StatusCode: 200, Body: "<h1>Welcome back!</h1>"}
	result := d.DetectResult(resp, "admin", "pass")
	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}
}

func TestDetectorFailure(t *testing.T) {
	cfg := &Config{FailureIndicator: "Invalid credentials"}
	d := NewDetector(cfg)

	resp := &LoginResponse{StatusCode: 200, Body: "Invalid credentials. Try again."}
	result := d.DetectResult(resp, "admin", "wrong")
	if result != "failure" {
		t.Errorf("expected 'failure', got '%s'", result)
	}
}

func TestDetectorBlocked(t *testing.T) {
	cfg := &Config{}
	d := NewDetector(cfg)

	resp := &LoginResponse{StatusCode: 429}
	result := d.DetectResult(resp, "admin", "pass")
	if result != "blocked" {
		t.Errorf("expected 'blocked', got '%s'", result)
	}
}

func TestDetectorError(t *testing.T) {
	cfg := &Config{}
	d := NewDetector(cfg)

	resp := &LoginResponse{Error: http.ErrHandlerTimeout}
	result := d.DetectResult(resp, "admin", "pass")
	if result != "error" {
		t.Errorf("expected 'error', got '%s'", result)
	}
}

func TestDetectorRedirect(t *testing.T) {
	cfg := &Config{SuccessStatus: 0}
	d := NewDetector(cfg)

	// Redirect to dashboard = success.
	resp := &LoginResponse{StatusCode: 302, Location: "/dashboard"}
	result := d.DetectResult(resp, "admin", "pass")
	if result != "success" {
		t.Errorf("expected 'success' on redirect to dashboard, got '%s'", result)
	}

	// Redirect back to login = failure.
	resp = &LoginResponse{StatusCode: 302, Location: "/login?error=1"}
	result = d.DetectResult(resp, "admin", "pass")
	if result != "failure" {
		t.Errorf("expected 'failure' on redirect to login, got '%s'", result)
	}
}

// ─────────────────────────────────────────────────────────────
// Utility Tests
// ─────────────────────────────────────────────────────────────

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"longpasswordhere", 10, "longpas..."},
		{"abc", 3, "abc"},
		{"ab", 1, "a"},
	}
	for _, tt := range tests {
		got := truncateString(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestApplyJitter(t *testing.T) {
	d := 1000 * time.Millisecond
	// With 0.3 factor, result should be in [700ms, 1300ms].
	for i := 0; i < 100; i++ {
		result := applyJitter(d, 0.3)
		if result < 700*time.Millisecond || result > 1300*time.Millisecond {
			t.Errorf("applyJitter(1s, 0.3) = %v, expected [700ms, 1300ms]", result)
		}
	}
}

func TestContains(t *testing.T) {
	if !contains("hello world", "world") {
		t.Error("expected contains to find 'world'")
	}
	if contains("hello", "world") {
		t.Error("expected contains not to find 'world'")
	}
	if contains("", "a") {
		t.Error("expected false for empty string")
	}
}

// ─────────────────────────────────────────────────────────────
// Form Detection Tests
// ─────────────────────────────────────────────────────────────

func TestDetectLoginForm(t *testing.T) {
	html := `<html><body>
		<form action="/login" method="post">
			<input type="text" name="username" />
			<input type="password" name="password" />
			<button type="submit">Login</button>
		</form>
	</body></html>`

	result := DetectLoginForm(html, "http://test.local")
	if !result.Found {
		t.Fatal("expected form to be found")
	}
	if result.UsernameField != "username" {
		t.Errorf("expected username field 'username', got '%s'", result.UsernameField)
	}
	if result.PasswordField != "password" {
		t.Errorf("expected password field 'password', got '%s'", result.PasswordField)
	}
	if result.FormAction != "http://test.local/login" {
		t.Errorf("expected form action 'http://test.local/login', got '%s'", result.FormAction)
	}
}

func TestDetectLoginFormNotFound(t *testing.T) {
	html := `<html><body><p>No form here</p></body></html>`
	result := DetectLoginForm(html, "http://test.local")
	if result.Found {
		t.Error("expected form not to be found")
	}
}

// ─────────────────────────────────────────────────────────────
// Integration Test — HTTP Server
// ─────────────────────────────────────────────────────────────

func TestBruteforcerIntegration(t *testing.T) {
	// Create a mock login server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		user := r.FormValue("username")
		pass := r.FormValue("password")

		if user == "admin" && pass == "secret123" {
			w.Header().Set("Location", "/dashboard")
			w.WriteHeader(302)
			w.Write([]byte(`<html>Redirecting to dashboard</html>`))
		} else {
			w.WriteHeader(200)
			w.Write([]byte(`<html>Invalid credentials. <a href="/login">Try again</a></html>`))
		}
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.TargetURL = server.URL
	cfg.TargetType = "form"
	cfg.UsernameField = "username"
	cfg.PasswordField = "password"
	cfg.Workers = 2
	cfg.Delay = 10
	cfg.BurstSize = 10
	cfg.Timeout = 5
	cfg.Verbose = false

	bf, err := NewBruteforcer(cfg)
	if err != nil {
		t.Fatalf("NewBruteforcer failed: %v", err)
	}
	defer bf.Close()

	users := &Wordlist{Name: "users", Entries: []string{"root", "admin", "user"}}
	passwords := &Wordlist{Name: "passwords", Entries: []string{"1234", "secret123", "password"}}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pairs := GeneratePairs(users, passwords)
	found := bf.Run(ctx, pairs)

	if len(found) == 0 {
		t.Fatal("expected to find credentials")
	}

	if found[0].Username != "admin" || found[0].Password != "secret123" {
		t.Errorf("expected admin:secret123, got %s:%s", found[0].Username, found[0].Password)
	}
}

func TestBruteforcerBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok && user == "testuser" && pass == "testpass" {
			w.WriteHeader(200)
			w.Write([]byte("Welcome"))
		} else {
			w.Header().Set("WWW-Authenticate", `Basic realm="Test"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorized"))
		}
	}))
	defer server.Close()

	cfg := DefaultConfig()
	cfg.TargetURL = server.URL
	cfg.TargetType = "basic"
	cfg.Workers = 1
	cfg.Delay = 10
	cfg.BurstSize = 5
	cfg.SuccessStatus = 200

	bf, err := NewBruteforcer(cfg)
	if err != nil {
		t.Fatalf("NewBruteforcer failed: %v", err)
	}
	defer bf.Close()

	users := &Wordlist{Name: "u", Entries: []string{"testuser", "other"}}
	passwords := &Wordlist{Name: "p", Entries: []string{"wrong", "testpass"}}

	ctx := context.Background()
	pairs := GeneratePairs(users, passwords)
	found := bf.Run(ctx, pairs)

	if len(found) == 0 {
		t.Fatal("expected to find credentials with basic auth")
	}

	if found[0].Username != "testuser" || found[0].Password != "testpass" {
		t.Errorf("expected testuser:testpass, got %s:%s", found[0].Username, found[0].Password)
	}
}
