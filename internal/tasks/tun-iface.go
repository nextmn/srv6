// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks

import (
	"context"

	app_api "github.com/nextmn/srv6/internal/app/api"
	"github.com/nextmn/srv6/internal/iproute2"
)

// TaskTunIface
type TaskTunIface struct {
	WithName
	WithState
	iface    *iproute2.TunIface
	registry app_api.Registry
}

// Create a new Task for TunIface
func NewTaskTunIface(name string, iface_name string, registry app_api.Registry) *TaskTunIface {
	return &TaskTunIface{
		WithName:  NewName(name),
		WithState: NewState(),
		iface:     iproute2.NewTunIface(iface_name),
		registry:  registry,
	}
}

// Create and set up the Iface
func (t *TaskTunIface) RunInit(ctx context.Context) error {
	if err := t.iface.CreateAndUp(); err != nil {
		return err
	}
	if t.registry != nil {
		if err := t.registry.RegisterTunIface(t.iface); err != nil {
			return err
		}
	}
	t.state = true
	return nil
}

// Delete the Iface
func (t *TaskTunIface) RunExit() error {
	if err := t.iface.Delete(); err != nil {
		return err
	}
	if t.registry != nil {
		t.registry.DeleteTunIface(t.iface.Name())
	}
	t.state = false
	return nil
}
