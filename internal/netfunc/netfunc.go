// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"github.com/nextmn/srv6/internal/iproute2"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

type NetFunc struct {
	stop    chan bool
	handler netfunc_api.Handler
}

func NewNetFunc(handler netfunc_api.Handler) *NetFunc {
	return &NetFunc{
		stop:    make(chan bool),
		handler: handler,
	}
}

// Handle packet continuously
func (n NetFunc) loop(tunIface *iproute2.TunIface) error {
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
				go func(iface *iproute2.TunIface) {
					if out, err := n.handler.Handle(packet[:nb]); err == nil {
						iface.Write(out)
					}
				}(tunIface)
			}
		}
	}
}

// Start the NetFunc goroutine
func (n *NetFunc) Start(tunIface *iproute2.TunIface) {
	go n.loop(tunIface)
}

// Stop the NetFunc goroutine
func (n *NetFunc) Stop() {
	n.stop <- true
}
