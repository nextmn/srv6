// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks

import "context"

// FakeTask is a dummy task that do nothing
type FakeTask struct {
	WithName
	WithState
}

// Create a new FakeTask
func NewFakeTask(name string) *FakeTask {
	return &FakeTask{
		WithName:  NewName(name),
		WithState: NewState(),
	}
}

// Init
func (t *FakeTask) RunInit(ctx context.Context) error {
	t.state = true
	return nil
}

// Exit
func (t *FakeTask) RunExit() error {
	t.state = false
	return nil
}
