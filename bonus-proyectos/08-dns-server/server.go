// Package main implements the DNS server's UDP listener and request
// processing pipeline. Each incoming query is handled in its own
// goroutine for maximum concurrency. The processing flow is:
//
//	UDP Packet In → Parse → Blocklist Check → Cache Lookup → Upstream Resolve → Cache Store → Respond
package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

// ────────────────────────────────────────────────────────
// Server Configuration
// ────────────────────────────────────────────────────────

// DNSServer is the main DNS server that listens for queries,
// applies blocklist filtering, serves from cache, and forwards
// to upstream resolvers.
type DNSServer struct {
	// ListenAddr is the UDP address to bind (e.g., ":5353").
	ListenAddr string

	// Resolver handles upstream DNS forwarding.
	Resolver *Resolver

	// Cache stores recent DNS responses with TTL.
	Cache *DNSCache

	// Blocklist filters known bad domains.
	Blocklist *Blocklist

	// DefaultTTL is used for cache storage when upstream TTL is 0.
	DefaultTTL uint32

	// conn is the underlying UDP packet connection.
	conn *net.UDPConn

	// Metrics.
	queryCount   atomic.Uint64
	blockedCount atomic.Uint64
	cacheHits    atomic.Uint64
	errors       atomic.Uint64

	// shutdown signals the server to stop.
	shutdown chan struct{}
	wg       sync.WaitGroup
}

// NewDNSServer creates a new DNS server with the given components.
func NewDNSServer(
	addr string,
	resolver *Resolver,
	cache *DNSCache,
	blocklist *Blocklist,
	defaultTTL uint32,
) *DNSServer {
	return &DNSServer{
		ListenAddr: addr,
		Resolver:   resolver,
		Cache:      cache,
		Blocklist:  blocklist,
		DefaultTTL: defaultTTL,
		shutdown:   make(chan struct{}),
	}
}

// ────────────────────────────────────────────────────────
// Lifecycle
// ────────────────────────────────────────────────────────

// Start begins listening for DNS queries on UDP. It blocks until
// Stop() is called or a fatal error occurs.
func (s *DNSServer) Start() error {
	addr, err := net.ResolveUDPAddr("udp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("server: resolve listen address: %w", err)
	}

	s.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("server: listen udp %s: %w", s.ListenAddr, err)
	}

	log.Printf("[server] DNS server listening on UDP %s", s.ListenAddr)
	log.Printf("[server] Upstream: %v", s.Resolver.Upstreams)
	log.Printf("[server] Cache TTL default: %ds", s.DefaultTTL)
	log.Printf("[server] Blocklist entries: %d (mode: %s)", s.Blocklist.Size(), blockModeString(s.Blocklist.Mode))

	// Read loop — each packet spawns a goroutine.
	s.wg.Add(1)
	go s.readLoop()

	return nil
}

// Stop gracefully shuts down the DNS server.
func (s *DNSServer) Stop() {
	log.Println("[server] Shutting down...")
	close(s.shutdown)
	if s.conn != nil {
		s.conn.Close()
	}
	s.wg.Wait()
	log.Printf("[server] Stopped. Queries: %d | Blocked: %d | Cache hits: %d | Errors: %d",
		s.queryCount.Load(), s.blockedCount.Load(), s.cacheHits.Load(), s.errors.Load())
}

// ────────────────────────────────────────────────────────
// Read Loop
// ────────────────────────────────────────────────────────

// readLoop continuously reads UDP packets and dispatches them
// to handlePacket in individual goroutines.
func (s *DNSServer) readLoop() {
	defer s.wg.Done()

	buf := make([]byte, 512) // Standard DNS UDP packet size

	for {
		n, remoteAddr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-s.shutdown:
				return // Expected shutdown
			default:
				log.Printf("[server] Read error: %v", err)
				continue
			}
		}

		if n == 0 {
			continue
		}

		// Copy the packet data to avoid overwriting in the next read.
		packetData := make([]byte, n)
		copy(packetData, buf[:n])

		// Handle each query in its own goroutine.
		s.wg.Add(1)
		go s.handlePacket(packetData, remoteAddr)
	}
}

// ────────────────────────────────────────────────────────
// Packet Handler (Pipeline)
// ────────────────────────────────────────────────────────

