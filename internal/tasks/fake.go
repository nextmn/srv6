// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

// FakeTask is a dummy task that do nothing
type FakeTask struct {
	state bool
}

// Create a new DummyTask
func NewFakeTask() *FakeTask {
	return &FakeTask{}
}

// Init
func (t *FakeTask) RunInit() error {
	t.state = true
	return nil
}

// Exit
func (t *FakeTask) RunExit() error {
	t.state = false
	return nil
}

// Returns state of the task
func (t *FakeTask) State() bool {
	return t.state
}
