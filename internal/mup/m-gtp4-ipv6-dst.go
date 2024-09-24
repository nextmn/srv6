// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package mup

import (
	"net/netip"
)

// RFC 9433, section 6.6 (End.M.GTP4.E):
// The End.M.GTP.E SID in S has the following format:
//
//	0                                                         127
//	+-----------------------+-------+----------------+---------+
//	|  SRGW-IPv6-LOC-FUNC   |IPv4DA |Args.Mob.Session|0 Padded |
//	+-----------------------+-------+----------------+---------+
//	       128-a-b-c            a            b           c
//	Figure 9: End.M.GTP4.E SID Encoding
type MGTP4IPv6Dst struct {
	prefix         netip.Prefix // prefix in canonical form
	ipv4           [IPV4_ADDR_SIZE_BYTE]byte
	argsMobSession *ArgsMobSession
}

// NewMGTP4IPv6Dst creates a new MGTP4IPv6Dst.
func NewMGTP4IPv6Dst(prefix netip.Prefix, ipv4 [IPV4_ADDR_SIZE_BYTE]byte, a *ArgsMobSession) *MGTP4IPv6Dst {
	return &MGTP4IPv6Dst{
		prefix:         prefix.Masked(),
		ipv4:           ipv4,
		argsMobSession: a,
	}
}

// ParseMGTP4IPv6Dst parses a given byte sequence into a MGTP4IPv6Dst according to the given prefixLength.
func ParseMGTP4IPv6Dst(ipv6Addr [IPV6_ADDR_SIZE_BYTE]byte, prefixLength uint) (*MGTP4IPv6Dst, error) {
	// prefix extraction
	a := netip.AddrFrom16(ipv6Addr)
	prefix := netip.PrefixFrom(a, int(prefixLength)).Masked()

	// ipv4 extraction
	var ipv4 [IPV4_ADDR_SIZE_BYTE]byte
	if src, err := fromIPv6(ipv6Addr, prefixLength, IPV4_ADDR_SIZE_BYTE); err != nil {
		return nil, err
	} else {
		copy(ipv4[:], src[:IPV4_ADDR_SIZE_BYTE])
	}

	// argMobSession extraction
	argsMobSessionSlice, err := fromIPv6(ipv6Addr, prefixLength+IPV4_ADDR_SIZE_BIT, ARGS_MOB_SESSION_SIZE_BYTE)
	argsMobSession, err := ParseArgsMobSession(argsMobSessionSlice)
	if err != nil {
		return nil, err
	}
	return &MGTP4IPv6Dst{
		prefix:         prefix,
		ipv4:           ipv4,
		argsMobSession: argsMobSession,
	}, nil
}

// IPv4 returns the IPv4 Address encoded in the MGTP4IPv6Dst.
func (e *MGTP4IPv6Dst) IPv4() netip.Addr {
	return netip.AddrFrom4(e.ipv4)
}

// ArgsMobSession returns the ArgsMobSession encoded in the MGTP4IPv6Dst.
func (e *MGTP4IPv6Dst) ArgsMobSession() *ArgsMobSession {
	return e.argsMobSession
}

// QFI returns the QFI encoded in the MGTP4IPv6Dst's ArgsMobSession.
func (e *MGTP4IPv6Dst) QFI() uint8 {
	return e.argsMobSession.QFI()
}

// R returns the R bit encoded in the MGTP4IPv6Dst's ArgsMobSession.
func (e *MGTP4IPv6Dst) R() bool {
	return e.argsMobSession.R()
}

// U returns the U bit encoded in the MGTP4IPv6Dst's ArgsMobSession.
func (e *MGTP4IPv6Dst) U() bool {
	return e.argsMobSession.U()
}

// PDUSessionID returns the PDUSessionID for this MGTP4IPv6Dst's ArgsMobSession.
func (a *MGTP4IPv6Dst) PDUSessionID() uint32 {
	return a.argsMobSession.PDUSessionID()
}

// Prefix returns the IPv6 Prefix for this MGTP4IPv6Dst.
func (e *MGTP4IPv6Dst) Prefix() netip.Prefix {
	return e.prefix
}

// MarshalLen returns the serial length of MGTP4IPv6Dst.
func (a *MGTP4IPv6Dst) MarshalLen() int {
	return IPV6_ADDR_SIZE_BYTE
}

// Marshal returns the byte sequence generated from MGTP4IPv6Dst.
func (a *MGTP4IPv6Dst) Marshal() ([]byte, error) {
	b := make([]byte, a.MarshalLen())
	if err := a.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
// warning: no caching is done, this result will be recomputed at each call
func (a *MGTP4IPv6Dst) MarshalTo(b []byte) error {
	if len(b) < a.MarshalLen() {
		return ErrTooShortToMarshal
	}
	// init ipv6 with the prefix
	prefix := a.prefix.Addr().As16()
	copy(b, prefix[:])

	ipv4 := netip.AddrFrom4(a.ipv4).AsSlice()
	bits := a.prefix.Bits()
	if bits == -1 {
		return ErrPrefixLength
	}

	// add ipv4
	if err := appendToSlice(b, uint(bits), ipv4); err != nil {
		return err
	}
	argsMobSessionB, err := a.argsMobSession.Marshal()
	if err != nil {
		return err
	}
	// add Args-Mob-Session
	if err := appendToSlice(b, uint(bits+IPV4_ADDR_SIZE_BIT), argsMobSessionB); err != nil {
		return err
	}
	return nil
}
