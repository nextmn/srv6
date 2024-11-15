// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package config

import (
	"github.com/nextmn/json-api/jsonapi"
)

type Control struct {
	Uri      jsonapi.ControlURI `yaml:"uri"`       // may contain domain name instead of ip address
	BindAddr string             `yaml:"bind-addr"` // in the form `ip:port`
}
