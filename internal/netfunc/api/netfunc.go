// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package netfunc

import (
	"context"

	"github.com/nextmn/srv6/internal/iproute2"
)

type NetFunc interface {
	Run(ctx context.Context, tunIface *iproute2.TunIface) error
}
