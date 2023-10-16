// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import (
	"net"
	"testing"
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
}
