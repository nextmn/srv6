// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"net"
	"strings"
)

type Bsid struct {
	BsidPrefix   *string  `yaml:"bsid-prefix,omitempty"`
	SegmentsList []string `yaml:"segments-list"`
}

func (a *Bsid) ToIPRoute2() string {
	return fmt.Sprintf(strings.Join(a.SegmentsList[:], ","))
}

func (a *Bsid) ReverseSegmentsList() []net.IP {
	res := []net.IP{}
	for i := len(a.SegmentsList) - 1; i >= 0; i-- {
		ip := net.ParseIP(a.SegmentsList[i])
		res = append(res, ip)
	}
	return res
}
