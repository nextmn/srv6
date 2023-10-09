// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"net/netip"

	"github.com/nextmn/srv6/internal/config"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

type Headend struct {
	NetFunc
}

func NewHeadend(ec *config.Headend) (netfunc_api.NetFunc, error) {
	p, err := netip.ParsePrefix(ec.To)
	if err != nil {
		return nil, err
	}

	// FIXME: switch on behavior to use a New<Behavior>(prefix)
	return &Headend{
		NetFunc: NewNetFunc(p),
	}, nil
}

func (e *Headend) Handle(packet []byte) error {
	return nil
}
