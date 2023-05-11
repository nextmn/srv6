// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

var SRv6 *SRv6Config

func Run() error {
	ipRoute2Init()
	linuxSRInit()
	goSRInit()
	for {
		select {}
	}
	return nil
}

func Exit() error {
	goSRExit()
	linuxSRExit()
	ipRoute2Exit()
	return nil
}
