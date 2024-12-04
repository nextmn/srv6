// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package ctrl

import (
	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n4tosrv6"
)

type ControllerRegistry struct {
	RemoteControlURI jsonapi.ControlURI // URI of the controller
	LocalControlURI  jsonapi.ControlURI // URI of the router, used to control it
	Locator          n4tosrv6.Locator
	Backbone         n4tosrv6.BackboneIP
	Resource         string
}
