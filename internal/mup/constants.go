// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package mup

// Sizes for IP Packets and IPv6 Fields
const (
	// Size of IPv6 Address
	IPV6_ADDR_SIZE_BYTE = 16
	IPV6_ADDR_SIZE_BIT  = IPV6_ADDR_SIZE_BYTE * 8

	// Size of IPv4 Address
	IPV4_ADDR_SIZE_BYTE = 4
	IPV4_ADDR_SIZE_BIT  = IPV4_ADDR_SIZE_BYTE * 8

	// Size of Args.Mob.Session
	ARGS_MOB_SESSION_SIZE_BYTE = 5
	ARGS_MOB_SESSION_SIZE_BIT  = ARGS_MOB_SESSION_SIZE_BYTE * 8
)

// NextMN MGTP4IPv6Src additional fields
const (
	// UDP Port Number field
	UDP_PORT_SIZE_BYTE = 2                      // size of the field in bytes
	UDP_PORT_SIZE_BIT  = UDP_PORT_SIZE_BYTE * 8 // size of the field in bits

	// "IPv6 Length" field
	IPV6_LEN_ENCODING_SIZE_BIT = 7                                          // size of the field in bits
	IPV6_LEN_ENCODING_POS_BIT  = 0                                          // position from right of the byte in bits
	IPV6_LEN_ENCODING_POS_BYTE = 15                                         // position from left in bytes
	IPV6_LEN_ENCODING_MASK     = (0xFF >> (8 - IPV6_LEN_ENCODING_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)
)

// Args.Mob.Session fields
const (
	// Field TEID
	TEID_SIZE_BYTE = 4                  // size of the field in bytes
	TEID_SIZE_BIT  = TEID_SIZE_BYTE * 8 // size of the field in bits
	TEID_POS_BYTE  = 1                  // position of the field from the left in bytes

	// Field QFI
	QFI_SIZE_BIT = 6                            // size of the field
	QFI_POS_BIT  = 2                            // position from right of the byte in bits
	QFI_POS_BYTE = 0                            // position from left in bytes
	QFI_MASK     = (0xFF >> (8 - QFI_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)

	// Field R
	R_SIZE_BIT = 1                          // size of the field
	R_POS_BIT  = 1                          // position from right of the byte in bits
	R_POS_BYTE = 0                          // position from left in bytes
	R_MASK     = (0xFF >> (8 - R_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)

	// Field U
	U_SIZE_BIT = 1                          // size of the field
	U_POS_BIT  = 0                          // position from right of the byte in bits
	U_POS_BYTE = 0                          // position from left in bytes
	U_MASK     = (0xFF >> (8 - U_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)
)
