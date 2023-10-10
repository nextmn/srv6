// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

import (
	"fmt"
	"strings"
)

type Policy struct {
	SegmentsList []string `yaml:"segments-list"`
}

func (p *Policy) ToIPRoute2() string {
	return fmt.Sprintf(strings.Join(p.SegmentsList[:], ","))
}
