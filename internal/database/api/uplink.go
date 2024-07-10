// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package database_api

import (
	"github.com/gofrs/uuid"
	"net/netip"
)

type Uplink interface {
	InsertAction(uplinkTeid uint32, SrgwIp netip.Addr, GnbIp netip.Addr, actionUuid uuid.UUID) error
	UpdateAction(uplinkTeid uint32, SrgwIp netip.Addr, GnbIp netip.Addr, actionUuid uuid.UUID) error
	GetUplinkAction(UplinkTeid uint32, SrgwIp netip.Addr, GnbIp netip.Addr) (uuid.UUID, error)
}
