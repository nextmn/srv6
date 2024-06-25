// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package ctrl_api

import (
	"github.com/gofrs/uuid"
	"github.com/nextmn/json-api/jsonapi"
	"net/netip"
)

type RulesRegistry interface {
	Action(UEIp netip.Addr) (uuid.UUID, jsonapi.Action, error)
	ByUUID(uuid uuid.UUID) (jsonapi.Action, error)
}
