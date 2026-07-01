// Package main implements upstream DNS resolution.
//
// The resolver forwards DNS queries to one or more upstream servers
// (e.g., 8.8.8.8, 1.1.1.1) over UDP, parses the responses, extracts
// the minimum TTL for cache storage, and returns the raw wire-format
// response bytes.
package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// ────────────────────────────────────────────────────────
// Resolver Configuration
// ────────────────────────────────────────────────────────

// Resolver handles forwarding DNS queries to upstream servers.
type Resolver struct {
	// Upstreams is the list of upstream DNS server addresses in "host:port" format.
	Upstreams []string
	// Timeout is the maximum time to wait for an upstream response.
	Timeout time.Duration
	// currentUpstream tracks the round-robin index for upstream selection.
	currentUpstream int
}

// NewResolver creates a new upstream DNS resolver.
// upstreams should be addresses like "8.8.8.8:53" or "1.1.1.1:53".
// If no upstreams are provided, it defaults to Google DNS (8.8.8.8:53).
func NewResolver(upstreams []string, timeout time.Duration) *Resolver {
	if len(upstreams) == 0 {
		upstreams = []string{"8.8.8.8:53"}
	}

	// Ensure all upstream addresses have a port.
	for i, u := range upstreams {
		_, _, err := net.SplitHostPort(u)
		if err != nil {
			// Assume port 53 if not specified.
			upstreams[i] = net.JoinHostPort(u, "53")
		}
	}

	return &Resolver{
		Upstreams: upstreams,
		Timeout:   timeout,
	}
}

// ────────────────────────────────────────────────────────
// Resolution
// ────────────────────────────────────────────────────────

// ResolveForward forwards a raw DNS query packet to an upstream server
// and returns the raw response bytes. It uses round-robin across configured
// upstreams and falls back to the next server on timeout/error.
func (r *Resolver) ResolveForward(queryData []byte) ([]byte, error) {
	if len(r.Upstreams) == 0 {
		return nil, fmt.Errorf("resolver: no upstream servers configured")
	}

	var lastErr error
	startIdx := r.currentUpstream

	// Try each upstream once (round-robin with fallback).
	for i := 0; i < len(r.Upstreams); i++ {
		idx := (startIdx + i) % len(r.Upstreams)
		upstream := r.Upstreams[idx]

		resp, err := r.queryUpstream(upstream, queryData)
		if err == nil {
			// Rotate to next upstream for the next query.
			r.currentUpstream = (idx + 1) % len(r.Upstreams)
			return resp, nil
		}

		log.Printf("[resolver] Upstream %s failed: %v", upstream, err)
		lastErr = err
	}

	return nil, fmt.Errorf("resolver: all upstreams failed, last error: %w", lastErr)
}

// queryUpstream sends a DNS query to a single upstream server via UDP.
func (r *Resolver) queryUpstream(upstream string, queryData []byte) ([]byte, error) {
	// Resolve the upstream address.
	addr, err := net.ResolveUDPAddr("udp", upstream)
	if err != nil {
		return nil, fmt.Errorf("resolve address %q: %w", upstream, err)
	}

	// Establish a UDP connection with timeout.
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", upstream, err)
	}
	defer conn.Close()

	// Set read/write deadlines.
	if err := conn.SetDeadline(time.Now().Add(r.Timeout)); err != nil {
		return nil, fmt.Errorf("set deadline: %w", err)
	}

	// Send the query.
	if _, err := conn.Write(queryData); err != nil {
		return nil, fmt.Errorf("write to %s: %w", upstream, err)
	}

	// Read the response (DNS over UDP max 512 bytes per RFC 1035 §2.3.4,
	// but we allow up to 4096 for EDNS0 support).
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read from %s: %w", upstream, err)
	}

	if n < dnsHeaderSize {
		return nil, fmt.Errorf("response from %s too short: %d bytes", upstream, n)
	}

	return buf[:n], nil
}

// ────────────────────────────────────────────────────────
// TTL Extraction
// ────────────────────────────────────────────────────────

// ExtractMinTTL parses a DNS response and returns the minimum TTL
// across all answer records. This is used to determine how long
// the response should be cached.
//
// If the response has no answer records, it returns defaultTTL.
func ExtractMinTTL(respData []byte, defaultTTL uint32) (uint32, error) {
	msg, err := ParseMessage(respData)
	if err != nil {
		return 0, fmt.Errorf("extract ttl: %w", err)
	}

	if len(msg.Answers) == 0 {
		return defaultTTL, nil
	}

	minTTL := msg.Answers[0].TTL
	for _, rr := range msg.Answers[1:] {
		if rr.TTL < minTTL {
			minTTL = rr.TTL
		}
	}

	// Ensure a minimum TTL of 1 second to avoid immediate expiry.
	if minTTL == 0 {
		minTTL = 1
	}

	return minTTL, nil
}