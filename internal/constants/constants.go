// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package constants

// ports
const GTPU_PORT = "2152"

// iproute2 rt protos
const RT_PROTO_NEXTMN = "nextmn"

// iproute2 rt tables
const RT_TABLE_NEXTMN_SRV6 = "nextmn/srv6"
const RT_TABLE_NEXTMN_GTP4 = "nextmn/gtp4"
const RT_TABLE_MAIN = "main"

// iproute2 ifaces
const IFACE_LINUX_SRV6 = "linux-srv6"

// golang/water ifaces
const IFACE_GOLANG_SRV6 = "golang-srv6"
const IFACE_GOLANG_GTP4 = "golang-gtp4"
