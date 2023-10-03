// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks_api

// Pairs of Task to be run
type Task interface {
	RunInit() error
	RunExit() error
	State() bool // true when the initialized and not yet exited
}
