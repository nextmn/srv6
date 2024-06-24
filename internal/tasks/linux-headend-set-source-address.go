// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"github.com/nextmn/srv6/internal/iproute2"
)

// TaskLinuxHeadendSetSourceAddress
type TaskLinuxHeadendSetSourceAddress struct {
	WithName
	WithState
	address string
}

// Create a new TaskLinuxHeadendSetSourceAddress
func NewTaskLinuxHeadendSetSourceAddress(name string, address string) *TaskLinuxHeadendSetSourceAddress {
	return &TaskLinuxHeadendSetSourceAddress{
		WithName:  NewName(name),
		WithState: NewState(),
		address:   address,
	}
}

// Init
func (t *TaskLinuxHeadendSetSourceAddress) RunInit() error {
	if err := iproute2.IPSrSetSourceAddress(t.address); err != nil {
		return err
	}
	t.state = true
	return nil
}

// Exit
func (t *TaskLinuxHeadendSetSourceAddress) RunExit() error {
	// :: resets to default behavior
	if err := iproute2.IPSrSetSourceAddress("::"); err != nil {
		return err
	}
	t.state = false
	return nil
}
