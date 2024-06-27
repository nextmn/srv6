// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"net/netip"
)

// Use this as a base for new handlers
type BaseHandler struct {
	prefix   netip.Prefix
	ttl      uint8
	hopLimit uint8
}

func NewBaseHandler(prefix netip.Prefix, ttl uint8, hopLimit uint8) BaseHandler {
	return BaseHandler{
		prefix:   prefix,
		ttl:      ttl,
		hopLimit: hopLimit,
	}
}

// Return prefix of the Handler as a *netip.Prefix
func (h BaseHandler) Prefix() netip.Prefix {
	return h.prefix
}

func (h BaseHandler) TTL() uint8 {
	return h.ttl
}

func (h BaseHandler) HopLimit() uint8 {
	return h.hopLimit
}

// Return the packet IP destination address (first network layer) if it is in the prefix range
func (h BaseHandler) CheckDAInPrefixRange(pqt *Packet) (netip.Addr, error) {
	return pqt.CheckDAInPrefixRange(h.Prefix())
}

func (h BaseHandler) GetSrcAddr(pqt *Packet) (netip.Addr, error) {
	return pqt.GetSrcAddr()
}