// handlePacket processes a single DNS query packet through the full pipeline:
//  1. Parse the incoming DNS message
//  2. Check if the domain is blocked
//  3. Check the cache for a valid response
//  4. Forward to upstream resolver
//  5. Cache the upstream response
//  6. Send the response to the client
func (s *DNSServer) handlePacket(data []byte, remote *net.UDPAddr) {
	defer s.wg.Done()
	s.queryCount.Add(1)

	// ── Step 1: Parse the incoming query ──
	query, err := ParseMessage(data)
	if err != nil {
		log.Printf("[server] Parse error from %s: %v", remote, err)
		s.errors.Add(1)
		return
	}

	// We only handle standard queries (opcode=0) with at least one question.
	if query.Header.Opcode() != OpcodeQuery || len(query.Questions) == 0 {
		log.Printf("[server] Unsupported opcode=%d from %s", query.Header.Opcode(), remote)
		s.errors.Add(1)
		return
	}

	q := query.Questions[0]
	log.Printf("[server] Query from %s: %s %s (id=%04x)",
		remote, q.Name, TypeToString(q.Type), query.Header.ID)

	// ── Step 2: Blocklist check ──
	if s.Blocklist.IsBlocked(q.Name) {
		s.blockedCount.Add(1)
		log.Printf("[server] BLOCKED: %s (query from %s)", q.Name, remote)

		var response []byte
		switch s.Blocklist.Mode {
		case BlockModeSinkhole:
			// Return an A record pointing to the sinkhole IP.
			rr := BuildARecord(q.Name, s.Blocklist.SinkholeIP, 300)
			response = BuildResponse(query, []ResourceRecord{rr}, RcodeNoError)
		default:
			// Return NXDOMAIN (domain does not exist).
			response = BuildNXDOMAINResponse(query)
		}

		if err := s.sendResponse(response, remote); err != nil {
			log.Printf("[server] Send blocked response error: %v", err)
		}
		return
	}

	// ── Step 3: Cache lookup ──
	if cached, ok := s.Cache.Lookup(q); ok {
		s.cacheHits.Add(1)
		log.Printf("[server] Cache HIT for %s %s", q.Name, TypeToString(q.Type))

		if err := s.sendResponse(cached, remote); err != nil {
			log.Printf("[server] Send cached response error: %v", err)
		}
		return
	}

	// ── Step 4: Forward to upstream resolver ──
	response, err := s.Resolver.ResolveForward(data)
	if err != nil {
		log.Printf("[server] Upstream resolution failed for %s: %v", q.Name, err)
		s.errors.Add(1)

		// Send SERVFAIL to the client.
		servfail := BuildResponse(query, nil, RcodeServerFailure)
		if err := s.sendResponse(servfail, remote); err != nil {
			log.Printf("[server] Send SERVFAIL error: %v", err)
		}
		return
	}

	// ── Step 5: Cache the upstream response ──
	ttl, err := ExtractMinTTL(response, s.DefaultTTL)
	if err != nil {
		log.Printf("[server] TTL extraction error for %s: %v (using default %ds)",
			q.Name, err, s.DefaultTTL)
		ttl = s.DefaultTTL
	}
	s.Cache.Store(q, response, ttl)

	log.Printf("[server] Upstream OK for %s %s (TTL=%ds)", q.Name, TypeToString(q.Type), ttl)

	// ── Step 6: Send the response to the client ──
	if err := s.sendResponse(response, remote); err != nil {
		log.Printf("[server] Send upstream response error: %v", err)
	}
}

// sendResponse writes a DNS response packet to the client via UDP.
func (s *DNSServer) sendResponse(data []byte, remote *net.UDPAddr) error {
	_, err := s.conn.WriteToUDP(data, remote)
	return err
}

// ────────────────────────────────────────────────────────
// Metrics
// ────────────────────────────────────────────────────────

// ServerMetrics holds a snapshot of the server's operational counters.
type ServerMetrics struct {
	Queries    uint64
	Blocked    uint64
	CacheHits  uint64
	Errors     uint64
	CacheSize  int
	BLockSize  int
	HitRate    float64
}

// Metrics returns the current server metrics.
func (s *DNSServer) Metrics() ServerMetrics {
	return ServerMetrics{
		Queries:   s.queryCount.Load(),
		Blocked:   s.blockedCount.Load(),
		CacheHits: s.cacheHits.Load(),
		Errors:    s.errors.Load(),
		CacheSize: s.Cache.Size(),
		BLockSize: s.Blocklist.Size(),
		HitRate:   s.Cache.HitRate(),
	}
}

// ────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────

// blockModeString returns a human-readable string for the block mode.
func blockModeString(mode BlockMode) string {
	switch mode {
	case BlockModeSinkhole:
		return "SINKHOLE"
	default:
		return "NXDOMAIN"
	}
}