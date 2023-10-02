// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import api "github.com/louisroyer/nextmn-srv6/cmd/nextmn-srv6/internal/api/tasks"

// HookMulti is a Task that runs 2 SingleHook
type HookMulti struct {
	init  api.TaskUnit
	exit  api.TaskUnit
	state bool
}

// Creates a new MultiHook
func NewMultiHook(init *string, exit *string) HookMulti {
	return HookMulti{init: NewSingleHook(init), exit: NewSingleHook(exit)}
}

// Init function
func (h *HookMulti) RunInit() error {
	if h.exit != nil {
		if err := h.init.Run(); err != nil {
			return err
		}
		h.state = true
		return nil
	}
	return nil
}

// Exit function
func (h *HookMulti) RunExit() error {
	if h.exit != nil {
		if err := h.exit.Run(); err != nil {
			return err
		}
		h.state = false
		return nil
	}
	return nil
}

// State
func (h *HookMulti) State() bool {
	return h.state
}
