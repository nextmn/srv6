// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

// HookMulti is a Task that runs 2 SingleHook
type HookMulti struct {
	init TaskUnit
	exit TaskUnit
}

// Creates a new MultiHook
func NewMultiHook(init *string, exit *string) HookMulti {
	return HookMulti{init: NewSingleHook(init), exit: NewSingleHook(exit)}
}

// Init function
func (h *HookMulti) RunInit() error {
	if h.exit != nil {
		return h.init.Run()
	}
	return nil
}

// Exit function
func (h *HookMulti) RunExit() error {
	if h.exit != nil {
		return h.exit.Run()
	}
	return nil
}
