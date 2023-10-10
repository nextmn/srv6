// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"log"
	"net"
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/nextmn/srv6/internal/mup"
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
func (h *EndpointMGTP4E) Handle(packet []byte) error {
	layerType, err := networkLayerType(packet)
	if err != nil {
		return err
	}
	if *layerType != layers.LayerTypeIPv6 {
		return fmt.Errorf("Endpoints can only handle IPv6 packets")
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
	// extract TEID from destination address
	// destination address is formed as follow : [ SID (netsize bits) + IPv4 DA (only if ipv4) + ArgsMobSession ]
	dst := pqt.NetworkLayer().(*layers.IPv6).DstIP.String()
	ip, err := netip.ParseAddr(dst)
	if err != nil {
		return err
	}
	dstarray := ip.As16()
	offset := 0
	if s.gtpIPVersion == 4 {
		offset = 32 / 8
	}
	// TODO: check segments left = 1, and if not send ICMP Parameter Problem to the Source Address (code 0, pointer to SegemntsLeft field), and drop the packet
	args, err := mup.ParseArgsMobSession(dstarray[(s.netsize/8)+offset:])
	if err != nil {
		return err
	}
	teid := args.PDUSessionID()
	// retrieve nextGTPNode (SHR[0])
	log.Printf("TEID retreived: %X\n", teid)

	nextGTPNode := ""
	if s.gtpIPVersion == 6 {
		// workaround: enforce use of gopacket_srv6 functions
		shr := gopacket.NewPacket(pqt.Layers()[1].LayerContents(), gopacket_srv6.LayerTypeIPv6Routing, gopacket.Default).Layers()[0].(*gopacket_srv6.IPv6Routing)
		log.Println("layer type", pqt.Layers()[1].LayerType())
		log.Println("RoutingType", shr.RoutingType)
		log.Println("LastEntry:", shr.LastEntry)
		log.Println("sourceRoutingIPs len:", len(shr.SourceRoutingIPs))
		log.Println("sourceRoutingIPs[0]:", shr.SourceRoutingIPs[0])
		nextGTPNode = fmt.Sprintf("[%s]:%s", shr.SourceRoutingIPs[0].String(), GTPU_PORT)
	} else {
		// IPv4
		ip_arr := dstarray[s.netsize/8 : (s.netsize/8)+4]
		ipv4_address := net.IPv4(ip_arr[0], ip_arr[1], ip_arr[2], ip_arr[3])
		nextGTPNode = fmt.Sprintf("%s:%s", ipv4_address, GTPU_PORT)
	}
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

	// RFC 9433 section 6.6. End.M.GTP4.E

	// S01. When an SRH is processed {
	// S02.   If (Segments Left != 0) {
	// S03.      Send an ICMP Parameter Problem to the Source Address with
	//              Code 0 (Erroneous header field encountered) and
	//              Pointer set to the Segments Left field,
	//              interrupt packet processing, and discard the packet.
	// S04.   }
	// S05.   Proceed to process the next header in the packet
	// S06. }

	// When processing the Upper-Layer header of a packet matching a FIB
	// entry locally instantiated as an End.M.GTP4.E SID, N does the
	// following:

	// S01.    Store the IPv6 DA and SA in buffer memory
	// S02.    Pop the IPv6 header and all its extension headers
	// S03.    Push a new IPv4 header with a UDP/GTP-U header
	// S04.    Set the outer IPv4 SA and DA (from buffer memory)
	// S05.    Set the outer Total Length, DSCP, Time To Live, and
	//            Next Header fields
	// S06.    Set the GTP-U TEID (from buffer memory)
	// S07.    Submit the packet to the egress IPv4 FIB lookup for
	//            transmission to the new destination
	return nil
}
