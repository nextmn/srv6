// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package ctrl

import (
	"net/netip"
)

type ControllerRegistry struct {
	RemoteControlURI string // URI of the controller
	LocalControlURI  string // URI of the router, used to control it
	Locator          string
	Backbone         netip.Addr
	Resource         string
}
