package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// ────────────────────────────────────────────────────────
// Protocol Tests
// ────────────────────────────────────────────────────────

func TestBuildName(t *testing.T) {
	tests := []struct {
		input    string
		expected []byte
	}{
		{"", []byte{0}},
		{"com", []byte{3, 'c', 'o', 'm', 0}},
		{"example.com", []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}},
		{"example.com.", []byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}},
		{"sub.example.com", []byte{3, 's', 'u', 'b', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}},
	}

	for _, tt := range tests {
		result := BuildName(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("BuildName(%q): got len %d, want %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("BuildName(%q)[%d]: got %d, want %d", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}

func TestBuildAndParseQuestion(t *testing.T) {
	q := Question{
		Name:  "example.com.",
		Type:  TypeA,
		Class: ClassIN,
	}

	data := BuildQuestion(q)
	if len(data) == 0 {
		t.Fatal("BuildQuestion returned empty data")
	}

	// Verify the encoded type and class are at the end.
	typeClass := data[len(data)-4:]
	decodedType := binary.BigEndian.Uint16(typeClass[0:2])
	decodedClass := binary.BigEndian.Uint16(typeClass[2:4])

	if decodedType != TypeA {
		t.Errorf("Question type: got %d, want %d", decodedType, TypeA)
	}
	if decodedClass != ClassIN {
		t.Errorf("Question class: got %d, want %d", decodedClass, ClassIN)
	}
}

func TestParseMessage_TooShort(t *testing.T) {
	shortData := []byte{0x00, 0x01, 0x02}
	_, err := ParseMessage(shortData)
	if err == nil {
		t.Error("ParseMessage should fail on packet shorter than 12 bytes")
	}
}

func TestBuildResponse_Structure(t *testing.T) {
	req := &Message{
		Header: Header{
			ID:      0x1234,
			QDCount: 1,
		},
		Questions: []Question{
			{Name: "example.com.", Type: TypeA, Class: ClassIN},
		},
	}
	req.Header.SetRD(true)

	ip := [4]byte{93, 184, 216, 34}
	rr := BuildARecord("example.com.", ip, 300)
	resp := BuildResponse(req, []ResourceRecord{rr}, RcodeNoError)

	if len(resp) < dnsHeaderSize {
		t.Fatal("Response too short")
	}

	// Verify ID matches.
	respID := binary.BigEndian.Uint16(resp[0:2])
	if respID != 0x1234 {
		t.Errorf("Response ID: got %04x, want %04x", respID, 0x1234)
	}

	// Verify QR bit is set.
	flags := binary.BigEndian.Uint16(resp[2:4])
	if flags&0x8000 == 0 {
		t.Error("QR bit should be set in response")
	}

	// Verify RA bit is set.
	if flags&0x0080 == 0 {
		t.Error("RA bit should be set in response")
	}

	// Verify RD bit preserved.
	if flags&0x0100 == 0 {
		t.Error("RD bit should be preserved in response")
	}

	// Verify RCODE is 0.
	if flags&0x000F != 0 {
		t.Errorf("RCODE: got %d, want 0", flags&0x000F)
	}

	// Verify ANCOUNT = 1.
	ancount := binary.BigEndian.Uint16(resp[6:8])
	if ancount != 1 {
		t.Errorf("ANCOUNT: got %d, want 1", ancount)
	}
}

func TestBuildNXDOMAINResponse(t *testing.T) {
	req := &Message{
		Header: Header{
			ID:      0xABCD,
			QDCount: 1,
		},
		Questions: []Question{
			{Name: "blocked.com.", Type: TypeA, Class: ClassIN},
		},
	}

	resp := BuildNXDOMAINResponse(req)
	flags := binary.BigEndian.Uint16(resp[2:4])
	rcode := flags & 0x000F
	if rcode != RcodeNameError {
		t.Errorf("NXDOMAIN rcode: got %d, want %d", rcode, RcodeNameError)
	}

	// Parse the response to verify it's valid.
	msg, err := ParseMessage(resp)
	if err != nil {
		t.Fatalf("Failed to parse NXDOMAIN response: %v", err)
	}
	if msg.Header.Rcode() != RcodeNameError {
		t.Errorf("Parsed rcode: got %d, want %d", msg.Header.Rcode(), RcodeNameError)
	}
}

func TestTypeToString(t *testing.T) {
	tests := []struct {
		t    uint16
		want string
	}{
		{TypeA, "A"},
		{TypeAAAA, "AAAA"},
		{TypeCNAME, "CNAME"},
		{TypeMX, "MX"},
		{TypeNS, "NS"},
		{999, "TYPE999"},
	}

	for _, tt := range tests {
		got := TypeToString(tt.t)
		if got != tt.want {
			t.Errorf("TypeToString(%d) = %q, want %q", tt.t, got, tt.want)
		}
	}
}

// ────────────────────────────────────────────────────────
// Cache Tests
// ────────────────────────────────────────────────────────

func TestCache_StoreAndLookup(t *testing.T) {
	cache := NewDNSCache(10 * time.Second)
	defer cache.Stop()

	q := Question{Name: "example.com.", Type: TypeA, Class: ClassIN}
	response := []byte{0x01, 0x02, 0x03, 0x04}

	cache.Store(q, response, 300)

	got, ok := cache.Lookup(q)
	if !ok {
		t.Fatal("Cache miss, expected hit")
	}
	if len(got) != len(response) {
		t.Errorf("Response length: got %d, want %d", len(got), len(response))
	}
}

func TestCache_Expired(t *testing.T) {
	cache := NewDNSCache(100 * time.Millisecond)
	defer cache.Stop()

	q := Question{Name: "shortlived.com.", Type: TypeA, Class: ClassIN}
	response := []byte{0x01, 0x02}

	// Store with TTL=1 second.
	cache.Store(q, response, 1)

	// Should hit immediately.
	if _, ok := cache.Lookup(q); !ok {
		t.Fatal("Expected cache hit immediately after store")
	}

	// Wait for expiration.
	time.Sleep(1100 * time.Millisecond)

	if _, ok := cache.Lookup(q); ok {
		t.Error("Expected cache miss after TTL expiration")
	}
}

func TestCache_ZeroTTL(t *testing.T) {
	cache := NewDNSCache(10 * time.Second)
	defer cache.Stop()

	q := Question{Name: "notcached.com.", Type: TypeA, Class: ClassIN}
	cache.Store(q, []byte{0x01}, 0) // TTL=0 should not cache.

	if _, ok := cache.Lookup(q); ok {
		t.Error("Zero-TTL entry should not be cached")
	}
}

func TestCache_HitRate(t *testing.T) {
	cache := NewDNSCache(10 * time.Second)
	defer cache.Stop()

	q1 := Question{Name: "a.com.", Type: TypeA, Class: ClassIN}
	q2 := Question{Name: "b.com.", Type: TypeA, Class: ClassIN}

	cache.Store(q1, []byte{0x01}, 300)

	// 1 hit, 1 miss.
	cache.Lookup(q1) // hit
	cache.Lookup(q2) // miss

	hits, misses := cache.Stats()
	if hits != 1 {
		t.Errorf("Hits: got %d, want 1", hits)
	}
	if misses != 1 {
		t.Errorf("Misses: got %d, want 1", misses)
	}

	rate := cache.HitRate()
	if rate < 49.0 || rate > 51.0 {
		t.Errorf("HitRate: got %.1f%%, want ~50%%", rate)
	}
}

func TestCache_Flush(t *testing.T) {
	cache := NewDNSCache(10 * time.Second)
	defer cache.Stop()

	q := Question{Name: "example.com.", Type: TypeA, Class: ClassIN}
	cache.Store(q, []byte{0x01}, 300)

	if cache.Size() != 1 {
		t.Errorf("Size before flush: got %d, want 1", cache.Size())
	}

	cache.Flush()

	if cache.Size() != 0 {
		t.Errorf("Size after flush: got %d, want 0", cache.Size())
	}
}

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"Example.COM.", "example.com"},
		{"EXAMPLE.COM.", "example.com"},
		{"example.com", "example.com"},
		{"UPPER.", "upper"},
	}

	for _, tt := range tests {
		got := normalizeDomain(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeDomain(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// ────────────────────────────────────────────────────────
// Blocklist Tests
// ────────────────────────────────────────────────────────

func TestBlocklist_ExactMatch(t *testing.T) {
	bl := NewBlocklist()
	bl.Add("ads.example.com")

	if !bl.IsBlocked("ads.example.com.") {
		t.Error("Expected ads.example.com. to be blocked")
	}
	if !bl.IsBlocked("ads.example.com") {
		t.Error("Expected ads.example.com to be blocked")
	}
	if bl.IsBlocked("other.example.com.") {
		t.Error("Expected other.example.com. to NOT be blocked")
	}
}

func TestBlocklist_WildcardMatch(t *testing.T) {
	bl := NewBlocklist()
	bl.Add("*.tracker.net")

	tests := []struct {
		domain  string
		blocked bool
	}{
		{"sub.tracker.net.", true},
		{"a.b.tracker.net.", true},
		{"tracker.net.", true},
		{"other.net.", false},
		{"nottracker.net.", false},
	}

	for _, tt := range tests {
		got := bl.IsBlocked(tt.domain)
		if got != tt.blocked {
			t.Errorf("IsBlocked(%q) = %v, want %v", tt.domain, got, tt.blocked)
		}
	}
}

func TestBlocklist_CaseInsensitive(t *testing.T) {
	bl := NewBlocklist()
	bl.Add("ADS.EXAMPLE.COM")

	if !bl.IsBlocked("ads.example.com.") {
		t.Error("Blocklist should be case-insensitive")
	}
	if !bl.IsBlocked("Ads.Example.Com.") {
		t.Error("Blocklist should be case-insensitive")
	}
}

func TestBlocklist_LoadFromFile(t *testing.T) {
	content := `# Comment line
ads.doubleclick.net
*.tracker.com
malware.net
// Another comment

`
	tmpFile, err := os.CreateTemp("", "blocklist-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(content)
	tmpFile.Close()

	bl := NewBlocklist()
	if err := bl.LoadFromFile(tmpFile.Name()); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if !bl.IsBlocked("ads.doubleclick.net") {
		t.Error("Expected ads.doubleclick.net to be blocked")
	}
	if !bl.IsBlocked("sub.tracker.com") {
		t.Error("Expected sub.tracker.com to be blocked via wildcard")
	}
	if !bl.IsBlocked("malware.net") {
		t.Error("Expected malware.net to be blocked")
	}
	if bl.IsBlocked("google.com") {
		t.Error("Expected google.com to NOT be blocked")
	}
}

func TestBlocklist_Disabled(t *testing.T) {
	bl := NewBlocklist()
	bl.Add("blocked.com")
	bl.Enabled = false

	if bl.IsBlocked("blocked.com.") {
		t.Error("Disabled blocklist should not block anything")
	}
}

func TestBlocklist_Remove(t *testing.T) {
	bl := NewBlocklist()
	bl.Add("temp.com")

	if !bl.IsBlocked("temp.com") {
		t.Error("Expected temp.com to be blocked before remove")
	}

	bl.Remove("temp.com")

	if bl.IsBlocked("temp.com") {
		t.Error("Expected temp.com to NOT be blocked after remove")
	}
}

// ────────────────────────────────────────────────────────
// Resolver Tests
// ────────────────────────────────────────────────────────

func TestNewResolver_DefaultUpstream(t *testing.T) {
	r := NewResolver(nil, 5*time.Second)
	if len(r.Upstreams) != 1 {
		t.Fatalf("Expected 1 default upstream, got %d", len(r.Upstreams))
	}
	if !strings.Contains(r.Upstreams[0], "8.8.8.8") {
		t.Errorf("Default upstream: got %s, want 8.8.8.8", r.Upstreams[0])
	}
}

func TestNewResolver_PortAppending(t *testing.T) {
	r := NewResolver([]string{"1.1.1.1", "8.8.4.4:5353"}, 5*time.Second)

	if r.Upstreams[0] != "1.1.1.1:53" {
		t.Errorf("Upstream[0]: got %s, want 1.1.1.1:53", r.Upstreams[0])
	}
	if r.Upstreams[1] != "8.8.4.4:5353" {
		t.Errorf("Upstream[1]: got %s, want 8.8.4.4:5353", r.Upstreams[1])
	}
}

// ────────────────────────────────────────────────────────
// Integration Test: Build → Parse round-trip
// ────────────────────────────────────────────────────────

func TestRoundTrip_BuildAndParse(t *testing.T) {
	// Build a response with an A record.
	req := &Message{
		Header: Header{
			ID:      0x5555,
			QDCount: 1,
		},
		Questions: []Question{
			{Name: "test.example.com.", Type: TypeA, Class: ClassIN},
		},
	}
	req.Header.SetRD(true)

	ip := [4]byte{192, 168, 1, 100}
	rr := BuildARecord("test.example.com.", ip, 600)
	resp := BuildResponse(req, []ResourceRecord{rr}, RcodeNoError)

	// Parse the response back.
	msg, err := ParseMessage(resp)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	if msg.Header.ID != 0x5555 {
		t.Errorf("ID: got %04x, want %04x", msg.Header.ID, 0x5555)
	}
	if !msg.Header.QR() {
		t.Error("QR should be set")
	}
	if !msg.Header.RD() {
		t.Error("RD should be set")
	}
	if !msg.Header.RA() {
		t.Error("RA should be set")
	}
	if len(msg.Questions) != 1 {
		t.Fatalf("Questions: got %d, want 1", len(msg.Questions))
	}
	if msg.Questions[0].Name != "test.example.com." {
		t.Errorf("Question name: got %q, want %q", msg.Questions[0].Name, "test.example.com.")
	}
	if len(msg.Answers) != 1 {
		t.Fatalf("Answers: got %d, want 1", len(msg.Answers))
	}
	if msg.Answers[0].Type != TypeA {
		t.Errorf("Answer type: got %d, want %d", msg.Answers[0].Type, TypeA)
	}
	if msg.Answers[0].TTL != 600 {
		t.Errorf("Answer TTL: got %d, want 600", msg.Answers[0].TTL)
	}

	// Verify the RDATA contains the correct IP.
	answerIP := IPFromRData(msg.Answers[0].RData)
	if answerIP != ip {
		t.Errorf("Answer IP: got %v, want %v", answerIP, ip)
	}
}

// ────────────────────────────────────────────────────────
// DNS Helper: build a raw query packet for integration tests
// ────────────────────────────────────────────────────────

// buildRawQuery constructs a minimal DNS A-record query packet.
func buildRawQuery(id uint16, domain string) []byte {
	q := Question{Name: domain, Type: TypeA, Class: ClassIN}

	var buf []byte
	hdr := make([]byte, dnsHeaderSize)
	binary.BigEndian.PutUint16(hdr[0:2], id)
	binary.BigEndian.PutUint16(hdr[2:4], 0x0100) // RD=1
	binary.BigEndian.PutUint16(hdr[4:6], 1)      // QDCOUNT=1
	buf = append(buf, hdr...)
	buf = append(buf, BuildQuestion(q)...)

	return buf
}

// ────────────────────────────────────────────────────────
// Integration Test: Full UDP round-trip
// ────────────────────────────────────────────────────────

func TestIntegration_UDPQuery(t *testing.T) {
	// Use a random high port.
	listenAddr := "127.0.0.1:0" // OS picks port

	// Create a mock upstream DNS server that always responds with an A record.
	mockUpstream := startMockDNS(t, "127.0.0.1:0")
	defer mockUpstream.Close()

	// Get the mock upstream address.
	mockAddr := mockUpstream.LocalAddr().(*net.UDPAddr)
	upstreamStr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", mockAddr.Port))

	// Create the server components.
	bl := NewBlocklist()
	bl.Add("blocked.test.")

	cache := NewDNSCache(10 * time.Second)
	defer cache.Stop()

	resolver := NewResolver([]string{upstreamStr}, 2*time.Second)

	server := NewDNSServer(listenAddr, resolver, cache, bl, 300)
	if err := server.Start(); err != nil {
		t.Fatalf("Server start failed: %v", err)
	}
	defer server.Stop()

	// Get the actual listening address.
	serverAddr := server.conn.LocalAddr().(*net.UDPAddr)

	t.Run("NormalQuery", func(t *testing.T) {
		queryData := buildRawQuery(0x1111, "example.com.")

		conn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		if _, err := conn.Write(queryData); err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		resp, err := ParseMessage(buf[:n])
		if err != nil {
			t.Fatalf("Parse response failed: %v", err)
		}

		if resp.Header.ID != 0x1111 {
			t.Errorf("Response ID: got %04x, want %04x", resp.Header.ID, 0x1111)
		}
		if resp.Header.Rcode() != RcodeNoError {
			t.Errorf("Rcode: got %d, want %d", resp.Header.Rcode(), RcodeNoError)
		}
	})

	t.Run("BlockedQuery", func(t *testing.T) {
		queryData := buildRawQuery(0x2222, "blocked.test.")

		conn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		if _, err := conn.Write(queryData); err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}

		resp, err := ParseMessage(buf[:n])
		if err != nil {
			t.Fatalf("Parse response failed: %v", err)
		}

		if resp.Header.Rcode() != RcodeNameError {
			t.Errorf("Blocked query rcode: got %d, want NXDOMAIN (%d)",
				resp.Header.Rcode(), RcodeNameError)
		}
	})

	t.Run("CacheHit", func(t *testing.T) {
		// First query populates cache.
		queryData := buildRawQuery(0x3333, "cached.test.")

		conn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		if _, err := conn.Write(queryData); err != nil {
			t.Fatalf("Write failed: %v", err)
		}

		buf := make([]byte, 512)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("First read failed: %v", err)
		}

		resp1, err := ParseMessage(buf[:n])
		if err != nil {
			t.Fatalf("First parse failed: %v", err)
		}

		// Second query should hit cache (same domain).
		queryData2 := buildRawQuery(0x4444, "cached.test.")
		if _, err := conn.Write(queryData2); err != nil {
			t.Fatalf("Second write failed: %v", err)
		}

		n, err = conn.Read(buf)
		if err != nil {
			t.Fatalf("Second read failed: %v", err)
		}

		resp2, err := ParseMessage(buf[:n])
		if err != nil {
			t.Fatalf("Second parse failed: %v", err)
		}

		// Both should succeed.
		if resp1.Header.Rcode() != RcodeNoError || resp2.Header.Rcode() != RcodeNoError {
			t.Error("Both queries should succeed")
		}

		// Verify cache metrics.
		hits, _ := cache.Stats()
		if hits == 0 {
			t.Error("Expected at least one cache hit")
		}
	})
}

// ────────────────────────────────────────────────────────
// Mock DNS Server for integration tests
// ────────────────────────────────────────────────────────

type mockDNS struct {
	conn *net.UDPConn
}

func startMockDNS(t *testing.T, addr string) *net.UDPConn {
	t.Helper()

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatalf("Mock DNS resolve addr: %v", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("Mock DNS listen: %v", err)
	}

	go func() {
		buf := make([]byte, 512)
		for {
			n, remote, err := conn.ReadFromUDP(buf)
			if err != nil {
				return // Closed
			}
			if n == 0 {
				continue
			}

			// Parse the query.
			query, err := ParseMessage(buf[:n])
			if err != nil {
				continue
			}

			// Build a mock A record response.
			var answers []ResourceRecord
			if len(query.Questions) > 0 {
				ip := [4]byte{93, 184, 216, 34} // Fixed mock IP
				rr := BuildARecord(query.Questions[0].Name, ip, 300)
				answers = append(answers, rr)
			}

			resp := BuildResponse(query, answers, RcodeNoError)
			conn.WriteToUDP(resp, remote)
		}
	}()

	return conn
}

// ────────────────────────────────────────────────────────
// Utility: parseIPv4 tests
// ────────────────────────────────────────────────────────

func TestParseIPv4(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		ip    [4]byte
	}{
		{"127.0.0.1", true, [4]byte{127, 0, 0, 1}},
		{"192.168.1.100", true, [4]byte{192, 168, 1, 100}},
		{"0.0.0.0", true, [4]byte{0, 0, 0, 0}},
		{"255.255.255.255", true, [4]byte{255, 255, 255, 255}},
		{"invalid", false, [4]byte{}},
		{"999.1.1.1", false, [4]byte{}},
		{"1.2.3", false, [4]byte{}},
	}

	for _, tt := range tests {
		ip, err := parseIPv4(tt.input)
		if tt.valid {
			if err != nil {
				t.Errorf("parseIPv4(%q): unexpected error: %v", tt.input, err)
			}
			if ip != tt.ip {
				t.Errorf("parseIPv4(%q): got %v, want %v", tt.input, ip, tt.ip)
			}
		} else {
			if err == nil {
				t.Errorf("parseIPv4(%q): expected error, got nil", tt.input)
			}
		}
	}
}

func TestParseUpstreamList(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"8.8.8.8", []string{"8.8.8.8:53"}},
		{"8.8.8.8,1.1.1.1", []string{"8.8.8.8:53", "1.1.1.1:53"}},
		{"8.8.8.8:5353,1.1.1.1", []string{"8.8.8.8:5353", "1.1.1.1:53"}},
		{" 8.8.8.8 , 1.1.1.1 ", []string{"8.8.8.8:53", "1.1.1.1:53"}},
	}

	for _, tt := range tests {
		got := parseUpstreamList(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("parseUpstreamList(%q): got len %d, want %d", tt.input, len(got), len(tt.expected))
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("parseUpstreamList(%q)[%d]: got %s, want %s", tt.input, i, got[i], tt.expected[i])
			}
		}
	}
}

