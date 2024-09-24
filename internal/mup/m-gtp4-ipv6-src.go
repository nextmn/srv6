// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package mup

import (
	"encoding/binary"
	"net/netip"

	"github.com/google/gopacket/layers"
)

// RFC 9433, section 6.6 (End.M.GTP4.E):
// The IPv6 Source Address has the following format:
//
//	0                                                         127
//	+----------------------+--------+--------------------------+
//	|  Source UPF Prefix   |IPv4 SA | any bit pattern(ignored) |
//	+----------------------+--------+--------------------------+
//	         128-a-b            a                  b
//	          Figure 10: IPv6 SA Encoding for End.M.GTP4.E
//
// With NextMN implementation, we choose to deviate from the RFC
// because RFC's proposal doesn't allow to retrieve
// the IPv4 SA without knowing the prefix length,
// which may be different for 2 packets issued from 2 different headends.
//
// To allow the endpoint to be stateless, we need to know the prefix.
// We propose to encode it on the 7 last bits of the IPv6 SA.
//
// The other option would have been to directly put the IPv4 SA at the end of the IPv6 SA (bytes 12 to 15),
// but this would imply matching on /128 if the IPv4 SA is used for source routing purpose,
// and thus breaking compatibility with future new patterns.
//
// We also introduce a new field that will carry the source UDP port to be used in the newly created GTP4 packet.
//
// This field is intended to help load balancing, as specified in [TS 129.281, section 4.4.2.0]:
//
// "For the GTP-U messages described below (other than the Echo Response message, see clause 4.4.2.2), the UDP Source Port
// or the Flow Label field (see IETF RFC 6437) should be set dynamically by the sending GTP-U entity to help
// balancing the load in the transport network".
//
// Since the headend has a better view than End.M.GTP4.E on
// the origin of the flows, and can be helped by the control plane,
// it makes sense to generate the source port number on headend side,
// and to carry it during transit through SR domain.
//
// Note: even with this proposal, the remaining space (73 bits) is bigger
// than what remains for LOC+FUNC in the SID (56 bits).
//
//	0                                                                         127
//	+----------------------+--------+----------------+--------------------------+---------------+
//	|  Source UPF Prefix   |IPv4 SA | UDP Source Port| any bit pattern(ignored) | Prefix length |
//	+----------------------+--------+----------------+--------------------------+---------------+
//	        128-a-b'-c-7  a (32 bits) c (16 bits)                 b'             7 bits
//	        IPv6 SA Encoding for End.M.GTP4.E in NextMN
//
// [TS 129.281, section 4.4.2.0]: https://www.etsi.org/deliver/etsi_ts/129200_129299/129281/17.04.00_60/ts_129281v170400p.pdf#page=16
type MGTP4IPv6Src struct {
	prefix netip.Prefix // prefix in canonical form
	ipv4   [IPV4_ADDR_SIZE_BYTE]byte
	udp    [UDP_PORT_SIZE_BYTE]byte
}

// NewMGTP4IPv6Src creates a nw MGTP4IPv6Src
func NewMGTP4IPv6Src(prefix netip.Prefix, ipv4 [IPV4_ADDR_SIZE_BYTE]byte, port [UDP_PORT_SIZE_BYTE]byte) *MGTP4IPv6Src {
	return &MGTP4IPv6Src{
		prefix: prefix.Masked(),
		ipv4:   ipv4,
		udp:    port,
	}
}

