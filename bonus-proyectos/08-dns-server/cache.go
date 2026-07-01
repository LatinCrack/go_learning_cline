// Package main implements a thread-safe DNS response cache with per-record
// TTL expiration. Entries are stored as raw wire-format bytes for zero-copy
// response delivery. A background janitor goroutine evicts expired entries.
package main

import (
	"log"
	"sync"
	"time"
)

// ────────────────────────────────────────────────────────
// Cache Entry
// ────────────────────────────────────────────────────────

// CacheEntry holds a cached DNS response packet along with its expiration time.
type CacheEntry struct {
	Response   []byte    // Raw wire-format DNS response bytes
	ExpiresAt  time.Time // Absolute expiration timestamp
	MinTTL     uint32    // The minimum TTL from the original answer records
}

// IsExpired returns true if the cache entry has passed its TTL.
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// RemainingTTL returns the number of seconds left before expiration.
// Returns 0 if already expired.
func (e *CacheEntry) RemainingTTL() uint32 {
	d := time.Until(e.ExpiresAt)
	if d <= 0 {
		return 0
	}
	return uint32(d.Seconds())
}

// ────────────────────────────────────────────────────────
// Cache Key
// ────────────────────────────────────────────────────────

// CacheKey uniquely identifies a DNS query by the question name, type, and class.
type CacheKey struct {
	Name  string
	Type  uint16
	Class uint16
}

// MakeCacheKey builds a CacheKey from the first question in a DNS message.
func MakeCacheKey(q Question) CacheKey {
	return CacheKey{
		Name:  normalizeDomain(q.Name),
		Type:  q.Type,
		Class: q.Class,
	}
}

// normalizeDomain lowercases and removes the trailing dot from a domain name
// for consistent cache key comparison.
func normalizeDomain(name string) string {
	// Fast path: lowercase in place.
	b := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			c += 32 // to lowercase
		}
		b[i] = c
	}
	s := string(b)
	// Remove trailing dot.
	if len(s) > 0 && s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s
}

// ────────────────────────────────────────────────────────
// DNS Cache
// ────────────────────────────────────────────────────────

// DNSCache is a thread-safe in-memory cache for DNS responses.
//
// It uses a RWMutex to allow concurrent read access (cache hits) while
// serializing writes (cache insertions and evictions). A background
// janitor goroutine periodically cleans expired entries.
type DNSCache struct {
	mu      sync.RWMutex
	entries map[CacheKey]*CacheEntry

	// Metrics (protected by mu).
	hits   uint64
	misses uint64

	// janitorDone signals the janitor goroutine to stop.
	janitorDone chan struct{}
}

// NewDNSCache creates a new DNS cache and starts the background
// eviction goroutine that runs every cleanupInterval.
func NewDNSCache(cleanupInterval time.Duration) *DNSCache {
	c := &DNSCache{
		entries:     make(map[CacheKey]*CacheEntry),
		janitorDone: make(chan struct{}),
	}

	go c.janitor(cleanupInterval)

	return c
}

// ────────────────────────────────────────────────────────
// Public API
// ────────────────────────────────────────────────────────

// Lookup retrieves a cached DNS response for the given question.
// Returns the raw response bytes and true on a cache hit (not expired).
// Returns nil, false on a cache miss or if the entry has expired.
func (c *DNSCache) Lookup(q Question) ([]byte, bool) {
	key := MakeCacheKey(q)

	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists || entry.IsExpired() {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	c.hits++
	c.mu.Unlock()

	return entry.Response, true
}

// Store inserts a DNS response into the cache with the given TTL.
// If TTL is 0, the entry is not cached (per RFC, TTL=0 means "do not cache").
func (c *DNSCache) Store(q Question, response []byte, ttl uint32) {
	if ttl == 0 {
		return // Do not cache zero-TTL responses.
	}

	// Cap maximum TTL to 24 hours to prevent stale entries.
	const maxTTL = 86400
	if ttl > maxTTL {
		ttl = maxTTL
	}

	key := MakeCacheKey(q)

	// Deep-copy the response to avoid aliasing the caller's buffer.
	copied := make([]byte, len(response))
	copy(copied, response)

	entry := &CacheEntry{
		Response:  copied,
		ExpiresAt: time.Now().Add(time.Duration(ttl) * time.Second),
		MinTTL:    ttl,
	}

	c.mu.Lock()
	c.entries[key] = entry
	c.mu.Unlock()
}

// Flush removes all entries from the cache.
func (c *DNSCache) Flush() {
	c.mu.Lock()
	c.entries = make(map[CacheKey]*CacheEntry)
	c.mu.Unlock()
}

// Size returns the number of entries currently in the cache (including expired).
func (c *DNSCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns cache hit/miss counters.
func (c *DNSCache) Stats() (hits, misses uint64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses
}

// HitRate returns the cache hit ratio as a percentage (0.0 - 100.0).
func (c *DNSCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0.0
	}
	return float64(c.hits) / float64(total) * 100.0
}

// Stop halts the background janitor goroutine.
func (c *DNSCache) Stop() {
	close(c.janitorDone)
}

// ────────────────────────────────────────────────────────
// Background Janitor
// ────────────────────────────────────────────────────────

// janitor runs as a background goroutine, periodically evicting
// expired entries from the cache.
func (c *DNSCache) janitor(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.evictExpired()
		case <-c.janitorDone:
			return
		}
	}
}

// evictExpired removes all expired entries from the cache.
func (c *DNSCache) evictExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	evicted := 0
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
			evicted++
		}
	}

	if evicted > 0 {
		log.Printf("[cache] Evicted %d expired entries (remaining: %d)", evicted, len(c.entries))
	}
}