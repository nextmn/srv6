// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"fmt"

	tasks_api "github.com/nextmn/srv6/internal/tasks/api"
	"github.com/sirupsen/logrus"
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
	logrus.WithFields(logrus.Fields{
		"name":   task.NameBase(),
		"status": "registered",
	}).Info("Task")
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
				logrus.WithFields(logrus.Fields{
					"name":   t.NameInit(),
					"status": "failure",
					"error":  err,
				}).Error("Task")
				return fmt.Errorf("Run init failure")
			}
			logrus.WithFields(logrus.Fields{
				"name":   t.NameInit(),
				"status": "success",
			}).Info("Task")
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
			logrus.WithFields(logrus.Fields{
				"name":   t.NameExit(),
				"status": "failure",
				"error":  err,
			}).Error("Task")
		} else {
			logrus.WithFields(logrus.Fields{
				"name":   t.NameExit(),
				"status": "success",
			}).Info("Task")
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
