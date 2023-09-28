// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/netip"
	"strconv"
	"strings"

	gopacket "github.com/google/gopacket"
	layers "github.com/google/gopacket/layers"
	gopacket_srv6 "github.com/louisroyer/gopacket-srv6"
	"github.com/songgao/water"
	waterutil "github.com/songgao/water/waterutil"
	gtpv1 "github.com/wmnsk/go-gtp/gtpv1"
	message "github.com/wmnsk/go-gtp/gtpv1/message"
)

var nextmnSR *water.Interface
var srNodes map[string](*SRToGTPNode)
var gtpNodes map[string](*gtpv1.UPlaneConn)

func goSRInit() error {
	// if at least one custom behavior, create tun
	nextmnSR, err := createTun(NextmnSRTunName)
	if err != nil {
		return err
	}
	// create endpoints
	if err := createEndpoints(nextmnSR); err != nil {
		return err
	}
	// listen on GTP Nodes if any endpoint use them
	if err := listenOnGTP(); err != nil {
		return err
	}
	// listen on SR TUN if any endpoint use it
	if err := listenOnTun(nextmnSR); err != nil {
		return err
	}
	return nil
}

func goSRExit() error {
	// stop the goroutine listening on SR TUN
	if err := stopListeningOnTun(nextmnSR); err != nil {
		return err
	}
	// remove SR TUNs
	if err := removeTun(nextmnSR); err != nil {
		return err
	}
	// stop GTP Nodes
	if err := stopListeningOnGTP(); err != nil {
		return err
	}
	return nil
}

func createGTPUEntity(ipAddress string, ipversion int) (*gtpv1.UPlaneConn, error) {
	// TODO: replace the GTP entity with a creation of GTP packet by gopacket (don't forget response to GTP Echo (option))
	switch ipversion {
	case 4:
		if !strings.HasSuffix(ipAddress, "/32") {
			return nil, fmt.Errorf("IP Address for this GTP Entity must be an IPv4 address with a /32 mask")
		}
		// we don't add ip address in ipv4 case because the address is expected to already exist
	case 6:
		if !strings.HasSuffix(ipAddress, "/128") {
			return nil, fmt.Errorf("IP Address for this GTP Entity must be an IPv6 address with a /128 mask")
		}
		// add ip address (padded with zeros) to iface LinuxSRLinkName
		// (we expect to have no address set in the ip range of the SID)
		if err := runIP("-6", "address", "replace", ipAddress, "dev", LinuxSRLinkName); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("IP version should be 4 or 6")
	}

	ipAddressNoMask := strings.SplitN(ipAddress, "/", 2)[0]

	// Prepare a GTPU Entity
	var udpaddr string
	if strings.Count(ipAddressNoMask, ":") > 0 {
		udpaddr = fmt.Sprintf("[%s]:%s", ipAddressNoMask, GTPU_PORT)
	} else {
		udpaddr = fmt.Sprintf("%s:%s", ipAddressNoMask, GTPU_PORT)
	}
	laddr, err := net.ResolveUDPAddr("udp", udpaddr)
	if err != nil {
		fmt.Println("Error while resolving UDP Address of local GTP entity")
		return nil, err
	}
	return gtpv1.NewUPlaneConn(laddr), nil
}

func segmentsList(list []string) []net.IP {
	res := make([]net.IP, 0)
	for _, i := range list {
		ip := net.ParseIP(i)
		res = append(res, ip)
	}
	return res
}

func segmentsLeft(list []string) uint8 {
	return uint8(len(list))
}

