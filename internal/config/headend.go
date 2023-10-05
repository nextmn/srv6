// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

type Headend struct {
	Provider      Provider
	Type          HeadendType
	Policy        *Policy // temporary field
	SourceAddress string  `yaml:"set-source-address"`
}

type Headends []*Headend

func (he Headends) Filter(provider Provider) Headends {
	newList := make([]*Headend, 0)
	for _, e := range he {
		if e.Provider == provider {
			newList = append(newList, e)
		}
	}
	return newList
}
