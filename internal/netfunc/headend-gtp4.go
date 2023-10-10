// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/nextmn/srv6/internal/constants"
)

type HeadendGTP4 struct {
	Handler
}

func NewHeadendGTP4(prefix netip.Prefix) *HeadendGTP4 {
	return &HeadendGTP4{
		Handler: NewHandler(prefix),
	}
}

// Handle a packet
func (h *HeadendGTP4) Handle(packet []byte) error {
	layerType, err := networkLayerType(packet)
	if err != nil {
		return err
	}
	if *layerType != layers.LayerTypeIPv4 {
		return fmt.Errorf("Headend for GTP4 can only handle IPv4 packets")
	}

	// create gopacket
	pqt := gopacket.NewPacket(packet, *layerType, gopacket.Default)

	// check prefix
	dstSlice := pqt.NetworkLayer().NetworkFlow().Dst().Raw()
	dst, ok := netip.AddrFromSlice(dstSlice)
	if !ok {
		return fmt.Errorf("Malformed address")
	}
	if !h.Prefix().Contains(dst) {
		return fmt.Errorf("Destination address not in handled range")
	}

	// RFC 9433 section 6.7. H.M.GTP4.D

	// S01. IF !(Payload == UDP/GTP-U) THEN Drop the packet
	transportLayer := pqt.TransportLayer()
	if transportLayer.LayerType() != layers.LayerTypeUDP {
		return fmt.Errorf("No UDP layer")
	}
	// TODO: use transportLayer.TransportFlow().Dst().Raw() to improve perfs
	if transportLayer.TransportFlow().Dst().String() != constants.GTPU_PORT {
		return fmt.Errorf("No GTP-U layer")
	}
	// S02. Pop the outer IPv4 header and UDP/GTP-U headers
	// S03. Copy IPv4 DA and TEID to form SID B
	//teid :=
	//argMobSession := mup.NewArgMobSession(qfi, reflectiveQosIndication, 0, teid)
	//sidB := createSID(destUPFPrefix, dstSlice, argMobSession)
	//IPv6SA := createSID(sourceUPFPrefix, srcSlice)
	// S04. Copy IPv4 SA to form IPv6 SA B'
	// S05. Encapsulate the packet into a new IPv6 header
	// S06. Set the IPv6 DA = B
	// S07. Forward along the shortest path to B

	return nil
}
