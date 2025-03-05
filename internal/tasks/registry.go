// Copyright 2023 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package tasks

import (
	"context"
	"fmt"
	"slices"

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
		"task-name":   task.NameBase(),
		"task-status": "registered",
	}).Info("Task registration")
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
				logrus.WithError(err).WithFields(logrus.Fields{
					"task-name":   t.NameInit(),
					"task-status": "failure",
				}).Error("Task runtime failure")
				return fmt.Errorf("Run init failure")
			}
			logrus.WithFields(logrus.Fields{
				"task-name":   t.NameInit(),
				"task-status": "success",
			}).Info("Task runtime success")
		}
		r.initializedTasks += 1
	}
	return nil
}

// Run exit tasks
func (r *Registry) RunExit() {
	for _, cancel := range slices.Backward(r.cancelFuncs) {
		cancel()
	}
	for _, t := range slices.Backward(r.Tasks[:r.initializedTasks]) {
		if !t.State() {
			continue
		}
		if err := t.RunExit(); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"task-name":   t.NameExit(),
				"task-status": "failure",
			}).Error("Task runtime failure")
		} else {
			logrus.WithFields(logrus.Fields{
				"task-name":   t.NameExit(),
				"task-status": "success",
			}).Info("Task runtime success")
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
