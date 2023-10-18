// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"fmt"
	"net"
	"net/netip"
)

type MGTP4IPv6DstFields struct {
	prefix         netip.Prefix // prefix in canonical form
	ipv4           [4]byte
	argsMobSession *ArgsMobSession
}

func NewMGTP4IPv6DstFieldsFromFields(prefix netip.Prefix, ipv4 []byte, a *ArgsMobSession) (*MGTP4IPv6DstFields, error) {
	if len(ipv4) != 4 {
		return nil, fmt.Errorf("Not a IPv4 Address")
	}
	return &MGTP4IPv6DstFields{
		prefix:         prefix.Masked(),
		ipv4:           [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
		argsMobSession: a,
	}, nil
}

func NewMGTP4IPv6DstFields(ipv6Addr []byte, prefixLength uint) (*MGTP4IPv6DstFields, error) {
	if len(ipv6Addr) != IPV6_ADDR_SIZE_BYTE {
		return nil, ErrNotAnIPv6Address
	}

	// prefix extraction
	a, ok := netip.AddrFromSlice(ipv6Addr)
	if !ok {
		return nil, ErrNotAnIPv6Address
	}
	prefix := netip.PrefixFrom(a, int(prefixLength)).Masked()

	// ipv4 extraction
	var ipv4 [IPV4_ADDR_SIZE_BYTE]byte
	if src, err := fromSlice(ipv6Addr, prefixLength, IPV4_ADDR_SIZE_BYTE); err != nil {
		return nil, err
	} else {
		copy(ipv4[:], src[:IPV4_ADDR_SIZE_BYTE])
	}

	// argMobSession extraction
	argsMobSessionSlice, err := fromSlice(ipv6Addr, prefixLength+IPV4_ADDR_SIZE_BIT, ARGS_MOB_SESSION_SIZE_BYTE)
	argsMobSession, err := ParseArgsMobSession(argsMobSessionSlice)
	if err != nil {
		return nil, err
	}
	return &MGTP4IPv6DstFields{
		prefix:         prefix,
		ipv4:           ipv4,
		argsMobSession: argsMobSession,
	}, nil
}

func (e *MGTP4IPv6DstFields) IPv4() net.IP {
	return net.IPv4(e.ipv4[0], e.ipv4[1], e.ipv4[2], e.ipv4[3])
}

func (e *MGTP4IPv6DstFields) ArgsMobSession() *ArgsMobSession {
	return e.argsMobSession
}

func (e *MGTP4IPv6DstFields) QFI() uint8 {
	return e.argsMobSession.QFI()
}
func (e *MGTP4IPv6DstFields) R() bool {
	return e.argsMobSession.R()
}
func (e *MGTP4IPv6DstFields) U() bool {
	return e.argsMobSession.U()
}

func (e *MGTP4IPv6DstFields) Prefix() netip.Prefix {
	return e.prefix
}

func (a *MGTP4IPv6DstFields) PDUSessionID() uint32 {
	return a.argsMobSession.PDUSessionID()
}

func (a *MGTP4IPv6DstFields) MarshalLen() int {
	return IPV6_ADDR_SIZE_BYTE
}

// Marshal returns the byte sequence generated from MGTP4IPv6DstFields.
func (a *MGTP4IPv6DstFields) Marshal() ([]byte, error) {
	b := make([]byte, a.MarshalLen())
	if err := a.MarshalTo(b); err != nil {
		return nil, err
	}
	return b, nil
}

// MarshalTo puts the byte sequence in the byte array given as b.
// warning: no caching is done, this result will be recomputed at each call
func (a *MGTP4IPv6DstFields) MarshalTo(b []byte) error {
	if len(b) < a.MarshalLen() {
		return ErrTooShortToMarshal
	}
	// init ipv6 with the prefix
	prefix := a.prefix.Addr().As16()
	copy(b, prefix[:])

	ipv4 := netip.AddrFrom4(a.ipv4).AsSlice()
	bits := a.prefix.Bits()
	if bits == -1 {
		return fmt.Errorf("Error with prefix length")
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
