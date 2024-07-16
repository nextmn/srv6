// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc

import (
	"encoding/binary"
	"fmt"
	"net/netip"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	json_api "github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/srv6/internal/constants"
	db_api "github.com/nextmn/srv6/internal/database/api"
)

type Packet struct {
	gopacket.Packet
	firstLayerType gopacket.LayerType
}

// (network) LayerType for this packet (LayerTypeIPv4 or LayerTypeIPv6)
func networkLayerType(packet []byte) (*gopacket.LayerType, error) {
	version := (packet[0] >> 4) & 0x0F
	switch version {
	case 4:
		return &layers.LayerTypeIPv4, nil
	case 6:
		return &layers.LayerTypeIPv6, nil
	default:
		return nil, fmt.Errorf("Malformed packet")

	}
}

func NewIPv6Packet(packet []byte) (*Packet, error) {
	if layerType, err := networkLayerType(packet); err != nil {
		return nil, err
	} else if *layerType != layers.LayerTypeIPv6 {
		return nil, fmt.Errorf("This handler can only receive IPv6 packets")
	}
	return &Packet{
		Packet:         gopacket.NewPacket(packet, layers.LayerTypeIPv6, gopacket.Default),
		firstLayerType: layers.LayerTypeIPv6,
	}, nil
}

func NewIPv4Packet(packet []byte) (*Packet, error) {
	if layerType, err := networkLayerType(packet); err != nil {
		return nil, err
	} else if *layerType != layers.LayerTypeIPv4 {
		return nil, fmt.Errorf("This handler can only receive IPv4 packets")
	}
	return &Packet{
		Packet:         gopacket.NewPacket(packet, layers.LayerTypeIPv4, gopacket.Default),
		firstLayerType: layers.LayerTypeIPv4,
	}, nil
}

// Return the packet IP destination address (first network layer) if it is in the prefix range
func (p *Packet) CheckDAInPrefixRange(prefix netip.Prefix) (netip.Addr, error) {
	// get destination address
	dstSlice := p.NetworkLayer().NetworkFlow().Dst().Raw()
	dst, ok := netip.AddrFromSlice(dstSlice)
	if !ok {
		return netip.Addr{}, fmt.Errorf("Malformed packet")
	}
	// check if in range
	if !prefix.Contains(dst) {
		return netip.Addr{}, fmt.Errorf("Destination address out of this handler’s range")
	}
	return dst, nil
}

func (p *Packet) GetSrcAddr() (netip.Addr, error) {
	// get destination address
	srcSlice := p.NetworkLayer().NetworkFlow().Src().Raw()
	src, ok := netip.AddrFromSlice(srcSlice)
	if !ok {
		return netip.Addr{}, fmt.Errorf("Malformed packet")
	}
	return src, nil
}

// Returns the DownlinkAction related to this packet
func (p *Packet) DownlinkAction(db db_api.Downlink) (json_api.Action, error) {
	dstSlice := p.NetworkLayer().NetworkFlow().Dst().Raw()
	dst, ok := netip.AddrFromSlice(dstSlice)
	if !ok {
		return json_api.Action{}, fmt.Errorf("Malformed packet")
	}
	return db.GetDownlinkAction(dst)
}

// Returns the first gopacket.Layer after IPv6 header / extension headers
func (p *Packet) PopIPv6Headers() (gopacket.Layer, error) {
	if p.firstLayerType != layers.LayerTypeIPv6 {
		return nil, fmt.Errorf("Not an IPv6 packet")
	}
	for _, l := range p.Layers()[1:] { // first layer is IPv6 header, we skip it
		if !layers.LayerClassIPv6Extension.Contains(l.LayerType()) {
			return l, nil
		}
	}
	return nil, fmt.Errorf("Nothing else than IPv6 Headers in the packet")
}

// Returns the first gopacket.Layer after IPv4/UDP/GTPU headers
func (p *Packet) PopGTP4Headers() (gopacket.Layer, error) {
	if p.firstLayerType != layers.LayerTypeIPv4 {
		return nil, fmt.Errorf("Not an IPv4 packet")
	}
	if len(p.Layers()) < 4 {
		return nil, fmt.Errorf("Not a GTP4 packet: not enough layers")
	}
	if p.Layers()[1].LayerType() != layers.LayerTypeUDP {
		return nil, fmt.Errorf("No UDP layer")
	}
	if binary.BigEndian.Uint16(p.TransportLayer().TransportFlow().Dst().Raw()) != constants.GTPU_PORT_INT {
		return nil, fmt.Errorf("No GTP-U layer")
	}
	return p.Layers()[3], nil
}
