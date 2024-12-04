// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks

import (
	"context"
	"net/netip"

	"github.com/nextmn/srv6/internal/constants"
	"github.com/nextmn/srv6/internal/iproute2"
)

// TaskIPRule
type TaskIPRule struct {
	WithName
	WithState
	prefix  netip.Prefix
	family4 bool
	table   iproute2.Table
}

// Create a new Task for IPRule
func NewTaskIP6Rule(name string, prefix netip.Prefix, table_name string) *TaskIPRule {
	return &TaskIPRule{
		WithName:  NewName(name),
		WithState: NewState(),
		family4:   false,
		prefix:    prefix,
		table:     iproute2.NewTable(table_name, constants.RT_PROTO_NEXTMN),
	}
}

// Create a new Task for IPRule
func NewTaskIP4Rule(name string, prefix netip.Prefix, table_name string) *TaskIPRule {
	return &TaskIPRule{
		WithName:  NewName(name),
		WithState: NewState(),
		family4:   true,
		prefix:    prefix,
		table:     iproute2.NewTable(table_name, constants.RT_PROTO_NEXTMN),
	}
}

// Setup ip rules
func (t *TaskIPRule) RunInit(ctx context.Context) error {
	if t.family4 {
		if err := t.table.AddRule4(t.prefix.String()); err != nil {
			return err
		}
	} else {
		if err := t.table.AddRule6(t.prefix.String()); err != nil {
			return err
		}
	}
	t.state = true
	return nil
}

// Delete ip rules
func (t *TaskIPRule) RunExit() error {
	if t.family4 {
		if err := t.table.DelRule4(t.prefix.String()); err != nil {
			return err
		}
	} else {
		if err := t.table.DelRule6(t.prefix.String()); err != nil {
			return err
		}
	}
	t.state = false
	return nil
}
