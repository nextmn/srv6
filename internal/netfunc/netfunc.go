// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"context"
	"log"

	"github.com/nextmn/srv6/internal/iproute2"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

type NetFunc struct {
	debug   bool
	handler netfunc_api.Handler
}

func NewNetFunc(handler netfunc_api.Handler, debug bool) *NetFunc {
	return &NetFunc{
		debug:   debug,
		handler: handler,
	}
}

func (n NetFunc) Debug() bool {
	return n.debug
}

// Run the NetFunc goroutine
func (n *NetFunc) Run(ctx context.Context, tunIface *iproute2.TunIface) error {
	// Get MTU
	mtu, err := tunIface.MTU()
	if err != nil {
		return err
	}
	// Read packets while no stop signal
	for {
		select {
		case <-ctx.Done():
			// Stop signal received
			return nil
		default:
			packet := make([]byte, mtu)
			if nb, err := tunIface.Read(packet); err == nil {
				go func(ctx context.Context, iface *iproute2.TunIface) {
					if out, err := n.handler.Handle(ctx, packet[:nb]); err == nil {
						iface.Write(out)
					} else if n.Debug() {
						log.Println(err)
					}
				}(ctx, tunIface)
			}
		}
	}
}
