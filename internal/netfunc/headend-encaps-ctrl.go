// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"net/netip"

	"github.com/nextmn/srv6/internal/ctrl"
)

type HeadendEncapsWithCtrl struct {
	RulesRegistry *ctrl.RulesRegistry
	BaseHandler
}

func NewHeadendEncapsWithCtrl(prefix netip.Prefix, rr *ctrl.RulesRegistry, ttl uint8, hopLimit uint8) *HeadendEncapsWithCtrl {
	return &HeadendEncapsWithCtrl{
		RulesRegistry: rr,
		BaseHandler:   NewBaseHandler(prefix, ttl, hopLimit),
	}
}

// Handle a packet
func (h HeadendEncapsWithCtrl) Handle(packet []byte) ([]byte, error) {
	return nil, fmt.Errorf("Not yet implemented")
}
