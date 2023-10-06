// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package iproute2

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/netfunc"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
	"github.com/songgao/water"
)

// TunIface
type TunIface struct {
	dispatcher Dispatcher
	queue      chan []byte
	name       string
	iface      *water.Interface
	netfuncs   map[string]netfunc_api.NetFunc
}

// Create a new TunIface
func NewTunIface(name string) *TunIface {
	return &TunIface{
		dispatcher: NewDispatcher(),
		name:       name,
		iface:      nil,
		netfuncs:   make(map[string]netfunc_api.NetFunc, 0),
	}
}

// Create the TunIface and set it up
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
	mtu, err := t.MTU()
	if err != nil {
		return err
	}
	t.dispatcher.Start(iface, mtu)
	return nil
}

// MTU of the TunIface
func (t *TunIface) MTU() (int64, error) {
	if strings.Contains(t.iface.Name(), "/") || strings.Contains(t.iface.Name(), ".") {
		return 0, fmt.Errorf("interface name contains illegal character")
	}
	filename := fmt.Sprintf("/sys/class/net/%s/mtu", t.iface.Name())
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimRight(string(content), "\n"), 10, 64)
}

// Stop TunIface related goroutines and delete the interface
func (t *TunIface) Delete() error {
	t.dispatcher.Stop()
	if t.iface == nil {
		return nil
	}
	if err := runIP("link", "del", t.iface.Name()); err != nil {
		return fmt.Errorf("Unable to delete interface %s:", t.iface.Name(), err)
	}
	return nil
}

// Name of the TunIface
func (t *TunIface) Name() string {
	return t.name
}

// Register a new Endpoint on this TunIface
func (t *TunIface) RegisterEndpoint(ep *config.Endpoint) error {
	if _, exists := t.netfuncs[ep.Sid]; exists {
		return fmt.Errorf("An endpoint with SID %s is already registered", ep.Sid)
	}
	e, err := netfunc.NewEndpoint(ep)
	if err != nil {
		return err
	}
	t.netfuncs[ep.Sid] = e
	t.dispatcher.Reload()
	return nil
}

// Delete an endpoint on this TunIface
func (t *TunIface) DeleteEndpoint(ep *config.Endpoint) error {
	if _, exists := t.netfuncs[ep.Sid]; !exists {
		return fmt.Errorf("Endpoint %s cannot be deleted because it is not registered", ep.Sid)
	}
	delete(t.netfuncs, ep.Sid)
	t.dispatcher.Reload()
	return nil
}

// Register a new Headend on this TunIface
func (t *TunIface) RegisterHeadend(he *config.Headend) error {
	if _, exists := t.netfuncs[he.Name]; exists {
		return fmt.Errorf("Headend %s is already registered", he.Name)
	}
	h, err := netfunc.NewHeadend(he)
	if err != nil {
		return err
	}
	t.netfuncs[he.Name] = h
	t.dispatcher.Reload()
	return nil
}

// Delete a Headend on this TunIface
func (t *TunIface) DeleteHeadend(he *config.Headend) error {
	if _, exists := t.netfuncs[he.SourceAddress]; !exists {
		return fmt.Errorf("Endpoint %s cannot be deleted because it is not registered", he.Name)
	}
	delete(t.netfuncs, he.SourceAddress)
	t.dispatcher.Reload()
	return nil
}
