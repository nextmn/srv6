// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"net"
	"net/netip"

	"github.com/nextmn/srv6/internal/ctrl"
)

type HeadendEncapsWithCtrl struct {
	RulesRegistry *ctrl.RulesRegistry
	BaseHandler
}

func NewHeadendEncapsWithCtrl(prefix netip.Prefix, rr *ctrl.RulesRegistry, ttl uint8, hopLimit uint8) *HeadendEncapsWithCtrl {
	return &HeadendEncapsWithCtrl{
		RulesRegistry: rr,
		BaseHandler:   NewBaseHandler(prefix, ttl, hopLimit),
	}
}

// Handle a packet
func (h HeadendEncapsWithCtrl) Handle(packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	if _, err := h.CheckDAInPrefixRange(pqt); err != nil {
		return nil, err
	}
	_, action, err := pqt.Action(h.RulesRegistry)
	if err != nil {
		return nil, err
	}
	src := net.ParseIP("::")            // FIXME: REQUIRED!!! use the right address
	nextHop := action.NextHop.AsSlice() // FIXME: allow multiple segments
	ipheader := &layers.IPv6{
		SrcIP: src,
		// S06. Set the IPv6 DA = B
		DstIP:      nextHop,
		Version:    6,
		NextHeader: layers.IPProtocolIPv6Routing, // IPv6-Route
		HopLimit:   h.HopLimit(),
		// TODO: Generate a FlowLabel with hash(IPv6SA + IPv6DA + policy)
		TrafficClass: 0, // FIXME: put this in Action
	}
	// FIXME: allow multiple segments
	//segList := append([]net.IP{seg0}, bsid.ReverseSegmentsList()...)
	segList := []net.IP{}
	for _, seg := range action.SRH {
		segList = append(segList, seg.AsSlice())
	}

	srh := &gopacket_srv6.IPv6Routing{
		RoutingType: 4,
		// the first item on segments list is the next endpoint
		SegmentsLeft:     uint8(len(segList) - 1), // pointer to next segment
		SourceRoutingIPs: segList,
		Tag:              0, // not used
		Flags:            0, // no flag defined
		GopacketIpv6ExtensionBase: gopacket_srv6.GopacketIpv6ExtensionBase{
			NextHeader: layers.IPProtocolIPv4,
		},
	}

	// Encapsulate the packet into a new IPv6 header
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf,
		gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		ipheader,
		srh,
		gopacket.Payload(pqt.Packet.Layers()[0].LayerContents()),
		gopacket.Payload(pqt.Packet.Layers()[0].LayerPayload()),
	); err != nil {
		return nil, err
	} else {
		// Forward along the shortest path to B
		return buf.Bytes(), nil
	}

	return nil, fmt.Errorf("Not yet implemented")
}
