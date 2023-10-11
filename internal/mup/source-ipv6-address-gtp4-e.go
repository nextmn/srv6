// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"fmt"
	"net/netip"
)

type SourceIPv6AddressGTP4E struct {
	sourceUPFPrefix netip.Prefix
	ipv4            netip.Addr
}

func NewSourceIPv6AddressGTP4E(addr netip.Addr, prefixLength int) (*SourceIPv6AddressGTP4E, error) {
	if !addr.Is6() {
		return nil, fmt.Errorf("addr must be IPv6 address")
	}
	// sourceUPFPrefix extraction
	sourceUPFPrefix, err := addr.Prefix(prefixLength)
	if err != nil {
		return nil, err
	}

	sidSlice := addr.AsSlice()

	// ipv4 extraction
	ipv4Slice, err := fromSlice(sidSlice, prefixLength, 4)
	if err != nil {
		return nil, err
	}
	ipv4, ok := netip.AddrFromSlice(ipv4Slice)
	if !ok {
		return nil, fmt.Errorf("Could not create ipv4 from slice")
	}

	return &SourceIPv6AddressGTP4E{
		sourceUPFPrefix: sourceUPFPrefix,
		ipv4:            ipv4,
	}, nil
}

func (s *SourceIPv6AddressGTP4E) IPv4() netip.Addr {
	return s.ipv4
}

func (s *SourceIPv6AddressGTP4E) SourceUPFPrefix() netip.Prefix {
	return s.sourceUPFPrefix
}
