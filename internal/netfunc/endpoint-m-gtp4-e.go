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
	gopacket_gtp "github.com/nextmn/gopacket-gtp"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"github.com/nextmn/srv6/internal/constants"
	"github.com/nextmn/srv6/internal/mup"
)

type EndpointMGTP4E struct {
	BaseHandler
}

func NewEndpointMGTP4E(prefix netip.Prefix, ttl uint8, hopLimit uint8) *EndpointMGTP4E {
	return &EndpointMGTP4E{
		BaseHandler: NewBaseHandler(prefix, ttl, hopLimit),
	}
}

// Get IPv6 Destination Address Fields from Packet
func (e EndpointMGTP4E) ipv6DAFields(p *Packet) (*mup.MGTP4IPv6DstFields, error) {
	layerIPv6 := p.Layer(layers.LayerTypeIPv6)
	if layerIPv6 == nil {
		return nil, fmt.Errorf("Malformed IPv6 packet")
	}
	// get destination address
	dstSlice := layerIPv6.(*layers.IPv6).NetworkFlow().Dst().Raw()
	prefix := e.Prefix().Bits()
	if prefix < 0 {
		return nil, fmt.Errorf("Wrong prefix")
	}
	if dst, err := mup.NewMGTP4IPv6DstFields(dstSlice, uint(prefix)); err != nil {
		return nil, err
	} else {
		return dst, nil
	}
}

// Get IPv6 Source Address Fields from Packet
func (e EndpointMGTP4E) ipv6SAFields(p *Packet) (*mup.MGTP4IPv6SrcFields, error) {
	layerIPv6 := p.Layer(layers.LayerTypeIPv6)
	if layerIPv6 == nil {
		return nil, fmt.Errorf("Malformed IPv6 packet")
	}
	// get destination address
	srcSlice := layerIPv6.(*layers.IPv6).NetworkFlow().Src().Raw()
	if src, err := mup.NewMGTP4IPv6SrcFields(srcSlice); err != nil {
		return nil, err
	} else {
		return src, nil
	}
}

