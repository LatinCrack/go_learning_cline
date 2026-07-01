package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// HTTPClient wraps an http.Client with proxy rotation, cookie management,
// custom headers, and request-level features for bruteforce attacks.
type HTTPClient struct {
	mu             sync.RWMutex
	config         *Config
	clients        []*http.Client   // One client per proxy (or one without proxy).
	proxyIndex     int              // Round-robin proxy index.
	proxyCount     atomic.Int64     // Total proxy rotations (for monitoring).
	cookieJar      *cookiejar.Jar   // Shared cookie jar.
}

// NewHTTPClient creates an HTTPClient configured from the provided Config.
func NewHTTPClient(cfg *Config) (*HTTPClient, error) {
	hc := &HTTPClient{
		config: cfg,
	}

	// Create a shared cookie jar.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}
	hc.cookieJar = jar

	// Set pre-configured cookies if provided.
	if len(cfg.Cookies) > 0 {
		u, err := url.Parse(cfg.TargetURL)
		if err == nil {
			var cookies []*http.Cookie
			for name, value := range cfg.Cookies {
				cookies = append(cookies, &http.Cookie{Name: name, Value: value})
			}
			jar.SetCookies(u, cookies)
		}
	}

	timeout := time.Duration(cfg.Timeout) * time.Second

	if len(cfg.Proxies) == 0 {
		// No proxies: create a single standard client.
		client := &http.Client{
			Timeout: timeout,
			Jar:     jar,
			// Don't follow redirects — return the original response
			// so the detector can analyze the 302 Location header.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		hc.clients = []*http.Client{client}
	} else {
		// Create one client per proxy for round-robin rotation.
		for _, proxyURL := range cfg.Proxies {
			proxyParsed, err := url.Parse(proxyURL)
			if err != nil {
				return nil, fmt.Errorf("parse proxy URL %s: %w", proxyURL, err)
			}
			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyParsed),
			}
			client := &http.Client{
				Timeout:   timeout,
				Jar:       jar,
				Transport: transport,
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			hc.clients = append(hc.clients, client)
		}
	}

	return hc, nil
}

// LoginResponse holds the result of a single login attempt.
type LoginResponse struct {
	StatusCode int
	Body       string
	Headers    http.Header
	Location   string   // Redirect location header, if any.
	Duration   time.Duration
	Error      error
}

// AttemptLogin sends a login request with the given credentials.
// For "form" type, it POSTs the username/password as form data.
// For "basic" type, it uses HTTP Basic Authentication.
func (hc *HTTPClient) AttemptLogin(username, password string) *LoginResponse {
	start := time.Now()
	resp := &LoginResponse{}

	client := hc.nextClient()

	var req *http.Request
	var err error

	if hc.config.TargetType == "basic" {
		req, err = http.NewRequest("GET", hc.config.TargetURL, nil)
		if err != nil {
			resp.Error = fmt.Errorf("create request: %w", err)
			resp.Duration = time.Since(start)
			return resp
		}
		req.SetBasicAuth(username, password)
	} else {
		// Form-based login.
		formData := url.Values{}
		formData.Set(hc.config.UsernameField, username)
		formData.Set(hc.config.PasswordField, password)

		req, err = http.NewRequest("POST", hc.config.TargetURL, strings.NewReader(formData.Encode()))
		if err != nil {
			resp.Error = fmt.Errorf("create request: %w", err)
			resp.Duration = time.Since(start)
			return resp
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// Set common headers.
	req.Header.Set("User-Agent", hc.config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	// Set custom headers.
	for key, value := range hc.config.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Execute the request.
	httpResp, err := client.Do(req)
	if err != nil {
		resp.Error = fmt.Errorf("execute request: %w", err)
		resp.Duration = time.Since(start)
		return resp
	}
	defer httpResp.Body.Close()

	resp.StatusCode = httpResp.StatusCode
	resp.Headers = httpResp.Header
	resp.Location = httpResp.Header.Get("Location")

	// Read response body (limited to 1MB to prevent memory issues).
	body, err := io.ReadAll(io.LimitReader(httpResp.Body, 1024*1024))
	if err != nil {
		resp.Error = fmt.Errorf("read body: %w", err)
		resp.Duration = time.Since(start)
		return resp
	}
	resp.Body = string(body)
	resp.Duration = time.Since(start)

	return resp
}

// nextClient returns the next client in round-robin order (proxy rotation).
func (hc *HTTPClient) nextClient() *http.Client {
	if len(hc.clients) == 1 {
		return hc.clients[0]
	}

	hc.mu.Lock()
	defer hc.mu.Unlock()

	client := hc.clients[hc.proxyIndex]
	hc.proxyIndex = (hc.proxyIndex + 1) % len(hc.clients)
	hc.proxyCount.Add(1)

	return client
}

// ProxyRotations returns the total number of proxy rotations performed.
func (hc *HTTPClient) ProxyRotations() int64 {
	return hc.proxyCount.Load()
}

// Close cleans up the HTTP client resources.
func (hc *HTTPClient) Close() {
	// Nothing to explicitly close; GC handles transport cleanup.
}