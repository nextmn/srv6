// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package netfunc

import (
	"context"
	"fmt"
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"github.com/nextmn/rfc9433/encoding"
	db_api "github.com/nextmn/srv6/internal/database/api"
)

type HeadendEncapsWithCtrl struct {
	BaseHandler
	db        db_api.Downlink
	srcPrefix netip.Prefix
}

func NewHeadendEncapsWithCtrl(prefix netip.Prefix, srcPrefix netip.Prefix, ttl uint8, hopLimit uint8, db db_api.Downlink) *HeadendEncapsWithCtrl {
	return &HeadendEncapsWithCtrl{
		BaseHandler: NewBaseHandler(prefix, ttl, hopLimit),
		db:          db,
		srcPrefix:   srcPrefix,
	}
}

// Handle a packet
func (h HeadendEncapsWithCtrl) Handle(ctx context.Context, packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	if _, err := h.CheckDAInPrefixRange(pqt); err != nil {
		return nil, err
	}
	action, err := pqt.DownlinkAction(ctx, h.db)
	if err != nil {
		return nil, err
	}
	if action.SourceGtp4 == nil {
		return nil, fmt.Errorf("Empty SourceGtp4 for downlink Action")
	}
	srgw_gtp_ip := *action.SourceGtp4
	ipv6Src := encoding.NewMGTP4IPv6Src(h.srcPrefix, srgw_gtp_ip.As4(), 2152) // FIXME:dont hardcode udp port number to 2152
	src, err := ipv6Src.Marshal()
	if err != nil {
		return nil, fmt.Errorf("Error during serialization of IPv6 SA: %w", err)
	}

	segs := action.SRH.AsSlice()
	ipheader := &layers.IPv6{
		SrcIP: src,
		// S06. Set the IPv6 DA = B
		DstIP:      segs[len(segs)-1],
		Version:    6,
		NextHeader: layers.IPProtocolIPv6Routing, // IPv6-Route
		HopLimit:   h.HopLimit(),
		// TODO: Generate a FlowLabel with hash(IPv6SA + IPv6DA + policy)
		TrafficClass: 0, // FIXME: put this in Action
	}

	srh := &gopacket_srv6.IPv6Routing{
		RoutingType: 4,
		// the first item on segments list is the next endpoint
		SegmentsLeft:     uint8(len(segs) - 1), // pointer to next segment
		SourceRoutingIPs: segs,
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
}
