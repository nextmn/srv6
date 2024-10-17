// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package database_api

import (
	"context"
	"net/netip"

	"github.com/nextmn/json-api/jsonapi"
)

type Uplink interface {
	GetUplinkAction(ctx context.Context, UplinkTeid uint32, GnbIp netip.Addr, UeIp netip.Addr, ServiceIp netip.Addr) (jsonapi.Action, error)
}
