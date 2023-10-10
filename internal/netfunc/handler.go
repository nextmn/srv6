// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import "net/netip"

// Use this as a base for new handlers
type Handler struct {
	prefix netip.Prefix
}

func NewHandler(prefix netip.Prefix) Handler {
	return Handler{
		prefix: prefix,
	}
}

// Return prefix of the Handler as a *netip.Prefix
func (h *Handler) Prefix() netip.Prefix {
	return h.prefix
}
