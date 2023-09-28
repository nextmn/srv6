// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package srv6

import "log"

var SRv6 *SRv6Config

func Run() error {
	log.Println("Starting…")
	if err := runHook(SRv6.IPRoute2.PreInitHook); err != nil {
		return err
	}
	if err := ipRoute2Init(); err != nil {
		return err
	}
	if err := linuxSRInit(); err != nil {
		return err
	}
	if err := goSRInit(); err != nil {
		return err
	}
	if err := runHook(SRv6.IPRoute2.PostInitHook); err != nil {
		return err
	}

	// Sleep infinity
	select {}

	return nil
}

func Exit() error {
	goSRExit()
	linuxSRExit()
	ipRoute2Exit()
	return nil
}
