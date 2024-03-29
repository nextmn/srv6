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

func IPSrSetSourceAddress(address string) error {
	return runIP("sr", "tunsrc", "set", address)
}
