// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"fmt"
	"net/netip"
)

type SidGTP4 struct {
	locFunc        netip.Prefix
	ipv4           netip.Addr
	argsMobSession *ArgsMobSession
}

func NewSidGTP4(sid netip.Addr, prefixLength int) (*SidGTP4, error) {
	if !sid.Is6() {
		return nil, fmt.Errorf("SID must be IPv6 address")
	}
	// locFunc extraction
	locFunc, err := sid.Prefix(prefixLength)
	if err != nil {
		return nil, err
	}

	sidSlice := sid.AsSlice()

	// ipv4 extraction
	ipv4Slice, err := fromSlice(sidSlice, prefixLength, 4)
	if err != nil {
		return nil, err
	}
	ipv4, ok := netip.AddrFromSlice(ipv4Slice)
	if !ok {
		return nil, fmt.Errorf("Could not create ipv4 from slice")
	}

	// argMobSession extraction
	argsMobSessionSlice, err := fromSlice(sidSlice, prefixLength+32, 5)
	argsMobSession, err := ParseArgsMobSession(argsMobSessionSlice)
	if err != nil {
		return nil, err
	}
	return &SidGTP4{
		locFunc:        locFunc,
		ipv4:           ipv4,
		argsMobSession: argsMobSession,
	}, nil
}

func (s *SidGTP4) IPv4() netip.Addr {
	return s.ipv4
}

func (s *SidGTP4) ArgsMobSession() *ArgsMobSession {
	return s.argsMobSession
}

func (s *SidGTP4) LocFunc() netip.Prefix {
	return s.locFunc
}
