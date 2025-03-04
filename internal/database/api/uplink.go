// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package database_api

import (
	"context"
	"net/netip"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n4tosrv6"
)

type Uplink interface {
	GetUplinkAction(ctx context.Context, UplinkFTeid jsonapi.Fteid, GnbIp netip.Addr, UeIp netip.Addr, ServiceIp netip.Addr) (n4tosrv6.Action, error)
}
