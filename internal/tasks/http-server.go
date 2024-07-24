// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"context"
	"fmt"
	"net/http"
	"time"

	app_api "github.com/nextmn/srv6/internal/app/api"
	ctrl_api "github.com/nextmn/srv6/internal/ctrl/api"
	"github.com/sirupsen/logrus"

	"github.com/nextmn/srv6/internal/ctrl"

	"github.com/gin-gonic/gin"
)

// HttpServerTask starts an http server
type HttpServerTask struct {
	WithName
	WithState
	srv               *http.Server
	httpAddr          string
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
		rulesRegistryHTTP: nil,
		setupRegistry:     setupRegistry,
	}
}

// Init
func (t *HttpServerTask) RunInit(ctx context.Context) error {
	if t.setupRegistry == nil {
		return fmt.Errorf("Registry is nil")
	}
	db, ok := t.setupRegistry.DB()
	if !ok {
		return fmt.Errorf("DB is not in Registry")
	}
	rr := ctrl.NewRulesRegistry(db)
	t.rulesRegistryHTTP = rr
	// TODO:  gin.SetMode(gin.ReleaseMode)
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

	go func() {
		if err := t.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Error("HTTP Server error: ", err)
		}
	}()
	t.state = true
	return nil
}

// Exit
func (t *HttpServerTask) RunExit() error {
	t.state = false
	ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second) // context.Background() is already Done()
	defer cancel()
	if err := t.srv.Shutdown(ctx); err != nil {
		logrus.Info("HTTP Server Shutdown: ", err)
	}
	return nil
}
