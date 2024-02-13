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
	Debug *bool  `yaml:"debug,omitempty"`
	Hooks *Hooks `yaml:"hooks"`

	// interface with controller
	HTTPAddress string  `yaml:"http-address"`
	HTTPPort    *string `yaml:"http-port,omitemty"` // default: 80
	// TODO: use a better type for this information
	ControllerURI string `yaml:controller-uri"` // example: http://192.0.2.2/8080

	// Backbone IPv6 address
	// TODO: use a better type for this information
	BackboneAddress string

	// headends
	LinuxHeadendSetSourceAddress *string  `yaml:"linux-headend-set-source-address,omitempty"`
	GTP4HeadendPrefix            *string  `yaml:"ipv4-headend-prefix,omitempty"` // example of prefix: 10.0.0.1/32 (if you use a single IPv4 headend) or 10.0.1.0/24 (with more headends)
	Headends                     Headends `yaml:"headends"`

	// endpoints
	Locator   *string   `yaml:"locator,omitempty"` // example of locator: fd00:51D5:0000:1::/64
	Endpoints Endpoints `yaml:"endpoints"`
}
