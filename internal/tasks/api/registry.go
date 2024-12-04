// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks_api

import "context"

type Registry interface {
	Register(task Task)
	Run(ctx context.Context) error
}