func tpduHandler(iface *water.Interface, srSourceAddr net.IP, c gtpv1.Conn, senderAddr net.Addr, msg message.Message, hoplimit uint8) error {
	// We have received a GTP packet, and we need to forward the payload using SRv6
	gtppacket := make([]byte, msg.MarshalLen())
	err := msg.MarshalTo(gtppacket)
	if err != nil {
		log.Println("Could not Marshal GTP Packet")
		return err
	}
	// to separate the payload from header, we need to create a message.Header
	var h message.Header
	err = h.UnmarshalBinary(gtppacket)
	if err != nil {
		log.Println("Could not UnMarshal GTP Packet")
		return err
	}
	pdu := h.Payload
	// create a SHR depending on the a policy
	ipheader := &layers.IPv6{
		SrcIP:      srSourceAddr,
		DstIP:      net.ParseIP(SRv6.Policy.SegmentsList[0]),
		Version:    6,
		NextHeader: 43, // IPv6-Route
		HopLimit:   hoplimit,
	}
	srh := &gopacket_srv6.IPv6Routing{
		RoutingType: 4,
		// the first item on segments list is the next endpoint
		SegmentsLeft:     segmentsLeft(SRv6.Policy.SegmentsList[1:]),
		SourceRoutingIPs: segmentsList(SRv6.Policy.SegmentsList[1:]),
		Tag:              0, // not used
		Flags:            0, // no flag defined
	}
	log.Println("Size of full policy:", len(segmentsList(SRv6.Policy.SegmentsList)))
	log.Println("Size of -1 policy:", len(segmentsList(SRv6.Policy.SegmentsList[1:])))
	// FIXME: allow creation of IPv4, IPv6, Ethernet, and Unstructured GTP Endpoints
	// We only implement IPv4v6 Endpoint.
	// Configure your SMF wisely! (it should establish PFCP Session with PDU Session Type = IPv4v6)
	// (IPv4 or IPv6 will also work, even if the following code is not optimal for this)
	if waterutil.IsIPv6(pdu) {
		srh.NextHeader = 41 // IPv6
	} else if waterutil.IsIPv4(pdu) {
		srh.NextHeader = 4 // IPv4
	} else {
		return fmt.Errorf("Only IPv4v6 PDUSession Type is supported")
	}

	// create buffer for the packet
	buf := gopacket.NewSerializeBuffer()
	// initialize buffer with the payload
	// Initial content of the buffer : [ ]
	// Updated content of the buffer : [ PDU ]
	err = gopacket.Payload(pdu).SerializeTo(buf, gopacket.SerializeOptions{
		FixLengths:       false,
		ComputeChecksums: false,
	})
	if err != nil {
		log.Println("Could not Serialize GTP Packet Payload into gopacket")
		return err
	}
	// lenght of outer header is computed automatically
	opts := gopacket.SerializeOptions{
		FixLengths:       true,  // to set LastEntry automatically
		ComputeChecksums: false, // no checksum in ipv6 headers
	}
	// prepend the SRH
	// Initial content of the buffer : [ PDU ]
	// Updated content of the buffer : [ SRH, PDU ]
	err = srh.SerializeTo(buf, opts)
	if err != nil {
		log.Println("Could not Serialize SRH into gopacket")
		return err
	}
	// prepend the IPv6 header
	// Initial content of the buffer : [ SRH, PDU ]
	// Updated content of the buffer : [ IPv6 Header, SRH, PDU ]
	err = ipheader.SerializeTo(buf, opts)
	if err != nil {
		log.Println("Could not Serialize IPv6 Header into gopacket")
		return err
	}
	srv6packet := buf.Bytes()
	// send the resulting packet on iface NextmnSRTunName
	iface.Write(srv6packet)
	return nil
}