// ParseMGTP4IPv6SrcNextMN parses a given IPv6 source address with NextMN bit pattern into a MGTP4IPv6Src
func ParseMGTP4IPv6SrcNextMN(addr [IPV6_ADDR_SIZE_BYTE]byte) (*MGTP4IPv6Src, error) {
	// Prefix length extraction
	prefixLen := uint(IPV6_LEN_ENCODING_MASK & (addr[IPV6_LEN_ENCODING_POS_BYTE] >> IPV6_LEN_ENCODING_POS_BIT))
	if prefixLen == 0 {
		// even if globally routable IPv6 Prefix size cannot currently be less than 32 (per ICANN policy),
		// nothing prevent the use of such prefix with ULA (fc00::/7)
		// or, in the future, a prefix from a currently not yet allocated address block.
		return nil, ErrPrefixLength
	}
	if prefixLen+IPV4_ADDR_SIZE_BIT+UDP_PORT_SIZE_BIT+IPV6_LEN_ENCODING_SIZE_BIT > IPV6_ADDR_SIZE_BIT {
		return nil, ErrOutOfRange
	}

	// udp port extraction
	var udp [UDP_PORT_SIZE_BYTE]byte
	if src, err := fromIPv6(addr, prefixLen+IPV4_ADDR_SIZE_BIT, UDP_PORT_SIZE_BYTE); err != nil {
		return nil, err
	} else {
		copy(udp[:], src[:UDP_PORT_SIZE_BYTE])
	}

	if r, err := ParseMGTP4IPv6Src(addr, prefixLen); err != nil {
		return nil, err
	} else {
		r.udp = udp
		return r, nil
	}
}

// ParseMGTP4IPv6SrcNextMN parses a given IPv6 source address without any specific bit pattern into a MGTP4IPv6Src
func ParseMGTP4IPv6Src(addr [IPV6_ADDR_SIZE_BYTE]byte, prefixLen uint) (*MGTP4IPv6Src, error) {
	// prefix extraction
	a := netip.AddrFrom16(addr)
	prefix := netip.PrefixFrom(a, int(prefixLen)).Masked()

	// ipv4 extraction
	var ipv4 [IPV4_ADDR_SIZE_BYTE]byte
	if src, err := fromIPv6(addr, prefixLen, IPV4_ADDR_SIZE_BYTE); err != nil {
		return nil, err
	} else {
		copy(ipv4[:], src[:IPV4_ADDR_SIZE_BYTE])
	}

	return &MGTP4IPv6Src{
		prefix: prefix,
		ipv4:   ipv4,
	}, nil
}

// IPv4 returns the IPv4 Address encoded in the MGTP4IPv6Src.
func (e *MGTP4IPv6Src) IPv4() netip.Addr {
	return netip.AddrFrom4(e.ipv4)
}

// UDPPortNumber returns the UDP Port Number encoded in the MGTP4IPv6Src (0 if not set).
func (e *MGTP4IPv6Src) UDPPortNumber() layers.UDPPort {
	return (layers.UDPPort)(binary.BigEndian.Uint16([]byte{e.udp[0], e.udp[1]}))
}

// MarshalLen returns the serial length of MGTP4IPv6Src.
func (a *MGTP4IPv6Src) MarshalLen() int {
	return IPV6_ADDR_SIZE_BYTE
}

// Marshal returns the byte sequence generated from MGTP4IPv6Src.
func (a *MGTP4IPv6Src) Marshal() ([]byte, error) {
	b := make([]byte, a.MarshalLen())
	if err := a.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
// warning: no caching is done, this result will be recomputed at each call
func (a *MGTP4IPv6Src) MarshalTo(b []byte) error {
	if len(b) < a.MarshalLen() {
		return ErrTooShortToMarshal
	}
	// init b with prefix
	prefix := a.prefix.Addr().As16()
	copy(b, prefix[:])

	ipv4 := netip.AddrFrom4(a.ipv4).AsSlice()
	udp := []byte{a.udp[0], a.udp[1]}
	bits := a.prefix.Bits()
	if bits == -1 {
		return ErrPrefixLength
	}

	// add ipv4
	if err := appendToSlice(b, uint(bits), ipv4); err != nil {
		return err
	}
	// add upd port
	if err := appendToSlice(b, uint(bits+IPV4_ADDR_SIZE_BIT), udp); err != nil {
		return err
	}
	// add prefix length
	b[IPV6_LEN_ENCODING_POS_BYTE] = byte(bits)
	return nil
}
