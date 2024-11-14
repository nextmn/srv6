// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package ctrl

import (
	"github.com/nextmn/json-api/jsonapi"
)

type ControllerRegistry struct {
	RemoteControlURI jsonapi.ControlURI // URI of the controller
	LocalControlURI  jsonapi.ControlURI // URI of the router, used to control it
	Locator          jsonapi.Locator
	Backbone         jsonapi.BackboneIP
	Resource         string
}
