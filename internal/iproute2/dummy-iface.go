// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package iproute2

// IPRoute2 Dummy interface
type DummyIface struct {
	name string
}

// Create a new DummyInterface
func NewDummyIface(name string) *DummyIface {
	return &DummyIface{name: name}
}

// Create iproute2 dummy interface
func (iface DummyIface) create() error {
	if err := runIP("link", "add", iface.name, "type", "dummy"); err != nil {
		return err
	}
	return nil
}

// Set iproute2 dummy interface link up
func (iface DummyIface) up() error {
	if err := runIP("link", "set", iface.name, "up"); err != nil {
		return err
	}
	return nil
}

// Create iproute2 dummy interface and set link up
func (iface DummyIface) CreateAndUp() error {
	if err := iface.create(); err != nil {
		return err
	}
	if err := iface.up(); err != nil {
		return err
	}
	return nil
}

// Delete iproute2 dummy interface
func (iface DummyIface) Delete() error {
	if err := runIP("link", "del", iface.Name()); err != nil {
		return err
	}
	return nil
}

// Returns name of the iface
func (iface DummyIface) Name() string {
	return iface.name
}
