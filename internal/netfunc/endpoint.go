// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"net/netip"

	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/iana"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

func NewEndpoint(ec *config.Endpoint) (netfunc_api.NetFunc, error) {
	p, err := netip.ParsePrefix(ec.Sid)
	if err != nil {
		return nil, err
	}
	switch ec.Behavior {
	case iana.End_M_GTP4_E:
		return NewNetFunc(NewEndpointMGTP4E(p)), nil
	default:
		return nil, fmt.Errorf("Unsupported endpoint behavior (%s) with this provider (%s)", ec.Behavior, ec.Provider)
	}
}