// Handle a packet
func (e EndpointMGTP4E) Handle(packet []byte) ([]byte, error) {
	pqt, err := NewIPv6Packet(packet)
	if err != nil {
		return nil, err
	}
	if err := e.CheckDAInPrefixRange(pqt); err != nil {
		return nil, err
	}

	// SRH is optionnal (unless the endpoint is configured to accept only packet with HMAC TLV)
	if layerSRH := pqt.Layer(gopacket_srv6.LayerTypeIPv6Routing); layerSRH != nil {
		srh := layerSRH.(*gopacket_srv6.IPv6Routing)
		// RFC 9433 section 6.6. End.M.GTP4.E
		// S01. When an SRH is processed {
		// S02.   If (Segments Left != 0) {
		// S03.      Send an ICMP Parameter Problem to the Source Address with
		//              Code 0 (Erroneous header field encountered) and
		//              Pointer set to the Segments Left field,
		//              interrupt packet processing, and discard the packet.
		// S04.   }
		if srh.SegmentsLeft != 0 {
			// TODO: Send ICMP response
			return nil, fmt.Errorf("Segments Left is not zero")
		}
		// TODO: check HMAC

		// S05.   Proceed to process the next header in the packet
		// S06. }
	} //TODO: else if HMAC -> error: no SRH

	// S01. Store the IPv6 DA and SA in buffer memory
	ipv6SA, err := e.ipv6SAFields(pqt)
	if err != nil {
		return nil, err
	}
	ipv6DA, err := e.ipv6DAFields(pqt)
	if err != nil {
		return nil, err
	}

	// S02. Pop the IPv6 header and all its extension headers
	payload, err := pqt.PopIPv6Headers()
	if err != nil {
		return nil, err
	}

	// S03. Push a new IPv4 header with a UDP/GTP-U header
	// S04. Set the outer IPv4 SA and DA (from buffer memory)
	// S05. Set the outer Total Length, DSCP, Time To Live, and
	//      Next Header fields
	ipv4 := layers.IPv4{
		// IPv4
		Version: 4,
		// Next Header: UDP
		Protocol: layers.IPProtocolUDP,
		// Fragmentation is inefficient and should be avoided (TS 129.281 section 4.2.2)
		// It is recommended to set the default inner MTU size instead.
		Flags: layers.IPv4DontFragment,
		// Destination IP from buffer
		SrcIP: ipv6SA.IPv4(),
		// Source IP from buffer
		DstIP: ipv6DA.IPv4(),
		// TOS = DSCP + ECN
		// We copy the QFI into the DSCP Field
		TOS: ipv6DA.QFI() << 2,
		// TTL from tun config
		TTL: e.TTL(),
		// other fields are initialized at zero
		// cheksum, and length are computed at serialization

	}

	udp := layers.UDP{
		// Source Port
		SrcPort: ipv6SA.UDPPortNumber(),
		DstPort: constants.GTPU_PORT_INT,
		// cheksum, and length are computed at serialization
	}

	// required for checksum
	udp.SetNetworkLayerForChecksum(&ipv4)

	// S06.    Set the GTP-U TEID (from buffer memory)
	pduSessionContainer := make([]byte, 2) // size should be (n×4-2) octets where n is a positive integer
	// Since End.M.GTP4.E is intended to be used on downlink, we use a DL PDU Session Information Message
	// If you want to use an endpoint behavior for uplink, please create a new one that would use
	// appropriate function arguments.
	// TS 138.415:
	// First byte
	// - [4 bits] PDU Type = 0 (DL PDU Session Information)
	// - [1 bit]  QMP      = 0 (not a QoS Monitoring Packet)
	// - [1 bit]  SNP      = 0 (QFI Sequence Number not present)
	// - [1 bit]  MSNP     = 0 (no MBS Sequence Number Presence)
	// - [1 bit]  Spare
	pduSessionContainer[0] = 0
	// Second byte
	// - [1 bit] PPP = 0 (Paging Policy Indicator not present)
	// - [1 bit] RQI from buffer memory
	// - [6 bits] QFI from buffer memory
	pduSessionContainer[1] = ipv6DA.QFI()
	if ipv6DA.R() {
		pduSessionContainer[1] |= 1 << 6
	}

	gtpExtensionHeaders := make([]gopacket_gtp.GTPExtensionHeader, 1)
	gtpExtensionHeaders[0] = gopacket_gtp.GTPExtensionHeader{
		Type:    0x85, // PDU Session Container
		Content: pduSessionContainer,
	}
	gtpExtensionHeadersLen := 0
	for _, g := range gtpExtensionHeaders {
		gtpExtensionHeadersLen += len(g.Content) + 2 // Type + Length = 2 bytes
	}

	gtpu := gopacket_gtp.GTPv1U{
		// Version should always be set to 1
		Version: 1,
		// TS 128281:
		// > This bit is used as a protocol discriminator between
		// > GTP (when PT is '1') and GTP' (whenPT is '0').
		ProtocolType: 1,
		// We use extension header "PDU Session Container"
		GTPExtensionHeaders: gtpExtensionHeaders,
		// TS 128281:
		// > Since the use of Sequence Numbers is optional for G-PDUs, the PGW,
		// > SGW, ePDG, eNodeB and TWAN should set the flag to '0'.
		SequenceNumberFlag: false,
		// message type: G-PDU
		MessageType: constants.GTPU_MESSAGE_TYPE_GPDU,
		TEID:        ipv6DA.PDUSessionID(),
		// Unfortunatelly, gopacket is not able to compute length at serialization for GTP…
		// We need to do it manually :(
		// TS 128281:
		// > This field indicates the length in octets of the payload, i.e. the rest of the packet following the mandatory
		// > part of the GTP header (that is the first 8 octets). The Sequence Number, the N-PDU Number or any Extension
		// > headers shall be considered to be part of the payload, i.e. included in the length count
		// We need to include
		// - payload length
		// - 4 bytes for :
		//   - Sequence Number (1st Octet): ignored
		//   - Sequence Number (2nd Octet): ignored
		//   - N-PDU Number: ignored
		//   - Next Extension Header Type (at end of the packet: no next header)
		// - total size of gtp extension headers
		MessageLength: uint16(len(payload.LayerContents()) + 4 + gtpExtensionHeadersLen),
	}
	// create buffer for the packet
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf,
		gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		&ipv4,
		&udp,
		&gtpu,
		gopacket.Payload(payload.LayerContents()),
		gopacket.Payload(payload.LayerPayload()),
	); err != nil {
		return nil, err
	} else {
		// S07. Submit the packet to the egress IPv4 FIB lookup for
		//      transmission to the new destination
		return buf.Bytes(), nil
	}
}
