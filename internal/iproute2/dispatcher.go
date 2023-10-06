// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package iproute2

import (
	"fmt"
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	netfunc "github.com/nextmn/srv6/internal/netfunc/api"
	"github.com/songgao/water"
)

type Dispatcher struct {
	reload   chan bool
	stop     chan bool
	prefixes []*netip.Prefix
	netfuncs []netfunc.NetFunc
}

// Create a new Dispatcher
func NewDispatcher() Dispatcher {
	return Dispatcher{
		reload:   make(chan bool),
		stop:     make(chan bool),
		prefixes: make([]*netip.Prefix, 0),
		netfuncs: make([]netfunc.NetFunc, 0),
	}
}

func (d *Dispatcher) loadNetFuncList(netfuncs *netfunc.NetFunc) {
	p := make([]*netip.Prefix, 0)
	n := make([]netfunc.NetFunc, 0)
	for _, nf := range d.netfuncs {
		p = append(p, nf.NetIPPrefix())
		n = append(n, nf)
	}
	d.prefixes = p
	d.netfuncs = n
}

func (d *Dispatcher) Start(iface *water.Interface, mtu int64) {
	go d.dispatch(iface, mtu)
}

func (d *Dispatcher) classify(packet []byte) error {
	// addr := … //TODO:
	var addrSlice []byte
	var layerType gopacket.LayerType
	switch (packet[0] >> 4) & 0x0F {
	case 4:
		layerType = layers.LayerTypeIPv4
	case 6:
		layerType = layers.LayerTypeIPv6
	default:
		return fmt.Errorf("Malformed packet")

	}
	pqt := gopacket.NewPacket(packet, layerType, gopacket.Default)
	dst := pqt.NetworkLayer().NetworkFlow().Dst // FIXME
	addr, ok := netip.AddrFromSlice(addrSlice)
	if !ok {
		return fmt.Errorf("Malformed address")
	}
	for i, prefix := range d.prefixes {
		if prefix.Contains(addr) {
			d.netfuncs[i].Handle(pqt) // TODO: packet should be a gopacket
		}
	}
	return nil
}

func (d *Dispatcher) dispatch(iface *water.Interface, mtu int64) error {
	for {
		select {
		case <-d.stop:
			return nil
		default:
			packet := make([]byte, mtu)
			if n, err := iface.Read(packet); err == nil {
				go d.classify(packet[:n])
			}
		}
	}
}

// Force the Dispatcher to reload list of Netfuncs
func (d *Dispatcher) Reload() {
	d.reload <- true
}

// Stop Dispatcher's goroutine
func (d *Dispatcher) Stop() {
	d.stop <- true
}
