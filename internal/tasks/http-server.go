// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nextmn/srv6/internal/ctrl"
)

// HttpServerTask starts an http server
type HttpServerTask struct {
	WithName
	WithState
	srv           *http.Server
	rulesRegistry *ctrl.RulesRegistry
}

// Create a new HttpServerTask
func NewHttpServerTask(name string, httpAddr string, rr *ctrl.RulesRegistry) *HttpServerTask {
	r := gin.Default()
	r.GET("/status", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})
	r.POST("/rules", rr.PostRule)
	r.GET("/rules/:uuid", rr.GetRule)
	r.GET("/rules", rr.GetRules)
	r.PATCH("/rules/:uuid/enable", rr.EnableRule)
	r.PATCH("/rules/:uuid/disable", rr.DisableRule)
	r.DELETE("/rules/:uuid", rr.DeleteRule)
	return &HttpServerTask{
		WithName:  NewName(name),
		WithState: NewState(),
		srv: &http.Server{
			Addr:    httpAddr,
			Handler: r,
		},
		rulesRegistry: rr,
	}
}

// Init
func (t *HttpServerTask) RunInit() error {
	t.state = true
	go func() {
		if err := t.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err)
		}
	}()
	return nil
}

// Exit
func (t *HttpServerTask) RunExit() error {
	t.state = false
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := t.srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP Server Shutdown: %s\n", err)
	}
	return nil
}
