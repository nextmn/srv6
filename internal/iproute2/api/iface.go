// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package iproute2_api

type Iface interface {
	CreateAndUp() error
	Delete() error
	Name() string
}
