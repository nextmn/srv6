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
	if err := t.DropIcmpRedirect(); err != nil {
		return err
	}
	if err := runIP("link", "set", "dev", t.iface.Name(), "up"); err != nil {
		return err
	}
	return nil
}

// Stop TunIface related goroutines and delete the interface
func (t *TunIface) Delete() error {
	if t.iface == nil {
		return nil
	}
	if err := runIP("link", "del", t.iface.Name()); err != nil {
		return fmt.Errorf("Unable to delete interface %s: %s", t.iface.Name(), err)
	}
	if err := t.CancelDropIcmpRedirect(); err != nil {
		return err
	}
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

// IPv6 Hop Limit of the TunIface
func (t *TunIface) IPv6HopLimit() (uint8, error) {
	if strings.Contains(t.iface.Name(), "/") || strings.Contains(t.iface.Name(), ".") {
		return 0, fmt.Errorf("interface name contains illegal character")
	}
	filename := fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/hop_limit", t.iface.Name())
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	ret, err := strconv.ParseUint(strings.TrimRight(string(content), "\n"), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(ret), nil
}

// IPv4 default TTL
func (t *TunIface) IPv4TTL() (uint8, error) {
	filename := "/proc/sys/net/ipv4/ip_default_ttl"
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	ret, err := strconv.ParseUint(strings.TrimRight(string(content), "\n"), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(ret), nil
}

// Drop ICMP/ICMPv6 redirects on the interface
func (t *TunIface) DropIcmpRedirect() error {
	if t.iface == nil {
		return nil
	}
	if err := runIPTables("-A", "OUTPUT", "-o", t.iface.Name(), "-p", "icmp", "--icmp-type", "redirect", "-j", "DROP"); err != nil {
		return fmt.Errorf("Unable to drop icmp redirect on interface %s: %s", t.iface.Name(), err)
	}
	if err := runIP6Tables("-A", "OUTPUT", "-o", t.iface.Name(), "-p", "icmpv6", "--icmpv6-type", "redirect", "-j", "DROP"); err != nil {
		return fmt.Errorf("Unable to drop icmpv6 redirect on interface %s: %s", t.iface.Name(), err)
	}
	return nil
}

// Cancel Drop ICMP/ICMPv6 redirects on the interface
func (t *TunIface) CancelDropIcmpRedirect() error {
	if t.iface == nil {
		return nil
	}
	if err := runIP6Tables("-D", "OUTPUT", "-o", t.iface.Name(), "-p", "icmpv6", "--icmpv6-type", "redirect", "-j", "DROP"); err != nil {
		return fmt.Errorf("Unable to drop icmpv6 redirect on interface %s: %s", t.iface.Name(), err)
	}
	if err := runIPTables("-D", "OUTPUT", "-o", t.iface.Name(), "-p", "icmp", "--icmp-type", "redirect", "-j", "DROP"); err != nil {
		return fmt.Errorf("Unable to drop icmp redirect on interface %s: %s", t.iface.Name(), err)
	}
	return nil
}

// Name of the TunIface
func (t *TunIface) Name() string {
	return t.name
}

// Read a packet from the water interface
func (t *TunIface) Read(b []byte) (int, error) {
	return t.iface.Read(b)
}

// Write a packet to the water interface
func (t *TunIface) Write(b []byte) (int, error) {
	return t.iface.Write(b)
}
