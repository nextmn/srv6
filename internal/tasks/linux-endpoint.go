// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks

import (
	"context"

	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/constants"
	"github.com/nextmn/srv6/internal/iproute2"
)

// TaskLinuxEndpoint creates a new linux endpoint
type TaskLinuxEndpoint struct {
	WithName
	WithState
	endpoint   *config.Endpoint
	table      iproute2.Table
	iface_name string
}

// Create a new TaskLinuxEndpoint
func NewTaskLinuxEndpoint(name string, endpoint *config.Endpoint, table_name string, iface_name string) *TaskLinuxEndpoint {
	return &TaskLinuxEndpoint{
		WithName:   NewName(name),
		WithState:  NewState(),
		endpoint:   endpoint,
		table:      iproute2.NewTable(table_name, constants.RT_PROTO_NEXTMN),
		iface_name: iface_name,
	}
}

// Init
func (t *TaskLinuxEndpoint) RunInit(ctx context.Context) error {
	if err := t.table.AddSeg6Local(t.endpoint.Prefix, t.endpoint.Behavior, t.iface_name); err != nil {
		return err
	}
	t.state = true
	return nil
}

// Exit
func (t *TaskLinuxEndpoint) RunExit() error {
	if err := t.table.DelSeg6Local(t.endpoint.Prefix, t.endpoint.Behavior, t.iface_name); err != nil {
		return err
	}
	t.state = false
	return nil
}
