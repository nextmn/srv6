// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package iproute2

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// Run ip command
func runIP(args ...string) error {
	cmd := exec.Command("ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		errLog := fmt.Sprintf("Error running %s: %s", cmd.Args, err)
		log.Println(errLog)
		return err
	}
	return nil
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
