// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

type Match struct {
	Teid                     *uint32 `yaml:"teid,omitempty"`
	InnerHeaderIPv4SrcPrefix *string `yaml:"inner-header-ipv4-src-prefix,omitempty"` // e.g. 192.168.0.1/32, Teid must be present
}
