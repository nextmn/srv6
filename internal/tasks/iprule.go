// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"github.com/nextmn/srv6/internal/constants"
	"github.com/nextmn/srv6/internal/iproute2"
)

// TaskIPRule
type TaskIPRule struct {
	WithState
	prefix  string
	family4 bool
	table   iproute2.Table
}

// Create a new Task for IPRule
func NewTaskIP6Rule(prefix string, tablename string) *TaskIPRule {
	return &TaskIPRule{
		WithState: NewState(),
		family4:   false,
		prefix:    prefix,
		table:     iproute2.NewTable(tablename, constants.RT_PROTO_NEXTMN),
	}
}

// Create a new Task for IPRule
func NewTaskIP4Rule(prefix string, tablename string) *TaskIPRule {
	return &TaskIPRule{
		WithState: NewState(),
		family4:   true,
		prefix:    prefix,
		table:     iproute2.NewTable(tablename, constants.RT_PROTO_NEXTMN),
	}
}

// Setup ip rules
func (t *TaskIPRule) RunInit() error {
	if t.family4 {
		if err := t.table.AddRule4(t.prefix); err != nil {
			return err
		}
	} else {
		if err := t.table.AddRule6(t.prefix); err != nil {
			return err
		}
	}
	t.state = true
	return nil
}

// Delete ip rules
func (t *TaskIPRule) RunExit() error {
	if t.family4 {
		if err := t.table.DelRule4(t.prefix); err != nil {
			return err
		}
	} else {
		if err := t.table.DelRule6(t.prefix); err != nil {
			return err
		}
	}
	t.state = false
	return nil
}
