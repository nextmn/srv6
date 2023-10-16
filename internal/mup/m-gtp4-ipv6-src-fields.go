// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"

	"github.com/google/gopacket/layers"
)

//-------------------------------------------------------
// RFC 9433: 6.6 End.M.GTP4.E
// > *  The IPv6 Source Address has the following format:
// >
// >       0                                                         127
// >       +----------------------+--------+--------------------------+
// >       |  Source UPF Prefix   |IPv4 SA | any bit pattern(ignored) |
// >       +----------------------+--------+--------------------------+
// >                128-a-b            a                  b
// >
// >                IPv6 SA Encoding for End.M.GTP4.E
//
//-------------------------------------------------------
// With NextMN implementation, we choose to deviate from the RFC
// because RFC's proposal doesn't allow to retrieve
// the IPv4 SA without knowing the prefix length,
// which may be different for 2 packets issued from 2 different headends.
//
// To allow the endpoint to be stateless, we need to know the prefix.
// We propose to encode it on the 7 last bits of the IPv6 SA.
//
// The other option would have be to directly put the IPv4 SA at the end of the IPv6 SA,
// but this would imply matching on /128 if the IPv4 SA is used for source routing purpose,
// and thus breaking compatibility with future new patterns.
//
// We also introcuce a new field that will carry the source UDP port to be used in the newly created GTP4 packet.
// This field is intended to help load balancing, as specified in TS 129.281 section 4.4.2.0:
// > For the GTP-U messages described below (other than the Echo Response message, see clause 4.4.2.2), the UDP Source
// > Port or the Flow Label field (see IETF RFC 6437 [37]) should be set dynamically by the sending GTP-U entity to help
// > balancing the load in the transport network
//
// Since the headend has a better view than End.M.GTP4.E on
// the origin of the flows, and can be helped by the control plane,
// it makes sense to generate the source port number on headend side,
// and to carry it during transit through SR domain.
//
// Note: even with this proposal, the remaining space (73 bits) is bigger
// than what remains for LOC+FUNC in the SID (56 bits).
//
//
//        0                                                                         127
//        +----------------------+--------+----------------+--------------------------+---------------+
//        |  Source UPF Prefix   |IPv4 SA | UDP Source Port| any bit pattern(ignored) | Prefix length |
//        +----------------------+--------+----------------+--------------------------+---------------+
//                 128-a-b'-c-7  a (32 bits) c (16 bits)                 b'             7 bits
//
//                 IPv6 SA Encoding for End.M.GTP4.E in NextMN

type MGTP4IPv6SrcFields struct {
	prefix netip.Prefix // prefix in canonical form
	ipv4   [IPV4_ADDR_SIZE_BYTE]byte
	udp    [UDP_PORT_SIZE_BYTE]byte
}

func NewMGTP4IPv6SrcFieldsFromFields(prefix netip.Prefix, ipv4 []byte, udp []byte) (*MGTP4IPv6SrcFields, error) {
	if len(ipv4) != 4 {
		return nil, fmt.Errorf("Not a IPv4 Address")
	}
	return &MGTP4IPv6SrcFields{
		prefix: prefix.Masked(),
		ipv4:   [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
		udp:    [2]byte{udp[0], udp[1]},
	}, nil
}

func NewMGTP4IPv6SrcFields(addr []byte) (*MGTP4IPv6SrcFields, error) {
	if len(addr) != IPV6_ADDR_SIZE_BYTE {
		return nil, ErrNotAnIPv6Address
	}

	// Prefix length extraction
	prefixLen := uint(IPV6_LEN_ENCODING_MASK & (addr[IPV6_LEN_ENCODING_POS_BYTE] >> IPV6_LEN_ENCODING_POS_BIT))
	if prefixLen == 0 {
		// even if globally routable IPv6 Prefix size cannot currently be less than 32 (per ICANN policy),
		// nothing prevent the use of such prefix with ULA (fc00::/7)
		// or, in the future, a prefix from a currently not yet allocated address block.
		return nil, ErrWrongValue
	}
	if prefixLen+IPV4_ADDR_SIZE_BIT+UDP_PORT_SIZE_BIT+IPV6_LEN_ENCODING_SIZE_BIT > IPV6_ADDR_SIZE_BIT {
		return nil, ErrWrongValue
	}

	// prefix extraction
	a, ok := netip.AddrFromSlice(addr)
	if !ok {
		return nil, ErrNotAnIPv6Address
	}
	prefix := netip.PrefixFrom(a, int(prefixLen)).Masked()

	// ipv4 extraction
	var ipv4 [IPV4_ADDR_SIZE_BYTE]byte
	if src, err := fromSlice(addr, prefixLen, IPV4_ADDR_SIZE_BYTE); err != nil {
		return nil, err
	} else {
		copy(ipv4[:], src[:IPV4_ADDR_SIZE_BYTE])
	}

	// udp port extraction
	var udp [UDP_PORT_SIZE_BYTE]byte
	if src, err := fromSlice(addr, prefixLen+IPV4_ADDR_SIZE_BIT, UDP_PORT_SIZE_BYTE); err != nil {
		return nil, err
	} else {
		copy(udp[:], src[:UDP_PORT_SIZE_BYTE])
	}

	return &MGTP4IPv6SrcFields{
		prefix: prefix,
		ipv4:   ipv4,
		udp:    udp,
	}, nil
}

func (e *MGTP4IPv6SrcFields) IPv4() net.IP {
	return net.IPv4(e.ipv4[0], e.ipv4[1], e.ipv4[2], e.ipv4[3])
}

func (e *MGTP4IPv6SrcFields) UDPPortNumber() layers.UDPPort {
	return (layers.UDPPort)(binary.BigEndian.Uint16([]byte{e.udp[0], e.udp[1]}))
}

func (a *MGTP4IPv6SrcFields) MarshalLen() int {
	return IPV6_ADDR_SIZE_BYTE
}

// Marshal returns the byte sequence generated from MGTP4IPv6SrcFields.
func (a *MGTP4IPv6SrcFields) Marshal() ([]byte, error) {
	b := make([]byte, a.MarshalLen())
	if err := a.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
// warning: no caching is done, this result will be recomputed at each call
func (a *MGTP4IPv6SrcFields) MarshalTo(b []byte) error {
	if len(b) < a.MarshalLen() {
		return ErrTooShortToMarshal
	}
	// init ipv6 with prefix
	ipv6 := a.prefix.Addr().AsSlice()

	ipv4 := netip.AddrFrom4(a.ipv4).AsSlice()
	udp := []byte{a.udp[0], a.udp[1]}
	bits := a.prefix.Bits()
	if bits == -1 {
		return fmt.Errorf("Error with prefix length")
	}

	// add ipv4
	if err := appendToSlice(ipv6, uint(bits), ipv4); err != nil {
		return err
	}
	// add upd port
	if err := appendToSlice(ipv6, uint(bits+IPV4_ADDR_SIZE_BYTE), udp); err != nil {
		return err
	}
	// add prefix length
	b[IPV6_LEN_ENCODING_POS_BYTE] = byte(bits)
	return nil
}
