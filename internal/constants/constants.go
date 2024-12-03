// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package constants

// ports
const GTPU_PORT = "2152"
const GTPU_PORT_INT = 2152

const GTPU_MESSAGE_TYPE_ECHO_REQUEST = 1
const GTPU_MESSAGE_TYPE_ECHO_RESPONSE = 2
const GTPU_MESSAGE_TYPE_GPDU = 255

// iproute2 rt protos
const RT_PROTO_NEXTMN = "nextmn"

// iproute2 rt tables
const RT_TABLE_NEXTMN_IPV6 = "nextmn/ipv6"
const RT_TABLE_NEXTMN_IPV4 = "nextmn/ipv4"

// iproute2 ifaces
const IFACE_LINUX = "nextmn-linux"

// golang/water ifaces
const IFACE_GOLANG_SRV6_PREFIX = "nextmn-srv6-"
const IFACE_GOLANG_IPV4_PREFIX = "nextmn-ipv4-" // ipv4 excluding gtp4
const IFACE_GOLANG_GTP4_PREFIX = "nextmn-gtp4-"
