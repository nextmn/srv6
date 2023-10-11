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
	"github.com/louisroyer/gopacket-srv6"
)

type EndpointMGTP4E struct {
	Handler
}

func NewEndpointMGTP4E(prefix netip.Prefix) *EndpointMGTP4E {
	return &EndpointMGTP4E{
		Handler: NewHandler(prefix),
	}
}

// Handle a packet
func (e *EndpointMGTP4E) Handle(packet []byte) ([]byte, error) {
	layerType, err := networkLayerType(packet)
	if err != nil {
		return nil, err
	}
	if *layerType != layers.LayerTypeIPv6 {
		return nil, fmt.Errorf("Endpoints can only handle IPv6 packets")
	}

	// create gopacket
	pqt := gopacket.NewPacket(packet, *layerType, gopacket.Default)

	// check prefix
	layerIPv6 := pqt.Layer(layers.LayerTypeIPv6)
	if layerIPv6 == nil {
		return nil, fmt.Errorf("Malformed IPv6 packet")
	}
	IPv6 := layerIPv6.(*layers.IPv6)
	dest, ok := netip.AddrFromSlice(IPv6.DstIP.To16())
	if !ok {
		return nil, fmt.Errorf("Malformed IPv6 packet")
	}
	if !e.Prefix().Contains(dest) {
		return nil, fmt.Errorf("Destination address out of this endpoint’s range")
	}
	layerSRH := pqt.Layer(gopacket_srv6.LayerTypeIPv6Routing)
	if layerSRH == nil {
		return nil, fmt.Errorf("No SRH")
	}
	srh := layerSRH.(*gopacket_srv6.IPv6Routing)
	// RFC 9433 section 6.6. End.M.GTP4.E
	// S01. When an SRH is processed {
	// S02.   If (Segments Left != 0) {
	// S03.      Send an ICMP Parameter Problem to the Source Address with
	//              Code 0 (Erroneous header field encountered) and
	//              Pointer set to the Segments Left field,
	//              interrupt packet processing, and discard the packet.
	// S04.   }
	if srh.SegmentsLeft == 0 {
		// TODO: Send ICMP response
		return nil, fmt.Errorf("Segments Left is zero")
	}
	// S05.   Proceed to process the next header in the packet
	// S06. }

	// When processing the Upper-Layer header of a packet matching a FIB
	// entry locally instantiated as an End.M.GTP4.E SID, N does the
	// following:

	// S01. Store the IPv6 DA and SA in buffer memory
	// Note: IPv6 DA is in variable `dest`
	source, ok := netip.AddrFromSlice(IPv6.SrcIP.To16())
	if !ok {
		return nil, fmt.Errorf("Malformed IPv6 packet")
	}
	// S02. Pop the IPv6 header and all its extension headers
	payload, err := popIPv6Headers(pqt)
	if err != nil {
		return err
	}
	// S03. Push a new IPv4 header with a UDP/GTP-U header
	ipv4 := layers.IPv4{
		Version: 4,
		// Fragmentation is inefficient and should be avoided (TS 129.281 section 4.2.2)
		// It is recommended to set the default inner MTU size instead.
		Flags:    layers.IPv4DontFragment,
		Id:       0, // IPv4 ID field can only be used for fragmentation (RFC 6864), and has no meaning with atomic datagrams
		Protocol: layers.IPProtocolUDP,
		Checksum: 0,                            // computed at serialization
		Options:  make([]layers.IPv4Option, 0), // no option

		// S04. Set the outer IPv4 SA and DA (from buffer memory)
		//SrcIP:
		//DstIP:

		// S05. Set the outer Total Length, DSCP, Time To Live, and
		//      Next Header fields
		//	TOS: 0, //TODO: from QFI
		//TTL: // TODO: from tun config
		//	IHL:    0, // computed at serialization
		//	Length: 0, // computed at serialization

	}

	// S06.    Set the GTP-U TEID (from buffer memory)
	//gtpu := layers.GTPv1U
	// create buffer for the packet
	//buf := gopacket.NewSerializeBuffer()
	// initialize buffer with the payload
	// Initial content of the buffer : [ ]
	// Updated content of the buffer : [ PDU ]
	//err = gopacket.Payload(pdu).SerializeTo(buf, gopacket.SerializeOptions{
	//	FixLengths:       true,
	//	ComputeChecksums: true,
	//})

	// S07.    Submit the packet to the egress IPv4 FIB lookup for
	//            transmission to the new destination

	// extract TEID from destination address
	// destination address is formed as follow : [ SID (netsize bits) + IPv4 DA (only if ipv4) + ArgsMobSession ]
	//	dstarray := dst.As16()
	//	offset := 0
	//	if s.gtpIPVersion == 4 {
	//		offset = 32 / 8
	//	}
	// TODO: check segments left = 1, and if not send ICMP Parameter Problem to the Source Address (code 0, pointer to SegemntsLeft field), and drop the packet
	//	args, err := mup.ParseArgsMobSession(dstarray[(s.netsize/8)+offset:])
	//	if err != nil {
	//		return err
	//	}
	//	teid := args.PDUSessionID()
	// retrieve nextGTPNode (SHR[0])

	//	nextGTPNode := ""
	//	if s.gtpIPVersion == 6 {
	//		// workaround: enforce use of gopacket_srv6 functions
	//		shr := gopacket.NewPacket(pqt.Layers()[1].LayerContents(), gopacket_srv6.LayerTypeIPv6Routing, gopacket.Default).Layers()[0].(*gopacket_srv6.IPv6Routing)
	//		log.Println("layer type", pqt.Layers()[1].LayerType())
	//		log.Println("RoutingType", shr.RoutingType)
	//		log.Println("LastEntry:", shr.LastEntry)
	//		log.Println("sourceRoutingIPs len:", len(shr.SourceRoutingIPs))
	//		log.Println("sourceRoutingIPs[0]:", shr.SourceRoutingIPs[0])
	//		nextGTPNode = fmt.Sprintf("[%s]:%s", shr.SourceRoutingIPs[0].String(), GTPU_PORT)
	//	} else {
	//		// IPv4
	//		ip_arr := dstarray[s.netsize/8 : (s.netsize/8)+4]
	//		ipv4_address := net.IPv4(ip_arr[0], ip_arr[1], ip_arr[2], ip_arr[3])
	//		nextGTPNode = fmt.Sprintf("%s:%s", ipv4_address, GTPU_PORT)
	//	}
	//	raddr, err := net.ResolveUDPAddr("udp", nextGTPNode)
	//	if err != nil {
	//		log.Println("Error while resolving ", nextGTPNode, "(remote node)")
	//		return nil
	//	}
	// retrieve payload
	//			pdu := pqt.Layers()[2].LayerContents() // We expect the packet to contains the following layers [ IPv6 Header (0) + IPv6Routing Ext Header (1) + PDU (2) ]
	//			// Search for existing Uconn with this peer and use it
	//			if s.uConn[nextGTPNode] == nil {
	//				// Start uConn with this peer
	//				ch := make(chan bool)
	//				go s.StartUconn(ch, nextGTPNode, raddr)
	//				_ = <-ch
	//			}
	//			s.uConn[nextGTPNode].WriteToGTP(teid, pdu, raddr)

	// create gopacket
	return nil, fmt.Errorf("TODO")
}
