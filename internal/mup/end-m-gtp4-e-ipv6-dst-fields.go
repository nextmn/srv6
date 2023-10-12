// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import "net"

type EndMGTP4EIPv6DstFields struct {
	ipv4           [4]byte
	argsMobSession *ArgsMobSession
}

func NewEndMGTP4EIPv6DstFields(ipv6Addr []byte, prefixLength uint) (*EndMGTP4EIPv6DstFields, error) {
	if len(ipv6Addr) != IPV6_ADDR_SIZE_BYTE {
		return nil, ErrNotAnIPv6Address
	}

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
	return &EndMGTP4EIPv6DstFields{
		ipv4:           ipv4,
		argsMobSession: argsMobSession,
	}, nil
}

func (e *EndMGTP4EIPv6DstFields) IPv4() net.IP {
	return net.IPv4(e.ipv4[0], e.ipv4[1], e.ipv4[2], e.ipv4[3])
}

func (e *EndMGTP4EIPv6DstFields) ArgsMobSession() *ArgsMobSession {
	return e.argsMobSession
}

func (e *EndMGTP4EIPv6DstFields) QFI() uint8 {
	return e.argsMobSession.QFI()
}
func (e *EndMGTP4EIPv6DstFields) R() bool {
	return e.argsMobSession.R()
}
func (e *EndMGTP4EIPv6DstFields) U() bool {
	return e.argsMobSession.U()
}

func (a *EndMGTP4EIPv6DstFields) PDUSessionID() uint32 {
	return a.argsMobSession.PDUSessionID()
}
