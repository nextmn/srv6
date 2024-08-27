// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package iproute2

import (
	"fmt"
	"os"
	"os/exec"
)

// Run ip command
func runIP(args ...string) error {
	cmd := exec.Command("ip", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running %s: %s", cmd.Args, err)
	}
	return nil
}

// Run iptables command
func runIPTables(args ...string) error {
	cmd := exec.Command("iptables", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running %s: %s", cmd.Args, err)
	}
	return nil
}

// Run ip6tables command
func runIP6Tables(args ...string) error {
	cmd := exec.Command("ip6tables", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error running %s: %s", cmd.Args, err)
	}
	return nil
}

func IPSrSetSourceAddress(address string) error {
	return runIP("sr", "tunsrc", "set", address)
}
