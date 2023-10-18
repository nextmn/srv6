// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"fmt"
	"net"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMGTP4IPv6SrcFields(t *testing.T) {
	ip_addr := []byte{
		0x20, 0x01, 0xDB, 0x08,
		192, 0, 2, 1,
		0x01, 0x23,
		0x55, 0x55, 0x55, 0x55, 0x55,
		32,
	}

	e, err := NewMGTP4IPv6SrcFields(ip_addr)
	if err != nil {
		t.Fatal(err)
	}
	if !e.IPv4().Equal(net.ParseIP("192.0.2.1")) {
		t.Fatalf("Cannot extract ipv4 correctly: %s", e.IPv4())
	}
	if e.UDPPortNumber() != 0x0123 {
		t.Fatalf("Cannot extract udp port number correctly: %x", e.UDPPortNumber())
	}
	ip_addr2, err := NewMGTP4IPv6SrcFieldsFromFields(netip.MustParsePrefix("fd00:1:1::/48"), []byte{10, 0, 4, 1}, []byte{0x12, 0x34})
	if err != nil {
		t.Fatal(err)
	}
	b, err := ip_addr2.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	res2 := []byte{
		0xfd, 0x00, 0x00, 0x01, 0x00, 0x01,
		10, 0, 4, 1,
		0x12, 0x34,
		0x00, 0x00, 0x00,
		48,
	}
	fmt.Println(b)
	fmt.Println(res2)
	if diff := cmp.Diff(b, res2); diff != "" {
		t.Error(diff)
	}

}
