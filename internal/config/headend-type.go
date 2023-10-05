// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

type HeadendType uint32

const (
	H_Encaps   HeadendType = iota // encapsulate the packet into a new IPv6 Header with a SRH
	H_Inline                      // add a SRH to an existing IPv6 Header
	H_M_GTP4_D                    // RFC 9433, section 6.7
)
