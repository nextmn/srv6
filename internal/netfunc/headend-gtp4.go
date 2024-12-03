// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"

	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/constants"

	gopacket_gtp "github.com/nextmn/gopacket-gtp"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"github.com/nextmn/rfc9433/encoding"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type HeadendGTP4 struct {
	policy []config.Policy
	BaseHandler
	sourceAddressPrefix netip.Prefix
}

func NewHeadendGTP4(prefix netip.Prefix, sourceAddressPrefix netip.Prefix, policy []config.Policy, ttl uint8, hopLimit uint8) *HeadendGTP4 {
	return &HeadendGTP4{
		sourceAddressPrefix: sourceAddressPrefix,
		policy:              policy,
		BaseHandler:         NewBaseHandler(prefix, ttl, hopLimit),
	}
}

// Handle a packet
func (h HeadendGTP4) Handle(ctx context.Context, packet []byte) ([]byte, error) {
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

	// TODO: create a dedicated parser for GTPU extension Headers
	// TODO: create a dedicated parser for PDU Session Container
	var qfi uint8 = 0
	var reflectiveQosIndication = false
	if gtpu.ExtensionHeaderFlag && len(gtpu.GTPExtensionHeaders) > 0 {
		// TS 129.281, Fig. 5.2.1-3:
		// > For a GTP-PDU with several Extension Headers, the PDU Session
		// > Container should be the first Extension Header.
		firstExt := gtpu.GTPExtensionHeaders[0]
		if firstExt.Type == 0x85 { // PDU Session Container
			b := firstExt.Content
			if (b[0] & 0xF0 >> 4) == 0 { // PDU Type == DL PDU Session Information
				qfi = uint8(b[1] & 0x3F)
				rqi := b[1] & 0x40 >> 6
				if rqi == 0 {
					reflectiveQosIndication = true
				}
			}
		}
	}
	ipv4DA := pqt.NetworkLayer().NetworkFlow().Dst().Raw()
	argsMobSession := encoding.NewArgsMobSession(qfi, reflectiveQosIndication, false, teid)

	var innerHeaderIPv4 netip.Addr
	isInnerHeaderIPv4 := false

	var bsid *config.Bsid

	// find a policy matching criteria
	for _, p := range h.policy {
		// catch-all policy (should be last policy in list)
		if p.Match == nil {
			bsid = &p.Bsid
			break
		}

		// otherwise, teid is mandatory
		if p.Match.Teid != nil {
			if *p.Match.Teid != teid {
				// teid doesn't match
				continue
			}
			if p.Match.InnerHeaderIPv4SrcPrefix != nil {
				// teid matches, and we need to check the prefix
				if !isInnerHeaderIPv4 {
					// init innerHeaderIPv4
					inner, ok := payload.(*layers.IPv4)
					if !ok {
						return nil, fmt.Errorf("Payload is not IPv4")
					}
					if inner.Version != 4 {
						return nil, fmt.Errorf("Payload is IPv%d instead of IPv4", inner.Version)
					}
					innerHeaderIPv4 = netip.AddrFrom4([4]byte{inner.SrcIP[0], inner.SrcIP[1], inner.SrcIP[2], inner.SrcIP[3]})
					isInnerHeaderIPv4 = true
				}
				prefix, err := netip.ParsePrefix(*p.Match.InnerHeaderIPv4SrcPrefix)
				if err != nil {
					return nil, fmt.Errorf("Malformed matching criteria (inner Header IPv4 Prefix): %s", err)
				}
				if prefix.Contains(innerHeaderIPv4) {
					// prefix matches
					bsid = &p.Bsid
					break
				}
				// prefix doesn't match: continue
			} else {
				// teid matches, and no prefix to check
				bsid = &p.Bsid
				break
			}
		}
	}
	if bsid == nil {
		return nil, fmt.Errorf("Could not found policy matching criteria")
	}

	if bsid.BsidPrefix == nil {
		return nil, fmt.Errorf("Error with policy found")
	}
	dstPrefix, err := netip.ParsePrefix(*bsid.BsidPrefix)
	if err != nil {
		return nil, err
	}
	ipv6DA := encoding.NewMGTP4IPv6Dst(dstPrefix, [4]byte(ipv4DA), argsMobSession)

	// S04. Copy IPv4 SA to form IPv6 SA B'
	ipv4SA := pqt.NetworkLayer().NetworkFlow().Src().Raw()
	udpSP := pqt.TransportLayer().TransportFlow().Src().Raw()

	srcPrefix := h.sourceAddressPrefix
	ipv6SA := encoding.NewMGTP4IPv6Src(srcPrefix, [4]byte(ipv4SA), binary.BigEndian.Uint16(udpSP))
	if err != nil {
		return nil, fmt.Errorf("Error during creation of IPv6 SA: %s", err)
	}

	src, err := ipv6SA.Marshal()
	if err != nil {
		return nil, fmt.Errorf("Error during serialization of IPv6 SA: %s", err)
	}

	seg0, err := ipv6DA.Marshal()
	if err != nil {
		return nil, fmt.Errorf("Error during serialization of Segment[0]: %s", err)
	}
	nextHop := seg0
	if len(bsid.SegmentsList) > 0 {
		nextHop = bsid.ReverseSegmentsList()[0]
	}

	ipheader := &layers.IPv6{
		SrcIP: src,
		// S06. Set the IPv6 DA = B
		DstIP:      nextHop,
		Version:    6,
		NextHeader: layers.IPProtocolIPv6Routing, // IPv6-Route
		HopLimit:   h.HopLimit(),
		// TODO: Generate a FlowLabel with hash(IPv6SA + IPv6DA + policy)
		TrafficClass: qfi << 2,
	}
	segList := append([]net.IP{seg0}, bsid.ReverseSegmentsList()...)
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
