// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import "github.com/louisroyer/nextmn-srv6/iproute2"

var RTTableNextMNSRv6 = iproute2.NewConfig("nextmn-srv6", iproute2.RTTable) // FIXME: number
var RTTableNextMNGTP4 = iproute2.NewConfig("nextmn-gtp4", iproute2.RTTable) // FIXME: number
var ProtoNextMN = iproute2.NewConfig("nextmn", iproute2.Proto)              // FIXME: number

const NextmnSRTunName = "nextmn-sr"
const LinuxSRLinkName = "linux-sr"
const GTPU_PORT = "2152"
