// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import (
	"fmt"
	"log"

	"github.com/songgao/water"
)

var nextmnSR *water.Interface

func goSRInit() error {
	// if at least one custom behavior, create tun
	nextmnSR, err := createTun(nextmnSRTunName)
	if err != nil {
		return err
	}
	go listenOnTun(nextmnSR)
	return nil
}

func listenOnTun(iface *water.Interface) error {
	mtu, err := getMTU(iface.Name())
	if err != nil {
		return err
	}
	for {
		packet := make([]byte, mtu)
		n, err := nextmnSR.Read(packet)
		if err != nil {
			return err
		}
		go handleSR(packet[:n])
	}
}

func handleSR(packet []byte) error {
	// DROP the packet
	fmt.Println("Received a new SR packet: dropping")
	return nil
}

func stopListenningOnTun() error {
	return nil
}

func goSRExit() error {
	if err := stopListenningOnTun(); err != nil {
		return err
	}
	removeTuns()
	return nil
}
func createTun(ifaceName string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = ifaceName
	iface, err := water.New(config)
	if err != nil {
		log.Println("Unable to allocate TUN interface:", err)
		return nil, err
	}
	if err = runIP("link", "set", "dev", iface.Name(), "up"); err != nil {
		log.Println("Unable to set", iface.Name(), "up")
		return nil, err
	}
	return iface, nil
}

func removeTun(iface *water.Interface) error {
	if iface == nil {
		return nil
	}
	if err := runIP("link", "del", iface.Name()); err != nil {
		log.Println("Unable to delete interface", iface.Name(), ":", err)
		return err
	}
	return nil
}

func removeTuns() error {
	if err := removeTun(nextmnSR); err != nil {
		return err
	}
	return nil
}
func createRoutes() {
}

func removeRoutes() {
}
