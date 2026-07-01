package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// ReverseProxyHandler wraps httputil.ReverseProxy with load balancing,
// header injection (X-Forwarded-For, X-Real-IP), and context-based timeouts.
type ReverseProxyHandler struct {
	balancer       *Balancer
	requestTimeout time.Duration
}

// NewReverseProxy creates a handler that reverse-proxies requests to backends
// selected by the provided Balancer.
func NewReverseProxy(balancer *Balancer, requestTimeout time.Duration) *ReverseProxyHandler {
	return &ReverseProxyHandler{
		balancer:       balancer,
		requestTimeout: requestTimeout,
	}
}

// ServeHTTP selects a backend via round-robin, creates a single-host reverse
// proxy, injects forwarding headers, and serves the request with a timeout.
func (rp *ReverseProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := rp.balancer.Next()
	if backend == nil {
		http.Error(w, "Service Unavailable — no healthy backends", http.StatusServiceUnavailable)
		return
	}

	// Build a reverse proxy targeting the selected backend.
	proxy := newSingleHostReverseProxy(backend.URL)

	// Inject standard proxy headers.
	// Note: X-Forwarded-For is handled automatically by httputil.ReverseProxy,
	// which appends r.RemoteAddr to the existing header value.
	// We only inject X-Real-IP, X-Forwarded-Host, and X-Forwarded-Proto.
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP == "" {
		r.Header.Set("X-Real-IP", clientIP(r))
	}
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Forwarded-Proto", schemeOf(r))

	// Create a context with timeout for the backend connection.
	ctx, cancel := context.WithTimeout(r.Context(), rp.requestTimeout)
	defer cancel()
	r = r.WithContext(ctx)

	// Serve the proxied request.
	proxy.ServeHTTP(w, r)
}

// newSingleHostReverseProxy builds an httputil.ReverseProxy for a single target.
func newSingleHostReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Custom error handler to mark backend down on transport failures.
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Bad Gateway — upstream error: %v\n", err)
	}

	// Configure the transport with sane defaults.
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout:  10 * time.Second,
		MaxIdleConns:           100,
		MaxIdleConnsPerHost:    10,
		IdleConnTimeout:        90 * time.Second,
		DisableCompression:     false,
		ForceAttemptHTTP2:      true,
	}

	return proxy
}

// clientIP extracts the client IP from the request, checking X-Forwarded-For
// first, then X-Real-IP, and finally the remote address.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain (original client).
		for i, c := range xff {
			if c == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// schemeOf determines the request scheme (http or https).
func schemeOf(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if r.URL.Scheme != "" {
		return r.URL.Scheme
	}
	return "http"
}