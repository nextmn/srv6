// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks

import (
	"context"
	"fmt"

	tasks_api "github.com/nextmn/srv6/internal/tasks/api"
)

// HookMulti is a Task that runs 2 SingleHook
type HookMulti struct {
	WithName
	WithState
	init tasks_api.TaskUnit
	exit tasks_api.TaskUnit
}

// Creates a new MultiHook
func NewMultiHook(init_name string, init *string, exit_name string, exit *string) *HookMulti {
	return &HookMulti{
		WithState: NewState(),
		init:      NewSingleHook(init_name, init),
		exit:      NewSingleHook(exit_name, exit),
	}
}

func (h *HookMulti) NameBase() string {
	return fmt.Sprintf("%s/%s", h.init.Name(), h.exit.Name())
}

func (h *HookMulti) NameInit() string {
	return h.init.Name()
}

func (h *HookMulti) NameExit() string {
	return h.exit.Name()
}

// Init function
func (h *HookMulti) RunInit(ctx context.Context) error {
	if h.init != nil {
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
