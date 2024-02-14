// Copyright 2023 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
)

// ControllerRegistry registers and unregisters into controller
type ControllerRegistry struct {
	WithState
	RemoteControlURI string // URI of the controller
	LocalControlURI  string // URI of the router, used to control it
	Locator          string
	Backbone         netip.Addr
	Resource         string
}

// Create a new ControllerRegistry
func NewControllerRegistry(remoteControlURI string, backbone netip.Addr, locator string, localControlURI string) *ControllerRegistry {
	return &ControllerRegistry{
		WithState:        NewState(),
		RemoteControlURI: remoteControlURI,
		LocalControlURI:  localControlURI,
		Locator:          locator,
		Backbone:         backbone,
		Resource:         "",
	}
}

// Init
func (t *ControllerRegistry) RunInit() error {
	data := map[string]string{
		"locator":  t.Locator,
		"backbone": t.Backbone.String(),
		"control":  t.LocalControlURI,
	}
	json_data, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// TODO: retry on timeout failure
	resp, err := http.Post(t.RemoteControlURI+"/routers", "application/json", bytes.NewBuffer(json_data))
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
		t.Resource = resp.Header.Get("Location")
	}

	t.state = true
	return nil
}

// Exit
func (t *ControllerRegistry) RunExit() error {
	// TODO: retry on timeout failure
	// TODO: if Resource has scheme, don't concatenate
	if t.Resource == "" {
		// nothing to do
		t.state = false
		return nil
	}
	req, err := http.NewRequest("DELETE", t.RemoteControlURI+t.Resource, nil)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
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
