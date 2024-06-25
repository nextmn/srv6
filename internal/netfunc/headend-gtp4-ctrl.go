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

	"database/sql"
	"github.com/gofrs/uuid"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	gopacket_srv6 "github.com/nextmn/gopacket-srv6"
	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/srv6/internal/ctrl"
)

type HeadendGTP4WithCtrl struct {
	RulesRegistry *ctrl.RulesRegistry
	BaseHandler
	db         *sql.DB
	get_action *sql.Stmt
	insert     *sql.Stmt
}

func NewHeadendGTP4WithCtrl(prefix netip.Prefix, rr *ctrl.RulesRegistry, ttl uint8, hopLimit uint8, db *sql.DB) (*HeadendGTP4WithCtrl, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS uplink_gtp4 (
		uplink_teid INTEGER,
		srgw_ip INET,
		action_uuid NOT NULL UUID,
		PRIMARY KEY(uplink_teid, srgw_ip)
		);
	`)
	if err != nil {
		return nil, fmt.Errorf("Could not create table uplink_gtp4 in database: %s", err)
	}

	get_action, err := db.Prepare(`SELECT action_uuid FROM uplink_gtp4 WHERE (uplink_teid = $1 AND srgw_ip = $2)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for get_action: %s", err)
	}

	insert, err := db.Prepare(`INSERT INTO uplink_gtp4 (uplink_teid, srgw_ip, action_uuid) VALUES($1, $2, $3)`)
	if err != nil {
		return nil, fmt.Errorf("Could not prepare statement for insert: %s", err)
	}

	return &HeadendGTP4WithCtrl{
		RulesRegistry: rr,
		BaseHandler:   NewBaseHandler(prefix, ttl, hopLimit),
		db:            db,
		get_action:    get_action,
		insert:        insert,
	}, nil
}

// Handle a packet
func (h HeadendGTP4WithCtrl) Handle(packet []byte) ([]byte, error) {
	pqt, err := NewIPv4Packet(packet)
	if err != nil {
		return nil, err
	}
	srgw_ip, err := h.CheckDAInPrefixRange(pqt)
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

	var action_uuid *uuid.UUID
	h.get_action.QueryRow(teid, srgw_ip).Scan(action_uuid)

	var action jsonapi.Action
	if action_uuid == nil {
		ue_ip_address, ok := netip.AddrFromSlice(gopacket.NewPacket(payload.LayerContents(), layers.LayerTypeIPv4, gopacket.Default).NetworkLayer().NetworkFlow().Src().Raw())
		if !ok {
			return nil, err
		}
		*action_uuid, action, err = h.RulesRegistry.Action(ue_ip_address)
		if err != nil {
			return nil, err
		}
		_, err := h.insert.Exec(teid, srgw_ip, action_uuid)
		if err != nil {
			log.Println("Warning: could not perform insert in headend gtp4 ctrl")
		}

	} else {
		action, err = h.RulesRegistry.ByUUID(*action_uuid)
		if err != nil {
			return nil, err
		}
	}

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
	segList := []net.IP{}
	for _, seg := range action.SRH {
		segList = append(segList, seg.AsSlice())
	}
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
