// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"fmt"
	"net"
	"net/netip"

	"database/sql"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"github.com/nextmn/srv6/internal/ctrl"
)

type HeadendGTP4WithCtrl struct {
	RulesRegistry *ctrl.RulesRegistry
	BaseHandler
	db        *sql.DB
	tableName string
}

func NewHeadendGTP4WithCtrl(prefix netip.Prefix, rr *ctrl.RulesRegistry, ttl uint8, hopLimit uint8, db *sql.DB) (*HeadendGTP4WithCtrl, error) {
	tableName := fmt.Sprintf("uplink-%s" + prefix.String())
	s, err := db.Prepare(`CREATE TABLE IF NOT EXISTS $1
		id INT NOT NULL AUTO_INCREMENT,
		uplink_teid INTEGER,
		gnb_ip INET,
		ue_ip_address INET,
		PRIMARY KEY (id);
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare request: %s", err)
	}
	defer s.Close()
	_, err = s.Exec(tableName)
	if err != nil {
		return nil, fmt.Errorf("Could not create table in database: %s", err)
	}

	return &HeadendGTP4WithCtrl{
		RulesRegistry: rr,
		BaseHandler:   NewBaseHandler(prefix, ttl, hopLimit),
		db:            db,
		tableName:     tableName,
	}, nil
}

// Handle a packet
func (h HeadendGTP4WithCtrl) Handle(packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	if err := h.CheckDAInPrefixRange(pqt); err != nil {
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
	//gtpu := layerGTPU.(*layers.GTPv1U)
	//teid := gtpu.TEID

	// TODO: create a dedicated parser for GTPU extension Headers
	// TODO: create a dedicated parser for PDU Session Container
	//var qfi uint8 = 0
	//var reflectiveQosIndication = false
	//if gtpu.ExtensionHeaderFlag && len(gtpu.GTPExtensionHeaders) > 0 {
	// TS 129.281, Fig. 5.2.1-3:
	// > For a GTP-PDU with several Extension Headers, the PDU Session
	// > Container should be the first Extension Header.
	//	firstExt := gtpu.GTPExtensionHeaders[0]
	//	if firstExt.Type == 0x85 { // PDU Session Container
	//		b := firstExt.Content
	//		if (b[0] & 0xF0 >> 4) == 0 { // PDU Type == DL PDU Session Information
	//			qfi = uint8(b[1] & 0x3F)
	//			rqi := b[1] & 0x40 >> 6
	//			if rqi == 0 {
	//				reflectiveQosIndication = true
	//			}
	//		}
	//	}
	//}
	nextHop := net.ParseIP("::") // FIXME: use right ip

	ipheader := &layers.IPv6{
		SrcIP: net.ParseIP("::"),
		// S06. Set the IPv6 DA = B
		DstIP:      nextHop,
		Version:    6,
		NextHeader: layers.IPProtocolIPv6Routing, // IPv6-Route
		HopLimit:   h.HopLimit(),
		// TODO: Generate a FlowLabel with hash(IPv6SA + IPv6DA + policy)
		//TrafficClass: qfi << 2,
		//TrafficClass: 0, // FIXME
	}
	segList := []net.IP{net.ParseIP("::")} //FIXME: fill this array
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
