// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package mup

import "net/netip"

func ExampleMGTP4IPv6Dst() {
	dst := NewMGTP4IPv6Dst(netip.MustParsePrefix("3fff::/20"), netip.MustParseAddr("203.0.113.1").As4(), NewArgsMobSession(0, false, false, 1))
	dst.Marshal()
}
