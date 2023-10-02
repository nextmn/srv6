// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	api "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/api/iproute2"
	"github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/iproute2"
)

type TaskIPRoute2 struct {
	iface api.Iface
	state bool
}

// Create a new Task for DummyIface
func NewTaskIPRoute2DummyIface(name string) *TaskIPRoute2 {
	return &TaskIPRoute2{iface: iproute2.NewDummyIface(name)}
}

// Create and set up the DummyIface
func (t *TaskIPRoute2) RunInit() error {
	if err := t.iface.CreateAndUp(); err != nil {
		return err
	}
	t.state = true
	return nil
}

// Delete the DummyIface
func (t *TaskIPRoute2) RunExit() error {
	if err := t.iface.Delete(); err != nil {
		return err
	}
	t.state = false
	return nil
}
