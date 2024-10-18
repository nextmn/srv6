// Copyright 2024 Louis Royer and the NextMN-SRv6 contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nextmn/srv6/internal/config"
)

type Healthcheck struct {
	uri string
}

// TODO: move this in json-api
type Status struct {
	Ready bool `json:"ready"`
}

func NewHealthcheck(conf *config.SRv6Config) *Healthcheck {
	httpPort := "80" // default http port
	if conf.HTTPPort != nil {
		httpPort = *conf.HTTPPort
	}
	httpURI := "http://"
	if conf.HTTPAddress.Is6() {
		httpURI = httpURI + "[" + conf.HTTPAddress.String() + "]:" + httpPort
	} else {
		httpURI = httpURI + conf.HTTPAddress.String() + ":" + httpPort
	}
	return &Healthcheck{
		uri: httpURI,
	}
}
func (h *Healthcheck) Run(ctx context.Context) error {
	client := http.Client{
		Timeout: 100 * time.Millisecond,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.uri+"/status", nil)
	if err != nil {
		logrus.WithError(err).Error("Error while creating http get request")
		return err
	}
	req.Header.Add("User-Agent", "go-github-nextmn-srv6")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Charset", "utf-8")
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{"remote-server": h.uri}).WithError(err).Info("No http response")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logrus.WithFields(logrus.Fields{"remote-server": h.uri}).WithError(err).Info("Http response is not 200 OK")
		return err
	}
	decoder := json.NewDecoder(resp.Body)
	var status Status
	if err := decoder.Decode(&status); err != nil {
		logrus.WithFields(logrus.Fields{"remote-server": h.uri}).WithError(err).Info("Could not decode json response")
		return err
	}
	if !status.Ready {
		err := fmt.Errorf("Server is not ready")
		logrus.WithFields(logrus.Fields{"remote-server": h.uri}).WithError(err).Info("Server is not ready")
		return err
	}
	return nil
}
