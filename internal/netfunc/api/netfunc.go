// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"net/netip"

	"github.com/google/gopacket"
)

type NetFunc interface {
	Handle(packet gopacket.Packet) error
	Prefix() string
	NetIPPrefix() *netip.Prefix
}
