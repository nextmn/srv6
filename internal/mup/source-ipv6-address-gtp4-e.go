// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"fmt"
	"net/netip"
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

type SourceIPv6AddressGTP4E struct {
	ipv4 netip.Addr
	udp  []byte
}

func NewSourceIPv6AddressGTP4E(addr netip.Addr) (*SourceIPv6AddressGTP4E, error) {
	if !addr.Is6() {
		return nil, fmt.Errorf("addr must be IPv6 address")
	}
	addrSlice := addr.AsSlice()

	// Prefix length extraction
	prefixLenSlice, err := fromSlice(addrSlice, 128-7, 1)
	if err != nil {
		return nil, err
	}
	prefixLen := int(prefixLenSlice[0])

	if prefixLen == 0 {
		return nil, fmt.Errorf("unknown prefix length")
	}
	if prefixLen+32+16+7 > 128 {
		return nil, fmt.Errorf("erroneous prefix length")
	}

	// ipv4 extraction
	ipv4Slice, err := fromSlice(addrSlice, prefixLen, 4)
	if err != nil {
		return nil, err
	}
	ipv4, ok := netip.AddrFromSlice(ipv4Slice)
	if !ok {
		return nil, fmt.Errorf("could not create ipv4 from slice")
	}

	// udp port extraction
	udp, err := fromSlice(addrSlice, prefixLen+32, 2)
	if err != nil {
		return nil, err
	}

	return &SourceIPv6AddressGTP4E{
		ipv4: ipv4,
		udp:  udp,
	}, nil
}

func (s *SourceIPv6AddressGTP4E) IPv4() netip.Addr {
	return s.ipv4
}

func (s *SourceIPv6AddressGTP4E) UDPPortNumber() []byte {
	return s.udp
}
