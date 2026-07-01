// Package main provides binary parsing and construction for DNS packets
// following RFC 1035 wire format. All structures map directly to the
// on-the-wire DNS message layout for maximum efficiency.
//
// DNS Message Format (RFC 1035 §4.1):
//
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                      ID                           |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE      |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    QDCOUNT                         |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    ANCOUNT                         |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    NSCOUNT                         |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
//	|                    ARCOUNT                         |
//	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// ────────────────────────────────────────────────────────
// DNS Constants
// ────────────────────────────────────────────────────────

// DNS opcodes (RFC 1035 §4.1.1).
const (
	OpcodeQuery  = 0
	OpcodeIQuery = 1
	OpcodeStatus = 2
)

// DNS response codes (RFC 1035 §4.1.1).
const (
	RcodeNoError        = 0
	RcodeFormatError    = 1
	RcodeServerFailure  = 2
	RcodeNameError      = 3 // NXDOMAIN
	RcodeNotImplemented = 4
	RcodeRefused        = 5
)

// DNS record types (RFC 1035 §3.2.2 + common extensions).
const (
	TypeA     uint16 = 1
	TypeNS    uint16 = 2
	TypeCNAME uint16 = 5
	TypeSOA   uint16 = 6
	TypeMX    uint16 = 15
	TypeTXT   uint16 = 16
	TypeAAAA  uint16 = 28
	TypeSRV   uint16 = 33
	TypeANY   uint16 = 255
)

// DNS class values (RFC 1035 §3.2.4).
const (
	ClassIN uint16 = 1 // Internet
	ClassCS uint16 = 2
	ClassCH uint16 = 3
	ClassHS uint16 = 4
)

// DNS header size in bytes (fixed: 12 bytes).
const dnsHeaderSize = 12

// ────────────────────────────────────────────────────────
// DNS Header (RFC 1035 §4.1.1)
// ────────────────────────────────────────────────────────

// Header represents the 12-byte DNS message header.
type Header struct {
	ID      uint16 // Packet identifier
	Flags   uint16 // QR(1) + OPCODE(4) + AA(1) + TC(1) + RD(1) + RA(1) + Z(3) + RCODE(4)
	QDCount uint16 // Number of questions
	ANCount uint16 // Number of answers
	NSCount uint16 // Number of authority records
	ARCount uint16 // Number of additional records
}

// QR returns true if this is a response (bit 15).
func (h *Header) QR() bool {
	return h.Flags&0x8000 != 0
}

// SetQR sets the QR bit.
func (h *Header) SetQR(isResponse bool) {
	if isResponse {
		h.Flags |= 0x8000
	} else {
		h.Flags &^= 0x8000
	}
}

// Opcode returns the 4-bit opcode field (bits 11-14).
func (h *Header) Opcode() uint8 {
	return uint8((h.Flags >> 11) & 0x0F)
}

// AA returns the Authoritative Answer bit.
func (h *Header) AA() bool {
	return h.Flags&0x0400 != 0
}

// TC returns the Truncation bit.
func (h *Header) TC() bool {
	return h.Flags&0x0200 != 0
}

// RD returns the Recursion Desired bit.
func (h *Header) RD() bool {
	return h.Flags&0x0100 != 0
}

// SetRD sets the Recursion Desired bit.
func (h *Header) SetRD(desired bool) {
	if desired {
		h.Flags |= 0x0100
	} else {
		h.Flags &^= 0x0100
	}
}

// RA returns the Recursion Available bit.
func (h *Header) RA() bool {
	return h.Flags&0x0080 != 0
}

// SetRA sets the Recursion Available bit.
func (h *Header) SetRA(available bool) {
	if available {
		h.Flags |= 0x0080
	} else {
		h.Flags &^= 0x0080
	}
}

// Rcode returns the 4-bit response code (bits 0-3).
func (h *Header) Rcode() uint8 {
	return uint8(h.Flags & 0x000F)
}

// SetRcode sets the 4-bit response code, preserving upper bits.
func (h *Header) SetRcode(rcode uint8) {
	h.Flags = (h.Flags & 0xFFF0) | (uint16(rcode) & 0x000F)
}

