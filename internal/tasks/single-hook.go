// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// HookSingle
type SingleHook struct {
	command *string
	name    string
}

// Creates a new SingleHook
func NewSingleHook(name string, cmd *string) SingleHook {
	return SingleHook{
		name:    name,
		command: cmd,
	}
}

func (h SingleHook) Name() string {
	return h.name
}

// Runs the command of the SingleHook
func (h SingleHook) Run() error {
	if h.command == nil {
		// nothing to do
		return nil
	}
	cmd := exec.Command(*h.command)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		errLog := fmt.Sprintf("Error running %s: %s", cmd.Args[0], err)
		log.Println(errLog)
		return err
	}
	return nil
}
