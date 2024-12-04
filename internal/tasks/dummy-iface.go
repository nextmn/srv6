// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks

import (
	"context"

	iproute2 "github.com/nextmn/srv6/internal/iproute2"
)

// TaskDummyIface
type TaskDummyIface struct {
	WithName
	WithState
	iface *iproute2.DummyIface
}

// Create a new Task for DummyIface
func NewTaskDummyIface(name string, iface_name string) *TaskDummyIface {
	return &TaskDummyIface{
		WithName:  NewName(name),
		WithState: NewState(),
		iface:     iproute2.NewDummyIface(iface_name),
	}
}

// Create and set up the Iface
func (t *TaskDummyIface) RunInit(ctx context.Context) error {
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