func createEndpoints(iface *water.Interface) error {
	srNodes = make(map[string](*SRToGTPNode))
	gtpNodes = make(map[string](*gtpv1.UPlaneConn))
	for _, e := range SRv6.Endpoints {
		switch e.Behavior {
		case "End.MAP":
			return fmt.Errorf("Not implemented")
		case "End.M.GTP6.D": // we receive GTP packet and send SRv6 packets with ArgsMobSession stored in arguments of SRH[0]
			return fmt.Errorf("Not implemented")
		case "End.M.GTP6.D.Di": // we receive GTP packet and send SRv6 packets, no ArgsMobSession is stored
			if e.Options == nil || e.Options.SourceAddress == nil {
				return fmt.Errorf("Options field must contain a set-source-address parameter")
				// TODO: after replacement of GTPU-Entity creation by gopacket, this parameter should become optional (default: dst addr of the received packet)
			}
			if !strings.HasSuffix(e.Sid, "/128") {
				return fmt.Errorf("SID of End.M.GTP6.Di must be a /128")
			}
			// FIXME: canonize gtpentityAddr
			gtpentityAddr := e.Sid                                                     // we receive GTP packets having this destination address
			srAddr := net.ParseIP(strings.SplitN(*e.Options.SourceAddress, "/", 2)[0]) // we send SR packets using this source address
			// add a GTP Node to be able to receive GTP packets
			if gtpNodes[gtpentityAddr] == nil {
				entity, err := createGTPUEntity(gtpentityAddr, 6)
				if err != nil {
					return err
				}
				gtpNodes[gtpentityAddr] = entity
			}
			// hop limit is set at start of the server, to avoid reading it at each packet reception
			hoplimit, err := getipv6hoplimit(NextmnSRTunName)
			if err != nil {
				return err
			}

			// add handler that will allow GTP decap & SR encap
			gtpNodes[gtpentityAddr].AddHandler(message.MsgTypeTPDU, func(c gtpv1.Conn, senderAddr net.Addr, msg message.Message) error {
				return tpduHandler(iface, srAddr, c, senderAddr, msg, hoplimit)
			})
		case "End.M.GTP6.E": // we receive SRv6 packets and send GTP6 packets
			if e.Options == nil || e.Options.SourceAddress == nil {
				// TODO: after replacement of GTPU-Entity creation by gopacket, this parameter should become optional (default: dst addr of the received packet)
				return fmt.Errorf("Options field must contain a set-source-address parameter")
			}
			if !strings.HasSuffix(*e.Options.SourceAddress, "/128") {
				return fmt.Errorf("set-source-address parameter of End.M.GTP6.E must be explicitly a /128 address")
			}
			if err := runIP("-6", "route", "add", e.Sid, "dev", NextmnSRTunName, "table", RTTableName, "proto", RTProtoName); err != nil {
				return err
			}
			maxNetSize := 128 - (8 + 8*4) // [ SID + QFI + R + U + TEID ]
			netSize, err := strconv.Atoi(strings.SplitN(e.Sid, "/", 2)[1])

			if err != nil {
				return err
			}
			if netSize > maxNetSize {
				return fmt.Errorf("Maximum network size for SID is /%d", maxNetSize)
			}
			if netSize%8 != 0 {
				return fmt.Errorf("Network size for SID must be multiple of 8") // FIXME: handle bit shifts
			}
			// FIXME: canonize srAddr
			srAddr := e.Sid                           // we receive SR packets having this destination address
			gtpentityAddr := *e.Options.SourceAddress // we send GTP packets using this source address
			// add a GTP Node to be able to respond to GTP Echo Requests
			if gtpNodes[gtpentityAddr] == nil {
				entity, err := createGTPUEntity(gtpentityAddr, 6)
				if err != nil {
					return err
				}
				gtpNodes[gtpentityAddr] = entity
			}
			// create SRToGTPNode
			n, err := NewSRToGTPNode(srAddr, gtpentityAddr, 6)
			if err != nil {
				return err
			} else {
				srNodes[srAddr] = n
			}
		case "End.M.GTP4.E": // we receive SRv6 packets and send GTP4 packets
			if e.Options == nil || e.Options.SourceAddress == nil {
				// TODO: after replacement of GTPU-Entity creation by gopacket, check the IPv4 source address from IPv6 dest addr argument space
				return fmt.Errorf("Options field must contain a set-source-address parameter")
			}
			if !strings.HasSuffix(*e.Options.SourceAddress, "/32") {
				return fmt.Errorf("set-source-address parameter of End.M.GTP4.E must be explicitly a /32 address")
			}
			if err := runIP("-6", "route", "add", e.Sid, "dev", NextmnSRTunName, "table", RTTableName, "proto", RTProtoName); err != nil {
				return err
			}
			maxNetSize := 128 - (32 + 8 + 8*4) // [ SID + IPv4 DA + QFI + R + U + TEID ]
			netSize, err := strconv.Atoi(strings.SplitN(e.Sid, "/", 2)[1])

			if err != nil {
				return err
			}
			if netSize > maxNetSize {
				return fmt.Errorf("Maximum network size for SID is /%d", maxNetSize)
			}
			if netSize%8 != 0 {
				return fmt.Errorf("Network size for SID must be multiple of 8") // FIXME: handle bit shifts
			}
			srAddr := e.Sid                           // we receive SR packets having this destination address
			gtpentityAddr := *e.Options.SourceAddress // we send GTP packets using this source address
			// add a GTP Node to be able to respond to GTP Echo Requests
			if gtpNodes[gtpentityAddr] == nil {
				entity, err := createGTPUEntity(gtpentityAddr, 4)
				if err != nil {
					return err
				}
				gtpNodes[gtpentityAddr] = entity
			}
			// create SRToGTPNode
			n, err := NewSRToGTPNode(srAddr, gtpentityAddr, 4)
			if err != nil {
				return err
			} else {
				srNodes[srAddr] = n
			}
		case "H.M.GTP4.D":
			if e.Options == nil || e.Options.SourceAddress == nil {
				// TODO: this parameter should be optional (default: sid + dst addr of the received packet)
				return fmt.Errorf("Options field must contain a set-source-address parameter")
			}
			if !strings.HasSuffix(e.Sid, "/32") {
				return fmt.Errorf("SID of H.GTP4.D must be a /32")
			}
			gtpentityAddr := e.Sid                                                     // we receive GTP packets having this destination address
			srAddr := net.ParseIP(strings.SplitN(*e.Options.SourceAddress, "/", 2)[0]) // we send SR packets using this source address
			// add a GTP Node to be able to receive GTP packets
			if gtpNodes[gtpentityAddr] == nil {
				entity, err := createGTPUEntity(gtpentityAddr, 4)
				if err != nil {
					return err
				}
				gtpNodes[gtpentityAddr] = entity
			}
			// hop limit is set at start of the server, to avoid reading it at each packet reception
			hoplimit, err := getipv6hoplimit(NextmnSRTunName)
			if err != nil {
				return err
			}

			// add handler that will allow GTP decap & SR encap
			gtpNodes[gtpentityAddr].AddHandler(message.MsgTypeTPDU, func(c gtpv1.Conn, senderAddr net.Addr, msg message.Message) error {
				return tpduHandler(iface, srAddr, c, senderAddr, msg, hoplimit)
			})
		case "End.Limit":
			return fmt.Errorf("Not implemented")
		default:
			// pass: other Behaviors can be implemented on linux side (see linux-sr.go)
		}
	}
	return nil
}

