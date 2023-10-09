// // Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// // Use of this source code is governed by a MIT-style license that can be
// // found in the LICENSE file.
// // SPDX-License-Identifier: MIT
package iproute2

//
//import (
//	"fmt"
//	"net/netip"
//
//	"github.com/google/gopacket"
//	netfunc "github.com/nextmn/srv6/internal/netfunc/api"
//	"github.com/songgao/water"
//)
//
//type Dispatcher struct {
//	reload   chan bool
//	stop     chan bool
//	prefixes []*netip.Prefix
//	netfuncs []netfunc.NetFunc
//}
//
//// Create a new Dispatcher
//func NewDispatcher() Dispatcher {
//	return Dispatcher{
//		reload:   make(chan bool),
//		stop:     make(chan bool),
//		prefixes: make([]*netip.Prefix, 0),
//		netfuncs: make([]netfunc.NetFunc, 0),
//	}
//}
//
//// (re)loads netfunc list
//func (d *Dispatcher) loadNetFuncList(netfuncs []netfunc.NetFunc) {
//	// we store prefixes and netfuncs in 2 lists
//	// with shared indexes
//	p := make([]*netip.Prefix, 0)
//	n := make([]netfunc.NetFunc, 0)
//	for _, nf := range d.netfuncs {
//		p = append(p, nf.NetIPPrefix())
//		n = append(n, nf)
//	}
//	d.prefixes = p
//	d.netfuncs = n
//}
//
//// Starts the dispatcher
//func (d *Dispatcher) Start(tunIface *TunIface, iface *water.Interface, mtu int64) {
//	go d.dispatch(tunIface, iface, mtu)
//}
//
//// Classify
//func (d *Dispatcher) classify(packet []byte) error {
//	layerType, err := networkLayerType(packet)
//	if err != nil {
//		return err
//	}
//	pqt := gopacket.NewPacket(packet, layerType, gopacket.Default)
//	addrSlice := pqt.NetworkLayer().NetworkFlow().Dst().Raw()
//	addr, ok := netip.AddrFromSlice(addrSlice)
//	if !ok {
//		return fmt.Errorf("Malformed address")
//	}
//	// XXX: this is not very efficient,
//	// but should do the job if there is not too many netfunc per tun iface.
//	// An improvement would be to create a TUN iface per netfunc
//	// and let Linux routing do all the work
//	// thus, removing the need for the loop & prefix check.
//	for i, prefix := range d.prefixes {
//		if prefix.Contains(addr) {
//			d.netfuncs[i].Handle(pqt)
//		}
//	}
//	return nil
//}
//
//// dispatch packets until stopped
//func (d *Dispatcher) dispatch(tunIface *TunIface, iface *water.Interface, mtu int64) error {
//	for {
//		select {
//		case <-d.stop:
//			return nil
//		case <-d.reload:
//			d.loadNetFuncList(maps.Values(tunIface.netfuncs.Values))
//		default:
//			packet := make([]byte, mtu)
//			if n, err := iface.Read(packet); err == nil {
//				go d.classify(packet[:n])
//			}
//		}
//	}
//}
//
//// Force the Dispatcher to reload list of Netfuncs
//func (d *Dispatcher) Reload() {
//	d.reload <- true
//}
//
//// Stop Dispatcher's goroutine
//func (d *Dispatcher) Stop() {
//	d.stop <- true
//}
