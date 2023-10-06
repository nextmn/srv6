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

type Endpoint struct {
	prefix netip.Prefix
}

func NewEndpoint(ec *config.Endpoint) (netfunc_api.NetFunc, error) {
	p, err := netip.ParsePrefix(ec.Sid)
	if err != nil {
		return nil, err
	}

	// FIXME: switch on behavior to use a New<Behavior>(prefix)
	return &Endpoint{
		prefix: p,
	}, nil
}

func (e *Endpoint) Handle(packet gopacket.Packet) error {
	return nil
}

func (e *Endpoint) NetIPPrefix() *netip.Prefix {
	return &e.prefix
}

func (e *Endpoint) Prefix() string {
	return e.prefix.String()
}
