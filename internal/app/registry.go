// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package app

import (
	"fmt"

	iproute2_api "github.com/nextmn/srv6/internal/iproute2/api"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

type Registry struct {
	ifaces    map[string]iproute2_api.Iface
	endpoints map[string]netfunc_api.Endpoints
}

func NewRegistry() *Registry {
	return &Registry{
		ifaces:    make(map[string]iproute2_api.Iface),
		endpoints: make(map[string]netfunc_api.Endpoint),
	}
}

func (r *Registry) Iface(name string) (iproute2.Iface, bool) {
	return r.ifaces[name]
}

func (r *Registry) RegisterIface(iface iproute2.Iface) error {
	if _, exists := r.ifaces[iface.Name()]; exists {
		return fmt.Errorf("Iface %s is already registered.", iface.Name())
	}
	r.ifaces[iface.Name()] = iface
	return nil
}

func (r *Registry) DeleteIface(name string) {
	delete(r.ifaces, name)
}

func (r *Registry) Endpoint(name string) (netfunc_api.Endpoint, bool) {
	return r.endpoints[name]
}

func (r *Registry) RegisterEndpoint(endpoint netfunc_api.Endpoint) error {
	if _, exists := r.endpoints[endpoint.Name()]; exists {
		return fmt.Errorf("Endpoint %s is already registered.", endpoint.Name())
	}
	r.endpoint[endpoint.Name()] = endpoint
	return nil
}
func (r *Registry) DeleteEndpoint(name string) error {
	delete(r.endpoints, name)
}