func listenOnGTP() error {
	// start created gtpNodes
	ctx, _ := context.WithCancel(context.Background())
	for key, _ := range gtpNodes {
		go gtpNodes[key].ListenAndServe(ctx)
	}
	return nil
}

func stopListeningOnGTP() error {
	for key, _ := range gtpNodes {
		gtpNodes[key].Close()
		gtpNodes[key] = nil
	}
	return nil
}

func listenOnTun(iface *water.Interface) error {
	mtu, err := getMTU(iface.Name())
	if err != nil {
		return err
	}
	for key, _ := range srNodes {
		go srNodes[key].ListenAndServe()
	}
	go func(iface *water.Interface, ifacemtu int64) error {
		for {
			packet := make([]byte, ifacemtu)
			log.Println("Waiting for next packet")
			n, err := iface.Read(packet)
			log.Println("New packet read")
			if err != nil {
				log.Println("Error while reading on iface nextmnSR")
				return err
			}
			go handleSR(packet[:n])
		}
	}(iface, mtu)
	return nil
}

func stopListeningOnTun(iface *water.Interface) error {
	for key, _ := range srNodes {
		srNodes[key].Close()
		srNodes[key] = nil
	}
	iface.Close()
	return nil
}

func handleSR(packet []byte) error {
	if !waterutil.IsIPv6(packet) {
		// SID can only be an IPv6 packet
		log.Println("Received non-IPv6 packet: dropping")
		return nil
	}
	log.Println("Received an IPv6 packet")
	dst := gopacket.NewPacket(packet, layers.LayerTypeIPv6, gopacket.Default).NetworkLayer().(*layers.IPv6).DstIP.String()
	for iprange, _ := range srNodes {
		network, err := netip.ParsePrefix(iprange)
		if err != nil {
			log.Println("Parsing error")
			return err
		}
		ip, err := netip.ParseAddr(dst)
		if err != nil {
			log.Println("Parsing error")
			return err
		}
		if network.Contains(ip) {
			// forwarding to the SR Node queue
			log.Println("Received IPv6 packet matching a SID: forwarding to", iprange)
			srNodes[iprange].Send(packet)
			return nil // avoid duplication of packet
		}
	}
	// no SID is matched
	log.Println("Received IPv6 packet matching no SID: dropping")
	return nil
}

func createTun(ifaceName string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = ifaceName
	iface, err := water.New(config)
	if err != nil {
		log.Println("Unable to allocate TUN interface:", err)
		return nil, err
	}
	if err = runIP("link", "set", "dev", iface.Name(), "up"); err != nil {
		log.Println("Unable to set", iface.Name(), "up")
		return nil, err
	}
	return iface, nil
}

func removeTun(iface *water.Interface) error {
	if iface == nil {
		return nil
	}
	if err := runIP("link", "del", iface.Name()); err != nil {
		log.Println("Unable to delete interface", iface.Name(), ":", err)
		return err
	}
	return nil
}
