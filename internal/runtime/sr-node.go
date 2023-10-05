// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

//import (
//	"context"
//	"fmt"
//	"log"
//	"net"
//	"net/netip"
//	"strconv"
//	"strings"
//
//	gopacket "github.com/google/gopacket"
//	layers "github.com/google/gopacket/layers"
//	gopacket_srv6 "github.com/louisroyer/gopacket-srv6"
//	mup "github.com/nextmn/srv6/internal/mup"
//	"github.com/wmnsk/go-gtp/gtpv1"
//)
//
//type SRToGTPNode struct {
//	queue         chan []byte
//	uConn         map[string](*gtpv1.UPlaneConn)
//	sid           string // CIDR
//	done          chan bool
//	netsize       int
//	gtpEntityAddr string // CIDR
//	gtpIPVersion  int
//}
//
//func NewSRToGTPNode(sid string, gtpentityaddr string, gtpIPVersion int) (*SRToGTPNode, error) {
//	if (gtpIPVersion != 4) && (gtpIPVersion != 6) {
//		return nil, fmt.Errorf("gtpIPVersion should be 6 or 4")
//	}
//	netSize, _ := strconv.Atoi(strings.SplitN(sid, "/", 2)[1])
//	if netSize%8 != 0 {
//		return nil, fmt.Errorf("SID networks must be multiple of 8") // FIXME
//	}
//	return &SRToGTPNode{
//		queue:         make(chan []byte),
//		uConn:         make(map[string](*gtpv1.UPlaneConn)),
//		sid:           sid,
//		done:          make(chan bool),
//		netsize:       netSize,
//		gtpEntityAddr: gtpentityaddr,
//		gtpIPVersion:  gtpIPVersion,
//	}, nil
//}
//
//func (s *SRToGTPNode) ListenAndServe() error {
//	log.Printf("SRToGTPNode: sid %s started\n", s.sid)
//	for {
//		select {
//		case packet := <-s.queue:
//			log.Printf("Received a packet on SID %s\n", s.sid)
//			pqt := gopacket.NewPacket(packet, layers.LayerTypeIPv6, gopacket.Default)
//			// extract TEID from destination address
//			// destination address is formed as follow : [ SID (netsize bits) + IPv4 DA (only if ipv4) + ArgsMobSession ]
//			dst := pqt.NetworkLayer().(*layers.IPv6).DstIP.String()
//			ip, err := netip.ParseAddr(dst)
//			if err != nil {
//				return err
//			}
//			dstarray := ip.As16()
//			offset := 0
//			if s.gtpIPVersion == 4 {
//				offset = 32 / 8
//			}
//			// TODO: check segments left = 1, and if not send ICMP Parameter Problem to the Source Address (code 0, pointer to SegemntsLeft field), and drop the packet
//			args, err := mup.ParseArgsMobSession(dstarray[(s.netsize/8)+offset:])
//			if err != nil {
//				return err
//			}
//			teid := args.PDUSessionID()
//			// retrieve nextGTPNode (SHR[0])
//			log.Printf("TEID retreived: %X\n", teid)
//
//			nextGTPNode := ""
//			if s.gtpIPVersion == 6 {
//				// workaround: enforce use of gopacket_srv6 functions
//				shr := gopacket.NewPacket(pqt.Layers()[1].LayerContents(), gopacket_srv6.LayerTypeIPv6Routing, gopacket.Default).Layers()[0].(*gopacket_srv6.IPv6Routing)
//				log.Println("layer type", pqt.Layers()[1].LayerType())
//				log.Println("RoutingType", shr.RoutingType)
//				log.Println("LastEntry:", shr.LastEntry)
//				log.Println("sourceRoutingIPs len:", len(shr.SourceRoutingIPs))
//				log.Println("sourceRoutingIPs[0]:", shr.SourceRoutingIPs[0])
//				nextGTPNode = fmt.Sprintf("[%s]:%s", shr.SourceRoutingIPs[0].String(), GTPU_PORT)
//			} else {
//				// IPv4
//				ip_arr := dstarray[s.netsize/8 : (s.netsize/8)+4]
//				ipv4_address := net.IPv4(ip_arr[0], ip_arr[1], ip_arr[2], ip_arr[3])
//				nextGTPNode = fmt.Sprintf("%s:%s", ipv4_address, GTPU_PORT)
//			}
//			raddr, err := net.ResolveUDPAddr("udp", nextGTPNode)
//			if err != nil {
//				log.Println("Error while resolving ", nextGTPNode, "(remote node)")
//				return nil
//			}
//			// retrieve payload
//			pdu := pqt.Layers()[2].LayerContents() // We expect the packet to contains the following layers [ IPv6 Header (0) + IPv6Routing Ext Header (1) + PDU (2) ]
//			// Search for existing Uconn with this peer and use it
//			if s.uConn[nextGTPNode] == nil {
//				// Start uConn with this peer
//				ch := make(chan bool)
//				go s.StartUconn(ch, nextGTPNode, raddr)
//				_ = <-ch
//			}
//			s.uConn[nextGTPNode].WriteToGTP(teid, pdu, raddr)
//
//		case <-s.done:
//			return nil
//
//		}
//	}
//}
//func (s *SRToGTPNode) StartUconn(ch chan bool, nextGTPNode string, raddr *net.UDPAddr) error {
//	laddrstr := fmt.Sprintf("[%s]:0", strings.SplitN(s.gtpEntityAddr, "/", 2)[0])
//	laddr, err := net.ResolveUDPAddr("udp", laddrstr)
//	if err != nil {
//		log.Println("Error while resolving ", laddrstr, "(local node)")
//		return nil
//	}
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//	uConn, err := gtpv1.DialUPlane(ctx, laddr, raddr)
//	if err != nil {
//		log.Println("Dial Failure", err)
//		return err
//	}
//	defer uConn.Close()
//	s.uConn[nextGTPNode] = uConn
//	close(ch)
//	for {
//		select {}
//	}
//}
//
//func (s *SRToGTPNode) Close() error {
//	close(s.queue)
//	s.done <- true
//	return nil
//}
//
//func (s *SRToGTPNode) Send(pdu []byte) {
//	s.queue <- pdu
//}
