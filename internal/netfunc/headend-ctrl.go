// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"net/netip"

	app_api "github.com/nextmn/srv6/internal/app/api"
	"github.com/nextmn/srv6/internal/config"
	netfunc_api "github.com/nextmn/srv6/internal/netfunc/api"
)

func NewHeadendWithCtrl(he *config.Headend, ttl uint8, hopLimit uint8, debug bool, setup_registry app_api.Registry) (netfunc_api.NetFunc, error) {
	p, err := netip.ParsePrefix(he.To)
	if err != nil {
		return nil, err
	}
	switch he.Behavior {
	case config.H_Encaps:
		db, ok := setup_registry.DB()
		if !ok {
			return nil, fmt.Errorf("No database in the registry")
		}
		return NewNetFunc(NewHeadendEncapsWithCtrl(p, ttl, hopLimit, db), debug), nil
	case config.H_M_GTP4_D:
		db, ok := setup_registry.DB()
		if !ok {
			return nil, fmt.Errorf("No database in the registry")
		}

		g, err := NewHeadendGTP4WithCtrl(p, ttl, hopLimit, db)
		if err != nil {
			return nil, err
		}
		return NewNetFunc(g, debug), nil
	default:
		return nil, fmt.Errorf("Unsupported headend behavior (%s) with this provider (%s)", he.Behavior, he.Provider)
	}
}
