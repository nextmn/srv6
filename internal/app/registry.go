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
	endpoints map[string]netfunc_api.Endpoint
	headends  map[string]netfunc_api.Headend
}

func NewRegistry() *Registry {
	return &Registry{
		ifaces:    make(map[string]iproute2_api.Iface),
		endpoints: make(map[string]netfunc_api.Endpoint),
		headends:  make(map[string]netfunc_api.Headend),
	}
}

func (r *Registry) Iface(name string) (iproute2_api.Iface, bool) {
	iface, exists := r.ifaces[name]
	return iface, exists
}

func (r *Registry) RegisterIface(iface iproute2_api.Iface) error {
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
	endpoint, exists := r.endpoints[name]
	return endpoint, exists
}

func (r *Registry) RegisterEndpoint(endpoint netfunc_api.Endpoint) error {
	if _, exists := r.endpoints[endpoint.Name()]; exists {
		return fmt.Errorf("Endpoint %s is already registered.", endpoint.Name())
	}
	r.endpoints[endpoint.Name()] = endpoint
	return nil
}
func (r *Registry) DeleteEndpoint(name string) {
	delete(r.endpoints, name)
}

func (r *Registry) Headend(name string) (netfunc_api.Headend, bool) {
	headend, exists := r.headends[name]
	return headend, exists
}

func (r *Registry) RegisterHeadend(headend netfunc_api.Headend) error {
	if _, exists := r.headends[headend.Name()]; exists {
		return fmt.Errorf("Headend %s is already registered.", headend.Name())
	}
	r.headends[headend.Name()] = headend
	return nil
}
func (r *Registry) DeleteHeadend(name string) {
	delete(r.headends, name)
}
