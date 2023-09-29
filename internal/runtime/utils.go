// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

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

func getMTU(iface string) (int64, error) {
	if strings.Contains(iface, "/") || strings.Contains(iface, ".") {
		return 0, fmt.Errorf("interface name contains illegal character")
	}
	filename := fmt.Sprintf("/sys/class/net/%s/mtu", iface)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimRight(string(content), "\n"), 10, 64)
}

func getipv6hoplimit(iface string) (uint8, error) {
	if strings.Contains(iface, "/") || strings.Contains(iface, ".") {
		return 0, fmt.Errorf("interface name contains illegal character")
	}
	filename := fmt.Sprintf("/proc/sys/net/ipv6/conf/%s/hop_limit", iface)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	ret, err := strconv.ParseUint(strings.TrimRight(string(content), "\n"), 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(ret), nil
}
