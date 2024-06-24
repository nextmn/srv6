// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import "fmt"

type WithName struct {
	name string
}

func NewName(name string) WithName {
	return WithName{
		name: name,
	}
}

func (wn *WithName) NameBase() string {
	return wn.name
}

func (wn *WithName) NameInit() string {
	return fmt.Sprintf("%s.init", wn.name)
}

func (wn *WithName) NameExit() string {
	return fmt.Sprintf("%s.exit", wn.name)
}
