// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"net/netip"

	"github.com/nextmn/srv6/internal/iproute2"
)

type NetFunc struct {
	prefix netip.Prefix
	stop   chan bool
}

func NewNetFunc(prefix netip.Prefix) NetFunc {
	return NetFunc{
		prefix: prefix,
		stop:   make(chan bool),
	}
}

// Handle a packet
func (n *NetFunc) Handle(packet []byte) error {
	return nil
}

// Return prefix of the NetFunc as a *netip.Prefix
func (n *NetFunc) NetIPPrefix() *netip.Prefix {
	return &n.prefix
}

// Returns prefix of the NetFunc as a String
func (n *NetFunc) Prefix() string {
	return n.prefix.String()
}

// Handle packet continuously
func (n *NetFunc) Loop(tunIface *iproute2.TunIface) error {
	// Get MTU
	mtu, err := tunIface.MTU()
	if err != nil {
		return err
	}
	// Read packets while no stop signal
	for {
		select {
		case <-n.stop:
			// Stop signal received
			return nil
		default:
			packet := make([]byte, mtu)
			if nb, err := tunIface.Read(packet); err == nil {
				go n.Handle(packet[:nb])
			}
		}
	}
}

// Start the NetFunc goroutine
func (n *NetFunc) Start(tunIface *iproute2.TunIface) {
	go n.Loop(tunIface)
}

// Stop the NetFunc goroutine
func (n *NetFunc) Stop() {
	n.stop <- true
}
