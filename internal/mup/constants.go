// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

// minimum sizes to hold values

const IPV4_ADDR_SIZE_BYTE = 4
const IPV4_ADDR_SIZE_BIT = IPV4_ADDR_SIZE_BYTE * 8

const IPV6_ADDR_SIZE_BYTE = 16
const IPV6_ADDR_SIZE_BIT = IPV6_ADDR_SIZE_BYTE * 8

const ARGS_MOB_SESSION_SIZE_BYTE = 5
const ARGS_MOB_SESSION_SIZE_BIT = ARGS_MOB_SESSION_SIZE_BYTE * 8

const UDP_PORT_SIZE_BYTE = 2
const UDP_PORT_SIZE_BIT = UDP_PORT_SIZE_BYTE * 8

const IPV6_LEN_ENCODING_SIZE_BIT = 7
const IPV6_LEN_ENCODING_POS_BIT = 0                                       // position from right of the byte in bits
const IPV6_LEN_ENCODING_POS_BYTE = 15                                     // position from left in bytes
const IPV6_LEN_ENCODING_MASK = (0xFF >> (8 - IPV6_LEN_ENCODING_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)

const TEID_SIZE_BYTE = 4
const TEID_SIZE_BIT = TEID_SIZE_BYTE * 8
const TEID_POS_BYTE = 1

const QFI_SIZE_BIT = 6
const QFI_POS_BIT = 2                         // position from right of the byte in bits
const QFI_POS_BYTE = 0                        // position from left in bytes
const QFI_MASK = (0xFF >> (8 - QFI_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)

const R_SIZE_BIT = 1
const R_POS_BIT = 1                       // position from right of the byte in bits
const R_POS_BYTE = 0                      // position from left in bytes
const R_MASK = (0xFF >> (8 - R_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)

const U_SIZE_BIT = 1
const U_POS_BIT = 0                       // position from right of the byte in bits
const U_POS_BYTE = 0                      // position from left in bytes
const U_MASK = (0xFF >> (8 - U_SIZE_BIT)) // mask (decoding: after shift to right; encoding before shift to left)