// ────────────────────────────────────────────────────────
// DNS Question (RFC 1035 §4.1.2)
// ────────────────────────────────────────────────────────

// Question represents a single DNS question entry.
type Question struct {
	Name  string // QNAME as dot-separated FQDN
	Type  uint16 // QTYPE
	Class uint16 // QCLASS (usually ClassIN = 1)
}

// ────────────────────────────────────────────────────────
// DNS Resource Record (RFC 1035 §4.1.3)
// ────────────────────────────────────────────────────────

// ResourceRecord represents a DNS answer/authority/additional record.
type ResourceRecord struct {
	Name   string // Owner name
	Type   uint16 // RR type
	Class  uint16 // RR class
	TTL    uint32 // Time to live in seconds
	RDLen  uint16 // RDATA length
	RData  []byte // Raw record data
}

// ────────────────────────────────────────────────────────
// DNS Message
// ────────────────────────────────────────────────────────

// Message represents a complete DNS message.
type Message struct {
	Header      Header
	Questions   []Question
	Answers     []ResourceRecord
	Authorities []ResourceRecord
	Additional  []ResourceRecord
	Raw         []byte // Original raw bytes (for cache storage)
}

// ────────────────────────────────────────────────────────
// Parsing Functions
// ────────────────────────────────────────────────────────

var (
	ErrPacketTooShort  = errors.New("dns: packet too short for header")
	ErrMalformedName   = errors.New("dns: malformed domain name")
	ErrPacketTruncated = errors.New("dns: packet truncated")
	ErrPointerOverflow = errors.New("dns: compression pointer overflow")
)

// ParseMessage decodes a raw DNS wire-format message into a Message struct.
// It performs zero-copy where possible and uses the original packet as backing
// store for RData slices.
func ParseMessage(data []byte) (*Message, error) {
	if len(data) < dnsHeaderSize {
		return nil, ErrPacketTooShort
	}

	msg := &Message{Raw: data}

	// ── Decode header ──
	msg.Header.ID = binary.BigEndian.Uint16(data[0:2])
	msg.Header.Flags = binary.BigEndian.Uint16(data[2:4])
	msg.Header.QDCount = binary.BigEndian.Uint16(data[4:6])
	msg.Header.ANCount = binary.BigEndian.Uint16(data[6:8])
	msg.Header.NSCount = binary.BigEndian.Uint16(data[8:10])
	msg.Header.ARCount = binary.BigEndian.Uint16(data[10:12])

	offset := dnsHeaderSize

	// ── Decode questions ──
	msg.Questions = make([]Question, 0, msg.Header.QDCount)
	for i := 0; i < int(msg.Header.QDCount); i++ {
		name, newOffset, err := parseName(data, offset)
		if err != nil {
			return nil, fmt.Errorf("dns: parse question %d name: %w", i, err)
		}
		offset = newOffset

		if offset+4 > len(data) {
			return nil, ErrPacketTruncated
		}

		q := Question{
			Name:  name,
			Type:  binary.BigEndian.Uint16(data[offset : offset+2]),
			Class: binary.BigEndian.Uint16(data[offset+2 : offset+4]),
		}
		offset += 4
		msg.Questions = append(msg.Questions, q)
	}

	// ── Decode answer, authority, and additional sections ──
	sectionCounts := []int{
		int(msg.Header.ANCount),
		int(msg.Header.NSCount),
		int(msg.Header.ARCount),
	}

	for s, count := range sectionCounts {
		records := make([]ResourceRecord, 0, count)
		for i := 0; i < count; i++ {
			rr, newOffset, err := parseResourceRecord(data, offset)
			if err != nil {
				return nil, fmt.Errorf("dns: parse section %d record %d: %w", s, i, err)
			}
			offset = newOffset
			records = append(records, *rr)
		}
		switch s {
		case 0:
			msg.Answers = records
		case 1:
			msg.Authorities = records
		case 2:
			msg.Additional = records
		}
	}

	return msg, nil
}

