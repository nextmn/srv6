// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/iproute2"
	"github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/runtime"
)

// TaskIPRule
type TaskIPRule struct {
	state   bool
	prefix  string
	family4 bool
	table   *iproute2.Table
}

// Create a new Task for IPRule
func NewTaskIP6Rule(tablename string, prefix string) *TaskIface {
	return &TaskIPRule{
		family4: false,
		state:   false,
		prefix:  prefix,
		table:   iproute2.NewTable(tablename, runtime.RT_PROTO_NEXTMN),
	}
}

// Create a new Task for IPRule
func NewTaskIP4Rule(tablename string, prefix string) *TaskIface {
	return &TaskIPRule{
		family4: true,
		state:   false,
		prefix:  prefix,
		table:   iproute2.NewTable(tablename, runtime.RT_PROTO_NEXTMN),
	}
}

// Setup ip rules
func (t *TaskIPRule) RunInit() error {
	if t.family4 {
		if err := t.table.AddIP4Rule(prefix); err != nil {
			return err
		}
	} else {
		if err := t.table.AddIP6Rule(prefix); err != nil {
			return err
		}
	}
	t.state = true
	return nil
}

// Delete ip rules
func (t *TaskIPRule) RunExit() error {
	if t.family4 {
		if err := t.table.DelIP4Rule(prefix); err != nil {
			return err
		}
	} else {
		if err := t.tableDelIP6Rule(prefix); err != nil {
			return err
		}
	}
	t.state = false
	return nil
}

// Returns state of the task
func (t *TaskIPRule) State() bool {
	return t.state
}
