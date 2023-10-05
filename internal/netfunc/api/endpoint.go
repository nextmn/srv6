// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package netfunc_api

import "github.com/nextmn/srv6/internal/iana"

type Endpoint interface {
	Name() string
	Behavior() iana.EndpointBehavior
}
