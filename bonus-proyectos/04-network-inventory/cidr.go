package main

import (
	"fmt"
	"net/netip"
)

// ExpandCIDR parses a CIDR notation string (e.g., "192.168.1.0/24") and returns
// a slice of all individual host IP addresses within that range. It uses the
// modern net/netip package (Go 1.18+) for safe, value-type IP manipulation.
//
// The function excludes the network address and broadcast address for subnets
// larger than /31. For /32 it returns exactly one address, and for /31 it returns
// both addresses (as per RFC 3021 point-to-point links).
func ExpandCIDR(cidr string) ([]netip.Addr, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR notation %q: %w", cidr, err)
	}

	// Canonicalize the prefix (mask host bits).
	prefix = prefix.Masked()

	bits := prefix.Bits()
	addr := prefix.Addr()

	// Validate that we're working with an IPv4 address.
	if !addr.Is4() {
		return nil, fmt.Errorf("only IPv4 CIDR ranges are supported, got %s", addr.String())
	}

	// For /32: single host, return immediately.
	if bits == 32 {
		return []netip.Addr{addr}, nil
	}

	// Calculate total addresses in the range.
	// For IPv4, bits ranges from 0 to 32.
	hostBits := 32 - bits
	totalAddresses := 1 << hostBits

	// Safety cap to prevent accidental memory exhaustion (e.g., /0 = 4 billion IPs).
	const maxAddresses = 1 << 20 // 1,048,576 addresses (up to a /12)
	if totalAddresses > maxAddresses {
		return nil, fmt.Errorf("CIDR range %s expands to %d addresses (max %d); use a more specific prefix",
			cidr, totalAddresses, maxAddresses)
	}

	// Convert the base address to a uint32 for iteration.
	baseInt := addrToUint32(addr)

	// Determine start and end offsets.
	// For /31 (point-to-point) include both addresses.
	// For larger subnets, skip network (first) and broadcast (last).
	start := 0
	end := totalAddresses
	if bits < 31 {
		start = 1                // skip network address
		end = totalAddresses - 1 // skip broadcast address
	}

	ips := make([]netip.Addr, 0, end-start)
	for i := start; i < end; i++ {
		ip := uint32ToAddr(baseInt + uint32(i))
		ips = append(ips, ip)
	}

	return ips, nil
}

// addrToUint32 converts a netip.Addr (IPv4) to its uint32 representation.
func addrToUint32(addr netip.Addr) uint32 {
	b := addr.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

// uint32ToAddr converts a uint32 back to a netip.Addr (IPv4).
func uint32ToAddr(n uint32) netip.Addr {
	return netip.AddrFrom4([4]byte{
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	})
}

// ValidateCIDR checks if a string is a valid CIDR notation without expanding it.
func ValidateCIDR(cidr string) error {
	_, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR notation %q: %w", cidr, err)
	}
	return nil
}
