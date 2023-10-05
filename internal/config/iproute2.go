// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package config

type IPRoute2 struct {
	PreInitHook  *string `yaml:"pre-init-hook,omitempty"`  // script to execute before interfaces are configured
	PostInitHook *string `yaml:"post-init-hook,omitempty"` // script to execute after interfaces are configured
	PreExitHook  *string `yaml:"pre-exit-hook,omitempty"`  // script to execute before interfaces are configured
	PostExitHook *string `yaml:"post-exit-hook,omitempty"` // script to execute after interfaces are configured
}
