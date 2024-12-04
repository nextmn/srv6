// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package config

import "github.com/nextmn/srv6/internal/iana"

type Endpoint struct {
	Provider Provider              `yaml:"provider"` // Linux, NextMN, …
	Prefix   string                `yaml:"prefix"`   // Prefix = LOC+FUNC example of prefix: fd00:51D5:0000:1:1:11/80
	Behavior iana.EndpointBehavior `yaml:"behavior"` // example of behavior: End.DX4
	Options  *BehaviorOptions      `yaml:"options,omitempty"`
}
type Endpoints []*Endpoint

func (el Endpoints) Filter(provider Provider) Endpoints {
	newList := make(Endpoints, 0)
	for _, e := range el {
		if e.Provider == provider {
			newList = append(newList, e)
		}
	}
	return newList
}