// parseName decodes a DNS domain name starting at offset in data.
// Supports standard label sequences and compression pointers (RFC 1035 §4.1.4).
// Returns the domain name and the new offset past the name (not past pointers).
func parseName(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", 0, ErrMalformedName
	}

	var labels []string
	originalOffset := offset
	jumped := false
	jumpOffset := 0
	maxJumps := 50 // Guard against infinite pointer loops

	for {
		if offset >= len(data) {
			return "", 0, ErrMalformedName
		}

		length := data[offset]

		// End of name (zero-length label).
		if length == 0 {
			if !jumped {
				offset++
			}
			break
		}

		// Compression pointer: top two bits are 11 (0xC0).
		if length&0xC0 == 0xC0 {
			if offset+1 >= len(data) {
				return "", 0, ErrPointerOverflow
			}
			pointer := int(binary.BigEndian.Uint16(data[offset:offset+2]) & 0x3FFF)
			if pointer >= len(data) {
				return "", 0, ErrPointerOverflow
			}
			maxJumps--
			if maxJumps <= 0 {
				return "", 0, ErrMalformedName
			}
			if !jumped {
				jumpOffset = offset + 2 // Advance past the 2-byte pointer
				jumped = true
			}
			offset = pointer
			continue
		}

		// Standard label: length byte + label bytes.
		if int(length) > 63 {
			return "", 0, fmt.Errorf("dns: label length %d exceeds maximum 63", length)
		}

		offset++
		if offset+int(length) > len(data) {
			return "", 0, ErrMalformedName
		}

		labels = append(labels, string(data[offset:offset+int(length)]))
		offset += int(length)
	}

	if jumped {
		offset = jumpOffset
	}

	name := strings.Join(labels, ".")
	if len(labels) > 0 {
		name += "." // Append trailing dot for FQDN
	}

	return name, originalOffset + (offset - originalOffset), nil
}

// parseResourceRecord decodes a single DNS resource record.
func parseResourceRecord(data []byte, offset int) (*ResourceRecord, int, error) {
	name, newOffset, err := parseName(data, offset)
	if err != nil {
		return nil, 0, err
	}
	offset = newOffset

	if offset+10 > len(data) {
		return nil, 0, ErrPacketTruncated
	}

	rr := &ResourceRecord{
		Name:  name,
		Type:  binary.BigEndian.Uint16(data[offset : offset+2]),
		Class: binary.BigEndian.Uint16(data[offset+2 : offset+4]),
		TTL:   binary.BigEndian.Uint32(data[offset+4 : offset+8]),
		RDLen: binary.BigEndian.Uint16(data[offset+8 : offset+10]),
	}
	offset += 10

	if offset+int(rr.RDLen) > len(data) {
		return nil, 0, ErrPacketTruncated
	}

	// Zero-copy slice pointing into the original packet data.
	rr.RData = data[offset : offset+int(rr.RDLen)]
	offset += int(rr.RDLen)

	return rr, offset, nil
}

// ────────────────────────────────────────────────────────
// Construction Functions
// ────────────────────────────────────────────────────────

// BuildName encodes a dot-separated domain name into DNS wire format
// (sequence of length-prefixed labels terminated by a zero byte).
func BuildName(name string) []byte {
	// Remove trailing dot if present.
	name = strings.TrimSuffix(name, ".")

	if name == "" {
		return []byte{0} // Root label
	}

	var buf []byte
	labels := strings.Split(name, ".")
	for _, label := range labels {
		if len(label) > 63 {
			label = label[:63] // Truncate oversized labels
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, []byte(label)...)
	}
	buf = append(buf, 0) // Terminate with zero-length label

	return buf
}

// BuildQuestion serializes a DNS question into wire format bytes.
func BuildQuestion(q Question) []byte {
	var buf []byte
	buf = append(buf, BuildName(q.Name)...)
	b := make([]byte, 4)
	binary.BigEndian.PutUint16(b[0:2], q.Type)
	binary.BigEndian.PutUint16(b[2:4], q.Class)
	buf = append(buf, b...)
	return buf
}

