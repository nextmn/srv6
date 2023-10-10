// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"net/netip"

	"github.com/nextmn/srv6/internal/iproute2"
)

type NetFunc interface {
	Handle(packet []byte) error
	Prefix() string
	NetIPPrefix() *netip.Prefix
	Loop(tunIface *iproute2.TunIface) error
	Start(tunIface *iproute2.TunIface)
	Stop()
}