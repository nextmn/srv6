// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// (network) LayerType for this packet (LayerTypeIPv4 or LayerTypeIPv6)
func networkLayerType(packet []byte) (*gopacket.LayerType, error) {
	version := (packet[0] >> 4) & 0x0F
	switch version {
	case 4:
		return &layers.LayerTypeIPv4, nil
	case 6:
		return &layers.LayerTypeIPv6, nil
	default:
		return nil, fmt.Errorf("Malformed packet")

	}
}
