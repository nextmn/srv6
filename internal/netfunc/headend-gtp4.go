// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"encoding/binary"
	"fmt"
	"net/netip"

	"github.com/google/gopacket/layers"
	"github.com/nextmn/srv6/internal/constants"
)

type HeadendGTP4 struct {
	BaseHandler
}

func NewHeadendGTP4(prefix netip.Prefix, ttl uint8, hopLimit uint8) *HeadendGTP4 {
	return &HeadendGTP4{
		BaseHandler: NewBaseHandler(prefix, ttl, hopLimit),
	}
}

// Handle a packet
func (h HeadendGTP4) Handle(packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	if err := h.CheckDAInPrefixRange(pqt); err != nil {
		return nil, err
	}

	// RFC 9433 section 6.7. H.M.GTP4.D

	// S01. IF !(Payload == UDP/GTP-U) THEN Drop the packet
	transportLayer := pqt.TransportLayer()
	if transportLayer.LayerType() != layers.LayerTypeUDP {
		return nil, fmt.Errorf("No UDP layer")
	}
	if binary.BigEndian.Uint16(transportLayer.TransportFlow().Dst().Raw()) != constants.GTPU_PORT_INT {
		return nil, fmt.Errorf("No GTP-U layer")
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

	return nil, fmt.Errorf("TODO")
}
