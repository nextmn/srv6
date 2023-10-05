// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	iproute2 "github.com/nextmn/srv6/internal/iproute2"
)

// TaskDummyIface
type TaskDummyIface struct {
	WithState
	iface *iproute2.DummyIface
}

// Create a new Task for DummyIface
func NewTaskDummyIface(name string) *TaskDummyIface {
	return &TaskDummyIface{
		WithState: NewState(),
		iface:     iproute2.NewDummyIface(name),
	}
}

// Create and set up the Iface
func (t *TaskDummyIface) RunInit() error {
	if err := t.iface.CreateAndUp(); err != nil {
		return err
	}
	t.state = true
	return nil
}

// Delete the Iface
func (t *TaskDummyIface) RunExit() error {
	if err := t.iface.Delete(); err != nil {
		return err
	}
	t.state = false
	return nil
}
