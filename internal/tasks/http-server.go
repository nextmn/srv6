// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	app_api "github.com/nextmn/srv6/internal/app/api"
	ctrl "github.com/nextmn/srv6/internal/ctrl"
	ctrl_api "github.com/nextmn/srv6/internal/ctrl/api"
)

// HttpServerTask starts an http server
type HttpServerTask struct {
	WithName
	WithState
	srv               *http.Server
	httpAddr          string
	rulesRegistry     ctrl_api.RulesRegistry
	rulesRegistryHTTP ctrl_api.RulesRegistryHTTP
	setupRegistry     app_api.Registry
}

// Create a new HttpServerTask
func NewHttpServerTask(name string, httpAddr string, setupRegistry app_api.Registry) *HttpServerTask {
	return &HttpServerTask{
		WithName:          NewName(name),
		WithState:         NewState(),
		srv:               nil,
		httpAddr:          httpAddr,
		rulesRegistry:     nil,
		rulesRegistryHTTP: nil,
		setupRegistry:     setupRegistry,
	}
}

// Init
func (t *HttpServerTask) RunInit() error {
	if t.setupRegistry == nil {
		return fmt.Errorf("Registry is nil")
	}
	db, ok := t.setupRegistry.DB()
	if !ok {
		return fmt.Errorf("DB is not in Registry")
	}
	rr := ctrl.NewRulesRegistry(db)
	t.rulesRegistry = rr
	t.rulesRegistryHTTP = rr
	r := gin.Default()
	r.GET("/status", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})
	r.POST("/rules", t.rulesRegistryHTTP.PostRule)
	r.GET("/rules/:uuid", t.rulesRegistryHTTP.GetRule)
	r.GET("/rules", t.rulesRegistryHTTP.GetRules)
	r.PATCH("/rules/:uuid/enable", t.rulesRegistryHTTP.EnableRule)
	r.PATCH("/rules/:uuid/disable", t.rulesRegistryHTTP.DisableRule)
	r.DELETE("/rules/:uuid", t.rulesRegistryHTTP.DeleteRule)
	t.srv = &http.Server{
		Addr:    t.httpAddr,
		Handler: r,
	}

	if t.setupRegistry != nil {
		t.setupRegistry.RegisterRulesRegistry(t.rulesRegistry)
	}

	go func() {
		if err := t.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err)
		}
	}()
	t.state = true
	return nil
}

// Exit
func (t *HttpServerTask) RunExit() error {
	t.state = false
	if t.setupRegistry != nil {
		t.setupRegistry.DeleteRulesRegistry()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := t.srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP Server Shutdown: %s\n", err)
	}
	return nil
}
