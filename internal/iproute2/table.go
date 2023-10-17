// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package iproute2

import "github.com/nextmn/srv6/internal/iana"

// IPRoute2 Table
type Table struct {
	table string // table name
	proto string // proto name
}

// Create a new Table
func NewTable(table string, proto string) Table {
	return Table{
		table: table,
		proto: proto,
	}
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
func (t Table) AddRule4(to string) error {
	return t.addRule4("to", to, "lookup", t.table)
}

// Delete a rule to lookup the table, for IPv4
func (t Table) DelRule4(to string) error {
	return t.delRule4("to", to, "lookup", t.table)
}

// Add a new rule to lookup the table, for IPv6
func (t Table) AddRule6(to string) error {
	return t.addRule6("to", to, "lookup", t.table)
}

// Delete a rule to lookup the table, for IPv6
func (t Table) DelRule6(to string) error {
	return t.delRule6("to", to, "lookup", t.table)
}

// Add a route on this table, protocol independent
func (t Table) AddRoute(args ...string) error {
	a := []string{"route", "add"}
	table := []string{"table", t.table}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP(a...)
}

// Delete a route on this table, protocol independent
func (t Table) DelRoute(args ...string) error {
	a := []string{"route", "del"}
	table := []string{"table", t.table}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP(a...)
}

// Add a route on this table, for IPv4
func (t Table) AddRoute4(args ...string) error {
	a := []string{"route", "add"}
	table := []string{"table", t.table}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP4(a...)
}

// Delete a route on this table, for IPv4
func (t Table) DelRoute4(args ...string) error {
	a := []string{"route", "del"}
	table := []string{"table", t.table}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP4(a...)
}

// Add a route on this table, for IPv6
func (t Table) AddRoute6(args ...string) error {
	a := []string{"route", "add"}
	table := []string{"table", t.table}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP6(a...)
}

// Delete a route on this table, for IPv6
func (t Table) DelRoute6(args ...string) error {
	a := []string{"route", "del"}
	table := []string{"table", t.table}
	a = append(a, args...)
	a = append(a, table...)
	return t.runIP6(a...)
}

// Add default blackhole routes
func (t Table) AddDefaultRoutesBlackhole() error {
	if err := t.AddRoute4("blackhole", "default"); err != nil {
		return err
	}
	if err := t.AddRoute6("blackhole", "default"); err != nil {
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

// Add Linux SRv6 Endpoint
func (t Table) AddSeg6Local(sid string, behavior iana.EndpointBehavior, dev string) error {
	linux_behavior, err := behavior.ToIPRoute2Action()
	if err != nil {
		return err
	}
	if err := t.AddRoute6(sid, "encap", "seg6local", "action", linux_behavior, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Delete Linux SRv6 Endpoint
func (t Table) DelSeg6Local(sid string, behavior iana.EndpointBehavior, dev string) error {
	linux_behavior, err := behavior.ToIPRoute2Action()
	if err != nil {
		return err
	}
	if err := t.DelRoute6(sid, "encap", "seg6local", "action", linux_behavior, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Add Linux Headend with encap
func (t Table) AddSeg6Encap(prefix string, segmentsList string, dev string) error {
	if err := t.AddRoute(prefix, "encap", "seg6", "mode", "encap", "segs", segmentsList, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Add Linux Headend with encap and MTU
func (t Table) AddSeg6EncapWithMTU(prefix string, segmentsList string, dev string, mtu string) error {
	if err := t.AddRoute(prefix, "encap", "seg6", "mode", "encap", "segs", segmentsList, "dev", dev, "mtu", mtu); err != nil {
		return err
	}
	return nil
}

// Delete Linux Headend with encap
func (t Table) DelSeg6Encap(prefix string, segmentsList string, dev string) error {
	if err := t.DelRoute(prefix, "encap", "seg6", "mode", "encap", "segs", segmentsList, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Add Linux Headend with inline
// Inline mode is only for incomming packets already having an IPv6 header
func (t Table) AddSeg6Inline(prefix string, segmentsList string, dev string) error {
	if err := t.AddRoute6(prefix, "encap", "seg6", "mode", "inline", "segs", segmentsList, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Delete Linux Headend with inline
// Inline mode is only for incomming packets already having an IPv6 header
func (t Table) DelSeg6Inline(prefix string, segmentsList string, dev string) error {
	if err := t.DelRoute6(prefix, "encap", "seg6", "mode", "inline", "segs", segmentsList, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Add IPv6 Route to Tun iface
func (t Table) AddRoute6Tun(prefix string, dev string) error {
	if err := t.AddRoute6(prefix, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Delete IPv6 Route to Tun iface
func (t Table) DelRoute6Tun(prefix string, dev string) error {
	if err := t.DelRoute6(prefix, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Add IPv4 Route to Tun iface
func (t Table) AddRoute4Tun(prefix string, dev string) error {
	if err := t.AddRoute4(prefix, "dev", dev); err != nil {
		return err
	}
	return nil
}

// Delete IPv4 Route to Tun iface
func (t Table) DelRoute4Tun(prefix string, dev string) error {
	if err := t.DelRoute4(prefix, "dev", dev); err != nil {
		return err
	}
	return nil
}
