// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Provider string

const (
	ProviderLinux      = "Linux"
	ProviderNextMNSRv6 = "NextMN-SRv6"
	ProviderNextMNGTP4 = "NextMN-GTP4"
)

func ParseConf(file string) (*SRv6Config, error) {
	var conf SRv6Config
	path, err := filepath.Abs(file)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(yamlFile, &conf); err != nil {
		return nil, err
	}
	for _, e := range conf.Endpoints {
		if err := e.InitDefaultProvider(); err != nil {
			return nil, err
		}
	}
	return conf, nil
}

type IPRoute2 struct {
	PreInitHook  *string `yaml:"pre-init-hook,omitempty"`  // script to execute before interfaces are configured
	PostInitHook *string `yaml:"post-init-hook,omitempty"` // script to execute after interfaces are configured
}

type BehaviorOptions struct {
	SourceAddress *string `yaml:"set-source-address,omitempty"` // mandatory for End.M.GTP6.(E|D)
}

type Endpoint struct {
	Provider *Provider        `yaml:"provider,omitempty"` // Linux, NextMN, …
	Sid      string           `yaml:"sid"`                // example of sid: fd00:51D5:0000:1:1:11/80
	Behavior string           `yaml:"behavior"`           // example of behavior: End.DX4
	Options  *BehaviorOptions `yaml:"options,omitempty"`
}

func (e Endpoint) InitDefaultProvider() error {
	if e.Provider != nil {
		if (e.Provider != ProviderLinux) && (e.Provider != ProviderNextMNSRv6)(e.Provider != ProviderNextMNGTP4) {
			return fmt.Errorf("Unknow provider for Endpoint %s: %s", e.Sid, e.Provider)
		}
		return nil
	}
	switch Behavior {
	case "End.MAP":
		e.Provider = ProviderNextMNSRv6
	case "End.M.GTP6.D":
		e.Provider = ProviderNextMNSRv6
	case "End.M.GTP6.D.Di":
		e.Provider = ProviderNextMNSRv6
	case "End.M.GTP6.E":
		e.Provider = ProviderNextMNSRv6
	case "End.M.GTP4.E":
		e.Provider = ProviderNextMNSRv6
	case "H.M.GTP4.D":
		e.Provider = ProviderNextMNGTP4
	case "End.Limit":
		e.Provider = ProviderNextMNSRv6
	default:
		e.Provider = ProviderLinux

	}
	return nil
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

type SRv6Config struct {
	IPRoute2          *IPRoute2   `yaml:"iproute2"`
	Locator           *string     `yaml:"locator,omitempty"`             // example of locator: fd00:51D5:0000:1::/64
	IPv4HeadendPrefix *string     `yaml:"ipv4-headend-prefix,omitempty"` // example of prefix: 10.0.0.1/32 (if you use a single IPv4 headend) or 10.0.1.0/24 (with more headends)
	Endpoints         []*Endpoint `yaml:"endpoints"`
	Policy            *Policy     `yaml:"policy"` // temporary field
}

type Policy struct { // temporary field
	SegmentsList []string `yaml:"segments-list"` // temporary field
}
