// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import tasks_api "github.com/nextmn/srv6/internal/tasks/api"
import (
	"fmt"
	"log"
)

type Registry struct {
	Tasks []tasks_api.Task
}

func NewRegistry() *Registry {
	return &Registry{
		Tasks: make([]tasks_api.Task, 0),
	}
}

// Register a new task
func (r *Registry) Register(task tasks_api.Task) {
	log.Printf("Task %s registered\n", task.NameBase())
	r.Tasks = append(r.Tasks, task)
}

// Run init tasks
func (r *Registry) RunInit() error {
	for _, t := range r.Tasks {
		if t.State() {
			continue
		}
		if err := t.RunInit(); err != nil {
			return fmt.Errorf("[Failed] %s: %s\n", t.NameInit(), err)
		}
		log.Printf("[OK] %s\n", t.NameInit())
	}
	return nil
}

// Run exit tasks
func (r *Registry) RunExit() {
	for i := len(r.Tasks) - 1; i >= 0; i-- {
		t := r.Tasks[i]
		if !t.State() {
			continue
		}
		if err := t.RunExit(); err != nil {
			log.Printf("[Failed] %s: %s\n", t.NameExit(), err)
		} else {
			log.Printf("[OK] %s\n", t.NameExit())
		}
	}
}
