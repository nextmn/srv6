// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"encoding/binary"
	"net"

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

type EndMGTP4EIPv6SrcFields struct {
	ipv4 [IPV4_ADDR_SIZE_BYTE]byte
	udp  [UDP_PORT_SIZE_BYTE]byte
}

func NewEndMGTP4EIPv6SrcFields(addr []byte) (*EndMGTP4EIPv6SrcFields, error) {
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

	return &EndMGTP4EIPv6SrcFields{
		ipv4: ipv4,
		udp:  udp,
	}, nil
}

func (e *EndMGTP4EIPv6SrcFields) IPv4() net.IP {
	return net.IPv4(e.ipv4[0], e.ipv4[1], e.ipv4[2], e.ipv4[3])
}

func (e *EndMGTP4EIPv6SrcFields) UDPPortNumber() layers.UDPPort {
	return (layers.UDPPort)(binary.BigEndian.Uint16([]byte{e.udp[0], e.udp[1]}))
}
