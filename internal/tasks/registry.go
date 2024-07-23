// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"fmt"
	"log"

	tasks_api "github.com/nextmn/srv6/internal/tasks/api"
)

type Registry struct {
	Tasks            []tasks_api.Task
	cancelFuncs      []context.CancelFunc
	initializedTasks int
}

func NewRegistry() *Registry {
	return &Registry{
		Tasks:            make([]tasks_api.Task, 0),
		cancelFuncs:      make([]context.CancelFunc, 0),
		initializedTasks: 0,
	}
}

// Register a new task
func (r *Registry) Register(task tasks_api.Task) {
	log.Printf("Task %s registered\n", task.NameBase())
	r.Tasks = append(r.Tasks, task)
}

// Run init tasks
func (r *Registry) RunInit(ctx context.Context) error {
	for _, t := range r.Tasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if t.State() {
				continue
			}
			taskCtx, cancel := context.WithCancel(ctx)
			r.cancelFuncs = append(r.cancelFuncs, cancel)
			if err := t.RunInit(taskCtx); err != nil {
				log.Printf("[Failed] %s: %s\n", t.NameInit(), err)
				return fmt.Errorf("Run init failure")
			}
			log.Printf("[OK] %s\n", t.NameInit())
		}
		r.initializedTasks += 1
	}
	return nil
}

// Run exit tasks
func (r *Registry) RunExit() {
	for i := len(r.cancelFuncs) - 1; i >= 0; i-- {
		r.cancelFuncs[i]()
	}
	for i := r.initializedTasks - 1; i >= 0; i-- {
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

func (r *Registry) Run(ctx context.Context) error {
	defer r.RunExit()
	if err := r.RunInit(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	}
}
