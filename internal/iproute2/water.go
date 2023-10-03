// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package iproute2

import (
	"fmt"

	"github.com/songgao/water"
)

// TunIface
type TunIface struct {
	name  string
	iface *water.Interface
}

// Create a new TunIface
func NewTunIface(name string) *TunIface {
	return &TunIface{
		name:  name,
		iface: nil,
	}
}

func (t *TunIface) CreateAndUp() error {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = t.name
	iface, err := water.New(config)
	if err != nil {
		return fmt.Errorf("Unable to allocate TUN interface: %s", err)
	}
	t.iface = iface
	if err := runIP("link", "set", "dev", t.iface.Name(), "up"); err != nil {
		return err
	}
	return nil
}
func (t *TunIface) Delete() error {
	if t.iface == nil {
		return nil
	}
	if err := runIP("link", "del", t.iface.Name()); err != nil {
		return fmt.Errorf("Unable to delete interface %s:", t.iface.Name(), err)
	}
	return nil
}

func (t *TunIface) Name() string {
	return t.name
}
