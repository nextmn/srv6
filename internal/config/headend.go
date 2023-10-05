// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

type HeadEnd struct {
	Provider      Provider
	Type          HeadEndType
	Policy        *Policy // temporary field
	SourceAddress string  `yaml:"set-source-address"`
}

func (he []*HeadEnd) Filter(provider Provider) []*HeadEnd {
	newList := make([]*HeadEnd, 0)
	for _, e := range conf.HeadEnd {
		if e.Provider == provider {
			newList := append(newList, e)
		}
	}
	return &newList
}
