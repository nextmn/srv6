// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"net/netip"

	"github.com/google/gopacket"
	"github.com/nextmn/srv6/internal/config"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

type Headend struct {
	prefix netip.Prefix
}

func NewHeadend(ec *config.Headend) (netfunc_api.NetFunc, error) {
	p, err := netip.ParsePrefix(ec.SourceAddress)
	if err != nil {
		return nil, err
	}

	// FIXME: switch on behavior to use a New<Behavior>(prefix)
	return &Headend{
		prefix: p,
	}, nil
}

func (e *Headend) NetIPPrefix() *netip.Prefix {
	return &e.prefix
}

func (e *Headend) Handle(packet gopacket.Packet) error {
	return nil
}

func (e *Headend) Prefix() string {
	return e.prefix.String()
}
