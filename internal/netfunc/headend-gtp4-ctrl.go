// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package netfunc

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/netip"

	"github.com/nextmn/srv6/internal/constants"
	db_api "github.com/nextmn/srv6/internal/database/api"

	gopacket_gtp "github.com/nextmn/gopacket-gtp"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"github.com/nextmn/rfc9433/encoding"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type HeadendGTP4WithCtrl struct {
	BaseHandler
	db        db_api.Uplink
	srcPrefix netip.Prefix
}

func NewHeadendGTP4WithCtrl(prefix netip.Prefix, srcPrefix netip.Prefix, ttl uint8, hopLimit uint8, db db_api.Uplink) (*HeadendGTP4WithCtrl, error) {
	return &HeadendGTP4WithCtrl{
		BaseHandler: NewBaseHandler(prefix, ttl, hopLimit),
		db:          db,
		srcPrefix:   srcPrefix,
	}, nil
}

// Handle a packet
func (h HeadendGTP4WithCtrl) Handle(ctx context.Context, packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	dest_addr, err := h.CheckDAInPrefixRange(pqt)
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

	// handle echo request
	if gtpu.MessageType == constants.GTPU_MESSAGE_TYPE_ECHO_REQUEST {
		if !gtpu.SequenceNumberFlag {
			return nil, fmt.Errorf("No sequence number flag in GTP Echo Request")
		}
		ipv4resp := layers.IPv4{
			// IPv4
			Version: 4,
			// Next Header: UDP
			Protocol: layers.IPProtocolUDP,
			// Fragmentation is inefficient and should be avoided (TS 129.281 section 4.2.2)
			// It is recommended to set the default inner MTU size instead.
			Flags: layers.IPv4DontFragment,
			// Destination IP from buffer
			SrcIP: dest_addr.AsSlice(),
			// Source IP from buffer
			DstIP: gnb_ip.AsSlice(),
			// TTL from tun config
			TTL: h.TTL(),
			// other fields are initialized at zero
			// cheksum, and length are computed at serialization
		}
		udpresp := layers.UDP{
			// Source Port
			SrcPort: constants.GTPU_PORT_INT,
			DstPort: pqt.Layer(layers.LayerTypeUDP).(*layers.UDP).SrcPort,
			// cheksum, and length are computed at serialization
		}
		// required for checksum
		udpresp.SetNetworkLayerForChecksum(&ipv4resp)
		gtpresp := gopacket_gtp.GTPv1U{
			Version:            1,
			ProtocolType:       1,
			SequenceNumberFlag: true,
			SequenceNumber:     gtpu.SequenceNumber,
			// message type: G-PDU
			MessageType:   constants.GTPU_MESSAGE_TYPE_ECHO_RESPONSE,
			TEID:          0,
			MessageLength: uint16(6), // recovery IE + seqNum + N-PDU Number (ignored) + next ext header type
		}
		payloadresp := []byte{14, 0} // empty recovery IE
		buf := gopacket.NewSerializeBuffer()
		if err := gopacket.SerializeLayers(buf,
			gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			},
			&ipv4resp,
			&udpresp,
			&gtpresp,
			gopacket.Payload(payloadresp),
		); err != nil {
			return nil, err
		} else {
			return buf.Bytes(), nil
		}

	}

	if gtpu.MessageType != constants.GTPU_MESSAGE_TYPE_GPDU {
		return nil, fmt.Errorf("GTP packet is not a G-PDU")
	}
	// Check payload is IPv4
	inner, ok := payload.(*layers.IPv4)
	if !ok {
		return nil, fmt.Errorf("Payload is not IPv4")
	}
	if inner.Version != 4 {
		return nil, fmt.Errorf("Payload is IPv%d instead of IPv4", inner.Version)
	}
	// Get Inner IPv4 Header Addresses
	innerHeaderSrcIPv4 := netip.AddrFrom4([4]byte{inner.SrcIP[0], inner.SrcIP[1], inner.SrcIP[2], inner.SrcIP[3]})
	innerHeaderDstIPv4 := netip.AddrFrom4([4]byte{inner.DstIP[0], inner.DstIP[1], inner.DstIP[2], inner.DstIP[3]})

	action, err := h.db.GetUplinkAction(ctx, teid, gnb_ip, innerHeaderSrcIPv4, innerHeaderDstIPv4)
	if err != nil {
		return nil, err
	}
	// S04. Copy IPv4 SA to form IPv6 SA B'
	ipv4SA := pqt.NetworkLayer().NetworkFlow().Src().Raw()
	udpSP := pqt.TransportLayer().TransportFlow().Src().Raw()

	ipv6SA := encoding.NewMGTP4IPv6Src(h.srcPrefix, [4]byte(ipv4SA), binary.BigEndian.Uint16(udpSP))

	src, err := ipv6SA.Marshal()
	segs := action.SRH.AsSlice()
	if err != nil {
		return nil, fmt.Errorf("Error during serialization of IPv6 SA: %w", err)
	}
	ipheader := &layers.IPv6{
		SrcIP: src,
		// S06. Set the IPv6 DA = B
		DstIP:      segs[len(segs)-1],
		Version:    6,
		NextHeader: layers.IPProtocolIPv6Routing, // IPv6-Route
		HopLimit:   h.HopLimit(),
		// TODO: Generate a FlowLabel with hash(IPv6SA + IPv6DA + policy)
		//TrafficClass: qfi << 2,
		TrafficClass: 0, // FIXME
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
