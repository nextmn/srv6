// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"fmt"

	app_api "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/app/api"
	"github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/iproute2"
	"github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/runtime"
)

// TaskBlackhole
type TaskBlackhole struct {
	registry   app_api.Registry
	iface_name string
	table      *iproute2.Table
	state      bool
}

// Create a new TaskBlackhole
func NewTaskBlackhole(iface_name string, registry app_api.Registry) *TaskBlackhole {
	return &TaskBlackhole{
		registry:   registry,
		iface_name: iface_name,
		table:      nil,
		state:      false,
	}
}

// Create table if not existing
func (t *TableBlackhole) createTable() error {
	if t.table != nil {
		// nothing to do
		return nil
	}
	if iface, exists := t.registry.Iface(t.iface_name); exists {
		t.table = NewTable(iface, runtime.RT_PROTO_NEXTMN)
		return nil
	}
	return fmt.Errorf("Interface %s not found in registry", t.iface_name)
}

// Create blackhole
func (t *TaskBlackhole) RunInit() error {
	if err := t.createTable(); err != nil {
		return err
	}
	if err := t.table.AddDefaultRoutesBlackhole(); err != nil {
		return err
	}
	t.state = true
	return nil
}

// Delete blackhole
func (t *TaskBlackhole) RunExit() error {
	if err := t.createTable(); err != nil {
		return err
	}
	if err := t.table.DelDefaultRoutesBlackhole(); err != nil {
		return err
	}
	t.state = false
	return nil
}

// Returns state of the task
func (t *TaskTable) State() bool {
	return t.state
}
