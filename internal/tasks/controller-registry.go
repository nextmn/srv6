// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	app_api "github.com/nextmn/srv6/internal/app/api"
	"github.com/nextmn/srv6/internal/ctrl"

	"github.com/nextmn/json-api/jsonapi"
	"github.com/nextmn/json-api/jsonapi/n4tosrv6"
)

const UserAgent = "go-github-nextmn-srv6"

// ControllerRegistry registers and unregisters into controller
type ControllerRegistryTask struct {
	WithName
	WithState
	ControllerRegistry *ctrl.ControllerRegistry
	SetupRegistry      app_api.Registry
	httpClient         http.Client
}

// Create a new ControllerRegistry
func NewControllerRegistryTask(name string, remoteControlURI jsonapi.ControlURI, backbone n4tosrv6.BackboneIP, locator n4tosrv6.Locator, localControlURI jsonapi.ControlURI, setup_registry app_api.Registry) *ControllerRegistryTask {
	return &ControllerRegistryTask{
		WithName:  NewName(name),
		WithState: NewState(),
		ControllerRegistry: &ctrl.ControllerRegistry{
			RemoteControlURI: remoteControlURI,
			LocalControlURI:  localControlURI,
			Locator:          locator,
			Backbone:         backbone,
			Resource:         "",
		},
		SetupRegistry: setup_registry,
		httpClient:    http.Client{},
	}
}

// Init
func (t *ControllerRegistryTask) RunInit(ctx context.Context) error {
	if t.SetupRegistry != nil {
		t.SetupRegistry.RegisterControllerRegistry(t.ControllerRegistry)
	} else {
		return fmt.Errorf("could not register controllerregistry")
	}
	data := n4tosrv6.Router{
		Locator:  t.ControllerRegistry.Locator,
		Backbone: t.ControllerRegistry.Backbone,
		Control:  t.ControllerRegistry.LocalControlURI,
	}
	reqBody, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// TODO: retry on timeout failure (use a new ctx)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.ControllerRegistry.RemoteControlURI.JoinPath("routers").String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 400 {
		return fmt.Errorf("HTTP Bad request\n")
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("HTTP Control Server: internal error\n")
	}
	if resp.StatusCode == 201 { // created
		t.ControllerRegistry.Resource = resp.Header.Get("Location")
	}

	t.state = true
	return nil
}

// Exit
func (t *ControllerRegistryTask) RunExit() error {
	// TODO: retry on timeout failure
	// TODO: if Resource has scheme, don't concatenate
	if t.SetupRegistry != nil {
		t.SetupRegistry.DeleteControllerRegistry()
	}

	if t.ControllerRegistry.Resource == "" {
		// nothing to do
		t.state = false
		return nil
	}
	// no context since Background Context is already Done
	req, err := http.NewRequest(http.MethodDelete, t.ControllerRegistry.RemoteControlURI.JoinPath(t.ControllerRegistry.Resource).String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", UserAgent)
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 400 {
		return fmt.Errorf("HTTP Bad request\n")
	}
	if resp.StatusCode >= 500 {
		return fmt.Errorf("HTTP Control Server: internal error %v\n", resp.Body)
	}
	t.state = false
	return nil
}
