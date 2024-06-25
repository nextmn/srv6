// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package app_api

import "github.com/nextmn/srv6/internal/iproute2"
import "github.com/nextmn/srv6/internal/ctrl"

type Registry interface {
	// ifaces
	TunIface(name string) (*iproute2.TunIface, bool)
	RegisterTunIface(iface *iproute2.TunIface) error
	DeleteTunIface(name string)
	RegisterControllerRegistry(*ctrl.ControllerRegistry)
	ControllerRegistry() (*ctrl.ControllerRegistry, bool)
	DeleteControllerRegistry()
}
