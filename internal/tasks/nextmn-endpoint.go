// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"fmt"

	app_api "github.com/nextmn/srv6/internal/app/api"
	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/constants"
	"github.com/nextmn/srv6/internal/iproute2"
	"github.com/nextmn/srv6/internal/netfunc"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

// TaskNextMNEndpoint creates a new endpoint
type TaskNextMNEndpoint struct {
	WithState
	endpoint   *config.Endpoint
	table      iproute2.Table
	registry   app_api.Registry
	iface_name string
	netfunc    netfunc_api.NetFunc
}

// Create a new TaskNextMNEndpoint
func NewTaskNextMNEndpoint(endpoint *config.Endpoint, table_name string, iface_name string, registry app_api.Registry) *TaskNextMNEndpoint {
	return &TaskNextMNEndpoint{
		WithState:  NewState(),
		endpoint:   endpoint,
		table:      iproute2.NewTable(table_name, constants.RT_PROTO_NEXTMN),
		iface_name: iface_name,
		registry:   registry,
		netfunc:    nil,
	}
}

// Init
func (t *TaskNextMNEndpoint) RunInit() error {
	// Create and start endpoint
	tunIface, ok := t.registry.TunIface(t.iface_name)
	if !ok {
		return fmt.Errorf("Interface %s is not in registry", t.iface_name)
	}
	if ep, err := netfunc.NewEndpoint(t.endpoint); err != nil {
		return err
	} else {
		t.netfunc = ep
	}
	t.netfunc.Start(tunIface)
	// Add route to endpoint
	if err := t.table.AddRoute6Tun(t.endpoint.Prefix, t.iface_name); err != nil {
		return err
	}
	t.state = true
	return nil
}

// Exit
func (t *TaskNextMNEndpoint) RunExit() error {
	// Remove route to endpoint
	if err := t.table.DelRoute6Tun(t.endpoint.Prefix, t.iface_name); err != nil {
		return err
	}
	// Stop endpoint
	t.netfunc.Stop()
	t.state = false
	return nil
}
