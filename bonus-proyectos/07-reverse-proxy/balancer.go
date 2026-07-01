package main

import (
	"log"
	"net/url"
	"sync"
	"sync/atomic"
)

// Backend represents a single upstream server with its URL and health state.
type Backend struct {
	URL    *url.URL
	Alive  atomic.Bool
	weight int // reserved for future weighted algorithms
}

// NewBackend creates a Backend from a raw URL string.
func NewBackend(rawURL string) *Backend {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("[balancer] invalid backend URL %q: %v", rawURL, err)
	}
	b := &Backend{URL: u, weight: 1}
	b.Alive.Store(true)
	return b
}

// Balancer distributes incoming requests across backends using Round Robin.
// It uses atomic operations for lock-free, thread-safe index advancement.
type Balancer struct {
	backends atomic.Pointer[[]*Backend]
	counter  atomic.Uint64
	mu       sync.RWMutex // protects the slice during health-check mutations
}

// NewBalancer creates a Balancer seeded with the provided backend URLs.
func NewBalancer(urls []string) *Balancer {
	b := &Balancer{}
	var backends []*Backend
	for _, u := range urls {
		backends = append(backends, NewBackend(u))
	}
	b.backends.Store(&backends)
	return b
}

// Next returns the next healthy backend using round-robin selection.
// If all backends are unhealthy, it returns nil.
func (b *Balancer) Next() *Backend {
	b.mu.RLock()
	backends := *b.backends.Load()
	total := len(backends)
	b.mu.RUnlock()

	if total == 0 {
		return nil
	}

	// Try every backend once; if all are unhealthy return nil.
	for i := 0; i < total; i++ {
		idx := b.counter.Add(1) - 1
		backend := backends[idx%uint64(total)]
		if backend.Alive.Load() {
			return backend
		}
	}

	return nil
}

// MarkDown sets a backend as unhealthy.
func (b *Balancer) MarkDown(target *url.URL) {
	b.mu.RLock()
	backends := *b.backends.Load()
	b.mu.RUnlock()

	for _, backend := range backends {
		if backend.URL.String() == target.String() {
			backend.Alive.Store(false)
			log.Printf("[balancer] backend marked DOWN: %s", target)
			return
		}
	}
}

// MarkUp sets a backend as healthy.
func (b *Balancer) MarkUp(target *url.URL) {
	b.mu.RLock()
	backends := *b.backends.Load()
	b.mu.RUnlock()

	for _, backend := range backends {
		if backend.URL.String() == target.String() {
			backend.Alive.Store(true)
			log.Printf("[balancer] backend marked UP: %s", target)
			return
		}
	}
}

// Backends returns a snapshot of all registered backends.
func (b *Balancer) Backends() []*Backend {
	b.mu.RLock()
	defer b.mu.RUnlock()
	backends := *b.backends.Load()
	result := make([]*Backend, len(backends))
	copy(result, backends)
	return result
}

// HealthyCount returns the number of backends currently marked as alive.
func (b *Balancer) HealthyCount() int {
	b.mu.RLock()
	backends := *b.backends.Load()
	b.mu.RUnlock()

	count := 0
	for _, backend := range backends {
		if backend.Alive.Load() {
			count++
		}
	}
	return count
}