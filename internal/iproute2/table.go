// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package iproute2

import (
	iproute2_api "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/iproute2/api"
)

// IPRoute2 Table
type Table struct {
	iface iproute2_api.Iface // interface
	proto string             // proto name
}

// Create a new Table
func NewTable(iface iproute2_api.Iface, proto string) Table {
	return Table{iface: iface, proto: proto}
}

// Run an IProute2 command using defined proto
func (t Table) runIP(args ...string) error {
	args = append(args, "protocol", t.proto)
	return runIP(args...)
}

// Run an IPRoute2 command using defined proto, for IPv4
func (t Table) runIP4(args ...string) error {
	a := []string{"-4"}
	a = append(a, args...)
	return t.runIP(a...)
}

// Run an IPRoute2 command using defined proto, for IPv6
func (t Table) runIP6(args ...string) error {
	a := []string{"-6"}
	a = append(a, args...)
	return t.runIP(a...)
}

// Add a new rule, for IPv4
func (t Table) addRule4(args ...string) error {
	a := []string{"rule", "add"}
	a = append(a, args...)
	return t.runIP4(a...)
}

// Delete a rule, for IPv4
func (t Table) delRule4(args ...string) error {
	a := []string{"rule", "del"}
	a = append(a, args...)
	return t.runIP4(a...)
}

// Add a new rule, for IPv6
func (t Table) addRule6(args ...string) error {
	a := []string{"rule", "add"}
	a = append(a, args...)
	return t.runIP6(a...)
}

// Delete a rule, for IPv6
func (t Table) delRule6(args ...string) error {
	a := []string{"rule", "del"}
	a = append(a, args...)
	return t.runIP6(a...)
}

// public methods

// Add a new rule to lookup the table, for IPv4
func (t Table) AddRule4(to string, table string) error {
	return t.addRule4("to", to, "lookup", t.iface.Name())
}

// Delete a rule to lookup the table, for IPv4
func (t Table) DelRule4(to string, table string) error {
	return t.delRule4("to", to, "lookup", t.iface.Name())
}

// Add a new rule to lookup the table, for IPv6
func (t Table) AddRule6(to string, table string) error {
	return t.addRule6("to", to, "lookup", t.iface.Name())
}

// Delete a rule to lookup the table, for IPv6
func (t Table) DelRule6(to string, table string) error {
	return t.delRule6("to", to, "lookup", t.iface.Name())
}

// Add a route on this table, for IPv4
func (t Table) AddRoute4(args ...string) error {
	a := []string{"route", "add"}
	table := []string{"table", t.iface.Name()}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP4(a...)
}

// Delete a route on this table, for IPv4
func (t Table) DelRoute4(args ...string) error {
	a := []string{"route", "del"}
	table := []string{"table", t.iface.Name()}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP4(a...)
}

// Add a route on this table, for IPv6
func (t Table) AddRoute6(args ...string) error {
	a := []string{"route", "add"}
	table := []string{"table", t.iface.Name()}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP6(a...)
}

// Delete a route on this table, for IPv6
func (t Table) DelRoute6(args ...string) error {
	a := []string{"route", "del"}
	table := []string{"table", t.iface.Name()}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP6(a...)
}

// Add default blackhole routes
func (t Table) AddDefaultRoutesBlackhole() error {
	if err := t.AddRoute4("blackhole", "default"); err != nil {
		return err
	}
	if err := t.DelRoute6("blackhole", "default"); err != nil {
		return err
	}
	return nil
}

// Delete default blackhole routes
func (t Table) DelDefaultRoutesBlackhole() error {
	if err := t.DelRoute4("blackhole", "default"); err != nil {
		return err
	}
	if err := t.DelRoute6("blackhole", "default"); err != nil {
		return err
	}
	return nil
}
