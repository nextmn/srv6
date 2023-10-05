// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ParseConf(file string) (*SRv6Config, error) {
	var conf SRv6Config
	path, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

type SRv6Config struct {
	IPRoute2          *IPRoute2   `yaml:"iproute2"`
	Locator           *string     `yaml:"locator,omitempty"`             // example of locator: fd00:51D5:0000:1::/64
	IPv4HeadendPrefix *string     `yaml:"ipv4-headend-prefix,omitempty"` // example of prefix: 10.0.0.1/32 (if you use a single IPv4 headend) or 10.0.1.0/24 (with more headends)
	HeadEnds          []*HeadEnd  `yaml:"headends"`
	Endpoints         []*Endpoint `yaml:"endpoints"`
}
