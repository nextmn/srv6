// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package mup

import "errors"

var (
	ErrTooShortToMarshal = errors.New("too short to serialize")
	ErrTooShortToParse   = errors.New("too short to parse")
)
