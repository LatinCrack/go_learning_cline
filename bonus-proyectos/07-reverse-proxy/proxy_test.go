package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

// --- Balancer Tests ---

func TestBalancerRoundRobin(t *testing.T) {
	urls := []string{
		"http://backend1:8080",
		"http://backend2:8080",
		"http://backend3:8080",
	}
	b := NewBalancer(urls)

	// Collect the order of backends returned by consecutive Next() calls.
	var order []string
	for i := 0; i < 6; i++ {
		backend := b.Next()
		if backend == nil {
			t.Fatal("expected a backend, got nil")
		}
		order = append(order, backend.URL.Host)
	}

	// With round-robin, the sequence should cycle through all three backends.
	// counter.Add(1) - 1 starts at 0:
	// idx 0 -> backend1, idx 1 -> backend2, idx 2 -> backend3,
	// idx 3 -> backend1, idx 4 -> backend2, idx 5 -> backend3
	expected := []string{
		"backend1:8080",
		"backend2:8080",
		"backend3:8080",
		"backend1:8080",
		"backend2:8080",
		"backend3:8080",
	}

	for i, want := range expected {
		if order[i] != want {
			t.Errorf("round %d: got %s, want %s", i, order[i], want)
		}
	}
}

func TestBalancerSkipsUnhealthy(t *testing.T) {
	urls := []string{
		"http://backend1:8080",
		"http://backend2:8080",
		"http://backend3:8080",
	}
	b := NewBalancer(urls)

	// Mark backend2 as down.
	backends := b.Backends()
	backends[1].Alive.Store(false)

	// Next should skip backend2 and return either backend1 or backend3.
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		backend := b.Next()
		if backend == nil {
			t.Fatal("expected a backend, got nil")
		}
		seen[backend.URL.Host] = true
	}

	if seen["backend2:8080"] {
		t.Error("unhealthy backend2 should not be returned")
	}
	if !seen["backend1:8080"] || !seen["backend3:8080"] {
		t.Error("healthy backends should all be returned")
	}
}

func TestBalancerAllUnhealthy(t *testing.T) {
	urls := []string{"http://b1:8080", "http://b2:8080"}
	b := NewBalancer(urls)

	// Mark all as down.
	for _, backend := range b.Backends() {
		backend.Alive.Store(false)
	}

	if got := b.Next(); got != nil {
		t.Errorf("expected nil when all backends are unhealthy, got %v", got)
	}
}

func TestBalancerHealthyCount(t *testing.T) {
	urls := []string{"http://b1:8080", "http://b2:8080", "http://b3:8080"}
	b := NewBalancer(urls)

	if got := b.HealthyCount(); got != 3 {
		t.Errorf("expected 3 healthy, got %d", got)
	}

	b.Backends()[0].Alive.Store(false)
	if got := b.HealthyCount(); got != 2 {
		t.Errorf("expected 2 healthy, got %d", got)
	}
}

func TestBalancerMarkUpDown(t *testing.T) {
	urls := []string{"http://b1:8080"}
	b := NewBalancer(urls)

	u, _ := url.Parse("http://b1:8080")
	b.MarkDown(u)

	if b.HealthyCount() != 0 {
		t.Error("expected 0 healthy after MarkDown")
	}

	b.MarkUp(u)
	if b.HealthyCount() != 1 {
		t.Error("expected 1 healthy after MarkUp")
	}
}

func TestBalancerConcurrency(t *testing.T) {
	urls := []string{
		"http://b1:8080",
		"http://b2:8080",
		"http://b3:8080",
	}
	b := NewBalancer(urls)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			backend := b.Next()
			if backend == nil {
				t.Error("got nil backend in concurrent test")
			}
		}()
	}
	wg.Wait()
}

// --- Proxy Tests ---

func TestReverseProxyForwarding(t *testing.T) {
	// Start a mock backend that echoes headers back.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Backend", "hit")
		w.Header().Set("X-Forwarded-For-Received", r.Header.Get("X-Forwarded-For"))
		w.Header().Set("X-Real-IP-Received", r.Header.Get("X-Real-IP"))
		fmt.Fprintf(w, "backend-response")
	}))
	defer backend.Close()

	// Configure balancer with the mock backend.
	b := NewBalancer([]string{backend.URL})
	proxy := NewReverseProxy(b, 5*time.Second)

	// Create a test request.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	resp := rec.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if string(body) != "backend-response" {
		t.Errorf("expected 'backend-response', got %q", string(body))
	}
	// httputil.ReverseProxy appends the IP from r.RemoteAddr (port stripped) to X-Forwarded-For.
	if got := resp.Header.Get("X-Forwarded-For-Received"); got != "192.168.1.100" {
		t.Errorf("expected X-Forwarded-For=192.168.1.100, got %q", got)
	}
	if got := resp.Header.Get("X-Real-IP-Received"); got != "192.168.1.100" {
		t.Errorf("expected X-Real-IP=192.168.1.100, got %q", got)
	}
}

