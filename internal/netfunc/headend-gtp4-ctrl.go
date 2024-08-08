// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	db_api "github.com/nextmn/srv6/internal/database/api"
	"github.com/nextmn/srv6/internal/mup"
)

type HeadendGTP4WithCtrl struct {
	BaseHandler
	db db_api.Uplink
}

func NewHeadendGTP4WithCtrl(prefix netip.Prefix, ttl uint8, hopLimit uint8, db db_api.Uplink) (*HeadendGTP4WithCtrl, error) {
	return &HeadendGTP4WithCtrl{
		BaseHandler: NewBaseHandler(prefix, ttl, hopLimit),
		db:          db,
	}, nil
}

// Handle a packet
func (h HeadendGTP4WithCtrl) Handle(ctx context.Context, packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	srgw_ip, err := h.CheckDAInPrefixRange(pqt)
	if err != nil {
		return nil, err
	}
	gnb_ip, err := h.GetSrcAddr(pqt)
	if err != nil {
		return nil, err
	}

	// RFC 9433 section 6.7. H.M.GTP4.D

	// S01. IF !(Payload == UDP/GTP-U) THEN Drop the packet
	// S02. Pop the outer IPv4 header and UDP/GTP-U headers
	payload, err := pqt.PopGTP4Headers()
	if err != nil {
		return nil, err
	}
	// S03. Copy IPv4 DA and TEID to form SID B
	layerGTPU := pqt.Layer(layers.LayerTypeGTPv1U)
	if layerGTPU == nil {
		return nil, fmt.Errorf("Could not parse GTPU layer")
	}
	gtpu := layerGTPU.(*layers.GTPv1U)
	teid := gtpu.TEID

	action, err := h.db.GetUplinkAction(ctx, teid, srgw_ip, gnb_ip)
	if err != nil {
		ue_ip_address, ok := netip.AddrFromSlice(gopacket.NewPacket(payload.LayerContents(), layers.LayerTypeIPv4, gopacket.Default).NetworkLayer().NetworkFlow().Src().Raw())
		if !ok {
			return nil, fmt.Errorf("Could not extract ue ip address (not IPv4 in payload?)")
		}
		select {
		case <-ctx.Done(): // TODO: check if err is no row instead of checking ctx
			return nil, ctx.Err()
		default:
			action, err = h.db.SetUplinkAction(ctx, teid, srgw_ip, gnb_ip, ue_ip_address)
		}

	}
	// S04. Copy IPv4 SA to form IPv6 SA B'
	ipv4SA := pqt.NetworkLayer().NetworkFlow().Src().Raw()
	udpSP := pqt.TransportLayer().TransportFlow().Src().Raw()

	srcPrefix := netip.MustParsePrefix("fc00:1:1::/48") // FIXME: dont hardcode
	ipv6SA, err := mup.NewMGTP4IPv6SrcFieldsFromFields(srcPrefix, ipv4SA, udpSP)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of IPv6 SA: %s", err)
	}

	src, err := ipv6SA.Marshal()
	if err != nil {
		return nil, fmt.Errorf("Error during serialization of IPv6 SA: %s", err)
	}
	nextHop := action.NextHop.AsSlice()

	ipheader := &layers.IPv6{
		SrcIP: src,
		// S06. Set the IPv6 DA = B
		DstIP:      nextHop,
		Version:    6,
		NextHeader: layers.IPProtocolIPv6Routing, // IPv6-Route
		HopLimit:   h.HopLimit(),
		// TODO: Generate a FlowLabel with hash(IPv6SA + IPv6DA + policy)
		//TrafficClass: qfi << 2,
		TrafficClass: 0, // FIXME
	}
	segList := []net.IP{}
	for _, seg := range action.SRH {
		segList = append(segList, seg.AsSlice())
	}
	segList = append(segList, nextHop)
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

	// S05. Encapsulate the packet into a new IPv6 header
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf,
		gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		ipheader,
		srh,
		gopacket.Payload(payload.LayerContents()),
		gopacket.Payload(payload.LayerPayload()),
	); err != nil {
		return nil, err
	} else {
		// S07. Forward along the shortest path to B
		return buf.Bytes(), nil
	}
}
