// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

import "github.com/nextmn/srv6/internal/iana"

type Endpoint struct {
	Provider Provider              `yaml:"provider"` // Linux, NextMN, …
	Sid      string                `yaml:"sid"`      // example of sid: fd00:51D5:0000:1:1:11/80
	Behavior iana.EndpointBehavior `yaml:"behavior"` // example of behavior: End.DX4
	Options  *BehaviorOptions      `yaml:"options,omitempty"`
}

func (el []*Endpoint) Filter(provider Provider) []*Endpoints {
	newList := make([]*Endpoints, 0)
	for _, e := range conf.Endpoints {
		if e.Provider == provider {
			newList := append(newList, e)
		}
	}
	return &newList
}