func TestReverseProxyNoHealthyBackends(t *testing.T) {
	// Create a balancer with all backends marked down.
	b := NewBalancer([]string{"http://unreachable:9999"})
	for _, backend := range b.Backends() {
		backend.Alive.Store(false)
	}

	proxy := NewReverseProxy(b, 2*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestReverseProxyHeaderInjection(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Got-XFF", r.Header.Get("X-Forwarded-For"))
		w.Header().Set("Got-XRI", r.Header.Get("X-Real-IP"))
		w.Header().Set("Got-XFH", r.Header.Get("X-Forwarded-Host"))
		w.Header().Set("Got-XFP", r.Header.Get("X-Forwarded-Proto"))
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	b := NewBalancer([]string{backend.URL})
	proxy := NewReverseProxy(b, 5*time.Second)

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	req.RemoteAddr = "10.0.0.1:5555"
	req.Host = "proxy.example.com"
	rec := httptest.NewRecorder()

	proxy.ServeHTTP(rec, req)

	resp := rec.Result()

	// httputil.ReverseProxy appends the IP from r.RemoteAddr (port stripped) to X-Forwarded-For.
	if got := resp.Header.Get("Got-XFF"); got != "10.0.0.1" {
		t.Errorf("X-Forwarded-For: got %q, want 10.0.0.1", got)
	}
	if got := resp.Header.Get("Got-XRI"); got != "10.0.0.1" {
		t.Errorf("X-Real-IP: got %q, want 10.0.0.1", got)
	}
	if got := resp.Header.Get("Got-XFH"); got != "proxy.example.com" {
		t.Errorf("X-Forwarded-Host: got %q, want proxy.example.com", got)
	}
	if got := resp.Header.Get("Got-XFP"); got != "http" {
		t.Errorf("X-Forwarded-Proto: got %q, want http", got)
	}
}

// --- Middleware Tests ---

func TestLoggingMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "hello")
	})

	handler := LoggingMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("expected 'hello', got %q", rec.Body.String())
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := RecoveryMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestChain(t *testing.T) {
	var order []string

	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1-after")
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2-after")
		})
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
	})

	handler := Chain(inner, mw1, mw2)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expected := []string{"mw1-before", "mw2-before", "handler", "mw2-after", "mw1-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, want := range expected {
		if order[i] != want {
			t.Errorf("position %d: got %q, want %q", i, order[i], want)
		}
	}
}

// --- Integration Test ---

func TestIntegrationRoundRobinProxy(t *testing.T) {
	// Start 3 mock backends that identify themselves.
	var mu sync.Mutex
	hits := make(map[string]int)

	var servers []*httptest.Server
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("server-%d", i)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			hits[id]++
			mu.Unlock()
			fmt.Fprintf(w, id)
		}))
		servers = append(servers, srv)
	}
	defer func() {
		for _, s := range servers {
			s.Close()
		}
	}()

	// Configure balancer with the 3 backends.
	var urls []string
	for _, s := range servers {
		urls = append(urls, s.URL)
	}
	b := NewBalancer(urls)
	handler := LoggingMiddleware(NewReverseProxy(b, 5*time.Second))

	// Send 9 requests — each backend should receive exactly 3.
	for i := 0; i < 9; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i, rec.Code)
		}
	}

	mu.Lock()
	defer mu.Unlock()

	// With round-robin across 3 backends and 9 requests, each should get 3.
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("server-%d", i)
		if hits[id] != 3 {
			t.Errorf("backend %s received %d hits, expected 3", id, hits[id])
		}
	}
}

// --- Utility Tests ---

func TestClientIPFromRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:12345"

	got := clientIP(req)
	if got != "10.0.0.5" {
		t.Errorf("got %q, want 10.0.0.5", got)
	}
}

func TestClientIPFromXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")

	got := clientIP(req)
	if got != "203.0.113.50" {
		t.Errorf("got %q, want 203.0.113.50", got)
	}
}

func TestClientIPFromXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:12345"
	req.Header.Set("X-Real-IP", "198.51.100.77")

	got := clientIP(req)
	if got != "198.51.100.77" {
		t.Errorf("got %q, want 198.51.100.77", got)
	}
}

func TestSchemeOfHTTP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	got := schemeOf(req)
	if got != "http" {
		t.Errorf("got %q, want http", got)
	}
}

func TestSchemeOfXForwardedProto(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	got := schemeOf(req)
	if got != "https" {
		t.Errorf("got %q, want https", got)
	}
}