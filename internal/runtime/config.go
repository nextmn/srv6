// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ParseConf(file string) error {
	path, err := filepath.Abs(file)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &SRv6)
	if err != nil {
		return err
	}
	return nil
}

type IPRoute2 struct {
	// TODO: make 100 and 101 default, but allow to change them
	GTP4RTTableNumber int32 `yaml:"gtp4-rttable-number"` // for example 100
	SRV6RTTableNumber int32 `yaml:"srv6-rttable-number"` // for example 101

	// TODO: make 100 default, but allow to change it
	RTProtoNumber int8 `yaml:"rtproto-number"` // for example 100, max value is 255

	PreInitHook  *string `yaml:"pre-init-hook,omitempty"`  // script to execute before interfaces are configured
	PostInitHook *string `yaml:"post-init-hook,omitempty"` // script to execute after interfaces are configured
}

type BehaviorOptions struct {
	SourceAddress *string `yaml:"set-source-address,omitempty"` // mandatory for End.M.GTP6.(E|D)
}

type Endpoint struct {
	Sid      string           `yaml:"sid"`      // example of sid: fd00:51D5:0000:1:1:11/80
	Behavior string           `yaml:"behavior"` // example of behavior: End.DX4
	Options  *BehaviorOptions `yaml:"options,omitempty"`
}

type SRv6Config struct {
	IPRoute2  *IPRoute2   `yaml:"iproute2"`
	Locator   string      `yaml:"locator` // example of locator: fd00:51D5:0000:1::/64
	Endpoints []*Endpoint `yaml:"endpoints"`
	Policy    *Policy     `yaml:"policy"` // temporary field
}

type Policy struct { // temporary field
	SegmentsList []string `yaml:"segments-list"` // temporary field
}
