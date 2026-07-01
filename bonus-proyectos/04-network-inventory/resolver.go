package main

import (
	"context"
	"net"
	"strings"
	"time"
)

const dnsTimeout = 2 * time.Second

// reverseLookup performs a DNS reverse lookup for the given IP address using
// net.LookupAddr. It returns the first resolved hostname (without trailing dot)
// or an empty string if the lookup fails or times out.
//
// This function is intentionally simple and synchronous — it is called from
// within a goroutine pool, so the concurrency is already managed externally.
func reverseLookup(ip string) string {
	// Use a custom resolver with a short timeout to avoid blocking.
	resolver := &net.Resolver{
		PreferGo: false, // Use OS resolver for broader compatibility.
	}

	ctx := &timeoutContext{timeout: dnsTimeout}
	names, err := resolver.LookupAddr(ctx.asContext(), ip)
	if err != nil {
		return ""
	}

	if len(names) == 0 {
		return ""
	}

	// DNS reverse lookups return FQDNs with a trailing dot (e.g., "host.example.com.").
	// Strip the trailing dot for cleaner display.
	hostname := strings.TrimSuffix(names[0], ".")
	return hostname
}

// timeoutContext is a minimal wrapper to provide a context-like timeout for
// the resolver without importing the full context package at this level.
// In practice, net.Resolver.LookupAddr accepts a context.Context, so we use
// a proper context with timeout.
type timeoutContext struct {
	timeout time.Duration
}

func (t *timeoutContext) asContext() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	// The cancel function is intentionally not returned to keep the API simple.
	// The context will be garbage collected after the timeout elapses.
	_ = cancel
	return ctx
}

// EnrichHosts takes a slice of HostResult and performs DNS reverse lookups
// for all alive hosts that don't yet have a hostname resolved. This can be
// used as a second-pass enrichment if needed.
//
// The enrichment is performed concurrently using a bounded worker pool.
func EnrichHosts(results []HostResult, maxWorkers int) []HostResult {
	if maxWorkers < 1 {
		maxWorkers = 10
	}

	sem := make(chan struct{}, maxWorkers)
	done := make(chan struct{})

	for i := range results {
		if !results[i].Alive || results[i].Hostname != "" {
			continue
		}

		sem <- struct{}{}
		go func(idx int) {
			defer func() { <-sem }()
			results[idx].Hostname = reverseLookup(results[idx].IP.String())
		}(i)
	}

	go func() {
		// Drain semaphore to ensure all workers complete.
		for i := 0; i < cap(sem); i++ {
			sem <- struct{}{}
		}
		close(done)
	}()

	<-done
	return results
}
