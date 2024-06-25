// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"github.com/nextmn/srv6/internal/config"
	"github.com/nextmn/srv6/internal/mup"
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
func (h HeadendGTP4) Handle(packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	if _, err := h.CheckDAInPrefixRange(pqt); err != nil {
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
	argsMobSession := mup.NewArgsMobSession(qfi, reflectiveQosIndication, false, teid)

	var innerHeaderIPv4 netip.Addr
	isInnerHeaderIPv4 := false

	var bsid *config.Bsid

	// find a policy matching criterias
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
		return nil, fmt.Errorf("Could not found policy matching criterias")
	}

	if bsid.BsidPrefix == nil {
		return nil, fmt.Errorf("Error with policy found")
	}
	dstPrefix, err := netip.ParsePrefix(*bsid.BsidPrefix)
	if err != nil {
		return nil, err
	}
	ipv6DA, err := mup.NewMGTP4IPv6DstFieldsFromFields(dstPrefix, ipv4DA, argsMobSession)
	if err != nil {
		return nil, fmt.Errorf("Error during creation of IPv6 DA: %s", err)
	}

	// S04. Copy IPv4 SA to form IPv6 SA B'
	ipv4SA := pqt.NetworkLayer().NetworkFlow().Src().Raw()
	udpSP := pqt.TransportLayer().TransportFlow().Src().Raw()

	srcPrefix := h.sourceAddressPrefix
	ipv6SA, err := mup.NewMGTP4IPv6SrcFieldsFromFields(srcPrefix, ipv4SA, udpSP)
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
