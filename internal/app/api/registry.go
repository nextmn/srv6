// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package app_api

import (
	iproute2_api "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/iproute2/api"
	netfunc_api "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/netfunc/api"
)

type Registry interface {
	// ifaces
	Iface(name string) (iproute2_api.Iface, bool)
	RegisterIface(iface iproute2_api.Iface) error
	DeleteIface(name string) error

	// endpoints
	Endpoint(name string) (netfunc_api.Endpoint, bool)
	RegisterEndpoint(endpoint netfunc_api.Endpoint) error
	DeleteEndpoint(name string) error
}