// ────────────────────────────────────────────────────────
// Benchmark
// ────────────────────────────────────────────────────────

func BenchmarkBuildName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildName("sub.domain.example.com")
	}
}

func BenchmarkParseMessage(b *testing.B) {
	req := &Message{
		Header: Header{ID: 0x1234, QDCount: 1},
		Questions: []Question{
			{Name: "example.com.", Type: TypeA, Class: ClassIN},
		},
	}
	ip := [4]byte{93, 184, 216, 34}
	rr := BuildARecord("example.com.", ip, 300)
	data := BuildResponse(req, []ResourceRecord{rr}, RcodeNoError)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMessage(data)
	}
}

func BenchmarkBlocklistLookup(b *testing.B) {
	bl := NewBlocklist()
	bl.Add("ads.example.com")
	bl.Add("tracker.malware.net")
	bl.Add("*.doubleclick.net")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bl.IsBlocked("ads.example.com.")
	}
}

func BenchmarkCacheLookup(b *testing.B) {
	cache := NewDNSCache(60 * time.Second)
	defer cache.Stop()

	q := Question{Name: "example.com.", Type: TypeA, Class: ClassIN}
	cache.Store(q, []byte{0x01, 0x02, 0x03, 0x04}, 300)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Lookup(q)
	}
}
