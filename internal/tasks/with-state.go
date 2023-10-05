// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

type WithState struct {
	state bool
}

func NewState() WithState {
	return WithState{
		state: false,
	}
}

func (ws *WithState) State() bool {
	return ws.state
}
