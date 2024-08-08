// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
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
	WithName
	WithState
	endpoint   *config.Endpoint
	table      iproute2.Table
	registry   app_api.Registry
	iface_name string
}

// Create a new TaskNextMNEndpoint
func NewTaskNextMNEndpoint(name string, endpoint *config.Endpoint, table_name string, iface_name string, registry app_api.Registry) *TaskNextMNEndpoint {
	return &TaskNextMNEndpoint{
		WithName:   NewName(name),
		WithState:  NewState(),
		endpoint:   endpoint,
		table:      iproute2.NewTable(table_name, constants.RT_PROTO_NEXTMN),
		iface_name: iface_name,
		registry:   registry,
	}
}

// Init
func (t *TaskNextMNEndpoint) RunInit(ctx context.Context) error {
	// Create and start endpoint
	tunIface, ok := t.registry.TunIface(t.iface_name)
	if !ok {
		return fmt.Errorf("Interface %s is not in registry", t.iface_name)
	}
	ttl, err := tunIface.IPv4TTL()
	if err != nil {
		return err
	}
	hopLimit, err := tunIface.IPv6HopLimit()
	if err != nil {
		return err
	}
	var n netfunc_api.NetFunc
	if ep, err := netfunc.NewEndpoint(t.endpoint, ttl, hopLimit); err != nil {
		return err
	} else {
		n = ep
	}
	go n.Run(ctx, tunIface)
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
	t.state = false
	return nil
}
