// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"fmt"

	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/constants"
	"github.com/nextmn/srv6/internal/iproute2"
)

// TaskLinuxHeadend creates a new linux headend
type TaskLinuxHeadend struct {
	WithState
	headend    *config.Headend
	table      iproute2.Table
	iface_name string
}

// Create a new TaskLinuxHeadend
func NewTaskLinuxHeadend(headend *config.Headend, table_name string, iface_name string) *TaskLinuxHeadend {
	return &TaskLinuxHeadend{
		WithState:  NewState(),
		headend:    headend,
		table:      iproute2.NewTable(table_name, constants.RT_PROTO_NEXTMN),
		iface_name: iface_name,
	}
}

// Init
func (t *TaskLinuxHeadend) RunInit() error {
	if t.headend.Policy == nil {
		return fmt.Errorf("No policy set for this headend")
	}
	seglist := ""
	for _, p := range t.headend.Policy {
		if p.Match == nil {
			seglist = p.Bsid.ToIPRoute2()
			break
		}
	}

	switch t.headend.Behavior {
	case config.H_Encaps:
		if t.headend.MTU != nil {
			if err := t.table.AddSeg6EncapWithMTU(t.headend.To, seglist, t.iface_name, *t.headend.MTU); err != nil {
				return err
			}
		} else {
			if err := t.table.AddSeg6Encap(t.headend.To, seglist, t.iface_name); err != nil {
				return err
			}
		}
	case config.H_Inline:
		if err := t.table.AddSeg6Inline(t.headend.To, seglist, t.iface_name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported headend behavior (%s) with this provider (%s)", t.headend.Behavior, t.headend.Provider)
	}
	t.state = true
	return nil
}

// Exit
func (t *TaskLinuxHeadend) RunExit() error {
	if t.headend.Policy == nil {
		return fmt.Errorf("No policy set for this headend")
	}
	seglist := ""
	for _, p := range t.headend.Policy {
		if p.Match == nil {
			seglist = p.Bsid.ToIPRoute2()
			break
		}

	}
	switch t.headend.Behavior {
	case config.H_Encaps:
		if err := t.table.DelSeg6Encap(t.headend.To, seglist, t.iface_name); err != nil {
			return err
		}
	case config.H_Inline:
		if err := t.table.DelSeg6Inline(t.headend.To, seglist, t.iface_name); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported headend behavior (%s) with this provider (%s).", t.headend.Behavior, t.headend.Provider)
	}
	t.state = false
	return nil
}