// BuildResponse constructs a DNS response message from an incoming request.
// It copies the request ID, sets QR=1, and appends the provided answers.
func BuildResponse(req *Message, answers []ResourceRecord, rcode uint8) []byte {
	// Build the header.
	var buf []byte
	hdr := make([]byte, dnsHeaderSize)
	binary.BigEndian.PutUint16(hdr[0:2], req.Header.ID)

	flags := uint16(0x8000) // QR = 1 (response)
	// Preserve RD from request.
	if req.Header.RD() {
		flags |= 0x0100
	}
	// Set RA = 1 (recursion available).
	flags |= 0x0080
	// Set response code.
	flags |= uint16(rcode) & 0x000F

	binary.BigEndian.PutUint16(hdr[2:4], flags)
	binary.BigEndian.PutUint16(hdr[4:6], uint16(len(req.Questions)))    // QDCOUNT
	binary.BigEndian.PutUint16(hdr[6:8], uint16(len(answers)))          // ANCOUNT
	binary.BigEndian.PutUint16(hdr[8:10], 0)                            // NSCOUNT
	binary.BigEndian.PutUint16(hdr[10:12], 0)                           // ARCOUNT
	buf = append(buf, hdr...)

	// Append original questions.
	for _, q := range req.Questions {
		buf = append(buf, BuildQuestion(q)...)
	}

	// Append answers.
	for _, rr := range answers {
		buf = append(buf, BuildResourceRecord(rr)...)
	}

	return buf
}

// BuildResourceRecord serializes a DNS resource record into wire format.
func BuildResourceRecord(rr ResourceRecord) []byte {
	var buf []byte
	buf = append(buf, BuildName(rr.Name)...)

	fixed := make([]byte, 10)
	binary.BigEndian.PutUint16(fixed[0:2], rr.Type)
	binary.BigEndian.PutUint16(fixed[2:4], rr.Class)
	binary.BigEndian.PutUint32(fixed[4:8], rr.TTL)
	binary.BigEndian.PutUint16(fixed[8:10], uint16(len(rr.RData)))
	buf = append(buf, fixed...)
	buf = append(buf, rr.RData...)

	return buf
}

// BuildARecord is a helper that builds a type-A (IPv4) resource record.
func BuildARecord(name string, ip [4]byte, ttl uint32) ResourceRecord {
	return ResourceRecord{
		Name:  name,
		Type:  TypeA,
		Class: ClassIN,
		TTL:   ttl,
		RDLen: 4,
		RData: ip[:],
	}
}

// BuildNXDOMAINResponse constructs a NXDOMAIN (Name Error) response.
// Returns the response with rcode=3 and zero answers.
func BuildNXDOMAINResponse(req *Message) []byte {
	return BuildResponse(req, nil, RcodeNameError)
}

// BuildREFUSEDResponse constructs a REFUSED response (rcode=5).
func BuildREFUSEDResponse(req *Message) []byte {
	return BuildResponse(req, nil, RcodeRefused)
}

// ────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────

// TypeToString returns the human-readable name for a DNS record type.
func TypeToString(t uint16) string {
	switch t {
	case TypeA:
		return "A"
	case TypeNS:
		return "NS"
	case TypeCNAME:
		return "CNAME"
	case TypeSOA:
		return "SOA"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeAAAA:
		return "AAAA"
	case TypeSRV:
		return "SRV"
	case TypeANY:
		return "ANY"
	default:
		return fmt.Sprintf("TYPE%d", t)
	}
}

// RcodeToString returns the human-readable name for a DNS response code.
func RcodeToString(rcode uint8) string {
	switch rcode {
	case RcodeNoError:
		return "NOERROR"
	case RcodeFormatError:
		return "FORMERR"
	case RcodeServerFailure:
		return "SERVFAIL"
	case RcodeNameError:
		return "NXDOMAIN"
	case RcodeNotImplemented:
		return "NOTIMP"
	case RcodeRefused:
		return "REFUSED"
	default:
		return fmt.Sprintf("RCODE%d", rcode)
	}
}

// IPFromRData extracts an IPv4 address from type-A RData.
func IPFromRData(data []byte) [4]byte {
	var ip [4]byte
	if len(data) >= 4 {
		copy(ip[:], data[:4])
	}
	return ip
}